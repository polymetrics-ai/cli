---
name: pm-interzoid
description: Interzoid connector knowledge and safe action guide.
---

# pm-interzoid

## Purpose

Reads Interzoid data-matching lookups: company-name, individual-name, and street-address similarity keys, plus organization-name standardization, via the Interzoid REST API.

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

- address
- address_match_algorithm
- base_url
- company
- company_match_algorithm
- fullname
- org
- api_key (secret)

## ETL Streams

- company_name_matching:
  - primary key: SimKey
  - fields: Code(), Credits(), SimKey(), query_company()
- individual_name_matching:
  - primary key: SimKey
  - fields: Code(), Credits(), SimKey(), query_fullname()
- street_address_matching:
  - primary key: SimKey
  - fields: Code(), Credits(), SimKey(), query_address()
- standardize_company_names:
  - primary key: Standard
  - fields: Code(), Credits(), Standard(), query_org()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Interzoid API single-lookup read; each read spends an API credit
- approval: none; read-only data-matching lookup API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect interzoid
```

### Inspect as structured JSON

```bash
pm connectors inspect interzoid --json
```

## Agent Rules

- Run pm connectors inspect interzoid before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
