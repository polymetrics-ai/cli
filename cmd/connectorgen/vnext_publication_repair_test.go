package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestVNextGenerationPublisherBootstrapsRetainedTerminalAuthorities(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := publisher.Publish(vNextPublicationArtifactsForTest("active", false)); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	marker := filepath.Join(root, "acme", vNextPublicationControlAuthorityMarkerFile)
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("authority marker = %v", err)
	}

	operation, err := publisher.openOperation(context.Background(), syscall.LOCK_SH, false)
	if err != nil {
		t.Fatal(err)
	}
	graph, err := publisher.controlAuthorityForReadLocked(operation)
	operation.close()
	if err != nil {
		t.Fatal(err)
	}
	defer graph.close()
	if !graph.marker {
		t.Fatal("authority graph did not retain protected-mode marker")
	}
	for _, target := range []string{vNextPublicationCurrentFile, vNextPublicationJournalFile} {
		head := graph.heads[target]
		if head == nil {
			t.Fatalf("authority graph has no %s terminal head", target)
		}
		if _, _, _, terminal := head.state.terminal(); !terminal {
			t.Fatalf("authority graph %s head is not terminal", target)
		}
	}
}

func TestVNextGenerationPublisherCheckRefusesMalformedPrivateTransactionBeforePublicDecode(t *testing.T) {
	root := t.TempDir()
	artifacts := vNextPublicationArtifactsForTest("active", false)
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := publisher.Publish(artifacts); err != nil {
		t.Fatal(err)
	}
	transaction := filepath.Join(root, "acme", vNextPublicationControlRepairDirectoryPrefix+strings.Repeat("a", 32))
	if err := os.Mkdir(transaction, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := publisher.Check(artifacts); err == nil || !strings.Contains(err.Error(), "prepared authority") {
		t.Fatalf("Check() after malformed private transaction = %v, want pre-decode authority refusal", err)
	}
}

func TestRunLockRenderCheckRefusesPendingPrivateAuthorityWithoutWriting(t *testing.T) {
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
	var stdout, stderr bytes.Buffer
	if code := runLockRender([]string{"lock-render", lock.Connector, "--defs", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("runLockRender() = %d; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	baseline, err := newVNextGenerationPublisher(root, lock.Connector, vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	active, err := baseline.Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	pointer := active.pointer
	active.Release()

	crash := errors.New("crash after durable private JOURNAL authority")
	writer, err := newVNextGenerationPublisher(root, lock.Connector, vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point == vNextPublicationAfterControlRepairPrepared {
			return crash
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	operation, err := writer.openOperation(context.Background(), syscall.LOCK_EX, true)
	if err != nil {
		t.Fatal(err)
	}
	err = writer.writeJournalLocked(operation, vNextGenerationJournal{New: pointer, State: "prepared"})
	operation.close()
	if !errors.Is(err, crash) {
		t.Fatalf("writeJournalLocked() error = %v, want injected crash", err)
	}
	before := vNextPublicationTreeSnapshotForTest(t, connectorRoot)
	stdout.Reset()
	stderr.Reset()
	if code := runLockRender([]string{"lock-render", lock.Connector, "--defs", root, "--check"}, &stdout, &stderr); code != 1 {
		t.Fatalf("runLockRender(--check) = %d, want 1; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), "pending repair") {
		t.Fatalf("pending --check output: stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
	if after := vNextPublicationTreeSnapshotForTest(t, connectorRoot); !bytes.Equal(before, after) {
		t.Fatal("lock-render --check changed a pending authority tree")
	}
}

func TestRunLockRenderCheckReadsAuthorizedTerminalAuthorityWithoutWriting(t *testing.T) {
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
	var stdout, stderr bytes.Buffer
	if code := runLockRender([]string{"lock-render", lock.Connector, "--defs", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("runLockRender() = %d; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	before := vNextPublicationTreeSnapshotForTest(t, connectorRoot)
	stdout.Reset()
	stderr.Reset()
	if code := runLockRender([]string{"lock-render", lock.Connector, "--defs", root, "--check"}, &stdout, &stderr); code != 0 {
		t.Fatalf("runLockRender(--check) = %d; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("successful --check stderr = %q, want empty", stderr.String())
	}
	if after := vNextPublicationTreeSnapshotForTest(t, connectorRoot); !bytes.Equal(before, after) {
		t.Fatal("lock-render --check changed an authorized authority tree")
	}
}

func TestRunLockRenderCheckRefusesDivergentTerminalAuthorityWithoutWriting(t *testing.T) {
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
	var stdout, stderr bytes.Buffer
	if code := runLockRender([]string{"lock-render", lock.Connector, "--defs", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("runLockRender() = %d; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	current := filepath.Join(connectorRoot, vNextPublicationCurrentFile)
	payload, err := os.ReadFile(current)
	if err != nil {
		t.Fatal(err)
	}
	vNextReplacePublicationControlForTest(t, root, vNextPublicationCurrentFile, payload)
	before := vNextPublicationTreeSnapshotForTest(t, connectorRoot)
	stdout.Reset()
	stderr.Reset()
	if code := runLockRender([]string{"lock-render", lock.Connector, "--defs", root, "--check"}, &stdout, &stderr); code != 1 {
		t.Fatalf("runLockRender(--check) = %d, want 1; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), "diverges from terminal authority") {
		t.Fatalf("divergent --check output: stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
	if after := vNextPublicationTreeSnapshotForTest(t, connectorRoot); !bytes.Equal(before, after) {
		t.Fatal("lock-render --check changed a divergent authority tree")
	}
}

func TestVNextGenerationPublisherRecoversTerminalDivergenceThroughSuccessor(t *testing.T) {
	root := t.TempDir()
	artifacts := vNextPublicationArtifactsForTest("active", false)
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	pointer, err := baseline.Publish(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "acme", vNextPublicationCurrentFile)
	prior, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	vNextReplacePublicationControlForTest(t, root, vNextPublicationCurrentFile, []byte(`{"generation":"`+strings.Repeat("f", 64)+`","integrity_digest":"`+strings.Repeat("e", 64)+`"}`+"\n"))

	fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.Check(artifacts); err == nil || !strings.Contains(err.Error(), "diverges from terminal authority") {
		t.Fatalf("Check() after terminal divergence = %v, want authority mismatch", err)
	}
	if err := fresh.Recover(context.Background()); err != nil {
		t.Fatalf("Recover() after terminal divergence = %v", err)
	}
	if got, err := os.ReadFile(target); err != nil || !bytes.Equal(got, prior) {
		t.Fatalf("CURRENT after successor recovery = %q, %v; want terminal %q", got, err, prior)
	}
	current, found, err := fresh.readCurrentForTest(pointer)
	if err != nil || !found || current != pointer {
		t.Fatalf("recovered CURRENT = %#v, found=%t, err=%v; want %#v", current, found, err, pointer)
	}
}

func TestVNextGenerationPublisherRemoveControlCapturesLateOccupantInsteadOfUnlinking(t *testing.T) {
	tests := []struct {
		name   string
		target string
		action string
	}{
		{name: "CURRENT/replacement", target: vNextPublicationCurrentFile, action: "replace"},
		{name: "CURRENT/unlink", target: vNextPublicationCurrentFile, action: "unlink"},
		{name: "JOURNAL/replacement", target: vNextPublicationJournalFile, action: "replace"},
		{name: "JOURNAL/unlink", target: vNextPublicationJournalFile, action: "unlink"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			artifacts := vNextPublicationArtifactsForTest("active", false)
			baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			pointer, err := baseline.Publish(artifacts)
			if err != nil {
				t.Fatal(err)
			}
			if test.target == vNextPublicationJournalFile {
				operation, openErr := baseline.openOperation(context.Background(), syscall.LOCK_EX, true)
				if openErr != nil {
					t.Fatal(openErr)
				}
				err = baseline.writeJournalLocked(operation, vNextGenerationJournal{New: pointer, State: "prepared"})
				operation.close()
				if err != nil {
					t.Fatal(err)
				}
			}
			target := filepath.Join(root, "acme", test.target)
			prior, err := os.ReadFile(target)
			if err != nil {
				t.Fatal(err)
			}
			replacement := []byte("late occupant before logical absence/" + test.name)
			attacked := false
			writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
				if point != vNextPublicationBeforeControlRepairCapture || attacked {
					return nil
				}
				attacked = true
				switch test.action {
				case "replace":
					vNextReplacePublicationControlForTest(t, root, test.target, replacement)
				case "unlink":
					if err := os.Remove(target); err != nil {
						t.Fatal(err)
					}
				default:
					t.Fatalf("unknown actor action %q", test.action)
				}
				return nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			operation, err := writer.openOperation(context.Background(), syscall.LOCK_EX, true)
			if err != nil {
				t.Fatal(err)
			}
			err = writer.removeControlLocked(operation, test.target)
			operation.close()
			if !attacked {
				t.Fatal("removeControlLocked() did not reach capture boundary")
			}
			if !errors.Is(err, errVNextPublicationControlConflict) {
				t.Fatalf("removeControlLocked() error = %v, want bounded conflict", err)
			}
			if got, readErr := os.ReadFile(target); readErr != nil || !bytes.Equal(got, prior) {
				t.Fatalf("%s after rejected removal = %q, %v; want prior %q", test.target, got, readErr, prior)
			}
			if test.action == "replace" {
				vNextPublicationFindPrivatePayloadForTest(t, root, replacement)
			}
			fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			if err := fresh.Recover(context.Background()); err != nil {
				t.Fatalf("fresh Recover() after %s = %v", test.name, err)
			}
			vNextPublicationAssertTerminalAuthoritiesMatchPublicForTest(t, fresh)
		})
	}
}

func TestVNextGenerationPublisherRejectsSubstitutedIntendedSourceWithoutOrphaningAuthority(t *testing.T) {
	root := t.TempDir()
	artifacts := vNextPublicationArtifactsForTest("active", false)
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	pointer, err := baseline.Publish(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	swapped := false
	writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point == vNextPublicationAfterFinalControlSourceIdentity && !swapped {
			swapped = true
			vNextReplacePublicationTemporaryForTest(t, root, []byte("substituted intended source"))
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	operation, err := writer.openOperation(context.Background(), syscall.LOCK_EX, true)
	if err != nil {
		t.Fatal(err)
	}
	err = writer.writeCurrentLocked(operation, pointer)
	operation.close()
	if !swapped {
		t.Fatal("writer did not reach final intended-source boundary")
	}
	if err == nil || !strings.Contains(err.Error(), "intended anchor identity changed") {
		t.Fatalf("writeCurrentLocked() error = %v, want intended-source identity refusal", err)
	}
	transactions, err := filepath.Glob(filepath.Join(root, "acme", vNextPublicationControlRepairDirectoryPrefix+"*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) == 0 {
		t.Fatal("rejected source removed every retained authority head")
	}
	for _, transaction := range transactions {
		if _, statErr := os.Stat(filepath.Join(transaction, vNextPublicationControlRepairPreparedFile)); statErr != nil {
			t.Fatalf("unprepared transaction %q survived rejected source: %v", transaction, statErr)
		}
	}
	fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.Recover(context.Background()); err != nil {
		t.Fatalf("Recover() after rejected source = %v", err)
	}
}
func TestVNextPublicationNoReplaceUnsupportedFilesystemIsTyped(t *testing.T) {
	for _, cause := range []error{unix.ENOSYS, unix.EINVAL, unix.ENOTSUP, unix.EOPNOTSUPP, unix.EXDEV} {
		err := vNextPublicationNoReplaceRenameError(cause)
		if !errors.Is(err, errVNextPublicationUnsupportedNoReplace) || !errors.Is(err, cause) {
			t.Fatalf("vNextPublicationNoReplaceRenameError(%v) = %v, want typed unsupported transition retaining the syscall cause", cause, err)
		}
	}
}

func TestVNextGenerationPublisherNoReplaceFailureRetainsAuthorityAndPublicControl(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		write       func(*vNextGenerationPublisher, *vNextPublicationOperation, vNextGenerationPointer) error
		cause       error
		unsupported bool
	}{
		{
			name:   "CURRENT/unsupported",
			target: vNextPublicationCurrentFile,
			write: func(publisher *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
				return publisher.writeCurrentLocked(operation, pointer)
			},
			cause:       unix.ENOSYS,
			unsupported: true,
		},
		{
			name:   "JOURNAL/unsupported",
			target: vNextPublicationJournalFile,
			write: func(publisher *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
				return publisher.writeJournalLocked(operation, vNextGenerationJournal{New: pointer, State: "prepared"})
			},
			cause:       unix.EXDEV,
			unsupported: true,
		},
		{
			name:   "CURRENT/EINVAL",
			target: vNextPublicationCurrentFile,
			write: func(publisher *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
				return publisher.writeCurrentLocked(operation, pointer)
			},
			cause:       unix.EINVAL,
			unsupported: true,
		},
		{
			name:   "JOURNAL/ENOTSUP",
			target: vNextPublicationJournalFile,
			write: func(publisher *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
				return publisher.writeJournalLocked(operation, vNextGenerationJournal{New: pointer, State: "prepared"})
			},
			cause:       unix.ENOTSUP,
			unsupported: true,
		},
		{
			name:   "CURRENT/EOPNOTSUPP",
			target: vNextPublicationCurrentFile,
			write: func(publisher *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
				return publisher.writeCurrentLocked(operation, pointer)
			},
			cause:       unix.EOPNOTSUPP,
			unsupported: true,
		},
		{
			name:   "CURRENT/collision",
			target: vNextPublicationCurrentFile,
			write: func(publisher *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
				return publisher.writeCurrentLocked(operation, pointer)
			},
			cause: fs.ErrExist,
		},
	}
	for _, test := range tests {
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
			targetPath := filepath.Join(root, "acme", test.target)
			before, err := os.ReadFile(targetPath)
			if test.target == vNextPublicationJournalFile && errors.Is(err, fs.ErrNotExist) {
				before = nil
			} else if err != nil {
				t.Fatal(err)
			}

			writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{
				ControlCaptureRename: func() error { return test.cause },
			})
			if err != nil {
				t.Fatal(err)
			}
			operation, err := writer.openOperation(context.Background(), syscall.LOCK_EX, true)
			if err != nil {
				t.Fatal(err)
			}
			err = test.write(writer, operation, pointer)
			operation.close()
			if !errors.Is(err, test.cause) {
				t.Fatalf("control transition error = %v, want %v", err, test.cause)
			}
			if test.unsupported != errors.Is(err, errVNextPublicationUnsupportedNoReplace) {
				t.Fatalf("control transition unsupported classification = %t, want %t: %v", errors.Is(err, errVNextPublicationUnsupportedNoReplace), test.unsupported, err)
			}
			if before == nil {
				if _, statErr := os.Lstat(targetPath); !errors.Is(statErr, fs.ErrNotExist) {
					t.Fatalf("failed no-replace %s created public control: %v", test.target, statErr)
				}
			} else if after, readErr := os.ReadFile(targetPath); readErr != nil || !bytes.Equal(after, before) {
				t.Fatalf("failed no-replace %s changed public control: err=%v got=%q want=%q", test.target, readErr, after, before)
			}

			inspection, err := writer.openOperation(context.Background(), syscall.LOCK_SH, false)
			if err != nil {
				t.Fatal(err)
			}
			graph, err := writer.scanControlAuthorityLocked(inspection)
			inspection.close()
			if err != nil {
				t.Fatal(err)
			}
			defer graph.close()
			head := graph.heads[test.target]
			if head == nil || head.state.latestPhase() != vNextPublicationControlRepairCaptureIntent {
				t.Fatalf("failed no-replace %s head = %#v, want pending capture intent", test.target, head)
			}
			if head.state.record.Predecessor == nil {
				t.Fatalf("failed no-replace %s lost predecessor authority", test.target)
			}
			predecessor := graph.states[head.state.record.Predecessor.Transaction]
			if predecessor == nil {
				t.Fatalf("failed no-replace %s predecessor transaction is absent", test.target)
			}
			if _, _, _, terminal := predecessor.terminal(); !terminal {
				t.Fatalf("failed no-replace %s predecessor is not terminal", test.target)
			}
		})
	}
}

func TestVNextGenerationPublisherNoReplaceSourceAbsentDoesNotClobber(t *testing.T) {
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
	before, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{
		ControlCaptureRename: func() error { return fs.ErrNotExist },
	})
	if err != nil {
		t.Fatal(err)
	}
	operation, err := writer.openOperation(context.Background(), syscall.LOCK_EX, true)
	if err != nil {
		t.Fatal(err)
	}
	err = writer.writeCurrentLocked(operation, pointer)
	operation.close()
	if !errors.Is(err, errVNextPublicationControlConflict) {
		t.Fatalf("source-absent transition error = %v, want bounded conflict", err)
	}
	if after, readErr := os.ReadFile(targetPath); readErr != nil || !bytes.Equal(after, before) {
		t.Fatalf("source-absent transition changed CURRENT: err=%v got=%q want=%q", readErr, after, before)
	}
	fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.Recover(context.Background()); err != nil {
		t.Fatalf("fresh Recover() after source-absent transition = %v", err)
	}
	vNextPublicationAssertTerminalAuthoritiesMatchPublicForTest(t, fresh)
}

func TestVNextGenerationPublisherTerminalAuthorityTransitionMatrix(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		priorPresent bool
		operation    string
		interfere    bool
		wantPresent  bool
		wantOutcome  string
	}{
		{
			name:         "CURRENT/install-intended",
			target:       vNextPublicationCurrentFile,
			priorPresent: true,
			operation:    "install",
			wantPresent:  true,
			wantOutcome:  vNextPublicationControlRepairCommitted,
		},
		{
			name:        "CURRENT/install-without-prior",
			target:      vNextPublicationCurrentFile,
			operation:   "install",
			wantPresent: true,
			wantOutcome: vNextPublicationControlRepairCommitted,
		},
		{
			name:        "CURRENT/restore-prior-absence",
			target:      vNextPublicationCurrentFile,
			operation:   "install",
			interfere:   true,
			wantOutcome: vNextPublicationControlRepairRolledBack,
		},
		{
			name:         "CURRENT/restore-prior",
			target:       vNextPublicationCurrentFile,
			priorPresent: true,
			operation:    "install",
			interfere:    true,
			wantPresent:  true,
			wantOutcome:  vNextPublicationControlRepairRolledBack,
		},
		{
			name:         "CURRENT/retire-to-absence",
			target:       vNextPublicationCurrentFile,
			priorPresent: true,
			operation:    "remove",
			wantOutcome:  vNextPublicationControlRepairCommitted,
		},
		{
			name:        "JOURNAL/install-without-prior",
			target:      vNextPublicationJournalFile,
			operation:   "install",
			wantPresent: true,
			wantOutcome: vNextPublicationControlRepairCommitted,
		},
		{
			name:         "JOURNAL/replace-prior",
			target:       vNextPublicationJournalFile,
			priorPresent: true,
			operation:    "install",
			wantPresent:  true,
			wantOutcome:  vNextPublicationControlRepairCommitted,
		},
		{
			name:        "JOURNAL/restore-prior-absence",
			target:      vNextPublicationJournalFile,
			operation:   "install",
			interfere:   true,
			wantOutcome: vNextPublicationControlRepairRolledBack,
		},
		{
			name:         "JOURNAL/retire-to-absence",
			target:       vNextPublicationJournalFile,
			priorPresent: true,
			operation:    "remove",
			wantOutcome:  vNextPublicationControlRepairCommitted,
		},
	}
	for _, test := range tests {
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
			if test.target == vNextPublicationJournalFile && test.priorPresent {
				operation, openErr := baseline.openOperation(context.Background(), syscall.LOCK_EX, true)
				if openErr != nil {
					t.Fatal(openErr)
				}
				err = baseline.writeJournalLocked(operation, vNextGenerationJournal{New: pointer, State: "prepared"})
				operation.close()
				if err != nil {
					t.Fatalf("write initial JOURNAL: %v", err)
				}
			}
			if test.target == vNextPublicationCurrentFile && !test.priorPresent {
				operation, openErr := baseline.openOperation(context.Background(), syscall.LOCK_EX, true)
				if openErr != nil {
					t.Fatal(openErr)
				}
				err = baseline.removeCurrentLocked(operation)
				operation.close()
				if err != nil {
					t.Fatalf("remove initial CURRENT: %v", err)
				}
			}

			targetPath := filepath.Join(root, "acme", test.target)
			before, readErr := os.ReadFile(targetPath)
			if !test.priorPresent {
				if !errors.Is(readErr, fs.ErrNotExist) {
					t.Fatalf("initial absent %s read error = %v", test.target, readErr)
				}
			} else if readErr != nil {
				t.Fatal(readErr)
			}
			replacement := []byte("unclassified public replacement for " + test.name)
			interfered := false
			writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
				if point != vNextPublicationBeforeControlRepairCapture || !test.interfere || interfered {
					return nil
				}
				interfered = true
				if test.priorPresent {
					vNextReplacePublicationControlForTest(t, root, test.target, replacement)
					return nil
				}
				if err := os.WriteFile(targetPath, replacement, 0o600); err != nil {
					t.Fatal(err)
				}
				return nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			operation, err := writer.openOperation(context.Background(), syscall.LOCK_EX, true)
			if err != nil {
				t.Fatal(err)
			}
			switch test.operation {
			case "install":
				if test.target == vNextPublicationCurrentFile {
					err = writer.writeCurrentLocked(operation, pointer)
				} else {
					err = writer.writeJournalLocked(operation, vNextGenerationJournal{New: pointer, State: "prepared"})
				}
			case "remove":
				err = writer.removeControlLocked(operation, test.target)
			default:
				t.Fatalf("unknown operation %q", test.operation)
			}
			operation.close()
			if test.interfere {
				if !interfered {
					t.Fatal("lock-ignoring actor did not reach the capture barrier")
				}
				if !errors.Is(err, errVNextPublicationControlConflict) {
					t.Fatalf("interfered %s transition error = %v, want conflict", test.target, err)
				}
				if got := vNextPublicationFindPrivatePayloadForTest(t, root, replacement); got == "" {
					t.Fatalf("interfered %s replacement was not retained", test.target)
				}
			} else if err != nil {
				t.Fatalf("%s transition error = %v", test.target, err)
			}

			inspection, err := writer.openOperation(context.Background(), syscall.LOCK_SH, false)
			if err != nil {
				t.Fatal(err)
			}
			graph, err := writer.controlAuthorityForReadLocked(inspection)
			if err != nil {
				t.Fatal(err)
			}
			head := graph.heads[test.target]
			if head == nil {
				graph.close()
				t.Fatalf("%s has no terminal authority head", test.target)
			}
			if head.outcome != test.wantOutcome || head.selected.Present != test.wantPresent {
				graph.close()
				t.Fatalf("%s terminal = outcome %q present %t, want outcome %q present %t", test.target, head.outcome, head.selected.Present, test.wantOutcome, test.wantPresent)
			}
			if test.wantPresent {
				_, found, identity, controlErr := vNextPublicationReadControlBound(inspection.connector, test.target, "public control")
				expected, _, identityErr := head.selected.identity(head.selected.Member)
				if controlErr != nil || identityErr != nil || !found || identity != expected {
					graph.close()
					t.Fatalf("%s terminal/public identity = found %t actual %#v expected %#v control=%v identity=%v", test.target, found, identity, expected, controlErr, identityErr)
				}
			} else if _, found, _, controlErr := vNextPublicationReadControlBound(inspection.connector, test.target, "public control"); controlErr != nil || found {
				graph.close()
				t.Fatalf("%s terminal/public absence = found %t error=%v", test.target, found, controlErr)
			}
			graph.close()
			inspection.close()
			if test.interfere && test.priorPresent {
				if after, err := os.ReadFile(targetPath); err != nil || !bytes.Equal(after, before) {
					t.Fatalf("rollback %s did not restore prior bytes: err=%v got=%q want=%q", test.target, err, after, before)
				}
			}

			fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			if err := fresh.Recover(context.Background()); err != nil {
				t.Fatalf("fresh Recover() after %s = %v", test.name, err)
			}
			freshInspection, err := fresh.openOperation(context.Background(), syscall.LOCK_SH, false)
			if err != nil {
				t.Fatal(err)
			}
			freshGraph, err := fresh.controlAuthorityForReadLocked(freshInspection)
			freshInspection.close()
			if err != nil {
				t.Fatal(err)
			}
			defer freshGraph.close()
			for _, target := range []string{vNextPublicationCurrentFile, vNextPublicationJournalFile} {
				head := freshGraph.heads[target]
				if head == nil || head.outcome == vNextPublicationControlRepairRetryRequired {
					t.Fatalf("fresh recovery %s head = %#v", target, head)
				}
				if _, _, _, terminal := head.state.terminal(); !terminal {
					t.Fatalf("fresh recovery %s head is not terminal", target)
				}
			}
		})
	}
}

func TestVNextGenerationPublisherRecoversEveryTerminalAuthorityDurableCut(t *testing.T) {
	points := []vNextPublicationFaultPoint{
		vNextPublicationAfterControlRepairPrepared,
		vNextPublicationAfterControlRepairCaptureDirectory,
		vNextPublicationBeforeControlRepairCapture,
		vNextPublicationAfterControlRepairCaptureRename,
		vNextPublicationAfterControlRepairCaptureDirectorySync,
		vNextPublicationAfterControlRepairCaptureRootSync,
		vNextPublicationAfterControlRepairCaptured,
		vNextPublicationAfterControlRepairInstall,
		vNextPublicationAfterControlRepairInstallSync,
		vNextPublicationAfterControlRepairSelected,
		vNextPublicationAfterFinalControlRepairValidation,
	}
	scenarios := []struct {
		name         string
		target       string
		priorPresent bool
	}{
		{name: "CURRENT/prior-present", target: vNextPublicationCurrentFile, priorPresent: true},
		{name: "CURRENT/prior-absent", target: vNextPublicationCurrentFile},
		{name: "JOURNAL/prior-present", target: vNextPublicationJournalFile, priorPresent: true},
		{name: "JOURNAL/prior-absent", target: vNextPublicationJournalFile},
	}
	for _, scenario := range scenarios {
		for _, point := range points {
			t.Run(scenario.name+"/"+string(point), func(t *testing.T) {
				root := t.TempDir()
				pointer := vNextPublicationPrepareTerminalAuthorityScenarioForTest(t, root, scenario.target, scenario.priorPresent)
				crash := errors.New("crash at durable terminal authority cut")
				failing, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(candidate vNextPublicationFaultPoint) error {
					if candidate == point {
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
				if scenario.target == vNextPublicationCurrentFile {
					err = failing.writeCurrentLocked(operation, pointer)
				} else {
					err = failing.writeJournalLocked(operation, vNextGenerationJournal{New: pointer, State: "prepared"})
				}
				operation.close()
				if !errors.Is(err, crash) {
					t.Fatalf("transition error = %v, want crash at %s", err, point)
				}

				fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
				if err != nil {
					t.Fatal(err)
				}
				if err := fresh.Recover(context.Background()); err != nil {
					t.Fatalf("fresh Recover() after %s = %v", point, err)
				}
				vNextPublicationAssertTerminalAuthoritiesMatchPublicForTest(t, fresh)
			})
		}
	}
}

func vNextPublicationPrepareTerminalAuthorityScenarioForTest(t *testing.T, root, target string, priorPresent bool) vNextGenerationPointer {
	t.Helper()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	pointer, err := publisher.Publish(vNextPublicationArtifactsForTest("active", false))
	if err != nil {
		t.Fatal(err)
	}
	operation, err := publisher.openOperation(context.Background(), syscall.LOCK_EX, true)
	if err != nil {
		t.Fatal(err)
	}
	switch {
	case target == vNextPublicationCurrentFile && !priorPresent:
		err = publisher.removeCurrentLocked(operation)
	case target == vNextPublicationJournalFile && priorPresent:
		err = publisher.writeJournalLocked(operation, vNextGenerationJournal{New: pointer, State: "prepared"})
	}
	operation.close()
	if err != nil {
		t.Fatal(err)
	}
	return pointer
}

func vNextPublicationAssertTerminalAuthoritiesMatchPublicForTest(t *testing.T, publisher *vNextGenerationPublisher) {
	t.Helper()
	operation, err := publisher.openOperation(context.Background(), syscall.LOCK_SH, false)
	if err != nil {
		t.Fatal(err)
	}
	defer operation.close()
	graph, err := publisher.controlAuthorityForReadLocked(operation)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.close()
	for _, target := range []string{vNextPublicationCurrentFile, vNextPublicationJournalFile} {
		head := graph.heads[target]
		if head == nil || head.outcome == vNextPublicationControlRepairRetryRequired {
			t.Fatalf("%s authority head = %#v", target, head)
		}
		if _, _, _, terminal := head.state.terminal(); !terminal {
			t.Fatalf("%s authority head is not terminal", target)
		}
		if _, _, _, err := vNextPublicationReadAuthorizedControlLocked(operation, graph, target, target); err != nil {
			t.Fatalf("%s authority/public read = %v", target, err)
		}
	}
}

func TestVNextGenerationPublisherCheckRefusesAuthorityTopologyBeforePublicDecode(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		inject   func(*testing.T, *vNextGenerationPublisher, string)
	}{
		{
			name:     "phase-gap",
			expected: "phase chain has a gap",
			inject: func(t *testing.T, publisher *vNextGenerationPublisher, root string) {
				transaction := vNextPublicationAuthorityTransactionForTest(t, publisher, vNextPublicationCurrentFile)
				missing := vNextPublicationFirstMissingControlPhaseForTest(t, filepath.Join(root, "acme", transaction))
				if missing == vNextPublicationControlRepairMaxPhases {
					t.Fatal("authority transaction has no room for a gapped phase")
				}
				if err := os.WriteFile(filepath.Join(root, "acme", transaction, vNextPublicationControlRepairPhaseName(missing+1)), []byte("{}\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:     "orphan-capture",
			expected: "unreferenced publication control capture",
			inject: func(t *testing.T, publisher *vNextGenerationPublisher, root string) {
				transaction := vNextPublicationAuthorityTransactionForTest(t, publisher, vNextPublicationCurrentFile)
				capture := vNextPublicationFirstMissingControlCaptureForTest(t, filepath.Join(root, "acme", transaction))
				if err := os.Mkdir(filepath.Join(root, "acme", transaction, vNextPublicationControlCaptureName(capture)), 0o700); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:     "fork",
			expected: "has multiple successors",
			inject: func(t *testing.T, publisher *vNextGenerationPublisher, _ string) {
				operation, err := publisher.openOperation(context.Background(), syscall.LOCK_EX, true)
				if err != nil {
					t.Fatal(err)
				}
				graph, err := publisher.scanControlAuthorityLocked(operation)
				if err != nil {
					operation.close()
					t.Fatal(err)
				}
				head := graph.heads[vNextPublicationCurrentFile]
				if head == nil {
					graph.close()
					operation.close()
					t.Fatal("CURRENT has no authority head")
				}
				sourceDirectory, err := head.state.openTransaction(operation)
				if err != nil {
					graph.close()
					operation.close()
					t.Fatal(err)
				}
				identity, present, err := head.state.anchor(sourceDirectory, head.selected, head.selected.Member)
				if err != nil || !present {
					_ = sourceDirectory.Close()
					graph.close()
					operation.close()
					t.Fatalf("CURRENT terminal anchor = identity %#v present %t error %v", identity, present, err)
				}
				intended := vNextPublicationControlStateWithMember(head.selected, vNextPublicationControlReplacementMember)
				source := &vNextPublicationControlAnchorSource{directory: sourceDirectory, name: head.selected.Member, identity: identity}
				first, err := publisher.createControlRepairLocked(operation, vNextPublicationCurrentFile, head, intended, source, false)
				if err == nil {
					first.close()
				}
				if err == nil {
					second, secondErr := publisher.createControlRepairLocked(operation, vNextPublicationCurrentFile, head, intended, source, false)
					if second != nil {
						second.close()
					}
					err = secondErr
				}
				if closeErr := sourceDirectory.Close(); closeErr != nil && err == nil {
					err = closeErr
				}
				graph.close()
				operation.close()
				if err != nil {
					t.Fatal(err)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			artifacts := vNextPublicationArtifactsForTest("active", false)
			publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := publisher.Publish(artifacts); err != nil {
				t.Fatal(err)
			}
			test.inject(t, publisher, root)
			if err := os.WriteFile(filepath.Join(root, "acme", vNextPublicationCurrentFile), []byte("not a CURRENT document\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			before := vNextPublicationTreeSnapshotForTest(t, filepath.Join(root, "acme"))
			err = publisher.Check(artifacts)
			if err == nil || !strings.Contains(err.Error(), test.expected) {
				t.Fatalf("Check() error = %v, want authority topology %q", err, test.expected)
			}
			if strings.Contains(err.Error(), "decode CURRENT") {
				t.Fatalf("Check() decoded public CURRENT before authority topology refusal: %v", err)
			}
			if after := vNextPublicationTreeSnapshotForTest(t, filepath.Join(root, "acme")); !bytes.Equal(before, after) {
				t.Fatal("Check() changed malformed authority topology")
			}
		})
	}
}

func vNextPublicationAuthorityTransactionForTest(t *testing.T, publisher *vNextGenerationPublisher, target string) string {
	t.Helper()
	operation, err := publisher.openOperation(context.Background(), syscall.LOCK_SH, false)
	if err != nil {
		t.Fatal(err)
	}
	defer operation.close()
	graph, err := publisher.controlAuthorityForReadLocked(operation)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.close()
	head := graph.heads[target]
	if head == nil {
		t.Fatalf("authority graph has no %s head", target)
	}
	return head.state.transactionName
}

func vNextPublicationFirstMissingControlPhaseForTest(t *testing.T, transaction string) int {
	t.Helper()
	for sequence := 1; sequence <= vNextPublicationControlRepairMaxPhases; sequence++ {
		_, err := os.Lstat(filepath.Join(transaction, vNextPublicationControlRepairPhaseName(sequence)))
		if errors.Is(err, fs.ErrNotExist) {
			return sequence
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Fatal("authority transaction has no missing phase")
	return 0
}

func vNextPublicationFirstMissingControlCaptureForTest(t *testing.T, transaction string) int {
	t.Helper()
	for attempt := 1; attempt <= vNextPublicationControlRepairMaxCaptureAttempts; attempt++ {
		_, err := os.Lstat(filepath.Join(transaction, vNextPublicationControlCaptureName(attempt)))
		if errors.Is(err, fs.ErrNotExist) {
			return attempt
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Fatal("authority transaction has no missing capture")
	return 0
}

func (p *vNextGenerationPublisher) readCurrentForTest(want vNextGenerationPointer) (vNextGenerationPointer, bool, error) {
	operation, err := p.openOperation(context.Background(), syscall.LOCK_SH, false)
	if err != nil {
		return vNextGenerationPointer{}, false, err
	}
	defer operation.close()
	return p.readCurrentLocked(operation)
}

func TestVNextGenerationPublisherRepeatedSubstitutionsRemainForensic(t *testing.T) {
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
	replacements := [][]byte{
		[]byte("first unclassified replacement"),
		[]byte("second unclassified replacement"),
		[]byte("third unclassified replacement"),
	}
	stage := 0
	writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		switch {
		case point == vNextPublicationBeforeControlRepairCapture && stage == 0:
			stage++
			vNextReplacePublicationControlForTest(t, root, vNextPublicationCurrentFile, replacements[0])
		case point == vNextPublicationAfterControlRepairCaptured && stage == 1:
			stage++
			if err := os.WriteFile(targetPath, replacements[1], 0o600); err != nil {
				t.Fatal(err)
			}
		case point == vNextPublicationAfterControlRepairSelected && stage == 2:
			stage++
			if err := os.Rename(targetPath, targetPath+".third-moved"); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(targetPath, replacements[2], 0o600); err != nil {
				t.Fatal(err)
			}
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	operation, err := writer.openOperation(context.Background(), syscall.LOCK_EX, true)
	if err != nil {
		t.Fatal(err)
	}
	err = writer.writeCurrentLocked(operation, pointer)
	operation.close()
	if !errors.Is(err, errVNextPublicationControlConflict) {
		t.Fatalf("writeCurrentLocked() error = %v, want bounded conflict", err)
	}
	if stage != 3 {
		t.Fatalf("substitution stage = %d, want all three barriers", stage)
	}

	fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.Recover(context.Background()); err != nil {
		t.Fatalf("fresh Recover() = %v", err)
	}
	paths := make([]string, len(replacements))
	for index, replacement := range replacements {
		paths[index] = vNextPublicationFindPrivatePayloadForTest(t, root, replacement)
	}
	for left := range paths {
		leftInfo, err := os.Stat(paths[left])
		if err != nil {
			t.Fatal(err)
		}
		for right := left + 1; right < len(paths); right++ {
			rightInfo, err := os.Stat(paths[right])
			if err != nil {
				t.Fatal(err)
			}
			if os.SameFile(leftInfo, rightInfo) {
				t.Fatalf("replacements %d and %d share retained inode", left, right)
			}
		}
	}

	inspection, err := fresh.openOperation(context.Background(), syscall.LOCK_SH, false)
	if err != nil {
		t.Fatal(err)
	}
	graph, err := fresh.controlAuthorityForReadLocked(inspection)
	if err != nil {
		inspection.close()
		t.Fatal(err)
	}
	defer graph.close()
	defer inspection.close()
	head := graph.heads[vNextPublicationCurrentFile]
	if head == nil || head.outcome == vNextPublicationControlRepairRetryRequired {
		t.Fatalf("CURRENT terminal head after recovery = %#v", head)
	}
	_, found, identity, err := vNextPublicationReadAuthorizedControlLocked(inspection, graph, vNextPublicationCurrentFile, "CURRENT")
	if err != nil || !found {
		t.Fatalf("authorized CURRENT after recovery = found %t identity %#v error %v", found, identity, err)
	}
	for _, state := range graph.states {
		if state.record.Target != vNextPublicationCurrentFile {
			continue
		}
		transaction, err := state.openTransaction(inspection)
		if err != nil {
			t.Fatalf("open CURRENT transaction %q: %v", state.transactionName, err)
		}
		err = state.validateAnchors(transaction)
		closeErr := transaction.Close()
		if err != nil {
			t.Fatalf("CURRENT transaction %q anchors = %v", state.transactionName, err)
		}
		if closeErr != nil {
			t.Fatalf("close CURRENT transaction %q: %v", state.transactionName, closeErr)
		}
		captures := 0
		for _, phase := range state.phases {
			if phase.record.State == vNextPublicationControlRepairCaptureIntent {
				captures++
			}
		}
		if captures > vNextPublicationControlRepairMaxCaptureAttempts {
			t.Fatalf("CURRENT transaction %q used %d capture attempts", state.transactionName, captures)
		}
	}
}

func TestVNextGenerationPublisherRefusesPrivateAuthorityReplacementBeforePublicDecode(t *testing.T) {
	tests := []struct {
		name     string
		point    vNextPublicationFaultPoint
		expected string
		replace  func(*testing.T, string, string)
	}{
		{
			name:     "transaction",
			point:    vNextPublicationAfterControlRepairCaptureDirectory,
			expected: "transaction identity changed",
			replace: func(t *testing.T, transaction, _ string) {
				vNextPublicationReplacePrivateDirectoryForTest(t, transaction)
			},
		},
		{
			name:     "prepared",
			point:    vNextPublicationAfterControlRepairCaptureDirectory,
			expected: "prepared authority identity changed",
			replace: func(t *testing.T, transaction, _ string) {
				vNextPublicationReplacePrivateFileForTest(t, filepath.Join(transaction, vNextPublicationControlRepairPreparedFile))
			},
		},
		{
			name:     "capture",
			point:    vNextPublicationAfterControlRepairCaptureDirectory,
			expected: "capture identity changed",
			replace: func(t *testing.T, transaction, _ string) {
				vNextPublicationReplacePrivateDirectoryForTest(t, filepath.Join(transaction, vNextPublicationControlCaptureName(1)))
			},
		},
		{
			name:     "phase",
			point:    vNextPublicationAfterControlRepairCaptured,
			expected: "phase",
			replace: func(t *testing.T, transaction, _ string) {
				vNextPublicationReplacePrivateFileForTest(t, filepath.Join(transaction, vNextPublicationControlRepairPhaseName(2)))
			},
		},
		{
			name:     "predecessor",
			point:    vNextPublicationAfterControlRepairCaptureDirectory,
			expected: "predecessor transaction",
			replace: func(t *testing.T, _, predecessor string) {
				vNextPublicationReplacePrivateDirectoryForTest(t, predecessor)
			},
		},
	}
	for _, test := range tests {
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
			crash := errors.New("crash after prepared authority")
			crashed, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
				if point == vNextPublicationAfterControlRepairPrepared {
					return crash
				}
				return nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			operation, err := crashed.openOperation(context.Background(), syscall.LOCK_EX, true)
			if err != nil {
				t.Fatal(err)
			}
			err = crashed.writeCurrentLocked(operation, pointer)
			operation.close()
			if !errors.Is(err, crash) {
				t.Fatalf("writeCurrentLocked() error = %v, want prepared crash", err)
			}
			transaction, predecessor := vNextPublicationPendingCurrentRepairPathsForTest(t, crashed, root)
			vNextReplacePublicationControlForTest(t, root, vNextPublicationCurrentFile, []byte("malformed public control must not be decoded"))
			attacked := false
			fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
				if point == test.point && !attacked {
					attacked = true
					test.replace(t, transaction, predecessor)
				}
				return nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			err = fresh.Recover(context.Background())
			if !attacked {
				t.Fatal("private replacement barrier was not reached")
			}
			if err == nil || !strings.Contains(err.Error(), test.expected) {
				t.Fatalf("Recover() error = %v, want private %s refusal", err, test.expected)
			}
			if strings.Contains(err.Error(), "decode CURRENT") {
				t.Fatalf("Recover() decoded public CURRENT before private replacement refusal: %v", err)
			}
			replaced := transaction
			switch test.name {
			case "prepared":
				replaced = filepath.Join(transaction, vNextPublicationControlRepairPreparedFile)
			case "capture":
				replaced = filepath.Join(transaction, vNextPublicationControlCaptureName(1))
			case "phase":
				replaced = filepath.Join(transaction, vNextPublicationControlRepairPhaseName(2))
			case "predecessor":
				replaced = predecessor
			}
			if _, statErr := os.Lstat(replaced); statErr != nil {
				t.Fatalf("private replacement %q was removed: %v", replaced, statErr)
			}
			if _, statErr := os.Lstat(replaced + ".attacker-moved"); statErr != nil {
				t.Fatalf("original private authority %q was removed: %v", replaced, statErr)
			}
		})
	}
}

func TestVNextGenerationPublisherCheckRevalidatesPrivateAuthorityBeforePublicDecode(t *testing.T) {
	root := t.TempDir()
	artifacts := vNextPublicationArtifactsForTest("active", false)
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := baseline.Publish(artifacts); err != nil {
		t.Fatal(err)
	}
	transaction := vNextPublicationAuthorityTransactionForTest(t, baseline, vNextPublicationCurrentFile)
	attacked := false
	checker, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point == vNextPublicationAfterControlAuthorityReadScan && !attacked {
			attacked = true
			vNextPublicationReplacePrivateDirectoryForTest(t, filepath.Join(root, "acme", transaction))
			vNextReplacePublicationControlForTest(t, root, vNextPublicationCurrentFile, []byte("malformed CURRENT must not be decoded"))
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	err = checker.Check(artifacts)
	if !attacked {
		t.Fatal("check did not reach authority-read scan boundary")
	}
	if err == nil || !strings.Contains(err.Error(), "transaction identity changed") {
		t.Fatalf("Check() error = %v, want private transaction identity refusal", err)
	}
	if strings.Contains(err.Error(), "decode CURRENT") {
		t.Fatalf("Check() decoded public CURRENT before private authority revalidation: %v", err)
	}
	before := vNextPublicationTreeSnapshotForTest(t, filepath.Join(root, "acme"))
	err = checker.Check(artifacts)
	if err == nil || strings.Contains(err.Error(), "decode CURRENT") {
		t.Fatalf("second Check() error = %v, want private authority refusal before decode", err)
	}
	if after := vNextPublicationTreeSnapshotForTest(t, filepath.Join(root, "acme")); !bytes.Equal(before, after) {
		t.Fatal("Check() changed a substituted private authority tree")
	}
}

func vNextPublicationPendingCurrentRepairPathsForTest(t *testing.T, publisher *vNextGenerationPublisher, root string) (string, string) {
	t.Helper()
	operation, err := publisher.openOperation(context.Background(), syscall.LOCK_SH, false)
	if err != nil {
		t.Fatal(err)
	}
	defer operation.close()
	graph, err := publisher.scanControlAuthorityLocked(operation)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.close()
	head := graph.heads[vNextPublicationCurrentFile]
	if head == nil || head.state.record.Predecessor == nil {
		t.Fatalf("pending CURRENT authority head = %#v", head)
	}
	return filepath.Join(root, "acme", head.state.transactionName), filepath.Join(root, "acme", head.state.record.Predecessor.Transaction)
}

func vNextPublicationReplacePrivateDirectoryForTest(t *testing.T, path string) {
	t.Helper()
	if err := os.Rename(path, path+".attacker-moved"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
}

func TestVNextPublicationOpenRecordedCaptureRejectsSubstitutedDirectory(t *testing.T) {
	root := t.TempDir()
	transaction, err := vNextPublicationOpenDirectory(root, "publication control repair transaction")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if closeErr := transaction.Close(); closeErr != nil {
			t.Errorf("close transaction: %v", closeErr)
		}
	}()

	const captureName = ".capture-001"
	if err := unix.Mkdirat(int(transaction.file.Fd()), captureName, 0o700); err != nil {
		t.Fatal(err)
	}
	captureDirectory, err := transaction.openDirectory(captureName, "publication control capture")
	if err != nil {
		t.Fatal(err)
	}
	identity, err := vNextPublicationIdentityFromFile(captureDirectory.file, "publication control capture")
	if closeErr := captureDirectory.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		t.Fatal(err)
	}
	capture := vNextPublicationControlRepairCapture{
		Attempt:  1,
		Name:     captureName,
		Identity: vNextPublicationRecordIdentity(identity),
	}

	capturePath := filepath.Join(root, captureName)
	vNextPublicationReplacePrivateDirectoryForTest(t, capturePath)
	if directory, openErr := vNextPublicationOpenRecordedCaptureLocked(transaction, capture); openErr == nil {
		if closeErr := directory.Close(); closeErr != nil {
			t.Errorf("close substituted capture: %v", closeErr)
		}
		t.Fatal("open recorded capture accepted a substituted directory")
	} else if !strings.Contains(openErr.Error(), "capture identity changed") {
		t.Fatalf("open recorded capture error = %v, want capture identity refusal", openErr)
	}
}

func vNextPublicationReplacePrivateFileForTest(t *testing.T, path string) {
	t.Helper()
	if err := os.Rename(path, path+".attacker-moved"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestVNextGenerationPublisherRetainsSupersededAuthorityWhenCleanupIsUnsafe(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	pointer, err := baseline.Publish(vNextPublicationArtifactsForTest("active", false))
	if err != nil {
		t.Fatal(err)
	}
	operation, err := baseline.openOperation(context.Background(), syscall.LOCK_EX, true)
	if err != nil {
		t.Fatal(err)
	}
	err = baseline.writeCurrentLocked(operation, pointer)
	operation.close()
	if err != nil {
		t.Fatal(err)
	}
	_, predecessor := vNextPublicationPendingCurrentRepairPathsForTest(t, baseline, root)
	if _, err := os.Stat(predecessor); err != nil {
		t.Fatalf("normal successor discarded predecessor authority: %v", err)
	}
	fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.Recover(context.Background()); err != nil {
		t.Fatalf("Recover() with valid successor = %v", err)
	}
	if _, err := os.Stat(predecessor); err != nil {
		t.Fatalf("valid recovery discarded predecessor authority: %v", err)
	}
	vNextPublicationReplacePrivateDirectoryForTest(t, predecessor)
	vNextReplacePublicationControlForTest(t, root, vNextPublicationCurrentFile, []byte("third control must not be decoded"))
	if err := fresh.Recover(context.Background()); err == nil ||
		(!strings.Contains(err.Error(), "predecessor transaction") &&
			!strings.Contains(err.Error(), "publication control repair transaction")) ||
		strings.Contains(err.Error(), "decode CURRENT") {
		t.Fatalf("Recover() after predecessor substitution = %v, want private authority refusal before public decode", err)
	}
	if _, err := os.Stat(predecessor); err != nil {
		t.Fatalf("substituted predecessor was removed: %v", err)
	}
	if _, err := os.Stat(predecessor + ".attacker-moved"); err != nil {
		t.Fatalf("superseded predecessor was removed: %v", err)
	}
}

func TestVNextGenerationPublisherCooperatingWritersSerializeAuthorityHeads(t *testing.T) {
	tests := []struct {
		name   string
		target string
		write  func(*vNextGenerationPublisher, *vNextPublicationOperation, vNextGenerationPointer) error
	}{
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
	for _, test := range tests {
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
			firstAtCapture := make(chan struct{})
			releaseFirst := make(chan struct{})
			first, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
				if point == vNextPublicationBeforeControlRepairCapture {
					close(firstAtCapture)
					<-releaseFirst
				}
				return nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			firstDone := make(chan error, 1)
			go func() {
				operation, err := first.openOperation(context.Background(), syscall.LOCK_EX, true)
				if err == nil {
					err = test.write(first, operation, pointer)
					operation.close()
				}
				firstDone <- err
			}()
			select {
			case <-firstAtCapture:
			case <-time.After(5 * time.Second):
				t.Fatal("first writer did not reach authority capture")
			}

			second, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			secondEntered := make(chan struct{})
			secondDone := make(chan error, 1)
			go func() {
				close(secondEntered)
				operation, err := second.openOperation(context.Background(), syscall.LOCK_EX, true)
				if err == nil {
					err = test.write(second, operation, pointer)
					operation.close()
				}
				secondDone <- err
			}()
			<-secondEntered
			select {
			case err := <-secondDone:
				t.Fatalf("second writer entered first authority transition: %v", err)
			case <-time.After(100 * time.Millisecond):
			}
			close(releaseFirst)
			select {
			case err := <-firstDone:
				if err != nil {
					t.Fatalf("first writer error = %v", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("first writer did not finish")
			}
			select {
			case err := <-secondDone:
				if err != nil {
					t.Fatalf("second writer error = %v", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("second writer did not finish")
			}
			inspection, err := baseline.openOperation(context.Background(), syscall.LOCK_SH, false)
			if err != nil {
				t.Fatal(err)
			}
			graph, err := baseline.controlAuthorityForReadLocked(inspection)
			inspection.close()
			if err != nil {
				t.Fatal(err)
			}
			defer graph.close()
			head := graph.heads[test.target]
			if head == nil {
				t.Fatalf("%s has no authority head", test.target)
			}
			if _, _, _, terminal := head.state.terminal(); !terminal {
				t.Fatalf("%s head is not terminal", test.target)
			}
		})
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

func vNextPublicationFindPrivatePayloadForTest(t *testing.T, root string, want []byte) string {
	t.Helper()
	connectorRoot := filepath.Join(root, "acme")
	transactions, err := filepath.Glob(filepath.Join(connectorRoot, vNextPublicationControlRepairDirectoryPrefix+"*"))
	if err != nil {
		t.Fatal(err)
	}
	for _, transaction := range transactions {
		entries, err := os.ReadDir(transaction)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			candidate := filepath.Join(transaction, entry.Name())
			if entry.IsDir() {
				candidate = filepath.Join(candidate, vNextPublicationControlCaptureMember)
			}
			payload, err := os.ReadFile(candidate)
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			if err != nil {
				continue
			}
			if bytes.Equal(payload, want) {
				return candidate
			}
		}
	}
	t.Fatalf("publication private authority does not retain %q", want)
	return ""
}

func vNextPublicationTreeSnapshotForTest(t *testing.T, root string) []byte {
	t.Helper()
	type entry struct {
		Path    string `json:"path"`
		Device  uint64 `json:"device"`
		Inode   uint64 `json:"inode"`
		Mode    uint32 `json:"mode"`
		Payload []byte `json:"payload,omitempty"`
	}
	var entries []entry
	var visit func(*vNextPublicationDirectory, string)
	visit = func(directory *vNextPublicationDirectory, relative string) {
		defer func() {
			if err := directory.Close(); err != nil {
				t.Errorf("close snapshot directory %q: %v", relative, err)
			}
		}()
		children, err := directory.readDir()
		if err != nil {
			t.Fatal(err)
		}
		for _, child := range children {
			name := child.Name()
			if relative != "" {
				name = filepath.Join(relative, name)
			}
			observed, err := directory.identityAt(child.Name(), "snapshot member "+name)
			if err != nil {
				t.Fatal(err)
			}
			if vNextPublicationSnapshotAfterObservationForTest != nil {
				vNextPublicationSnapshotAfterObservationForTest(directory, child.Name(), observed)
			}
			snapshot := entry{Path: name, Device: observed.device, Inode: observed.inode, Mode: observed.mode}
			switch observed.mode {
			case unix.S_IFDIR:
				childDirectory, err := directory.openDirectory(child.Name(), "snapshot directory "+name)
				if err != nil {
					t.Fatal(err)
				}
				actual, identityErr := vNextPublicationIdentityFromFile(childDirectory.file, "snapshot directory "+name)
				if identityErr != nil {
					_ = childDirectory.Close()
					t.Fatal(identityErr)
				}
				if actual != observed {
					_ = childDirectory.Close()
					t.Fatalf("snapshot directory %q identity changed", name)
				}
				entries = append(entries, snapshot)
				visit(childDirectory, name)
			case unix.S_IFREG:
				file, err := directory.openRegular(child.Name(), "snapshot file "+name, unix.O_RDONLY)
				if err != nil {
					t.Fatal(err)
				}
				actual, identityErr := vNextPublicationIdentityFromFile(file, "snapshot file "+name)
				if identityErr != nil {
					_ = file.Close()
					t.Fatal(identityErr)
				}
				if actual != observed {
					_ = file.Close()
					t.Fatalf("snapshot file %q identity changed", name)
				}
				payload, readErr := io.ReadAll(file)
				closeErr := file.Close()
				if readErr != nil {
					t.Fatalf("read snapshot file %q: %v", name, readErr)
				}
				if closeErr != nil {
					t.Fatalf("close snapshot file %q: %v", name, closeErr)
				}
				snapshot.Payload = payload
				entries = append(entries, snapshot)
			default:
				entries = append(entries, snapshot)
			}
		}
	}
	directory, err := vNextPublicationOpenDirectory(root, "snapshot root")
	if err != nil {
		t.Fatal(err)
	}
	visit(directory, "")
	snapshot, err := json.Marshal(entries)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}
func TestVNextGenerationPublisherResumesInterruptedBaseAuthorityPreparation(t *testing.T) {
	tests := []struct {
		name    string
		crashAt int
	}{
		{name: "CURRENT first base authority", crashAt: 1},
		{name: "JOURNAL second base authority", crashAt: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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

			crash := errors.New("crash after durable base prepared authority")
			hits := 0
			var stdout, stderr bytes.Buffer
			code := runLockRenderContextWithHooks(context.Background(), []string{"lock-render", lock.Connector, "--defs", root}, &stdout, &stderr, vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
				if point != vNextPublicationAfterBaseControlRepairPrepared {
					return nil
				}
				hits++
				if hits == test.crashAt {
					return crash
				}
				return nil
			}})
			if code != 1 || !strings.Contains(stderr.String(), crash.Error()) {
				t.Fatalf("runLockRender() after %s interruption = %d; stdout=%q stderr=%q", test.name, code, stdout.String(), stderr.String())
			}
			if hits != test.crashAt {
				t.Fatalf("base prepared fault hits = %d, want %d", hits, test.crashAt)
			}

			beforeCheck := vNextPublicationTreeSnapshotForTest(t, connectorRoot)
			stdout.Reset()
			stderr.Reset()
			if code := runLockRender([]string{"lock-render", lock.Connector, "--defs", root, "--check"}, &stdout, &stderr); code != 1 {
				t.Fatalf("runLockRender(--check) during %s interruption = %d, want 1; stdout=%q stderr=%q", test.name, code, stdout.String(), stderr.String())
			}
			if stdout.Len() != 0 {
				t.Fatalf("runLockRender(--check) during %s interruption wrote stdout %q", test.name, stdout.String())
			}
			if afterCheck := vNextPublicationTreeSnapshotForTest(t, connectorRoot); !bytes.Equal(beforeCheck, afterCheck) {
				t.Fatalf("runLockRender(--check) changed interrupted %s authority state", test.name)
			}

			fresh, err := newVNextGenerationPublisher(root, lock.Connector, vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			if err := fresh.Recover(context.Background()); err != nil {
				t.Fatalf("fresh Recover() after %s interruption = %v", test.name, err)
			}
			vNextPublicationAssertTerminalAuthoritiesMatchPublicForTest(t, fresh)
			for _, target := range []string{vNextPublicationCurrentFile, vNextPublicationJournalFile} {
				if _, err := os.Lstat(filepath.Join(connectorRoot, target)); !errors.Is(err, fs.ErrNotExist) {
					t.Fatalf("%s after recovering %s base authority = %v, want original absence", target, test.name, err)
				}
			}

			stdout.Reset()
			stderr.Reset()
			if code := runLockRender([]string{"lock-render", lock.Connector, "--defs", root}, &stdout, &stderr); code != 0 {
				t.Fatalf("runLockRender() retry after %s interruption = %d; stdout=%q stderr=%q", test.name, code, stdout.String(), stderr.String())
			}
		})
	}
}
