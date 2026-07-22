# Summary: #478

Status: Cycle 13 local GREEN `e0101044bb68f8a6b4cf45960029aac8d8b1ff78` follows artifact PLAN
`27e7b5d2736c62b80618de020e743df49abf76b6` and executable RED
`a36188f4e3f2ee532f10bd38fbcaa1d7ce43e6ae`. Focused is 1061 total / 1060 pass / 0 fail / 1
intentional skip; Cycle 13 targeted is 68/68 on consecutive runs. Strict owned and all-production
TypeScript pass, pinned Pi 0.80.6 offline RPC returns `true`, exact base/scope/integrity gates pass,
and the 60,000-opener scanner case completes in roughly 14-24 ms. The broad route remains non-zero
only in 65 unchanged managed-sandbox process/lease tests, and no fresh exact-head review ran, so
`verificationPassed` and `reviewCoveragePassed` remain false. This is local evidence, not a frozen
or review-clean candidate.

The plan-first checkpoint fixes the immutable base, owned file boundary, strict RED→GREEN→REFACTOR
sequence, fake-only transport policy, exact-head review policy, human gates, and coordinator-bounded
verification matrix. The test-only RED checkpoint is now captured: all three matching test files
fail because their production modules are intentionally absent.

Minimal GREEN completed at 21 focused passes. A subsequent test-only adversarial RED checkpoint
captured 10 receipt, restart, planned-review binding, parent handoff, hostile-array, duplicate
finding, and disposition failures against unchanged production. The matching correction is pushed
at `40ce66d4b5010b92089895a05709687143d15a05` and passes 27/27 focused tests.

The delivered boundary now provides exact-shape authoritative evidence validation, a declarative
independent Codex-only review route, reconcile-before-mutate issue/PR/roster publication, exact
generation/marker/base/head integration receipts, merged-PR restart recovery, upstream child and
parent handoff capture, and a broker-consumed exact-generation/head parent ready gate. It never
wires sessions or performs a parent merge.

All parent-authorized local gates pass at the implementation head: the serialized Shepherd suite
reports 290 pass, 0 fail, and one intentional sandbox skip; strict owned and all-production
TypeScript pass with TypeScript 5.9.3 against cached Pi 0.80.6; pinned offline Pi RPC returns `true`
for `pm-shepherd`; and exact diff, immutable-base, and owned-path checks pass. No reviewer was
started; stable-head independent review remains parent-owned.

Ready PR https://github.com/polymetrics-ai/cli/pull/487 targets
`feat/471-pi-agent-session-shepherd` with the required Conventional Commit title and `Refs #478` /
`Refs #471` linkage. The worker requested no reviewer and performed no integration or merge.

## Functional review correction in progress

The deep stable-head review at `093b3c90` supersedes the earlier clean-local-gates statement. Eleven
accepted correctness findings now require a fresh strict TDD slice covering authoritative changed
paths and integrations, trusted CI/session provenance, deterministic review selection, keyed
idempotency, positive generations, canonical Git refs, complete pagination evidence, and mutation
recovery. This artifact checkpoint precedes the required single test-only RED commit; no correction
production code has been edited yet. Manual GSD fallback and `local_critical_path` remain recorded.

The correction now has test-only RED `4e02d059` and coherent GREEN `8e32896a`. Focused 38/38,
strict owned/all-production TypeScript, offline pinned Pi registration, and base/head/diff/scope
checks pass. The full serialized command ran all 302 tests; 65 unrelated tests fail solely because
the managed sandbox rejects their process-identity child process with `spawn EPERM`, while every
#478 test passes. Accordingly the phase remains `in_progress` with `verificationPassed: false`.
GitHub DNS resolution failed on every push attempt, so plan, RED, and GREEN commits remain local and
PR #487 could not be updated from this environment.

## Cycle 3 correction status

Two corrected deep-review ledgers for frozen candidate `3f285722` converge on one fourteen-invariant
architectural correction. The new slice covers durable canonical plan provenance, mutating-only
top-level children, complete PR/review/integration binding, independent complete evidence and causal
freshness, cross-instance conditional mutations, exact ancestry proofs, deterministic same-marker
attestations, symbolic-ref rejection, versioned plan-bound CI policy, monotonic roster publication,
an exported controller attestation API, and adversarial bounds/partial-effect safety. This is the
artifact-only checkpoint; tests and production remain unchanged until the required single RED
commit. Network publication remains deferred under the existing DNS blocker.

Cycle 3 is locally complete at GREEN `41e8e76e`: all fourteen invariants are implemented and
53/53 focused tests pass. Strict owned/all-production TypeScript, pinned Pi 0.80.6 offline
discovery, immutable-base/diff/17-path ownership, and credential scans pass. The serialized suite
records 251 pass, 65 unrelated sandbox `spawn EPERM` failures, and one intentional skip across
317 tests; every #478 test passes. Exact-head review and publication remain parent-owned/deferred.

Cycle 4 begins from frozen `d3b6b5e2` after two final deep reviews. The consolidated correction
separates stable receipt topology from observations, validates canonical child topology at restart
and readiness, introduces bounded cancellable/redacted external ports and a current policy source,
completes pseudo-ref/CAS/dense-bound/tuple-key safety, and retains all Cycle 3 contracts. This
artifact-only checkpoint precedes the single required test-only RED; production/tests are unchanged.

Cycle 4 is locally complete at PLAN `607e203e`, single test-only RED `abbf388b`, and architectural
GREEN `b92b5ff7`. All ten consolidated contracts pass 68/68 focused tests and strict owned plus
all-production TypeScript against pinned Pi 0.80.6. Offline RPC still discovers `pm-shepherd`, and
base/ancestry/diff/17-path/data scans pass. The serialized suite records 266 pass, 65 unrelated
sandbox `spawn EPERM` failures, and one intentional skip across 332 tests; all #478 tests pass.
No live/network/prohibited action ran. The evidence commit and clean candidate are handed to the
parent for two fresh exact-head `xhigh` reviews.

## Cycle 5 correction status

Two independent reviews block frozen Cycle 4 candidate `ca6f6873`. Their unique union is one
artifact-first correction: exact broker records; generation-wide policy authority; independently
re-evaluated controller receipt authorization; consistent child eligibility; CAS-bound stable
mutation identity; cookie/session redaction; caller-linked lifecycle with live-call key ownership
and bounded stop/join; pre-materialization raw/schema-directed envelope bounds; and atomic current
run state. This plan checkpoint precedes the one comprehensive test-only RED; production and tests
remain frozen. Push/network/GitHub and parent-owned review remain excluded.

Cycle 5 is locally complete at PLAN `7cf9c88d`, comprehensive test-only RED `6cb21902`, and
architectural GREEN `3ae10dc2`. Focused #478 passes 109/109; strict owned/all-production
TypeScript and pinned Pi 0.80.6 offline discovery pass; immutable-base/ancestry, full-range diff,
exact 17-path scope, JSON, and credential scans pass. The serialized suite records 307 pass, 65
unrelated managed-sandbox `spawn EPERM` failures, and 1 intentional skip across 373 tests; every
#478 test passes. Post-RED test edits only align support fixtures with the strengthened production
contract and do not weaken expectations. No prohibited, network, GitHub, reviewer, integration,
or merge action ran; two fresh exact-head `xhigh` reviews remain parent-owned.

## Cycle 6 correction status

Both Cycle 5 reviews block exact candidate `63ac436f`. Cycle 6 consolidates their unique union and
both warnings: real production broker composition; bounded canonical decision provenance;
conditional ready authorization and rollback; intrinsic signal/raw ownership; ordered review
attempts with stable semantic authority; one complete credential grammar; receipt chronology; and
non-self-referential current RUN-STATE. The completed broker contract map amended the exact scope
from 18 to 21 paths before RED: broker-owned `readRecord` and its native test are required to compose
the real compact poll/evidence-only consume API without a second repository, while
`human-decision.ts` and its native test own canonical chronology and reuse the shared
`review-router.ts` credential grammar.

Cycle 6 is locally complete at amended PLAN `2832993b`, comprehensive five-test-file RED
`ca4d97d1`, and architectural GREEN `2c6371e7`. The focused route passes 206 assertions with zero
failures and one intentional live-GitHub skip; retained Cycle 5 remains 109/109. Strict owned and all
20 production TypeScript modules pass, pinned Pi 0.80.6 offline discovery succeeds, and immutable
base/ancestry, full-range diff, exact 21-path scope, JSON, and synthetic credential scans pass. The
broad serialized suite records 361 pass, 65 unrelated managed-sandbox `spawn EPERM` failures, and
one intentional skip across 427 tests. RUN-STATE now uses non-circular `HEAD` evidence semantics
with exact completed checkpoints. No push, network, live GitHub, Go, connector, `make`, #479,
reviewer, integration, or merge action ran; two fresh exact-head reviews remain parent-owned.

## Cycle 7 correction status

Both Cycle 6 exact-head reports against historical candidate `dbce5b7d` were consolidated into one
46-row correction. PLAN `2c649798`, test-only RED `10033bc5`, architectural GREEN `5bab0bc7`,
REFACTOR proof `87e70401`, audit RED `b1560e76`, and audit GREEN `915882c2` implement one mandatory
durable parent-ready authority boundary, stable semantic
authorization separate from freshness, complete uncertain-effect quarantine/settlement/rollback,
authoritative review-attempt provenance, owned-clock broker chronology, finite credential schemas,
one current `HEAD` run-state semantic, and a public production-port-only #479 prepare/commit seam.

The reports were replayed line by line after REFACTOR, and named passing tests represent every
consolidated family. Exact 500 ms before/after effects following a 100 ms timeout, cancellation,
restart quarantine, read failure, transient rollback retry, harmless refresh, real semantic
movement, forged/full-attempt provenance, future timestamps, public-port-only wiring, and the
absence of a legacy transport mutation fallback are explicit. The #479 proof separates
production-typed transport, authority, and journal roles instead of projecting the test fake.

Cycle 7 is locally verified: the focused five-file suite records 297 total, 296 pass, 0 fail, and 1
intentional live-GitHub skip; strict owned/all-production TypeScript and pinned Pi 0.80.6 offline
discovery pass. The serialized suite is an environmental failure with 517 total, 451 pass, 65
unchanged unrelated managed-sandbox `spawn EPERM` failures, and 1 intentional skip. Immutable base
and reviewed-candidate ancestry, full-range diff, exact 21-path scope, JSON, and explicit
test-synthetic marker scans pass. No external/prohibited action ran; fresh exact-head review remains
parent-owned.

## Cycle 8 correction status

Both Cycle 7 exact-head reports against `b90037df` are fully accepted as one 48-row correction.
The seven families are provider-neutral credential suffix closure, strict #479 type/recovery
composition, uncertain immediate rejection recovery, real-broker prepared resume after expiry,
bounded fenced rollback attempts, reconstructed durable restart, and refreshed freshness delivery.

The durable ownership decision is frozen before RED: a stable recovery identity plus ordered
attempt fence and original ready mutation identity bind every rollback request without revision
guessing; the authority durably supersedes predecessors and may
only restore draft; the controller aborts each response wait, ignores superseded results, keeps key
and stop ownership, and releases quarantine only on the matching fenced durable draft result.
Durable backing—not `WeakMap` or adapter identity—owns cross-instance truth.

PLAN `bccee8e6`, comprehensive five-test-file RED `851bb3bf`, coherent GREEN `013bdc8b`, and
bounded REFACTOR `26a7d476` complete that architecture. RED records 374 total / 314 pass / 59
intended failures / 1 skip with four intended strict diagnostics and all production blobs frozen.
After REFACTOR, targeted Cycle 8 is 46/46 and the complete focused route is 374 total / 373 pass /
0 fail / 1 intentional live-sandbox skip. Both Cycle 7 reports were re-read completely after
REFACTOR and every family maps to named passing evidence.

Strict TypeScript passes for all five owned production/test pairs and all 20 production modules.
Pinned Pi 0.80.6 offline RPC discovers `pm-shepherd`. Serialized Shepherd is an environmental
failure at 594 total / 528 pass / 65 unchanged unrelated sandbox `spawn EPERM` failures / 1 skip;
all Cycle 8 and focused tests pass. Immutable-base and reviewed-candidate ancestry, exact merge
base, full diff, exact 21-path scope, three JSON parses, synthetic-marker confinement, and clean
pre-evidence status pass.

Required skills/contracts and `manual_gsd_fallback` are recorded. The evidence candidate remains
non-self-referential `HEAD`; its exact SHA is reported after commit. No Go, connector, parent/main
worktree, #475, network/GitHub, push, reviewer, integration, or merge action ran. Parent owns
publication, two fresh exact-head reviews, dispositions, integration, and every human gate.

## Cycle 9 correction status

Both Cycle 8 reports against `f97a698d` were consolidated into one completed 69-row correction. The
four families are uncertain-result consistency (8), durable dangerous-point restart and
original-writer fencing (13), total assignment parsing across five consumers (40), and the exact
typed/value-serialized #479 production-role fixture (8).

The central architecture is authority-owned. A canonical durable record progresses from
`ready_invoking` to `ready_effect_applied` and only then, through an explicit response settlement,
to `ready_settled`. Uncertainty instead claims a monotonic recovery fence before rollback, making
the original writer and older attempts stale, and ends only at exact `draft_restored`. Prepare,
commit, and reconcile consult this truth before ready reuse; visible non-draft state cannot override
an unsettled record. Stop/key ownership lasts through terminal draft settlement.

The provider-neutral parser now consumes the complete uppercase shell assignment name, including a
leading underscore, before suffix classification; 127/128/129/256/largest-in-field/over-field
boundaries and the exact safe exception pass through every shared durable/outbound consumer.
The #479 fixture decodes serialized values into `unknown` and applies exported production
validators without `any`, casts, fake projection, private shortcuts, or object-identity restart.

PLAN `7ad23ed4` preceded test-only RED `9278e97e` and coherent GREEN `593ba1cf`. RED recorded
398 pass / 43 intended fail / 1 skip with frozen production; GREEN records 449 pass / 0 fail / 1
skip across 450 focused cases. Strict owned/all-production TypeScript, pinned offline RPC, exact
base/ancestry/merge-base/diff/21-path ownership, three JSON parses, marker confinement, and both
Cycle 8 report replays pass. Serialized Shepherd is 670 total / 604 pass / 65 unchanged
managed-sandbox process-identity `spawn EPERM` failures / 1 skip. No standalone production
refactor was necessary and no prohibited or external action ran.

## Cycle 10 correction status

Both Cycle 9 reviews blocked `a49e4df2` and were consolidated into one plan-first correction. The
seven blockers are authority-first recovery ordering, settlement/recovery CAS arbitration, exact
applied-revision proof, bounded post-rollback confirmation, terminal pre-application rejection,
complete assignment operator/case/index policy, and warning-finding disposition chronology. The
same slice hardens the value-serialized restart snapshot and corrects verification truth.

The frozen design separates durable begin from the uncertain ready effect, stores exact applied
revision, queries authority before unrelated readiness gates, and makes terminal settlement versus
fenced recovery mutually exclusive. `+=`, mixed/lowercase identifiers, and indexed assignments are
classified by complete base name; only exact unindexed `FEATURE_TOKEN` is public. Snapshot maps and
journal entries require unique reciprocal canonical state. At the historical Cycle 10 PLAN
checkpoint, no Cycle 10 test or production edit had run.

Cycle 10 is now locally complete at PLAN `470a8a85`, RED `2256971a`, GREEN `5f46206e`, and
refactor `8946b67b`. Focused evidence is 687 total / 686 pass / 0 fail / 1 intentional skip;
strict owned/all-production TypeScript passes. The broad route remains non-zero at 841 pass / 65
managed-sandbox failures / 1 skip, so verification remains false pending parent-owned review.

## Cycle 11 correction status

Both Cycle 10 reviews block exact candidate `3b39cfce`. Their complete union is frozen into one
artifact-first correction: a lost/rejected/timed-out/cancelled/malformed durable begin retains its
invocation, key, and stop owner through settlement and a subsequent authority reconciliation;
every typed non-applied compare conflict returns exact atomic tombstone proof for the requested
invocation; the #479 decoder validates one coherent settlement/authority/visibility/revision/
mutation history; C10-CONFIRM uses causal latches rather than timing guesses; and sensitive
assignment values redact escaped and substitution tails completely while validators stay generic.

Cycle 10 did complete locally at PLAN `470a8a85`, RED `2256971a`, GREEN `5f46206e`, refactor
`8946b67b`, and evidence `3b39cfce`, but its independent review is blocked. Cycle 11 PLAN
`863bf94a` precedes executable RED `1b4aa6f1` and coherent GREEN `e765e0d3`. The exact 21-path
scope is unchanged; parent ownership of publication, reviews, integration, merge, ready, and human
gates is unchanged.

The complete Cycle 11 RED now executes 791 focused tests: 743 pass, 42 intended failing leaves plus
five failing parent containers, and one intentional live-sandbox skip. Failures isolate all six
begin settlement trajectories, thirteen cross-history snapshot gaps, ten missing typed-coordinate
terminal proofs, three persistent moved/foreign tombstones, and ten incomplete direct redactions.
All fifty durable/outbound consumer rows already reject generically without marker or `API_KEY`
reflection. GREEN makes all 60 assignment rows pass and closes every begin, terminal-conflict, and
unified-history row. Focused evidence is 791 total / 790 pass / 0 fail / 1 intentional skip. The
causally latched C10-CONFIRM family passes five consecutive 5/5 runs at both RED and GREEN. Strict
TypeScript passes for the five owned pairs and all 20 production modules; pinned offline RPC
returns `true`; immutable base/ancestry/merge-base, full-range diff, exact 21 paths, three JSON
parses, marker confinement, and both Cycle 10 report replays pass. The serialized route is 1011
total / 945 pass / 65 unchanged managed-sandbox `spawn EPERM` failures / 1 skip, so
`verificationPassed` remains false; `reviewCoveragePassed` is also false pending parent-owned fresh
review. No network, publication, review, integration, ready, merge, or human-gate action has run.

## Cycle 12 correction status

Both Cycle 11 reviews block exact historical candidate `4f0e17df`. Their 399-line union was frozen
in an artifact-first plan against the unchanged base, merge base, 21 paths, and five production
blobs. Original executable RED `2649cf6d` is 942 total / 885 pass / 56 intended fail / 1 skip:
six BEGIN, seven orphan-role, five global-sequence, fifteen recovery-claimed, and eighteen direct
assignment leaves plus five parent containers. All ninety initial consumers and the ordinary
newline control passed while all production blobs remained frozen.

Independent implementation review strengthened each BEGIN permutation to use a distinct foreign
repository, marker, generation, PR, and head, and added ANSI-C escaped-quote, case-pattern, and
heredoc command-substitution assignment forms. The resulting reviewer-gap RED is 978 total / 963
pass / 14 intended fail / 1 skip: twelve leaves plus two parent containers. Its thirty new consumer
rows already rejected generically.

Coherent GREEN `723fdc122cea75a5d6f146fb8b39383e9e5795e3` now retains separate requested and
observed durable owners after valid mismatched begin, performs foreign recovery only against the
returned state's exact repository/marker/generation coordinate, and keeps keyed/stop ownership
until both terminal proofs join. Requested effects remain zero. The restart decoder reverse-
consumes every retained role into exactly one prepared/decision/PR history, validates all
recovery-claimed visibility windows, and enforces one globally unique causal mutation sequence.
The scanner consumes all twelve bounded multiline/composite forms for both `=` and `+=`.

Focused GREEN is 978 total / 977 pass / 0 fail / 1 intentional skip. C12-BEGIN is 6/6, graph
orphan/sequence/claim are 7/7, 5/5, and 15/15, and C12-ASSIGN is 144/144: 24 direct redactions plus
120 generic/no-marker consumer checks. Strict five-pair and all-20-production TypeScript pass with
TypeScript 5.9.3 against cached Pi 0.80.6; pinned offline RPC discovers `pm-shepherd` from
`extension`. The broad route is an honest environmental non-zero result at 1198 total / 1132 pass
/ 65 managed-sandbox process-spawn/lease failures / 1 skip.

Immutable base and reviewed-candidate ancestry, exact merge base `3addb1f4`, full-range diff,
exact 21 paths, three JSON parses, and Cycle 12 marker confinement pass. The two Cycle 11 reports
remain exactly 399 lines with SHA-256
`f2aa1e4a89686c6ae1748252c994d18a602167c56f61f28583ff52162b0d5d27` and
`d8e0fdfca0696f6446c0e85af43fd2471e8112a693688f053cccd547c1e430a1`. Final production blobs
are orchestrator `ca07667f4e598fee472ae174b2a3c55bc708db55`, router
`2c5fd80e4ee5ba536fb7f608ca4e424661a5431e`, broker `7be67851`, evidence `058ad162`, and
human decision `fc1c6230`. `verificationPassed` and `reviewCoveragePassed` remain false. No
network, publication, reviewer dispatch, integration, ready, merge, or human gate ran.

## Cycle 13 correction status

Both Cycle 12 reports were read completely: 487 total lines with SHA-256
`b7724f6845e0c48ac23f88e942fffe84d86faac532a2a08c914259e94eeea06e` and
`38ccafdc48e4cf49043cc6bb5946b91910aaf18a74d17488d20209f763234593`. Their union is frozen
against candidate/tree `baef7615` / `6bf70b7a`, immutable/exact merge base `3addb1f4`, clean exact
21-path ownership, and five production blobs.

The complete union is implemented at GREEN `e0101044`: every mismatched begin joins requested and
exact returned-coordinate proof; coherent authority-terminal/journal-pending windows return
deterministic explicit settlement repairs; prepared authority binds the exact consumed affirmative
decision and unique canonical marker owner; and the assignment scanner is a forward stack with
O(1) maintained composite/heredoc references. The split remains BEGIN 16, cross-store 10, decision
9, marker 4, scanner 30, and artifact 4.

Focused passes 1060/1061 with one intentional skip; targeted Cycle 13 passes 68/68 twice. The
single-pass source gate excludes `lastIndexOf`, per-character frame search/reverse, and frame-copy
patterns; the maximum dense input falls from the reviewed 11.5-second implementation probe to
14-24 ms. Post-RED expectation alignments only make the head-divergence fixture deterministic,
record explicit-stop cancellation of an already excluded waiter, and stop treating Cycle 12's
predeclared cross-bound duplicate-marker helper as canonical while retaining all twelve leaf
identities and single-owner orphan/sequence controls.

Strict TypeScript passes for five owned pairs and all 20 production modules against cached Pi
0.80.6; offline RPC discovers `pm-shepherd`; exact merge base, ancestry, full diff, 21 paths, three
JSON parses, marker confinement, and both 487-line report identities pass. Serialized broad is an
environmental 1281 total / 1215 pass / 65 unchanged managed-sandbox failures / 1 skip. Current
production blobs are orchestrator `63e1f68354de7b499aa727ab133caa84e8e1a35d`, router
`8eb32b882ad030c9c9bd9bc7ba7f5d91884b293d`, broker `7be67851`, evidence `058ad162`, and human
decision `fc1c6230`. The adapter remains `manual_gsd_fallback`; execution is `local_critical_path`.
No network, GitHub, push, reviewer, integration, ready, merge, or human-gate action ran.
