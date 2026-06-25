# Observability

## Local Run Metadata

Each ETL run records:

- status
- counts
- batch count
- checkpoint map
- error when failed
- timestamps

## Phase Metrics

Benchmark coverage records elapsed time and records per second for append and deduped sync paths.

## Deferred Runtime Metrics

PostgreSQL, DragonflyDB, and Temporal metrics are outside this phase because the dependency-free JSONL semantics are the implementation target.

