# TDD Ledger: #478

| Slice | RED evidence | GREEN evidence | Refactor/broad evidence | State |
| --- | --- | --- | --- | --- |
| bounded child planning | focused run fails with `ERR_MODULE_NOT_FOUND` for absent `github-orchestrator.ts` | bounded plan/DAG tests pass | strict owned TypeScript pass | green |
| reconcile-before-mutate idempotency | same test-only run; timeout/restart and collision contracts are present before production | timeout/restart/collision tests pass | strict owned TypeScript pass | green |
| stacked PR topology and linkage | same test-only run; parent `Closes` versus child `Refs` assertions are present | topology/linkage test passes | strict owned TypeScript pass | green |
| authoritative CI/thread/disposition evidence | focused run fails with `ERR_MODULE_NOT_FOUND` for absent `github-evidence.ts` | all evidence cases pass | strict owned TypeScript pass | green |
| exact-head independent Codex review | focused run fails with `ERR_MODULE_NOT_FOUND` for absent `review-router.ts` | all route/rejection cases pass | strict owned TypeScript pass | green |
| scoped provisional integration | same absent-orchestrator failure; handoff/scope/head recheck cases are present | capture/scope/re-read/integration cases pass | strict owned TypeScript pass | green |
| exact-generation/head parent human gate | same absent-orchestrator failure; incomplete/pending/reject/head-move cases are present | parent readiness/broker cases pass | strict owned TypeScript pass | green |

## Initial state

- Exact base: `3addb1f48be1afe8b1e2b59b54247679d7293805`.
- Production files and matching tests do not exist at plan time.
- RED will be committed with tests/fixtures only before any production file is added.
- The missing GSD adapter command is recorded as `manual_gsd_fallback`; strict test-first behavior
  remains mandatory.

## RED checkpoint

Command:

```bash
node --test .pi/extensions/shepherd/github-orchestrator.test.ts \
  .pi/extensions/shepherd/review-router.test.ts \
  .pi/extensions/shepherd/github-evidence.test.ts
```

Result: exit 1 with 0 pass / 3 test-file failures. Each matching test file fails deterministically
with `ERR_MODULE_NOT_FOUND` for its intentionally absent production module. The first attempted run
also exposed and corrected one test-only illegal nested `await`; the recorded RED run contains no
test syntax error. `scripts/tdd-gate.mjs` is not present, so the command output plus unchanged
production-file absence is the manual strict-TDD gate.

## Minimal GREEN checkpoint

- Focused command: 21 pass, 0 fail.
- `review-router`: 5 pass; declarative work records only.
- `github-evidence`: 6 pass; authoritative checks, requested changes, threads, dispositions, and
  exact-range review policy.
- `github-orchestrator`: 10 pass; bounded plans, reconcile-before-mutate publication, stacked
  topology, handoff capture, roster updates, exact-head child integration, and broker-gated parent
  readiness.
- Strict no-emit TypeScript over all three owned modules and matching tests passes using TypeScript
  5.9.3 with the cached Pi 0.80.6 Node type root.

## Adversarial correction RED

The same focused command against unchanged GREEN production reports 27 total: 17 pass and 10
expected failures. Failures cover proxy-trap execution, duplicate finding ambiguity,
`not_actionable` bypass, missing parent handoff capture, unbound receipt schema, plan/review binding,
merged-PR restart reuse, and the downstream parent completeness cases that now require bound
receipts. Production remains exactly at pushed GREEN checkpoint `90321ffb` for this test-only RED.

## Adversarial correction GREEN and refactor

- Correction implementation checkpoint:
  `40ce66d4b5010b92089895a05709687143d15a05`.
- Focused command: 27 pass, 0 fail in 230.914417 ms.
- Strict owned TypeScript: pass with TypeScript 5.9.3 and the cached Pi 0.80.6 Node type root.
- Receipts now bind child ID, PR, generation, stable marker, base SHA, exact head SHA, and parent
  branch; a merged PR with an exact receipt is reused safely after restart.
- The eligible review returned by authoritative evidence must also bind the planned repository,
  work item, generation, changed paths, and exact allowed scopes before integration/readiness.
- Parent handoff setup uses the upstream `captureHandoff` boundary. Transport arrays and caller
  DTOs are descriptor-validated without invoking Proxy/accessor code. Cross-review finding IDs are
  unique, and a blocking finding requires an exact-head `fixed` disposition.
- The fake-only PR number hint was removed from the production transport contract.

## Final authorized verification

- Focused #478: 27/27 pass.
- Complete serialized Shepherd: 291 total, 290 pass, 0 fail, 1 intentional sandbox skip;
  127120.23075 ms.
- Strict all-production TypeScript: 20 production modules pass with TypeScript 5.9.3 using the
  cached Pi 0.80.6 package resolver and Node type root.
- Pinned Pi 0.80.6 offline RPC `get_commands`: `true` for `pm-shepherd` from `extension`.
- Immutable merge base equals `3addb1f48be1afe8b1e2b59b54247679d7293805`; full-range
  `git diff --check` and coordinator-owned path validation pass.
- No Go, connector, certification, runtime-service, `make`, live orchestration transport,
  reviewer, or merge command ran.

## Stable-head functional review correction RED matrix

Reviewed production baseline: `093b3c90409cedc6b7008b7510f53937eb1ebbc1`.

| Finding | Behavior-level RED contract | State |
| --- | --- | --- |
| CR-01 | exact authoritative changed-path equality rejects empty/subset/superset and accepts reordered equality | green |
| CR-02 | authoritative integration lookup plus current-parent ancestry rejects forged, stale, mismatched, and orphaned receipts | green |
| CR-03 | capture, ensure, and integrate reject every forged immutable materialized-child topology field | green |
| CR-04 | required CI contexts and trusted producer IDs require a complete successful deterministic rollup | green |
| CR-05 | controller-owned session/run attestation and digest binding reject reviewer-self-attested execution metadata | green |
| CR-06 | expected review target is direct input and equal-generation selection is permutation-independent | green |
| CR-07 | keyed concurrent ensures create once, reconcile after create, and reject post-create ambiguity | green |
| WR-01 | zero and negative generations fail at plan, expected evidence, target, review, receipt, and attestation boundaries | green |
| WR-02 | canonical ref validation rejects spaces, `.lock`, leading dot, `@`, and other invalid Git ref forms | green |
| WR-03 | incomplete bounded lookup snapshots fail closed for issue, PR, roster, and integration reconciliation | green |
| WR-04 | timeout/malformed-response recovery covers PR create, integration, ready transition, lookup failure, and restart | green |

Checkpoint discipline: the next commit after the plan checkpoint is test/fixture-only. Before that
RED commit, `git diff 093b3c90 -- .pi/extensions/shepherd/*.ts` must show changes only in test files,
and production blob IDs must match the frozen reviewed head. GREEN evidence is intentionally blank
until the RED suite has failed for the new behavior.

## Stable-head correction RED and GREEN evidence

- Artifact-only checkpoint: `5dd7897e1a906fd16a88001cc5830a0db305c5ba`.
- Test-only RED checkpoint: `4e02d059050aa8fe6f9a60b519c61500b00d9f44`; 38 tests, 9 pass, 29
  expected fail. Before commit, all three production blob IDs exactly matched frozen reviewed head
  `093b3c90409cedc6b7008b7510f53937eb1ebbc1`.
- Coherent GREEN checkpoint: `8e32896aff5a0a04e47efc437aeb6bac1e0d3967`; focused 38/38 pass
  in 175.527833 ms, strict owned TypeScript 5.9.3 passes, and strict all-production Shepherd
  TypeScript passes against pinned Pi 0.80.6 types.
- Full serialized Shepherd command ran 302 tests: 236 pass, 65 fail, 1 intentional skip. Every
  #478 test passed; all 65 failures are outside the owned files and report the managed sandbox's
  `spawn EPERM` from the process-identity child-process probe. This is recorded as an environmental
  broad-gate failure, not a GREEN claim.
- Pinned Pi 0.80.6 offline RPC registration returned `true` for `pm-shepherd`; startup emitted only
  sandbox write warnings for the global settings lock.
- Push attempts for all three checkpoints failed with `ssh: Could not resolve hostname github.com:
  -65563`; commits remain local.

## Cycle 3 corrected-review RED matrix

Frozen production candidate: `3f285722a505ea426d53a34f95716781d1aca7c2`.

| Invariant | Test-only RED contract | State |
| --- | --- | --- |
| canonical persisted plan provenance | clone/deserialization/proxy/accessor/cycle/unknown/oversize/tamper cases fail closed at every public boundary; canonical serialization/digest survives persistence | planned RED |
| mutating child topology | top-level `read_only` children and empty mutating scopes are rejected | planned RED |
| outer evidence/review identity | PR number, branches, SHAs, repository, work item, generation, paths, and scopes must all agree | planned RED |
| canonical integration receipt | full child PR snapshot plus controller/transport mutation provenance is required and rechecked before parent readiness | planned RED |
| independent evidence completeness | expected diff source is independently complete; nested path/check/change/thread/review/disposition evidence and minimum observation revision are complete/fresh | planned RED |
| causal chronology | future/backward timestamps, pending completions, stale revision, pre-finding dispositions, and pre-blocker clean reviews fail; sequence rollups are authoritative | planned RED |
| durable cross-instance mutations | two orchestrators create/integrate/ready/publish once; retries reconcile visibility; rejected local queues drain FIFO and remain bounded | planned RED |
| ancestry proof | malformed/truthy/wrong-coordinate/stale proofs fail; exact literal-true proof passes | planned RED |
| same-marker attempts | marker+digest+target matching is permutation-invariant and true ambiguity fails | planned RED |
| canonical Git refs | `HEAD`, symbolic, and pseudo refs fail at plan/evidence/receipt boundaries | planned RED |
| versioned CI policy | repository/base/revision/context/producer/digest is plan-bound; movement/staleness/missing producer fail | planned RED |
| monotonic roster CAS | stale revision/status epoch cannot overwrite newer roster state | planned RED |
| exported attestation API | controller constructor/digest/validator round-trip exact targets and reject tampering | planned RED |
| bounds/partial effects/secrets | proxies/accessors/cycles/oversize and malformed post-effect responses do not leak secrets or duplicate effects | planned RED |

The next commit after the Cycle 3 plan checkpoint must modify only tests/fixtures. Before that commit,
the blob IDs for `github-orchestrator.ts`, `github-evidence.ts`, and `review-router.ts` must equal
their IDs at `3f285722`. GREEN and verification cells remain blank until the single RED checkpoint
has been captured.

## Cycle 3 GREEN and verification evidence

- Plan `d97faf44` plus policy correction `d2c7f374` preceded test and production edits. RED
  `faf2e8f8` ran 52 tests (37 pass, 15 expected fail) with all production blobs equal to
  frozen `3f285722`.
- GREEN `41e8e76e` passes 53/53 focused tests, including bounded retry-before-visibility for
  every durable mutation and authoritative receipt-provenance tamper rejection.
- Strict owned/all-production TypeScript, pinned Pi 0.80.6 offline RPC, immutable base/ancestry,
  17-path scope, full-range diff, and credential-literal scan pass.
- Serialized Shepherd ran 317 tests: 251 pass, 65 unrelated sandbox `spawn EPERM` failures, and
  1 intentional skip; every #478 test passes.

## Cycle 4 consolidated-review RED matrix

Frozen candidate: `d3b6b5e226b17db6ec8350163acdbb41368ec3bf`.

| Contract | Behavior-level RED requirement | State |
| --- | --- | --- |
| stable receipt identity | observation revision/time and merged state can advance without identity drift; wrong planned head branch fails | planned RED |
| canonical restart/readiness topology | recomputed wrong branch/base/marker/path receipts and PRs fail before broker/ready | planned RED |
| bounded cancellable ports | every never-settling external read/mutation/source/workspace/broker/policy call aborts, drains its key, and late effects reconcile once | planned RED |
| sensitive text and external errors | valid-field secret shapes never persist/publish/escape; Error/string/object/undefined rejections become bounded codes | planned RED |
| current CI policy source | current exact policy passes; movement/incomplete/stale/wrong coordinates block existing plan/PR | planned RED |
| complete pseudo-ref grammar | all pseudo/symbolic and segment variants fail across plan/policy/evidence/review/snapshot/receipt | planned RED |
| post-mutation CAS progression | unchanged/regressed roster or ready revision and out-of-order stale writers fail | planned RED |
| descriptor-first dense bounds | large dense and million-length sparse nested arrays reject before descriptor traversal/effects; exact lengths compare | planned RED |
| collision-free tuple identities | colon-bearing distinct session/run pairs pass and exact tuple duplicates reject | planned RED |
| Cycle 3 retention | prior 53 focused contracts and partial-effect/proxy/accessor/cycle/error cases remain green after correction | planned RED |

The next commit after the Cycle 4 plan checkpoint may change only tests and issue-478 fixtures.
Production blob identity and the focused expected-failure matrix must be recorded before GREEN.

## Cycle 4 RED, GREEN, and verification evidence

- PLAN `607e203ef1f76ff112c130ccff5d155973d984f6` preceded every Cycle 4 test and
  production edit.
- Single test-only RED `abbf388b8a852836e0dd10a55b9f17720b9fde22` ran 68 tests: 50 pass and
  18 expected fail. The production blobs still exactly matched frozen candidate `d3b6b5e2`:
  orchestrator `ed576e6455384f123b075d078aed33ca242f2339`, evidence
  `a3076e39e53e2e5c9e7d3dfbc3e52d94af322a7a`, router
  `ca0c8116266be9b31d87657713b40c57e0b10759`.
- Architectural GREEN `b92b5ff7dd3738dc3b3350ebb4d2f2b42074f954` passes all 68 focused tests.
  Stable PR identity/observation, canonical readiness topology, cancellable external ports,
  sensitive-text/error normalization, current policy observation, pseudo-ref grammar, CAS
  progression, descriptor-first dense arrays, and tuple identities are now enforced together.
- Strict owned TypeScript 5.9.3 passes without dependency-check suppression. All 20 production
  Shepherd modules pass strict TypeScript against cached Pi 0.80.6 with third-party declaration
  checking skipped; project source remains fully checked.
- Serialized Shepherd: 332 total, 266 pass, 65 fail, 1 intentional skip. Every #478 test passes;
  every failure is outside owned files and reports the managed sandbox's `spawn EPERM` at the
  process-identity child-process boundary.
- Pinned Pi 0.80.6 offline RPC discovers `pm-shepherd` from `extension`; only expected global
  settings lock warnings are emitted. Immutable base/ancestry, full-range diff, 17-path scope,
  `git diff --check`, and credential-literal scan pass.

## Cycle 5 consolidated-review RED matrix

Frozen candidate: `ca6f6873d168db707bbe58291b5ee1b582e9404f`.

| Contract | Behavior-level RED requirement | State |
| --- | --- | --- |
| exact broker runtime records | malformed/extra request records and request/poll/consume records with wrong ID, gate, marker, options, allowlist, binding, generation, head, question, lifetime, status, actor, source, or chronology never reach `markParentReady` | planned RED |
| complete policy-set refresh | movement/missing/incomplete/stale evidence for either plan-bound coordinate blocks receipt reuse, readiness before/after decision, and ready-effect recovery; every stage queries the complete exact set | planned RED |
| authoritative initial policy topology | a #479-shaped caller constructs a plan through the authoritative full-bundle/config source and public async port only; wrong/ambiguous bundle coordinates fail | planned RED |
| non-self-auth controller authorization | forged/re-digested controller revision/time, changed paths, review result, missing attestation/review, or transport-only receipt blocks both reuse and readiness after independent evidence re-evaluation | planned RED |
| centralized child PR eligibility | open/non-draft and documented open-to-merged/merged transitions pass; draft open, draft merged, closed, and regressed transitions fail identically at reuse/readiness | planned RED |
| CAS-conditioned stable mutation identity | same logical coordinates with different expected revisions produce different authenticated identities; PR observation refresh preserves child integration identity; timeout-before/after-effect recovery remains single-effect | planned RED |
| cookie/session grammar | synthetic `Cookie`, `Set-Cookie`, session ID/token, and response-header values are rejected/redacted in every durable/outbound text field without storing real credentials | planned RED |
| caller lifecycle and join ownership | caller abort/deadline propagates to ports; a timed-out uncooperative keyed call remains excluded; bounded stop reports incomplete/unacknowledged until underlying settlement and clean only after join | planned RED |
| pre-materialization envelopes | oversized raw UTF-8 JSON rejects before parse; oversized object envelopes reject after bounded schema-directed inspection and before generic descriptors/effects; proxy/accessor/closed/dense cases remain fail-closed | planned RED |
| current durable run state | JSON names Cycle 5 frozen candidate, blocked Cycle 4 reviews, exact available checkpoints, verification/review truth, and no stale Cycle 3 current head | planned RED |
| Cycle 4 retention | the existing focused 68 tests pass unchanged at RED and after GREEN | planned RED |

The next commit after the Cycle 5 plan checkpoint may modify only the three matching tests and the
two issue-478 fixtures. The retained 68 tests, each new expected failure row, and exact frozen
production blob IDs must be recorded before any Cycle 5 production edit.

## Cycle 5 RED and first GREEN evidence

- PLAN `7cf9c88ddadee395020444c19ee9f001b0807a53` preceded every Cycle 5 test and
  production edit.
- Single test-only RED `6cb21902244e4bccf390c4e7556eb615e5e1697f` retained 68/68 Cycle 4
  cases and produced 37 intended failures across every new top-level Cycle 5 contract. Production
  blobs remained exactly `c60f2f09b62a11b2eb17fb48fc8197f938ec8eff` (orchestrator),
  `165d483aaaea5d4d67a2b0f88efce1b36118460a` (evidence), and
  `ab8718eb8c2f5a4e2fe3a993283a166fd8e4e961` (router).
- First architectural GREEN passes 109/109 focused tests: retained 68 plus all 41 Cycle 5 cases.
  Strict owned TypeScript 5.9.3 and all 20 production Shepherd modules pass against the cached
  Pi 0.80.6 package and Node declarations. Broader serialized, offline RPC, scope, and data gates
  remain pending until after the GREEN checkpoint.

## Cycle 5 final local verification

- Architectural GREEN `3ae10dc2303409230153e32e6b6231b27b18cdcf` passes 109/109
  focused tests and strict owned plus all-production TypeScript.
- The serialized suite classifies 373 tests as 307 pass, 65 unrelated managed-sandbox
  process-identity `spawn EPERM` failures, and one intentional live-GitHub skip. Every #478 test
  passes; the environmental broad-gate failure is not reported as a full-suite pass.
- Pinned Pi 0.80.6 offline RPC discovers `pm-shepherd` from `extension`; immutable base and frozen
  candidate are ancestors; the exact merge base is `3addb1f4`; full-range `git diff --check`,
  exact 17-path ownership, JSON parsing, and high-confidence credential-literal scans pass.
- GREEN adjusted only `github-orchestrator.test.ts` support fixtures after RED: the fake broker now
  emits the newly required full canonical records, seeded receipts carry authoritative provenance,
  lifecycle recovery observes the created resource, and RUN-STATE asserts the exact RED SHA. These
  edits align fixtures with the stronger contract and do not remove or weaken any RED assertion.

## Cycle 6 consolidated-review RED matrix

Frozen candidate: `63ac436fdac5fc46be7004f8109c4f068aa5749c`.

Frozen production blobs:

- `github-orchestrator.ts`: `dfa8c189bc116c212f5623df508ef918e4d17943`
- `github-evidence.ts`: `ee3cd46f663c6a4cd15db137fb5a8f2cc539f773`
- `review-router.ts`: `611aa0679a84c4a470d19c102b540ea5b3fc103d`
- scope expansion `human-decision.ts`: `0c26e61808a577b3197f9373b659d95739ab20b3`
- scope expansion `github-decision-broker.ts`:
  `a04d331443e6ffda4b66996766a6d4111d664931`

| Contract | Behavior-level RED requirement | State |
| --- | --- | --- |
| real broker composition | strict TypeScript and runtime test instantiate actual `GitHubDecisionBroker`; adapter drives real request record, pending/decided poll result, consumed evidence, and canonical repository reload without invented fields | green |
| bounded broker DTO/provenance | request/poll/consume/reload reject wide, accessor, normal/revoked proxy, unknown-field, missing request comment, and impossible updated/decision/consume chronology with typed sanitized errors before ready | green |
| conditional parent-ready authorization | token digest binds complete policies, exact consumed decision, current parent review/path authority, child receipt+ancestry roster, plan, head, and PR revision; movement inside effect leaves draft true; after-effect drift rolls back idempotently | green |
| intrinsic signal lease | pre-aborted/live genuine signals with own shadows never start/escape calls; signal proxies/incompatible receivers reject; attach/recheck closes races; early/duplicate/late abort acknowledgement reports accurately for ordinary and real-broker calls | green |
| ordered reviews/stable semantic authority | later findings blocks earlier clean, findings-then-clean recovers, equivalent later clean keeps stable authorization/mutation intent, and timeout/restart/receipt reuse remains deterministic | green |
| intrinsic raw/proxy totality | shadowed `Uint8Array.byteLength` cannot bypass predecode bounds; incompatible/subclass/revoked proxies and revoked array/object inputs reject with normalized bounded errors before traps/materialization | green |
| complete shared credential grammar | synthetic npm `_authToken`, netrc whitespace password, lowercase cloud credential keys, and credential-file forms reject/redact in plan/outbound/review/human-decision fields through one exported helper | green |
| receipt chronology | integration before snapshot/path/review/policy/controller observations or in the future rejects; repaired happy fixtures integrate after every authority observation | green |
| non-self-referential RUN-STATE | current cycle 6 uses `candidateRef: "HEAD"`, exact completed checkpoint commits, explicit Cycle 5 blocked review truth, no Cycle 4-as-current state, and no null Cycle 5/6 evidence field | green |
| Cycle 5 retention | all existing 109 focused tests pass unchanged at RED and after GREEN; policy, reauthorization, eligibility, CAS, lifecycle, bounds, and no-merge guarantees remain intact | green |

The first Cycle 6 plan checkpoint is `88513259ffc31fd0853679234c6a42ab6cd04ef6`.
The completed broker map then proved the production class must expose a canonical `readRecord` over
its own repository: supplying a second repository to the adapter could validate a different store,
and the current public compact poll/evidence consume results cannot reconstruct a canonical record.
The amended exact scope is the prior 17 paths plus `github-decision-broker.ts`, its test,
`human-decision.ts`, and its test (21 total). The comprehensive RED may modify the five matching
tests and two issue-478 fixtures. Existing production behavior remains untouched until RED evidence
is captured; the composition row uses the actual broker without casts or invented fields. Before
GREEN, retained 109/109, every intended new failure, and all frozen production blob IDs above must
be recorded.

## Cycle 6 RED, GREEN, and local verification evidence

- Artifact-only scope amendment `2832993b93d07ea20197bad52ec23700fe21fc1e` fixed the exact
  21-path range before RED.
- Comprehensive five-test-file RED `ca4d97d1100b1b44176da9d7dfd6ee6f56f4e1e6` retained the
  prior 109/109 Cycle 5 assertions. Its full run recorded 207 total: 152 pass, 54 intentional
  contract failures, and 1 intentional live skip. Strict TypeScript reported only the six intended
  absent-contract errors, and all five production blob IDs remained byte-identical to `63ac436f`.
- Architectural GREEN `2c6371e725d58b2dc05902d68f9e6812904664d6` passes the five focused
  files with 206 pass, 0 fail, and 1 intentional live-GitHub skip. Strict TypeScript passes for the
  owned production/tests and all 20 Shepherd production modules.
- Post-RED test support edits preserve every expectation: the review-router fixture description no
  longer contains a credential-grammar trigger word; the native decision fixture first persists its
  request comment; the orchestrator fakes now expose canonical request-comment/consumed state,
  validate and roll back the typed ready authorization, carry stable review provenance, use
  post-authority integration times, and explicitly remove request-comment provenance only in the
  negative test. No RED assertion was removed or weakened.
- Pinned Pi 0.80.6 offline discovery, immutable base/ancestry, full-range diff, exact 21-path
  ownership, JSON parsing, and synthetic credential-marker scans pass. The broad serialized suite
  records 427 total, 361 pass, 65 unchanged unrelated managed-sandbox process-identity `spawn
  EPERM` failures, and 1 intentional skip; every Cycle 6 focused assertion passes.

## Cycle 7 consolidated-review RED ledger

Frozen exact candidate: `dbce5b7d0c698bc802594211072fed77eff23c1c`; immutable base:
`3addb1f48be1afe8b1e2b59b54247679d7293805`. Both Cycle 6 reports were read completely. The
planned matrix contains 46 behavior rows and retains the complete Cycle 6 focused suite.

| ID | Rows | RED contract | State |
| --- | ---: | --- | --- |
| C7-AUTH | 10 | mandatory production durable authority boundary atomically rejects movement of policy, review, paths, receipt, ancestry, decision, plan, head, PR revision, or authorization state without clearing draft, an optional transport fallback, or ordinary recovery | green |
| C7-LATE | 6 | uncertain mark settlement is quarantined and joined across before/after-effect timeout, restart-before-visibility, read failure, rollback retry, and stop/key ownership | green |
| C7-STABLE | 7 | semantic authorization and mutation identity survive harmless policy/ancestry/review refresh and restart; actual semantic movement blocks | green |
| C7-REVIEW | 5 | receipt full-attempt digest/time must occur in exact authoritative attestation history; equivalent clean remains compatible and later findings blocks | green |
| C7-CLOCK | 6 | each broker event timestamp and the combined future chronology reject under broker/controller owned clocks through the actual adapter and readiness path | green |
| C7-SECRET | 8 | finite Kubernetes, Docker, AWS assignment/prefix families reject across persistence, comments, plans, titles/bodies, findings, and dispositions without marker reflection | green |
| C7-STATE | 2 | HEAD is the only current candidate semantic; historical SHA in a current slot fails schema invariant | green |
| C7-479 | 1 | production-port-only public prepare/journal/commit/settle trajectory composes real broker, policy, and separate transport/authority/journal roles with rollback, stop, and join | green |
| C7-RET | 1 | all Cycle 6 focused assertions and intentional skip classification remain unchanged | green |

Frozen production blobs before RED: orchestrator `b3515a94e932a6206f2c32f083c1188882a01dfe`,
broker `25c98a3c224d660c7fe6b5de16a30fdf73f95621`, human decision
`4202ba001dd0d48b83d68a65b7004c8db49d0b65`, review router
`a113b4d6bb77f001e8b377c2696c934136b4ceb9`, and GitHub evidence
`23efd2c51280ba83836feef4fcb459e7da4571c0`. RED may change only the five matching tests,
existing issue-478 fixtures when needed, and phase artifacts; production remains byte-identical.

The public prepare/commit split is a testable product contract, not a fake seam: #479 can persist
the prepared intent and exact consumed decision before invoking the conditional effect, then record
settlement afterward. The atomic effect and durable quarantine/rollback share one production
authority-boundary interface. `reconcileParentReadiness` remains the convenience composition.

## Cycle 7 RED, GREEN, REFACTOR, and verification evidence

- PLAN `2c64979829048d3de0d1ff1575c2a4f43cb699ba` precedes every Cycle 7 test and
  production edit.
- Test-only RED `10033bc532d06967ce960e408c2bc9725020478a` records 290 total: 217 pass,
  72 intentional failures, and 1 intentional live-GitHub skip. Strict owned TypeScript reports only
  the 14 intentionally absent contracts. The five production blob IDs above remain unchanged.
- Architectural GREEN `5bab0bc7e56292171eb28618cc2f37488ed1b7a4` implements the public
  authority/prepare/commit contracts, stable authorization and freshness split, late-effect
  quarantine/recovery, review-attempt provenance, broker clock, and finite credential schemas.
- REFACTOR proof `87e704010f3e2226d8393d12e1a1bdf72df212a0` changes the late-effect
  proof to exact 500 ms effects after a 100 ms timeout, adds caller cancellation, updates both
  decision chronology fields in the semantic-movement fixture, and removes a temporary
  canonicalizer repair. No assertion is removed or weakened.
- Independent architecture-audit RED `b1560e76a3abbac5efcd33b2740b7275b6acc137`
  records 297 total, 294 pass, 2 intentional failures, and 1 skip. The failures prove the remaining
  optional legacy ready-mutation path and structural fake projection before the audit GREEN.
- Audit GREEN `915882c219f52da2c1edebce84d2bf90c61a4592` requires the production
  authority, removes ready mutation/rollback from the transport, returns typed atomic conflicts,
  and proves separate production-typed transport, authority, and journal roles: 297 total,
  296 pass, 0 fail, and 1 skip.
- Final focused five-file suite: 297 total, 296 pass, 0 fail, 1 intentional live-GitHub skip.
  Strict TypeScript passes for the 10 owned production/test files and all 20 Shepherd production
  modules. Pinned Pi 0.80.6 offline RPC discovers `pm-shepherd` from `extension`.
- Serialized Shepherd is an environmental failure: 517 total, 451 pass, 65 unchanged unrelated
  managed-sandbox process-identity `spawn EPERM` failures, and 1 intentional skip. Every Cycle 7
  assertion passes.
- Both reports were replayed after REFACTOR. Named tests cover all atomic coordinates, exact
  before/after timing and cancellation, durable/keyed quarantine, restart/read/rollback failure,
  stable harmless refresh versus semantic movement, authoritative full-attempt provenance,
  owned-clock future chronology, all eight finite credential forms, current-HEAD schema, and the
  public production-port-only #479 prepare/journal/commit seam. No optional authority route,
  deprecated transport ready mutation, or fake-only authorization validator exists.
- Immutable base and reviewed candidate are ancestors; exact merge base, full-range diff check,
  exact 21-path ownership, three JSON parses, and the five-path explicit synthetic AWS marker
  allowlist pass with zero production/artifact or unexpected candidates.

## Cycle 8 consolidated-review RED ledger

Frozen exact candidate: `b90037df1fff38c755ebc8025579120d17031330`; immutable base:
`3addb1f48be1afe8b1e2b59b54247679d7293805`. Both Cycle 7 reports were read completely. Their
seven unique families map to 48 planned behavior rows and retain all 297 Cycle 7 focused cases.

| ID | Rows | RED contract | State |
| --- | ---: | --- | --- |
| C8-SECRET | 20 | every recognized credential-assignment suffix rejects under an unknown provider prefix across all durable/outbound consumers; exact safe-name exception is applied only after classification and finite kube/docker/AWS forms remain | red_then_green |
| C8-479 | 6 | separate production-typed transport/authority/journal roles prove success, typed conflict, uncertainty, rollback, stop incomplete/joined, and settlement without `any`, casts, fake projection, or private shortcuts | red_then_green |
| C8-UNCERTAIN | 4 | immediate apply-then-reject starts durable recovery despite failed reads, restores draft, blocks reentry, and retains stop/key ownership | red_then_green |
| C8-BROKER | 4 | actual broker rereads exact consumed state before new-request expiry validation, resumes prepared commit once after expiry, and rejects truly new expired request/decision | red_then_green |
| C8-FENCE | 5 | per-attempt deadline abort, durable predecessor fencing, successful later retry, ignored superseded result, authoritative draft observation, and eventual join are enforced | red_then_green |
| C8-RESTART | 4 | serialized prepared/journal state recreates controller/broker/journal/transport/authority over shared durable backing and resumes without `WeakMap` or object identity | red_then_green |
| C8-FRESH | 5 | revalidation forwards refreshed policy/ancestry/equivalent-clean freshness while original authorization, key, and intent remain exact | red_then_green |

Frozen production blobs before RED: orchestrator `668a55af55413c1cc595424e87ce352c355eec88`,
broker `7be6785190176a8c15660fb180fc95c207b76d5b`, human decision
`b1c0c198c33c95c8fabb0f911a42513d2305cb17`, review router
`a586405153e2e666a57b832e7d4b48df80e3265c`, and GitHub evidence
`23efd2c51280ba83836feef4fcb459e7da4571c0`. RED may modify only the five matching test files
inside the existing 21-path range. No fixture change is planned; production must remain byte-exact.

Rollback ownership is explicit before RED: the controller binds stable recovery identity to the
prepared authorization/key/intent, carries the original ready mutation identity instead of guessing
resource revisions, and sends an ordered attempt fence in every rollback request.
The durable authority must claim that fence and supersede its predecessor before returning, while
every attempt can only idempotently restore the exact draft. Controller deadlines abort response
waits, not durable cleanup; late/superseded results cannot release quarantine or settle state. A
matching fenced durable result carrying the exact draft is the authoritative observation that
releases the key. Reconstructed adapters rely on shared durable backing, never module object
identity.

Cycle 8 followed strict artifact PLAN -> comprehensive test-only RED -> coherent GREEN -> bounded
REFACTOR -> exact evidence. PLAN is `bccee8e6cdbcb6e38419114f264222b1f5616f66`; RED is
`851bb3bfa3e23042211a8b37f3a97253cc6fedf5`; GREEN is
`013bdc8b264e1ce8808d4af2558e2ec40b85ee49`; REFACTOR is
`26a7d476bdfaa4e263196fb76f7f43b5a3ad799e`.

- RED: 374 total / 314 pass / 59 intended fail / 1 intentional live-sandbox skip. Strict owned
  TypeScript reported only four intended diagnostics for the missing recovery fence/validator.
  All five production blob IDs remained exact.
- GREEN/REFACTOR: targeted Cycle 8 orchestrator 46/46; complete focused five-file route 374 total /
  373 pass / 0 fail / 1 skip. Strict owned and all-20-production TypeScript pass against pinned Pi
  0.80.6. The exact #479 fixture uses typed public ports in success, conflict, uncertainty,
  rollback, stop/join, journal settlement, and serialized reconstruction paths.
- Both complete Cycle 7 reports were replayed after REFACTOR. No expectation was removed, weakened,
  skipped, or converted to fake-only proof. The only post-RED test support changes teach fakes the
  public recovery fence/durable backing and update two historical immediate-rejection expectations
  to the newly required rollback semantics.
- Offline RPC, immutable base/reviewed-candidate ancestry, exact merge base, full diff, exact
  21-path scope, three JSON parses, marker confinement, and clean pre-evidence status pass.
  Serialized Shepherd is environmental failure: 594 total / 528 pass / 65 unchanged unrelated
  sandbox `spawn EPERM` failures / 1 intentional skip; all Cycle 8/focused tests pass.

## Cycle 9 consolidated-review RED ledger

Frozen exact candidate: `f97a698df90010ae072554e04563a8134a8e5f6e`; immutable base:
`3addb1f48be1afe8b1e2b59b54247679d7293805`. Both Cycle 8 reports were read completely. The retained
374 focused cases plus one intentional live skip remain mandatory.

| ID | Rows | RED contract | State |
| --- | ---: | --- | --- |
| C9-RESULT | 8 | an applied-then-rejected uncertain result can only be blocked/quarantined; visibility cannot report ready; keyed recovery, reentry, stop/join, draft restoration, and blocked settlement stay consistent | red_then_green |
| C9-DURABLE | 13 | one canonically queryable authority record implements `ready_invoking -> ready_effect_applied -> ready_settled` or fenced `recovery_claimed -> draft_restored`; restart occurs at visible-ready/unsettled state and fences the original writer | red_then_green |
| C9-ASSIGN | 40 | eight boundary/control shapes run through each of five shared consumers: leading underscore, 127, 128, 129, 256, largest in-field, over-field, and exact safe exception; full-name suffix classification fails closed without marker reflection | red_then_green |
| C9-479 | 8 | a public typed broker and canonical unknown decoders reconstruct all five roles and validate decision/prepared/journal/authority/recovery/fence/mutation/settlement values across the exact #479 lifecycle | red_then_green |

Frozen production blobs before RED: orchestrator `ab9b2c0ed254ecdbffa10c4ca2b13420de01268a`,
broker `7be6785190176a8c15660fb180fc95c207b76d5b`, human decision
`fc1c62307ccca0c2590ea0a7cd61626876f3f71f`, review router
`31234c70ade7341a2af01aeac2d81a015b696e6b`, and evidence
`23efd2c51280ba83836feef4fcb459e7da4571c0`. RED may modify only the five matching test files.

The authority record is the test oracle, not visible PR state: fence 0 original writes require an
exact `ready_invoking` CAS; a recovery claim moves to fence >=1 before rollback and makes every
fence-0/older writer stale. Only explicit applied-result settlement can produce `ready_settled`;
only exact fenced draft proof can produce `draft_restored`. Any other state forces a public blocked
outcome and retains lifecycle ownership.

Cycle 9 followed artifact PLAN `7ad23ed4` -> complete test-only RED `9278e97e` -> coherent GREEN
`593ba1cf`. RED executed 442 cases: 398 pass, 43 intended fail, and 1 intentional skip, with the
five production blobs unchanged. GREEN executes 450 cases: 449 pass, 0 fail, and 1 intentional
skip. Strict owned/all-production TypeScript and the exact public value-serialized #479 fixture
pass. Serialized Shepherd is 670 total / 604 pass / 65 unchanged managed-sandbox process-identity
`spawn EPERM` failures / 1 skip. Post-RED test edits align retained uncertain-result expectations
with the stronger typed blocked contract and teach only public fake/port implementations the
canonical authority state; no Cycle 9 assertion was removed, skipped, or weakened.

## Cycle 10 consolidated-review RED ledger

Frozen exact candidate: `a49e4df2798281d1e64c722ccbcab5f4a678c3e1`; immutable base and merge base:
`3addb1f48be1afe8b1e2b59b54247679d7293805`. Both Cycle 9 reports were read completely and their
union is one strict RED→GREEN correction. The retained 450 focused cases, including the one
intentional live-sandbox skip, remain mandatory.

| ID | Rows | RED contract | State |
| --- | ---: | --- | --- |
| C10-ORDER | 42 | prepare/reconcile query and recover each unsettled phase before independently broken roster, review, policy, broker, pending/expired request, or rejected decision gates | planned_red |
| C10-CAS | 6 | settlement-wins lost responses terminate recovery without rollback; recovery-wins conflicts settlement, blocks, restores draft once, and joins | planned_red |
| C10-REVISION | 15 | prepare/commit/reconcile require exact stored applied revision and provenance, rejecting original/lower/higher mismatch | planned_red |
| C10-CONFIRM | 4 | bounded post-rollback confirmation failure/hang is superseded by a newer fence and cannot freeze key/stop ownership | planned_red |
| C10-NOT-STARTED | 4 | pre-application effect rejection/timeout/cancel/malformed paths have durable invoking state and terminal no-op draft restoration | planned_red |
| C10-ASSIGN | 135 | all suffixes with `+=` cross five consumers; case/index policy is fail-closed; exact unindexed `FEATURE_TOKEN` controls remain public without marker reflection | planned_red |
| C10-WARNING | 6 | warning findings require exact-current-head later fixed disposition and a later clean review | planned_red |
| C10-SNAPSHOT | 17 | decoded JSON rejects duplicate, orphan, missing reciprocal, revision/fence/settlement-incoherent, oversized, and extra-field state while reordered equivalents remain canonical | planned_red |

Frozen production blobs before RED: orchestrator `538962e4e30410dea6e714d565018639e23d1efa`,
broker `7be6785190176a8c15660fb180fc95c207b76d5b`, GitHub evidence
`23efd2c51280ba83836feef4fcb459e7da4571c0`, human decision
`fc1c62307ccca0c2590ea0a7cd61626876f3f71f`, and review router
`8b14fb1fd54938d9e49a880d75b2089c978766c0`. RED may change only the five matching test files and
must remain executable. Exact failing row groups and counts are recorded after the RED run; no
production blob may move before that evidence exists.
