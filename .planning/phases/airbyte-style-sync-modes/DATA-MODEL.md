# Data Model

## Stream State

- `connection`
- `stream`
- `cursor`
- `generation_id`
- `last_successful_run_id`
- `records_loaded`
- `updated_at`

## Raw Record

- `_polymetrics_raw_id`
- `_polymetrics_run_id`
- `_polymetrics_sync_id`
- `_polymetrics_generation_id`
- `_polymetrics_extracted_at`
- `_polymetrics_loaded_at`
- `_polymetrics_cursor`
- `_polymetrics_primary_key`
- `_polymetrics_deleted`
- `record`

## Final Record

Final records are the original connector records enriched with Polymetrics metadata fields. Deduped final records are materialized from raw records.

