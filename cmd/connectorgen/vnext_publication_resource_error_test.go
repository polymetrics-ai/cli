package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"golang.org/x/sys/unix"
)

func TestVNextPublicationRemoveTreeBoundsChildDescriptorsOnReplacement(t *testing.T) {
	previousGC := debug.SetGCPercent(-1)
	defer debug.SetGCPercent(previousGC)
	root := t.TempDir()
	parent, err := vNextPublicationOpenDirectory(root, "F-02 repaired parent")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := parent.Close(); err != nil {
			t.Errorf("close F-02 repaired parent: %v", err)
		}
	})
	before := vNextPublicationNumericFDCountForTest(t)
	const attempts = 24
	fired := 0
	vNextPublicationRemoveTreeAfterIdentityForTest = func(directory *vNextPublicationDirectory, name string) {
		fired++
		originalA, err := directory.identityAt(name, "F-02 repaired A")
		if err != nil {
			t.Fatal(err)
		}
		if err := unix.Renameat(int(directory.file.Fd()), name, int(directory.file.Fd()), name+".A"); err != nil {
			t.Fatal(err)
		}
		if err := unix.Mkdirat(int(directory.file.Fd()), name, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name, "B"), []byte("replacement B"), 0o600); err != nil {
			t.Fatal(err)
		}
		replacementB, err := directory.identityAt(name, "F-02 repaired B")
		if err != nil {
			t.Fatal(err)
		}
		if originalA == replacementB || originalA.mode != unix.S_IFDIR || replacementB.mode != unix.S_IFDIR {
			t.Fatalf("invalid repaired F-02 A/B identities: A=%#v B=%#v", originalA, replacementB)
		}
	}
	t.Cleanup(func() { vNextPublicationRemoveTreeAfterIdentityForTest = nil })
	for attempt := 0; attempt < attempts; attempt++ {
		name := fmt.Sprintf("target-%02d", attempt)
		if err := unix.Mkdirat(int(parent.file.Fd()), name, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name, "A"), []byte("retained A"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := parent.removeTree(name, "F-02 repaired target"); err == nil || !strings.Contains(err.Error(), "identity changed") {
			t.Fatalf("removeTree(%q) error = %v, want identity refusal", name, err)
		}
		if got, err := os.ReadFile(filepath.Join(root, name+".A", "A")); err != nil || string(got) != "retained A" {
			t.Fatalf("retained A for %q = %q, %v", name, got, err)
		}
		if got, err := os.ReadFile(filepath.Join(root, name, "B")); err != nil || string(got) != "replacement B" {
			t.Fatalf("replacement B for %q = %q, %v", name, got, err)
		}
	}
	if fired != attempts {
		t.Fatalf("F-02 replacement seam fired %d times, want %d", fired, attempts)
	}
	if after := vNextPublicationNumericFDCountForTest(t); after != before {
		t.Fatalf("F-02 child descriptors grew despite %d identity refusals: before=%d after=%d", attempts, before, after)
	}
}

func TestVNextPublicationRemoveTreeClosesChildOnFstatError(t *testing.T) {
	previousGC := debug.SetGCPercent(-1)
	defer debug.SetGCPercent(previousGC)
	root := t.TempDir()
	parent, err := vNextPublicationOpenDirectory(root, "F-02 fstat parent")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := parent.Close(); err != nil {
			t.Errorf("close F-02 fstat parent: %v", err)
		}
	})
	if err := unix.Mkdirat(int(parent.file.Fd()), "target", 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "target", "A"), []byte("retained A"), 0o600); err != nil {
		t.Fatal(err)
	}
	before := vNextPublicationNumericFDCountForTest(t)
	injected := errors.New("injected child fstat error")
	vNextPublicationRemoveTreeChildIdentityForTest = func(file *os.File, label string) (vNextPublicationIdentity, error) {
		// Confirm the real descriptor was opened, then inject the fstat-boundary
		// failure while leaving that descriptor for removeTree's deferred owner.
		if _, err := file.Stat(); err != nil {
			return vNextPublicationIdentity{}, err
		}
		return vNextPublicationIdentity{}, injected
	}
	t.Cleanup(func() { vNextPublicationRemoveTreeChildIdentityForTest = nil })
	err = parent.removeTree("target", "F-02 fstat target")
	if !errors.Is(err, injected) {
		t.Fatalf("removeTree fstat error = %v, want injected cause", err)
	}
	if after := vNextPublicationNumericFDCountForTest(t); after != before {
		t.Fatalf("F-02 child descriptor remained open after fstat error: before=%d after=%d", before, after)
	}
	if got, err := os.ReadFile(filepath.Join(root, "target", "A")); err != nil || string(got) != "retained A" {
		t.Fatalf("fstat refusal changed target bytes: got=%q err=%v", got, err)
	}
}

func TestVNextPublicationPublishReturnsWritableCloseErrorAndFreshRecovery(t *testing.T) {
	root := t.TempDir()
	injected := errors.New("injected writable temporary close failure")
	closed := 0
	failing, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{
		CloseAtomicTemporary: func(file *os.File) error {
			closed++
			if err := file.Close(); err != nil {
				return fmt.Errorf("real temporary close: %w", err)
			}
			return injected
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := failing.Publish(vNextPublicationArtifactsForTest("close-error", true)); !errors.Is(err, injected) {
		t.Fatalf("Publish close outcome = %v, want injected cause", err)
	}
	if closed != 1 {
		t.Fatalf("real atomic temporary closes = %d, want first prepared JOURNAL close exactly once", closed)
	}
	journal, err := os.ReadFile(filepath.Join(root, "acme", vNextPublicationJournalFile))
	if err != nil || len(journal) == 0 {
		t.Fatalf("close error did not retain the prepared JOURNAL for fresh recovery: journal=%q err=%v", journal, err)
	}
	if _, err := os.Stat(filepath.Join(root, "acme", vNextPublicationCurrentFile)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("first close failure advanced CURRENT unexpectedly: %v", err)
	}
	fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.Recover(context.Background()); err != nil {
		t.Fatalf("fresh recovery after durable prepared JOURNAL: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "acme", vNextPublicationJournalFile)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("fresh recovery retained JOURNAL: %v", err)
	}
	if _, err := fresh.Publish(vNextPublicationArtifactsForTest("retry", true)); err != nil {
		t.Fatalf("normal retry after close error recovery: %v", err)
	}
}

func TestVNextPublicationWriteAtomicPreservesPrimaryAndCloseCauses(t *testing.T) {
	root := t.TempDir()
	primary := errors.New("injected journal sync failure")
	secondary := errors.New("injected writable temporary close failure")
	closed := 0
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{
		At: func(point vNextPublicationFaultPoint) error {
			if point == vNextPublicationBeforeJournalSync {
				return primary
			}
			return nil
		},
		CloseAtomicTemporary: func(file *os.File) error {
			closed++
			if err := file.Close(); err != nil {
				return err
			}
			return secondary
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = publisher.Publish(vNextPublicationArtifactsForTest("compound", true))
	if !errors.Is(err, primary) || !errors.Is(err, secondary) {
		t.Fatalf("compound atomic result = %v, want primary %v and close %v", err, primary, secondary)
	}
	if closed != 1 {
		t.Fatalf("compound path real close calls = %d, want 1", closed)
	}
}

func TestVNextPublicationAtomicCloseFailuresFollowTheirDurableCuts(t *testing.T) {
	for _, test := range []struct {
		name                  string
		closeAt               int
		journalState          string
		expectCurrent         bool
		expectNewAfterRecover bool
	}{
		{name: "prepared-journal", closeAt: 1, journalState: "prepared", expectCurrent: false, expectNewAfterRecover: false},
		{name: "current", closeAt: 2, journalState: "prepared", expectCurrent: true, expectNewAfterRecover: true},
		{name: "committed-journal", closeAt: 3, journalState: "committed", expectCurrent: true, expectNewAfterRecover: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			injected := errors.New("injected writable temporary close failure")
			closed := 0
			writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{
				CloseAtomicTemporary: func(file *os.File) error {
					closed++
					if err := file.Close(); err != nil {
						return err
					}
					if closed == test.closeAt {
						return injected
					}
					return nil
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := writer.Publish(vNextPublicationArtifactsForTest(test.name, true)); !errors.Is(err, injected) {
				t.Fatalf("Publish close outcome = %v, want injected cause", err)
			}
			if closed != test.closeAt {
				t.Fatalf("real close calls before %s cut = %d, want %d", test.name, closed, test.closeAt)
			}
			journalBytes, err := os.ReadFile(filepath.Join(root, "acme", vNextPublicationJournalFile))
			if err != nil {
				t.Fatalf("read durable %s JOURNAL: %v", test.name, err)
			}
			var journal vNextGenerationJournal
			if err := json.Unmarshal(journalBytes, &journal); err != nil || journal.State != test.journalState {
				t.Fatalf("durable journal at %s cut = %#v, decode=%v, want state %q", test.name, journal, err, test.journalState)
			}
			_, currentErr := os.Stat(filepath.Join(root, "acme", vNextPublicationCurrentFile))
			if got := currentErr == nil; got != test.expectCurrent {
				t.Fatalf("CURRENT presence at %s cut = %t (err=%v), want %t", test.name, got, currentErr, test.expectCurrent)
			}
			fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			if err := fresh.Recover(context.Background()); err != nil {
				t.Fatalf("fresh recovery at %s cut: %v", test.name, err)
			}
			if _, err := os.Stat(filepath.Join(root, "acme", vNextPublicationJournalFile)); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("fresh recovery retained JOURNAL at %s cut: %v", test.name, err)
			}
			if test.expectNewAfterRecover {
				handle, err := fresh.Open(context.Background())
				if err != nil {
					t.Fatalf("fresh Open after %s recovery: %v", test.name, err)
				}
				assertVNextPublicationMarker(t, handle, test.name)
				handle.Release()
			} else if _, err := fresh.Publish(vNextPublicationArtifactsForTest(test.name+"-retry", true)); err != nil {
				t.Fatalf("fresh retry after %s recovery: %v", test.name, err)
			}
		})
	}
}

func TestVNextPublicationRollbackRetainsValidationAndCloseCauses(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	old, err := baseline.Publish(vNextPublicationArtifactsForTest("old", true))
	if err != nil {
		t.Fatal(err)
	}
	oldCurrent, err := os.ReadFile(filepath.Join(root, "acme", vNextPublicationCurrentFile))
	if err != nil {
		t.Fatal(err)
	}
	validation := errors.New("injected active validation failure")
	closeFailure := errors.New("injected rollback CURRENT close failure")
	closed := 0
	writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{
		CloseAtomicTemporary: func(file *os.File) error {
			closed++
			if err := file.Close(); err != nil {
				return err
			}
			if closed == 3 {
				return closeFailure
			}
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	broken := vNextPublicationArtifactsForTest("rejected", true)
	validationCalls := 0
	broken.Validate = func(fs.FS) error {
		validationCalls++
		if validationCalls == 2 {
			return validation
		}
		return nil
	}
	if _, err := writer.Publish(broken); !errors.Is(err, validation) || !errors.Is(err, closeFailure) {
		t.Fatalf("rollback result = %v, want validation and close causes", err)
	}
	if closed != 3 {
		t.Fatalf("rollback close count = %d, want 3", closed)
	}
	if validationCalls != 2 {
		t.Fatalf("rollback validation calls = %d, want staged and active validation", validationCalls)
	}
	if current, err := os.ReadFile(filepath.Join(root, "acme", vNextPublicationCurrentFile)); err != nil || !bytes.Equal(current, oldCurrent) {
		t.Fatalf("rollback close outcome did not retain old CURRENT: current=%q err=%v", current, err)
	}
	fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.Recover(context.Background()); err != nil {
		t.Fatalf("fresh recovery after rollback close failure: %v", err)
	}
	handle, err := fresh.Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer handle.Release()
	if handle.pointer != old {
		t.Fatalf("recovery selected %#v, want old %#v", handle.pointer, old)
	}
	assertVNextPublicationMarker(t, handle, "old")
}

func TestVNextPublicationRecoveryCallersReportRestoreCurrentCloseError(t *testing.T) {
	for _, test := range []struct {
		name string
		call func(*vNextGenerationPublisher) error
	}{
		{name: "Recover", call: func(publisher *vNextGenerationPublisher) error { return publisher.Recover(context.Background()) }},
		{name: "Open", call: func(publisher *vNextGenerationPublisher) error {
			handle, err := publisher.Open(context.Background())
			if handle != nil {
				handle.Release()
			}
			return err
		}},
		{name: "Prune", call: func(publisher *vNextGenerationPublisher) error { return publisher.Prune(context.Background()) }},
		{name: "Publish initial recovery", call: func(publisher *vNextGenerationPublisher) error {
			_, err := publisher.Publish(vNextPublicationArtifactsForTest("retry-after-recovery-close", true))
			return err
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			old, oldCurrent, rejectedMetadataPath, rejectedMetadata := vNextPublicationPrepareRejectedNewForRecoveryCloseTest(t, root)
			injected := errors.New("injected recovery restore CURRENT close failure")
			closed := 0
			failing, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{
				CloseAtomicTemporary: func(file *os.File) error {
					closed++
					if err := file.Close(); err != nil {
						return err
					}
					return injected
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			if err := test.call(failing); !errors.Is(err, injected) {
				t.Fatalf("%s recovery close result = %v, want injected cause", test.name, err)
			}
			if closed != 1 {
				t.Fatalf("%s recovery real close calls = %d, want 1", test.name, closed)
			}
			if current, err := os.ReadFile(filepath.Join(root, "acme", vNextPublicationCurrentFile)); err != nil || !bytes.Equal(current, oldCurrent) {
				t.Fatalf("%s recovery did not durably restore old CURRENT before close error: current=%q err=%v", test.name, current, err)
			}
			if _, err := os.Stat(filepath.Join(root, "acme", vNextPublicationJournalFile)); err != nil {
				t.Fatalf("%s recovery unexpectedly cleared JOURNAL after close error: %v", test.name, err)
			}
			if _, err := os.Stat(rejectedMetadataPath); err != nil {
				t.Fatalf("%s recovery removed the unvalidated rejected generation before fixture restoration: %v", test.name, err)
			}
			if err := os.WriteFile(rejectedMetadataPath, rejectedMetadata, 0o600); err != nil {
				t.Fatalf("restore test-owned rejected generation before fresh recovery: %v", err)
			}
			fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			if err := fresh.Recover(context.Background()); err != nil {
				t.Fatalf("fresh recovery after %s close error: %v", test.name, err)
			}
			handle, err := fresh.Open(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			defer handle.Release()
			if handle.pointer != old {
				t.Fatalf("fresh recovery selected %#v, want old %#v", handle.pointer, old)
			}
			assertVNextPublicationMarker(t, handle, "old")
		})
	}
}

func vNextPublicationPrepareRejectedNewForRecoveryCloseTest(t *testing.T, root string) (vNextGenerationPointer, []byte, string, []byte) {
	t.Helper()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	old, err := baseline.Publish(vNextPublicationArtifactsForTest("old", true))
	if err != nil {
		t.Fatal(err)
	}
	oldCurrent, err := os.ReadFile(filepath.Join(root, "acme", vNextPublicationCurrentFile))
	if err != nil {
		t.Fatal(err)
	}
	crash := errors.New("interrupted after durable new CURRENT")
	interrupted, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point == vNextPublicationAfterCurrentParent {
			return crash
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	newSet := vNextPublicationArtifactsForTest("rejected", true)
	if _, err := interrupted.Publish(newSet); !errors.Is(err, crash) {
		t.Fatalf("interrupted Publish = %v, want durable CURRENT interruption", err)
	}
	newGeneration := vNextPublicationGenerationID(newSet.Files)
	metadataPath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, newGeneration, "metadata.json")
	metadata, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(metadataPath, []byte(`{"marker":"tampered"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	return old, oldCurrent, metadataPath, metadata
}

func TestRunLockRenderReportsWritableCloseFailureWithoutSuccessOutput(t *testing.T) {
	root := t.TempDir()
	lock := minimalVNextLockForTest()
	connectorRoot := filepath.Join(root, lock.Connector)
	if err := os.MkdirAll(connectorRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(connectorRoot, "source.lock.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	injected := errors.New("injected writable temporary close failure")
	closed := 0
	var stdout, stderr bytes.Buffer
	code := runLockRenderContextWithHooks(context.Background(), []string{"lock-render", lock.Connector, "--defs", root}, &stdout, &stderr, vNextPublicationHooks{
		CloseAtomicTemporary: func(file *os.File) error {
			closed++
			if err := file.Close(); err != nil {
				return err
			}
			return injected
		},
	})
	if code != 1 || closed != 1 || stdout.Len() != 0 || !strings.Contains(stderr.String(), injected.Error()) {
		t.Fatalf("lock-render writable close result: code=%d closes=%d stdout=%q stderr=%q", code, closed, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := runLockRender([]string{"lock-render", lock.Connector, "--defs", root}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "published vNext execution generation") || stderr.Len() != 0 {
		t.Fatalf("lock-render retry after close recovery: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestVNextPublicationFailedRepairCreationCleansLinkedPredecessorAnchor(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := baseline.Publish(vNextPublicationArtifactsForTest("active", true)); err != nil {
		t.Fatal(err)
	}
	before := vNextPublicationRepairDirectoriesForTest(t, root)
	injected := errors.New("injected predecessor close failure")
	closed := 0
	writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{
		CloseRepairPredecessor: func(directory *vNextPublicationDirectory) error {
			closed++
			if err := directory.Close(); err != nil {
				return err
			}
			return injected
		},
	})
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
	defer graph.close()
	head := graph.heads[vNextPublicationCurrentFile]
	if head == nil {
		t.Fatal("missing CURRENT predecessor authority")
	}
	state, err := writer.createControlRepairLocked(operation, vNextPublicationCurrentFile, head, vNextPublicationAbsentControlState(), nil, false)
	if state != nil || !errors.Is(err, injected) {
		t.Fatalf("failed repair creation = state=%v err=%v, want injected predecessor close", state, err)
	}
	if closed != 1 {
		t.Fatalf("predecessor real close calls = %d, want 1", closed)
	}
	if after := vNextPublicationRepairDirectoriesForTest(t, root); strings.Join(after, ",") != strings.Join(before, ",") {
		t.Fatalf("failed repair creation stranded linked anchor transaction: before=%v after=%v", before, after)
	}
}

func vNextPublicationNumericFDCountForTest(t *testing.T) int {
	t.Helper()
	lsof, err := exec.LookPath("lsof")
	if err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(lsof, "-n", "-P", "-p", strconv.Itoa(os.Getpid()), "-Ff").Output()
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, line := range strings.Split(string(output), "\n") {
		if len(line) > 1 && line[0] == 'f' {
			if _, err := strconv.Atoi(line[1:]); err == nil {
				count++
			}
		}
	}
	return count
}

func vNextPublicationRepairDirectoriesForTest(t *testing.T, root string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, "acme"))
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), vNextPublicationControlRepairDirectoryPrefix) {
			names = append(names, entry.Name())
		}
	}
	return names
}
