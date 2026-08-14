## Summary

Implements dynamic typed PostgreSQL source catalog discovery for the configured
database/schema. The runtime derives base relations, ordered columns,
nullability, ordered primary keys, and supported native/logical type metadata
from the live PostgreSQL catalog rather than a hard-coded connector schema.

## Linked Issue

Refs #3976
Refs #3972

## Stacked PR

- Parent issue: #4097
- Parent branch: `integration/4015-mvp-flat-r1`
- PR base branch: `integration/4015-mvp-flat-r1`
- Sub-issue: #3976
- Child PR: #4065
- Integration synchronization: merge commit `1dde1b00d7da5a2abb59b202d30a37e2ec9eadab`
  absorbs current base head `5245b288776a30d808c934c51adc25593ddd5d1d`; the
  prior merge commit `0df3d5d4d` had absorbed
  `fbd06e7d7c5c0632182e98cbb3a223ba25b19883`. A lease-bound restoration moved
  the remote branch from the cancelled pipeline's stale `c82e17db` head to the
  verified proof head; no verified work was discarded.

## Parent Orchestration

- Orchestrator: canonical inline delivery worker
- State ledger: `.planning/phases/issue-3976-postgres-dynamic-catalog/RUN-STATE.json`
- Worker handoff: this phase directory
- Merge owner: human/captain for parent PR
- Integration state: held until corrected #4058 is green and merged

## Connector Implementation Scope

- Applies: yes
- Target connector scope: native PostgreSQL source catalog adapter only
- Connector-owned paths: `internal/connectors/native/postgres/**` and focused live-test evidence
- Ownership guard evidence: #3976 owns source catalog/type/fingerprint discovery; #3980/#3982/#3983/#3987 remain untouched
- Changed-path compliance: passed — only PostgreSQL source-catalog discovery,
  its focused tests, connector-description generated artifacts, and this GSD
  evidence changed. The inherited PostgreSQL surface audit found only
  test-only fixture schemas and test setup DDL; no live hard-coded table or
  column shape remains in #3976's discovery path.
- Foundation issue/PR path: #3974 / #4034 typed database foundation
- Shared runtime/tooling or unrelated connector changes: none planned
- no-mistakes foundation split status: not applicable; stop if a shared-foundation path is required

## Verification

RED is committed in `db7e06d36`; catalog GREEN is `24d0055f5`; live-read
RED/GREEN is recorded in `traces/live-reads-{red,green}.txt`. Focused PostgreSQL,
database-foundation, engine, and CLI tests, the PostgreSQL race suite, `go vet
./...`, build, lint, docs, smoke, contract, connector-generation/boundary, and
release workflow gates are green. The opt-in real Docker proof through Colima
discovered the seeded catalog, returned full IDs `1,2,3,4,5`, and returned only
`3,4,5` after cursor `10`:

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

The full proof is also posted on issue #3976.
The historical CDC integration test remains an intentional fail-closed skip,
not claimed coverage. No-mistakes is pending the final evidence commit.

## CI Status and External Snyk Gap

Every GitHub Actions check passed. The external, non-required
`security/snyk (karthik-sivadas)` status has reported one failed test, but no
target URL, log, or repository-local configuration is available to inspect or
reproduce its content. Under the recorded captain decision, it is not treated
as a merge blocker and remains a known external-service gap rather than a
clean result.

## Automated Review

- Primary route: pending until the child is non-draft and locally green
- Fallback route: Copilot only if Claude is unavailable and coverage blocks progress
- PR base/default branch: `integration/4015-mvp-flat-r1` / `main`
- Latest reviewed commit: pending
- Reviewed range: pending
- Coverage route: pending
- Coverage status: pending
- Disposition summary: pending
- Follow-up review status: pending
