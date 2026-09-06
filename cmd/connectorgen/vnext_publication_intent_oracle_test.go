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
			writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
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
				cut, err = vNextPublicationCaptureExpectedCut(root, prior)
				return err
			}})
			if err != nil {
				t.Fatal(err)
			}
			_, err = writer.Publish(actual)
			if !fired || err == nil || !strings.Contains(err.Error(), "identity") {
				t.Fatalf("actual final-prune cut fired=%t err=%v", fired, err)
			}
			vNextPublicationAssertExpectedTree(t, filepath.Join(root, "acme"), cut)
			_, verdict := vNextPublicationFinalPruneVerdict(root, prior, old, desired)
			if wrong && verdict == nil {
				t.Fatal("oracle accepted coherent Q although fixture intended N")
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
			known := vNextPublicationRepairTransactionsForTest(t, filepath.Join(root, "acme"))
			fault := errors.New("actual prepared Sync completed")
			fired := 0
			writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{ControlRecord: vNextPublicationControlRecordHooks{Sync: func(file *os.File, label string) error {
				err := file.Sync()
				if label == "publication control repair prepared authority" {
					fired++
					return errors.Join(err, fault)
				}
				return err
			}}})
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
			verdict := vNextPublicationPendingPreparedVerdict(writer, filepath.Join(root, "acme"), vNextPublicationJournalFile, false, true, known, desired)
			if wrong && verdict == nil {
				t.Fatal("oracle accepted wrong well-formed intended anchor payload")
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
	wantA, err := vNextPublicationObserveExpectedTree(stage)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := vNextPublicationExpectedStableAuthority(root)
	if err != nil {
		t.Fatal(err)
	}
	var cut vNextPublicationExpectedTree
	var moved string
	writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationBeforeStageRemoval || moved != "" {
			return nil
		}
		moved = vNextReplaceOwnedStageForTest(t, stage, pointer, stageName, "replacement")
		cut, err = vNextPublicationCaptureExpectedCut(root, prior)
		return err
	}})
	if err != nil {
		t.Fatal(err)
	}
	err = writer.Recover(context.Background())
	if moved == "" || err == nil || !strings.Contains(err.Error(), "identity") {
		t.Fatalf("actual early-stage cut moved=%q err=%v", moved, err)
	}
	if err := vNextPublicationEarlyStageVerdict(root, moved, stage, cut, wantA); err != nil {
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
	if err := vNextPublicationEarlyStageVerdict(root, moved, stage, cut, wantA); err == nil {
		t.Fatalf("oracle accepted displaced A %s", mutation)
	}
}
