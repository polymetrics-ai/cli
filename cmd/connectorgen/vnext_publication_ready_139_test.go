package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

const vNextPublicationReadyBarrierEnv = "PM_CONNECTORGEN_CP13_READY_BARRIER"

// A readiness name belongs to one child in its private test directory. Publish
// only a closed, complete payload; parent readers deliberately reject malformed
// JSON instead of treating it as a reason to retry.
func vNextPublicationWriteReadyForTest(path string, payload []byte) (result error) {
	f, err := os.CreateTemp(filepath.Dir(path), ".ready-*")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(f.Name()); err != nil && !errors.Is(err, os.ErrNotExist) {
			result = errors.Join(result, err)
		}
	}()
	if err := vNextPublicationReadyBarrierForTest(path); err != nil {
		return errors.Join(err, f.Close())
	}
	_, err = f.Write(payload)
	if err := errors.Join(err, f.Close()); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

func vNextPublicationReadyBarrierForTest(path string) error {
	root := os.Getenv(vNextPublicationReadyBarrierEnv)
	if root == "" {
		return nil
	}
	// The schedule notification is itself completed before publication.
	pending, opened := filepath.Join(root, "pending"), filepath.Join(root, "opened")
	if err := os.WriteFile(pending, []byte(path), 0o600); err != nil {
		return err
	}
	if err := os.Rename(pending, opened); err != nil {
		return err
	}
	return vNextPublicationWaitForContentionReleaseForTest(filepath.Join(root, "release"), 2*time.Second)
}

func TestVNextPublication139ReadyPublication(t *testing.T) {
	for _, tc := range []struct{ name, selector string }{
		{"contention", "^TestConnectorgenMainSignalsOnlyAfterExactLockContention$/^interrupt$"},
		{"snapshot", "^TestVNextPublicationTreeSnapshotRefusesAtoBReplacement$/^fifo$"},
		{"retained witness", "^TestCP11F04RWitnessRetainsOpenedObjectAcrossReplacement$/^fifo$"},
	} {
		for _, paused := range []bool{true, false} {
			name := tc.name + "/completed control"
			if paused {
				name = tc.name + "/before payload write"
			}
			t.Run(name, func(t *testing.T) {
				root := t.TempDir()
				command := exec.Command(os.Args[0], "-test.run="+tc.selector, "-test.v")
				command.Env = append(os.Environ(), vNextPublicationReadyBarrierEnv+"=")
				if paused {
					command.Env = append(command.Env, vNextPublicationReadyBarrierEnv+"="+root)
				}
				var output bytes.Buffer
				command.Stdout, command.Stderr = &output, &output
				child := vNextPublicationStartBoundedChildForTest(t, command, tc.name+" readiness parent")
				if paused {
					// Release even after an assertion/setup failure; the original
					// parent retains its bounded cleanup and direct child Wait.
					release := func() error { return os.WriteFile(filepath.Join(root, "release"), []byte("release"), 0o600) }
					t.Cleanup(func() {
						if err := release(); err != nil {
							t.Error(err)
						}
					})
					if err := vNextPublicationWaitForContentionReleaseForTest(filepath.Join(root, "opened"), 2*time.Second); err != nil {
						t.Fatal(err)
					}
					readyPath, err := os.ReadFile(filepath.Join(root, "opened"))
					if err != nil {
						t.Fatal(err)
					}
					_, visibleErr := os.Lstat(string(readyPath))
					if err := release(); err != nil {
						t.Fatal(err)
					}
					if !errors.Is(visibleErr, os.ErrNotExist) {
						t.Errorf("readiness visible before completed payload: path=%q err=%v", readyPath, visibleErr)
					}
				}
				if err, completed := child.waitWithin(5 * time.Second); !completed {
					child.killAndWait(t, tc.name+" readiness parent")
					t.Fatalf("readiness parent did not finish: %s", output.String())
				} else if err != nil {
					t.Fatalf("actual producer/reader failed: %v\n%s", err, output.String())
				}
			})
		}
	}
}
