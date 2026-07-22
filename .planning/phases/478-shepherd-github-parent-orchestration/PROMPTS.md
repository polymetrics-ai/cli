# Prompt Trace: #478

## Kickoff snapshot

- Coordinator objective: implement issue #478 in exactly the three assigned production modules,
  matching tests/fixtures, and this issue-local phase directory.
- Model role: `openai-codex/gpt-5.6-sol:high` implementation worker.
- Immutable base: `3addb1f48be1afe8b1e2b59b54247679d7293805` on
  `feat/471-pi-agent-session-shepherd`.
- GSD preflight: `scripts/gsd doctor` passed; `scripts/gsd prompt programming-loop init --phase
  478-shepherd-github-parent-orchestration --dry-run` returned `unknown GSD command:
  programming-loop`.
- Runtime decision: `manual_gsd_fallback` plus `local_critical_path`. Two attempts to dispatch the
  read-only API-recon subtask were rejected by the runtime agent-thread limit.
- Contract resolution: the live #478 issue/coordinator handoff supersedes an older local draft;
  only exact-head `codex_independent` review using `openai-codex/gpt-5.6-sol:xhigh` can satisfy the
  automated review gate. Claude, Copilot, generic Codex, and human records are rejected for that
  gate.
- Human gates: all policy exceptions and the final parent ready/merge decision route through the
  existing broker. This worker will represent but not cross them.

## Stable-head resume

- The parent temporarily paused #478 review work while #475 established a stable predecessor,
  then explicitly resumed this branch without resetting its pushed TDD history.
- Reconciliation found pushed head `db9fbc33` (the adversarial test-only RED checkpoint), with
  prior production GREEN at `90321ffb`; the uncommitted matching correction was preserved.
- Parent review policy: finish and push an exact GREEN candidate, but do not start any reviewer.
  The parent orchestrator owns the later stable-head specialist campaign.
- Correction GREEN: `40ce66d4b5010b92089895a05709687143d15a05`; 27/27 focused tests and
  strict owned TypeScript pass before broader verification.
- Authorized broader verification only: serialized Shepherd TypeScript tests, strict pinned Pi
  0.80.6 TypeScript, pinned offline Pi RPC, and immutable-base/diff/owned-scope checks. No Go,
  connector, certification, runtime-service, or `make` command is permitted.

## Functional review correction kickoff

- Frozen reviewed head: `093b3c90409cedc6b7008b7510f53937eb1ebbc1`; frozen base:
  `3addb1f48be1afe8b1e2b59b54247679d7293805`.
- Review input: `/tmp/478-REVIEW-FUNCTIONAL.md`, 7 critical and 4 warning findings, all accepted
  for one strict correction slice.
- Sequence: artifact-only plan commit, one behavior-level test-only RED commit covering all eleven
  findings with production byte identity proof, coherent GREEN, authorized Shepherd verification,
  push/update PR #487, then parent-owned fresh xhigh review.
- Scope exclusion: no #475/#479/controller/top-level wiring; the session provenance addition is a
  scoped interface and fixture only.
- GSD: `manual_gsd_fallback`; cycle decision `local_critical_path` in the isolated #478 clone.

## Cycle 3 corrected-review prompt

- Inputs: `/tmp/478-REVIEW-CORRECTED-1.md` and `/tmp/478-REVIEW-CORRECTED-2.md`, both read in full.
- Frozen candidate/base: `3f285722a505ea426d53a34f95716781d1aca7c2` /
  `3addb1f48be1afe8b1e2b59b54247679d7293805`.
- Implementation/correction route: Codex `openai-codex/gpt-5.6-sol:high`,
  `local_critical_path`; `xhigh` remains parent-owned planning/review/orchestration only, and no
  Claude/Copilot finding is imported.
- Sequence: artifact-only plan commit; exactly one test/fixture-only RED commit proving all three
  production blobs unchanged; architectural GREEN/refactor; focused/authorized broad verification;
  evidence commit. Network publication remains deferred under the recorded DNS failure.
- Fourteen-invariant batch: canonical persisted plans; mutating-only children; exact outer/nested
  evidence identity, completeness, chronology, and freshness; receipt provenance; durable
  cross-instance mutations; exact ancestry; deterministic same-marker review attempts; symbolic-ref
  rejection; versioned plan-bound CI; monotonic roster CAS; exported attestation API; adversarial
  bounds, partial effects, and secret safety.
- Exclusions: #479 controller, parent planning artifacts, live GitHub, Go/connectors/certification,
  runtime services, `make`, external reviewer, integration, or merge.

## Cycle 4 consolidated-review prompt

- Inputs: `/tmp/478-REVIEW-CYCLE3-1.md` and `/tmp/478-REVIEW-CYCLE3-2.md`, read completely.
- Frozen candidate/base: `d3b6b5e226b17db6ec8350163acdbb41368ec3bf` /
  `3addb1f48be1afe8b1e2b59b54247679d7293805`.
- Implementation route: `openai-codex/gpt-5.6-sol:high`; parent-owned review is `xhigh`.
- Strict sequence: artifact-only PLAN, exactly one all-contract behavior-level test/fixture RED
  with frozen production blobs, one architectural GREEN/refactor, then authorized evidence.
- Batch: stable receipt/observation split; canonical readiness topology; cancellable bounded and
  redacted ports; sensitive text; current CI policy source; complete pseudo-ref rejection;
  post-mutation CAS progression; descriptor-first dense bounds; tuple-safe identities; all Cycle 3
  invariants and #479-facing typed ports.
- Exclusions: #479/controller, parent artifacts, Go/connectors/certification/runtime/`make`,
  network/live GitHub, reviewer, integration, or merge.

## Cycle 5 consolidated-review prompt

- Inputs: `/tmp/478-REVIEW-CYCLE4-1.md` and `/tmp/478-REVIEW-CYCLE4-2.md`, read completely.
- Frozen candidate/base: `ca6f6873d168db707bbe58291b5ee1b582e9404f` /
  `3addb1f48be1afe8b1e2b59b54247679d7293805`.
- Strict sequence: artifact-only PLAN/finding matrix; one comprehensive test/fixture-only RED with
  retained 68 and frozen production proof; one coherent GREEN/refactor; authorized local evidence.
- Batch: exact broker records; complete policy-set refresh and authoritative initial bundle source;
  re-evaluated controller receipt authorization; centralized child eligibility; CAS-conditioned
  stable mutation identity; cookie/session grammar; caller-linked lifecycle with live-call keyed
  ownership and bounded stop/join; byte-bounded raw decoder plus schema-directed object envelopes;
  atomic Cycle 5 run state; complete Cycle 4 retention.
- Exclusions: push/network/GitHub, #479 implementation, parent artifacts, Go/connectors/
  certification/runtime/`make`, reviewer, integration, or merge.

## Cycle 6 consolidated-review prompt

- Inputs: `/tmp/478-REVIEW-CYCLE5-1.md` and `/tmp/478-REVIEW-CYCLE5-2.md`, read completely.
- Frozen candidate/base: `63ac436fdac5fc46be7004f8109c4f068aa5749c` /
  `3addb1f48be1afe8b1e2b59b54247679d7293805`.
- Strict sequence: artifact-only PLAN/scope/finding matrix; one comprehensive test/fixture-only RED
  retaining 109 and all frozen production blobs; one coherent GREEN/refactor; authorized evidence.
- Batch: real `GitHubDecisionBroker` adapter/composition; hostile-safe broker provenance;
  conditional parent-ready authorization plus rollback; intrinsic signal ownership; ordered review
  attempts with stable semantic authority; intrinsic raw/proxy totality; one complete credential
  grammar; receipt chronology; non-self-referential exact-current RUN-STATE; Cycle 5 retention.
- Scope after the completed broker contract map: prior exact 17 paths plus
  `.pi/extensions/shepherd/github-decision-broker.ts`, its test, `human-decision.ts`, and its test
  (21 total); any further path requires stop-and-replan. The broker exposes a repository-owned
  `readRecord`, and the orchestrator adapter proves the actual request/full-record,
  poll/compact-result, consume/evidence composition without a second repository or invented fields.
  The bounded read-only explorer has completed; implementation remains the local critical path.
- Exclusions: push/network/live GitHub, #479, Go/connectors/certification/runtime/`make`, reviewer,
  integration, or merge.

## Cycle 7 consolidated-review prompt

- Inputs: `/tmp/478-REVIEW-CYCLE6-1.md` and `/tmp/478-REVIEW-CYCLE6-2.md`, both read completely.
- Exact reviewed candidate/base: `dbce5b7d0c698bc802594211072fed77eff23c1c` /
  `3addb1f48be1afe8b1e2b59b54247679d7293805`.
- Strict sequence: artifact-only PLAN with 46-row matrix; one comprehensive test/fixture-only RED
  retaining Cycle 6 and frozen production; one coherent GREEN/refactor; exact-head evidence.
- Architecture: public prepare/commit split lets #479 journal prepared intent plus consumed decision
  before effect and settlement afterward; one production durable authority boundary atomically
  compare-consumes stable policy/review/path/receipt/ancestry/decision/plan/head/PR-CAS authority,
  while a separate freshness envelope carries observation metadata. The same boundary durably
  quarantines uncertain effects and idempotently rolls them back before key/join release.
- Other batch invariants: authoritative attested review-attempt history, owned-clock broker skew,
  finite Kubernetes/Docker/AWS secret grammar, one HEAD current run-state semantic, true port-only
  #479-shaped wiring, and complete Cycle 6 retention.
- Exact scope remains the existing 21 owned paths; no new artifact is necessary. No Go, connector,
  parent/main worktree, dependency, push/network/live GitHub, #479 implementation, reviewer,
  integration, or merge action is authorized.

## Cycle 7 correction outcome

- Checkpoints: PLAN `2c64979829048d3de0d1ff1575c2a4f43cb699ba`; test-only RED
  `10033bc532d06967ce960e408c2bc9725020478a`; architectural GREEN
  `5bab0bc7e56292171eb28618cc2f37488ed1b7a4`; REFACTOR proof
  `87e704010f3e2226d8393d12e1a1bdf72df212a0`; audit RED
  `b1560e76a3abbac5efcd33b2740b7275b6acc137`; audit GREEN
  `915882c219f52da2c1edebce84d2bf90c61a4592`; current evidence candidate `HEAD`.
- Both input reports were replayed after REFACTOR. All consolidated families and the #479 public
  prepare/journal/commit seam have named passing tests, including exact 500 ms late effects after a
  100 ms timeout, cancellation, durable keyed quarantine, rollback retry, harmless-refresh stable
  identity, semantic movement, full authoritative attempt provenance, future chronology, and
  production-port-only wiring. The required authority has no legacy transport fallback, and the
  #479 proof uses separate production-typed transport, authority, and journal roles.
- Focused result: 297 total, 296 pass, 0 fail, 1 intentional live skip. Broad serialized result:
  environmental failure, 517 total, 451 pass, 65 unrelated managed-sandbox `spawn EPERM`
  failures, 1 skip. Strict TypeScript, pinned offline RPC, base/ancestry/diff/exact-scope/JSON, and
  explicit synthetic-marker gates pass.

## Cycle 8 consolidated-review prompt

- Inputs: `/tmp/478-REVIEW-CYCLE7-1.md` and `/tmp/478-REVIEW-CYCLE7-2.md`, both read completely.
- Exact frozen candidate/base: `b90037df1fff38c755ebc8025579120d17031330` /
  `3addb1f48be1afe8b1e2b59b54247679d7293805`; existing exact scope remains 21 paths.
- Sequence: artifact-only PLAN with 48-row seven-family matrix; one comprehensive five-test-file
  RED with frozen production; one coherent GREEN and bounded REFACTOR; complete local gates; exact
  evidence. Do not freeze, review, or hand off a partial family subset.
- Credential contract: classify all recognized assignment suffixes under arbitrary provider
  prefixes, then apply only narrow exact safe-name exceptions; retain kube/docker/AWS forms and
  table every suffix through every durable/outbound text consumer without marker reflection.
- Recovery contract: every uncertain non-value result starts quarantine. Rollback requests carry a
  stable recovery identity, the original ready mutation key/intent, and an ordered attempt fence;
  resource revisions are never guessed. Each durable authority attempt fences its
  predecessor and can only restore the exact draft. Controller response waits are deadline-aborted;
  superseded results never settle state; quarantine releases only on the matching fenced durable
  draft result. Cross-instance truth lives in shared durable authority/journal backing, never a
  module `WeakMap` or adapter identity.
- #479 contract: separate production-typed transport, authority, and journal roles; exact ancestry,
  compare, and durable mutation return types; success, conflict, uncertainty, rollback,
  incomplete/joined stop, reconstruction, and settlement; no `any`, casts, fake projection,
  private shortcuts, or same-object restart.
- Broker/freshness contract: actual broker rereads existing exact state before new-request expiry
  validation; pre-expiry consumed evidence can resume after expiry while new expired events fail.
  Commit sends refreshed freshness but original stable authorization/key/intent.
- GSD: doctor passes; adapter lacks `programming-loop`, so `manual_gsd_fallback`. One read-only
  explorer maps contracts; worker owns local critical path. Exclude Go/connectors/`make`, parent or
  main worktrees, #475, dependencies, runtime services, network/GitHub, push, reviewers,
  integration, and merge.
- Outcome: PLAN `bccee8e6`, test-only RED `851bb3bf`, GREEN `013bdc8b`, and REFACTOR `26a7d476`
  complete the prompt without scope expansion. Targeted Cycle 8 is 46/46; focused is 374 total /
  373 pass / 0 fail / 1 skip; strict owned/all-production and offline RPC pass. Serialized Shepherd
  is environmentally failing only in the unchanged 65 sandbox `spawn EPERM` cases (594 total /
  528 pass / 1 skip). Both input reports were re-read after REFACTOR; parent owns fresh exact-head
  review and every external action.

## Cycle 9 consolidated-review prompt

- Inputs: `/tmp/478-REVIEW-CYCLE8-1.md` and `/tmp/478-REVIEW-CYCLE8-2.md`, both read completely;
  frozen candidate `f97a698df90010ae072554e04563a8134a8e5f6e`, exact base `3addb1f4`, exact 21 paths.
- Execute one indivisible 69-row PLAN -> five-file RED -> coherent GREEN -> bounded REFACTOR ->
  evidence correction. Preserve all 374 focused cases and frozen production before RED.
- Replace local quarantine truth with a canonically queryable authority record carrying invocation
  and recovery IDs, exact target/revision/head, stable mutation identity, phase/status, and monotonic
  fence. Only explicit `ready_settled` may report ready; `recovery_claimed` fences the original
  writer before rollback; only exact `draft_restored` releases key/stop ownership.
- Treat immediate applied-then-rejected and every uncertain `ExternalPortError` uniformly as
  blocked/quarantined, even under healthy visible-ready reads. Consult authority before every
  prepare/commit/reconcile ready shortcut and reconstruct at the visible-ready/pre-rollback point.
- Parse the complete uppercase shell assignment name, including leading underscore, to the
  delimiter. Exercise leading underscore, 127/128/129/256, largest accepted field, over-field, and
  exact `FEATURE_TOKEN` control through all five consumers without reflecting markers.
- Rebuild the #479 proof with a public typed decision broker and `JSON.parse` assigned to `unknown`;
  canonically validate decision/prepared/journal/authority/recovery/fence/mutation/settlement
  snapshots before constructing new roles. No `any`, casts, fake projection, private shortcut, or
  object identity.
- Adapter remains `manual_gsd_fallback`; all slots are occupied so execution is
  `local_critical_path`. Exclude Go/connectors/`make`, runtime/services, dependencies, parent/main/
  #475, network/GitHub, push, reviewers, integration, merge, and human-gate actions.
- Outcome: PLAN `7ad23ed4`, RED `9278e97e`, and GREEN `593ba1cf` complete the prompt. The focused
  five-file route is 450 total / 449 pass / 0 fail / 1 intentional skip; strict owned and
  all-production TypeScript plus pinned offline RPC pass. Serialized Shepherd is 670 total / 604
  pass / 65 unchanged managed-sandbox `spawn EPERM` failures / 1 skip. A separate production
  refactor was not needed; final artifact edits record evidence only. Parent retains fresh-review
  ownership.
