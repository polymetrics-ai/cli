# Verification — #3856 polling conformance corpus

## Current target boundary

The earlier broad-validation record below applies only to
`aa4d8c8a9fabfb230c3de44764f13b945814153d`. The exact broad-validation target
was `97d4f49ada7806a4674e8915a52ccee878e76427`; the evidence-only handoff
commit `a4498eb2dbc5c67dd9049bd9d536d4838a27d510` follows it and changes only
phase evidence, not the polling corpus or runner. The validation target differs
from the passed no-mistakes pipeline head
`82a95c609e3a62f37041c5503b75b1e18aa6dd1c` only by the required #3856
review-evidence reconciliation. Its source target remains
`cd92037fe8add4b6fc3e09c5b0a7648e5a0ab6c5`, which changes conformance
admission and checkpoint provenance. The serial exact-head broad validation
completed successfully on 2026-08-13; its focused RED/GREEN evidence remains
in `TDD-LEDGER.md`.

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
| Exact binary | clean child-head `pm` byte count, SHA-256, and applicable no-credential smoke | pass at `97d4f49`; `148200034` bytes, `1f6e973cd25bc935c84731e0aafbea6a37840622586da93a906dfe6960232033`, `pm help` and bare `pm connectors` pass |
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

## Exact-head broad revalidation — 2026-08-13

Validation began from clean exact head
`97d4f49ada7806a4674e8915a52ccee878e76427`. Each command ran one at a time;
all returned success:

- `go test -timeout 20m ./internal/connectors/engine -run '^(TestPollingWatermarkConformance|TestBoundedOverlapReferenceLaneDerivesTheOverlapRequest)$' -count=1`
- `go test -timeout 20m ./internal/connectors/engine -count=1`
- `go test -timeout 20m ./internal/synccontract -count=1`
- `go test -timeout 20m ./internal/app -count=1`
- `go test -timeout 20m ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1`
- `go test -race -timeout 20m ./internal/connectors/engine -run '^(TestPollingWatermarkConformance|TestBoundedOverlapReferenceLaneDerivesTheOverlapRequest)$' -count=1`
- `gofmt -w internal/connectors/engine/polling_conformance.go internal/connectors/engine/polling_conformance_test.go` (no changes)
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

The generated `pm` binary is `148200034` bytes with SHA-256
`1f6e973cd25bc935c84731e0aafbea6a37840622586da93a906dfe6960232033`; both
applicable no-credential smoke commands above passed. The generic #3810 corpus
digest remains `7366873f91b326d7c35c15d36af74702fe413b9d62906b3fbc3279016a9efaa1`;
the separate polling fixture digest is
`3acf1e9bf13615c5355cc305a705cdcddec5d08ab80ece8024459860bb03e1a4`.

`make smoke-no-build` remains not applicable because it creates local
credentials and runs sample ETL/reverse-ETL flows, which exceed this
test/corpus-only child's authorization. No live provider credentials, calls,
mutations, database integration, public CDC promotion, push, PR, or merge
occurred during this validation.

## Historical serialized validation hold

Historical state before the 2026-08-13 release: #3856 focused implementation
complete; awaiting serialized broad validation gate.

At that point, the captain's resume record permitted focused implementation and
tests only. No broad repository suite, race-heavy sweep, no-mistakes run, push,
PR creation, CI, or merge had started. The exact-head broad result is recorded
above; the focused GREEN evidence and exact commands remain in `TDD-LEDGER.md`.

## Released broad validation

The released, implementation-bearing candidate was
`aa4d8c8a9fabfb230c3de44764f13b945814153d`. The later evidence-only validation
record is owned by [Current target boundary](#current-target-boundary). #4072
remains serialized and untouched.

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
