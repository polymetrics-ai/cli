# #4070 TDD ledger

**Fresh correction counter:** **0 / 5**. This is independent of the immutable
#3976 5/5 run. No sixth #3976 loop is permitted.

| Cycle | State | Evidence required before advancing |
| --- | --- | --- |
| RED | complete in `e23a945` / `d890902` | A test-only change proves every reserved schema currently reaches transport/catalog discovery; a fresh run-owned PostgreSQL session proves a held `pg_temp_N` table is accepted by the candidate. See `traces/red-unit.md` and `traces/red-live-boundary.md`. |
| GREEN | complete; implementation commit pending | The narrow guard returns a named typed scope error before pool creation for every required exact/prefix case; allowed schemas remain live/dynamic. See `traces/green-live-boundary.md`. |
| Refactor/docs | complete; implementation commit pending | No refactor beyond one narrow helper was needed. Authoritative PostgreSQL docs and derived website catalog data state the same boundary. |
| Verify/review | blocked on refactor | Focused/unit/live/race/build/vet/lint/docs/generator/diff checks, GSD verification, and deep review are recorded. |
| No-mistakes | blocked on verification | One new #4070 `axi run` is driven to a terminal ready state, maximum five correction loops, with no out-of-pipeline edits while active. |

## Red contract

The RED is behavioral, not a compile-only reference to a future sentinel. It
uses an otherwise valid live configuration with an unreachable loopback target:
the current candidate reports a pool/transport failure instead of an
identifier-free system-schema scope outcome. The database-integration RED holds
a real temporary relation in `pg_temp_N`; the current candidate returns live
metadata for it. The captured commands and outcomes are recorded without raw
connection strings or secret values.

## Green contract

The green path must reject before `newTypedCatalogResources`, operation-context
creation, and `openTypedCatalogPool`. It returns one named scope error whose
message contains no configured schema or connection material. It must not
change query results for an allowed exact user schema.
