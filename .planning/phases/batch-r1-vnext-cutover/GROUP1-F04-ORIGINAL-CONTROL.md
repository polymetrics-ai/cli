# CP11 F-04-R original-behaviour negative control

This is the exact record requested by Firstmate steer 053 before changing the
tested witness helpers. It proves an **oracle/helper** defect; it is not a
production repair GREEN and does not reopen accepted production R3-01.

## Immutable execution record

- Parent local HEAD: `67ff7a7ababdbd4d91d9a0b5f9b9d6705fb3c189`.
- Command: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestCP11F04ROriginalWitnessCanMixPathObservations$' -v`
- Shell exit: `0`; elapsed wall time: `4.5s`; Go package duration:
  `ok polymetrics.ai/cmd/connectorgen 1.199s`.
- Exact complete output:

```text
=== RUN   TestCP11F04ROriginalWitnessCanMixPathObservations
=== RUN   TestCP11F04ROriginalWitnessCanMixPathObservations/fifo
    vnext_publication_witness_observation_test.go:52: original unsafe witness fifo recorded distinct A/B identities before its pathname read
    vnext_publication_witness_observation_test.go:58: bounded cleanup observed direct Wait completion for unsafe FIFO witness bound
=== RUN   TestCP11F04ROriginalWitnessCanMixPathObservations/symlink
    vnext_publication_witness_observation_test.go:52: original unsafe witness symlink recorded distinct A/B identities before its pathname read
=== RUN   TestCP11F04ROriginalWitnessCanMixPathObservations/directory
    vnext_publication_witness_observation_test.go:52: original unsafe witness directory recorded distinct A/B identities before its pathname read
--- PASS: TestCP11F04ROriginalWitnessCanMixPathObservations (0.28s)
    --- PASS: TestCP11F04ROriginalWitnessCanMixPathObservations/fifo (0.19s)
    --- PASS: TestCP11F04ROriginalWitnessCanMixPathObservations/symlink (0.05s)
    --- PASS: TestCP11F04ROriginalWitnessCanMixPathObservations/directory (0.04s)
PASS
ok  	polymetrics.ai/cmd/connectorgen	1.199s
```

The parent received a JSON A/B identity witness before the FIFO child entered
the bounded hang. It then used the existing immediate child owner to issue
cleanup `SIGKILL` and observe direct `Wait` completion. That cleanup is not
signal-success evidence. The symlink and directory children exited normally
only because their test assertions observed the original A-classification /
B-payload defect.

## Exact source and test-only seam identity

- Original unsafe helper source after adding only its pre-repair seam:
  `cmd/connectorgen/vnext_publication_durable_matrix_test.go`, SHA-256
  `38e7416617d4702d14784b50fb0eb28111897dac6645f4f3e285dad127fd2eb4`.
- Complete bounded control and inert hook snapshot:
  `cmd/connectorgen/vnext_publication_witness_observation_test.go`, SHA-256
  `783be93cd71e7aa6345458906e002d95acb7cbf0905047c873a00a84741fd33b`.
- The hook was `vNextPublicationWitnessAfterLstatForTest`; it exists only in
  a `_test.go` file and fires exactly after old `Lstat` but before the unsafe
  pathname `ReadFile`/directory member read. It does not alter product source.

The exact source snapshot below plus its SHA-256 and parent `67ff7a7` makes the
small helper delta reconstructable without putting an otherwise redundant raw
patch inside a Markdown evidence file. Later replacement of the unsafe test
seam cannot rewrite this original-behaviour evidence.

```go
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
	vNextPublicationUnsafeWitnessScenarioEnv = "PM_CONNECTORGEN_CP11_F04_UNSAFE_WITNESS_SCENARIO"
	vNextPublicationUnsafeWitnessPathEnv     = "PM_CONNECTORGEN_CP11_F04_UNSAFE_WITNESS_PATH"
)

var vNextPublicationWitnessAfterLstatForTest func(kind, path string, original os.FileInfo)

type vNextPublicationUnsafeWitnessBoundary struct {
	Scenario string `json:"scenario"`
	ADevice  uint64 `json:"a_device"`
	AInode   uint64 `json:"a_inode"`
	BDevice  uint64 `json:"b_device"`
	BInode   uint64 `json:"b_inode"`
}

func TestCP11F04ROriginalWitnessCanMixPathObservations(t *testing.T) {
	if scenario := os.Getenv(vNextPublicationUnsafeWitnessScenarioEnv); scenario != "" {
		vNextPublicationUnsafeWitnessChild(t, scenario)
		return
	}

	for _, scenario := range []string{"fifo", "symlink", "directory"} {
		t.Run(scenario, func(t *testing.T) {
			witnessPath := filepath.Join(t.TempDir(), "boundary.json")
			command := exec.Command(os.Args[0], "-test.run=^TestCP11F04ROriginalWitnessCanMixPathObservations$", "-test.v")
			command.Env = append(os.Environ(), vNextPublicationUnsafeWitnessScenarioEnv+"="+scenario, vNextPublicationUnsafeWitnessPathEnv+"="+witnessPath)
			var output bytes.Buffer
			command.Stdout = &output
			command.Stderr = &output
			child := vNextPublicationStartBoundedChildForTest(t, command, "unsafe witness child")
			boundary := vNextPublicationWaitForUnsafeWitnessBoundaryForTest(t, witnessPath, time.Second)
			if boundary.Scenario != scenario || (boundary.ADevice == boundary.BDevice && boundary.AInode == boundary.BInode) {
				t.Fatalf("unsafe witness %s boundary = %#v, want distinct A/B identities", scenario, boundary)
			}
			t.Logf("original unsafe witness %s recorded distinct A/B identities before its pathname read", scenario)

			if scenario == "fifo" {
				if _, completed := child.waitWithin(150 * time.Millisecond); completed {
					t.Fatalf("unsafe FIFO witness returned instead of blocking; output=%q", output.String())
				}
				child.killAndWait(t, "unsafe FIFO witness bound")
				return
			}
			if err, completed := child.waitWithin(time.Second); !completed {
				child.killAndWait(t, "unsafe witness timeout")
				t.Fatalf("unsafe %s witness did not complete; output=%q", scenario, output.String())
			} else if err != nil {
				t.Fatalf("unsafe %s witness did not observe the mixed object: %v; output=%q", scenario, err, output.String())
			}
		})
	}
}

func vNextPublicationUnsafeWitnessChild(t *testing.T, scenario string) {
	t.Helper()
	root := t.TempDir()
	target := filepath.Join(root, "target")
	secret := []byte("replacement-B payload")
	switch scenario {
	case "fifo", "symlink":
		if err := os.WriteFile(target, []byte("original-A payload"), 0o600); err != nil {
			t.Fatal(err)
		}
	case "directory":
		if err := os.Mkdir(target, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(target, "metadata.json"), []byte("original-A metadata"), 0o600); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatalf("unknown unsafe witness scenario %q", scenario)
	}

	fired := false
	vNextPublicationWitnessAfterLstatForTest = func(kind, path string, original os.FileInfo) {
		if path != target || fired {
			return
		}
		fired = true
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
			replacement := filepath.Join(root, "replacement-B")
			if err := os.WriteFile(replacement, secret, 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(target); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(replacement, target); err != nil {
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
			if err := os.WriteFile(filepath.Join(target, "metadata.json"), secret, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		replacement, err := os.Lstat(target)
		if err != nil {
			t.Fatal(err)
		}
		boundaryPath := os.Getenv(vNextPublicationUnsafeWitnessPathEnv)
		if boundaryPath == "" {
			t.Fatal("missing unsafe witness boundary path")
		}
		boundary := vNextPublicationUnsafeWitnessBoundary{
			Scenario: scenario,
			ADevice:  vNextPublicationInfoDeviceForTest(original),
			AInode:   vNextPublicationInfoInodeForTest(original),
			BDevice:  vNextPublicationInfoDeviceForTest(replacement),
			BInode:   vNextPublicationInfoInodeForTest(replacement),
		}
		encoded, err := json.Marshal(boundary)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(boundaryPath, encoded, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { vNextPublicationWitnessAfterLstatForTest = nil })

	if scenario == "directory" {
		got := vNextPublicationDirectoryWitnessForTest(t, target)
		if !bytes.Equal(got.payload, secret) {
			t.Fatalf("unsafe directory witness payload = %q, want replacement B bytes", got.payload)
		}
	} else {
		got := vNextPublicationFileWitnessForTest(t, target)
		if scenario == "symlink" && !bytes.Equal(got.payload, secret) {
			t.Fatalf("unsafe symlink witness payload = %q, want replacement B bytes", got.payload)
		}
	}
	if !fired {
		t.Fatal("unsafe witness did not reach the Lstat/read boundary")
	}
}

func vNextPublicationWaitForUnsafeWitnessBoundaryForTest(t *testing.T, path string, timeout time.Duration) vNextPublicationUnsafeWitnessBoundary {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		payload, err := os.ReadFile(path)
		if err == nil {
			var boundary vNextPublicationUnsafeWitnessBoundary
			if err := json.Unmarshal(payload, &boundary); err != nil {
				t.Fatalf("decode unsafe witness boundary %q: %v", payload, err)
			}
			return boundary
		}
		if !os.IsNotExist(err) {
			t.Fatalf("read unsafe witness boundary: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("unsafe witness child did not reach its parent-visible boundary %q", path)
	return vNextPublicationUnsafeWitnessBoundary{}
}
```
