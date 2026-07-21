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
- [ ] Minimal GREEN implemented.
- [ ] Refactor and all authorized gates pass.
- [ ] Ready stacked PR opened with correct title/base/body.
- [ ] Final exact-head independent review and human parent merge remain parent-owned.
