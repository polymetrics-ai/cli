package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	// maxOperationBinaryDownloadBytes is the absolute ceiling no operation
	// and no caller can exceed.
	//
	// CAPTAIN DECISION, PENDING: the default ceiling is deliberately set to
	// the largest value any bundle currently declares (100 MiB, declared by
	// 56 of the 83 binary operations) so that no already-declared operation
	// is silently unsatisfiable. connsdk's buffered path caps at 64 MiB,
	// which is why DoStream exists. A lower ceiling (e.g. for laptops) is a
	// policy choice, not a technical constraint — change this constant.
	maxOperationBinaryDownloadBytes = 100 << 20

	// defaultBinaryDownloadStallTimeout bounds how long a download may make
	// NO progress. It is deliberately not a wall-clock deadline: reusing the
	// JSON path's 30s wall clock would turn the byte cap into a bandwidth
	// requirement, failing a legitimate 100 MiB download on a slow link.
	defaultBinaryDownloadStallTimeout = 60 * time.Second

	// binaryDownloadSniffBytes matches http.DetectContentType's window.
	binaryDownloadSniffBytes = 512

	binaryDownloadFileMode = 0o600
)

// BinaryDownloadRequest is one bounded binary/file download.
//
// It stays engine-local on purpose: connectors.OperationBinaryDownloadRequest
// is the cross-package contract, and Connector.OperationBinaryDownload is the
// seam between them, so the executor's own shape can change without moving the
// connectors package.
type BinaryDownloadRequest struct {
	Operation    string
	Config       connectors.RuntimeConfig
	PathParams   map[string]string
	Query        map[string]string
	Headers      map[string]string
	HeaderValues map[string][]string
	// MaxBytes optionally lowers the operation's declared cap. It can never
	// raise it.
	MaxBytes int64
	// DestRoot is the directory downloads are confined to. Required: there is
	// no implicit destination.
	DestRoot string
	// FileName optionally names the file within DestRoot. It must be a local,
	// single-segment name; traversal is refused.
	FileName string
	// RedactFields names record fields to redact before the record is emitted.
	// It matters here because source_ref carries the resolved endpoint path,
	// including whatever path parameters the caller supplied.
	RedactFields []string
}

// BinaryDownloadResult carries the flat record describing what landed on disk.
// Bytes are never inlined: records are flat map[string]any and pass through
// schema projection, and a 25 MiB attachment would become a 34 MiB JSON line.
type BinaryDownloadResult struct {
	Connector string
	Operation string
	Method    string
	Path      string
	Record    connectors.Record
	Status    int
	Headers   map[string]connectors.OperationResponseHeader
	Receipt   *connectors.ProviderResponseReceipt
}

// OperationBinaryDownload executes a declared binary_download or text_export
// operation, streaming the response to a file confined beneath req.DestRoot.
//
// This is the executor half of declaration-owned file operations: both kinds
// are in the schema enum and block map, BinaryOperationSpec already carries
// method/path/max_bytes/allow_overwrite/extract_archives, and GET-only +
// positive-max_bytes validation already runs at bundle load. text_export is the
// closed CSV variant and still produces a destination manifest, never stdout.
//
// Bounded by construction:
//   - the body is read one byte PAST the limit and rejected on overflow, so a
//     truncated file is never mistaken for a complete one;
//   - limits clamp request -> spec -> ceiling, never upward;
//   - the endpoint must be declared in api_surface;
//   - all filesystem access goes through os.Root, which refuses traversal and
//     escaping symlinks and closes the TOCTOU race that a lexical path check
//     (safety.ValidateLocalWritePath) cannot;
//   - extract_archives is refused outright.
//
// pmcert:executes binary_download,text_export
func OperationBinaryDownload(ctx context.Context, b Bundle, req BinaryDownloadRequest, h Hooks) (BinaryDownloadResult, error) {
	if err := ctx.Err(); err != nil {
		return BinaryDownloadResult{}, err
	}
	op, err := operationBinaryDownloadSpec(b, req.Operation)
	if err != nil {
		return BinaryDownloadResult{}, err
	}
	spec := op.Binary
	redirect, err := operationRedirectPolicy(op)
	if err != nil {
		return BinaryDownloadResult{}, err
	}
	if strings.TrimSpace(req.DestRoot) == "" {
		return BinaryDownloadResult{}, fmt.Errorf("binary download requires a destination root")
	}

	cfg := materializeConfigDefaults(b, req.Config)
	effectivePathParams, err := materializeOperationBinaryDownloadPathParams(op, cfg, req.PathParams)
	if err != nil {
		return BinaryDownloadResult{}, err
	}
	resolvedPath, err := resolveSurfaceEndpointPath(spec.Path, cfg, effectivePathParams)
	if err != nil {
		return BinaryDownloadResult{}, err
	}
	queryMap, err := operationBinaryDownloadQuery(op, req.Query)
	if err != nil {
		return BinaryDownloadResult{}, err
	}
	query, err := directReadQuery(queryMap)
	if err != nil {
		return BinaryDownloadResult{}, err
	}
	headers, err := operationRequestHeaders(b, op, req.Headers, req.HeaderValues)
	if err != nil {
		return BinaryDownloadResult{}, err
	}
	maxBytes := clampOperationBinaryDownloadMaxBytes(req.MaxBytes, spec.MaxBytes)

	stall := defaultBinaryDownloadStallTimeout
	if spec.StallTimeoutSeconds > 0 {
		stall = time.Duration(spec.StallTimeoutSeconds) * time.Second
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	baseURL, err := resolveOperationRoute(b, cfg, op.Route, op.ID, spec.Path, op.SourceURL)
	if err != nil {
		return BinaryDownloadResult{}, err
	}
	rt, err := newRuntimeForOperationRoute(ctx, b, cfg, h, op.Route, op.ID, spec.Path, op.SourceURL)
	if err != nil {
		return BinaryDownloadResult{}, err
	}

	requestPath := normalizeDirectReadPathForBaseURL(resolvedPath, baseURL)
	requester, err := rt.requesterFor(http.MethodGet, spec.Path)
	if err != nil {
		return BinaryDownloadResult{}, err
	}
	requester, err = requesterWithOperationHeaders(requester, op, headers)
	if err != nil {
		return BinaryDownloadResult{}, err
	}
	resp, err := requester.DoStream(ctx, http.MethodGet, requestPath, query, connsdk.StreamOptions{
		Accept:         spec.Accept,
		RedirectPolicy: redirect,
	})
	if err != nil && resp == nil {
		result := BinaryDownloadResult{Connector: b.Name, Operation: op.ID, Method: http.MethodGet, Path: resolvedPath}
		result.Receipt = providerResponseReceiptFromHTTPError(b, err, cfg.Secrets)
		if result.Receipt != nil {
			result.Status = result.Receipt.Status
		}
		class, hint := applyErrorMap(b.HTTP.ErrorMap, err)
		msg := completeEngineErrorText(err)
		if hint != "" {
			msg = msg + ": " + hint
		}
		if class != "" {
			msg = class + ": " + msg
		}
		return result, formatResponseError(fmt.Sprintf("binary download GET %s: %s", spec.Path, msg), err)
	}
	defer func() { _ = resp.Body.Close() }()
	responseReceipt := connectors.ProviderResponseReceipt{
		ResponseReceived: true,
		Status:           resp.Status,
		Headers:          completeProviderResponseHeaders(b, resp.Header),
	}
	responseReceipt = connectors.SanitizeProviderResponseReceiptForOutput(responseReceipt, cfg.Secrets)
	result := BinaryDownloadResult{
		Connector: b.Name,
		Operation: op.ID,
		Method:    http.MethodGet,
		Path:      resolvedPath,
		Status:    resp.Status,
		Receipt:   &responseReceipt,
	}
	if err != nil {
		captureErr := captureBinaryResponseMetadata(resp.Body, result.Receipt, maxBytes, stall, cancel)
		return result, errors.Join(err, captureErr)
	}
	if err := validateOperationBinaryResponseMediaType(op, resp.Header); err != nil {
		captureErr := captureBinaryResponseMetadata(resp.Body, result.Receipt, maxBytes, stall, cancel)
		return result, errors.Join(err, captureErr)
	}
	responseHeaders, err := operationResponseHeaders(b, op, resp.Header, cfg.Secrets)
	if err != nil {
		captureErr := captureBinaryResponseMetadata(resp.Body, result.Receipt, maxBytes, stall, cancel)
		return result, errors.Join(err, captureErr)
	}
	result.Headers = responseHeaders

	fileName, err := resolveBinaryDownloadFileName(req.FileName, resp.Header.Get("Content-Disposition"), op.ID)
	if err != nil {
		captureErr := captureBinaryResponseMetadata(resp.Body, result.Receipt, maxBytes, stall, cancel)
		return result, errors.Join(err, captureErr)
	}

	written, digest, sniffed, err := streamBinaryDownloadToRoot(resp.Body, req.DestRoot, fileName, maxBytes, spec.AllowOverwrite, stall, cancel)
	if err != nil {
		result.Receipt.BodyPresent = written != 0
		result.Receipt.BodyBytes = written
		return result, err
	}

	record := redactBinaryDownloadRecord(connectors.Record{
		"file_path":       filepath.Join(req.DestRoot, fileName),
		"file_name":       fileName,
		"file_size_bytes": written,
		"file_sha256":     digest,
		"content_type":    resp.Header.Get("Content-Type"),
		// Sniffed independently: never trust Content-Type and never infer
		// from the URL path. One provider serves CSV bytes from a path
		// ending .json. The mismatch is surfaced, not rejected.
		"content_type_sniffed": sniffed,
		"source_operation":     op.ID,
		// The connector-relative source reference carries no signed-URL
		// credentials and remains stable across public receipt projection.
		"source_ref":    resolvedPath,
		"downloaded_at": time.Now().UTC().Format(time.RFC3339),
		// Always false: overflow is a hard error rather than a silent
		// truncation. The field exists so consumers can rely on it and so
		// a future ranged/resumable mode has somewhere to report.
		"truncated": false,
	}, req.RedactFields)
	receipt := connectors.ProviderResponseReceipt{
		ResponseReceived: true,
		Status:           resp.Status,
		Headers:          completeProviderResponseHeaders(b, resp.Header),
		BodyPresent:      written != 0,
		BodyBytes:        written,
		Body:             map[string]any{"file_size_bytes": written, "file_sha256": digest},
	}
	receipt = connectors.SanitizeProviderResponseReceiptForOutput(receipt, cfg.Secrets)
	result.Record = record
	result.Receipt = &receipt
	return result, nil
}

func captureBinaryResponseMetadata(body io.Reader, receipt *connectors.ProviderResponseReceipt, maxBytes int64, stall time.Duration, cancel context.CancelFunc) error {
	if body == nil || receipt == nil {
		return nil
	}
	written, err := io.Copy(io.Discard, io.LimitReader(newStallReader(body, stall, cancel), maxBytes+1))
	receipt.BodyPresent = written != 0
	receipt.BodyBytes = written
	if err != nil {
		return fmt.Errorf("capture binary response metadata: %w", err)
	}
	if written > maxBytes {
		return fmt.Errorf("binary response metadata exceeds limit %d bytes", maxBytes)
	}
	return nil
}

func PreflightOperationBinaryDownload(b Bundle, operation, method, path string) error {
	op, err := operationBinaryDownloadSpec(b, operation)
	if err != nil {
		return err
	}
	if !strings.EqualFold(strings.TrimSpace(method), http.MethodGet) {
		return fmt.Errorf("binary download command requires GET")
	}
	if path != op.Binary.Path {
		return fmt.Errorf("binary download command path %q does not match declared operation path %q", path, op.Binary.Path)
	}
	return nil
}

func operationBinaryDownloadSpec(b Bundle, operation string) (OperationSpec, error) {
	op, err := findOperation(b, operation)
	if err != nil {
		return OperationSpec{}, err
	}
	if (op.Kind != "binary_download" && op.Kind != "text_export") || op.Binary == nil {
		return OperationSpec{}, fmt.Errorf("file download requires binary_download or text_export operation, got %q", op.Kind)
	}
	if err := validateOperationRouteForOperation(b, op.Route, op.ID, op.Binary.Path, op.SourceURL); err != nil {
		return OperationSpec{}, err
	}
	spec := op.Binary
	if err := requireOperationBinaryResponseContract(op); err != nil {
		return OperationSpec{}, err
	}
	if _, err := requireOperationSuccessStatusPolicy(op); err != nil {
		return OperationSpec{}, err
	}
	if _, err := operationRedirectPolicy(op); err != nil {
		return OperationSpec{}, err
	}
	if op.Kind == "text_export" {
		if spec.MaxBytes <= 0 {
			return OperationSpec{}, fmt.Errorf("text export requires positive max_bytes")
		}
		if !strings.EqualFold(strings.TrimSpace(spec.Accept), "text/csv") {
			return OperationSpec{}, fmt.Errorf("text export requires the closed text/csv accept contract")
		}
	}
	if spec.ExtractArchives {
		return OperationSpec{}, fmt.Errorf("operation %q declares extract_archives, which is not supported: archive extraction is a separate capability", op.ID)
	}
	if method := strings.ToUpper(strings.TrimSpace(spec.Method)); method != http.MethodGet {
		return OperationSpec{}, fmt.Errorf("binary download requires GET, got %s", method)
	}
	if isAbsoluteHTTPURL(spec.Path) {
		return OperationSpec{}, fmt.Errorf("binary download endpoint must be connector-relative, got absolute URL")
	}
	if err := requireOperationSurfaceEndpoint(b, http.MethodGet, spec.Path); err != nil {
		return OperationSpec{}, err
	}
	return op, nil
}

func validateOperationBinaryContentTypes(op OperationSpec) error {
	if op.Binary == nil {
		return fmt.Errorf("operation has no binary declaration")
	}
	for _, declared := range op.Binary.ContentTypes {
		mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(declared))
		if err != nil || !validOperationMediaRange(mediaType) {
			return fmt.Errorf("content type %q is not a valid media type", declared)
		}
	}
	if charset := strings.TrimSpace(op.Binary.Charset); charset != "" {
		if _, _, err := mime.ParseMediaType("text/plain; charset=" + charset); err != nil {
			return fmt.Errorf("charset %q is invalid", op.Binary.Charset)
		}
	}
	return nil
}

func requireOperationBinaryResponseContract(op OperationSpec) error {
	if op.Binary == nil || len(op.Binary.ContentTypes) == 0 {
		return fmt.Errorf("operation %q requires non-empty declared response content_types", op.ID)
	}
	if op.Kind == "text_export" && strings.TrimSpace(op.Binary.Charset) == "" {
		return fmt.Errorf("operation %q requires declared response charset", op.ID)
	}
	return validateOperationBinaryContentTypes(op)
}

func validOperationMediaRange(mediaType string) bool {
	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || parts[0] == "*" {
		return false
	}
	return !strings.Contains(parts[0], "*") && (parts[1] == "*" || !strings.Contains(parts[1], "*"))
}

// validateOperationBinaryResponseMediaType rejects an undeclared response
// before a destination file is created. This keeps media policy bounded while
// preserving every ordinary field of a response the declaration admits.
func validateOperationBinaryResponseMediaType(op OperationSpec, headers http.Header) error {
	if err := requireOperationBinaryResponseContract(op); err != nil {
		return fmt.Errorf("operation %q response media declaration: %w", op.ID, err)
	}
	mediaType, params, err := mime.ParseMediaType(headers.Get("Content-Type"))
	if err != nil || mediaType == "" {
		return fmt.Errorf("operation %q response has no valid Content-Type", op.ID)
	}
	if op.Kind == "text_export" && !strings.EqualFold(mediaType, "text/csv") {
		return fmt.Errorf("text export response is not text/csv")
	}
	matched := false
	for _, declared := range op.Binary.ContentTypes {
		allowed, declaredParams, _ := mime.ParseMediaType(strings.TrimSpace(declared))
		allowedParts := strings.Split(allowed, "/")
		actualParts := strings.Split(mediaType, "/")
		if len(allowedParts) == 2 && len(actualParts) == 2 && strings.EqualFold(actualParts[0], allowedParts[0]) && (allowedParts[1] == "*" || strings.EqualFold(actualParts[1], allowedParts[1])) && operationMediaParametersMatch(declaredParams, params, op.Binary.Charset) {
			matched = true
			break
		}
	}
	if !matched {
		return fmt.Errorf("operation %q response media type or parameters %q are not declared", op.ID, headers.Get("Content-Type"))
	}
	return nil
}

func operationMediaParametersMatch(declared, actual map[string]string, charset string) bool {
	expected := make(map[string]string, len(declared)+1)
	for name, value := range declared {
		expected[strings.ToLower(strings.TrimSpace(name))] = value
	}
	if charset = strings.TrimSpace(charset); charset != "" {
		if declaredCharset, present := expected["charset"]; present && !strings.EqualFold(declaredCharset, charset) {
			return false
		}
		expected["charset"] = charset
	}
	if len(expected) != len(actual) {
		return false
	}
	for name, value := range actual {
		expectedValue, present := expected[strings.ToLower(strings.TrimSpace(name))]
		if !present {
			return false
		}
		if strings.EqualFold(name, "charset") {
			if !strings.EqualFold(value, expectedValue) {
				return false
			}
			continue
		}
		if value != expectedValue {
			return false
		}
	}
	return true
}

// redactBinaryDownloadRecord applies the command's declared redact_fields to
// the record describing the download, using the same field-name matching as a
// direct read so one declaration means the same thing on both executors.
//
// The record is flat, so the result is always a flat record again.
func redactBinaryDownloadRecord(record connectors.Record, fields []string) connectors.Record {
	if len(fields) == 0 {
		return record
	}
	redacted, ok := redactNamedJSONFields(map[string]any(record), fields).(map[string]any)
	if !ok {
		return record
	}
	return connectors.Record(redacted)
}

// clampOperationBinaryDownloadMaxBytes clamps request -> spec -> ceiling,
// mirroring clampOperationDirectReadMaxBytes. A caller can only ever lower the
// limit.
func clampOperationBinaryDownloadMaxBytes(requested int64, operationMax int) int64 {
	maxBytes := requested
	if maxBytes <= 0 {
		maxBytes = int64(operationMax)
	}
	if maxBytes <= 0 || maxBytes > maxOperationBinaryDownloadBytes {
		maxBytes = maxOperationBinaryDownloadBytes
	}
	if operationMax > 0 && maxBytes > int64(operationMax) {
		return int64(operationMax)
	}
	return maxBytes
}

// streamBinaryDownloadToRoot writes body into a file confined beneath destRoot,
// bounded to maxBytes, returning the byte count, hex SHA-256, and sniffed
// content type.
//
// Every filesystem operation goes through os.Root, which refuses traversal and
// escaping symlinks. The bytes land in an owned hidden temp file inside the
// SAME root, are fsync'd, then publish with either a replacing rename (when
// explicitly allowed) or an atomic hard-link claim (when no overwrite was
// approved). The containing directory is synced after the name transition.
func streamBinaryDownloadToRoot(body io.Reader, destRoot, fileName string, maxBytes int64, allowOverwrite bool, stall time.Duration, cancel context.CancelFunc) (int64, string, string, error) {
	root, err := os.OpenRoot(destRoot)
	if err != nil {
		return 0, "", "", fmt.Errorf("binary download destination: %w", err)
	}
	defer func() { _ = root.Close() }()

	// Do not reserve fileName while staging: a process death must leave the
	// visible final name absent. A hidden temp is owned solely by this attempt;
	// its eventual hard-link publication below atomically claims fileName only
	// if no competing final exists.
	tempName := "." + fileName + ".part-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	temp, err := root.OpenFile(tempName, os.O_CREATE|os.O_EXCL|os.O_WRONLY, binaryDownloadFileMode)
	if err != nil {
		return 0, "", "", fmt.Errorf("binary download temp file: %w", err)
	}
	cleanup := func() {
		_ = temp.Close()
		_ = root.Remove(tempName)
	}

	hash := sha256.New()
	sniff := &sniffBuffer{}
	// Read ONE byte past the limit: io.LimitReader signals exhaustion with
	// EOF, not an error, so a download stopping exactly at the cap is
	// indistinguishable from a complete one unless we look for the extra byte.
	limited := io.LimitReader(newStallReader(body, stall, cancel), maxBytes+1)
	written, err := io.Copy(io.MultiWriter(temp, hash, sniff), limited)
	if err != nil {
		cleanup()
		return written, "", "", fmt.Errorf("binary download: %w", err)
	}
	if written > maxBytes {
		cleanup()
		return written, "", "", fmt.Errorf("binary download response too large: exceeds limit %d bytes", maxBytes)
	}
	// fsync before rename, or the rename can yield a zero-length file.
	if err := temp.Sync(); err != nil {
		cleanup()
		return 0, "", "", fmt.Errorf("binary download sync: %w", err)
	}
	if err := temp.Close(); err != nil {
		_ = root.Remove(tempName)
		return 0, "", "", fmt.Errorf("binary download close: %w", err)
	}
	if allowOverwrite {
		if err := root.Rename(tempName, fileName); err != nil {
			_ = root.Remove(tempName)
			return 0, "", "", fmt.Errorf("binary download rename: %w", err)
		}
	} else if err := root.Link(tempName, fileName); err != nil {
		_ = root.Remove(tempName)
		return 0, "", "", fmt.Errorf("binary download publish without overwrite: %w", err)
	} else if err := root.Remove(tempName); err != nil {
		// The final link is already published. Never remove it on a cleanup
		// error: doing so could delete the just-created artifact or a foreign
		// replacement. Report the owned-temp cleanup failure for recovery.
		return 0, "", "", fmt.Errorf("binary download publish cleanup: %w", err)
	}
	if err := syncBinaryDownloadDirectory(root); err != nil {
		// Publication is durable enough to be visible but its directory entry
		// was not confirmed to stable storage. Preserve the artifact; callers
		// can inspect it and a retry will safely collide rather than overwrite.
		return 0, "", "", fmt.Errorf("binary download directory sync: %w", err)
	}
	return written, hex.EncodeToString(hash.Sum(nil)), http.DetectContentType(sniff.head), nil
}

func syncBinaryDownloadDirectory(root *os.Root) error {
	directory, err := root.Open(".")
	if err != nil {
		return err
	}
	defer func() { _ = directory.Close() }()
	return directory.Sync()
}

// sniffBuffer captures only the first 512 bytes, matching
// http.DetectContentType's window, so sniffing costs no extra I/O and no
// unbounded memory.
type sniffBuffer struct{ head []byte }

func (s *sniffBuffer) Write(p []byte) (int, error) {
	if n := binaryDownloadSniffBytes - len(s.head); n > 0 {
		if n > len(p) {
			n = len(p)
		}
		s.head = append(s.head, p[:n]...)
	}
	return len(p), nil
}

// stallReader cancels the request when NO progress is made for the stall
// window. This is deliberately a stall timeout rather than a wall-clock
// deadline: a wall clock would turn the byte cap into a bandwidth requirement
// and fail legitimate large downloads on slow links.
type stallReader struct {
	reader   io.Reader
	lastRead atomic.Int64
	done     chan struct{}
}

// newStallReader wraps r and calls cancel when no bytes arrive for the stall
// window.
//
// cancel MUST be the cancel func of the context the in-flight HTTP request was
// built with. Deriving a fresh context here instead would produce a watchdog
// that fires correctly and yet aborts nothing, because the request body would
// still be governed by the original context — the read would hang on until the
// client's own timeout.
func newStallReader(r io.Reader, stall time.Duration, cancel context.CancelFunc) io.Reader {
	if stall <= 0 || cancel == nil {
		return r
	}
	sr := &stallReader{reader: r, done: make(chan struct{})}
	sr.lastRead.Store(time.Now().UnixNano())
	tick := stall / 4
	if tick <= 0 {
		tick = stall
	}
	go func() {
		ticker := time.NewTicker(tick)
		defer ticker.Stop()
		for {
			select {
			case <-sr.done:
				return
			case now := <-ticker.C:
				if now.UnixNano()-sr.lastRead.Load() > int64(stall) {
					cancel()
					return
				}
			}
		}
	}()
	return sr
}

func (s *stallReader) Read(p []byte) (int, error) {
	n, err := s.reader.Read(p)
	if n > 0 {
		s.lastRead.Store(time.Now().UnixNano())
	}
	if err != nil {
		s.stop()
	}
	return n, err
}

func (s *stallReader) stop() {
	select {
	case <-s.done:
	default:
		close(s.done)
	}
}

// resolveBinaryDownloadFileName derives a safe, root-local file name.
//
// Preference order: an explicit caller name, then the provider's
// Content-Disposition, then a connector-controlled identifier derived from the
// operation id. Provider text is never trusted:
//
//   - mime.ParseMediaType DOES decode RFC 5987/6266 filename*, and the decoded
//     value lands under the UNSTARRED key, so params["filename"] is read and
//     params["filename*"] never is;
//   - RFC 6266 counts BOTH / and \ as path separators, and filepath.Base on
//     Linux returns `..\..\etc\passwd` unchanged, so both are stripped first;
//   - filepath.Localize and filepath.IsLocal then reject anything that is not a
//     single local segment.
//
// A caller-supplied name that fails these checks is an ERROR (the caller asked
// for something specific and must be told it was refused), while unusable
// provider text falls back to the operation-derived name.
func resolveBinaryDownloadFileName(callerName, contentDisposition, operationID string) (string, error) {
	if strings.TrimSpace(callerName) != "" {
		// A caller-supplied name is validated STRICTLY, never rewritten. The
		// caller asked for something specific, so silently basename-ing
		// "../escape.txt" into "escape.txt" would hide a traversal attempt
		// instead of reporting it.
		if !isLocalSingleSegment(callerName) {
			return "", fmt.Errorf("binary download file name %q must be a single local file name", callerName)
		}
		return callerName, nil
	}
	if contentDisposition != "" {
		if _, params, err := mime.ParseMediaType(contentDisposition); err == nil {
			if name, ok := sanitizeBinaryDownloadFileName(params["filename"]); ok {
				return name, nil
			}
		}
	}
	name, ok := sanitizeBinaryDownloadFileName(strings.ReplaceAll(operationID, ".", "_"))
	if !ok {
		return "", fmt.Errorf("cannot derive a safe file name for operation %q", operationID)
	}
	return name, nil
}

// isLocalSingleSegment reports whether raw is already a safe, single-segment,
// root-local file name. It rewrites nothing: any traversal, separator (either
// flavour — RFC 6266 counts both, and filepath.Base on Linux leaves
// backslash-separated Windows traversal fully intact), absolute path, or
// reserved name is simply refused.
func isLocalSingleSegment(raw string) bool {
	name := strings.TrimSpace(raw)
	if name == "" || name == "." || name == ".." {
		return false
	}
	if strings.ContainsAny(name, `/\`) {
		return false
	}
	localized, err := filepath.Localize(name)
	if err != nil || localized != name {
		return false
	}
	return filepath.IsLocal(name)
}

// sanitizeBinaryDownloadFileName is the LENIENT counterpart, used only for
// untrusted provider text (Content-Disposition) and for the operation-derived
// fallback, where rewriting to something safe is preferable to failing the
// download outright.
func sanitizeBinaryDownloadFileName(raw string) (string, bool) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", false
	}
	// Strip BOTH separators before Base: on Linux, filepath.Base leaves
	// backslash-separated Windows-style traversal fully intact.
	name = strings.ReplaceAll(name, "\\", "/")
	name = filepath.Base(name)
	if name == "." || name == ".." || name == "/" || name == "" {
		return "", false
	}
	localized, err := filepath.Localize(name)
	if err != nil {
		return "", false
	}
	if !filepath.IsLocal(localized) {
		return "", false
	}
	if strings.ContainsRune(localized, os.PathSeparator) {
		return "", false
	}
	return localized, true
}
