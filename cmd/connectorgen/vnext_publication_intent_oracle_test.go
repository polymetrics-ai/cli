package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

// These controls call the same verdicts as the public cleanup/F03A rows.
// Producing Q deliberately is an oracle challenge, not a production corruption.
func TestCP11FinalPruneOracleRejectsWrongFixture(t *testing.T) {
	for _, wrong := range []bool{false, true} {
		name := "correct-N"
		if wrong {
			name = "coherent-Q-not-N"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			old, err := baseline.Publish(vNextPublicationArtifactsForTest("old", false))
			if err != nil {
				t.Fatal(err)
			}
			desired := vNextPublicationArtifactsForTest("desired-N", false)
			plan, _ := vNextPublicationExpectedPublish(t, root, vNextPublicationArtifactsForTest("old", false), desired, "committed")
			actual := desired
			if wrong {
				actual = vNextPublicationArtifactsForTest("coherent-Q", false)
			}
			prior, err := vNextPublicationExpectedStableAuthority(root)
			if err != nil {
				t.Fatal(err)
			}
			lease := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, old.Generation, vNextPublicationLeaseFile)
			var cut vNextPublicationExpectedTree
			fired := false
			writer, err := newVNextGenerationPublisher(root, "acme", plan.hooks(vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
				if point != vNextPublicationAfterGenerationLeaseIdentity || fired {
					return nil
				}
				fired = true
				if err := os.Rename(lease, lease+"-A"); err != nil {
					return err
				}
				if err := os.WriteFile(lease, []byte("actual lease B"), 0600); err != nil {
					return err
				}
				plan.leaseDelta(t, old.Generation, lease, lease+"-A")
				cut, err = vNextPublicationObserveExpectedTree(filepath.Join(root, "acme"))
				return err
			}}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = writer.Publish(actual)
			if !fired || err == nil || !strings.Contains(err.Error(), "identity") {
				t.Fatalf("actual final-prune cut fired=%t err=%v", fired, err)
			}
			vNextPublicationAssertExpectedTree(t, filepath.Join(root, "acme"), cut)
			_, verdict := vNextPublicationFinalPruneVerdict(root, prior, old, desired, plan)
			if wrong && (verdict == nil || !strings.Contains(verdict.Error(), "JOURNAL replacement differs from independent payload")) {
				t.Fatalf("wrong fixture must fail independent prepared JOURNAL check: %v", verdict)
			}
			if !wrong && verdict != nil {
				t.Fatalf("correct N rejected: %v", verdict)
			}
		})
	}
}

func TestCP11PreparedOracleRejectsWrongIntendedPayload(t *testing.T) {
	for _, wrong := range []bool{false, true} {
		name := "correct-prepared"
		if wrong {
			name = "well-formed-committed-not-prepared"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			old, err := baseline.Publish(vNextPublicationArtifactsForTest("old", false))
			if err != nil {
				t.Fatal(err)
			}
			desired, err := vNextPublicationJSON(vNextGenerationJournal{New: old, State: "prepared"})
			if err != nil {
				t.Fatal(err)
			}
			actual := vNextGenerationJournal{New: old, State: "prepared"}
			if wrong {
				actual.State = "committed"
			}

			plan := vNextPublicationNewExpectedPlan(t, root, map[string][]byte{vNextPublicationCurrentFile: vNextPublicationExpectedJSON(t, old), vNextPublicationJournalFile: nil}, vNextPublicationExpectedTransition{target: vNextPublicationJournalFile, intended: desired, phases: 0})
			known := vNextPublicationRepairTransactionsForTest(t, filepath.Join(root, "acme"))
			fault := errors.New("actual prepared Sync completed")
			fired := 0
			writer, err := newVNextGenerationPublisher(root, "acme", plan.hooks(vNextPublicationHooks{ControlRecord: vNextPublicationControlRecordHooks{Sync: func(file *os.File, label string) error {
				err := file.Sync()
				if label == "publication control repair prepared authority" {
					fired++
					return errors.Join(err, fault)
				}
				return err
			}}}))
			if err != nil {
				t.Fatal(err)
			}
			operation, err := writer.openOperation(context.Background(), unix.LOCK_EX, true)
			if err != nil {
				t.Fatal(err)
			}
			err = writer.writeJournalLocked(operation, actual)
			operation.close()
			if fired != 1 || !errors.Is(err, fault) {
				t.Fatalf("retained prepared cut fired=%d err=%v", fired, err)
			}
			verdict := vNextPublicationPendingPreparedVerdict(writer, filepath.Join(root, "acme"), vNextPublicationJournalFile, plan, known)
			if wrong && (verdict == nil || !strings.Contains(verdict.Error(), "differs from independent payload")) {
				t.Fatalf("wrong intended payload must fail independent anchor check: %v", verdict)
			}
			if !wrong && verdict != nil {
				t.Fatalf("correct intended payload rejected: %v", verdict)
			}
		})
	}
}

func TestCP11EarlyStageOracleRejectsChangedNonemptyA(t *testing.T) {
	vNextPublicationExerciseEarlyStageOracle(t, "changed-bytes")
}
func TestCP11EarlyStageOracleRejectsSameByteReplacementA(t *testing.T) {
	vNextPublicationExerciseEarlyStageOracle(t, "replacement-identity")
}
func vNextPublicationExerciseEarlyStageOracle(t *testing.T, mutation string) {
	t.Helper()
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	pointer, err := baseline.Publish(vNextPublicationArtifactsForTest("active", false))
	if err != nil {
		t.Fatal(err)
	}
	stageName := ".stage-owned"
	stage := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, stageName)
	vNextWriteOwnedStageForTest(t, stage, pointer, stageName, "original")
	wantA := vNextPublicationExpectedOwnedStage(t, stage, pointer, stageName, "original")
	plan := vNextPublicationNewExpectedPlan(t, root, map[string][]byte{vNextPublicationCurrentFile: vNextPublicationExpectedJSON(t, vNextPublicationFixturePointer(t, vNextPublicationArtifactsForTest("active", false))), vNextPublicationJournalFile: nil})
	prior, err := vNextPublicationExpectedStableAuthority(root)
	if err != nil {
		t.Fatal(err)
	}
	var cut vNextPublicationExpectedTree
	var moved string
	writer, err := newVNextGenerationPublisher(root, "acme", plan.hooks(vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationBeforeStageRemoval || moved != "" {
			return nil
		}
		moved = vNextReplaceOwnedStageForTest(t, stage, pointer, stageName, "replacement")
		plan.stageDelta(t, stage, moved, vNextPublicationExpectedOwnedStage(t, stage, pointer, stageName, "replacement"))
		cut, err = vNextPublicationCaptureExpectedCut(root, prior, plan)
		return err
	}}))
	if err != nil {
		t.Fatal(err)
	}
	err = writer.Recover(context.Background())
	if moved == "" || err == nil || !strings.Contains(err.Error(), "identity") {
		t.Fatalf("actual early-stage cut moved=%q err=%v", moved, err)
	}
	if err := vNextPublicationEarlyStageVerdict(root, moved, cut, wantA); err != nil {
		t.Fatalf("true renamed A rejected: %v", err)
	}
	if mutation == "changed-bytes" {
		if err := os.WriteFile(filepath.Join(moved, "sentinel.txt"), []byte("another nonempty value"), 0600); err != nil {
			t.Fatal(err)
		}
	} else {
		original := moved + "-actual-A"
		if err := os.Rename(moved, original); err != nil {
			t.Fatal(err)
		}
		vNextCopyDirectoryForTest(t, original, moved)
	}
	if err := vNextPublicationEarlyStageVerdict(root, moved, cut, wantA); err == nil || !strings.Contains(err.Error(), "identity/type/bytes changed") {
		t.Fatalf("displaced A %s must fail retained witness comparison: %v", mutation, err)
	}
}
