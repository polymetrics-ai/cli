# Issue 389 TDD Ledger — proof-recovery repair

## Baseline for this repair run

- Branch: `fix/389-shepherd-proof-recovery`
- Start head: `db13cbaa8e27cbc86130ce2547f3e60b82b5217c`
- GSD command: `scripts/gsd prompt programming-loop init --phase issue-389-shepherd-hardening --dry-run`
- GSD adapter checks run: `scripts/gsd doctor`, `scripts/gsd list`
- Current evidence status: prior claims of independent validation, ratification, recovery planning,
  final verification, and canary readiness are invalid for this repair run until re-proven against a
  new exact candidate head.

## Skills recorded

`gsd-core`, `polymetrics-issue-delivery`, `gsd-programming-loop`, `golang-how-to`,
`golang-testing`, `golang-error-handling`, `golang-safety`, `golang-security`, `golang-context`,
`golang-concurrency`, `golang-design-patterns`, `golang-structs-interfaces`,
`golang-observability`, `golang-lint`.

## Read-only recon evidence

- Scout finding: `cmd/shepherd/main.go:persistSuccessProof` currently writes
  `Validator="openai-codex/gpt-5.6-sol"`, `Thinking="high"`, `Verdict="PROCEED"`, and
  `Ratified=true` directly.
- Scout finding: `internal/authority/ratification.go` contains `authority.Ratify`, but production
  code does not call it.
- Scout finding: store proof tests do not yet reject missing ratification or non-PROCEED verdicts.
- Orchestration decision: start with RED tests around proof/attestation before any production edit.

## Slice A — Real independent validation and ratification

RED tests:
- [x] `internal/store/proof_test.go`: reject `ArtifactProof{Ratified:false}`.
- [x] `internal/store/proof_test.go`: reject attestation verdicts other than real successful
      ratification (for example `RETRY`/`HALT`).

RED evidence:
- FAIL (expected) `cd agent-runtime/shepherd && go test ./internal/store -run 'TestArtifactProofRejectsUnratifiedResult|TestAttestationRejectsNonProceedVerdicts' -count=1`
  - `TestArtifactProofRejectsUnratifiedResult`: `unratified artifact proof accepted`
- [x] `cmd/shepherd/main_test.go`: successful supervise path must not create a ratified proof when
      independent validator evidence is absent.
- [x] `cmd/shepherd/main_test.go`: stale candidate head leaves canonical branch unchanged.
- [x] `cmd/shepherd/main_test.go`: GPT-5.5 validator evidence is rejected.
- [x] `cmd/shepherd/main_test.go`: `RETRY` and `HALT` verdicts are rejected.
- [x] `cmd/shepherd/main_test.go`: every rejected path leaves canonical HEAD and canonical `.gsd`
      unchanged.
- [x] `cmd/shepherd/main_test.go`: successful path validates before promotion and persists proof plus
      attestation after ratification.

RED evidence:
- FAIL (expected) `cd agent-runtime/shepherd && go test ./cmd/shepherd -run 'TestSuperviseRejectsInvalidIndependentValidationWithoutPromotion|TestSuperviseRatifiesBeforePromotingCandidate' -count=1`
  - compile failed because `internal/validation` and `independentValidatorFactory` did not exist yet.

GREEN evidence (partial store hardening only):
- PASS `cd agent-runtime/shepherd && gofmt -w internal/store/proof_test.go internal/store/store.go && go test ./internal/store -run 'TestArtifactProofRejectsUnratifiedResult|TestAttestationRejectsNonProceedVerdicts|TestArtifactProofBindsExactHeadsAndRatification|TestAttestationPersistsValidatorProof' -count=1`
- Production change: `PutArtifactProof` now rejects unratified proofs; `PutAttestation` now rejects non-`PROCEED` verdicts.

False-green evidence for commit `19d051c6`:
- The previous Slice A completion claim is invalid. Tests only proved fake-validator port behavior and
  did not test the real production validator implementation.
- Production validation had no real result producer; it invoked canonical `validate-milestone`, which can
  mutate GSD workflow state and does not produce Shepherd validation results.
- Production trusted worker-controlled `.gsd/shepherd-validation.json`, fabricated fallback session IDs,
  used generation as state version, hard-coded PR base as `main`, and blindly required/claimed UAT.
- Slice A remains open; Slice B, PR creation, final Sol review, and canaries remain blocked.

Retry RED tests added against the real production validator:
- [x] `internal/validation`: no validation-result producer exists.
- [x] `internal/validation`: stale pre-existing result is ignored/rejected.
- [x] `internal/validation`: no new validator session is rejected.
- [x] `internal/validation`: validator session model GPT-5.5 is rejected.
- [x] `internal/validation`: thinking below high is rejected.
- [x] `internal/validation`: result head/evidence/request nonce mismatch is rejected.
- [x] `internal/validation`: candidate moving during validation is rejected.
- [x] `internal/validation`: stale base branch/governance state version is rejected.
- [x] `internal/validation`/`internal/authority`: RETRY/HALT or missing required gates are rejected.
- [x] `internal/validation`: failed validation paths leave candidate `.gsd` unchanged and do not mutate Git except the explicit candidate-move fixture.

RED evidence: these tests target production `validation.GSDValidator` with a helper process and would fail against `19d051c6` because it used canonical `validate-milestone`, trusted worker-local `.gsd/shepherd-validation.json`, accepted derived session IDs, lacked protected nonce-bound result transport, and did not bind state/base/gates.

Corrected Slice A GREEN evidence:
- Added dedicated non-canonical validator process launch (`headless shepherd-validate`) with Sol/high, exact candidate worktree, bounded request path, and exclusive result path.
- Shepherd now writes nonce-bound validation requests under protected state, outside the candidate worktree, and rejects stale/reused/malformed result files.
- Removed derived/fabricated validator session identity; a new session after validator start with exact worktree, Sol/high model, and high thinking is mandatory.
- Result proof binds request ID, nonce, base branch/head, candidate/observed head, durable governance state version, contract/evidence hash, verdict, gates, issue time, and expiry.
- Extended durable attestation schema/migration to persist repository, PR, base branch, base/candidate/observed head, delivery, generation, unit, attempt, state version, hashes, session ID, model/thinking, verdict, gates, issued, and expiry.
- Required gates are derived from official unit metadata; UAT is required only for UAT phase metadata.
- PASS `cd agent-runtime/shepherd && go test ./internal/validation ./internal/store ./internal/authority ./internal/workspace ./cmd/shepherd`.
- PASS `cd agent-runtime/shepherd && go test ./...`.
- PASS `cd agent-runtime/shepherd && go test -race ./...`.
- PASS `cd agent-runtime/shepherd && go vet ./...`.
- FAIL `cd agent-runtime/shepherd && golangci-lint run ./...` with the same 30 pre-existing findings; the earlier new `validation/validator.go` staticcheck finding was fixed.
- PASS `cd agent-runtime/shepherd && go build ./cmd/shepherd && make verify && cd ../.. && scripts/tests/shepherd-module-boundary.sh && git diff --check && go list ./...`.

## Slice B — Durable attempt lifecycle and crash recovery

Planned RED tests:
- [ ] attempt state enum/store supports `created`, `prepared`, `running`, `validated`, `ratified`,
      `promoting`, `promoted`, `retained_for_recovery`, `cleanup_pending`, `cleanup_complete`, and
      `cleanup_blocked`.
- [ ] restart reconciles database-owned orphan worktrees/branches and leaves unknown/live worktrees
      untouched.
- [ ] preparation/query/runtime failures transition explicitly.
- [ ] retry after retained failure creates a fresh attempt worktree.

GREEN evidence: pending.
Refactor evidence: pending.

## Slice C — Crash-safe GSD-state promotion

Planned RED tests:
- [ ] inject failure before Git promotion.
- [ ] inject failure after Git promotion.
- [ ] inject failure before state swap.
- [ ] inject failure after state swap.
- [ ] restart is idempotent and converges to one consistent Git/GSD state.

GREEN evidence: pending.
Refactor evidence: pending.

## Slice D — Official GSD 1.11 registry loading

Planned RED tests:
- [ ] parse real pinned GSD 1.11 registry fixture with array spreads.
- [ ] preserve allowed, required, and forbidden tools.
- [ ] route models only from official metadata.
- [ ] null/unknown units fail closed or are explicitly governed sidecars.

GREEN evidence: pending.
Refactor evidence: pending.

## Slice E — Real Sol/high recovery planning

Planned RED tests:
- [ ] static recovery-planning text is insufficient.
- [ ] planner result must include observed model/thinking, evidence hash, typed action, and bounded
      plan.
- [ ] unallowlisted action fails closed.
- [ ] budget exhaustion persists `awaiting_decision` across restart.

GREEN evidence: pending.
Refactor evidence: pending.

## Slice F — Authority-gated external effects

Planned RED tests:
- [ ] no direct `SyncDecisionComment` production path.
- [ ] outbox pending, claim, send, failure, restart, and idempotent replay.
- [ ] worker ports cannot directly mutate GitHub.

GREEN evidence: pending.
Refactor evidence: pending.

## Slice G — Real supervise integration coverage

Planned RED tests:
- [ ] success path reaches `final_human_gate` only after independent Sol/high validation,
      ratification, and promotion.
- [ ] missing/GPT-5.5 validator evidence fails without canonical mutation.
- [ ] stale candidate head fails without canonical mutation.
- [ ] validator `RETRY`/`HALT` fails without canonical mutation.
- [ ] crash/restart at each promotion boundary.
- [ ] retained failed attempt followed by fresh attempt.
- [ ] recovery planning and `awaiting_decision` restart.
- [ ] outbox restart and duplicate suppression.
- [ ] official registry spread metadata.
- [ ] canonical worktree unchanged on every failed path.

GREEN evidence: pending.
Refactor evidence: pending.

## Verification log

No production-code verification is claimed yet for this repair run. Focused RED tests will be recorded
before the first production edit, followed by focused GREEN and full nested-module gates after each
coherent slice.
