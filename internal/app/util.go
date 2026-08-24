package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/safety"
)

func cloneStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneStringSliceMap(in map[string][]string) map[string][]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string][]string, len(in))
	for key, values := range in {
		out[key] = append([]string(nil), values...)
	}
	return out
}

func cloneRecord(in connectors.Record) connectors.Record {
	out := make(connectors.Record, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneRecords(in []connectors.Record) []connectors.Record {
	out := make([]connectors.Record, 0, len(in))
	for _, record := range in {
		out = append(out, cloneRecord(record))
	}
	return out
}

func RedactReversePlanRecords(in []connectors.Record, fields []string) []connectors.Record {
	out := make([]connectors.Record, 0, len(in))
	for _, record := range in {
		out = append(out, redactReversePlanRecord(record, fields))
	}
	return out
}

func redactReversePlanRecord(in connectors.Record, fields []string) connectors.Record {
	out := deepCloneRecord(in)
	for _, field := range fields {
		redactReversePlanField(out, strings.TrimPrefix(strings.TrimSpace(field), "record."))
	}
	return out
}

func deepCloneRecord(in connectors.Record) connectors.Record {
	out := make(connectors.Record, len(in))
	for k, v := range in {
		out[k] = deepCloneRecordValue(v)
	}
	return out
}

func deepCloneRecordValue(v any) any {
	switch typed := v.(type) {
	case connectors.Record:
		return deepCloneRecord(typed)
	case map[string]any:
		out := make(map[string]any, len(typed))
		for k, value := range typed {
			out[k] = deepCloneRecordValue(value)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, value := range typed {
			out[i] = deepCloneRecordValue(value)
		}
		return out
	case []connectors.Record:
		out := make([]connectors.Record, len(typed))
		for i, value := range typed {
			out[i] = deepCloneRecord(value)
		}
		return out
	default:
		return v
	}
}

// resolveRecordParent walks a dotted field path and returns the map that holds
// the leaf together with the leaf key. It reports false when any intermediate
// segment is absent or is not itself a record, which is the single traversal
// every field helper below shares.
func resolveRecordParent(record connectors.Record, field string) (map[string]any, string, bool) {
	if field == "" {
		return nil, "", false
	}
	parts := strings.Split(field, ".")
	current := map[string]any(record)
	for _, part := range parts[:len(parts)-1] {
		next, ok := current[part]
		if !ok {
			return nil, "", false
		}
		nested, ok := asRecordMap(next)
		if !ok {
			return nil, "", false
		}
		current = nested
	}
	return current, parts[len(parts)-1], true
}

func redactReversePlanField(record connectors.Record, field string) {
	parent, leaf, ok := resolveRecordParent(record, field)
	if !ok {
		return
	}
	if _, present := parent[leaf]; present {
		parent[leaf] = "redacted"
	}
}

// withholdRecordFields returns a clone of in with every declared field removed
// outright, alongside the fields it actually removed. The key is absent rather
// than set to a placeholder: an absent key is unambiguous, where a placeholder
// is indistinguishable from an operator who really supplied that string and
// would be dispatched verbatim.
//
// The second return value is what makes re-supply solvable. A declared field
// the operator never supplied was never in the record and was never withheld,
// so demanding it back would be an unsatisfiable precondition on a plan whose
// hash was computed without it. Only the fields removed here are owed back.
//
// fields must already be record-relative; connectorCommandRedactFields strips
// the declaring surface's prefix.
func withholdRecordFields(in connectors.Record, fields []string) (connectors.Record, []string) {
	out := deepCloneRecord(in)
	var withheld []string
	for _, field := range fields {
		trimmed := strings.TrimSpace(field)
		if deleteRecordField(out, trimmed) {
			withheld = append(withheld, trimmed)
		}
	}
	return out, withheld
}

// deleteRecordField removes a dotted field and reports whether it was present.
func deleteRecordField(record connectors.Record, field string) bool {
	parent, leaf, ok := resolveRecordParent(record, field)
	if !ok {
		return false
	}
	if _, present := parent[leaf]; !present {
		return false
	}
	delete(parent, leaf)
	return true
}

// recordHasField reports whether a dotted field is present. A plan written
// before fields were withheld still carries them, so nothing is re-supplied for
// it.
func recordHasField(record connectors.Record, field string) bool {
	parent, leaf, ok := resolveRecordParent(record, strings.TrimSpace(field))
	if !ok {
		return false
	}
	_, present := parent[leaf]
	return present
}

// mergeRecordFields overlays supplied onto base, descending into nested maps so
// a withheld leaf is restored without discarding its siblings.
func mergeRecordFields(base, overlay connectors.Record) connectors.Record {
	out := deepCloneRecord(base)
	for key, value := range overlay {
		out[key] = mergeRecordValue(out[key], value)
	}
	return out
}

func mergeRecordValue(existing, overlay any) any {
	existingMap, okExisting := asRecordMap(existing)
	overlayMap, okOverlay := asRecordMap(overlay)
	if !okExisting || !okOverlay {
		return overlay
	}
	merged := make(map[string]any, len(existingMap)+len(overlayMap))
	for key, value := range existingMap {
		merged[key] = value
	}
	for key, value := range overlayMap {
		merged[key] = mergeRecordValue(merged[key], value)
	}
	return merged
}

func asRecordMap(value any) (map[string]any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		return typed, true
	case connectors.Record:
		return map[string]any(typed), true
	default:
		return nil, false
	}
}

// connectorCommandRedactFields answers "what does this plan withhold?" for a
// connector-command plan. It dispatches on mode and never falls back between
// the two branches: operation IDs and write-action names are separate
// namespaces that collide by name in at least one bundle, so a fallback could
// withhold an unrelated action's fields. An operation whose metadata cannot be
// resolved is an error, never an empty withhold set.
func connectorCommandRedactFields(connector connectors.Connector, operation, actionName string) ([]string, bool, error) {
	operation = strings.TrimSpace(operation)
	prefix := connectorCommandRecordPrefix(operation)
	if operation == "" {
		return recordRelativeFields(reversePlanRedactFields(connector, actionName), prefix), false, nil
	}
	metadata, err := connectorCommandDirectWriteMetadata(connector, operation)
	if err != nil {
		return nil, false, err
	}
	return recordRelativeFields(metadata.RedactFields, prefix), metadata.StructuredBody, nil
}

func connectorCommandDirectWriteMetadata(connector connectors.Connector, operation string) (connectors.OperationDirectWriteMetadata, error) {
	provider, ok := connector.(connectors.OperationDirectWriteMetadataProvider)
	if !ok {
		return connectors.OperationDirectWriteMetadata{}, fmt.Errorf("connector %q does not expose direct-write metadata for operation %q", connector.Name(), operation)
	}
	metadata, err := provider.OperationDirectWriteMetadata(operation)
	if err != nil {
		return connectors.OperationDirectWriteMetadata{}, err
	}
	if metadata.Operation != operation {
		return connectors.OperationDirectWriteMetadata{}, fmt.Errorf("connector %q direct-write metadata did not match operation %q", connector.Name(), operation)
	}
	return metadata, nil
}

func connectorCommandPlanRecords(connector connectors.Connector, operation string, structuredBody bool, record, redacted connectors.Record, fields []string) (connectors.Record, []string, []connectors.Record, error) {
	if !structuredBody {
		withheld, withheldFields := withholdRecordFields(record, fields)
		return withheld, withheldFields, RedactReversePlanRecords([]connectors.Record{cloneRecord(redacted)}, fields), nil
	}
	transformer, ok := connector.(connectors.OperationDirectWriteBodyPlanTransformer)
	if !ok {
		return nil, nil, nil, fmt.Errorf("connector %q does not expose direct-write body plan transformation", connector.Name())
	}
	withheld, withheldFields, err := transformer.WithholdOperationDirectWriteBodyFields(operation, map[string]any(record), fields)
	if err != nil {
		return nil, nil, nil, err
	}
	sample, err := transformer.RedactOperationDirectWriteBodyFields(operation, map[string]any(redacted), fields)
	if err != nil {
		return nil, nil, nil, err
	}
	return connectors.Record(withheld), withheldFields, []connectors.Record{connectors.Record(sample)}, nil
}

// connectorCommandRecordPrefix is the field-path prefix the declaring surface
// uses for the record a connector-command plan carries. A direct_write
// operation's record IS its request body, so it declares body.<path>; a write
// action declares record.<path>. commandrunner.ReconstituteWithheldFields
// dispatches on the same rule, and the two halves have to agree or a withheld
// field silently stays on disk.
func connectorCommandRecordPrefix(operation string) string {
	if strings.TrimSpace(operation) != "" {
		return "body."
	}
	return "record."
}

// recordRelativeFields strips the declaring surface's prefix so every field
// path a plan persists is relative to the record itself.
func recordRelativeFields(fields []string, prefix string) []string {
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		trimmed := strings.TrimPrefix(strings.TrimSpace(field), prefix)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func reversePlanRedactFields(connector connectors.Connector, actionName string) []string {
	for _, action := range connectors.ManifestOf(connector).WriteActions {
		if action.Name == actionName {
			return append([]string(nil), action.RedactFields...)
		}
	}
	return nil
}

func mapReverseRecords(records []connectors.Record, mappings map[string]string) []connectors.Record {
	mapped := make([]connectors.Record, 0, len(records))
	for _, record := range records {
		out := connectors.Record{}
		for source, dest := range mappings {
			out[dest] = record[source]
		}
		mapped = append(mapped, out)
	}
	return mapped
}

func hashString(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func hashJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return hashString(string(b)), nil
}

func reversePlanHash(planName, sourceTable, destinationConnector, destinationCredential, action string, destinationConfig, mappings map[string]string, mapped []connectors.Record, payloadIdentity []PayloadIdentity) (string, error) {
	payload := map[string]any{
		"name":                   planName,
		"source_table":           sourceTable,
		"destination_connector":  destinationConnector,
		"destination_credential": destinationCredential,
		"destination_config":     cloneStringMap(destinationConfig),
		"action":                 action,
		"mappings":               cloneStringMap(mappings),
		"record_count":           len(mapped),
		"records":                cloneRecords(mapped),
	}
	if len(payloadIdentity) > 0 {
		payload["payload_identity"] = append([]PayloadIdentity(nil), payloadIdentity...)
	}
	return hashJSON(payload)
}

func connectorCommandPlanHash(planName, connector, credential string, config map[string]string, command string, path []string, action string, record connectors.Record, payloadIdentity []PayloadIdentity) (string, error) {
	payload := map[string]any{
		"name":         planName,
		"connector":    connector,
		"credential":   credential,
		"config":       cloneStringMap(config),
		"command":      command,
		"path":         append([]string(nil), path...),
		"action":       action,
		"record_count": 1,
		"record":       cloneRecord(record),
	}
	if len(payloadIdentity) > 0 {
		payload["payload_identity"] = append([]PayloadIdentity(nil), payloadIdentity...)
	}
	return hashJSON(payload)
}

// operationConnectorCommandPlanHash is the direct_write variant of a
// connector-command plan hash. The path/query/body fields are all part of the
// approved request, so they must be bound before a preview can mint a
// single-use approval token.
func operationConnectorCommandPlanHash(planName, connector, credential string, config map[string]string, command string, path []string, operation string, pathParams, query, headers map[string]string, headerValues map[string][]string, body connectors.Record, payloadIdentity []PayloadIdentity) (string, error) {
	payload := map[string]any{
		"name":         planName,
		"connector":    connector,
		"credential":   credential,
		"config":       cloneStringMap(config),
		"command":      command,
		"path":         append([]string(nil), path...),
		"operation":    operation,
		"path_params":  cloneStringMap(pathParams),
		"query":        cloneStringMap(query),
		"headers":      cloneStringMap(headers),
		"record_count": 1,
		"body":         cloneRecord(body),
	}
	if len(headerValues) > 0 {
		payload["header_values"] = cloneStringSliceMap(headerValues)
	}
	if len(payloadIdentity) > 0 {
		payload["payload_identity"] = append([]PayloadIdentity(nil), payloadIdentity...)
	}
	return hashJSON(payload)
}

func approvedPayloadSHA256(identities []PayloadIdentity) map[string]string {
	if len(identities) == 0 {
		return nil
	}
	approved := make(map[string]string, len(identities))
	for _, identity := range identities {
		if identity.ContentSHA256 == "" {
			continue
		}
		approved[connectors.PayloadApprovalKey(identity.RecordIndex, identity.Field)] = identity.ContentSHA256
	}
	if len(approved) == 0 {
		return nil
	}
	return approved
}

func payloadIdentitiesForRecords(projectDir string, records []connectors.Record) ([]PayloadIdentity, error) {
	fields := make(map[string]struct{})
	for _, record := range records {
		for key := range record {
			if isPayloadPathField(key) {
				fields[key] = struct{}{}
			}
		}
	}
	return payloadIdentitiesForDeclaredFields(projectDir, records, fieldSetSlice(fields))
}

// payloadIdentitiesForConnectorCommand uses the connector's closed operation
// metadata for direct-write file fields and its binary-upload declaration for
// the public binary_upload path. Legacy generic reverse-ETL commands retain
// their established file_path discovery behavior.
func payloadIdentitiesForConnectorCommand(projectDir string, connector connectors.Connector, operation, intent, action string, record connectors.Record) ([]PayloadIdentity, error) {
	if strings.TrimSpace(operation) == "" {
		if intent == "binary_upload" {
			provider, ok := connector.(interface {
				PreflightBinaryUploadAction(string) ([]connectors.BinaryUploadSource, error)
			})
			if !ok {
				return nil, fmt.Errorf("connector %q does not expose a binary-upload declaration", connector.Name())
			}
			sources, err := provider.PreflightBinaryUploadAction(action)
			if err != nil {
				return nil, err
			}
			fields := make([]string, 0, len(sources))
			caps := make(map[string]int64, len(sources))
			for _, source := range sources {
				fields = append(fields, source.Field)
				caps[source.Field] = source.MaxBytes
			}
			return payloadIdentitiesForDeclaredFieldsWithCaps(projectDir, []connectors.Record{record}, fields, caps)
		}
		return payloadIdentitiesForRecords(projectDir, []connectors.Record{record})
	}
	provider, ok := connector.(connectors.OperationDirectWriteMetadataProvider)
	if !ok {
		return nil, fmt.Errorf("connector %q does not expose direct-write metadata", connector.Name())
	}
	metadata, err := provider.OperationDirectWriteMetadata(operation)
	if err != nil {
		return nil, err
	}
	if metadata.Operation != operation {
		return nil, fmt.Errorf("connector %q direct-write metadata did not match operation %q", connector.Name(), operation)
	}
	if metadata.PayloadFileFields != nil {
		return payloadIdentitiesForDeclaredFieldsWithCaps(projectDir, []connectors.Record{record}, metadata.PayloadFileFields, metadata.PayloadFileMaxBytes)
	}
	return payloadIdentitiesForRecords(projectDir, []connectors.Record{record})
}

// payloadIdentitiesForDeclaredFields captures only declaration-owned file
// paths. It never guesses from arbitrary user fields, so multipart support
// cannot broaden the accepted local-file input surface.
func payloadIdentitiesForDeclaredFields(projectDir string, records []connectors.Record, declaredFields []string) ([]PayloadIdentity, error) {
	return payloadIdentitiesForDeclaredFieldsWithCaps(projectDir, records, declaredFields, nil)
}

func payloadIdentitiesForDeclaredFieldsWithCaps(projectDir string, records []connectors.Record, declaredFields []string, maxBytes map[string]int64) ([]PayloadIdentity, error) {
	var identities []PayloadIdentity
	fields := fieldSetSlice(stringSliceSet(declaredFields))
	for i, record := range records {
		for _, field := range fields {
			value, present := recordPathValue(record, field)
			raw, ok := value.(string)
			if !present || !ok || strings.TrimSpace(raw) == "" {
				continue
			}
			limit := int64(0)
			if maxBytes != nil {
				var ok bool
				limit, ok = maxBytes[field]
				if !ok || limit <= 0 {
					return nil, fmt.Errorf("payload identity for %s: missing declared byte cap", field)
				}
			}
			identity, err := payloadIdentityForPathWithCap(projectDir, i, field, raw, limit)
			if err != nil {
				return nil, err
			}
			identities = append(identities, identity)
		}
	}
	return identities, nil
}

func stringSliceSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			set[value] = struct{}{}
		}
	}
	return set
}

func fieldSetSlice(fields map[string]struct{}) []string {
	keys := make([]string, 0, len(fields))
	for field := range fields {
		keys = append(keys, field)
	}
	sort.Strings(keys)
	return keys
}

func recordPathValue(record connectors.Record, field string) (any, bool) {
	var current any = map[string]any(record)
	for _, segment := range strings.Split(field, ".") {
		if strings.TrimSpace(segment) == "" {
			return nil, false
		}
		var values map[string]any
		switch typed := current.(type) {
		case map[string]any:
			values = typed
		case connectors.Record:
			values = map[string]any(typed)
		default:
			return nil, false
		}
		value, ok := values[segment]
		if !ok {
			return nil, false
		}
		current = value
	}
	return current, true
}

func isPayloadPathField(name string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(name, "-", "_"))
	return strings.Contains(normalized, "file_path")
}

func payloadIdentityForPathWithCap(projectDir string, recordIndex int, field, raw string, maxBytes int64) (PayloadIdentity, error) {
	resolved, err := resolvePayloadPath(projectDir, raw)
	if err != nil {
		return PayloadIdentity{}, fmt.Errorf("payload identity for %s: %w", field, err)
	}
	contentDigest, info, err := digestPayloadFileWithCap(resolved, maxBytes)
	if err != nil {
		return PayloadIdentity{}, fmt.Errorf("payload identity for %s: %w", field, err)
	}
	return PayloadIdentity{
		RecordIndex:     recordIndex,
		Field:           field,
		PathHash:        hashString(resolved),
		ContentSHA256:   contentDigest,
		SizeBytes:       info.Size(),
		ModTimeUnixNano: info.ModTime().UTC().UnixNano(),
	}, nil
}

func digestPayloadFile(path string) (string, os.FileInfo, error) {
	return digestPayloadFileWithCap(path, 0)
}

func digestPayloadFileWithCap(path string, maxBytes int64) (string, os.FileInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", nil, err
	}
	defer func() { _ = file.Close() }()

	before, err := file.Stat()
	if err != nil {
		return "", nil, err
	}
	if !before.Mode().IsRegular() {
		return "", nil, fmt.Errorf("file must be a regular file")
	}
	if maxBytes > 0 && before.Size() > maxBytes {
		return "", nil, fmt.Errorf("payload file exceeds declared byte cap %d", maxBytes)
	}
	hash := sha256.New()
	reader := io.Reader(file)
	if maxBytes > 0 {
		reader = io.LimitReader(file, maxBytes)
	}
	if _, err := io.Copy(hash, reader); err != nil {
		return "", nil, err
	}
	if maxBytes > 0 {
		var extra [1]byte
		n, readErr := file.Read(extra[:])
		if n > 0 {
			return "", nil, fmt.Errorf("payload file exceeds declared byte cap %d", maxBytes)
		}
		if readErr != nil && readErr != io.EOF {
			return "", nil, readErr
		}
	}
	after, err := file.Stat()
	if err != nil {
		return "", nil, err
	}
	if before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
		return "", nil, fmt.Errorf("payload file changed while computing approval identity")
	}
	if maxBytes > 0 && after.Size() > maxBytes {
		return "", nil, fmt.Errorf("payload file exceeds declared byte cap %d", maxBytes)
	}
	return hex.EncodeToString(hash.Sum(nil)), after, nil
}

func resolvePayloadPath(projectDir, raw string) (string, error) {
	if strings.TrimSpace(projectDir) == "" {
		projectDir = "."
	}
	if err := safety.ValidateLocalWritePath(projectDir, raw, "payload file path", false); err != nil {
		return "", err
	}
	rootAbs, err := filepath.Abs(projectDir)
	if err != nil {
		return "", fmt.Errorf("resolve project root: %w", err)
	}
	if resolvedRoot, err := filepath.EvalSymlinks(rootAbs); err == nil {
		rootAbs = resolvedRoot
	}
	candidate := raw
	if !filepath.IsAbs(raw) {
		candidate = filepath.Join(rootAbs, filepath.Clean(raw))
	}
	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(rootAbs, resolved)
	if err != nil {
		return "", fmt.Errorf("compare payload file path to project root: %w", err)
	}
	if rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)) {
		return resolved, nil
	}
	return "", fmt.Errorf("payload file path outside the project root is not allowed")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func findCatalogStream(catalog connectors.Catalog, name string) (connectors.Stream, bool) {
	for _, stream := range catalog.Streams {
		if stream.Name == name {
			return stream, true
		}
	}
	return connectors.Stream{}, false
}
