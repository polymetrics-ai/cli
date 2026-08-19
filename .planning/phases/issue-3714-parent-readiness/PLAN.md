# PLAN — issue #3714 parent readiness

Issue: #3714. Parent PR: #3723. Active branch:
`refactor/3714-canonical-delivery-flow`.

## GSD lifecycle

- `scripts/gsd doctor`: passed.
- `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review|ship`:
  resolved through the installed adapter.
- `scripts/gsd prompt discuss-phase issue-3714-parent-readiness` and
  `scripts/gsd prompt plan-phase issue-3714-parent-readiness --tdd`: generated and executed
  inline.
- `programming-loop` is intentionally not invoked: it is absent from the installed adapter and
  excluded by the canonical contract. This phase follows the installed
  `discuss-phase → plan-phase --tdd → execute-phase → verify-work → code-review` sequence.
- The canonical delivery contract prohibits spawned roles. Execution decision: `local_critical_path`.

## Required skills loaded

- `golang-how-to` — routes this Go generator/integration validation task.
- `golang-testing` — treats ancestry and generator drift as observable integration contracts.
- `golang-error-handling`, `golang-safety`, `golang-security`, and `golang-lint` — preserve
  bounded conflict resolution, generated-file safety, the write-confirmation boundary, and static
  validation.
- `gsd-programming-loop` — provides the strict TDD artifact lifecycle; the installed adapter
  command itself is unavailable and is not fabricated.

## RED → GREEN → REFACTOR

### Slice 1 — establish the parent readiness gap

1. **RED:** prove the current parent does not yet contain `origin/main` with
   `git merge-base --is-ancestor origin/main HEAD`; record its nonzero result.
2. **GREEN:** merge current `origin/main` into the existing parent branch. Resolve conflicts by
   preserving mainline destructive-write confirmation behavior and the canonical contract's complete
   Codex, Claude, and Pi metadata.
3. **REFACTOR:** run `go run ./cmd/agentcontractgen sync` only if the regenerated projection set
   differs, then use the drift checker to prove every registered projection matches the source.

### Slice 2 — focused verification and parent readiness

1. Re-run the ancestry assertion; it must now pass.
2. Run the generator drift check and focused `agentcontract`/generator tests, plus tests covering
   packages changed by current `main` that overlap the merge.
3. Run the scoped repository gates appropriate to the integrated parent; do not replace full CI
   with an assertion of local full-suite success.
4. Commit the reconciliation and phase evidence on the existing parent branch, push that branch,
   let PR #3723 CI finish, then mark that same PR ready only if all checks are green.

## Conflict rules

- Never resolve a conflict by deleting an entry from `harness_policies`, `codex`, `pi_harness`, or
  `projections`. The final canonical source must render both roles for all three harnesses.
- Never hand-edit a generated projection. Edit only the canonical source when required, regenerate
  through `agentcontractgen sync`, then run `agentcontractgen check`.
- Preserve #3730's destructive-write confirmation checks and tests as `main` authored them. Any
  interaction must tighten compatibility, never bypass confirmation.

## Verification plan

- `git merge-base --is-ancestor origin/main HEAD` (red before merge, green after merge).
- `go run ./cmd/agentcontractgen sync` followed by `go run ./cmd/agentcontractgen check`.
- `go test ./internal/agentcontract ./cmd/agentcontractgen`.
- Focused package tests selected from the #3730/#3731/#3726 merge paths after conflict inspection.
- `go vet` on affected packages, `make agent-contract-check`, and applicable non-monolithic
  repository gates; remote CI owns the full suite.
- `git diff --check`, `git status --short`, PR checks, and full-file generated-projection audit.

## Commit / push checkpoints

1. Plan/TDD/verification checkpoint before merge work.
2. Reconciliation + regeneration checkpoint after focused validation.
3. Review-fix checkpoint only if the code-review or no-mistakes pipeline applies an in-scope fix.

## Refresh cycle — 2026-08-05

`origin/main` advanced to `4871df2f8` (Recurly parity) after the prior clean parent
integration. The new change set is outside the canonical delivery contract and all generated
Codex, Claude, and Pi projections, but the parent is deliberately refreshed again rather than
claiming stale ancestry.

1. **RED:** `git merge-base --is-ancestor origin/main HEAD` exited nonzero at
   `554235545`.
2. **GREEN:** merge `origin/main` into this same parent branch without rewriting its history.
3. **REFACTOR:** regenerate with `agentcontractgen sync`; use its drift check and the clean Pi
   agent check to prove all registered projections remain source-derived and complete.
