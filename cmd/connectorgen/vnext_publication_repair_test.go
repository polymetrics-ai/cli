package main

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

type vNextPublicationControlRepairTestControl struct {
	name    string
	target  string
	noPrior bool
	write   func(*vNextGenerationPublisher, *vNextPublicationOperation, vNextGenerationPointer) error
}

func TestVNextGenerationPublisherRecoversEveryControlRepairDurableCut(t *testing.T) {
	controls := []vNextPublicationControlRepairTestControl{
		{
			name:   "CURRENT/existing",
			target: vNextPublicationCurrentFile,
			write: func(publisher *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
				return publisher.writeCurrentLocked(operation, pointer)
			},
		},
		{
			name:   "JOURNAL/existing",
			target: vNextPublicationJournalFile,
			write: func(publisher *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
				return publisher.writeJournalLocked(operation, vNextGenerationJournal{New: pointer, State: "prepared"})
			},
		},
		{
			name:    "CURRENT/no-prior",
			target:  vNextPublicationCurrentFile,
			noPrior: true,
			write: func(publisher *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
				return publisher.writeCurrentLocked(operation, pointer)
			},
		},
		{
			name:    "JOURNAL/no-prior",
			target:  vNextPublicationJournalFile,
			noPrior: true,
			write: func(publisher *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
				return publisher.writeJournalLocked(operation, vNextGenerationJournal{New: pointer, State: "prepared"})
			},
		},
	}
	normalCuts := []vNextPublicationFaultPoint{
		vNextPublicationAfterControlRepairBackupSync,
		vNextPublicationAfterControlRepairPrepared,
		vNextPublicationAfterControlRepairInstall,
		vNextPublicationAfterControlRepairInstallSync,
		vNextPublicationAfterControlRepairClearSync,
	}
	for _, control := range controls {
		for _, cut := range normalCuts {
			t.Run(control.name+"/"+string(cut), func(t *testing.T) {
				vNextPublicationExerciseControlRepairCut(t, control, cut, nil)
			})
		}
	}
}

func TestVNextGenerationPublisherRecoversEveryRetainedReplacementCut(t *testing.T) {
	controls := []vNextPublicationControlRepairTestControl{
		{
			name:   "CURRENT/existing",
			target: vNextPublicationCurrentFile,
			write: func(publisher *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
				return publisher.writeCurrentLocked(operation, pointer)
			},
		},
		{
			name:   "JOURNAL/existing",
			target: vNextPublicationJournalFile,
			write: func(publisher *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
				return publisher.writeJournalLocked(operation, vNextGenerationJournal{New: pointer, State: "prepared"})
			},
		},
		{
			name:    "CURRENT/no-prior",
			target:  vNextPublicationCurrentFile,
			noPrior: true,
			write: func(publisher *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
				return publisher.writeCurrentLocked(operation, pointer)
			},
		},
		{
			name:    "JOURNAL/no-prior",
			target:  vNextPublicationJournalFile,
			noPrior: true,
			write: func(publisher *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
				return publisher.writeJournalLocked(operation, vNextGenerationJournal{New: pointer, State: "prepared"})
			},
		},
	}
	cuts := []vNextPublicationFaultPoint{
		vNextPublicationAfterControlRepairReplacementRetainSync,
		vNextPublicationAfterControlRepairReplacementSync,
		vNextPublicationAfterControlRepairPublicRestoreSync,
		vNextPublicationAfterControlRepairRestoreSync,
		vNextPublicationAfterControlRepairClearSync,
	}
	for _, control := range controls {
		for _, cut := range cuts {
			t.Run(control.name+"/"+string(cut), func(t *testing.T) {
				vNextPublicationExerciseControlRepairCut(t, control, cut, func(pointer vNextGenerationPointer) []byte {
					return vNextPublicationControlRepairReplacementForTest(t, control.target, pointer)
				})
			})
		}
	}
}

func vNextPublicationExerciseControlRepairCut(t *testing.T, control vNextPublicationControlRepairTestControl, cut vNextPublicationFaultPoint, replacement func(vNextGenerationPointer) []byte) {
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
	vNextPublicationPrepareControlRepairTargetForTest(t, baseline, root, control, pointer)

	crash := errors.New("injected control repair crash")
	swapped := false
	var replacementPayload []byte
	if replacement != nil {
		replacementPayload = replacement(pointer)
	}
	failing, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if replacementPayload != nil && point == vNextPublicationAfterFinalControlSourceIdentity && !swapped {
			swapped = true
			vNextReplacePublicationTemporaryForTest(t, root, replacementPayload)
		}
		if point == cut {
			return crash
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	operation, err := failing.openOperation(context.Background(), syscall.LOCK_EX, true)
	if err != nil {
		t.Fatal(err)
	}
	err = control.write(failing, operation, pointer)
	operation.close()
	if !errors.Is(err, crash) {
		t.Fatalf("write at %s error = %v, want injected crash", cut, err)
	}
	if replacementPayload != nil && !swapped {
		t.Fatal("replacement scenario did not reach final source-identity barrier")
	}
	connectorRoot := filepath.Join(root, "acme")
	pendingAuthority := cut != vNextPublicationAfterControlRepairBackupSync && cut != vNextPublicationAfterControlRepairClearSync
	if _, err := os.Stat(filepath.Join(connectorRoot, vNextPublicationControlRepairFile)); pendingAuthority && err != nil {
		t.Fatalf("repair authority at %s error = %v, want pending authority", cut, err)
	} else if !pendingAuthority && !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("repair authority at %s = %v, want absent", cut, err)
	}

	recovered, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if err := recovered.Recover(context.Background()); err != nil {
		t.Fatalf("Recover() after %s error = %v", cut, err)
	}
	if _, err := os.Stat(filepath.Join(connectorRoot, vNextPublicationControlRepairFile)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("repair authority after recovery = %v, want absent", err)
	}
	completed := cut == vNextPublicationAfterControlRepairClearSync && replacementPayload == nil
	if control.noPrior && control.target == vNextPublicationCurrentFile && !completed {
		if _, err := os.Stat(filepath.Join(connectorRoot, control.target)); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("no-prior CURRENT after %s = %v, want absent", cut, err)
		}
	} else {
		active, err := recovered.Open(context.Background())
		if err != nil {
			t.Fatalf("Open() after %s error = %v", cut, err)
		}
		assertVNextPublicationMarker(t, active, "active")
		active.Release()
	}
	if replacementPayload != nil {
		vNextPublicationFindControlRepairReplacementForTest(t, root, replacementPayload)
	}
}

func vNextPublicationPrepareControlRepairTargetForTest(t *testing.T, publisher *vNextGenerationPublisher, root string, control vNextPublicationControlRepairTestControl, pointer vNextGenerationPointer) {
	t.Helper()
	if control.target == vNextPublicationJournalFile && !control.noPrior {
		payload, err := vNextPublicationJSON(vNextGenerationJournal{New: pointer, State: "prepared"})
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(root, "acme", control.target)
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write(payload); err != nil {
			_ = file.Close()
			t.Fatal(err)
		}
		if err := file.Sync(); err != nil {
			_ = file.Close()
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if control.target == vNextPublicationCurrentFile && control.noPrior {
		operation, err := publisher.openOperation(context.Background(), syscall.LOCK_EX, true)
		if err != nil {
			t.Fatal(err)
		}
		err = publisher.removeCurrentLocked(operation)
		operation.close()
		if err != nil {
			t.Fatal(err)
		}
	}
}

func vNextPublicationFindControlRepairReplacementForTest(t *testing.T, root string, want []byte) string {
	t.Helper()
	connectorRoot := filepath.Join(root, "acme")
	entries, err := os.ReadDir(connectorRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), ".connectorgen-quarantine-") {
			continue
		}
		path := filepath.Join(connectorRoot, entry.Name(), vNextPublicationControlReplacementMember)
		payload, err := os.ReadFile(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if string(payload) == string(want) {
			return path
		}
	}
	t.Fatalf("control repair does not retain replacement %q", want)
	return ""
}

func vNextPublicationControlRepairReplacementForTest(t *testing.T, target string, old vNextGenerationPointer) []byte {
	t.Helper()
	replacement := vNextGenerationPointer{
		Generation:      strings.Repeat("a", 64),
		IntegrityDigest: strings.Repeat("b", 64),
	}
	var (
		payload []byte
		err     error
	)
	switch target {
	case vNextPublicationCurrentFile:
		payload, err = vNextPublicationJSON(replacement)
	case vNextPublicationJournalFile:
		payload, err = vNextPublicationJSON(vNextGenerationJournal{Old: &old, New: replacement, State: "prepared"})
	default:
		t.Fatalf("unsupported control repair target %q", target)
	}
	if err != nil {
		t.Fatal(err)
	}
	return payload
}
