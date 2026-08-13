# RED — live PostgreSQL temporary-schema boundary

**Candidate:** freshly built `pm` from the unmodified #4070 candidate.
**Production code modified:** no.
**Boundary:** a unique run-owned PostgreSQL 12 cluster on loopback only; no
ambient PostgreSQL environment variables, shared database, raw connection
string, or secret value was used or recorded.

## Safe command shapes

```sh
initdb -D <run-owned-data-dir> --auth=trust --username=<run-owned-role>
pg_ctl -D <run-owned-data-dir> -o '-h 127.0.0.1 -p <run-port>' -w start
pm init
PM_4070_TEST_PASSWORD=<redacted> pm credentials add <name> \
  --connector postgres --from-env password=PM_4070_TEST_PASSWORD \
  --config host=127.0.0.1 --config port=<run-port> \
  --config database=<run-owned-db> --config username=<run-owned-role> \
  --config schema=<schema> --config sslmode=disable
pm catalog refresh --connection <name> --json
```

## Server-owned oracle setup

- Created allowed schemas `audit_alpha_4070` and `audit_beta_4070` with
  materially different base-table/column/key shapes.
- Created `audit_unsupported_4070.events` with an enum column to retain the
  existing fail-closed unsupported-shape proof.
- In a second still-open session, created temporary base table
  `catalog_scope_temp_probe_4070` and read its physical namespace from
  `pg_my_temp_schema()`: `pg_temp_3` for this run.
- A separate PostgreSQL catalog query observed
  `pg_temp_3.catalog_scope_temp_probe_4070:r` while the owning session remained
  open. Stock system relation counts were nonzero for `pg_catalog`,
  `information_schema`, and `pg_toast`.

## Shipping-path observations

| Production invocation | Observed candidate result | Independent oracle |
| --- | --- | --- |
| `pm catalog refresh --connection pg4070-alpha --json` | `audit_alpha_4070.accounts` (four fields, primary key `account_id`) and `audit_alpha_4070.audit` | `information_schema` returned the matching ordered names, ordinal positions, and nullable flags; the view remained absent. |
| `pm catalog refresh --connection pg4070-beta --json` | only `audit_beta_4070.accounts`, with a two-column primary key and different fields | `information_schema` returned the different ordered relation shape. |
| `pm catalog refresh --connection pg4070-unsupported --json` | explicit unsupported-catalog-shape error | Enum-based allowed-schema object remained fail-closed. |
| `pm catalog refresh --connection pg4070-temp --json` while the held temp session was open | **incorrect success:** stream `pg_temp_3.catalog_scope_temp_probe_4070` with `probe_id`, `marker`, and primary key `probe_id` | PostgreSQL catalog query observed the same live temporary relation. |

The alpha/beta results prove the binary uses dynamic source metadata rather than
a connector JSON or canned static table list: two independently created allowed
schemas produced different catalogs without any connector code or JSON edit.
The temporary-table success reproduces the #4070 defect at the real PM/registry
boundary; it is not a SQL-query-only result.

The temporary session remains open only until the subsequent Green rerun begins;
final idempotent cleanup and zero-residue proof are recorded in verification.
