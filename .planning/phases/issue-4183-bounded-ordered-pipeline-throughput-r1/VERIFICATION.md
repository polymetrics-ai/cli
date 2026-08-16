# Verification — bounded ordered pipeline throughput R1

## Planned local gates

- `go test -count=1 -timeout 20m ./internal/synctransport ./internal/connectors/native/postgres ./internal/app`
- `go test -race -count=1 -timeout 20m ./internal/synctransport`
- `go test -count=1 -timeout 20m ./internal/cli`
- `go vet ./internal/synctransport ./internal/connectors/native/postgres ./internal/app ./internal/cli`
- `go build ./cmd/pm`
- `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, and `make release-workflow-check`
- `pnpm --dir website run gen:docs`, then repeat it and verify the generated output is byte-stable.
- With the explicit shared endpoint only: `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m ./internal/cli -run '<named production binary regressions>'`.
- The exact opt-in 5 GB proof before and after implementation; record host load, stage timings, logical bytes, elapsed seconds, decimal MB/s, MiB/s, and pass/fail gate in a new committed artifact.

## GSD and review record

The manual inline fallback is authorized by the unavailable compatible Pi role runtime and the repository's no-role-spawning contract. `scripts/gsd doctor`, all five `scripts/gsd sources` commands, and `go run ./cmd/agentcontractgen check` passed before planning. The generated discuss, TDD plan, execute, verify-work, and code-review prompts are resolved; their implementation, verification, and review evidence will be appended here rather than silently omitted.

## Results

### TDD and atomicity

Green:

```bash
go test -v -count=1 -timeout 20m ./internal/synctransport -run 'TestOrchestrator(FullOverwritePublishesAllBoundedPagesBeforeOneCheckpoint|ArrowFullOverwrite(OverlapsNextExtractionWithPreviousCopy|DepthOneKeepsSerialEmitBehavior|RefusesUndeclaredPipelineBeforeIO|PipelineFailureAbortsBeforePublicationOrCheckpoint|PipelineCancellationDrainsBeforeAbort))'
go test -race -count=1 -timeout 20m ./internal/synctransport
go test -v -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransport(FullOverwrite(SourceFailureAbortsWithoutCheckpoint|CancellationBeforePublishAbortsWithoutCheckpoint|StaleWriterAfterReceiptReadBackFinalizesLosingRun|ReceiptReadBackThenStaleFinalCheckpointFinalizesLosingRun|CompletionRebasesUnrelatedStateAfterFinalCheckpoint|CompletionMissingRunIsTypedConflictAfterFinalCheckpoint)|PersistsActiveCheckpointBeforeSourceFailureForPerPageAcknowledgementModes)$'
go test -count=1 -timeout 20m ./internal/app -run '^TestRunETLRefusesExplicitPipelineDepth'
```

The fakes prove depth-two overlap, source order, no page-three admission above a depth-two bound, byte-credit peak, serial depth-one behavior, typed pre-I/O refusal, failure abort/no publish/no read-back/no checkpoint, and cancellation drain/abort. The preserved full-overwrite app tests prove PR #4184's source-failure, pre-publish cancellation, post-read-back stale writer, final checkpoint, and unrelated-state behavior.

Green production binary proof:

```bash
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/cli -run '^TestPMBinaryPostgres(FullOverwriteRetainsEverySourcePage|TransformedFullOverwriteUsesArrowCOPY)$'
```

It independently read every two-page overwrite row and the transformed Arrow-COPY target after the default admitted pipeline path; both tests passed.

### Package, CLI, and generated-surface checks

#### PR #4190 help-dispatch gap (resolved)

The PR verify run found that `pm etl run --help` and `pm connections create --help` reached `withApp` before their in-handler help guards, so invoking help from a directory without `.polymetrics` failed. This is a user-facing regression, not an environmental failure. The red step was:

```bash
go test -count=1 -timeout 20m ./internal/cli -run '^TestChangedCLICommandHelpIsExecutable$'
```

after the test changed into `t.TempDir()`; both subtests failed with `open project at .polymetrics`.

Green moves dispatch for only those known leaf manuals ahead of `withApp`, retains the in-handler guard for direct callers, and makes the test assert zero stderr as well as exit zero and flag text. The full `internal/cli` suite, both golden-transcript passes (one update, one check), all scoped safety packages and race tests, the two tagged PostgreSQL binary regressions, and every individual non-aggregate repository gate below were rerun successfully. Direct built-binary calls from a fresh empty directory also succeeded for both leaf manuals. No project was created to satisfy the test, and the 5 GB performance proof remains deliberately pending.

Green:

```bash
go test -count=1 -timeout 20m ./internal/synctransport ./internal/app ./internal/cli ./internal/connectors/engine ./internal/connectors/native/postgres
go test -count=1 -timeout 20m ./internal/connectors/native/postgres ./internal/connectors/database ./internal/connectors/engine
go test -count=1 -timeout 20m ./internal/cli -run 'Test(ChangedCLICommandHelpIsExecutable|ETLTransport(PlanAndPreviewShowBoundedTargetCopyCapacity|SafeOutputOmitsApprovalAndDestinationInternals)|GoldenTranscripts|Parse(MaxInFlightBatchesRequiresExecutableBound|TargetCopyWorkersRequiresGlobalBound))'
POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test -count=1 -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$'
go test -count=1 -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$'
go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors
go run ./cmd/pm docs validate --connectors-dir docs/connectors
pnpm --dir website run gen:docs
pnpm --dir website run gen:docs
```

The repeated website generator produced the same SHA-256 on both runs. `pm help etl`, bare `pm etl`, `pm etl run --help`, `pm help connections`, bare `pm connections`, and `pm connections create --help` were exercised from the built binary; both changed leaf help paths now succeed and print their flags.

### Repository gates

Green:

```bash
gofmt -w cmd internal
go vet ./...
go build ./cmd/pm
make tidy-check
make docs-check-no-build
make smoke-no-build
make lint
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make connectorgen-certification-matrix
make connector-runtime-preflight
make connector-boundary
make connector-canon-check
make release-workflow-check
make certify-timing
```

`connectorgen-certification-matrix` initially reported PostgreSQL drift after the transport declaration changed. `go run ./cmd/connectorgen certification-matrix --connector postgres` regenerated the declared artifact, after which the check passed. `certify-timing` passed within its 3m30s budget.

The repository explicitly directs agents under a per-command timeout not to invoke aggregate `go test ./...` or `make verify` as one command; the changed packages plus `internal/cli` were run with `-timeout 20m` above, and every non-aggregate `make verify` constituent was run individually. CI remains responsible for the aggregate suite.

### Performance proof

**PENDING by captain decision; not a merge gate.** The host was contended before implementation (load average 8.60 on 12 CPUs, with independent `connectorgen`/certification work consuming full cores), so no post-change throughput value was run or published. The unchanged historical quiet-host value is 5,368,947,776 logical input bytes in 48.000221500 s = 111.85 decimal MB/s (106.67 MiB/s), using pre-transform projected source Arrow buffers. The audit's ~172.5 MB/s is an optimistic upper bound only, not a result. A future run needs the existing explicit database/performance opt-in, the shared Docker Unix socket endpoint, pre-run load recorded beside each result, and the harness's unchanged 200 MB/s/25 s gate.

### Review

Manual inline review is recorded in `REVIEW.md`. It found and fixed the queue off-by-one admission and no-op explicit-flag paths before final validation. Automated review coverage is pending the trusted-author PR-open trigger; it will be recorded with the API-reported base/head/SHA after PR creation.
