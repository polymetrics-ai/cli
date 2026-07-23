package connsdk

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/safety"
	"polymetrics.ai/internal/telemetry"
)

// maxErrorBody bounds how much of an error response body is captured in HTTPError.
const maxErrorBody = 8 << 10 // 8 KiB
const defaultMaxResponseBody = 64 << 20

// Response is a captured HTTP response with its body already read.
type Response struct {
	Status     int
	Header     http.Header
	Body       []byte
	requestURL string
}

// HTTPError is returned when a request completes with a 4xx/5xx status after
// exhausting retries. The body is truncated and never assumed to be secret-free
// by callers, but connsdk itself never logs it.
type HTTPError struct {
	Status int
	URL    string
	Body   string
}

func (e *HTTPError) Error() string {
	msg := http.StatusText(e.Status)
	if msg == "" {
		msg = "http error"
	}
	return safety.RedactErrorText(fmt.Sprintf("http %d for %s: %s", e.Status, safeErrorURL(e.URL), msg))
}

// StatusCode returns the HTTP response status for safe telemetry classification.
func (e *HTTPError) StatusCode() int {
	if e == nil {
		return 0
	}
	return e.Status
}

// TelemetryErrorClass returns a stable error class for secret-safe tracing.
func (e *HTTPError) TelemetryErrorClass() string { return "http" }

// TelemetryErrorCode returns a stable error code for secret-safe tracing.
func (e *HTTPError) TelemetryErrorCode() string { return "http_status" }

// MultipartForm describes bounded fields and files for a multipart request.
type MultipartForm struct {
	Fields   map[string]string
	Files    []MultipartFile
	MaxBytes int64
}

// MultipartFile describes one file part and its approval-time identity.
type MultipartFile struct {
	FieldName      string
	Path           string
	FileName       string
	ContentType    string
	MaxBytes       int64
	ExpectedSHA256 string
}

type requestBody struct {
	Reader      io.Reader
	ContentType string
	Cleanup     func() error
}

type countingReader struct {
	reader io.Reader
	bytes  int64
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.bytes += int64(n)
	return n, err
}

func metricByteCount(n int64) int {
	maxInt := int64(^uint(0) >> 1)
	if n > maxInt {
		return int(maxInt)
	}
	return int(n)
}

// Requester performs JSON HTTP requests with auth, retry, and rate-limit handling.
// The zero value is usable once Client/BaseURL are set; sensible defaults are
// applied for the rest on first use.
type Requester struct {
	// Client is the HTTP client. Defaults to a client with a 60s timeout.
	Client *http.Client
	// BaseURL is prepended to relative paths. A path beginning with http:// or
	// https:// is treated as absolute and used as-is (e.g. Link-header next URLs).
	BaseURL string
	// Auth, when set, is applied to every request before it is sent.
	Auth Authenticator
	// UserAgent and DefaultHeaders are applied to every request.
	UserAgent      string
	DefaultHeaders map[string]string
	// Accept overrides the Accept header (defaults to application/json).
	Accept string

	// MaxRetries is the number of additional attempts after the first (default 4).
	MaxRetries int
	// BaseBackoff and MaxBackoff bound exponential backoff (defaults 500ms / 30s).
	BaseBackoff time.Duration
	MaxBackoff  time.Duration
	// RetryStatuses are HTTP statuses that trigger a retry. Defaults to
	// 429, 500, 502, 503, 504.
	RetryStatuses map[int]bool
	// Sleep waits for d or until ctx is cancelled. Injectable for tests.
	Sleep func(ctx context.Context, d time.Duration) error
}

func (r *Requester) client() *http.Client {
	if r.Client != nil {
		return r.Client
	}
	return &http.Client{Timeout: 60 * time.Second}
}

func (r *Requester) maxRetries() int {
	if r.MaxRetries > 0 {
		return r.MaxRetries
	}
	return 4
}

func (r *Requester) baseBackoff() time.Duration {
	if r.BaseBackoff > 0 {
		return r.BaseBackoff
	}
	return 500 * time.Millisecond
}

func (r *Requester) maxBackoff() time.Duration {
	if r.MaxBackoff > 0 {
		return r.MaxBackoff
	}
	return 30 * time.Second
}

func (r *Requester) shouldRetry(status int) bool {
	if r.RetryStatuses != nil {
		return r.RetryStatuses[status]
	}
	switch status {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	default:
		return false
	}
}

func (r *Requester) sleep(ctx context.Context, d time.Duration) error {
	if r.Sleep != nil {
		return r.Sleep(ctx, d)
	}
	return ctxSleep(ctx, d)
}

// ctxSleep waits for d or returns early if ctx is cancelled.
func ctxSleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// resolveURL builds the absolute request URL from path + query.
func (r *Requester) resolveURL(path string, query url.Values) (string, error) {
	raw := path
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		base := strings.TrimRight(r.BaseURL, "/")
		raw = base + "/" + strings.TrimLeft(path, "/")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse url %q: %w", raw, err)
	}
	if len(query) > 0 {
		existing := u.Query()
		for k, vs := range query {
			existing.Del(k)
			for _, v := range vs {
				existing.Add(k, v)
			}
		}
		u.RawQuery = existing.Encode()
	}
	return u.String(), nil
}

// Do performs an HTTP request with an optional JSON body, retrying on transient
// failures, and returns the captured response. A 4xx/5xx after retries is
// returned as *HTTPError.
func (r *Requester) Do(ctx context.Context, method, path string, query url.Values, body any) (*Response, error) {
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
	}
	return r.do(ctx, method, path, query, payload, "application/json", defaultMaxResponseBody)
}

// DoLimited performs Do while bounding the captured successful response body to
// maxBodyBytes+1. Callers can reject len(resp.Body) > maxBodyBytes without ever
// buffering the default 64 MiB response cap.
func (r *Requester) DoLimited(ctx context.Context, method, path string, query url.Values, body any, maxBodyBytes int) (*Response, error) {
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
	}
	return r.do(ctx, method, path, query, payload, "application/json", maxBodyBytes+1)
}

// DoForm performs an HTTP request with an application/x-www-form-urlencoded body,
// reusing the same auth, retry, and rate-limit handling as Do. It is the form
// counterpart used by APIs (e.g. Stripe) whose write endpoints take form bodies.
// A nil/empty form sends no body.
func (r *Requester) DoForm(ctx context.Context, method, path string, query, form url.Values) (*Response, error) {
	var payload []byte
	contentType := ""
	if len(form) > 0 {
		payload = []byte(form.Encode())
		contentType = "application/x-www-form-urlencoded"
	}
	return r.do(ctx, method, path, query, payload, contentType, defaultMaxResponseBody)
}

// DoMultipart performs an HTTP request with a multipart/form-data body. File
// parts are opened for each retry attempt, so callers may use it with the same
// retry policy as JSON/form requests without reusing a consumed reader.
func (r *Requester) DoMultipart(ctx context.Context, method, path string, query url.Values, form MultipartForm) (*Response, error) {
	if err := validateMultipartForm(form); err != nil {
		return nil, err
	}
	prepared, cleanup, err := snapshotApprovedMultipartFiles(ctx, form)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	return r.doWithBody(ctx, method, path, query, defaultMaxResponseBody, func() (*requestBody, error) {
		return multipartBody(prepared)
	})
}

func validateMultipartForm(form MultipartForm) error {
	var total int64
	for i, file := range form.Files {
		if strings.TrimSpace(file.FieldName) == "" {
			return fmt.Errorf("multipart file %d field name is required", i)
		}
		if strings.TrimSpace(file.Path) == "" {
			return fmt.Errorf("multipart file %q path is required", file.FieldName)
		}
		if file.ExpectedSHA256 != "" {
			digest, err := hex.DecodeString(file.ExpectedSHA256)
			if err != nil || len(digest) != sha256.Size {
				return fmt.Errorf("multipart file %q expected SHA-256 is invalid", file.FieldName)
			}
		}
		info, err := os.Stat(file.Path)
		if err != nil {
			return fmt.Errorf("multipart file %q: %w", file.FieldName, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("multipart file %q must be a regular file", file.FieldName)
		}
		if file.MaxBytes > 0 && info.Size() > file.MaxBytes {
			return fmt.Errorf("multipart file %q too large: %d bytes exceeds limit %d", file.FieldName, info.Size(), file.MaxBytes)
		}
		total += info.Size()
		if form.MaxBytes > 0 && total > form.MaxBytes {
			return fmt.Errorf("multipart files too large: %d bytes exceeds limit %d", total, form.MaxBytes)
		}
	}
	return nil
}

func snapshotApprovedMultipartFiles(ctx context.Context, form MultipartForm) (MultipartForm, func(), error) {
	prepared := form
	prepared.Files = append([]MultipartFile(nil), form.Files...)
	tempPaths := make([]string, 0, len(form.Files))
	cleanup := func() {
		for _, path := range tempPaths {
			_ = os.Remove(path)
		}
	}
	var total int64
	for i, file := range form.Files {
		if file.ExpectedSHA256 == "" {
			info, err := os.Stat(file.Path)
			if err != nil {
				cleanup()
				return MultipartForm{}, func() {}, fmt.Errorf("multipart file %q: %w", file.FieldName, err)
			}
			total += info.Size()
			continue
		}
		limit := file.MaxBytes
		if limit <= 0 {
			limit = -1
		}
		if form.MaxBytes > 0 {
			remaining := form.MaxBytes - total
			if remaining < 0 {
				cleanup()
				return MultipartForm{}, func() {}, fmt.Errorf("multipart files too large: exceeds limit %d", form.MaxBytes)
			}
			if limit < 0 || remaining < limit {
				limit = remaining
			}
		}
		tempPath, size, digest, err := snapshotMultipartFile(ctx, file, limit)
		if err != nil {
			cleanup()
			return MultipartForm{}, func() {}, err
		}
		tempPaths = append(tempPaths, tempPath)
		expected, _ := hex.DecodeString(file.ExpectedSHA256)
		if !bytes.Equal(digest, expected) {
			cleanup()
			return MultipartForm{}, func() {}, fmt.Errorf("multipart file %q changed since approval", file.FieldName)
		}
		if prepared.Files[i].FileName == "" {
			prepared.Files[i].FileName = filepath.Base(file.Path)
		}
		prepared.Files[i].Path = tempPath
		total += size
	}
	return prepared, cleanup, nil
}

func snapshotMultipartFile(ctx context.Context, file MultipartFile, maxBytes int64) (string, int64, []byte, error) {
	source, err := os.Open(file.Path)
	if err != nil {
		return "", 0, nil, fmt.Errorf("multipart file %q: %w", file.FieldName, err)
	}
	defer source.Close()
	temp, err := os.CreateTemp("", "polymetrics-upload-*")
	if err != nil {
		return "", 0, nil, fmt.Errorf("snapshot multipart file %q: %w", file.FieldName, err)
	}
	tempPath := temp.Name()
	removeTemp := true
	defer func() {
		_ = temp.Close()
		if removeTemp {
			_ = os.Remove(tempPath)
		}
	}()

	hash := sha256.New()
	reader := io.Reader(&contextReader{ctx: ctx, reader: source})
	if maxBytes >= 0 {
		reader = io.LimitReader(reader, maxBytes)
	}
	written, err := io.Copy(io.MultiWriter(temp, hash), reader)
	if err != nil {
		return "", written, nil, fmt.Errorf("snapshot multipart file %q: %w", file.FieldName, err)
	}
	if maxBytes >= 0 && written == maxBytes {
		var extra [1]byte
		n, readErr := (&contextReader{ctx: ctx, reader: source}).Read(extra[:])
		if n > 0 {
			return "", written, nil, fmt.Errorf("multipart file %q too large: exceeds limit %d", file.FieldName, maxBytes)
		}
		if readErr != nil && readErr != io.EOF {
			return "", written, nil, fmt.Errorf("snapshot multipart file %q: %w", file.FieldName, readErr)
		}
	}
	if err := temp.Close(); err != nil {
		return "", written, nil, fmt.Errorf("snapshot multipart file %q: %w", file.FieldName, err)
	}
	removeTemp = false
	return tempPath, written, hash.Sum(nil), nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *contextReader) Read(p []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
		return r.reader.Read(p)
	}
}

func multipartBody(form MultipartForm) (*requestBody, error) {
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	done := make(chan error, 1)
	go func() {
		err := writeMultipartForm(mw, form)
		if closeErr := mw.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			_ = pw.CloseWithError(err)
		} else {
			_ = pw.Close()
		}
		done <- err
	}()
	return &requestBody{
		Reader:      pr,
		ContentType: mw.FormDataContentType(),
		Cleanup: func() error {
			_ = pr.Close()
			return <-done
		},
	}, nil
}

func writeMultipartForm(mw *multipart.Writer, form MultipartForm) error {
	keys := make([]string, 0, len(form.Fields))
	for key := range form.Fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := mw.WriteField(key, form.Fields[key]); err != nil {
			return err
		}
	}
	var total int64
	for _, file := range form.Files {
		limit := file.MaxBytes
		if limit <= 0 {
			limit = -1
		}
		if form.MaxBytes > 0 {
			remaining := form.MaxBytes - total
			if limit < 0 || remaining < limit {
				limit = remaining
			}
		}
		written, err := writeMultipartFile(mw, file, limit)
		if err != nil {
			return err
		}
		total += written
	}
	return nil
}

func writeMultipartFile(mw *multipart.Writer, file MultipartFile, maxBytes int64) (int64, error) {
	name := file.FileName
	if strings.TrimSpace(name) == "" {
		name = filepath.Base(file.Path)
	}
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, file.FieldName, name))
	if file.ContentType != "" {
		header.Set("Content-Type", file.ContentType)
	}
	part, err := mw.CreatePart(header)
	if err != nil {
		return 0, err
	}
	f, err := os.Open(file.Path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	if maxBytes < 0 {
		written, err := io.Copy(part, f)
		return written, err
	}
	written, err := io.CopyN(part, f, maxBytes)
	if err != nil && err != io.EOF {
		return written, err
	}
	if written < maxBytes {
		return written, nil
	}
	var extra [1]byte
	n, readErr := f.Read(extra[:])
	if n > 0 {
		return written, fmt.Errorf("multipart file %q too large: exceeds limit %d", file.FieldName, maxBytes)
	}
	if readErr != nil && readErr != io.EOF {
		return written, readErr
	}
	return written, nil
}

// do is the shared request core for Do/DoForm. payload is the already-encoded
// body (nil for none) and contentType is the Content-Type to set when a body is
// present.
func (r *Requester) do(ctx context.Context, method, path string, query url.Values, payload []byte, contentType string, maxBodyBytes int) (*Response, error) {
	return r.doWithBody(ctx, method, path, query, maxBodyBytes, func() (*requestBody, error) {
		if payload == nil {
			return nil, nil
		}
		return &requestBody{Reader: bytes.NewReader(payload), ContentType: contentType}, nil
	})
}

func (r *Requester) doWithBody(ctx context.Context, method, path string, query url.Values, maxBodyBytes int, bodyFactory func() (*requestBody, error)) (*Response, error) {
	fullURL, err := r.resolveURL(path, query)
	if err != nil {
		return nil, err
	}
	if maxBodyBytes <= 0 {
		maxBodyBytes = defaultMaxResponseBody
	}

	attempts := r.maxRetries() + 1
	attrs := append(telemetry.HTTPAttrs(method, fullURL), telemetry.IntAttr("pm.http.max_attempts", attempts))
	ctx, span := telemetry.StartSpan(ctx, "pm.connector.http", attrs...)
	defer span.End()
	started := time.Now()
	responseBytes := 0
	defer func() {
		telemetry.RecordConnectorOperation(ctx, method, time.Since(started), responseBytes)
	}()
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		attemptAttr := telemetry.IntAttr("pm.http.attempt", attempt+1)
		span.AddEvent("pm.connector.http.attempt", attemptAttr)
		if err := ctx.Err(); err != nil {
			span.RecordError(err)
			return nil, err
		}

		body, err := bodyFactory()
		if err != nil {
			return nil, err
		}
		var reader io.Reader
		var contentType string
		var sent *countingReader
		if body != nil {
			sent = &countingReader{reader: body.Reader}
			reader = sent
			contentType = body.ContentType
		}
		req, err := http.NewRequestWithContext(ctx, method, fullURL, reader)
		if err != nil {
			cleanupRequestBody(body)
			return nil, fmt.Errorf("build request: %w", err)
		}
		r.applyHeaders(req, body != nil, contentType)
		if r.Auth != nil {
			if err := r.Auth.Apply(ctx, req); err != nil {
				cleanupRequestBody(body)
				return nil, fmt.Errorf("apply auth: %w", err)
			}
		}

		resp, err := r.client().Do(req)
		bodyErr := cleanupRequestBody(body)
		requestBytes := 0
		if sent != nil {
			requestBytes = metricByteCount(sent.bytes)
		}
		telemetry.RecordAPICall(ctx, method, requestBytes)
		if err != nil {
			lastErr = fmt.Errorf("send request: %w", err)
			span.AddEvent("pm.connector.http.error", attemptAttr)
			if bodyErr != nil {
				lastErr = fmt.Errorf("send request body: %w", bodyErr)
			}
			if attempt < attempts-1 {
				telemetry.RecordAPIRetry(ctx, method)
				if werr := r.sleep(ctx, r.backoff(attempt, "")); werr != nil {
					return nil, werr
				}
				continue
			}
			span.RecordError(lastErr)
			return nil, lastErr
		}
		if bodyErr != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("send request body: %w", bodyErr)
		}

		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, int64(maxBodyBytes)))
		resp.Body.Close()
		responseBytes += len(respBody)

		if r.shouldRetry(resp.StatusCode) && attempt < attempts-1 {
			lastErr = &HTTPError{Status: resp.StatusCode, URL: safeErrorURL(fullURL), Body: truncate(respBody)}
			span.AddEvent("pm.connector.http.retry", attemptAttr, telemetry.IntAttr("pm.http.status_code", resp.StatusCode), telemetry.BoolAttr("pm.http.retry", true))
			telemetry.RecordAPIRetry(ctx, method)
			wait := r.backoff(attempt, resp.Header.Get("Retry-After"))
			if resp.StatusCode == http.StatusTooManyRequests {
				telemetry.RecordRateLimitWait(ctx, method, wait)
			}
			if werr := r.sleep(ctx, wait); werr != nil {
				return nil, werr
			}
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			err := &HTTPError{Status: resp.StatusCode, URL: safeErrorURL(fullURL), Body: truncate(respBody)}
			span.SetAttributes(attemptAttr, telemetry.IntAttr("pm.http.status_code", resp.StatusCode))
			span.RecordError(err)
			return nil, err
		}

		span.SetAttributes(attemptAttr, telemetry.IntAttr("pm.http.status_code", resp.StatusCode), telemetry.BoolAttr("pm.http.retry", false))
		return &Response{Status: resp.StatusCode, Header: resp.Header, Body: respBody, requestURL: fullURL}, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("request to %s failed after %d attempts", safeErrorURL(fullURL), attempts)
	}
	span.RecordError(lastErr)
	return nil, lastErr
}

func cleanupRequestBody(body *requestBody) error {
	if body == nil || body.Cleanup == nil {
		return nil
	}
	return body.Cleanup()
}

// DoJSON performs a request and decodes a successful response into out (which may
// be nil to discard the body). Numbers are decoded with json.Number to preserve
// integer fidelity, matching the rest of the codebase.
func (r *Requester) DoJSON(ctx context.Context, method, path string, query url.Values, body, out any) error {
	resp, err := r.Do(ctx, method, path, query, body)
	if err != nil {
		return err
	}
	if out == nil || len(resp.Body) == 0 {
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(resp.Body))
	dec.UseNumber()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}
	return nil
}

func (r *Requester) applyHeaders(req *http.Request, hasBody bool, contentType string) {
	accept := r.Accept
	if accept == "" {
		accept = "application/json"
	}
	req.Header.Set("Accept", accept)
	if r.UserAgent != "" {
		req.Header.Set("User-Agent", r.UserAgent)
	}
	if hasBody {
		if contentType == "" {
			contentType = "application/json"
		}
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range r.DefaultHeaders {
		req.Header.Set(k, v)
	}
}

// backoff computes the wait before the next attempt. A Retry-After header (delay
// seconds or HTTP date) takes precedence, otherwise exponential backoff capped at
// MaxBackoff is used.
func (r *Requester) backoff(attempt int, retryAfter string) time.Duration {
	if d, ok := parseRetryAfter(retryAfter); ok {
		if d > r.maxBackoff() {
			return r.maxBackoff()
		}
		return d
	}
	d := r.baseBackoff() << attempt
	if d <= 0 || d > r.maxBackoff() {
		return r.maxBackoff()
	}
	return d
}

// parseRetryAfter parses a Retry-After header value as either delay-seconds or an
// HTTP date relative to now.
func parseRetryAfter(value string) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	if secs, err := strconv.Atoi(value); err == nil {
		if secs < 0 {
			return 0, false
		}
		return time.Duration(secs) * time.Second, true
	}
	if t, err := http.ParseTime(value); err == nil {
		d := time.Until(t)
		if d < 0 {
			return 0, true
		}
		return d, true
	}
	return 0, false
}

func safeErrorURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return safety.RedactErrorText(raw)
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	parsed.Fragment = ""
	parsed.RawPath = ""
	parsed.Host = safety.SanitizeTerminalLine(parsed.Host)
	parsed.Path = safety.SanitizeTerminalLine(parsed.Path)
	parsed.Opaque = safety.SanitizeTerminalLine(parsed.Opaque)
	return parsed.String()
}

func truncate(body []byte) string {
	if len(body) > maxErrorBody {
		return string(body[:maxErrorBody])
	}
	return string(body)
}
