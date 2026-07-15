package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	shepherdgit "github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/git"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/store"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/workspace"
)

type promotionBoundary string

const (
	promotionAfterJournalCreated  promotionBoundary = "after_journal_created"
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
	if _, err := authority.FinalizePromotionJournalStage(ctx, lease, journalID, staged.ManifestJSON,
		staged.ManifestHash, staged.BackupManifestJSON, staged.BackupManifestHash); err != nil {
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
	journals, err := authority.ListIncompletePromotionJournals(ctx, delivery.ID)
	if err != nil {
		return err
	}
	for _, journal := range journals {
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
				if _, err := authority.FinalizePromotionJournalStage(ctx, lease, journalID, staged.ManifestJSON,
					staged.ManifestHash, staged.BackupManifestJSON, staged.BackupManifestHash); err != nil {
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
