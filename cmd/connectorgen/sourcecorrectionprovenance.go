package main

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strconv"
	"strings"
)

// sourceDescriptorCorrectionProvenance is a deliberately small authoring-only
// declaration for a corrected descriptor that is derived from an immutable
// source-lock snapshot. It never changes provider source membership, lane
// classification, or runtime admission. The matrix keeps both the original
// upstream identity and the derived descriptor identity so it cannot claim
// that a locally corrected descriptor came byte-for-byte from the archive.
type sourceDescriptorCorrectionProvenance struct {
	SchemaVersion int                                       `json:"schema_version"`
	Connector     string                                    `json:"connector"`
	Kind          string                                    `json:"kind"`
	BaseSnapshot  sourceDescriptorCorrectionBaseSnapshot    `json:"base_snapshot"`
	Target        sourceDescriptorCorrectionTarget          `json:"target"`
	SourceLock    sourceDescriptorCorrectionSourceLock      `json:"source_lock"`
	Corrections   []sourceDescriptorCorrectionSourceBinding `json:"corrections"`
}

type sourceDescriptorCorrectionBaseSnapshot struct {
	Ref             string `json:"ref"`
	Commit          string `json:"commit"`
	Materialization string `json:"materialization"`
}

type sourceDescriptorCorrectionIdentity struct {
	Path        string `json:"path"`
	GitBlobSHA1 string `json:"git_blob_sha1"`
	Bytes       int64  `json:"bytes"`
}

type sourceDescriptorCorrectionTarget struct {
	Path     string                             `json:"path"`
	Original sourceDescriptorCorrectionIdentity `json:"original"`
	Derived  sourceDescriptorCorrectionIdentity `json:"derived"`
}

type sourceDescriptorCorrectionSourceLock struct {
	Path      string `json:"path"`
	SourceURL string `json:"source_url"`
	SHA256    string `json:"sha256"`
	Bytes     int64  `json:"bytes"`
}

type sourceDescriptorCorrectionSourceBinding struct {
	SourceID          string                                 `json:"source_id"`
	OperationID       string                                 `json:"operation_id"`
	Method            string                                 `json:"method"`
	Path              string                                 `json:"path"`
	SourceLocation    string                                 `json:"source_location"`
	DescriptorPointer string                                 `json:"descriptor_pointer"`
	SourceSchema      json.RawMessage                        `json:"source_schema"`
	RemovedGaps       []sourceDescriptorCorrectionGapBinding `json:"removed_gaps"`
	RetainedGaps      []sourceDescriptorCorrectionGapBinding `json:"retained_gaps"`
	Rationale         string                                 `json:"rationale"`
}

type sourceDescriptorCorrectionGapBinding struct {
	Scope        string `json:"scope"`
	Foundation   string `json:"foundation"`
	Location     string `json:"location"`
	ReasonSHA256 string `json:"reason_sha256"`
}

type sourceDescriptorCorrectionMatrix struct {
	Connector      string `json:"connector"`
	SourceSnapshot struct {
		SourceSnapshotRef              string                               `json:"source_snapshot_ref"`
		SourceSnapshotCommit           string                               `json:"source_snapshot_commit"`
		Materialization                string                               `json:"materialization"`
		BaseMaterialization            string                               `json:"base_materialization"`
		DescriptorCorrectionProvenance *sourceDescriptorCorrectionIdentity  `json:"descriptor_correction_provenance,omitempty"`
		BaseRetainedFiles              []sourceDescriptorCorrectionIdentity `json:"base_retained_files,omitempty"`
		RetainedFiles                  []sourceDescriptorCorrectionIdentity `json:"retained_files"`
	} `json:"source_snapshot"`
}

type sourceDescriptorCorrectionLock struct {
	Connector string `json:"connector"`
	Rest      struct {
		SourceURL  string                                    `json:"source_url"`
		SHA256     string                                    `json:"sha256"`
		Bytes      int64                                     `json:"bytes"`
		Operations []sourceDescriptorCorrectionLockOperation `json:"operations"`
	} `json:"rest"`
}

type sourceDescriptorCorrectionLockOperation struct {
	ID             string `json:"id"`
	Method         string `json:"method"`
	Path           string `json:"path"`
	OperationID    string `json:"operation_id"`
	SourceLocation string `json:"source_location"`
}

const sourceDescriptorCorrectionKind = "source_lock_backed_descriptor_correction"

// validateSourceDescriptorCorrectionProvenance validates a matrix-declared
// descriptor correction for any connector. A connector without a correction
// declaration stays on its existing source-snapshot path. The correction is
// data-driven: no connector, operation ID, schema name, or provider URL is
// selected in Go code.
func validateSourceDescriptorCorrectionProvenance(fsys fs.FS, connector string, lockRaw, descriptorRaw []byte) error {
	matrixPath := path.Join(connector, "sources", connector+"-source-lane-matrix.json")
	matrixRaw, err := fs.ReadFile(fsys, matrixPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read source lane matrix: %w", err)
	}
	var matrix sourceDescriptorCorrectionMatrix
	if err := json.Unmarshal(matrixRaw, &matrix); err != nil {
		return fmt.Errorf("parse source lane matrix: %w", err)
	}
	if matrix.SourceSnapshot.DescriptorCorrectionProvenance == nil {
		if matrix.SourceSnapshot.Materialization == sourceDescriptorCorrectionKind {
			return fmt.Errorf("derived descriptor materialization declares no correction provenance")
		}
		return nil
	}
	if matrix.Connector != connector {
		return fmt.Errorf("matrix connector %q does not match %q", matrix.Connector, connector)
	}
	if matrix.SourceSnapshot.Materialization != sourceDescriptorCorrectionKind || matrix.SourceSnapshot.BaseMaterialization != "git_archive_byte_identical" {
		return fmt.Errorf("correction provenance must distinguish derived descriptor materialization from git_archive_byte_identical base")
	}
	provenanceIdentity := *matrix.SourceSnapshot.DescriptorCorrectionProvenance
	provenanceRaw, err := sourceDescriptorCorrectionReadIdentity(fsys, connector, provenanceIdentity)
	if err != nil {
		return fmt.Errorf("correction provenance file: %w", err)
	}
	var provenance sourceDescriptorCorrectionProvenance
	if err := decodeSourceStrictJSON(provenanceRaw, &provenance); err != nil {
		return fmt.Errorf("parse correction provenance: %w", err)
	}
	if err := sourceDescriptorCorrectionValidateMatrixIdentity(matrix, provenance); err != nil {
		return err
	}

	baseFiles, err := sourceDescriptorCorrectionIdentityMap(matrix.SourceSnapshot.BaseRetainedFiles)
	if err != nil {
		return fmt.Errorf("base retained files: %w", err)
	}
	retainedFiles, err := sourceDescriptorCorrectionIdentityMap(matrix.SourceSnapshot.RetainedFiles)
	if err != nil {
		return fmt.Errorf("retained files: %w", err)
	}
	if err := sourceDescriptorCorrectionValidateRetainedFiles(fsys, connector, provenance, baseFiles, retainedFiles); err != nil {
		return err
	}

	var lock sourceDescriptorCorrectionLock
	if err := json.Unmarshal(lockRaw, &lock); err != nil {
		return fmt.Errorf("parse source lock evidence: %w", err)
	}
	if err := sourceDescriptorCorrectionValidateSourceLock(lock, connector, provenance, retainedFiles); err != nil {
		return err
	}

	descriptor, err := sourceDescriptorCorrectionDecodeAny(descriptorRaw)
	if err != nil {
		return fmt.Errorf("parse descriptor evidence: %w", err)
	}
	return sourceDescriptorCorrectionValidateBindings(descriptor, lock, provenance)
}

func sourceDescriptorCorrectionValidateMatrixIdentity(matrix sourceDescriptorCorrectionMatrix, provenance sourceDescriptorCorrectionProvenance) error {
	if provenance.SchemaVersion != 1 || provenance.Connector != matrix.Connector || provenance.Kind != sourceDescriptorCorrectionKind {
		return fmt.Errorf("correction provenance has unsupported identity")
	}
	if provenance.BaseSnapshot.Ref != matrix.SourceSnapshot.SourceSnapshotRef || provenance.BaseSnapshot.Commit != matrix.SourceSnapshot.SourceSnapshotCommit || provenance.BaseSnapshot.Materialization != matrix.SourceSnapshot.BaseMaterialization {
		return fmt.Errorf("correction provenance base snapshot does not match matrix")
	}
	if !sourceDescriptorCorrectionValidPath(provenance.Target.Path) || provenance.Target.Path != provenance.Target.Original.Path || provenance.Target.Path != provenance.Target.Derived.Path {
		return fmt.Errorf("correction provenance target path is invalid")
	}
	if !sourceDescriptorCorrectionValidIdentity(provenance.Target.Original) || !sourceDescriptorCorrectionValidIdentity(provenance.Target.Derived) || sourceDescriptorCorrectionSameIdentity(provenance.Target.Original, provenance.Target.Derived) {
		return fmt.Errorf("correction provenance target identities must be distinct and complete")
	}
	if !sourceDescriptorCorrectionValidPath(provenance.SourceLock.Path) || provenance.SourceLock.SourceURL == "" || !sourceDescriptorCorrectionValidHex(provenance.SourceLock.SHA256, 64) || provenance.SourceLock.Bytes <= 0 {
		return fmt.Errorf("correction provenance source lock evidence is incomplete")
	}
	if len(provenance.Corrections) == 0 {
		return fmt.Errorf("correction provenance declares no corrections")
	}
	return nil
}

func sourceDescriptorCorrectionValidateRetainedFiles(fsys fs.FS, connector string, provenance sourceDescriptorCorrectionProvenance, baseFiles, retainedFiles map[string]sourceDescriptorCorrectionIdentity) error {
	if len(baseFiles) == 0 || len(baseFiles) != len(retainedFiles) {
		return fmt.Errorf("base and derived retained-file inventories differ")
	}
	for filePath, base := range baseFiles {
		retained, found := retainedFiles[filePath]
		if !found {
			return fmt.Errorf("derived retained-file inventory omits %q", filePath)
		}
		if filePath == provenance.Target.Path {
			if !sourceDescriptorCorrectionSameIdentity(base, provenance.Target.Original) || !sourceDescriptorCorrectionSameIdentity(retained, provenance.Target.Derived) {
				return fmt.Errorf("descriptor correction identities do not match the matrix retained-file inventories")
			}
		} else if !sourceDescriptorCorrectionSameIdentity(base, retained) {
			return fmt.Errorf("uncorrected retained source file identity changed %q", filePath)
		}
		if _, err := sourceDescriptorCorrectionReadIdentity(fsys, connector, retained); err != nil {
			return fmt.Errorf("derived retained file %q: %w", filePath, err)
		}
	}
	if _, found := baseFiles[provenance.Target.Path]; !found {
		return fmt.Errorf("correction target %q is absent from base retained files", provenance.Target.Path)
	}
	return nil
}

func sourceDescriptorCorrectionValidateSourceLock(lock sourceDescriptorCorrectionLock, connector string, provenance sourceDescriptorCorrectionProvenance, retainedFiles map[string]sourceDescriptorCorrectionIdentity) error {
	if lock.Connector != connector || provenance.SourceLock.SourceURL != lock.Rest.SourceURL || provenance.SourceLock.SHA256 != lock.Rest.SHA256 || provenance.SourceLock.Bytes != lock.Rest.Bytes {
		return fmt.Errorf("correction provenance source-lock evidence does not match the retained lock")
	}
	if _, found := retainedFiles[provenance.SourceLock.Path]; !found {
		return fmt.Errorf("correction provenance source lock %q is not a retained file", provenance.SourceLock.Path)
	}
	return nil
}

func sourceDescriptorCorrectionValidateBindings(descriptor any, lock sourceDescriptorCorrectionLock, provenance sourceDescriptorCorrectionProvenance) error {
	descriptorObject, ok := descriptor.(map[string]any)
	if !ok {
		return fmt.Errorf("descriptor is not an object")
	}
	descriptorOperations, ok := descriptorObject["operations"].([]any)
	if !ok {
		return fmt.Errorf("descriptor has no operations")
	}
	seen := map[string]bool{}
	for _, correction := range provenance.Corrections {
		if err := sourceDescriptorCorrectionValidateBindingFields(correction); err != nil {
			return err
		}
		if seen[correction.SourceID] {
			return fmt.Errorf("correction provenance duplicates source ID %q", correction.SourceID)
		}
		seen[correction.SourceID] = true
		lockOperation, err := sourceDescriptorCorrectionFindLockOperation(lock.Rest.Operations, correction.OperationID)
		if err != nil {
			return err
		}
		if !strings.EqualFold(lockOperation.Method, correction.Method) || lockOperation.Path != correction.Path || lockOperation.SourceLocation != correction.SourceLocation {
			return fmt.Errorf("correction source-lock operation evidence does not match declared operation %q", correction.OperationID)
		}
		descriptorOperation, err := sourceDescriptorCorrectionFindObject(descriptorOperations, "source_id", correction.SourceID)
		if err != nil {
			return fmt.Errorf("correction descriptor source ID: %w", err)
		}
		if sourceDescriptorCorrectionString(descriptorOperation, "operation_id") != correction.OperationID || !strings.EqualFold(sourceDescriptorCorrectionString(descriptorOperation, "method"), correction.Method) || sourceDescriptorCorrectionString(descriptorOperation, "path") != correction.Path {
			return fmt.Errorf("correction descriptor operation does not match declared source-lock operation %q", correction.OperationID)
		}
		source, ok := descriptorOperation["source"].(map[string]any)
		if !ok || sourceDescriptorCorrectionString(source, "url") != provenance.SourceLock.SourceURL || sourceDescriptorCorrectionString(source, "sha256") != provenance.SourceLock.SHA256 || sourceDescriptorCorrectionNumber(source, "bytes") != provenance.SourceLock.Bytes || sourceDescriptorCorrectionString(source, "location") != correction.SourceLocation {
			return fmt.Errorf("correction descriptor provenance does not match the source lock for %q", correction.OperationID)
		}
		actualSchema, err := sourceDescriptorCorrectionPointer(descriptorOperation, correction.DescriptorPointer)
		if err != nil {
			return fmt.Errorf("correction descriptor pointer for %q: %w", correction.OperationID, err)
		}
		expectedSchema, err := sourceDescriptorCorrectionDecodeAny(correction.SourceSchema)
		if err != nil || !sourceDescriptorCorrectionEquivalent(actualSchema, expectedSchema) {
			return fmt.Errorf("correction source schema evidence does not match descriptor pointer for %q", correction.OperationID)
		}
		if err := sourceDescriptorCorrectionValidateGapBindings(descriptorObject, descriptorOperation, correction); err != nil {
			return fmt.Errorf("correction gap evidence for %q: %w", correction.OperationID, err)
		}
	}
	return nil
}

func sourceDescriptorCorrectionValidateBindingFields(correction sourceDescriptorCorrectionSourceBinding) error {
	if correction.SourceID == "" || correction.OperationID == "" || correction.Method == "" || correction.Path == "" || correction.SourceLocation == "" || correction.DescriptorPointer == "" || strings.Contains(correction.Rationale, "\n") || strings.TrimSpace(correction.Rationale) == "" || len(correction.SourceSchema) == 0 || len(correction.RemovedGaps) == 0 || len(correction.RetainedGaps) == 0 {
		return fmt.Errorf("correction provenance has incomplete binding fields")
	}
	return nil
}

func sourceDescriptorCorrectionValidateGapBindings(descriptor, operation map[string]any, correction sourceDescriptorCorrectionSourceBinding) error {
	removed := make(map[string]struct{}, len(correction.RemovedGaps))
	for _, gap := range correction.RemovedGaps {
		if err := sourceDescriptorCorrectionValidateGapBinding(gap); err != nil {
			return err
		}
		key := sourceDescriptorCorrectionGapBindingKey(gap)
		if _, duplicate := removed[key]; duplicate {
			return fmt.Errorf("removed gap is declared more than once at %s", gap.Location)
		}
		removed[key] = struct{}{}
		gaps, err := sourceDescriptorCorrectionGapsForScope(descriptor, operation, gap.Scope)
		if err != nil {
			return err
		}
		if sourceDescriptorCorrectionHasGap(gaps, gap) {
			return fmt.Errorf("removed gap remains at %s", gap.Location)
		}
	}
	for _, gap := range correction.RetainedGaps {
		if err := sourceDescriptorCorrectionValidateGapBinding(gap); err != nil {
			return err
		}
		key := sourceDescriptorCorrectionGapBindingKey(gap)
		if _, overlap := removed[key]; overlap {
			return fmt.Errorf("gap is declared both removed and retained at %s", gap.Location)
		}
		gaps, err := sourceDescriptorCorrectionGapsForScope(descriptor, operation, gap.Scope)
		if err != nil {
			return err
		}
		if !sourceDescriptorCorrectionHasGap(gaps, gap) {
			return fmt.Errorf("retained gap is absent at %s", gap.Location)
		}
	}
	return nil
}

func sourceDescriptorCorrectionGapBindingKey(gap sourceDescriptorCorrectionGapBinding) string {
	return strings.Join([]string{gap.Scope, gap.Foundation, gap.Location, gap.ReasonSHA256}, "\x00")
}

func sourceDescriptorCorrectionValidateGapBinding(gap sourceDescriptorCorrectionGapBinding) error {
	if (gap.Scope != "operation_runtime" && gap.Scope != "descriptor_aggregate") || gap.Foundation == "" || gap.Location == "" || !sourceDescriptorCorrectionValidHex(gap.ReasonSHA256, 64) {
		return fmt.Errorf("gap binding is incomplete")
	}
	return nil
}

func sourceDescriptorCorrectionGapsForScope(descriptor, operation map[string]any, scope string) ([]any, error) {
	var source any
	switch scope {
	case "operation_runtime":
		runtime, ok := operation["runtime"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("operation runtime gaps are absent")
		}
		source = runtime["gaps"]
	case "descriptor_aggregate":
		source = descriptor["gaps"]
	default:
		return nil, fmt.Errorf("unknown gap scope %q", scope)
	}
	if source == nil {
		return nil, nil
	}
	gaps, ok := source.([]any)
	if !ok {
		return nil, fmt.Errorf("gap scope %q is not an array", scope)
	}
	return gaps, nil
}

func sourceDescriptorCorrectionHasGap(gaps []any, expected sourceDescriptorCorrectionGapBinding) bool {
	for _, raw := range gaps {
		gap, ok := raw.(map[string]any)
		if !ok || sourceDescriptorCorrectionString(gap, "foundation") != expected.Foundation || sourceDescriptorCorrectionString(gap, "location") != expected.Location {
			continue
		}
		digest := sha256.Sum256([]byte(sourceDescriptorCorrectionString(gap, "reason")))
		if hex.EncodeToString(digest[:]) == expected.ReasonSHA256 {
			return true
		}
	}
	return false
}

func sourceDescriptorCorrectionFindLockOperation(operations []sourceDescriptorCorrectionLockOperation, operationID string) (sourceDescriptorCorrectionLockOperation, error) {
	var found sourceDescriptorCorrectionLockOperation
	for _, operation := range operations {
		if operation.OperationID != operationID {
			continue
		}
		if found.OperationID != "" {
			return found, fmt.Errorf("source lock duplicates operation %q", operationID)
		}
		found = operation
	}
	if found.OperationID == "" {
		return found, fmt.Errorf("source lock omits operation %q", operationID)
	}
	return found, nil
}

func sourceDescriptorCorrectionFindObject(values []any, key, want string) (map[string]any, error) {
	var found map[string]any
	for _, value := range values {
		object, ok := value.(map[string]any)
		if !ok || sourceDescriptorCorrectionString(object, key) != want {
			continue
		}
		if found != nil {
			return nil, fmt.Errorf("duplicates %s %q", key, want)
		}
		found = object
	}
	if found == nil {
		return nil, fmt.Errorf("omits %s %q", key, want)
	}
	return found, nil
}

func sourceDescriptorCorrectionPointer(value any, pointer string) (any, error) {
	if pointer == "" || !strings.HasPrefix(pointer, "/") {
		return nil, fmt.Errorf("must be a JSON Pointer")
	}
	current := value
	for _, token := range strings.Split(pointer[1:], "/") {
		token = strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
		switch typed := current.(type) {
		case map[string]any:
			var found bool
			current, found = typed[token]
			if !found {
				return nil, fmt.Errorf("token %q is absent", token)
			}
		case []any:
			index, err := strconv.Atoi(token)
			if err != nil || index < 0 || index >= len(typed) {
				return nil, fmt.Errorf("array token %q is invalid", token)
			}
			current = typed[index]
		default:
			return nil, fmt.Errorf("token %q crosses a scalar", token)
		}
	}
	return current, nil
}

func sourceDescriptorCorrectionIdentityMap(values []sourceDescriptorCorrectionIdentity) (map[string]sourceDescriptorCorrectionIdentity, error) {
	result := make(map[string]sourceDescriptorCorrectionIdentity, len(values))
	for _, value := range values {
		if !sourceDescriptorCorrectionValidIdentity(value) {
			return nil, fmt.Errorf("identity is incomplete")
		}
		if _, duplicate := result[value.Path]; duplicate {
			return nil, fmt.Errorf("identity duplicates path %q", value.Path)
		}
		result[value.Path] = value
	}
	return result, nil
}

func sourceDescriptorCorrectionReadIdentity(fsys fs.FS, connector string, identity sourceDescriptorCorrectionIdentity) ([]byte, error) {
	if !sourceDescriptorCorrectionValidIdentity(identity) {
		return nil, fmt.Errorf("identity is incomplete")
	}
	raw, err := fs.ReadFile(fsys, path.Join(connector, identity.Path))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) != identity.Bytes || sourceDescriptorCorrectionGitBlobSHA1(raw) != identity.GitBlobSHA1 {
		return nil, fmt.Errorf("byte/blob identity drift")
	}
	return raw, nil
}

func sourceDescriptorCorrectionValidIdentity(identity sourceDescriptorCorrectionIdentity) bool {
	return sourceDescriptorCorrectionValidPath(identity.Path) && sourceDescriptorCorrectionValidHex(identity.GitBlobSHA1, 40) && identity.Bytes > 0
}

func sourceDescriptorCorrectionValidPath(value string) bool {
	return strings.HasPrefix(value, "sources/") && fs.ValidPath(value) && path.Clean(value) == value
}

func sourceDescriptorCorrectionValidHex(value string, length int) bool {
	if len(value) != length || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func sourceDescriptorCorrectionSameIdentity(left, right sourceDescriptorCorrectionIdentity) bool {
	return left.Path == right.Path && left.GitBlobSHA1 == right.GitBlobSHA1 && left.Bytes == right.Bytes
}

func sourceDescriptorCorrectionGitBlobSHA1(raw []byte) string {
	hash := sha1.New()
	_, _ = hash.Write([]byte(fmt.Sprintf("blob %d\x00", len(raw))))
	_, _ = hash.Write(raw)
	return hex.EncodeToString(hash.Sum(nil))
}

func sourceDescriptorCorrectionDecodeAny(raw []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, fmt.Errorf("multiple JSON values")
		}
		return nil, err
	}
	return value, nil
}

func sourceDescriptorCorrectionEquivalent(left, right any) bool {
	leftRaw, leftErr := marshalNoEscapeHTML(left)
	rightRaw, rightErr := marshalNoEscapeHTML(right)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftRaw, rightRaw)
}

func sourceDescriptorCorrectionString(object map[string]any, key string) string {
	value, _ := object[key].(string)
	return value
}

func sourceDescriptorCorrectionNumber(object map[string]any, key string) int64 {
	value, ok := object[key].(json.Number)
	if !ok {
		return 0
	}
	result, err := value.Int64()
	if err != nil {
		return 0
	}
	return result
}
