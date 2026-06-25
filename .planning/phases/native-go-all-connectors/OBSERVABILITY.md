# Observability Plan

## Local Evidence

- Native conformance reports include slug, runtime family, fixture mode, test results, and benchmark hooks.
- ETL and reverse ETL run records retain counts, status, checkpoints, and errors.
- Native destination receipts provide per-run evidence for fixture-backed writes.

## Future Metrics

- Per connector: check latency, catalog latency, read records/sec, write records/sec, query latency, CDC event lag.
- Per runtime family: retry count, rate-limit waits, error class counts, and conformance failure count.
