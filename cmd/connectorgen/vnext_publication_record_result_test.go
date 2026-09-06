package main

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

// All three immutable record callers use this writer. Their six completion
// frontiers and adoption/recovery behavior are exercised separately by
// TestCP11OwnershipRecordFrontiersRecover; these controls cover its remaining
// result branches without multiplying unrelated caller/state combinations.
func TestCP11OwnershipRecordResultBoundaries(t *testing.T) {
	for _, boundary := range []string{"not-created-collision", "replaced-after-close", "incomplete-disposal-refused"} {
		t.Run(boundary, func(t *testing.T) {
			root := t.TempDir()
			name := vNextPublicationControlRepairPreparedFile
			path := filepath.Join(root, name)
			directory, err := vNextPublicationOpenDirectory(root, "record result root")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := directory.Close(); err != nil {
					t.Errorf("close test-owned directory: %v", err)
				}
			}()
			primary := errors.New("actual partial write completion")
			completion := errors.New("actual record Close completion")
			payload := []byte("complete immutable record\n")
			writes, closes := 0, 0
			var actual *os.File
			var expected vNextPublicationExpectedTree
			capture := func() {
				var err error
				expected, err = vNextPublicationObserveExpectedTree(root)
				if err != nil {
					t.Fatal(err)
				}
			}
			replace := func() {
				if err := os.Rename(path, path+"-owned-A"); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("foreign replacement B"), 0o600); err != nil {
					t.Fatal(err)
				}
				capture()
			}
			if boundary == "not-created-collision" {
				if err := os.WriteFile(path, []byte("preexisting record B"), 0o600); err != nil {
					t.Fatal(err)
				}
				capture()
			}
			result, err := vNextPublicationWriteControlRepairRecord(directory, name, "result record", payload, vNextPublicationControlRecordHooks{
				Write: func(file *os.File, _ string, b []byte) (int, error) {
					actual = file
					writes++
					if boundary == "incomplete-disposal-refused" {
						n, err := file.Write(b[:1])
						replace()
						return n, errors.Join(err, primary)
					}
					return file.Write(b)
				},
				Close: func(file *os.File, _ string) error {
					closes++
					if file != actual {
						t.Error("Close does not own the written descriptor")
					}
					err := file.Close()
					if boundary == "replaced-after-close" {
						replace()
					}
					return errors.Join(err, completion)
				},
			})
			if boundary == "not-created-collision" {
				if result.created || result.contentComplete || result.identity != (vNextPublicationIdentity{}) || result.disposition != vNextPublicationRecordNotCreated || writes != 0 || closes != 0 || !errors.Is(err, fs.ErrExist) {
					t.Fatalf("unowned collision result=%#v writes=%d closes=%d err=%v", result, writes, closes, err)
				}
			} else {
				if !result.created || writes != 1 || closes != 1 || !errors.Is(err, completion) || !strings.Contains(err.Error(), "identity changed") {
					t.Fatalf("owned result=%#v writes=%d closes=%d err=%v", result, writes, closes, err)
				}
				if result.identity != expected[name+"-owned-A"].identity || result.identity == expected[name].identity {
					t.Fatal("result adopted replacement B")
				}
				if boundary == "replaced-after-close" {
					if !result.contentComplete || result.disposition != vNextPublicationRecordUnknown || string(expected[name+"-owned-A"].payload) != string(payload) {
						t.Fatalf("final binding result=%#v", result)
					}
				} else if result.contentComplete || result.disposition != vNextPublicationRecordRetainedIncomplete || !errors.Is(err, primary) || !errors.Is(err, io.ErrShortWrite) || string(expected[name+"-owned-A"].payload) != string(payload[:1]) {
					t.Fatalf("failed bound disposal result=%#v err=%v", result, err)
				}
			}
			vNextPublicationAssertExpectedTree(t, root, expected)
			// Unknown/replaced records are forensic residue, not valid recoverable
			// authority. No fresh-Recover success is promised for these fixtures.
		})
	}
}

func TestCP11OwnershipRemainingControlStates(t *testing.T) {
	for _, test := range []struct {
		name, target    string
		prior, intended bool
	}{
		{"CURRENT-absent-present", vNextPublicationCurrentFile, false, true},
		{"CURRENT-present-absent", vNextPublicationCurrentFile, true, false},
		{"JOURNAL-present-present", vNextPublicationJournalFile, true, true},
		{"JOURNAL-absent-absent-noop", vNextPublicationJournalFile, false, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			artifacts := vNextPublicationArtifactsForTest("state-boundary", false)
			pointer, err := baseline.Publish(artifacts)
			if err != nil {
				t.Fatal(err)
			}
			operation, err := baseline.openOperation(context.Background(), unix.LOCK_EX, true)
			if err != nil {
				t.Fatal(err)
			}
			if test.target == vNextPublicationCurrentFile && !test.prior {
				err = baseline.removeCurrentLocked(operation)
			}
			if test.target == vNextPublicationJournalFile && test.prior {
				err = baseline.writeJournalLocked(operation, vNextGenerationJournal{New: pointer, State: "prepared"})
			}
			operation.close()
			if err != nil {
				t.Fatal(err)
			}
			connector := filepath.Join(root, "acme")
			prior, err := vNextPublicationExpectedAuthority(root)
			if err != nil {
				t.Fatal(err)
			}
			_, present := prior[test.target]
			if present != test.prior {
				t.Fatalf("fixture prior present=%t want=%t", present, test.prior)
			}
			known := vNextPublicationRepairTransactionsForTest(t, connector)
			fault := errors.New("complete prepared record Sync completion")
			fired := 0
			var beforeRecord vNextPublicationExpectedTree
			writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{
				At: func(point vNextPublicationFaultPoint) error {
					if point == vNextPublicationBeforeControlRepairRecord {
						var err error
						beforeRecord, err = vNextPublicationExpectedAuthority(root)
						return err
					}
					return nil
				},
				ControlRecord: vNextPublicationControlRecordHooks{Sync: func(file *os.File, label string) error {
					err := file.Sync()
					if label == "publication control repair prepared authority" {
						fired++
						return errors.Join(err, fault)
					}
					return err
				}},
			})
			if err != nil {
				t.Fatal(err)
			}
			operation, err = writer.openOperation(context.Background(), unix.LOCK_EX, true)
			if err != nil {
				t.Fatal(err)
			}
			if !test.intended {
				err = writer.removeControlLocked(operation, test.target)
			} else if test.target == vNextPublicationCurrentFile {
				err = writer.writeCurrentLocked(operation, pointer)
			} else {
				err = writer.writeJournalLocked(operation, vNextGenerationJournal{New: pointer, State: "committed"})
			}
			operation.close()
			// Absence-to-absence is a logical no-op, still an explicit durable transition.

			if fired != 1 || !errors.Is(err, fault) {
				t.Fatalf("actual prepared completion fired=%d err=%v", fired, err)
			}
			after, err := vNextPublicationExpectedAuthority(root)
			if err != nil {
				t.Fatal(err)
			}
			if beforeRecord == nil {
				t.Fatal("missing pre-record anchor observation")
			}
			if err := vNextPublicationCompareExpectedMembers(after, beforeRecord, false); err != nil {
				t.Fatal(err)
			}
			if err := vNextPublicationCompareExpectedMembers(after, prior, false); err != nil {
				t.Fatal(err)
			}
			vNextPublicationAssertPendingPreparedGraphForTest(t, writer, connector, test.target, test.prior, test.intended, known)
			beforeCheck, err := vNextPublicationObserveExpectedTree(connector)
			if err != nil {
				t.Fatal(err)
			}
			fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			checkErr := fresh.Check(artifacts)
			if checkErr == nil {
				t.Fatal("Check accepted pending durable transition")
			}
			vNextPublicationAssertExpectedTree(t, connector, beforeCheck)
			if err := fresh.Recover(context.Background()); err != nil {
				t.Fatal(err)
			}
			// Recovery may validly select prior absence for interrupted CURRENT
			// installation. A subsequent Publish establishes the queryable head.
			if _, err := fresh.Publish(artifacts); err != nil {
				t.Fatal(err)
			}
			if err := fresh.Check(artifacts); err != nil {
				t.Fatal(err)
			}
		})
	}
}
