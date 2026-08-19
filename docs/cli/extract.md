```
NAME
  pm extract - route a natural-language data request safely

SYNOPSIS
  pm extract --request <text> [--sql query] [--limit n] [--in table] [--out table] [--json]

DESCRIPTION
  Classifies a bounded natural-language request and routes it to a typed local
  query or RLM analysis path. It never accepts an unrestricted shell, HTTP, or
  SQL write operation.

OPTIONS
  --request text       request to classify
  --sql query          optional validated local query
  --limit n            maximum returned rows; default 100
  --in table           source table for an executable RLM route
  --out table          destination table for an executable RLM route
  --json               render machine-readable JSON

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
