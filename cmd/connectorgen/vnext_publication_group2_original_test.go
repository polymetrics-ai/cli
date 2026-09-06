package main

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

func TestCP11F03ARepairPostRecordFailureRetainsCoherentAuthority(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	oldArtifacts := vNextPublicationArtifactsForTest("old", true)
	if _, err := baseline.Publish(oldArtifacts); err != nil {
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
	if stranded {
		t.Fatal("post-record failure stranded prepared.json after its anchors were removed")
	}
	fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.Recover(context.Background()); err != nil {
		t.Fatalf("fresh Recover after post-record failure = %v, want coherent recovery", err)
	}
	if err := fresh.Check(oldArtifacts); err != nil {
		t.Fatalf("fresh Check after post-record recovery = %v", err)
	}
	if _, err := fresh.Publish(vNextPublicationArtifactsForTest("retry", true)); err != nil {
		t.Fatalf("ordinary retry after post-record recovery = %v", err)
	}
}

func TestCP11F03CRepairTemporaryCleanupPreservesReplacementB(t *testing.T) {
	for _, test := range []struct {
		name    string
		foreign []byte
	}{
		{name: "empty"},
		{name: "nonempty", foreign: []byte("foreign replacement B")},
	} {
		t.Run(test.name, func(t *testing.T) {
			vNextPublicationAssertTemporaryReplacementBForTest(t, test.foreign)
		})
	}
}

func vNextPublicationAssertTemporaryReplacementBForTest(t *testing.T, foreign []byte) {
	t.Helper()
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
		if err := vNextPublicationCreateReplacementBForTest(parent, name, foreign); err != nil {
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
	if _, err := os.Lstat(replacementB); err != nil {
		t.Fatalf("temporary cleanup removed replacement B: %v", err)
	}
	if len(foreign) != 0 {
		payload, err := os.ReadFile(filepath.Join(replacementB, "foreign"))
		if err != nil || string(payload) != string(foreign) {
			t.Fatalf("temporary cleanup changed nonempty replacement B: payload=%q err=%v", payload, err)
		}
	}
}

func TestCP11F03CRepairQuarantineCleanupPreservesReplacementB(t *testing.T) {
	for _, test := range []struct {
		name    string
		foreign []byte
	}{
		{name: "empty"},
		{name: "nonempty", foreign: []byte("foreign replacement B")},
	} {
		t.Run(test.name, func(t *testing.T) {
			vNextPublicationAssertQuarantineReplacementBForTest(t, test.foreign)
		})
	}
}

func vNextPublicationAssertQuarantineReplacementBForTest(t *testing.T, foreign []byte) {
	t.Helper()
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
		if err := vNextPublicationCreateReplacementBForTest(parent, name, foreign); err != nil {
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
	if _, err := os.Lstat(replacementB); err != nil {
		t.Fatalf("quarantine cleanup removed replacement B: %v", err)
	}
	if len(foreign) != 0 {
		payload, err := os.ReadFile(filepath.Join(replacementB, "foreign"))
		if err != nil || string(payload) != string(foreign) {
			t.Fatalf("quarantine cleanup changed nonempty replacement B: payload=%q err=%v", payload, err)
		}
	}
	if !baseline.GenerationExists(old.Generation) {
		t.Fatal("original quarantine allocation interference removed the stale generation")
	}
}

func vNextPublicationCreateReplacementBForTest(parent *vNextPublicationDirectory, name string, foreign []byte) error {
	if err := unix.Mkdirat(int(parent.file.Fd()), name, 0o700); err != nil {
		return err
	}
	if len(foreign) == 0 {
		return nil
	}
	file, err := parent.openFile(name+"/foreign", "replacement B foreign member", unix.O_CREAT|unix.O_EXCL|unix.O_WRONLY, 0o600, false)
	if err != nil {
		return err
	}
	if _, err := file.Write(foreign); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func TestCP11F03BRepairCompoundCausesRemainInspectable(t *testing.T) {
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
		if !errors.Is(err, definitionsClose) || !errors.Is(err, connectorClose) {
			t.Fatalf("definitions/connector compound result = %v, want both close causes inspectable", err)
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
		if !found || !errors.Is(err, completion) || !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("missing-control compound result = found=%t err=%v, want both absence and completion causes", found, err)
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
		if !fired || !errors.Is(err, primary) || !errors.Is(err, completion) {
			t.Fatalf("staged-file compound result = fired=%t err=%v, want both primary and completion causes", fired, err)
		}
	})

	t.Run("failed stage and owned cleanup", func(t *testing.T) {
		root := t.TempDir()
		primary := errors.New("injected staged-file primary failure")
		cleanup := errors.New("injected owned stage cleanup failure")
		publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
			switch point {
			case vNextPublicationBeforeFileSync:
				return primary
			case vNextPublicationBeforeStageRemoval:
				return cleanup
			default:
				return nil
			}
		}})
		if err != nil {
			t.Fatal(err)
		}
		_, err = publisher.Publish(vNextPublicationArtifactsForTest("compound-stage-cleanup", true))
		if !errors.Is(err, primary) || !errors.Is(err, cleanup) {
			t.Fatalf("failed stage/cleanup result = %v, want primary and owned cleanup causes", err)
		}
	})

	t.Run("capture sync and close", func(t *testing.T) {
		root := t.TempDir()
		primary := errors.New("injected capture pre-close failure")
		completion := errors.New("injected capture close completion")
		publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{
			At: func(point vNextPublicationFaultPoint) error {
				if point == vNextPublicationBeforeControlRepairCaptureClose {
					return primary
				}
				return nil
			},
			CloseDirectory: func(file *os.File, label string) error {
				if err := file.Close(); err != nil {
					return err
				}
				if label == "publication control capture" {
					return completion
				}
				return nil
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = publisher.Publish(vNextPublicationArtifactsForTest("compound-capture", true))
		if !errors.Is(err, primary) || !errors.Is(err, completion) {
			t.Fatalf("capture sync/close result = %v, want primary and completion causes", err)
		}
	})
}
