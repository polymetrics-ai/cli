package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"golang.org/x/sys/unix"
)

// vNextPublicationPathWitness binds a test assertion to an actual pathname
// identity and payload. It deliberately records both: equal bytes do not make
// a substitute object the same authority member.
type vNextPublicationPathWitness struct {
	info    os.FileInfo
	payload []byte
}

func vNextPublicationFileWitnessForTest(t *testing.T, path string) vNextPublicationPathWitness {
	t.Helper()
	parent, err := vNextPublicationOpenDirectory(filepath.Dir(path), "witness parent")
	if err != nil {
		t.Fatalf("open witness parent for %q: %v", path, err)
	}
	file, err := parent.openRegular(filepath.Base(path), "file witness", unix.O_RDONLY)
	if err != nil {
		_ = parent.Close()
		t.Fatalf("open file witness %q: %v", path, err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		_ = parent.Close()
		t.Fatalf("stat file witness %q: %v", path, err)
	}
	if vNextPublicationWitnessAfterOpenForTest != nil {
		vNextPublicationWitnessAfterOpenForTest("file", path, info)
	}
	payload, readErr := io.ReadAll(file)
	closeFileErr := file.Close()
	closeParentErr := parent.Close()
	if readErr != nil {
		t.Fatalf("read witness %q: %v", path, readErr)
	}
	if closeFileErr != nil {
		t.Fatalf("close file witness %q: %v", path, closeFileErr)
	}
	if closeParentErr != nil {
		t.Fatalf("close witness parent for %q: %v", path, closeParentErr)
	}
	return vNextPublicationPathWitness{info: info, payload: payload}
}

func vNextPublicationDirectoryWitnessForTest(t *testing.T, path string) vNextPublicationPathWitness {
	t.Helper()
	directory, err := vNextPublicationOpenDirectory(path, "directory witness")
	if err != nil {
		t.Fatalf("open directory witness %q: %v", path, err)
	}
	info, err := directory.file.Stat()
	if err != nil {
		_ = directory.Close()
		t.Fatalf("stat directory witness %q: %v", path, err)
	}
	if vNextPublicationWitnessAfterOpenForTest != nil {
		vNextPublicationWitnessAfterOpenForTest("directory", path, info)
	}
	payload, readErr := directory.readFile("metadata.json", "directory witness metadata")
	closeErr := directory.Close()
	if readErr != nil {
		t.Fatalf("read directory witness metadata %q: %v", path, readErr)
	}
	if closeErr != nil {
		t.Fatalf("close directory witness %q: %v", path, closeErr)
	}
	return vNextPublicationPathWitness{info: info, payload: payload}
}

func vNextPublicationAssertWitnessForTest(t *testing.T, label, path string, want vNextPublicationPathWitness) {
	t.Helper()
	got := vNextPublicationFileWitnessForTest(t, path)
	if got.info.Mode().Type() != want.info.Mode().Type() || !os.SameFile(got.info, want.info) {
		t.Fatalf("%s identity/type changed: got mode=%v dev/inode=%v/%v, want mode=%v dev/inode=%v/%v", label, got.info.Mode(), vNextPublicationInfoDeviceForTest(got.info), vNextPublicationInfoInodeForTest(got.info), want.info.Mode(), vNextPublicationInfoDeviceForTest(want.info), vNextPublicationInfoInodeForTest(want.info))
	}
	if !bytes.Equal(got.payload, want.payload) {
		t.Fatalf("%s bytes changed: got=%q want=%q", label, got.payload, want.payload)
	}
}

func vNextPublicationAssertDirectoryWitnessForTest(t *testing.T, label, path string, want vNextPublicationPathWitness) {
	t.Helper()
	got := vNextPublicationDirectoryWitnessForTest(t, path)
	if got.info.Mode().Type() != want.info.Mode().Type() || !os.SameFile(got.info, want.info) {
		t.Fatalf("%s identity/type changed: got mode=%v dev/inode=%v/%v, want mode=%v dev/inode=%v/%v", label, got.info.Mode(), vNextPublicationInfoDeviceForTest(got.info), vNextPublicationInfoInodeForTest(got.info), want.info.Mode(), vNextPublicationInfoDeviceForTest(want.info), vNextPublicationInfoInodeForTest(want.info))
	}
	if !bytes.Equal(got.payload, want.payload) {
		t.Fatalf("%s bytes changed: got=%q want=%q", label, got.payload, want.payload)
	}
}

func vNextPublicationAssertDistinctWitnessForTest(t *testing.T, label string, first, second vNextPublicationPathWitness) {
	t.Helper()
	if first.info.Mode().Type() != second.info.Mode().Type() {
		t.Fatalf("%s types differ: first=%v second=%v", label, first.info.Mode(), second.info.Mode())
	}
	if os.SameFile(first.info, second.info) {
		t.Fatalf("%s has the same dev/inode %v/%v, want distinct objects", label, vNextPublicationInfoDeviceForTest(first.info), vNextPublicationInfoInodeForTest(first.info))
	}
}

func vNextPublicationInfoDeviceForTest(info os.FileInfo) uint64 {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0
	}
	return uint64(stat.Dev)
}

func vNextPublicationInfoInodeForTest(info os.FileInfo) uint64 {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0
	}
	return uint64(stat.Ino)
}

func vNextPublicationReadCurrentJournalForTest(t *testing.T, root string) (vNextGenerationPointer, vNextGenerationJournal, bool, []byte, []byte) {
	t.Helper()
	connectorRoot := filepath.Join(root, "acme")
	currentBytes, err := os.ReadFile(filepath.Join(connectorRoot, vNextPublicationCurrentFile))
	if err != nil {
		t.Fatalf("read CURRENT: %v", err)
	}
	var current vNextGenerationPointer
	if err := vNextPublicationDecode(currentBytes, &current); err != nil {
		t.Fatalf("decode CURRENT: %v", err)
	}
	journalBytes, err := os.ReadFile(filepath.Join(connectorRoot, vNextPublicationJournalFile))
	if errors.Is(err, fs.ErrNotExist) {
		return current, vNextGenerationJournal{}, false, currentBytes, nil
	}
	if err != nil {
		t.Fatalf("read JOURNAL: %v", err)
	}
	var journal vNextGenerationJournal
	if err := vNextPublicationDecode(journalBytes, &journal); err != nil {
		t.Fatalf("decode JOURNAL: %v", err)
	}
	return current, journal, true, currentBytes, journalBytes
}

func vNextPublicationAssertCurrentJournalForTest(t *testing.T, root, label string, wantCurrent vNextGenerationPointer, wantState string, wantOld *vNextGenerationPointer, wantNew vNextGenerationPointer) {
	t.Helper()
	current, journal, found, _, _ := vNextPublicationReadCurrentJournalForTest(t, root)
	if current != wantCurrent {
		t.Fatalf("%s CURRENT = %#v, want %#v", label, current, wantCurrent)
	}
	if wantState == "" {
		if found {
			t.Fatalf("%s retains JOURNAL %#v, want absent", label, journal)
		}
		return
	}
	if !found {
		t.Fatalf("%s has no JOURNAL, want %s", label, wantState)
	}
	if journal.State != wantState || journal.New != wantNew {
		t.Fatalf("%s JOURNAL = %#v, want state=%q new=%#v", label, journal, wantState, wantNew)
	}
	if wantOld == nil {
		if journal.Old != nil {
			t.Fatalf("%s JOURNAL old = %#v, want absent", label, journal.Old)
		}
	} else if journal.Old == nil || *journal.Old != *wantOld {
		t.Fatalf("%s JOURNAL old = %#v, want %#v", label, journal.Old, wantOld)
	}
}

func vNextPublicationRestoreLeaseForTest(t *testing.T, retainedPath, leasePath string) {
	t.Helper()
	if err := os.Remove(leasePath); err != nil {
		t.Fatalf("remove test replacement lease %q: %v", leasePath, err)
	}
	if err := os.Rename(retainedPath, leasePath); err != nil {
		t.Fatalf("restore test lease %q: %v", leasePath, err)
	}
}

func TestVNextPublicationCommittedJournalNewSelectedRecoveryRejectsLateLeaseReplacement(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	old, err := baseline.Publish(vNextPublicationArtifactsForTest("old", false))
	if err != nil {
		t.Fatal(err)
	}
	held, err := baseline.Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	newSet := vNextPublicationArtifactsForTest("new", false)
	crash := errors.New("stop after committed JOURNAL before prune")
	writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point == vNextPublicationBeforePrune {
			return crash
		}
		return nil
	}})
	if err != nil {
		held.Release()
		t.Fatal(err)
	}
	if _, err := writer.Publish(newSet); !errors.Is(err, crash) {
		held.Release()
		t.Fatalf("Publish(new) = %v, want committed cut interruption", err)
	}
	held.Release()

	newPointer, committedJournal, foundCommittedJournal, _, _ := vNextPublicationReadCurrentJournalForTest(t, root)
	if !foundCommittedJournal || committedJournal.State != "committed" || committedJournal.New != newPointer || committedJournal.Old == nil || *committedJournal.Old != old {
		t.Fatalf("true committed interruption JOURNAL = %#v, current=%#v", committedJournal, newPointer)
	}
	vNextPublicationAssertCurrentJournalForTest(t, root, "true committed interruption", newPointer, "committed", &old, newPointer)
	leasePath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, old.Generation, vNextPublicationLeaseFile)
	retainedPath := leasePath + ".committed-A"
	var leaseA, leaseB vNextPublicationPathWitness
	hookHit := false
	guard, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationAfterGenerationLeaseIdentity || hookHit {
			return nil
		}
		hookHit = true
		leaseA = vNextPublicationFileWitnessForTest(t, leasePath)
		if err := os.Rename(leasePath, retainedPath); err != nil {
			return err
		}
		if err := os.WriteFile(leasePath, []byte("committed-recovery-B"), 0o600); err != nil {
			return err
		}
		leaseB = vNextPublicationFileWitnessForTest(t, leasePath)
		vNextPublicationAssertDistinctWitnessForTest(t, "committed recovery A/B", leaseA, leaseB)
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	err = guard.Recover(context.Background())
	if !hookHit || err == nil || !strings.Contains(err.Error(), "identity") {
		t.Fatalf("Recover true committed cut = %v, hook=%t, want late identity refusal", err, hookHit)
	}
	vNextPublicationAssertWitnessForTest(t, "committed recovery displaced A", retainedPath, leaseA)
	vNextPublicationAssertWitnessForTest(t, "committed recovery installed B", leasePath, leaseB)
	vNextPublicationAssertCurrentJournalForTest(t, root, "committed recovery refusal", newPointer, "committed", &old, newPointer)

	vNextPublicationRestoreLeaseForTest(t, retainedPath, leasePath)
	fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.Recover(context.Background()); err != nil {
		t.Fatalf("fresh Recover after committed lease restore: %v", err)
	}
	vNextPublicationAssertCurrentJournalForTest(t, root, "committed recovery completion", newPointer, "", nil, vNextGenerationPointer{})
	if fresh.GenerationExists(old.Generation) {
		t.Fatal("fresh committed recovery retained stale old generation")
	}
}

func TestVNextPublicationSuccessfulPublishFinalPruneRejectsLateLeaseReplacement(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	old, err := baseline.Publish(vNextPublicationArtifactsForTest("old", false))
	if err != nil {
		t.Fatal(err)
	}
	newSet := vNextPublicationArtifactsForTest("new", false)
	leasePath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, old.Generation, vNextPublicationLeaseFile)
	retainedPath := leasePath + ".publish-final-prune-A"
	var leaseA, leaseB vNextPublicationPathWitness
	hookHit := false
	guard, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationAfterGenerationLeaseIdentity || hookHit {
			return nil
		}
		hookHit = true
		leaseA = vNextPublicationFileWitnessForTest(t, leasePath)
		if err := os.Rename(leasePath, retainedPath); err != nil {
			return err
		}
		if err := os.WriteFile(leasePath, []byte("final-prune-B"), 0o600); err != nil {
			return err
		}
		leaseB = vNextPublicationFileWitnessForTest(t, leasePath)
		vNextPublicationAssertDistinctWitnessForTest(t, "successful Publish final prune A/B", leaseA, leaseB)
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := guard.Publish(newSet); err == nil || !strings.Contains(err.Error(), "identity") {
		t.Fatalf("successful Publish final prune result = %v, want late identity refusal", err)
	}
	if !hookHit {
		t.Fatal("Publish did not reach its final stale-generation prune")
	}
	newPointer, committedJournal, foundCommittedJournal, _, _ := vNextPublicationReadCurrentJournalForTest(t, root)
	if !foundCommittedJournal || committedJournal.State != "committed" || committedJournal.New != newPointer || committedJournal.Old == nil || *committedJournal.Old != old {
		t.Fatalf("successful Publish final-prune JOURNAL = %#v, current=%#v", committedJournal, newPointer)
	}
	vNextPublicationAssertWitnessForTest(t, "successful Publish displaced A", retainedPath, leaseA)
	vNextPublicationAssertWitnessForTest(t, "successful Publish installed B", leasePath, leaseB)
	vNextPublicationAssertCurrentJournalForTest(t, root, "successful Publish final prune refusal", newPointer, "committed", &old, newPointer)

	vNextPublicationRestoreLeaseForTest(t, retainedPath, leasePath)
	fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.Recover(context.Background()); err != nil {
		t.Fatalf("fresh Recover after Publish final-prune restore: %v", err)
	}
	vNextPublicationAssertCurrentJournalForTest(t, root, "successful Publish final-prune completion", newPointer, "", nil, vNextGenerationPointer{})
	if fresh.GenerationExists(old.Generation) {
		t.Fatal("successful Publish recovery retained stale old generation")
	}
}

func TestVNextPublicationFreshRejectedNewRecoveryRejectsLateLeaseReplacement(t *testing.T) {
	for _, replacement := range [][]byte{nil, []byte("nonempty rejected-new B")} {
		name := "empty-B"
		if len(replacement) != 0 {
			name = "nonempty-B"
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
			newSet := vNextPublicationArtifactsForTest("rejected-new", false)
			crash := errors.New("stop after stage rename before CURRENT selection")
			interrupted, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
				if point == vNextPublicationAfterStageRename {
					return crash
				}
				return nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := interrupted.Publish(newSet); !errors.Is(err, crash) {
				t.Fatalf("Publish(rejected new) = %v, want stage-rename interruption", err)
			}
			_, interruptedJournal, foundInterruptedJournal, _, _ := vNextPublicationReadCurrentJournalForTest(t, root)
			if !foundInterruptedJournal {
				t.Fatal("stage-rename interruption left no prepared JOURNAL")
			}
			newPointer := interruptedJournal.New
			vNextPublicationAssertCurrentJournalForTest(t, root, "old-selected rejected-new interruption", old, "prepared", &old, newPointer)
			newRoot := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, newPointer.Generation)
			if _, err := os.Stat(newRoot); err != nil {
				t.Fatalf("finalized rejected new tree: %v", err)
			}
			leasePath := filepath.Join(newRoot, vNextPublicationLeaseFile)
			retainedPath := leasePath + ".rejected-A"
			var leaseA, leaseB vNextPublicationPathWitness
			hookHit := false
			fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
				if point != vNextPublicationAfterGenerationLeaseIdentity || hookHit {
					return nil
				}
				hookHit = true
				leaseA = vNextPublicationFileWitnessForTest(t, leasePath)
				if err := os.Rename(leasePath, retainedPath); err != nil {
					return err
				}
				if err := os.WriteFile(leasePath, replacement, 0o600); err != nil {
					return err
				}
				leaseB = vNextPublicationFileWitnessForTest(t, leasePath)
				vNextPublicationAssertDistinctWitnessForTest(t, "fresh rejected-new A/B", leaseA, leaseB)
				return nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			err = fresh.Recover(context.Background())
			if !hookHit || err == nil || !strings.Contains(err.Error(), "identity") {
				t.Fatalf("fresh Recover rejected-new result = %v, hook=%t, want identity refusal", err, hookHit)
			}
			vNextPublicationAssertWitnessForTest(t, "fresh rejected-new displaced A", retainedPath, leaseA)
			vNextPublicationAssertWitnessForTest(t, "fresh rejected-new installed B", leasePath, leaseB)
			vNextPublicationAssertCurrentJournalForTest(t, root, "fresh rejected-new refusal", old, "prepared", &old, newPointer)
			if _, err := os.Stat(newRoot); err != nil {
				t.Fatalf("fresh rejected-new refusal removed rejected root: %v", err)
			}

			vNextPublicationRestoreLeaseForTest(t, retainedPath, leasePath)
			completed, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			if err := completed.Recover(context.Background()); err != nil {
				t.Fatalf("fresh Recover after rejected-new fixture restore: %v", err)
			}
			vNextPublicationAssertCurrentJournalForTest(t, root, "fresh rejected-new completion", old, "", nil, vNextGenerationPointer{})
			if completed.GenerationExists(newPointer.Generation) {
				t.Fatal("fresh rejected-new completion retained rejected generation")
			}
		})
	}
}

func TestVNextPublicationImmediateRollbackRejectsLateLeaseReplacementIdentityVariants(t *testing.T) {
	for _, replacement := range [][]byte{nil, []byte("nonempty rollback B")} {
		name := "empty-B"
		if len(replacement) != 0 {
			name = "nonempty-B"
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
			broken := vNextPublicationArtifactsForTest("rollback-new", false)
			validation := errors.New("active validation failure")
			calls := 0
			broken.Validate = func(fs.FS) error {
				calls++
				if calls == 2 {
					return validation
				}
				return nil
			}
			newGeneration := vNextPublicationGenerationID(broken.Files)
			newRoot := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, newGeneration)
			leasePath := filepath.Join(newRoot, vNextPublicationLeaseFile)
			retainedPath := leasePath + ".rollback-A"
			var leaseA, leaseB vNextPublicationPathWitness
			hookHit := false
			guard, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
				if point != vNextPublicationAfterGenerationLeaseIdentity || hookHit {
					return nil
				}
				hookHit = true
				leaseA = vNextPublicationFileWitnessForTest(t, leasePath)
				if err := os.Rename(leasePath, retainedPath); err != nil {
					return err
				}
				if err := os.WriteFile(leasePath, replacement, 0o600); err != nil {
					return err
				}
				leaseB = vNextPublicationFileWitnessForTest(t, leasePath)
				vNextPublicationAssertDistinctWitnessForTest(t, "immediate rollback A/B", leaseA, leaseB)
				return nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			_, err = guard.Publish(broken)
			if !hookHit || !errors.Is(err, validation) || !strings.Contains(err.Error(), "identity") {
				t.Fatalf("immediate rollback result = %v, hook=%t, want validation and identity refusal", err, hookHit)
			}
			if calls != 2 {
				t.Fatalf("immediate rollback validation calls = %d, want 2", calls)
			}
			_, rollbackJournal, foundRollbackJournal, _, _ := vNextPublicationReadCurrentJournalForTest(t, root)
			if !foundRollbackJournal {
				t.Fatal("immediate rollback identity refusal left no prepared JOURNAL")
			}
			newPointer := rollbackJournal.New
			vNextPublicationAssertWitnessForTest(t, "immediate rollback displaced A", retainedPath, leaseA)
			vNextPublicationAssertWitnessForTest(t, "immediate rollback installed B", leasePath, leaseB)
			vNextPublicationAssertCurrentJournalForTest(t, root, "immediate rollback refusal", old, "prepared", &old, newPointer)

			vNextPublicationRestoreLeaseForTest(t, retainedPath, leasePath)
			fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			if err := fresh.Recover(context.Background()); err != nil {
				t.Fatalf("fresh Recover after immediate rollback fixture restore: %v", err)
			}
			vNextPublicationAssertCurrentJournalForTest(t, root, "immediate rollback completion", old, "", nil, vNextGenerationPointer{})
		})
	}
}
