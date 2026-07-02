```
NAME
  pm query - inspect local warehouse data

SYNOPSIS
  pm query run --table <table> [--limit n] [--json]
  pm query run --sql "select * from <table> limit n" [--json]
  pm query run --table <table> --agent-mode summary --fields id,email --sample 3
  pm query run --table <table> --agent-mode stream --fields id,email

DESCRIPTION
  The MVP query engine supports table reads and a small SELECT * FROM parser.
  Agent mode can emit compact summary JSON or projected NDJSON rows to reduce
  token usage for external agents.

FLAGS
  --table table              local warehouse table to scan
  --sql sql                  read-only SQL query; takes precedence over --table
  --limit n                  maximum rows to read; default 100
  --fields a,b               project output to selected fields
  --agent-mode summary       emit a count, sorted field list, and sample rows
  --agent-mode stream        emit one projected JSON object per line
  --sample n                 summary sample size; default 3

SECURITY
  Query output can contain data rows. Agent callers should use --fields and
  small limits or --agent-mode summary.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
