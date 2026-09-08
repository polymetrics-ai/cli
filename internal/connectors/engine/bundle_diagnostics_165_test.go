package engine

import (
	"encoding/json"
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

// The tuple is public diagnostic metadata; the original cause stays inspectable
// through the normal error chain. No declaration values belong in its rendering.
type bundleLocation165 interface {
	BundleLocation() (string, string, string, string)
}

func TestBundleDiagnosticIdentityAndCause165(t *testing.T) {
	var declaration map[string]any
	if err := json.Unmarshal([]byte(validProviderCitedRateLimits), &declaration); err != nil {
		t.Fatal(err)
	}
	policy := declaration["policies"].([]any)[0].(map[string]any)
	policy["id"] = "synthetic-private-token-165"
	declaration["policies"] = []any{policy, policy}
	duplicate, err := json.Marshal(declaration)
	if err != nil {
		t.Fatal(err)
	}
	var metadata map[string]any
	if err := json.Unmarshal([]byte(validMetadata("acme")), &metadata); err != nil {
		t.Fatal(err)
	}
	delete(metadata, "name")
	missingName, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, file, data, field string
		syntax, missing         bool
	}{
		{name: "rate_syntax", file: "rate_limits.json", data: `{"schema_version":}`, field: "/", syntax: true},
		{name: "rate_duplicate", file: "rate_limits.json", data: string(duplicate), field: "/policies/1/id"},
		{name: "rate_state", file: "rate_limits.json", data: `{"schema_version":1,"state":"synthetic-private-token-165","reason":"synthetic-private-token-165"}`, field: "/state"},
		{name: "metadata_syntax", file: "metadata.json", data: `{"name":}`, field: "/", syntax: true},
		{name: "missing_metadata_name", file: "metadata.json", data: string(missingName), field: "/name"},
		{name: "stream_schema_compile", file: "schemas/widgets.json", data: `{"type":"object","unknown_keyword":true}`, field: "/<member:1>"},
		{name: "missing_metadata", file: "metadata.json", field: "/", missing: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.syntax {
				var syntax *json.SyntaxError
				var value any
				control := json.Unmarshal([]byte(tc.data), &value)
				if !errors.As(control, &syntax) {
					t.Fatalf("invalid syntax oracle fixture: %v", control)
				}
			}
			files := fullValidBundleFS("acme")
			if tc.missing {
				delete(files, "acme/"+tc.file)
			} else {
				files["acme/"+tc.file] = &fstest.MapFile{Data: []byte(tc.data)}
			}
			_, err := Load(files, "acme")
			if err == nil {
				t.Fatal("malformed bundle accepted")
			}
			var location bundleLocation165
			if !errors.As(err, &location) {
				t.Errorf("selected error lost structured connector/generation/file/field: %v", err)
			} else {
				name, generation, file, field := location.BundleLocation()
				if name != "acme" || generation != "embedded-v1" || file != tc.file || field != tc.field {
					t.Errorf("wrong diagnostic tuple %q %q %q %q", name, generation, file, field)
				}
			}
			if strings.Contains(err.Error(), "synthetic-private-token-165") {
				t.Error("declaration value leaked through diagnostic rendering")
			}
			if tc.syntax {
				var syntax *json.SyntaxError
				if !errors.As(err, &syntax) {
					t.Errorf("original JSON syntax cause lost: %v", err)
				}
			}
			if tc.missing && !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("required-file absence cause lost: %v", err)
			}
		})
	}
	// A positive input control reaches the real loader with the same fixture.
	if _, err := Load(fullValidBundleFS("acme"), "acme"); err != nil {
		t.Fatal(err)
	}
}

// Existing authoring rejection tests inspect the actual retained cause. Public
// diagnostics are checked separately above and must never print these values.
func bundleCauseText165(err error) string {
	if err == nil {
		return ""
	}
	var diagnostic *BundleDiagnosticError
	if errors.As(err, &diagnostic) {
		return diagnostic.Cause.Error()
	}
	return err.Error()
}

func TestBundleCauseOracle165(t *testing.T) {
	for _, message := range []string{"expected rejection", "wrong but readable rejection"} {
		cause := errors.New(message)
		err := &BundleDiagnosticError{Connector: "acme", Generation: "g", File: "rate_limits.json", Field: "/", Cause: cause}
		if bundleCauseText165(err) != message || !errors.Is(err, cause) {
			t.Fatal("cause oracle lost actual producer result")
		}
	}
}
