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
