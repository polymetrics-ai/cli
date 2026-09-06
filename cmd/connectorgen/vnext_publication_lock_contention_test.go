package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

const (
	vNextPublicationContentionScenarioEnv = "PM_CONNECTORGEN_CP11_F08_CONTENTION_SCENARIO"
	vNextPublicationContentionRootEnv     = "PM_CONNECTORGEN_CP11_F08_CONTENTION_ROOT"
	vNextPublicationContentionPreLockEnv  = "PM_CONNECTORGEN_CP11_F08_PRELOCK_PATH"
	vNextPublicationContentionReleaseEnv  = "PM_CONNECTORGEN_CP11_F08_RELEASE_PATH"
	vNextPublicationContentionAckEnv      = "PM_CONNECTORGEN_CP11_F08_ACK_PATH"
	vNextPublicationContentionResultEnv   = "PM_CONNECTORGEN_CP11_F08_RESULT_PATH"
)

type vNextPublicationContentionAck struct {
	Device uint64 `json:"device"`
	Inode  uint64 `json:"inode"`
}

type vNextPublicationContentionResult struct {
	Code   int    `json:"code"`
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

func TestConnectorgenMainSignalsOnlyAfterExactLockContention(t *testing.T) {
	if scenario := os.Getenv(vNextPublicationContentionScenarioEnv); scenario != "" {
		vNextPublicationContentionChild(t, scenario)
		return
	}

	for _, test := range []struct {
		name   string
		signal os.Signal
	}{
		{name: "interrupt", signal: os.Interrupt},
		{name: "terminate", signal: syscall.SIGTERM},
	} {
		t.Run(test.name, func(t *testing.T) {
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
			var initialOut, initialErr bytes.Buffer
			if code := runLockRender([]string{"lock-render", lock.Connector, "--defs", root}, &initialOut, &initialErr); code != 0 {
				t.Fatalf("initial lock-render = %d; stdout=%q stderr=%q", code, initialOut.String(), initialErr.String())
			}
			currentPath := filepath.Join(connectorRoot, vNextPublicationCurrentFile)
			currentBefore, err := os.ReadFile(currentPath)
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
			heldIdentity, err := vNextPublicationIdentityFromFile(held, "test-held publication lock")
			if err != nil {
				t.Fatal(err)
			}

			preLockPath := filepath.Join(root, "pre-lock.json")
			releasePath := filepath.Join(root, "release-lock-attempt")
			ackPath := filepath.Join(root, "contention.json")
			resultPath := filepath.Join(root, "result.json")
			command := exec.Command(os.Args[0], "-test.run=^TestConnectorgenMainSignalsOnlyAfterExactLockContention$", "-test.v")
			command.Env = append(os.Environ(),
				vNextPublicationContentionScenarioEnv+"="+test.name,
				vNextPublicationContentionRootEnv+"="+root,
				vNextPublicationContentionPreLockEnv+"="+preLockPath,
				vNextPublicationContentionReleaseEnv+"="+releasePath,
				vNextPublicationContentionAckEnv+"="+ackPath,
				vNextPublicationContentionResultEnv+"="+resultPath,
			)
			var output bytes.Buffer
			command.Stdout = &output
			command.Stderr = &output
			child := vNextPublicationStartBoundedChildForTest(t, command, "real-main contention child")

			// lsof is deliberately only the old, insufficient directory-open
			// observation. The pre-lock gate proves it cannot authorize the
			// signal below on its own.
			vNextPublicationWaitForProcessOpenPathForTest(t, command.Process.Pid, connectorRoot, 2*time.Second)
			vNextPublicationWaitForContentionFileForTest(t, preLockPath, time.Second)
			time.Sleep(100 * time.Millisecond)
			if _, err := os.Stat(ackPath); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("directory-open readiness reached lock contention before the explicit release: err=%v", err)
			}

			if err := os.WriteFile(releasePath, []byte("release"), 0o600); err != nil {
				t.Fatal(err)
			}
			ack := vNextPublicationReadContentionAckForTest(t, ackPath, time.Second)
			if ack.Device != heldIdentity.device || ack.Inode != heldIdentity.inode {
				t.Fatalf("contention ack identity = %d/%d, want held connector directory %d/%d", ack.Device, ack.Inode, heldIdentity.device, heldIdentity.inode)
			}
			if err := command.Process.Signal(test.signal); err != nil {
				t.Fatalf("signal contended real-main subprocess with %v: %v", test.signal, err)
			}
			if err, completed := child.waitWithin(1500 * time.Millisecond); !completed {
				child.killAndWait(t, "real-main contention timeout")
				t.Fatalf("real-main contention child did not exit after %v; output=%q", test.signal, output.String())
			} else if err != nil {
				t.Fatalf("real-main contention child failed: %v; output=%q", err, output.String())
			}
			result := vNextPublicationReadContentionResultForTest(t, resultPath)
			if result.Code != 1 || result.Stdout != "" || !strings.Contains(result.Stderr, context.Canceled.Error()) {
				t.Fatalf("real-main contention result after %v = %#v, want exit 1/no stdout/context cancellation", test.signal, result)
			}
			if after, err := os.ReadFile(currentPath); err != nil || !bytes.Equal(currentBefore, after) {
				t.Fatalf("contended lock-render changed CURRENT: err=%v before=%q after=%q", err, currentBefore, after)
			}
			if stateAfter := vNextPublicationTreeSnapshotForTest(t, connectorRoot); !bytes.Equal(stateBefore, stateAfter) {
				t.Fatalf("contended lock-render changed selected/control/authority/generation state: before=%s after=%s", stateBefore, stateAfter)
			}

			unlockVNextPublicationFile(held)
			held = nil
			var retryOut, retryErr bytes.Buffer
			if code := runLockRender([]string{"lock-render", lock.Connector, "--defs", root, "--check"}, &retryOut, &retryErr); code != 0 {
				t.Fatalf("lock-render retry = %d; stdout=%q stderr=%q", code, retryOut.String(), retryErr.String())
			}
		})
	}
}

func vNextPublicationContentionChild(t *testing.T, scenario string) {
	t.Helper()
	root := os.Getenv(vNextPublicationContentionRootEnv)
	preLockPath := os.Getenv(vNextPublicationContentionPreLockEnv)
	releasePath := os.Getenv(vNextPublicationContentionReleaseEnv)
	ackPath := os.Getenv(vNextPublicationContentionAckEnv)
	resultPath := os.Getenv(vNextPublicationContentionResultEnv)
	if root == "" || preLockPath == "" || releasePath == "" || ackPath == "" || resultPath == "" {
		t.Fatal("missing real-main contention child paths")
	}
	var once sync.Once
	vNextPublicationLockRenderHooksForTest = vNextPublicationHooks{
		At: func(point vNextPublicationFaultPoint) error {
			if point != vNextPublicationBeforeLockAcquire {
				return nil
			}
			if err := os.WriteFile(preLockPath, []byte(scenario), 0o600); err != nil {
				return err
			}
			return vNextPublicationWaitForContentionReleaseForTest(releasePath, 5*time.Second)
		},
		LockContention: func(identity vNextPublicationIdentity) error {
			var writeErr error
			once.Do(func() {
				payload, err := json.Marshal(vNextPublicationContentionAck{Device: identity.device, Inode: identity.inode})
				if err != nil {
					writeErr = err
					return
				}
				writeErr = os.WriteFile(ackPath, payload, 0o600)
			})
			return writeErr
		},
	}
	t.Cleanup(func() { vNextPublicationLockRenderHooksForTest = vNextPublicationHooks{} })
	var stdout, stderr bytes.Buffer
	code := runMain([]string{"lock-render", "acme", "--defs", root, "--check"}, &stdout, &stderr)
	payload, err := json.Marshal(vNextPublicationContentionResult{Code: code, Stdout: stdout.String(), Stderr: stderr.String()})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(resultPath, payload, 0o600); err != nil {
		t.Fatal(err)
	}
}

func vNextPublicationWaitForContentionReleaseForTest(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return nil
		} else if !os.IsNotExist(err) {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
	return errors.New("contention pre-lock release did not arrive")
}

func vNextPublicationWaitForContentionFileForTest(t *testing.T, path string, timeout time.Duration) {
	t.Helper()
	if err := vNextPublicationWaitForContentionReleaseForTest(path, timeout); err != nil {
		t.Fatal(err)
	}
}

func vNextPublicationReadContentionAckForTest(t *testing.T, path string, timeout time.Duration) vNextPublicationContentionAck {
	t.Helper()
	vNextPublicationWaitForContentionFileForTest(t, path, timeout)
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var ack vNextPublicationContentionAck
	if err := json.Unmarshal(payload, &ack); err != nil {
		t.Fatal(err)
	}
	return ack
}

func vNextPublicationReadContentionResultForTest(t *testing.T, path string) vNextPublicationContentionResult {
	t.Helper()
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var result vNextPublicationContentionResult
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatal(err)
	}
	return result
}
