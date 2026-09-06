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
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"

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

func TestVNextGenerationPublisherOpenKeepsValidatedDirectory(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	pointer, err := publisher.Publish(vNextPublicationArtifactsForTest("validated-A", false))
	if err != nil {
		t.Fatalf("Publish(validated A) error = %v", err)
	}

	validatedPath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, pointer.Generation)
	retainedPath := filepath.Join(root, "retained-validated-A")
	validatedA := vNextPublicationDirectoryWitnessForTest(t, validatedPath)
	var retainedA, replacementB vNextPublicationPathWitness
	reader, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationAfterOpenValidation {
			return nil
		}
		if err := os.Rename(validatedPath, retainedPath); err != nil {
			return fmt.Errorf("move validated A: %w", err)
		}
		retainedA = vNextPublicationDirectoryWitnessForTest(t, retainedPath)
		if !os.SameFile(validatedA.info, retainedA.info) || !bytes.Equal(validatedA.payload, retainedA.payload) {
			return fmt.Errorf("retained validated A changed during displacement")
		}
		if err := vNextPublicationCopyTreeForTest(retainedPath, validatedPath); err != nil {
			return fmt.Errorf("install replacement B: %w", err)
		}
		if err := os.WriteFile(filepath.Join(validatedPath, "metadata.json"), []byte(`{"marker":"replacement-B"}`), 0o600); err != nil {
			return err
		}
		replacementB = vNextPublicationDirectoryWitnessForTest(t, validatedPath)
		vNextPublicationAssertDistinctWitnessForTest(t, "validated reader A/B", retainedA, replacementB)
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}

	handle, err := reader.Open(context.Background())
	if err != nil {
		t.Fatalf("Open() after A/B substitution error = %v", err)
	}
	defer handle.Release()
	if got, want := handle.Generation(), pointer.Generation; got != want {
		t.Fatalf("Open() generation = %q, want validated %q", got, want)
	}
	assertVNextPublicationMarker(t, handle, "validated-A")
	handleInfo, err := handle.filesRoot.file.Stat()
	if err != nil {
		t.Fatalf("stat returned A descriptor: %v", err)
	}
	if !handleInfo.IsDir() || !os.SameFile(handleInfo, retainedA.info) {
		t.Fatalf("returned A descriptor identity/type = mode %v dev/inode=%v/%v, want retained A mode %v dev/inode=%v/%v", handleInfo.Mode(), vNextPublicationInfoDeviceForTest(handleInfo), vNextPublicationInfoInodeForTest(handleInfo), retainedA.info.Mode(), vNextPublicationInfoDeviceForTest(retainedA.info), vNextPublicationInfoInodeForTest(retainedA.info))
	}
	if got, err := handle.ReadFile("metadata.json"); err != nil || !bytes.Equal(got, retainedA.payload) {
		t.Fatalf("returned A metadata = %q, err=%v, want retained A bytes %q", got, err, retainedA.payload)
	}
	vNextPublicationAssertDirectoryWitnessForTest(t, "displaced validated A", retainedPath, retainedA)
	vNextPublicationAssertDirectoryWitnessForTest(t, "installed replacement B", validatedPath, replacementB)
}

func TestVNextGenerationPublisherHeldGenerationUsesStableCleanupLock(t *testing.T) {
	for _, cleanup := range []string{"prune", "publish", "recover"} {
		t.Run(cleanup, func(t *testing.T) {
			root := t.TempDir()
			publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			old, err := publisher.Publish(vNextPublicationArtifactsForTest("held-A", false))
			if err != nil {
				t.Fatalf("Publish(A) error = %v", err)
			}
			held, err := publisher.Open(context.Background())
			if err != nil {
				t.Fatalf("Open(A) error = %v", err)
			}
			defer held.Release()

			filesA := vNextPublicationArtifactsForTest("held-A", false)
			filesB := vNextPublicationArtifactsForTest("current-B", false)
			filesC := vNextPublicationArtifactsForTest("current-C", false)
			desiredA := vNextPublicationFixturePointer(t, filesA)
			desiredB := vNextPublicationFixturePointer(t, filesB)
			desiredC := vNextPublicationFixturePointer(t, filesC)
			if old != desiredA {
				t.Fatal("held A setup differs from fixture")
			}
			var plan *vNextPublicationExpectedPlan
			leasePath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, old.Generation, vNextPublicationLeaseFile)
			retainedLease := filepath.Join(root, "retained-"+old.Generation+".lease")
			originalLeaseA := vNextPublicationFileWitnessForTest(t, leasePath)
			var displacedLeaseA, replacementLeaseB vNextPublicationPathWitness
			var heldExpected, priorAuthority vNextPublicationExpectedTree

			replaceLease := func() error {
				if err := os.Rename(leasePath, retainedLease); err != nil {
					return fmt.Errorf("move held lease: %w", err)
				}
				displacedLeaseA = vNextPublicationFileWitnessForTest(t, retainedLease)
				if !os.SameFile(originalLeaseA.info, displacedLeaseA.info) || !bytes.Equal(originalLeaseA.payload, displacedLeaseA.payload) {
					return fmt.Errorf("displaced held lease A changed")
				}
				if err := os.WriteFile(leasePath, nil, 0o600); err != nil {
					return err
				}
				replacementLeaseB = vNextPublicationFileWitnessForTest(t, leasePath)
				if plan != nil {
					plan.leaseDelta(t, old.Generation, leasePath, retainedLease)
				}
				if !bytes.Equal(displacedLeaseA.payload, replacementLeaseB.payload) {
					return fmt.Errorf("equal-byte lease mutation control has unequal bytes")
				}
				vNextPublicationAssertDistinctWitnessForTest(t, "held cleanup equal-byte A/B", displacedLeaseA, replacementLeaseB)
				var err error
				heldExpected, err = vNextPublicationObserveExpectedTree(filepath.Dir(leasePath))
				if err != nil {
					return err
				}
				priorAuthority, err = vNextPublicationExpectedStableAuthority(root)
				if err != nil {
					return err
				}

				return nil
			}
			restoreLease := func() {
				t.Helper()
				if err := os.Remove(leasePath); err != nil {
					t.Fatalf("remove replacement lease: %v", err)
				}
				if err := os.Rename(retainedLease, leasePath); err != nil {
					t.Fatalf("restore held lease: %v", err)
				}
			}

			var expectedCurrent vNextGenerationPointer
			switch cleanup {
			case "prune":
				current, err := publisher.Publish(vNextPublicationArtifactsForTest("current-B", false))
				if err != nil {
					t.Fatalf("Publish(B) error = %v", err)
				}
				if current != desiredB {
					t.Fatal("held prune setup B differs from fixture")
				}
				expectedCurrent = desiredB
				if err := replaceLease(); err != nil {
					t.Fatal(err)
				}
				plan = vNextPublicationNewExpectedPlan(t, root, map[string][]byte{vNextPublicationCurrentFile: vNextPublicationExpectedJSON(t, desiredB), vNextPublicationJournalFile: nil})
				publisher.hooks = plan.hooks(vNextPublicationHooks{})
				if err := publisher.Prune(context.Background()); err != nil {
					t.Fatalf("Prune() with held A error = %v", err)
				}
			case "publish":
				currentB, err := publisher.Publish(vNextPublicationArtifactsForTest("current-B", false))
				if err != nil {
					t.Fatalf("Publish(B) error = %v", err)
				}
				if err := replaceLease(); err != nil {
					t.Fatal(err)
				}
				if currentB != desiredB {
					t.Fatal("held publish setup B differs from fixture")
				}
				plan, _ = vNextPublicationExpectedPublish(t, root, filesB, filesC, "complete")
				plan.expectPrunedGeneration(desiredB.Generation)
				publisher.hooks = plan.hooks(vNextPublicationHooks{})
				currentC, err := publisher.Publish(vNextPublicationArtifactsForTest("current-C", false))
				if err != nil {
					t.Fatalf("Publish(C) with held A error = %v", err)
				}
				if currentC == currentB {
					t.Fatal("Publish(C) did not intentionally advance CURRENT")
				}
				if currentC != desiredC {
					t.Fatal("held publish C differs from fixture")
				}
				expectedCurrent = desiredC
			case "recover":
				plan, _ = vNextPublicationExpectedPublish(t, root, filesA, filesB, "prepared")
				plan.steps = append(plan.steps, vNextPublicationExpectedTransition{target: vNextPublicationJournalFile, intended: nil, phases: 4})

				crash := errors.New("crash after CURRENT selection before committed JOURNAL replacement and lease substitution")
				writer, err := newVNextGenerationPublisher(root, "acme", plan.hooks(vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
					if point != vNextPublicationAfterCommitSync {
						return nil
					}
					if err := replaceLease(); err != nil {
						return err
					}
					return crash
				}}))
				if err != nil {
					t.Fatal(err)
				}
				if _, err := writer.Publish(vNextPublicationArtifactsForTest("current-B", false)); !errors.Is(err, crash) {
					t.Fatalf("Publish(B) error = %v, want injected crash", err)
				}
				var preparedJournal vNextGenerationJournal
				var foundPreparedJournal bool
				var observedCurrent vNextGenerationPointer
				observedCurrent, preparedJournal, foundPreparedJournal, _, _ = vNextPublicationReadCurrentJournalForTest(t, root)
				expectedCurrent = desiredB
				if observedCurrent != desiredB || !foundPreparedJournal || preparedJournal.State != "prepared" || preparedJournal.Old == nil || *preparedJournal.Old != old || preparedJournal.New != expectedCurrent {
					t.Fatalf("AfterCommitSync interruption durable state = current %#v journal %#v, want prepared old/new-selected", expectedCurrent, preparedJournal)
				}
				recovered, err := newVNextGenerationPublisher(root, "acme", plan.hooks(vNextPublicationHooks{}))
				if err != nil {
					t.Fatal(err)
				}
				if err := recovered.Recover(context.Background()); err != nil {
					t.Fatalf("Recover() with held A error = %v", err)
				}
			}

			if err := plan.check(); err != nil {
				t.Fatal(err)
			}
			vNextPublicationAssertExpectedTree(t, filepath.Dir(leasePath), heldExpected)
			actualAuthority, authorityErr := vNextPublicationExpectedStableAuthority(root)
			if authorityErr != nil {
				t.Fatal(authorityErr)
			}
			if err := vNextPublicationCompareExpectedMembers(actualAuthority, priorAuthority, false); err != nil {
				t.Fatal(err)
			}
			actualControls, controlErr := vNextPublicationExpectedPublicControls(root)
			if controlErr != nil {
				t.Fatal(controlErr)
			}
			if err := vNextPublicationCompareExpectedMembers(actualControls, plan.controls, true); err != nil {
				t.Fatalf("held cleanup controls after explicit transition cut: %v", err)
			}
			if !publisher.GenerationExists(old.Generation) {
				t.Fatal("cleanup removed the held generation after its lease pathname was replaced")
			}
			assertVNextPublicationMarker(t, held, "held-A")
			leaseInfo, err := os.Lstat(leasePath)
			if err != nil {
				t.Fatalf("replacement lease after cleanup = %v", err)
			}
			if leaseInfo.Size() != 0 || !leaseInfo.Mode().IsRegular() {
				t.Fatalf("replacement lease after cleanup = mode %v size %d, want empty regular file", leaseInfo.Mode(), leaseInfo.Size())
			}
			vNextPublicationAssertWitnessForTest(t, "held cleanup displaced lease A", retainedLease, displacedLeaseA)
			vNextPublicationAssertWitnessForTest(t, "held cleanup replacement lease B", leasePath, replacementLeaseB)
			vNextPublicationAssertCurrentJournalForTest(t, root, "held cleanup selected state", expectedCurrent, "", nil, vNextGenerationPointer{})

			held.Release()
			restoreLease()
			if err := publisher.Prune(context.Background()); err != nil {
				t.Fatalf("Prune() after release/restoration error = %v", err)
			}
			if publisher.GenerationExists(old.Generation) {
				t.Fatal("Prune() retained stale generation after reader release")
			}
		})
	}
}

const vNextPublicationDescriptorLimitScenarioEnv = "PM_CONNECTORGEN_DESCRIPTOR_LIMIT_SCENARIO"

func TestVNextGenerationPublisherBoundsAuthorityScanDescriptors(t *testing.T) {
	if scenario := os.Getenv(vNextPublicationDescriptorLimitScenarioEnv); scenario != "" {
		vNextPublicationDescriptorLimitSubprocess(t, scenario)
		return
	}
	for _, scenario := range []string{"check", "open", "recover", "prune", "publish"} {
		t.Run(scenario, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestVNextGenerationPublisherBoundsAuthorityScanDescriptors$", "-test.v")
			command.Env = append(os.Environ(), vNextPublicationDescriptorLimitScenarioEnv+"="+scenario)
			output, err := command.CombinedOutput()
			if ctx.Err() != nil {
				t.Fatalf("descriptor-limit %s child timed out: %s", scenario, output)
			}
			if err != nil {
				t.Fatalf("descriptor-limit %s child failed: %v; output=%s", scenario, err, output)
			}
		})
	}
}

func vNextPublicationDescriptorLimitSubprocess(t *testing.T, scenario string) {
	t.Helper()
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	var current vNextPublicationArtifacts
	for index := 0; index < 24; index++ {
		current = vNextPublicationArtifactsForTest(fmt.Sprintf("retained-%02d", index), false)
		if _, err := publisher.Publish(current); err != nil {
			t.Fatalf("Publish(retained history %d) error = %v", index, err)
		}
	}
	var limit unix.Rlimit
	if err := unix.Getrlimit(unix.RLIMIT_NOFILE, &limit); err != nil {
		t.Fatalf("get descriptor limit: %v", err)
	}
	const target = uint64(96)
	if limit.Max < target {
		t.Fatalf("descriptor maximum = %d, need at least %d for bounded-history witness", limit.Max, target)
	}
	if err := unix.Setrlimit(unix.RLIMIT_NOFILE, &unix.Rlimit{Cur: target, Max: limit.Max}); err != nil {
		t.Fatalf("set descriptor limit: %v", err)
	}
	var actual unix.Rlimit
	if err := unix.Getrlimit(unix.RLIMIT_NOFILE, &actual); err != nil {
		t.Fatalf("read effective descriptor limit: %v", err)
	}
	if actual.Cur != target {
		t.Fatalf("effective descriptor limit = %d, want %d", actual.Cur, target)
	}
	t.Logf("descriptor-limit scenario %s uses RLIMIT_NOFILE=%d after history construction; the child retains runtime headroom", scenario, actual.Cur)

	switch scenario {
	case "check":
		if err := publisher.Check(current); err != nil {
			t.Fatalf("Check() under bounded descriptor limit error = %v", err)
		}
	case "open":
		handle, err := publisher.Open(context.Background())
		if err != nil {
			t.Fatalf("Open() under bounded descriptor limit error = %v", err)
		}
		defer handle.Release()
		assertVNextPublicationMarker(t, handle, "retained-23")
	case "recover":
		if err := publisher.Recover(context.Background()); err != nil {
			t.Fatalf("Recover() under bounded descriptor limit error = %v", err)
		}
	case "prune":
		if err := publisher.Prune(context.Background()); err != nil {
			t.Fatalf("Prune() under bounded descriptor limit error = %v", err)
		}
	case "publish":
		if _, err := publisher.Publish(vNextPublicationArtifactsForTest("retained-next", false)); err != nil {
			t.Fatalf("Publish(next) under bounded descriptor limit error = %v", err)
		}
	default:
		t.Fatalf("unknown descriptor-limit scenario %q", scenario)
	}
}

func vNextPublicationCopyTreeForTest(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return os.Mkdir(destination, 0o755)
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.Mkdir(target, 0o755)
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("copy fixture member %q is not regular", relative)
		}
		payload, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, payload, 0o600)
	})
}

func TestConnectorgenMainPreservesNonConsumingSignalTermination(t *testing.T) {
	binary := vNextConnectorgenBinaryForTest(t)
	for _, test := range []struct {
		name       string
		signals    []os.Signal
		wantSignal syscall.Signal
	}{
		{name: "interrupt", signals: []os.Signal{os.Interrupt}, wantSignal: syscall.SIGINT},
		{name: "terminate", signals: []os.Signal{syscall.SIGTERM}, wantSignal: syscall.SIGTERM},
		{name: "repeated-interrupt", signals: []os.Signal{os.Interrupt, os.Interrupt}, wantSignal: syscall.SIGINT},
	} {
		t.Run(test.name, func(t *testing.T) {
			workingDirectory, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			root := t.TempDir()
			bundle := filepath.Join(root, "github")
			source := filepath.Join(workingDirectory, "..", "..", "internal", "connectors", "defs", "github")
			if err := vNextPublicationCopyTreeForTest(source, bundle); err != nil {
				t.Fatalf("copy validate bundle fixture: %v", err)
			}
			fifo := filepath.Join(bundle, "spec.json")
			if err := os.Remove(fifo); err != nil {
				t.Fatal(err)
			}
			if err := unix.Mkfifo(fifo, 0o600); err != nil {
				t.Fatal(err)
			}
			command := exec.Command(binary, "validate", root, "--connector", "github")
			var stdout, stderr bytes.Buffer
			command.Stdout = &stdout
			command.Stderr = &stderr
			child := vNextPublicationStartBoundedChildForTest(t, command, "validate subprocess")
			writer := vNextPublicationOpenFIFOWriterForTest(t, fifo)
			t.Cleanup(func() {
				if err := writer.Close(); err != nil {
					t.Errorf("close validate FIFO writer: %v", err)
				}
			})
			for _, signal := range test.signals {
				if err := command.Process.Signal(signal); err != nil && !errors.Is(err, os.ErrProcessDone) {
					t.Fatalf("signal validate subprocess with %v: %v", signal, err)
				}
				time.Sleep(25 * time.Millisecond)
			}
			err, completed := child.waitWithin(1500 * time.Millisecond)
			if !completed {
				child.killAndWait(t, "validate subprocess timeout")
				t.Fatalf("validate subprocess did not terminate after %v; global signal interception kept a non-consuming command alive; stdout=%q stderr=%q", test.signals, stdout.String(), stderr.String())
			}
			if err == nil {
				t.Fatalf("validate subprocess exited successfully after %v; stdout=%q stderr=%q", test.signals, stdout.String(), stderr.String())
			}
			exit, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatalf("validate subprocess error = %T %v, want signal termination", err, err)
			}
			status, ok := exit.Sys().(syscall.WaitStatus)
			if !ok || !status.Signaled() || status.Signal() != test.wantSignal {
				t.Fatalf("validate subprocess status = %#v, want exact signal %v; stdout=%q stderr=%q", exit.Sys(), test.wantSignal, stdout.String(), stderr.String())
			}
		})
	}
}

func TestConnectorgenMainLockRenderSignalCancellationAndRetry(t *testing.T) {
	binary := vNextConnectorgenBinaryForTest(t)
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
	initial := exec.Command(binary, "lock-render", lock.Connector, "--defs", root)
	var initialStdout, initialStderr bytes.Buffer
	initial.Stdout = &initialStdout
	initial.Stderr = &initialStderr
	initialChild := vNextPublicationStartBoundedChildForTest(t, initial, "initial lock-render subprocess")
	if err, completed := initialChild.waitWithin(5 * time.Second); !completed {
		initialChild.killAndWait(t, "initial lock-render timeout")
		t.Fatalf("initial lock-render did not finish; stdout=%q stderr=%q", initialStdout.String(), initialStderr.String())
	} else if err != nil {
		t.Fatalf("initial lock-render error = %v; stdout=%q stderr=%q", err, initialStdout.String(), initialStderr.String())
	}
	currentPath := filepath.Join(connectorRoot, vNextPublicationCurrentFile)
	before, err := os.ReadFile(currentPath)
	if err != nil {
		t.Fatal(err)
	}
	stateBefore := vNextPublicationTreeSnapshotForTest(t, connectorRoot)
	held := vNextPublicationHoldLockForTest(t, root)
	t.Cleanup(func() {
		if held != nil {
			unlockVNextPublicationFile(held)
		}
	})
	command := exec.Command(binary, "lock-render", lock.Connector, "--defs", root, "--check")
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	child := vNextPublicationStartBoundedChildForTest(t, command, "contended lock-render subprocess")
	vNextPublicationWaitForProcessOpenPathForTest(t, command.Process.Pid, connectorRoot, 2*time.Second)
	if err := command.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("signal contended lock-render subprocess: %v", err)
	}
	err, completed := child.waitWithin(1500 * time.Millisecond)
	if !completed {
		child.killAndWait(t, "contended lock-render timeout")
		t.Fatalf("contended lock-render did not exit after interruption; stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
	unlockVNextPublicationFile(held)
	held = nil
	if err == nil {
		t.Fatalf("contended lock-render exited successfully after signal; stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 1 {
		t.Fatalf("contended lock-render error = %T %v, want cancellation exit 1; stdout=%q stderr=%q", err, err, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), context.Canceled.Error()) {
		t.Fatalf("contended lock-render output after cancellation: stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
	if after, err := os.ReadFile(currentPath); err != nil || !bytes.Equal(before, after) {
		t.Fatalf("contended lock-render changed CURRENT: err=%v before=%q after=%q", err, before, after)
	}
	if _, err := os.Lstat(filepath.Join(connectorRoot, vNextPublicationJournalFile)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("contended lock-render changed JOURNAL: %v", err)
	}
	if stateAfter := vNextPublicationTreeSnapshotForTest(t, connectorRoot); !bytes.Equal(stateBefore, stateAfter) {
		t.Fatalf("contended lock-render changed selected/control/authority/generation state: before=%s after=%s", stateBefore, stateAfter)
	}
	retry := exec.Command(binary, "lock-render", lock.Connector, "--defs", root, "--check")
	var retryStdout, retryStderr bytes.Buffer
	retry.Stdout = &retryStdout
	retry.Stderr = &retryStderr
	retryChild := vNextPublicationStartBoundedChildForTest(t, retry, "lock-render retry subprocess")
	if err, completed := retryChild.waitWithin(5 * time.Second); !completed {
		retryChild.killAndWait(t, "lock-render retry timeout")
		t.Fatalf("lock-render retry did not finish; stdout=%q stderr=%q", retryStdout.String(), retryStderr.String())
	} else if err != nil {
		t.Fatalf("lock-render retry error = %v; stdout=%q stderr=%q", err, retryStdout.String(), retryStderr.String())
	}
}

func vNextConnectorgenBinaryForTest(t *testing.T) string {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "connectorgen")
	command := exec.Command("go", "build", "-o", binary, ".")
	command.Dir = workingDirectory
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build connectorgen test binary: %v; output=%s", err, output)
	}
	return binary
}

type vNextPublicationBoundedChild struct {
	command *exec.Cmd
	wait    chan error
	reaped  bool
}

func vNextPublicationStartBoundedChildForTest(t *testing.T, command *exec.Cmd, label string) *vNextPublicationBoundedChild {
	t.Helper()
	if err := command.Start(); err != nil {
		t.Fatalf("start %s: %v", label, err)
	}
	child := &vNextPublicationBoundedChild{command: command, wait: make(chan error, 1)}
	go func() { child.wait <- command.Wait() }()
	t.Cleanup(func() {
		if !child.reaped {
			child.killAndWait(t, label+" cleanup")
		}
	})
	return child
}

func (child *vNextPublicationBoundedChild) waitWithin(timeout time.Duration) (error, bool) {
	if child.reaped {
		return nil, true
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case err := <-child.wait:
		child.reaped = true
		return err, true
	case <-timer.C:
		return nil, false
	}
}

func (child *vNextPublicationBoundedChild) killAndWait(t *testing.T, label string) {
	t.Helper()
	if child.reaped {
		return
	}
	if child.command.Process != nil {
		if err := child.command.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			t.Errorf("kill %s: %v", label, err)
		}
	}
	if _, completed := child.waitWithin(time.Second); !completed {
		t.Errorf("bounded cleanup did not reap %s", label)
	} else {
		t.Logf("bounded cleanup observed direct Wait completion for %s", label)
	}
}

func vNextPublicationWaitForProcessOpenPathForTest(t *testing.T, pid int, target string, timeout time.Duration) {
	t.Helper()
	if pid <= 0 {
		t.Fatal("invalid child PID for lock readiness")
	}
	targetInfo, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	lsof, err := exec.LookPath("lsof")
	if err != nil {
		t.Fatalf("locate lsof for bounded lock readiness: %v", err)
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		command := exec.CommandContext(ctx, lsof, "-Fn", "-p", strconv.Itoa(pid))
		output, runErr := command.Output()
		timedOut := errors.Is(ctx.Err(), context.DeadlineExceeded)
		cancel()
		if runErr == nil {
			for _, line := range strings.Split(string(output), "\n") {
				if !strings.HasPrefix(line, "n") {
					continue
				}
				name := strings.TrimSpace(strings.TrimPrefix(line, "n"))
				info, statErr := os.Stat(name)
				if statErr == nil && os.SameFile(targetInfo, info) {
					t.Logf("observed contended lock-render child PID %d holding the connector-directory descriptor", pid)
					return
				}
			}
		} else if timedOut {
			t.Fatalf("bounded lock readiness observation timed out: %v", runErr)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("contended lock-render child PID %d never opened the held connector directory", pid)
}

func vNextPublicationOpenFIFOWriterForTest(t *testing.T, path string) *os.File {
	t.Helper()
	writer, err := vNextPublicationWaitForFIFOWriterForTest(path, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	return writer
}

func vNextPublicationWaitForFIFOWriterForTest(path string, timeout time.Duration) (*os.File, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		fd, err := unix.Open(path, unix.O_WRONLY|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
		if err == nil {
			writer := os.NewFile(uintptr(fd), "validate metadata FIFO writer")
			if writer == nil {
				_ = unix.Close(fd)
				return nil, fmt.Errorf("construct validate metadata FIFO writer")
			}
			return writer, nil
		}
		if !errors.Is(err, unix.ENXIO) {
			return nil, fmt.Errorf("open validate metadata FIFO writer: %w", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil, fmt.Errorf("validate subprocess never reached deterministic FIFO-read readiness")
}

func TestVNextPublicationBoundedChildReapsWithheldFIFOReadiness(t *testing.T) {
	fifo := filepath.Join(t.TempDir(), "withheld-readiness")
	if err := unix.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("/bin/sh", "-c", "trap '' TERM; while :; do :; done")
	child := vNextPublicationStartBoundedChildForTest(t, command, "nonterminating withheld-readiness child")
	if writer, err := vNextPublicationWaitForFIFOWriterForTest(fifo, 100*time.Millisecond); err == nil {
		if closeErr := writer.Close(); closeErr != nil {
			t.Errorf("close unexpected FIFO writer: %v", closeErr)
		}
		t.Fatal("withheld FIFO readiness unexpectedly completed")
	}
	child.killAndWait(t, "nonterminating withheld-readiness child")
	if !child.reaped {
		t.Fatal("withheld readiness child was not reaped")
	}
	t.Log("withheld FIFO readiness returned within its bound and the owned nonterminating child reached direct Wait completion")
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

	expectedCut, err := vNextPublicationObserveExpectedTree(filepath.Join(root, "acme"))
	if err != nil {
		t.Fatal(err)
	}

	if err := publisher.Check(newSet); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	vNextPublicationAssertDurableCutWitnessForTest(t, root, "non-destructive Check", expectedCut, oldPointer)

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

	expected, err := vNextPublicationObserveExpectedTree(filepath.Join(root, "acme"))
	if err != nil {
		t.Fatal(err)
	}
	if err := publisher.Check(artifacts); err == nil {
		t.Fatal("Check() accepted a pending journal")
	}
	if got, err := os.ReadFile(journalPath); err != nil || !bytes.Equal(got, journal) {
		t.Fatalf("Check() recovered or rewrote journal: err=%v got=%q", err, got)
	}
	vNextPublicationAssertExpectedTree(t, filepath.Join(root, "acme"), expected)
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
	t.Cleanup(func() {
		if err := connector.Close(); err != nil {
			t.Errorf("close connector publication root: %v", err)
		}
	})
	lock, identity, err := vNextPublicationOpenLock(connector)
	if err != nil {
		t.Fatalf("vNextPublicationOpenLock() error = %v", err)
	}
	t.Cleanup(func() {
		if err := lock.Close(); err != nil {
			t.Errorf("close connector publication lock: %v", err)
		}
	})
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
			if test.name == "JOURNAL" {
				if err != nil {
					t.Fatalf("Recover() after divergent JOURNAL inode = %v", err)
				}
				if _, statErr := os.Stat(filepath.Join(root, "acme", vNextPublicationJournalFile)); !errors.Is(statErr, fs.ErrNotExist) {
					t.Fatalf("JOURNAL after authority recovery = %v, want absent", statErr)
				}
				return
			}
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
	wantStageA := vNextPublicationExpectedOwnedStage(t, stagePath, pointer, stageName, "original")

	var movedPath string
	hookHit := false
	desiredPointer := vNextPublicationFixturePointer(t, vNextPublicationArtifactsForTest("active", false))
	if pointer != desiredPointer {
		t.Fatal("stage setup differs from declared fixture")
	}
	plan := vNextPublicationNewExpectedPlan(t, root, map[string][]byte{vNextPublicationCurrentFile: vNextPublicationExpectedJSON(t, desiredPointer), vNextPublicationJournalFile: nil})
	priorAuthority, priorErr := vNextPublicationExpectedStableAuthority(root)
	if priorErr != nil {
		t.Fatal(priorErr)
	}
	var expectedCut vNextPublicationExpectedTree
	guard, err := newVNextGenerationPublisher(root, "acme", plan.hooks(vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationBeforeStageRemoval || hookHit {
			return nil
		}
		hookHit = true
		movedPath = vNextReplaceOwnedStageForTest(t, stagePath, pointer, stageName, "replacement")
		plan.stageDelta(t, stagePath, movedPath, vNextPublicationExpectedOwnedStage(t, stagePath, desiredPointer, stageName, "replacement"))

		var observeErr error
		expectedCut, observeErr = vNextPublicationCaptureExpectedCut(root, priorAuthority, plan)
		if observeErr != nil {
			return observeErr
		}
		return nil
	}}))
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
	vNextPublicationAssertCurrentJournalForTest(t, root, "stage-cleanup identity refusal", pointer, "", nil, vNextGenerationPointer{})
	if err := vNextPublicationEarlyStageVerdict(root, movedPath, expectedCut, wantStageA); err != nil {
		t.Fatal(err)
	}

	fresh, freshErr := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if freshErr != nil {
		t.Fatal(freshErr)
	}
	checkErr := fresh.Check(vNextPublicationArtifactsForTest("active", false))
	if checkErr == nil {
		t.Fatal("Check accepted retained stale/replaced stage")
	}
	if err := vNextPublicationEarlyStageVerdict(root, movedPath, expectedCut, wantStageA); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(stagePath); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(movedPath, stagePath); err != nil {
		t.Fatal(err)
	}
	if err := fresh.Recover(context.Background()); err != nil {
		t.Fatalf("fresh recovery after fixture stage restoration: %v", err)
	}
	retry := vNextPublicationArtifactsForTest("stage-retry", false)
	if _, err := fresh.Publish(retry); err != nil {
		t.Fatal(err)
	}
	if err := fresh.Check(retry); err != nil {
		t.Fatal(err)
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
	desiredPointer := vNextPublicationFixturePointer(t, vNextPublicationArtifactsForTest("active", false))
	if pointer != desiredPointer {
		t.Fatal("stage setup differs from declared fixture")
	}
	plan := vNextPublicationNewExpectedPlan(t, root, map[string][]byte{vNextPublicationCurrentFile: vNextPublicationExpectedJSON(t, desiredPointer), vNextPublicationJournalFile: nil})
	priorAuthority, priorErr := vNextPublicationExpectedStableAuthority(root)
	if priorErr != nil {
		t.Fatal(priorErr)
	}
	var expectedCut vNextPublicationExpectedTree
	guard, err := newVNextGenerationPublisher(root, "acme", plan.hooks(vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationFaultPoint("after_stage_removal_identity") || hookHit {
			return nil
		}
		hookHit = true
		movedPath = vNextReplaceDirectoryAtPathForTest(t, stagePath)
		replacement, observeErr := vNextPublicationObserveExpectedTree(stagePath)
		if observeErr != nil {
			return observeErr
		}
		if len(replacement) != 1 || replacement["."].identity.mode != unix.S_IFDIR {
			return fmt.Errorf("late stage B is not empty directory")
		}
		plan.stageDelta(t, stagePath, movedPath, replacement)

		expectedCut, observeErr = vNextPublicationCaptureExpectedCut(root, priorAuthority, plan)
		if observeErr != nil {
			return observeErr
		}
		return nil
	}}))
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
	vNextPublicationAssertExpectedTree(t, filepath.Join(root, "acme"), expectedCut)
	fresh, freshErr := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if freshErr != nil {
		t.Fatal(freshErr)
	}
	checkErr := fresh.Check(vNextPublicationArtifactsForTest("active", false))
	if checkErr == nil {
		t.Fatal("Check accepted retained stale/replaced stage")
	}
	vNextPublicationAssertExpectedTree(t, filepath.Join(root, "acme"), expectedCut)
	if err := os.RemoveAll(stagePath); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(movedPath, stagePath); err != nil {
		t.Fatal(err)
	}
	if err := fresh.Recover(context.Background()); err != nil {
		t.Fatalf("fresh recovery after fixture stage restoration: %v", err)
	}
	retry := vNextPublicationArtifactsForTest("stage-retry", false)
	if _, err := fresh.Publish(retry); err != nil {
		t.Fatal(err)
	}
	if err := fresh.Check(retry); err != nil {
		t.Fatal(err)
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
	for _, test := range []struct {
		name        string
		replacement []byte
	}{
		{name: "nonempty replacement", replacement: []byte("late lease replacement")},
		{name: "empty replacement", replacement: nil},
	} {
		t.Run(test.name, func(t *testing.T) {
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
			leasePath := filepath.Join(generationPath, vNextPublicationLeaseFile)
			movedPath := leasePath + ".late-original"
			leaseInfo, err := os.Lstat(leasePath)
			if err != nil {
				t.Fatal(err)
			}
			leaseBytes, err := os.ReadFile(leasePath)
			if err != nil {
				t.Fatal(err)
			}
			if !leaseInfo.Mode().IsRegular() || len(leaseBytes) != 0 {
				t.Fatalf("original lease = mode %v bytes %q, want empty regular file", leaseInfo.Mode(), leaseBytes)
			}
			generationInfo, err := os.Stat(generationPath)
			if err != nil {
				t.Fatal(err)
			}
			currentPath := filepath.Join(root, "acme", vNextPublicationCurrentFile)
			currentInfo, err := os.Stat(currentPath)
			if err != nil {
				t.Fatal(err)
			}
			currentBytes, err := os.ReadFile(currentPath)
			if err != nil {
				t.Fatal(err)
			}
			activePath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, current.Generation)
			activeInfo, err := os.Stat(activePath)
			if err != nil {
				t.Fatal(err)
			}
			activeMetadata, err := os.ReadFile(filepath.Join(activePath, "metadata.json"))
			if err != nil {
				t.Fatal(err)
			}

			hookHit := false
			var replacementInfo os.FileInfo
			guard, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
				if point != vNextPublicationAfterGenerationLeaseIdentity || hookHit {
					return nil
				}
				hookHit = true
				if err := os.Rename(leasePath, movedPath); err != nil {
					return err
				}
				if err := os.WriteFile(leasePath, test.replacement, 0o600); err != nil {
					return err
				}
				replacementInfo, err = os.Lstat(leasePath)
				if err != nil {
					return err
				}
				if !replacementInfo.Mode().IsRegular() || os.SameFile(leaseInfo, replacementInfo) {
					return fmt.Errorf("late replacement does not have a distinct regular identity")
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
				_, movedErr := os.Lstat(movedPath)
				_, replacementErr := os.Lstat(leasePath)
				_, generationErr := os.Stat(generationPath)
				t.Fatalf("Prune() after late %s lease replacement error = %v, want identity refusal; moved=%v replacement=%v generation=%v", test.name, err, movedErr, replacementErr, generationErr)
			}
			if info, statErr := os.Lstat(movedPath); statErr != nil || !os.SameFile(info, leaseInfo) {
				t.Fatalf("late lease replacement did not preserve original A: err=%v info=%v", statErr, info)
			}
			if got, readErr := os.ReadFile(movedPath); readErr != nil || !bytes.Equal(got, leaseBytes) {
				t.Fatalf("late lease replacement changed original A bytes: err=%v got=%q want=%q", readErr, got, leaseBytes)
			}
			if info, statErr := os.Lstat(leasePath); statErr != nil || !os.SameFile(info, replacementInfo) {
				t.Fatalf("late lease replacement did not preserve replacement B: err=%v info=%v", statErr, info)
			}
			if got, readErr := os.ReadFile(leasePath); readErr != nil || !bytes.Equal(got, test.replacement) {
				t.Fatalf("late lease replacement changed replacement B bytes: err=%v got=%q want=%q", readErr, got, test.replacement)
			}
			if info, statErr := os.Stat(generationPath); statErr != nil || !os.SameFile(info, generationInfo) {
				t.Fatalf("late lease replacement changed generation root: err=%v info=%v", statErr, info)
			}
			if got, readErr := os.ReadFile(currentPath); readErr != nil || !bytes.Equal(got, currentBytes) {
				t.Fatalf("late lease replacement changed CURRENT bytes: err=%v got=%q want=%q", readErr, got, currentBytes)
			}
			if info, statErr := os.Stat(currentPath); statErr != nil || !os.SameFile(info, currentInfo) {
				t.Fatalf("late lease replacement changed CURRENT identity: err=%v info=%v", statErr, info)
			}
			if info, statErr := os.Stat(activePath); statErr != nil || !os.SameFile(info, activeInfo) {
				t.Fatalf("late lease replacement changed active generation: err=%v info=%v", statErr, info)
			}
			if got, readErr := os.ReadFile(filepath.Join(activePath, "metadata.json")); readErr != nil || !bytes.Equal(got, activeMetadata) {
				t.Fatalf("late lease replacement changed active metadata: err=%v got=%q want=%q", readErr, got, activeMetadata)
			}
		})
	}
}

func TestVNextGenerationPublisherRefusesLateLeaseReplacementAcrossPublicCleanupCallers(t *testing.T) {
	for _, test := range []struct {
		name   string
		invoke func(*vNextGenerationPublisher) error
	}{
		{name: "prune", invoke: func(publisher *vNextGenerationPublisher) error {
			return publisher.Prune(context.Background())
		}},
		{name: "no journal recovery", invoke: func(publisher *vNextGenerationPublisher) error {
			return publisher.Recover(context.Background())
		}},
		{name: "publish initial recovery", invoke: func(publisher *vNextGenerationPublisher) error {
			_, err := publisher.Publish(vNextPublicationArtifactsForTest("next", false))
			return err
		}},
		{name: "open transitive recovery", invoke: func(publisher *vNextGenerationPublisher) error {
			handle, err := publisher.Open(context.Background())
			if handle != nil {
				handle.Release()
			}
			return err
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
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

			leasePath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, old.Generation, vNextPublicationLeaseFile)
			leaseA := vNextPublicationFileWitnessForTest(t, leasePath)
			var displacedLeaseA, replacementLeaseB vNextPublicationPathWitness
			hookHit := false
			desiredCurrent := vNextPublicationFixturePointer(t, vNextPublicationArtifactsForTest("current", false))
			if current != desiredCurrent {
				t.Fatal("static cleanup setup differs from fixture")
			}
			plan := vNextPublicationNewExpectedPlan(t, root, map[string][]byte{vNextPublicationCurrentFile: vNextPublicationExpectedJSON(t, desiredCurrent), vNextPublicationJournalFile: nil})
			priorAuthority, priorErr := vNextPublicationExpectedStableAuthority(root)
			if priorErr != nil {
				t.Fatal(priorErr)
			}
			var expectedCut vNextPublicationExpectedTree
			guard, err := newVNextGenerationPublisher(root, "acme", plan.hooks(vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
				if point != vNextPublicationAfterGenerationLeaseIdentity || hookHit {
					return nil
				}
				hookHit = true
				if err := os.Rename(leasePath, leasePath+".late-original"); err != nil {
					return err
				}
				displacedLeaseA = vNextPublicationFileWitnessForTest(t, leasePath+".late-original")
				if !os.SameFile(leaseA.info, displacedLeaseA.info) || !bytes.Equal(leaseA.payload, displacedLeaseA.payload) {
					return fmt.Errorf("late public caller displaced A changed")
				}
				if err := os.WriteFile(leasePath, []byte("late public caller replacement"), 0o600); err != nil {
					return err
				}
				replacementLeaseB = vNextPublicationFileWitnessForTest(t, leasePath)
				plan.leaseDelta(t, old.Generation, leasePath, leasePath+".late-original")
				vNextPublicationAssertDistinctWitnessForTest(t, "late public caller A/B", displacedLeaseA, replacementLeaseB)
				var observeErr error
				expectedCut, observeErr = vNextPublicationCaptureExpectedCut(root, priorAuthority, plan)
				if observeErr != nil {
					return observeErr
				}
				return nil
			}}))
			if err != nil {
				t.Fatal(err)
			}
			err = test.invoke(guard)
			if !hookHit {
				t.Fatal("public cleanup caller did not reach late generation-lease identity barrier")
			}
			if err == nil || !strings.Contains(err.Error(), "identity") {
				t.Fatalf("public %s after late lease replacement error = %v, want identity refusal", test.name, err)
			}
			vNextPublicationAssertWitnessForTest(t, "late public caller displaced A", leasePath+".late-original", displacedLeaseA)
			vNextPublicationAssertWitnessForTest(t, "late public caller replacement B", leasePath, replacementLeaseB)
			vNextPublicationAssertCurrentJournalForTest(t, root, "late public caller refusal", current, "", nil, vNextGenerationPointer{})
			vNextPublicationAssertDurableCutWitnessForTest(t, root, "late public caller refusal", expectedCut, old)
		})
	}

	t.Run("prepared journal/new-selected recovery", func(t *testing.T) {
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
		newFiles := vNextPublicationArtifactsForTest("current", false)
		plan, wantedNew := vNextPublicationExpectedPublish(t, root, vNextPublicationArtifactsForTest("old", false), newFiles, "prepared")
		crash := errors.New("prepared JOURNAL/new-selected before committed replacement")
		writer, err := newVNextGenerationPublisher(root, "acme", plan.hooks(vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
			if point == vNextPublicationAfterCommitSync {
				return crash
			}
			return nil
		}}))
		if err != nil {
			handle.Release()
			t.Fatal(err)
		}
		if _, err := writer.Publish(newFiles); !errors.Is(err, crash) {
			handle.Release()
			t.Fatalf("Publish(current) error = %v, want prepared-JOURNAL interruption", err)
		}
		handle.Release()
		selectedNew, preparedJournal, foundPreparedJournal, _, _ := vNextPublicationReadCurrentJournalForTest(t, root)
		if selectedNew != wantedNew || !foundPreparedJournal || preparedJournal.State != "prepared" || preparedJournal.Old == nil || *preparedJournal.Old != old || preparedJournal.New != selectedNew {
			t.Fatalf("AfterCommitSync durable state = current %#v journal %#v, want prepared old/new-selected", selectedNew, preparedJournal)
		}

		leasePath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, old.Generation, vNextPublicationLeaseFile)
		leaseA := vNextPublicationFileWitnessForTest(t, leasePath)
		var displacedLeaseA, replacementLeaseB vNextPublicationPathWitness
		hookHit := false
		priorAuthority, priorErr := vNextPublicationExpectedStableAuthority(root)
		if priorErr != nil {
			t.Fatal(priorErr)
		}
		var expectedCut vNextPublicationExpectedTree
		guard, err := newVNextGenerationPublisher(root, "acme", plan.hooks(vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
			if point != vNextPublicationAfterGenerationLeaseIdentity || hookHit {
				return nil
			}
			hookHit = true
			if err := os.Rename(leasePath, leasePath+".late-original"); err != nil {
				return err
			}
			displacedLeaseA = vNextPublicationFileWitnessForTest(t, leasePath+".late-original")
			if !os.SameFile(leaseA.info, displacedLeaseA.info) || !bytes.Equal(leaseA.payload, displacedLeaseA.payload) {
				return fmt.Errorf("late prepared-journal displaced A changed")
			}
			if err := os.WriteFile(leasePath, []byte("late prepared replacement"), 0o600); err != nil {
				return err
			}
			replacementLeaseB = vNextPublicationFileWitnessForTest(t, leasePath)
			plan.leaseDelta(t, old.Generation, leasePath, leasePath+".late-original")
			vNextPublicationAssertDistinctWitnessForTest(t, "late prepared-journal A/B", displacedLeaseA, replacementLeaseB)
			var observeErr error
			expectedCut, observeErr = vNextPublicationCaptureExpectedCut(root, priorAuthority, plan)
			if observeErr != nil {
				return observeErr
			}
			return nil
		}}))
		if err != nil {
			t.Fatal(err)
		}
		err = guard.Recover(context.Background())
		if !hookHit {
			t.Fatal("prepared-journal recovery did not reach late generation-lease identity barrier")
		}
		if err == nil || !strings.Contains(err.Error(), "identity") {
			t.Fatalf("Recover() after prepared journal late lease replacement error = %v, want identity refusal", err)
		}
		vNextPublicationAssertWitnessForTest(t, "prepared-journal recovery displaced A", leasePath+".late-original", displacedLeaseA)
		vNextPublicationAssertWitnessForTest(t, "prepared-journal recovery replacement B", leasePath, replacementLeaseB)
		vNextPublicationAssertCurrentJournalForTest(t, root, "prepared-journal recovery refusal", wantedNew, "prepared", &old, wantedNew)
		vNextPublicationAssertDurableCutWitnessForTest(t, root, "prepared-journal recovery refusal", expectedCut)
	})
}

func TestVNextGenerationPublisherLateLeaseReplacementRetainsPublicGenerationCollision(t *testing.T) {
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

	generationsPath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory)
	generationPath := filepath.Join(generationsPath, old.Generation)
	leasePath := filepath.Join(generationPath, vNextPublicationLeaseFile)
	movedLeaseName := vNextPublicationLeaseFile + ".late-original"
	movedLeasePath := filepath.Join(generationPath, movedLeaseName)
	leaseInfo, err := os.Lstat(leasePath)
	if err != nil {
		t.Fatal(err)
	}
	leaseBytes, err := os.ReadFile(leasePath)
	if err != nil {
		t.Fatal(err)
	}
	generationInfo, err := os.Stat(generationPath)
	if err != nil {
		t.Fatal(err)
	}
	currentPath := filepath.Join(root, "acme", vNextPublicationCurrentFile)
	currentInfo, err := os.Stat(currentPath)
	if err != nil {
		t.Fatal(err)
	}
	currentBytes, err := os.ReadFile(currentPath)
	if err != nil {
		t.Fatal(err)
	}
	activePath := filepath.Join(generationsPath, current.Generation)
	activeInfo, err := os.Stat(activePath)
	if err != nil {
		t.Fatal(err)
	}
	activeMetadata, err := os.ReadFile(filepath.Join(activePath, "metadata.json"))
	if err != nil {
		t.Fatal(err)
	}

	replacementBytes := []byte("late lease replacement")
	publicBytes := []byte(`{"marker":"public-C"}`)
	var replacementInfo, retainedInfo, publicInfo os.FileInfo
	var retainedPath string
	identityHit := false
	restoreHit := false
	guard, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		switch point {
		case vNextPublicationAfterGenerationLeaseIdentity:
			if identityHit {
				return nil
			}
			identityHit = true
			if err := os.Rename(leasePath, movedLeasePath); err != nil {
				return err
			}
			if err := os.WriteFile(leasePath, replacementBytes, 0o600); err != nil {
				return err
			}
			var err error
			replacementInfo, err = os.Lstat(leasePath)
			if err != nil {
				return err
			}
			if !replacementInfo.Mode().IsRegular() || os.SameFile(leaseInfo, replacementInfo) {
				return fmt.Errorf("late replacement does not have a distinct regular identity")
			}
			return nil
		case vNextPublicationBeforeQuarantineRestore:
			if !identityHit || restoreHit {
				return nil
			}
			restoreHit = true
			if _, err := os.Lstat(generationPath); !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("public generation before quarantine restore = %v, want absent", err)
			}
			entries, err := os.ReadDir(generationsPath)
			if err != nil {
				return err
			}
			for _, entry := range entries {
				if !entry.IsDir() || !strings.HasPrefix(entry.Name(), ".connectorgen-quarantine-") {
					continue
				}
				candidate := filepath.Join(generationsPath, entry.Name(), vNextPublicationQuarantineMember)
				info, err := os.Stat(candidate)
				if err == nil && os.SameFile(info, generationInfo) {
					retainedPath, retainedInfo = candidate, info
					break
				}
			}
			if retainedPath == "" {
				return fmt.Errorf("quarantine does not retain original generation before restoration")
			}
			if err := os.Mkdir(generationPath, 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(generationPath, "metadata.json"), publicBytes, 0o600); err != nil {
				return err
			}
			publicInfo, err = os.Stat(generationPath)
			return err
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

	if !identityHit || !restoreHit {
		t.Fatalf("late lease collision barriers reached identity=%t restore=%t, want both", identityHit, restoreHit)
	}
	if err == nil || !strings.Contains(err.Error(), "identity") || !errors.Is(err, fs.ErrExist) {
		t.Fatalf("Prune() after late lease/public C collision error = %v, want identity refusal retaining no-replace conflict", err)
	}
	if info, statErr := os.Stat(retainedPath); statErr != nil || !os.SameFile(info, retainedInfo) || !os.SameFile(info, generationInfo) {
		t.Fatalf("quarantined generation A/B identity changed: err=%v info=%v", statErr, info)
	}
	retainedMoved := filepath.Join(retainedPath, movedLeaseName)
	retainedLease := filepath.Join(retainedPath, vNextPublicationLeaseFile)
	if info, statErr := os.Lstat(retainedMoved); statErr != nil || !os.SameFile(info, leaseInfo) {
		t.Fatalf("quarantined original lease A changed: err=%v info=%v", statErr, info)
	}
	if got, readErr := os.ReadFile(retainedMoved); readErr != nil || !bytes.Equal(got, leaseBytes) {
		t.Fatalf("quarantined original lease A bytes changed: err=%v got=%q want=%q", readErr, got, leaseBytes)
	}
	if info, statErr := os.Lstat(retainedLease); statErr != nil || !os.SameFile(info, replacementInfo) {
		t.Fatalf("quarantined replacement lease B changed: err=%v info=%v", statErr, info)
	}
	if got, readErr := os.ReadFile(retainedLease); readErr != nil || !bytes.Equal(got, replacementBytes) {
		t.Fatalf("quarantined replacement lease B bytes changed: err=%v got=%q want=%q", readErr, got, replacementBytes)
	}
	if info, statErr := os.Stat(generationPath); statErr != nil || !os.SameFile(info, publicInfo) {
		t.Fatalf("public C identity changed: err=%v info=%v", statErr, info)
	}
	if got, readErr := os.ReadFile(filepath.Join(generationPath, "metadata.json")); readErr != nil || !bytes.Equal(got, publicBytes) {
		t.Fatalf("public C bytes changed: err=%v got=%q want=%q", readErr, got, publicBytes)
	}
	if info, statErr := os.Stat(currentPath); statErr != nil || !os.SameFile(info, currentInfo) {
		t.Fatalf("late lease collision changed CURRENT identity: err=%v info=%v", statErr, info)
	}
	if got, readErr := os.ReadFile(currentPath); readErr != nil || !bytes.Equal(got, currentBytes) {
		t.Fatalf("late lease collision changed CURRENT bytes: err=%v got=%q want=%q", readErr, got, currentBytes)
	}
	if info, statErr := os.Stat(activePath); statErr != nil || !os.SameFile(info, activeInfo) {
		t.Fatalf("late lease collision changed active generation: err=%v info=%v", statErr, info)
	}
	if got, readErr := os.ReadFile(filepath.Join(activePath, "metadata.json")); readErr != nil || !bytes.Equal(got, activeMetadata) {
		t.Fatalf("late lease collision changed active metadata: err=%v got=%q want=%q", readErr, got, activeMetadata)
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

func TestVNextGenerationPublisherRefusesLateReplacedRollbackGenerationLeaseCleanup(t *testing.T) {
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
	leasePath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, generation, vNextPublicationLeaseFile)
	hookHit := false
	guard, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationAfterGenerationLeaseIdentity || hookHit {
			return nil
		}
		hookHit = true
		if err := os.Rename(leasePath, leasePath+".late-original"); err != nil {
			return err
		}
		return os.WriteFile(leasePath, []byte("late rollback lease replacement"), 0o600)
	}})
	if err != nil {
		t.Fatal(err)
	}
	_, publishErr := guard.Publish(newSet)
	if !hookHit {
		t.Fatal("rollback generation cleanup did not reach late lease identity barrier")
	}
	if publishErr == nil || !strings.Contains(publishErr.Error(), activeFailure.Error()) || !strings.Contains(publishErr.Error(), "identity") {
		t.Fatalf("Publish(new) after late rollback lease replacement error = %v, want active-validation and identity refusal", publishErr)
	}
	if validationCalls != 2 {
		t.Fatalf("validation calls = %d, want staged then active validation", validationCalls)
	}
	if _, statErr := os.Lstat(leasePath + ".late-original"); statErr != nil {
		t.Fatalf("rollback cleanup removed late original lease: %v", statErr)
	}
	if got, readErr := os.ReadFile(leasePath); readErr != nil || string(got) != "late rollback lease replacement" {
		t.Fatalf("rollback cleanup changed late replacement lease: err=%v got=%q", readErr, got)
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

const vNextPublicationFIFOReaderScenarioEnv = "PM_CONNECTORGEN_FIFO_READER_SCENARIO"

func TestVNextPublicationFIFOReaderRefusesBeforeBlockingOpen(t *testing.T) {
	if scenario := os.Getenv(vNextPublicationFIFOReaderScenarioEnv); scenario != "" {
		vNextPublicationFIFOReaderSubprocess(t, scenario)
		return
	}

	for _, scenario := range []string{"stale-stage-owner", "public-current-recover", "public-journal-check", "admission-filesystem"} {
		t.Run(scenario, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestVNextPublicationFIFOReaderRefusesBeforeBlockingOpen$", "-test.v")
			command.Env = append(os.Environ(), vNextPublicationFIFOReaderScenarioEnv+"="+scenario)
			output, err := command.CombinedOutput()
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				t.Fatalf("FIFO %s child did not refuse before blocking open; output=%q", scenario, output)
			}
			if err != nil {
				t.Fatalf("FIFO %s child failed: %v; output=%q", scenario, err, output)
			}
		})
	}
}

func vNextPublicationFIFOReaderSubprocess(t *testing.T, scenario string) {
	t.Helper()
	switch scenario {
	case "stale-stage-owner":
		root := t.TempDir()
		publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
		if err != nil {
			t.Fatal(err)
		}
		pointer, err := publisher.Publish(vNextPublicationArtifactsForTest("active", false))
		if err != nil {
			t.Fatal(err)
		}
		stageName := ".stage-fifo-owner"
		stagePath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, stageName)
		vNextWriteOwnedStageForTest(t, stagePath, pointer, stageName, "stale")
		markerPath := filepath.Join(stagePath, vNextPublicationStageOwnerFile)
		if err := os.Remove(markerPath); err != nil {
			t.Fatal(err)
		}
		if err := unix.Mkfifo(markerPath, 0o600); err != nil {
			t.Fatal(err)
		}
		before := vNextPublicationTreeSnapshotForTest(t, filepath.Join(root, "acme"))
		fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
		if err != nil {
			t.Fatal(err)
		}
		if err := fresh.Recover(context.Background()); err == nil || !strings.Contains(err.Error(), "not a regular file") {
			t.Fatalf("Recover() with FIFO stage marker error = %v, want prompt regular-file refusal", err)
		}
		if info, err := os.Lstat(markerPath); err != nil || info.Mode()&os.ModeNamedPipe == 0 {
			t.Fatalf("FIFO stage marker changed after refusal: info=%v err=%v", info, err)
		}
		if after := vNextPublicationTreeSnapshotForTest(t, filepath.Join(root, "acme")); !bytes.Equal(before, after) {
			t.Fatal("FIFO stage marker refusal changed publication tree")
		}
		operation, err := fresh.openOperation(context.Background(), syscall.LOCK_EX, true)
		if err != nil {
			t.Fatalf("operation lock remained held after FIFO refusal: %v", err)
		}
		operation.close()
	case "public-current-recover", "public-journal-check":
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
		if code := runLockRender([]string{"lock-render", lock.Connector, "--defs", root}, &stdout, &stderr); code != 0 {
			t.Fatalf("initial lock-render = %d; stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		target := vNextPublicationCurrentFile
		if scenario == "public-journal-check" {
			target = vNextPublicationJournalFile
		}
		targetPath := filepath.Join(connectorRoot, target)
		savedPath := targetPath + ".before-fifo"
		_, err = os.Lstat(targetPath)
		hadOriginal := err == nil
		if !hadOriginal && !errors.Is(err, fs.ErrNotExist) {
			t.Fatal(err)
		}
		if hadOriginal {
			if err := os.Rename(targetPath, savedPath); err != nil {
				t.Fatal(err)
			}
		}
		if err := unix.Mkfifo(targetPath, 0o600); err != nil {
			t.Fatal(err)
		}
		before := vNextPublicationTreeSnapshotForTest(t, connectorRoot)
		if scenario == "public-current-recover" {
			publisher, err := newVNextGenerationPublisher(root, lock.Connector, vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			if err := publisher.Recover(context.Background()); err == nil || !strings.Contains(err.Error(), "not a regular file") {
				t.Fatalf("Recover() with FIFO %s error = %v, want prompt regular-file refusal", target, err)
			}
		} else {
			stdout.Reset()
			stderr.Reset()
			if code := runLockRender([]string{"lock-render", lock.Connector, "--defs", root, "--check"}, &stdout, &stderr); code != 1 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "not a regular file") {
				t.Fatalf("lock-render --check with FIFO %s = %d; stdout=%q stderr=%q", target, code, stdout.String(), stderr.String())
			}
		}
		if info, err := os.Lstat(targetPath); err != nil || info.Mode()&os.ModeNamedPipe == 0 {
			t.Fatalf("FIFO %s changed after refusal: info=%v err=%v", target, info, err)
		}
		if after := vNextPublicationTreeSnapshotForTest(t, connectorRoot); !bytes.Equal(before, after) {
			t.Fatalf("FIFO %s refusal changed public/private authority tree", target)
		}
		publisher, err := newVNextGenerationPublisher(root, lock.Connector, vNextPublicationHooks{})
		if err != nil {
			t.Fatal(err)
		}
		operation, err := publisher.openOperation(context.Background(), syscall.LOCK_EX, true)
		if err != nil {
			t.Fatalf("operation lock remained held after FIFO %s refusal: %v", target, err)
		}
		operation.close()
		if err := os.Remove(targetPath); err != nil {
			t.Fatal(err)
		}
		if hadOriginal {
			if err := os.Rename(savedPath, targetPath); err != nil {
				t.Fatal(err)
			}
		}
		stdout.Reset()
		stderr.Reset()
		if code := runLockRender([]string{"lock-render", lock.Connector, "--defs", root, "--check"}, &stdout, &stderr); code != 0 {
			t.Fatalf("lock-render --check after FIFO %s restoration = %d; stdout=%q stderr=%q", target, code, stdout.String(), stderr.String())
		}
	case "admission-filesystem":
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "spec.json"), []byte(`{"marker":"regular"}`), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(filepath.Join(root, "schemas"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "schemas", "record.json"), []byte(`{"type":"object"}`), 0o600); err != nil {
			t.Fatal(err)
		}
		directory, err := vNextPublicationOpenDirectory(root, "FIFO admission root")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := directory.Close(); err != nil {
				t.Errorf("close FIFO admission root: %v", err)
			}
		})
		stage := vNextPublicationStageFS{connector: "acme", root: vNextPublicationDirectoryFS{root: directory}}
		if _, err := fs.ReadDir(stage, "acme"); err != nil {
			t.Fatalf("enumerate regular staged filesystem: %v", err)
		}
		if _, err := fs.ReadDir(stage, "acme/schemas"); err != nil {
			t.Fatalf("enumerate nested staged schema directory: %v", err)
		}
		if payload, err := fs.ReadFile(stage, "acme/schemas/record.json"); err != nil || string(payload) != `{"type":"object"}` {
			t.Fatalf("read nested staged schema = %q, %v", payload, err)
		}
		member := filepath.Join(root, "spec.json")
		if err := os.Remove(member); err != nil {
			t.Fatal(err)
		}
		if err := unix.Mkfifo(member, 0o600); err != nil {
			t.Fatal(err)
		}
		before := vNextPublicationTreeSnapshotForTest(t, root)
		if _, err := fs.ReadFile(stage, "acme/spec.json"); err == nil || !strings.Contains(err.Error(), "not a regular file") {
			t.Fatalf("admission filesystem FIFO read error = %v, want prompt regular-file refusal", err)
		}
		if info, err := os.Lstat(member); err != nil || info.Mode()&os.ModeNamedPipe == 0 {
			t.Fatalf("FIFO admission member changed after refusal: info=%v err=%v", info, err)
		}
		if after := vNextPublicationTreeSnapshotForTest(t, root); !bytes.Equal(before, after) {
			t.Fatal("adapter-only FIFO refusal changed the staged filesystem")
		}
	default:
		t.Fatalf("unknown FIFO reader subprocess scenario %q", scenario)
	}
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

func TestVNextGenerationPublisherRefusesSecondPublicReplacementAtQuarantineRestore(t *testing.T) {
	const candidatePayload = "second public replacement"
	const replacementPayload = "third public replacement"

	t.Run("stage", func(t *testing.T) {
		root := t.TempDir()
		baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
		if err != nil {
			t.Fatal(err)
		}
		pointer, err := baseline.Publish(vNextPublicationArtifactsForTest("active", false))
		if err != nil {
			t.Fatal(err)
		}
		stageName := ".stage-second-replacement"
		stagePath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, stageName)
		vNextWriteOwnedStageForTest(t, stagePath, pointer, stageName, "original")

		witness := newVNextPublicationSecondReplacementWitnessForTest(t, stagePath, vNextPublicationAfterStageRemovalIdentity, []byte(candidatePayload), []byte(replacementPayload))
		guard, err := newVNextGenerationPublisher(root, "acme", witness.hooks())
		if err != nil {
			t.Fatal(err)
		}
		err = guard.Recover(context.Background())
		if err == nil || !strings.Contains(err.Error(), "identity") {
			t.Fatalf("Recover() after second stage replacement error = %v, want identity refusal", err)
		}
		witness.assertPreserved(t)
		if !errors.Is(err, fs.ErrExist) {
			t.Fatalf("Recover() after second stage replacement error = %v, want no-replace conflict", err)
		}
		if got, readErr := os.ReadFile(filepath.Join(witness.originalPath, "sentinel.txt")); readErr != nil || string(got) != "original" {
			t.Fatalf("stage original did not remain intact: err=%v got=%q", readErr, got)
		}
	})

	t.Run("generation", func(t *testing.T) {
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
		current, publishErr := baseline.Publish(vNextPublicationArtifactsForTest("current", false))
		handle.Release()
		if publishErr != nil {
			t.Fatal(publishErr)
		}
		generationPath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, old.Generation)

		witness := newVNextPublicationSecondReplacementWitnessForTest(t, generationPath, vNextPublicationAfterGenerationRemovalIdentity, []byte(candidatePayload), []byte(replacementPayload))
		guard, err := newVNextGenerationPublisher(root, "acme", witness.hooks())
		if err != nil {
			t.Fatal(err)
		}
		operation, err := guard.openOperation(context.Background(), syscall.LOCK_EX, true)
		if err != nil {
			t.Fatal(err)
		}
		err = guard.pruneLocked(operation, current.Generation)
		operation.close()
		if err == nil || !strings.Contains(err.Error(), "identity") {
			t.Fatalf("Prune() after second generation replacement error = %v, want identity refusal", err)
		}
		witness.assertPreserved(t)
		if !errors.Is(err, fs.ErrExist) {
			t.Fatalf("Prune() after second generation replacement error = %v, want no-replace conflict", err)
		}
		if got, readErr := os.ReadFile(filepath.Join(witness.originalPath, "metadata.json")); readErr != nil || string(got) != `{"marker":"old"}` {
			t.Fatalf("generation original did not remain intact: err=%v got=%q", readErr, got)
		}
	})

	t.Run("rollback", func(t *testing.T) {
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
		generationPath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, vNextPublicationGenerationID(newSet.Files))

		witness := newVNextPublicationSecondReplacementWitnessForTest(t, generationPath, vNextPublicationAfterGenerationRemovalIdentity, []byte(candidatePayload), []byte(replacementPayload))
		guard, err := newVNextGenerationPublisher(root, "acme", witness.hooks())
		if err != nil {
			t.Fatal(err)
		}
		_, err = guard.Publish(newSet)
		if err == nil || !strings.Contains(err.Error(), "identity") {
			t.Fatalf("Publish(new) after second rollback replacement error = %v, want identity refusal", err)
		}
		if validationCalls != 2 {
			t.Fatalf("validation calls = %d, want staged then active validation", validationCalls)
		}
		witness.assertPreserved(t)
		if !errors.Is(err, fs.ErrExist) {
			t.Fatalf("Publish(new) after second rollback replacement error = %v, want no-replace conflict", err)
		}
		if got, readErr := os.ReadFile(filepath.Join(witness.originalPath, "metadata.json")); readErr != nil || string(got) != `{"marker":"new"}` {
			t.Fatalf("rollback generation original did not remain intact: err=%v got=%q", readErr, got)
		}
	})
}

func TestVNextGenerationPublisherRefusesFinalGenerationActivationCollision(t *testing.T) {
	root := t.TempDir()
	var generationPath string
	var collisionInfo os.FileInfo
	collisionHit := false
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationBeforeStageRename || collisionHit {
			return nil
		}
		collisionHit = true
		if err := os.Mkdir(generationPath, 0o755); err != nil {
			t.Fatal(err)
		}
		var statErr error
		collisionInfo, statErr = os.Stat(generationPath)
		if statErr != nil {
			t.Fatal(statErr)
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	operation, err := publisher.openOperation(context.Background(), syscall.LOCK_EX, true)
	if err != nil {
		t.Fatal(err)
	}
	stageName, pointer, err := publisher.stageLocked(operation, vNextPublicationArtifactsForTest("active", false).Files, nil)
	if err != nil {
		operation.close()
		t.Fatal(err)
	}
	stagePath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, stageName)
	stageInfo, err := os.Stat(stagePath)
	if err != nil {
		operation.close()
		t.Fatal(err)
	}
	generationPath = filepath.Join(root, "acme", vNextPublicationGenerationDirectory, pointer.Generation)
	err = publisher.activateStageLocked(operation, stageName, pointer.Generation)
	operation.close()

	if !collisionHit {
		t.Fatal("activateStageLocked() did not reach final generation collision seam")
	}
	if !errors.Is(err, fs.ErrExist) {
		t.Fatalf("activateStageLocked() collision error = %v, want no-replace destination conflict", err)
	}
	if info, statErr := os.Stat(stagePath); statErr != nil || !os.SameFile(info, stageInfo) {
		t.Fatalf("activation collision did not retain staged generation: err=%v info=%v", statErr, info)
	}
	if info, statErr := os.Stat(generationPath); statErr != nil || !os.SameFile(info, collisionInfo) {
		t.Fatalf("activation collision did not retain destination: err=%v info=%v", statErr, info)
	}
}

type vNextPublicationSecondReplacementWitness struct {
	publicPath         string
	afterIdentity      vNextPublicationFaultPoint
	candidatePayload   []byte
	replacementPayload []byte
	originalPath       string
	originalInfo       os.FileInfo
	candidatePath      string
	candidateInfo      os.FileInfo
	publicInfo         os.FileInfo
	identityHit        bool
	restoreHit         bool
}

func newVNextPublicationSecondReplacementWitnessForTest(t *testing.T, publicPath string, afterIdentity vNextPublicationFaultPoint, candidatePayload, replacementPayload []byte) *vNextPublicationSecondReplacementWitness {
	t.Helper()
	return &vNextPublicationSecondReplacementWitness{
		publicPath:         publicPath,
		afterIdentity:      afterIdentity,
		candidatePayload:   candidatePayload,
		replacementPayload: replacementPayload,
	}
}

func (w *vNextPublicationSecondReplacementWitness) hooks() vNextPublicationHooks {
	return vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point == w.afterIdentity && !w.identityHit {
			w.identityHit = true
			originalPath, err := vNextReplaceDirectoryWithRegularFileForTest(w.publicPath, w.candidatePayload)
			if err != nil {
				return err
			}
			w.originalPath = originalPath
			originalInfo, err := os.Stat(w.originalPath)
			if err != nil {
				return err
			}
			w.originalInfo = originalInfo
			return nil
		}
		if point != vNextPublicationBeforeQuarantineRestore || !w.identityHit || w.restoreHit {
			return nil
		}
		w.restoreHit = true
		if _, err := os.Lstat(w.publicPath); !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("public replacement before quarantine restore = %v, want absent", err)
		}
		candidatePath, err := vNextPublicationQuarantinedCandidateForTest(filepath.Dir(w.publicPath), w.candidatePayload)
		if err != nil {
			return err
		}
		w.candidatePath = candidatePath
		candidateInfo, err := os.Stat(w.candidatePath)
		if err != nil {
			return err
		}
		if !candidateInfo.Mode().IsRegular() {
			return fmt.Errorf("quarantined second replacement mode = %v, want regular file", candidateInfo.Mode())
		}
		w.candidateInfo = candidateInfo
		if err := os.WriteFile(w.publicPath, w.replacementPayload, 0o600); err != nil {
			return err
		}
		publicInfo, err := os.Stat(w.publicPath)
		if err != nil {
			return err
		}
		if !publicInfo.Mode().IsRegular() {
			return fmt.Errorf("third public replacement mode = %v, want regular file", publicInfo.Mode())
		}
		w.publicInfo = publicInfo
		return nil
	}}
}

func (w *vNextPublicationSecondReplacementWitness) assertPreserved(t *testing.T) {
	t.Helper()
	if !w.identityHit {
		t.Fatal("cleanup did not reach post-identity replacement seam")
	}
	if !w.restoreHit {
		t.Fatal("cleanup did not reach quarantine restoration seam")
	}
	if info, err := os.Stat(w.originalPath); err != nil || !os.SameFile(info, w.originalInfo) {
		t.Fatalf("moved original identity changed: err=%v info=%v", err, info)
	}
	if got, err := os.ReadFile(w.publicPath); err != nil || !bytes.Equal(got, w.replacementPayload) {
		t.Fatalf("third public replacement was not retained: err=%v got=%q want=%q", err, got, w.replacementPayload)
	}
	if info, err := os.Stat(w.publicPath); err != nil || !os.SameFile(info, w.publicInfo) {
		t.Fatalf("third public replacement inode changed: err=%v info=%v", err, info)
	}
	if got, err := os.ReadFile(w.candidatePath); err != nil || !bytes.Equal(got, w.candidatePayload) {
		t.Fatalf("quarantined second replacement was not retained: err=%v got=%q want=%q", err, got, w.candidatePayload)
	}
	if info, err := os.Stat(w.candidatePath); err != nil || !os.SameFile(info, w.candidateInfo) {
		t.Fatalf("quarantined second replacement inode changed: err=%v info=%v", err, info)
	}
}

func vNextReplaceDirectoryWithRegularFileForTest(path string, payload []byte) (string, error) {
	moved := path + ".late-original"
	if err := os.Rename(path, moved); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		return "", err
	}
	return moved, nil
}

func vNextPublicationQuarantinedCandidateForTest(parent string, payload []byte) (string, error) {
	entries, err := os.ReadDir(parent)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), ".connectorgen-quarantine-") {
			continue
		}
		candidate := filepath.Join(parent, entry.Name(), vNextPublicationQuarantineMember)
		got, err := os.ReadFile(candidate)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		if bytes.Equal(got, payload) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("publication quarantine does not retain %q", payload)
}
