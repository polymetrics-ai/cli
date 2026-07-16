package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	shepherdgithub "github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/github"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/store"
	_ "modernc.org/sqlite"
)

func TestCompletedCleanPromotionResolutionPersistsAcrossRestart(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "" {
		return
	}
	repo, contextPath, config, runner := setupFakeSuperviseRuntime(t)
	_ = repo
	restoreValidator := installFakeIndependentValidator(t, &fakeIndependentValidator{
		result: validFakeValidationResult("openai-codex/gpt-5.6-sol", "PROCEED"),
	})
	defer restoreValidator()
	if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath,
		false, "shepherd", "fake runtime"); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(config.StateDir, "authority.db")
	if count := countSQLiteQueryForTest(t, dbPath, "SELECT COUNT(*) FROM promotion_journals WHERE state = 'complete' AND cleanup_complete = 1 AND decisions_resolved = 1"); count != 1 {
		t.Fatalf("resolved completed journals=%d want 1", count)
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE promotion_journals SET decisions_resolved = 0`); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath,
		false, "shepherd", "fake runtime"); err != nil {
		t.Fatalf("resolution-only restart: %v", err)
	}
	if count := countSQLiteQueryForTest(t, dbPath, "SELECT COUNT(*) FROM promotion_journals WHERE decisions_resolved = 0"); count != 0 {
		t.Fatalf("unresolved journals after restart=%d", count)
	}
}

func TestPromotionFailureBoundariesRecoverForwardIdempotently(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "" {
		return
	}
	boundaries := []promotionBoundary{
		promotionAfterJournalCreated,
		promotionAfterStateStaged,
		promotionBeforeGit,
		promotionAfterGit,
		promotionAfterGitJournaled,
		promotionBeforeBackupRename,
		promotionAfterBackupRename,
		promotionAfterStateInstall,
		promotionAfterStateJournaled,
		promotionAfterJournalComplete,
	}
	for _, boundary := range boundaries {
		t.Run(string(boundary), func(t *testing.T) {
			repo, contextPath, config, runner := setupFakeSuperviseRuntime(t)
			baseHead := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD"))
			oldState := readFileForTest(t, filepath.Join(repo, ".gsd", "STATE.md"))
			restoreValidator := installFakeIndependentValidator(t, &fakeIndependentValidator{result: validFakeValidationResult("openai-codex/gpt-5.6-sol", "PROCEED")})
			defer restoreValidator()
			injected := errors.New("injected promotion stop")
			restoreFailpoint := installPromotionFailpoint(t, func(observed promotionBoundary) error {
				if observed == boundary {
					return injected
				}
				return nil
			})
			err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime")
			restoreFailpoint()
			if err == nil {
				t.Fatal("injected promotion boundary unexpectedly succeeded")
			}
			dbPath := filepath.Join(config.StateDir, "authority.db")
			wantIncomplete := 1
			if boundary == promotionAfterJournalComplete {
				wantIncomplete = 0
			}
			if count := countSQLiteQueryForTest(t, dbPath, "SELECT COUNT(*) FROM promotion_journals WHERE state <> 'complete'"); count != wantIncomplete {
				t.Fatalf("incomplete journals=%d want %d", count, wantIncomplete)
			}
			if boundary == promotionAfterJournalCreated || boundary == promotionAfterStateStaged || boundary == promotionBeforeGit {
				decisionReplyPollerFactory = func() decisionReplyPoller {
					return decisionReplyPollerFunc(func(_ context.Context, request shepherdgithub.QuestionRequest,
						actor string,
					) ([]shepherdgithub.Reply, error) {
						return []shepherdgithub.Reply{{RequestID: request.RequestID, Option: "retry", Author: actor,
							AuthorID: 6113982, CommentID: request.QuestionCommentID + 1,
							CreatedAt: time.Now().UTC()}}, nil
					})
				}
				if got := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD")); got != baseHead {
					t.Fatalf("HEAD=%s want base=%s", got, baseHead)
				}
				if got := readFileForTest(t, filepath.Join(repo, ".gsd", "STATE.md")); got != oldState {
					t.Fatalf("state changed before git: %q", got)
				}
			}
			if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err != nil {
				t.Fatalf("restart recovery: %v", err)
			}
			candidateHead := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD"))
			if candidateHead == baseHead {
				t.Fatal("Git did not converge to candidate")
			}
			if got := readFileForTest(t, filepath.Join(repo, ".gsd", "STATE.md")); got != "fake completed state\n" {
				t.Fatalf("installed state=%q", got)
			}
			if count := countSQLiteQueryForTest(t, dbPath, "SELECT COUNT(*) FROM promotion_journals WHERE state = 'complete'"); count != 1 {
				t.Fatalf("complete journals=%d", count)
			}
			if count := countSQLiteQueryForTest(t, dbPath, "SELECT COUNT(*) FROM attempt_worktrees WHERE state = 'cleanup_complete'"); count != 1 {
				t.Fatalf("cleanup-complete attempts=%d", count)
			}
			if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err != nil {
				t.Fatalf("idempotent restart: %v", err)
			}
			if got := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD")); got != candidateHead {
				t.Fatalf("duplicate Git effect: %s != %s", got, candidateHead)
			}
		})
	}
}

func TestPreGitHardCrashWithoutDecisionRecoversDeterministically(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "" {
		return
	}
	repo, contextPath, config, runner := setupFakeSuperviseRuntime(t)
	baseHead := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD"))
	restoreValidator := installFakeIndependentValidator(t, &fakeIndependentValidator{
		result: validFakeValidationResult("openai-codex/gpt-5.6-sol", "PROCEED"),
	})
	defer restoreValidator()
	restoreFailpoint := installPromotionFailpoint(t, failAtPromotionBoundary(promotionBeforeGit))
	if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath,
		false, "shepherd", "fake runtime"); err == nil {
		t.Fatal("injected pre-Git boundary succeeded")
	}
	restoreFailpoint()
	db, err := sql.Open("sqlite", filepath.Join(config.StateDir, "authority.db"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM decision_requests`); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE delivery_runs SET state = 'running', owner = 'crashed-controller'`); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD")); got != baseHead {
		t.Fatalf("precondition HEAD=%s want %s", got, baseHead)
	}
	if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath,
		false, "shepherd", "fake runtime"); err != nil {
		t.Fatalf("hard-crash restart recovery: %v", err)
	}
	candidateHead := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD"))
	if candidateHead == baseHead {
		t.Fatal("hard-crash restart did not promote candidate")
	}
	if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath,
		false, "shepherd", "fake runtime"); err != nil {
		t.Fatalf("idempotent hard-crash restart: %v", err)
	}
	if got := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD")); got != candidateHead {
		t.Fatalf("second restart changed promoted HEAD: %s != %s", got, candidateHead)
	}
}

func TestPreGitPromotionRecoveryWaitsForUnexpiredDecision(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "" {
		return
	}
	repo, contextPath, config, runner := setupFakeSuperviseRuntime(t)
	baseHead := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD"))
	restoreValidator := installFakeIndependentValidator(t, &fakeIndependentValidator{
		result: validFakeValidationResult("openai-codex/gpt-5.6-sol", "PROCEED"),
	})
	defer restoreValidator()
	restoreFailpoint := installPromotionFailpoint(t, failAtPromotionBoundary(promotionBeforeGit))
	if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath,
		false, "shepherd", "fake runtime"); err == nil {
		t.Fatal("injected pre-Git boundary succeeded")
	}
	restoreFailpoint()
	if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath,
		false, "shepherd", "fake runtime"); err == nil || !strings.Contains(err.Error(), "decision defers") {
		t.Fatalf("unanswered pre-Git decision did not defer recovery: %v", err)
	}
	if got := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD")); got != baseHead {
		t.Fatalf("unanswered pre-Git decision advanced HEAD: %s != %s", got, baseHead)
	}
	if count := countSQLiteQueryForTest(t, filepath.Join(config.StateDir, "authority.db"),
		"SELECT COUNT(*) FROM decision_requests WHERE status = 'published'"); count != 1 {
		t.Fatalf("published pre-Git decisions=%d want 1", count)
	}
}

func TestPromotionRecoveryRevalidatesStageBeforeGit(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "" {
		return
	}
	repo, contextPath, config, runner := setupFakeSuperviseRuntime(t)
	base := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD"))
	oldState := readFileForTest(t, filepath.Join(repo, ".gsd", "STATE.md"))
	restoreValidator := installFakeIndependentValidator(t, &fakeIndependentValidator{result: validFakeValidationResult("openai-codex/gpt-5.6-sol", "PROCEED")})
	defer restoreValidator()
	restoreFailpoint := installPromotionFailpoint(t, failAtPromotionBoundary(promotionBeforeGit))
	if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err == nil {
		t.Fatal("injected pre-Git boundary succeeded")
	}
	restoreFailpoint()
	db, err := sql.Open("sqlite", filepath.Join(config.StateDir, "authority.db"))
	if err != nil {
		t.Fatal(err)
	}
	var stagePath string
	if err := db.QueryRow(`SELECT stage_path FROM promotion_journals WHERE state = 'state_staged'`).Scan(&stagePath); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stagePath, "STATE.md"), []byte("tampered\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err == nil {
		t.Fatal("tampered stage advanced Git")
	}
	if got := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD")); got != base {
		t.Fatalf("Git advanced before staged-state validation: %s", got)
	}
	if got := readFileForTest(t, filepath.Join(repo, ".gsd", "STATE.md")); got != oldState {
		t.Fatalf("canonical GSD changed: %q", got)
	}
	if count := countSQLiteQueryForTest(t, filepath.Join(config.StateDir, "authority.db"), "SELECT COUNT(*) FROM promotion_journals WHERE state = 'blocked'"); count != 1 {
		t.Fatalf("blocked journals=%d", count)
	}
}

func TestPromotionRecoveryMovedHeadFailsClosed(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "" {
		return
	}
	repo, contextPath, config, runner := setupFakeSuperviseRuntime(t)
	restoreValidator := installFakeIndependentValidator(t, &fakeIndependentValidator{result: validFakeValidationResult("openai-codex/gpt-5.6-sol", "PROCEED")})
	defer restoreValidator()
	restoreFailpoint := installPromotionFailpoint(t, failAtPromotionBoundary(promotionBeforeGit))
	if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err == nil {
		t.Fatal("injected failure succeeded")
	}
	restoreFailpoint()
	if err := os.WriteFile(filepath.Join(repo, "moved.txt"), []byte("moved\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitForTest(t, repo, "add", "moved.txt")
	runGitForTest(t, repo, "commit", "-m", "move canonical head")
	moved := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD"))
	if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err == nil {
		t.Fatal("moved head recovered")
	}
	if got := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD")); got != moved {
		t.Fatalf("moved head changed: %s", got)
	}
	if count := countSQLiteQueryForTest(t, filepath.Join(config.StateDir, "authority.db"), "SELECT COUNT(*) FROM promotion_journals WHERE state = 'blocked'"); count != 1 {
		t.Fatalf("blocked journals=%d", count)
	}
}

func TestPromotionRecoveryDirtyCanonicalFailsClosed(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "" {
		return
	}
	repo, contextPath, config, runner := setupFakeSuperviseRuntime(t)
	restoreValidator := installFakeIndependentValidator(t, &fakeIndependentValidator{result: validFakeValidationResult("openai-codex/gpt-5.6-sol", "PROCEED")})
	defer restoreValidator()
	restoreFailpoint := installPromotionFailpoint(t, failAtPromotionBoundary(promotionBeforeGit))
	if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err == nil {
		t.Fatal("injected failure succeeded")
	}
	restoreFailpoint()
	if err := os.WriteFile(filepath.Join(repo, "dirty.txt"), []byte("dirty\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err == nil {
		t.Fatal("dirty canonical recovered")
	}
	if _, err := os.Stat(filepath.Join(repo, "agent-runtime", "shepherd", "canary.txt")); !os.IsNotExist(err) {
		t.Fatalf("candidate promoted into dirty canonical: %v", err)
	}
}

func TestPromotionRecoveryMovedOrDirtyAfterGitFailsClosed(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "" {
		return
	}
	for _, test := range []struct {
		name     string
		boundary promotionBoundary
		mutate   func(*testing.T, string)
	}{
		{name: "moved", boundary: promotionAfterGit, mutate: func(t *testing.T, repo string) {
			if err := os.WriteFile(filepath.Join(repo, "moved-after-git.txt"), []byte("moved\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			runGitForTest(t, repo, "add", "moved-after-git.txt")
			runGitForTest(t, repo, "commit", "-m", "move after promotion")
		}},
		{name: "dirty", boundary: promotionAfterGit, mutate: func(t *testing.T, repo string) {
			if err := os.WriteFile(filepath.Join(repo, "dirty-after-git.txt"), []byte("dirty\n"), 0o600); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "dirty_after_state_installed", boundary: promotionAfterStateJournaled, mutate: func(t *testing.T, repo string) {
			if err := os.WriteFile(filepath.Join(repo, "dirty-after-state.txt"), []byte("dirty\n"), 0o600); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo, contextPath, config, runner := setupFakeSuperviseRuntime(t)
			restoreValidator := installFakeIndependentValidator(t, &fakeIndependentValidator{result: validFakeValidationResult("openai-codex/gpt-5.6-sol", "PROCEED")})
			defer restoreValidator()
			restoreFailpoint := installPromotionFailpoint(t, failAtPromotionBoundary(test.boundary))
			if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err == nil {
				t.Fatal("injected post-Git boundary succeeded")
			}
			restoreFailpoint()
			test.mutate(t, repo)
			if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err == nil {
				t.Fatal("moved/dirty post-Git state recovered")
			}
			if count := countSQLiteQueryForTest(t, filepath.Join(config.StateDir, "authority.db"), "SELECT COUNT(*) FROM promotion_journals WHERE state = 'blocked'"); count != 1 {
				t.Fatalf("blocked journals=%d", count)
			}
			if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath,
				false, "shepherd", "fake runtime"); err == nil {
				t.Fatal("second restart bypassed blocked promotion journal")
			}
			if count := countSQLiteQueryForTest(t, filepath.Join(config.StateDir, "authority.db"), "SELECT COUNT(*) FROM promotion_journals WHERE state = 'blocked'"); count != 1 {
				t.Fatalf("blocked journals after second restart=%d", count)
			}
		})
	}
}

func TestPromotionRecoveryExpiredBeforeGitFailsClosed(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "" {
		return
	}
	repo, contextPath, config, runner := setupFakeSuperviseRuntime(t)
	base := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD"))
	restoreValidator := installFakeIndependentValidator(t, &fakeIndependentValidator{result: validFakeValidationResult("openai-codex/gpt-5.6-sol", "PROCEED")})
	defer restoreValidator()
	restoreFailpoint := installPromotionFailpoint(t, failAtPromotionBoundary(promotionBeforeGit))
	if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err == nil {
		t.Fatal("injected failure succeeded")
	}
	restoreFailpoint()
	db, err := sql.Open("sqlite", filepath.Join(config.StateDir, "authority.db"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE attestations SET expires_at = 1`); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err == nil {
		t.Fatal("expired pre-Git authority recovered")
	}
	if got := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD")); got != base {
		t.Fatalf("expired proof changed Git head: %s", got)
	}
	if count := countSQLiteQueryForTest(t, filepath.Join(config.StateDir, "authority.db"), "SELECT COUNT(*) FROM promotion_journals WHERE state = 'blocked'"); count != 1 {
		t.Fatalf("blocked journals=%d", count)
	}
}

func TestValidatedGSDManifestHashDistinguishesLegacyFromCorruption(t *testing.T) {
	legacyRaw := `{"unit_type":"execute-task","phase_chain":["execution"],"required_workflow_tools":["gsd_task_complete"],"artifacts":[{"path":"STATE.md","hash":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}]}`
	legacyDigest := sha256.Sum256([]byte(legacyRaw))
	_, err := validatedGSDManifestHash(store.ArtifactProof{ExpectedArtifact: legacyRaw,
		ArtifactHash: "sha256:" + hex.EncodeToString(legacyDigest[:])})
	if !errors.Is(err, errLegacyGSDManifest) {
		t.Fatalf("legacy manifest err=%v", err)
	}
	for _, raw := range []string{
		"{}", "null", legacyRaw + " trailing",
		strings.Replace(legacyRaw, `"artifacts"`, `"gsd_manifest_hash":null,"artifacts"`, 1),
		strings.Replace(legacyRaw, `"artifacts"`, `"gsd_manifest_hash":"","artifacts"`, 1),
		strings.Replace(legacyRaw, `"artifacts"`, `"gsd_manifest_hash":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","gsd_manifest_hash":"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","artifacts"`, 1),
	} {
		digest := sha256.Sum256([]byte(raw))
		if _, err := validatedGSDManifestHash(store.ArtifactProof{ExpectedArtifact: raw,
			ArtifactHash: "sha256:" + hex.EncodeToString(digest[:])}); err == nil ||
			errors.Is(err, errLegacyGSDManifest) {
			t.Fatalf("corrupt manifest %q err=%v", raw, err)
		}
	}
}

func TestPromotionRecoveryBlocksLegacyProofBeforeGit(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "" {
		return
	}
	repo, contextPath, config, runner := setupFakeSuperviseRuntime(t)
	base := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD"))
	restoreValidator := installFakeIndependentValidator(t, &fakeIndependentValidator{
		result: validFakeValidationResult("openai-codex/gpt-5.6-sol", "PROCEED"),
	})
	defer restoreValidator()
	restoreFailpoint := installPromotionFailpoint(t, failAtPromotionBoundary(promotionBeforeGit))
	if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath,
		false, "shepherd", "fake runtime"); err == nil {
		t.Fatal("injected failure succeeded")
	}
	restoreFailpoint()
	db, err := sql.Open("sqlite", filepath.Join(config.StateDir, "authority.db"))
	if err != nil {
		t.Fatal(err)
	}
	downgradePromotionProofToLegacy(t, db)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath,
		false, "shepherd", "fake runtime"); err == nil {
		t.Fatal("legacy pre-Git proof recovered without normalized state binding")
	}
	if got := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD")); got != base {
		t.Fatalf("legacy pre-Git recovery changed head: %s != %s", got, base)
	}
	if count := countSQLiteQueryForTest(t, filepath.Join(config.StateDir, "authority.db"),
		"SELECT COUNT(*) FROM promotion_journals WHERE state = 'blocked' AND blocked_reason = 'legacy_pre_git_manifest_requires_human_reconciliation'"); count != 1 {
		t.Fatalf("blocked legacy journals=%d", count)
	}
}

func TestPromotionRecoveryAfterGitCompletesLegacyProofForward(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "" {
		return
	}
	repo, contextPath, config, runner := setupFakeSuperviseRuntime(t)
	base := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD"))
	validator := &fakeIndependentValidator{result: validFakeValidationResult("openai-codex/gpt-5.6-sol", "PROCEED")}
	restoreValidator := installFakeIndependentValidator(t, validator)
	defer restoreValidator()
	restoreFailpoint := installPromotionFailpoint(t, failAtPromotionBoundary(promotionAfterGit))
	if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err == nil {
		t.Fatal("injected failure succeeded")
	}
	restoreFailpoint()
	if got := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD")); got == base {
		t.Fatal("Git was not promoted before boundary")
	}
	db, err := sql.Open("sqlite", filepath.Join(config.StateDir, "authority.db"))
	if err != nil {
		t.Fatal(err)
	}
	downgradePromotionProofToLegacy(t, db)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	calls := validator.calls
	if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err != nil {
		t.Fatalf("forward recovery: %v", err)
	}
	if validator.calls != calls {
		t.Fatalf("recovery requested new verdict: calls %d -> %d", calls, validator.calls)
	}
	if got := readFileForTest(t, filepath.Join(repo, ".gsd", "STATE.md")); got != "fake completed state\n" {
		t.Fatalf("state=%q", got)
	}
}

func downgradePromotionProofToLegacy(t *testing.T, db *sql.DB) {
	t.Helper()
	var proofID, expectedArtifact string
	if err := db.QueryRow(`SELECT proof_id, expected_artifact FROM artifact_proofs`).Scan(&proofID, &expectedArtifact); err != nil {
		t.Fatal(err)
	}
	var legacyManifest map[string]any
	if err := json.Unmarshal([]byte(expectedArtifact), &legacyManifest); err != nil {
		t.Fatal(err)
	}
	delete(legacyManifest, "gsd_manifest_hash")
	legacyRaw, err := json.Marshal(legacyManifest)
	if err != nil {
		t.Fatal(err)
	}
	legacyDigest := sha256.Sum256(legacyRaw)
	legacyHash := "sha256:" + hex.EncodeToString(legacyDigest[:])
	if _, err := db.Exec(`UPDATE artifact_proofs SET expected_artifact = ?, artifact_hash = ? WHERE proof_id = ?`,
		string(legacyRaw), legacyHash, proofID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE attestations SET evidence_hash = ?, expires_at = 1`, legacyHash); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE promotion_journals SET evidence_hash = ?`, legacyHash); err != nil {
		t.Fatal(err)
	}
}

func TestPromotingAttemptWithoutJournalRemainsHumanGated(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "" {
		return
	}
	_, contextPath, config, runner := setupFakeSuperviseRuntime(t)
	restoreValidator := installFakeIndependentValidator(t, &fakeIndependentValidator{result: validFakeValidationResult("openai-codex/gpt-5.6-sol", "PROCEED")})
	defer restoreValidator()
	restoreFailpoint := installPromotionFailpoint(t, failAtPromotionBoundary(promotionAfterJournalCreated))
	if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err == nil {
		t.Fatal("injected journal boundary succeeded")
	}
	restoreFailpoint()
	dbPath := filepath.Join(config.StateDir, "authority.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	var attemptPath string
	if err := db.QueryRow(`SELECT path FROM attempt_worktrees WHERE state = 'promoting'`).Scan(&attemptPath); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM promotion_journals`); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := runSupervise(context.Background(), runner, config, testUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err == nil {
		t.Fatal("promoting attempt without journal did not human-gate")
	}
	if _, err := os.Stat(attemptPath); err != nil {
		t.Fatalf("ambiguous no-journal attempt was changed: %v", err)
	}
	if count := countSQLiteQueryForTest(t, dbPath, "SELECT COUNT(*) FROM promotion_journals"); count != 0 {
		t.Fatalf("unexpected promotion journals=%d", count)
	}
}

func installPromotionFailpoint(t *testing.T, failpoint func(promotionBoundary) error) func() {
	t.Helper()
	previous := promotionFailureInjector
	promotionFailureInjector = failpoint
	return func() { promotionFailureInjector = previous }
}

func failAtPromotionBoundary(want promotionBoundary) func(promotionBoundary) error {
	return func(observed promotionBoundary) error {
		if observed == want {
			return errors.New("injected promotion stop")
		}
		return nil
	}
}
