# Plan: #478 GitHub Parent Orchestration

Issue: https://github.com/polymetrics-ai/cli/issues/478
Parent: #471
Parent PR: #472
Branch: `feat/478-shepherd-github-parent-orchestration`
Exact base: `3addb1f48be1afe8b1e2b59b54247679d7293805`
PR base: `feat/471-pi-agent-session-shepherd`

## Objective

Implement a typed, fakeable GitHub orchestration boundary that turns a parent objective into
bounded dependency-linked child records, reconciles parent/child issue and stacked-PR publication
before mutation, evaluates authoritative CI/review/disposition evidence at an exact head, and
keeps parent ready/merge transitions behind an exact-generation/head consumed human decision.

## Workflow and required skills

- GSD mode: `manual_gsd_fallback`. `scripts/gsd doctor` passes, while
  `scripts/gsd prompt programming-loop init --phase 478-shepherd-github-parent-orchestration
  --dry-run` reports `unknown GSD command: programming-loop`.
- Skills read completely: `gsd-programming-loop`, `github-issue-first-delivery`,
  `gsd-workstreams`, `architecture-patterns`, and `javascript-testing-patterns`, plus the required
  repository routing, issue/parent contracts, orchestration loops, and Pi adapter/runtime policy.
- Execution decision: `local_critical_path`. A read-only explorer spawn was attempted twice and
  rejected because the runtime is at its agent-thread limit. The three owned modules share one
  ordered RED-before-GREEN contract, so implementation continues in this isolated issue worktree.
- The live #478 issue and coordinator handoff supersede the older local issue draft where review
  policy differs: eligible automated review evidence is only `codex_independent` using
  `openai-codex/gpt-5.6-sol` at `xhigh`; Claude, Copilot, generic Codex, and human approvals cannot
  satisfy the automated review gate.

## Owned slice

1. `github-evidence.ts`: validate bounded authoritative issue/PR/check/review-thread/disposition
   snapshots, exact reviewed ranges, workspace handoff evidence, and eligible independent Codex
   review records. Reject unknown fields, hostile values, stale heads, incomplete checks,
   requested changes, unresolved threads, and undispositioned findings.
2. `review-router.ts`: pure declarative routing and exact-head invalidation. It returns review work
   records only; it does not spawn `AgentSession` or edit controller/extension wiring reserved for
   #479.
3. `github-orchestrator.ts`: validate bounded parent/child plans by reusing
   `DependencyWorkItem`/`validateDependencyGraph`; derive child branches, parent PR bases, skills,
   verification, gates, and stable markers; reconcile typed transport snapshots before every
   publication/integration mutation; reuse `selectReadyWork`, `reconcileAutonomy`,
   `ClaimedWorkspace`/`WorkspaceHandoffEvidence`, and `GitHubDecisionBroker` request/poll/consume
   contracts instead of duplicating their domains.
4. Tests and issue-local fixtures use fake transports only. No live GitHub mutation, generic HTTP
   write, shell escape hatch, credential value, default-branch mutation, merge, or ready-for-review
   action is permitted.

## Required behavior and RED matrix

| Requirement | Test-first failure |
| --- | --- |
| bounded canonical child records | reject extra fields, oversized rosters/lists/text, invalid dependency/scope graph, unsafe branch/base, duplicate marker, or missing skills/verification/gates |
| retry-safe publish/restart | reconcile first; recover timeout-after-publish without duplicate mutation; reject same marker on mismatched resource |
| stacked topology/linkage | child PR base equals parent branch; child body uses `Refs #child` and `Refs #parent`; no child `Closes`; parent PR alone may use `Closes #parent` |
| authoritative quality evidence | require successful completed checks, no requested changes, no unresolved threads, and a disposition for every actionable finding |
| exact-head review coverage | accept only clean `codex_independent` review by `openai-codex/gpt-5.6-sol:xhigh` covering the exact base..head; invalidate on any head or range movement |
| scoped integration | require passed `WorkspaceHandoffEvidence`, clean workspace, allowed changed scope, open stacked PR, and current exact head before provisional integration |
| parent hard gate | remain draft/blocked until every child is integrated and a broker decision is consumed for the exact parent PR, generation, and head; never perform merge |

## Typed transport boundary

The orchestration transport is separate from `GitHubDecisionTransport`. It exposes only bounded,
domain-specific reads and mutations for issue/PR/status records. Every mutation is preceded by a
fresh read keyed by a stable marker. Existing matching resources are reused; duplicate markers or
marker/resource collisions fail closed. Mutations return the canonical record that the next
reconciliation validates.

The human-decision dependency is the existing `GitHubDecisionBroker`-compatible
`request`/`poll`/`consume` surface. The orchestrator never widens its comment-only transport and
never treats review text, CI, reactions, bots, or provider-generated records as human approval.

## Threat model and hard stops

- Treat all GitHub payloads, titles, bodies, branches, scopes, review text, URLs, and transport
  results as untrusted; validate exact shapes and bounded sizes before use.
- Stable markers are derived from canonical immutable intent. A reused marker with different
  content is a collision, not an update opportunity.
- Timeout/retry ambiguity is resolved by reading canonical state again, never by blind retry.
- Exact head/base/range binding prevents stale review or human approval from authorizing new code.
- Failed/pending/skipped/missing checks, requested changes, unresolved threads, missing
  dispositions, dirty/out-of-scope handoff, or head movement fail closed.
- Stop before auth changes, secrets, dependencies, live external mutation, policy exceptions,
  parent ready transition, parent merge, or any merge to `main`.

## TDD lifecycle

1. Commit this plan checkpoint while production and tests remain absent.
2. Add only three matching test files and bounded JSON fixtures. Run the focused command and record
   deterministic module-absence RED evidence.
3. Commit/push the test-only RED checkpoint.
4. Implement the smallest typed evidence and review-policy primitives, then the reconcile-first
   orchestration service. Run focused tests until GREEN.
5. Refactor validation, canonicalization, and fake-transport helpers without adding controller
   wiring or new capabilities.
6. Run only the coordinator-authorized gates:

```bash
node --test .pi/extensions/shepherd/github-orchestrator.test.ts \
  .pi/extensions/shepherd/review-router.test.ts \
  .pi/extensions/shepherd/github-evidence.test.ts
node --test --test-concurrency=1 .pi/extensions/shepherd/*.test.ts
# strict no-emit TypeScript against pinned Pi 0.80.6 declarations
# offline Pi RPC get_commands discovery
git diff --check 3addb1f48be1afe8b1e2b59b54247679d7293805..HEAD
```

7. Verify exact base ancestry and restrict changed paths to the assigned six files, issue fixtures,
   and this phase directory. Do not run Go, connector, certification, runtime, or `make` gates.
8. Commit/push coherent GREEN/refactor and evidence checkpoints. Open a ready stacked PR titled
   `feat(shepherd): orchestrate parent issues and stacked PRs`, based on
   `feat/471-pi-agent-session-shepherd`, with `Refs #478` and `Refs #471` and no `Closes` keyword.
   Return the handoff without merging or requesting Claude/Copilot review.

## Completion checklist

- [x] Exact base/branch/parent topology confirmed.
- [x] Plan, TDD ledger, and verification checklist created before production edits.
- [x] Test-only RED matrix committed and pushed.
- [x] Minimal GREEN implemented.
- [x] Refactor and all authorized gates pass.
- [x] Ready stacked PR #487 opened with correct title/base/body.
- [ ] Final exact-head independent review and human parent merge remain parent-owned.

## Adversarial correction slice: 2026-07-21

Local review of GREEN head `90321ffb` found restart and binding gaps that need another strict TDD
slice before broad verification:

1. Bind child integration receipts to the plan generation, stable PR marker, reviewed base SHA,
   exact head SHA, child ID, PR number, and parent branch.
2. Reuse an exact receipt after GitHub transitions the integrated child PR from open to merged;
   do not falsely regress a successful integration on restart.
3. Require the clean review record used for integration/readiness to match the planned repository,
   work item, generation, changed paths, and allowed scopes in addition to base/head.
4. Capture parent branch setup through the same upstream `captureHandoff` boundary used by children.
5. Reject proxied arrays before invoking traps, reject duplicate cross-review finding IDs, and do
   not let a local `not_actionable` label satisfy a blocking finding without a brokered exception.
6. Remove the fake-only PR-number hint from the production transport request during refactor.

The test-only correction run passes 17 and fails 10 expected cases against unchanged GREEN
production. A review-agent spawn was retried after GREEN and remains unavailable because the
runtime is at its thread cap, so this is recorded as another `local_critical_path` correction.

The matching correction is pushed at `40ce66d4b5010b92089895a05709687143d15a05`. Focused #478
tests pass 27/27; the serialized Shepherd suite passes 290 with one intentional sandbox skip;
strict owned and all-production TypeScript pass with TypeScript 5.9.3 against the cached Pi 0.80.6
package/type surface; pinned offline Pi RPC discovers `pm-shepherd`; and immutable-base, diff, and
owned-path checks pass. Review dispatch remains parent-owned under the later stable-head policy.

## Stable-head functional review correction slice: 2026-07-21

The deep functional review froze base `3addb1f48be1afe8b1e2b59b54247679d7293805` and reviewed
head `093b3c90409cedc6b7008b7510f53937eb1ebbc1`. Its eleven accepted findings are one strict
RED-before-GREEN correction slice. Manual GSD remains the recorded fallback because the repository
adapter does not expose `programming-loop`; this isolated worker takes the `local_critical_path`
without editing #475, #479, controller, extension, or top-level integration files.

### Correction contract

1. Carry one authoritative expected changed-path set through pull-request evidence, review target,
   session attestation, integration, and parent readiness. Require exact set equality; empty,
   subset, and superset claims fail while reordered equality passes.
2. Reconcile every child integration through complete authoritative transport lookup, canonical
   materialized child topology, the current parent PR head, and a typed ancestry proof. Reject
   forged, stale, mismatched, or force-push-orphaned receipts.
3. Validate every immutable `MaterializedChildRecord` field against plan plus canonical issue before
   capture, publication, and integration. No unchecked DTO field reaches transport.
4. Bind CI to a required context set and trusted producer identifiers. Require one deterministic
   successful terminal rollup per required context; absent, pending, duplicate-ambiguous,
   unsuccessful, untrusted, or incomplete results fail closed.
5. Separate reviewer output from controller-owned execution evidence through an issue-local scoped
   `AgentSession` attestation interface/fixture. The attestation binds session/run identity, launch
   provider/model/effort/read-only policy, repository, PR, generation, base/head, exact paths, and
   result digest. Self-asserted reviewer execution metadata cannot satisfy the gate.
6. Pass the expected independent-review target directly into evidence evaluation and select equal
   generations deterministically, independent of transport and pagination ordering.
7. Serialize issue and PR ensure operations by stable marker within the process, always reconcile
   authoritatively after create, and fail closed when post-create state is absent or ambiguous.
8. Enforce positive generation at every correction boundary.
9. Centralize a pure canonical Git ref validator with check-ref-format-equivalent restrictions; no
   subprocess is introduced.
10. Make every marker/integration lookup return bounded `{ items, complete }` evidence. An
    incomplete lookup is never treated as absence or uniqueness.
11. Add deterministic partial-failure regressions for PR creation, integration, ready transition,
    malformed mutation recovery, concurrent ensures, and incomplete pagination while preserving
    broker semantics and secret safety.

### Strict checkpoint sequence

1. Commit and push this plan/TDD/verification/review-artifact update with production and tests
   byte-identical to reviewed head `093b3c90`.
2. Add one behavior-level test-only RED commit covering all eleven findings. Run the focused suite
   and prove every production file under `.pi/extensions/shepherd/*.ts` is byte-identical to
   `093b3c90` before the RED commit.
3. Implement a coherent architectural GREEN only in the three #478 production modules, matching
   tests/fixtures, and this phase directory.
4. Run focused #478 tests, then the full serialized Shepherd test suite, strict TypeScript over
   owned and all Shepherd production modules against pinned Pi 0.80.6, offline Pi RPC registration,
   immutable-base/head/diff/scope checks, and `git diff --check`. Do not run Go, connector,
   certification, runtime-service, or `make` gates.
5. Commit/push GREEN and final evidence checkpoints and update PR #487. Fresh stable-head xhigh
   review remains parent-owned.

### Correction checkpoint outcome

- [x] Artifact-only plan checkpoint committed: `5dd7897e`.
- [x] One test-only RED committed: `4e02d059`; production byte-identical to `093b3c90`.
- [x] Coherent architectural GREEN committed: `8e32896a`; focused 38/38 and strict TypeScript pass.
- [x] Offline Pi RPC and immutable base/head/diff/scope checks pass.
- [ ] Full serialized suite is green: blocked by 65 unrelated `spawn EPERM` sandbox failures
      (236 pass, 65 fail, 1 skip; all #478 tests pass).
- [ ] Push/update PR #487: blocked by GitHub DNS resolution; all commits retained locally.
