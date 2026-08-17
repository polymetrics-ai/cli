---
name: pm-nexiopay
description: Nexio Pay connector knowledge and safe action guide.
---

# pm-nexiopay

## Purpose

Reads Nexio Pay card tokens, payout recipients, spendbacks, payment types, terminals, and the API user via the Nexio REST API.

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
- mode
- api_key (secret)
- username (secret)

## ETL Streams

- card_tokens:
  - primary key: key
  - fields: cardHolderName(), cardType(), createdDate(), currency(), expirationMonth(), expirationYear(), key(), lastFour()
- recipients:
  - primary key: recipientId
  - fields: createdDate(), currency(), email(), name(), recipientId(), status(), updatedDate()
- spendbacks:
  - primary key: id
  - fields: amount(), createdDate(), currency(), id(), recipientId(), status()
- payment_types:
  - primary key: id
  - fields: displayName(), enabled(), id(), name()
- terminal_list:
  - primary key: terminalId
  - fields: merchantId(), name(), status(), terminalId()
- user:
  - primary key: accountId
  - fields: accountId(), email(), merchantId(), role(), username()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Nexio Pay API read of card tokens, payout, and account data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect nexiopay
```

### Inspect as structured JSON

```bash
pm connectors inspect nexiopay --json
```

## Agent Rules

- Run pm connectors inspect nexiopay before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
