---
coverage:
  - id: D1
    description: The four API/database family half-paths carry exact values through the warehouse seam for every canonical mode.
    verification:
      - kind: unit
        ref: internal/synctransport/family_conformance_test.go:TestTransportFamilyHalfPathConformance
        status: pass
    human_judgment: false
  - id: D2
    description: Shared transport refuses an unbound source before I/O and never commits without a durable acknowledgement.
    verification:
      - kind: unit
        ref: internal/synctransport/family_conformance_test.go:TestTransportFamilyHalfPathConformanceRefusesUnboundSourceBeforeIO
        status: pass
      - kind: unit
        ref: internal/synctransport/family_conformance_test.go:TestTransportFamilyHalfPathConformanceAcknowledgementAndCancellation
        status: pass
    human_judgment: false
  - id: D3
    description: Verified-auth fencing, rate parking/resume, durable conflict, cancellation, and restart preserve shared transport safety.
    verification:
      - kind: unit
        ref: internal/coordination/transport_conformance_test.go:TestSharedTransportCoordinationConformance
        status: pass
      - kind: unit
        ref: go test -race -count=1 -timeout 20m ./internal/synctransport ./internal/coordination
        status: pass
    human_judgment: false
---

# Summary: Issue #3866 shared transport family conformance

Implemented a deterministic, fixture-only conformance matrix at the shared
transport seam. It proves API/database source and destination *families* as
four named warehouse half-paths. Every `synccontract.AllModes()` member is
executed with exact records, descriptor strategy, acknowledgement, and
checkpoint assertions.

The companion coordination test covers only secret-free, in-memory auth/rate
state. It verifies typed pre-send refusals, durable conflict handling, exact
restart checkpoint resumption without replay, and cancellation.

This is simulated CI coverage. It neither calls a provider/database nor claims
live certification. PR #4195 remains the distinct deterministic proof for the
four registered GitHub/PostgreSQL connector routes and sealed-workset ordering.
