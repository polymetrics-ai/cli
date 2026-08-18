# Local verification

Date: 2026-08-18

All commands ran from the isolated task worktree. The database tests selected the existing direct Colima Docker Unix socket without starting, stopping, or restarting the runtime.

## Passed commands

```text
gofmt -w internal/synctransport/transport_test.go internal/cli/postgres_transport_binary_integration_test.go
gofmt -w internal/synctransport/orchestrator.go internal/synctransport/arrow_fast_path_controller.go internal/synctransport/arrow_fast_path_pipeline.go
gofmt -w internal/cli/postgres_transport_binary_integration_test.go
git diff --check
go test -count=1 -timeout 20m -run '^TestOrchestratorSourceCheckpointFollowsRefreshSemantics$' -v ./internal/synctransport
go test -count=1 -timeout 20m ./internal/synctransport                         # PASS 0.747s
go test -count=1 -timeout 20m ./internal/app                                   # PASS 237.994s
go test -count=1 -timeout 20m ./internal/cli                                   # PASS 543.951s
go vet ./...                                                                   # PASS
go build ./cmd/pm                                                              # PASS
make tidy-check                                                                # PASS
make lint                                                                      # PASS, 0 issues
make docs-check                                                                 # PASS
make smoke-no-build                                                             # PASS
make agent-contract-check                                                       # PASS
make connectorgen-validate                                                      # PASS, 552 checked, 0 findings
make connectorgen-surface-sync                                                  # PASS, 552 scanned, 0 drift
make connector-boundary                                                         # PASS, clean, 292 files, 552 connectors, 0 findings/warnings
make release-workflow-check                                                     # PASS
scripts/verify-gsd-workflow                                                     # PASS
/Users/karthiksivadas/karthik-agent-workspace/bin/fm-ensure-agents-md.sh .      # PASS, unchanged
```

Both green database commands passed repeatedly with the exact counts recorded in `traces/green-full-refresh-checkpoint.md`:

```text
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -run '^TestPMBinaryPostgresFullOverwriteRetainsEverySourcePage$' -v ./internal/cli

POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -run '^TestPMBinaryPostgresIncrementalUpsertStillSkipsUnchangedSource$' -v ./internal/cli
```

Supplemental changed-package lint passed:

```text
task_base=$(git merge-base HEAD origin/integration/4015-mvp-flat-r1)
golangci-lint run --new-from-rev "$task_base" ./internal/synctransport ./internal/cli
# 0 issues
```

## Honest non-green and excluded commands

`golangci-lint run ./internal/synctransport ./internal/cli` is not the repository lint target and reported 16 pre-existing findings (three errcheck, nine staticcheck/deprecation/style, and four unused test helpers). None is on an added line; the target-branch diff-scoped invocation returned zero findings.

The opt-in monolithic `TestPMBinaryExecutesPostgresWarehousePostgres` independently printed the required unchanged incremental proof (`1001/1001`, then `0/0`, with 1,001 rows and sample `id=1001 sequence=10010 label="event-1001"` unchanged), then failed later on its unrelated CDC process-restart section with `invalid_checkpoint: PostgreSQL polling checkpoint mechanism is not resumable`. That known base-branch defect is the subject of open PR #4253. The new dedicated incremental test isolates and passes the contract relevant to this fix.

Per `AGENTS.md`, `go test -timeout 20m ./...` and `make verify` were not run as single local commands: the 550+ connector suite is routinely cut off by the per-command harness and a cutoff is indistinguishable from a hang. Changed packages and `internal/cli` ran separately, every listed non-monolithic Make gate ran individually, and CI owns the full suite.
