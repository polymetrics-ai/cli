```
NAME
  pm query - inspect local warehouse data

SYNOPSIS
  pm query run --table <table> [--limit n] [--json]
  pm query run --sql "select * from <table> limit n" [--json]

DESCRIPTION
  The MVP query engine supports table reads and a small SELECT * FROM parser.

SECURITY
  Query output can contain data rows. Agent callers should use small limits.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
