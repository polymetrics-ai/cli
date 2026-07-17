```
NAME
  pm perf - compare dependency-free and dependency-backed runtime paths

SYNOPSIS
  pm perf compare [--iterations n] [--runtime] [--json]
  pm perf sync-modes [--records n] [--json]

DESCRIPTION
  Runs repeated local ETL loops and reports elapsed time, average operation time,
  and records per second. Without --runtime, only the dependency-free path runs.
  With --runtime, the command also checks PostgreSQL, DragonflyDB, and Temporal,
  acquires a Dragonfly lease, appends a PostgreSQL ledger record, and compares
  that path against the dependency-free baseline.

  The sync-modes subcommand runs a synthetic local file-to-warehouse benchmark
  for every supported ETL sync mode and reports each mode's duration and records
  per second. Workload flags are bounded: --iterations accepts 1 through 1000,
  and --records accepts 1 through 100000.

  With --runtime --json, output also includes runtime_report with PostgreSQL,
  DragonflyDB, and Temporal health-check metadata: check name, status, endpoint,
  latency, and redacted error text when a check fails. PostgreSQL endpoint is redacted
  and omits query strings. DragonflyDB and Temporal endpoints are topology metadata,
  not credentials.

SECURITY
  Performance output includes counts, durations, and, when --runtime is used,
  runtime health metadata. No decrypted secrets are printed. Runtime errors use
  the shared redaction path before JSON or text output.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
  3 validation error, including invalid --iterations or --records

```
