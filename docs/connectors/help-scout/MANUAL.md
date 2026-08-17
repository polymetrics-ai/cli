# pm connectors inspect help-scout

```text
NAME
  pm connectors inspect help-scout - Help Scout connector manual

SYNOPSIS
  pm connectors inspect help-scout
  pm connectors inspect help-scout --json
  pm credentials add <name> --connector help-scout [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Help Scout conversations, customers, mailboxes, and users through the Mailbox API using OAuth2 client-credentials authentication.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  start_date
  token_url
  client_id (secret)
  client_secret (secret)

ETL STREAMS
  conversations:
    primary key: id
    cursor: userUpdatedAt
    fields: assigneeId(), closedAt(), createdAt(), folderId(), id(), mailboxId(), number(), preview(), state(), status(), subject(), threads(), type(), userUpdatedAt()
  customers:
    primary key: id
    cursor: updatedAt
    fields: age(), createdAt(), firstName(), gender(), id(), jobTitle(), lastName(), organization(), photoUrl(), updatedAt()
  mailboxes:
    primary key: id
    cursor: updatedAt
    fields: createdAt(), email(), id(), name(), slug(), updatedAt()
  users:
    primary key: id
    cursor: updatedAt
    fields: createdAt(), email(), firstName(), id(), jobTitle(), lastName(), role(), timezone(), type(), updatedAt()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Help Scout API read of conversation, customer, mailbox, and user data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect help-scout

  # Inspect as structured JSON
  pm connectors inspect help-scout --json

AGENT WORKFLOW
  - Run pm connectors inspect help-scout before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
