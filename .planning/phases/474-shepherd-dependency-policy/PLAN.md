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
- [x] Focused three-file test command passes.
- [x] Full `.pi/extensions/shepherd/*.test.ts` suite passes with exact test count.
- [x] Strict `tsc --noEmit` passes against pinned Pi 0.80.6 types.
- [x] Pi 0.80.6 offline RPC extension discovery passes. The requested
  `pi --list-extensions` spelling is unsupported by that version and is recorded as a tooling
  deviation, not silently skipped.
- [x] `git diff --check` passes.
- [x] Parent policy declared the phase-equivalent child verification gate to be the focused tests,
  full Shepherd suite, strict TypeScript, offline Pi RPC smoke, and diff check. The attempted full
  `make verify` was intentionally cancelled by the parent under that superseding policy, not failed.
- [x] Supplemental `go vet ./...`, `go test ./...`, and `go build ./cmd/pm` pass.
- [x] Diff remains inside owned files.
- [x] Ready stacked PR #483 targets `feat/471-pi-agent-session-shepherd`, includes `Refs #474` and
  `Refs #471`, and does not request Claude or Copilot.

## Commit/push checkpoints

1. plan artifacts;
2. RED tests with captured expected failure;
3. GREEN implementation and focused suite;
4. refactor/full verification/final artifacts;
5. push and ready stacked PR.

## Verification policy update

During final verification, the parent orchestrator relayed a new explicit user policy: full-repo
`make verify` is no longer a required child-lane gate. It intentionally terminated that run and
declared the phase-equivalent child gate listed above. Artifacts therefore record
`cancelled_by_parent_policy`, not a functional failure. `verificationPassed` refers to the declared
phase-equivalent child gate, consistent with the universal loop's “full make verify (or declared
phase equivalent)” rule.

## Execution decisions

This child issue is already isolated in its assigned worktree. Its three modules form one cohesive
pure policy boundary, so each child lifecycle cycle runs as `local_critical_path`; no nested worker
is necessary or authorized.
