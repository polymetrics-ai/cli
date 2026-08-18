# Verification checklist — Issue #3976

## R2 — resumable source reads (verified locally; live execution pending)

- [x] RR1 happy path: production construction reaches the shared polling executor and
      asserts exact records plus successful checkpoint resume.
- [x] RR2 bad path: unset stream cursor returns the named typed refusal before source I/O.
- [x] RR3 edge: nullable cursor fixture cannot omit a null-cursor record.
- [x] RR4 edge: stale/invalid resume checkpoint returns its named reason before I/O.
- [x] RR5 edge: two streams with distinct cursor columns bind independently.
- [x] The private PostgreSQL paging loop is absent from the source path.
- [x] Focused PostgreSQL/engine/CLI tests, race where applicable, vet, build, and
      repository gates pass.
- [x] Generated `verify-work` / `code-review` manual-inline fallback records contain
      production reach, red/green, and finding dispositions.
- [x] Live PostgreSQL dbtest status is explicitly recorded as pending; no shared
      container runtime is started or restarted for this task.

### R2 local evidence

- Production entry-point: `TestPMBinaryExecutesPostgresFixturePollingResume` produces 3 rows
  in 2 pages, persists `polling_watermark`, then resumes with 0 rows in 1 acknowledged empty page.
- Typed pre-I/O refusals: `TestPostgresPollingTransportRefusesMissingPerStreamCursorBeforeIO`,
  `TestPostgresPollingReadPlanRefusesNullableCursor`,
  `TestPostgresPollingTransportRefusesInvalidCheckpointWithoutRestart`, and
  `TestPostgresPollingTransportRefusesStaleSchemaCheckpointBeforePageRead`, and
  `TestPMBinaryRefusesPostgresFixturePollingUnknownStreamCursorBeforePageRead`.
- The tagged `databaseintegration` PostgreSQL and CLI packages compile with no tests executed.
  The live test was not run: the shared container runtime was not started or restarted.

## Acceptance proof

- [x] RED: `dynamic_catalog_integration_test.go` was committed before
      production implementation and its exact expected compiler failure is in
      `traces/dynamic-catalog-red.txt`.
- [ ] Live typed catalog discovery proves two distinct PostgreSQL schema
      fixtures yield different correct catalogs by independent PostgreSQL
      catalog oracle inspection.
- [x] Configured database/schema/table identity and deterministic relation,
      column, and key ordering are retained.
- [x] Columns include nullability, native type identity/modifiers, supported
      logical type, ordinal, and ordered primary-key membership.
- [x] Unsupported or unsafe native shapes fail closed with named secret-safe
      errors rather than coarse static/string fallback.
- [x] The #4034 typed catalog foundation is used at PostgreSQL's shipping
      runtime boundary; no second disconnected static catalog remains.
- [x] Static behavior owned by #3980/#3982/#3983/#3987 is recorded and not
      changed by this child.
- [x] PostgreSQL `write`, `query`, and `cdc` capability truth remains unchanged.

## Planned local gates

- [x] focused PostgreSQL catalog unit tests
- [x] opt-in live `dbtest` PostgreSQL catalog/read test through Docker and
      Colima's explicit Unix endpoint; its command and verbatim output are in
      `traces/live-reads-green.txt` and issue #3976
- [x] `go test -race -timeout 20m ./internal/connectors/native/postgres`
- [x] `go test -timeout 20m ./internal/cli -count=1`
- [x] `go vet ./...`
- [x] `go build ./cmd/pm`
- [x] `make tidy-check`
- [x] `make lint`
- [x] `make docs-check`
- [x] `make smoke-no-build`
- [x] `make agent-contract-check`
- [x] `make connectorgen-validate`
- [x] `make connectorgen-surface-sync`
- [x] `make connector-boundary`
- [x] `make release-workflow-check`
- [x] generated `verify-work` and `code-review` manual-inline records
- [ ] no-mistakes child pipeline without `--yes`, maximum five fresh correction
      loops

## Live-proof resumption

- [x] RED trace records the pre-base harness-constructor incompatibility;
      `traces/live-reads-green.txt` replaces that historical state with the
      exact passing Docker/Colima command and output.
- [x] PostgreSQL dbtest uses an explicit `docker` or `podman` direct Unix
      endpoint and a pinned capacity probe, never a global runtime default.
- [x] A real catalog discovery reports the seeded table's native/logical
      metadata; a full read returns all asserted rows; a cursor-advanced read
      returns only its asserted rows.
- [x] The live proof records the actual no-cursor, nonexistent-column,
      nullable-cursor, and two-table connection-level cursor outcomes.
- [x] The historical CDC integration test is recorded as intentionally
      fail-closed/skipped, not treated as live success or re-enabled.

## Delivery holds

- [x] Draft child PR #4065 exists with `Refs #3976` and `Refs #3972`, targets
      `integration/4015-mvp-flat-r1`, and includes merge commit `0df3d5d4d`
      absorbing base head `fbd06e7d7c5c0632182e98cbb3a223ba25b19883`.
- [ ] Parent integration remains held until corrected #4058 is green and merged.
- [ ] Automated review coverage is recorded and every actionable finding is
      dispositioned before parent-branch integration.

## CLI/docs/website applicability

No new command, flag, or help topic is added. The existing PostgreSQL connector
description changed from `information_schema` to PostgreSQL system-catalog
discovery, so its defs documentation, generated connector manuals/catalog, and
website connector data were regenerated. Connector docs validation remains a
required final gate.

## Final local outcomes

- Focused package run: PostgreSQL, `database`, and `engine` all passed.
- PostgreSQL race run passed; `internal/cli` passed in 162.331 seconds.
- `go vet ./...`, `go build ./cmd/pm`, lint, docs, smoke, agent-contract,
  connector validation/surface sync/boundary, and release-workflow gates passed.
- The tagged live regression passed against real PostgreSQL 16.10 through
  Docker and Colima. It discovered the seeded catalog, returned full IDs
  `1,2,3,4,5`, and returned only `3,4,5` after cursor `10`.

Command and verbatim output:

```sh
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -run '^TestPostgresDynamicTypedCatalogUsesLiveMetadata$' -v ./internal/connectors/native/postgres
```

```text
=== RUN   TestPostgresDynamicTypedCatalogUsesLiveMetadata
    dynamic_catalog_integration_test.go:90: live PostgreSQL full read read_events: ids=1,2,3,4,5 labels=alpha,bravo,charlie,delta,echo
    dynamic_catalog_integration_test.go:90: live PostgreSQL cursor read read_events after=10: ids=3,4,5 labels=charlie,delta,echo
    dynamic_catalog_integration_test.go:90: live PostgreSQL cursor_field absent with stored cursor=12: ids=1,2,3,4,5
    dynamic_catalog_integration_test.go:90: live PostgreSQL nonexistent cursor column: read rejected
    dynamic_catalog_integration_test.go:90: live PostgreSQL nullable cursor rows after=1: ids=23; null cursor row omitted
    dynamic_catalog_integration_test.go:90: live PostgreSQL connection-level cursor_field=sequence: alternate_events rejected because it requires alternate_cursor
    dynamic_catalog_integration_test.go:76: PostgreSQL database test target image-store free bytes: before=100015849472 after=100015857664
--- PASS: TestPostgresDynamicTypedCatalogUsesLiveMetadata (5.41s)
PASS
ok  	polymetrics.ai/internal/connectors/native/postgres	6.146s
```
