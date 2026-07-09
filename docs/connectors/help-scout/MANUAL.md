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
  write risk: documented Help Scout mutations are tracked as blocked-by-default operation metadata until typed reverse-ETL actions and approval policies are implemented
  approval: no runtime writes are executable in this slice; future reverse-ETL writes must use plan, preview, approval, and execute
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Inspect and plan Help Scout Inbox API operations safely.
  Usage: pm help-scout <resource> <action> [flags]
  Source CLI: Help Scout Inbox API (https://developer.helpscout.com/mailbox-api/)
  Global flags:
    --json (boolean): Write machine-readable JSON output.
    --connection (string): Use a saved Help Scout connector credential.: maps_to=connection
  Core record streams
    conversation list - List Help Scout conversations as an ETL stream. [intent=etl availability=implemented stream=conversations]; flags: --modified-since
    conversation view - View one conversation. [intent=direct_read availability=planned]; notes: Planned for #217 as a bounded direct read; not executable in this metadata slice.
    conversation thread list - List threads for one conversation. [intent=direct_read availability=planned]; notes: Planned for #217 as a bounded direct read.
    conversation reply create - Create a reply thread on a conversation. [intent=reverse_etl availability=planned]; approval: Must be implemented only as a typed reverse-ETL action with plan, preview, approval, and execute.; risk: Sends a visible customer-support reply through Help Scout.; notes: No runtime write action is declared in this slice.
    conversation note create - Create an internal note on a conversation. [intent=reverse_etl availability=planned]; approval: Must be implemented only as a typed reverse-ETL action with plan, preview, approval, and execute.; risk: Adds internal support context to a Help Scout conversation.; notes: No runtime write action is declared in this slice.
    customer list - List Help Scout customers as an ETL stream. [intent=etl availability=implemented stream=customers]; flags: --modified-since
    customer view - View one customer. [intent=direct_read availability=planned]; notes: Planned for #217 as a bounded direct read.
    customer create - Create a Help Scout customer. [intent=reverse_etl availability=planned]; approval: Must be implemented only as a typed reverse-ETL action with plan, preview, approval, and execute.; risk: Creates customer profile data in Help Scout.; notes: No runtime write action is declared in this slice.
    customer update - Update a Help Scout customer. [intent=reverse_etl availability=planned]; approval: Must be implemented only as a typed reverse-ETL action with plan, preview, approval, and execute.; risk: Mutates customer profile data in Help Scout.; notes: No runtime write action is declared in this slice.
    customer delete - Delete a Help Scout customer. [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Would require destructive typed confirmation and human approval if ever implemented.; risk: Deletes customer data from Help Scout.; notes: Blocked by default; #219 owns destructive/admin policy decisions.
    mailbox list - List Help Scout mailboxes as an ETL stream. [intent=etl availability=implemented stream=mailboxes]
    mailbox saved-reply list - List saved replies for a mailbox. [intent=direct_read availability=planned]; notes: Planned for direct-read or stream classification after operation-ledger review.
    user list - List Help Scout users as an ETL stream. [intent=etl availability=implemented stream=users]
    user create - Create a Help Scout user. [intent=reverse_etl availability=planned]; approval: Requires typed schema, preflight, preview, approval, and execute.; risk: Administrative user provisioning.; notes: No runtime write action is declared in this slice.
    user delete - Delete a Help Scout user. [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Would require destructive typed confirmation and human approval if ever implemented.; risk: Destructive user removal.; notes: Blocked by default; #219 owns destructive/admin policy decisions.
  Bounded direct reads planned for later lanes
    report company - Read company-level report metrics. [intent=direct_read availability=planned]; notes: Report endpoints require provider-specific query/body support and bounded output policy in #218.
  Approval-gated reverse ETL candidates
  Blocked by default until a policy lane implements safeguards
  Other Commands
    attachment download - Download or inspect a conversation attachment. [intent=direct_read availability=planned]; notes: Binary payload; blocked until #218 adds a bounded binary/output policy.
    tag list - List Help Scout tags. [intent=direct_read availability=planned]; notes: Candidate stream or direct read; #216 owns final operation classification.
    team list - List Help Scout teams. [intent=direct_read availability=planned]; notes: Candidate stream or direct read; #216 owns final operation classification.
    workflow list - List Help Scout workflows. [intent=direct_read availability=planned]; notes: Candidate stream or direct read; #216 owns final operation classification.
    webhook create - Create a Help Scout webhook. [intent=reverse_etl availability=planned]; approval: Must require typed schema, URL validation, preview, approval, and execute.; risk: Creates an outbound webhook configuration that can send Help Scout event data to another URL.; notes: No runtime write action is declared in this slice.
    workflow run - Run a Help Scout manual workflow. [intent=reverse_etl availability=planned]; approval: Must require plan, preview, approval, and execute.; risk: Triggers provider-side workflow automation.; notes: No runtime write action is declared in this slice.
    team member update - Update Help Scout team membership. [intent=reverse_etl availability=planned]; approval: Requires typed schema, preflight, preview, approval, and execute.; risk: Administrative membership change.; notes: No runtime write action is declared in this slice.
  Help topics:
    help-scout-auth - Configure Help Scout OAuth2 client credentials from environment variables or stdin; never put secrets in prompt text.
    help-scout-writes - Help Scout write candidates are metadata-only until typed reverse-ETL schemas and approval policies are implemented.

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
