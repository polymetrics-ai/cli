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
  full_refresh_overwrite_deduped   compatibility name for typed full_overwrite admission
  incremental_append               append records at or after the saved cursor
  incremental_append_deduped       compatibility name for typed incremental_dedupe admission

  Incremental modes and deduped compatibility names require --cursor. Deduped
  modes require --primary-key. A static connector manifest advertises the full
  deduped compatibility name only with both fields, and incremental modes only
  with a declared incremental executor. The two deduped compatibility names use
  their typed contract and refuse before source I/O until a matching transport
  is admitted. When a connector manifest declares defaults, pm fills them during
  connection creation.

POLLING-WATERMARK LIMITS
  polling_watermark is not a general connection mode and is not CDC. It can be
  selected only by a connector's declared native source/object/destination
  binding after runtime preflight succeeds. An admitted source resumes from its
  declared watermark plus unique tie-breaker after durable downstream
  acknowledgement, so replay is at least once. A polling scan cannot observe a
  hard delete after the row is gone; delete-aware history requires a declared
  cursor-advancing soft-delete mapping. Incompatible state, source identity
  changes, snapshot expiry, and retention failures require explicit
  rebootstrap rather than an automatic full scan.

SECURITY
  Connections reference credentials by name only.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
