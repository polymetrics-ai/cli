## Summary

- add a bounded, typed GitHub parent/child orchestration transport with reconcile-before-mutate
  issue, stacked-PR, roster, integration, and parent-ready operations
- validate authoritative CI, review-thread, requested-change, finding-disposition, and exact-range
  independent Codex review evidence
- reuse dependency scheduling, autonomy reconciliation, workspace handoffs, and the existing human
  decision broker while keeping review execution and parent merge outside this slice

## GSD / TDD

- mode: `manual_gsd_fallback` because the healthy repo adapter does not expose
  `programming-loop`
- required skills: `gsd-programming-loop`, `github-issue-first-delivery`, `gsd-workstreams`,
  `architecture-patterns`, `javascript-testing-patterns`
- initial test-only RED: 0 pass / 3 absent-module failures
- minimal GREEN: 21/21 focused tests
- adversarial test-only RED: 17 pass / 10 expected failures against unchanged production
- corrected GREEN: 27/27 focused tests at `40ce66d4b5010b92089895a05709687143d15a05`

## Verification

- focused #478 tests: 27/27 pass
- serialized Shepherd tests: 290 pass, 0 fail, 1 intentional sandbox skip
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
