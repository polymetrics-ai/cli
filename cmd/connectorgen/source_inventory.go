package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"syscall"
	"unicode"
)

// sourceOperationKey identifies retained evidence, never an execution unit.
type sourceOperationKey struct {
	Connector string `json:"connector"`
	Inventory string `json:"inventory"`
	ID        string `json:"id"`
}

type sourceArtifactPin struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Bytes  int64  `json:"bytes"`
}

type sourceInventoryAnchor struct {
	Connector       string              `json:"connector"`
	Inventory       string              `json:"inventory"`
	Class           string              `json:"class"`
	Path            string              `json:"path"`
	SHA256          string              `json:"sha256"`
	AcquisitionRef  string              `json:"acquisition_ref"`
	AcquisitionPath string              `json:"acquisition_path"`
	ExpectedIDs     []string            `json:"expected_ids"`
	ExpectedCount   int                 `json:"expected_count"`
	Artifacts       []sourceArtifactPin `json:"artifacts"`
}

type sourceLaneCohort struct {
	SchemaVersion int                     `json:"schema_version"`
	CohortID      string                  `json:"cohort_id"`
	Inventories   []sourceInventoryAnchor `json:"inventories"`
}

type sourceLaneDiagnostic struct {
	Key      sourceOperationKey `json:"key"`
	Lanes    []string           `json:"lanes"`
	Stage    string             `json:"stage"`
	Code     string             `json:"code"`
	Pointer  string             `json:"pointer"`
	Owner    string             `json:"owner"`
	Severity string             `json:"severity"`
}

type retainedSourceOperation struct {
	Key            sourceOperationKey     `json:"key"`
	Class          string                 `json:"class"`
	DocumentID     string                 `json:"document_id"`
	Pointer        string                 `json:"pointer"`
	SourceLocation string                 `json:"source_location"`
	Observed       bool                   `json:"observed"`
	Node           json.RawMessage        `json:"source_node"`
	Diagnostics    []sourceLaneDiagnostic `json:"diagnostics"`
}

type retainedSourceDocument struct {
	ContentType            string          `json:"content_type"`
	ID                     string          `json:"id"`
	Path                   string          `json:"path"`
	RetainedFileSHA256     string          `json:"retained_file_sha256"`
	Bytes                  int64           `json:"bytes"`
	UpstreamDeclaredSHA256 string          `json:"upstream_declared_sha256"`
	UpstreamDeclaredBytes  int64           `json:"upstream_declared_bytes"`
	UpstreamBytesVerified  bool            `json:"upstream_bytes_verified"`
	Payload                json.RawMessage `json:"payload"`
}

type retainedSourceInventory struct {
	Operations  []retainedSourceOperation `json:"operations"`
	Documents   []retainedSourceDocument  `json:"documents"`
	Diagnostics []sourceLaneDiagnostic    `json:"diagnostics"`
}

// loadRetainedSourceInventory allocates the anchored universe before reading
// provider evidence. A failed source cannot change that universe.
func loadRetainedSourceInventory(ctx context.Context, repo string, cohort sourceLaneCohort) retainedSourceInventory {
	result := retainedSourceInventory{Operations: []retainedSourceOperation{}, Documents: []retainedSourceDocument{}, Diagnostics: []sourceLaneDiagnostic{}}
	index := map[sourceOperationKey]int{}
	for _, anchor := range cohort.Inventories {
		for _, id := range anchor.ExpectedIDs {
			key := sourceOperationKey{Connector: anchor.Connector, Inventory: anchor.Inventory, ID: id}
			if _, exists := index[key]; exists {
				continue
			}
			index[key] = len(result.Operations)
			result.Operations = append(result.Operations, retainedSourceOperation{Key: key, Class: anchor.Class, DocumentID: anchor.Connector + ":" + anchor.Inventory, Node: json.RawMessage("null"), Diagnostics: []sourceLaneDiagnostic{}})
		}
	}
	add := func(key sourceOperationKey, code, pointer string) {
		d := sourceLaneDiagnostic{Key: key, Lanes: sourceLaneNames(), Stage: "inventory", Code: code, Pointer: pointer, Owner: key.Connector, Severity: "error"}
		result.Diagnostics = append(result.Diagnostics, d)
		if n, ok := index[key]; ok {
			result.Operations[n].Diagnostics = append(result.Operations[n].Diagnostics, d)
		}
	}
	root, err := os.OpenRoot(repo)
	if err != nil {
		for key := range index {
			add(key, "source_root_unavailable", "")
		}
		sortSourceInventory(&result)
		return result
	}
	defer func() { _ = root.Close() }() // Root holds read-only directory authority, no durable writes.
	var totalBytes int64
	for _, anchor := range cohort.Inventories {
		base := sourceOperationKey{Connector: anchor.Connector, Inventory: anchor.Inventory}
		fail := func(code string) {
			for _, id := range anchor.ExpectedIDs {
				key := base
				key.ID = id
				add(key, code, anchor.Path)
			}
		}
		if err := ctx.Err(); err != nil {
			fail("source_canceled")
			continue
		}
		data, err := readSourceInput(root, anchor.Path, 64<<20)
		if err != nil {
			fail("source_unavailable")
			continue
		}
		totalBytes += int64(len(data))
		if totalBytes > 512<<20 {
			fail("source_budget_exceeded")
			continue
		}
		if sourceBytesHash(data) != anchor.SHA256 {
			fail("source_hash_mismatch")
			continue
		}
		validArtifacts := true
		for _, pin := range anchor.Artifacts {
			raw, readErr := readSourceInput(root, pin.Path, 64<<20)
			totalBytes += int64(len(raw))
			if readErr != nil || int64(len(raw)) != pin.Bytes || sourceBytesHash(raw) != pin.SHA256 || totalBytes > 512<<20 {
				validArtifacts = false
			}
		}
		if !validArtifacts {
			fail("source_artifact_invalid")
			continue
		}
		var envelope map[string]json.RawMessage
		if err := decodeSourceJSON(data, &envelope); err != nil {
			fail("source_invalid")
			continue
		}
		var version int
		if err := json.Unmarshal(envelope["schema_version"], &version); err != nil || (version != 2 && version != 3) {
			fail("source_schema_invalid")
			continue
		}
		var connector string
		if err := json.Unmarshal(envelope["connector"], &connector); err != nil || connector != anchor.Connector {
			fail("source_connector_mismatch")
			continue
		}
		var rest struct {
			SHA256     string            `json:"sha256"`
			Bytes      int64             `json:"bytes"`
			Operations []json.RawMessage `json:"operations"`
			Documents  []struct {
				Operations []json.RawMessage `json:"operations"`
			} `json:"source_documents"`
		}
		if err := json.Unmarshal(envelope["rest"], &rest); err != nil {
			fail("source_invalid")
			continue
		}
		result.Documents = append(result.Documents, retainedSourceDocument{ID: anchor.Connector + ":" + anchor.Inventory, Path: anchor.Path, RetainedFileSHA256: anchor.SHA256, Bytes: int64(len(data)), UpstreamDeclaredSHA256: rest.SHA256, UpstreamDeclaredBytes: rest.Bytes, Payload: append(json.RawMessage(nil), data...)})
		rows := rest.Operations
		pointers := []string{}
		if version == 3 {
			rows = []json.RawMessage{}
			for d, doc := range rest.Documents {
				for n, row := range doc.Operations {
					rows = append(rows, row)
					pointers = append(pointers, fmt.Sprintf("/rest/source_documents/%d/operations/%d", d, n))
				}
			}
		} else {
			for n := range rows {
				pointers = append(pointers, fmt.Sprintf("/rest/operations/%d", n))
			}
		}
		observed := map[string]bool{}
		for n, row := range rows {
			var node struct {
				ID       string `json:"id"`
				Location string `json:"source_location"`
			}
			if err := json.Unmarshal(row, &node); err != nil || !validSourceID(node.ID) {
				add(base, "source_id_invalid", pointers[n])
				continue
			}
			key := base
			key.ID = node.ID
			if observed[node.ID] {
				add(key, "source_duplicate", pointers[n])
				continue
			}
			observed[node.ID] = true
			target, exists := index[key]
			if !exists {
				add(key, "source_unexpected", pointers[n])
				continue
			}
			result.Operations[target].Observed = true
			result.Operations[target].Node = append(json.RawMessage(nil), row...)
			result.Operations[target].Pointer = pointers[n]
			result.Operations[target].SourceLocation = node.Location
		}
		for _, id := range anchor.ExpectedIDs {
			if !observed[id] {
				key := base
				key.ID = id
				add(key, "source_missing", anchor.Path)
			}
		}
		var counts struct {
			Total *int `json:"total"`
		}
		if err := json.Unmarshal(envelope["counts"], &counts); err != nil || counts.Total == nil || *counts.Total != len(observed) || *counts.Total != anchor.ExpectedCount || anchor.ExpectedCount != len(anchor.ExpectedIDs) {
			add(base, "source_count_mismatch", "/counts/total")
		}
	}
	sortSourceInventory(&result)
	return result
}

func sourceLaneNames() []string {
	return []string{"direct_read", "direct_write", "binary_download", "binary_upload", "etl", "reverse_etl", "sync_transport"}
}

func sourceBytesHash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func validSourceID(id string) bool {
	return id != "" && strings.IndexFunc(id, unicode.IsControl) < 0
}

func sourceKeyLess(a, b sourceOperationKey) bool {
	if a.Connector != b.Connector {
		return a.Connector < b.Connector
	}
	if a.Inventory != b.Inventory {
		return a.Inventory < b.Inventory
	}
	return a.ID < b.ID
}

func sortSourceInventory(inv *retainedSourceInventory) {
	sort.Slice(inv.Operations, func(i, j int) bool { return sourceKeyLess(inv.Operations[i].Key, inv.Operations[j].Key) })
	sort.Slice(inv.Documents, func(i, j int) bool { return inv.Documents[i].ID < inv.Documents[j].ID })
	sort.Slice(inv.Diagnostics, func(i, j int) bool {
		a, b := inv.Diagnostics[i], inv.Diagnostics[j]
		if a.Key != b.Key {
			return sourceKeyLess(a.Key, b.Key)
		}
		if a.Code != b.Code {
			return a.Code < b.Code
		}
		return a.Pointer < b.Pointer
	})
}

// readSourceInput is confined to a caller-owned root and opens only regular
// no-follow files. Nonblocking open prevents substituted FIFOs from hanging.
func readSourceInput(root *os.Root, name string, limit int64) ([]byte, error) {
	if name == "" || strings.Contains(name, "\\") || strings.HasPrefix(name, "/") || path.Clean(name) != name || strings.IndexFunc(name, unicode.IsControl) >= 0 {
		return nil, fmt.Errorf("invalid source path %q", name)
	}
	prefix := ""
	for _, part := range strings.Split(name, "/") {
		if part == "." || part == ".." || part == "" {
			return nil, fmt.Errorf("invalid source path %q", name)
		}
		prefix = path.Join(prefix, part)
		info, err := root.Lstat(prefix)
		if err != nil {
			return nil, fmt.Errorf("inspect source: %w", err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("source path is a symlink")
		}
	}
	f, err := root.OpenFile(name, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, fmt.Errorf("open source: %w", err)
	}
	info, statErr := f.Stat()
	if statErr != nil || !info.Mode().IsRegular() {
		if statErr == nil {
			statErr = fmt.Errorf("source is not a regular file")
		}
		return nil, errors.Join(statErr, f.Close())
	}
	data, readErr := io.ReadAll(io.LimitReader(f, limit+1))
	closeErr := f.Close()
	if readErr != nil || closeErr != nil {
		return nil, errors.Join(readErr, closeErr)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("source byte limit exceeded")
	}
	return data, nil
}

// Provider objects deliberately remain opaque; the envelope around them is
// validated separately. Numbers retain their lexical precision.
func decodeSourceJSON(data []byte, destination any) error {
	depth := 0
	inString, escaped := false, false
	for _, b := range data {
		if inString {
			if escaped {
				escaped = false
			} else if b == '\\' {
				escaped = true
			} else if b == '"' {
				inString = false
			}
			continue
		}
		switch b {
		case '"':
			inString = true
		case '{', '[':
			depth++
			if depth > 256 {
				return fmt.Errorf("source nesting limit exceeded")
			}
		case '}', ']':
			depth--
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := rejectDuplicateJSONMembers(decoder); err != nil {
		return err
	}
	decoder = json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	return nil
}
