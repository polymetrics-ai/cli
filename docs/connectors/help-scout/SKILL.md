---
name: pm-help-scout
description: Help Scout connector knowledge and safe action guide.
---

# pm-help-scout

## Purpose

Reads Help Scout conversations, customers, mailboxes, and users through the Mailbox API using OAuth2 client-credentials authentication.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- start_date
- token_url
- client_id (secret)
- client_secret (secret)

## ETL Streams

- conversations:
  - primary key: id
  - cursor: userUpdatedAt
  - fields: assigneeId(), closedAt(), createdAt(), folderId(), id(), mailboxId(), number(), preview(), state(), status(), subject(), threads(), type(), userUpdatedAt()
- customers:
  - primary key: id
  - cursor: updatedAt
  - fields: age(), createdAt(), firstName(), gender(), id(), jobTitle(), lastName(), organization(), photoUrl(), updatedAt()
- mailboxes:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(), email(), id(), name(), slug(), updatedAt()
- users:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(), email(), firstName(), id(), jobTitle(), lastName(), role(), timezone(), type(), updatedAt()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Help Scout API read of conversation, customer, mailbox, and user data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect help-scout
```

### Inspect as structured JSON

```bash
pm connectors inspect help-scout --json
```

## Agent Rules

- Run pm connectors inspect help-scout before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
