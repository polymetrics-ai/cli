package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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

// Authority expectations contain explicit public heads and every private member.
// Temporary write directories are not authority and may be removed on return.
func vNextPublicationExpectedAuthority(root string) (vNextPublicationExpectedTree, error) {
	all, err := vNextPublicationObserveExpectedTree(filepath.Join(root, "acme"))
	if err != nil {
		return nil, err
	}
	authority := make(vNextPublicationExpectedTree)
	for name, member := range all {
		first := strings.Split(filepath.ToSlash(name), "/")[0]
		if first == vNextPublicationCurrentFile || first == vNextPublicationJournalFile || first == vNextPublicationControlAuthorityMarkerFile || strings.HasPrefix(first, vNextPublicationControlRepairDirectoryPrefix) {
			authority[name] = member
		}
	}
	return authority, nil
}

func vNextPublicationCompareExpectedMembers(got, want vNextPublicationExpectedTree, exact bool) error {
	if exact && len(got) != len(want) {
		return fmt.Errorf("authority membership got %d want %d", len(got), len(want))
	}
	for path, expected := range want {
		actual, found := got[path]
		if !found || actual.identity != expected.identity || !bytes.Equal(actual.payload, expected.payload) {
			return fmt.Errorf("authority %s identity/type/bytes changed", path)
		}
	}
	return nil
}

func vNextPublicationExpectedStableAuthority(root string) (vNextPublicationExpectedTree, error) {
	authority, err := vNextPublicationExpectedAuthority(root)
	if err != nil {
		return nil, err
	}
	delete(authority, vNextPublicationCurrentFile)
	delete(authority, vNextPublicationJournalFile)
	return authority, nil
}

func vNextPublicationCaptureExpectedCut(root string, prior vNextPublicationExpectedTree) (vNextPublicationExpectedTree, error) {
	authority, err := vNextPublicationExpectedStableAuthority(root)
	if err != nil {
		return nil, err
	}
	// Existing authority members are immutable. New phases may be appended, but
	// never excuse replacement or loss of an earlier prepared/anchor/capture.
	if err := vNextPublicationCompareExpectedMembers(authority, prior, false); err != nil {
		return nil, err
	}
	if err := vNextPublicationExpectedCompletedTransitions(authority, prior); err != nil {
		return nil, err
	}
	return vNextPublicationObserveExpectedTree(filepath.Join(root, "acme"))
}

func vNextPublicationExpectedPublicControls(root string) (vNextPublicationExpectedTree, error) {
	authority, err := vNextPublicationExpectedAuthority(root)
	if err != nil {
		return nil, err
	}
	controls := make(vNextPublicationExpectedTree)
	for _, name := range []string{vNextPublicationCurrentFile, vNextPublicationJournalFile} {
		if member, found := authority[name]; found {
			controls[name] = member
		}
	}
	return controls, nil
}

// Cleanup fixtures have no public-control interference before the lease/root
// cut. Every newly completed successor must therefore have the four prescribed
// phases and commit its intended selection. This independent schedule check
// supplements byte/identity preservation and the production graph decoder.
func vNextPublicationExpectedCompletedTransitions(current, prior vNextPublicationExpectedTree) error {
	for path, member := range current {
		if filepath.Base(path) != vNextPublicationControlRepairPreparedFile {
			continue
		}
		if _, existed := prior[path]; existed {
			continue
		}
		var prepared vNextPublicationControlRepair
		if err := json.Unmarshal(member.payload, &prepared); err != nil {
			return err
		}
		states := []string{"capture_intent", "captured", "selected", "terminal"}
		if prepared.Predecessor == nil {
			states = []string{"terminal"}
		}
		transaction := filepath.Dir(path)
		for i, state := range states {
			name := filepath.Join(transaction, vNextPublicationControlRepairPhaseName(i+1))
			var phase vNextPublicationControlRepairPhase
			record, found := current[name]
			if !found {
				return fmt.Errorf("expected completed cleanup transition missing %s", name)
			}
			if err := json.Unmarshal(record.payload, &phase); err != nil {
				return err
			}
			if phase.Sequence != i+1 || phase.State != state {
				return fmt.Errorf("wrong cleanup phase %s: sequence=%d state=%s", name, phase.Sequence, phase.State)
			}
			if state == "terminal" && (phase.Outcome != "committed" || phase.Selected == nil || !phase.Selected.sameLogical(prepared.Intended)) {
				return fmt.Errorf("wrong cleanup terminal selection %s", name)
			}
		}
		if _, extra := current[filepath.Join(transaction, vNextPublicationControlRepairPhaseName(len(states)+1))]; extra {
			return fmt.Errorf("unexpected extra cleanup phase in %s", transaction)
		}
	}
	return nil
}

func TestCP11ExpectedCleanupScheduleRejectsWrongPhase(t *testing.T) {
	root := t.TempDir()
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := publisher.Publish(vNextPublicationArtifactsForTest("schedule", false)); err != nil {
		t.Fatal(err)
	}
	expected, err := vNextPublicationExpectedStableAuthority(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := vNextPublicationExpectedCompletedTransitions(expected, nil); err != nil {
		t.Fatal(err)
	}
	changed := false
	for path, member := range expected {
		if filepath.Base(path) != vNextPublicationControlRepairPhaseName(3) {
			continue
		}
		var phase vNextPublicationControlRepairPhase
		if err := json.Unmarshal(member.payload, &phase); err != nil {
			t.Fatal(err)
		}
		phase.State = "captured" // readable valid enum, wrong position in actual schedule
		member.payload, err = json.Marshal(phase)
		if err != nil {
			t.Fatal(err)
		}
		expected[path] = member
		changed = true
		break
	}
	if !changed {
		t.Fatal("no actual successor third phase for negative control")
	}
	if err := vNextPublicationExpectedCompletedTransitions(expected, nil); err == nil {
		t.Fatal("accepted wrong cleanup phase")
	}
}
