package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"time"

	shepherdgit "github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/git"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/store"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/workspace"
)

type promotionBoundary string

const (
	promotionAfterJournalCreated  promotionBoundary = "after_journal_created"
	promotionAfterStateStaged     promotionBoundary = "after_state_staged"
	promotionBeforeGit            promotionBoundary = "before_git_promotion"
	promotionAfterGit             promotionBoundary = "after_git_promotion"
	promotionAfterGitJournaled    promotionBoundary = "after_git_promoted_journaled"
	promotionBeforeBackupRename   promotionBoundary = "before_backup_rename"
	promotionAfterBackupRename    promotionBoundary = "after_backup_rename"
	promotionAfterStateInstall    promotionBoundary = "after_state_install"
	promotionAfterStateJournaled  promotionBoundary = "after_state_installed_journaled"
	promotionAfterJournalComplete promotionBoundary = "after_journal_complete"
)

var promotionFailureInjector func(promotionBoundary) error

func promoteRatifiedAttempt(
	ctx context.Context,
	authority *store.Store,
	manager *workspace.Manager,
	repositoryLock *workspace.RepositoryLock,
	lease store.Lease,
	delivery store.Delivery,
	attempt workspace.AttemptWorktree,
	lifecycle *attemptLifecycle,
) error {
	if lifecycle == nil {
		return errors.New("durable attempt lifecycle missing before promotion")
	}
	proof, attestation, err := authority.GetRatifiedPromotionAuthority(ctx, lifecycle.key)
	if err != nil {
		return err
	}
	validatedManifestHash, err := validatedGSDManifestHash(proof)
	if err != nil {
		return err
	}
	journalID := promotionJournalID(lifecycle.key, proof.CandidateHead)
	paths, err := workspace.PlanPromotionPaths(manager.RepoRoot, journalID)
	if err != nil {
		return err
	}
	if err := manager.VerifyOwnedAttemptAtHead(ctx, attempt, proof.CandidateHead); err != nil {
		return fmt.Errorf("verify ratified attempt ownership: %w", err)
	}
	journal := store.PromotionJournal{
		JournalID: journalID, DeliveryID: lifecycle.key.DeliveryID, Generation: lifecycle.key.Generation,
		UnitID: lifecycle.key.UnitID, Attempt: lifecycle.key.Attempt, BaseHead: proof.StartHead,
		CandidateHead: proof.CandidateHead, ValidatedHead: proof.ValidatedHead, ProofID: proof.ProofID,
		EvidenceHash: attestation.EvidenceHash, ValidatorSessionID: attestation.ValidatorSessionID,
		AttestationRepository: attestation.Repository, AttestationPR: attestation.PR,
		AttestationBaseBranch: attestation.BaseBranch, AttestationContractHash: attestation.ContractHash,
		AttestationCreatedAt: attestation.CreatedAt, AttestationValidator: attestation.Validator, AttestationThinking: attestation.Thinking,
		AttestationVerdict: attestation.Verdict, AttestationLocalGates: attestation.LocalGates,
		AttestationUAT: attestation.UAT, AttestationMilestoneValid: attestation.MilestoneValid,
		GovernanceStateVersion: attestation.StateVersion, AttestationExpiresAt: attestation.ExpiresAt,
		StagePath: paths.Stage, BackupPath: paths.Backup, CanonicalPath: paths.Canonical,
		State: store.PromotionJournalCreated, ControllerOwner: lease.Owner, ControllerEpoch: lease.Epoch,
	}
	if err := authority.CreatePromotionJournal(ctx, lease, journal); err != nil {
		return err
	}
	lifecycle.state = store.AttemptWorktreePromoting
	if err := hitPromotionBoundary(promotionAfterJournalCreated); err != nil {
		return err
	}
	staged, err := workspace.StageGSDState(ctx, attempt.Root+"/.gsd", paths)
	if err != nil {
		return fmt.Errorf("stage ratified GSD state: %w", err)
	}
	if staged.ManifestHash != validatedManifestHash {
		return promotionIntegrityError{reason: "staged_state_differs_from_validated_manifest",
			err: errors.New("normalized stage does not match validator-bound GSD state")}
	}
	if _, err := authority.FinalizePromotionJournalStage(ctx, lease, journalID, staged.ManifestJSON,
		staged.ManifestHash, staged.BackupManifestJSON, staged.BackupManifestHash); err != nil {
		return err
	}
	if err := hitPromotionBoundary(promotionAfterStateStaged); err != nil {
		return err
	}
	if err := continuePromotion(ctx, authority, manager, repositoryLock, lease, delivery, attempt, journalID); err != nil {
		return err
	}
	lifecycle.state = store.AttemptWorktreePromoted
	return nil
}

func recoverPromotionJournals(
	ctx context.Context,
	authority *store.Store,
	manager *workspace.Manager,
	repositoryLock *workspace.RepositoryLock,
	delivery store.Delivery,
	lease store.Lease,
) error {
	return recoverPromotionJournalsMatching(ctx, authority, manager, repositoryLock, delivery, lease,
		func(store.PromotionJournal) (bool, error) { return true, nil })
}

func validatePreGitPromotionJournals(
	ctx context.Context,
	authority *store.Store,
	manager *workspace.Manager,
	delivery store.Delivery,
	lease store.Lease,
) error {
	journals, err := authority.ListIncompletePromotionJournals(ctx, delivery.ID)
	if err != nil {
		return err
	}
	for _, journal := range journals {
		if journal.State != store.PromotionJournalCreated && journal.State != store.PromotionJournalStateStaged &&
			journal.State != store.PromotionJournalGitPromoting {
			continue
		}
		journal, err = authority.ClaimPromotionJournalRecovery(ctx, journal.JournalID, lease)
		if err != nil {
			return err
		}
		validatedManifestHash, proofErr := journalValidatedGSDManifestHash(ctx, authority, journal)
		if proofErr != nil {
			reason := "validated_manifest_authority_invalid"
			if errors.Is(proofErr, errLegacyGSDManifest) {
				reason = "legacy_pre_git_manifest_requires_human_reconciliation"
			}
			return blockPromotionJournal(ctx, authority, lease, journal.JournalID, reason)
		}
		if journal.ManifestHash != "" && journal.ManifestHash != validatedManifestHash {
			return blockPromotionJournal(ctx, authority, lease, journal.JournalID,
				"staged_state_differs_from_validated_manifest")
		}
		if _, snapshotErr := canonicalPromotionSnapshot(ctx, manager.RepoRoot, delivery.Branch,
			journal.BaseHead, journal.BaseHead); snapshotErr != nil {
			return blockPromotionJournal(ctx, authority, lease, journal.JournalID,
				"canonical_git_moved_or_dirty")
		}
		paths, pathErr := workspace.PlanPromotionPaths(manager.RepoRoot, journal.JournalID)
		if pathErr != nil || paths.Canonical != journal.CanonicalPath || paths.Stage != journal.StagePath ||
			paths.Backup != journal.BackupPath {
			return blockPromotionJournal(ctx, authority, lease, journal.JournalID,
				"promotion_path_ownership_mismatch")
		}
		attemptRecord, getErr := authority.GetAttemptWorktree(ctx, journal.AttemptKey())
		if getErr != nil {
			return blockPromotionJournal(ctx, authority, lease, journal.JournalID, "promotion_attempt_missing")
		}
		attempt := workspace.AttemptWorktree{Root: attemptRecord.Path, Branch: attemptRecord.Branch,
			Identity: workspace.AttemptIdentity{DeliveryID: attemptRecord.DeliveryID,
				Generation: attemptRecord.Generation, UnitID: attemptRecord.UnitID,
				Attempt: attemptRecord.Attempt, BaseHead: attemptRecord.BaseHead}}
		validationErr := manager.VerifyOwnedAttemptAtHead(ctx, attempt, journal.CandidateHead)
		if validationErr == nil {
			validationErr = authority.ValidatePromotionAuthority(ctx, lease, journal.JournalID, time.Now().UTC())
		}
		if validationErr == nil && journal.ManifestHash != "" {
			validationErr = workspace.ValidateStagedGSDState(ctx, paths, journal.ManifestHash,
				journal.BackupManifestHash)
		}
		if validationErr != nil {
			return blockPromotionJournal(ctx, authority, lease, journal.JournalID,
				promotionFailureClass(promotionIntegrityError{reason: "pre_git_authority_or_stage_invalid",
					err: validationErr}))
		}
	}
	return nil
}

func recoverPostGitPromotionJournals(
	ctx context.Context,
	authority *store.Store,
	manager *workspace.Manager,
	repositoryLock *workspace.RepositoryLock,
	delivery store.Delivery,
	lease store.Lease,
) error {
	return recoverPromotionJournalsMatching(ctx, authority, manager, repositoryLock, delivery, lease,
		func(journal store.PromotionJournal) (bool, error) {
			switch journal.State {
			case store.PromotionJournalGitPromoting:
				snapshot, err := shepherdgit.Inspect(ctx, manager.RepoRoot)
				if err != nil {
					return false, err
				}
				return snapshot.Branch != delivery.Branch || snapshot.HeadSHA != journal.BaseHead, nil
			case store.PromotionJournalGitPromoted, store.PromotionJournalStateSwapStarted,
				store.PromotionJournalStateInstalled, store.PromotionJournalComplete:
				return true, nil
			default:
				return false, nil
			}
		})
}

func recoverPromotionJournalsMatching(
	ctx context.Context,
	authority *store.Store,
	manager *workspace.Manager,
	repositoryLock *workspace.RepositoryLock,
	delivery store.Delivery,
	lease store.Lease,
	matches func(store.PromotionJournal) (bool, error),
) error {
	journals, err := authority.ListIncompletePromotionJournals(ctx, delivery.ID)
	if err != nil {
		return err
	}
	for _, journal := range journals {
		matched, matchErr := matches(journal)
		if matchErr != nil {
			return matchErr
		}
		if !matched {
			continue
		}
		if journal.State == store.PromotionJournalComplete && journal.CleanupComplete && !journal.DecisionsResolved {
			// Promotion and owned-resource cleanup are already durable. Only the
			// post-poll decision-resolution projection remains; replaying the
			// promotion claim would incorrectly require a cleaned attempt.
			continue
		}
		journal, err = authority.ClaimPromotionJournalRecovery(ctx, journal.JournalID, lease)
		if err != nil {
			return err
		}
		expectedPaths, pathErr := workspace.PlanPromotionPaths(manager.RepoRoot, journal.JournalID)
		if pathErr != nil || expectedPaths.Canonical != journal.CanonicalPath ||
			expectedPaths.Stage != journal.StagePath || expectedPaths.Backup != journal.BackupPath {
			return blockPromotionJournal(ctx, authority, lease, journal.JournalID, "promotion_path_ownership_mismatch")
		}
		attemptRecord, getErr := authority.GetAttemptWorktree(ctx, journal.AttemptKey())
		if getErr != nil {
			return blockPromotionJournal(ctx, authority, lease, journal.JournalID, "promotion_attempt_missing")
		}
		attempt := workspace.AttemptWorktree{Root: attemptRecord.Path, Branch: attemptRecord.Branch,
			Identity: workspace.AttemptIdentity{DeliveryID: attemptRecord.DeliveryID, Generation: attemptRecord.Generation,
				UnitID: attemptRecord.UnitID, Attempt: attemptRecord.Attempt, BaseHead: attemptRecord.BaseHead}}
		if err := continuePromotion(ctx, authority, manager, repositoryLock, lease, delivery, attempt, journal.JournalID); err != nil {
			if isPromotionIntegrityFailure(err) {
				return blockPromotionJournal(ctx, authority, lease, journal.JournalID, promotionFailureClass(err))
			}
			return err
		}
	}
	return nil
}

func continuePromotion(
	ctx context.Context,
	authority *store.Store,
	manager *workspace.Manager,
	repositoryLock *workspace.RepositoryLock,
	lease store.Lease,
	delivery store.Delivery,
	attempt workspace.AttemptWorktree,
	journalID string,
) error {
	for {
		journal, err := authority.GetPromotionJournal(ctx, journalID)
		if err != nil {
			return err
		}
		paths := workspace.PromotionPaths{Canonical: journal.CanonicalPath, Stage: journal.StagePath, Backup: journal.BackupPath}
		validatedManifestHash, err := journalValidatedGSDManifestHash(ctx, authority, journal)
		legacyManifest := errors.Is(err, errLegacyGSDManifest)
		if err != nil && !legacyManifest {
			return promotionIntegrityError{reason: "validated_manifest_authority_invalid", err: err}
		}
		if legacyManifest {
			snapshot, snapshotErr := canonicalPromotionSnapshot(ctx, manager.RepoRoot, delivery.Branch,
				journal.BaseHead, journal.CandidateHead)
			if snapshotErr != nil {
				return promotionIntegrityError{reason: "legacy_manifest_canonical_state_invalid", err: snapshotErr}
			}
			if snapshot.HeadSHA != journal.CandidateHead {
				return promotionIntegrityError{reason: "legacy_pre_git_manifest_requires_human_reconciliation",
					err: errLegacyGSDManifest}
			}
		}
		if !legacyManifest && journal.ManifestHash != "" && journal.ManifestHash != validatedManifestHash {
			return promotionIntegrityError{reason: "staged_state_differs_from_validated_manifest",
				err: errors.New("journaled stage does not match validator-bound GSD state")}
		}
		switch journal.State {
		case store.PromotionJournalCreated:
			if journal.ManifestHash == "" {
				if err := manager.VerifyOwnedAttemptAtHead(ctx, attempt, journal.CandidateHead); err != nil {
					return promotionIntegrityError{reason: "attempt_ownership_invalid", err: err}
				}
				if err := workspace.ResetPromotionStage(ctx, paths); err != nil {
					return promotionIntegrityError{reason: "partial_stage_reset_failed", err: err}
				}
				staged, err := workspace.StageGSDState(ctx, attempt.Root+"/.gsd", paths)
				if err != nil {
					return promotionIntegrityError{reason: "state_staging_failed", err: err}
				}
				if !legacyManifest && staged.ManifestHash != validatedManifestHash {
					return promotionIntegrityError{reason: "staged_state_differs_from_validated_manifest",
						err: errors.New("normalized stage does not match validator-bound GSD state")}
				}
				if _, err := authority.FinalizePromotionJournalStage(ctx, lease, journalID, staged.ManifestJSON,
					staged.ManifestHash, staged.BackupManifestJSON, staged.BackupManifestHash); err != nil {
					return err
				}
				if err := hitPromotionBoundary(promotionAfterStateStaged); err != nil {
					return err
				}
				continue
			}
			if err := workspace.ValidateStagedGSDState(ctx, paths, journal.ManifestHash, journal.BackupManifestHash); err != nil {
				return promotionIntegrityError{reason: "staged_state_invalid", err: err}
			}
			if _, err := authority.TransitionPromotionJournal(ctx, lease, journalID, store.PromotionJournalStateStaged, ""); err != nil {
				return err
			}
		case store.PromotionJournalStateStaged:
			if err := workspace.ValidateStagedGSDState(ctx, paths, journal.ManifestHash, journal.BackupManifestHash); err != nil {
				return promotionIntegrityError{reason: "staged_state_invalid", err: err}
			}
			if err := manager.VerifyOwnedAttemptAtHead(ctx, attempt, journal.CandidateHead); err != nil {
				return promotionIntegrityError{reason: "attempt_ownership_invalid", err: err}
			}
			if err := authority.ValidatePromotionAuthority(ctx, lease, journalID, time.Now().UTC()); err != nil {
				return promotionIntegrityError{reason: "promotion_authority_invalid", err: err}
			}
			if err := hitPromotionBoundary(promotionBeforeGit); err != nil {
				return err
			}
			if _, err := authority.TransitionPromotionJournal(ctx, lease, journalID, store.PromotionJournalGitPromoting, ""); err != nil {
				return err
			}
		case store.PromotionJournalGitPromoting:
			if err := repositoryLock.Check(); err != nil {
				return err
			}
			snapshot, err := canonicalPromotionSnapshot(ctx, manager.RepoRoot, delivery.Branch, journal.BaseHead, journal.CandidateHead)
			if err != nil {
				return promotionIntegrityError{reason: "canonical_git_moved_or_dirty", err: err}
			}
			if snapshot.HeadSHA == journal.BaseHead {
				if err := workspace.ValidateStagedGSDState(ctx, paths, journal.ManifestHash, journal.BackupManifestHash); err != nil {
					return promotionIntegrityError{reason: "staged_state_invalid", err: err}
				}
				if err := manager.VerifyOwnedAttemptAtHead(ctx, attempt, journal.CandidateHead); err != nil {
					return promotionIntegrityError{reason: "attempt_ownership_invalid", err: err}
				}
				if err := authority.ValidatePromotionAuthority(ctx, lease, journalID, time.Now().UTC()); err != nil {
					return promotionIntegrityError{reason: "promotion_authority_invalid", err: err}
				}
				if err := manager.PromoteCandidate(ctx, attempt, journal.CandidateHead); err != nil {
					return promotionIntegrityError{reason: "candidate_git_promotion_failed", err: err}
				}
			}
			if err := hitPromotionBoundary(promotionAfterGit); err != nil {
				return err
			}
			if _, err := authority.TransitionPromotionJournal(ctx, lease, journalID, store.PromotionJournalGitPromoted, ""); err != nil {
				return err
			}
			if err := hitPromotionBoundary(promotionAfterGitJournaled); err != nil {
				return err
			}
		case store.PromotionJournalGitPromoted:
			if _, err := canonicalPromotionSnapshot(ctx, manager.RepoRoot, delivery.Branch, journal.CandidateHead, journal.CandidateHead); err != nil {
				return promotionIntegrityError{reason: "promoted_git_moved_or_dirty", err: err}
			}
			if _, err := authority.TransitionPromotionJournal(ctx, lease, journalID, store.PromotionJournalStateSwapStarted, ""); err != nil {
				return err
			}
		case store.PromotionJournalStateSwapStarted:
			if _, err := canonicalPromotionSnapshot(ctx, manager.RepoRoot, delivery.Branch, journal.CandidateHead, journal.CandidateHead); err != nil {
				return promotionIntegrityError{reason: "promoted_git_moved_or_dirty", err: err}
			}
			if err := workspace.RecoverGSDState(ctx, paths, journal.ManifestHash, journal.BackupManifestHash, workspacePromotionFailpoint); err != nil {
				return promotionIntegrityError{reason: "state_swap_recovery_failed", err: err}
			}
			if _, err := authority.TransitionPromotionJournal(ctx, lease, journalID, store.PromotionJournalStateInstalled, ""); err != nil {
				return err
			}
			if err := hitPromotionBoundary(promotionAfterStateJournaled); err != nil {
				return err
			}
		case store.PromotionJournalStateInstalled:
			if _, err := canonicalPromotionSnapshot(ctx, manager.RepoRoot, delivery.Branch, journal.CandidateHead, journal.CandidateHead); err != nil {
				return promotionIntegrityError{reason: "installed_state_git_moved_or_dirty", err: err}
			}
			if err := workspace.ValidateInstalledGSDState(ctx, paths, journal.ManifestHash, journal.BackupManifestHash); err != nil {
				return promotionIntegrityError{reason: "installed_state_invalid", err: err}
			}
			if _, err := authority.TransitionPromotionJournal(ctx, lease, journalID, store.PromotionJournalComplete, ""); err != nil {
				return err
			}
			if err := hitPromotionBoundary(promotionAfterJournalComplete); err != nil {
				return err
			}
		case store.PromotionJournalComplete:
			if _, err := authority.CompletePromotionAttempt(ctx, lease, journalID); err != nil {
				return err
			}
			if err := workspace.CleanupPromotionArtifacts(ctx, paths, journal.ManifestHash, journal.BackupManifestHash); err != nil {
				return promotionIntegrityError{reason: "promotion_backup_cleanup_failed", err: err}
			}
			if err := authority.MarkPromotionCleanupComplete(ctx, lease, journalID); err != nil {
				return err
			}
			return nil
		case store.PromotionJournalBlocked:
			return errors.New("promotion journal requires human recovery")
		default:
			return promotionIntegrityError{reason: "unknown_promotion_state", err: errors.New("unknown promotion journal state")}
		}
	}
}

func validateFinalGateGSDState(ctx context.Context, authority *store.Store, stateDir, workDir,
	deliveryID string, generation int64, headSHA string,
) error {
	journal, err := authority.GetFinalGatePromotionJournal(ctx, deliveryID, generation, headSHA)
	if err != nil {
		return err
	}
	manifest, err := workspace.SnapshotGSDManifest(ctx, filepath.Join(workDir, ".gsd"), stateDir)
	if err != nil {
		return err
	}
	if manifest.Hash != journal.ManifestHash {
		return errors.New("canonical GSD state differs from the resolved promotion")
	}
	return nil
}

func canonicalPromotionSnapshot(ctx context.Context, workDir, branch, base, candidate string) (shepherdgit.Snapshot, error) {
	snapshot, err := shepherdgit.Inspect(ctx, workDir)
	if err != nil {
		return shepherdgit.Snapshot{}, err
	}
	if snapshot.Branch != branch || snapshot.HeadSHA != base && snapshot.HeadSHA != candidate {
		return shepherdgit.Snapshot{}, errors.New("canonical branch or head is outside journal bounds")
	}
	if err := shepherdgit.RequireClean(snapshot); err != nil {
		return shepherdgit.Snapshot{}, err
	}
	return snapshot, nil
}

func promotionJournalID(key store.AttemptWorktreeKey, candidateHead string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%d\x00%s\x00%d\x00%s", key.DeliveryID, key.Generation, key.UnitID, key.Attempt, candidateHead)))
	return "promotion-" + hex.EncodeToString(sum[:16])
}

func workspacePromotionFailpoint(boundary workspace.PromotionBoundary) error {
	switch boundary {
	case workspace.PromotionBoundaryBeforeBackupRename:
		return hitPromotionBoundary(promotionBeforeBackupRename)
	case workspace.PromotionBoundaryAfterBackupRename:
		return hitPromotionBoundary(promotionAfterBackupRename)
	case workspace.PromotionBoundaryAfterStageInstall:
		return hitPromotionBoundary(promotionAfterStateInstall)
	default:
		return nil
	}
}

var errLegacyGSDManifest = errors.New("ratified proof predates normalized GSD manifest binding")

func validatedGSDManifestHash(proof store.ArtifactProof) (string, error) {
	digest := sha256.Sum256([]byte(proof.ExpectedArtifact))
	if proof.ArtifactHash != "sha256:"+hex.EncodeToString(digest[:]) {
		return "", errors.New("ratified proof artifact manifest digest mismatch")
	}
	fields, err := strictManifestFields([]byte(proof.ExpectedArtifact))
	if err != nil {
		return "", errors.New("ratified proof has invalid artifact manifest")
	}
	for _, required := range []string{"unit_type", "phase_chain", "required_workflow_tools", "artifacts"} {
		if _, ok := fields[required]; !ok {
			return "", errors.New("ratified proof has incomplete artifact manifest")
		}
	}
	for name := range fields {
		if name != "unit_type" && name != "phase_chain" && name != "required_workflow_tools" &&
			name != "observed_workflow_tools" && name != "gsd_manifest_hash" && name != "artifacts" {
			return "", errors.New("ratified proof has unknown artifact manifest field")
		}
	}
	_, hasManifestHash := fields["gsd_manifest_hash"]
	if hasManifestHash {
		if _, ok := fields["observed_workflow_tools"]; !ok {
			return "", errors.New("ratified proof has no observed workflow transition")
		}
	}
	var manifest struct {
		UnitType              string   `json:"unit_type"`
		PhaseChain            []string `json:"phase_chain"`
		RequiredWorkflowTools []string `json:"required_workflow_tools"`
		ObservedWorkflowTools []string `json:"observed_workflow_tools"`
		GSDManifestHash       string   `json:"gsd_manifest_hash"`
		Artifacts             []struct {
			Path    string `json:"path"`
			Hash    string `json:"hash"`
			Deleted bool   `json:"deleted,omitempty"`
		} `json:"artifacts"`
	}
	decoder := json.NewDecoder(bytes.NewReader([]byte(proof.ExpectedArtifact)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil || decoder.Decode(&struct{}{}) != io.EOF ||
		manifest.UnitType == "" || len(manifest.PhaseChain) == 0 || len(manifest.Artifacts) == 0 {
		return "", errors.New("ratified proof has invalid artifact manifest")
	}
	for _, artifact := range manifest.Artifacts {
		if artifact.Path == "" || len(artifact.Hash) != len("sha256:")+64 ||
			artifact.Hash[:len("sha256:")] != "sha256:" {
			return "", errors.New("ratified proof has invalid artifact entry")
		}
		if _, err := hex.DecodeString(artifact.Hash[len("sha256:"):]); err != nil {
			return "", errors.New("ratified proof has invalid artifact entry")
		}
		if artifact.Deleted != (artifact.Hash == shepherdgit.DeletionSentinelHash) {
			return "", errors.New("ratified proof has inconsistent deletion artifact entry")
		}
	}
	if !hasManifestHash {
		return "", errLegacyGSDManifest
	}
	if manifest.GSDManifestHash == "" {
		return "", errors.New("ratified proof has empty GSD manifest hash")
	}
	for _, required := range manifest.RequiredWorkflowTools {
		if !slices.Contains(manifest.ObservedWorkflowTools, required) {
			return "", errors.New("ratified proof lacks a required observed workflow transition")
		}
	}
	if len(manifest.GSDManifestHash) != len("sha256:")+64 ||
		manifest.GSDManifestHash[:len("sha256:")] != "sha256:" {
		return "", errors.New("ratified proof has no bounded GSD manifest hash")
	}
	if _, err := hex.DecodeString(manifest.GSDManifestHash[len("sha256:"):]); err != nil {
		return "", errors.New("ratified proof has invalid GSD manifest hash")
	}
	return manifest.GSDManifestHash, nil
}

func strictManifestFields(raw []byte) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return nil, errors.New("artifact manifest must be an object")
	}
	fields := make(map[string]json.RawMessage)
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyToken.(string)
		if !ok {
			return nil, errors.New("artifact manifest key must be a string")
		}
		if _, duplicate := fields[key]; duplicate {
			return nil, errors.New("artifact manifest has duplicate fields")
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		fields[key] = value
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') || decoder.Decode(&struct{}{}) != io.EOF {
		return nil, errors.New("artifact manifest has trailing data")
	}
	return fields, nil
}

func journalValidatedGSDManifestHash(ctx context.Context, authority *store.Store,
	journal store.PromotionJournal,
) (string, error) {
	proof, err := authority.GetArtifactProof(ctx, journal.ProofID)
	if err != nil {
		return "", err
	}
	if !proof.Ratified || proof.DeliveryID != journal.DeliveryID || proof.Generation != journal.Generation ||
		proof.UnitID != journal.UnitID || proof.Attempt != journal.Attempt || proof.StartHead != journal.BaseHead ||
		proof.CandidateHead != journal.CandidateHead || proof.ValidatedHead != journal.ValidatedHead ||
		proof.ArtifactHash != journal.EvidenceHash {
		return "", errors.New("promotion journal proof identity changed")
	}
	return validatedGSDManifestHash(proof)
}

func hitPromotionBoundary(boundary promotionBoundary) error {
	if promotionFailureInjector == nil {
		return nil
	}
	return promotionFailureInjector(boundary)
}

type promotionIntegrityError struct {
	reason string
	err    error
}

func (e promotionIntegrityError) Error() string { return e.reason + ": " + e.err.Error() }
func (e promotionIntegrityError) Unwrap() error { return e.err }

func isPromotionIntegrityFailure(err error) bool {
	var target promotionIntegrityError
	return errors.As(err, &target)
}

func promotionFailureClass(err error) string {
	var target promotionIntegrityError
	if errors.As(err, &target) {
		return target.reason
	}
	return "promotion_recovery_failed"
}

func blockPromotionJournal(ctx context.Context, authority *store.Store, lease store.Lease, journalID, reason string) error {
	_, transitionErr := authority.TransitionPromotionJournal(ctx, lease, journalID, store.PromotionJournalBlocked, reason)
	return errors.Join(errors.New("promotion recovery requires human intervention"), transitionErr)
}
