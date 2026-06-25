# SPEC: Airbyte-Style Sync Modes

## Behavior

`pm` normalizes every configured stream mode into:

- source sync mode: `full_refresh` or `incremental`
- destination sync mode: `append`, `overwrite`, `append_dedup`, or internal `overwrite_dedup`

Incremental modes require a cursor. Deduped modes require a primary key. Deduped final output keeps the newest record per primary-key tuple by cursor value, extracted timestamp, then raw ID.

## Local Warehouse Semantics

- Append modes append to the final table.
- Overwrite modes write to a temp final file and atomically rename on success.
- Deduped modes write accepted records to raw JSONL and materialize the final file from raw records.
- Full-refresh deduped modes materialize only from the current generation.
- Incremental deduped modes materialize from accumulated raw history.
- Delete/tombstone records are omitted from final output when they are the newest record for a primary key.

## State

Per-stream state stores cursor value, generation ID, last successful run ID, and counters. State advances only after successful finalization.

## Compatibility

The existing `full_refresh_overwrite` value remains valid. Connections without a mode still default to `full_refresh_overwrite`.

