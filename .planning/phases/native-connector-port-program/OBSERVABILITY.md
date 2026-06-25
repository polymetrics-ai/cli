# Observability

This phase adds planning metadata, not long-running connector execution.

Future native ports must add metrics for:

- API calls, pages, bytes, records, retries, rate-limit waits
- CDC lag, checkpoint LSN/binlog GTID/resume token age
- write batch sizes, per-record failures, idempotency conflicts
- secret-redaction failures in test fixtures
