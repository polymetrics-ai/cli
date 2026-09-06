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

func vNextPublicationCaptureExpectedCut(root string, prior vNextPublicationExpectedTree, plan *vNextPublicationExpectedPlan) (vNextPublicationExpectedTree, error) {
	if plan == nil {
		return nil, fmt.Errorf("missing independent cut plan")
	}
	if err := plan.check(); err != nil {
		return nil, err
	}
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

// Final-prune verdict binds observed controls to fixture intent and the actual
// pre-write acquisition/transition plan shared by the public row and controls.
func vNextPublicationFinalPruneVerdict(root string, prior vNextPublicationExpectedTree, old vNextGenerationPointer, desired vNextPublicationArtifacts, plan *vNextPublicationExpectedPlan) (vNextPublicationExpectedTree, error) {
	cut, err := vNextPublicationCaptureExpectedCut(root, prior, plan)
	if err != nil {
		return nil, err
	}
	var current vNextGenerationPointer
	var journal vNextGenerationJournal
	if err := json.Unmarshal(cut[vNextPublicationCurrentFile].payload, &current); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(cut[vNextPublicationJournalFile].payload, &journal); err != nil {
		return nil, err
	}
	_, _, wanted, deriveErr := vNextPublicationIntegrityForFiles("acme", vNextPublicationGenerationID(desired.Files), desired.Files)
	if deriveErr != nil {
		return nil, deriveErr
	}
	if current != wanted || journal.State != "committed" || journal.New != wanted || journal.Old == nil || *journal.Old != old {
		return nil, fmt.Errorf("final-prune control mismatch")
	}
	return cut, nil
}

// The same two-root verdict serves the actual Recover/Check row and controls.
// desiredA was acquired before Recover and before the fixture renamed A.
func vNextPublicationEarlyStageVerdict(root, moved string, cut, desiredA vNextPublicationExpectedTree) error {
	if desiredA == nil {
		return fmt.Errorf("missing premove stage A witness")
	}
	if err := vNextPublicationCompareExpectedTree(moved, desiredA); err != nil {
		return err
	}
	return vNextPublicationCompareExpectedTree(filepath.Join(root, "acme"), cut)
}

func vNextPublicationExpectedOwnedStage(t *testing.T, path string, pointer vNextGenerationPointer, name, sentinel string) vNextPublicationExpectedTree {
	t.Helper()
	tree, err := vNextPublicationObserveExpectedTree(path)
	if err != nil {
		t.Fatal(err)
	}
	marker, err := vNextPublicationJSON(vNextPublicationStageOwner{Version: 1, Connector: "acme", Generation: pointer.Generation, Stage: name})
	if err != nil {
		t.Fatal(err)
	}
	if len(tree) != 3 || tree["."].identity.mode != unix.S_IFDIR || tree["sentinel.txt"].identity.mode != unix.S_IFREG || !bytes.Equal(tree["sentinel.txt"].payload, []byte(sentinel)) || tree[vNextPublicationStageOwnerFile].identity.mode != unix.S_IFREG || !bytes.Equal(tree[vNextPublicationStageOwnerFile].payload, marker) {
		t.Fatal("stage does not match declared fixture")
	}
	return tree
}

// A fixture supplies desired bytes before the operation. Runtime-created
// identities are bound only after those bytes/shape have been checked.
type vNextPublicationExpectedTransition struct {
	target   string
	intended []byte // nil is logical absence; an empty present file is not used here.
	base     bool
	phases   int // zero for interrupted preparation, four for ordinary successor.
}
type vNextPublicationExpectedPlan struct {
	tree              vNextPublicationExpectedTree
	root              string
	prior, authority  vNextPublicationExpectedTree
	controls          vNextPublicationExpectedTree
	steps             []vNextPublicationExpectedTransition
	next              int
	active            *vNextPublicationExpectedTransaction
	heads             map[string]*vNextPublicationControlRepairPredecessor
	problem           error
	fixtureFiles      map[string][]byte
	fixturePointer    vNextGenerationPointer
	generationWitness vNextPublicationExpectedTree
}
type vNextPublicationExpectedTransaction struct {
	name            string
	step            vNextPublicationExpectedTransition
	prepared        vNextPublicationControlRepair
	prior, intended vNextPublicationExpectedMember
	preparedMember  vNextPublicationExpectedMember
	last            vNextPublicationControlRepairPhaseReference
	sequence        int
	capture         *vNextPublicationControlRepairCapture
}

func vNextPublicationExpectedJSON(t *testing.T, value any) []byte {
	t.Helper()
	payload, err := vNextPublicationJSON(value)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}
func vNextPublicationFixturePointer(t *testing.T, artifacts vNextPublicationArtifacts) vNextGenerationPointer {
	t.Helper()
	files := make(map[string][]byte, len(artifacts.Files))
	for name, payload := range artifacts.Files {
		files[name] = bytes.Clone(payload)
	}
	_, _, pointer, err := vNextPublicationIntegrityForFiles("acme", vNextPublicationGenerationID(files), files)
	if err != nil {
		t.Fatal(err)
	}
	return pointer
}
func vNextPublicationCloneExpectedTree(tree vNextPublicationExpectedTree) vNextPublicationExpectedTree {
	copy := make(vNextPublicationExpectedTree, len(tree))
	for name, member := range tree {
		member.payload = bytes.Clone(member.payload)
		copy[name] = member
	}
	return copy
}

func vNextPublicationNewExpectedPlan(t *testing.T, root string, desiredControls map[string][]byte, steps ...vNextPublicationExpectedTransition) *vNextPublicationExpectedPlan {
	t.Helper()
	authority, err := vNextPublicationExpectedAuthority(root)
	if err != nil {
		t.Fatal(err)
	}
	plan := &vNextPublicationExpectedPlan{root: root, prior: vNextPublicationCloneExpectedTree(authority), authority: vNextPublicationCloneExpectedTree(authority), controls: make(vNextPublicationExpectedTree), heads: make(map[string]*vNextPublicationControlRepairPredecessor)}
	plan.tree, err = vNextPublicationObserveExpectedTree(filepath.Join(root, "acme"))
	if err != nil {
		t.Fatal(err)
	}

	for _, target := range []string{vNextPublicationCurrentFile, vNextPublicationJournalFile} {
		payload, declared := desiredControls[target]
		if !declared {
			t.Fatalf("missing declared prior %s", target)
		}
		member, present := authority[target]
		if present != (payload != nil) || (present && (member.identity.mode != unix.S_IFREG || !bytes.Equal(member.payload, payload))) {
			t.Fatalf("prior %s differs from declared fixture", target)
		}
		if present {
			plan.controls[target] = member
		}
	}
	for _, step := range steps {
		step.intended = bytes.Clone(step.intended)
		plan.steps = append(plan.steps, step)
	}
	// Existing graph is prior-state evidence, never the desired future value.
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	operation, err := publisher.openOperation(context.Background(), unix.LOCK_SH, false)
	if err != nil {
		t.Fatal(err)
	}
	graph, err := publisher.scanControlAuthorityLocked(operation)
	if err != nil {
		operation.close()
		t.Fatal(err)
	}
	for target, head := range graph.heads {
		if head == nil {
			continue
		}
		plan.heads[target] = &vNextPublicationControlRepairPredecessor{Transaction: head.state.transactionName, TransactionIdentity: vNextPublicationRecordIdentity(head.state.transactionIdentity), Terminal: vNextPublicationControlRepairPhaseReference{Name: head.terminal.name, Identity: vNextPublicationRecordIdentity(head.terminal.identity), Digest: head.terminal.digest}, Selected: head.selected}
	}
	graph.close()
	operation.close()
	return plan
}

func (plan *vNextPublicationExpectedPlan) hooks(original vNextPublicationHooks) vNextPublicationHooks {
	hooks := original
	hooks.At = func(point vNextPublicationFaultPoint) error {
		if plan.problem == nil {
			switch point {
			case vNextPublicationBeforeControlRepairRecord:
				plan.problem = plan.acquireAnchors()
			case vNextPublicationAfterControlRepairCaptureDirectory:
				plan.problem = plan.acquireCapture()
			case vNextPublicationBeforeStageRename:
				if plan.fixtureFiles != nil {
					plan.problem = plan.acquireGeneration()
				}
			}
		}
		// Oracle disagreement is recorded, not injected into the production path:
		// the existing fault must still reach its named durable cut.
		if original.At != nil {
			return original.At(point)
		}
		return nil
	}
	hooks.ControlRecord.Write = func(file *os.File, label string, payload []byte) (int, error) {
		if plan.problem == nil && label != "publication control authority marker" {
			plan.problem = plan.checkRecord(file, label, payload)
		}
		if original.ControlRecord.Write != nil {
			return original.ControlRecord.Write(file, label, payload)
		}
		return file.Write(payload)
	}
	return hooks
}

func (plan *vNextPublicationExpectedPlan) acquireAnchors() error {
	if plan.next >= len(plan.steps) {
		return fmt.Errorf("unplanned control transition")
	}
	if plan.active != nil && plan.active.sequence != plan.active.step.phases {
		return fmt.Errorf("preceding transition phase count mismatch")
	}
	all, err := vNextPublicationExpectedAuthority(plan.root)
	if err != nil {
		return err
	}
	var name string
	for path, member := range all {
		if strings.HasPrefix(path, vNextPublicationControlRepairDirectoryPrefix) && !strings.Contains(path, string(filepath.Separator)) && member.identity.mode == unix.S_IFDIR {
			if _, known := plan.authority[path]; !known {
				if name != "" {
					return fmt.Errorf("multiple new transactions")
				}
				name = path
			}
		}
	}
	if name == "" {
		return fmt.Errorf("missing newly allocated transaction")
	}
	step := plan.steps[plan.next]
	tx := &vNextPublicationExpectedTransaction{name: name, step: step}
	tx.prior = plan.controls[step.target]
	expectedMembers := 1
	for _, anchor := range []struct {
		name    string
		payload []byte
		prior   bool
	}{
		{vNextPublicationControlBackupMember, tx.prior.payload, true}, {vNextPublicationControlReplacementMember, step.intended, false},
	} {
		member, found := all[filepath.Join(name, anchor.name)]
		present := anchor.payload != nil
		if found != present || (present && (member.identity.mode != unix.S_IFREG || !bytes.Equal(member.payload, anchor.payload))) {
			return fmt.Errorf("%s %s differs from independent payload", step.target, anchor.name)
		}
		if present {
			expectedMembers++
			if (anchor.prior || step.base) && member.identity != tx.prior.identity {
				return fmt.Errorf("%s prior/base anchor identity changed", step.target)
			}
			if !anchor.prior {
				tx.intended = member
			}
		}
	}
	observed := 0
	for path := range all {
		if path == name || strings.HasPrefix(path, name+string(filepath.Separator)) {
			observed++
		}
	}
	if observed != expectedMembers {
		return fmt.Errorf("new transaction has unexpected members")
	}
	predecessor := plan.heads[step.target]
	if step.base {
		if predecessor != nil {
			return fmt.Errorf("base unexpectedly has predecessor")
		}
	} else if predecessor == nil {
		return fmt.Errorf("successor missing prior head")
	}
	priorState := vNextPublicationAbsentControlState()
	intendedState := vNextPublicationAbsentControlState()
	if tx.prior.payload != nil {
		priorState = vNextPublicationPresentControlState(vNextPublicationControlBackupMember, tx.prior.identity)
	}
	if step.intended != nil {
		intendedState = vNextPublicationPresentControlState(vNextPublicationControlReplacementMember, tx.intended.identity)
	}
	tx.prepared = vNextPublicationControlRepair{Version: vNextPublicationControlRepairVersion, Target: step.target, Transaction: name, TransactionIdentity: vNextPublicationRecordIdentity(all[name].identity), Predecessor: predecessor, Prior: priorState, Intended: intendedState, MaxCaptureAttempts: vNextPublicationControlRepairMaxCaptureAttempts}
	for path, member := range all {
		if path == name || strings.HasPrefix(path, name+string(filepath.Separator)) {
			plan.authority[path] = member
		}
	}
	plan.next++
	plan.active = tx
	return nil
}
func (plan *vNextPublicationExpectedPlan) acquireCapture() error {
	tx := plan.active
	if tx == nil || tx.step.base || tx.sequence != 0 {
		return fmt.Errorf("capture outside planned first attempt")
	}
	name := vNextPublicationControlCaptureName(1)
	path := filepath.Join(tx.name, name)
	tree, err := vNextPublicationObserveExpectedTree(filepath.Join(plan.root, "acme", path))
	if err != nil {
		return err
	}
	if len(tree) != 1 || tree["."].identity.mode != unix.S_IFDIR {
		return fmt.Errorf("capture is not the expected empty directory")
	}
	plan.authority[path] = tree["."]
	tx.capture = &vNextPublicationControlRepairCapture{Attempt: 1, Name: name, Identity: vNextPublicationRecordIdentity(tree["."].identity)}
	return nil
}
func (plan *vNextPublicationExpectedPlan) checkRecord(file *os.File, label string, payload []byte) error {
	tx := plan.active
	if tx == nil {
		return fmt.Errorf("record outside planned transaction: %s", label)
	}
	var expected any
	name := vNextPublicationControlRepairPreparedFile
	switch label {
	case "publication control repair prepared authority":
		expected = tx.prepared
	case "publication control repair phase":
		sequence := tx.sequence + 1
		states := []string{"capture_intent", "captured", "selected", "terminal"}
		if tx.step.base {
			states = []string{"terminal"}
		}
		if sequence > len(states) || sequence > tx.step.phases {
			return fmt.Errorf("unplanned phase %d", sequence)
		}
		phase := vNextPublicationControlRepairPhase{Version: vNextPublicationControlRepairVersion, Sequence: sequence, State: states[sequence-1], PreparedDigest: vNextPublicationDigest(tx.preparedMember.payload), PreparedIdentity: vNextPublicationRecordIdentity(tx.preparedMember.identity), Previous: tx.last}
		switch phase.State {
		case "capture_intent":
			if tx.capture == nil {
				return fmt.Errorf("capture identity not acquired")
			}
			phase.Capture = tx.capture
		case "captured":
			phase.Capture = tx.capture
			candidate := vNextPublicationControlStateWithMember(tx.prepared.Prior, vNextPublicationControlCaptureMember)
			phase.Candidate = &candidate
			if candidate.Present {
				plan.authority[filepath.Join(tx.name, tx.capture.Name, vNextPublicationControlCaptureMember)] = tx.prior
			}
		case "selected", "terminal":
			selected := tx.prepared.Intended
			phase.Selected = &selected
			if phase.State == "terminal" {
				phase.Outcome = "committed"
			}
		}
		expected = phase
		name = vNextPublicationControlRepairPhaseName(sequence)
	default:
		return fmt.Errorf("unplanned record label %s", label)
	}
	wanted, err := vNextPublicationJSON(expected)
	if err != nil {
		return err
	}
	if !bytes.Equal(payload, wanted) {
		return fmt.Errorf("%s %s differs from planned record", tx.step.target, name)
	}
	identity, err := vNextPublicationIdentityFromFile(file, "expected newly written record")
	if err != nil {
		return err
	}
	member := vNextPublicationExpectedMember{identity: identity, payload: bytes.Clone(wanted)}
	plan.authority[filepath.Join(tx.name, name)] = member
	reference := vNextPublicationControlRepairPhaseReference{Name: name, Identity: vNextPublicationRecordIdentity(identity), Digest: vNextPublicationDigest(wanted)}
	if name == vNextPublicationControlRepairPreparedFile {
		tx.preparedMember = member
	} else {
		tx.sequence++
		if tx.sequence == tx.step.phases {
			if tx.step.intended == nil {
				delete(plan.controls, tx.step.target)
				delete(plan.authority, tx.step.target)
			} else {
				plan.controls[tx.step.target] = tx.intended
				plan.authority[tx.step.target] = tx.intended
			}
			plan.heads[tx.step.target] = &vNextPublicationControlRepairPredecessor{Transaction: tx.name, TransactionIdentity: tx.prepared.TransactionIdentity, Terminal: reference, Selected: tx.prepared.Intended}
		}
	}
	tx.last = reference
	return nil
}
func (plan *vNextPublicationExpectedPlan) check() error {
	if plan.fixtureFiles != nil && plan.problem == nil {
		if plan.generationWitness == nil {
			return fmt.Errorf("missing validated new generation acquisition")
		}
		if err := vNextPublicationCompareExpectedTree(filepath.Join(plan.root, "acme", vNextPublicationGenerationDirectory, plan.fixturePointer.Generation), plan.generationWitness); err != nil {
			return err
		}
	}
	if plan.problem != nil {
		return plan.problem
	}
	if plan.next != len(plan.steps) {
		return fmt.Errorf("planned transitions got %d want %d", plan.next, len(plan.steps))
	}
	if plan.active != nil && plan.active.sequence != plan.active.step.phases {
		return fmt.Errorf("final planned phase count mismatch")
	}
	actual, err := vNextPublicationExpectedAuthority(plan.root)
	if err != nil {
		return err
	}
	if err := vNextPublicationCompareExpectedMembers(actual, plan.authority, true); err != nil {
		return err
	}
	expected := vNextPublicationCloneExpectedTree(plan.tree)
	for key := range expected {
		first := strings.Split(filepath.ToSlash(key), "/")[0]
		if first == vNextPublicationCurrentFile || first == vNextPublicationJournalFile || first == vNextPublicationControlAuthorityMarkerFile || strings.HasPrefix(first, vNextPublicationControlRepairDirectoryPrefix) {
			delete(expected, key)
		}
	}
	for key, member := range plan.authority {
		expected[key] = member
	}
	return vNextPublicationCompareExpectedTree(filepath.Join(plan.root, "acme"), expected)
}

func (plan *vNextPublicationExpectedPlan) watchGeneration(t *testing.T, artifacts vNextPublicationArtifacts) {
	t.Helper()
	plan.fixtureFiles = make(map[string][]byte, len(artifacts.Files))
	for name, payload := range artifacts.Files {
		plan.fixtureFiles[name] = bytes.Clone(payload)
	}
	plan.fixturePointer = vNextPublicationFixturePointer(t, artifacts)
}
func (plan *vNextPublicationExpectedPlan) acquireGeneration() error {
	if plan.generationWitness != nil {
		return fmt.Errorf("extra generation acquisition")
	}
	base := filepath.Join(plan.root, "acme", vNextPublicationGenerationDirectory)
	tree, err := vNextPublicationObserveExpectedTree(base)
	if err != nil {
		return err
	}
	name := ""
	for path, member := range tree {
		if !strings.Contains(path, string(filepath.Separator)) && strings.HasPrefix(path, ".stage-") && member.identity.mode == unix.S_IFDIR {
			if name != "" {
				return fmt.Errorf("multiple staged generations")
			}
			name = path
		}
	}
	if name == "" {
		return fmt.Errorf("missing expected staged generation")
	}
	files := make(map[string][]byte, len(plan.fixtureFiles)+3)
	for path, payload := range plan.fixtureFiles {
		files[path] = payload
	}
	_, integrity, _, err := vNextPublicationIntegrityForFiles("acme", plan.fixturePointer.Generation, plan.fixtureFiles)
	if err != nil {
		return err
	}
	files[vNextPublicationIntegrityFile] = integrity
	files[vNextPublicationLeaseFile] = []byte{}
	marker, err := vNextPublicationJSON(vNextPublicationStageOwner{Version: 1, Connector: "acme", Generation: plan.fixturePointer.Generation, Stage: name})
	if err != nil {
		return err
	}
	files[vNextPublicationStageOwnerFile] = marker
	expected := make(vNextPublicationExpectedTree)
	allowed := map[string]bool{".": true}
	for path, payload := range files {
		member, found := tree[filepath.Join(name, path)]
		if !found || member.identity.mode != unix.S_IFREG || !bytes.Equal(member.payload, payload) {
			return fmt.Errorf("staged %s differs from fixture", path)
		}
		expected[path] = member
		allowed[path] = true
		for parent := filepath.Dir(path); parent != "."; parent = filepath.Dir(parent) {
			allowed[parent] = true
		}
	}
	for path, member := range tree {
		if path != name && !strings.HasPrefix(path, name+string(filepath.Separator)) {
			continue
		}
		relative, err := filepath.Rel(name, path)
		if err != nil {
			return err
		}
		if !allowed[relative] {
			return fmt.Errorf("unexpected staged member %s", relative)
		}
		if _, file := files[relative]; !file {
			if member.identity.mode != unix.S_IFDIR {
				return fmt.Errorf("staged directory type %s", relative)
			}
			expected[relative] = member
		}
	}
	if len(expected) != len(allowed) {
		return fmt.Errorf("missing staged directory")
	}
	plan.generationWitness = expected
	for relative, member := range expected {
		plan.tree[filepath.Join(vNextPublicationGenerationDirectory, plan.fixturePointer.Generation, relative)] = member
	}
	return nil
}

// The actual actor-created lease B and displaced A are an explicit fixture
// delta. Only that named member is replaced; other generation members stay bound.
func (plan *vNextPublicationExpectedPlan) leaseDelta(t *testing.T, generation, lease, retained string) {
	t.Helper()
	connector := filepath.Join(plan.root, "acme")
	key := filepath.Join(vNextPublicationGenerationDirectory, generation, vNextPublicationLeaseFile)
	before, found := plan.tree[key]
	if !found {
		t.Fatal("lease A absent from independently expected tree")
	}
	moved, err := vNextPublicationObserveExpectedTree(retained)
	if err != nil {
		t.Fatal(err)
	}
	if err := vNextPublicationCompareExpectedMembers(moved, vNextPublicationExpectedTree{".": before}, true); err != nil {
		t.Fatal(err)
	}
	actual, err := vNextPublicationObserveExpectedTree(lease)
	if err != nil {
		t.Fatal(err)
	}
	if actual["."].identity.mode != unix.S_IFREG || actual["."].identity == before.identity {
		t.Fatal("invalid actor-created lease B")
	}
	plan.tree[key] = actual["."]
	relative, err := filepath.Rel(connector, retained)
	if err != nil {
		t.Fatal(err)
	}
	if relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		plan.tree[relative] = before
	}
	if generation == plan.fixturePointer.Generation {
		plan.generationWitness[vNextPublicationLeaseFile] = actual["."]
		relative, err = filepath.Rel(filepath.Join(connector, vNextPublicationGenerationDirectory, generation), retained)
		if err != nil {
			t.Fatal(err)
		}
		if relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			plan.generationWitness[relative] = before
		}
	}
}

func (plan *vNextPublicationExpectedPlan) stageDelta(t *testing.T, from, moved string, replacement vNextPublicationExpectedTree) {
	t.Helper()
	connector := filepath.Join(plan.root, "acme")
	fromKey, err := filepath.Rel(connector, from)
	if err != nil {
		t.Fatal(err)
	}
	movedKey, err := filepath.Rel(connector, moved)
	if err != nil {
		t.Fatal(err)
	}
	original := make(vNextPublicationExpectedTree)
	for key, member := range plan.tree {
		if key == fromKey || strings.HasPrefix(key, fromKey+string(filepath.Separator)) {
			relative, err := filepath.Rel(fromKey, key)
			if err != nil {
				t.Fatal(err)
			}
			original[relative] = member
			delete(plan.tree, key)
		}
	}
	if len(original) == 0 {
		t.Fatal("stage A missing from independent prior tree")
	}
	if err := vNextPublicationCompareExpectedTree(moved, original); err != nil {
		t.Fatal(err)
	}
	for key, member := range replacement {
		plan.tree[filepath.Join(fromKey, key)] = member
	}
	if movedKey != ".." && !strings.HasPrefix(movedKey, ".."+string(filepath.Separator)) {
		for key, member := range original {
			plan.tree[filepath.Join(movedKey, key)] = member
		}
	}
}

// The held-reader Publish row explicitly permits pruning its known unheld B.
func (plan *vNextPublicationExpectedPlan) expectPrunedGeneration(generation string) {
	prefix := filepath.Join(vNextPublicationGenerationDirectory, generation)
	for key := range plan.tree {
		if key == prefix || strings.HasPrefix(key, prefix+string(filepath.Separator)) {
			delete(plan.tree, key)
		}
	}
}

// Each schedule is the accepted concrete Publish cut, not an observed count.
func vNextPublicationExpectedPublish(t *testing.T, root string, oldFiles, newFiles vNextPublicationArtifacts, cut string) (*vNextPublicationExpectedPlan, vNextGenerationPointer) {
	t.Helper()
	old := vNextPublicationFixturePointer(t, oldFiles)
	next := vNextPublicationFixturePointer(t, newFiles)
	prepared := vNextPublicationExpectedJSON(t, vNextGenerationJournal{Old: &old, New: next, State: "prepared"})
	committed := vNextPublicationExpectedJSON(t, vNextGenerationJournal{Old: &old, New: next, State: "committed"})
	steps := []vNextPublicationExpectedTransition{{target: vNextPublicationJournalFile, intended: prepared, phases: 4}}
	switch cut {
	case "rejected":
	case "prepared":
		steps = append(steps, vNextPublicationExpectedTransition{target: vNextPublicationCurrentFile, intended: vNextPublicationExpectedJSON(t, next), phases: 4})
	case "committed", "complete":
		steps = append(steps, vNextPublicationExpectedTransition{target: vNextPublicationCurrentFile, intended: vNextPublicationExpectedJSON(t, next), phases: 4}, vNextPublicationExpectedTransition{target: vNextPublicationJournalFile, intended: committed, phases: 4})
		if cut == "complete" {
			steps = append(steps, vNextPublicationExpectedTransition{target: vNextPublicationJournalFile, intended: nil, phases: 4})
		}
	case "rollback":
		steps = append(steps, vNextPublicationExpectedTransition{target: vNextPublicationCurrentFile, intended: vNextPublicationExpectedJSON(t, next), phases: 4}, vNextPublicationExpectedTransition{target: vNextPublicationCurrentFile, intended: vNextPublicationExpectedJSON(t, old), phases: 4})
	default:
		t.Fatalf("unknown fixture cut %q", cut)
	}
	plan := vNextPublicationNewExpectedPlan(t, root, map[string][]byte{vNextPublicationCurrentFile: vNextPublicationExpectedJSON(t, old), vNextPublicationJournalFile: nil}, steps...)
	plan.watchGeneration(t, newFiles)
	return plan, next
}
