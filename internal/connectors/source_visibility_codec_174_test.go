package connectors_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func sourceGzip174(t *testing.T, raw []byte) string {
	t.Helper()
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	if _, err := w.Write(raw); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func sourceRaw174(t *testing.T, payload string) []byte {
	t.Helper()
	r, err := gzip.NewReader(bytes.NewReader([]byte(payload)))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(io.LimitReader(r, connectors.SourceVisibilityConnectorLimit+1))
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	if len(raw) > connectors.SourceVisibilityConnectorLimit {
		t.Fatal("fixture over budget")
	}
	return raw
}

// JSON assignment lets this unchanged regression reach the old real decoder
// before the internal Encoding field exists, instead of failing compilation.
func sourceEncoding174(t *testing.T, a connectors.SourceVisibilityArtifact, encoding string) connectors.SourceVisibilityArtifact {
	t.Helper()
	raw, err := json.Marshal(map[string]string{"Encoding": encoding})
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &a); err != nil {
		t.Fatal(err)
	}
	return a
}

func TestSourceVisibilityGzipConsumer174(t *testing.T) {
	a := sourceArtifact162(t, "notion")
	raw := []byte(a.Payload)
	// Read the actual emitted fixture independently of the production decoder.
	if len(raw) >= 2 && raw[0] == 0x1f && raw[1] == 0x8b {
		r, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			t.Fatal(err)
		}
		raw, err = io.ReadAll(io.LimitReader(r, connectors.SourceVisibilityConnectorLimit+1))
		if err != nil {
			t.Fatal(err)
		}
		if err := r.Close(); err != nil {
			t.Fatal(err)
		}
	}
	var expected connectors.SourceVisibility
	if err := json.Unmarshal(raw, &expected); err != nil {
		t.Fatal(err)
	}
	if expected.Connector != "notion" || expected.OperationCount != 49 || expected.CellCount != 343 {
		t.Fatal("independent complete fixture identity changed")
	}
	encoded := sourceEncoding174(t, a, "gzip")
	encoded.Payload = sourceGzip174(t, raw)
	for _, tc := range []struct {
		name   string
		mutate func(*connectors.SourceVisibilityArtifact)
	}{
		{"healthy", func(*connectors.SourceVisibilityArtifact) {}},
		{"raw_encoding", func(a *connectors.SourceVisibilityArtifact) {
			*a = sourceEncoding174(t, *a, "")
			a.Payload = string(raw)
		}},
		{"unknown_encoding", func(a *connectors.SourceVisibilityArtifact) {
			*a = sourceEncoding174(t, *a, "unknown")
			a.Payload = string(raw)
		}},
		{"truncated", func(a *connectors.SourceVisibilityArtifact) { a.Payload = a.Payload[:len(a.Payload)-1] }},
		{"checksum", func(a *connectors.SourceVisibilityArtifact) {
			b := []byte(a.Payload)
			b[len(b)-8] ^= 1
			a.Payload = string(b)
		}},
		{"trailing_byte", func(a *connectors.SourceVisibilityArtifact) { a.Payload += "x" }},
		{"second_member", func(a *connectors.SourceVisibilityArtifact) { a.Payload += sourceGzip174(t, []byte("x")) }},
		{"decoded_short", func(a *connectors.SourceVisibilityArtifact) { a.Bytes++ }},
		{"decoded_long", func(a *connectors.SourceVisibilityArtifact) { a.Bytes-- }},
		{"over_budget", func(a *connectors.SourceVisibilityArtifact) { a.Bytes = connectors.SourceVisibilityConnectorLimit + 1 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			candidate := encoded
			tc.mutate(&candidate)
			got, err := connectors.DecodeSourceVisibility(t.Context(), candidate)
			if tc.name != "healthy" {
				if err == nil {
					t.Fatal("invalid representation accepted")
				}
				return
			}
			if err != nil {
				t.Fatalf("valid independently compressed selected payload refused: %v", err)
			}
			if !reflect.DeepEqual(got, expected) {
				t.Fatal("decoded operations/cells/facts differ from exact independent fixture")
			}
		})
	}
}
