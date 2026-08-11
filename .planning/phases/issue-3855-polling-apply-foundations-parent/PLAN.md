# PLAN — #3855 polling/apply foundations parent scaffold

## Objective

Create a deliberately small, draft-only parent delivery surface for #3855. It records the existing
child dependency order and reusable polling provenance while preserving the transport dependency
and final certification attachment rules.

## Scope

Owned paths are only:

- `.planning/phases/issue-3855-polling-apply-foundations-parent/**`

Out of scope:

- production or test implementation for #3856, #3857, #3858, #3859, #3864, or PostgreSQL, and
  documentation delivery for #3860;
- `cmd/`, `internal/`, connector definitions, generated files, credentials, provider calls, and
  runtime services;
- certification, executable product behavior, child integration, or any merge;
- targeting `main` or treating #3880 as completion of a child.

## GSD and skill record

The following adapter checks succeeded before this seed:

```text
scripts/gsd doctor
scripts/gsd sources discuss-phase
scripts/gsd sources plan-phase
scripts/gsd sources execute-phase
scripts/gsd sources verify-work
scripts/gsd sources code-review
go run ./cmd/agentcontractgen check
```

Generated prompts for the five lifecycle commands were resolved with phase argument `3855` and
executed inline. The single-worker contract forbids the required role spawning, so this is the
documented manual-GSD fallback. It does not omit planning, verification, or review evidence.

Loaded skills:

- `github-issue-first-delivery` for issue/PR topology and issue-first linkage;
- `no-mistakes` for the required local gate;
- `golang-documentation` for bounded phase-evidence accuracy; it does not authorize edits to
  inherited #4015 architecture content.
- `golang-how-to` and `golang-lint` for the scoped documentation/lint review; no Go source path is
  in scope for static analysis.

No Go implementation, CLI, connector-runtime, website, or design skill applies: this seed changes
only planning artifacts and does not describe or alter product behavior.

## Child acceptance ledger

| Order | Existing issue | Independently deliverable outcome | Preconditions | Parent acceptance rule |
| --- | --- | --- | --- | --- |
| 1 | #3856 | Immutable, executable polling-watermark conformance corpus with an all-fixtures/no-skip runner | Shared #3810 contract reuse | Must land before #3857 begins. |
| 2 | #3857 | Declarative polling descriptor and real transport preflight | #3856; reviewed #3864 generic seam | Must consume, not duplicate, #3864 transport registration/preflight; must not advertise polling as public CDC. |
| 3a | #3858 | Page-safe polling source executor through the #3864 source port | #3857 | May run in parallel with #3859 only after #3857 lands. Reuse #3880 mechanics without claiming #3880 completed it. |
| 3b | #3859 | Native apply strategies through the #3864 destination port | #3857 | May run in parallel with #3858 only after #3857 lands. Shared mode semantics stay out of PostgreSQL-specific work. |
| 4 | #3860 | Truthful polling-watermark eligibility and limitation documentation | #3856–#3859 | Follows completion of all four core children; it is not a parallel lane in the core implementation DAG. |

## Parent topology ledger

| Item | Required state |
| --- | --- |
| Primary issue | #3855; retain existing children #3856–#3860 and create none. |
| Parent branch | `feat/3855-polling-apply-foundations`, created from current `origin/feat/3862-any-to-any-transport` head `30b2fb4aeb121641b6158903fe1d3b54668599a6`. |
| Initial PR | One draft PR only; base `feat/3862-any-to-any-transport`, head `feat/3855-polling-apply-foundations`. |
| Temporary-base meaning | Dependency-only; it never authorizes #3855 integration into #3862. |
| Retarget rule | Before final parent integration, retarget to `docs/4015-connector-release-certification` after the reviewed transport seam is present there. |
| Merge rule | Never merge into the temporary #3862 base or `main`; after retargeting, final integration remains human-gated. |
| Historical reuse | #3880 / `dc1a6a7171ca901c8dbaf8cd528f67b18e57d9bb` is partial reusable polling implementation, not closure for any child. |

## Plan checkpoints

1. **RED — absent parent surface:** establish the pre-seed absence of the requested local branch
   and durable current parent acceptance ledger; preserve the live absence of a draft PR rather
   than replacing it with an assumption.
2. **GREEN — bounded seed:** create this phase ledger on the named branch from the verified remote
   head, with exact child order, scope fence, and reuse classification.
3. **GREEN — local topology validation:** verify tracked paths, content invariants, formatting/check
   hygiene, and canonical GSD projection health without touching product code.
4. **GREEN — local no-mistakes gate:** run the exact child-style argv with
   `--skip=push,pr,ci`, never `--yes`; own every synchronous decision gate.
5. **GREEN — draft parent PR:** after local gates, push only this branch and use `gh-axi` to create
   exactly one draft PR with the explicit temporary base/head. Re-read live state with `gh-axi` and
   record its exact draft/base/head result.

## Refactor policy

There is no code refactor in this phase. If a validation exposes a product defect, search for an
existing child first and otherwise create a narrowly scoped #3855 child before changing code. A
pipeline or topology recovery that changes no product code does not consume a correction round.

## Verification plan

- `git diff --check` and a changed-path assertion must show planning artifacts only.
- `go run ./cmd/agentcontractgen check` must remain green.
- Run applicable repository gates (`make tidy-check`, `make docs-check`, and `make lint`) without
  lowering their standards.
- Run the no-mistakes local pipeline using the contract-owned skip vector.
- Validate the final GitHub object only with `gh-axi`: draft state, exact base/head, one PR, and no
  certification or executable-behavior claim.
