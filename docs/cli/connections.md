```
NAME
  pm connections - configure source-to-destination sync connections

SYNOPSIS
  pm connections create <name> --source connector:credential --destination connector:credential --stream stream [--sync-mode mode] [--cursor field] [--primary-key field] [--table table]
  pm connections list [--json]

DESCRIPTION
  A connection joins one source endpoint to one destination endpoint and stores
  stream-level sync settings.

SYNC MODES
  full_refresh_append              read all source records and append them
  full_refresh_overwrite           read all source records and replace final output
  full_refresh_overwrite_deduped   replace final output and keep latest row per primary key
  incremental_append               append records at or after the saved cursor
  incremental_append_deduped       append raw history and materialize latest row per primary key

  Incremental modes require --cursor. Deduped modes require --primary-key. When
  a connector manifest declares defaults, pm fills them during connection
  creation.

SECURITY
  Connections reference credentials by name only.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
