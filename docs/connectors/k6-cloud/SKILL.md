---
name: pm-k6-cloud
description: k6 Cloud connector knowledge and safe action guide.
---

# pm-k6-cloud

## Purpose

Reads k6 Cloud organizations, projects, and load tests through the k6 Cloud REST API.

## Icon

- id: k6cloud
- asset: icons/k6cloud.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://k6.io/docs/cloud/cloud-reference/cloud-rest-api/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- page_size
- api_token (secret) (required)

## ETL Streams

- organizations:
  - primary key: id
  - fields: billing_address(string), billing_country(string), billing_email(string), created(string), description(string), id(integer), is_default(boolean), is_saml_org(boolean), name(string), owner_id(integer), updated(string), vat_number(string)
- k6_tests:
  - primary key: id
  - fields: created(string), id(integer), last_test_run_id(string), name(string), project_id(integer), script(string), test_run_ids(array), updated(string), user_id(integer)
- projects:
  - primary key: id
  - fields: created(string), description(string), id(integer), is_default(boolean), name(string), organization_id(integer), updated(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external k6 Cloud API read of organizations, projects, and load tests
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect k6-cloud
```

### Inspect as structured JSON

```bash
pm connectors inspect k6-cloud --json
```

## Agent Rules

- Run pm connectors inspect k6-cloud before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
