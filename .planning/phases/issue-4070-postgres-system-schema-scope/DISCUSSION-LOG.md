# Issue #4070 discussion log

## Invocation

- Requested workflow: `discuss-phase issue-4070-postgres-system-schema-scope --auto`.
- Phase lookup: `phase_found: false`; this issue-specific correction has no
  numbered roadmap phase.
- Resolution: documented inline/manual GSD fallback, permitted by the project
  adapter when a compatible formal phase cannot be initialized.

## Auto-resolved decisions

| Area | Locked input | Selected implementation boundary |
| --- | --- | --- |
| System namespace scope | #4070 and the promoted Sol report name the exact schemas. | Reject only `pg_catalog`, `information_schema`, `pg_toast`, `pg_toast_*`, and `pg_temp_*` before a pool exists. |
| Error contract | The correction is a typed-catalog boundary, not a general connection error. | Use one named, identifier-free scope error that callers and tests can classify. |
| Dynamic behavior | #3976 requires live discovery rather than connector JSON or fixture substitution. | Preserve exact valid user-schema discovery through the existing catalog SQL and legacy projection. |
| Evidence | Ship handoff mandates a committed RED plus a fresh local PostgreSQL proof. | Commit tests before production code; run a unique loopback fixture through `pm catalog refresh` and PostgreSQL oracle queries. |
| Delivery | #3976 is parked at 5/5; #4070 starts independently at 0/5. | Carry `49a9386` on a fresh branch, then run one new no-mistakes pipeline after implementation. |

## Deferred / explicitly excluded

- Destination DDL/write, Parquet, CDC, outbound delivery, sync-mode apply,
  transport registration, and generic certification wiring are not part of #4070.
- Any actual registry wiring failure is routed to the existing #4015 owner,
  rather than widened here.
