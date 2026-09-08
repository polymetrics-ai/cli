package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/manifestindex"
	"polymetrics.ai/internal/connectors/manifeststore"
)

type malformedRateFS167 struct{ fs.FS }

func (f malformedRateFS167) Open(name string) (fs.File, error) {
	if name == "github/rate_limits.json" {
		return fstest.MapFS{name: &fstest.MapFile{Data: []byte(`{"state":}`)}}.Open(name)
	}
	return f.FS.Open(name)
}
func TestCLIPublicBundleDiagnostic167(t *testing.T) {
	for _, args := range [][]string{{"github", "label", "delete", "--json"}, {"connectors", "inspect", "github", "--json"}} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			loads := 0
			registry, err := connectors.NewLazyRegistry([]connectors.Metadata{{Name: "github", DisplayName: "GitHub", IntegrationType: "api"}}, func(context.Context, string) (connectors.Connector, error) {
				loads++
				_, failure := engine.Load(malformedRateFS167{defs.FS}, "github")
				return nil, fmt.Errorf("private-prefix-167: %w", errors.Join(failure, errors.New("private-sentinel-167")))
			})
			if err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			code := runWithPreflightRegistryAndApprovalReader(args, &stdout, &stderr, registry, nil)
			var response struct {
				Error struct {
					Category, Code, Message string
					Bundle                  map[string]any
				}
			}
			if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
				t.Fatalf("actual CLI JSON: %v %q", err, stdout.String())
			}
			if code != 1 || response.Error.Category != "internal" || response.Error.Code != "internal_error" || loads != 1 {
				t.Errorf("wrong selection classification/counters: %d %+v loads=%d", code, response, loads)
			}
			b := response.Error.Bundle
			if b["connector"] != "github" || b["generation"] != "embedded-v1" || b["file"] != "rate_limits.json" || b["field"] != "/" || b["reason_code"] != "invalid_json" || b["reason"] != "malformed JSON" {
				t.Errorf("wrong structured projection: %v", b)
			}
			if !strings.Contains(response.Error.Message, "malformed JSON") || stderr.String() != "error: "+response.Error.Message+"\n" || strings.Contains(stdout.String()+stderr.String(), "private-") {
				t.Errorf("public parity/secrecy: stdout=%q stderr=%q", stdout.String(), stderr.String())
			}
		})
	}
}

func TestCLILazyMalformedUnselected168(t *testing.T) {
	loads := map[string]int{}
	registry, err := connectors.NewLazyRegistry([]connectors.Metadata{{Name: "github", DisplayName: "GitHub", IntegrationType: "api"}, {Name: "gitlab", DisplayName: "GitLab", IntegrationType: "api"}}, func(_ context.Context, name string) (connectors.Connector, error) {
		loads[name]++
		b, e := engine.Load(malformedRateFS167{defs.FS}, name)
		if e != nil {
			return nil, e
		}
		return engine.New(b, nil), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := runWithPreflightRegistryAndApprovalReader([]string{"connectors", "list", "--json"}, &stdout, &stderr, registry, nil); code != 0 {
		t.Fatalf("list failed: %d %s", code, stderr.String())
	}
	var result struct {
		Kind       string
		Connectors []struct{ Name string }
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Kind != "ConnectorList" || len(result.Connectors) != 2 || result.Connectors[0].Name != "github" || result.Connectors[1].Name != "gitlab" || len(loads) != 0 {
		t.Fatalf("metadata listing decoded or lost identities: %+v loads=%v", result, loads)
	}
	stdout.Reset()
	stderr.Reset()
	if code := runWithPreflightRegistryAndApprovalReader([]string{"connectors", "inspect", "gitlab", "--json"}, &stdout, &stderr, registry, nil); code != 0 {
		t.Fatalf("healthy actual selection failed: %d %s", code, stderr.String())
	}
	var selected struct {
		Kind      string
		Connector struct{ Name string }
	}
	if err := json.Unmarshal(stdout.Bytes(), &selected); err != nil {
		t.Fatal(err)
	}
	if selected.Kind != "Connector" || selected.Connector.Name != "gitlab" || loads["gitlab"] != 1 || loads["github"] != 0 {
		t.Fatalf("healthy identity/selection changed: %+v loads=%v", selected, loads)
	}
}

func TestCLIBundlePreformattedAndReported168(t *testing.T) {
	_, original := engine.Load(malformedRateFS167{defs.FS}, "github")
	if original == nil {
		t.Fatal("malformed loader control succeeded")
	}
	wrapped := &cliError{category: categoryInternal, code: "internal_error", message: "private-prefmt-168: " + original.Error(), err: errors.Join(original, errors.New("private-companion-168"))}
	var stdout, stderr bytes.Buffer
	if code := writeError(&stdout, &stderr, wrapped, true); code != 1 || strings.Contains(stdout.String()+stderr.String(), "private-") || !strings.Contains(stdout.String(), "malformed JSON") {
		t.Fatalf("preformatted public boundary: code=%d out=%q err=%q", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	reported := alreadyReportedExecutionError(wrapped)
	if code := writeError(&stdout, &stderr, reported, true); code != 1 || stdout.Len() != 0 || stderr.Len() != 0 {
		t.Errorf("already-reported contract lost: code=%d out=%q err=%q", code, stdout.String(), stderr.String())
	}
	if !errors.Is(classifyError(wrapped), original) {
		t.Error("classification lost original graph")
	}
}

func TestCLISelectedStoreBundleDiagnostic168(t *testing.T) {
	for _, asJSON := range []bool{false, true} {
		t.Run(fmt.Sprintf("json=%t", asJSON), func(t *testing.T) {
			var selected manifestindex.Entry
			for _, entry := range manifestindex.GeneratedEntries() {
				if entry.Connector == "github" {
					selected = entry
					break
				}
			}
			if selected.Connector != "github" {
				t.Fatal("missing real generated GitHub entry")
			}
			selected.Generation = "selected-cli-generation-168"
			index, err := manifestindex.New([]manifestindex.Entry{selected}, 1)
			if err != nil {
				t.Fatal(err)
			}
			loads := 0
			companion := errors.New("private-companion-168")
			store, err := manifeststore.NewBundleStore(index, manifeststore.Limits{Entries: 1, Bytes: selected.Bytes}, func(_ context.Context, entry manifestindex.Entry) (manifeststore.LoadedBundle, error) {
				loads++
				_, failure := engine.Load(malformedRateFS167{defs.FS}, entry.Connector)
				if failure == nil {
					return manifeststore.LoadedBundle{}, errors.New("malformed fixture unexpectedly loaded")
				}
				return manifeststore.LoadedBundle{}, errors.Join(failure, companion)
			})
			if err != nil {
				t.Fatal(err)
			}
			var observed error
			registry, err := connectors.NewLazyRegistry([]connectors.Metadata{{Name: "github", DisplayName: "GitHub", IntegrationType: "api"}}, func(ctx context.Context, name string) (connectors.Connector, error) {
				handle, failure := store.Acquire(ctx, name)
				if failure != nil {
					observed = failure
					return nil, failure
				}
				defer handle.Release()
				return engine.New(*handle.Bundle(), nil), nil
			})
			if err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			args := []string{"connectors", "inspect", "github"}
			if asJSON {
				args = append(args, "--json")
			}
			code := runWithPreflightRegistryAndApprovalReader(args, &stdout, &stderr, registry, nil)
			var diagnostic *engine.BundleDiagnosticError
			var syntax *json.SyntaxError
			if code != 1 || loads != 1 || !errors.As(observed, &diagnostic) || !errors.As(observed, &syntax) || !errors.Is(observed, companion) {
				t.Fatalf("selected store graph/counter failed: code=%d loads=%d error=%v", code, loads, observed)
			}
			if diagnostic.Connector != selected.Connector || diagnostic.Generation != selected.Generation || diagnostic.Digest != selected.Digest || diagnostic.File != "rate_limits.json" || diagnostic.Field != "/" || diagnostic.ReasonCode != "invalid_json" {
				t.Fatalf("store identity mismatch: %+v", diagnostic)
			}
			if stderr.String() != "error: "+diagnostic.Error()+"\n" || strings.Contains(stdout.String()+stderr.String(), "private-") {
				t.Fatalf("unsafe or misbound text: stdout=%q stderr=%q", stdout.String(), stderr.String())
			}
			if asJSON {
				var response struct {
					Error struct {
						Code   string
						Bundle map[string]any
					}
				}
				if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
					t.Fatal(err)
				}
				b := response.Error.Bundle
				if response.Error.Code != "internal_error" || b["connector"] != selected.Connector || b["generation"] != selected.Generation || b["digest"] != selected.Digest || b["file"] != "rate_limits.json" || b["field"] != "/" || b["reason_code"] != "invalid_json" || b["reason"] != "malformed JSON" || len(b) != 7 {
					t.Fatalf("wrong selected public tuple: %+v", response)
				}
			} else if stdout.Len() != 0 {
				t.Fatalf("plain failure emitted successful output: %q", stdout.String())
			}
		})
	}
}
