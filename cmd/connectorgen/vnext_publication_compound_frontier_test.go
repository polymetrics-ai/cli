package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

func TestCP11CompoundOpenedControlFrontiers(t *testing.T) {
	for _, frontier := range []string{"identity", "stat", "bounded-read"} {
		t.Run(frontier, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "CURRENT")
			payload := []byte("known actual control bytes")
			if err := os.WriteFile(path, payload, 0o600); err != nil {
				t.Fatal(err)
			}
			want, err := vNextPublicationObserveExpectedTree(path)
			if err != nil {
				t.Fatal(err)
			}
			directory, err := vNextPublicationOpenDirectory(root, "control root")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := directory.Close(); err != nil {
					t.Errorf("close test-owned directory: %v", err)
				}
			}()
			primary := errors.New("opened-control " + frontier)
			completion := errors.New("opened-control close")
			operated, closed := 0, 0
			var actual *os.File
			t.Cleanup(func() {
				vNextPublicationReadControlIdentityForTest = nil
				vNextPublicationReadControlStatForTest = nil
				vNextPublicationReadControlReaderForTest = nil
				vNextPublicationCloseReadControlForTest = nil
			})
			vNextPublicationCloseReadControlForTest = func(file *os.File, _ string) error {
				closed++
				if file != actual {
					t.Error("Close did not own the actual observed descriptor")
				}
				return errors.Join(file.Close(), completion)
			}
			switch frontier {
			case "identity":
				vNextPublicationReadControlIdentityForTest = func(file *os.File, label string) (vNextPublicationIdentity, error) {
					actual = file
					operated++
					identity, err := vNextPublicationIdentityFromFile(file, label)
					if err == nil && identity != want["."].identity {
						t.Error("actual identity mismatch")
					}
					return identity, errors.Join(err, primary)
				}
			case "stat":
				vNextPublicationReadControlStatForTest = func(file *os.File, _ string) (os.FileInfo, error) {
					actual = file
					operated++
					info, err := file.Stat()
					if err == nil && info.Size() != int64(len(payload)) {
						t.Error("actual Stat did not observe expected bytes")
					}
					return info, errors.Join(err, primary)
				}
			case "bounded-read":
				vNextPublicationReadControlReaderForTest = func(file *os.File, _ string) io.Reader {
					actual = file
					operated++
					return io.MultiReader(io.LimitReader(file, 1), vNextPublicationControlErrorReader{err: primary})
				}
			}
			got, found, identity, err := vNextPublicationReadControlBound(directory, "CURRENT", "control")
			if !found || len(got) != 0 || identity != (vNextPublicationIdentity{}) || operated != 1 || closed != 1 || !errors.Is(err, primary) || !errors.Is(err, completion) {
				t.Fatalf("compound boundary: found=%t bytes=%q identity=%#v operated=%d closed=%d err=%v", found, got, identity, operated, closed, err)
			}
			vNextPublicationAssertExpectedTree(t, path, want)
		})
	}
}

func TestCP11CompoundRecordUnknownIdentityRetainsOwnedEffects(t *testing.T) {
	for _, name := range []string{vNextPublicationControlRepairPreparedFile, vNextPublicationControlAuthorityMarkerFile, vNextPublicationControlRepairPhaseName(1)} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			directory, err := vNextPublicationOpenDirectory(root, "record root")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := directory.Close(); err != nil {
					t.Errorf("close test-owned directory: %v", err)
				}
			}()
			primary := errors.New("actual identity completion")
			completion := errors.New("actual close completion")
			closed := 0
			var want vNextPublicationExpectedTree
			result, err := vNextPublicationWriteControlRepairRecord(directory, name, "identity record", []byte("{}\n"), vNextPublicationControlRecordHooks{
				Identity: func(file *os.File, label string) (vNextPublicationIdentity, error) {
					_, err := vNextPublicationIdentityFromFile(file, label)
					if err != nil {
						return vNextPublicationIdentity{}, err
					}
					want, err = vNextPublicationObserveExpectedTree(filepath.Join(root, name))
					return vNextPublicationIdentity{}, errors.Join(err, primary)
				},
				Write: func(*os.File, string, []byte) (int, error) {
					t.Error("wrote after unknown identity")
					return 0, errors.New("unexpected write")
				},
				Close: func(file *os.File, _ string) error { closed++; return errors.Join(file.Close(), completion) },
			})
			if !result.created || result.identity != (vNextPublicationIdentity{}) || result.disposition != vNextPublicationRecordUnknown || closed != 1 || !errors.Is(err, primary) || !errors.Is(err, completion) {
				t.Fatalf("unknown result %#v closed=%d err=%v", result, closed, err)
			}
			vNextPublicationAssertExpectedTree(t, filepath.Join(root, name), want)
		})
	}
}

func TestCP11CompoundFailedPredecessorLink(t *testing.T) {
	root := t.TempDir()
	artifacts := vNextPublicationArtifactsForTest("link-active", false)
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	current, err := baseline.Publish(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := vNextPublicationExpectedAuthority(root)
	if err != nil {
		t.Fatal(err)
	}
	completion := errors.New("predecessor real Close completion")
	closed, blocked := 0, 0
	fixtureCompleted := false
	var blockerIdentity vNextPublicationIdentity
	writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{
		BeforeRepairPredecessorLink: func(transaction *vNextPublicationDirectory) error {
			blocked++
			file, err := transaction.openFile(vNextPublicationControlBackupMember, "fixture link blocker", unix.O_CREAT|unix.O_EXCL|unix.O_WRONLY, 0o600, false)
			if err != nil {
				return err
			}
			blockerIdentity, err = vNextPublicationIdentityFromFile(file, "fixture blocker")
			_, writeErr := file.Write([]byte("foreign link blocker"))
			fixtureErr := errors.Join(err, writeErr, file.Close())
			fixtureCompleted = fixtureErr == nil
			return fixtureErr
		},
		CloseRepairPredecessor: func(directory *vNextPublicationDirectory) error {
			closed++
			return errors.Join(directory.Close(), completion)
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	operation, err := writer.openOperation(context.Background(), unix.LOCK_EX, true)
	if err != nil {
		t.Fatal(err)
	}
	err = writer.writeCurrentLocked(operation, current)
	operation.close()
	if !fixtureCompleted || blocked != 1 || closed != 1 || !errors.Is(err, os.ErrExist) || !errors.Is(err, completion) {
		t.Fatalf("actual failed link plus Close: blocked=%d closed=%d err=%v", blocked, closed, err)
	}
	actual, err := vNextPublicationExpectedAuthority(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := vNextPublicationCompareExpectedMembers(actual, prior, false); err != nil {
		t.Fatal(err)
	}
	var blockerPath string
	for path, member := range actual {
		if member.identity == blockerIdentity {
			if !bytes.Equal(member.payload, []byte("foreign link blocker")) {
				t.Fatal("blocker bytes changed")
			}
			blockerPath = path
		}
	}
	if blockerPath == "" {
		t.Fatal("failed link adopted/deleted foreign blocker")
	}
	txPath := filepath.Dir(filepath.Join(root, "acme", blockerPath))
	if _, err := os.Lstat(filepath.Join(txPath, vNextPublicationControlRepairPreparedFile)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("failed link created prepared authority: %v", err)
	}
	// The fixture-created blocker is removed only after preservation assertions.
	if err := os.Remove(filepath.Join(root, "acme", blockerPath)); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(txPath); err != nil {
		t.Fatal(err)
	}
	fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := fresh.Check(artifacts); err != nil {
		t.Fatal(err)
	}
}

func TestCP11CompoundCapturePrimarySyncClose(t *testing.T) {
	root := t.TempDir()
	primary := errors.New("capture preparation primary")
	syncCompletion := errors.New("capture actual Sync completion")
	closeCompletion := errors.New("capture actual Close completion")
	armed := false
	syncSeen, closed := 0, 0
	var expected vNextPublicationExpectedTree
	closes := make(map[*os.File]int)
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{
		At: func(point vNextPublicationFaultPoint) error {
			switch point {
			case vNextPublicationBeforeControlRepairCaptureClose:
				armed = true
				var err error
				expected, err = vNextPublicationExpectedAuthority(root)
				return errors.Join(err, primary)
			case vNextPublicationAfterControlRepairCaptureSync:
				if armed {
					syncSeen++
					return syncCompletion
				}
			}
			return nil
		},
		CloseDirectory: func(file *os.File, label string) error {
			closes[file]++
			err := file.Close()
			if armed && label == "publication control capture" {
				closed++
				return errors.Join(err, closeCompletion)
			}
			return err
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	artifacts := vNextPublicationArtifactsForTest("capture-compound", false)
	_, err = publisher.Publish(artifacts)
	if syncSeen != 1 || closed != 1 || !errors.Is(err, primary) || !errors.Is(err, syncCompletion) || !errors.Is(err, closeCompletion) {
		t.Fatalf("capture three causes sync=%d close=%d err=%v", syncSeen, closed, err)
	}
	for file, count := range closes {
		if count != 1 {
			t.Fatalf("capture/parent/readDir instance %p Close=%d", file, count)
		}
	}
	after, observeErr := vNextPublicationExpectedAuthority(root)
	if observeErr != nil {
		t.Fatal(observeErr)
	}
	if expected == nil {
		t.Fatal("missing capture pre-fault authority")
	}
	if err := vNextPublicationCompareExpectedMembers(after, expected, true); err != nil {
		t.Fatal(err)
	}
	fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := fresh.Publish(artifacts); err != nil {
		t.Fatal(err)
	}
	if err := fresh.Check(artifacts); err != nil {
		t.Fatal(err)
	}
}

func TestCP11CompoundStagedActualSyncCloseCleanup(t *testing.T) {
	root := t.TempDir()
	primary := errors.New("stage actual Sync completion")
	completion := errors.New("stage real Close completion")
	cleanup := errors.New("owned stage cleanup refusal")
	synced, closed, refused := 0, 0, 0
	var expected vNextPublicationExpectedTree
	var staged *os.File
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{
		At: func(point vNextPublicationFaultPoint) error {
			switch point {
			case vNextPublicationAfterFileSync:
				synced++
				var err error
				expected, err = vNextPublicationObserveExpectedTree(filepath.Join(root, "acme"))
				return errors.Join(err, primary)
			case vNextPublicationBeforeStageRemoval:
				refused++
				return cleanup
			}
			return nil
		},
		CloseStageFile: func(file *os.File) error { closed++; staged = file; return errors.Join(file.Close(), completion) },
	})
	if err != nil {
		t.Fatal(err)
	}
	artifacts := vNextPublicationArtifactsForTest("stage-compound", false)
	_, err = publisher.Publish(artifacts)
	if synced != 1 || closed != 1 || refused != 1 || !errors.Is(err, primary) || !errors.Is(err, completion) || !errors.Is(err, cleanup) {
		t.Fatalf("staged actual completion: sync=%d close=%d cleanup=%d err=%v", synced, closed, refused, err)
	}
	if expected == nil || staged == nil {
		t.Fatal("missing staged effect witness")
	}
	if _, err := staged.Stat(); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("staged descriptor remains usable: %v", err)
	}
	vNextPublicationAssertExpectedTree(t, filepath.Join(root, "acme"), expected)
	fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := fresh.Publish(artifacts); err != nil {
		t.Fatal(err)
	}
	if err := fresh.Check(artifacts); err != nil {
		t.Fatal(err)
	}
}

// The actual io.ReadAll/LimitReader path consumes a real byte before this
// reader supplies its injected error; it is not a post-ReadAll completion hook.
type vNextPublicationControlErrorReader struct{ err error }

func (reader vNextPublicationControlErrorReader) Read([]byte) (int, error) { return 0, reader.err }
