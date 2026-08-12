```
NAME
  pm connections - configure source-to-destination sync connections

SYNOPSIS
  pm connections create <name> --source connector:credential --destination connector:credential --stream stream [--sync-mode mode] [--cursor field] [--primary-key field] [--table table]
  pm connections list [--json]

DESCRIPTION
  A connection joins one source endpoint to one destination endpoint and stores
  stream-level sync settings.

CONNECTION NAMES
  A name may contain letters, digits, '-' and '_', must start with a letter or
  digit, and is limited to 128 characters. Names that differ only by letter case
  are refused. Ambiguous names are rejected at creation rather than rewritten,
  because two connections that cannot be told apart cannot own separate data.
  The name is a display value: the local warehouse keys its directories on a
  generated identifier, so renaming is safe and never moves data.

STREAM AND TABLE NAMES
  Against the local warehouse destination, --stream and --table become path
  components, so each may contain only letters, digits, '.', '-' and '_'. They
  are checked when the connection is created, because a name the warehouse
  cannot materialize would otherwise fail every sync of that connection.

  One local-warehouse connection cannot configure distinct --table spellings
  that differ only by ASCII letter case, such as records and RECORDS: DuckDB
  treats them as one identifier. Creation refuses that inventory before saving
  it. A legacy inventory is left unchanged on open, but any local sync refuses
  before changing run or warehouse state; create replacement connections whose
  destination table names differ by more than ASCII letter case.

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
