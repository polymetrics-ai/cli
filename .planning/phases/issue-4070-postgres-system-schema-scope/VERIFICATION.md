---
phase: issue-4070-postgres-system-schema-scope
issue: 4070
status: passed_locally
formal_gsd_phase: unavailable
verification_mode: manual_gsd_fallback
---

# #4070 verification — PostgreSQL system-owned schema rejection

## GSD verification route

`gsd-sdk query init.phase-op issue-4070-postgres-system-schema-scope` reported
`phase_found: false`: issue #4070 has no formal ROADMAP.md phase. Per the
repository contract, the issue-specific phase directory records the inline
manual fallback rather than pretending a formal phase exists. The completed
sequence is context/discussion → TDD plan → committed RED → minimum Green →
documentation/refactor → focused/live verification → deep review. The old #3976
pipeline was neither continued nor used as a sixth correction loop.

## Goal-backward acceptance evidence

| Required result | Evidence | Result |
| --- | --- | --- |
| Reserved names are refused before pool creation | `TestTypedCatalogRejectsReservedConfiguredSchemasBeforeConnect` closes a unique loopback listener, then invokes the current connector for `pg_catalog`, `information_schema`, `pg_toast`, `pg_toast_4070`, and `pg_temp_4070`; all return `ErrSystemCatalogSchema`, rather than a dial error. Fresh `pm catalog refresh` against a reserved schema configured with an unused loopback port returns the same scope error. | Pass |
| Temporary physical namespaces are covered at the real boundary | A run-owned PostgreSQL session created a temporary table and PostgreSQL reported physical namespace `pg_temp_3`. The Red PM binary exposed that table; fresh Green `pm catalog refresh` rejected it. The database-integration test keeps the same live `pg_temp_N` shape in its acceptance path. | Pass |
| Allowed discovery remains dynamic | One fresh PM binary refreshed `audit_alpha_4070` and `audit_beta_4070`, created during the run with materially different tables, fields, and key shapes. The results differed without connector JSON/code schema edits and matched independent `information_schema`/`pg_catalog` queries. | Pass |
| Identity, ordering, nullability, primary keys, and supported mapping remain correct | Independent server queries matched ordered relation and column identity, nullability, and primary/unique key order for the alpha/beta results. Existing typed-catalog unit/integration assertions cover supported native/logical type mapping. | Pass |
| Unsupported allowed schemas fail closed | Enum-backed `audit_unsupported_4070.events` returned the explicit unsupported-catalog-shape outcome instead of a partial or static catalog. | Pass |
| Fixture/static metadata cannot substitute | The PM runs used live credential/connection configuration and the production registry; alpha and beta differed solely because server schemas differed. `typed_catalog.go` still queries the configured live schema, while fixture mode exits before typed discovery. | Pass |
| System behavior is documented and generated data stays derived | `internal/connectors/defs/postgres/docs.md` describes the exact rejected forms. `npm run gen:website-data`, `make docs-check`, `connectorgen validate`, and `surface-sync --check` passed. | Pass |
| Run-owned cleanup is idempotent and leaves no residue | PostgreSQL oracle reported zero created schemas after drops and zero temporary relations after session close; repeated `IF EXISTS` cleanup succeeded. The stopped, uniquely named scratch directory was moved to operating-system Trash after its unique marker was checked. | Pass |

## Commands run and outcomes

All commands below were executed from this isolated worktree. Live commands use
only a unique loopback PostgreSQL cluster and redacted environment aliases;
they do not contain a raw connection string or secret.

```text
go test -count=1 ./internal/connectors/native/postgres                                      PASS
go test -tags=databaseintegration -count=1 -run '^$' ./internal/connectors/native/postgres PASS
go test -race -count=1 ./internal/connectors/native/postgres                                PASS
go test -timeout 20m ./internal/connectors/native/postgres ./internal/connectors/database \
  ./internal/connectors ./internal/cli                                                       PASS
go test -timeout 20m ./internal/cli                                                         PASS
go vet ./...                                                                                PASS
go build ./cmd/pm                                                                           PASS
make tidy-check                                                                             PASS
make lint                                                                                   PASS (0 issues)
make docs-check                                                                             PASS
make connector-boundary                                                                     PASS (outcome clean; 552 bundles)
make connector-runtime-preflight                                                            PASS
make connector-canon-check                                                                  PASS
make release-workflow-check                                                                 PASS
go run ./cmd/agentcontractgen check                                                        PASS
go run ./cmd/connectorgen validate internal/connectors/defs                                PASS
go run ./cmd/connectorgen surface-sync --check                                              PASS
npm run gen:website-data                                                                    PASS
git diff --check                                                                            PASS after evidence whitespace normalization
```

The only `git diff --check` finding was trailing whitespace in new planning
Markdown. It was normalized before this evidence commit; it was not a source or
runtime finding.

## Live-boundary command shapes

```sh
initdb -D <unique-run-data-dir> --auth=trust --username=<run-owned-role>
pg_ctl -D <unique-run-data-dir> -o '-h 127.0.0.1 -p <run-port>' -w start
pm init
PM_4070_TEST_PASSWORD=<redacted> pm credentials add <connection> \
  --connector postgres --from-env password=PM_4070_TEST_PASSWORD \
  --config host=127.0.0.1 --config port=<run-port> \
  --config database=<run-owned-db> --config username=<run-owned-role> \
  --config schema=<configured-schema> --config sslmode=disable
pm catalog refresh --connection <connection> --json
```

The full Red, Green, server-oracle, and cleanup observations are retained in
the trace files. No secret values were printed or persisted.

## Remaining gate

Local verification is complete. The new #4070 no-mistakes pipeline is
intentionally not started because Firstmate requested a pause while the shared
daemon is needed for GitHub run `01KZSJG7P7QREZSZ7N52TPBA5F`. This is an
external coordination pause, not a verification failure.
