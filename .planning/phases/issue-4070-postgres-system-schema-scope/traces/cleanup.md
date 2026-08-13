# Run-owned PostgreSQL cleanup

## Proven object ownership

The only database resources removed were created by the #4070 fixture:

- schemas `audit_alpha_4070`, `audit_beta_4070`, and
  `audit_unsupported_4070`;
- database `pm4070_red_catalog`;
- bootstrap role, data directory, PM project, binaries, and log contained
  under one unique task-owned scratch directory;
- temporary probe tables that belonged to sessions already closed.

No ambient credentials, shared database, shared container, or pre-existing
schema was inspected or touched.

## Commands and observations

```sh
DROP SCHEMA <run-owned-schema> CASCADE
SELECT count(*) FROM pg_namespace WHERE nspname IN (<run-owned-schemas>)
DROP SCHEMA IF EXISTS <run-owned-schema> CASCADE
DROP DATABASE <run-owned-database>
DROP DATABASE IF EXISTS <run-owned-database>
pg_ctl -D <run-owned-data-dir> -w stop
trash <run-owned-scratch-dir>
```

- First-drop PostgreSQL oracle count: `0` remaining run-owned schemas.
- Repeated schema and database drops returned only expected `IF EXISTS`
  notices; cleanup is idempotent.
- `pg_ctl status` after stop confirmed no server for the run-owned data
  directory.
- The validated scratch target matched only
  `.scratch/issue-4070-red.*`, contained its own `postgres-data/PG_VERSION`
  marker, and was moved to the operating-system Trash. The worktree has zero
  run-owned residue; the final move is recoverable.
- A PostgreSQL query after each temporary-session close returned `0` matching
  temporary probe relations.
