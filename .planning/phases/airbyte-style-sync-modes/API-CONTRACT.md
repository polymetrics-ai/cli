# API Contract

## Stream Config

`StreamConfig` continues to expose:

- `sync_mode`
- `cursor_field`
- `primary_key`
- `destination_table`

The accepted `sync_mode` values are:

- `full_refresh_append`
- `full_refresh_overwrite`
- `full_refresh_overwrite_deduped`
- `incremental_append`
- `incremental_append_deduped`

## Run Output

`Run.Checkpoint` may include:

- `records_read`
- `records_transformed`
- `records_loaded`
- `records_failed`
- `batches`
- `cursor`
- `sync_mode`
- `state_key`
- `generation_id`

JSON output remains deterministic and contains no secret values.

