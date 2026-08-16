# TDD ledger — PostgreSQL certification profile

## Red 1 — observed

Create a focused profile test that asserts PostgreSQL has definition-owned certification metadata. On the base branch it must fail because no `certification.json` exists.

Command: `go test -timeout 20m ./internal/connectors/engine -run '^TestBundleLoadEmbeddedPostgresCertification$'`

Observed failure: `PostgreSQL Certification is nil; defs.FS must embed certification.json`.

## Green 1 — observed

Add PostgreSQL-owned declaration data that makes the focused profile test pass. Validate it uses no connector-specific condition in shared Go.

The green slice adds PostgreSQL-owned credential defaults only: a bounded live `read_limit=100` and `sslmode=disabled`, matching the connector's existing local-container test transport. It deliberately has no fabricated REST direct-read command, binary candidate, or `writes.json` pairing.

## Red 2 — observed

Build the current `pm` binary and run `connectors certify postgres --full` against a live, seeded PostgreSQL container. Before the profile existed, the build could not load PostgreSQL certification metadata. The focused live test now requires a live catalog/read report and independently verifies its source count.

## Green 2 — observed

`POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -timeout 20m -tags=databaseintegration ./internal/connectors/native/postgres -run '^TestPostgresCertificationProfileRunsBuiltBinaryLive$' -count=1` passed. It built `./cmd/pm`, catalogued and read five typed rows, and independently confirmed the source stayed unchanged.

## Red 3 — observed

After `go run ./cmd/connectorgen validate internal/connectors/defs/postgres` passed, scratch-changing the profile default to `sslmode=bananas` made that same built-binary live test fail: `exit=2`, with a report emitted by the binary. The scratch change is absent from the working diff.

## Green 3 — observed

Restoring `sslmode=disabled` and rerunning the same command passed.

## Red 4 / Green 4 — terminal roll-up

Red: an injected `DeclaredTransportCertificationProof{Declared:true}` with no executable adapter produced a `skipped:` stage and terminal pass on the base behavior.

Green: `StageResult.Status` distinguishes `skipped`, `unexecutable`, `fail`, and `pass`; `allStagesPassed` exempts only `skipped`. `TestCertificationDeclaredTransportPairFailsConnectorWithoutAdapter` now asserts `status=unexecutable`, report failure, and exit 2. The registered GitHub declared pair remains passing.

## Red 5 / Green 5 — database-shaped declared pair

Red: the first PostgreSQL `--full --write` run exposed the API-shaped assumption that a dynamic database catalog must advertise a universal `cursor_fields` list; it failed before the declared pair. A second red run proved the adapter also had no primary-key/cursor identity on its generated connection.

Green: the shared selector recognizes only an already-declared native-database polling pair and an explicit selected relation plus `cursor_field`; it derives the catalog primary key without inventing a universal database watermark. PostgreSQL's exact adapter then executes that one relation and passes the key/watermark into every declared managed-target connection. The generic incremental warehouse stage is a concrete `skipped` result because the six actual PostgreSQL incremental modes run through the declared pair instead. The live test verifies all six plan/preview/approval/runs, checkpoint evidence, independent target read-back, and unchanged source rows.

## Red 6 / Green 6 — live matrix evidence

Red: a database transport result had no accepted evidence importer and every PostgreSQL matrix cell remained `live_tested=false`.

Green: `connectorgen certification-evidence transport --connector postgres` validates the report against PostgreSQL's loaded native-database descriptor and rejects any report without a passing exact pair and all six target/read/checkpoint results. Its unit test rejects a deliberately incomplete mode before writing any record. The opt-in owned-container live run wrote 12 HMAC-redacted records; the generated matrix marks exactly the read and managed-target-write cells for all six modes, including `incremental_dedupe_history`.

## Red 7 / Green 7 — boundary and help artifacts

Red: the whole-tree `connectorgen boundary` gate found PostgreSQL literals/imports in shared generator and certification code; the changed connector-help text also made `TestGoldenTranscripts` red.

Green: shared code now derives native-database selector and evidence checks from the loaded descriptor, while PostgreSQL-specific execution remains in its exact adapter. No boundary exception was added. The project golden generator refreshed the changed help transcript. The final whole-tree boundary scan is clean and the full `internal/cli` suite is green.
