---
phase: issue-4070-postgres-system-schema-scope
issue: 4070
status: passed_automated_boundary_uat
visual_review: not_applicable
---

# #4070 acceptance confirmation

All issue acceptance points are observable connector behavior rather than a
subjective UI workflow. They were exercised with a freshly built PM binary,
the production registry, and an independently queried unique local PostgreSQL
fixture:

1. Two materially different allowed user schemas yielded different server-
   derived catalogs without a connector JSON or code schema edit.
2. A supported catalog retained ordered identity, columns, nullability, and
   primary-key details; an allowed enum-backed catalog failed closed.
3. Exact `pg_catalog`, `information_schema`, `pg_toast` and prefixed
   `pg_toast_*`, `pg_temp_*` configuration values returned the safe typed scope
   error before a connection attempt.
4. A live temporary table's actual physical `pg_temp_N` namespace reproduced
   the defect before the Green change and was rejected afterward.
5. All run-owned schemas, temporary relations, database/process state, and
   scratch data were cleaned idempotently with zero remaining run-owned server
   objects.

There is no visual surface to review. The pending no-mistakes run and requested
draft PR are delivery gates, not unmet product acceptance.
