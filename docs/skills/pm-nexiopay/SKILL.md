---
name: pm-nexiopay
description: Nexio Pay connector knowledge and safe action guide.
---

# pm-nexiopay

## Purpose

Reads Nexio Pay card tokens, payout recipients, spendbacks, payment types, terminals, and the API user via the Nexio REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- api_key (secret) (required)
- username (secret) (required)

## ETL Streams

- card_tokens:
  - primary key: key
  - fields: cardHolderName(string), cardType(string), createdDate(string), currency(string), expirationMonth(string), expirationYear(string), key(string), lastFour(string)
- recipients:
  - primary key: recipientId
  - fields: createdDate(string), currency(string), email(string), name(string), recipientId(string), status(string), updatedDate(string)
- spendbacks:
  - primary key: id
  - fields: amount(number), createdDate(string), currency(string), id(string), recipientId(string), status(string)
- payment_types:
  - primary key: id
  - fields: displayName(string), enabled(boolean), id(string), name(string)
- terminal_list:
  - primary key: terminalId
  - fields: merchantId(string), name(string), status(string), terminalId(string)
- user:
  - primary key: accountId
  - fields: accountId(string), email(string), merchantId(string), role(string), username(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

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
