package store

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func closeStoreForTest(t *testing.T, db *Store) {
	t.Helper()
	if err := db.Close(); err != nil {
		t.Error(err)
	}
}

func TestArtifactProofBindsExactHeadsAndRatification(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "shepherd.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer closeStoreForTest(t, db)
	if err := db.EnsureDelivery(ctx, testDelivery("issue-389", 389)); err != nil {
		t.Fatal(err)
	}
	proof := ArtifactProof{
		ProofID: "proof-1", DeliveryID: "issue-389", Generation: 1, UnitID: "validate-milestone/M001", Attempt: 1,
		StartHead: strings.Repeat("a", 40), CandidateHead: strings.Repeat("b", 40), ValidatedHead: strings.Repeat("b", 40),
		ExpectedArtifact: ".gsd/phases/M001/VALIDATION.md", ArtifactHash: "sha256:" + strings.Repeat("c", 64),
		Validator: "openai-codex/gpt-5.6-sol", Thinking: "high", Ratified: true,
	}
	if err := db.PutArtifactProof(ctx, proof); err != nil {
		t.Fatal(err)
	}
	loaded, err := db.GetArtifactProof(ctx, proof.ProofID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.CandidateHead != proof.ValidatedHead || !loaded.Ratified {
		t.Fatalf("proof=%+v", loaded)
	}
	proof.ProofID = "proof-moved"
	proof.ValidatedHead = strings.Repeat("d", 40)
	if err := db.PutArtifactProof(ctx, proof); err == nil {
		t.Fatal("moved-head proof accepted")
	}
}

func TestArtifactProofRejectsUnratifiedResult(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "shepherd.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer closeStoreForTest(t, db)
	if err := db.EnsureDelivery(ctx, testDelivery("issue-389", 389)); err != nil {
		t.Fatal(err)
	}
	proof := ArtifactProof{
		ProofID: "proof-unratified", DeliveryID: "issue-389", Generation: 1, UnitID: "validate-milestone/M001", Attempt: 1,
		StartHead: strings.Repeat("a", 40), CandidateHead: strings.Repeat("b", 40), ValidatedHead: strings.Repeat("b", 40),
		ExpectedArtifact: ".gsd/phases/M001/VALIDATION.md", ArtifactHash: "sha256:" + strings.Repeat("c", 64),
		Validator: "openai-codex/gpt-5.6-sol", Thinking: "high", Ratified: false,
	}
	if err := db.PutArtifactProof(ctx, proof); err == nil {
		t.Fatal("unratified artifact proof accepted")
	}
}

func TestAttestationRejectsNonProceedVerdicts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "shepherd.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer closeStoreForTest(t, db)
	for _, verdict := range []string{"RETRY", "HALT"} {
		t.Run(verdict, func(t *testing.T) {
			record := AttestationRecord{RunID: "issue-389", HeadSHA: strings.Repeat(strings.ToLower(verdict[:1]), 40), Validator: "openai-codex/gpt-5.6-sol", Thinking: "high", Verdict: verdict, CreatedAt: time.Unix(1_700_000_000, 0).UTC()}
			if err := db.PutAttestation(ctx, record); err == nil {
				t.Fatalf("%s attestation accepted", verdict)
			}
		})
	}
}

func TestAttestationPersistsValidatorProof(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "shepherd.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer closeStoreForTest(t, db)
	record := AttestationRecord{RunID: "issue-389", HeadSHA: strings.Repeat("e", 40), Validator: "openai-codex/gpt-5.6-sol", Thinking: "high", Verdict: "PROCEED", CreatedAt: time.Unix(1_700_000_000, 0).UTC()}
	if err := db.PutAttestation(ctx, record); err != nil {
		t.Fatal(err)
	}
	loaded, err := db.GetAttestation(ctx, record.RunID, record.HeadSHA)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Validator != record.Validator || loaded.Thinking != "high" || loaded.Verdict != "PROCEED" {
		t.Fatalf("attestation=%+v", loaded)
	}
	record.HeadSHA = strings.Repeat("f", 40)
	record.Validator = "openai-codex/gpt-5.5"
	if err := db.PutAttestation(ctx, record); err == nil {
		t.Fatal("downgraded validator proof accepted")
	}
}
