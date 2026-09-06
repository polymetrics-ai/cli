package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

func TestCP11F03ARepairPostRecordFailureRetainsCoherentAuthority(t *testing.T) {
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	oldArtifacts := vNextPublicationArtifactsForTest("old", true)
	if _, err := baseline.Publish(oldArtifacts); err != nil {
		t.Fatal(err)
	}
	injected := errors.New("injected post-record preparation failure")
	fired := false
	writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point == vNextPublicationAfterControlRepairRecord && !fired {
			fired = true
			return injected
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Publish(vNextPublicationArtifactsForTest("new", true)); !errors.Is(err, injected) {
		t.Fatalf("Publish post-record failure = %v, want injected cause", err)
	}
	if !fired {
		t.Fatal("post-record failure hook did not fire")
	}
	stranded := false
	transactions, err := filepath.Glob(filepath.Join(root, "acme", vNextPublicationControlRepairDirectoryPrefix+"*"))
	if err != nil {
		t.Fatal(err)
	}
	for _, transaction := range transactions {
		entries, err := os.ReadDir(transaction)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) == 1 && entries[0].Name() == vNextPublicationControlRepairPreparedFile {
			stranded = true
		}
	}
	if stranded {
		t.Fatal("post-record failure stranded prepared.json after its anchors were removed")
	}
	fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.Recover(context.Background()); err != nil {
		t.Fatalf("fresh Recover after post-record failure = %v, want coherent recovery", err)
	}
	if err := fresh.Check(oldArtifacts); err != nil {
		t.Fatalf("fresh Check after post-record recovery = %v", err)
	}
	if _, err := fresh.Publish(vNextPublicationArtifactsForTest("retry", true)); err != nil {
		t.Fatalf("ordinary retry after post-record recovery = %v", err)
	}
}

func TestCP11F03ARepairPreparationFrontierMatrix(t *testing.T) {
	for _, target := range []struct {
		name     string
		control  string
		prior    bool
		intended bool
		prepare  func(*vNextGenerationPublisher, *vNextPublicationOperation, vNextGenerationPointer) error
		write    func(*vNextGenerationPublisher, *vNextPublicationOperation, vNextGenerationPointer) error
	}{
		{name: "JOURNAL-absent-to-present", control: vNextPublicationJournalFile, prior: false, intended: true, write: func(p *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
			return p.writeJournalLocked(operation, vNextGenerationJournal{New: pointer, State: "prepared"})
		}},
		{name: "CURRENT-present-to-present", control: vNextPublicationCurrentFile, prior: true, intended: true, write: func(p *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
			return p.writeCurrentLocked(operation, pointer)
		}},
		{name: "JOURNAL-present-to-absent", control: vNextPublicationJournalFile, prior: true, intended: false, prepare: func(p *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
			return p.writeJournalLocked(operation, vNextGenerationJournal{New: pointer, State: "prepared"})
		}, write: func(p *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
			return p.removeJournalLocked(operation)
		}},
	} {
		for _, frontier := range []struct {
			name      string
			keepGraph bool
			hooks     func(error) vNextPublicationHooks
		}{
			{name: "before-record-operation-failure", hooks: func(injected error) vNextPublicationHooks {
				return vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
					if point == vNextPublicationBeforeControlRepairRecord {
						return injected
					}
					return nil
				}}
			}},
			{name: "record-short-write", hooks: func(injected error) vNextPublicationHooks {
				return vNextPublicationHooks{ControlRecord: vNextPublicationControlRecordHooks{Write: func(file *os.File, label string, payload []byte) (int, error) {
					if label != "publication control repair prepared authority" {
						return file.Write(payload)
					}
					if len(payload) < 2 {
						t.Fatalf("prepared record payload too short for actual short-write control")
					}
					return file.Write(payload[:len(payload)-1])
				}}}
			}},
			{name: "record-sync-completion", keepGraph: true, hooks: func(injected error) vNextPublicationHooks {
				return vNextPublicationHooks{ControlRecord: vNextPublicationControlRecordHooks{Sync: func(file *os.File, label string) error {
					if err := file.Sync(); err != nil {
						return err
					}
					if label == "publication control repair prepared authority" {
						return injected
					}
					return nil
				}}}
			}},
			{name: "record-close-completion", keepGraph: true, hooks: func(injected error) vNextPublicationHooks {
				return vNextPublicationHooks{ControlRecord: vNextPublicationControlRecordHooks{Close: func(file *os.File, label string) error {
					if err := file.Close(); err != nil {
						return err
					}
					if label == "publication control repair prepared authority" {
						return injected
					}
					return nil
				}}}
			}},
			{name: "post-record", keepGraph: true, hooks: func(injected error) vNextPublicationHooks {
				return vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
					if point == vNextPublicationAfterControlRepairRecord {
						return injected
					}
					return nil
				}}
			}},
			{name: "post-transaction-sync", keepGraph: true, hooks: func(injected error) vNextPublicationHooks {
				return vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
					if point == vNextPublicationAfterControlRepairTransactionSync {
						return injected
					}
					return nil
				}}
			}},
			{name: "post-connector-sync", keepGraph: true, hooks: func(injected error) vNextPublicationHooks {
				return vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
					if point == vNextPublicationAfterControlRepairConnectorSync {
						return injected
					}
					return nil
				}}
			}},
		} {
			t.Run(target.name+"/"+frontier.name, func(t *testing.T) {
				root := t.TempDir()
				artifacts := vNextPublicationArtifactsForTest("old", true)
				baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
				if err != nil {
					t.Fatal(err)
				}
				old, err := baseline.Publish(artifacts)
				if err != nil {
					t.Fatal(err)
				}
				if target.prepare != nil {
					operation, err := baseline.openOperation(context.Background(), unix.LOCK_EX, true)
					if err != nil {
						t.Fatal(err)
					}
					err = target.prepare(baseline, operation, old)
					operation.close()
					if err != nil {
						t.Fatalf("prepare %s predecessor: %v", target.name, err)
					}
				}
				knownTransactions := vNextPublicationRepairTransactionsForTest(t, filepath.Join(root, "acme"))
				injected := errors.New("injected " + frontier.name)
				hooks := frontier.hooks(injected)
				var beforeRecord vNextPublicationExpectedTree
				originalAt := hooks.At
				hooks.At = func(point vNextPublicationFaultPoint) error {
					if point == vNextPublicationBeforeControlRepairRecord {
						var err error
						beforeRecord, err = vNextPublicationExpectedAuthority(root)
						if err != nil {
							return err
						}
					}
					if originalAt != nil {
						return originalAt(point)
					}
					return nil
				}
				writer, err := newVNextGenerationPublisher(root, "acme", hooks)
				if err != nil {
					t.Fatal(err)
				}
				operation, err := writer.openOperation(context.Background(), unix.LOCK_EX, true)
				if err != nil {
					t.Fatal(err)
				}
				err = target.write(writer, operation, old)
				operation.close()
				if !errors.Is(err, injected) && frontier.name != "record-short-write" {
					t.Fatalf("%s %s error = %v, want injected cause", target.name, frontier.name, err)
				}
				if frontier.name == "record-short-write" && (err == nil || !strings.Contains(err.Error(), "short write")) {
					t.Fatalf("%s short-write error = %v, want actual short write", target.name, err)
				}

				fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
				if err != nil {
					t.Fatal(err)
				}
				connectorRoot := filepath.Join(root, "acme")
				if frontier.keepGraph {
					afterRecord, err := vNextPublicationExpectedAuthority(root)
					if err != nil {
						t.Fatal(err)
					}
					if beforeRecord == nil {
						t.Fatal("missing pre-record anchor witness")
					}
					if err := vNextPublicationCompareExpectedMembers(afterRecord, beforeRecord, false); err != nil {
						t.Fatalf("pre-record controls/anchors changed: %v", err)
					}
					vNextPublicationAssertPendingPreparedGraphForTest(t, fresh, connectorRoot, target.control, target.prior, target.intended, knownTransactions)
					beforeCheck := vNextPublicationTreeSnapshotForTest(t, connectorRoot)
					if err := fresh.Check(artifacts); err == nil {
						t.Fatalf("pending %s %s Check succeeded", target.name, frontier.name)
					}
					if afterCheck := vNextPublicationTreeSnapshotForTest(t, connectorRoot); !bytes.Equal(beforeCheck, afterCheck) {
						t.Fatalf("pending %s %s Check mutated the authority graph", target.name, frontier.name)
					}
				} else {
					vNextPublicationAssertNoPreparedGraphForTest(t, connectorRoot, knownTransactions)
				}
				if err := fresh.Recover(context.Background()); err != nil {
					t.Fatalf("fresh Recover after %s %s = %v", target.name, frontier.name, err)
				}
				if err := fresh.Check(artifacts); err != nil {
					t.Fatalf("fresh Check after %s %s recovery = %v", target.name, frontier.name, err)
				}
				if _, err := fresh.Publish(vNextPublicationArtifactsForTest("retry-"+frontier.name, true)); err != nil {
					t.Fatalf("fresh retry after %s %s = %v", target.name, frontier.name, err)
				}
			})
		}
	}
}

func TestCP11F03ARepairBasePresentPresentAuthorityRecovers(t *testing.T) {
	for _, target := range []string{vNextPublicationCurrentFile, vNextPublicationJournalFile} {
		t.Run(target, func(t *testing.T) {
			root := t.TempDir()
			artifacts := vNextPublicationArtifactsForTest("old", true)
			baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			old, err := baseline.Publish(artifacts)
			if err != nil {
				t.Fatal(err)
			}
			if target == vNextPublicationJournalFile {
				operation, err := baseline.openOperation(context.Background(), unix.LOCK_EX, true)
				if err != nil {
					t.Fatal(err)
				}
				err = baseline.writeJournalLocked(operation, vNextGenerationJournal{New: old, State: "prepared"})
				operation.close()
				if err != nil {
					t.Fatalf("prepare present JOURNAL control: %v", err)
				}
			}
			connectorRoot := filepath.Join(root, "acme")
			// Retain the published generation and selected public control, but
			// remove historical authority to construct the marker-missing
			// bootstrap state whose observed target (and therefore prior and
			// intended anchors) is present. This is fixture-only in the isolated
			// temp root.
			vNextPublicationRemoveAuthorityGraphForTest(t, connectorRoot)

			crash := errors.New("crash after durable present-present base authority")
			writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
				if point == vNextPublicationAfterBaseControlRepairPrepared {
					return crash
				}
				return nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			operation, err := writer.openOperation(context.Background(), unix.LOCK_EX, true)
			if err != nil {
				t.Fatal(err)
			}
			_, err = writer.createBaseControlAuthorityLocked(operation, target)
			operation.close()
			if !errors.Is(err, crash) {
				t.Fatalf("create %s present-present base authority = %v, want injected interruption", target, err)
			}

			fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			vNextPublicationAssertPendingPreparedGraphForTest(t, fresh, connectorRoot, target, true, true, map[string]vNextPublicationExpectedTree{})
			beforeCheck := vNextPublicationTreeSnapshotForTest(t, connectorRoot)
			if err := fresh.Check(artifacts); err == nil {
				t.Fatalf("Check() during %s present-present base interruption succeeded", target)
			}
			if afterCheck := vNextPublicationTreeSnapshotForTest(t, connectorRoot); !bytes.Equal(beforeCheck, afterCheck) {
				t.Fatalf("Check() changed %s present-present interrupted authority state", target)
			}
			if err := fresh.Recover(context.Background()); err != nil {
				t.Fatalf("fresh Recover() after %s present-present base interruption = %v", target, err)
			}
			if err := fresh.Check(artifacts); err != nil {
				t.Fatalf("fresh Check() after %s present-present base interruption = %v", target, err)
			}
			if _, err := fresh.Publish(vNextPublicationArtifactsForTest("retry-"+target, true)); err != nil {
				t.Fatalf("retry after %s present-present base interruption = %v", target, err)
			}
		})
	}
}

func vNextPublicationRemoveAuthorityGraphForTest(t *testing.T, connectorRoot string) {
	t.Helper()
	marker := filepath.Join(connectorRoot, vNextPublicationControlAuthorityMarkerFile)
	if err := os.Remove(marker); err != nil {
		t.Fatalf("remove fixture authority marker: %v", err)
	}
	for name := range vNextPublicationRepairTransactionsForTest(t, connectorRoot) {
		if err := os.RemoveAll(filepath.Join(connectorRoot, name)); err != nil {
			t.Fatalf("remove fixture authority transaction %q: %v", name, err)
		}
	}
}

func vNextPublicationRepairTransactionsForTest(t *testing.T, connectorRoot string) map[string]vNextPublicationExpectedTree {
	t.Helper()
	transactions, err := filepath.Glob(filepath.Join(connectorRoot, vNextPublicationControlRepairDirectoryPrefix+"*"))
	if err != nil {
		t.Fatal(err)
	}
	result := make(map[string]vNextPublicationExpectedTree, len(transactions))
	for _, transaction := range transactions {
		snapshot, err := vNextPublicationObserveExpectedTree(transaction)
		if err != nil {
			t.Fatal(err)
		}
		result[filepath.Base(transaction)] = snapshot
	}
	return result
}

func vNextPublicationAssertNoPreparedGraphForTest(t *testing.T, connectorRoot string, knownTransactions map[string]vNextPublicationExpectedTree) {
	t.Helper()
	for name, expected := range knownTransactions {
		vNextPublicationAssertExpectedTree(t, filepath.Join(connectorRoot, name), expected)
	}
	for name := range vNextPublicationRepairTransactionsForTest(t, connectorRoot) {
		if _, existed := knownTransactions[name]; existed {
			continue
		}
		transaction := filepath.Join(connectorRoot, name)
		for _, member := range []string{vNextPublicationControlRepairPreparedFile, vNextPublicationControlBackupMember, vNextPublicationControlReplacementMember} {
			if _, err := os.Lstat(filepath.Join(transaction, member)); !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("unexposed repair retained %q in %q: %v", member, transaction, err)
			}
		}
	}
}

func vNextPublicationAssertPendingPreparedGraphForTest(t *testing.T, publisher *vNextGenerationPublisher, connectorRoot, target string, wantPriorPresent, wantIntendedPresent bool, knownTransactions map[string]vNextPublicationExpectedTree) {
	t.Helper()
	for name, expected := range knownTransactions {
		vNextPublicationAssertExpectedTree(t, filepath.Join(connectorRoot, name), expected)
	}
	operation, err := publisher.openOperation(context.Background(), unix.LOCK_SH, false)
	if err != nil {
		t.Fatal(err)
	}
	defer operation.close()
	graph, err := publisher.scanControlAuthorityLocked(operation)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.close()
	var state *vNextPublicationControlRepairState
	for name, candidate := range graph.states {
		if _, existed := knownTransactions[name]; !existed && candidate.record.Target == target && candidate.preparedIdentity.mode == unix.S_IFREG && len(candidate.phases) == 0 {
			state = candidate
			break
		}
	}
	if state == nil {
		t.Fatalf("pending %s authority graph has no new prepared phase-free state", target)
	}
	priorIdentity, priorPresent, err := state.record.Prior.identity(vNextPublicationControlBackupMember)
	if err != nil {
		t.Fatal(err)
	}
	intendedIdentity, intendedPresent, err := state.record.Intended.identity(vNextPublicationControlReplacementMember)
	if err != nil {
		t.Fatal(err)
	}
	if priorPresent != wantPriorPresent || intendedPresent != wantIntendedPresent {
		t.Fatalf("pending %s repair states = prior-present:%t intended-present:%t, want prior-present:%t intended-present:%t", target, priorPresent, intendedPresent, wantPriorPresent, wantIntendedPresent)
	}
	transaction, err := state.openTransaction(operation)
	if err != nil {
		t.Fatal(err)
	}
	if priorPresent {
		actual, found, anchorErr := state.anchor(transaction, state.record.Prior, vNextPublicationControlBackupMember)
		if anchorErr != nil || !found || actual != priorIdentity {
			_ = transaction.Close()
			t.Fatalf("pending %s prior anchor = identity=%#v found=%t err=%v, want %#v", target, actual, found, anchorErr, priorIdentity)
		}
	}
	if intendedPresent {
		actual, found, anchorErr := state.anchor(transaction, state.record.Intended, vNextPublicationControlReplacementMember)
		if anchorErr != nil || !found || actual != intendedIdentity {
			_ = transaction.Close()
			t.Fatalf("pending %s intended anchor = identity=%#v found=%t err=%v, want %#v", target, actual, found, anchorErr, intendedIdentity)
		}
	}
	if err := transaction.Close(); err != nil {
		t.Fatal(err)
	}
	_, publicFound, publicIdentity, err := vNextPublicationReadControlBound(operation.connector, target, "pending public control")
	if err != nil {
		t.Fatal(err)
	}
	if publicFound != wantPriorPresent {
		t.Fatalf("pending %s public control = found:%t identity:%#v, want prior-present:%t", target, publicFound, publicIdentity, wantPriorPresent)
	}
	if wantPriorPresent && publicIdentity != priorIdentity {
		t.Fatalf("pending %s public control identity = %#v, want prior %#v", target, publicIdentity, priorIdentity)
	}
	preparedPath := filepath.Join(connectorRoot, state.transactionName, vNextPublicationControlRepairPreparedFile)
	if _, err := os.Stat(preparedPath); err != nil {
		t.Fatalf("pending %s prepared record missing: %v", target, err)
	}
}

func TestCP11F03CRepairTemporaryCleanupPreservesReplacementB(t *testing.T) {
	for _, test := range []struct {
		name    string
		foreign []byte
	}{
		{name: "empty"},
		{name: "nonempty", foreign: []byte("foreign replacement B")},
	} {
		t.Run(test.name, func(t *testing.T) {
			vNextPublicationAssertTemporaryReplacementBForTest(t, test.foreign)
		})
	}
}

func TestCP11F03CRepairCurrentAndJournalTemporaryPathsPreserveReplacementB(t *testing.T) {
	for _, target := range []struct {
		name  string
		write func(*vNextGenerationPublisher, *vNextPublicationOperation, vNextGenerationPointer) error
	}{
		{name: "JOURNAL", write: func(p *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
			return p.writeJournalLocked(operation, vNextGenerationJournal{New: pointer, State: "prepared"})
		}},
		{name: "CURRENT", write: func(p *vNextGenerationPublisher, operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
			return p.writeCurrentLocked(operation, pointer)
		}},
	} {
		for _, foreign := range [][]byte{nil, []byte("foreign replacement B")} {
			name := "empty"
			if len(foreign) != 0 {
				name = "nonempty"
			}
			t.Run(target.name+"/"+name, func(t *testing.T) {
				root := t.TempDir()
				baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
				if err != nil {
					t.Fatal(err)
				}
				oldArtifacts := vNextPublicationArtifactsForTest("old", true)
				old, err := baseline.Publish(oldArtifacts)
				if err != nil {
					t.Fatal(err)
				}
				var movedA, replacementB string
				var originalIdentity vNextPublicationIdentity
				fired := false
				writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{AfterTemporaryOpen: func(parent *vNextPublicationDirectory, temporaryName string, temporary *vNextPublicationDirectory) error {
					if fired {
						return nil
					}
					fired = true
					var err error
					originalIdentity, err = vNextPublicationIdentityFromFile(temporary.file, "CURRENT/JOURNAL temporary A")
					if err != nil {
						return err
					}
					if err := unix.Renameat(int(parent.file.Fd()), temporaryName, int(parent.file.Fd()), temporaryName+".A"); err != nil {
						return err
					}
					if err := vNextPublicationCreateReplacementBForTest(parent, temporaryName, foreign); err != nil {
						return err
					}
					movedA = filepath.Join(root, "acme", temporaryName+".A")
					replacementB = filepath.Join(root, "acme", temporaryName)
					return nil
				}})
				if err != nil {
					t.Fatal(err)
				}
				operation, err := writer.openOperation(context.Background(), unix.LOCK_EX, true)
				if err != nil {
					t.Fatal(err)
				}
				err = target.write(writer, operation, old)
				operation.close()
				if err == nil || !strings.Contains(err.Error(), "identity changed") {
					t.Fatalf("%s temporary transition = %v, want owned-cleanup identity refusal", target.name, err)
				}
				if !fired {
					t.Fatalf("%s temporary allocation hook did not fire", target.name)
				}
				vNextPublicationAssertTemporaryReplacementForTest(t, movedA, replacementB, originalIdentity, foreign)
				fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
				if err != nil {
					t.Fatal(err)
				}
				if err := fresh.Recover(context.Background()); err != nil {
					t.Fatalf("fresh %s Recover = %v", target.name, err)
				}
				if err := fresh.Check(oldArtifacts); err != nil {
					t.Fatalf("fresh %s Check = %v", target.name, err)
				}
				if _, err := fresh.Publish(vNextPublicationArtifactsForTest("retry-"+target.name+"-"+name, true)); err != nil {
					t.Fatalf("fresh %s retry after temporary cleanup refusal = %v", target.name, err)
				}
			})
		}
	}
}

func vNextPublicationAssertTemporaryReplacementForTest(t *testing.T, movedA, replacementB string, wantA vNextPublicationIdentity, foreign []byte) {
	t.Helper()
	originalInfo, err := os.Stat(movedA)
	if err != nil || !originalInfo.IsDir() {
		t.Fatalf("moved temporary A = info=%v err=%v, want retained directory", originalInfo, err)
	}
	original, err := vNextPublicationOpenDirectory(movedA, "moved temporary A")
	if err != nil {
		t.Fatal(err)
	}
	actualA, identityErr := vNextPublicationIdentityFromFile(original.file, "moved temporary A")
	closeErr := original.Close()
	if identityErr != nil || closeErr != nil || actualA != wantA {
		t.Fatalf("moved temporary A identity = %#v identityErr=%v closeErr=%v, want %#v", actualA, identityErr, closeErr, wantA)
	}
	replacementInfo, err := os.Lstat(replacementB)
	if err != nil || !replacementInfo.IsDir() || os.SameFile(originalInfo, replacementInfo) {
		t.Fatalf("replacement temporary B = info=%v err=%v, want distinct retained directory", replacementInfo, err)
	}
	if len(foreign) != 0 {
		payload, err := os.ReadFile(filepath.Join(replacementB, "foreign"))
		if err != nil || !bytes.Equal(payload, foreign) {
			t.Fatalf("replacement temporary B bytes = %q err=%v, want %q", payload, err, foreign)
		}
	}
}

func vNextPublicationAssertTemporaryReplacementBForTest(t *testing.T, foreign []byte) {
	t.Helper()
	root := t.TempDir()
	var movedA, replacementB string
	fired := false
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{AfterTemporaryOpen: func(parent *vNextPublicationDirectory, name string, temporary *vNextPublicationDirectory) error {
		if fired {
			return nil
		}
		fired = true
		blocker, err := temporary.openFile(vNextPublicationTemporaryFile, "original temporary blocker", unix.O_CREAT|unix.O_EXCL|unix.O_WRONLY, 0o600, false)
		if err != nil {
			return err
		}
		if err := blocker.Close(); err != nil {
			return err
		}
		if err := unix.Renameat(int(parent.file.Fd()), name, int(parent.file.Fd()), name+".A"); err != nil {
			return err
		}
		if err := vNextPublicationCreateReplacementBForTest(parent, name, foreign); err != nil {
			return err
		}
		movedA = filepath.Join(root, "acme", name+".A")
		replacementB = filepath.Join(root, "acme", name)
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := publisher.Publish(vNextPublicationArtifactsForTest("temporary-retry", true)); err == nil {
		t.Fatal("Publish swallowed failed allocation identity/cleanup refusal")
	}
	if !fired {
		t.Fatal("temporary allocation hook did not fire")
	}
	if _, err := os.Stat(filepath.Join(movedA, vNextPublicationTemporaryFile)); err != nil {
		t.Fatalf("original temporary A/blocker was not retained: %v", err)
	}
	if _, err := os.Lstat(replacementB); err != nil {
		t.Fatalf("temporary cleanup removed replacement B: %v", err)
	}
	if len(foreign) != 0 {
		payload, err := os.ReadFile(filepath.Join(replacementB, "foreign"))
		if err != nil || string(payload) != string(foreign) {
			t.Fatalf("temporary cleanup changed nonempty replacement B: payload=%q err=%v", payload, err)
		}
	}
}

func TestCP11F03CRepairQuarantineCleanupPreservesReplacementB(t *testing.T) {
	for _, test := range []struct {
		name    string
		foreign []byte
	}{
		{name: "empty"},
		{name: "nonempty", foreign: []byte("foreign replacement B")},
	} {
		t.Run(test.name, func(t *testing.T) {
			vNextPublicationAssertQuarantineReplacementBForTest(t, test.foreign)
		})
	}
}

func TestCP11F03CRepairStaleStageQuarantinePreservesReplacementB(t *testing.T) {
	for _, foreign := range [][]byte{nil, []byte("foreign replacement B")} {
		name := "empty"
		if len(foreign) != 0 {
			name = "nonempty"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := baseline.Publish(vNextPublicationArtifactsForTest("active", true)); err != nil {
				t.Fatal(err)
			}
			operation, err := baseline.openOperation(context.Background(), unix.LOCK_EX, true)
			if err != nil {
				t.Fatal(err)
			}
			stageName, _, err := baseline.stageLocked(operation, vNextPublicationArtifactsForTest("stale-stage", true).Files, nil)
			operation.close()
			if err != nil {
				t.Fatal(err)
			}
			injected := errors.New("injected stale-stage quarantine completion")
			var movedA, replacementB string
			var originalIdentity vNextPublicationIdentity
			fired := false
			writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{AfterQuarantineOpen: func(parent *vNextPublicationDirectory, name string, quarantine *vNextPublicationDirectory, identity vNextPublicationIdentity) error {
				if fired {
					return nil
				}
				fired = true
				if actual, err := vNextPublicationIdentityFromFile(quarantine.file, "stale-stage quarantine A"); err != nil || actual != identity {
					return errors.New("opened stale-stage quarantine identity was not retained before replacement")
				}
				originalIdentity = identity
				if err := unix.Renameat(int(parent.file.Fd()), name, int(parent.file.Fd()), name+".A"); err != nil {
					return err
				}
				if err := vNextPublicationCreateReplacementBForTest(parent, name, foreign); err != nil {
					return err
				}
				movedA = filepath.Join(root, "acme", vNextPublicationGenerationDirectory, name+".A")
				replacementB = filepath.Join(root, "acme", vNextPublicationGenerationDirectory, name)
				return injected
			}})
			if err != nil {
				t.Fatal(err)
			}
			if err := writer.Prune(context.Background()); !errors.Is(err, injected) {
				t.Fatalf("Prune stale stage after quarantine replacement = %v, want injected cause", err)
			}
			if !fired {
				t.Fatal("stale-stage quarantine allocation hook did not fire")
			}
			vNextPublicationAssertTemporaryReplacementForTest(t, movedA, replacementB, originalIdentity, foreign)
			if _, err := os.Stat(filepath.Join(root, "acme", vNextPublicationGenerationDirectory, stageName)); err != nil {
				t.Fatalf("stale stage residue was removed after allocation failure: %v", err)
			}
			if err := os.RemoveAll(replacementB); err != nil {
				t.Fatalf("remove test-owned replacement B: %v", err)
			}
			if err := os.RemoveAll(movedA); err != nil {
				t.Fatalf("remove test-owned moved A: %v", err)
			}
			fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
			if err != nil {
				t.Fatal(err)
			}
			if err := fresh.Recover(context.Background()); err != nil {
				t.Fatalf("fresh Recover after stale-stage allocation completion = %v", err)
			}
			if err := fresh.Prune(context.Background()); err != nil {
				t.Fatalf("fresh Prune after stale-stage fixture restoration = %v", err)
			}
		})
	}
}

func vNextPublicationAssertQuarantineReplacementBForTest(t *testing.T, foreign []byte) {
	t.Helper()
	root := t.TempDir()
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	old, err := baseline.Publish(vNextPublicationArtifactsForTest("old", true))
	if err != nil {
		t.Fatal(err)
	}
	held, err := baseline.Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := baseline.Publish(vNextPublicationArtifactsForTest("new", true)); err != nil {
		held.Release()
		t.Fatal(err)
	}
	held.Release()
	injected := errors.New("injected opened quarantine completion failure")
	var movedA, replacementB string
	fired := false
	writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{AfterQuarantineOpen: func(parent *vNextPublicationDirectory, name string, quarantine *vNextPublicationDirectory, identity vNextPublicationIdentity) error {
		if fired {
			return nil
		}
		fired = true
		if actual, err := vNextPublicationIdentityFromFile(quarantine.file, "original quarantine A"); err != nil || actual != identity {
			return errors.New("opened quarantine identity was not retained before replacement")
		}
		if err := unix.Renameat(int(parent.file.Fd()), name, int(parent.file.Fd()), name+".A"); err != nil {
			return err
		}
		if err := vNextPublicationCreateReplacementBForTest(parent, name, foreign); err != nil {
			return err
		}
		movedA = filepath.Join(root, "acme", vNextPublicationGenerationDirectory, name+".A")
		replacementB = filepath.Join(root, "acme", vNextPublicationGenerationDirectory, name)
		return injected
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := writer.Prune(context.Background()); !errors.Is(err, injected) {
		t.Fatalf("Prune after original quarantine allocation interference = %v, want injected cause", err)
	}
	if !fired {
		t.Fatal("quarantine allocation hook did not fire")
	}
	if _, err := os.Stat(movedA); err != nil {
		t.Fatalf("opened quarantine A was not retained: %v", err)
	}
	if _, err := os.Lstat(replacementB); err != nil {
		t.Fatalf("quarantine cleanup removed replacement B: %v", err)
	}
	if len(foreign) != 0 {
		payload, err := os.ReadFile(filepath.Join(replacementB, "foreign"))
		if err != nil || string(payload) != string(foreign) {
			t.Fatalf("quarantine cleanup changed nonempty replacement B: payload=%q err=%v", payload, err)
		}
	}
	if !baseline.GenerationExists(old.Generation) {
		t.Fatal("original quarantine allocation interference removed the stale generation")
	}
}

func vNextPublicationCreateReplacementBForTest(parent *vNextPublicationDirectory, name string, foreign []byte) error {
	if err := unix.Mkdirat(int(parent.file.Fd()), name, 0o700); err != nil {
		return err
	}
	if len(foreign) == 0 {
		return nil
	}
	file, err := parent.openFile(name+"/foreign", "replacement B foreign member", unix.O_CREAT|unix.O_EXCL|unix.O_WRONLY, 0o600, false)
	if err != nil {
		return err
	}
	if _, err := file.Write(foreign); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func TestCP11F03BRepairCompoundCausesRemainInspectable(t *testing.T) {
	// Retain pointers across sequential controls to distinguish resource instances
	// despite descriptor-number reuse. Each callback performs one actual Close.
	closedFiles := make(map[*os.File]int)
	closeFile := func(file *os.File) error {
		closedFiles[file]++
		if closedFiles[file] != 1 {
			t.Errorf("compound resource instance %p Close=%d", file, closedFiles[file])
		}
		return file.Close()
	}

	t.Run("definitions and connector close", func(t *testing.T) {
		root := t.TempDir()
		definitionsClose := errors.New("injected definitions close failure")
		connectorClose := errors.New("injected connector close failure")
		publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{CloseDirectory: func(file *os.File, label string) error {
			if err := closeFile(file); err != nil {
				return err
			}
			switch label {
			case "publication definitions root":
				return definitionsClose
			case `connector publication root "acme"`:
				return connectorClose
			default:
				return nil
			}
		}})
		if err != nil {
			t.Fatal(err)
		}
		connector, err := publisher.openConnectorRoot(true)
		if connector != nil {
			t.Fatalf("openConnectorRoot returned connector after both real closes: %v", connector)
		}
		if !errors.Is(err, definitionsClose) || !errors.Is(err, connectorClose) {
			t.Fatalf("definitions/connector compound result = %v, want both close causes inspectable", err)
		}
	})

	t.Run("missing control parent close", func(t *testing.T) {
		root := t.TempDir()
		completion := errors.New("injected missing-control parent close failure")
		publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{CloseDirectory: func(file *os.File, label string) error {
			if err := closeFile(file); err != nil {
				return err
			}
			if label == "original missing control" {
				return completion
			}
			return nil
		}})
		if err != nil {
			t.Fatal(err)
		}
		operation, err := publisher.openOperation(context.Background(), unix.LOCK_EX, true)
		if err != nil {
			t.Fatal(err)
		}
		defer operation.close()
		_, found, _, err := vNextPublicationReadControlBound(operation.connector, "missing-control", "original missing control")
		if !found || !errors.Is(err, completion) || !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("missing-control compound result = found=%t err=%v, want both absence and completion causes", found, err)
		}
	})

	t.Run("staged file sync and close", func(t *testing.T) {
		root := t.TempDir()
		primary := errors.New("injected staged-file sync frontier")
		completion := errors.New("injected staged-file close completion")
		fired := false
		publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{
			At: func(point vNextPublicationFaultPoint) error {
				if point == vNextPublicationBeforeFileSync && !fired {
					fired = true
					return primary
				}
				return nil
			},
			CloseStageFile: func(file *os.File) error {
				if err := closeFile(file); err != nil {
					return err
				}
				return completion
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = publisher.Publish(vNextPublicationArtifactsForTest("compound-stage", true))
		if !fired || !errors.Is(err, primary) || !errors.Is(err, completion) {
			t.Fatalf("staged-file compound result = fired=%t err=%v, want both primary and completion causes", fired, err)
		}
	})

	t.Run("failed stage and owned cleanup", func(t *testing.T) {
		root := t.TempDir()
		primary := errors.New("injected staged-file primary failure")
		cleanup := errors.New("injected owned stage cleanup failure")
		publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
			switch point {
			case vNextPublicationBeforeFileSync:
				return primary
			case vNextPublicationBeforeStageRemoval:
				return cleanup
			default:
				return nil
			}
		}})
		if err != nil {
			t.Fatal(err)
		}
		_, err = publisher.Publish(vNextPublicationArtifactsForTest("compound-stage-cleanup", true))
		if !errors.Is(err, primary) || !errors.Is(err, cleanup) {
			t.Fatalf("failed stage/cleanup result = %v, want primary and owned cleanup causes", err)
		}
	})

	t.Run("capture sync and close", func(t *testing.T) {
		root := t.TempDir()
		primary := errors.New("injected capture post-sync completion")
		completion := errors.New("injected capture close completion")
		publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{
			At: func(point vNextPublicationFaultPoint) error {
				if point == vNextPublicationAfterControlRepairCaptureSync {
					return primary
				}
				return nil
			},
			CloseDirectory: func(file *os.File, label string) error {
				if err := closeFile(file); err != nil {
					return err
				}
				if label == "publication control capture" {
					return completion
				}
				return nil
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = publisher.Publish(vNextPublicationArtifactsForTest("compound-capture", true))
		if !errors.Is(err, primary) || !errors.Is(err, completion) {
			t.Fatalf("capture sync/close result = %v, want primary and completion causes", err)
		}
	})

	t.Run("failed open plus parent close", func(t *testing.T) {
		root := t.TempDir()
		completion := errors.New("injected failed-open parent close")
		directory, err := vNextPublicationOpenDirectoryWithCloseForTest(root, "compound root", func(file *os.File, label string) error {
			if err := closeFile(file); err != nil {
				return err
			}
			if label == "compound missing control" {
				return completion
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := directory.Close(); err != nil {
				t.Errorf("close test-owned directory: %v", err)
			}
		}()
		if _, err := directory.openFile("missing-control", "compound missing control", unix.O_RDONLY, 0, false); !errors.Is(err, fs.ErrNotExist) || !errors.Is(err, completion) {
			t.Fatalf("failed open/parent close = %v, want absence and completion causes", err)
		}
	})

	t.Run("parent close plus opened-file close", func(t *testing.T) {
		root := t.TempDir()
		parentCompletion := errors.New("injected opened-file parent close")
		fileCompletion := errors.New("injected opened-file completion")
		directory, err := vNextPublicationOpenDirectoryWithCloseForTest(root, "compound root", func(file *os.File, label string) error {
			if err := closeFile(file); err != nil {
				return err
			}
			if label == "compound opened file" {
				return parentCompletion
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := directory.Close(); err != nil {
				t.Errorf("close test-owned directory: %v", err)
			}
		}()
		vNextPublicationCloseOpenedFileAfterParentCloseForTest = func(fd int, label string) error {
			if err := unix.Close(fd); err != nil {
				return err
			}
			return fileCompletion
		}
		t.Cleanup(func() { vNextPublicationCloseOpenedFileAfterParentCloseForTest = nil })
		if _, err := directory.openFile("opened", "compound opened file", unix.O_CREAT|unix.O_EXCL|unix.O_WRONLY, 0o600, false); !errors.Is(err, parentCompletion) || !errors.Is(err, fileCompletion) {
			t.Fatalf("parent/opened-file close = %v, want both completion causes", err)
		}
	})

	t.Run("opened control read plus close and pure absence", func(t *testing.T) {
		root := t.TempDir()
		directory, err := vNextPublicationOpenDirectory(root, "compound read root")
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := directory.Close(); err != nil {
				t.Errorf("close test-owned directory: %v", err)
			}
		}()
		if payload, found, _, err := vNextPublicationReadControlBound(directory, "pure-absence", "pure absent control"); payload != nil || found || err != nil {
			t.Fatalf("pure absence = payload=%q found=%t err=%v, want clean absence", payload, found, err)
		}
		file, err := directory.openFile("control", "compound read control", unix.O_CREAT|unix.O_EXCL|unix.O_WRONLY, 0o600, false)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write([]byte("control payload")); err != nil {
			_ = closeFile(file)
			t.Fatal(err)
		}
		if err := closeFile(file); err != nil {
			t.Fatal(err)
		}
		primary := errors.New("injected read completion")
		completion := errors.New("injected read close completion")
		vNextPublicationAfterReadOpenControlForTest = func(file *os.File, label string) error {
			if label != "compound read control" {
				t.Fatalf("read hook label = %q", label)
			}
			return primary
		}
		vNextPublicationCloseReadControlForTest = func(file *os.File, label string) error {
			if err := closeFile(file); err != nil {
				return err
			}
			return completion
		}
		t.Cleanup(func() {
			vNextPublicationAfterReadOpenControlForTest = nil
			vNextPublicationCloseReadControlForTest = nil
		})
		if _, found, _, err := vNextPublicationReadControlBound(directory, "control", "compound read control"); !found || !errors.Is(err, primary) || !errors.Is(err, completion) {
			t.Fatalf("opened control read/close = found=%t err=%v, want primary and close causes", found, err)
		}
	})

	t.Run("shared record writers preserve short sync and close", func(t *testing.T) {
		for _, label := range []struct {
			name string
			file string
		}{
			{name: "authority-marker", file: vNextPublicationControlAuthorityMarkerFile},
			{name: "prepared", file: vNextPublicationControlRepairPreparedFile},
			{name: "phase", file: vNextPublicationControlRepairPhaseName(1)},
		} {
			t.Run(label.name, func(t *testing.T) {
				root := t.TempDir()
				directory, err := vNextPublicationOpenDirectory(root, "record writer root")
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					if err := directory.Close(); err != nil {
						t.Errorf("close test-owned directory: %v", err)
					}
				}()
				payload := []byte("shared record payload")
				record, err := vNextPublicationWriteControlRepairRecord(directory, label.file, label.name, payload, vNextPublicationControlRecordHooks{Write: func(file *os.File, _ string, payload []byte) (int, error) {
					return file.Write(payload[:len(payload)-1])
				}})
				if !record.created || record.identity.mode != unix.S_IFREG || !errors.Is(err, io.ErrShortWrite) || record.disposition != vNextPublicationRecordRemoved {
					t.Fatalf("%s actual short write = identity=%#v created=%t err=%v", label.name, record.identity, record.created, err)
				}
				if _, err := os.Lstat(filepath.Join(root, label.file)); !errors.Is(err, fs.ErrNotExist) {
					t.Fatal(err)
				}
				syncCompletion := errors.New(label.name + " sync completion")
				closeCompletion := errors.New(label.name + " close completion")
				closed := 0
				record, err = vNextPublicationWriteControlRepairRecord(directory, label.file, label.name, payload, vNextPublicationControlRecordHooks{
					Sync: func(file *os.File, _ string) error {
						if err := file.Sync(); err != nil {
							return err
						}
						return syncCompletion
					},
					Close: func(file *os.File, _ string) error {
						closed++
						if err := closeFile(file); err != nil {
							return err
						}
						return closeCompletion
					},
				})
				if !record.created || record.disposition != vNextPublicationRecordRetainedComplete || !errors.Is(err, syncCompletion) || !errors.Is(err, closeCompletion) || closed != 1 {
					t.Fatalf("%s sync/close = created=%t closed=%d err=%v", label.name, record.created, closed, err)
				}
			})
		}
	})

	t.Run("linked predecessor close", func(t *testing.T) {
		root := t.TempDir()
		baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
		if err != nil {
			t.Fatal(err)
		}
		old, err := baseline.Publish(vNextPublicationArtifactsForTest("old", true))
		if err != nil {
			t.Fatal(err)
		}
		knownTransactions := vNextPublicationRepairTransactionsForTest(t, filepath.Join(root, "acme"))
		completion := errors.New("injected linked predecessor close")
		closed := 0
		writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{CloseRepairPredecessor: func(directory *vNextPublicationDirectory) error {
			closed++
			if err := directory.Close(); err != nil {
				return err
			}
			return completion
		}})
		if err != nil {
			t.Fatal(err)
		}
		operation, err := writer.openOperation(context.Background(), unix.LOCK_EX, true)
		if err != nil {
			t.Fatal(err)
		}
		err = writer.writeCurrentLocked(operation, old)
		operation.close()
		if !errors.Is(err, completion) || closed != 1 {
			t.Fatalf("linked predecessor close = closed=%d err=%v, want completion once", closed, err)
		}
		vNextPublicationAssertNoPreparedGraphForTest(t, filepath.Join(root, "acme"), knownTransactions)
	})
}
