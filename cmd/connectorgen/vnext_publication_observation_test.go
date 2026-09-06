package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// vNextPublicationSnapshotAfterObservationForTest is a nil-by-default test
// seam at the only classification-to-open boundary in the snapshot oracle.
// Production code neither declares nor observes it.
var vNextPublicationSnapshotAfterObservationForTest func(*vNextPublicationDirectory, string, vNextPublicationIdentity)

func TestVNextPublicationTreeSnapshotRefusesReplacementChild(t *testing.T) {
	mode := os.Getenv("CP11_F04_GREEN_CHILD")
	if mode == "" {
		return
	}
	root := t.TempDir()
	target := filepath.Join(root, "target")
	secret := []byte("replacement-B bytes must never enter the safe snapshot")
	switch mode {
	case "fifo", "symlink":
		if err := os.WriteFile(target, []byte("original-A"), 0o600); err != nil {
			t.Fatal(err)
		}
	case "directory":
		if err := os.Mkdir(target, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(target, "a.json"), []byte("original-A"), 0o600); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatalf("unknown safe snapshot child mode %q", mode)
	}

	type witness struct {
		AInode uint64 `json:"a_inode"`
		BInode uint64 `json:"b_inode"`
		Mode   string `json:"mode"`
	}
	fired := false
	vNextPublicationSnapshotAfterObservationForTest = func(directory *vNextPublicationDirectory, name string, originalA vNextPublicationIdentity) {
		if name != "target" || fired {
			return
		}
		fired = true
		switch mode {
		case "fifo":
			if err := os.Remove(target); err != nil {
				t.Fatal(err)
			}
			if err := unix.Mkfifo(target, 0o600); err != nil {
				t.Fatal(err)
			}
		case "symlink":
			external := filepath.Join(t.TempDir(), "replacement-B")
			if err := os.WriteFile(external, secret, 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(target); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(external, target); err != nil {
				t.Fatal(err)
			}
		case "directory":
			if err := os.Rename(target, target+"-A"); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(target, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(target, "b.json"), secret, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		replacementB, err := directory.identityAt(name, "replacement B")
		if err != nil {
			t.Fatal(err)
		}
		if originalA == replacementB {
			t.Fatal("fixture replacement did not change identity")
		}
		witnessPath := os.Getenv("CP11_F04_GREEN_WITNESS")
		if witnessPath == "" {
			t.Fatal("missing safe snapshot witness path")
		}
		payload, err := json.Marshal(witness{AInode: originalA.inode, BInode: replacementB.inode, Mode: mode})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(witnessPath, payload, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { vNextPublicationSnapshotAfterObservationForTest = nil })

	// The fixed oracle must terminate the child with a boundary refusal. A
	// return would mean it accepted a substituted B after recording A.
	vNextPublicationTreeSnapshotForTest(t, root)
	if !fired {
		t.Fatal("safe snapshot seam was not reached")
	}
	t.Fatal("safe snapshot accepted replacement B")
}

func TestVNextPublicationTreeSnapshotRefusesAtoBReplacement(t *testing.T) {
	for _, mode := range []string{"fifo", "symlink", "directory"} {
		t.Run(mode, func(t *testing.T) {
			witnessPath := filepath.Join(t.TempDir(), "safe-snapshot-boundary")
			command := exec.Command(os.Args[0], "-test.run=^TestVNextPublicationTreeSnapshotRefusesReplacementChild$", "-test.v")
			command.Env = append(os.Environ(), "CP11_F04_GREEN_CHILD="+mode, "CP11_F04_GREEN_WITNESS="+witnessPath)
			var output bytes.Buffer
			command.Stdout = &output
			command.Stderr = &output
			child := vNextPublicationStartBoundedChildForTest(t, command, "safe snapshot replacement child")
			deadline := time.Now().Add(time.Second)
			for time.Now().Before(deadline) {
				payload, err := os.ReadFile(witnessPath)
				if err == nil {
					var actual struct {
						AInode uint64 `json:"a_inode"`
						BInode uint64 `json:"b_inode"`
						Mode   string `json:"mode"`
					}
					if err := json.Unmarshal(payload, &actual); err != nil || actual.Mode != mode || actual.AInode == actual.BInode {
						t.Fatalf("invalid safe snapshot boundary witness %q: %#v %v", payload, actual, err)
					}
					break
				}
				if !os.IsNotExist(err) {
					t.Fatalf("read safe snapshot witness: %v", err)
				}
				time.Sleep(10 * time.Millisecond)
			}
			if _, err := os.Stat(witnessPath); err != nil {
				t.Fatalf("safe snapshot child did not reach A/B boundary: %v", err)
			}
			t.Logf("safe snapshot %s witness recorded distinct A/B identities before its descriptor-bound open", mode)
			err, completed := child.waitWithin(time.Second)
			if !completed {
				child.killAndWait(t, "safe snapshot replacement timeout")
				t.Fatalf("safe snapshot %s replacement child did not refuse; output=%s", mode, output.String())
			}
			if err == nil || bytes.Contains(output.Bytes(), []byte("replacement-B bytes")) {
				t.Fatalf("safe snapshot %s replacement was not refused before reading B: err=%v output=%s", mode, err, output.String())
			}
			switch mode {
			case "fifo":
				if !strings.Contains(output.String(), "not a regular file") {
					t.Fatalf("safe FIFO refusal lacks regular-file boundary: %s", output.String())
				}
			case "symlink":
				if !strings.Contains(output.String(), "snapshot file target") {
					t.Fatalf("safe symlink refusal lacks no-follow file boundary: %s", output.String())
				}
			case "directory":
				if !strings.Contains(output.String(), "snapshot directory \"target\" identity changed") {
					t.Fatalf("safe directory refusal lacks descriptor identity boundary: %s", output.String())
				}
			}
		})
	}
}

func TestVNextPublicationTreeSnapshotRetainsNestedRegularBytes(t *testing.T) {
	root := t.TempDir()
	payload := []byte(`{"nested":true}`)
	if err := os.Mkdir(filepath.Join(root, "nested"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "nested", "schema.json"), payload, 0o600); err != nil {
		t.Fatal(err)
	}
	var entries []struct {
		Path    string `json:"path"`
		Payload []byte `json:"payload"`
	}
	if err := json.Unmarshal(vNextPublicationTreeSnapshotForTest(t, root), &entries); err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Path == filepath.Join("nested", "schema.json") && bytes.Equal(entry.Payload, payload) {
			return
		}
	}
	t.Fatal("snapshot omitted nested regular-file bytes")
}
