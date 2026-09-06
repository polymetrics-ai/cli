package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestCP11ExpectedWitnessRejectsReplacedPriorRoot(t *testing.T) {
	root := t.TempDir()
	publisher, _ := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	old, err := publisher.Publish(vNextPublicationArtifactsForTest("old", true))
	if err != nil {
		t.Fatal(err)
	}
	held, err := publisher.Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := publisher.Publish(vNextPublicationArtifactsForTest("new", false)); err != nil {
		held.Release()
		t.Fatal(err)
	}
	held.Release()
	oldPath := filepath.Join(root, "acme", vNextPublicationGenerationDirectory, old.Generation)
	expected, err := vNextPublicationObserveExpectedTree(oldPath)
	if err != nil {
		t.Fatal(err)
	}
	before := vNextPublicationDirectoryIdentityForTest(t, oldPath, "review prior generation A")
	if err := os.Rename(oldPath, oldPath+"-A"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(oldPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldPath, "foreign"), []byte("replacement B, not generation A"), 0o600); err != nil {
		t.Fatal(err)
	}
	after := vNextPublicationDirectoryIdentityForTest(t, oldPath, "review replacement generation B")
	if before == after {
		t.Fatal("replacement control failed")
	}
	if err := vNextPublicationCompareExpectedTree(oldPath, expected); err == nil {
		t.Fatal("accepted replaced prior root B")
	}
}

// Expected trees are captured before interference, never reconstructed from the
// post-operation graph. Membership is exact; permitted transitions must replace
// explicit expected members at an independently asserted cut.
type vNextPublicationExpectedMember struct {
	identity vNextPublicationIdentity
	payload  []byte
}
type vNextPublicationExpectedTree map[string]vNextPublicationExpectedMember

func vNextPublicationObserveExpectedTree(path string) (tree vNextPublicationExpectedTree, resultErr error) {
	parent, err := vNextPublicationOpenDirectory(filepath.Dir(path), "expected tree parent")
	if err != nil {
		return nil, err
	}
	defer func() { resultErr = errors.Join(resultErr, parent.Close()) }()
	tree = make(vNextPublicationExpectedTree)
	var visit func(*vNextPublicationDirectory, string, string) error
	visit = func(parent *vNextPublicationDirectory, name, key string) (visitErr error) {
		stat, err := parent.lstat(name, "expected member")
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		if vNextPublicationStatIsDir(stat) {
			child, err := parent.openDirectory(name, "expected directory")
			if err != nil {
				return err
			}
			defer func() { visitErr = errors.Join(visitErr, child.Close()) }()
			identity, err := vNextPublicationIdentityFromFile(child.file, "expected directory")
			if err != nil {
				return err
			}
			tree[key] = vNextPublicationExpectedMember{identity: identity}
			entries, err := child.readDir()
			if err != nil {
				return err
			}
			for _, entry := range entries {
				if err := visit(child, entry.Name(), filepath.Join(key, entry.Name())); err != nil {
					return err
				}
			}
			return nil
		}
		if !vNextPublicationStatIsRegular(stat) {
			return fmt.Errorf("expected member %s is not a regular file or directory", key)
		}
		file, err := parent.openRegular(name, "expected file", unix.O_RDONLY)
		if err != nil {
			return err
		}
		identity, statErr := vNextPublicationIdentityFromFile(file, "expected file")
		payload, readErr := io.ReadAll(file)
		if err := errors.Join(statErr, readErr, file.Close()); err != nil {
			return err
		}
		tree[key] = vNextPublicationExpectedMember{identity: identity, payload: payload}
		return nil
	}
	if err := visit(parent, filepath.Base(path), "."); err != nil {
		return nil, err
	}
	return tree, nil
}

func vNextPublicationCompareExpectedTree(path string, want vNextPublicationExpectedTree) error {
	got, err := vNextPublicationObserveExpectedTree(path)
	if err != nil {
		return err
	}
	if len(got) != len(want) {
		return fmt.Errorf("%s membership changed: got %d want %d", path, len(got), len(want))
	}
	for name, expected := range want {
		actual, found := got[name]
		if !found || actual.identity != expected.identity || !bytes.Equal(actual.payload, expected.payload) {
			return fmt.Errorf("%s member %s identity/type/bytes changed", path, name)
		}
	}
	return nil
}

func vNextPublicationAssertExpectedTree(t *testing.T, path string, want vNextPublicationExpectedTree) {
	t.Helper()
	if err := vNextPublicationCompareExpectedTree(path, want); err != nil {
		t.Fatal(err)
	}
}

func TestCP11ExpectedTreeRejectsWrongReadableState(t *testing.T) {
	for _, name := range []string{"same-bytes-wrong-inode", "empty-B-replaced-by-C", "nonempty-B-replaced-by-C", "missing-history", "replaced-history", "changed-bytes", "extra-member"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "B")
			if err := os.Mkdir(path, 0o700); err != nil {
				t.Fatal(err)
			}
			history := filepath.Join(path, "prepared.json")
			if name != "empty-B-replaced-by-C" {
				if err := os.WriteFile(history, []byte("retained authority"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			want, err := vNextPublicationObserveExpectedTree(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := vNextPublicationCompareExpectedTree(path, want); err != nil {
				t.Fatalf("unchanged positive control: %v", err)
			}
			switch name {
			case "empty-B-replaced-by-C", "nonempty-B-replaced-by-C":
				if err := os.Rename(path, path+"-actual-B"); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(path, 0o700); err != nil {
					t.Fatal(err)
				}
				if name == "nonempty-B-replaced-by-C" {
					if err := os.WriteFile(history, []byte("retained authority"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
			case "same-bytes-wrong-inode", "replaced-history":
				if err := os.Rename(history, filepath.Join(root, "retained-A")); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(history, []byte("retained authority"), 0o600); err != nil {
					t.Fatal(err)
				}
			case "missing-history":
				if err := os.Remove(history); err != nil {
					t.Fatal(err)
				}
			case "changed-bytes":
				if err := os.WriteFile(history, []byte("other authority"), 0o600); err != nil {
					t.Fatal(err)
				}
			case "extra-member":
				if err := os.WriteFile(filepath.Join(path, "foreign"), nil, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if err := vNextPublicationCompareExpectedTree(path, want); err == nil {
				t.Fatal("accepted deliberately wrong readable state")
			}
		})
	}
}
