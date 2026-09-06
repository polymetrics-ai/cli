package main

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"golang.org/x/sys/unix"
)

func TestVNextPublicationRecordedCaptureRejectsReplacementBeforeDependentUse(t *testing.T) {
	for _, test := range []struct {
		name  string
		check func(*vNextPublicationDirectory, vNextPublicationControlRepairCapture) error
	}{
		{
			name:  "validation",
			check: vNextPublicationValidateCaptureLocked,
		},
		{
			name: "candidate",
			check: func(transaction *vNextPublicationDirectory, capture vNextPublicationControlRepairCapture) error {
				return vNextPublicationValidateCapturedCandidateLocked(transaction, capture, vNextPublicationAbsentControlState())
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			root, transaction, capture := vNextPublicationRecordedEmptyCaptureForTest(t)
			capturePath := filepath.Join(root, capture.Name)
			vNextPublicationReplacePrivateDirectoryForTest(t, capturePath)
			if err := test.check(transaction, capture); err == nil || !strings.Contains(err.Error(), "capture identity changed") {
				t.Fatalf("%s accepted capture B or returned wrong error: %v", test.name, err)
			}
			vNextPublicationAssertCaptureReplacementForTest(t, capturePath)
		})
	}
}

func TestVNextPublicationCompleteCaptureRejectsReplacementBeforeCurrentMovement(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := baseline.Publish(vNextPublicationArtifactsForTest("active", false)); err != nil {
		t.Fatal(err)
	}
	writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	operation, err := writer.openOperation(context.Background(), syscall.LOCK_EX, true)
	if err != nil {
		t.Fatal(err)
	}
	defer operation.close()
	graph, err := writer.scanControlAuthorityLocked(operation)
	if err != nil {
		t.Fatal(err)
	}
	head := graph.heads[vNextPublicationCurrentFile]
	graph.close()
	if head == nil {
		t.Fatal("missing CURRENT authority head")
	}
	state, err := writer.createControlRepairLocked(operation, vNextPublicationCurrentFile, head, vNextPublicationAbsentControlState(), nil, false)
	if err != nil {
		t.Fatal(err)
	}
	defer state.close()
	if err := writer.beginControlCaptureLocked(operation, state); err != nil {
		t.Fatal(err)
	}
	capture := *state.phases[len(state.phases)-1].record.Capture
	currentPath := filepath.Join(root, "acme", vNextPublicationCurrentFile)
	currentBefore, err := os.ReadFile(currentPath)
	if err != nil {
		t.Fatal(err)
	}
	capturePath := filepath.Join(root, "acme", state.transactionName, capture.Name)
	vNextPublicationReplacePrivateDirectoryForTest(t, capturePath)
	if err := writer.completeControlCaptureLocked(operation, state); err == nil || !strings.Contains(err.Error(), "capture identity changed") {
		t.Fatalf("complete capture accepted B or returned wrong error: %v", err)
	}
	if currentAfter, readErr := os.ReadFile(currentPath); readErr != nil || !bytes.Equal(currentAfter, currentBefore) {
		t.Fatalf("capture B mutation changed CURRENT: err=%v got=%q want=%q", readErr, currentAfter, currentBefore)
	}
	if _, err := os.Lstat(filepath.Join(capturePath, vNextPublicationControlCaptureMember)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("capture B received CURRENT before refusal: %v", err)
	}
	vNextPublicationAssertCaptureReplacementForTest(t, capturePath)
}

func TestVNextPublicationDurableCaptureRefusesReplacementForCurrentAndJournal(t *testing.T) {
	for _, test := range []struct {
		target   string
		name     string
		swapAt   int
		exercise func(*vNextGenerationPublisher, *vNextPublicationOperation, *vNextPublicationControlRepairState, vNextPublicationControlRepairCapture) error
	}{
		{
			target: vNextPublicationCurrentFile,
			name:   "CURRENT/validation-open",
			swapAt: 1,
			exercise: func(_ *vNextGenerationPublisher, operation *vNextPublicationOperation, state *vNextPublicationControlRepairState, capture vNextPublicationControlRepairCapture) error {
				transaction, err := state.openTransaction(operation)
				if err != nil {
					return err
				}
				defer vNextPublicationCloseAfter(&err, transaction, "CURRENT validation transaction")
				return vNextPublicationValidateCaptureLocked(transaction, capture)
			},
		},
		{
			target: vNextPublicationCurrentFile,
			name:   "CURRENT/candidate-open",
			swapAt: 2,
			exercise: func(_ *vNextGenerationPublisher, operation *vNextPublicationOperation, state *vNextPublicationControlRepairState, capture vNextPublicationControlRepairCapture) error {
				transaction, err := state.openTransaction(operation)
				if err != nil {
					return err
				}
				defer vNextPublicationCloseAfter(&err, transaction, "CURRENT candidate transaction")
				return vNextPublicationValidateCapturedCandidateLocked(transaction, capture, vNextPublicationAbsentControlState())
			},
		},
		{
			target: vNextPublicationCurrentFile,
			name:   "CURRENT/mutating-open",
			swapAt: 2,
			exercise: func(writer *vNextGenerationPublisher, operation *vNextPublicationOperation, state *vNextPublicationControlRepairState, _ vNextPublicationControlRepairCapture) error {
				return writer.completeControlCaptureLocked(operation, state)
			},
		},
		{
			target: vNextPublicationJournalFile,
			name:   "JOURNAL/validation-open",
			swapAt: 1,
			exercise: func(_ *vNextGenerationPublisher, operation *vNextPublicationOperation, state *vNextPublicationControlRepairState, capture vNextPublicationControlRepairCapture) error {
				transaction, err := state.openTransaction(operation)
				if err != nil {
					return err
				}
				defer vNextPublicationCloseAfter(&err, transaction, "JOURNAL validation transaction")
				return vNextPublicationValidateCaptureLocked(transaction, capture)
			},
		},
		{
			target: vNextPublicationJournalFile,
			name:   "JOURNAL/candidate-open",
			swapAt: 2,
			exercise: func(_ *vNextGenerationPublisher, operation *vNextPublicationOperation, state *vNextPublicationControlRepairState, capture vNextPublicationControlRepairCapture) error {
				transaction, err := state.openTransaction(operation)
				if err != nil {
					return err
				}
				defer vNextPublicationCloseAfter(&err, transaction, "JOURNAL candidate transaction")
				return vNextPublicationValidateCapturedCandidateLocked(transaction, capture, vNextPublicationAbsentControlState())
			},
		},
		{
			target: vNextPublicationJournalFile,
			name:   "JOURNAL/mutating-open",
			swapAt: 2,
			exercise: func(writer *vNextGenerationPublisher, operation *vNextPublicationOperation, state *vNextPublicationControlRepairState, _ vNextPublicationControlRepairCapture) error {
				return writer.completeControlCaptureLocked(operation, state)
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			active, err := baseline.Publish(vNextPublicationArtifactsForTest("active", true))
			if err != nil {
				t.Fatal(err)
			}
			writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			operation, err := writer.openOperation(context.Background(), syscall.LOCK_EX, true)
			if err != nil {
				t.Fatal(err)
			}
			if test.target == vNextPublicationJournalFile {
				if err := writer.writeJournalLocked(operation, vNextGenerationJournal{New: active, State: "prepared"}); err != nil {
					operation.close()
					t.Fatal(err)
				}
			}
			publicPath := filepath.Join(root, "acme", test.target)
			publicBefore, err := os.ReadFile(publicPath)
			if err != nil {
				operation.close()
				t.Fatal(err)
			}
			graph, err := writer.scanControlAuthorityLocked(operation)
			if err != nil {
				operation.close()
				t.Fatal(err)
			}
			head := graph.heads[test.target]
			graph.close()
			if head == nil {
				operation.close()
				t.Fatalf("missing %s authority head", test.target)
			}
			state, err := writer.createControlRepairLocked(operation, test.target, head, vNextPublicationAbsentControlState(), nil, false)
			if err != nil {
				operation.close()
				t.Fatal(err)
			}
			if err := writer.beginControlCaptureLocked(operation, state); err != nil {
				operation.close()
				t.Fatal(err)
			}
			capture := *state.phases[len(state.phases)-1].record.Capture
			capturePath := filepath.Join(root, "acme", state.transactionName, capture.Name)
			captureA, entriesA := vNextPublicationCaptureSnapshotForTest(t, capturePath)
			if len(entriesA) != 0 {
				operation.close()
				t.Fatalf("initial %s capture entries = %v, want empty", test.target, entriesA)
			}
			opens := 0
			var captureB vNextPublicationIdentity
			vNextPublicationBeforeRecordedCaptureOpenForTest = func(transaction *vNextPublicationDirectory, observed vNextPublicationControlRepairCapture) {
				opens++
				if observed.Name != capture.Name || opens != test.swapAt {
					return
				}
				if err := os.Rename(capturePath, capturePath+".attacker-moved"); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(capturePath, 0o700); err != nil {
					t.Fatal(err)
				}
				captureB, _ = vNextPublicationCaptureSnapshotForTest(t, capturePath)
			}
			t.Cleanup(func() { vNextPublicationBeforeRecordedCaptureOpenForTest = nil })
			err = test.exercise(writer, operation, state, capture)
			vNextPublicationBeforeRecordedCaptureOpenForTest = nil
			if err == nil || !strings.Contains(err.Error(), "capture identity changed") {
				operation.close()
				t.Fatalf("%s accepted replacement or returned wrong error: %v", test.name, err)
			}
			if opens != test.swapAt {
				operation.close()
				t.Fatalf("%s observed capture opens = %d, want replacement at %d", test.name, opens, test.swapAt)
			}
			if got, err := os.ReadFile(publicPath); err != nil || !bytes.Equal(got, publicBefore) {
				operation.close()
				t.Fatalf("%s replacement changed public %s: got=%q err=%v want=%q", test.name, test.target, got, err, publicBefore)
			}
			movedA, movedEntries := vNextPublicationCaptureSnapshotForTest(t, capturePath+".attacker-moved")
			currentB, replacementEntries := vNextPublicationCaptureSnapshotForTest(t, capturePath)
			if captureA != movedA || captureA == captureB || captureA == currentB || captureB != currentB || captureA.mode != unix.S_IFDIR || captureB.mode != unix.S_IFDIR || len(movedEntries) != 0 || len(replacementEntries) != 0 {
				operation.close()
				t.Fatalf("%s capture A/B observation invalid: initial=%#v moved=%#v B=%#v entriesA=%v entriesB=%v", test.name, captureA, movedA, currentB, movedEntries, replacementEntries)
			}
			if _, err := os.Lstat(filepath.Join(capturePath, vNextPublicationControlCaptureMember)); !errors.Is(err, fs.ErrNotExist) {
				operation.close()
				t.Fatalf("%s moved public %s into replacement B: %v", test.name, test.target, err)
			}
			if err := os.Remove(capturePath); err != nil {
				operation.close()
				t.Fatal(err)
			}
			if err := os.Rename(capturePath+".attacker-moved", capturePath); err != nil {
				operation.close()
				t.Fatal(err)
			}
			operation.close()
			fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			if err := fresh.Recover(context.Background()); err != nil {
				t.Fatalf("%s lock reacquisition recovery after restoring owned capture A: %v", test.name, err)
			}
			if _, err := fresh.Publish(vNextPublicationArtifactsForTest(strings.ReplaceAll(test.name, "/", "-")+"-retry", true)); err != nil {
				t.Fatalf("%s normal retry after lock reacquisition: %v", test.name, err)
			}
		})
	}
}

func vNextPublicationCaptureSnapshotForTest(t *testing.T, path string) (vNextPublicationIdentity, []string) {
	t.Helper()
	directory, err := vNextPublicationOpenDirectory(path, "capture snapshot")
	if err != nil {
		t.Fatal(err)
	}
	identity, err := vNextPublicationIdentityFromFile(directory.file, "capture snapshot")
	entries, readErr := directory.readDir()
	closeErr := directory.Close()
	if err != nil || readErr != nil || closeErr != nil {
		t.Fatalf("capture snapshot %q: identity=%v read=%v close=%v", path, err, readErr, closeErr)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return identity, names
}

func vNextPublicationRecordedEmptyCaptureForTest(t *testing.T) (string, *vNextPublicationDirectory, vNextPublicationControlRepairCapture) {
	t.Helper()
	root := t.TempDir()
	transaction, err := vNextPublicationOpenDirectory(root, "recorded capture transaction")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closeErr := transaction.Close(); closeErr != nil {
			t.Errorf("close recorded capture transaction: %v", closeErr)
		}
	})
	const name = ".capture-001"
	if err := unix.Mkdirat(int(transaction.file.Fd()), name, 0o700); err != nil {
		t.Fatal(err)
	}
	directory, err := transaction.openDirectory(name, "recorded capture A")
	if err != nil {
		t.Fatal(err)
	}
	identity, err := vNextPublicationIdentityFromFile(directory.file, "recorded capture A")
	if closeErr := directory.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		t.Fatal(err)
	}
	return root, transaction, vNextPublicationControlRepairCapture{Attempt: 1, Name: name, Identity: vNextPublicationRecordIdentity(identity)}
}

func vNextPublicationAssertCaptureReplacementForTest(t *testing.T, capturePath string) {
	t.Helper()
	if info, err := os.Stat(capturePath + ".attacker-moved"); err != nil || !info.IsDir() {
		t.Fatalf("retained capture A = info=%v err=%v", info, err)
	}
	entries, err := os.ReadDir(capturePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("capture B is not empty after refusal: %#v", entries)
	}
	if info, err := os.Lstat(capturePath); err != nil || !info.IsDir() {
		t.Fatalf("capture B = info=%v err=%v", info, err)
	}
}
