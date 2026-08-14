---
coverage:
  - id: D1
    description: Canonical closed transport dispatch uses one warehouse-mediated orchestrator for all fake API/database pairings.
    verification:
      - kind: unit
        ref: internal/app/transport_dispatch_test.go: TestRunETLCanonicalFullAppendUsesRegisteredTransports
        status: pass
      - kind: unit
        ref: internal/synctransport/transport_test.go: TestOrchestratorDispatchesFourClosedPairingsWithoutPairBranches
        status: pass
    human_judgment: false
  - id: D2
    description: Closed preflight and acknowledgement boundaries fail before source reads or checkpoint advancement.
    verification:
      - kind: unit
        ref: internal/synctransport/transport_test.go: TestPreflightRejectsClosedAdmissionFailuresBeforeSourceRead
        status: pass
      - kind: unit
        ref: internal/synctransport/transport_test.go: TestOrchestratorCommitsOnlyAfterDurableAcknowledgement
        status: pass
      - kind: unit
        ref: internal/synctransport/transport_test.go: TestOrchestratorStopsOnCancellationBetweenWarehouseStageAndApply
        status: pass
    human_judgment: false
  - id: D3
    description: Inspection and connector help expose closed-transport eligibility without claiming certification.
    verification:
      - kind: unit
        ref: internal/cli/sync_transport_cli_test.go: TestConnectorInspectProjectsUnsupportedSyncTransport
        status: pass
      - kind: other
        ref: freshly built pm binary help and inspect probe recorded in VERIFICATION.md
        status: pass
    human_judgment: false
  - id: D4
    description: Closed generic executor identifiers fail consistently across underscore and hyphen spelling.
    verification:
      - kind: unit
        ref: internal/connectors/sync_transport_test.go: TestSyncTransportDescriptorRejectsHyphenatedGenericExecutorReference
        status: pass
    human_judgment: false
  - id: D5
    description: Inspection reports every structurally valid destination acknowledgement while runtime preflight remains durable-warehouse-only.
    verification:
      - kind: unit
        ref: internal/connectors/sync_transport_test.go: TestSyncTransportEligibilityProjectsDeclaredNoneAcknowledgement
        status: pass
      - kind: unit
        ref: internal/synctransport/transport_test.go: TestPreflightRejectsClosedAdmissionFailuresBeforeSourceRead
        status: pass
    human_judgment: false
---

# Summary — #3864 closed transport dispatch

## Delivered

- Added a transport-neutral, closed descriptor and executor registry with source/destination
  roles, integration-family admission, mode/stream/strategy checks, and an external conformance
  verifier seam that defaults to unavailable.
- Added one warehouse-mediated orchestrator used by fake API→API, API→database, database→API,
  and database→database test pairings. It resolves the descriptor-owned destination strategy,
  stages bounded pages, applies only after planning, and calls the #3810 acknowledgement/commit
  seam without altering #3810 semantics.
- Added the bounded `App.RunETL` dispatch bridge and metadata-only inspect/help projections.
  No existing connector is promoted to a transport executor, and no provider/database protocol or
  live call is introduced.
- Recorded five independently scoped review corrections: [#4021](https://github.com/polymetrics-ai/cli/issues/4021)
  for empty authored descriptors reaching preflight, and
  [#4023](https://github.com/polymetrics-ai/cli/issues/4023) for normalized generic executor IDs,
  plus [#4029](https://github.com/polymetrics-ai/cli/issues/4029) for declared `none`
  acknowledgement inspection. The latter preserves honest metadata without relaxing the
  durable-warehouse runtime gate. Loop 4/5 adds [#4046](https://github.com/polymetrics-ai/cli/issues/4046)
  interim checkpoint persistence, [#4045](https://github.com/polymetrics-ai/cli/issues/4045)
  complete-descriptor preflight, [#4048](https://github.com/polymetrics-ai/cli/issues/4048)
  typed-nil verifier closure, [#4047](https://github.com/polymetrics-ai/cli/issues/4047)
  binary record isolation, and #4029 help/manual parity. These corrections stay in the one child;
  loop 5/5 keeps [#4046](https://github.com/polymetrics-ai/cli/issues/4046) in that same child to
  validate acknowledgement-stamped checkpoint identity and atomically advance only the expected
  stream entry, preserving prior completion metadata and rejecting stale writers. The outer
  delivery owner will record the correction commit separately.

## Manual-GSD fallback

The canonical phase worker runtime cannot create compatible isolated roles for this non-numbered
issue phase. The required `discuss-phase → plan-phase --tdd → execute-phase → verify-work →
code-review` sequence is therefore recorded inline in this phase directory. TDD RED/GREEN output
is in `TDD-LEDGER.md`; local verification is in `VERIFICATION.md`; review disposition is in
`REVIEW.md`.

## Honest boundary

This summary records fake-backed dispatch and local behavior only. It does not assert accepted
#3810 conformance, a real provider/database leg, live credentials, automatic Shepherd acceptance,
or certification. The required child-local gate remains pending as
`no-mistakes axi run --intent <complete issue intent> --skip=push,pr,ci`; the outer delivery owner
alone handles any later push, stacked sub-PR, and CI work.
