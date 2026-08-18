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

### CI fix round 1

The independent repairs are complete and recorded without changing the fast-path benchmark result.

- Green — `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$' -count=1` regenerated the three stale `connections` transcript entries. Inspection before regeneration established that each carried only the two intended changes: the `--transform-file plan.json` synopsis token and the closed `TransformPlanV1` manual section. The rerun passes.
- Green — `pnpm --dir website run gen:docs` regenerated `website/lib/docs.generated.ts`; a second generator run preserved its SHA-256, and `pnpm --dir website run typecheck` passes.
- Green — `go mod why -m github.com/apache/thrift` establishes `internal/warehouse -> apache/arrow-go/v18/parquet -> thrift`. Upgrading the existing direct Arrow dependency from `v18.1.0` to `v18.7.0` selects Apache Thrift `v0.24.0` instead of `v0.21.0`. `go mod verify`, `go build ./cmd/pm`, and `go test -timeout 20m ./internal/warehouse ./internal/connectors/database ./internal/connectors/native/postgres ./internal/synctransport ./internal/cli` pass.
- Green — the complete CLI package, `go vet ./internal/warehouse ./internal/connectors/database ./internal/connectors/native/postgres ./internal/synctransport ./internal/cli`, and `make docs-check` pass after the generated-output repair.

### Full-overwrite acknowledgement-contract decision

The accepted decision and its independent second opinion were read in full before this repair. Production publish-then-checkpoint semantics remain unchanged; this is an App test-double and test-matrix correction only.

- Red — `TestAppTransportDestinationAfterApplyRunsAfterOrdinaryAcknowledgement` first observed zero acknowledgement calls in the callback; the historical full-overwrite stale-writer wait timed out at 21.28 s because page apply is not a receipt boundary. The full-overwrite source/cancellation inverse first observed no abort (`1/0/0/0` apply/abort/publish/read-back).
- Green — the ordinary callback now follows acknowledgement construction. `TestRunETLTransportFullOverwriteSourceFailureAbortsWithoutCheckpoint`, `TestRunETLTransportFullOverwriteCancellationBeforePublishAbortsWithoutCheckpoint`, `TestRunETLTransportFullOverwriteStaleWriterAfterReceiptReadBackFinalizesLosingRun`, `TestRunETLTransportFullOverwriteCompletionRebasesUnrelatedStateAfterFinalCheckpoint`, and `TestRunETLTransportFullOverwriteCompletionMissingRunIsTypedConflictAfterFinalCheckpoint` pass alongside their retained per-page matrices. These prove one abort and zero publish/read-back/checkpoint before publication, a same-target stale winner after receipt read-back, and the two truthful final-checkpoint terminal races.
- Green — `go test -timeout 20m ./internal/app -run '^TestRunETLTransportFullOverwriteStaleWriterAfterReceiptReadBackFinalizesLosingRun$' -count=20` passes at the existing 20-second test timeout; no timeout was raised. The focused race run for the full-overwrite stale/rebase/missing tests also passes.
- The two end-to-end post-final-checkpoint cancellation/failure cases remain executed for acknowledgement-per-page modes and are explicitly inapplicable for `full_overwrite`: the controller has no real observation point after its sole final checkpoint. No test is skipped or disabled, and no fake hook was added.

### Full local regression gate

Green — `make test` passes in full after the decision repair.

### CodeQL follow-up

Red — CodeQL on `585d620e` reports the new Postgres map-capacity arithmetic and the impossible unsigned `Bavail < 0` comparison.

Green — the capacity hint now performs no arithmetic, and the filesystem probe retains its signed block-size rejection and multiplication-overflow clamp while removing only the dead unsigned-negative comparison. The Postgres compiler receives only a valid immutable plan created by the closed parser and its regular-file admission capped at 64 KiB; the native package cannot construct its representation. The bound therefore makes overflow practically unreachable, but the fix does not rely on it. `go test -count=1 -timeout 20m ./internal/connectors/native/postgres ./internal/warehouse`, `go vet ./...`, and full `make test` pass. The existing Postgres plan happy/bad tests and warehouse capacity happy/bad/edge tests retain their observable results. Remote CodeQL is the remaining confirmation. `security/snyk` is intentionally not chased because it fails identically on the base branch.

### Arrow v18.7 lint follow-up

Red — `make verify` on `585d620e` has zero test `--- FAIL:` lines but fails `make lint`: staticcheck `SA1019` reports the Arrow v18.7-deprecated `arrow.Record` and `NewRecord` APIs introduced by the CVE-fixing dependency update.

Green — all eight scoped calls use `arrow.RecordBatch` or `NewRecordBatch`, including the two additional zero-row/null test constructors found by direct source scan. `go test -count=1 -timeout 20m ./internal/connectors/native/postgres` passes, `make lint` exits 0 with `0 issues.`, and full `make test` exits 0. The type/constructor migration preserves existing COPY happy/bad/zero-row-null results and record release lifecycle; no benchmark or behavior changed. Both mandatory command results are appended to the status file before push.

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
