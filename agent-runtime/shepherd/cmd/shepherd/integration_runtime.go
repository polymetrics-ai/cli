//go:build integration

package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// integrationGSDExecutor exposes one compile-time-only process seam for the
// build-tagged black-box harness. Release binaries compile the no-op variant.
func integrationGSDExecutor() (string, error) {
	if err := configureIntegrationCrashBoundary(); err != nil {
		return "", err
	}
	configured := os.Getenv("SHEPHERD_INTEGRATION_GSD_EXECUTOR")
	if configured == "" {
		return "", nil
	}
	if !filepath.IsAbs(configured) || filepath.Clean(configured) != configured ||
		strings.ContainsAny(configured, "\r\n\x00") {
		return "", errors.New("integration GSD executor must be one clean absolute path")
	}
	resolved, err := filepath.EvalSymlinks(configured)
	if err != nil {
		return "", err
	}
	info, err := os.Lstat(resolved)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 {
		return "", errors.New("integration GSD executor must be an executable regular file")
	}
	return resolved, nil
}

func configureIntegrationCrashBoundary() error {
	configured := os.Getenv("SHEPHERD_INTEGRATION_CRASH_BOUNDARY")
	if configured == "" || configured == "final_gate_projected" || configured == "retry_persisted" {
		return nil
	}
	allowed := map[promotionBoundary]struct{}{
		promotionAfterJournalCreated: {}, promotionAfterStateStaged: {}, promotionBeforeGit: {},
		promotionAfterGit: {}, promotionAfterGitJournaled: {}, promotionBeforeBackupRename: {},
		promotionAfterBackupRename: {}, promotionAfterStateInstall: {},
		promotionAfterStateJournaled: {}, promotionAfterJournalComplete: {},
	}
	boundary := promotionBoundary(configured)
	if _, ok := allowed[boundary]; !ok {
		return errors.New("unknown integration crash boundary")
	}
	promotionFailureInjector = func(observed promotionBoundary) error {
		if observed == boundary {
			os.Exit(97)
		}
		return nil
	}
	return nil
}

func integrationFinalGateBoundary() {
	if os.Getenv("SHEPHERD_INTEGRATION_CRASH_BOUNDARY") == "final_gate_projected" {
		os.Exit(97)
	}
}

func integrationEffectTTL() time.Duration {
	if os.Getenv("SHEPHERD_INTEGRATION_SHORT_EFFECT_TTL") == "1" {
		return integrationShortEffectTTL()
	}
	return effectClaimTTL
}

func integrationDecisionTTL() time.Duration {
	if os.Getenv("SHEPHERD_INTEGRATION_SHORT_DECISION_TTL") == "1" {
		return 10 * time.Second
	}
	return 24 * time.Hour
}

func integrationEffectEnqueuedBoundary() {
	if os.Getenv("SHEPHERD_INTEGRATION_OUTBOX_CRASH") == "post_enqueue" {
		os.Exit(99)
	}
}

func integrationRetryBoundary() {
	if os.Getenv("SHEPHERD_INTEGRATION_CRASH_BOUNDARY") == "retry_persisted" {
		os.Exit(96)
	}
}

func integrationExitAwaitingDecision() bool {
	return os.Getenv("SHEPHERD_INTEGRATION_EXIT_AWAITING_DECISION") == "1"
}

func integrationPostValidationBoundary(workDir string) error {
	if os.Getenv("SHEPHERD_INTEGRATION_MUTATE_AFTER_VALIDATION") != "1" {
		return nil
	}
	return os.WriteFile(filepath.Join(workDir, ".gsd", "STATE.md"),
		[]byte("mutated after validation\n"), 0o600)
}
