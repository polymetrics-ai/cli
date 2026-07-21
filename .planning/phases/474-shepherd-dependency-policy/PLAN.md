# Issue #474 Dependency Policy Plan

## Contract

- Objective: deliver a pure, deterministic dependency policy and reconciler for the in-process
  Shepherd parent lifecycle.
- Parent: #471 / PR #472.
- Immutable base: `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`.
- Branch: `feat/474-shepherd-dependency-policy`.
- Allowed production scope: `autonomy-policy.ts`, `dependency-graph.ts`, `reconciler.ts` and their
  matching tests only.
- External effects: none. Git, GitHub, filesystem, clocks, credentials, and agent sessions are
  canonical input data or integration-layer concerns.

## GSD mode and skills

- GSD command attempted:
  `scripts/gsd prompt programming-loop init --phase 474-shepherd-dependency-policy --dry-run`.
- Result: `scripts/gsd: unknown GSD command: programming-loop`.
- Mode: `manual_gsd_fallback`; execute the full plan -> RED -> GREEN -> refactor -> verify ->
  review/handoff lifecycle locally.
- Required skills loaded completely: `gsd-programming-loop`, `gsd-workstreams`, `gsd-plan-phase`,
  `github-issue-first-delivery`, `architecture-patterns`, `javascript-testing-patterns`.
- Required routing reviewed: issue-agent contract, universal GSD loop, Pi adapter, runtime/Pi
  integration, worker handoff, autonomous Shepherd lifecycle, and validator boundary.
- Go skills: not applicable to the TypeScript-only production slice; repository-wide Go gates still
  run unchanged.

## Architecture boundaries

1. `autonomy-policy.ts` owns lifecycle stages/transitions, repository blocker vocabulary, and
   bounded retry/correction decisions.
2. `dependency-graph.ts` owns closed-world DAG validation, canonical write scopes, collision
   detection, and maximum-cardinality deterministic ready selection under concurrency.
3. `reconciler.ts` consumes persisted intent plus canonical observed facts and returns one pure
   decision. It performs no I/O and never invents truth.
4. Existing `domain.ts`, controller, state store, runner, SDK runner, evidence capture, extension,
   and shared parent artifacts are typed integration boundaries and remain untouched.

## TDD slices

### Slice 1: lifecycle and retry policy

- RED: table tests for every legal lifecycle edge, unsafe/skipped edges, research skip conditions,
  exact-head human wait, transient verification/review retry budgets, and hard human gates.
- GREEN: minimal immutable transition and budget functions.
- Refactor: centralize frozen transition vocabulary and validate all untrusted numeric inputs.

### Slice 2: dependency graph and ready queue

- RED: cycles, self/unknown dependencies, duplicate IDs, ambiguous/broad/path-traversing scopes,
  ancestor collisions, read-only coexistence, occupied running scopes, deterministic ordering, and
  a star-conflict case requiring maximum safe selection rather than first-fit selection.
- GREEN: closed-world graph validator plus bounded branch-and-bound independent-set selection.
- Refactor: canonical comparison keys and immutable copies.

### Slice 3: reconciler

- RED: fail-closed invalid snapshot, persisted/canonical drift, safe transition, correction retry,
  hard human wait, ready spawn, every no-spawn blocker category, complete state, and repeated-call
  deep equality/idempotence.
- GREEN: deterministic priority-ordered decision pipeline.
- Refactor: exactly-one blocker invariant for every `no_spawn` result.

## Verification checklist

- [ ] Genuine focused RED test failures captured before production files exist.
- [ ] Focused three-file test command passes.
- [ ] Full `.pi/extensions/shepherd/*.test.ts` suite passes with exact test count.
- [ ] Strict `tsc --noEmit` passes against pinned Pi 0.80.6 types.
- [ ] `pi --list-extensions` passes.
- [ ] `git diff --check` passes.
- [ ] `go vet ./...` passes.
- [ ] `go test ./...` passes.
- [ ] `go build ./cmd/pm` passes.
- [ ] `make verify` passes; only then may `verificationPassed` be true.
- [ ] Diff remains inside owned files.
- [ ] Ready stacked PR targets `feat/471-pi-agent-session-shepherd`, includes `Refs #474` and
  `Refs #471`, and does not request Claude or Copilot.

## Commit/push checkpoints

1. plan artifacts;
2. RED tests with captured expected failure;
3. GREEN implementation and focused suite;
4. refactor/full verification/final artifacts;
5. push and ready stacked PR.

## Execution decisions

This child issue is already isolated in its assigned worktree. Its three modules form one cohesive
pure policy boundary, so each child lifecycle cycle runs as `local_critical_path`; no nested worker
is necessary or authorized.
