# Verification — #3856 polling conformance corpus

## Current target boundary

The passing broad-validation record below applies only to
`aa4d8c8a9fabfb230c3de44764f13b945814153d`. The current source target
`cd92037fe8add4b6fc3e09c5b0a7648e5a0ab6c5` changes conformance admission and
checkpoint provenance; evidence refresh committed at
`82a95c609e3a62f37041c5503b75b1e18aa6dd1c`. Its focused RED/GREEN evidence is
recorded in `TDD-LEDGER.md`; the outer validation phase must re-run broad
validation before this record is used as current-head certification.

## Broad validation record

| Area | Command/check | Status |
| --- | --- | --- |
| R1 RED/GREEN | focused conformance test before/after implementation | pass; the mandatory RED and focused GREEN are recorded in `TDD-LEDGER.md` |
| Polling engine | focused and full `internal/connectors/engine` tests | pass |
| Generic contract non-regression | `internal/synccontract` tests and unchanged `v1.json` digest | pass; generic digest remains `7366873f91b326d7c35c15d36af74702fe413b9d62906b3fbc3279016a9efaa1` |
| App regression | `go test -timeout 20m ./internal/app -count=1` | pass |
| Runtime preflight sweep | focused commandrunner test | pass |
| Race safety | focused engine `-race` test | pass |
| Format/vet/build | scoped `gofmt`, `go vet ./...`, `go build ./cmd/pm` | pass |
| Repository gates | individual AGENTS.md gates | pass; the credential-mutating smoke target is explicitly not applicable |
| Exact binary | clean child-head `pm` byte count, SHA-256, and applicable no-credential smoke | pass; `148200034` bytes, `cd1744c464daca2c4e73d1eaa8869c3aa85494c895fe3a381fd0876647ede0e5`, `pm help` and bare `pm connectors` pass |
| GSD verify/review | generated prompt, durable verification/review report | pass using the documented inline/manual fallback; see `3856-UAT.md` and `REVIEW.md` |
| Independent audit | required Sol audit after final candidate | pending |
| no-mistakes | `--skip=push,pr,ci`, no `--yes` | pass at `82a95c60`; review, focused test/race, document, and lint passed while push/PR/CI were intentionally skipped |

## No-mistakes correction evidence

Run `01KZVGR0A7Y9ZQ31HK4AYK9DGX` passed at pipeline head
`82a95c609e3a62f37041c5503b75b1e18aa6dd1c`. The #4075 causal RED command was:

```text
$ go test -timeout 20m ./internal/connectors/engine -run '^(TestPollingWatermarkConformanceRegistrationRejectsUnsafeDescriptor|TestPollingWatermarkConformanceSuiteRejectsPersistedCheckpointDescriptorMismatch)$' -count=1
FAIL: unsafe keyset, cursor policy, bounded overlap, and bounded commit lag registrations returned <nil>
FAIL: source identity, source generation, schema fingerprint, and mechanism checkpoint mutations returned <nil>
```

After the shared-runner correction, the gate passed:

- `go test -v -timeout 20m ./internal/connectors/engine -run '^(TestPollingWatermarkConformance.*|TestBoundedOverlapReferenceLaneDerivesTheOverlapRequest)$' -count=1`
- the same target with `-race`
- separate polling corpus and generic #3810 SHA-256 evidence; the generic digest remains unchanged

The gate created the no-skip runner transcript and eleven-scenario inventory in
its required temporary evidence directory. Push, PR, and CI were deliberately
skipped; no remote branch, PR, or CI state was created by this run.

## Serialized validation hold

paused: #3856 focused implementation complete; awaiting serialized broad validation gate

The captain's resume record permits focused implementation and tests only at
this point. No broad repository suite, race-heavy sweep, no-mistakes run,
push, PR creation, CI, or merge has been started. The focused GREEN evidence
and exact commands are recorded in `TDD-LEDGER.md`.

## Released broad validation

The released, implementation-bearing candidate was
`aa4d8c8a9fabfb230c3de44764f13b945814153d`. The evidence-only commit that
updates this record does not alter the polling corpus or runner. #4072 remains
serialized and untouched.

The following commands completed successfully:

- `go test -timeout 20m ./internal/connectors/engine -run '^(TestPollingWatermarkConformance|TestBoundedOverlapReferenceLaneDerivesTheOverlapRequest)$' -count=1`
- `go test -timeout 20m ./internal/connectors/engine -count=1`
- `go test -timeout 20m ./internal/synccontract -count=1`
- `go test -timeout 20m ./internal/app -count=1`
- `go test -timeout 20m ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1`
- `go test -race -timeout 20m ./internal/connectors/engine -run '^(TestPollingWatermarkConformance|TestBoundedOverlapReferenceLaneDerivesTheOverlapRequest)$' -count=1`
- `gofmt -w internal/connectors/engine/polling_conformance.go internal/connectors/engine/polling_conformance_test.go`
- `git diff --check origin/feat/3855-polling-apply-foundations...HEAD`
- `go vet ./...`
- `make tidy-check`
- `make lint`
- `go build ./cmd/pm`
- `./pm help`
- `./pm connectors`
- `make docs-check-no-build`
- `make agent-contract-check`
- `make connectorgen-validate`
- `make connectorgen-surface-sync`
- `make github-parity-artifacts-check`
- `make connectorgen-certification-matrix`
- `make connector-boundary`
- `make connector-canon-check`
- `make connector-runtime-preflight`
- `make release-workflow-check`
- `scripts/verify-gsd-workflow origin/feat/3855-polling-apply-foundations`

`make smoke-no-build` is deliberately not run: it creates local credentials and
performs sample ETL/reverse-ETL flows, which are outside this test/corpus-only
child's authorization. Live credentials, provider calls, provider mutation,
database integration, and public CDC promotion are likewise not applicable.

The independent inline review found and corrected #4074 before this run: the
bounded-overlap reference lane now derives its lower request from the durable
checkpoint rather than copying the fixture expectation. The subsequent
`make lint` S1016 cleanup was a direct equivalent struct conversion, not a
production correction. The correction ledger is therefore 2/5. GSD
`verify-work` and `code-review` source resolution and generated prompts were
executed; the repository's canonical single-worker contract forbids spawning
their agent roles, so `3856-UAT.md` and `REVIEW.md` record the documented
inline/manual result. Shepherd #3995 / PR #4062 remains excluded.

## Explicit exclusions

Shepherd #3995 / PR #4062 is parked and excluded. No provider, credential,
database, live mutation, CDC promotion, or certification test occurs in this
corpus-only phase. Credential-backed Transport is deferred to the first
executable provider-boundary child and final #4019 gate.
