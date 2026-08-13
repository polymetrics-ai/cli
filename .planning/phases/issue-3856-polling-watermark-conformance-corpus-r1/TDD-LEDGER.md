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

### R2/R3 focused GREEN and correction evidence

The embedded `v1` corpus now has eleven immutable fixture IDs, its own
SHA-256 (`3acf1e9bf13615c5355cc305a705cdcddec5d08ab80ece8024459860bb03e1a4`),
defensive copies (including raw cursor values), full fixture-owned enumeration,
and no filter or skip input. It remains separate from the generic #3810 corpus.

Focused RED/GREEN corrections:

```text
$ go test -timeout 20m ./internal/connectors/engine -run '^TestPollingWatermarkConformanceSuiteNeverRegressesDurableCheckpointForOverlap$' -count=1
FAIL: bounded overlap regressed durable checkpoint to 2026-08-06T09:59:00Z/late,
want the prior committed 2026-08-06T10:00:00Z/a

$ go test -timeout 20m ./internal/connectors/engine -run '^TestPollingWatermarkConformanceSuiteNeverRegressesDurableCheckpointForOverlap$' -count=1
ok   polymetrics.ai/internal/connectors/engine

$ go test -timeout 20m ./internal/connectors/engine -run '^TestPollingWatermarkConformanceSuiteRejectsUntypedRecoveryObservation$' -count=1
FAIL [build failed]: PollingWatermarkConformanceObservation lacked RecoveryError

$ go test -timeout 20m ./internal/connectors/engine -run '^TestPollingWatermarkConformanceSuiteRejectsUntypedRecoveryObservation$' -count=1
ok   polymetrics.ai/internal/connectors/engine

$ go test -timeout 20m ./internal/connectors/engine -run '^TestPollingWatermarkConformanceSuiteRejectsCursorPolicyResultMismatch$' -count=1
FAIL [build failed]: PollingWatermarkConformanceObservation lacked CursorSampleResults

$ go test -timeout 20m ./internal/connectors/engine -run '^TestPollingWatermarkConformanceSuiteRejectsCursorPolicyResultMismatch$' -count=1
ok   polymetrics.ai/internal/connectors/engine
```

The GREEN runner checks concrete source requests, raw NULL/precision/coercion
results, typed rebootstrap recovery, durable checkpoint positions and commits,
replay state, tombstone/history state, hard-delete invisibility, and admission
rejection. It uses the merged #3880 lossless scalar and timestamp algorithms
from tests without changing their history.

Focused results before the serialized validation gate:

```text
$ go test -timeout 20m ./internal/connectors/engine -run '^TestPollingWatermarkConformance' -count=1
ok   polymetrics.ai/internal/connectors/engine

$ go test -timeout 20m ./internal/connectors/engine -count=1
ok   polymetrics.ai/internal/connectors/engine

$ go test -timeout 20m ./internal/synccontract -run '^TestConformanceFixturesAreVersionedAndDefensivelyCopied$' -count=1
ok   polymetrics.ai/internal/synccontract
```

`gh-axi issue subissue list 3856` returned zero existing children. The
overlap correction stays within the uncommitted primary #3856 scope under the
resume record's no-new-issue custody rule.

paused: #3856 focused implementation complete; awaiting serialized broad validation gate

### R3 correction loop 2/5 — #4074 bounded-overlap derivation

The released GSD inline code review found that the reference lane copied the
fixture's expected overlap request instead of deriving it from the durable
checkpoint and replayed source record. `gh-axi issue subissue list 3856` was
empty, so the released correction protocol created and linked #4074 before the
test change.

```text
$ go test -timeout 20m ./internal/connectors/engine -run '^TestBoundedOverlapReferenceLaneDerivesTheOverlapRequest$' -count=1
FAIL: runBoundedOverlapCommitLag error = <nil>, want copied expectation rejection

$ go test -timeout 20m ./internal/connectors/engine -run '^TestBoundedOverlapReferenceLaneDerivesTheOverlapRequest$' -count=1
ok   polymetrics.ai/internal/connectors/engine
```

The lane now derives the lower overlap request with an empty tie breaker,
requires its timestamp to precede the durable checkpoint, and rejects a corpus
expectation that does not match that derived request.

### Validation cleanup — staticcheck (non-substantive)

`make lint` reported staticcheck S1016 on the reference factory's equivalent
struct construction. The factory now uses the direct Go conversion; the focused
conformance target and a rerun of `make lint` pass. This style-only cleanup does
not increment the substantive correction count (still 2/5).

### R3 correction loop 3/5 — #4075 shared admission and checkpoint provenance

The subsequent review confirmed two shared-runner false-certification gaps: a
registered lane could declare an unsafe stable-keyset/cursor/overlap/commit-lag
contract, and a structurally valid persisted envelope could name another
source, generation, schema fingerprint, or mechanism. Fixture descriptors
remain permissive enough to model rejected cases; registration is now the
separate fail-closed safety boundary. Persisted envelopes are checked against
the fixture descriptor through `ValidateResume`, schema fingerprint, and
mechanism before their expected position is accepted.

```text
$ go test -timeout 20m ./internal/connectors/engine -run '^(TestPollingWatermarkConformanceRegistrationRejectsUnsafeDescriptor|TestPollingWatermarkConformanceSuiteRejectsPersistedCheckpointDescriptorMismatch)$' -count=1
FAIL: unsafe keyset, cursor policy, bounded overlap, and bounded commit lag registrations returned <nil>
FAIL: source identity, source generation, schema fingerprint, and mechanism checkpoint mutations returned <nil>
```

After the shared validation corrections:

```text
$ go test -v -timeout 20m ./internal/connectors/engine -run '^(TestPollingWatermarkConformance.*|TestBoundedOverlapReferenceLaneDerivesTheOverlapRequest)$' -count=1
ok   polymetrics.ai/internal/connectors/engine

$ go test -timeout 20m ./internal/connectors/engine -count=1
ok   polymetrics.ai/internal/connectors/engine

$ go vet ./internal/connectors/engine
```

The post-fix no-mistakes test gate independently passed the same complete
focused runner target and its `-race` variant at `cd92037fe`; its no-skip
runner transcript, scenario inventory, and separate polling/#3810 SHA-256
evidence were captured under the gate's required temporary evidence directory.
Push, PR, and CI were intentionally skipped by that run.
