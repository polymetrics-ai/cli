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

func TestCP11OwnershipPreparedOpenCompletionMustRemainRecoverable(t *testing.T) {
	root := t.TempDir()
	baseline, _ := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	old := vNextPublicationArtifactsForTest("old", true)
	if _, err := baseline.Publish(old); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(root, "acme", "CURRENT"))
	if err != nil {
		t.Fatal(err)
	}
	completion := errors.New("review: real prepared-open parent Close completed with error")
	armed, fired := false, false
	writer, _ := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{
		At: func(point vNextPublicationFaultPoint) error {
			if point == vNextPublicationBeforeControlRepairRecord {
				armed = true
			}
			return nil
		},
		CloseDirectory: func(file *os.File, label string) error {
			err := file.Close()
			if armed && !fired && label == "publication control repair prepared authority" {
				fired = true
				return errors.Join(err, completion)
			}
			return err
		},
	})
	_, publishErr := writer.Publish(vNextPublicationArtifactsForTest("new", true))
	if !fired || !errors.Is(publishErr, completion) {
		t.Fatalf("fault placement failed: fired=%t err=%v", fired, publishErr)
	}
	after, err := os.ReadFile(filepath.Join(root, "acme", "CURRENT"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("CURRENT changed: %v", err)
	}
	txs, err := filepath.Glob(filepath.Join(root, "acme", vNextPublicationControlRepairDirectoryPrefix+"*"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tx := range txs {
		entries, err := os.ReadDir(tx)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) == 1 && entries[0].Name() == vNextPublicationControlRepairPreparedFile {
			payload, err := os.ReadFile(filepath.Join(tx, entries[0].Name()))
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("stranded transaction=%s members=[prepared.json] bytes=%d", filepath.Base(tx), len(payload))
		}
	}
	fresh, _ := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	checkErr := fresh.Check(old)
	recoveryErr := fresh.Recover(context.Background())
	t.Logf("Publish=%v; Check=%v; Recover=%v; CURRENT unchanged=true", publishErr, checkErr, recoveryErr)
	if recoveryErr != nil {
		t.Fatalf("F03-A first real creation frontier must recover after owned prepared-open parent Close failure: %v", recoveryErr)
	}
}

func TestCP11OwnershipTempRetryMustNotSwallowCompletion(t *testing.T) {
	root := t.TempDir()
	completion := errors.New("review: failed allocator owned-directory real Close completion")
	fired, closeFired := false, false
	var aPath, bPath string
	var aBefore, bBefore os.FileInfo
	writer, _ := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{
		AfterTemporaryOpen: func(_ *vNextPublicationDirectory, name string, opened *vNextPublicationDirectory) error {
			if fired {
				return nil
			}
			fired = true
			var err error
			aBefore, err = opened.file.Stat()
			if err != nil {
				return err
			}
			blocker, err := opened.openFile("control", "review real EEXIST blocker", unix.O_CREAT|unix.O_EXCL|unix.O_WRONLY, 0o600, false)
			if err != nil {
				return err
			}
			if _, err := blocker.Write([]byte("A blocker")); err != nil {
				return errors.Join(err, blocker.Close())
			}
			if err := blocker.Close(); err != nil {
				return err
			}
			bPath = filepath.Join(root, "acme", name)
			aPath = bPath + "-A"
			if err := os.Rename(bPath, aPath); err != nil {
				return err
			}
			if err := os.Mkdir(bPath, 0o700); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(bPath, "foreign"), []byte("B bytes"), 0o600); err != nil {
				return err
			}
			bBefore, err = os.Lstat(bPath)
			return err
		},
		CloseDirectory: func(file *os.File, label string) error {
			err := file.Close()
			if fired && !closeFired && label == "publication temporary root" {
				closeFired = true
				return errors.Join(err, completion)
			}
			return err
		},
	})
	pointer, err := writer.Publish(vNextPublicationArtifactsForTest("new", true))
	if !fired || !closeFired {
		t.Fatalf("fault did not hit failed allocator: allocation=%t close=%t err=%v", fired, closeFired, err)
	}
	aAfter, aErr := os.Lstat(aPath)
	bAfter, bErr := os.Lstat(bPath)
	if aErr != nil || bErr != nil || !os.SameFile(aBefore, aAfter) || !os.SameFile(bBefore, bAfter) {
		t.Fatalf("A/B preservation failed: %v %v", aErr, bErr)
	}
	blocker, _ := os.ReadFile(filepath.Join(aPath, "control"))
	foreign, _ := os.ReadFile(filepath.Join(bPath, "foreign"))
	if string(blocker) != "A blocker" || string(foreign) != "B bytes" {
		t.Fatalf("A/B bytes changed")
	}
	t.Logf("Publish pointer=%#v error=%v A/B identities and bytes preserved=true", pointer, err)
	if !errors.Is(err, completion) {
		t.Fatalf("EEXIST retry swallowed meaningful owned Close completion: %v", err)
	}
}

func TestCP11OwnershipRecordFrontiersRecover(t *testing.T) {
	for _, kind := range []struct{ name, label, file string }{
		{"prepared", "publication control repair prepared authority", vNextPublicationControlRepairPreparedFile},
		{"marker", "publication control authority marker", vNextPublicationControlAuthorityMarkerFile},
		{"phase", "publication control repair phase", vNextPublicationControlRepairPhaseName(1)},
	} {
		for _, frontier := range []string{"parent-close-after-create", "partial-write-close", "short-write-close", "sync-close", "full-write-error-close", "close"} {
			t.Run(kind.name+"/"+frontier, func(t *testing.T) {
				root := t.TempDir()
				failure := errors.New("record frontier " + frontier)
				completion := errors.New("record close completion")
				fired := false
				closed := 0
				written := 0
				var created vNextPublicationIdentity
				var payload []byte
				hooks := vNextPublicationHooks{}
				hooks.CloseDirectory = func(file *os.File, label string) error {
					var stat unix.Stat_t
					hit := frontier == "parent-close-after-create" && !fired && label == kind.label && unix.Fstatat(int(file.Fd()), kind.file, &stat, unix.AT_SYMLINK_NOFOLLOW) == nil && stat.Size == 0 && uint32(stat.Mode)&unix.S_IFMT == unix.S_IFREG
					if hit {
						fired = true
						created = vNextPublicationIdentityFromStat(stat)
					}
					err := file.Close()
					if hit {
						return errors.Join(err, failure)
					}
					return err
				}
				hooks.ControlRecord.Write = func(file *os.File, label string, b []byte) (int, error) {
					if fired || label != kind.label || frontier == "parent-close-after-create" {
						return file.Write(b)
					}
					fired = true
					var err error
					created, err = vNextPublicationIdentityFromFile(file, label)
					if err != nil {
						return 0, err
					}
					payload = bytes.Clone(b)
					if frontier == "partial-write-close" || frontier == "short-write-close" {
						written, err = file.Write(b[:1])
						if frontier == "partial-write-close" {
							err = errors.Join(err, failure)
						}
						return written, err
					}
					written, err = file.Write(b)
					if frontier == "full-write-error-close" {
						err = errors.Join(err, failure)
					}
					return written, err
				}
				hooks.ControlRecord.Sync = func(file *os.File, label string) error {
					err := file.Sync()
					identity, idErr := vNextPublicationIdentityFromFile(file, label)
					if fired && identity == created && frontier == "sync-close" {
						return errors.Join(err, idErr, failure)
					}
					return errors.Join(err, idErr)
				}
				hooks.ControlRecord.Close = func(file *os.File, label string) error {
					identity, idErr := vNextPublicationIdentityFromFile(file, label)
					hit := fired && identity == created
					err := file.Close()
					if hit {
						closed++
						return errors.Join(idErr, err, completion)
					}
					return errors.Join(idErr, err)
				}
				writer, err := newVNextGenerationPublisher(root, "acme", hooks)
				if err != nil {
					t.Fatal(err)
				}
				artifacts := vNextPublicationArtifactsForTest("record-frontier", true)
				_, publishErr := writer.Publish(artifacts)
				if !fired || closed != 1 || !errors.Is(publishErr, completion) {
					t.Fatalf("first effect/one Close/public error: fired=%t closed=%d err=%v", fired, closed, publishErr)
				}
				if frontier != "short-write-close" && frontier != "close" && !errors.Is(publishErr, failure) {
					t.Fatalf("lost primary cause: %v", publishErr)
				}
				if frontier == "short-write-close" && !errors.Is(publishErr, io.ErrShortWrite) {
					t.Fatalf("lost short write: %v", publishErr)
				}
				complete := frontier == "sync-close" || frontier == "full-write-error-close" || frontier == "close"
				connector := filepath.Join(root, "acme")
				tree, err := vNextPublicationObserveExpectedTree(connector)
				if err != nil {
					t.Fatal(err)
				}
				retained := 0
				for _, member := range tree {
					if member.identity == created {
						retained++
						if !bytes.Equal(member.payload, payload) {
							t.Fatal("complete record bytes changed")
						}
					}
				}
				if complete && (retained != 1 || written != len(payload)) {
					t.Fatalf("complete authority lost: retained=%d written=%d want=%d", retained, written, len(payload))
				}
				if !complete && retained != 0 {
					t.Fatalf("owned incomplete record retained at %d paths", retained)
				}
				fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
				if err != nil {
					t.Fatal(err)
				}
				if err := fresh.Check(artifacts); err == nil {
					t.Fatal("interrupted initial publication unexpectedly passed Check")
				}
				vNextPublicationAssertExpectedTree(t, connector, tree)
				if err := fresh.Recover(context.Background()); err != nil {
					t.Fatalf("fresh recovery: %v", err)
				}
				if _, err := fresh.Publish(artifacts); err != nil {
					t.Fatalf("ordinary retry: %v", err)
				}
				if err := fresh.Check(artifacts); err != nil {
					t.Fatalf("completed retry Check: %v", err)
				}
			})
		}
	}
}

func TestCP11OwnershipAllocatorClassifiesActualCleanup(t *testing.T) {
	for _, mode := range []string{"pure-collision", "exhaustion", "parent-close", "root-close", "completion-matches-exist"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			directory, err := vNextPublicationOpenDirectory(root, "allocator test root")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := directory.Close(); err != nil {
					t.Errorf("close test-owned directory: %v", err)
				}
			}()
			completion := errors.New("actual allocation completion")
			attempts := 0
			closes := 0
			expected := map[string]vNextPublicationExpectedTree{}
			directory.closeForTest = func(file *os.File, label string) error {
				err := file.Close()
				hit := attempts == 1 && ((mode == "parent-close" && label == "publication temporary") || ((mode == "root-close" || mode == "completion-matches-exist") && label == "publication temporary root"))
				if hit {
					closes++
					if mode == "completion-matches-exist" {
						return errors.Join(err, completion, os.ErrExist)
					}
					return errors.Join(err, completion)
				}
				return err
			}
			name, temporary, file, resultErr := vNextPublicationCreateTemp(directory, func(_ *vNextPublicationDirectory, name string, opened *vNextPublicationDirectory) error {
				attempts++
				if attempts > 1 && mode != "exhaustion" {
					return nil
				}
				fd, err := unix.Openat(int(opened.file.Fd()), vNextPublicationTemporaryFile, unix.O_CREAT|unix.O_EXCL|unix.O_WRONLY, 0o600)
				if err != nil {
					return err
				}
				blocker := os.NewFile(uintptr(fd), "actual collision blocker")
				if blocker == nil {
					return errors.Join(errors.New("invalid blocker fd"), unix.Close(fd))
				}
				_, writeErr := blocker.Write([]byte("retained blocker"))
				if err := errors.Join(writeErr, blocker.Close()); err != nil {
					return err
				}
				path := filepath.Join(root, name)
				want, err := vNextPublicationObserveExpectedTree(path)
				if err != nil {
					return err
				}
				expected[path] = want
				return nil
			})
			if file != nil {
				if err := file.Close(); err != nil {
					t.Fatal(err)
				}
			}
			if temporary != nil {
				if err := temporary.Close(); err != nil {
					t.Fatal(err)
				}
			}
			for path, want := range expected {
				vNextPublicationAssertExpectedTree(t, path, want)
			}
			switch mode {
			case "pure-collision":
				if resultErr != nil || attempts != 2 || name == "" {
					t.Fatalf("pure collision retry: attempts=%d result=%v", attempts, resultErr)
				}
			case "exhaustion":
				if attempts != 128 || !errors.Is(resultErr, os.ErrExist) || name != "" {
					t.Fatalf("bounded collision exhaustion: attempts=%d result=%v", attempts, resultErr)
				}
			default:
				if attempts != 1 || closes != 1 || !errors.Is(resultErr, completion) || !errors.Is(resultErr, os.ErrExist) || name != "" {
					t.Fatalf("compound collision must not retry: attempts=%d closes=%d result=%v", attempts, closes, resultErr)
				}
			}
		})
	}
}
