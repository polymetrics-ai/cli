package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

func TestCP11AllocationCallersRetainActualAB(t *testing.T) {
	for _, caller := range []string{"public-publish", "CURRENT", "JOURNAL", "generation-prune", "stage-recover"} {
		for _, nonempty := range []bool{false, true} {
			variant := "empty"
			if nonempty {
				variant = "nonempty"
			}
			t.Run(caller+"/"+variant, func(t *testing.T) {
				root := t.TempDir()
				active := vNextPublicationArtifactsForTest("allocation-active", true)
				next := vNextPublicationArtifactsForTest("allocation-next", true)
				baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
				if err != nil {
					t.Fatal(err)
				}
				current, err := baseline.Publish(active)
				if err != nil {
					t.Fatal(err)
				}
				if caller == "generation-prune" {
					held, err := baseline.Open(context.Background())
					if err != nil {
						t.Fatal(err)
					}
					current, err = baseline.Publish(next)
					held.Release()
					if err != nil {
						t.Fatal(err)
					}
					active = next
				}
				if caller == "stage-recover" {
					operation, err := baseline.openOperation(context.Background(), unix.LOCK_EX, true)
					if err != nil {
						t.Fatal(err)
					}
					_, _, err = baseline.stageLocked(operation, next.Files, nil)
					operation.close()
					if err != nil {
						t.Fatal(err)
					}
				}
				connector := filepath.Join(root, "acme")
				before, err := vNextPublicationObserveExpectedTree(connector)
				if err != nil {
					t.Fatal(err)
				}
				injected := errors.New("actual quarantine allocation completion")
				attempts := 0
				var allocated *os.File
				closes := make(map[*os.File]int)
				var aPath, bPath string
				var expected vNextPublicationExpectedTree
				interfere := func(parent *vNextPublicationDirectory, name string, opened *vNextPublicationDirectory, temporary bool) error {
					attempts++
					allocated = opened.file
					if attempts != 1 {
						return errors.New("compound allocation silently retried")
					}
					aIdentity, err := vNextPublicationIdentityFromFile(opened.file, "opened allocation A")
					if err != nil {
						return err
					}
					if temporary {
						fd, err := unix.Openat(int(opened.file.Fd()), vNextPublicationTemporaryFile, unix.O_CREAT|unix.O_EXCL|unix.O_WRONLY, 0o600)
						if err != nil {
							return err
						}
						file := os.NewFile(uintptr(fd), "actual EEXIST blocker")
						if file == nil {
							return errors.Join(unix.Close(fd), errors.New("invalid blocker fd"))
						}
						_, writeErr := file.Write([]byte("owned A collision blocker"))
						if err := errors.Join(writeErr, file.Close()); err != nil {
							return err
						}
					}
					if err := unix.Renameat(int(parent.file.Fd()), name, int(parent.file.Fd()), name+".retained-A"); err != nil {
						return err
					}
					var foreign []byte
					if nonempty {
						foreign = []byte("actual actor-created B")
					}
					if err := vNextPublicationCreateReplacementBForTest(parent, name, foreign); err != nil {
						return err
					}
					parentPath := connector
					if !temporary {
						parentPath = filepath.Join(connector, vNextPublicationGenerationDirectory)
					}
					aPath = filepath.Join(parentPath, name+".retained-A")
					bPath = filepath.Join(parentPath, name)
					expected, err = vNextPublicationObserveExpectedTree(connector)
					if err != nil {
						return err
					}
					aKey, _ := filepath.Rel(connector, aPath)
					bKey, _ := filepath.Rel(connector, bPath)
					if expected[aKey].identity != aIdentity || expected[bKey].identity == aIdentity || expected[bKey].identity.mode != unix.S_IFDIR {
						return errors.New("actor A/B creation identity mismatch")
					}
					// Every pre-existing member remains exact at this allocation boundary.
					if err := vNextPublicationCompareExpectedMembers(expected, before, false); err != nil {
						return err
					}
					if temporary {
						return nil
					}
					return injected
				}
				hooks := vNextPublicationHooks{CloseDirectory: func(file *os.File, _ string) error { closes[file]++; return file.Close() }}
				if caller == "generation-prune" || caller == "stage-recover" {
					hooks.AfterQuarantineOpen = func(parent *vNextPublicationDirectory, name string, opened *vNextPublicationDirectory, _ vNextPublicationIdentity) error {
						return interfere(parent, name, opened, false)
					}
				} else {
					hooks.AfterTemporaryOpen = func(parent *vNextPublicationDirectory, name string, opened *vNextPublicationDirectory) error {
						return interfere(parent, name, opened, true)
					}
				}
				writer, err := newVNextGenerationPublisher(root, "acme", hooks)
				if err != nil {
					t.Fatal(err)
				}
				switch caller {
				case "public-publish":
					_, err = writer.Publish(next)
				case "generation-prune":
					err = writer.Prune(context.Background())
				case "stage-recover":
					err = writer.Recover(context.Background())
				default:
					operation, openErr := writer.openOperation(context.Background(), unix.LOCK_EX, true)
					if openErr != nil {
						t.Fatal(openErr)
					}
					if caller == "CURRENT" {
						err = writer.writeCurrentLocked(operation, current)
					} else {
						err = writer.writeJournalLocked(operation, vNextGenerationJournal{New: current, State: "prepared"})
					}
					operation.close()
				}
				if allocated == nil || closes[allocated] != 1 {
					t.Fatalf("allocated A instance Close=%d", closes[allocated])
				}
				for file, count := range closes {
					if count != 1 {
						t.Fatalf("allocation directory %p Close=%d", file, count)
					}
				}
				if attempts != 1 || err == nil || !strings.Contains(err.Error(), "identity changed") {
					t.Fatalf("actual failed allocation: attempts=%d err=%v", attempts, err)
				}
				if caller == "generation-prune" || caller == "stage-recover" {
					if !errors.Is(err, injected) {
						t.Fatalf("lost allocation completion: %v", err)
					}
				} else if !errors.Is(err, os.ErrExist) {
					t.Fatalf("did not produce actual collision: %v", err)
				}
				vNextPublicationAssertExpectedTree(t, connector, expected)
				vNextPublicationAssertCurrentJournalForTest(t, root, "failed allocation prior selection", current, "", nil, vNextGenerationPointer{})
				fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
				if err != nil {
					t.Fatal(err)
				}
				_ = fresh.Check(active) // Success depends on whether this caller left an owned stage.
				vNextPublicationAssertExpectedTree(t, connector, expected)
				// Only after actual A/B/control/history assertions remove test-owned interference.
				for _, path := range []string{bPath, aPath} {
					if err := os.RemoveAll(path); err != nil {
						t.Fatal(err)
					}
				}
				if err := fresh.Recover(context.Background()); err != nil {
					t.Fatalf("fresh recovery: %v", err)
				}
				if _, err := fresh.Publish(next); err != nil {
					t.Fatalf("ordinary retry: %v", err)
				}
				if err := fresh.Check(next); err != nil {
					t.Fatalf("retry Check: %v", err)
				}
			})
		}
	}
}
