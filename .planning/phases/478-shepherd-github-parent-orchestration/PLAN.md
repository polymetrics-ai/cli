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

## Cycle 3 deep-review correction: 2026-07-22

Two independent corrected review ledgers examined frozen candidate
`3f285722a505ea426d53a34f95716781d1aca7c2` against immutable base
`3addb1f48be1afe8b1e2b59b54247679d7293805`. Their overlapping findings are accepted as one
fourteen-invariant batch. The implementation/correction route remains local Codex
`openai-codex/gpt-5.6-sol:high`; `xhigh` is reserved for parent-owned planning, review, and
orchestration. No Claude/Copilot finding or controller/#479 work is imported.

### Cycle 3 correction contract

1. Give every plan a persistence-safe opaque canonical serialization and digest. Revalidate and
   rebuild it at every public mutation/readiness boundary, rejecting clones, deserialized or
   tampered envelopes, proxies, accessors, cycles, unknown fields, oversized input, and altered
   repository/generation/branch/marker/topology values.
2. Reject top-level `read_only` children. Every planned child mutates through an issue/PR/integration
   lifecycle and therefore requires non-empty scopes; read-only roles remain #479-internal.
3. Bind outer PR evidence and the independent-review target to the same repository, work item,
   generation, PR number, base/head SHA, base/head branch, exact paths, and scopes.
4. Bind an integration receipt to the complete canonical child PR snapshot and its controller and
   durable-transport mutation provenance before parent readiness can consume it.
5. Require independently complete changed-path evidence plus complete checks, requested changes,
   threads, reviews, dispositions, and a controller minimum observation revision/time. No expected
   path set may be copied tautologically from the evidence snapshot under evaluation.
6. Enforce causal event chronology and freshness: no future/backward event, no fake pending
   completion, sequence-authoritative check rollups, disposition after finding, and a clean review
   after all blocking findings and satisfying dispositions. Stale same-head evidence blocks.
7. Introduce durable, cross-instance conditional mutation/idempotency contracts for issue, PR,
   integration, parent-ready, and roster publication. Keep local keyed queues as bounded FIFO
   optimization only; serialize integration/ready, reconcile after success and retry visibility,
   and prove two orchestrators produce one external effect.
8. Replace ancestry booleans with an exact runtime-validated proof binding repository, ancestor,
   descendant, result, revision, and observation time. Only literal `true` at exact coordinates
   satisfies readiness.
9. Reconcile same-marker review attempts by marker, result digest, and target; reject true
   ambiguity and remain permutation-invariant.
10. Reject `HEAD` and all pseudo/symbolic refs through the single canonical Git-ref validator.
11. Inject a repository/base-branch-scoped, versioned required-CI policy whose contexts,
   producers, revision, and digest are bound into the canonical plan/generation. Policy movement
   or stale digest blocks; no hard-coded single-check fallback remains.
12. Give roster snapshots monotonic revision/status epochs and conditional durable publication so
   an older orchestrator cannot overwrite newer status.
13. Export the controller-side attestation result digest, constructor, and validator so #479 and
   tests consume one canonical protocol rather than copying a private hash implementation.
14. Extend partial-effect and bound tests across proxy/accessor/cycle/oversize/secret inputs,
   retry-before-visibility, queue rejection/cleanup, and concurrent orchestrator instances.

### Cycle 3 strict lifecycle

1. Commit this artifact-only plan checkpoint while every #478 production/test blob remains equal
   to frozen candidate `3f285722`.
2. Make exactly one test/fixture-only RED commit covering all fourteen invariants. Before committing,
   prove every #478 production blob is identical to `3f285722`; record the focused failing matrix.
3. Implement one architectural GREEN/refactor within the three owned production modules plus their
   tests and issue-478 fixtures. No #479 controller or parent planning file may change.
4. Run focused #478, strict owned and all-production TypeScript, serialized Shepherd, pinned offline
   Pi RPC, immutable-base/diff/scope checks, and secret scans. Record sandbox `spawn EPERM` as an
   environmental broad-gate failure if it recurs; do not run Go, connectors, certification,
   runtime services, `make`, or live GitHub.
5. Commit verification evidence. Push/PR synchronization remains deferred while the recorded DNS
   failure persists; fresh exact-head independent review and human integration remain parent-owned.

### Cycle 3 checkpoints

- [x] Artifact-only PLAN commit: `d97faf44` (model-policy correction: `d2c7f374`).
- [x] Single all-invariants test-only RED commit with production blob identity proof: `faf2e8f8`.
- [x] Architectural GREEN/refactor commit: `41e8e76e`.
- [x] Focused and authorized broad verification evidence commit (this artifact checkpoint).
- [ ] Push/PR #487 synchronization when network access returns.

## Cycle 4 consolidated stable-head correction: 2026-07-22

Two final deep-review ledgers examined frozen candidate
`d3b6b5e226b17db6ec8350163acdbb41368ec3bf` against immutable base
`3addb1f48be1afe8b1e2b59b54247679d7293805`. All findings are accepted as one ten-contract
correction batch. Implementation uses `openai-codex/gpt-5.6-sol:high`; fresh review remains
parent-owned `xhigh`.

Required routing loaded: `gsd-programming-loop`, required-skills routing, the Pi adapter
reference, and the Pi/runtime integration reference. `scripts/gsd doctor` passes, but
`scripts/gsd prompt programming-loop ...` is unavailable, so the recorded route is
`manual_gsd_fallback`. The collaboration runtime is at its thread cap, so execution is the
`local_critical_path`.

### Cycle 4 correction contract

1. Split stable integrated-PR identity from volatile observation evidence. Stable identity binds
   repository, work item, PR, generation, marker, both branches, reviewed SHAs, exact paths/scopes,
   and required-policy identity; observation revision/time remains separate. Later identical
   topology and merged-state observations reuse while wrong head branch fails.
2. At restart and readiness, reconstruct the authoritative materialized child from its canonical
   issue and validate current PR plus receipt through the same topology used at integration,
   including exact branches/marker/generation/SHAs, path-within-scope, stable identity, controller
   provenance, and durable transport provenance. Recomputed digests never authorize wrong topology.
3. Put a bounded deadline and `AbortSignal` on every transport, evidence, workspace, policy, and
   broker call. Normalize timeout into an uncertain typed failure, release keyed queues in
   `finally`, and reconcile late effects by durable mutation key without unmanaged bare races.
4. Apply one sensitive-text grammar before canonical persistence or outbound human text, covering
   titles, objectives, verification, decision questions, and issue/PR/roster bodies. Normalize
   every external rejection shape to bounded redacted typed errors/codes.
5. Add a controller-owned `RequiredCheckPolicySource` that returns exactly one complete current
   repository/base observation with revision, digest, and observation time. Re-read immediately
   before child integration and parent readiness; incomplete, stale, moved, or wrong-coordinate
   policy blocks the existing generation.
6. Centralize one canonical Git branch/ref validator and reject the complete pseudo/symbolic set:
   `HEAD`, `FETCH_HEAD`, `ORIG_HEAD`, `MERGE_HEAD`, `CHERRY_PICK_HEAD`,
   `REVERT_HEAD`, `REBASE_HEAD`, `BISECT_HEAD`, `AUTO_MERGE`, `refs/*`, and segment forms.
7. Validate post-mutation CAS semantics. Durable result intent/revision and authoritative resource
   revision must strictly advance beyond the expected revision for roster and parent-ready
   mutations; stale/regressing out-of-order writers fail.
8. Pre-bound every array/envelope from its own `length` data descriptor before descriptor/key
   materialization, require dense indices, and compare exact array lengths in canonical equality.
   Large dense and million-length sparse nested values reject before traversal or effects.
9. Use collision-free tuple keys for session/run and every compound identity. Colon-bearing
   distinct pairs remain distinct while exact duplicate tuples reject.
10. Retain every Cycle 3 invariant, extend partial-effect/proxy/accessor/cycle/bounds/error
    regressions, and preserve the typed #479-facing ports without controller implementation.

### Cycle 4 strict lifecycle

1. Commit this artifact-only plan while all production and test blobs equal frozen `d3b6b5e2`.
2. Make exactly one behavior-level test/fixture-only RED commit covering all ten contracts. Before
   committing, prove `github-orchestrator.ts`, `github-evidence.ts`, and `review-router.ts`
   retain blob IDs `ed576e64`, `a3076e39`, and `ca0c8116`.
3. Implement one architectural GREEN/refactor only in #478-owned modules/tests/fixtures.
4. Run focused #478, serialized Shepherd, strict owned and all-production TypeScript against
   pinned Pi 0.80.6, offline RPC, immutable-base/diff/scope/data scans. Record unrelated sandbox
   `spawn EPERM` separately. Do not run Go, connectors, certification, runtime, `make`, network,
   live GitHub, #479 controller, reviewer, or merge actions.
5. Commit evidence and report exact PLAN/RED/GREEN/evidence SHAs plus the clean candidate for two
   fresh parent-owned reviews. Preserve the existing DNS-deferred local-only state.

### Cycle 4 checkpoints

- [x] Artifact-only PLAN commit: `607e203ef1f76ff112c130ccff5d155973d984f6`.
- [x] One all-contract test/fixture-only RED with frozen production blob proof:
      `abbf388b8a852836e0dd10a55b9f17720b9fde22`.
- [x] One architectural GREEN/refactor commit:
      `b92b5ff7dd3738dc3b3350ebb4d2f2b42074f954`.
- [x] Authorized verification/evidence commit and clean candidate; the evidence commit hash is
      reported in the worker handoff because a commit cannot contain its own identity.
- [ ] Two fresh exact-head independent reviews remain parent-owned.

Cycle 4 local verification: focused 68/68; strict owned and all-production TypeScript pass against
pinned Pi 0.80.6; pinned offline RPC discovers `pm-shepherd`; immutable-base, ancestry, diff,
17-path ownership, and credential-literal scans pass. The serialized suite ran 332 tests: 266 pass,
65 unrelated managed-sandbox `spawn EPERM` failures, and one intentional skip; every #478 test
passes. No prohibited, network, reviewer, controller/#479, or merge action ran.

## Cycle 5 consolidated stable-head correction: 2026-07-22

Two independent Cycle 4 ledgers reviewed exact frozen candidate
`ca6f6873d168db707bbe58291b5ee1b582e9404f` against immutable base
`3addb1f48be1afe8b1e2b59b54247679d7293805`. Their unique union plus the stale run-state warning
is accepted as one architecture correction. Cycle 5 retains all 68 focused Cycle 4 contracts.

Required routing loaded: `gsd-programming-loop`, `caveman`, required-skills routing, the Pi adapter
reference, the runtime/Pi integration reference, and required GSD project artifacts. `scripts/gsd
doctor` passes. The adapter does not expose `programming-loop` and `scripts/programming-loop.mjs`
is absent, so the recorded route remains `manual_gsd_fallback`. One read-only explorer maps coupled
orchestrator symbols while this isolated worker owns the ordered production critical path.

### Cycle 5 correction contract

1. Runtime-validate each broker request, poll, and consume result as an exact canonical human
   decision record. Exact-match request ID, `parent_merge` gate, idempotency marker, options,
   actor allowlist, target/binding/generation/head, question, expiry/lifetime, timestamps,
   status/decision/consumption coherence, allowlisted actor, and exact GitHub source before any
   decision branch or ready effect. TypeScript interface claims are never runtime evidence.
2. Refresh the complete set of plan-bound policy coordinates before fresh or reused child
   authorization, initial readiness, post-decision revalidation, and ready-effect recovery.
   Missing, incomplete, stale, ambiguous, moved, or wrong-coordinate observations block. Extend
   the typed policy source with one authoritative full initial bundle/config read and expose an
   async plan-construction boundary so #479 can wire only ports and never invent policy authority.
3. Replace self-authenticating receipt provenance with controller authorization bound to the
   stable PR identity, exact changed-path observation digest, attested Codex xhigh review-result
   digest, current policy observation, canonical plan, and integration intent/result. Until a
   durable #479 authorization store exists, reuse and readiness independently re-evaluate current
   controller evidence and require exact authorization equality; loose revision/time inequalities
   and transport-only recomputation fail.
4. Centralize current child PR eligibility across integration reuse and readiness. Reject draft,
   closed, and malformed draft+merged evidence consistently; document and allow only unchanged
   open, open-to-merged, and unchanged merged post-integration transitions.
5. Bind `expectedResourceRevision` into both durable mutation key and intent digest. Build child
   integration mutation identity from stable logical authorization/PR identity while keeping
   volatile read revision/time outside it. Same coordinates with a different CAS precondition
   receive a different authenticated identity; observation refresh and timeout-before/after-effect
   recovery remain deterministic.
6. Extend the one shared sensitive-text grammar for `Cookie`, `Set-Cookie`, session identifiers,
   and session/auth/CSRF response-header forms. Table-test every durable/outbound title, objective,
   verification, question, issue/PR/roster body, review finding, and disposition field with
   synthetic values only.
7. Add optional caller lifecycle context to every public async orchestration entry point. Link
   caller abort/deadline with the local deadline; pass explicit abort acknowledgement to ports;
   track every invocation through settlement. An `AsyncLocalStorage` ensure scope retains the keyed
   gate while a timed-out invocation is live. A bounded `stop` aborts and joins active calls, and
   reports incomplete/unacknowledged work rather than claiming a clean join for an uncooperative
   never-settling port.
8. Replace bulk object descriptor expansion with one shared schema-directed bounded reader: inspect
   only declared data descriptors and stop unknown-key enumeration after schema cardinality plus
   one. Add a byte-bounded UTF-8 JSON decoder that rejects oversized raw envelopes before parsing;
   raw or schema-directed representations are the only external boundary forms. Preserve proxy,
   accessor, closed-shape, dense-array, cycle, and exact comparison rejection, with oversized raw
   and oversized object regressions before effects.
9. Rewrite `RUN-STATE.json` atomically to name Cycle 5, exact frozen candidate, exact checkpoints,
   blocked Cycle 4 review state, current local verification truth, and pending parent-owned review.
   No Cycle 3 or superseded Cycle 4 checkpoint may remain the current durable status.
10. Retain all Cycle 4 behavior and the 17-path issue-owned boundary. Do not run Go, connectors,
    certification, runtime services, `make`, network/live GitHub, #479 implementation, reviewer,
    integration, or merge actions.

### Cycle 5 strict lifecycle

1. Commit this artifact-only plan/finding-to-RED matrix while all three production modules and all
   focused tests remain byte-identical to `ca6f6873`.
2. Make one comprehensive test/fixture-only RED commit. Run the retained 68 tests separately and
   require 68/68 green; run every new matrix row and require its intended failure. Prove the three
   production blob IDs still exactly equal frozen `ca6f6873` before GREEN.
3. Implement one coherent GREEN/refactor only in the three #478 production modules plus matching
   tests/fixtures. Report the first focused GREEN immediately.
4. Run focused three-file tests, strict owned/all-production TypeScript against pinned Pi 0.80.6,
   pinned offline RPC, serialized Shepherd classification, immutable-base/ancestry/diff/17-path
   scope, `git diff --check`, and synthetic credential scans.
5. Commit evidence with exact PLAN/RED/GREEN/evidence SHAs and a clean candidate for two fresh
   parent-owned exact-head `xhigh` reviews. No push, network, or GitHub mutation runs.

### Cycle 5 checkpoints

- [x] Artifact-only PLAN/finding-to-RED commit: `7cf9c88ddadee395020444c19ee9f001b0807a53`.
- [x] One comprehensive test/fixture-only RED: `6cb21902244e4bccf390c4e7556eb615e5e1697f`;
      retained 68/68 passed, 37 intended Cycle 5 assertions failed, and production blobs remained
      frozen.
- [x] One coherent architectural GREEN/refactor commit:
      `3ae10dc2303409230153e32e6b6231b27b18cdcf`. Focused 109/109 and strict
      owned/all-production TypeScript pass.
- [x] Authorized local evidence commit and clean candidate (this commit; exact SHA is reported by
      the worker after commit because a commit cannot contain its own identity).
- [ ] Two fresh exact-head independent reviews remain parent-owned.

## Cycle 6 consolidated stable-head correction: 2026-07-22

Two independent Cycle 5 ledgers reviewed exact frozen candidate
`63ac436fdac5fc46be7004f8109c4f068aa5749c` against immutable base
`3addb1f48be1afe8b1e2b59b54247679d7293805`. Their unique blocker union and both run-state/
receipt warnings are accepted as one architecture batch. Cycle 6 retains all 109 focused Cycle 5
contracts.

Required routing loaded: `gsd-programming-loop`, `caveman`, `architecture-patterns`,
`javascript-testing-patterns`, required-skills routing, the issue-agent contract, Pi adapter and
runtime/Pi references, and required GSD project artifacts. `scripts/gsd doctor` passes; the adapter
still does not expose `programming-loop`, so `manual_gsd_fallback` remains explicit. One read-only
explorer maps real broker/shared-boundary symbols while this isolated worker owns the ordered
PLAN/RED/GREEN/evidence critical path (`read_only_spawned` plus `local_critical_path`).

### Cycle 6 bounded scope expansion

The first Cycle 6 plan checkpoint conservatively proposed an 18-path range. The completed read-only
broker contract map proved that range cannot support an honest production composition: the real
broker owns its repository and publicly returns a full record from `request`, a compact poll result
from `poll`, and decision evidence from `consume`. A consumer-supplied repository would permit the
adapter to reread a different store, while an orchestrator-only adapter cannot canonically reread
the broker-owned record. This artifact-only amendment therefore expands the reviewed 17-path range
by exactly four coupled paths:

- `.pi/extensions/shepherd/github-decision-broker.ts`
- `.pi/extensions/shepherd/github-decision-broker.test.ts`
- `.pi/extensions/shepherd/human-decision.ts`
- `.pi/extensions/shepherd/human-decision.test.ts`

`GitHubDecisionBroker.readRecord` must use the broker's own repository, validate the canonical
record, and enforce the exact binding. The explicit adapter remains in `github-orchestrator.ts` and
maps the real request/poll/consume shapes to the controller contract only after broker-owned
canonical rereads and exact result/reload coherence. Its composition test instantiates the actual
broker without type casts or invented fields. `human-decision.ts` imports the authoritative
credential assertion from `review-router.ts`; focused human-decision tests prove that shared grammar
and the descriptor-safe canonical record chronology at their native boundary.

Expected exact base-to-candidate allowlist: the prior 17 paths plus the four paths above (21 total).
Any additional path is a stop-and-replan event, never a silent scope expansion.

### Cycle 6 exact symbol map

| Contract | Required production symbols |
| --- | --- |
| real broker composition | `GitHubDecisionBroker.readRecord`; an exported orchestrator adapter over `request`, `poll`, `consume`, and the broker-owned reread; real composition coverage in both broker and orchestrator tests |
| shared credential grammar | `review-router.ts` `redactSensitiveText`/`assertNoSensitiveText`; `human-decision.ts` `normalizeQuestion` imports the shared assertion and removes its duplicate grammar |
| intrinsic signal lease | `GitHubParentOrchestrator.lifecycleContext`, `withLifecycle`, `waitForLifecycle`, `callExternal`, `stop`, `#lifecycleScope`, and `ExternalCallContext.acknowledgeAbort` |
| intrinsic raw/proxy bounds | `review-router.ts` `decodeBoundedJsonPayload`, `readBoundedExactRecord`, and `exactArrayValues`; `github-evidence.ts` and `github-orchestrator.ts` `boundedArray` |
| ordered stable reviews | `review-router.ts` `reconcileIndependentReview` and a separate stable semantic clean-authorization digest while retaining the full attempt digest for attestation |
| canonical decision DTO | `human-decision.ts` record/field readers, normalizers, `validateHumanDecisionRequestComment`, and `validateHumanDecisionRecord` |

### Cycle 6 correction contract

1. Preserve the real `GitHubDecisionBroker` request/full-record, poll/result, and
   consume/evidence shapes through one explicit production adapter backed by the broker's own
   canonical `readRecord` boundary. The adapter owns context linkage, bounded validation, and exact
   request/result/reload coherence; it cannot accept a second repository or synthesize decision
   evidence. Strict compile and runtime composition use an actual `GitHubDecisionBroker` through
   pending -> decided -> consumed. Test fakes expose the same production shapes.
2. Capture every request, poll, consume, and reload value through byte/descriptor-bounded,
   proxy/accessor-safe closed DTOs before canonical human-decision validation. All failures become
   stable typed/sanitized boundary errors. Decided/consumed records require persisted request-comment
   provenance and ordered create/comment/decision/consume/update timestamps.
3. Represent parent readiness as one controller-authorized conditional effect. A canonical token
   binds the complete current policy set, exact consumed decision record, current parent review and
   path authority, complete child receipt plus ancestry roster, canonical plan, PR head, and PR
   resource revision. The transport compares/consumes that token atomically with clearing draft;
   authorization movement yields no visible ready effect. Any post-effect drift invokes a typed,
   idempotent rollback-to-draft effect and verifies rollback before return.
4. Lease caller/stop signals only through the intrinsic `AbortSignal.aborted` getter and native
   `EventTarget` add/remove methods. Reject proxies/incompatible receivers, attach then recheck,
   never execute own shadows, and accept abort acknowledgement only after local abort. The real
   broker adapter remains inside tracked settlement/key/stop accounting.
5. Order all attested exact-head review attempts by authoritative completion before verdict. A
   later findings attempt invalidates earlier clean; a later clean resolves findings. Keep full
   attempt digest/timestamp for attestation but use one stable semantic clean-authorization digest
   for child mutation/receipt identity, so equivalent later clean does not fork one key+intent and
   timeout/restart recovery remains deterministic.
6. Read raw `Uint8Array` size with the intrinsic byte-length getter after exact compatible-receiver
   validation and before decode. Reject normal/revoked proxies before `Array.isArray`; normalize
   host errors across record/array readers while retaining dense/accessor/cardinality bounds.
7. Make one exported credential grammar authoritative for orchestration, evidence, review, and
   human decisions. Add npm `_authToken`, netrc whitespace assignment, lowercase cloud credential
   keys, and well-known credential-file forms; table-test every #478/human-decision durable or
   outbound text boundary using synthetic markers only.
8. Require receipt `integratedAt` no earlier than its PR snapshot, observation, path evidence,
   review completion, policy observation, or controller observation and never impossibly in the
   future. Repair fixture chronology and add each failure row.
9. Replace self-referential durable candidate state with exact non-circular semantics:
   `candidateRef: "HEAD"`, current cycle/state, and exact prior checkpoint commits; parent handoff
   binds the resulting evidence commit SHA externally. Cycle 4 is historical, Cycle 5 review truth
   is explicit, and no current Cycle 5/6 evidence field is null.
10. Retain all Cycle 5 policy refresh, receipt reauthorization, child eligibility, CAS identity,
    caller/key retention, bounded stop/raw envelopes, topology/path/scope checks, and no-parent-main-
    merge guarantees.

### Cycle 6 strict lifecycle

1. Commit this PLAN/scope decision/finding matrix before any test or production edit. Record all
   five relevant production blob IDs at frozen `63ac436f`.
2. Make one comprehensive test/fixture-only RED commit. Run retained Cycle 5 tests separately and
   require 109/109 pass; every new row must fail intentionally. The real broker composition test
   must instantiate `GitHubDecisionBroker` itself, not a stronger fake. Prove all relevant
   production blobs still equal frozen `63ac436f` before GREEN.
3. Implement one coherent architectural GREEN/refactor. Any post-RED fixture edit must preserve
   production shapes, must not weaken/remove expectations, and must be enumerated in evidence.
4. Run the focused five files including the real broker composition test, strict owned and all 20
   production Shepherd modules against pinned Pi 0.80.6, pinned offline RPC, serialized Shepherd
   classification, immutable-base/ancestry/diff/exact-expanded-scope, JSON, and synthetic-marker
   scans. Do not run Go, connectors, certification, runtime services, `make`, network/live GitHub,
   #479 implementation, reviewer, integration, or merge actions.
5. Commit evidence and report exact PLAN/RED/GREEN/evidence SHAs plus the clean candidate for two
   fresh parent-owned exact-head `xhigh` reviews.

### Cycle 6 checkpoints

- [x] Initial artifact-only PLAN/finding-matrix commit:
      `88513259ffc31fd0853679234c6a42ab6cd04ef6`.
- [x] Artifact-only broker-map scope amendment from 18 to 21 paths:
      `2832993b93d07ea20197bad52ec23700fe21fc1e`.
- [x] One comprehensive test/fixture-only RED with retained 109/109 and frozen production proof:
      `ca4d97d1100b1b44176da9d7dfd6ee6f56f4e1e6`.
- [x] One coherent architectural GREEN/refactor:
      `2c6371e725d58b2dc05902d68f9e6812904664d6`.
- [x] Authorized local verification/evidence uses non-circular `HEAD` state; the parent handoff
      records the resulting evidence commit SHA externally.
- [ ] Two fresh exact-head independent reviews remain parent-owned.

### Cycle 6 local result

The coherent GREEN composes the actual repository-owning `GitHubDecisionBroker`, closes hostile
decision DTO and chronology boundaries, makes ready transition a complete conditional
authorization with verified rollback, leases cancellation through native intrinsics, orders all
attested exact-head review attempts while preserving stable semantic authorization, hardens raw
and proxy bounds, centralizes the credential grammar, and enforces receipt chronology.

The five focused files pass 206 assertions with zero failures and one intentional live-GitHub skip;
the retained Cycle 5 slice remains 109/109. Strict TypeScript passes for all owned files and all 20
production Shepherd modules, and pinned Pi 0.80.6 offline discovery succeeds. Immutable-base,
ancestry, full-range diff, exact 21-path ownership, JSON, and synthetic-marker scans pass. The broad
serialized suite is recorded separately because the managed sandbox still rejects unrelated
process-identity child spawns with `EPERM`. No prohibited or external action ran.

## Cycle 7 consolidated exact-head correction: 2026-07-22

Two independent Cycle 6 reviews inspected exact candidate
`dbce5b7d0c698bc802594211072fed77eff23c1c` against immutable base
`3addb1f48be1afe8b1e2b59b54247679d7293805`. Both reports were read completely. Their unique union
is accepted as one coherent correction: the test-only parent-ready authority oracle, escaping late
effects, volatile ready identity, unauthenticated review-attempt audit fields, future broker
chronology, incomplete credential grammar, ambiguous run state, and the missing #479-shaped
production-port-only seam must close together.

Required routing loaded before edits: `gsd-programming-loop`, `architecture-patterns`,
`javascript-testing-patterns`, `github-issue-first-delivery`, `caveman`, required-skills routing,
the issue-agent contract, Pi adapter and universal-loop references, and required GSD project
artifacts. `scripts/gsd doctor` passes. `scripts/gsd prompt programming-loop init --phase
478-shepherd-github-parent-orchestration --dry-run` reports that `programming-loop` is unavailable,
so Cycle 7 records `manual_gsd_fallback` and follows strict PLAN -> RED -> GREEN -> REFACTOR. The
phase-specific coverage matrix remains in this PLAN and `TDD-LEDGER.md`; adding a new
`PRD-COVERAGE.md` path is unnecessary and would violate the exact owned scope.

### Cycle 7 exact scope and frozen production

The base-to-candidate allowlist remains exactly the Cycle 6 21 paths. No new path is required.
Production blobs frozen at reviewed Cycle 6 candidate `dbce5b7d` are:

- `github-orchestrator.ts`: `b3515a94e932a6206f2c32f083c1188882a01dfe`
- `github-decision-broker.ts`: `25c98a3c224d660c7fe6b5de16a30fdf73f95621`
- `human-decision.ts`: `4202ba001dd0d48b83d68a65b7004c8db49d0b65`
- `review-router.ts`: `a113b4d6bb77f001e8b377c2696c934136b4ceb9`
- `github-evidence.ts`: `23efd2c51280ba83836feef4fcb459e7da4571c0`

No agent slot is available at planning time; the parent, #475 worker, #479 preflight, and this
ordered #478 worker occupy all four slots. Cycle 7 therefore records `local_critical_path` rather
than delaying or duplicating work. The parent-provided #479 preflight observation is incorporated
as part of the typed-wiring finding below.

### Cycle 7 architectural contract

1. Remove parent-ready mutation and rollback from `GitHubOrchestrationTransport`; construction
   requires one explicit production `ParentReadyDurableAuthorityBoundary` with no optional or
   legacy fallback. Its conditional method owns the
   authoritative compare, durable authorization consumption, exact PR draft/revision CAS, and
   ready effect as one indivisible port operation. Its recovery method durably quarantines that
   authorization before performing an idempotent rollback-to-draft. Export closed canonical
   builders/validators so an implementation can compare current policy, review, exact paths,
   canonical child receipts and literal ancestry, consumed decision, plan/head, and PR CAS without
   reproducing private controller logic. Every coordinate movement inside the boundary conflicts
   before draft is cleared; controller rereads and rollback are exceptional recovery only.
2. Expose a production-typed prepare/commit split and public `ParentReadyOperationJournal`.
   `prepareParentReadiness` returns a canonical
   prepared operation containing the authoritative consumed decision and durable mutation intent;
   `commitPreparedParentReadiness` revalidates and submits it to the durable authority boundary.
   This lets #479 durably journal prepare/decision-consume before the conditional effect and record
   settlement afterward without private callbacks or invented protocol. The existing
   `reconcileParentReadiness` remains a convenience wrapper. Prepared values are closed,
   digest-bound, and unusable after semantic movement.
3. Split parent-ready authorization into a stable semantic projection and a separately validated
   freshness envelope. Stable identity includes repository/base/revision/policy digest (never
   policy `observedAt`), exact ancestry repository/ancestor/descendant/literal result (never proof
   revision/time), `independentReviewAuthorizationDigest`, stable exact paths, consumed decision
   identity, canonical child-receipt identities, plan/head, and PR revision CAS. Equivalent policy
   or ancestry observations and a later equivalent clean attempt preserve both idempotency key and
   intent digest across retry/restart; real semantic movement conflicts.
4. Bind each receipt's `reviewResultDigest` and `reviewCompletedAt` to the complete authoritative
   controller-owned attestation history returned for that exact target. Stable review
   authorization remains attempt-independent, but a forged full-attempt digest/time can never be
   reused or authorize readiness. Later equivalent clean attempts remain compatible; the latest
   ordered findings attempt still blocks.
5. On uncertain parent-ready timeout/cancellation, immediately enter keyed quarantine, retain the
   ensure lease through the original invocation's settlement and recovery, and drive the durable
   boundary's quarantine+rollback until an exact newer draft is confirmed. Timeout-before-effect,
   timeout-after-effect, late settlement, transient read/rollback failures, and restart-before-
   visibility cannot release an unowned live effect or report a clean terminal join. `stop()` joins
   finalizers and reports incomplete/unacknowledged work when a port remains uncooperative.
6. Validate every canonical human-decision record against a broker/controller-owned observation
   clock with exported explicit GitHub timestamp skew. Bound `createdAt`, request-comment time,
   `decidedAt`, `consumedAt`, and `updatedAt` independently and together. The actual
   `GitHubDecisionBroker` repository reread and the orchestrator adapter both enforce the bound;
   impossible future evidence never reaches parent-ready.
7. Replace the open-ended uppercase credential-assignment suffix with finite shared schema/prefix
   rules. Add Kubernetes `client-key-data`/token forms, Docker config `auth`/identity-token forms,
   AWS access-key/session assignments and closely related well-known prefixes. Table-drive every
   new synthetic marker through human-decision persistence/comment, parent/child plan and
   title/body fields, findings, and dispositions; rejection precedes persistence/effect and errors
   never reflect the marker.
8. Make `candidateRef: "HEAD"` the only current-candidate semantic in `RUN-STATE.json`. Nest exact
   Cycle 6 SHA `dbce5b7d...` under historical prior-review state, remove the stale Cycle 5
   `frozenCandidate` current field, and add a schema-level invariant rejecting any historical SHA
   in current-candidate fields.
9. Add one true #479-shaped wiring trajectory using only exported production interfaces and public
   prepare/commit methods. It composes actual policy, real broker adapter, durable prepared-intent
   journal, atomic authority/readiness, quarantine/rollback, and stop/join contracts. It has no
   fake-only validator property, `Object.assign` seam, type cast, or private helper duplication.
10. Retain all 206 passing Cycle 6 focused assertions plus the intentional live skip. Do not touch
    Go, connectors, main/parent worktrees, controller/#479 implementation, dependencies, network,
    live GitHub, reviewer dispatch, integration, or merge.

### Cycle 7 RED matrix (46 planned behavior rows)

| Family | Rows | Required failing behavior before GREEN |
| --- | ---: | --- |
| atomic authority movement | 10 | policy, review, paths, receipt, ancestry, decision, plan, head, PR revision, and durable authorization state move inside the effect; all remain draft without relying on reread/rollback |
| late-effect ownership | 6 | timeout before/after effect, restart before visibility, read failure, rollback failure/retry, and no unowned live effect/false clean join |
| semantic identity vs freshness | 7 | policy observation refresh, ancestry metadata refresh, equivalent later clean, combined retry, restart, and true policy/review/path semantic movement |
| attested receipt attempt provenance | 5 | forged result digest, forged completion time, later equivalent clean, later findings, and restart/history ordering |
| owned-clock broker chronology | 6 | future create, request comment, decision, consume, update, and all-at-once through actual broker plus ready path |
| shared credential forms | 8 | Kubernetes key/token, Docker auth/identity-token, AWS access/secret/session assignments, and well-known temporary-access prefix across every durable/outbound boundary |
| RUN-STATE schema | 2 | one HEAD current semantic and synthetic historical-SHA/current-field rejection |
| #479 port-only wiring | 1 | public prepare/journal/commit/settle path over production interfaces only |
| Cycle 6 retention | 1 | retained 207-test focused suite classification is unchanged |

### Cycle 7 strict lifecycle and verification

1. Commit this artifact-only plan, scope, finding matrix, run-state correction, and verification
   checklist before any test or production edit.
2. Make one comprehensive test/fixture-only RED commit. Retained Cycle 6 assertions must remain
   green; each new contract family must demonstrate its intended failure. Prove all five frozen
   production blobs above remain exact before GREEN.
3. Implement one coherent GREEN/refactor only in the owned production modules and aligned tests.
   Any post-RED test support change must preserve every expectation and be enumerated.
4. Run the complete five-file focused matrix, strict owned and all-production TypeScript against
   pinned Pi 0.80.6, pinned offline Pi RPC, immutable-base/ancestry/diff/exact-21-path/JSON/synthetic-
   marker gates. Broad Shepherd may run only serialized; `/bin/ps`/spawn `EPERM` remains a separately
   reported environment classification.
5. Commit evidence only after the worktree is cleanable and artifacts describe exact HEAD. Report
   PLAN/RED/GREEN/evidence commits, exact base and candidate, path list/count, counts, and known
   environmental failures. Parent owns fresh review; no self-review or integration occurs.

### Cycle 7 checkpoints

- [x] Artifact-only PLAN/finding matrix commit
      `2c64979829048d3de0d1ff1575c2a4f43cb699ba`.
- [x] Comprehensive test-only RED `10033bc532d06967ce960e408c2bc9725020478a`:
      290 total, 217 pass, 72 intentional failures, 1 intentional live-GitHub skip; strict
      TypeScript reports only the 14 absent Cycle 7 contracts and all five production blobs remain
      frozen.
- [x] Coherent architectural GREEN `5bab0bc7e56292171eb28618cc2f37488ed1b7a4`.
- [x] REFACTOR proof `87e704010f3e2226d8393d12e1a1bdf72df212a0` fixes the timeout
      contract at a 500 ms late effect after a 100 ms timeout, adds caller cancellation before the
      same late effect, keeps semantic chronology strict, and weakens no RED assertion.
- [x] Independent pre-freeze architecture audit RED
      `b1560e76a3abbac5efcd33b2740b7275b6acc137`: 297 total, 294 pass, 2 intentional
      failures, and 1 intentional skip expose the remaining optional legacy ready-mutation route
      and fake-shaped #479 role projection.
- [x] Audit GREEN `915882c219f52da2c1edebce84d2bf90c61a4592`: 297 total, 296 pass,
      0 fail, and 1 intentional skip; authority is mandatory, transport has no ready mutation,
      compare conflicts are typed, and transport/authority/journal are separate production roles.
- [x] Exact-head local verification/evidence uses the non-self-referential `HEAD` candidate and is
      committed only after the evidence below is complete.
- [ ] Fresh exact-head review remains parent-owned.

### Cycle 7 local result

Both Cycle 6 reports were replayed line by line after REFACTOR. Named passing tests cover all ten
atomic movement coordinates; before- and after-effect 500/100 timing; caller cancellation;
keyed/durable quarantine across restart; read failure; rollback retry before key/join release;
stable key and mutation intent under harmless policy, ancestry, and equivalent-clean refresh;
semantic movement; authoritative full-attempt digest/time provenance; owned-clock future event
coordinates through the real broker adapter; all eight finite Kubernetes, Docker, and AWS forms;
one current `HEAD` run-state semantic; and the public production-port-only #479-shaped
prepare/journal/commit seam. The public transport has no parent-ready mutation fallback, the
authority boundary is mandatory, and the #479 proof uses separate production-typed transport,
authority, and journal roles rather than a structural `FakeTransport` projection.

The five-file focused suite records 297 total, 296 pass, 0 fail, and 1 intentional live-GitHub
skip. Strict owned and all-20-production TypeScript pass, and pinned Pi 0.80.6 offline RPC discovers
`pm-shepherd` from `extension`. The serialized Shepherd suite is an environmental failure, not a
pass: 517 total, 451 pass, 65 unchanged unrelated managed-sandbox process-identity `spawn EPERM`
failures, and 1 intentional skip. Immutable base and reviewed-candidate ancestry, exact merge base,
full-range `git diff --check`, exact 21-path scope, three JSON documents, and the explicit
test-synthetic credential-marker allowlist pass. No prohibited or external action ran.

## Cycle 8 exact-head recovery correction: 2026-07-22

Two independent reviews block frozen Cycle 7 candidate
`b90037df1fff38c755ebc8025579120d17031330` against immutable base
`3addb1f48be1afe8b1e2b59b54247679d7293805`. Both reports were read completely. Their seven
unique families form one indivisible correction: provider-neutral credential assignments,
strictly typed #479 composition, every uncertain non-value outcome, real-broker resume after
expiry, bounded fenced rollback, reconstructed durable restart, and refreshed freshness delivery.
No family may be frozen or reviewed separately.

Required routing loaded before edits: `gsd-programming-loop`, `architecture-patterns`,
`javascript-testing-patterns`, `github-issue-first-delivery`, `caveman`, required-skills routing,
the issue-agent contract, runtime/Pi integration reference, GSD adapter/universal-loop references,
and project loop artifacts. `scripts/gsd doctor` passes; `scripts/gsd prompt programming-loop init
--phase 478-shepherd-github-parent-orchestration --dry-run` returns `unknown GSD command:
programming-loop`, so the recorded route remains `manual_gsd_fallback`. One read-only explorer maps
the seven coupled contracts; this isolated worker owns the ordered PLAN -> RED -> GREEN -> REFACTOR
critical path.

### Cycle 8 exact scope and frozen production

The owned immutable-base range remains exactly the existing 21 paths. No new file is needed.
Production blobs frozen at candidate `b90037df` before RED are:

- `github-orchestrator.ts`: `668a55af55413c1cc595424e87ce352c355eec88`
- `github-decision-broker.ts`: `7be6785190176a8c15660fb180fc95c207b76d5b`
- `human-decision.ts`: `b1c0c198c33c95c8fabb0f911a42513d2305cb17`
- `review-router.ts`: `a586405153e2e666a57b832e7d4b48df80e3265c`
- `github-evidence.ts`: `23efd2c51280ba83836feef4fcb459e7da4571c0`

### Cycle 8 architecture contract

1. Classify assignment-shaped uppercase names by the complete recognized credential suffix
   grammar before applying a narrow exact safe-name exception. Unknown provider/vendor prefixes
   never bypass classification. Preserve the finite Kubernetes, Docker, and AWS forms. Every
   suffix is table-driven through review routing, native decision persistence/comment creation,
   broker persistence/outbound comments, PR title/body/findings/dispositions, and orchestration
   plan/decision text; rejection precedes effects and never reflects the synthetic marker.
2. Make the #479 fixture compile without `any`, casts, fake projections, or private shortcuts.
   Separate production-typed transport, authority, and journal adapters return exact
   `GitAncestryProof`, `ParentReadyCompareEffectResult`, and
   `DurableMutationResult<GitHubPullRequestEvidence>` values across success, typed conflict,
   uncertainty, rollback, stop/join, and settlement.
3. Treat every uncertain non-value result from `callExternal` alike. Timeout, cancellation, and
   immediate rejection after a possible effect all start the durable quarantine finalizer. The
   original invocation plus recovery retain keyed and stop/join ownership even when ordinary
   visibility reads fail.
4. Split canonical request normalization from new-request lifetime validation. An existing exact
   broker record is read and compared before creation-time expiry is considered. A decision
   created/decided before expiry and durably consumed may revalidate after restart; a genuinely
   new request or decision after expiry remains rejected.
5. Bind every rollback to a stable recovery ID and an ordered attempt fence. Each authority call
   must durably claim its attempt and supersede its predecessor before returning; its only permitted
   effect is idempotent restoration of the exact draft owned by the original ready mutation
   key/intent. The request carries that original mutation identity; the controller never guesses or
   alternates resource revisions. The controller enforces a real per-attempt
   deadline, aborts the response wait, retains timed-out promises until a later fenced result proves
   durable handoff, and never lets a superseded result release quarantine or settle controller
   state. A matching fenced `DurableMutationResult` containing the exact draft is the authoritative
   observation that permits release. Until then reentry is blocked and `stop()` is incomplete.
6. Remove module `WeakMap`/authority-object identity from restart evidence. Serialize the prepared
   operation and journal, recreate controller, broker, journal, transport, and authority adapter
   instances over shared durable backing, reconcile the uncertain settlement, and resume once.
   Cross-instance fencing belongs to the durable authority backing; in-process maps are only local
   scheduling aids.
7. Public prepared commit retains the journaled stable authorization, idempotency key, and intent,
   but after successful revalidation sends the newly observed freshness envelope to the atomic
   authority. Harmless policy, ancestry, and equivalent-clean review refresh must be visible in
   that request while stable identity remains byte-equivalent.

### Cycle 8 RED matrix (48 behavior rows)

| Family | Rows | Required failing behavior before GREEN |
| --- | ---: | --- |
| provider-neutral credential closure | 20 | every recognized authorization/token/access-token/refresh-token/API-key/password/secret/client-secret/private-key/database-URL/credential(s)/cookie(s)/set-cookie/session/id/token/cookie/CSRF-token suffix under an unknown prefix rejects through every durable/outbound consumer; exact safe-name control remains allowed only after classification |
| strict #479 production roles | 6 | exact typed success, non-applied conflict, uncertain late effect, quarantine/rollback, live `stop()` incomplete then joined, and durable settlement without `any`, cast, fake projection, or private shortcut |
| every uncertain non-value outcome | 4 | apply-then-reject starts recovery with all reads failing, restores draft, blocks keyed reentry, and keeps stop incomplete until joined |
| real-broker expiry resume | 4 | existing consumed request resumes after expiry, prepared commit applies once, new expired request rejects, and new post-expiry decision rejects through the actual adapter |
| bounded fenced rollback | 5 | ignored first response receives an abort signal, later attempt succeeds without revision guessing, superseded late result cannot settle, quarantine lasts through authoritative draft observation, and stop eventually joins |
| reconstructed durable restart | 4 | serialized prepared/journal state recreates all five production roles, shares only durable backing, reconciles uncertainty without `WeakMap` identity, and resumes exactly once |
| refreshed freshness delivery | 5 | refreshed policy, ancestry, equivalent-clean review, combined envelope, and unchanged authorization/key/intent are asserted at the atomic authority |

### Cycle 8 strict lifecycle and verification

1. Commit this artifact-only PLAN and 48-row ledger before any Cycle 8 test or production edit.
2. Commit one complete five-test-file RED. Retain all 297 Cycle 7 focused cases and the intentional
   live skip; prove every new contract fails and all five frozen production blobs remain exact.
3. Implement one coherent GREEN, then one bounded REFACTOR if structure or lifecycle clarity
   requires it. No expectation may be removed, weakened, skipped, or converted to a fake-only proof.
4. Run focused five-file tests, strict owned and all-production TypeScript against pinned Pi 0.80.6,
   pinned offline RPC, immutable-base/reviewed-candidate ancestry, exact merge base, full-range diff,
   exact 21-path scope, three JSON parses, explicit synthetic-marker classification, and clean status.
   Serialized Shepherd runs only after focused GREEN and is reported as pass or environmental
   failure without reinterpretation. Go, connectors, `make`, runtime services, and live actions stay
   out of scope.
5. Replay both reports after REFACTOR. Commit exact evidence only once every seven-family contract
   and retained gate is truthful. Parent owns publication, fresh exact-head review, integration,
   and every human gate.

### Cycle 8 checkpoints

- [x] Both Cycle 7 reports read completely and consolidated without dropping warnings.
- [x] Exact base, frozen candidate, clean worktree, 21-path range, and five frozen production blobs
      confirmed before edits.
- [x] Required skills/contracts loaded; healthy adapter plus missing command records
      `manual_gsd_fallback`; one read-only contract mapper records `read_only_spawned`.
- [x] Artifact-only PLAN `bccee8e6cdbcb6e38419114f264222b1f5616f66` precedes all test and
      production edits.
- [x] Comprehensive test-only RED `851bb3bfa3e23042211a8b37f3a97253cc6fedf5` proves all 48 rows with
      frozen production.
- [x] Coherent GREEN `013bdc8b264e1ce8808d4af2558e2ec40b85ee49` and bounded REFACTOR
      `26a7d476bdfaa4e263196fb76f7f43b5a3ad799e` pass retained and new contracts.
- [x] Exact evidence uses the non-self-referential current candidate `HEAD`; its commit SHA is
      reported externally after commit. Two fresh reviews remain parent-owned.

### Cycle 8 local evidence

- RED: 374 total, 314 pass, 59 intended failures, 1 intentional live-sandbox skip. Strict
  TypeScript reported only four intended missing-contract diagnostics. All five production blob
  IDs remained frozen.
- GREEN/REFACTOR: 46/46 targeted Cycle 8 orchestrator cases and the complete focused five-file
  route pass at 374 total / 373 pass / 0 fail / 1 intentional skip. Both complete Cycle 7 reports
  were replayed after REFACTOR; all seven families map to named passing cases.
- Strict TypeScript passes for the five owned production/test pairs and all 20 Shepherd production
  modules against pinned Pi 0.80.6 declarations. Offline pinned Pi RPC discovers `pm-shepherd`
  from the explicit `index.ts` extension.
- Serialized Shepherd is honestly classified as an environmental failure: 594 total / 528 pass /
  65 unchanged unrelated managed-sandbox process-identity `spawn EPERM` failures / 1 intentional
  skip. Every Cycle 8 and focused assertion passes.
- Immutable base and reviewed candidate ancestry, exact merge base, full-range diff check, exact
  21-path ownership, three JSON parses, synthetic-marker confinement, and clean pre-evidence status
  pass. No prohibited or external action ran.

## Cycle 9 consolidated-review correction

Frozen reviewed candidate: `f97a698df90010ae072554e04563a8134a8e5f6e`; immutable base:
`3addb1f48be1afe8b1e2b59b54247679d7293805`. Both complete Cycle 8 reports were read before this
plan. Their two blocker sets and typed-fixture warning are accepted as one indivisible four-family
architecture correction. The exact 21-path boundary is unchanged. All five production blobs are
frozen before RED: orchestrator `ab9b2c0ed254ecdbffa10c4ca2b13420de01268a`, broker
`7be6785190176a8c15660fb180fc95c207b76d5b`, human decision
`fc1c62307ccca0c2590ea0a7cd61626876f3f71f`, review router
`31234c70ade7341a2af01aeac2d81a015b696e6b`, and evidence
`23efd2c51280ba83836feef4fcb459e7da4571c0`.

### Cycle 9 authority-owned recovery protocol

The durable authority, never a controller-local `Set`, owns one exact record for a parent-ready
invocation. The record persists `invocationId`, `recoveryId`, repository/PR/marker/generation/head,
the original PR revision, stable ready mutation key and intent, optional rollback mutation, phase,
status, and a monotonic fence. It is canonically validated at every read/write boundary.

The only valid state transitions are:

1. `ready_invoking` (`unsettled`, fence 0): persisted before the ready effect. The original writer
   must re-read/CAS this exact state immediately before writing.
2. `ready_effect_applied` (`unsettled`, fence 0): the effect is durable, but the caller has not yet
   validated and durably settled its response. Visible non-draft state is not reusable here.
3. `ready_settled` (`settled`, fence 0): reached only through a separate exact settlement CAS after
   the applied result is validated. This is the only authority state that may authorize a ready
   result or already-ready reuse.
4. `recovery_claimed` (`unsettled`, fence >= 1): an atomic recovery claim increments the fence and
   thereby invalidates every lower rollback attempt plus the original fence-0 writer/replay before
   any draft effect. It may be entered from either unsettled ready state, including reconstructed
   state after response loss.
5. `draft_restored` (`settled`, fence >= 1): reached only after the matching fenced rollback
   mutation proves the exact authorized PR/head is draft. Stale writers and stale rollback results
   cannot transition or settle it; a later public call may then prepare one fresh invocation.

The authority exposes typed read and settlement operations in addition to compare/effect and
rollback. Prepare, commit, and reconcile query it before any already-ready shortcut. Every uncertain
`ExternalPortError`, including immediate promise rejection after an applied effect, returns only a
typed blocked/quarantined outcome while recovery remains owned by the keyed lifecycle. Healthy
visibility never overrides unsettled authority. `stop()` remains incomplete until matching
`draft_restored` settlement joins. A reconstructed controller obtains prepared operation data from
the journal and recovery truth from serialized authority values; no object identity is evidence.

### Cycle 9 RED matrix (69 behavior rows)

| Family | Rows | Required failing behavior before GREEN |
| --- | ---: | --- |
| uncertain result consistency | 8 | immediate apply-then-reject with healthy ready visibility returns blocked, never ready; recovery remains keyed, prevents reentry, holds stop incomplete, restores exact draft, joins stop, records blocked settlement, and treats every uncertain `ExternalPortError` alike |
| durable dangerous-point restart and original-writer fence | 13 | queryable invocation/recovery IDs, target/revision/head, stable mutation identity, five phases, monotonic fence, pre-shortcut prepare/commit/reconcile checks, serialization while ready is visible before rollback, reconstructed five-role recovery, stale original writer suppression, exact one-time draft restoration, truthful stop/join, and one fresh resume |
| total provider-neutral assignment parsing | 40 | leading underscore, lengths 127/128/129/256, each consumer's largest in-field assignment, over-field assignment, and exact `FEATURE_TOKEN` control are tabled through each of five shared durable/outbound consumers; the complete name is parsed to its delimiter and failures never reflect the marker |
| exact #479 value-serialized production fixture | 8 | public typed broker plus `JSON.parse` into `unknown` canonically decode decision, prepared operation, journal, authority recovery, fence, mutation, and settlement snapshots across success, conflict, uncertainty, restart, incomplete/joined stop, and final settlement without `any`, casts, fake projection, or private shortcuts |

### Cycle 9 lifecycle and verification

1. Commit these nine artifact-only updates before every Cycle 9 test or production edit.
2. Commit one complete five-test-file RED. Retain all 374 focused cases and the intentional live
   skip; prove all 69 rows fail for the intended missing contracts while all five production blobs
   remain byte-exact.
3. Implement one coherent GREEN followed by a bounded structural REFACTOR if needed. The durable
   record and state transitions are one architecture slice; no partial family may be frozen.
4. Run targeted Cycle 9, the complete focused five-file route, strict owned and all-production
   TypeScript against pinned Pi 0.80.6, pinned offline RPC, exact base/ancestry/merge-base/diff/
   21-path checks, three JSON parses, marker-confinement scans, report replay, and clean status.
   Serialized Shepherd is run only after focused GREEN and classified truthfully. Go, connectors,
   `make`, runtime services, parent/main/#475, dependencies, network/GitHub, push, reviewers,
   integration, and merge remain prohibited.
5. Evidence remains non-self-referential `HEAD`; parent owns publication, two fresh exact-head
   independent reviews, dispositions, integration, and every human gate.

### Cycle 9 checkpoints

- [x] Both Cycle 8 reports read completely and frozen candidate/base/scope/blobs confirmed clean.
- [x] Required skills/contracts loaded; doctor passes, unavailable programming-loop command records
      `manual_gsd_fallback`; all agent slots are occupied so execution records `local_critical_path`.
- [ ] Artifact-only Cycle 9 PLAN commit precedes tests and production.
- [ ] Comprehensive five-file RED preserves all prior behavior and freezes production blobs.
- [ ] Coherent GREEN and bounded REFACTOR close all 69 rows.
- [ ] Exact local evidence and both-report replay are recorded; fresh reviews remain parent-owned.
