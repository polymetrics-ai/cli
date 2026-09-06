package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

// This witness names the actual caller above appendControlRepairPhaseLocked;
// a shared phase-write hook alone cannot prove which transition reached it.
func vNextPublicationPhaseCallerForTest() string {
	var pcs [32]uintptr
	n := runtime.Callers(2, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	afterAppender := false
	for {
		frame, more := frames.Next()
		if afterAppender {
			return frame.Function
		}
		if strings.HasSuffix(frame.Function, ".appendControlRepairPhaseLocked") {
			afterAppender = true
		}
		if !more {
			break
		}
	}
	return ""
}

func TestCP11OwnershipSevenPhaseCallers(t *testing.T) {
	for _, caller := range []string{"createControlRepairLocked", "resumeBaseControlAuthorityLocked", "beginControlCaptureLocked", "completeControlCaptureLocked", "selectControlStateLocked", "terminalizeRetryLocked", "resolveControlRepairLocked"} {
		for _, target := range []string{vNextPublicationCurrentFile, vNextPublicationJournalFile} {
			for _, frontier := range []string{"parent-close-after-create", "partial-write", "sync-completion"} {
				t.Run(caller+"/"+target+"/"+frontier, func(t *testing.T) {
					root := t.TempDir()
					expectedSequence := 1
					expectedState := vNextPublicationControlRepairTerminal
					switch caller {
					case "beginControlCaptureLocked":
						expectedState = vNextPublicationControlRepairCaptureIntent
					case "completeControlCaptureLocked":
						expectedSequence = 2
						expectedState = vNextPublicationControlRepairCaptured
					case "selectControlStateLocked":
						expectedSequence = 3
						expectedState = vNextPublicationControlRepairSelected
					case "terminalizeRetryLocked", "resolveControlRepairLocked":
						expectedSequence = 4
					}
					artifacts := vNextPublicationArtifactsForTest("phase-frontier", false)
					baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
					if err != nil {
						t.Fatal(err)
					}
					var pointer vNextGenerationPointer
					baseCase := caller == "createControlRepairLocked" || caller == "resumeBaseControlAuthorityLocked"
					if !baseCase {
						pointer, err = baseline.Publish(artifacts)
						if err != nil {
							t.Fatal(err)
						}
					}
					if caller == "resumeBaseControlAuthorityLocked" {
						interrupted := errors.New("seed valid phase-empty base")
						lastTarget := ""
						baseline.hooks.ControlRecord.Write = func(file *os.File, label string, payload []byte) (int, error) {
							if label == "publication control repair prepared authority" {
								var prepared vNextPublicationControlRepair
								if err := json.Unmarshal(payload, &prepared); err != nil {
									return 0, err
								}
								lastTarget = prepared.Target
							}
							return file.Write(payload)
						}
						baseline.hooks.At = func(point vNextPublicationFaultPoint) error {
							if point == vNextPublicationAfterBaseControlRepairPrepared && lastTarget == target {
								return interrupted
							}
							return nil
						}
						if _, err := baseline.Publish(artifacts); !errors.Is(err, interrupted) {
							t.Fatalf("seed phase-empty %s: %v", target, err)
						}
					}
					prior, err := vNextPublicationExpectedAuthority(root)
					if err != nil && !baseCase {
						t.Fatal(err)
					}
					if baseCase && prior == nil {
						prior = make(vNextPublicationExpectedTree)
					}
					byDigest := map[string]string{}
					for path, member := range prior {
						if filepath.Base(path) == vNextPublicationControlRepairPreparedFile {
							var prepared vNextPublicationControlRepair
							if err := json.Unmarshal(member.payload, &prepared); err != nil {
								t.Fatal(err)
							}
							byDigest[vNextPublicationDigest(member.payload)] = prepared.Target
						}
					}
					failure := errors.New("phase " + frontier)
					fired := false
					closed := 0
					substitutions := 0
					var created vNextPublicationIdentity
					var expected vNextPublicationExpectedTree
					observe := func() error {
						var err error
						expected, err = vNextPublicationExpectedAuthority(root)
						if err != nil {
							return err
						}
						// Existing stable private records must not be lost before this cut.
						stable := make(vNextPublicationExpectedTree)
						for path, member := range prior {
							if strings.HasPrefix(path, vNextPublicationControlRepairDirectoryPrefix) || path == vNextPublicationControlAuthorityMarkerFile {
								stable[path] = member
							}
						}
						if err := vNextPublicationCompareExpectedMembers(expected, stable, false); err != nil {
							return err
						}
						if frontier != "sync-completion" {
							for path, member := range expected {
								if member.identity == created {
									delete(expected, path)
								}
							}
						}
						return nil
					}
					hooks := vNextPublicationHooks{}
					hooks.At = func(point vNextPublicationFaultPoint) error {
						if caller == "terminalizeRetryLocked" && point == vNextPublicationAfterControlRepairSelected && substitutions == 0 {
							substitutions++
							path := filepath.Join(root, "acme", target)
							if err := os.Rename(path, path+".fixture-selected"); err != nil {
								return err
							}
							return os.WriteFile(path, []byte("fixture late occupant"), 0o600)
						}
						return nil
					}
					hooks.ControlRecord.Write = func(file *os.File, label string, payload []byte) (int, error) {
						if label == "publication control repair prepared authority" {
							var prepared vNextPublicationControlRepair
							if err := json.Unmarshal(payload, &prepared); err != nil {
								return 0, err
							}
							byDigest[vNextPublicationDigest(payload)] = prepared.Target
						}
						if fired || label != "publication control repair phase" || !strings.HasSuffix(vNextPublicationPhaseCallerForTest(), "."+caller) {
							return file.Write(payload)
						}
						var phase vNextPublicationControlRepairPhase
						if err := json.Unmarshal(payload, &phase); err != nil {
							return 0, err
						}
						if byDigest[phase.PreparedDigest] != target || frontier == "parent-close-after-create" {
							return file.Write(payload)
						}
						if phase.Sequence != expectedSequence || phase.State != expectedState {
							return 0, errors.New("actual phase does not match independently expected caller/sequence")
						}
						if expectedState == vNextPublicationControlRepairTerminal {
							outcome := vNextPublicationControlRepairCommitted
							if caller == "terminalizeRetryLocked" {
								outcome = vNextPublicationControlRepairRetryRequired
							}
							if phase.Outcome != outcome {
								return 0, errors.New("unexpected terminal outcome")
							}
						}
						var err error
						created, err = vNextPublicationIdentityFromFile(file, label)
						if err != nil {
							return 0, err
						}
						fired = true
						if frontier == "partial-write" {
							n, err := file.Write(payload[:1])
							return n, errors.Join(err, failure, observe())
						}
						n, err := file.Write(payload)
						return n, errors.Join(err, observe())
					}
					hooks.CloseDirectory = func(file *os.File, label string) error {
						hit := !fired && frontier == "parent-close-after-create" && label == "publication control repair phase" && strings.HasSuffix(vNextPublicationPhaseCallerForTest(), "."+caller)
						if hit {
							fd, err := unix.Openat(int(file.Fd()), vNextPublicationControlRepairPreparedFile, unix.O_RDONLY|unix.O_NOFOLLOW, 0)
							if err != nil {
								return errors.Join(err, file.Close())
							}
							preparedFile := os.NewFile(uintptr(fd), "phase expected prepared")
							if preparedFile == nil {
								return errors.Join(unix.Close(fd), file.Close(), errors.New("invalid prepared descriptor"))
							}
							payload, readErr := io.ReadAll(preparedFile)
							closeErr := preparedFile.Close()
							var prepared vNextPublicationControlRepair
							if err := errors.Join(readErr, closeErr, json.Unmarshal(payload, &prepared)); err != nil {
								return errors.Join(err, file.Close())
							}
							hit = prepared.Target == target
							if hit {
								for _, sequence := range []int{expectedSequence} {
									var stat unix.Stat_t
									if unix.Fstatat(int(file.Fd()), vNextPublicationControlRepairPhaseName(sequence), &stat, unix.AT_SYMLINK_NOFOLLOW) == nil && stat.Size == 0 && uint32(stat.Mode)&unix.S_IFMT == unix.S_IFREG {
										created = vNextPublicationIdentityFromStat(stat)
										fired = true
										break
									}
								}
								hit = fired
								if !hit {
									return file.Close()
								}
								if err := observe(); err != nil {
									return errors.Join(err, file.Close())
								}
							}
						}
						err := file.Close()
						if hit {
							return errors.Join(err, failure)
						}
						return err
					}
					hooks.ControlRecord.Sync = func(file *os.File, label string) error {
						err := file.Sync()
						identity, idErr := vNextPublicationIdentityFromFile(file, label)
						if fired && identity == created && frontier == "sync-completion" {
							return errors.Join(err, idErr, failure)
						}
						return errors.Join(err, idErr)
					}
					hooks.ControlRecord.Close = func(file *os.File, label string) error {
						identity, idErr := vNextPublicationIdentityFromFile(file, label)
						if fired && identity == created {
							closed++
						}
						return errors.Join(idErr, file.Close())
					}
					writer, err := newVNextGenerationPublisher(root, "acme", hooks)
					if err != nil {
						t.Fatal(err)
					}
					switch caller {
					case "createControlRepairLocked":
						_, err = writer.Publish(artifacts)
					case "resumeBaseControlAuthorityLocked":
						err = writer.Recover(context.Background())
					default:
						operation, openErr := writer.openOperation(context.Background(), unix.LOCK_EX, true)
						if openErr != nil {
							t.Fatal(openErr)
						}
						if target == vNextPublicationCurrentFile {
							err = writer.writeCurrentLocked(operation, pointer)
						} else {
							err = writer.writeJournalLocked(operation, vNextGenerationJournal{New: pointer, State: "prepared"})
						}
						operation.close()
					}
					if !fired || closed != 1 || !errors.Is(err, failure) {
						t.Fatalf("actual caller/cut cause: fired=%t closed=%d err=%v", fired, closed, err)
					}
					if caller == "terminalizeRetryLocked" && substitutions != 1 {
						t.Fatal("retry terminal did not follow actual late selection")
					}
					actual, err := vNextPublicationExpectedAuthority(root)
					if err != nil {
						t.Fatal(err)
					}
					if err := vNextPublicationCompareExpectedMembers(actual, expected, true); err != nil {
						t.Fatalf("post-cut authority: %v", err)
					}
					fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
					if err != nil {
						t.Fatal(err)
					}
					checkErr := fresh.Check(artifacts)
					if frontier != "sync-completion" || (caller != "resolveControlRepairLocked" && caller != "createControlRepairLocked" && caller != "resumeBaseControlAuthorityLocked") {
						if checkErr == nil {
							t.Fatal("pending authority Check succeeded")
						}
					}
					afterCheck, err := vNextPublicationExpectedAuthority(root)
					if err != nil {
						t.Fatal(err)
					}
					if err := vNextPublicationCompareExpectedMembers(afterCheck, actual, true); err != nil {
						t.Fatalf("Check mutation: %v", err)
					}
					if err := fresh.Recover(context.Background()); err != nil {
						t.Fatalf("fresh recovery of %s %s: %v", caller, target, err)
					}
					if _, err := fresh.Publish(artifacts); err != nil {
						t.Fatalf("ordinary retry: %v", err)
					}
					if err := fresh.Check(artifacts); err != nil {
						t.Fatalf("retry Check: %v", err)
					}
				})
			}
		}
	}
}

func TestCP11ExpectedPhaseWitnessRejectsWrongCaller(t *testing.T) {
	if got := vNextPublicationPhaseCallerForTest(); got != "" {
		t.Fatalf("non-phase stack accepted as %q", got)
	}
}
