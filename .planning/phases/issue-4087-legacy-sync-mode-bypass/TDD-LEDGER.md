# TDD ledger: issue #4087

## Slice 1 — typed aliases

- **Red:** recorded 2026-08-14. Added focused parsing and `RunETL` tests for `full_refresh_overwrite_deduped` and `incremental_append_deduped`; `go test -timeout 20m ./internal/app -run 'Test(DedupedLegacyAliasesUseTypedContractsBeforeSourceIO|CanonicalSyncModesRetainParsedContracts)$'` failed as expected. Both normal and persisted-legacy parsing returned an empty `ContractMode` and `LegacyCompatibility:true`; failures are at `sync_modes_test.go:289` before the typed pre-I/O assertion.
- **Green:** recorded 2026-08-14. `internal/synccontract/public_modes.go` is the single connector-neutral authority for public names, typed contracts, and generic capability projection; direct and persisted-legacy parsing consume it through `internal/app/sync_modes.go`. The aliases now return `ModeFullOverwrite` and `ModeIncrementalDedupe` with typed admission; the focused test passed and `go test -timeout 20m ./internal/app` passed in 159.862s.
- **Refactor:** recorded 2026-08-14. Removed tests that asserted the now-invalid legacy dedupe execution, retained unrelated cursor and warehouse isolation coverage on `incremental_append`, and added closed-mode control assertions for every canonical contract. The control also verifies the pre-existing `incremental_append` persisted-compatibility execution remains unchanged, while the other typed contracts retain their pre-I/O refusal without a transport.

## Slice 3 — surface and certification parity

- **Red:** recorded 2026-08-14. Before the help-source change, `go test -timeout 20m ./internal/cli -run TestETLHelpListsAllSyncModes$` failed because neither compatibility description identified the typed canonical admission. The real CLI certification sweep likewise treated the typed `Error`/exit-1 response as a failed expectation for an `ETLRun`.
- **Green:** recorded 2026-08-14. Updated generic help, generated `docs/cli` output (via `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors`), website docs, and the connector-neutral certification expectation. `go test -timeout 20m ./internal/connectors/certify -run TestSourceStagesAgainstSample$` and `go test -timeout 20m ./internal/cli -run 'Test(CertifyCLISingleConnectorPassExitsZero|ETLHelpListsAllSyncModes)$'` pass.

## Slice 4 — generated static-surface reconciliation

- **Red:** recorded 2026-08-14. Review found that a connector-literal static-surface test violated the connector-neutrality boundary and that generated connector manuals, skills, and catalog data still used the superseded primary-key-only projection. A stream with a primary key but no cursor could therefore be advertised with `full_refresh_overwrite_deduped` even though `internal/synccontract/public_modes.go` rejected that public mode.
- **Green:** recorded 2026-08-14. Removed the connector-specific test, retained the generic engine truth table, and made the connector-doc validator compare every generator-owned manual, skill, index, and catalog against the same render functions used by the writer. `internal/synccontract/public_modes.go` remains the sole connector-neutral authority; `internal/connectors/engine` consumes its capability projection. The reference and website docs now state that full dedupe requires a cursor and primary key, while incremental modes also require an executor.
- **Generated artifact inventory:** `pm docs generate` owns `docs/cli`, connector manuals, connector skills, the catalog JSON/Markdown, index, and icon assets; 192 manuals, 192 skills, two CLI pages, and `all-connectors.json` changed. `pm skills generate` owns `docs/skills` (six files changed); the golden-transcript updater owns `internal/cli/testdata/golden_transcripts.json`; and `npm run gen:docs` owns `website/lib/docs.generated.ts`. `connectorgen certification-matrix` owns capability, flow, and status matrices and was rerun successfully without a content change. The website connector-bundle generator consumes raw bundle fields rather than the sync-mode projection, so it has no projection-derived artifact to regenerate.
- **Generator record:** `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors` reported `Generated docs in docs/cli and connector docs in docs/connectors`; `go run ./cmd/pm skills generate --dir docs/skills` reported `Generated skills in docs/skills`; `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$'` reported `ok polymetrics.ai/internal/cli 22.847s`; `npm run gen:docs` reported `Wrote 12 docs pages to lib/docs.generated.ts.`; and `go run ./cmd/connectorgen certification-matrix` reported `generated certification matrices: connectors=556 capability_complete=0 certified=0`.
- **Check record:** `go run ./cmd/pm docs validate --connectors-dir docs/connectors` reported `Validated connector docs in docs/connectors`; `go run ./cmd/connectorgen surface-sync --check` reported `552 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)`; and `go run ./cmd/connectorgen certification-matrix --check` reported `certification matrices are current: connectors=556 capability_complete=0 certified=0`. The scoped alias/canonical, engine, docs, skills, and golden tests passed; the app, synccontract, and engine packages reported `10.620s`, `0.445s`, and `1.885s` respectively, and the explicit CLI invocation completed successfully with no emitted runner output.

## Planned checks

- `go test -timeout 20m ./internal/app -run 'Test(.*Legacy.*Typed|.*Canonical.*)'`
- `go test -timeout 20m ./internal/app`
- `go test -timeout 20m ./internal/synccontract ./internal/synctransport`
- `go test -timeout 20m ./internal/connectors/certify`
- `gofmt -w internal/app/sync_modes.go internal/app/sync_modes_test.go`
- `go vet ./...`
- `go build ./cmd/pm`
- individual `make verify` gates, including `connectorgen-surface-sync`

## Slice 4 — CI standard-library vulnerability repair

- **Red:** recorded 2026-08-14 from the PR `govulncheck` job. Its explicit `GOTOOLCHAIN=go1.25.12` selected a runtime with seven reachable standard-library advisories (GO-2026-6218, GO-2026-6091, GO-2026-6090, GO-2026-6089, GO-2026-6088, GO-2026-5972, and GO-2026-5026), each fixed in Go 1.25.13. The job exited 3; neither a dependency upgrade nor a scanner suppression is a correct remediation.
- **Green:** recorded 2026-08-14. Advanced `go.mod` and every explicit executable workflow pin together to Go 1.25.13, then ran the exact fixed-toolchain command: `GOTOOLCHAIN=go1.25.13 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` reported `No vulnerabilities found.` The repair leaves the Go language directive and dependency graph unchanged.

## Slice 5 — schema cursor capability reconciliation

- **Red:** recorded 2026-08-14. Review found that Chargebee `coupons` declared `incremental.cursor_field: updated_at` without the schema `x-cursor-field`, so the schema-derived catalog omitted the cursor-dependent public modes despite the executable incremental declaration.
- **Green:** recorded 2026-08-14. Declared the schema cursor, made the generic validator reject a nonempty incremental cursor that disagrees with its schema, and kept valid empty-cursor incremental declarations eligible through their schema cursor. Regenerating connector documentation restored `coupons` to all five canonical public modes in the catalog, manual, and skill output. The focused engine and connectorgen regressions, Chargebee validation, and connector-doc validation passed.

## Slice 6 — CI warehouse-fixture recovery

- **Red:** recorded 2026-08-14 from the supplied `verify` job and reproduced locally with `go test -count=1 -timeout 20m ./internal/app -run '^TestQuerySQLScopesConnectionOwnedAndUnattributedViews$'`. The warehouse-isolation fixture configured `incremental_append_deduped` and then called `RunETL`; the required typed-admission behavior correctly returned `sync mode "incremental_dedupe" is not executable` before materialization.
- **Green:** recorded 2026-08-14. Changed only that materializing fixture to the already-supported canonical `incremental_append` mode; alias-specific tests continue to cover typed pre-I/O refusal. `go test -count=1 -timeout 20m ./internal/app -run '^(TestQuerySQLScopesConnectionOwnedAndUnattributedViews|TestDedupedLegacyAliasesUseTypedContractsBeforeSourceIO|TestCanonicalSyncModesRetainParsedContracts)$'` passed, and the full affected package `go test -count=1 -timeout 20m ./internal/app` passed in 188.630s. `GOTOOLCHAIN=go1.26.6 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` reported `No vulnerabilities found`; `go mod verify` passed.

## Slice 7 — effective cursor correction

- **Red:** review found that `DerivedSyncModes` treated only `x-cursor-field` as a cursor and that the validator rejected a nonempty incremental cursor when the schema omitted that optional extension; the existing truth-table case codified the resulting nonincremental projection.
- **Green:** recorded 2026-08-14. A shared effective-cursor resolver now projects a schema cursor or a standalone incremental cursor into catalog and definition views, while an explicit schema/incremental conflict remains separately gated. `go test -timeout 20m ./internal/connectors/engine ./cmd/connectorgen -run 'Test(DerivedSyncModesTruthTable|ConnectorIncrementalCursorWithoutSchemaFieldProjectsEffectiveCursor|Validate_RejectsIncrementalCursorSchemaMismatch|Validate_AcceptsIncrementalCursorWithoutSchemaCursor)$'` passed.

## Slice 8 — certification report contract preservation

- **Red:** recorded 2026-08-14. `go test -count=1 -timeout 20m ./internal/connectors/certify -run '^(TestSourceStagesAgainstSample|TestReportSaveAllowsEmptyPreIORefusalDataSource)$'` failed: `Save` rejected the pre-existing empty data source for a typed pre-I/O refusal, while the runner emitted schema version `2` and serialized `pre_io_refusal` for both compatibility aliases. This contradicts the version-1 report contract and the supplied acceptance constraint.
- **Green:** recorded 2026-08-14. Restored the version-1 report implementation and its unvalidated string data-source field; typed pre-I/O refusals now retain the established empty source with their existing reason instead of adding a new serialized value. The focused certification test passed, as did the exact base comparison for `report.go` and both protected architecture design documents.
