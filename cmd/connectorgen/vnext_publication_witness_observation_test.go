package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

const (
	vNextPublicationSafeWitnessScenarioEnv = "PM_CONNECTORGEN_CP11_F04_SAFE_WITNESS_SCENARIO"
	vNextPublicationSafeWitnessPathEnv     = "PM_CONNECTORGEN_CP11_F04_SAFE_WITNESS_PATH"
)

// vNextPublicationWitnessAfterOpenForTest is a nil-by-default test seam at
// the helper's actual retained-descriptor boundary. Production code has no
// witness helper and no configuration path to this callback.
var vNextPublicationWitnessAfterOpenForTest func(kind, path string, opened os.FileInfo)

type vNextPublicationSafeWitnessBoundary struct {
	Scenario string `json:"scenario"`
	ADevice  uint64 `json:"a_device"`
	AInode   uint64 `json:"a_inode"`
	BDevice  uint64 `json:"b_device"`
	BInode   uint64 `json:"b_inode"`
}

func TestCP11F04RWitnessRetainsOpenedObjectAcrossReplacement(t *testing.T) {
	if scenario := os.Getenv(vNextPublicationSafeWitnessScenarioEnv); scenario != "" {
		vNextPublicationSafeWitnessChild(t, scenario)
		return
	}

	for _, scenario := range []string{"fifo", "symlink", "directory"} {
		t.Run(scenario, func(t *testing.T) {
			witnessPath := filepath.Join(t.TempDir(), "boundary.json")
			command := exec.Command(os.Args[0], "-test.run=^TestCP11F04RWitnessRetainsOpenedObjectAcrossReplacement$", "-test.v")
			command.Env = append(os.Environ(), vNextPublicationSafeWitnessScenarioEnv+"="+scenario, vNextPublicationSafeWitnessPathEnv+"="+witnessPath)
			var output bytes.Buffer
			command.Stdout = &output
			command.Stderr = &output
			child := vNextPublicationStartBoundedChildForTest(t, command, "safe witness child")
			boundary := vNextPublicationWaitForSafeWitnessBoundaryForTest(t, witnessPath, time.Second)
			if boundary.Scenario != scenario || (boundary.ADevice == boundary.BDevice && boundary.AInode == boundary.BInode) {
				t.Fatalf("safe witness %s boundary = %#v, want distinct A/B identities", scenario, boundary)
			}
			if err, completed := child.waitWithin(3 * time.Second); !completed {
				child.killAndWait(t, "safe witness timeout")
				t.Fatalf("safe %s witness did not complete; output=%q", scenario, output.String())
			} else if err != nil {
				t.Fatalf("safe %s witness did not retain opened A: %v; output=%q", scenario, err, output.String())
			}
		})
	}
}

func vNextPublicationSafeWitnessChild(t *testing.T, scenario string) {
	t.Helper()
	root := t.TempDir()
	target := filepath.Join(root, "target")
	original := []byte("original-A payload")
	replacement := []byte("replacement-B payload")
	switch scenario {
	case "fifo", "symlink":
		if err := os.WriteFile(target, original, 0o600); err != nil {
			t.Fatal(err)
		}
	case "directory":
		if err := os.Mkdir(target, 0o700); err != nil {
			t.Fatal(err)
		}
		original = []byte("original-A metadata")
		if err := os.WriteFile(filepath.Join(target, "metadata.json"), original, 0o600); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatalf("unknown safe witness scenario %q", scenario)
	}

	fired := false
	var openedA os.FileInfo
	vNextPublicationWitnessAfterOpenForTest = func(kind, path string, observed os.FileInfo) {
		if path != target || fired {
			return
		}
		fired = true
		openedA = observed
		switch scenario {
		case "fifo":
			if kind != "file" {
				t.Fatalf("FIFO witness kind = %q, want file", kind)
			}
			if err := os.Remove(target); err != nil {
				t.Fatal(err)
			}
			if err := unix.Mkfifo(target, 0o600); err != nil {
				t.Fatal(err)
			}
		case "symlink":
			if kind != "file" {
				t.Fatalf("symlink witness kind = %q, want file", kind)
			}
			external := filepath.Join(root, "replacement-B")
			if err := os.WriteFile(external, replacement, 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(target); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(external, target); err != nil {
				t.Fatal(err)
			}
		case "directory":
			if kind != "directory" {
				t.Fatalf("directory witness kind = %q, want directory", kind)
			}
			if err := os.Rename(target, target+"-A"); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(target, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(target, "metadata.json"), replacement, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		replacementInfo, err := os.Lstat(target)
		if err != nil {
			t.Fatal(err)
		}
		witnessPath := os.Getenv(vNextPublicationSafeWitnessPathEnv)
		if witnessPath == "" {
			t.Fatal("missing safe witness boundary path")
		}
		boundary := vNextPublicationSafeWitnessBoundary{
			Scenario: scenario,
			ADevice:  vNextPublicationInfoDeviceForTest(observed),
			AInode:   vNextPublicationInfoInodeForTest(observed),
			BDevice:  vNextPublicationInfoDeviceForTest(replacementInfo),
			BInode:   vNextPublicationInfoInodeForTest(replacementInfo),
		}
		encoded, err := json.Marshal(boundary)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(witnessPath, encoded, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { vNextPublicationWitnessAfterOpenForTest = nil })

	var got vNextPublicationPathWitness
	if scenario == "directory" {
		got = vNextPublicationDirectoryWitnessForTest(t, target)
	} else {
		got = vNextPublicationFileWitnessForTest(t, target)
	}
	if !fired || openedA == nil {
		t.Fatal("safe witness did not reach the retained-descriptor boundary")
	}
	if !os.SameFile(got.info, openedA) {
		t.Fatalf("safe witness identity = dev/inode %d/%d, want opened A %d/%d", vNextPublicationInfoDeviceForTest(got.info), vNextPublicationInfoInodeForTest(got.info), vNextPublicationInfoDeviceForTest(openedA), vNextPublicationInfoInodeForTest(openedA))
	}
	if !bytes.Equal(got.payload, original) {
		t.Fatalf("safe witness payload = %q, want retained A %q", got.payload, original)
	}
}

func vNextPublicationWaitForSafeWitnessBoundaryForTest(t *testing.T, path string, timeout time.Duration) vNextPublicationSafeWitnessBoundary {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		payload, err := os.ReadFile(path)
		if err == nil {
			var boundary vNextPublicationSafeWitnessBoundary
			if err := json.Unmarshal(payload, &boundary); err != nil {
				t.Fatalf("decode safe witness boundary %q: %v", payload, err)
			}
			return boundary
		}
		if !os.IsNotExist(err) {
			t.Fatalf("read safe witness boundary: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("safe witness child did not reach its parent-visible boundary %q", path)
	return vNextPublicationSafeWitnessBoundary{}
}
