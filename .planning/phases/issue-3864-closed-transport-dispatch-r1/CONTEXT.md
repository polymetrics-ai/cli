# #3864 — Closed source/destination transport dispatch

**Issue:** [#3864](https://github.com/polymetrics-ai/cli/issues/3864)\
**Parent:** [#3862](https://github.com/polymetrics-ai/cli/issues/3862) / draft [PR #4019](https://github.com/polymetrics-ai/cli/pull/4019)\
**Child branch:** `feat/3864-closed-transport-dispatch`\
**Base:** `origin/feat/3862-any-to-any-transport`\
**Status:** active manual-GSD fallback; correction-loop 5 focused evidence is recorded.

## Lifecycle fallback

`scripts/gsd doctor`, `scripts/gsd sources`, and `go run ./cmd/agentcontractgen check`
passed before planning. The adapter resolved these official commands:

```text
/gsd-discuss-phase issue-3864-closed-transport-dispatch-r1
/gsd-plan-phase issue-3864-closed-transport-dispatch-r1 --tdd
/gsd-execute-phase issue-3864-closed-transport-dispatch-r1
/gsd-verify-work issue-3864-closed-transport-dispatch-r1
/gsd-code-review issue-3864-closed-transport-dispatch-r1
```

The supplied issue key is not a numbered roadmap phase, so the official Pi workflow
cannot create its normal phase directory. The canonical delivery contract also forbids
role spawning here. This directory is the explicit inline/manual GSD fallback: it
records the discuss, plan/TDD, execution, verification/gap, review, and no-mistakes
evidence in the same order. This is a tooling/topology fallback, not a lifecycle
waiver.

## Binding inputs and decisions

1. #3810 owns the exact seven `synccontract.Mode` values, checkpoint envelope,
   tombstone/history/recovery types, and durable acknowledgement gate. #3864 imports
   and passes those values through; it must not add a mode, restate checkpoint
   semantics, fabricate a durable acknowledgement, or change #3810 source files.
2. The topology scout report
   `data/cli-sync-transport-topology-scout-r1/report.md` establishes that the
   #3810 corpus is metadata-only. A transport must not become executable because an
   executor returns `RequiredConformanceEvidence()` or any other self-reported
   digest/fixture list. The #3864 registry therefore uses an externally supplied
   conformance verifier and defaults closed when one is unavailable. Fake tests use
   a separate test verifier, never executor self-assertion. The missing evidence
   hardening remains a #3810-owned prerequisite for real admission.
3. Connector canon requires warehouse mediation. The common fake dispatch path has
   a typed warehouse-stage seam between source and destination; it never calls one
   provider transport from another. No live source/destination transport, provider
   adapter, database protocol/DDL, or local database call is added by this issue.
4. A descriptor is closed data: source/destination executor reference, eligible
   streams/actions, supported canonical modes, delivery declarations, an
   externally-verifiable conformance reference, and per-mode destination apply
   strategy. `integration_type` determines the only eligible executor families:
   API (`declarative_api` or `native_api`), database (`native_database`), file
   (`file`), queue (`queue`). Generic SQL/HTTP/shell and generic action execution
   are rejected.
5. Runtime preflight occurs before a transport source can read. It rejects an absent
   descriptor/executor, family/type mismatch, unadvertised mode/stream/action,
   missing externally verified conformance, unsafe acknowledgement policy, or a
   missing declared strategy. The existing legacy warehouse and connector paths stay
   available only when neither side declares a transport descriptor.
6. The new orchestrator resolves its destination apply strategy from the descriptor
   for the requested #3810 mode. It does not use the legacy hard-coded `upsert`
   path and does not inject `_polymetrics_*` metadata into provider records. It
   advances a candidate checkpoint only through
   `synccontract.CommitAfterDownstreamAcknowledgement` after the destination's
   durable acknowledgement.
7. Inspection is metadata-only. `pm connectors inspect <name> --json` always
   projects source/destination eligibility and reports missing descriptors as
   `unsupported`; every structurally valid declared destination remains
   `declared`, including `acknowledgement: none`. A rendered connector manual adds
   the declared-role projection when a descriptor exists. The help and website docs
   say this is not a claim that a transport has passed conformance or is certified.
   Runtime preflight remains closed unless acknowledgement is `durable_warehouse`.

## Explicit non-goals

- Provider-specific GitHub/harness work, live calls, database drivers/protocols/DDL,
  and #3859 database apply implementation.
- #3865 authentication cohorts/fencing; #3867 rate parking; #3866 full four-pair ×
  seven-mode test matrix.
- A generic HTTP, SQL, shell, record-action, or direct source-to-destination escape
  hatch.
- A real #3810 conformance corpus/evidence implementation or any certification claim.

## Required skills used

`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
`golang-security`, `golang-safety`, `golang-documentation`,
`golang-design-patterns`, `golang-structs-interfaces`, `golang-context`,
`golang-concurrency`, `golang-database`, `golang-lint`, `golang-naming`,
`golang-code-style`, `github-issue-first-delivery`, `gsd-discuss-phase`,
`gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, `gsd-code-review`, and
`no-mistakes`. The canonical child-local gate is `no-mistakes axi run --intent <complete issue
intent> --skip=push,pr,ci`, never with `--yes`; the outer delivery owner alone handles later
push, sub-PR creation to `feat/3862-any-to-any-transport`, and CI. It must never create another
parent/default-branch PR or merge. Product corrections remain in this one child: #4021 owns only
app's empty-authored-descriptor fallback (loop 1/5), #4023 owns only normalized generic executor
identifiers (loop 2/5), and #4029 owns the declared `none` acknowledgement projection (loop 3/5).
Loop 4/5 retains this one child: #4046 owns acknowledged interim checkpoint persistence, #4045
owns complete-descriptor runtime validation, #4048 owns typed-nil verifier admission, #4047 owns
binary record isolation, and #4029 also owns help/manual parity. A topology restart is not a
product correction loop. Loop 5/5 remains under #4046: acknowledgement-stamped checkpoints must
match the active resume identity before persistence, and only the captured stream entry may advance
through a compare-and-swap that preserves prior completion metadata and rejects stale writers.
