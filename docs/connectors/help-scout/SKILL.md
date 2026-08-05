---
name: pm-help-scout
description: Help Scout connector knowledge and safe action guide.
---

# pm-help-scout

## Purpose

Reads Help Scout conversations, customers, mailboxes, and users through the Mailbox API using OAuth2 client-credentials authentication.

## Icon

- id: simple-icons-helpscout
- asset: icons/simple-icons/helpscout.svg
- title: Help Scout
- simple_icon_slug: helpscout
- simple_icon_hex: 1292EE
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Help%20Scout
- match: exact-name-or-slug
- matched_by: help-scout

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

## Mechanism

- kind: official_api
- sanctioned_by_provider: true (official)

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
