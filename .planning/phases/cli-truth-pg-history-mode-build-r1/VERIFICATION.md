# Verification — PostgreSQL history-mode truthfulness repair

Status: passed locally.

## Deliverable evidence

- Outer contract: `TestOpenPostgresHistoryModeResolvesRegisteredExecutors`
  resolves `postgres_polling_watermark` and `postgres_managed_target` with the
  declared `dedupe_history` strategy. The prior red output is retained in
  `traces/red-outer-history-admission.txt`.
- Happy/update/replay: `TestPMBinaryExecutesPostgresIncrementalDedupeHistory`
  builds a fresh `pm`, creates a saved PostgreSQL history connection, runs the
  normal plan/approval route against live PostgreSQL, then queries the separate
  target database. It observes v1 current, then v1 closed exactly at v2's
  `_valid_from` and v2 current, then an unchanged target after a zero-row
  replay. It also verifies the source checkpoint changes only after the target
  receipt/read-back path. Result: PASS (30.70s), retained in
  `traces/green-binary-history-live.txt`.
- Bad route: `TestIncrementalDedupeHistoryRefusesEachNonPostgresRouteBeforeSessionMutation`
  passes in `./internal/connectors/database`; it proves the typed route refusal
  leaves driver session, ledger, and target state unchanged. The native
  `TestManagedTargetHistoryRouteRefusesSourceWithoutTypedDefinition` adds the
  equivalent outer-adapter refusal.
- Edge/replay: the live test's third execution asserts zero records loaded and
  byte-for-value identical independently queried history rows. The history
  target's existing order-fence and version tests pass through the native
  PostgreSQL suite.

## Commands and results

| Command | Result |
| --- | --- |
| `go test -timeout 20m -count=1 ./internal/app` | PASS (230.728s) |
| `go test -timeout 20m -count=1 ./internal/connectors/database` | PASS (6.579s) |
| `go test -timeout 20m -count=1 ./internal/connectors/engine` | PASS (6.206s) |
| `go test -timeout 20m -count=1 ./internal/connectors/native/postgres` | PASS (1.416s) |
| `go test -timeout 20m -count=1 ./internal/synctransport` | PASS (0.587s) |
| `go test -timeout 20m -count=1 ./internal/cli` | PASS (472.760s) |
| `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -timeout 20m -count=1 -tags=databaseintegration ./internal/cli -run '^TestPMBinaryExecutesPostgresIncrementalDedupeHistory$' -v` | PASS (30.70s) |
| `go vet ./...` and `go build ./cmd/pm` | PASS |
| `go run ./cmd/connectorgen validate`; `surface-sync --check`; `certification-matrix --check`; `boundary` | PASS (552 connectors; zero validation/surface findings; boundary zero findings) |
| `pnpm --dir website run gen:docs` twice; `git diff --exit-code -- website` | PASS; second generation byte-stable |
| `make tidy-check`; `make lint`; `make docs-check-no-build`; `make smoke-no-build`; `make agent-contract-check`; `make connectorgen-validate`; `make connectorgen-surface-sync`; `make release-workflow-check` | PASS |
| `scripts/verify-gsd-workflow` and `git diff --check` | PASS |

`make docs-check-no-build` first identified the expected stale generated
connector catalog. `./pm docs generate --dir docs/cli` regenerated
`docs/connectors/catalog/all-connectors.json`; the second docs check passed.
This is the required generated manual/catalog parity update, not a wording
change.

## Lifecycle and review

`discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and
`code-review` were resolved with `scripts/gsd prompt ...` and executed inline:
the direct-PR worker has no compatible Pi role runtime, and repository policy
forbids role spawning. Manual standard review covered the descriptor,
source/destination route seal, keys/positions, registry preflight, live binary
test, and generated catalog. No unresolved findings remain.
