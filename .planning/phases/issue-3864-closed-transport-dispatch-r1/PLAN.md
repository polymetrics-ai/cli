# #3864 plan — closed transport dispatch (TDD)

## Scope and file ownership

This child owns only the following new/changed paths. It will not edit provider bundles,
native connector implementations, `internal/synccontract/**`, or sibling issue areas.

| Area | Paths |
| --- | --- |
| Closed descriptor/projection | `internal/connectors/sync_transport.go` (new), `internal/connectors/definition.go`, `internal/connectors/guide.go`, `internal/connectors/sync_transport_test.go` (new) |
| Transport-neutral runtime | `internal/synctransport/types.go` (new), `internal/synctransport/registry.go` (new), `internal/synctransport/orchestrator.go` (new), their focused tests |
| Bounded app dispatch | `internal/app/app.go`, `internal/app/transport_dispatch.go` (new), `internal/app/transport_dispatch_test.go` (new) |
| Inspection/help | `internal/cli/cli.go`, focused existing/new CLI tests, `docs/cli/connectors.md`, `website/content/docs/agent-guide.mdx` |
| Delivery evidence | `.planning/phases/issue-3864-closed-transport-dispatch-r1/**` |

## Task 1 — RED: characterize the hard-coded failure

Add `TestRunETLCanonicalFullAppendUsesRegisteredTransports` using only fake API and
database connectors/executors, a fake external conformance verifier, and a fake warehouse
stage. It must initially receive the existing `ModeNotExecutableError`, proving that
canonical modes do not reach a registered source/destination path. Add narrow negative
preflight tests for missing executor, family mismatch, unsupported mode, unsafe
acknowledgement, and absent apply strategy; add a strategy test that fails if `upsert` is
chosen instead of the descriptor declaration.

Run each named test and record its actual RED output in `TDD-LEDGER.md`.

## Task 2 — GREEN: closed descriptor and registry

Add closed executor-family, role, delivery, acknowledgement, strategy, and conformance
reference descriptor types. Validate identifiers and all closed values. Add source and
destination registries guarded by a mutex. Preflight resolves only exact registered refs,
checks integration type/family compatibility, descriptor eligibility/mode/strategy, and
asks a supplied external conformance verifier; it never trusts evidence returned by an
executor. Its default verifier rejects admission as unavailable.

## Task 3 — GREEN: one transport-neutral orchestrator

Define small consumer-owned source, warehouse-stage, and destination interfaces. The
orchestrator obtains the descriptor-resolved strategy, obtains bounded source pages, copies
provider records into the typed stage, plans/applies the resulting workset, validates the
opaque #3810 durable acknowledgement, and calls `CommitAfterDownstreamAcknowledgement`.
It checks cancellation between stages and commits no checkpoint on stage/apply/ack errors.

## Task 4 — GREEN: app dispatch and public inspection

Wire an app-local transport registry/orchestrator. Move canonical-mode admission until after
source/destination resolution and route a connection only if it has transport descriptor(s);
the registry rejects incomplete sides before `Read`. Preserve legacy routes when both sides
have no descriptor. The new path accepts declared strategy output rather than the legacy
`upsert` writer and preserves provider record maps. Add descriptor-derived JSON/manual
inspection; absent descriptors are explicit unsupported roles.

## Task 5 — verify and review

Run targeted package tests, `-race`, cancellation test, `internal/app`, `internal/cli`,
static/build and connector preflight/help checks. Run the non-`go test ./...` `make verify`
components individually. Complete manual `verify-work`, any `--gaps` loop, code review,
and `no-mistakes axi run --intent ... --skip=push,pr,ci` without `--yes`. At most five
combined correction rounds. Only then push and open a child PR to
`feat/3862-any-to-any-transport`.

## Foundation-check disposition

| Need | Evidence targeted here | Honest status after this child |
| --- | --- | --- |
| Closed dispatch/preflight | Focused fake tests plus real `RunETL` test and inspection checks | Implemented transport-neutral seam only |
| Real conformance | External verifier dependency defaults unavailable; no self-assertion | Blocked on #3810 evidence hardening |
| Real warehouse/apply legs | No provider/database implementation changed | Not implemented; owned elsewhere |
| Reverse delivery | No change; plan → preview → approval → execute remains | Not claimed |
| Certification | No live call or certification metadata changed | Not claimed |
