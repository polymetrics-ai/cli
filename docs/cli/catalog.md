```
NAME
  pm catalog - discover and display source streams

SYNOPSIS
  pm catalog refresh --connection <name> [--json]
  pm catalog show --connection <name> [--json]

DESCRIPTION
  Catalog refresh calls the source connector and stores a local snapshot;
  catalog show reads that persisted snapshot.
  refresh deliberately fetches a new provider catalog; show reads the
  existing snapshot and marks it stale when its discovery expiry has passed.
  Refresh a stale catalog before relying on fields added by the provider.

SECURITY
  Catalog output includes schemas and stream names, never secret values.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
