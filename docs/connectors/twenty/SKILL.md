---
name: pm-twenty
description: Twenty CRM connector knowledge and safe action guide.
---

# pm-twenty

## Purpose

Reads Twenty CRM companies, people, opportunities, notes, tasks, messages, calendar events, workflows, workspace members, and the rest of the 28-object workspace surface, and writes create/update/batch/delete mutations through the Twenty REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- replication_start_date
- api_key (secret)

## Security

- read risk: external Twenty CRM API read of CRM, messaging, calendar, workflow, and workspace-member data
- write risk: creates, updates, batch-writes, and deletes records across all 28 Twenty CRM REST objects
- approval: required for every update_<object>, batch_<object>, and delete_<object> action across the Twenty REST object surface; create_<object> actions require no approval (low-risk, non-destructive)
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect twenty
```

### Inspect as structured JSON

```bash
pm connectors inspect twenty --json
```

## Agent Rules

- Run pm connectors inspect twenty before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
