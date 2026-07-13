package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"unicode"
)

const (
	asanaBaselineIdentityCount  = 250
	asanaBaselineIdentityDigest = "aa652bd7c4e127ca517f615e6628f7d92c34388089a7490b392f177e2eb530d6"
	asanaOpenAPIRevision        = "2b84af9d1f2bdd840052a1c7ddbee0ecb6a80869"
	asanaOpenAPISHA256          = "0592cd896d2cb52f5a01393c0f86258ead9e896b769780c5891e2375aa4a5150"
)

type asanaAPISurface struct {
	Scope          string                    `json:"scope"`
	SourceEvidence *asanaAPISurfaceEvidence  `json:"source_evidence"`
	Endpoints      []asanaAPISurfaceEndpoint `json:"endpoints"`
}

type asanaAPISurfaceEvidence struct {
	Kind                       string `json:"kind"`
	SourceURL                  string `json:"source_url"`
	Revision                   string `json:"revision"`
	SHA256                     string `json:"sha256"`
	CapturedAt                 string `json:"captured_at"`
	BaselineOperationCount     int    `json:"baseline_operation_count"`
	SupplementalOperationsNote string `json:"supplemental_operations_note"`
}

type asanaAPISurfaceEndpoint struct {
	Method    string         `json:"method"`
	Path      string         `json:"path"`
	CoveredBy map[string]any `json:"covered_by"`
	Excluded  map[string]any `json:"excluded"`
}

func TestAsanaAPISurfaceBaseline(t *testing.T) {
	surface := loadAsanaAPISurface(t)

	digest, err := asanaIdentityDigest(surface.Endpoints)
	if err != nil {
		t.Fatalf("validate Asana operation identities: %v", err)
	}
	if got := len(surface.Endpoints); got != asanaBaselineIdentityCount {
		t.Fatalf("endpoint count = %d, want %d", got, asanaBaselineIdentityCount)
	}
	if digest != asanaBaselineIdentityDigest {
		t.Fatalf("identity digest = %q, want %q", digest, asanaBaselineIdentityDigest)
	}

	t.Run("classifier changes do not affect identity digest", func(t *testing.T) {
		changed := append([]asanaAPISurfaceEndpoint(nil), surface.Endpoints...)
		changed[0].CoveredBy = nil
		changed[0].Excluded = map[string]any{"category": "test-only"}

		got, err := asanaIdentityDigest(changed)
		if err != nil {
			t.Fatalf("digest reclassified identities: %v", err)
		}
		if got != digest {
			t.Fatalf("classifier-only change altered identity digest: got %q, want %q", got, digest)
		}
	})

	for _, tc := range []struct {
		name      string
		endpoints []asanaAPISurfaceEndpoint
		wantError string
	}{
		{
			name: "duplicate identity",
			endpoints: []asanaAPISurfaceEndpoint{
				{Method: "GET", Path: "/tasks/{task_gid}"},
				{Method: "GET", Path: "/tasks/{task_gid}"},
			},
			wantError: "duplicate operation identity",
		},
		{
			name:      "lowercase method",
			endpoints: []asanaAPISurfaceEndpoint{{Method: "get", Path: "/tasks"}},
			wantError: "uppercase HTTP method",
		},
		{
			name:      "absolute URL path",
			endpoints: []asanaAPISurfaceEndpoint{{Method: "GET", Path: "https://app.asana.com/api/1.0/tasks"}},
			wantError: "connector-relative path",
		},
		{
			name:      "query-bearing path",
			endpoints: []asanaAPISurfaceEndpoint{{Method: "GET", Path: "/tasks?workspace=123"}},
			wantError: "canonical path",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := asanaIdentityDigest(tc.endpoints); err == nil || !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("asanaIdentityDigest() error = %v, want containing %q", err, tc.wantError)
			}
		})
	}
}

func TestAsanaAPISurfaceProvenance(t *testing.T) {
	surface := loadAsanaAPISurface(t)
	if surface.SourceEvidence == nil {
		t.Fatal("source_evidence is required for the Asana baseline")
	}

	evidence := surface.SourceEvidence
	if evidence.Kind != "openapi" {
		t.Fatalf("source_evidence.kind = %q, want openapi", evidence.Kind)
	}
	if evidence.Revision != asanaOpenAPIRevision {
		t.Fatalf("source_evidence.revision = %q, want %q", evidence.Revision, asanaOpenAPIRevision)
	}
	wantURL := "https://raw.githubusercontent.com/Asana/openapi/" + asanaOpenAPIRevision + "/defs/asana_oas.yaml"
	if evidence.SourceURL != wantURL {
		t.Fatalf("source_evidence.source_url = %q, want %q", evidence.SourceURL, wantURL)
	}
	if evidence.SHA256 != asanaOpenAPISHA256 {
		t.Fatalf("source_evidence.sha256 = %q, want %q", evidence.SHA256, asanaOpenAPISHA256)
	}
	if evidence.CapturedAt != "2026-07-04" {
		t.Fatalf("source_evidence.captured_at = %q, want 2026-07-04", evidence.CapturedAt)
	}
	if evidence.BaselineOperationCount != asanaBaselineIdentityCount {
		t.Fatalf("source_evidence.baseline_operation_count = %d, want %d", evidence.BaselineOperationCount, asanaBaselineIdentityCount)
	}
	if !strings.Contains(evidence.SupplementalOperationsNote, "/users/me") ||
		!strings.Contains(evidence.SupplementalOperationsNote, "not an executable capability grant") {
		t.Fatalf("source_evidence supplemental note must identify /users/me and deny an executable grant: %q", evidence.SupplementalOperationsNote)
	}
}

func TestAsanaAPISurfaceClassifiers(t *testing.T) {
	surface := loadAsanaAPISurface(t)

	streamRows := 0
	writeRows := 0
	exclusions := 0
	streams := make(map[string]struct{})
	writes := make(map[string]struct{})

	for i, endpoint := range surface.Endpoints {
		hasCoveredBy := endpoint.CoveredBy != nil
		hasExcluded := endpoint.Excluded != nil
		if hasCoveredBy == hasExcluded {
			t.Fatalf("endpoint %d %s %s must have exactly one of covered_by or excluded", i, endpoint.Method, endpoint.Path)
		}
		if hasExcluded {
			exclusions++
			continue
		}
		if len(endpoint.CoveredBy) != 1 {
			t.Fatalf("endpoint %d %s %s covered_by must have exactly one classifier, got %v", i, endpoint.Method, endpoint.Path, endpoint.CoveredBy)
		}
		if stream, ok := endpoint.CoveredBy["stream"].(string); ok && stream != "" {
			streamRows++
			streams[stream] = struct{}{}
			continue
		}
		if write, ok := endpoint.CoveredBy["write"].(string); ok && write != "" {
			writeRows++
			writes[write] = struct{}{}
			continue
		}
		t.Fatalf("endpoint %d %s %s has malformed covered_by classifier %v", i, endpoint.Method, endpoint.Path, endpoint.CoveredBy)
	}

	if streamRows != 13 || len(streams) != 12 {
		t.Fatalf("stream accounting = %d rows / %d unique streams, want 13 / 12", streamRows, len(streams))
	}
	if writeRows != 13 || len(writes) != 13 {
		t.Fatalf("write accounting = %d rows / %d unique writes, want 13 / 13", writeRows, len(writes))
	}
	if exclusions != 224 {
		t.Fatalf("exclusion accounting = %d, want 224", exclusions)
	}
	if !strings.Contains(surface.Scope, "12 streams covering 13 GET rows") {
		t.Fatalf("scope does not describe 12 streams covering 13 GET rows: %q", surface.Scope)
	}
}

func loadAsanaAPISurface(t *testing.T) asanaAPISurface {
	t.Helper()

	raw, err := os.ReadFile("../../internal/connectors/defs/asana/api_surface.json")
	if err != nil {
		t.Fatalf("read Asana api_surface.json: %v", err)
	}
	var surface asanaAPISurface
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal Asana api_surface.json: %v", err)
	}
	return surface
}

// asanaIdentityDigest validates canonical connector-relative operation
// identities, sorts them, and hashes METHOD + " " + path records joined by LF
// without a trailing LF. Classifiers are intentionally outside this digest.
func asanaIdentityDigest(endpoints []asanaAPISurfaceEndpoint) (string, error) {
	identities := make([]string, 0, len(endpoints))
	seen := make(map[string]struct{}, len(endpoints))
	for i, endpoint := range endpoints {
		if endpoint.Method == "" || endpoint.Method != strings.ToUpper(endpoint.Method) ||
			strings.IndexFunc(endpoint.Method, func(r rune) bool { return r < 'A' || r > 'Z' }) >= 0 {
			return "", fmt.Errorf("endpoint %d must use a non-empty uppercase HTTP method", i)
		}
		if !strings.HasPrefix(endpoint.Path, "/") || strings.HasPrefix(endpoint.Path, "//") {
			return "", fmt.Errorf("endpoint %d must use a connector-relative path", i)
		}
		if strings.ContainsAny(endpoint.Path, "?#\\") ||
			strings.IndexFunc(endpoint.Path, unicode.IsSpace) >= 0 ||
			strings.IndexFunc(endpoint.Path, func(r rune) bool { return unicode.IsControl(r) }) >= 0 {
			return "", fmt.Errorf("endpoint %d must use a canonical path without query, fragment, backslash, whitespace, or control characters", i)
		}
		for _, segment := range strings.Split(strings.TrimPrefix(endpoint.Path, "/"), "/") {
			if segment == "" || segment == "." || segment == ".." {
				return "", fmt.Errorf("endpoint %d must use a canonical path without empty or dot segments", i)
			}
		}

		identity := endpoint.Method + " " + endpoint.Path
		if _, exists := seen[identity]; exists {
			return "", fmt.Errorf("duplicate operation identity %q", identity)
		}
		seen[identity] = struct{}{}
		identities = append(identities, identity)
	}

	sort.Strings(identities)
	sum := sha256.Sum256([]byte(strings.Join(identities, "\n")))
	return hex.EncodeToString(sum[:]), nil
}
