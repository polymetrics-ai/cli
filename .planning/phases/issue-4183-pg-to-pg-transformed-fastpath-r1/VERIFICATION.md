# Verification — issue #4183

## Planned local gates

- `go test -count=1 -timeout 20m ./internal/connectors/database ./internal/connectors/native/postgres ./internal/synctransport ./internal/warehouse`
- `go test -count=1 -timeout 20m ./internal/app ./internal/cli`
- `go vet ./internal/connectors/database ./internal/connectors/native/postgres ./internal/synctransport ./internal/warehouse ./internal/app ./internal/cli`
- `go build ./cmd/pm`
- `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, `make release-workflow-check`
- With explicit opt-in and a direct local Docker/Podman Unix endpoint only: `go test -tags=databaseintegration -count=1 -timeout 20m` for the named live PostgreSQL test.
- Qualified host: `TestPMBinaryPostgresTransformedFullOverwriteFiveGigabyteProof` with both database and performance opt-ins. It hard-stops below 3 GiB free space, records peak disk, and scores 5 GB measured logical bytes at ≥200 MB/s / ≤25.0 s.

## Acceptance record

### Live binary correctness

`POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/cli -run '^TestPMBinaryPostgres(FullOverwriteRetainsEverySourcePage|TransformedFullOverwriteUsesArrowCOPY)$'`

Green: the production binary retained all two-page source rows and independently read the transformed, filtered target rows; it observed one full-overwrite receipt and durable phase measurement.

### 5 GB proof and disk safety

No dangling image was enumerated for removal. Host free space was 134,480,212 KiB before the proof and never fell below 132,682,396 KiB; the 3 GiB guard never approached its refusal threshold. The harness emitted [fastpath-5gb-proof.json](fastpath-5gb-proof.json) before container cleanup.

Input bytes are pre-transform projected source Arrow buffer bytes; they exclude Parquet, pgwire, target storage, and checkpoint bytes. The realistic mapping moved 5,368,947,776 logical input bytes in 48.032609708 s: **111.78 decimal MB/s / 106.60 MiB/s**. This is below the 200 MB/s / 25 s gate. The named bottleneck is the sequential source-read plus binary-COPY critical path: source read 16.699 s and COPY 27.638 s (with transform 1.677 s, Parquet close/fsync 1.907 s, index/constraint build 1.787 ms, publish/receipt 9.159 ms, checkpoint 10.872 ms). The identity control was 105.68 MB/s / 100.79 MiB/s in 50.802 s.

### Local checks

- Green: `go test -count=1 -timeout 20m ./internal/connectors/database ./internal/connectors/native/postgres ./internal/synctransport ./internal/warehouse`
- Green: targeted `internal/app` transform/approval/measurement tests and `internal/cli` transform-file tests; tagged 5 GB test visibly skips without the performance opt-in.
- Green: scoped `go vet`, `go build ./cmd/pm`, runtime connection help/bare namespace/help flag checks, generated documentation validation, and `make smoke-no-build`.
- Green: `agentcontractgen check`, `connectorgen validate`, `surface-sync --check`, connector-boundary, connector-canon, pinned-build dependency, Homebrew notification, and release-target parity gates.
- Manual GSD fallback: resolved `execute-phase`, `verify-work`, and `code-review` prompts. This environment cannot launch the required compatible isolated Pi role, so the worker performed the corresponding focused test/verification/review pass inline and records this fallback rather than silently skipping GSD.

The 5 GB performance assertion intentionally remains an opt-in failing gate until the bottleneck is removed; it is not counted as a Green throughput result.
