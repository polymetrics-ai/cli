---
name: pm-fillout
description: Fillout connector knowledge and safe action guide.
---

# pm-fillout

## Purpose

Reads Fillout forms and manages webhooks/submission deletion through the Fillout REST API. Question definitions and submissions LIST remain on the legacy connector pending an engine fan_out fallback-mode gap (see docs.md Known limits).

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

- base_url
- mode
- api_key (secret)

## ETL Streams

- forms:
  - primary key: id
  - fields: id(), name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_webhook:
  - endpoint: POST /webhook/create
  - risk: registers a new outbound webhook subscription that will POST live form-submission data to an external URL; external mutation, approval required
- remove_webhook:
  - endpoint: POST /webhook/delete
  - risk: permanently removes a webhook subscription; event delivery to its target URL stops immediately
- delete_submission_by_id:
  - endpoint: DELETE /forms/{{ record.form_id }}/submissions/{{ record.submission_id }}
  - required fields: form_id, submission_id
  - risk: permanently deletes a single form response; irreversible, approval required

## Security

- read risk: external Fillout API read of form metadata
- write risk: creates/removes outbound webhook subscriptions and deletes individual form submissions; external mutation, approval required
- approval: required for write actions; none for read
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect fillout
```

### Inspect as structured JSON

```bash
pm connectors inspect fillout --json
```

## Agent Rules

- Run pm connectors inspect fillout before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
