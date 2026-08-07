package engine

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

const RuntimeOperationEndpointLedgerFile = "operation_endpoint_ledger.json"

type OperationEndpointLedgerEntry struct {
	Method   string `json:"method"`
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	MaxBytes int    `json:"max_bytes"`
}

type operationEndpointLedger struct {
	entries []OperationEndpointLedgerEntry
}

func loadOperationEndpointLedgers(fsys fs.FS) (map[string]*operationEndpointLedger, error) {
	if !fileExists(fsys, RuntimeOperationEndpointLedgerFile) {
		return nil, nil
	}
	raw, err := readFile(fsys, RuntimeOperationEndpointLedgerFile)
	if err != nil {
		return nil, err
	}
	var source map[string][]OperationEndpointLedgerEntry
	if err := strictDecode(raw, &source); err != nil {
		return nil, fmt.Errorf("%s: %w", RuntimeOperationEndpointLedgerFile, err)
	}
	ledgers := make(map[string]*operationEndpointLedger, len(source))
	for name, entries := range source {
		if !namePattern.MatchString(name) {
			return nil, fmt.Errorf("%s: invalid connector name %q", RuntimeOperationEndpointLedgerFile, name)
		}
		if err := validateOperationEndpointLedgerEntries(entries); err != nil {
			return nil, fmt.Errorf("%s: connector %s: %w", RuntimeOperationEndpointLedgerFile, name, err)
		}
		ledgers[name] = &operationEndpointLedger{entries: append([]OperationEndpointLedgerEntry(nil), entries...)}
	}
	return ledgers, nil
}

func validateOperationEndpointLedgerEntries(entries []OperationEndpointLedgerEntry) error {
	seen := make(map[string]struct{}, len(entries))
	for i, entry := range entries {
		method := strings.ToUpper(strings.TrimSpace(entry.Method))
		if method != entry.Method || (method != "GET" && method != "POST") {
			return fmt.Errorf("entry %d has unsupported method %q", i+1, entry.Method)
		}
		if strings.TrimSpace(entry.Path) == "" || strings.TrimSpace(entry.Path) != entry.Path || isAbsoluteHTTPURL(entry.Path) || strings.HasPrefix(entry.Path, "//") {
			return fmt.Errorf("entry %d has invalid connector-relative path %q", i+1, entry.Path)
		}
		if entry.Kind != "rest_read" && entry.Kind != "provider_search" {
			return fmt.Errorf("entry %d has unsupported operation kind %q", i+1, entry.Kind)
		}
		if entry.MaxBytes <= 0 {
			return fmt.Errorf("entry %d requires positive max_bytes", i+1)
		}
		key := entry.Method + "\x00" + entry.Path + "\x00" + entry.Kind + "\x00" + fmt.Sprint(entry.MaxBytes)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("entry %d duplicates %s %s %s", i+1, entry.Method, entry.Path, entry.Kind)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func deriveOperationDirectReadEndpointLedger(operations []OperationSpec, surface *APISurface) *operationEndpointLedger {
	if surface == nil {
		return nil
	}
	entries := make([]OperationEndpointLedgerEntry, 0)
	seen := make(map[string]struct{})
	for _, operation := range operations {
		if (operation.Kind != "rest_read" && operation.Kind != "provider_search") || operation.REST == nil {
			continue
		}
		method := strings.ToUpper(strings.TrimSpace(operation.REST.Method))
		if (method != "GET" && method != "POST") || operation.REST.Path == "" || operation.REST.MaxBytes <= 0 {
			continue
		}
		if !hasOperationDirectReadSurfaceEndpoint(surface, method, operation.REST.Path) {
			continue
		}
		entry := OperationEndpointLedgerEntry{
			Method:   method,
			Path:     operation.REST.Path,
			Kind:     operation.Kind,
			MaxBytes: operation.REST.MaxBytes,
		}
		key := entry.Method + "\x00" + entry.Path + "\x00" + entry.Kind + "\x00" + fmt.Sprint(entry.MaxBytes)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Method != entries[j].Method {
			return entries[i].Method < entries[j].Method
		}
		if entries[i].Path != entries[j].Path {
			return entries[i].Path < entries[j].Path
		}
		if entries[i].Kind != entries[j].Kind {
			return entries[i].Kind < entries[j].Kind
		}
		return entries[i].MaxBytes < entries[j].MaxBytes
	})
	return &operationEndpointLedger{entries: entries}
}

func hasOperationDirectReadSurfaceEndpoint(surface *APISurface, method, endpointPath string) bool {
	for _, endpoint := range surface.Endpoints {
		if !strings.EqualFold(endpoint.Method, method) || endpoint.Path != endpointPath {
			continue
		}
		if endpoint.Operation != nil && endpoint.Operation.Model == "direct_read" {
			return true
		}
		if endpoint.CoveredBy != nil && (strings.TrimSpace(endpoint.CoveredBy.DirectRead) != "" || len(endpoint.CoveredBy.DirectReads) > 0) {
			return true
		}
	}
	return false
}

func operationDirectReadEndpointLedger(b Bundle) *operationEndpointLedger {
	if b.directReadLedger != nil {
		return b.directReadLedger
	}
	return deriveOperationDirectReadEndpointLedger(b.Operations, b.Surface)
}

func OperationDirectReadEndpointLedgerEntries(b Bundle) []OperationEndpointLedgerEntry {
	ledger := operationDirectReadEndpointLedger(b)
	if ledger == nil {
		return nil
	}
	return append([]OperationEndpointLedgerEntry{}, ledger.entries...)
}
