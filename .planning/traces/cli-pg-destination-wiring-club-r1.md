# PostgreSQL production transport wiring trace

## Manual GSD execution

The project-local GSD adapter generated the discuss/plan/execute prompts. This
worker executed them inline because the issue-agent contract requires one
canonical worker and forbids spawning lifecycle roles.

## Red

- R1 command: `go test ./internal/connectors/native/postgres ./internal/app -run 'TestPostgresDefinitionDeclaresBoundedSnapshotTransportSource|TestOpenRegistersDefinitionOwnedProductionTransports' -count=1`
- Observed missing behavior: PostgreSQL source is fixed to the synthetic
  `snapshot` stream, and `app.Open` cannot preflight PostgreSQL as a destination
  because no destination declaration/factory exists.
- This is a production-composition assertion: the test obtains the registry
  built by `app.Open`; it does not hand-register a destination executor.

## Green

- Production registration: `TestOpenRegistersDefinitionOwnedProductionTransports`
  resolves PostgreSQL's exact native destination for both declared PostgreSQL
  and GitHub sources while the public PostgreSQL `write` capability stays false.
- API binary proof: `TestPMBinaryExecutesAuthenticatedGitHubWarehousePostgres`
  passed with a real GitHub credential and 50 issues from `rails/rails`. The
  built binary staged WAL/Parquet/manifest artifacts; independent reads observed
  exactly 50 Parquet rows and 50 live PostgreSQL rows with the expected logical
  types, a durable ledger receipt, and an advanced checkpoint.
- Decisive database binary proof:
  `TestPMBinaryExecutesPostgresWarehousePostgres` passed in 67.55s with 1,001
  rows over a 1,000-row boundary, exact replay, consumed-token/schema/auth/
  permission refusals, empty input, bootstrap transaction, process kill, and
  LSN resume assertions.
- Focused live components: `TestPostgresManagedTargetWorksetDeliveryLive`
  passed in 8.11s; `TestPostgresGapFreeBootstrapContainerHarness` and
  `TestPostgresBootstrapSnapshotFailureRequiresExplicitRebootstrap` passed in
  5.15s and 6.14s.
- Focused RED/GREEN for the added API seam: database schema projection and
  structured-JSON tests initially failed on missing implementation, while App
  production selection failed to select GitHub→PostgreSQL. All are green; a
  focused legacy GitHub-mode regression additionally caught and prevented an
  accidental GitHub→warehouse dispatch change.

## Merge-gate gap Red / Green

- Red command: `go test -timeout 20m ./internal/app -run 'TestIssueLabelTransportSourceCollectsBoundedBatchForManagedTarget|TestOpenRegistersDefinitionOwnedProductionTransports' -count=1`.
- Red evidence: the source refused missing `transport_source_issue_number`, and
  production dispatch refused a GitHub→PostgreSQL batch larger than one.
- Green design: singleton configuration retains its exact one-issue behavior;
  absent singleton configuration selects a buffer-before-emit collection mode
  capped at 1,000 records and ten provider pages. GitHub's mutating destination
  remains batch-size one.
- Live command: focused `databaseintegration`
  `TestPMBinaryExecutesAuthenticatedGitHubWarehousePostgres` with certification
  credential injection and real Docker PostgreSQL.
- Live result: pass in 47.700s; observed Parquet rows=50 and PostgreSQL rows=50.
  PostgreSQL types were `number=bigint`, `node_id=text`, `labels=jsonb`,
  `updated_at=text`, and `locked=boolean`.
- Refusals: `ErrPostgresManagedTargetApprovalRequired`,
  `AuthorizationTokenReplayError`, and `context.Canceled` after a confirmed
  in-flight provider request each preserved business-row counts and byte-exact
  stream checkpoint state.

## Declared boundary

The exact live managed-target commit-death/out-of-order control assertion is
the pre-existing #4158 failure and was explicitly excluded by the captain.
This phase does not report that individual assertion as green; it relies on the
passing unknown-commit/baseline unit and live-driver evidence around it and
records the exclusion in the PR edge table.
