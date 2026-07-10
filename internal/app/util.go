package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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

func mapReverseRecords(records []connectors.Record, mappings map[string]string, planID string) []connectors.Record {
	mapped := make([]connectors.Record, 0, len(records))
	for _, record := range records {
		out := connectors.Record{}
		for source, dest := range mappings {
			out[dest] = record[source]
		}
		if planID != "" {
			out["_polymetrics_reverse_plan_id"] = planID
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

func payloadIdentitiesForRecords(projectDir string, records []connectors.Record) ([]PayloadIdentity, error) {
	var identities []PayloadIdentity
	for i, record := range records {
		keys := make([]string, 0, len(record))
		for key := range record {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if !isPayloadPathField(key) {
				continue
			}
			raw, ok := record[key].(string)
			if !ok || strings.TrimSpace(raw) == "" {
				continue
			}
			identity, err := payloadIdentityForPath(projectDir, i, key, raw)
			if err != nil {
				return nil, err
			}
			identities = append(identities, identity)
		}
	}
	return identities, nil
}

func isPayloadPathField(name string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(name, "-", "_"))
	return strings.Contains(normalized, "file_path")
}

func payloadIdentityForPath(projectDir string, recordIndex int, field, raw string) (PayloadIdentity, error) {
	resolved, err := resolvePayloadPath(projectDir, raw)
	if err != nil {
		return PayloadIdentity{}, fmt.Errorf("payload identity for %s: %w", field, err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return PayloadIdentity{}, fmt.Errorf("payload identity for %s: %w", field, err)
	}
	if !info.Mode().IsRegular() {
		return PayloadIdentity{}, fmt.Errorf("payload identity for %s: file must be a regular file", field)
	}
	return PayloadIdentity{
		RecordIndex:     recordIndex,
		Field:           field,
		PathHash:        hashString(resolved),
		SizeBytes:       info.Size(),
		ModTimeUnixNano: info.ModTime().UTC().UnixNano(),
	}, nil
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

func parseSelectAll(sql string) (string, int, error) {
	fields := strings.Fields(strings.TrimSpace(strings.TrimSuffix(sql, ";")))
	if len(fields) < 4 {
		return "", 0, errors.New("only SELECT * FROM <table> [LIMIT n] is supported in the MVP")
	}
	if !strings.EqualFold(fields[0], "select") || fields[1] != "*" || !strings.EqualFold(fields[2], "from") {
		return "", 0, errors.New("only SELECT * FROM <table> [LIMIT n] is supported in the MVP")
	}
	table := fields[3]
	limit := 100
	if len(fields) > 4 {
		if len(fields) != 6 || !strings.EqualFold(fields[4], "limit") {
			return "", 0, errors.New("only SELECT * FROM <table> [LIMIT n] is supported in the MVP")
		}
		n, err := strconv.Atoi(fields[5])
		if err != nil || n <= 0 {
			return "", 0, fmt.Errorf("invalid limit %q", fields[5])
		}
		limit = n
	}
	return table, limit, nil
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
