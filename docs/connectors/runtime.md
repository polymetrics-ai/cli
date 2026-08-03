```
NAME
  pm runtime - inspect external runtime dependencies

SYNOPSIS
  pm runtime doctor [--json]

DESCRIPTION
  Checks PostgreSQL, DragonflyDB, and Temporal using the configured endpoints.
  Defaults match the local Compose stack in deploy/compose.

ENVIRONMENT
  POLYMETRICS_POSTGRES_URL
  POLYMETRICS_DRAGONFLY_ADDR
  POLYMETRICS_TEMPORAL_ADDR

SECURITY
  PostgreSQL passwords are redacted in command output.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
