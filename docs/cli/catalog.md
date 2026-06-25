```
NAME
  pm catalog - discover and display source streams

SYNOPSIS
  pm catalog refresh --connection <name> [--json]
  pm catalog show --connection <name> [--json]

DESCRIPTION
  Catalog commands call the source connector and store a local snapshot.

SECURITY
  Catalog output includes schemas and stream names, never secret values.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
