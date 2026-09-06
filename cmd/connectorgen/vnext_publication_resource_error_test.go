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
	"sort"
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

// TestVNextPublicationPublicNestedQuarantineBoundsChildOwnership exercises the
// public cleanup callers after their root moved into the private quarantine.
// The direct removeTree controls above prove the recursive primitive; these
// rows prove the real Recover, Prune, and Publish initial-recovery paths retain the
// candidate and only test-owned interference is restored after a nested
// descriptor-bound refusal.
func TestVNextPublicationPublicNestedQuarantineBoundsChildOwnership(t *testing.T) {
	previousGC := debug.SetGCPercent(-1)
	defer debug.SetGCPercent(previousGC)

	tests := []struct {
		name    string
		prepare func(*testing.T) (string, func(*vNextGenerationPublisher) error)
	}{
		{
			name: "recover-owned-stage",
			prepare: func(t *testing.T) (string, func(*vNextGenerationPublisher) error) {
				root := t.TempDir()
				baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
				if err != nil {
					t.Fatal(err)
				}
				pointer, err := baseline.Publish(vNextPublicationArtifactsForTest("active", false))
				if err != nil {
					t.Fatal(err)
				}
				stageName := ".stage-public-nested"
				stagePath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, stageName)
				vNextWriteOwnedStageForTest(t, stagePath, pointer, stageName, "owned stage")
				vNextPublicationWriteNestedMemberForTest(t, stagePath)
				return stagePath, func(publisher *vNextGenerationPublisher) error {
					return publisher.Recover(context.Background())
				}
			},
		},
		{
			name: "prune-stale-generation",
			prepare: func(t *testing.T) (string, func(*vNextGenerationPublisher) error) {
				root := t.TempDir()
				baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
				if err != nil {
					t.Fatal(err)
				}
				old, err := baseline.Publish(vNextPublicationArtifactsWithNestedMemberForTest("old"))
				if err != nil {
					t.Fatal(err)
				}
				handle, err := baseline.Open(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				if _, err := baseline.Publish(vNextPublicationArtifactsForTest("current", false)); err != nil {
					handle.Release()
					t.Fatal(err)
				}
				handle.Release()
				return filepath.Join(root, "acme", vNextPublicationGenerationDirectory, old.Generation), func(publisher *vNextGenerationPublisher) error {
					return publisher.Prune(context.Background())
				}
			},
		},
		{
			name: "publish-initial-recovery-generation",
			prepare: func(t *testing.T) (string, func(*vNextGenerationPublisher) error) {
				root := t.TempDir()
				baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
				if err != nil {
					t.Fatal(err)
				}
				old, err := baseline.Publish(vNextPublicationArtifactsWithNestedMemberForTest("old"))
				if err != nil {
					t.Fatal(err)
				}
				handle, err := baseline.Open(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				if _, err := baseline.Publish(vNextPublicationArtifactsForTest("current", false)); err != nil {
					handle.Release()
					t.Fatal(err)
				}
				handle.Release()
				return filepath.Join(root, "acme", vNextPublicationGenerationDirectory, old.Generation), func(publisher *vNextGenerationPublisher) error {
					_, err := publisher.Publish(vNextPublicationArtifactsForTest("next", false))
					return err
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, fault := range []struct {
				name  string
				fstat bool
			}{
				{name: "nested-replacement"},
				{name: "opened-nested-child-fstat", fstat: true},
			} {
				t.Run(fault.name, func(t *testing.T) {
					before := vNextPublicationNumericFDCountForTest(t)
					const attempts = 4
					for attempt := 0; attempt < attempts; attempt++ {
						targetPath, invoke := test.prepare(t)
						worktreeRoot := filepath.Dir(filepath.Dir(filepath.Dir(targetPath)))
						expectedAuthority, authorityErr := vNextPublicationExpectedAuthority(worktreeRoot)
						if authorityErr != nil {
							t.Fatal(authorityErr)
						}
						rootIdentity := vNextPublicationDirectoryIdentityForTest(t, targetPath, "public cleanup root")
						if strings.HasPrefix(filepath.Base(targetPath), ".stage-") {
							stageMarker := vNextPublicationFileWitnessForTest(t, filepath.Join(targetPath, vNextPublicationStageOwnerFile))
							if !stageMarker.info.Mode().IsRegular() || vNextPublicationInfoInodeForTest(stageMarker.info) == 0 {
								t.Fatalf("owned stage marker is not a regular identity-bound file: mode=%v inode=%d", stageMarker.info.Mode(), vNextPublicationInfoInodeForTest(stageMarker.info))
							}
							var owner vNextPublicationStageOwner
							if err := vNextPublicationDecode(stageMarker.payload, &owner); err != nil || owner.Connector != "acme" || owner.Generation == "" || owner.Stage != filepath.Base(targetPath) {
								t.Fatalf("owned stage marker = %#v decode=%v, want acme/%q ownership", owner, err, filepath.Base(targetPath))
							}
						}
						candidateParent := filepath.Dir(targetPath)
						backupPath := filepath.Join(t.TempDir(), "fixture-owned-original")
						vNextCopyDirectoryForTest(t, targetPath, backupPath)
						injected := errors.New("injected public nested child fstat failure")
						var nestedA, nestedB vNextPublicationIdentity
						replacementHit := false
						fstatHit := false
						var nestedOpened, quarantineOpened *os.File
						closes := make(map[*os.File]int)

						if fault.fstat {
							vNextPublicationRemoveTreeChildIdentityForTest = func(file *os.File, label string) (vNextPublicationIdentity, error) {
								if !strings.HasSuffix(label, "/nested") || fstatHit {
									return vNextPublicationIdentityFromFile(file, label)
								}
								if _, err := file.Stat(); err != nil {
									return vNextPublicationIdentity{}, err
								}
								fstatHit = true
								nestedOpened = file
								return vNextPublicationIdentity{}, injected
							}
							t.Cleanup(func() { vNextPublicationRemoveTreeChildIdentityForTest = nil })
						} else {
							vNextPublicationRemoveTreeChildIdentityForTest = func(file *os.File, label string) (vNextPublicationIdentity, error) {
								if strings.HasSuffix(label, "/nested") {
									nestedOpened = file
								}
								return vNextPublicationIdentityFromFile(file, label)
							}
							t.Cleanup(func() { vNextPublicationRemoveTreeChildIdentityForTest = nil })
							vNextPublicationRemoveTreeAfterIdentityForTest = func(directory *vNextPublicationDirectory, name string) {
								if name != "nested" || replacementHit {
									return
								}
								var err error
								nestedA, err = directory.identityAt(name, "nested original A")
								if err != nil {
									t.Fatal(err)
								}
								if nestedA.mode != unix.S_IFDIR {
									t.Fatalf("nested original A mode = %#o, want directory", nestedA.mode)
								}
								if err := unix.Renameat(int(directory.file.Fd()), name, int(directory.file.Fd()), name+".A"); err != nil {
									t.Fatal(err)
								}
								if err := unix.Mkdirat(int(directory.file.Fd()), name, 0o700); err != nil {
									t.Fatal(err)
								}
								foreign, err := directory.openFile(name+"/foreign", "nested replacement B", unix.O_CREAT|unix.O_EXCL|unix.O_WRONLY, 0o600, false)
								if err != nil {
									t.Fatal(err)
								}
								if _, err := foreign.Write([]byte("nested replacement B")); err != nil {
									_ = foreign.Close()
									t.Fatal(err)
								}
								if err := foreign.Close(); err != nil {
									t.Fatal(err)
								}
								nestedB, err = directory.identityAt(name, "nested replacement B")
								if err != nil {
									t.Fatal(err)
								}
								if nestedA == nestedB || nestedB.mode != unix.S_IFDIR {
									t.Fatalf("nested replacement identities A=%#v B=%#v", nestedA, nestedB)
								}
								replacementHit = true
							}
							t.Cleanup(func() { vNextPublicationRemoveTreeAfterIdentityForTest = nil })
						}

						guard, err := newVNextGenerationPublisher(worktreeRoot, "acme", vNextPublicationHooks{
							AfterQuarantineOpen: func(_ *vNextPublicationDirectory, _ string, opened *vNextPublicationDirectory, _ vNextPublicationIdentity) error {
								quarantineOpened = opened.file
								return nil
							},
							CloseDirectory: func(file *os.File, _ string) error { closes[file]++; return file.Close() },
						})
						if err != nil {
							t.Fatal(err)
						}
						err = invoke(guard)
						for file, count := range closes {
							if count != 1 {
								t.Fatalf("directory instance %p Close count=%d", file, count)
							}
						}
						if nestedOpened == nil || closes[nestedOpened] != 1 || quarantineOpened == nil || closes[quarantineOpened] != 1 {
							t.Fatalf("actual nested/quarantine owners Close=%d/%d", closes[nestedOpened], closes[quarantineOpened])
						}
						if _, err := nestedOpened.Stat(); !errors.Is(err, fs.ErrClosed) {
							t.Fatalf("nested descriptor remains usable: %v", err)
						}
						if fault.fstat {
							if !fstatHit {
								t.Fatal("public caller did not open the nested child before the fstat injection")
							}
							if !errors.Is(err, injected) {
								t.Fatalf("public nested fstat error = %v, want %v", err, injected)
							}
						} else {
							if !replacementHit {
								t.Fatal("public caller did not reach nested post-identity/pre-open replacement")
							}
							if err == nil || !strings.Contains(err.Error(), "identity changed") {
								t.Fatalf("public nested replacement error = %v, want identity refusal", err)
							}
						}

						if _, statErr := os.Lstat(targetPath); !errors.Is(statErr, fs.ErrNotExist) {
							t.Fatalf("public cleanup root after quarantine = %v, want absent", statErr)
						}
						candidatePath := vNextPublicationQuarantineCandidateDirectoryForTest(t, candidateParent)
						if got := vNextPublicationDirectoryIdentityForTest(t, candidatePath, "quarantined cleanup candidate"); got != rootIdentity {
							t.Fatalf("quarantined candidate identity = %#v, want original root %#v", got, rootIdentity)
						}
						if fault.fstat {
							vNextPublicationAssertNestedPayloadForTest(t, filepath.Join(candidatePath, "nested", "deeper", "child.json"), []byte("nested original A"))
						} else {
							movedAPath := filepath.Join(candidatePath, "nested.A")
							replacementBPath := filepath.Join(candidatePath, "nested")
							if got := vNextPublicationDirectoryIdentityForTest(t, movedAPath, "quarantined nested original A"); got != nestedA {
								t.Fatalf("quarantined nested original A identity = %#v, want %#v", got, nestedA)
							}
							if got := vNextPublicationDirectoryIdentityForTest(t, replacementBPath, "quarantined nested replacement B"); got != nestedB {
								t.Fatalf("quarantined nested replacement B identity = %#v, want %#v", got, nestedB)
							}
							vNextPublicationAssertNestedPayloadForTest(t, filepath.Join(movedAPath, "deeper", "child.json"), []byte("nested original A"))
							vNextPublicationAssertNestedPayloadForTest(t, filepath.Join(replacementBPath, "foreign"), []byte("nested replacement B"))
						}
						if residue := vNextPublicationTreeSnapshotForTest(t, candidatePath); len(residue) == 0 {
							t.Fatal("quarantined candidate has no observable residue")
						}
						actualAuthority, authorityErr := vNextPublicationExpectedAuthority(worktreeRoot)
						if authorityErr != nil {
							t.Fatal(authorityErr)
						}
						if err := vNextPublicationCompareExpectedMembers(actualAuthority, expectedAuthority, true); err != nil {
							t.Fatal(err)
						}
						vNextPublicationAssertDurableControlsAndAuthorityForTest(t, worktreeRoot, "public nested cleanup refusal")
						// The public caller intentionally makes no all-or-nothing
						// rollback promise for a nested removal failure. After every
						// A/B/residue assertion, replace only this test fixture with its
						// pre-call backup. Production does not promise that restoration.
						if err := os.RemoveAll(candidatePath); err != nil {
							t.Fatalf("remove observed fixture quarantine candidate: %v", err)
						}
						if err := os.Remove(filepath.Dir(candidatePath)); err != nil {
							t.Fatalf("remove observed empty quarantine fixture: %v", err)
						}
						if err := os.Rename(backupPath, targetPath); err != nil {
							t.Fatalf("restore fixture-owned cleanup root: %v", err)
						}

						fresh, err := newVNextGenerationPublisher(worktreeRoot, "acme", vNextPublicationHooks{})
						if err != nil {
							t.Fatal(err)
						}
						if err := fresh.Recover(context.Background()); err != nil {
							t.Fatalf("fresh recovery after fixture-only restoration: %v", err)
						}
						vNextPublicationAssertNoOpenDescriptorsUnderForTest(t, worktreeRoot)
					}
					if after := vNextPublicationNumericFDCountForTest(t); after != before {
						t.Fatalf("public nested cleanup descriptors grew: before=%d after=%d", before, after)
					}
				})
			}
		})
	}
}

func vNextPublicationArtifactsWithNestedMemberForTest(marker string) vNextPublicationArtifacts {
	artifacts := vNextPublicationArtifactsForTest(marker, false)
	artifacts.Files["nested/deeper/child.json"] = []byte("nested original A")
	return artifacts
}

func vNextPublicationWriteNestedMemberForTest(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "nested", "deeper"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "nested", "deeper", "child.json"), []byte("nested original A"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func vNextPublicationDirectoryIdentityForTest(t *testing.T, path, label string) vNextPublicationIdentity {
	t.Helper()
	directory, err := vNextPublicationOpenDirectory(path, label)
	if err != nil {
		t.Fatalf("open %s %q: %v", label, path, err)
	}
	identity, identityErr := vNextPublicationIdentityFromFile(directory.file, label)
	closeErr := directory.Close()
	if identityErr != nil {
		t.Fatalf("observe %s %q: %v", label, path, identityErr)
	}
	if closeErr != nil {
		t.Fatalf("close %s %q: %v", label, path, closeErr)
	}
	return identity
}

func vNextPublicationQuarantineCandidateDirectoryForTest(t *testing.T, parent string) string {
	t.Helper()
	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), ".connectorgen-quarantine-") {
			candidate := filepath.Join(parent, entry.Name(), vNextPublicationQuarantineMember)
			if identity := vNextPublicationDirectoryIdentityForTest(t, candidate, "quarantined cleanup candidate"); identity.mode == unix.S_IFDIR {
				return candidate
			}
		}
	}
	t.Fatal("public cleanup did not retain a quarantined directory candidate")
	return ""
}

func vNextPublicationAssertNestedPayloadForTest(t *testing.T, path string, want []byte) {
	t.Helper()
	witness := vNextPublicationFileWitnessForTest(t, path)
	if !bytes.Equal(witness.payload, want) {
		t.Fatalf("nested payload %q = %q, want %q", path, witness.payload, want)
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

// vNextPublicationAssertNoOpenDescriptorsUnderForTest makes the F-02 public
// path resource boundary concrete: after each fixture-only recovery, no root,
// quarantine, child, or connector-lock descriptor may still name that unique
// test tree. The surrounding numeric count and disabled GC catch a separate
// descriptor accumulating without relying on finalizers.
func vNextPublicationAssertNoOpenDescriptorsUnderForTest(t *testing.T, root string) {
	t.Helper()
	lsof, err := exec.LookPath("lsof")
	if err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(lsof, "-n", "-P", "-p", strconv.Itoa(os.Getpid()), "-Ffn").Output()
	if err != nil {
		t.Fatal(err)
	}
	var descriptors []string
	fd := ""
	for _, line := range strings.Split(string(output), "\n") {
		if len(line) < 2 {
			continue
		}
		switch line[0] {
		case 'f':
			if _, err := strconv.Atoi(line[1:]); err == nil {
				fd = line[1:]
			} else {
				fd = ""
			}
		case 'n':
			if fd != "" && strings.HasPrefix(line[1:], root) {
				descriptors = append(descriptors, fd+":"+line[1:])
			}
		}
	}
	if len(descriptors) != 0 {
		sort.Strings(descriptors)
		t.Fatalf("public nested cleanup retained descriptors under %q: %v", root, descriptors)
	}
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
