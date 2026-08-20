---
name: pm-etl
description: Run bounded ETL syncs from configured connections.
---

# pm-etl

- Use `pm etl run --connection <name> --stream <stream> --json`.
- Use `--batch-size` for large streams when the caller requests bounded memory behavior.
- Supported sync modes are `full_refresh_append`, `full_refresh_overwrite`, `full_refresh_overwrite_deduped`, `incremental_append`, and `incremental_append_deduped`.
- For the closed managed-PostgreSQL route, run `pm etl transport postgres-managed-target plan --connection <name> --stream <stream>`, preview the plan, then send its one-time token through `pm etl run ... --approval-token-stdin --confirm destructive`; PostgreSQL and declared API sources use sealed catalogs, and callers never supply raw SQL or target identifiers.
- For a connector-declared typed destination, run `pm etl transport declarative-typed-destination plan --connection <name> --stream <stream>`, preview it, then use the ordinary approved ETL run. The saved stream's `destination_action` selects the connector-owned `writes.json` action; never accept it, a route, method, body, mapping, connector, or evidence as a run-time flag.
- In `pm etl run --json` and `pm etl status --json`, inspect `run.destination_results` for each acknowledged typed action's complete provider result. Provider-returned fields, keys, and values remain verbatim even when they equal configured credential bytes; system-generated plans, logs, request diagnostics, and synthetic errors remain secret-taint-safe.
- Incremental modes and deduped compatibility names require a cursor. Deduped modes require a primary key; static manifests advertise the full deduped compatibility name only with both fields and incremental modes only with a declared incremental executor. The deduped compatibility names refuse before source I/O until a matching transport is admitted.
- Inspect `batch_count` and `checkpoint` in JSON output after runs.
