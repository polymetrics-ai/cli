## Worker Handoff

Sub-issue: #478

Parent issue: #471

Worker agent: Codex `gpt-5.6-sol` / high

Branch: `feat/478-shepherd-github-parent-orchestration`

Sub-PR: https://github.com/polymetrics-ai/cli/pull/487

Parent PR: #472

Base branch: `feat/471-pi-agent-session-shepherd`

Worker directory: `/tmp/shepherd-478-correction`

Prior frozen reviewed candidate: `baef761544b8f0f58e2662058ae0c1715f345300`

Cycle 13 status: local GREEN `e0101044bb68f8a6b4cf45960029aac8d8b1ff78` follows PLAN
`27e7b5d2` and RED `a36188f4`. Focused is 1061 / 1060 / 0 fail / 1 intentional skip; targeted Cycle
13 passes 68/68 twice. Strict owned/all-production TypeScript, pinned offline RPC, exact
base/scope/integrity, and report-replay gates pass. Broad is environmental 1281/1215/65/1 and no
fresh exact-head review ran, so verification and review coverage remain false. This is local
evidence, not a frozen or review-clean candidate.

## Scope Delivered

- Typed, bounded, fakeable GitHub orchestration transport for parent objectives, child issues,
  stacked PRs, rosters, and exact child integration receipts, plus separate mandatory durable
  authority and journal ports for parent-ready transitions.
- Exact-shape authoritative checks, requested changes, review threads, review findings,
  dispositions, and exact-range independent Codex review evidence.
- Reuse of dependency graph scheduling, autonomy reconciliation, workspace handoff evidence, and
  the existing request/poll/consume decision broker without controller/session wiring.
- Retry-safe marker reconciliation, timeout-after-publish recovery, merged-child restart reuse,
  exact generation/head binding, and fail-closed parent human gating without merge capability.

## Files Changed

- `.pi/extensions/shepherd/github-orchestrator.ts` and matching test: orchestration domain/port.
- `.pi/extensions/shepherd/github-evidence.ts` and matching test: authoritative evidence policy.
- `.pi/extensions/shepherd/review-router.ts` and matching test: declarative independent review work.
- `.pi/extensions/shepherd/github-decision-broker.ts` and matching test: broker-owned canonical
  record rereads and actual production composition.
- `.pi/extensions/shepherd/human-decision.ts` and matching test: closed decision DTO chronology and
  the shared sensitive-text grammar.
- `.pi/extensions/shepherd/fixtures/issue-478/**`: bounded fake evidence/objective fixtures.
- `.planning/phases/478-shepherd-github-parent-orchestration/**`: plan, TDD, verification, and handoff.

## GSD / TDD / Skill Evidence

- GSD mode: `manual_gsd_fallback`.
- GSD command: `scripts/gsd prompt programming-loop init --phase
  478-shepherd-github-parent-orchestration --dry-run` returned `unknown GSD command:
  programming-loop`; `scripts/gsd doctor` passed.
- GSD adapter source: `.agents/agentic-delivery/references/gsd-pi-adapter.md`.
- Required skills source: `.agents/agentic-delivery/references/required-skills-routing.md`.
- Required Go skills loaded: not applicable; this is a TypeScript-only bounded slice.
- Required design skills loaded: not applicable.
- Skills loaded: `gsd-programming-loop`, `github-issue-first-delivery`, `gsd-workstreams`,
  `architecture-patterns`, `javascript-testing-patterns`.
- Red test evidence: initial absent-module RED was 0 pass / 3 file failures; adversarial correction
  RED at `db9fbc33` was 17 pass / 10 expected failures with production unchanged at `90321ffb`.
- Green implementation evidence: focused 27/27 and strict owned TypeScript pass at `40ce66d4`.
- Refactor evidence: serialized Shepherd, strict production, offline RPC, and diff/base/scope pass.

## CLI Help / Docs / Website Parity

- Applies: no; no CLI command, flag, help, docs, website, or generated manual surface changed.
- Runtime help checked: not applicable.
- Bare namespace behavior checked: not applicable.
- `docs/cli/**` updated: not applicable.
- `website/**` updated: not applicable.
- Generated help/manual artifacts updated: not applicable.
- Parity exemptions: internal Pi extension orchestration boundary only.

## Verification

```bash
node --test .pi/extensions/shepherd/github-orchestrator.test.ts \
  .pi/extensions/shepherd/review-router.test.ts \
  .pi/extensions/shepherd/github-evidence.test.ts \
  .pi/extensions/shepherd/github-decision-broker.test.ts \
  .pi/extensions/shepherd/human-decision.test.ts
node --test --test-concurrency=1 .pi/extensions/shepherd/*.test.ts
# strict TypeScript 5.9.3 with cached Pi 0.80.6 package/types
# pinned Pi 0.80.6 offline RPC get_commands discovery
git diff --check 3addb1f48be1afe8b1e2b59b54247679d7293805..HEAD
```

Historical Cycle 7 result: focused 297 total / 296 pass / 0 fail / 1 intentional live-GitHub skip;
serialized Shepherd 517 total / 451 pass / 65 unchanged unrelated sandbox `spawn EPERM` failures /
1 intentional skip; strict owned and all-production TypeScript pass; pinned offline RPC returns
`true`; immutable-base, exact-21-path scope, diff, JSON, and credential scans pass.

## Automated Review

- Primary route: `codex_independent` using `openai-codex/gpt-5.6-sol:xhigh` (parent policy).
- Fallback route: none.
- Coverage route: parent-owned stable-head specialist campaign.
- Coverage status: pending.
- Review URL: pending.
- Disposition summary: no review was started by this worker.
- Unresolved findings: all accepted Cycle 11 findings have local executable coverage and GREEN;
  exact-head independent validation remains required after this terminal evidence commit.

## Merge Recommendation

- Recommended state: `local_green_broad_nonzero_awaiting_exact_head_review`.
- Reason: Cycle 12 focused/strict/offline/integrity gates pass, but the declared broad route exits
  non-zero with 65 managed-sandbox process-spawn/lease failures and both frozen-candidate reviews
  remain historical blockers until fresh exact-head review covers current `HEAD`.
- Human gates: parent ready, exact-head parent merge, and default-branch merge remain active.
- Follow-up issues: #479 owns controller/session integration; this worker must not add it.

## Superseding correction note

The deep functional review of `093b3c90` found eleven accepted issues, so the earlier merge
recommendation and local verification are historical. The correction is now executing as a strict
artifact-only plan checkpoint, one behavior-level test-only RED commit, coherent GREEN, and fresh
authorized verification. Controller wiring remains #479-owned; #478 may add only the scoped
session-attestation contract and fixtures required to verify independent-review provenance.

## Correction handoff status

- Plan: `5dd7897e`; test-only RED: `4e02d059`; coherent GREEN: `8e32896a`.
- Focused: 38/38 pass. Strict owned and all-production TypeScript 5.9.3 against pinned Pi 0.80.6:
  pass. Offline pinned Pi RPC: `true`. Base/head/diff/scope: pass.
- Serialized Shepherd: 302 total, 236 pass, 65 fail, 1 intentional skip. All correction tests pass;
  every failure is outside #478 scope and reports sandbox `spawn EPERM` from the process-identity
  child-process probe. `verificationPassed` remains false.
- Push/PR #487 update: blocked by `ssh: Could not resolve hostname github.com: -65563`; local
  commits retained. No live reviewer, GitHub mutation, controller/#479 edit, or merge ran.

## Cycle 3 handoff in progress

- Frozen candidate: `3f285722a505ea426d53a34f95716781d1aca7c2`; exact base:
  `3addb1f48be1afe8b1e2b59b54247679d7293805`.
- Both corrected review ledgers are accepted as one fourteen-invariant correction batch.
- Strict sequence: artifact-only plan, exactly one test/fixture-only RED with all production blobs
  unchanged, one architectural GREEN/refactor, then authorized verification evidence.
- New contracts cover canonical persisted plan provenance, mutating-only child topology, complete
  evidence/receipt identity and chronology, durable conditional mutations, exact ancestry, CI-policy
  provenance, monotonic roster state, and the exported controller attestation protocol.
- Parent still owns #479 wiring, parent planning, exact-head independent review, integration, and
  human merge gates. Existing GitHub DNS failure defers push/PR synchronization only.

## Cycle 3 completed local handoff

- Checkpoints: plan `d97faf44`, policy correction `d2c7f374`, RED `faf2e8f8`, GREEN `41e8e76e`.
- Focused #478: 53/53 pass. Strict TypeScript, pinned Pi 0.80.6 offline RPC, immutable
  base/ancestry, full-range diff, 17-path scope, and credential scan pass.
- Serialized Shepherd: 317 total, 251 pass, 65 unrelated sandbox `spawn EPERM` failures, and
  1 intentional skip; every #478 test passes. No prohibited or live action ran.

## Cycle 4 correction in progress

- Frozen candidate `d3b6b5e2`; immutable base `3addb1f4`; initial worktree clean.
- Both final ledgers were read completely and consolidated into ten contracts before production
  changes. Required `gsd-programming-loop`/routing/runtime references were loaded.
- `scripts/gsd doctor` passes; the adapter lacks `programming-loop`, so
  `manual_gsd_fallback` is active. Thread capacity forces `local_critical_path`.
- Implementation route is `openai-codex/gpt-5.6-sol:high`. Parent owns #479, planning, two fresh
  exact-head reviews, publication, integration, and human merge gates.

## Cycle 4 completed local handoff

- Checkpoints: PLAN `607e203ef1f76ff112c130ccff5d155973d984f6`; single RED
  `abbf388b8a852836e0dd10a55b9f17720b9fde22`; GREEN
  `b92b5ff7dd3738dc3b3350ebb4d2f2b42074f954`; evidence hash is reported by the worker after commit.
- Focused #478: 68/68 pass. Strict TypeScript passes for owned production/tests and all 20
  production modules against pinned Pi 0.80.6. Offline RPC discovers `pm-shepherd` from
  `extension`.
- Serialized Shepherd: 332 total, 266 pass, 65 unrelated managed-sandbox `spawn EPERM` failures,
  1 intentional skip; every #478 test passes. This environmental broad-gate result is not claimed
  as a full-suite pass.
- Immutable base/frozen-candidate ancestry, full-range `git diff --check`, 17-path owned scope, and
  credential-literal scan pass. No Go, connector, certification, runtime, `make`, network, live
  GitHub, controller/#479, reviewer, Claude/Copilot, or merge action ran.
- No unresolved local design blocker remains. Parent next runs two fresh independent exact-head
  `openai-codex/gpt-5.6-sol:xhigh` reviews, then owns disposition, publication, integration, and
  human gates.

## Cycle 5 correction in progress

- Frozen candidate `ca6f6873`; immutable base `3addb1f4`; initial worktree clean.
- Both Cycle 4 review ledgers were read completely and consolidated before test/production edits.
- Envelope solution: byte-bound raw JSON before parse plus shared schema-directed capped record
  reads; no bulk descriptor expansion of untrusted objects.
- Lifecycle solution: caller-linked abort/deadline, tracked port settlement/acknowledgement,
  `AsyncLocalStorage` ensure scopes retaining keyed ownership while live, and bounded stop/join that
  reports uncooperative work incomplete.
- Adapter doctor passes; missing command/helper activates `manual_gsd_fallback`. One read-only
  explorer maps coupled symbols; isolated worker owns implementation. No push/network/GitHub.

## Cycle 5 completed local handoff

- Checkpoints: PLAN `7cf9c88ddadee395020444c19ee9f001b0807a53`; comprehensive test-only
  RED `6cb21902244e4bccf390c4e7556eb615e5e1697f`; architectural GREEN
  `3ae10dc2303409230153e32e6b6231b27b18cdcf`; the evidence commit is reported by the worker
  after commit because it cannot contain its own identity.
- Focused #478: 109/109 pass. Strict TypeScript 5.9.3 passes for the six owned production/test
  files and all 20 Shepherd production modules against pinned Pi 0.80.6 declarations. Pinned
  offline Pi RPC discovers `pm-shepherd` from `extension`.
- Serialized Shepherd: 373 total, 307 pass, 65 unrelated managed-sandbox process-identity
  `spawn EPERM` failures, and 1 intentional live-GitHub skip. Every #478 test passes; this
  environmental result is not represented as a full-suite pass.
- Immutable base and frozen candidate ancestry, exact merge base, full-range `git diff --check`,
  exact 17-path ownership, JSON parsing, and high-confidence credential-literal scans pass.
- GREEN changed only support fixtures in `github-orchestrator.test.ts` after RED: fake broker
  records became fully canonical, seeded receipts gained authoritative provenance, lifecycle
  recovery observes the created resource, and RUN-STATE checks the exact RED checkpoint. No RED
  expectation was removed or weakened.
- No Go, connector, certification, runtime, `make`, network, live GitHub, controller/#479,
  reviewer, integration, or merge action ran. Parent next runs two fresh exact-head
  `openai-codex/gpt-5.6-sol:xhigh` reviews and owns disposition, publication, integration, and all
  human gates.

## Cycle 6 correction in progress

- Frozen candidate `63ac436f`; immutable base `3addb1f4`; initial worktree clean.
- Both Cycle 5 review ledgers were read completely and consolidated before test/production edits.
- The initial 18-path plan was amended before RED after the completed broker contract map proved the
  minimum honest boundary is 21 paths: add `github-decision-broker.ts` plus its test and
  `human-decision.ts` plus its test to the prior 17.
- `GitHubDecisionBroker.readRecord` must use its own repository. The orchestrator adapter consumes
  the real request/full-record, poll/compact-result, and consume/evidence shapes, then verifies each
  against that broker-owned canonical reread; a second repository or reconstructed record is
  forbidden. Native broker and orchestrator composition tests prove the boundary.
- `human-decision.ts` imports the authoritative credential assertion from `review-router.ts` and
  owns descriptor-safe closed-record chronology; native tests exercise both contracts.
- Adapter command unavailable after healthy doctor, so `manual_gsd_fallback` remains recorded.
  The bounded read-only explorer completed its exact symbol map; isolated worker owns implementation. No push,
  network, live GitHub, Go, connector, runtime, `make`, reviewer, integration, or merge action.

## Cycle 6 completed local handoff

- Checkpoints: amended PLAN `2832993b93d07ea20197bad52ec23700fe21fc1e`; comprehensive
  five-test-file RED `ca4d97d1100b1b44176da9d7dfd6ee6f56f4e1e6`; architectural GREEN
  `2c6371e725d58b2dc05902d68f9e6812904664d6`; the evidence commit is reported after commit
  because RUN-STATE deliberately uses the non-circular `HEAD` reference.
- Actual `GitHubDecisionBroker` composition passes pending -> decided -> consumed using its own
  repository-backed `readRecord`; no second repository, stronger fake API, or synthesized evidence
  is involved.
- Focused five-file route: 207 total, 206 pass, 0 fail, 1 intentional live-GitHub skip. Retained
  Cycle 5 is 109/109. Strict TypeScript 5.9.3 passes for the five owned production/test pairs and
  all 20 Shepherd production modules. Pinned Pi 0.80.6 offline RPC discovers `pm-shepherd` from
  `extension`.
- Serialized Shepherd: 427 total, 361 pass, 65 unchanged unrelated managed-sandbox
  process-identity `spawn EPERM` failures, 1 intentional skip. Every Cycle 6 focused assertion
  passes; the environmental result is not represented as a full-suite pass.
- Immutable base and frozen candidate ancestry, exact merge base, full-range `git diff --check`,
  exact 21-path ownership, JSON parsing, and high-confidence credential-literal scans pass.
- Post-RED test edits are support-only: one fixture description stopped triggering the credential
  grammar; direct human-decision setup now persists its request comment; orchestrator fakes now
  match canonical broker/authorization/receipt/provenance shapes and the negative test explicitly
  removes request-comment evidence. No assertion was removed or weakened.
- No Go, connector, certification, runtime, `make`, network, live GitHub, controller/#479,
  reviewer, integration, or merge action ran. Parent next publishes the exact evidence candidate,
  runs two fresh exact-head `openai-codex/gpt-5.6-sol:xhigh` reviews, and owns disposition,
  integration, and all human gates.

## Cycle 7 completed local handoff

- Checkpoints: PLAN `2c64979829048d3de0d1ff1575c2a4f43cb699ba`; comprehensive test-only
  RED `10033bc532d06967ce960e408c2bc9725020478a`; architectural GREEN
  `5bab0bc7e56292171eb28618cc2f37488ed1b7a4`; REFACTOR proof
  `87e704010f3e2226d8393d12e1a1bdf72df212a0`; architecture-audit RED
  `b1560e76a3abbac5efcd33b2740b7275b6acc137`; audit GREEN
  `915882c219f52da2c1edebce84d2bf90c61a4592`; evidence candidate is the deliberate
  non-circular `HEAD` and the exact evidence SHA is reported after commit.
- Focused five-file suite: 297 total, 296 pass, 0 fail, 1 intentional live-GitHub skip. Strict
  TypeScript passes for the five owned production/test pairs and all 20 production modules. Pinned
  Pi 0.80.6 offline RPC discovers `pm-shepherd` from `extension`.
- Serialized Shepherd: environmental failure, 517 total, 451 pass, 65 unchanged unrelated
  managed-sandbox process-identity `spawn EPERM` failures, 1 intentional skip. This is not reported
  as a broad pass; all issue #478 Cycle 7 assertions pass.
- Both Cycle 6 reports were replayed after REFACTOR. Passing named cases cover ten atomic movement
  coordinates, exact 500/100 before/after timing, cancellation, keyed/durable restart quarantine,
  read failure, rollback retry, stable harmless refresh versus semantic movement, authoritative
  full-attempt provenance, future broker chronology, all finite credential schemas, current-HEAD
  state, and the public production-port-only #479 prepare/journal/commit seam. The public transport
  has no ready mutation, authority is mandatory, compare conflicts are typed, and the proof uses
  separate production-typed transport, authority, and journal roles rather than the test fake.
- Immutable base and reviewed candidate are ancestors; exact merge base, full diff check, exact
  21-path ownership, JSON parsing, and explicit test-synthetic marker classification pass.
- No Go, connector, certification, runtime, `make`, dependency, network, push, live GitHub,
  controller/#479 implementation, reviewer/self-review, integration, or merge ran. Parent owns
  publication, two fresh exact-head reviews, dispositions, integration, and human gates.

## Cycle 8 plan handoff (historical checkpoint)

- Frozen candidate `b90037df1fff38c755ebc8025579120d17031330`; immutable base
  `3addb1f48be1afe8b1e2b59b54247679d7293805`; exact 21 paths; clean start.
- Both Cycle 7 reports were read fully. Seven-family union is one 48-row correction: credential
  suffix closure, strict #479 typing/recovery, all uncertain non-values, real-broker expiry resume,
  bounded fenced rollback, reconstructed durable restart, and refreshed freshness.
- Durable recovery ownership: stable recovery ID plus ordered attempt fence; authority durably
  supersedes predecessor before returning, identifies the exact original ready mutation without
  revision guessing, and only restores draft; controller aborts bounded
  response waits, ignores stale results, blocks reentry, and holds key/stop ownership until the
  matching fenced durable draft result. Cross-instance state lives in shared durable backing, not
  `WeakMap`/adapter identity.
- GSD doctor passes; missing adapter command records `manual_gsd_fallback`. Required skills and
  contracts loaded. One read-only explorer maps the contracts; worker owns the local critical path.
- Planned checkpoints, completed below: artifact-only PLAN commit; complete five-test-file RED with
  all five production blobs frozen; coherent GREEN/refactor; focused/strict/offline/scope/data/
  serialized classification; exact evidence. Parent owns publication, fresh reviews, integration,
  and gates.

## Cycle 8 completed local handoff

- Checkpoints: PLAN `bccee8e6cdbcb6e38419114f264222b1f5616f66`; comprehensive five-test-file
  RED `851bb3bfa3e23042211a8b37f3a97253cc6fedf5`; coherent GREEN
  `013bdc8b264e1ce8808d4af2558e2ec40b85ee49`; bounded REFACTOR
  `26a7d476bdfaa4e263196fb76f7f43b5a3ad799e`; evidence is current `HEAD` and its exact SHA is
  reported after commit.
- RED records 374 total / 314 pass / 59 intended failures / 1 skip and only four intended strict
  diagnostics, with all five production blobs frozen. Targeted Cycle 8 is 46/46 after REFACTOR.
- Focused five-file route: 374 total / 373 pass / 0 fail / 1 intentional live-sandbox skip. Strict
  TypeScript passes for all five production/test pairs and all 20 production modules. Pinned Pi
  0.80.6 offline RPC discovers `pm-shepherd` from the explicit extension.
- Serialized Shepherd is environmental failure: 594 total / 528 pass / 65 unchanged unrelated
  managed-sandbox `spawn EPERM` failures / 1 intentional skip. Every Cycle 8/focused assertion
  passes; no broad-pass claim is made.
- Both Cycle 7 reports were re-read completely after REFACTOR. Exact provider-neutral suffixes,
  real-broker expiry replay, uncertain immediate rejection, fenced deadlines/supersession,
  refreshed freshness, exact typed #479 recovery, and serialized cross-instance role
  reconstruction all have named passing evidence.
- Immutable base/reviewed-candidate ancestry, exact merge base, full-range diff check, exact
  21-path ownership, three JSON parses, synthetic-marker confinement, and clean pre-evidence status
  pass. No Go, connector, `make`, dependency, parent/main/#475, network/GitHub, push, reviewer,
  integration, or merge action ran. Parent owns publication, two fresh exact-head reviews,
  dispositions, integration, and all human gates.

## Cycle 9 completion handoff

- Frozen reviewed candidate `f97a698df90010ae072554e04563a8134a8e5f6e`; exact base `3addb1f4`;
  clean exact 21-path scope; both Cycle 8 reports read completely; all five production blobs frozen.
- One indivisible 69-row correction: uncertain-result consistency 8, durable dangerous-point
  restart/original-writer fence 13, total assignment parsing 40, exact typed serialized #479 proof
  8. Prior focused 374 cases and the intentional live skip remain mandatory.
- Authority-owned record states are `ready_invoking` -> `ready_effect_applied` -> explicit
  `ready_settled`, or monotonic `recovery_claimed` -> exact `draft_restored`. Recovery claims fence
  the original writer and older rollback attempts before effects. Unsettled authority always blocks
  ready reuse and retains key/stop ownership across reconstructed controllers.
- Complete shell-style uppercase names, including leading underscore, are parsed to their delimiter
  before suffix classification. Exact boundary/control rows traverse each of the five shared
  durable/outbound consumers without marker reflection.
- The #479 fixture must use public typed roles and canonical unknown decoders for all serialized
  decision/prepared/journal/authority/recovery/fence/mutation/settlement state; no cast, `any`, fake
  projection, private shortcut, or identity-dependent restart.
- Doctor passes; missing programming-loop command records `manual_gsd_fallback`. All runtime slots
  are occupied, so the isolated worker owns `local_critical_path`. Parent owns publication/reviews/
  integration/human gates; every prohibited or external action remains excluded.
- Checkpoints: artifact PLAN `7ad23ed4`, test-only RED `9278e97e`, coherent GREEN `593ba1cf`.
  Focused five-file evidence is 450 total / 449 pass / 0 fail / 1 intentional skip; strict owned
  and all-production TypeScript and pinned Pi 0.80.6 offline RPC pass. Serialized Shepherd is 670
  total / 604 pass / 65 unchanged managed-sandbox process-identity `spawn EPERM` failures / 1 skip.
  Immutable base, exact merge base, full-range diff, exact 21 paths, three JSON parses, Cycle 9
  marker confinement, and both complete Cycle 8 report replays pass. No post-GREEN production edit
  or standalone refactor ran.

## Cycle 10 plan handoff

- Frozen candidate/tree/base: `a49e4df2798281d1e64c722ccbcab5f4a678c3e1` /
  `9167ebaf82f92c1229e56b1b8334262a356dcd3c` / `3addb1f4`; clean exact 21 paths; both Cycle 9
  reports read completely; all five production blobs frozen.
- One transition table owns every ordering, begin, effect, settlement, recovery claim, rollback,
  confirmation, restart, and stale-writer interruption. Durable begin precedes effect; exact applied
  revision is mandatory; settlement-wins and recovery-wins are exclusive terminal CAS outcomes.
- One planned executable RED matrix covers authority-first broken gates, settlement races, revision
  mismatches, bounded confirmation, pre-application rejection, `+=`/case/index assignments across
  five consumers, warning disposition chronology, and canonical snapshot consistency.
- `verificationPassed` is false while the complete broad route exits non-zero. Doctor passes;
  unavailable command records `manual_gsd_fallback`; agent-capacity rejection records
  `local_critical_path`. Parent retains publication/review/integration/human-gate ownership.

## Cycle 10 completion handoff

- Checkpoints: PLAN `470a8a85`, RED `2256971a`, GREEN `5f46206e`, refactor `8946b67b`.
- Focused five-file route is 687 total / 686 pass / 0 fail / 1 intentional skip; authority target
  is 89/89; strict owned and all-20-production TypeScript pass.
- Serialized Shepherd is 907 total / 841 pass / 65 unchanged managed-sandbox process/lease
  failures / 1 skip. This non-zero route keeps `verificationPassed` false.
- Exact base/merge-base, full-range diff, exact 21 paths, and all three JSON parses pass. No network,
  push, GitHub, reviewer, integration, merge, parent-ready, or other human-gate action ran.

## Cycle 11 correction handoff

- Frozen candidate/tree/base: `3b39cfce9b4a99940b0451302df6bf5c17b49c02` /
  `962160e1ccae2e52f6f645185edb96819bd4a9f5` / `3addb1f4`; clean exact 21 paths; both Cycle 10
  reports read completely; all five production blobs frozen.
- Durable begin is uncertain. The original invocation, same-key lock, and stop token remain owned
  until that invocation settles and a later authority read proves absence or drives the exact
  invocation terminal. A late signal-ignoring write cannot escape after an early null read.
- Every typed non-applied conflict must carry exact atomic tombstone proof bound to the requested
  invocation. The compare boundary removes only that `ready_invoking` reservation and never
  rewrites a moved draft or foreign non-draft PR; the controller preserves the coordinate only
  after validating the terminal proof.
- The #479 decoder validates the cross-component history before constructing roles: settlement,
  authority phase, current PR visibility/revision/head, applied/rollback values, prepared decision,
  and unique causal mutation sequence must agree. Legitimate claim-before-effect and pruned-history
  windows remain explicit.
- C10-CONFIRM uses observable latches for confirmation entry, deadline/new fence entry, and release;
  isolated and focused runs repeat without retry wrappers, timing relaxation, or fixed-delay phase
  guesses. Escaped quote/space/continuation and command/parameter substitution tails are fully
  redacted for `=`/`+=`; five validators reject generically without marker reflection.
- `verificationPassed` and `reviewCoveragePassed` remain false. The broad route is non-zero and the
  fresh Cycle 11 head has not been reviewed. Parent retains publication, review, integration,
  ready, merge, and every human gate.
- RED evidence: 791 total / 743 pass / 47 intended TAP failures / 1 intentional skip. The exact 42
  failing leaves are BEGIN 6, SNAPSHOT 13, terminal-proof 10, persistent-tombstone 3, and direct
  redaction 10; five parent containers account for the remaining failures. All five production
  blobs remain exact. C10-CONFIRM passes five consecutive causally latched 5/5 runs.
- GREEN `e765e0d31c426ecf201162509519fa03d460d871` closes every named row. Focused is 791 total /
  790 pass / 0 fail / 1 intentional skip. All 60 assignment rows pass; begin is 6/6, typed conflict
  is 10/10, persistent tombstone is 3/3, and all 16 named snapshot rows plus their controls pass.
- C10-CONFIRM repeats five consecutive 5/5 runs at GREEN without retries, longer deadlines, phase
  sleeps, or weaker assertions. Strict TypeScript 5.9.3 passes over the five owned pairs and all 20
  production modules against pinned Pi 0.80.6 declarations.
- Pinned offline RPC returns `true` for `pm-shepherd`; exact base/reviewed-candidate ancestry and
  merge base, full-range diff, exact 21 paths, three JSON parses, Cycle 11 marker confinement, and
  both complete Cycle 10 report replays pass.
- Serialized Shepherd is an honest environmental failure: 1011 total / 945 pass / 65 unchanged
  managed-sandbox `spawn EPERM` failures / 1 skip. This keeps machine verification false. No
  standalone code refactor followed GREEN because the coherent change already isolates the
  tombstone and restart-history validators.
- No Go, connector, `make`, service, dependency, parent/main/#475, network/GitHub, push, reviewer,
  integration, ready, merge, or human-gate action ran.

## Cycle 12 correction handoff

- Frozen candidate/tree/base: `4f0e17df` / `4f9797b2` / `3addb1f4`; exact 21 paths, clean start,
  both Cycle 11 reports read completely, all five production blobs frozen.
- PLAN `7f96718c4d8c692cd618ff220ab0d53d2e6546a2` precedes executable RED
  `2649cf6dcad1fcc943261c0838bbca8e56c5c595`. Original RED is 942 / 885 / 56 / 1: 51 named
  leaves plus five parents with all production frozen. Reviewer-gap RED is 978 / 963 / 14 / 1:
  twelve strengthened leaves plus two parents; thirty new consumers already pass.
- Coherent GREEN `723fdc122cea75a5d6f146fb8b39383e9e5795e3` retains two exact-coordinate
  requested/foreign begin owners, reverse-consumes the total restart graph, validates every
  recovery-claimed window and global mutation sequence, and completely consumes all twelve
  multiline/composite assignment forms for both operators.
- Focused is 978 / 977 / 0 / 1. BEGIN 6/6, graph orphan 7/7, sequence 5/5, claim 15/15, and
  assignment 144/144 (24 direct + 120 consumer) pass.
- Strict five-pair and all-20-production TypeScript pass with TypeScript 5.9.3 / cached Pi 0.80.6.
  Pinned offline RPC returns `true` for `pm-shepherd` from `extension`.
- Serialized Shepherd is honestly environmental at 1198 total / 1132 pass / 65 managed-sandbox
  process-spawn/lease failures / 1 skip. This keeps `verificationPassed` false.
- Immutable base/reviewed-candidate ancestry, exact merge base, full diff, exact 21 paths, JSON3,
  and five-test Cycle 12 marker confinement pass. Both Cycle 11 reports total 399 lines with
  SHA-256 `f2aa1e4a89686c6ae1748252c994d18a602167c56f61f28583ff52162b0d5d27` and
  `d8e0fdfca0696f6446c0e85af43fd2471e8112a693688f053cccd547c1e430a1`.
- Final production blobs are orchestrator `ca07667f4e598fee472ae174b2a3c55bc708db55`, router
  `2c5fd80e4ee5ba536fb7f608ca4e424661a5431e`, broker `7be67851`, evidence `058ad162`, and
  human decision `fc1c6230`.
- `verificationPassed` and `reviewCoveragePassed` remain false; publication, review dispatch,
  integration, ready, merge, and every human gate remain parent-owned and did not run.

## Cycle 13 correction handoff

- Frozen candidate/tree/base: `baef7615` / `6bf70b7a` / `3addb1f4`; exact 21 paths and five
  production blobs; both Cycle 12 reports read completely (487 lines, exact hashes recorded).
- C13-BEGIN has 16 rows: absent, three unsettled phases, two terminal phases, and two stale-terminal
  rereads, each in both proof orders. Public return, key release, and stop completion await both the
  requested and exact returned-coordinate proof; requested coordinates are never synthesized.
- C13-CROSS-STORE has 10 rows: deterministic typed journal repair for two coherent terminal
  authority windows, idempotent replay, and six impossible-history rejections.
- C13-DECISION/MARKER has 13 rows: exact consumed `parent_merge`/`approve-merge` authorization
  binding plus one unique current canonical parent marker owner across histories.
- C13-SCANNER has 30 rows: malformed case/heredoc ownership, later public closer preservation,
  five generic consumers, and bounded single-pass/no-repeated-`lastIndexOf` behavior.
- PLAN `27e7b5d2`, RED `a36188f4`, and GREEN `e0101044` preserve the ordered TDD history. Focused
  GREEN is 1061/1060/0/1 and targeted Cycle 13 is 68/68 twice; all intended RED leaves close.
- Strict owned/all-20-production TypeScript and pinned Pi 0.80.6 offline RPC pass. Exact immutable
  base/reviewed-candidate ancestry, merge base, full diff, 21 paths, three JSON parses, Cycle 13
  marker confinement, and both Cycle 12 report identities pass.
- The forward scanner maintains O(1) composite/heredoc references and completes 60,000 unmatched
  openers in approximately 14-24 ms. Current production blobs are orchestrator `63e1f683`, router
  `8eb32b88`, broker `7be67851`, evidence `058ad162`, and human decision `fc1c6230`.
- Broad is truthfully classified 1281/1215/65/1 exit 1 in unchanged managed-sandbox process/lease
  tests; both machine gates remain false and fresh exact-head review is parent-owned.
- Post-RED expectation alignment only distinguishes explicit-stop cancellation, guarantees the
  divergent-head mutation, and reclassifies the invalid duplicate-marker Cycle 12 helper while
  retaining all twelve leaves and legal single-owner orphan/sequence controls.
- Adapter/runtime decisions are `manual_gsd_fallback` and `local_critical_path`. No network,
  GitHub, push, reviewer, integration, ready, merge, or human gate is worker-owned or ran.
