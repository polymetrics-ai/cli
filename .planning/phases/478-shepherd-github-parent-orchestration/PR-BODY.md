## Summary

- add a bounded, typed GitHub parent/child orchestration transport with reconcile-before-mutate
  issue, stacked-PR, roster, and integration operations, plus separate mandatory durable authority
  and journal ports for parent-ready transitions
- validate authoritative CI, review-thread, requested-change, finding-disposition, and exact-range
  independent Codex review evidence
- reuse dependency scheduling, autonomy reconciliation, workspace handoffs, and the existing human
  decision broker while keeping review execution and parent merge outside this slice

## Initial delivery GSD / TDD (historical)

- mode: `manual_gsd_fallback` because the healthy repo adapter does not expose
  `programming-loop`
- required skills: `gsd-programming-loop`, `github-issue-first-delivery`, `gsd-workstreams`,
  `architecture-patterns`, `javascript-testing-patterns`
- initial test-only RED: 0 pass / 3 absent-module failures
- minimal GREEN: 21/21 focused tests
- adversarial test-only RED: 17 pass / 10 expected failures against unchanged production
- corrected GREEN: 27/27 focused tests at `40ce66d4b5010b92089895a05709687143d15a05`

## Cycle 7 verification (historical)

- focused #478 tests: 297 total, 296 pass, 0 fail, 1 intentional live-GitHub skip
- serialized Shepherd tests: environmental failure, 517 total, 451 pass, 65 unchanged unrelated
  managed-sandbox `spawn EPERM` failures, 1 intentional skip
- strict TypeScript 5.9.3 over owned tests/modules and all 20 production modules against cached Pi
  0.80.6: pass
- pinned Pi 0.80.6 offline RPC: `pm-shepherd` discovered from `extension`
- immutable-base ancestry, full-range diff check, and owned-path scope: pass
- Go, connector, certification, runtime-service, and `make` gates: not run by parent policy

## Review and safety

- fake orchestration transports only; no live issue/comment/ready/merge transport ran
- no reviewer was started; the parent owns the stable-head `codex_independent`
  `openai-codex/gpt-5.6-sol:xhigh` campaign
- no Claude, Copilot, parent merge, or default-branch mutation was requested

Refs #478

Refs #471

## Stable-head review correction (in progress)

- accepted 11 functional findings from the deep review of `093b3c90`
- planned one strict artifact-only → test-only RED → architectural GREEN sequence
- correction scope covers authoritative path/integration/CI/session evidence, deterministic review
  selection, keyed idempotency, canonical generations/refs, complete lookups, and partial failures
- no controller/#479 wiring, live GitHub mutation, secret access, Go gate, or merge is included
- fresh exact-head xhigh review remains parent-owned after corrected local verification

Correction evidence: RED `4e02d059` (9 pass / 29 expected fail with production identical to
`093b3c90`), GREEN `8e32896a` (38/38 focused and strict owned/all-production TypeScript pass).
Pinned Pi 0.80.6 offline registration and base/head/diff/scope pass. The serialized suite ran 302
tests but is environmentally blocked at 236 pass / 65 unrelated `spawn EPERM` failures / 1 skip.
Push and live PR-body update are blocked by GitHub DNS resolution in the worker environment.

## Cycle 3 corrected-review slice (in progress)

- accepted both corrected review ledgers as one fourteen-invariant batch against frozen candidate
  `3f285722`
- planned persisted canonical plan/digest validation, durable cross-instance conditional mutation
  ports, full PR/integration/review provenance, independent nested evidence freshness, exact ancestry
  proofs, plan-bound CI policy, monotonic roster CAS, and exported controller attestation helpers
- retained one artifact-only → one all-invariants test-only RED → architectural GREEN → verification
  lifecycle; no controller/#479, parent planning, live GitHub, Go, connector, `make`, or reviewer work
- push and PR synchronization remain deferred because the existing GitHub DNS failure is external
  to the local correction lifecycle

## Cycle 4 consolidated-review correction

- PLAN `607e203e`, single test-only RED `abbf388b`, and architectural GREEN `b92b5ff7` close all
  ten findings from the final two Cycle 3 review ledgers
- stable PR identity is separated from observation evidence; restart/readiness reconstructs the
  authoritative issue-derived child and validates exact current PR/receipt topology
- every external port receives a deadline/`AbortSignal` and normalized bounded errors; policy
  freshness, sensitive text, pseudo refs, CAS progression, dense bounds, and tuple identities fail
  closed
- focused 68/68, strict owned/all-production TypeScript, pinned Pi 0.80.6 offline discovery, and
  base/diff/scope/data gates pass; serialized Shepherd is environmentally blocked only by 65
  unrelated sandbox `spawn EPERM` failures (266 pass, 1 skip)
- no controller/#479, live GitHub, reviewer, network, Go, connector, `make`, or merge action ran;
  two fresh exact-head `xhigh` reviews remain parent-owned

## Cycle 5 consolidated-review correction

- PLAN `7cf9c88d`, comprehensive test-only RED `6cb21902`, and architectural GREEN `3ae10dc2`
  close the unique union of both blocking Cycle 4 review ledgers
- full canonical broker records, generation-wide policy refresh, independent receipt
  reauthorization, centralized child eligibility, revision-bound CAS identity, and expanded
  cookie/session redaction now fail closed
- caller-linked deadlines/cancellation, tracked settlement and abort acknowledgement, live-call
  keyed exclusion, bounded stop/join, byte-bounded raw JSON, and schema-directed record reads
  harden lifecycle and pre-materialization boundaries
- focused 109/109, strict owned/all-production TypeScript, pinned Pi 0.80.6 offline discovery, and
  base/diff/exact-17-path/JSON/credential gates pass; serialized Shepherd records 307 pass, 65
  unrelated sandbox `spawn EPERM` failures, and 1 skip across 373 tests
- post-RED test edits only align support fixtures with the stronger contract; no RED expectation
  was removed or weakened
- no controller/#479, live GitHub, reviewer, network, Go, connector, `make`, integration, or merge
  action ran; two fresh exact-head `xhigh` reviews remain parent-owned

## Cycle 6 consolidated-review correction

- amended PLAN `2832993b`, comprehensive five-test-file RED `ca4d97d1`, and architectural GREEN
  `2c6371e7` close the unique union of both blocking Cycle 5 review ledgers within the exact
  expanded 21-path scope
- the actual `GitHubDecisionBroker` now composes through its own canonical repository reread while
  preserving full request, compact poll, and evidence-only consume shapes; hostile records and
  impossible request/comment/decision/consume chronology fail closed
- parent-ready carries a complete canonical conditional authorization and has a typed idempotent
  verified rollback; intrinsic abort/raw/proxy bounds, ordered stable review authority, one shared
  credential grammar, and receipt chronology are enforced
- focused five-file route records 207 total / 206 pass / 0 fail / 1 intentional skip; strict owned
  and all-production TypeScript plus pinned Pi 0.80.6 offline discovery pass
- serialized Shepherd records 427 total / 361 pass / 65 unchanged unrelated sandbox `spawn EPERM`
  failures / 1 skip; immutable-base/ancestry, exact merge base, full diff, exact 21-path scope, JSON,
  and synthetic credential scans pass
- post-RED test changes only align fixture descriptions, request-comment persistence, canonical
  fake broker/authorization/provenance shapes, and post-authority receipt times; no expectation was
  removed or weakened
- RUN-STATE uses non-circular `HEAD` evidence semantics with exact completed checkpoints; no push,
  network, live GitHub, reviewer, Go, connector, `make`, #479, integration, or merge action ran;
  two fresh exact-head `xhigh` reviews remain parent-owned

## Cycle 7 consolidated-review correction

- PLAN `2c649798`, comprehensive test-only RED `10033bc5`, architectural GREEN `5bab0bc7`,
  timing/chronology REFACTOR proof `87e70401`, audit RED `b1560e76`, and audit GREEN `915882c2`
  close the union of both Cycle 6 exact-head reviews without expanding the exact 21-path boundary
- parent readiness now exposes a public prepare/commit split over one durable authority boundary;
  stable semantic authorization and a separate freshness envelope cover policy, review, paths,
  receipts, ancestry, decision, plan, head, and PR revision; authority is mandatory and the public
  transport has no optional or deprecated ready-mutation route
- uncertain before/after effects remain durably quarantined and keyed until original settlement,
  verified rollback, and retry completion; exact 500 ms effects after 100 ms timeout plus caller
  cancellation are covered
- receipts bind authoritative full review-attempt digest/time provenance; actual broker rereads use
  a controller-owned clock; finite Kubernetes, Docker, and AWS credential forms fail closed; current
  RUN-STATE uses only `HEAD`
- focused five-file route records 297 total / 296 pass / 0 fail / 1 intentional skip; strict owned
  and all-production TypeScript and pinned Pi 0.80.6 offline discovery pass
- the #479 proof composes separate production-typed transport, authority, and journal roles with
  typed compare conflicts; it does not structurally project the test fake as a production role
- serialized Shepherd is honestly classified as environmental failure: 517 total / 451 pass / 65
  unchanged unrelated sandbox `spawn EPERM` failures / 1 skip; base/ancestry, diff, exact 21-path
  scope, JSON, and explicit test-synthetic marker scans pass
- no push, network, live GitHub, reviewer, self-review, Go, connector, `make`, controller/#479
  implementation, integration, or merge action ran; two fresh exact-head `xhigh` reviews remain
  parent-owned

## Cycle 8 consolidated-review correction

- frozen reviewed candidate/base: `b90037df1fff38c755ebc8025579120d17031330` /
  `3addb1f48be1afe8b1e2b59b54247679d7293805`; both blocked reports were read completely
- PLAN `bccee8e6`, comprehensive five-test-file RED `851bb3bf`, coherent GREEN `013bdc8b`, and
  bounded REFACTOR `26a7d476` cover all 48 rows and seven unique families inside the unchanged
  exact 21-path boundary
- provider-neutral assignment suffixes reject before all durable/outbound consumers; finite
  kube/docker/AWS forms remain, and only exact `FEATURE_TOKEN` is allowed after classification
- every uncertain non-value authority result starts durable recovery; rollback attempts use real
  response deadlines and ordered durable fences, may only restore the exact draft, and cannot let
  superseded results settle or release key/stop ownership
- the #479 proof uses exact production return types and reconstructs separate transport, authority,
  journal, broker, and controller adapters over serialized durable state; success, conflict,
  uncertainty, rollback, incomplete/joined stop, and settlement run without `any`, unchecked casts,
  fake projection, private shortcuts, same-object reuse, or module `WeakMap` identity
- the actual broker resumes exact pre-expiry consumed evidence after restart/expiry while rejecting
  new expired events; commit sends refreshed freshness with the original authorization/key/intent
- focused five-file result is 374 total / 373 pass / 0 fail / 1 intentional skip; strict owned and
  all-20-production TypeScript plus pinned Pi 0.80.6 offline discovery pass
- serialized Shepherd is an environmental failure at 594 total / 528 pass / 65 unchanged unrelated
  managed-sandbox `spawn EPERM` failures / 1 intentional skip; exact base/ancestry/diff/21-path/
  JSON/marker gates pass
- GSD doctor passes and unavailable command records `manual_gsd_fallback`; no Go, connector,
  `make`, dependency, parent/main worktree, #475, push/network/GitHub, reviewer, integration, or
  merge action is authorized
