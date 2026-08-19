# Verification — Issue 4211 provable credential-scope contract

**Status:** passed locally; pending direct-PR automated review

## Required results

- [x] Red guard: unverified report refuses the full-parity construction path.
- [x] Green guard: passed full-parity stage produces a `full_parity` record.
- [x] Green bounded path: a direct-read-like report can publish
  `observed_operations` evidence with protocol-exchange proof.
- [x] Validator rejects missing/mismatched v2 scope proof without removing the
  legacy exact full-parity validator.
- [x] Fourteen fresh PostgreSQL records re-issued with an explicitly bounded
  scope and proof, with no old scope restamped.
- [x] Matrix/sweep generated surfaces are checked twice for byte stability.
- [x] Repository checks and exact results recorded before PR creation.

## Red and green proof

- **Red:** before the contract change,
  `go test -timeout 20m ./cmd/connectorgen -run '^TestCertificationBoundedScopePublishesObservedOperations$' -count=1`
  failed with `completed live test did not use a full-parity credential`.
- **Green:**
  `go test -timeout 20m ./cmd/connectorgen -run '^(TestCertificationBoundedScopePublishesObservedOperations|TestCertificationFullParityScopeRequiresPassedReportStage|TestCertificationPublishesNarrowCredentialEvidence)$' -count=1`
  passed. The full-parity test first receives the missing-stage refusal, then
  proves a passed `full_parity` report produces the only accepted full claim.

## Fresh PostgreSQL re-issue

The original fourteen records were **unverified**, not merely unrecorded: their
live invocations used `--full --write`, while `RequireFullParity` is driven by
`--full-parity`; the full-parity stage was absent, and the importer independently
hard-coded the old caller attestation. They were replaced, not retained.

- 12 transport records: `TestPostgresCertificationProfileRunsBuiltBinaryLive`
  passed in 41.779s.
- 2 CDC records: `TestPMBinaryDispatchesPostgresChangeCaptureToWarehouse`
  passed in 30.230s.
- Both runs used the opt-in `databaseintegration` tag and the documented direct
  Colima Unix endpoint. They wrote all 14 new v2 records.
- `jq` field audit found exactly one tuple for every record and every generated
  matrix pointer: `observed_operations`, `protocol_exchanges`, and the explicit
  no-broader-scope note.

## Commands and results

All following commands passed unless stated otherwise:

```text
go test -timeout 20m ./cmd/connectorgen
go test -timeout 20m ./internal/connectors/certify
go test -timeout 20m ./internal/cli
go test -timeout 20m ./internal/agentcontract
go vet ./...
make fmt
make tidy-check
make build
make docs-check
make smoke-no-build
make lint
make connectorgen-validate
make connectorgen-surface-sync
make github-parity-artifacts-check
make connectorgen-certification-matrix
make connectorgen-certification-sweep
make connector-boundary
make connector-canon-check
make release-workflow-check
go run ./cmd/agentcontractgen check
go run ./cmd/connectorgen certification-matrix --check
```

`make verify` was deliberately not run as one process: the repository instruction
for agents under a per-command timeout prohibits that aggregate because it embeds
the full 550+ connector suite and is routinely cut off. Its non-suite gates above
and all packages changed by this PR (including `./cmd/connectorgen`) were run
individually with `-timeout 20m`. `security/snyk` remains the identical known
base-branch failure and was not attributed to this change.
