# GitHub Rails Pull Requests ETL Benchmark

Date: 2026-06-24

## Target

- Connector: `github`
- Repository: `rails/rails`
- Stream: `pull_requests`
- Pagination: `max_pages=0`, `per_page=100`
- ETL batch size: `100`
- Destination: local JSONL warehouse
- Runtime mode: dependency-free

## Result

```text
status: completed
records_read: 38715
records_transformed: 38715
records_loaded: 38715
records_failed: 0
batch_count: 388
warehouse_size: 62,875,797 bytes
real_time: 595.68s
user_time: 11.19s
sys_time: 1.78s
max_resident_set_size: 43,122,688 bytes
peak_memory_footprint: 29,869,312 bytes
```

Derived throughput:

```text
records_per_second: 64.99
batches_per_second: 0.65
average_records_per_batch: 99.78
```

## Notes

- The benchmark used an authenticated GitHub CLI token without printing or storing the token in benchmark output.
- Setup time for project initialization, credentials, connection creation, and catalog refresh was outside the timed ETL section.
- The timed section was only `pm etl run`.
- The result confirms the latest bounded-batch ETL path completed the full Rails PR history without the old page cap.
