---
name: pm-dixa
description: Dixa connector knowledge and safe action guide.
---

# pm-dixa

## Purpose

Reads Dixa conversations (and their queue, rating, and assignment projections) from the Dixa conversation_export API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- asset: icons/dixa.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.dixa.io/openapi/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- batch_size
- mode
- start_date
- api_token (secret)

## ETL Streams

- conversations:
  - primary key: id
  - cursor: updated_at
  - fields: closed_at(), created_at(), direction(), handling_duration(), id(), initial_channel(), last_message_created_at(), originating_country(), requester_email(), requester_id(), requester_name(), status(), subject(), total_duration(), updated_at()
- conversation_queue:
  - primary key: id
  - cursor: updated_at
  - fields: direction(), id(), initial_channel(), queue_id(), queue_name(), queued_at(), updated_at()
- conversation_rating:
  - primary key: id
  - cursor: updated_at
  - fields: id(), rating_message(), rating_score(), status(), updated_at()
- conversation_assignment:
  - primary key: id
  - cursor: updated_at
  - fields: assigned_at(), assignee_email(), assignee_id(), assignee_name(), id(), transfer_time(), transferee_name(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Dixa API reads performed by the legacy connector via a Tier-2 hook
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect dixa
```

### Inspect as structured JSON

```bash
pm connectors inspect dixa --json
```

## Agent Rules

- Run pm connectors inspect dixa before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
