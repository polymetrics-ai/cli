# PLAN — canonical issue-first delivery contract

Issue: #3718. Parent: #3714. Branch: `fm/cli-agents-wave1-canonical-contract-r1`.

## GSD path

- `scripts/gsd doctor`: passed; 69 registered commands.
- `scripts/gsd list`: passed.
- `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review`: passed.
- `scripts/gsd sources programming-loop`: failed with `unknown GSD command or prompt`, as expected.
- Discuss prompt: `scripts/gsd prompt discuss-phase issue-3718-canonical-delivery-contract --auto`.
- Plan prompt: `scripts/gsd prompt plan-phase issue-3718-canonical-delivery-contract --tdd --skip-research`.
- Execute prompt: `scripts/gsd prompt execute-phase issue-3718-canonical-delivery-contract`;
  executed inline through the committed RED/GREEN/REFACTOR slices.
- Verify prompt: `scripts/gsd prompt verify-work issue-3718-canonical-delivery-contract`;
  automated coverage and command evidence are recorded in `SUMMARY.md`, `UAT.md`, and
  `VERIFICATION.md`.
- Review prompt: `scripts/gsd prompt code-review issue-3718-canonical-delivery-contract --files
  ...`; executed inline with dispositions in `REVIEW.md`.
- Inline execution fallback: required because this issue prohibits spawned roles. No GSD or
  specialist role will be spawned.

## Required skills loaded

- `no-mistakes` v1.41.2 — pipeline authority and exact-finding response rules.
- `golang-how-to` — routed this Go generator task.
- `golang-testing` — observable red/green contract tests and deterministic fixtures.
- `golang-cli` — stdlib flag command, exit codes, and stdout/stderr discipline.
- `golang-error-handling` — contextual errors and no swallowed write/read failures.
- `golang-safety` — bounded paths, no broad writes, and defensive validation.
- GSD phase skills/read-through: `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`,
  `gsd-verify-work`, and `gsd-code-review`, constrained by the repo-local adapter and the
  issue's no-spawn rule.

## Deliverable slices

### Slice 1 — canonical schema/source and GSD command contract

1. RED: add a command-resolution test against a temporary contract that names known-absent
   `programming-loop`; execute and record the failure.
2. GREEN: add the versioned JSON canonical source and parser/validator. Read its referenced GSD
   commands and require `scripts/gsd sources <command>` to resolve each one.
3. REFACTOR: centralize required state/authority/prohibition validation and deterministic rendering.

### Slice 2 — projection generation and drift enforcement

1. RED: add a checker stub plus a test that seeds a divergent projection and proves the current
   implementation accepts it; execute and record the failure.
2. GREEN: render canonical generated blocks, compare registered existing projections exactly, and
   add a bounded sync command that only replaces marked blocks. Missing Wave 2–4 optional targets
   remain allowed; a present divergent target fails.
3. REFACTOR: add stable-output, required-field, forbidden-delegation, and overlay-inheritance tests.
4. Wire `agent-contract-check` into `make verify` so CI executes the drift/contract check.

### Slice 3 — shared delivery policy convergence

1. Update `AGENTS.md`, GSD adapter/routing references, task skill matrix, and issue contract to the
   verified installed lifecycle; remove the stale mandatory `programming-loop` path.
2. Reframe `parent-orchestrator-contract.md` as single-worker parent/job ownership. Preserve
   child-target/check/review coverage gates, provisional integration semantics, parent full gates,
   and explicit captain approval before final merge.
3. Ensure the canonical source states the stacked draft-parent-first topology, excludes GSD ship,
   documents the no-mistakes child workaround/full-parent pipeline, encodes the exact away boundary,
   and records the Wayfinder rejection.

## TDD and verification

- Focused: `go test ./internal/agentcontract ./cmd/agentcontractgen`.
- Contract: `go run ./cmd/agentcontractgen check`.
- GSD: `scripts/gsd doctor`, `scripts/gsd list`, and `scripts/gsd sources` for every canonical command.
- Formatting/static: `gofmt -w cmd internal`, `go vet ./internal/agentcontract ./cmd/agentcontractgen`.
- Broader non-timeout gates: `go test ./internal/agentcontract ./cmd/agentcontractgen`,
  `make tidy-check`, `make release-workflow-check`, and `make agent-contract-check`.
- Policy grep: no changed canonical issue-first policy surface invokes `programming-loop`;
  forbidden role projections and Wave 2–4 harness files are absent from this diff.
- GSD verify prompt after implementation; gap-plan/re-execute only if verification finds gaps.
- GSD code-review prompt and reasoned disposition before the final commit.

## Commit checkpoints

1. Plan/context/TDD checkpoint.
2. Executed RED test checkpoint(s), retaining failing evidence in the ledger.
3. Green generator/check and canonical source.
4. Shared contract reframe and review/verification fixes.

## Captain topology disposition

The initial Wave 1 branch started from `origin/main` before the required parent topology existed.
Before any child push, PR, or CI phase, the outer delivery executor must create
`refactor/3714-canonical-delivery-flow` from current `origin/main`, add a deliberate seed commit,
and open its draft PR to `main` as the Waves 1–7 integration point. It must then restack this branch
as the first child without squashing, resetting, rebuilding, or dropping any implementation or
gate-fix commit. The active no-mistakes review step records this captain decision but does not own
branch-topology, push, PR, or CI mutations.
