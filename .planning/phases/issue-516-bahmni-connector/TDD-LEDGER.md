# TDD ledger — Bahmni-docker connector (issue #516)

Manual-GSD fallback (repo-local `scripts/gsd` unavailable in this worktree). The connector is
data, not Go, so the fail-first gates are the bundle validator (`cmd/connectorgen/validate.go`) and
the fixture-replay conformance harness (`internal/connectors/conformance/`), both of which run over
the real `internal/connectors/defs` tree.

## Red

Captured against the pre-change tree (`internal/connectors/defs/bahmni/` did not exist):

- `go test ./internal/connectors/conformance -run 'TestConformance/bahmni'` matched **no
  subtest** — the connector had zero conformance coverage, the defining absence this phase closes.
- `go run ./cmd/connectorgen validate internal/connectors/defs` reported **547 connector(s)
  checked**, i.e. the bundle was absent from the fleet.

Real red observed while landing the bundle, each fixed before moving on:

- `Red:` `connectorgen validate internal/connectors/defs/bahmni` exited 1 with 2 findings
  (`missing_file` on `fixtures`/`schemas`). Cause: `validateDir` treats each *subdirectory* of the
  given path as a bundle, so the validator must be pointed at the parent `defs` tree, not at a
  single bundle directory. Corrected the invocation rather than the bundle.
- `Red:` `TestNewLoadsDeclarativeBundlesWithHooksAndNativeOverrides` failed:
  `bundle count = 548, want 547`. This is the fail-first proof that the new bundle actually
  registered through the production `//go:embed` in `internal/connectors/defs/defs.go`.
- `Red:` `TestConnectorCatalogCLIJSON` failed: catalog JSON missing `"count": 551` (runtime catalog
  is now 552 = 548 declarative bundles + 4 local primitives).
- `Red:` `TestGoldenTranscripts` failed after correcting the `pm connectors` CATALOG help text,
  because four recorded transcripts still carried the stale "551 bare-name entries: 547 declarative
  bundles" string.
- `Red:` `./pm docs validate --connectors-dir docs/connectors` exited with
  `connector catalog json has 551 entries, want 552`.

## Green

- `go run ./cmd/connectorgen validate internal/connectors/defs` → **548 connector(s) checked, 0
  findings** (whole fleet, not just this bundle).
- `go test ./internal/connectors/conformance -run 'TestConformance/bahmni'` → **PASS**, with
  real engine replay rather than skips. Per-check dump:
  - All 10 static checks pass: `spec_schema_valid`, `stream_schemas_valid`, `pk_fields_exist`,
    `cursor_fields_exist`, `interpolations_resolve`, `write_schemas_valid`, `surface_complete`,
    `docs_present`, `secret_redaction`, `fixtures_present`.
  - Dynamic replay passes: `check_fixture`; `read_fixture_nonempty` for the 6 fixtured streams
    (patients, encounters, concepts, locations, providers, appointments — the last exercising the
    root-array `records.path: ""` path); `pagination_terminates`; `records_match_schema`.
  - Legitimately skipped: `cursor_advances` (no incremental streams by design), the 6 unfixtured
    streams, and `write_request_shape`/`delete_semantics` (no write fixtures, matching gong).
- Count goldens made green: bundle registry 547→548, catalog CLI 551→552, `internal/cli/docs.go`
  CATALOG help text, and `internal/cli/testdata/golden_transcripts.json` regenerated with the
  documented `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1` flow.
- `./pm docs validate --connectors-dir docs/connectors` → `Validated connector docs`.
- `gofmt -l cmd internal` clean; `go vet ./...` clean; `go build ./cmd/pm` ok.
- `go test -timeout 20m` green across the tree.

## Refactor / hardening

- Kept the bundle purely declarative after confirming no existing connector ships a `native/`
  adapter, so the engine stays the single interpreter.
- Chose `offset_limit` paging to match the OpenMRS `limit`/`startIndex` convention, and sized the
  first stream's fixture below the page size so `pagination_terminates` asserts a real stop rather
  than a fixture coincidence.
- Kept generated-artifact churn scoped: `pm docs generate` rewrites all ~450 committed connector
  manuals with pre-existing drift unrelated to this connector, so only the bahmni entries
  were spliced in. The website generators are not drifty and were run wholesale.

The follow-up phase `.planning/phases/issue-516-bahmni-rename-parity-followup/` records the captain-authorized fixes for the four real automated-review defects; the two captain-owned product decisions remain recorded in VERIFICATION.md.
