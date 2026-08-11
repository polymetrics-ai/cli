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
`upsert` writer and preserves provider record maps. Add descriptor-derived JSON inspection
with explicit unsupported roles and a manual projection when a descriptor is declared.

## Task 5 — verify and review

Run targeted package tests, `-race`, cancellation test, `internal/app`, `internal/cli`,
static/build and connector preflight/help checks. Run the non-`go test ./...` `make verify`
components individually. Complete manual `verify-work`, any `--gaps` loop, and code review.
The child-local delivery command is `no-mistakes axi run --intent <complete issue intent>
--skip=push,pr,ci`, never with `--yes`; it deliberately ends before external push, PR, or CI.
After those child-local gates, the delivery owner may use `gh-axi` to open at most one
conventional-commit sub-PR from this existing child branch to
`feat/3862-any-to-any-transport`. The integrated parent, not this child, owns the full
push/PR/CI pipeline. Never create a second parent/default-branch PR or merge any PR. A
topology restart is not a product correction round. At most five product correction rounds.

## Foundation-check disposition

| Need | Evidence targeted here | Honest status after this child |
| --- | --- | --- |
| Closed dispatch/preflight | Focused fake tests plus real `RunETL` test and inspection checks | Implemented transport-neutral seam only |
| Real conformance | External verifier dependency defaults unavailable; no self-assertion | Blocked on #3810 evidence hardening |
| Real warehouse/apply legs | No provider/database implementation changed | Not implemented; owned elsewhere |
| Reverse delivery | No change; plan → preview → approval → execute remains | Not claimed |
| Certification | No live call or certification metadata changed | Not claimed |

## Correction log

1. **#4021 — authored invalid descriptor fallback (loop 1/5).** An inline review
   found that an empty authored `SyncTransportDescriptor` was treated as absent
   because app dispatch looked only for individual roles. The recorded RED test
   reproduced the legacy/no-descriptor fallback. The correction routes any
   authored descriptor into closed preflight, preserves #3810 semantics, and
   remains inside this child branch. Commit `9775f420c` references `Refs #4021`,
   `Refs #3864`, and `Refs #3862`.
2. **#4023 — normalized generic executor identifiers (loop 2/5).** A distinct
   inline review found that `generic-http` bypassed the same generic-executor
   rejection applied to `generic_http`. Its RED/GREEN test stays exclusively in
   `internal/connectors`; #4021 remains scoped to app routing. This correction
   normalizes hyphen spelling before closed generic-identifier rejection. Commit
   `9775f420c` references `Refs #4023`, `Refs #3864`, and `Refs #3862`.
3. **#4029 — declared `none` acknowledgement inspection (loop 3/5).** Inspection
   must project every structurally valid declared destination, including
   `acknowledgement: none`, without treating that metadata as runtime admission.
   `TestSyncTransportEligibilityProjectsDeclaredNoneAcknowledgement` first failed
   because that role remained `unsupported`; the projection now includes every
   structurally valid destination and retains its acknowledgement. Registry.Preflight's
   `durable_warehouse` execution gate remains unchanged.
4. **#4046, #4045, #4048, #4047, and #4029 — review correction (loop 4/5).**
   RED slices established that an acknowledged page was not persisted before source
   failure, cancellation, or final-state saving returned; full endpoint descriptors
   were not fail-closed when inspection projected a valid sibling; typed-nil verifier
   invocation could panic; binary records aliased provider data; and `connectorsHelp`
   drifted from its generated manual. GREEN persists only the active stream's interim
   checkpoint and generation before the callback returns, full-validates each selected
   descriptor before role access, fails typed nil closed, copies byte backing storage,
   and regenerates help artifacts. It keeps the durable-warehouse execution gate, all
   seven closed modes, fake-only warehouse mediation, and child-local
   `--skip=push,pr,ci` topology intact.
5. **#4046 — checkpoint identity and stale-writer correction (loop 5/5).**
   Distinct RED slices showed that an acknowledged checkpoint could be persisted for another
   credential, stream, or source generation, and that a stale app instance could replace a
   newer target-stream checkpoint. GREEN validates the stamped envelope against the active
   resume expectation before any state mutation, then compares only the captured target stream
   entry under the JSON-store lock. It preserves earlier successful-run metadata during interim
   persistence, advances its expected entry page by page, retains unrelated project updates,
   compares opaque bytes exactly, and keeps the fake warehouse-mediated seven-mode boundary.
