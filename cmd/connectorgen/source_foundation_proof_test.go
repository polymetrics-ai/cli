package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func sourceFoundationProofObservationFixture(t *testing.T) (string, sourceFoundationProofDocument) {
	t.Helper()
	repo, _ := sourceFoundationUniverseFixture(t)
	project, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(project, sourceFoundationProofPath))
	if err != nil {
		t.Fatal(err)
	}
	var document sourceFoundationProofDocument
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	// Later declaration tests reach real confined source files. These copies
	// come from the unchanged recorded test/owner inputs, not fabricated ASTs.
	paths := map[string]bool{}
	for _, record := range document.Records {
		paths[record.Test.File] = true
		for _, input := range record.Inputs {
			paths[input.Path] = true
		}
		paths[record.Capture.Path] = true
		paths[record.Output.Path] = true
		for _, owner := range record.OwnerSymbols {
			paths[owner.File] = true
		}
	}
	for name := range paths {
		contents, err := os.ReadFile(filepath.Join(project, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(filepath.Join(repo, name)), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(repo, name), contents, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return repo, document
}

func TestSourceFoundationProofOriginalExecution(t *testing.T) {
	for _, scenario := range []string{"missing dependency", "changed dependency", "failed command", "wrong output pin", "zero selected count", "changed command", "wrong environment", "skipped output", "duplicate selected pass", "parent only"} {
		t.Run(scenario, func(t *testing.T) {
			repo, document := sourceFoundationProofObservationFixture(t)
			writeSourceFoundationObservationFixture(t, repo, document)
			if control, err := readSourceFoundationProofObservations(t.Context(), repo); err != nil || len(control) != 2 {
				t.Fatalf("original complete capture/control frontier unavailable: %v", err)
			}
			r := &document.Records[0]
			if scenario == "parent only" {
				r = &document.Records[1]
			}
			captureName := filepath.Join(repo, r.Capture.Path)
			raw, err := os.ReadFile(captureName)
			if err != nil {
				t.Fatal(err)
			}
			var capture map[string]any
			if err := json.Unmarshal(raw, &capture); err != nil {
				t.Fatal(err)
			}
			switch scenario {
			case "missing dependency":
				if err := os.Remove(filepath.Join(repo, "go.sum")); err != nil {
					t.Fatal(err)
				}
			case "changed dependency":
				if err := os.WriteFile(filepath.Join(repo, "go.sum"), []byte("readable unrelated dependency"), 0600); err != nil {
					t.Fatal(err)
				}
			case "failed command":
				capture["exit_code"] = float64(1)
			case "wrong output pin":
				capture["output_sha256"] = sourceBytesHash([]byte("unrelated"))
			case "zero selected count":
				capture["selected_test_events"] = float64(0)
			case "changed command":
				capture["command"] = []any{"sh", "-c", "go test ./..."}
			case "wrong environment":
				capture["go_environment"].(map[string]any)["GOFLAGS"] = "-overlay=unreviewed.json"
			case "skipped output", "duplicate selected pass", "parent only":
				outputName := filepath.Join(repo, r.Output.Path)
				output, err := os.ReadFile(outputName)
				if err != nil {
					t.Fatal(err)
				}
				var rewritten []byte
				for _, line := range bytes.Split(bytes.TrimSpace(output), []byte{'\n'}) {
					var event map[string]any
					if err := json.Unmarshal(line, &event); err != nil {
						t.Fatal(err)
					}
					selected := event["Test"] == r.Test.Selected
					if scenario == "parent only" && selected {
						continue
					}
					if scenario == "skipped output" && selected && event["Action"] == "pass" {
						event["Action"] = "skip"
					}
					encoded, err := json.Marshal(event)
					if err != nil {
						t.Fatal(err)
					}
					rewritten = append(append(rewritten, encoded...), '\n')
					if scenario == "duplicate selected pass" && selected && event["Action"] == "pass" {
						rewritten = append(append(rewritten, encoded...), '\n')
					}
				}
				if err := os.WriteFile(outputName, rewritten, 0600); err != nil {
					t.Fatal(err)
				}
				r.Output.SHA256, r.Output.Bytes = sourceBytesHash(rewritten), int64(len(rewritten))
				capture["output_sha256"], capture["output_bytes"] = r.Output.SHA256, r.Output.Bytes
				// Counts are made internally consistent deliberately: the real
				// parser must reject events rather than an unrelated stale count.
				parsed := parseSourceProofResult(rewritten)
				runs, passes := 0, 0
				for _, result := range parsed.tests {
					runs, passes = runs+result.runs, passes+result.passes
				}
				capture["selected_test_events"], capture["passing_test_events"] = runs, passes
			}
			raw, err = json.Marshal(capture)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(captureName, raw, 0600); err != nil {
				t.Fatal(err)
			}
			r.Capture.SHA256, r.Capture.Bytes = sourceBytesHash(raw), int64(len(raw))
			writeSourceFoundationObservationFixture(t, repo, document)
			if got, err := readSourceFoundationProofObservations(t.Context(), repo); err == nil || len(got) != 0 {
				t.Fatalf("invalid execution/input survived real proof reader: observations=%d error=%v", len(got), err)
			}
		})
	}
}

func writeSourceFoundationObservationFixture(t *testing.T, repo string, document sourceFoundationProofDocument) {
	t.Helper()
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, sourceFoundationProofPath), raw, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestSourceFoundationProofObservation(t *testing.T) {
	repo, document := sourceFoundationProofObservationFixture(t)
	writeSourceFoundationObservationFixture(t, repo, document)
	got, err := readSourceFoundationProofObservations(t.Context(), repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].record.ID != "transport.sync-contract.v1.mode-vocabulary" || got[1].record.ID != "runtime.direct-execution.v1.reject-undeclared-check-status" {
		t.Fatal("real Atlas/record projection lost the independently named observations")
	}
	if got[0].record.Test.Selected != "TestModeVocabularyIsClosed" || got[1].record.Test.Selected != "TestCheck_ExactSuccessStatusesPreserveDeclaredOutcome/bad_rejects_undeclared_200" {
		t.Fatal("real top-level/child selector changed")
	}
	first, err := sourceFoundationProofRecordHash(got[0].record)
	if err != nil {
		t.Fatal(err)
	}
	changed := got[0].record
	changed.Assertion.Statement = "unrelated assertion"
	second, err := sourceFoundationProofRecordHash(changed)
	if err != nil || first == second {
		t.Fatal("record digest omitted assertion meaning")
	}
}

func TestSourceFoundationProofClosedObservations(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*sourceFoundationProofDocument)
	}{
		{"duplicate id first", func(d *sourceFoundationProofDocument) { d.Records[1].ID = d.Records[0].ID }},
		{"duplicate id reversed", func(d *sourceFoundationProofDocument) { d.Records[0].ID = d.Records[1].ID }},
		{"duplicate contract selection", func(d *sourceFoundationProofDocument) {
			other := d.Records[0]
			other.ID = "other-assertion"
			other.Assertion.Statement = "conflicting assertion"
			d.Records = append(d.Records, other)
		}},
		{"duplicate contract reversed", func(d *sourceFoundationProofDocument) {
			other := d.Records[0]
			other.ID = "other-assertion"
			other.Assertion.Statement = "conflicting assertion"
			d.Records = append([]sourceFoundationProofRecord{other}, d.Records...)
		}},
		{"empty record id", func(d *sourceFoundationProofDocument) { d.Records[0].ID = "" }},
		{"missing owners", func(d *sourceFoundationProofDocument) { d.Records[0].OwnerSymbols = nil }},
		{"duplicate owners", func(d *sourceFoundationProofDocument) {
			d.Records[0].OwnerSymbols = append(d.Records[0].OwnerSymbols, d.Records[0].OwnerSymbols[0])
		}},
		{"missing inputs", func(d *sourceFoundationProofDocument) { d.Records[0].Inputs = nil }},
		{"duplicate input", func(d *sourceFoundationProofDocument) {
			d.Records[0].Inputs = append(d.Records[0].Inputs, d.Records[0].Inputs[0])
		}},
		{"wrong input role", func(d *sourceFoundationProofDocument) { d.Records[0].Inputs[0].Role = "executable" }},
		{"missing limits", func(d *sourceFoundationProofDocument) { d.Records[0].Limitations = nil }},
		{"wrong scope", func(d *sourceFoundationProofDocument) { d.Records[0].Scope = "connector_hermetic" }},
		{"wrong execution class", func(d *sourceFoundationProofDocument) { d.Records[0].ExecutionClass = "C1" }},
		{"invalid digest", func(d *sourceFoundationProofDocument) { d.Records[0].Capture.SHA256 = "invalid" }},
		{"foreign receipt", func(d *sourceFoundationProofDocument) { d.Records[0].Capture.Path = "go.mod" }},
		{"invalid assertion span", func(d *sourceFoundationProofDocument) {
			d.Records[0].Assertion.EndLine = d.Records[0].Assertion.StartLine - 1
		}},
		{"unknown atlas control", func(d *sourceFoundationProofDocument) { d.Records[0].AtlasID = "unknown.v1" }},
		{"stale atlas control", func(d *sourceFoundationProofDocument) { d.Atlas.Bytes++ }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo, document := sourceFoundationProofObservationFixture(t)
			// Each malformed candidate starts from an actually read, valid
			// Atlas/record pair. No missing command or synthetic error baseline.
			writeSourceFoundationObservationFixture(t, repo, document)
			if control, err := readSourceFoundationProofObservations(t.Context(), repo); err != nil || len(control) != 2 {
				t.Fatalf("positive observation frontier not reached: %v", err)
			}
			tc.mutate(&document)
			writeSourceFoundationObservationFixture(t, repo, document)
			if got, err := readSourceFoundationProofObservations(t.Context(), repo); err == nil || len(got) != 0 {
				t.Fatalf("malformed foundation record survived actual observation frontier: observations=%d error=%v", len(got), err)
			}
		})
	}
}

func TestSourceFoundationProofAtlasBindings(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*sourceFoundationProofRecord)
	}{
		{"unknown owner", func(r *sourceFoundationProofRecord) { r.OwnerSymbols[0].Name = "NonexistentModeOwner" }},
		{"owner file mismatch", func(r *sourceFoundationProofRecord) { r.OwnerSymbols[0].File = "internal/synccontract/contract.go" }},
		{"unknown test selector", func(r *sourceFoundationProofRecord) { r.Test.Symbol, r.Test.Selected = "TestUnknown", "TestUnknown" }},
		{"test package mismatch", func(r *sourceFoundationProofRecord) { r.Test.Package = "./internal/connectors/engine" }},
		{"test file mismatch", func(r *sourceFoundationProofRecord) {
			r.Test.File = "internal/synccontract/other_test.go"
			r.Inputs = append(r.Inputs, sourceFoundationProofInput{Path: r.Test.File, SHA256: r.Assertion.SourceSHA256, Role: "test"})
		}},
		{"contract value mismatch", func(r *sourceFoundationProofRecord) { r.Contract.ValueSHA256 = sourceBytesHash([]byte(`"unrelated"`)) }},
		{"missing contract occurrence", func(r *sourceFoundationProofRecord) { r.Contract.Pointer = "/supported_contracts/guarantees/99999" }},
		{"wrong occurrence", func(r *sourceFoundationProofRecord) { r.Contract.Pointer = "/supported_contracts/guarantees/0" }},
		{"borrow another foundation", func(r *sourceFoundationProofRecord) { r.AtlasID = "runtime.direct-execution.v1" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo, document := sourceFoundationProofObservationFixture(t)
			writeSourceFoundationObservationFixture(t, repo, document)
			if control, err := readSourceFoundationProofObservations(t.Context(), repo); err != nil || len(control) != 2 {
				t.Fatalf("valid source/Atlas observation unavailable: %v", err)
			}
			tc.mutate(&document.Records[0])
			sort.Slice(document.Records[0].Inputs, func(i, j int) bool {
				return document.Records[0].Inputs[i].Path < document.Records[0].Inputs[j].Path
			})
			if !sourceFoundationProofRecordShape(document.Records[0]) {
				t.Fatal("relationship counterexample fails earlier record shape")
			}
			writeSourceFoundationObservationFixture(t, repo, document)
			if got, err := readSourceFoundationProofObservations(t.Context(), repo); err == nil || len(got) != 0 {
				t.Fatalf("incorrect owner/test/contract relationship survived: observations=%d error=%v", len(got), err)
			}
		})
	}
}

func TestSourceFoundationProofReviewedAssertions(t *testing.T) {
	for _, scenario := range []string{"original reviewed assertions", "forged assertion scope"} {
		t.Run(scenario, func(t *testing.T) {
			repo, document := sourceFoundationProofObservationFixture(t)
			want := "current"
			if scenario == "forged assertion scope" {
				// Execution, capture and input hashes still agree. Only semantic
				// review can refuse the forged meaning of a genuine passing test.
				document.Records[0].Assertion.Statement = "Every provider receiver is implemented and proven"
				want = "unreviewed"
			}
			writeSourceFoundationObservationFixture(t, repo, document)
			got, err := readSourceFoundationProofObservations(t.Context(), repo)
			if err != nil || len(got) != 2 {
				t.Fatalf("real capture/result/input frontier not reached: %v", err)
			}
			if got[0].status != want {
				t.Fatalf("first assertion status = %q, want %q after actual full evidence read", got[0].status, want)
			}
			if scenario == "original reviewed assertions" && got[1].status != "current" {
				t.Fatalf("actual reviewed child assertion status = %q, want current", got[1].status)
			}
		})
	}
}

func TestSourceFoundationProofRecordCapacity(t *testing.T) {
	repo, document := sourceFoundationProofObservationFixture(t)
	writeSourceFoundationObservationFixture(t, repo, document)
	if control, err := readSourceFoundationProofObservations(t.Context(), repo); err != nil || len(control) != 2 {
		t.Fatalf("two reviewed record control unavailable: %v", err)
	}
	// The third value is deliberately malformed. Capacity must be refused
	// before decoding/allocating that out-of-policy record, not afterward.
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatal(err)
	}
	object["records"] = append(bytes.TrimSuffix(object["records"], []byte("]")), []byte(`,{"out_of_capacity":true}]`)...)
	raw, err = json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, sourceFoundationProofPath), raw, 0600); err != nil {
		t.Fatal(err)
	}
	if got, err := readSourceFoundationProofObservations(t.Context(), repo); err == nil || len(got) != 0 || !strings.Contains(err.Error(), "record capacity") {
		t.Fatalf("out-of-policy entry reached decoding instead of capacity frontier: observations=%d error=%v", len(got), err)
	}
}

func TestSourceFoundationProofNullCaptureFields(t *testing.T) {
	for _, field := range []string{"exit_code", "wall_seconds", "started_epoch", "tracked_status_before", "GOFLAGS", "GOEXPERIMENT", "GOWORK"} {
		t.Run(field, func(t *testing.T) {
			repo, document := sourceFoundationProofObservationFixture(t)
			writeSourceFoundationObservationFixture(t, repo, document)
			if control, err := readSourceFoundationProofObservations(t.Context(), repo); err != nil || len(control) != 2 || control[0].status != "current" {
				t.Fatalf("original reviewed capture frontier unavailable: %v", err)
			}
			r := &document.Records[0]
			name := filepath.Join(repo, r.Capture.Path)
			raw, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			var capture map[string]json.RawMessage
			if err := json.Unmarshal(raw, &capture); err != nil {
				t.Fatal(err)
			}
			if strings.HasPrefix(field, "GO") {
				var environment map[string]json.RawMessage
				if err := json.Unmarshal(capture["go_environment"], &environment); err != nil {
					t.Fatal(err)
				}
				environment[field] = json.RawMessage("null")
				capture["go_environment"], err = json.Marshal(environment)
				if err != nil {
					t.Fatal(err)
				}
			} else {
				capture[field] = json.RawMessage("null")
			}
			raw, err = json.Marshal(capture)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(name, raw, 0600); err != nil {
				t.Fatal(err)
			}
			r.Capture.SHA256, r.Capture.Bytes = sourceBytesHash(raw), int64(len(raw))
			writeSourceFoundationObservationFixture(t, repo, document)
			if got, err := readSourceFoundationProofObservations(t.Context(), repo); err == nil || len(got) != 0 {
				t.Fatalf("required capture null became a legitimate zero value: observations=%d error=%v", len(got), err)
			}
		})
	}
}

// Closed wire grammar must reject omitted/null required members before their
// Go zero values lose the distinction. Use the actual authored document.
func TestSourceFoundationProofRequiredWireMembers(t *testing.T) {
	project, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(filepath.Join(project, sourceFoundationProofPath))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decodeSourceFoundationProofDocument(t.Context(), original, 2); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"atlas/bytes", "records/0/capture/bytes", "records/0/output/bytes", "records/0/inputs/0/bytes", "records/0/test/selected", "records/0/assertion/start_line", "records/0/owner_symbols/0/name"} {
		for _, mode := range []string{"missing", "null"} {
			t.Run(path+"/"+mode, func(t *testing.T) {
				var value any
				if err := json.Unmarshal(original, &value); err != nil {
					t.Fatal(err)
				}
				parts := strings.Split(path, "/")
				node := value
				for _, part := range parts[:len(parts)-1] {
					switch object := node.(type) {
					case map[string]any:
						node = object[part]
					case []any:
						node = object[0] // Every array selection in this literal fixture is zero.
					default:
						t.Fatal("fixture path does not resolve")
					}
				}
				object, ok := node.(map[string]any)
				if !ok {
					t.Fatal("fixture target not object")
				}
				field := parts[len(parts)-1]
				if mode == "missing" {
					delete(object, field)
				} else {
					object[field] = nil
				}
				raw, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := decodeSourceFoundationProofDocument(t.Context(), raw, 2); err == nil {
					t.Fatal("required wire member lost through zero-value normalization")
				}
			})
		}
	}
}

func TestSourceFoundationProofFileCustody(t *testing.T) {
	for _, target := range []string{"capture", "input", "output"} {
		for _, fault := range []string{"replacement inode", "changed content", "cancel"} {
			t.Run(target+"/"+fault, func(t *testing.T) {
				repo, document := sourceFoundationProofObservationFixture(t)
				writeSourceFoundationObservationFixture(t, repo, document)
				name := "go.sum"
				if target == "capture" {
					name = document.Records[0].Capture.Path
				}
				if target == "output" {
					name = document.Records[0].Output.Path
				}
				full := filepath.Join(repo, name)
				original, err := os.ReadFile(full)
				if err != nil {
					t.Fatal(err)
				}
				before, err := os.Stat(full)
				if err != nil {
					t.Fatal(err)
				}
				if target == "capture" && len(original) <= 4<<20 {
					t.Fatal("original metadata no longer proves larger-than-4MiB capture")
				}
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				fired := false
				var actor os.FileInfo
				actorBytes := original
				got, readErr := readSourceFoundationProofsObserved(ctx, repo, reviewedSourceFoundationProofs(), func(event sourceProofReadEvent) {
					if event.Path != name || event.Phase != "initial" || fired {
						return
					}
					if !event.Success || event.Bytes != int64(len(original)) || !os.SameFile(before, event.Before) || !os.SameFile(before, event.After) {
						t.Fatal("fault was not after witnessed original read")
					}
					fired = true
					switch fault {
					case "replacement inode":
						replacement, err := os.CreateTemp(filepath.Dir(full), "actor-")
						if err != nil {
							t.Fatal(err)
						}
						if _, err := replacement.Write(original); err != nil {
							_ = replacement.Close()
							t.Fatal(err)
						}
						if err := replacement.Close(); err != nil {
							t.Fatal(err)
						}
						if err := os.Rename(replacement.Name(), full); err != nil {
							t.Fatal(err)
						}
					case "changed content":
						actorBytes = append(append([]byte{}, original...), '\n')
						if err := os.WriteFile(full, actorBytes, 0600); err != nil {
							t.Fatal(err)
						}
					case "cancel":
						cancel()
					}
					actor, err = os.Stat(full)
					if err != nil {
						t.Fatal(err)
					}
					if fault == "replacement inode" && os.SameFile(before, actor) {
						t.Fatal("replacement retained original identity")
					}
					if fault != "replacement inode" && !os.SameFile(before, actor) {
						t.Fatal("content/cancel case unexpectedly replaced original inode")
					}
				})
				if !fired {
					t.Fatal("actual initial-read frontier not reached")
				}
				if readErr == nil || len(got) != 0 {
					t.Fatalf("fault leaked observations: %d, %v", len(got), readErr)
				}
				after, err := os.Stat(full)
				if err != nil {
					t.Fatal(err)
				}
				afterBytes, err := os.ReadFile(full)
				if err != nil || !os.SameFile(actor, after) || !bytes.Equal(actorBytes, afterBytes) {
					t.Fatalf("reader changed actor-owned outcome: %v", err)
				}
				if fault == "cancel" && !errors.Is(readErr, context.Canceled) {
					t.Fatalf("actual cancellation cause lost: %v", readErr)
				}
			})
		}
	}
}

func TestSourceFoundationProofBoundedMetadata(t *testing.T) {
	for _, extra := range []int64{0, 1} {
		name := "exact 64MiB"
		if extra != 0 {
			name = "over 64MiB"
		}
		t.Run(name, func(t *testing.T) {
			repo, document := sourceFoundationProofObservationFixture(t)
			r := &document.Records[0]
			full := filepath.Join(repo, r.Capture.Path)
			original, err := os.ReadFile(full)
			if err != nil {
				t.Fatal(err)
			}
			raw := make([]byte, (64<<20)+extra)
			copy(raw, original)
			for i := len(original); i < len(raw); i++ {
				raw[i] = ' '
			}
			if err := os.WriteFile(full, raw, 0600); err != nil {
				t.Fatal(err)
			}
			r.Capture.SHA256, r.Capture.Bytes = sourceBytesHash(raw), int64(len(raw))
			writeSourceFoundationObservationFixture(t, repo, document)
			initial, final := 0, 0
			var charge int64
			got, readErr := readSourceFoundationProofsObserved(t.Context(), repo, reviewedSourceFoundationProofs(), func(event sourceProofReadEvent) {
				if event.Path != r.Capture.Path {
					return
				}
				if event.Phase == "initial" {
					initial++
					charge = event.Charged
					if event.Success != (extra == 0) {
						t.Fatal("unexpected actual bounded-read outcome")
					}
				} else {
					final++
				}
			})
			if initial != 1 {
				t.Fatalf("initial physical reads = %d", initial)
			}
			if extra == 0 {
				if readErr != nil || len(got) != 2 || got[0].status != "unreviewed" || final != 1 || charge != 64<<20 {
					t.Fatalf("exact typed metadata bound: observations=%d final=%d charge=%d error=%v", len(got), final, charge, readErr)
				}
			} else if readErr == nil || len(got) != 0 || final != 0 || charge != (64<<20)+1 {
				t.Fatalf("failed attempted read uncharged or accepted: observations=%d final=%d charge=%d error=%v", len(got), final, charge, readErr)
			}
		})
	}
}

func TestSourceFoundationProofDuplicateTupleWithinCapacity(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		name := "forward"
		if reverse {
			name = "reverse"
		}
		t.Run(name, func(t *testing.T) {
			repo, document := sourceFoundationProofObservationFixture(t)
			copyRecord := document.Records[0]
			copyRecord.ID += ".duplicate"
			document.Records = []sourceFoundationProofRecord{document.Records[0], copyRecord}
			if reverse {
				document.Records[0], document.Records[1] = document.Records[1], document.Records[0]
			}
			writeSourceFoundationObservationFixture(t, repo, document)
			got, err := readSourceFoundationProofObservations(t.Context(), repo)
			if err == nil || len(got) != 0 || !strings.Contains(err.Error(), "record invalid or duplicate") {
				t.Fatalf("two distinct IDs sharing tuple did not reach tuple refusal: observations=%d error=%v", len(got), err)
			}
		})
	}
}
