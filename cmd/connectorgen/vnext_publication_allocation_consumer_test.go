package main

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

func TestCP11AllocationConsumersPreserveCompoundAndState(t *testing.T) {
	for _, cut := range []string{"prepared-JOURNAL", "CURRENT-selection", "committed-JOURNAL", "rollback", "restore-Recover", "restore-Open", "restore-Prune", "restore-Publish"} {
		t.Run(cut, func(t *testing.T) {
			root := t.TempDir()
			baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			oldSet := vNextPublicationArtifactsForTest("old", true)
			newSet := vNextPublicationArtifactsForTest("allocation-consumer-new", true)
			restoring := len(cut) > 8 && cut[:8] == "restore-"
			var old vNextGenerationPointer
			var rejectedPath string
			var rejectedBytes []byte
			if restoring {
				old, _, rejectedPath, rejectedBytes = vNextPublicationPrepareRejectedNewForRecoveryCloseTest(t, root)
				newSet = vNextPublicationArtifactsForTest("rejected", true)
			} else {
				old, err = baseline.Publish(oldSet)
				if err != nil {
					t.Fatal(err)
				}
			}
			connector := filepath.Join(root, "acme")
			newPointer := vNextGenerationPointer{Generation: vNextPublicationGenerationID(newSet.Files)}
			// Compare generation names below: the selected pointer's remaining
			// integrity fields are validated by the real fresh Open/marker proof.
			failAt := 1
			if cut == "CURRENT-selection" {
				failAt = 2
			}
			if cut == "committed-JOURNAL" || cut == "rollback" {
				failAt = 3
			}
			completion := errors.New("actual allocator parent Close completion")
			validation := errors.New("actual active validation rejection")
			attempts, parentCloses, rootCloses, validationCalls := 0, 0, 0, 0
			var failedRoot *os.File
			var expected vNextPublicationExpectedTree
			hooks := vNextPublicationHooks{
				AfterTemporaryOpen: func(_ *vNextPublicationDirectory, _ string, opened *vNextPublicationDirectory) error {
					attempts++
					if attempts != failAt {
						return nil
					}
					failedRoot = opened.file
					fd, err := unix.Openat(int(opened.file.Fd()), vNextPublicationTemporaryFile, unix.O_CREAT|unix.O_EXCL|unix.O_WRONLY, 0o600)
					if err != nil {
						return err
					}
					blocker := os.NewFile(uintptr(fd), "allocator actual collision")
					if blocker == nil {
						return errors.Join(unix.Close(fd), errors.New("invalid blocker fd"))
					}
					_, writeErr := blocker.Write([]byte("retained allocator collision A"))
					if err := errors.Join(writeErr, blocker.Close()); err != nil {
						return err
					}
					expected, err = vNextPublicationObserveExpectedTree(connector)
					return err
				},
				CloseDirectory: func(file *os.File, label string) error {
					isRoot := failedRoot != nil && file == failedRoot
					isParent := failedRoot != nil && label == "publication temporary" && parentCloses == 0
					err := file.Close()
					if isRoot {
						rootCloses++
					}
					if isParent {
						parentCloses++
						return errors.Join(err, completion)
					}
					return err
				},
			}
			writer, err := newVNextGenerationPublisher(root, "acme", hooks)
			if err != nil {
				t.Fatal(err)
			}
			if cut == "rollback" {
				newSet.Validate = func(fs.FS) error {
					validationCalls++
					if validationCalls == 2 {
						return validation
					}
					return nil
				}
			}
			switch cut {
			case "restore-Recover":
				err = writer.Recover(context.Background())
			case "restore-Open":
				var handle *vNextGenerationHandle
				handle, err = writer.Open(context.Background())
				if handle != nil {
					handle.Release()
					t.Error("Open returned handle across failed restoration")
				}
			case "restore-Prune":
				err = writer.Prune(context.Background())
			default:
				_, err = writer.Publish(newSet)
			}
			if attempts != failAt || parentCloses != 1 || rootCloses != 1 || !errors.Is(err, fs.ErrExist) || !errors.Is(err, completion) {
				t.Fatalf("allocation consumer attempts=%d want=%d parent/root Close=%d/%d err=%v", attempts, failAt, parentCloses, rootCloses, err)
			}
			if cut == "rollback" && (validationCalls != 2 || !errors.Is(err, validation)) {
				t.Fatalf("rollback lost validation cause: calls=%d err=%v", validationCalls, err)
			}
			if expected == nil {
				t.Fatal("missing actual allocation cut")
			}
			vNextPublicationAssertExpectedTree(t, connector, expected)
			var current vNextGenerationPointer
			if err := json.Unmarshal(expected[vNextPublicationCurrentFile].payload, &current); err != nil {
				t.Fatal(err)
			}
			wantCurrent := old.Generation
			if cut == "committed-JOURNAL" || cut == "rollback" || restoring {
				wantCurrent = newPointer.Generation
			}
			if current.Generation != wantCurrent {
				t.Fatalf("selected generation=%s want=%s", current.Generation, wantCurrent)
			}
			journalMember, hasJournal := expected[vNextPublicationJournalFile]
			if hasJournal != (cut != "prepared-JOURNAL") {
				t.Fatalf("unexpected JOURNAL presence=%t", hasJournal)
			}
			if hasJournal {
				var journal vNextGenerationJournal
				if err := json.Unmarshal(journalMember.payload, &journal); err != nil {
					t.Fatal(err)
				}
				if journal.State != "prepared" || journal.New.Generation != newPointer.Generation || journal.Old == nil || *journal.Old != old {
					t.Fatalf("wrong independently expected journal %#v", journal)
				}
			}
			fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			// Check may be healthy at the failed first JOURNAL allocation, but
			// must not alter any selected control, history, generation or residue.
			checkErr := fresh.Check(oldSet)
			if cut != "prepared-JOURNAL" && checkErr == nil {
				t.Fatal("Check accepted interrupted publication")
			}
			vNextPublicationAssertExpectedTree(t, connector, expected)
			recoverErr := fresh.Recover(context.Background())
			if restoring {
				// The deliberately corrupted rejected generation is not owned deletion
				// authority. Recovery restores old CURRENT, then must refuse to prune it.
				if recoverErr == nil || !strings.Contains(recoverErr.Error(), "without validated publication ownership") {
					t.Fatalf("invalid rejected generation refusal: %v", recoverErr)
				}
				restored, readErr := os.ReadFile(filepath.Join(connector, vNextPublicationCurrentFile))
				var selected vNextGenerationPointer
				if readErr != nil {
					t.Fatal(readErr)
				}
				if err := json.Unmarshal(restored, &selected); err != nil || selected != old {
					t.Fatalf("restoration before refusal selected=%#v err=%v", selected, err)
				}
				if err := os.WriteFile(rejectedPath, rejectedBytes, 0o600); err != nil {
					t.Fatal(err)
				}
				if err := fresh.Recover(context.Background()); err != nil {
					t.Fatal(err)
				}
			} else if recoverErr != nil {
				t.Fatal(recoverErr)
			}
			handle, err := fresh.Open(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			wantMarker := "old"
			if cut == "committed-JOURNAL" || cut == "rollback" {
				wantMarker = "allocation-consumer-new"
			}
			assertVNextPublicationMarker(t, handle, wantMarker)
			handle.Release()
			if _, err := os.Lstat(filepath.Join(connector, vNextPublicationJournalFile)); !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("recovery retained journal: %v", err)
			}
			if _, err := fresh.Publish(vNextPublicationArtifactsForTest("consumer-retry", true)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
