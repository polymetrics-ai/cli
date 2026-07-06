---
name: pm-vwo
description: VWO connector knowledge and safe action guide.
---

# pm-vwo

## Purpose

Reads and writes VWO (Visual Website Optimizer) A/B testing campaigns.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_id
- base_url
- campaign_type
- page_size
- platform
- start_date
- api_key (secret)

## ETL Streams

- campaigns:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), id(), name(), status()
- campaign_variations:
  - primary key: campaign_id, id
  - fields: campaign_id(), id(), is_control(), is_disabled(), name(), percent_split(), platform()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_campaign:
  - endpoint: POST /accounts/{{ config.account_id }}/campaigns
  - risk: creates a new A/B testing campaign visible to the workspace; external mutation, approval required
- update_campaign:
  - endpoint: PATCH /accounts/{{ config.account_id }}/campaigns/{{ record.id }}
  - required fields: id
  - risk: updates an existing campaign's configuration (name, status, traffic allocation, etc.); can start/pause/stop a live experiment; external mutation, approval required

## Security

- read risk: external VWO API read of experiment/campaign metadata
- write risk: external mutation of VWO campaign configuration, including starting/pausing/stopping live A/B experiments; approval required
- approval: read: none, read-only campaign directory sync. write: required for all mutating actions (create/update campaigns).
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect vwo
```

### Inspect as structured JSON

```bash
pm connectors inspect vwo --json
```

## Agent Rules

- Run pm connectors inspect vwo before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
