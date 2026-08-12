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

### R1 RED recorded before production edits

The test file existed before `polling_conformance.go`, its embedded corpus, or
any runner implementation. The first required command failed as intended:

```text
$ go test -timeout 20m ./internal/connectors/engine -run '^TestPollingWatermarkConformanceSuiteRunsEveryMandatoryFixture$' -count=1
# polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
internal/connectors/engine/polling_conformance_test.go:15:17: undefined: RunPollingWatermarkConformanceSuite
internal/connectors/engine/polling_conformance_test.go:17:3: undefined: newReferencePollingWatermarkConformanceLaneFactory
internal/connectors/engine/polling_conformance_test.go:23:13: undefined: RequiredPollingWatermarkConformanceEvidence
internal/connectors/engine/polling_conformance_test.go:27:30: undefined: RunPollingWatermarkConformanceSuite
internal/connectors/engine/polling_conformance_test.go:38:51: undefined: PollingWatermarkConformancePosition
FAIL    polymetrics.ai/internal/connectors/engine [build failed]
FAIL
```

This RED proves the required reusable factory, corpus evidence, no-skip runner,
and composite position are absent. No production implementation file existed
when it was recorded.
