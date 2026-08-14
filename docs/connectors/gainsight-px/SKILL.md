---
name: pm-gainsight-px
description: Gainsight PX connector knowledge and safe action guide.
---

# pm-gainsight-px

## Purpose

Reads Gainsight PX accounts, users, features, and segments through the aptrinsic REST API (read-only).

## Icon

- id: gainsight-px
- asset: icons/gainsight-px.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://support.gainsight.com/PX/API_for_Developers/02Usage_of_Different_APIs

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- mode
- page_size
- api_key (secret) (required)

## ETL Streams

- accounts:
  - primary key: id
  - fields: createDate(string), id(string), industry(string), lastModifiedDate(string), lastSeenDate(string), location(string), name(string), numberOfEmployees(string), numberOfUsers(string), plan(string), sfdcId(string), trackedSubscriptionId(string), website(string)
- users:
  - primary key: id
  - fields: accountId(string), aptrinsicId(string), createDate(integer), email(string), firstName(string), id(string), lastModifiedDate(integer), lastName(string), lastSeenDate(integer), numberOfVisits(integer), role(string), score(number), signUpDate(integer), title(string), type(string)
- feature:
  - primary key: id
  - fields: id(string), name(string), parentFeatureId(string), propertyKey(string), status(string), type(string)
- segments:
  - primary key: id
  - fields: createdBy(string), createdDate(string), description(string), id(string), modifiedBy(string), modifiedDate(string), name(string), priority(string), productId(string), productName(string), status(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Gainsight PX (aptrinsic) API read of account, user, feature, and segment data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect gainsight-px
```

### Inspect as structured JSON

```bash
pm connectors inspect gainsight-px --json
```

## Agent Rules

- Run pm connectors inspect gainsight-px before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
