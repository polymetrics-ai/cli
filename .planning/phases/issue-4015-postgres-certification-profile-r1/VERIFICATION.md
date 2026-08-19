# Verification checklist — PostgreSQL certification profile

## Executed focused checks

- [x] `go test -timeout 20m ./internal/connectors/engine -run '^TestBundleLoadEmbeddedPostgresCertification$'` — pass.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/postgres` — pass; it also passed while the scratch-invalid runtime `sslmode=bananas` was installed, proving schema compilation alone cannot certify behavior.
- [x] `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -timeout 20m -tags=databaseintegration ./internal/connectors/native/postgres -run '^TestPostgresCertificationProfileRunsBuiltBinaryLive$' -count=1` — green after restoration; red with scratch `sslmode=bananas` (`pm` exit 2).
- [x] `go run ./cmd/connectorgen certification-matrix --connector postgres` twice — first regenerated the prior stale `incremental_dedupe_history` declared/implemented flags; second was byte-stable.
- [x] `go test -timeout 20m ./internal/connectors/certify -run 'Test(PostgresPollingCertification|DeclaredWriteActionNamesTreatsAbsent|PostgresProfileSkips|CertificationDeclaredTransportPair)' -count=1` — pass; covers exact dynamic-watermark selection, no direct write inventory, profile direct-write skip, and both declared transport outcomes.
- [x] `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -v -timeout 20m -tags=databaseintegration ./internal/connectors/native/postgres -run '^TestPostgresCertificationProfileRunsBuiltBinaryLive$' -count=1` — pass: built binary, catalog, typed read, all six managed-target modes, independent read-back, and source immutability.
- [x] Same command with `POLYMETRICS_WRITE_POSTGRES_CERTIFICATION_EVIDENCE=1` — pass and wrote 12 immutable proof-bearing records.
- [x] `go run ./cmd/connectorgen certification-matrix --connector postgres` twice, `cmp -s` of the two generated shards, and `go run ./cmd/connectorgen certification-matrix --check` — pass; exact matrix cells are live-marked and byte-stable.
- [x] `go test -timeout 20m ./cmd/connectorgen -count=1` — pass (227.026s); required generator consumer package for the profile, evidence, and matrix.
- [x] `go test -timeout 20m ./internal/app -count=1` — pass (312.139s).
- [x] `go test -timeout 20m ./internal/connectors/certify -count=1` — pass (10.676s).
- [x] `go test -timeout 20m ./internal/connectors/engine -count=1` — pass (6.421s).
- [x] `go test -timeout 20m ./internal/connectors/native/postgres -count=1` — pass (1.417s); the tagged live test above passed again after the definition-owned selector repair (31.67s).
- [x] `go test -timeout 20m ./internal/cli -count=1` — pass (690.336s) after regenerating the changed connector-help golden transcript with `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$' -count=1`.
- [x] `go build ./cmd/pm`, `./pm help connectors`, `./pm connectors`, and `./pm connectors certify --help` — pass.
- [x] `gofmt -d` on changed Go files, `git diff --check`, and `go vet ./...` — pass.
- [x] `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connectorgen-certification-matrix`, `make release-workflow-check`, `make github-parity-artifacts-check`, and `make connector-canon-check` — pass.
- [x] `make connector-boundary` — clean: 552 connectors, 284 shared files, zero findings. It initially caught three shared PostgreSQL branches/imports; those were removed without adding a boundary exception.

## Result log

The built binary has a passing PostgreSQL certificate. Its `declared_transport_pair` is a live, exact PostgreSQL proof with six modes and independent managed-target reads. The matrix has twelve evidence-backed `live_tested` cells; it remains globally uncertified until independent fixture requirements are satisfied, which the shard represents honestly.
