# TDD ledger — #3856 polling conformance corpus

## R1 — mandatory no-skip suite

**RED command (required before implementation):**

```text
go test -timeout 20m ./internal/connectors/engine -run '^TestPollingWatermarkConformanceSuiteRunsEveryMandatoryFixture$' -count=1
```

The test must fail because the polling corpus loader, registration evidence,
lane interface, and no-skip runner do not exist. It must name a reusable lane
factory and assert real transcript/state behavior, not only IDs.

**GREEN target:** the runner validates the registration, drives every embedded
fixture, reports the exact immutable ID set, and proves the first fixture's
equal-watermark split plus post-ack persistence failure safely restarts from
the prior committed #3810 envelope.

## R2 — corpus completeness and immutability

**RED target:** absent, mutable, duplicate, or incomplete fixture metadata
either loads or permits a lane to choose a subset.

**GREEN target:** embedded `v1` JSON has exact digest/version, defensive copies,
unique IDs, every mandatory scenario kind, and no runner skip/filter argument.

## R3 — behavior, state, and admission

**RED target:** a fake lane can silently accept an unsafe cursor/keyset/lag,
erase incompatible checkpoint state, claim a hard delete, or enter with no
registered executor/evidence.

**GREEN target:** every fixture observes source request, acknowledgement,
checkpoint/recovery, history/tombstone, or typed admission outcome as
applicable.

## Run log

Pending R1 RED. Production implementation files do not exist yet.
