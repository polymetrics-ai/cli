---
name: pm-etl
description: Run bounded ETL syncs from configured connections.
---

# pm-etl

- Use `pm etl run --connection <name> --stream <stream> --json`.
- Use `--batch-size` for large streams when the caller requests bounded memory behavior.
- Supported sync modes are `full_refresh_append`, `full_refresh_overwrite`, `full_refresh_overwrite_deduped`, `incremental_append`, and `incremental_append_deduped`.
- Incremental modes and deduped compatibility names require a cursor. Deduped modes require a primary key; static manifests advertise the full deduped compatibility name only with both fields and incremental modes only with a declared incremental executor. The deduped compatibility names refuse before source I/O until a matching transport is admitted.
- Inspect `batch_count` and `checkpoint` in JSON output after runs.
