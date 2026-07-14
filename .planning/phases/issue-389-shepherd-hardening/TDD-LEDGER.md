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
- [ ] `cmd/shepherd/main_test.go`: successful supervise path must not create a ratified proof when
      independent validator evidence is absent.
- [ ] `cmd/shepherd/main_test.go`: stale candidate head leaves canonical branch unchanged.

GREEN evidence (partial store hardening only):
- PASS `cd agent-runtime/shepherd && gofmt -w internal/store/proof_test.go internal/store/store.go && go test ./internal/store -run 'TestArtifactProofRejectsUnratifiedResult|TestAttestationRejectsNonProceedVerdicts|TestArtifactProofBindsExactHeadsAndRatification|TestAttestationPersistsValidatorProof' -count=1`
- Production change: `PutArtifactProof` now rejects unratified proofs; `PutAttestation` now rejects non-`PROCEED` verdicts.

Refactor evidence: pending; production `persistSuccessProof` still requires follow-up because it manufactures validator/ratification evidence.

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
