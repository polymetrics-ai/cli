package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/manifestindex"
)

func (p *vNextGenerationPublisher) Publish(artifacts vNextPublicationArtifacts) (vNextGenerationPointer, error) {
	return p.PublishContext(context.Background(), artifacts)
}

func (p *vNextGenerationPublisher) Check(artifacts vNextPublicationArtifacts) error {
	return p.CheckContext(context.Background(), artifacts)
}

func runLockRender(args []string, stdout, stderr io.Writer) int {
	return runLockRenderContext(context.Background(), args, stdout, stderr)
}

func TestVNextGenerationPublisherActivatesClosedSetAndDefersHeldPrune(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatalf("newVNextGenerationPublisher() error = %v", err)
	}

	old := vNextPublicationArtifactsForTest("old", true)
	oldPointer, err := publisher.Publish(old)
	if err != nil {
		t.Fatalf("Publish(old) error = %v", err)
	}
	oldHandle, err := publisher.Open(context.Background())
	if err != nil {
		t.Fatalf("Open(old) error = %v", err)
	}
	defer oldHandle.Release()
	if err := publisher.Check(vNextPublicationArtifactsForTest("new", false)); err == nil {
		t.Fatal("Check(new) accepted stale optional rate_limits.json in the active old generation")
	}

	newSet := vNextPublicationArtifactsForTest("new", false)
	newPointer, err := publisher.Publish(newSet)
	if err != nil {
		t.Fatalf("Publish(new) error = %v", err)
	}
	if newPointer.Generation == oldPointer.Generation {
		t.Fatal("different closed artifact sets received the same generation")
	}

	active, err := publisher.Open(context.Background())
	if err != nil {
		t.Fatalf("Open(new) error = %v", err)
	}
	defer active.Release()
	assertVNextPublicationMarker(t, active, "new")
	if _, err := active.ReadFile("rate_limits.json"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("active generation retained stale optional rate_limits.json: %v", err)
	}
	if got, want := active.Files(), []string{"atlas.json", "index.json", "manifest.json", "metadata.json", "proof.json", "provenance.json", "spec.json"}; !sameStrings(got, want) {
		t.Fatalf("active closed files = %#v, want %#v", got, want)
	}
	if err := publisher.Check(newSet); err != nil {
		t.Fatalf("Check(new) error = %v", err)
	}

	assertVNextPublicationMarker(t, oldHandle, "old")
	if err := publisher.Prune(context.Background()); err != nil {
		t.Fatalf("Prune() while old handle held error = %v", err)
	}
	if !publisher.GenerationExists(oldPointer.Generation) {
		t.Fatal("Prune() removed an in-use old generation")
	}

	oldHandle.Release()
	if err := publisher.Prune(context.Background()); err != nil {
		t.Fatalf("Prune() after old handle release error = %v", err)
	}
	if publisher.GenerationExists(oldPointer.Generation) {
		t.Fatal("Prune() retained an unheld stale generation")
	}
}

func TestVNextGenerationPublisherDefersHeldGenerationPrune(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	old := vNextPublicationArtifactsForTest("old", false)
	oldPointer, err := publisher.Publish(old)
	if err != nil {
		t.Fatalf("Publish(old) error = %v", err)
	}
	held, err := publisher.Open(context.Background())
	if err != nil {
		t.Fatalf("Open(old) error = %v", err)
	}
	defer held.Release()
	if _, err := publisher.Publish(vNextPublicationArtifactsForTest("new", false)); err != nil {
		t.Fatalf("Publish(new) error = %v", err)
	}
	if err := publisher.Prune(context.Background()); err != nil {
		t.Fatalf("Prune() while old generation is held error = %v", err)
	}
	if !publisher.GenerationExists(oldPointer.Generation) {
		t.Fatal("Prune() removed the held generation")
	}
}

func TestVNextGenerationPublisherRefusesPruneWithInvalidLease(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	old, err := publisher.Publish(vNextPublicationArtifactsForTest("old", false))
	if err != nil {
		t.Fatal(err)
	}
	held, err := publisher.Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	current, err := publisher.Publish(vNextPublicationArtifactsForTest("current", false))
	held.Release()
	if err != nil {
		t.Fatal(err)
	}
	leasePath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, old.Generation, vNextPublicationLeaseFile)
	if err := os.WriteFile(leasePath, []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := publisher.Prune(context.Background()); err == nil || !strings.Contains(err.Error(), "lease") {
		t.Fatalf("Prune() error = %v, want invalid lease refusal", err)
	}
	if !publisher.GenerationExists(old.Generation) {
		t.Fatal("Prune() removed the generation with an invalid lease")
	}
	if !publisher.GenerationExists(current.Generation) {
		t.Fatal("Prune() removed the active generation")
	}
}

func TestVNextGenerationPublisherCheckIsReadOnly(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	oldPointer, err := publisher.Publish(vNextPublicationArtifactsForTest("old", true))
	if err != nil {
		t.Fatal(err)
	}
	oldHandle, err := publisher.Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	newSet := vNextPublicationArtifactsForTest("new", false)
	if _, err := publisher.Publish(newSet); err != nil {
		t.Fatal(err)
	}
	oldHandle.Release()
	currentPath := filepath.Join(root, "acme", vNextPublicationCurrentFile)
	beforeCurrent, err := os.ReadFile(currentPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := publisher.Check(newSet); err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	afterCurrent, err := os.ReadFile(currentPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(afterCurrent, beforeCurrent) {
		t.Fatal("Check() rewrote CURRENT")
	}
	if !publisher.GenerationExists(oldPointer.Generation) {
		t.Fatal("Check() pruned an unheld stale generation")
	}
	if _, err := os.Stat(filepath.Join(root, "acme", vNextPublicationJournalFile)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Check() created a journal: %v", err)
	}
}

func TestVNextGenerationPublisherCheckRefusesPendingJournalWithoutRecovery(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	artifacts := vNextPublicationArtifactsForTest("active", false)
	if _, err := publisher.Publish(artifacts); err != nil {
		t.Fatal(err)
	}
	journalPath := filepath.Join(root, "acme", vNextPublicationJournalFile)
	journal := []byte(`{"not":"a publication journal"}` + "\n")
	if err := os.WriteFile(journalPath, journal, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := publisher.Check(artifacts); err == nil {
		t.Fatal("Check() accepted a pending journal")
	}
	if got, err := os.ReadFile(journalPath); err != nil || !bytes.Equal(got, journal) {
		t.Fatalf("Check() recovered or rewrote journal: err=%v got=%q", err, got)
	}
}

func TestVNextGenerationPublisherRepublishingActiveSetRemainsRecoverable(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	artifacts := vNextPublicationArtifactsForTest("active", false)
	if _, err := publisher.Publish(artifacts); err != nil {
		t.Fatal(err)
	}
	faultHit := false
	guard, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point == vNextPublicationBeforeCurrentRename {
			faultHit = true
			return errors.New("repeat publication rewrote CURRENT")
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := guard.Publish(artifacts); err != nil {
		t.Fatalf("Publish(repeated active set) error = %v", err)
	}
	if faultHit {
		t.Fatal("Publish(repeated active set) rewrote CURRENT")
	}
	if err := publisher.Recover(context.Background()); err != nil {
		t.Fatalf("Recover() after repeat publish error = %v", err)
	}
	if err := publisher.Check(artifacts); err != nil {
		t.Fatalf("Check() after interrupted repeat publish error = %v", err)
	}
}

func TestVNextGenerationPublisherRefusesSymlinkedStaleEntryWithoutDeletion(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := publisher.Publish(vNextPublicationArtifactsForTest("active", false)); err != nil {
		t.Fatal(err)
	}
	authorOwned := filepath.Join(root, "author-owned")
	if err := os.Mkdir(authorOwned, 0o755); err != nil {
		t.Fatal(err)
	}
	sentinelPath := filepath.Join(authorOwned, "sentinel.txt")
	if err := os.WriteFile(sentinelPath, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	stageLink := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, ".stage-author-owned")
	if err := os.Symlink(authorOwned, stageLink); err != nil {
		t.Fatal(err)
	}

	if err := publisher.Recover(context.Background()); err == nil {
		t.Fatal("Recover() accepted a symlinked stale stage")
	}
	if got, err := os.ReadFile(sentinelPath); err != nil || string(got) != "keep" {
		t.Fatalf("Recover() altered author-owned target: err=%v got=%q", err, got)
	}
	if info, err := os.Lstat(stageLink); err != nil || info.Mode()&fs.ModeSymlink == 0 {
		t.Fatalf("Recover() removed or replaced symlinked stale entry: info=%v err=%v", info, err)
	}
}

func TestVNextGenerationPublisherRetainsDescriptorsAcrossConnectorReplacement(t *testing.T) {
	type operation struct {
		name string
		run  func(t *testing.T, root string, publisher *vNextGenerationPublisher, replacement *vNextGenerationPublisher) error
	}
	operations := []operation{
		{
			name: "Publish",
			run: func(t *testing.T, _ string, publisher *vNextGenerationPublisher, replacement *vNextGenerationPublisher) error {
				t.Helper()
				if _, err := publisher.Publish(vNextPublicationArtifactsForTest("old", false)); err != nil {
					return err
				}
				_, err := replacement.Publish(vNextPublicationArtifactsForTest("new", false))
				return err
			},
		},
		{
			name: "Recover",
			run: func(t *testing.T, root string, publisher *vNextGenerationPublisher, replacement *vNextGenerationPublisher) error {
				t.Helper()
				pointer, err := publisher.Publish(vNextPublicationArtifactsForTest("active", false))
				if err != nil {
					return err
				}
				return vNextCreateOwnedStageForTest(root, pointer, ".stage-original", replacement.Recover)
			},
		},
		{
			name: "Prune",
			run: func(t *testing.T, _ string, publisher *vNextGenerationPublisher, replacement *vNextGenerationPublisher) error {
				t.Helper()
				old, err := publisher.Publish(vNextPublicationArtifactsForTest("old", false))
				if err != nil {
					return err
				}
				handle, err := publisher.Open(context.Background())
				if err != nil {
					return err
				}
				if _, err := publisher.Publish(vNextPublicationArtifactsForTest("new", false)); err != nil {
					handle.Release()
					return err
				}
				handle.Release()
				if !publisher.GenerationExists(old.Generation) {
					return errors.New("fixture did not retain held old generation")
				}
				return replacement.Prune(context.Background())
			},
		},
	}
	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			root := t.TempDir()
			publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			external := t.TempDir()
			hookHit := false
			stageName := ".stage-external"
			if operation.name == "Recover" {
				stageName = ".stage-original"
			}
			replacement, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
				want := vNextPublicationBeforePrune
				if operation.name == "Recover" {
					want = vNextPublicationBeforeStageCleanup
				}
				if point != want || hookHit {
					return nil
				}
				hookHit = true
				return vNextReplaceConnectorRootForTest(root, external, stageName)
			}})
			if err != nil {
				t.Fatal(err)
			}

			runErr := operation.run(t, root, publisher, replacement)
			if !hookHit {
				t.Fatalf("%s() did not reach root-replacement hook", operation.name)
			}
			sentinel := filepath.Join(external, vNextPublicationGenerationDirectory, stageName, "sentinel.txt")
			if got, err := os.ReadFile(sentinel); err != nil || string(got) != "keep" {
				t.Fatalf("%s() touched replacement-root sentinel: err=%v got=%q operation_err=%v", operation.name, err, got, runErr)
			}
			if runErr != nil {
				t.Fatalf("%s() error after retained-root replacement = %v", operation.name, runErr)
			}
		})
	}
}

func TestRunLockRenderRetainsSourceDescriptorAcrossConnectorReplacement(t *testing.T) {
	root := t.TempDir()
	lock := minimalVNextLockForTest()
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	connectorRoot := filepath.Join(root, lock.Connector)
	if err := os.MkdirAll(connectorRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(connectorRoot, "source.lock.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	external := t.TempDir()
	hookHit := false
	var stdout, stderr bytes.Buffer
	code := runLockRenderContextWithHooks(context.Background(), []string{"lock-render", lock.Connector, "--defs", root}, &stdout, &stderr, vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationAfterLockAcquire || hookHit {
			return nil
		}
		hookHit = true
		return vNextReplaceConnectorRootForTest(root, external, ".stage-external")
	}})
	if code != 0 {
		t.Fatalf("runLockRenderContextWithHooks() = %d; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !hookHit {
		t.Fatal("lock render did not reach connector-replacement hook")
	}
	sentinel := filepath.Join(external, vNextPublicationGenerationDirectory, ".stage-external", "sentinel.txt")
	if got, err := os.ReadFile(sentinel); err != nil || string(got) != "keep" {
		t.Fatalf("lock render touched replacement-root sentinel: err=%v got=%q", err, got)
	}
	movedGenerationRoot := filepath.Join(root, lock.Connector+"-moved", vNextPublicationGenerationDirectory)
	entries, err := os.ReadDir(movedGenerationRoot)
	if err != nil {
		t.Fatalf("read published moved generation root: %v", err)
	}
	for _, entry := range entries {
		if vNextPublicationGenerationIDValid(entry.Name()) {
			return
		}
	}
	t.Fatalf("lock render did not publish through retained connector descriptor: %#v", entries)
}

func TestRunLockRenderRejectsSourceMutationBeforeGenerationCreation(t *testing.T) {
	root := t.TempDir()
	lock := minimalVNextLockForTest()
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	connectorRoot := filepath.Join(root, lock.Connector)
	if err := os.MkdirAll(connectorRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	sourcePath := filepath.Join(connectorRoot, "source.lock.json")
	if err := os.WriteFile(sourcePath, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := runLockRenderContextWithHooks(context.Background(), []string{"lock-render", lock.Connector, "--defs", root}, &stdout, &stderr, vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationAfterLockAcquire {
			return nil
		}
		return os.WriteFile(sourcePath, append(append([]byte(nil), raw...), '\n'), 0o600)
	}})
	if code != 1 || !strings.Contains(stderr.String(), "source lock changed during admission") {
		t.Fatalf("runLockRenderContextWithHooks() = %d; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(connectorRoot, vNextPublicationGenerationDirectory)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("source mutation created a generation root: %v", err)
	}
}

func TestVNextGenerationPublisherRefusesSymlinkedConnectorRootWithoutTouchingTarget(t *testing.T) {
	operations := []struct {
		name string
		run  func(*vNextGenerationPublisher) error
	}{
		{
			name: "Publish",
			run: func(publisher *vNextGenerationPublisher) error {
				_, err := publisher.Publish(vNextPublicationArtifactsForTest("active", false))
				return err
			},
		},
		{name: "Check", run: func(publisher *vNextGenerationPublisher) error {
			return publisher.Check(vNextPublicationArtifactsForTest("active", false))
		}},
		{name: "Recover", run: func(publisher *vNextGenerationPublisher) error {
			return publisher.Recover(context.Background())
		}},
		{
			name: "Open",
			run: func(publisher *vNextGenerationPublisher) error {
				handle, err := publisher.Open(context.Background())
				if handle != nil {
					handle.Release()
				}
				return err
			},
		},
		{name: "Prune", run: func(publisher *vNextGenerationPublisher) error {
			return publisher.Prune(context.Background())
		}},
	}
	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			defsRoot := t.TempDir()
			external := t.TempDir()
			sentinel := filepath.Join(external, "sentinel.txt")
			if err := os.WriteFile(sentinel, []byte("keep"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(external, filepath.Join(defsRoot, "acme")); err != nil {
				t.Fatal(err)
			}
			publisher, err := newVNextGenerationPublisher(defsRoot, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}

			err = operation.run(publisher)
			if err == nil || !strings.Contains(err.Error(), "symlink") {
				t.Fatalf("%s() error = %v, want connector-root symlink refusal", operation.name, err)
			}
			if got, err := os.ReadFile(sentinel); err != nil || string(got) != "keep" {
				t.Fatalf("%s() touched external sentinel: err=%v got=%q", operation.name, err, got)
			}
			entries, err := os.ReadDir(external)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 1 || entries[0].Name() != "sentinel.txt" {
				t.Fatalf("%s() created external connector state: %#v", operation.name, entries)
			}
		})
	}
}

func TestVNextGenerationPublisherSerializesMatchedLockAnchorReplacement(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := baseline.Publish(vNextPublicationArtifactsForTest("old", false)); err != nil {
		t.Fatalf("Publish(old) error = %v", err)
	}

	connectorRoot := filepath.Join(root, "acme")
	vNextPublicationWriteMatchedLockAnchorPairForTest(t, connectorRoot)

	acquired := make(chan struct{})
	release := make(chan struct{})
	firstResult := make(chan error, 1)
	first, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationAfterLockAcquire {
			return nil
		}
		close(acquired)
		<-release
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		_, publishErr := first.Publish(vNextPublicationArtifactsForTest("first", false))
		firstResult <- publishErr
	}()
	<-acquired

	lockPath := filepath.Join(connectorRoot, ".connectorgen.lock")
	anchorPath := filepath.Join(connectorRoot, ".connectorgen.lock.anchor")
	oldLockPath := filepath.Join(t.TempDir(), "old-lock")
	oldAnchorPath := filepath.Join(t.TempDir(), "old-anchor")
	if err := os.Rename(lockPath, oldLockPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(anchorPath, oldAnchorPath); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(lockPath, []byte("replacement"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(lockPath, anchorPath); err != nil {
		t.Fatal(err)
	}
	replacementLock, err := os.Stat(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	replacementAnchor, err := os.Stat(anchorPath)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(replacementLock, replacementAnchor) {
		t.Fatal("replacement lock and anchor are not a matching pair")
	}

	secondAcquired := make(chan struct{})
	secondResult := make(chan error, 1)
	second, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point == vNextPublicationAfterLockAcquire {
			close(secondAcquired)
			return errors.New("second publication acquired serialization domain")
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		_, publishErr := second.Publish(vNextPublicationArtifactsForTest("second", false))
		secondResult <- publishErr
	}()

	select {
	case <-secondAcquired:
		t.Fatal("matched replacement lock pair admitted a second publisher while first remained active")
	case <-time.After(200 * time.Millisecond):
	}

	close(release)
	select {
	case err := <-firstResult:
		if err != nil {
			t.Fatalf("first Publish() after matched lock replacement error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("first Publish() did not complete after matched lock replacement")
	}
	select {
	case <-secondAcquired:
	case <-time.After(time.Second):
		t.Fatal("second Publish() did not acquire after first completed")
	}
	select {
	case err := <-secondResult:
		if err == nil || !strings.Contains(err.Error(), "second publication acquired") {
			t.Fatalf("second Publish() error = %v, want post-first acquisition sentinel", err)
		}
	case <-time.After(time.Second):
		t.Fatal("second Publish() did not return after first completed")
	}
	if err := baseline.Recover(context.Background()); err != nil {
		t.Fatalf("Recover() after matched lock replacement error = %v", err)
	}
	if err := baseline.Check(vNextPublicationArtifactsForTest("first", false)); err != nil {
		t.Fatalf("Check(first) after matched lock replacement error = %v", err)
	}
}

func TestVNextPublicationOpenLockBindsConnectorDirectory(t *testing.T) {
	root := t.TempDir()
	connectorPath := filepath.Join(root, "acme")
	if err := os.Mkdir(connectorPath, 0o755); err != nil {
		t.Fatal(err)
	}
	connector, err := vNextPublicationOpenDirectory(connectorPath, "connector publication root")
	if err != nil {
		t.Fatal(err)
	}
	defer connector.Close()
	lock, identity, err := vNextPublicationOpenLock(connector)
	if err != nil {
		t.Fatalf("vNextPublicationOpenLock() error = %v", err)
	}
	defer lock.Close()
	connectorIdentity, err := vNextPublicationIdentityFromFile(connector.file, "connector publication root")
	if err != nil {
		t.Fatal(err)
	}
	if identity != connectorIdentity {
		t.Fatalf("lock identity = %#v, want connector identity %#v", identity, connectorIdentity)
	}
	if err := vNextPublicationAssertLockBound(connector, lock, identity); err != nil {
		t.Fatalf("vNextPublicationAssertLockBound() error = %v", err)
	}
	entries, err := os.ReadDir(connectorPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("vNextPublicationOpenLock() created sibling state: %#v", entries)
	}
}

func vNextPublicationWriteMatchedLockAnchorPairForTest(t *testing.T, connectorRoot string) {
	t.Helper()
	lockPath := filepath.Join(connectorRoot, ".connectorgen.lock")
	anchorPath := filepath.Join(connectorRoot, ".connectorgen.lock.anchor")
	if _, err := os.Lstat(lockPath); errors.Is(err, fs.ErrNotExist) {
		if err := os.WriteFile(lockPath, []byte("original"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Link(lockPath, anchorPath); err != nil {
			t.Fatal(err)
		}
	}
	lockInfo, err := os.Stat(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	anchorInfo, err := os.Stat(anchorPath)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(lockInfo, anchorInfo) {
		t.Fatal("original lock and anchor are not a matching pair")
	}
}

func TestRunLockRenderRefusesSymlinkedConnectorRootWithoutTouchingTarget(t *testing.T) {
	defsRoot := t.TempDir()
	external := t.TempDir()
	sentinel := filepath.Join(external, "source.lock.json")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(defsRoot, "acme")); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	if code := runLockRender([]string{"lock-render", "acme", "--defs", defsRoot}, &stdout, &stderr); code != 1 {
		t.Fatalf("runLockRender() = %d, want symlink refusal; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "symlink") {
		t.Fatalf("runLockRender() error lacks symlink refusal: %s", stderr.String())
	}
	if got, err := os.ReadFile(sentinel); err != nil || string(got) != "keep" {
		t.Fatalf("runLockRender() touched external source lock: err=%v got=%q", err, got)
	}
}

func TestVNextGenerationPublisherPreservesUnownedStageAcrossMutations(t *testing.T) {
	operations := []struct {
		name string
		run  func(*vNextGenerationPublisher) error
	}{
		{name: "Recover", run: func(publisher *vNextGenerationPublisher) error {
			return publisher.Recover(context.Background())
		}},
		{
			name: "Publish",
			run: func(publisher *vNextGenerationPublisher) error {
				_, err := publisher.Publish(vNextPublicationArtifactsForTest("new", false))
				return err
			},
		},
		{
			name: "Open",
			run: func(publisher *vNextGenerationPublisher) error {
				handle, err := publisher.Open(context.Background())
				if handle != nil {
					handle.Release()
				}
				return err
			},
		},
		{name: "Prune", run: func(publisher *vNextGenerationPublisher) error {
			return publisher.Prune(context.Background())
		}},
	}
	stageCases := []struct {
		name   string
		marker []byte
	}{
		{name: "missing marker"},
		{name: "malformed marker", marker: []byte(`{"version":1,"connector":"acme"` + "\n")},
	}
	for _, stageCase := range stageCases {
		for _, operation := range operations {
			t.Run(stageCase.name+"/"+operation.name, func(t *testing.T) {
				root := t.TempDir()
				publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := publisher.Publish(vNextPublicationArtifactsForTest("active", false)); err != nil {
					t.Fatal(err)
				}
				stage := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, ".stage-author-owned")
				if err := os.Mkdir(stage, 0o755); err != nil {
					t.Fatal(err)
				}
				sentinel := filepath.Join(stage, "sentinel.txt")
				if err := os.WriteFile(sentinel, []byte("keep"), 0o600); err != nil {
					t.Fatal(err)
				}
				if stageCase.marker != nil {
					if err := os.WriteFile(filepath.Join(stage, vNextPublicationStageOwnerFile), stageCase.marker, 0o600); err != nil {
						t.Fatal(err)
					}
				}

				err = operation.run(publisher)
				if err == nil || !strings.Contains(err.Error(), "ownership") {
					t.Fatalf("%s() error = %v, want stage ownership refusal", operation.name, err)
				}
				if got, err := os.ReadFile(sentinel); err != nil || string(got) != "keep" {
					t.Fatalf("%s() removed unowned stage sentinel: err=%v got=%q", operation.name, err, got)
				}
				if info, err := os.Lstat(stage); err != nil || !info.IsDir() {
					t.Fatalf("%s() removed unowned stage: info=%v err=%v", operation.name, info, err)
				}
			})
		}
	}
}

func TestVNextGenerationPublisherRecoversDurablyOwnedStage(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	artifacts := vNextPublicationArtifactsForTest("active", false)
	pointer, err := publisher.Publish(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	stageName := ".stage-publisher-owned"
	stage := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, stageName)
	if err := os.Mkdir(stage, 0o755); err != nil {
		t.Fatal(err)
	}
	marker, err := vNextPublicationJSON(vNextPublicationStageOwner{
		Version:    1,
		Connector:  "acme",
		Generation: pointer.Generation,
		Stage:      stageName,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stage, vNextPublicationStageOwnerFile), marker, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := publisher.Recover(context.Background()); err != nil {
		t.Fatalf("Recover() error = %v", err)
	}
	if _, err := os.Lstat(stage); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Recover() retained a durably owned stale stage: %v", err)
	}
	if err := publisher.Check(artifacts); err != nil {
		t.Fatalf("Check() after owned-stage recovery error = %v", err)
	}
}

func TestVNextGenerationPublisherRejectsSelfConsistentRenamedGeneration(t *testing.T) {
	operations := []struct {
		name string
		run  func(*vNextGenerationPublisher, vNextPublicationArtifacts) error
	}{
		{name: "Check", run: func(publisher *vNextGenerationPublisher, artifacts vNextPublicationArtifacts) error {
			return publisher.Check(artifacts)
		}},
		{
			name: "Open",
			run: func(publisher *vNextGenerationPublisher, _ vNextPublicationArtifacts) error {
				handle, err := publisher.Open(context.Background())
				if handle != nil {
					handle.Release()
				}
				return err
			},
		},
		{name: "Recover", run: func(publisher *vNextGenerationPublisher, _ vNextPublicationArtifacts) error {
			return publisher.Recover(context.Background())
		}},
		{name: "Prune", run: func(publisher *vNextGenerationPublisher, _ vNextPublicationArtifacts) error {
			return publisher.Prune(context.Background())
		}},
	}
	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			root := t.TempDir()
			publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			artifacts := vNextPublicationArtifactsForTest("active", false)
			pointer, err := publisher.Publish(artifacts)
			if err != nil {
				t.Fatal(err)
			}
			renamed := vNextRewriteGenerationIdentityForTest(t, root, pointer)
			currentPath := filepath.Join(root, "acme", vNextPublicationCurrentFile)
			before, err := os.ReadFile(currentPath)
			if err != nil {
				t.Fatal(err)
			}

			err = operation.run(publisher, artifacts)
			if err == nil || !strings.Contains(err.Error(), "content address") {
				t.Fatalf("%s() error = %v, want renamed content-address refusal", operation.name, err)
			}
			if got, err := os.ReadFile(currentPath); err != nil || !bytes.Equal(got, before) {
				t.Fatalf("%s() rewrote CURRENT after refusing renamed generation: err=%v got=%q", operation.name, err, got)
			}
			if !publisher.GenerationExists(renamed.Generation) {
				t.Fatalf("%s() removed the rejected generation", operation.name)
			}
		})
	}
}

func TestVNextGenerationPublisherRejectsUnexpectedDirectoryAndNonemptyLease(t *testing.T) {
	cases := []struct {
		name    string
		corrupt func(t *testing.T, root string, pointer vNextGenerationPointer)
		want    string
	}{
		{
			name: "unexpected directory",
			corrupt: func(t *testing.T, root string, pointer vNextGenerationPointer) {
				t.Helper()
				if err := os.Mkdir(filepath.Join(root, "acme", vNextPublicationGenerationDirectory, pointer.Generation, "author-owned"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			want: "unexpected directory",
		},
		{
			name: "nonempty lease",
			corrupt: func(t *testing.T, root string, pointer vNextGenerationPointer) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(root, "acme", vNextPublicationGenerationDirectory, pointer.Generation, vNextPublicationLeaseFile), []byte("not empty"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			want: "lease",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			artifacts := vNextPublicationArtifactsForTest("active", false)
			pointer, err := publisher.Publish(artifacts)
			if err != nil {
				t.Fatal(err)
			}
			test.corrupt(t, root, pointer)

			if err := publisher.Check(artifacts); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Check() error = %v, want %q refusal", err, test.want)
			}
			handle, err := publisher.Open(context.Background())
			if handle != nil {
				handle.Release()
			}
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Open() error = %v, want %q refusal", err, test.want)
			}
		})
	}
}

func TestVNextGenerationPublisherRejectsDuplicatePublicationControlMembers(t *testing.T) {
	cases := []struct {
		name    string
		corrupt func(t *testing.T, root string, pointer vNextGenerationPointer)
	}{
		{
			name: "CURRENT",
			corrupt: func(t *testing.T, root string, pointer vNextGenerationPointer) {
				t.Helper()
				payload := []byte(fmt.Sprintf(`{"generation":%q,"generation":%q,"integrity_digest":%q}`, pointer.Generation, pointer.Generation, pointer.IntegrityDigest))
				if err := os.WriteFile(filepath.Join(root, "acme", vNextPublicationCurrentFile), payload, 0o600); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "JOURNAL",
			corrupt: func(t *testing.T, root string, pointer vNextGenerationPointer) {
				t.Helper()
				payload := []byte(fmt.Sprintf(`{"new":{"generation":%q,"integrity_digest":%q},"state":"prepared","state":"prepared"}`, pointer.Generation, pointer.IntegrityDigest))
				if err := os.WriteFile(filepath.Join(root, "acme", vNextPublicationJournalFile), payload, 0o600); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "integrity root",
			corrupt: func(t *testing.T, root string, pointer vNextGenerationPointer) {
				t.Helper()
				vNextDuplicateIntegrityMemberForTest(t, root, pointer, fmt.Sprintf(`"generation":%q`, pointer.Generation))
			},
		},
		{
			name: "integrity file",
			corrupt: func(t *testing.T, root string, pointer vNextGenerationPointer) {
				t.Helper()
				integrityPath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, pointer.Generation, vNextPublicationIntegrityFile)
				payload, err := os.ReadFile(integrityPath)
				if err != nil {
					t.Fatal(err)
				}
				var integrity vNextGenerationIntegrity
				if err := json.Unmarshal(payload, &integrity); err != nil {
					t.Fatal(err)
				}
				vNextDuplicateIntegrityMemberForTest(t, root, pointer, fmt.Sprintf(`"digest":%q`, integrity.Files[0].Digest))
			},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			pointer, err := publisher.Publish(vNextPublicationArtifactsForTest("active", false))
			if err != nil {
				t.Fatal(err)
			}
			test.corrupt(t, root, pointer)

			err = publisher.Recover(context.Background())
			if err == nil || !strings.Contains(err.Error(), "duplicate JSON object member") {
				t.Fatalf("Recover() error = %v, want duplicate-control-member refusal", err)
			}
		})
	}
}

func TestVNextGenerationPublisherRejectsOversizedIntegrityControl(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	pointer, err := publisher.Publish(vNextPublicationArtifactsForTest("active", false))
	if err != nil {
		t.Fatal(err)
	}
	integrityPath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, pointer.Generation, vNextPublicationIntegrityFile)
	payload := bytes.Repeat([]byte("x"), vNextPublicationControlMaxBytes+1)
	if err := os.WriteFile(integrityPath, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	current, err := vNextPublicationJSON(vNextGenerationPointer{
		Generation:      pointer.Generation,
		IntegrityDigest: vNextPublicationDigest(payload),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "acme", vNextPublicationCurrentFile), current, 0o600); err != nil {
		t.Fatal(err)
	}

	err = publisher.Recover(context.Background())
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("Recover() error = %v, want bounded-integrity refusal", err)
	}
}

func TestVNextGenerationPublisherContextCancelsContendedLockWithoutStateChange(t *testing.T) {
	operations := []struct {
		name string
		run  func(context.Context, *vNextGenerationPublisher, vNextPublicationArtifacts) error
	}{
		{
			name: "Publish",
			run: func(ctx context.Context, publisher *vNextGenerationPublisher, artifacts vNextPublicationArtifacts) error {
				_, err := publisher.PublishContext(ctx, artifacts)
				return err
			},
		},
		{
			name: "Check",
			run: func(ctx context.Context, publisher *vNextGenerationPublisher, artifacts vNextPublicationArtifacts) error {
				return publisher.CheckContext(ctx, artifacts)
			},
		},
		{
			name: "Recover",
			run: func(ctx context.Context, publisher *vNextGenerationPublisher, _ vNextPublicationArtifacts) error {
				return publisher.Recover(ctx)
			},
		},
		{
			name: "Open",
			run: func(ctx context.Context, publisher *vNextGenerationPublisher, _ vNextPublicationArtifacts) error {
				handle, err := publisher.Open(ctx)
				if handle != nil {
					handle.Release()
				}
				return err
			},
		},
		{
			name: "Prune",
			run: func(ctx context.Context, publisher *vNextGenerationPublisher, _ vNextPublicationArtifacts) error {
				return publisher.Prune(ctx)
			},
		},
	}
	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			root := t.TempDir()
			publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			artifacts := vNextPublicationArtifactsForTest("active", false)
			if _, err := publisher.Publish(artifacts); err != nil {
				t.Fatal(err)
			}
			currentPath := filepath.Join(root, "acme", vNextPublicationCurrentFile)
			currentBefore, err := os.ReadFile(currentPath)
			if err != nil {
				t.Fatal(err)
			}
			journalPath := filepath.Join(root, "acme", vNextPublicationJournalFile)
			if _, err := os.Lstat(journalPath); !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("publication unexpectedly left JOURNAL before contention: %v", err)
			}

			held := vNextPublicationHoldLockForTest(t, root)
			defer func() { unlockVNextPublicationFile(held) }()
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			err = operation.run(ctx, publisher, artifacts)
			cancel()
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("%s under held lock error = %v, want context deadline exceeded", operation.name, err)
			}
			if currentAfter, err := os.ReadFile(currentPath); err != nil || !bytes.Equal(currentAfter, currentBefore) {
				t.Fatalf("%s changed CURRENT while lock acquisition was cancelled: err=%v current=%q", operation.name, err, currentAfter)
			}
			if _, err := os.Lstat(journalPath); !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("%s changed JOURNAL while lock acquisition was cancelled: %v", operation.name, err)
			}
			unlockVNextPublicationFile(held)
			held = nil
			if err := operation.run(context.Background(), publisher, artifacts); err != nil {
				t.Fatalf("%s did not retry after lock release: %v", operation.name, err)
			}
		})
	}
}

func TestVNextGenerationPublisherCancelsAfterLockAcquireBeforeMutationAndRetries(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	oldPointer, err := baseline.Publish(vNextPublicationArtifactsForTest("old", true))
	if err != nil {
		t.Fatalf("Publish(old) error = %v", err)
	}
	currentPath := filepath.Join(root, "acme", vNextPublicationCurrentFile)
	currentBefore, err := os.ReadFile(currentPath)
	if err != nil {
		t.Fatal(err)
	}
	acquired := make(chan struct{})
	release := make(chan struct{})
	stageEntered := make(chan struct{}, 1)
	blocked, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		switch point {
		case vNextPublicationAfterLockAcquire:
			close(acquired)
			<-release
		case vNextPublicationBeforeStageDirectory:
			stageEntered <- struct{}{}
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	newSet := vNextPublicationArtifactsForTest("new", false)
	go func() {
		_, err := blocked.PublishContext(ctx, newSet)
		result <- err
	}()
	select {
	case <-acquired:
	case <-time.After(5 * time.Second):
		t.Fatal("publication did not acquire its nonblocking lock")
	}
	cancel()
	close(release)
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("PublishContext() error = %v, want context cancellation", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("publication did not return after post-lock cancellation")
	}
	select {
	case <-stageEntered:
		t.Fatal("publication entered staged-tree mutation after post-lock cancellation")
	default:
	}
	if currentAfter, err := os.ReadFile(currentPath); err != nil || !bytes.Equal(currentAfter, currentBefore) {
		t.Fatalf("cancelled publication changed CURRENT: err=%v current=%q", err, currentAfter)
	}
	if _, err := os.Lstat(filepath.Join(root, "acme", vNextPublicationJournalFile)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("cancelled publication changed JOURNAL: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(root, "acme", vNextPublicationGenerationDirectory))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != oldPointer.Generation {
		t.Fatalf("cancelled publication changed generation tree: %#v", entries)
	}
	if _, err := baseline.Publish(newSet); err != nil {
		t.Fatalf("Publish(new) retry error = %v", err)
	}
}

func TestVNextGenerationPublisherRefusesReplacedAtomicControlTemporary(t *testing.T) {
	tests := []struct {
		name   string
		target string
		point  vNextPublicationFaultPoint
		write  func(*vNextGenerationPublisher, *vNextPublicationOperation, vNextGenerationPointer) error
	}{
		{
			name:   "CURRENT",
			target: vNextPublicationCurrentFile,
			point:  vNextPublicationBeforeCurrentRename,
			write: func(publisher *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
				return publisher.writeCurrentLocked(operation, pointer)
			},
		},
		{
			name:   "JOURNAL",
			target: vNextPublicationJournalFile,
			point:  vNextPublicationAfterJournalSync,
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
			targetPath := filepath.Join(root, "acme", test.target)
			if test.target == vNextPublicationJournalFile {
				payload, err := vNextPublicationJSON(vNextGenerationJournal{New: pointer, State: "prepared"})
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(targetPath, payload, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			before, err := os.ReadFile(targetPath)
			if err != nil {
				t.Fatal(err)
			}

			var replacementPath, movedPath string
			hookHit := false
			guard, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
				if point != test.point || hookHit {
					return nil
				}
				hookHit = true
				replacementPath, movedPath = vNextReplacePublicationTemporaryForTest(t, root, []byte("unrelated replacement"))
				return nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			operation, err := guard.openOperation(context.Background(), syscall.LOCK_EX, true)
			if err != nil {
				t.Fatal(err)
			}
			err = test.write(guard, operation, pointer)
			operation.close()

			if !hookHit {
				t.Fatal("atomic control write did not reach replacement barrier")
			}
			if err == nil || !strings.Contains(err.Error(), "identity") {
				t.Fatalf("atomic %s replacement error = %v, want identity refusal", test.target, err)
			}
			if got, readErr := os.ReadFile(targetPath); readErr != nil || !bytes.Equal(got, before) {
				t.Fatalf("atomic %s replacement changed prior control: err=%v got=%q want=%q", test.target, readErr, got, before)
			}
			if got, readErr := os.ReadFile(replacementPath); readErr != nil || string(got) != "unrelated replacement" {
				t.Fatalf("atomic %s replacement moved unrelated object: err=%v got=%q", test.target, readErr, got)
			}
			if _, statErr := os.Stat(movedPath); statErr != nil {
				t.Fatalf("atomic %s replacement removed original temporary: %v", test.target, statErr)
			}
		})
	}
}

func TestVNextGenerationPublisherRefusesLateReplacedAtomicControlTemporary(t *testing.T) {
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
			targetPath := filepath.Join(root, "acme", test.target)
			if test.target == vNextPublicationJournalFile {
				payload, err := vNextPublicationJSON(vNextGenerationJournal{New: pointer, State: "prepared"})
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(targetPath, payload, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			before, err := os.ReadFile(targetPath)
			if err != nil {
				t.Fatal(err)
			}

			var replacementPath, movedPath string
			hookHit := false
			guard, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
				if point != vNextPublicationFaultPoint("after_control_source_identity") || hookHit {
					return nil
				}
				hookHit = true
				replacementPath, movedPath = vNextReplacePublicationTemporaryForTest(t, root, []byte("late unrelated replacement"))
				return nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			operation, err := guard.openOperation(context.Background(), syscall.LOCK_EX, true)
			if err != nil {
				t.Fatal(err)
			}
			err = test.write(guard, operation, pointer)
			operation.close()

			if !hookHit {
				t.Fatal("atomic control write did not reach post-identity rename barrier")
			}
			if err == nil || !strings.Contains(err.Error(), "identity") {
				t.Fatalf("late atomic %s replacement error = %v, want identity refusal", test.target, err)
			}
			if got, readErr := os.ReadFile(targetPath); readErr != nil || !bytes.Equal(got, before) {
				t.Fatalf("late atomic %s replacement changed prior control: err=%v got=%q want=%q", test.target, readErr, got, before)
			}
			if got, readErr := os.ReadFile(replacementPath); readErr != nil || string(got) != "late unrelated replacement" {
				t.Fatalf("late atomic %s replacement moved unrelated object: err=%v got=%q", test.target, readErr, got)
			}
			if _, statErr := os.Stat(movedPath); statErr != nil {
				t.Fatalf("late atomic %s replacement removed original temporary: %v", test.target, statErr)
			}
		})
	}
}

func vNextReplacePublicationTemporaryForTest(t *testing.T, root string, replacement []byte) (string, string) {
	t.Helper()
	connectorRoot := filepath.Join(root, "acme")
	entries, err := os.ReadDir(connectorRoot)
	if err != nil {
		t.Fatal(err)
	}
	temporary := ""
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".connectorgen-publication-") {
			if temporary != "" {
				t.Fatalf("multiple publication temporaries: %q and %q", temporary, entry.Name())
			}
			temporary = entry.Name()
		}
	}
	if temporary == "" {
		t.Fatal("publication replacement barrier found no temporary")
	}
	path := filepath.Join(connectorRoot, temporary)
	if info, statErr := os.Stat(path); statErr != nil {
		t.Fatal(statErr)
	} else if info.IsDir() {
		path = filepath.Join(path, vNextPublicationTemporaryFile)
	}
	moved := path + ".moved"
	if err := os.Rename(path, moved); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, replacement, 0o600); err != nil {
		t.Fatal(err)
	}
	return path, moved
}

func TestVNextGenerationPublisherRefusesReplacedValidatedStageCleanup(t *testing.T) {
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
	stagePath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, stageName)
	vNextWriteOwnedStageForTest(t, stagePath, pointer, stageName, "original")

	var movedPath string
	hookHit := false
	guard, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationBeforeStageRemoval || hookHit {
			return nil
		}
		hookHit = true
		movedPath = vNextReplaceOwnedStageForTest(t, stagePath, pointer, stageName, "replacement")
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	err = guard.Recover(context.Background())

	if !hookHit {
		t.Fatal("Recover() did not reach stage cleanup barrier")
	}
	if err == nil || !strings.Contains(err.Error(), "identity") {
		t.Fatalf("Recover() after stage replacement error = %v, want identity refusal", err)
	}
	for _, path := range []string{
		filepath.Join(movedPath, "sentinel.txt"),
		filepath.Join(stagePath, "sentinel.txt"),
	} {
		if got, readErr := os.ReadFile(path); readErr != nil || string(got) == "" {
			t.Fatalf("Recover() removed a validated-stage replacement object %q: err=%v got=%q", path, readErr, got)
		}
	}
}

func TestVNextGenerationPublisherRefusesReplacedValidatedGenerationCleanup(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	old, err := baseline.Publish(vNextPublicationArtifactsForTest("old", false))
	if err != nil {
		t.Fatal(err)
	}
	handle, err := baseline.Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	current, err := baseline.Publish(vNextPublicationArtifactsForTest("current", false))
	if err != nil {
		handle.Release()
		t.Fatal(err)
	}
	handle.Release()
	generationPath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, old.Generation)

	var movedPath string
	hookHit := false
	guard, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationBeforeGenerationRemoval || hookHit {
			return nil
		}
		hookHit = true
		movedPath = vNextReplaceGenerationForTest(t, generationPath)
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	operation, err := guard.openOperation(context.Background(), syscall.LOCK_EX, true)
	if err != nil {
		t.Fatal(err)
	}
	err = guard.pruneLocked(operation, current.Generation)
	operation.close()

	if !hookHit {
		t.Fatal("Prune() did not reach generation cleanup barrier")
	}
	if err == nil || !strings.Contains(err.Error(), "identity") {
		t.Fatalf("Prune() after generation replacement error = %v, want identity refusal", err)
	}
	for _, path := range []string{
		filepath.Join(movedPath, "metadata.json"),
		filepath.Join(generationPath, "metadata.json"),
	} {
		if got, readErr := os.ReadFile(path); readErr != nil || string(got) != `{"marker":"old"}` {
			t.Fatalf("Prune() removed a validated-generation replacement object %q: err=%v got=%q", path, readErr, got)
		}
	}
}

func TestVNextGenerationPublisherRefusesReplacedCommittedJournalCleanup(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := baseline.Publish(vNextPublicationArtifactsForTest("old", false)); err != nil {
		t.Fatal(err)
	}
	journalPath := filepath.Join(root, "acme", vNextPublicationJournalFile)
	movedPath := journalPath + ".moved"
	hookHit := false
	guard, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationBeforeControlRemoval || hookHit {
			return nil
		}
		hookHit = true
		payload, err := os.ReadFile(journalPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(journalPath, movedPath); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(journalPath, payload, 0o600); err != nil {
			t.Fatal(err)
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = guard.Publish(vNextPublicationArtifactsForTest("new", false))

	if !hookHit {
		t.Fatal("Publish() did not reach committed-journal barrier")
	}
	if err == nil || !strings.Contains(err.Error(), "identity") {
		t.Fatalf("Publish() after JOURNAL replacement error = %v, want identity refusal", err)
	}
	for _, path := range []string{journalPath, movedPath} {
		if payload, readErr := os.ReadFile(path); readErr != nil || len(payload) == 0 {
			t.Fatalf("Publish() removed a committed JOURNAL object %q: err=%v payload=%q", path, readErr, payload)
		}
	}
}

func vNextWriteOwnedStageForTest(t *testing.T, stagePath string, pointer vNextGenerationPointer, stageName, sentinel string) {
	t.Helper()
	if err := os.Mkdir(stagePath, 0o755); err != nil {
		t.Fatal(err)
	}
	marker, err := vNextPublicationJSON(vNextPublicationStageOwner{
		Version:    1,
		Connector:  "acme",
		Generation: pointer.Generation,
		Stage:      stageName,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stagePath, vNextPublicationStageOwnerFile), marker, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stagePath, "sentinel.txt"), []byte(sentinel), 0o600); err != nil {
		t.Fatal(err)
	}
}

func vNextReplaceOwnedStageForTest(t *testing.T, stagePath string, pointer vNextGenerationPointer, stageName, sentinel string) string {
	t.Helper()
	movedPath := filepath.Join(t.TempDir(), filepath.Base(stagePath))
	if err := os.Rename(stagePath, movedPath); err != nil {
		t.Fatal(err)
	}
	vNextWriteOwnedStageForTest(t, stagePath, pointer, stageName, sentinel)
	return movedPath
}

func vNextReplaceGenerationForTest(t *testing.T, generationPath string) string {
	t.Helper()
	movedPath := filepath.Join(t.TempDir(), filepath.Base(generationPath))
	if err := os.Rename(generationPath, movedPath); err != nil {
		t.Fatal(err)
	}
	vNextCopyDirectoryForTest(t, movedPath, generationPath)
	return movedPath
}

func vNextCopyDirectoryForTest(t *testing.T, source, destination string) {
	t.Helper()
	info, err := os.Stat(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(destination, info.Mode().Perm()); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(source, entry.Name())
		destinationPath := filepath.Join(destination, entry.Name())
		if entry.IsDir() {
			vNextCopyDirectoryForTest(t, sourcePath, destinationPath)
			continue
		}
		payload, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatal(err)
		}
		info, err := entry.Info()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(destinationPath, payload, info.Mode().Perm()); err != nil {
			t.Fatal(err)
		}
	}
}

func vNextPublicationHoldLockForTest(t *testing.T, root string) *os.File {
	t.Helper()
	lock, err := os.Open(filepath.Join(root, "acme"))
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		_ = lock.Close()
		t.Fatal(err)
	}
	return lock
}

func TestVNextGenerationPublisherRollsBackFailedActiveValidationWithoutOrphan(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	oldPointer, err := publisher.Publish(vNextPublicationArtifactsForTest("old", true))
	if err != nil {
		t.Fatalf("Publish(old) error = %v", err)
	}
	newSet := vNextPublicationArtifactsForTest("new", false)
	validationCalls := 0
	activeFailure := errors.New("active validation failure")
	newSet.Validate = func(fs.FS) error {
		validationCalls++
		if validationCalls == 2 {
			return activeFailure
		}
		return nil
	}
	if _, err := publisher.Publish(newSet); !errors.Is(err, activeFailure) {
		t.Fatalf("Publish(new) error = %v, want active validation failure", err)
	}
	if validationCalls != 2 {
		t.Fatalf("validation calls = %d, want staged then active validation", validationCalls)
	}
	entries, err := os.ReadDir(filepath.Join(root, "acme", vNextPublicationGenerationDirectory))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != oldPointer.Generation {
		t.Fatalf("failed active validation retained entries = %#v, want only restored old %q", entries, oldPointer.Generation)
	}
	active, err := publisher.Open(context.Background())
	if err != nil {
		t.Fatalf("Open() after rollback error = %v", err)
	}
	defer active.Release()
	assertVNextPublicationMarker(t, active, "old")
}

func TestVNextGenerationPublisherRefusesReplacedRollbackGenerationCleanup(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := baseline.Publish(vNextPublicationArtifactsForTest("old", true)); err != nil {
		t.Fatalf("Publish(old) error = %v", err)
	}
	newSet := vNextPublicationArtifactsForTest("new", false)
	validationCalls := 0
	activeFailure := errors.New("active validation failure")
	newSet.Validate = func(fs.FS) error {
		validationCalls++
		if validationCalls == 2 {
			return activeFailure
		}
		return nil
	}
	var movedPath string
	hookHit := false
	guard, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationBeforeGenerationRemoval || hookHit {
			return nil
		}
		hookHit = true
		generation := vNextPublicationGenerationID(newSet.Files)
		generationPath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, generation)
		movedPath = vNextReplaceGenerationForTest(t, generationPath)
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := guard.Publish(newSet); err == nil || !strings.Contains(err.Error(), "identity") {
		t.Fatalf("Publish(new) after rollback generation replacement error = %v, want identity refusal", err)
	}
	if validationCalls != 2 {
		t.Fatalf("validation calls = %d, want staged then active validation", validationCalls)
	}
	if !hookHit {
		t.Fatal("rollback generation cleanup barrier was not reached")
	}
	generation := vNextPublicationGenerationID(newSet.Files)
	for _, path := range []string{
		filepath.Join(movedPath, "metadata.json"),
		filepath.Join(root, "acme", vNextPublicationGenerationDirectory, generation, "metadata.json"),
	} {
		payload, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read preserved generation artifact %q: %v", path, err)
		}
		if string(payload) != `{"marker":"new"}` {
			t.Fatalf("preserved generation artifact %q = %s, want new marker", path, payload)
		}
	}
}

func TestVNextGenerationPublisherRefusesLateReplacedValidatedStageCleanup(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	pointer, err := baseline.Publish(vNextPublicationArtifactsForTest("active", false))
	if err != nil {
		t.Fatal(err)
	}
	stageName := ".stage-late-owned"
	stagePath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, stageName)
	vNextWriteOwnedStageForTest(t, stagePath, pointer, stageName, "original")

	hookHit := false
	var movedPath string
	guard, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationFaultPoint("after_stage_removal_identity") || hookHit {
			return nil
		}
		hookHit = true
		movedPath = vNextReplaceDirectoryAtPathForTest(t, stagePath)
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	err = guard.Recover(context.Background())

	if !hookHit {
		t.Fatal("Recover() did not reach late stage-removal identity barrier")
	}
	if err == nil || !strings.Contains(err.Error(), "identity") {
		t.Fatalf("Recover() after late stage replacement error = %v, want identity refusal", err)
	}
	if got, readErr := os.ReadFile(filepath.Join(movedPath, "sentinel.txt")); readErr != nil || string(got) != "original" {
		t.Fatalf("late stage replacement did not preserve original: err=%v got=%q", readErr, got)
	}
	if info, statErr := os.Stat(stagePath); statErr != nil || !info.IsDir() {
		t.Fatalf("late stage replacement did not preserve replacement: err=%v info=%v", statErr, info)
	}
}

func TestVNextGenerationPublisherRefusesLateReplacedValidatedGenerationCleanup(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	old, err := baseline.Publish(vNextPublicationArtifactsForTest("old", false))
	if err != nil {
		t.Fatal(err)
	}
	handle, err := baseline.Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	current, err := baseline.Publish(vNextPublicationArtifactsForTest("current", false))
	handle.Release()
	if err != nil {
		t.Fatal(err)
	}
	generationPath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, old.Generation)

	hookHit := false
	var movedPath string
	guard, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationFaultPoint("after_generation_removal_identity") || hookHit {
			return nil
		}
		hookHit = true
		movedPath = vNextReplaceDirectoryAtPathForTest(t, generationPath)
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	operation, err := guard.openOperation(context.Background(), syscall.LOCK_EX, true)
	if err != nil {
		t.Fatal(err)
	}
	err = guard.pruneLocked(operation, current.Generation)
	operation.close()

	if !hookHit {
		t.Fatal("Prune() did not reach late generation-removal identity barrier")
	}
	if err == nil || !strings.Contains(err.Error(), "identity") {
		t.Fatalf("Prune() after late generation replacement error = %v, want identity refusal", err)
	}
	if got, readErr := os.ReadFile(filepath.Join(movedPath, "metadata.json")); readErr != nil || string(got) != `{"marker":"old"}` {
		t.Fatalf("late generation replacement did not preserve original: err=%v got=%q", readErr, got)
	}
	if info, statErr := os.Stat(generationPath); statErr != nil || !info.IsDir() {
		t.Fatalf("late generation replacement did not preserve replacement: err=%v info=%v", statErr, info)
	}
}

func TestVNextGenerationPublisherRefusesLateReplacedGenerationLeaseCleanup(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	old, err := baseline.Publish(vNextPublicationArtifactsForTest("old", false))
	if err != nil {
		t.Fatal(err)
	}
	handle, err := baseline.Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	current, err := baseline.Publish(vNextPublicationArtifactsForTest("current", false))
	handle.Release()
	if err != nil {
		t.Fatal(err)
	}
	leasePath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, old.Generation, vNextPublicationLeaseFile)
	movedPath := leasePath + ".late-original"

	hookHit := false
	guard, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationFaultPoint("after_generation_lease_identity") || hookHit {
			return nil
		}
		hookHit = true
		if err := os.Rename(leasePath, movedPath); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(leasePath, []byte("late lease replacement"), 0o600); err != nil {
			t.Fatal(err)
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	operation, err := guard.openOperation(context.Background(), syscall.LOCK_EX, true)
	if err != nil {
		t.Fatal(err)
	}
	err = guard.pruneLocked(operation, current.Generation)
	operation.close()

	if !hookHit {
		t.Fatal("Prune() did not reach late generation-lease identity barrier")
	}
	if err == nil || !strings.Contains(err.Error(), "identity") {
		t.Fatalf("Prune() after late lease replacement error = %v, want identity refusal", err)
	}
	for _, path := range []string{leasePath, movedPath} {
		if _, statErr := os.Stat(path); statErr != nil {
			t.Fatalf("Prune() removed a late generation-lease object %q: %v", path, statErr)
		}
	}
}

func TestVNextGenerationPublisherRefusesLateReplacedRollbackGenerationCleanup(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := baseline.Publish(vNextPublicationArtifactsForTest("old", true)); err != nil {
		t.Fatal(err)
	}
	newSet := vNextPublicationArtifactsForTest("new", false)
	validationCalls := 0
	activeFailure := errors.New("active validation failure")
	newSet.Validate = func(fs.FS) error {
		validationCalls++
		if validationCalls == 2 {
			return activeFailure
		}
		return nil
	}
	generation := vNextPublicationGenerationID(newSet.Files)
	generationPath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, generation)
	hookHit := false
	var movedPath string
	guard, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationFaultPoint("after_generation_removal_identity") || hookHit {
			return nil
		}
		hookHit = true
		movedPath = vNextReplaceDirectoryAtPathForTest(t, generationPath)
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	_, publishErr := guard.Publish(newSet)
	if !hookHit {
		t.Fatal("rollback generation cleanup did not reach late identity barrier")
	}
	if publishErr == nil || !strings.Contains(publishErr.Error(), "identity") {
		t.Fatalf("Publish(new) after late rollback generation replacement error = %v, want identity refusal", publishErr)
	}
	if validationCalls != 2 {
		t.Fatalf("validation calls = %d, want staged then active validation", validationCalls)
	}
	if got, readErr := os.ReadFile(filepath.Join(movedPath, "metadata.json")); readErr != nil || string(got) != `{"marker":"new"}` {
		t.Fatalf("late rollback generation replacement did not preserve original: err=%v got=%q", readErr, got)
	}
	if info, statErr := os.Stat(generationPath); statErr != nil || !info.IsDir() {
		t.Fatalf("late rollback generation replacement did not preserve replacement: err=%v info=%v", statErr, info)
	}
}

func TestVNextGenerationPublisherRefusesLateReplacedControlCleanup(t *testing.T) {
	for _, name := range []string{vNextPublicationCurrentFile, vNextPublicationJournalFile} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			pointer, err := baseline.Publish(vNextPublicationArtifactsForTest("active", false))
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(root, "acme", name)
			if name == vNextPublicationJournalFile {
				payload, err := vNextPublicationJSON(vNextGenerationJournal{New: pointer, State: "prepared"})
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, payload, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			movedPath := path + ".late-original"
			hookHit := false
			guard, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
				if point != vNextPublicationFaultPoint("after_control_removal_identity") || hookHit {
					return nil
				}
				hookHit = true
				if err := os.Rename(path, movedPath); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("late control replacement"), 0o600); err != nil {
					t.Fatal(err)
				}
				return nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			operation, err := guard.openOperation(context.Background(), syscall.LOCK_EX, true)
			if err != nil {
				t.Fatal(err)
			}
			if name == vNextPublicationCurrentFile {
				err = guard.removeCurrentLocked(operation)
			} else {
				err = guard.removeJournalLocked(operation)
			}
			operation.close()

			if !hookHit {
				t.Fatalf("remove %s did not reach late control-removal identity barrier", name)
			}
			if err == nil || !strings.Contains(err.Error(), "identity") {
				t.Fatalf("remove %s after late replacement error = %v, want identity refusal", name, err)
			}
			for _, preserved := range []string{path, movedPath} {
				if _, statErr := os.Stat(preserved); statErr != nil {
					t.Fatalf("remove %s deleted a late control object %q: %v", name, preserved, statErr)
				}
			}
		})
	}
}

func vNextReplaceDirectoryAtPathForTest(t *testing.T, path string) string {
	t.Helper()
	moved := path + ".late-original"
	if err := os.Rename(path, moved); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
	return moved
}

func TestVNextPublicationGenerationIDDelimitsArtifactNamesAndBytes(t *testing.T) {
	one := map[string][]byte{"a": []byte("payload\x00b\x00other")}
	two := map[string][]byte{"a": []byte("payload"), "b": []byte("other")}
	if gotOne, gotTwo := vNextPublicationGenerationID(one), vNextPublicationGenerationID(two); gotOne == gotTwo {
		t.Fatalf("ambiguous artifact sets share generation ID %q", gotOne)
	}
}
func TestVNextGenerationPublisherRejectsUnsafeArtifactPathsBeforeWriting(t *testing.T) {
	for _, name := range []string{"../outside.json", "schemas/../../outside.json", "/absolute.json", ".hidden.json", "schemas/.hidden.json", "schemas\\windows.json", "nul\x00name.json", "integrity.json", ".lease"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := publisher.Publish(vNextPublicationArtifacts{Files: map[string][]byte{name: []byte(`{}`)}}); err == nil {
				t.Fatalf("Publish() accepted unsafe artifact path %q", name)
			}
			if _, err := os.Stat(filepath.Join(root, "acme", vNextPublicationGenerationDirectory)); !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("Publish(%q) created generation root before path refusal: %v", name, err)
			}
		})
	}
}

func TestVNextGenerationPublisherRejectsTamperedActiveDigestAndUnexpectedMember(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	artifacts := vNextPublicationArtifactsForTest("old", true)
	pointer, err := publisher.Publish(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	generationRoot := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, pointer.Generation)
	if err := os.WriteFile(filepath.Join(generationRoot, "unexpected.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := publisher.Check(artifacts); err == nil {
		t.Fatal("Check() accepted unexpected executable generation member")
	}
	if err := os.Remove(filepath.Join(generationRoot, "unexpected.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(generationRoot, "metadata.json"), []byte(`{"marker":"tampered"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := publisher.Check(artifacts); err == nil {
		t.Fatal("Check() accepted an artifact whose digest differs from integrity")
	}
	if _, err := publisher.Open(context.Background()); err == nil {
		t.Fatal("Open() accepted a tampered active generation")
	}
}

func TestVNextGenerationPublisherRejectsSymlinkedCurrentWithoutReadingTarget(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	artifacts := vNextPublicationArtifactsForTest("active", false)
	if _, err := publisher.Publish(artifacts); err != nil {
		t.Fatal(err)
	}
	currentPath := filepath.Join(root, "acme", vNextPublicationCurrentFile)
	current, err := os.ReadFile(currentPath)
	if err != nil {
		t.Fatal(err)
	}
	authorOwned := filepath.Join(root, "author-owned-current.json")
	if err := os.WriteFile(authorOwned, current, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(currentPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(authorOwned, currentPath); err != nil {
		t.Fatal(err)
	}

	if err := publisher.Check(artifacts); err == nil {
		t.Fatal("Check() accepted a symlinked CURRENT")
	}
	if got, err := os.ReadFile(authorOwned); err != nil || !bytes.Equal(got, current) {
		t.Fatalf("Check() altered the symlink target: err=%v got=%q", err, got)
	}
}

func TestVNextGenerationPublisherBoundsControlMetadataReads(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	artifacts := vNextPublicationArtifactsForTest("active", false)
	if _, err := publisher.Publish(artifacts); err != nil {
		t.Fatal(err)
	}
	currentPath := filepath.Join(root, "acme", vNextPublicationCurrentFile)
	if err := os.WriteFile(currentPath, bytes.Repeat([]byte("x"), vNextPublicationControlMaxBytes+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := publisher.Check(artifacts); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("Check() error = %v, want bounded control metadata refusal", err)
	}
}

func TestVNextGenerationPublisherStagesDirectlyUnderGenerationRoot(t *testing.T) {
	root := t.TempDir()
	stop := errors.New("stop after observing stage")
	observed := false
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationBeforeStageDirectory {
			return nil
		}
		generationsRoot := filepath.Join(root, "acme", vNextPublicationGenerationDirectory)
		entries, err := os.ReadDir(generationsRoot)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if !strings.HasPrefix(entry.Name(), ".stage-") {
				continue
			}
			stage := filepath.Join(generationsRoot, entry.Name())
			relative, err := filepath.Rel(generationsRoot, stage)
			if err != nil || filepath.Dir(relative) != "." || !entry.IsDir() {
				t.Fatalf("stage %q is not a direct generation-root child: relative=%q err=%v", stage, relative, err)
			}
			observed = true
		}
		return stop
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := publisher.Publish(vNextPublicationArtifactsForTest("active", false)); !errors.Is(err, stop) {
		t.Fatalf("Publish() error = %v, want %v", err, stop)
	}
	if !observed {
		t.Fatal("Publish() did not create a same-parent staging directory")
	}
}

func TestVNextGenerationPublisherRefusesToPruneUnownedGeneration(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := publisher.Publish(vNextPublicationArtifactsForTest("active", false)); err != nil {
		t.Fatal(err)
	}
	foreignGeneration := "g-" + strings.Repeat("a", 64)
	foreignRoot := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, foreignGeneration)
	if err := os.Mkdir(foreignRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(foreignRoot, vNextPublicationLeaseFile), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	authorOwned := filepath.Join(foreignRoot, "author-owned.txt")
	if err := os.WriteFile(authorOwned, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := publisher.Prune(context.Background()); err == nil {
		t.Fatal("Prune() removed a generation without validated publisher ownership")
	}
	if got, err := os.ReadFile(authorOwned); err != nil || string(got) != "keep" {
		t.Fatalf("Prune() deleted author-owned file: err=%v got=%q", err, got)
	}
}

func TestVNextGenerationPublisherRefusesSymlinkedGenerationRootWithoutDeletingTarget(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := publisher.Publish(vNextPublicationArtifactsForTest("active", false)); err != nil {
		t.Fatal(err)
	}
	generationsRoot := filepath.Join(root, "acme", vNextPublicationGenerationDirectory)
	if err := os.RemoveAll(generationsRoot); err != nil {
		t.Fatal(err)
	}
	authorOwned := filepath.Join(root, "author-owned-generation-root")
	stage := filepath.Join(authorOwned, ".stage-keep")
	if err := os.MkdirAll(stage, 0o755); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(stage, "sentinel.txt")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(authorOwned, generationsRoot); err != nil {
		t.Fatal(err)
	}

	if err := publisher.Recover(context.Background()); err == nil {
		t.Fatal("Recover() accepted a symlinked generation root")
	}
	if got, err := os.ReadFile(sentinel); err != nil || string(got) != "keep" {
		t.Fatalf("Recover() deleted author-owned generation-root target: err=%v got=%q", err, got)
	}
}
func TestRunLockRenderPublishesOnlyClosedGeneration(t *testing.T) {
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
	lockPath := filepath.Join(connectorRoot, "source.lock.json")
	if err := os.WriteFile(lockPath, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	const sentinel = "authoring-adjacent flat output"
	flatMetadata := filepath.Join(connectorRoot, "metadata.json")
	if err := os.WriteFile(flatMetadata, []byte(sentinel), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	if code := runLockRender([]string{"lock-render", lock.Connector, "--defs", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("runLockRender() = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := runLockRender([]string{"lock-render", lock.Connector, "--defs", root, "--check"}, &stdout, &stderr); code != 0 {
		t.Fatalf("runLockRender(--check) = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if got, err := os.ReadFile(lockPath); err != nil || !bytes.Equal(got, raw) {
		t.Fatalf("lock render changed author-owned source.lock.json: err=%v got=%s", err, got)
	}
	if got, err := os.ReadFile(flatMetadata); err != nil || string(got) != sentinel {
		t.Fatalf("lock render wrote flat artifact instead of closed generation: err=%v got=%q", err, got)
	}

	publisher, err := newVNextGenerationPublisher(root, lock.Connector, vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	active, err := publisher.Open(context.Background())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer active.Release()
	for _, name := range []string{"metadata.json", "spec.json", "manifest.json", "provenance.json", "atlas.json", "index.json", "proof.json"} {
		if _, err := active.ReadFile(name); err != nil {
			t.Fatalf("closed published generation lacks %s: %v", name, err)
		}
	}
	artifacts, err := vNextPublicationArtifactsForStage(raw, lock.Connector, mustCanonicalVNextLockForTest(t, lock).Staged)
	if err != nil {
		t.Fatalf("vNextPublicationArtifactsForStage() error = %v", err)
	}
	if err := publisher.Check(artifacts); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestVNextGenerationPublisherCheckRefusesPhysicalClosedSetMutation(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	artifacts := vNextPublicationArtifactsForTest("active", false)
	pointer, err := publisher.Publish(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	unexpected := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, pointer.Generation, "unexpected.json")
	if err := os.WriteFile(unexpected, []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := publisher.Check(artifacts); err == nil {
		t.Fatal("Check() accepted a physically mutated closed generation")
	}
	if _, err := os.Stat(unexpected); err != nil {
		t.Fatalf("Check() removed physical closed-set mutation: %v", err)
	}
}

func TestVNextGenerationPublisherRecoversPublisherWrittenDurableStage(t *testing.T) {
	stop := errors.New("stop before stage rename")
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point == vNextPublicationBeforeStageRename {
			return stop
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	artifacts := vNextPublicationArtifactsForTest("active", false)
	if _, err := publisher.Publish(artifacts); !errors.Is(err, stop) {
		t.Fatalf("Publish() error = %v, want %v", err, stop)
	}
	generationsRoot := filepath.Join(root, "acme", vNextPublicationGenerationDirectory)
	entries, err := os.ReadDir(generationsRoot)
	if err != nil {
		t.Fatal(err)
	}
	stageName := ""
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".stage-") {
			stageName = entry.Name()
		}
	}
	if stageName == "" {
		t.Fatal("Publish() did not retain its pre-rename durable stage")
	}
	stagePath := filepath.Join(generationsRoot, stageName)
	payload, err := os.ReadFile(filepath.Join(stagePath, vNextPublicationStageOwnerFile))
	if err != nil {
		t.Fatal(err)
	}
	var owner vNextPublicationStageOwner
	if err := vNextPublicationDecode(payload, &owner); err != nil {
		t.Fatal(err)
	}
	if owner.Version != 1 || owner.Connector != "acme" || owner.Generation != vNextPublicationGenerationID(artifacts.Files) || owner.Stage != stageName {
		t.Fatalf("publisher-written durable stage marker = %#v", owner)
	}
	if err := publisher.Recover(context.Background()); err != nil {
		t.Fatalf("Recover() error = %v", err)
	}
	if _, err := os.Lstat(stagePath); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Recover() retained publisher-written durable stage: %v", err)
	}
}

func TestVNextGenerationPublisherPhysicallyPreflightsImplementedStagedCommand(t *testing.T) {
	lock, raw, stage := vNextPhysicalPreflightStageForTest(t, false)
	artifacts, err := vNextPublicationArtifactsForStage(raw, lock.Connector, stage)
	if err != nil {
		t.Fatalf("vNextPublicationArtifactsForStage() error = %v", err)
	}
	publisher, err := newVNextGenerationPublisher(t.TempDir(), lock.Connector, vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := publisher.Publish(artifacts); err != nil {
		t.Fatalf("Publish() physical implemented-command preflight error = %v", err)
	}
	if err := publisher.Check(artifacts); err != nil {
		t.Fatalf("Check() physical implemented-command preflight error = %v", err)
	}
}

func TestVNextGenerationPublisherRefusesPhysicallyStagedCommandPreflight(t *testing.T) {
	lock, raw, stage := vNextPhysicalPreflightStageForTest(t, true)
	artifacts, err := vNextPublicationArtifactsForStage(raw, lock.Connector, stage)
	if err != nil {
		t.Fatalf("vNextPublicationArtifactsForStage() error = %v", err)
	}
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, lock.Connector, vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := publisher.Publish(artifacts); err == nil || !strings.Contains(err.Error(), `preflight staged command "widgets get"`) {
		t.Fatalf("Publish() physical staged command preflight error = %v, want staged preflight refusal", err)
	}
	connectorRoot := filepath.Join(root, lock.Connector)
	for _, name := range []string{vNextPublicationCurrentFile, vNextPublicationJournalFile} {
		if _, err := os.Lstat(filepath.Join(connectorRoot, name)); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("physical staged preflight refusal left %s: %v", name, err)
		}
	}
	entries, err := os.ReadDir(filepath.Join(connectorRoot, vNextPublicationGenerationDirectory))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("physical staged preflight refusal left generation members: %#v", entries)
	}
}

func vNextPhysicalPreflightStageForTest(t *testing.T, invalidCommand bool) (vNextSourceLock, []byte, vNextStagedGeneration) {
	t.Helper()
	lock := operationDirectReadLockForSemanticAdmissionTest()
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	stage := mustCanonicalVNextLockForTest(t, lock).Staged
	if !invalidCommand {
		return lock, raw, stage
	}
	outputs := make(map[string][]byte, len(stage.Outputs))
	for name, payload := range stage.Outputs {
		outputs[name] = append([]byte(nil), payload...)
	}
	outputs["cli_surface.json"] = []byte(`{"usage":"pm acme <command>","tagline":"Acme commands","commands":[{"path":"widgets get","summary":"Get widgets","intent":"direct_read","availability":"implemented","operation":"widgets.get","api_surface":[{"method":"GET","path":"/widgets"}],"output_policy":"json_redacted","flags":[{"name":"bogus","type":"string","maps_to":"unsupported.bogus"}]}]}`)
	bundle, err := engine.Load(newVNextExecutionFS(lock.Connector, outputs), lock.Connector)
	if err != nil {
		t.Fatalf("load staged physical preflight fixture: %v", err)
	}
	selection, err := vNextSelectedRuntime(lock.Connector)
	if err != nil {
		t.Fatalf("select staged physical preflight runtime: %v", err)
	}
	runtime := engine.New(bundle, selection.Hooks)
	index, err := manifestindex.New([]manifestindex.Entry{vNextManifestEntry(bundle, runtime, selection)}, 1)
	if err != nil {
		t.Fatalf("index staged physical preflight fixture: %v", err)
	}
	stage.Outputs = outputs
	stage.Identity = bundle.Identity
	stage.Manifest = vNextManifestEntry(bundle, runtime, selection)
	stage.Index = index
	return lock, raw, stage
}

func TestRunLockRenderContextCancelsContendedPublicationAndRetries(t *testing.T) {
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
	args := []string{"lock-render", lock.Connector, "--defs", root, "--check"}
	if code := runLockRender([]string{"lock-render", lock.Connector, "--defs", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("runLockRender() = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	currentPath := filepath.Join(connectorRoot, vNextPublicationCurrentFile)
	currentBefore, err := os.ReadFile(currentPath)
	if err != nil {
		t.Fatal(err)
	}
	held := vNextPublicationHoldLockForTest(t, root)
	defer func() { unlockVNextPublicationFile(held) }()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	stdout.Reset()
	stderr.Reset()
	code := runLockRenderContext(ctx, args, &stdout, &stderr)
	cancel()
	if code != 1 || !strings.Contains(stderr.String(), context.DeadlineExceeded.Error()) {
		t.Fatalf("runLockRenderContext() = %d, want cancelled publication; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if currentAfter, err := os.ReadFile(currentPath); err != nil || !bytes.Equal(currentAfter, currentBefore) {
		t.Fatalf("cancelled lock-render changed CURRENT: err=%v current=%q", err, currentAfter)
	}
	if _, err := os.Lstat(filepath.Join(connectorRoot, vNextPublicationJournalFile)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("cancelled lock-render changed JOURNAL: %v", err)
	}
	unlockVNextPublicationFile(held)
	held = nil
	stdout.Reset()
	stderr.Reset()
	if code := runLockRenderContext(context.Background(), args, &stdout, &stderr); code != 0 {
		t.Fatalf("runLockRenderContext retry = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

func TestVNextGenerationPublisherRecoversStagedOrphanBeforePreparedJournal(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	old := vNextPublicationArtifactsForTest("old", true)
	oldPointer, err := publisher.Publish(old)
	if err != nil {
		t.Fatalf("Publish(old) error = %v", err)
	}

	crash := errors.New("injected crash")
	failing, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point == vNextPublicationBeforeJournalSync {
			return crash
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := failing.Publish(vNextPublicationArtifactsForTest("new", false)); !errors.Is(err, crash) {
		t.Fatalf("Publish(new) error = %v, want injected crash", err)
	}

	recovered, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if err := recovered.Recover(context.Background()); err != nil {
		t.Fatalf("Recover() error = %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(root, "acme", vNextPublicationGenerationDirectory))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != oldPointer.Generation {
		t.Fatalf("Recover() retained orphan generation entries = %#v, want only %q", entries, oldPointer.Generation)
	}
	active, err := recovered.Open(context.Background())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer active.Release()
	assertVNextPublicationMarker(t, active, "old")
}

func TestVNextGenerationPublisherWritesPreparedJournalBeforeFinalRename(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	oldPointer, err := publisher.Publish(vNextPublicationArtifactsForTest("old", true))
	if err != nil {
		t.Fatalf("Publish(old) error = %v", err)
	}
	crash := errors.New("injected crash")
	failing, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point == vNextPublicationBeforeStageRename {
			return crash
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	newSet := vNextPublicationArtifactsForTest("new", false)
	newGeneration := vNextPublicationGenerationID(newSet.Files)
	if _, err := failing.Publish(newSet); !errors.Is(err, crash) {
		t.Fatalf("Publish(new) error = %v, want injected crash", err)
	}
	connectorRoot := filepath.Join(root, "acme")
	journalPayload, err := os.ReadFile(filepath.Join(connectorRoot, vNextPublicationJournalFile))
	if err != nil {
		t.Fatalf("read prepared JOURNAL: %v", err)
	}
	var journal vNextGenerationJournal
	if err := vNextPublicationDecode(journalPayload, &journal); err != nil {
		t.Fatalf("decode prepared JOURNAL: %v", err)
	}
	if journal.State != "prepared" || journal.New.Generation != newGeneration {
		t.Fatalf("prepared JOURNAL = %#v, want new generation %q", journal, newGeneration)
	}
	if _, err := os.Stat(filepath.Join(connectorRoot, vNextPublicationGenerationDirectory, newGeneration)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("final generation exists before final rename: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(connectorRoot, vNextPublicationGenerationDirectory))
	if err != nil {
		t.Fatal(err)
	}
	hasStage := false
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".stage-") {
			hasStage = true
			break
		}
	}
	if !hasStage {
		t.Fatalf("prepared journal left no owned stage: %#v", entries)
	}
	if err := publisher.Recover(context.Background()); err != nil {
		t.Fatalf("Recover() error = %v", err)
	}
	active, err := publisher.Open(context.Background())
	if err != nil {
		t.Fatalf("Open() after recovery error = %v", err)
	}
	defer active.Release()
	if active.pointer != oldPointer {
		t.Fatalf("active pointer after recovery = %#v, want %#v", active.pointer, oldPointer)
	}
	assertVNextPublicationMarker(t, active, "old")
}

func TestVNextGenerationPublisherRecoversPreparedJournalAfterFinalRename(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	oldPointer, err := publisher.Publish(vNextPublicationArtifactsForTest("old", true))
	if err != nil {
		t.Fatalf("Publish(old) error = %v", err)
	}
	crash := errors.New("injected crash")
	failing, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point == vNextPublicationAfterStageRename {
			return crash
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	newSet := vNextPublicationArtifactsForTest("new", false)
	newGeneration := vNextPublicationGenerationID(newSet.Files)
	if _, err := failing.Publish(newSet); !errors.Is(err, crash) {
		t.Fatalf("Publish(new) error = %v, want injected crash", err)
	}
	connectorRoot := filepath.Join(root, "acme")
	journalPayload, err := os.ReadFile(filepath.Join(connectorRoot, vNextPublicationJournalFile))
	if err != nil {
		t.Fatalf("read prepared JOURNAL: %v", err)
	}
	var journal vNextGenerationJournal
	if err := vNextPublicationDecode(journalPayload, &journal); err != nil {
		t.Fatalf("decode prepared JOURNAL: %v", err)
	}
	if journal.State != "prepared" || journal.New.Generation != newGeneration {
		t.Fatalf("prepared JOURNAL = %#v, want new generation %q", journal, newGeneration)
	}
	if _, err := os.Stat(filepath.Join(connectorRoot, vNextPublicationGenerationDirectory, newGeneration)); err != nil {
		t.Fatalf("final generation missing after final rename: %v", err)
	}
	currentPayload, err := os.ReadFile(filepath.Join(connectorRoot, vNextPublicationCurrentFile))
	if err != nil {
		t.Fatal(err)
	}
	var current vNextGenerationPointer
	if err := vNextPublicationDecode(currentPayload, &current); err != nil {
		t.Fatal(err)
	}
	if current != oldPointer {
		t.Fatalf("CURRENT after final rename = %#v, want prior %#v", current, oldPointer)
	}
	if err := publisher.Recover(context.Background()); err != nil {
		t.Fatalf("Recover() error = %v", err)
	}
	if publisher.GenerationExists(newGeneration) {
		t.Fatalf("Recover() retained rejected generation %q", newGeneration)
	}
	active, err := publisher.Open(context.Background())
	if err != nil {
		t.Fatalf("Open() after recovery error = %v", err)
	}
	defer active.Release()
	if active.pointer != oldPointer {
		t.Fatalf("active pointer after recovery = %#v, want %#v", active.pointer, oldPointer)
	}
	assertVNextPublicationMarker(t, active, "old")
}

func TestVNextGenerationPublisherRecoversEveryDurableCutPoint(t *testing.T) {
	points := []vNextPublicationFaultPoint{
		vNextPublicationBeforeFileSync,
		vNextPublicationAfterFileSync,
		vNextPublicationBeforeStageDirectory,
		vNextPublicationAfterStageDirectory,
		vNextPublicationBeforeJournalSync,
		vNextPublicationAfterJournalSync,
		vNextPublicationBeforeStageRename,
		vNextPublicationAfterStageRename,
		vNextPublicationBeforeCurrentTempSync,
		vNextPublicationAfterCurrentTempSync,
		vNextPublicationBeforeCurrentRename,
		vNextPublicationAfterCurrentRename,
		vNextPublicationBeforeCurrentParent,
		vNextPublicationAfterCurrentParent,
		vNextPublicationBeforeActiveValidate,
		vNextPublicationAfterActiveValidate,
		vNextPublicationBeforeCommitSync,
		vNextPublicationAfterCommitSync,
		vNextPublicationBeforePrune,
		vNextPublicationAfterPrune,
	}
	for _, point := range points {
		t.Run(string(point), func(t *testing.T) {
			root := t.TempDir()
			publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := publisher.Publish(vNextPublicationArtifactsForTest("old", true)); err != nil {
				t.Fatalf("Publish(old) error = %v", err)
			}
			crash := errors.New("injected crash")
			failing, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(candidate vNextPublicationFaultPoint) error {
				if candidate == point {
					return crash
				}
				return nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := failing.Publish(vNextPublicationArtifactsForTest("new", false)); !errors.Is(err, crash) {
				t.Fatalf("Publish(new) error = %v, want injected crash at %s", err, point)
			}

			recovered, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			if err := recovered.Recover(context.Background()); err != nil {
				t.Fatalf("Recover() after %s error = %v", point, err)
			}
			active, err := recovered.Open(context.Background())
			if err != nil {
				t.Fatalf("Open() after %s error = %v", point, err)
			}
			markerPayload, err := active.ReadFile("metadata.json")
			active.Release()
			if err != nil {
				t.Fatalf("ReadFile(metadata.json) after %s error = %v", point, err)
			}
			marker := ""
			switch string(markerPayload) {
			case `{"marker":"old"}`:
				marker = "old"
			case `{"marker":"new"}`:
				marker = "new"
			default:
				t.Fatalf("active marker after %s = %s, want complete old or new generation", point, markerPayload)
			}
			if err := recovered.Check(vNextPublicationArtifactsForTest(marker, marker == "old")); err != nil {
				t.Fatalf("Check(%s) after %s error = %v", marker, point, err)
			}
		})
	}
}

func TestVNextGenerationPublisherSerializesWritersAndReadersSeeWholeGeneration(t *testing.T) {
	root := t.TempDir()
	firstAtRename := make(chan struct{})
	releaseFirst := make(chan struct{})
	first, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point == vNextPublicationBeforeCurrentRename {
			close(firstAtRename)
			<-releaseFirst
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	secondEntered := make(chan struct{})
	second, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point == vNextPublicationBeforeFileSync {
			select {
			case <-secondEntered:
			default:
				close(secondEntered)
			}
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}

	firstResult := make(chan error, 1)
	go func() {
		_, err := first.Publish(vNextPublicationArtifactsForTest("first", true))
		firstResult <- err
	}()
	<-firstAtRename

	secondStarted := make(chan struct{})
	secondResult := make(chan error, 1)
	go func() {
		close(secondStarted)
		_, err := second.Publish(vNextPublicationArtifactsForTest("second", false))
		secondResult <- err
	}()
	<-secondStarted
	select {
	case <-secondEntered:
		t.Fatal("second writer reached staging while first writer held connector lock")
	default:
	}

	readerResult := make(chan string, 1)
	readerErr := make(chan error, 1)
	go func() {
		reader, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
		if err != nil {
			readerErr <- err
			return
		}
		handle, err := reader.Open(context.Background())
		if err != nil {
			readerErr <- err
			return
		}
		defer handle.Release()
		metadata, err := handle.ReadFile("metadata.json")
		if err != nil {
			readerErr <- err
			return
		}
		index, err := handle.ReadFile("index.json")
		if err != nil {
			readerErr <- err
			return
		}
		if string(metadata) != string(index) {
			readerErr <- errors.New("reader observed mixed metadata/index generation")
			return
		}
		readerResult <- string(metadata)
	}()

	close(releaseFirst)
	if err := <-firstResult; err != nil {
		t.Fatalf("first Publish() error = %v", err)
	}
	<-secondEntered
	if err := <-secondResult; err != nil {
		t.Fatalf("second Publish() error = %v", err)
	}
	select {
	case err := <-readerErr:
		t.Fatalf("reader error = %v", err)
	case marker := <-readerResult:
		if marker != `{"marker":"first"}` && marker != `{"marker":"second"}` {
			t.Fatalf("reader marker = %s, want whole first or second generation", marker)
		}
	}

	checker, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	final, err := checker.Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer final.Release()
	assertVNextPublicationMarker(t, final, "second")
	if err := checker.Check(vNextPublicationArtifactsForTest("second", false)); err != nil {
		t.Fatalf("Check(second) error = %v", err)
	}
}

func vNextRewriteGenerationIdentityForTest(t *testing.T, root string, pointer vNextGenerationPointer) vNextGenerationPointer {
	t.Helper()
	renamed := vNextGenerationPointer{Generation: "g-" + strings.Repeat("f", 64)}
	oldRoot := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, pointer.Generation)
	newRoot := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, renamed.Generation)
	if err := os.Mkdir(newRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(oldRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			t.Fatalf("test fixture has unexpected nested generation directory %q", entry.Name())
		}
		payload, err := os.ReadFile(filepath.Join(oldRoot, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(newRoot, entry.Name()), payload, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	integrityPath := filepath.Join(newRoot, vNextPublicationIntegrityFile)
	integrityRaw, err := os.ReadFile(integrityPath)
	if err != nil {
		t.Fatal(err)
	}
	var integrity vNextGenerationIntegrity
	if err := json.Unmarshal(integrityRaw, &integrity); err != nil {
		t.Fatal(err)
	}
	integrity.Generation = renamed.Generation
	integrityRaw, err = vNextPublicationJSON(integrity)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(integrityPath, integrityRaw, 0o600); err != nil {
		t.Fatal(err)
	}
	renamed.IntegrityDigest = vNextPublicationDigest(integrityRaw)

	ownerPath := filepath.Join(newRoot, vNextPublicationStageOwnerFile)
	ownerRaw, err := os.ReadFile(ownerPath)
	if err != nil {
		t.Fatal(err)
	}
	var owner vNextPublicationStageOwner
	if err := json.Unmarshal(ownerRaw, &owner); err != nil {
		t.Fatal(err)
	}
	owner.Generation = renamed.Generation
	ownerRaw, err = vNextPublicationJSON(owner)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ownerPath, ownerRaw, 0o600); err != nil {
		t.Fatal(err)
	}

	current, err := vNextPublicationJSON(renamed)

	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "acme", vNextPublicationCurrentFile), current, 0o600); err != nil {
		t.Fatal(err)
	}
	return renamed
}

func vNextDuplicateIntegrityMemberForTest(t *testing.T, root string, pointer vNextGenerationPointer, member string) {
	t.Helper()
	integrityPath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, pointer.Generation, vNextPublicationIntegrityFile)
	payload, err := os.ReadFile(integrityPath)
	if err != nil {
		t.Fatal(err)
	}
	index := strings.Index(string(payload), member)
	if index < 0 {
		t.Fatalf("integrity fixture lacks member %q", member)
	}
	suffix := append([]byte(nil), payload[index+len(member):]...)
	payload = append([]byte(nil), payload[:index]...)
	payload = append(payload, member...)
	payload = append(payload, ',')
	payload = append(payload, member...)
	payload = append(payload, suffix...)
	if err := os.WriteFile(integrityPath, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	current, err := vNextPublicationJSON(vNextGenerationPointer{
		Generation:      pointer.Generation,
		IntegrityDigest: vNextPublicationDigest(payload),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "acme", vNextPublicationCurrentFile), current, 0o600); err != nil {
		t.Fatal(err)
	}
}

func mustCanonicalVNextLockForTest(t *testing.T, lock vNextSourceLock) vNextCanonicalDescriptor {
	t.Helper()
	canonical, err := canonicalizeVNextSourceLock(lock)
	if err != nil {
		t.Fatalf("canonicalizeVNextSourceLock() error = %v", err)
	}
	return canonical
}

func vNextPublicationArtifactsForTest(marker string, includeRate bool) vNextPublicationArtifacts {
	files := map[string][]byte{
		"metadata.json":   []byte(`{"marker":"` + marker + `"}`),
		"spec.json":       []byte(`{"marker":"` + marker + `"}`),
		"manifest.json":   []byte(`{"marker":"` + marker + `"}`),
		"provenance.json": []byte(`[{"marker":"` + marker + `"}]`),
		"atlas.json":      []byte(`[{"marker":"` + marker + `"}]`),
		"index.json":      []byte(`{"marker":"` + marker + `"}`),
		"proof.json":      []byte(`{"marker":"` + marker + `"}`),
	}
	if includeRate {
		files["rate_limits.json"] = []byte(`{"marker":"` + marker + `"}`)
	}
	return vNextPublicationArtifacts{Files: files}
}

func assertVNextPublicationMarker(t *testing.T, handle *vNextGenerationHandle, want string) {
	t.Helper()
	for _, name := range []string{"metadata.json", "index.json", "manifest.json", "proof.json"} {
		payload, err := handle.ReadFile(name)
		if err != nil {
			t.Fatalf("ReadFile(%s) error = %v", name, err)
		}
		if string(payload) != `{"marker":"`+want+`"}` {
			t.Fatalf("ReadFile(%s) = %s, want marker %q", name, payload, want)
		}
	}
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}
func vNextCreateOwnedStageForTest(root string, pointer vNextGenerationPointer, stageName string, recover func(context.Context) error) error {
	stage := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, stageName)
	if err := os.Mkdir(stage, 0o755); err != nil {
		return err
	}
	marker, err := vNextPublicationJSON(vNextPublicationStageOwner{
		Version:    1,
		Connector:  "acme",
		Generation: pointer.Generation,
		Stage:      stageName,
	})
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(stage, vNextPublicationStageOwnerFile), marker, 0o600); err != nil {
		return err
	}
	return recover(context.Background())
}

func vNextReplaceConnectorRootForTest(root, external, stageName string) error {
	stage := filepath.Join(external, vNextPublicationGenerationDirectory, stageName)
	if err := os.MkdirAll(stage, 0o755); err != nil {
		return err
	}
	marker, err := vNextPublicationJSON(vNextPublicationStageOwner{
		Version:    1,
		Connector:  "acme",
		Generation: "g-" + strings.Repeat("a", 64),
		Stage:      stageName,
	})
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(stage, vNextPublicationStageOwnerFile), marker, 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(stage, "sentinel.txt"), []byte("keep"), 0o600); err != nil {
		return err
	}
	connectorRoot := filepath.Join(root, "acme")
	if err := os.Rename(connectorRoot, connectorRoot+"-moved"); err != nil {
		return err
	}
	return os.Symlink(external, connectorRoot)
}
