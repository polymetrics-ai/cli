package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// The actor in these tests deliberately bypasses the publisher's cooperative
// Flock. It exercises the public-name race that a descriptor-bound lock cannot
// prevent.
func TestVNextGenerationPublisherCheckRefusesTerminalAuthorityFreeReplacement(t *testing.T) {
	root := t.TempDir()
	artifacts := vNextPublicationArtifactsForTest("active", false)
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	pointer, err := baseline.Publish(artifacts)
	if err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(root, "acme", vNextPublicationCurrentFile)
	priorPayload, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	priorInfo, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}

	attacked := false
	writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point != vNextPublicationAfterFinalControlRepairValidation || attacked {
			return nil
		}
		attacked = true
		if err := os.Rename(target, target+".attacker"); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, priorPayload, 0o600); err != nil {
			t.Fatal(err)
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	operation, err := writer.openOperation(context.Background(), syscall.LOCK_EX, true)
	if err != nil {
		t.Fatal(err)
	}
	err = writer.writeCurrentLocked(operation, pointer)
	operation.close()
	if !errors.Is(err, errVNextPublicationControlConflict) {
		t.Fatalf("writeCurrentLocked() error = %v, want retained terminal conflict", err)
	}
	if !attacked {
		t.Fatal("writer did not reach the final authority validation barrier")
	}
	currentInfo, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(priorInfo, currentInfo) {
		t.Fatal("direct actor did not replace CURRENT after final validation")
	}
	if got, err := os.ReadFile(target); err != nil || !bytes.Equal(got, priorPayload) {
		t.Fatalf("attacker CURRENT = %q, %v; want valid copied public payload", got, err)
	}

	fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.Check(artifacts); err == nil {
		t.Fatal("Check() accepted a valid but authority-free replacement CURRENT inode")
	}
}

func TestVNextGenerationPublisherCheckRefusesPendingPrivateAuthority(t *testing.T) {
	root := t.TempDir()
	artifacts := vNextPublicationArtifactsForTest("active", false)
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	pointer, err := baseline.Publish(artifacts)
	if err != nil {
		t.Fatal(err)
	}

	crash := errors.New("crash after private JOURNAL authority preparation")
	writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point == vNextPublicationAfterControlRepairPrepared {
			return crash
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	operation, err := writer.openOperation(context.Background(), syscall.LOCK_EX, true)
	if err != nil {
		t.Fatal(err)
	}
	err = writer.writeJournalLocked(operation, vNextGenerationJournal{New: pointer, State: "prepared"})
	operation.close()
	if !errors.Is(err, crash) {
		t.Fatalf("writeJournalLocked() error = %v, want injected crash", err)
	}
	if _, err := os.Stat(filepath.Join(root, "acme", vNextPublicationJournalFile)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("JOURNAL after prepared crash = %v, want absent", err)
	}

	fresh, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.Check(artifacts); err == nil {
		t.Fatal("Check() accepted CURRENT while a private no-prior JOURNAL authority is pending")
	}
}

func TestVNextGenerationPublisherCapturesLatePublicControlReplacementWithoutClobber(t *testing.T) {
	root := t.TempDir()
	artifacts := vNextPublicationArtifactsForTest("active", false)
	baseline, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	pointer, err := baseline.Publish(artifacts)
	if err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(root, "acme", vNextPublicationCurrentFile)
	priorPayload, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	replacement := []byte("late public CURRENT replacement")
	attacked := false
	writer, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{At: func(point vNextPublicationFaultPoint) error {
		if point == vNextPublicationBeforeControlRepairCapture && !attacked {
			attacked = true
			vNextReplacePublicationControlForTest(t, root, vNextPublicationCurrentFile, replacement)
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	operation, err := writer.openOperation(context.Background(), syscall.LOCK_EX, true)
	if err != nil {
		t.Fatal(err)
	}
	err = writer.writeCurrentLocked(operation, pointer)
	operation.close()
	if !attacked {
		t.Fatal("writer did not reach the no-replace capture boundary")
	}
	if err == nil {
		t.Fatal("writeCurrentLocked() accepted a late public replacement")
	}
	if got, err := os.ReadFile(target); err != nil || !bytes.Equal(got, priorPayload) {
		t.Fatalf("CURRENT after rejected replacement = %q, %v; want restored prior control", got, err)
	}

	transactions, err := filepath.Glob(filepath.Join(root, "acme", vNextPublicationControlRepairDirectoryPrefix+"*"))
	if err != nil {
		t.Fatal(err)
	}
	for _, transaction := range transactions {
		captured, err := os.ReadFile(filepath.Join(transaction, vNextPublicationControlCaptureName(1), vNextPublicationControlCaptureMember))
		if err == nil && bytes.Equal(captured, replacement) {
			return
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			t.Fatal(err)
		}
	}
	t.Fatalf("no-replace capture did not retain late public replacement %q", replacement)
}
