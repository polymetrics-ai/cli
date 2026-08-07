# TDD ledger — warehouse connection namespacing

Phase: `cli-warehouse-connection-namespacing-r1`
PR: #3901 · Branch: `fm/cli-warehouse-connection-namespacing-r1`

## GSD provenance

**Manual-GSD fallback.** This work ran inline rather than through `scripts/gsd`: it is a single
focused defect fix on a release blocker, dispatched directly with a fixed brief and driven through
the `no-mistakes` pipeline for review, test, document, lint, push, PR and CI gates. The canonical
contract permits inline execution when the runtime cannot provide compatible isolated agents or
when spawning is not warranted; this is recorded here as that fallback rather than left implicit.

Review evidence is not self-attested: the change went through **three** automated review rounds and
two document rounds inside the pipeline, with every actionable finding either fixed at the root or
explicitly dispositioned by the maintainer. Two consequences were deliberately deferred and filed
as their own issues (#3897, #3900) rather than silently dropped.

## Red → Green

The regression was reproduced as a **failing test first**, before any fix existed.

### Red

`TestSecondConnectionDoesNotDestroyFirstConnectionRows`
(`internal/app/warehouse_connection_isolation_test.go`) — two connections with distinct
credentials, both `incremental_append_deduped`, both materializing a table named `records`.

```
--- FAIL: TestSecondConnectionDoesNotDestroyFirstConnectionRows
    row "a1" is missing after the second connection synced; warehouse rows = [g1]
```

Committed red at `test(warehouse): reproduce cross-connection warehouse data loss`. This matches
the independently measured behaviour: after the acme sync 2 rows, after the globex sync 1 row —
acme's rows destroyed, no error, exit status 0.

### Green

`feat(warehouse): nest materialization under workspace, connector and connection` turns it green.
Tables now resolve to `<workspace-id>/<connector>/<connection-id>/tables/<table>.jsonl`, so two
connections never share a parent directory and the collision is unrepresentable.

The same test also asserts that **no single table file holds two connections' rows**, which covers
the non-deduped append failure mode (silent interleaving) as well as the deduped one (outright
destruction).

## Subsequent red/green cycles

Each review finding that was fixed got its own regression test before or alongside the fix:

| Behaviour pinned | Test |
|---|---|
| Both tenants intact; unscoped read refused; `--connection` resolves | `TestBothConnectionsKeepTheirOwnRowsAndAreReadableByName` |
| Ownership mismatch is a hard error and changes no rows | `TestSyncRefusesWhenAnotherConnectionOwnsTheDirectory` |
| Legacy flat layout refused, nothing deleted | `TestSyncRefusesLegacyFlatWarehouse` |
| Ambiguous/case-colliding connection names refused at creation | `TestCreateConnectionRejectsAmbiguousNames` |
| The five measured name collisions cannot all remain valid | `TestValidateConnectionNameRejectsNamesThatFoldTogether` |
| Path components leak no display name or credential material | `TestConnectionIdentityIsOpaqueAndNotDerivedFromNameOrCredential` |
| A reintroduced shared table path fails loudly | `TestAssertOwnedTableCatchesAReintroducedSharedPath` |
| One damaged ownership record does not deny healthy connections | `TestOneDamagedOwnershipRecordDoesNotDenyHealthyConnections` |
| A damaged record is reported, never reported as an absent table | `TestTablesReportsAnUnreadableOwnershipRecord` |
| A path with no selector names no flag it cannot accept | `TestQuerySQLAmbiguityNamesNoSelectorItCannotAccept` |
| `ValidateWrite` cannot drift from `Write` | `TestWarehouseValidateWriteAgreesWithWrite` |

**No test was weakened, skipped, or deleted at any point.** Three existing directory-sync-chain
tests and one reverse-ETL helper were updated to the new layout with their intent preserved; the
two-writer sync-chain case was strengthened, since each writer now asserts its own distinct chain
rather than sharing one expectation.

## Local verification

`gofmt` · `go vet` on the default and `-tags duckdb` builds · `go build ./cmd/pm` ·
`internal/app` (203s) · `internal/warehouse` · `internal/connectors` · `internal/cli` (611s) ·
`internal/flow` · `internal/rlm` · `internal/perf` · `internal/state` · `internal/durability` ·
`make tidy-check` · `make lint` · `make docs-check` · `make smoke-no-build` ·
`make agent-contract-check` · `make connectorgen-validate` · `make connectorgen-surface-sync` ·
`make connector-boundary` · `make release-workflow-check` · `git diff --check`.

Note: `go test` was run with `-timeout 20m`. `internal/cli` exceeds Go's 10-minute default on a
loaded machine and the resulting timeout panic reads exactly like a hang — this was hit once and
is now recorded in `AGENTS.md`.

## End-to-end proof against the real binary

Beyond the unit tests, the built `pm` binary was exercised directly:

- two connections keep their own rows; the layout nests as designed with `owner.json` written
- an unscoped read of a shared table name is refused and names both owners
- `--connection acme`, `--connection globex` and `--connection _unattributed` each resolve
- `acme.prod`, `acme prod`, `acme:prod`, `acme/prod`, `acme#prod` and `ACME` are all refused at
  creation
- a legacy `_pm_raw` warehouse is refused on **both** the read and write paths, with its files
  left untouched
- with one connection's `owner.json` deliberately corrupted, the healthy connection still reads
  while the damaged one surfaces a named fault with a recovery step
