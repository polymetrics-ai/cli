package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestSourceFoundationAvailabilitySchema153(t *testing.T) {
	project, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	for _, fault := range []string{"current", "missing_document", "missing_selected_record", "missing_dependency", "stale_dependency"} {
		t.Run(fault, func(t *testing.T) {
			repo, baseline := sourceFoundationCommandFixture(t)
			if fault != "current" {
				sourceFoundationUnavailableFault150(t, repo, fault)
			}
			var out, diag bytes.Buffer
			if code := runSourceDemandsPolicy(t.Context(), []string{"source-demands", "--repo", repo}, &out, &diag, baseline); code != 0 {
				t.Fatal(diag.String())
			}
			candidate := filepath.Join(t.TempDir(), "register.json")
			if err := os.WriteFile(candidate, out.Bytes(), 0600); err != nil {
				t.Fatal(err)
			}
			command := exec.CommandContext(t.Context(), "python3", filepath.Join(project, ".planning/phases/cp13-foundation-demand-register/register-schema-check.py"), "--document", candidate)
			result, err := command.CombinedOutput()
			t.Logf("actual generated register published-schema comparison: %s", result)
			if err != nil {
				t.Fatalf("published schema parity: %v", err)
			}
		})
	}
}

func TestSourceFoundationOptionalInput153(t *testing.T) {
	for _, role := range []string{"owner", "test", "code", "fixture", "module", "sum", "capture", "output"} {
		for _, fault := range []string{"missing", "stale"} {
			t.Run(role+"/"+fault, func(t *testing.T) {
				repo, document := sourceFoundationProofObservationFixture(t)
				writeSourceFoundationObservationFixture(t, repo, document)
				control, err := readSourceFoundationProofBatch(t.Context(), repo, reviewedSourceFoundationProofs(), nil)
				if err != nil || len(control.records) != 2 {
					t.Fatalf("healthy original proof frontier: %v", err)
				}
				record := document.Records[0]
				name := map[string]string{"owner": record.OwnerSymbols[0].File, "test": record.Test.File, "module": "go.mod", "sum": "go.sum", "capture": record.Capture.Path, "output": record.Output.Path}[role]
				if role == "code" || role == "fixture" {
					for _, input := range document.Records[1].Inputs {
						if input.Role != role {
							continue
						}
						owner := false
						for _, o := range document.Records[1].OwnerSymbols {
							owner = owner || o.File == input.Path
						}
						if !owner {
							name = input.Path
							break
						}
					}
					if name == "" {
						t.Fatal("actual independent code/fixture dependency absent")
					}
				}
				if fault == "missing" {
					if err := os.Remove(filepath.Join(repo, name)); err != nil {
						t.Fatal(err)
					}
				} else {
					proofWrite(t, repo, name, []byte("safe changed historical input"))
				}
				got, err := readSourceFoundationProofBatch(t.Context(), repo, reviewedSourceFoundationProofs(), nil)
				if err != nil || len(got.records) != 2 {
					t.Fatalf("optional %s erased observations: %v", role, err)
				}
				for _, proof := range got.records {
					depends := proof.record.Capture.Path == name || proof.record.Output.Path == name
					for _, input := range proof.record.Inputs {
						depends = depends || input.Path == name
					}
					if !depends {
						if proof.status != "current" {
							t.Fatal("unrelated proof lost current observation")
						}
						continue
					}
					wantStatus, wantCode := "proof_stale", "proof_input_stale"
					if fault == "missing" {
						wantStatus, wantCode = "proof_unavailable", "proof_input_missing"
					}
					if proof.status != wantStatus || !reflect.DeepEqual(proof.issues, []sourceFoundationProofIssue{{Code: wantCode, Path: name}}) {
						t.Fatalf("typed issue mismatch: %s %+v", proof.status, proof.issues)
					}
				}
				found := false
				for _, pin := range got.pins {
					if pin.Path == name {
						found = true
						if pin.SHA256 != sourceBytesHash([]byte("safe changed historical input")) {
							t.Fatal("historical expected pin substituted for observed bytes")
						}
					}
				}
				if found != (fault == "stale") {
					t.Fatal("actual input pins fabricated missing bytes or omitted stale observation")
				}
			})
		}
	}
}

func TestSourceFoundationOptionalCheck153(t *testing.T) {
	for _, fault := range []string{"missing_document", "missing_selected_record", "missing_dependency", "stale_dependency"} {
		t.Run(fault, func(t *testing.T) {
			repo, baseline := sourceFoundationCommandFixture(t)
			sourceFoundationUnavailableFault150(t, repo, fault)
			var out, diag bytes.Buffer
			args := []string{"source-demands", "--repo", repo}
			if code := runSourceDemandsPolicy(t.Context(), args, &out, &diag, baseline); code != 0 {
				t.Fatalf("optional evidence generation: %s", diag.String())
			}
			original := append([]byte{}, out.Bytes()...)
			proofWrite(t, repo, "candidate.json", original)
			before, err := os.Stat(filepath.Join(repo, "candidate.json"))
			if err != nil {
				t.Fatal(err)
			}
			out.Reset()
			diag.Reset()
			if code := runSourceDemandsPolicy(t.Context(), append(args, "--check", "candidate.json"), &out, &diag, baseline); code != 0 || out.Len() != 0 {
				t.Fatalf("optional saved check failed or emitted bytes: %s", diag.String())
			}
			after, err := os.Stat(filepath.Join(repo, "candidate.json"))
			if err != nil || !os.SameFile(before, after) || before.Mode() != after.Mode() || before.ModTime() != after.ModTime() {
				t.Fatal("saved check changed identity/metadata")
			}
			raw, err := os.ReadFile(filepath.Join(repo, "candidate.json"))
			if err != nil || !bytes.Equal(raw, original) {
				t.Fatal("saved check changed complete bytes")
			}
		})
	}
}

func TestSourceFoundationCommandRoleOverlap153(t *testing.T) {
	for _, name := range []string{sourceLaneCohortPath, sourceLaneAnnotationsPath, sourceDemandAtlasPath} {
		t.Run(name, func(t *testing.T) {
			repo, baseline := sourceFoundationCommandFixture(t)
			var control, diag bytes.Buffer
			args := []string{"source-demands", "--repo", repo}
			if code := runSourceDemandsPolicy(t.Context(), args, &control, &diag, baseline); code != 0 {
				t.Fatal(diag.String())
			}
			raw, err := os.ReadFile(filepath.Join(repo, sourceFoundationProofPath))
			if err != nil {
				t.Fatal(err)
			}
			var document sourceFoundationProofDocument
			if err := json.Unmarshal(raw, &document); err != nil {
				t.Fatal(err)
			}
			record := &document.Records[0]
			inputRaw, err := os.ReadFile(filepath.Join(repo, name))
			if err != nil {
				t.Fatal(err)
			}
			pin := sourceFoundationProofInput{Path: name, SHA256: sourceBytesHash(inputRaw), Bytes: int64(len(inputRaw)), Role: "fixture"}
			record.Inputs = append(record.Inputs, pin)
			sort.Slice(record.Inputs, func(i, j int) bool { return record.Inputs[i].Path < record.Inputs[j].Path })
			captureRaw, err := os.ReadFile(filepath.Join(repo, record.Capture.Path))
			if err != nil {
				t.Fatal(err)
			}
			var capture map[string]any
			if err := json.Unmarshal(captureRaw, &capture); err != nil {
				t.Fatal(err)
			}
			// Deliberately altered test metadata exercises role overlap. Its
			// changed record has no matching review and must stay unreviewed.
			capture["inputs"].(map[string]any)[name] = map[string]any{"sha256": pin.SHA256, "bytes": pin.Bytes, "snapshot": "test-only-unreviewed-overlap"}
			captureRaw, err = json.Marshal(capture)
			if err != nil {
				t.Fatal(err)
			}
			proofWrite(t, repo, record.Capture.Path, captureRaw)
			record.Capture.SHA256, record.Capture.Bytes = sourceBytesHash(captureRaw), int64(len(captureRaw))
			writeSourceFoundationObservationFixture(t, repo, document)
			events := []sourceProofReadEvent{}
			var out bytes.Buffer
			if code := runSourceDemandsObserved(t.Context(), args, &out, &diag, baseline, func(e sourceProofReadEvent) {
				if e.Path == name {
					events = append(events, e)
				}
			}); code != 0 {
				t.Fatalf("actual CLI byte-role overlap refused: %s", diag.String())
			}
			var got, want sourceFoundationRegister
			if json.Unmarshal(out.Bytes(), &got) != nil || json.Unmarshal(control.Bytes(), &want) != nil {
				t.Fatal("full output invalid")
			}
			if !reflect.DeepEqual(got.Coverage, want.Coverage) || len(got.Requirements) != 1 || got.Requirements[0].Proofs[0].Status != "unreviewed" || len(got.Adopters) != 0 {
				t.Fatal("role overlap altered source or authorized modified proof")
			}
			if len(events) != 2 || events[0].Phase != "initial" || events[1].Phase != "final" || !events[0].Success || !events[1].Success || events[0].Bytes != int64(len(inputRaw)) || events[1].Bytes != int64(len(inputRaw)) {
				t.Fatalf("CLI overlap multiplied reads or lost bounded content: %+v", events)
			}
		})
	}
}

func TestSourceFoundationMissingCannotMaskInvalid153(t *testing.T) {
	repo, document := sourceFoundationProofObservationFixture(t)
	writeSourceFoundationObservationFixture(t, repo, document)
	if _, err := readSourceFoundationProofObservations(t.Context(), repo); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(repo, "go.sum")); err != nil {
		t.Fatal(err)
	}
	record := &document.Records[0]
	raw, err := os.ReadFile(filepath.Join(repo, record.Capture.Path))
	if err != nil {
		t.Fatal(err)
	}
	var capture map[string]any
	if err := json.Unmarshal(raw, &capture); err != nil {
		t.Fatal(err)
	}
	capture["command"] = []string{"sh", "-c", "untrusted"}
	raw, err = json.Marshal(capture)
	if err != nil {
		t.Fatal(err)
	}
	proofWrite(t, repo, record.Capture.Path, raw)
	record.Capture.SHA256, record.Capture.Bytes = sourceBytesHash(raw), int64(len(raw))
	writeSourceFoundationObservationFixture(t, repo, document)
	if got, err := readSourceFoundationProofObservations(t.Context(), repo); err == nil || len(got) != 0 {
		t.Fatal("missing dependency masked present matching-pin invalid command")
	}
}

func TestSourceFoundationPlannedContent153(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		t.Run(map[bool]string{false: "content_first", true: "hash_first"}[reverse], func(t *testing.T) {
			repo := t.TempDir()
			name := "internal/owner.go"
			raw := []byte("package fixture\n")
			proofWrite(t, repo, name, raw)
			root, err := os.OpenRoot(repo)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = root.Close() }()
			events := []sourceProofReadEvent{}
			cache := sourceProofFileCache{ctx: t.Context(), root: root, files: map[string]*sourceProofFile{}, limits: sourceProofFileLimits{UniqueBytes: 128, Files: 4}, afterRead: func(e sourceProofReadEvent) { events = append(events, e) }}
			if err := cache.plan(name, 64, true); err != nil {
				t.Fatal(err)
			}
			if err := cache.plan(name, int64(len(raw)), false); err != nil {
				t.Fatal(err)
			}
			if reverse {
				cache.get(name, 64, false)
			}
			for range 3 {
				got, file := cache.getContent(name, 64)
				if file.code != "" || !bytes.Equal(got, raw) {
					t.Fatal("planned hash/content overlap lost original bytes")
				}
			}
			cache.finalize()
			if len(events) != 2 || !events[0].Success || !events[1].Success || events[0].Phase != "initial" || events[1].Phase != "final" || cache.stats.ReadCalls != 2 || cache.stats.UniqueBytes != int64(len(raw)) || cache.stats.PhysicalCharged != 2*int64(len(raw)) {
				t.Fatalf("hits reread or changed accounting: %+v %+v", cache.stats, events)
			}
		})
	}
	t.Run("late_role_refused", func(t *testing.T) {
		repo := t.TempDir()
		proofWrite(t, repo, "input", []byte("abc"))
		root, err := os.OpenRoot(repo)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = root.Close() }()
		cache := sourceProofFileCache{ctx: t.Context(), root: root, files: map[string]*sourceProofFile{}, limits: sourceProofFileLimits{UniqueBytes: 16, Files: 2}}
		cache.get("input", 16, false)
		if err := cache.plan("input", 16, true); err == nil {
			t.Fatal("late content role allowed hidden reread")
		}
		if cache.stats.ReadCalls != 1 {
			t.Fatal("late role performed I/O")
		}
	})
	t.Run("minimum_cap_refused", func(t *testing.T) {
		repo := t.TempDir()
		proofWrite(t, repo, "input", []byte("abc"))
		root, err := os.OpenRoot(repo)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = root.Close() }()
		cache := sourceProofFileCache{ctx: t.Context(), root: root, files: map[string]*sourceProofFile{}, limits: sourceProofFileLimits{UniqueBytes: 16, Files: 2}}
		if err := cache.plan("input", 16, true); err != nil {
			t.Fatal(err)
		}
		if err := cache.plan("input", 2, false); err != nil {
			t.Fatal(err)
		}
		_, file := cache.getContent("input", 16)
		if file.code == "" || cache.stats.PhysicalCharged != 3 {
			t.Fatalf("larger role bypassed smaller cap: %+v %+v", file, cache.stats)
		}
	})
}

func TestSourceFoundationAbsentCustody153(t *testing.T) {
	for _, fault := range []string{"create", "symlink", "cancel"} {
		t.Run(fault, func(t *testing.T) {
			repo, document := sourceFoundationProofObservationFixture(t)
			writeSourceFoundationObservationFixture(t, repo, document)
			if _, err := readSourceFoundationProofObservations(t.Context(), repo); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(filepath.Join(repo, "go.sum")); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			reached := false
			got, err := readSourceFoundationProofsObserved(ctx, repo, reviewedSourceFoundationProofs(), func(e sourceProofReadEvent) {
				if reached || e.Phase != "final" || !e.Success {
					return
				}
				reached = true
				switch fault {
				case "create":
					proofWrite(t, repo, "go.sum", []byte("actor replacement"))
				case "symlink":
					if err := os.Symlink("go.mod", filepath.Join(repo, "go.sum")); err != nil {
						t.Fatal(err)
					}
				case "cancel":
					cancel()
				}
			})
			if !reached || err == nil || len(got) != 0 {
				t.Fatalf("absence changed without revoking batch: reached=%v err=%v", reached, err)
			}
			if fault == "cancel" && !errors.Is(err, context.Canceled) {
				t.Fatal("cancel cause lost")
			}
			if fault == "create" {
				raw, err := os.ReadFile(filepath.Join(repo, "go.sum"))
				if err != nil || string(raw) != "actor replacement" {
					t.Fatal("actor replacement not preserved")
				}
			}
		})
	}
}

func TestSourceFoundationOptionalCommandCustody153(t *testing.T) {
	for _, fault := range []string{"absent_created", "stale_replaced", "stale_changed", "stale_deleted"} {
		t.Run(fault, func(t *testing.T) {
			repo, baseline := sourceFoundationCommandFixture(t)
			name := filepath.Join(repo, "go.sum")
			if fault == "absent_created" {
				if err := os.Remove(name); err != nil {
					t.Fatal(err)
				}
			} else {
				proofWrite(t, repo, "go.sum", []byte("stable stale bytes"))
			}
			var out, diag bytes.Buffer
			args := []string{"source-demands", "--repo", repo}
			if code := runSourceDemandsPolicy(t.Context(), args, &out, &diag, baseline); code != 0 {
				t.Fatalf("stable optional observation control: %s", diag.String())
			}
			var before os.FileInfo
			if fault != "absent_created" {
				var err error
				before, err = os.Stat(name)
				if err != nil {
					t.Fatal(err)
				}
			}
			out.Reset()
			diag.Reset()
			fired := false
			code := runSourceDemandsObserved(t.Context(), args, &out, &diag, baseline, func(e sourceProofReadEvent) {
				if fired || e.Phase != "final" || !e.Success {
					return
				}
				fired = true
				switch fault {
				case "absent_created":
					proofWrite(t, repo, "go.sum", []byte("actor created"))
				case "stale_replaced":
					proofWrite(t, repo, "replacement", []byte("stable stale bytes"))
					if err := os.Rename(filepath.Join(repo, "replacement"), name); err != nil {
						t.Fatal(err)
					}
				case "stale_changed":
					if err := os.WriteFile(name, []byte("actor changed"), 0600); err != nil {
						t.Fatal(err)
					}
				case "stale_deleted":
					if err := os.Remove(name); err != nil {
						t.Fatal(err)
					}
				}
			})
			if !fired || code == 0 || out.Len() != 0 {
				t.Fatal("optional in-flight change exposed tentative stdout")
			}
			if fault == "stale_deleted" {
				if _, err := os.Stat(name); !errors.Is(err, os.ErrNotExist) {
					t.Fatal("actor deletion not retained")
				}
				return
			}
			after, err := os.Stat(name)
			if err != nil {
				t.Fatal(err)
			}
			if fault == "stale_replaced" && os.SameFile(before, after) {
				t.Fatal("same-byte replacement identity not reached")
			}
			if fault == "stale_changed" && !os.SameFile(before, after) {
				t.Fatal("same-inode content mutation not reached")
			}
			raw, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			want := map[string]string{"absent_created": "actor created", "stale_replaced": "stable stale bytes", "stale_changed": "actor changed"}[fault]
			if string(raw) != want {
				t.Fatal("actor bytes modified by command")
			}
		})
	}
}
