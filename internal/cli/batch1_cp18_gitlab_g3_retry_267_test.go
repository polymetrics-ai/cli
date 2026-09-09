package cli_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// TestBatch1CP18GitLabG3DeleteRetryDeclarations makes the source/policy
// boundary executable: a retained GitLab DELETE operation selects the
// source-admitted HTTP-idempotent retry policy. That is an intended-resource
// effect claim, not a provider idempotency-key or exactly-once claim.
func TestBatch1CP18GitLabG3DeleteRetryDeclarations(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "internal", "connectors", "defs", "gitlab", "sources", "gitlab-operation-source-lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(raw)
	if got := hex.EncodeToString(digest[:]); got != "de7010cc2a2088b8da33f50312db4d5129fd167b53ef1e028ccf72cae5c72b8d" {
		t.Fatalf("retained GitLab source capture sha256=%s", got)
	}
	var source struct {
		REST struct {
			Operations []struct {
				ID, Method, Path string
			} `json:"operations"`
		} `json:"rest"`
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	bundle, err := engine.Load(defs.FS, "gitlab")
	if err != nil {
		t.Fatal(err)
	}
	writes := make(map[string]engine.WriteAction, len(bundle.Writes))
	for _, write := range bundle.Writes {
		writes[write.Name] = write
	}
	var sourceDeletes, renderedDeletes int
	for _, operation := range source.REST.Operations {
		if strings.ToUpper(operation.Method) != "DELETE" {
			continue
		}
		sourceDeletes++
		name := "source_write_" + hex.EncodeToString([]byte("DELETE "+strings.TrimPrefix(operation.Path, "/api/v4")))
		write, ok := writes[name]
		if !ok {
			continue
		}
		renderedDeletes++
		if write.Kind != "delete" || write.Delete == nil {
			t.Fatalf("source DELETE %s rendered as %+v", operation.ID, write)
		}
		if !write.Delete.Idempotent {
			t.Fatalf("source DELETE %s does not select declared idempotent retry", operation.ID)
		}
		if write.IdempotencyKeyHeader != "" {
			t.Fatalf("source DELETE %s invents provider idempotency header %q", operation.ID, write.IdempotencyKeyHeader)
		}
	}
	if sourceDeletes != 212 {
		t.Fatalf("retained GitLab DELETE operations=%d, want 212", sourceDeletes)
	}
	if renderedDeletes != 82 {
		t.Fatalf("rendered GitLab DELETE actions=%d, want 82", renderedDeletes)
	}
	fixtureRaw, err := os.ReadFile(filepath.Join("testdata", "batch1-cp18-gitlab", "remaining252", "write-contracts.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Action                 string `json:"action"`
			Variant                string `json:"variant"`
			ExpectedAmbiguousSends int    `json:"expected_ambiguous_sends"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(fixtureRaw, &fixture); err != nil {
		t.Fatal(err)
	}
	var retryWitnesses int
	for _, tc := range fixture.Cases {
		if tc.ExpectedAmbiguousSends == 0 {
			continue
		}
		retryWitnesses++
		write, ok := writes[tc.Action]
		if !ok || tc.Variant != "selected_source_request_success" || write.Kind != "delete" || write.Delete == nil || !write.Delete.Idempotent || tc.ExpectedAmbiguousSends != 5 {
			t.Fatalf("unexpected retry witness action=%q variant=%q sends=%d", tc.Action, tc.Variant, tc.ExpectedAmbiguousSends)
		}
	}
	if retryWitnesses != 82 {
		t.Fatalf("idempotent DELETE retry witnesses=%d, want 82", retryWitnesses)
	}
}
