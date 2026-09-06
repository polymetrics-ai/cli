package main

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

func TestCP11F03AOriginalPostRecordFailureStrandsPreparedAuthority(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := baseline.Publish(vNextPublicationArtifactsForTest("old", true)); err != nil {
		t.Fatal(err)
	}
	injected := errors.New("injected post-record preparation failure")
	fired := false
	writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point == vNextPublicationAfterControlRepairRecord && !fired {
			fired = true
			return injected
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Publish(vNextPublicationArtifactsForTest("new", true)); !errors.Is(err, injected) {
		t.Fatalf("Publish post-record failure = %v, want injected cause", err)
	}
	if !fired {
		t.Fatal("post-record failure hook did not fire")
	}
	stranded := false
	transactions, err := filepath.Glob(filepath.Join(root, "acme", vNextPublicationControlRepairDirectoryPrefix+"*"))
	if err != nil {
		t.Fatal(err)
	}
	for _, transaction := range transactions {
		entries, err := os.ReadDir(transaction)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) == 1 && entries[0].Name() == vNextPublicationControlRepairPreparedFile {
			stranded = true
		}
	}
	if !stranded {
		t.Fatal("original post-record failure did not strand prepared.json after its anchors were removed")
	}
	fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.Recover(context.Background()); err == nil || !strings.Contains(err.Error(), "anchor") {
		t.Fatalf("fresh Recover after stranded prepared authority = %v, want missing-anchor refusal", err)
	}
}

func TestCP11F03COriginalTemporaryCleanupDeletesReplacementB(t *testing.T) {
	root := t.TempDir()
	var movedA, replacementB string
	fired := false
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{AfterTemporaryOpen: func(parent *vNextPublicationDirectory, name string, temporary *vNextPublicationDirectory) error {
		if fired {
			return nil
		}
		fired = true
		blocker, err := temporary.openFile(vNextPublicationTemporaryFile, "original temporary blocker", unix.O_CREAT|unix.O_EXCL|unix.O_WRONLY, 0o600, false)
		if err != nil {
			return err
		}
		if err := blocker.Close(); err != nil {
			return err
		}
		if err := unix.Renameat(int(parent.file.Fd()), name, int(parent.file.Fd()), name+".A"); err != nil {
			return err
		}
		if err := unix.Mkdirat(int(parent.file.Fd()), name, 0o700); err != nil {
			return err
		}
		movedA = filepath.Join(root, "acme", name+".A")
		replacementB = filepath.Join(root, "acme", name)
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := publisher.Publish(vNextPublicationArtifactsForTest("temporary-retry", true)); err != nil {
		t.Fatalf("Publish after original temporary allocation interference = %v", err)
	}
	if !fired {
		t.Fatal("temporary allocation hook did not fire")
	}
	if _, err := os.Stat(filepath.Join(movedA, vNextPublicationTemporaryFile)); err != nil {
		t.Fatalf("original temporary A/blocker was not retained: %v", err)
	}
	if _, err := os.Lstat(replacementB); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("original temporary cleanup retained replacement B: %v, want the demonstrated erroneous deletion", err)
	}
}

func TestCP11F03COriginalQuarantineCleanupDeletesReplacementB(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	old, err := baseline.Publish(vNextPublicationArtifactsForTest("old", true))
	if err != nil {
		t.Fatal(err)
	}
	held, err := baseline.Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := baseline.Publish(vNextPublicationArtifactsForTest("new", true)); err != nil {
		held.Release()
		t.Fatal(err)
	}
	held.Release()
	injected := errors.New("injected opened quarantine completion failure")
	var movedA, replacementB string
	fired := false
	writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{AfterQuarantineOpen: func(parent *vNextPublicationDirectory, name string, quarantine *vNextPublicationDirectory, identity vNextPublicationIdentity) error {
		if fired {
			return nil
		}
		fired = true
		if actual, err := vNextPublicationIdentityFromFile(quarantine.file, "original quarantine A"); err != nil || actual != identity {
			return errors.New("opened quarantine identity was not retained before replacement")
		}
		if err := unix.Renameat(int(parent.file.Fd()), name, int(parent.file.Fd()), name+".A"); err != nil {
			return err
		}
		if err := unix.Mkdirat(int(parent.file.Fd()), name, 0o700); err != nil {
			return err
		}
		movedA = filepath.Join(root, "acme", vNextPublicationGenerationDirectory, name+".A")
		replacementB = filepath.Join(root, "acme", vNextPublicationGenerationDirectory, name)
		return injected
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := writer.Prune(context.Background()); !errors.Is(err, injected) {
		t.Fatalf("Prune after original quarantine allocation interference = %v, want injected cause", err)
	}
	if !fired {
		t.Fatal("quarantine allocation hook did not fire")
	}
	if _, err := os.Stat(movedA); err != nil {
		t.Fatalf("opened quarantine A was not retained: %v", err)
	}
	if _, err := os.Lstat(replacementB); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("original quarantine cleanup retained replacement B: %v, want the demonstrated erroneous deletion", err)
	}
	if !baseline.GenerationExists(old.Generation) {
		t.Fatal("original quarantine allocation interference removed the stale generation")
	}
}

func TestCP11F03BOriginalCompoundCausesAreNotInspectable(t *testing.T) {
	t.Run("definitions and connector close", func(t *testing.T) {
		root := t.TempDir()
		definitionsClose := errors.New("injected definitions close failure")
		connectorClose := errors.New("injected connector close failure")
		publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{CloseDirectory: func(file *os.File, label string) error {
			if err := file.Close(); err != nil {
				return err
			}
			switch label {
			case "publication definitions root":
				return definitionsClose
			case `connector publication root "acme"`:
				return connectorClose
			default:
				return nil
			}
		}})
		if err != nil {
			t.Fatal(err)
		}
		connector, err := publisher.openConnectorRoot(true)
		if connector != nil {
			t.Fatalf("openConnectorRoot returned connector after both real closes: %v", connector)
		}
		if !errors.Is(err, connectorClose) || errors.Is(err, definitionsClose) {
			t.Fatalf("original definitions/connector compound result = %v, want only final close cause inspectable", err)
		}
	})

	t.Run("missing control parent close", func(t *testing.T) {
		root := t.TempDir()
		completion := errors.New("injected missing-control parent close failure")
		publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{CloseDirectory: func(file *os.File, label string) error {
			if err := file.Close(); err != nil {
				return err
			}
			if label == "original missing control" {
				return completion
			}
			return nil
		}})
		if err != nil {
			t.Fatal(err)
		}
		operation, err := publisher.openOperation(context.Background(), unix.LOCK_EX, true)
		if err != nil {
			t.Fatal(err)
		}
		defer operation.close()
		_, found, _, err := vNextPublicationReadControlBound(operation.connector, "missing-control", "original missing control")
		if !found || !errors.Is(err, completion) || errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("original missing-control compound result = found=%t err=%v, want flattened absence and inspectable completion only", found, err)
		}
	})

	t.Run("staged file sync and close", func(t *testing.T) {
		root := t.TempDir()
		primary := errors.New("injected staged-file sync frontier")
		completion := errors.New("injected staged-file close completion")
		fired := false
		publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{
			At: func(point vNextPublicationFaultPoint) error {
				if point == vNextPublicationBeforeFileSync && !fired {
					fired = true
					return primary
				}
				return nil
			},
			CloseStageFile: func(file *os.File) error {
				if err := file.Close(); err != nil {
					return err
				}
				return completion
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = publisher.Publish(vNextPublicationArtifactsForTest("compound-stage", true))
		if !fired || !errors.Is(err, primary) || errors.Is(err, completion) {
			t.Fatalf("original staged-file compound result = fired=%t err=%v, want only primary inspectable", fired, err)
		}
	})
}
