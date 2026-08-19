---
name: pm-interzoid
description: Interzoid connector knowledge and safe action guide.
---

# pm-interzoid

## Purpose

Reads Interzoid data-matching lookups: company-name, individual-name, and street-address similarity keys, plus organization-name standardization, via the Interzoid REST API.

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

- address
- address_match_algorithm
- base_url
- company
- company_match_algorithm
- fullname
- org
- api_key (secret) (required)

## ETL Streams

- company_name_matching:
  - primary key: SimKey
  - fields: Code(string), Credits(string), SimKey(string), query_company(string)
- individual_name_matching:
  - primary key: SimKey
  - fields: Code(string), Credits(string), SimKey(string), query_fullname(string)
- street_address_matching:
  - primary key: SimKey
  - fields: Code(string), Credits(string), SimKey(string), query_address(string)
- standardize_company_names:
  - primary key: Standard
  - fields: Code(string), Credits(string), Standard(string), query_org(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

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
