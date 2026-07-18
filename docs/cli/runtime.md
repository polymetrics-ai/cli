```
NAME
  pm runtime - inspect external runtime dependencies

SYNOPSIS
  pm runtime doctor [--json]

DESCRIPTION
  Checks PostgreSQL, DragonflyDB, and Temporal using the configured endpoints.
  Defaults match the local Compose stack in deploy/compose. Runtime doctor does
  not require live services or credentials; unavailable optional services are
  reported as per-check error status and degraded mode in the output.

ENVIRONMENT
  POLYMETRICS_POSTGRES_URL
  POLYMETRICS_DRAGONFLY_ADDR
  POLYMETRICS_TEMPORAL_ADDR

SECURITY
  Reported endpoints are sanitized before command output. Userinfo, query
  strings, fragments, and control characters are removed from PostgreSQL,
  DragonflyDB, and Temporal endpoints.

EXIT STATUS
  0 report emitted, including degraded or absent optional services
  2 usage error

```
