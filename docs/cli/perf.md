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
  per second.

SECURITY
  Performance output contains counts and durations only.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
