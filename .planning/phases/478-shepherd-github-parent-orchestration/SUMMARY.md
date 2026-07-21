# Summary: #478

Status: ready stacked PR #487 open; parent-owned exact-head review pending.

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
