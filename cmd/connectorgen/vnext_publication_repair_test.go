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
		vNextPublicationAfterControlRepairInstalledPhaseSync,
		vNextPublicationAfterControlRepairAuthorityRetireSync,
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
		vNextPublicationAfterControlRepairAuthorityRetireSync,
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

// A publisher's Flock serializes cooperative writers, but it does not stop a
// same-permission actor from renaming a public control while that writer holds
// the lock. The prepared recovery authority must therefore live outside both
// public controls before either substitution can happen.
func TestVNextGenerationPublisherPrivatePreparedAuthoritySurvivesPublicControlSubstitution(t *testing.T) {
	controls := []vNextPublicationControlRepairTestControl{
		{
			name:   "CURRENT",
			target: vNextPublicationCurrentFile,
			write: func(publisher *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
				return publisher.writeCurrentLocked(operation, pointer)
			},
		},
		{
			name:   "JOURNAL",
			target: vNextPublicationJournalFile,
			write: func(publisher *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
				return publisher.writeJournalLocked(operation, vNextGenerationJournal{New: pointer, State: "prepared"})
			},
		},
	}
	attacks := []struct {
		name string
		run  func(t *testing.T, root string, target string, point vNextPublicationFaultPoint) error
	}{
		{
			name: "target-rename",
			run: func(t *testing.T, root string, target string, point vNextPublicationFaultPoint) error {
				if point != vNextPublicationAfterControlRepairPrepared {
					return nil
				}
				vNextReplacePublicationControlForTest(t, root, target, []byte("substituted public control"))
				return errors.New("crash after target substitution")
			},
		},
		{
			name: "target-unlink",
			run: func(t *testing.T, root string, target string, point vNextPublicationFaultPoint) error {
				if point != vNextPublicationAfterControlRepairPrepared {
					return nil
				}
				if err := os.Remove(filepath.Join(root, "acme", target)); err != nil {
					t.Fatal(err)
				}
				return errors.New("crash after target unlink")
			},
		},
		{
			name: "source-rename",
			run: func(t *testing.T, root string, target string, point vNextPublicationFaultPoint) error {
				if point == vNextPublicationAfterFinalControlSourceIdentity {
					vNextReplacePublicationTemporaryForTest(t, root, []byte("substituted private source"))
				}
				if point == vNextPublicationAfterControlRepairInstall {
					return errors.New("crash after source substitution")
				}
				return nil
			},
		},
	}

	for _, control := range controls {
		for _, attack := range attacks {
			t.Run(control.name+"/"+attack.name, func(t *testing.T) {
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

				failing, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
					return attack.run(t, root, control.target, point)
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
				if err == nil || !strings.Contains(err.Error(), "crash after") {
					t.Fatalf("write after %s substitution error = %v, want injected crash", attack.name, err)
				}

				connectorRoot := filepath.Join(root, "acme")
				entries, err := os.ReadDir(connectorRoot)
				if err != nil {
					t.Fatal(err)
				}
				privateAuthorities := 0
				for _, entry := range entries {
					if entry.IsDir() && strings.HasPrefix(entry.Name(), ".connectorgen-control-repair-") {
						privateAuthorities++
					}
				}
				if privateAuthorities != 1 {
					t.Fatalf("private prepared recovery authorities = %d, want one after %s substitution", privateAuthorities, attack.name)
				}
				if _, err := os.Lstat(filepath.Join(connectorRoot, ".connectorgen-control-repair.json")); !errors.Is(err, fs.ErrNotExist) {
					t.Fatalf("mutable root repair authority after %s substitution = %v, want absent", attack.name, err)
				}

				recovered, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
				if err != nil {
					t.Fatal(err)
				}
				if err := recovered.Recover(context.Background()); err != nil {
					t.Fatalf("Recover() after %s substitution = %v", attack.name, err)
				}
				active, err := recovered.Open(context.Background())
				if err != nil {
					t.Fatalf("Open() after %s substitution = %v", attack.name, err)
				}
				assertVNextPublicationMarker(t, active, "active")
				active.Release()
			})
		}
	}
}

func TestVNextGenerationPublisherWritesBoundImmutableRecoveryPhaseChains(t *testing.T) {
	cases := []struct {
		name       string
		cut        vNextPublicationFaultPoint
		substitute bool
		wantStates []string
	}{
		{
			name:       "installed",
			cut:        vNextPublicationAfterControlRepairInstalledPhaseSync,
			wantStates: []string{vNextPublicationControlRepairInstalled},
		},
		{
			name:       "replacement-restored",
			cut:        vNextPublicationAfterControlRepairRestoreSync,
			substitute: true,
			wantStates: []string{vNextPublicationControlRepairReplacementRetained, vNextPublicationControlRepairRestored},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			pointer, err := baseline.Publish(vNextPublicationArtifactsForTest("active", false))
			if err != nil {
				t.Fatal(err)
			}

			crash := errors.New("crash after durable recovery phase")
			substituted := false
			failing, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
				if test.substitute && point == vNextPublicationAfterFinalControlSourceIdentity && !substituted {
					substituted = true
					vNextReplacePublicationTemporaryForTest(t, root, []byte("phase-chain substitute"))
				}
				if point == test.cut {
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
			err = failing.writeCurrentLocked(operation, pointer)
			operation.close()
			if !errors.Is(err, crash) {
				t.Fatalf("writeCurrentLocked() error = %v, want injected crash", err)
			}
			if test.substitute && !substituted {
				t.Fatal("replacement phase path did not replace the final source")
			}

			transactionPath := vNextPublicationPendingControlRepairTransactionPathForTest(t, root)
			transaction, err := vNextPublicationOpenDirectory(transactionPath, "test control repair transaction")
			if err != nil {
				t.Fatal(err)
			}
			prepared, err := transaction.readFile(vNextPublicationControlRepairPreparedFile, "test control repair prepared authority")
			if err != nil {
				_ = transaction.Close()
				t.Fatal(err)
			}
			preparedIdentity, err := transaction.identityAt(vNextPublicationControlRepairPreparedFile, "test control repair prepared authority")
			if err != nil {
				_ = transaction.Close()
				t.Fatal(err)
			}
			previous := vNextPublicationControlRepairPhaseReference{
				Name:     vNextPublicationControlRepairPreparedFile,
				Identity: vNextPublicationRecordIdentity(preparedIdentity),
				Digest:   vNextPublicationDigest(prepared),
			}
			for index, wantState := range test.wantStates {
				name := vNextPublicationControlRepairPhaseName(index + 1)
				payload, err := transaction.readFile(name, "test control repair phase")
				if err != nil {
					_ = transaction.Close()
					t.Fatal(err)
				}
				var phase vNextPublicationControlRepairPhase
				if err := vNextPublicationDecode(payload, &phase); err != nil {
					_ = transaction.Close()
					t.Fatal(err)
				}
				if phase.Sequence != index+1 || phase.State != wantState {
					_ = transaction.Close()
					t.Fatalf("phase %d = %#v, want state %q", index+1, phase, wantState)
				}
				if phase.PreparedDigest != vNextPublicationDigest(prepared) || phase.PreparedIdentity != vNextPublicationRecordIdentity(preparedIdentity) || phase.Previous != previous {
					_ = transaction.Close()
					t.Fatalf("phase %d is not bound to the immutable prepared authority", index+1)
				}
				phaseIdentity, err := transaction.identityAt(name, "test control repair phase")
				if err != nil {
					_ = transaction.Close()
					t.Fatal(err)
				}
				previous = vNextPublicationControlRepairPhaseReference{
					Name:     name,
					Identity: vNextPublicationRecordIdentity(phaseIdentity),
					Digest:   vNextPublicationDigest(payload),
				}
			}
			after, err := transaction.readFile(vNextPublicationControlRepairPreparedFile, "test control repair prepared authority")
			if err != nil {
				_ = transaction.Close()
				t.Fatal(err)
			}
			if string(after) != string(prepared) {
				_ = transaction.Close()
				t.Fatal("append-only phase persistence changed the prepared authority")
			}
			if err := transaction.Close(); err != nil {
				t.Fatal(err)
			}

			recovered, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			if err := recovered.Recover(context.Background()); err != nil {
				t.Fatalf("Recover() after %s phase chain = %v", test.name, err)
			}
		})
	}
}

func TestVNextGenerationPublisherFinalTargetSubstitutionCannotClearPrivateAuthority(t *testing.T) {
	controls := []vNextPublicationControlRepairTestControl{
		{
			name:   "CURRENT",
			target: vNextPublicationCurrentFile,
			write: func(publisher *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
				return publisher.writeCurrentLocked(operation, pointer)
			},
		},
		{
			name:   "JOURNAL",
			target: vNextPublicationJournalFile,
			write: func(publisher *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
				return publisher.writeJournalLocked(operation, vNextGenerationJournal{New: pointer, State: "prepared"})
			},
		},
	}
	for _, control := range controls {
		t.Run(control.name, func(t *testing.T) {
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
			replacement := []byte("final public control substitute")
			swapped := false
			failing, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
				if point == vNextPublicationBeforeControlRepairAuthorityClear && !swapped {
					swapped = true
					vNextReplacePublicationControlForTest(t, root, control.target, replacement)
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
			if !swapped {
				t.Fatal("final cleanup barrier did not permit the same-permission target replacement")
			}
			if err == nil || !strings.Contains(err.Error(), "final installed repair cleanup") {
				t.Fatalf("write after final target substitution error = %v, want final cleanup refusal", err)
			}
			if got := vNextPublicationPendingControlRepairAuthoritiesForTest(t, root); got != 1 {
				t.Fatalf("pending authority after final target substitution = %d, want one", got)
			}

			recovered, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			if err := recovered.Recover(context.Background()); err != nil {
				t.Fatalf("Recover() after final target substitution = %v", err)
			}
			vNextPublicationFindControlRepairReplacementForTest(t, root, replacement)
			active, err := recovered.Open(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			assertVNextPublicationMarker(t, active, "active")
			active.Release()
		})
	}
}

func TestVNextGenerationPublisherFinalNoPriorTargetUnlinkCannotClearPrivateAuthority(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	pointer, err := baseline.Publish(vNextPublicationArtifactsForTest("active", false))
	if err != nil {
		t.Fatal(err)
	}
	control := vNextPublicationControlRepairTestControl{
		target:  vNextPublicationCurrentFile,
		noPrior: true,
		write: func(publisher *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
			return publisher.writeCurrentLocked(operation, pointer)
		},
	}
	vNextPublicationPrepareControlRepairTargetForTest(t, baseline, root, control, pointer)
	unlinked := false
	failing, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point == vNextPublicationBeforeControlRepairAuthorityClear && !unlinked {
			unlinked = true
			if err := os.Remove(filepath.Join(root, "acme", vNextPublicationCurrentFile)); err != nil {
				t.Fatal(err)
			}
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
	if !unlinked {
		t.Fatal("final cleanup barrier did not permit the same-permission target unlink")
	}
	if err == nil || !strings.Contains(err.Error(), "final installed repair cleanup") {
		t.Fatalf("write after final target unlink error = %v, want final cleanup refusal", err)
	}
	if got := vNextPublicationPendingControlRepairAuthoritiesForTest(t, root); got != 1 {
		t.Fatalf("pending authority after final target unlink = %d, want one", got)
	}
	recovered, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if err := recovered.Recover(context.Background()); err != nil {
		t.Fatalf("Recover() after final target unlink = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "acme", vNextPublicationCurrentFile)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("CURRENT after final target unlink recovery = %v, want absent", err)
	}
}

func TestVNextGenerationPublisherRefusesSubstitutedPreparedAuthorityBeforePublicControlExposure(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	pointer, err := baseline.Publish(vNextPublicationArtifactsForTest("active", false))
	if err != nil {
		t.Fatal(err)
	}
	targetPath := filepath.Join(root, "acme", vNextPublicationCurrentFile)
	prior, err := os.Stat(targetPath)
	if err != nil {
		t.Fatal(err)
	}

	var moved string
	substituted := false
	failing, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationAfterFinalControlSourceIdentity || substituted {
			return nil
		}
		substituted = true
		transaction := vNextPublicationPendingControlRepairTransactionPathForTest(t, root)
		prepared := filepath.Join(transaction, vNextPublicationControlRepairPreparedFile)
		payload, err := os.ReadFile(prepared)
		if err != nil {
			t.Fatal(err)
		}
		moved = prepared + ".moved"
		if err := os.Rename(prepared, moved); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(prepared, payload, 0o600); err != nil {
			t.Fatal(err)
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
	err = failing.writeCurrentLocked(operation, pointer)
	operation.close()

	if !substituted {
		t.Fatal("final source-identity barrier did not run prepared-authority substitution")
	}
	if err == nil || !strings.Contains(err.Error(), "prepared authority identity changed") {
		t.Fatalf("write after prepared authority substitution error = %v, want pre-exposure identity refusal", err)
	}
	after, statErr := os.Stat(targetPath)
	if statErr != nil {
		t.Fatal(statErr)
	}
	if !os.SameFile(prior, after) {
		t.Fatal("prepared-authority substitution installed a public control before refusal")
	}
	if _, err := os.Stat(moved); err != nil {
		t.Fatalf("prepared-authority substitution removed original authority: %v", err)
	}
	recovered, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	err = recovered.Recover(context.Background())
	if err == nil || !strings.Contains(err.Error(), "publication control repair") || strings.Contains(err.Error(), "decode CURRENT") {
		t.Fatalf("Recover() after pre-exposure authority substitution error = %v, want private-authority refusal before public control decode", err)
	}
}

func TestVNextGenerationPublisherRefusesSubstitutedPreparedAuthorityBeforePublicControlTrust(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	pointer, err := baseline.Publish(vNextPublicationArtifactsForTest("active", false))
	if err != nil {
		t.Fatal(err)
	}
	substituted := false
	failing, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationBeforeControlRepairAuthorityClear || substituted {
			return nil
		}
		substituted = true
		transaction := vNextPublicationPendingControlRepairTransactionPathForTest(t, root)
		prepared := filepath.Join(transaction, vNextPublicationControlRepairPreparedFile)
		payload, err := os.ReadFile(prepared)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(prepared, prepared+".moved"); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(prepared, payload, 0o600); err != nil {
			t.Fatal(err)
		}
		vNextReplacePublicationControlForTest(t, root, vNextPublicationCurrentFile, []byte("untrusted CURRENT after authority substitution"))
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	operation, err := failing.openOperation(context.Background(), syscall.LOCK_EX, true)
	if err != nil {
		t.Fatal(err)
	}
	err = failing.writeCurrentLocked(operation, pointer)
	operation.close()
	if !substituted {
		t.Fatal("final cleanup barrier did not run authority substitution")
	}
	if err == nil || !strings.Contains(err.Error(), "prepared authority identity changed") {
		t.Fatalf("write after prepared authority substitution error = %v, want identity refusal", err)
	}
	recovered, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	err = recovered.Recover(context.Background())
	if err == nil || !strings.Contains(err.Error(), "publication control repair") || strings.Contains(err.Error(), "decode CURRENT") {
		t.Fatalf("Recover() after prepared authority substitution error = %v, want private-authority refusal before public control decode", err)
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
	pendingAuthority := cut != vNextPublicationAfterControlRepairBackupSync && cut != vNextPublicationAfterControlRepairAuthorityRetireSync && cut != vNextPublicationAfterControlRepairClearSync
	if got := vNextPublicationPendingControlRepairAuthoritiesForTest(t, root); pendingAuthority && got != 1 {
		t.Fatalf("repair authority count at %s = %d, want one pending authority", cut, got)
	} else if !pendingAuthority && got != 0 {
		t.Fatalf("repair authority count at %s = %d, want absent", cut, got)
	}

	recovered, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if err := recovered.Recover(context.Background()); err != nil {
		t.Fatalf("Recover() after %s error = %v", cut, err)
	}
	if got := vNextPublicationPendingControlRepairAuthoritiesForTest(t, root); got != 0 {
		t.Fatalf("repair authority count after recovery = %d, want absent", got)
	}
	if control.noPrior && control.target == vNextPublicationCurrentFile && (replacementPayload != nil || cut == vNextPublicationAfterControlRepairBackupSync || cut == vNextPublicationAfterControlRepairPrepared) {
		if _, err := os.Stat(filepath.Join(root, "acme", control.target)); !errors.Is(err, fs.ErrNotExist) {
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

func vNextReplacePublicationControlForTest(t *testing.T, root, name string, replacement []byte) {
	t.Helper()
	path := filepath.Join(root, "acme", name)
	if err := os.Rename(path, path+".moved"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, replacement, 0o600); err != nil {
		t.Fatal(err)
	}
}

func vNextPublicationPendingControlRepairAuthoritiesForTest(t *testing.T, root string) int {
	t.Helper()
	connectorRoot := filepath.Join(root, "acme")
	entries, err := os.ReadDir(connectorRoot)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), vNextPublicationControlRepairDirectoryPrefix) {
			continue
		}
		_, err := os.Stat(filepath.Join(connectorRoot, entry.Name(), vNextPublicationControlRepairPreparedFile))
		if err == nil {
			count++
			continue
		}
		if !errors.Is(err, fs.ErrNotExist) {
			t.Fatal(err)
		}
	}
	return count
}

func vNextPublicationPendingControlRepairTransactionPathForTest(t *testing.T, root string) string {
	t.Helper()
	connectorRoot := filepath.Join(root, "acme")
	entries, err := os.ReadDir(connectorRoot)
	if err != nil {
		t.Fatal(err)
	}
	transaction := ""
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), vNextPublicationControlRepairDirectoryPrefix) {
			continue
		}
		path := filepath.Join(connectorRoot, entry.Name())
		if _, err := os.Stat(filepath.Join(path, vNextPublicationControlRepairPreparedFile)); errors.Is(err, fs.ErrNotExist) {
			continue
		} else if err != nil {
			t.Fatal(err)
		}
		if transaction != "" {
			t.Fatalf("multiple pending control repair transactions: %q and %q", transaction, path)
		}
		transaction = path
	}
	if transaction == "" {
		t.Fatal("no pending control repair transaction")
	}
	return transaction
}

func vNextPublicationFindControlRepairReplacementForTest(t *testing.T, root string, want []byte) string {
	t.Helper()
	connectorRoot := filepath.Join(root, "acme")
	entries, err := os.ReadDir(connectorRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), vNextPublicationControlRepairDirectoryPrefix) {
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
