---
name: pm-jira
description: Jira connector knowledge and safe action guide.
---

# pm-jira

## Purpose

Reads Jira issues, projects, and users through the Jira Cloud REST API v3 using HTTP Basic auth (email + API token). Read-only.

## Icon

- asset: icons/jira.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.atlassian.com/changelog/#

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- email
- api_token (secret)

## ETL Streams

- issues:
  - primary key: id
  - cursor: updated
  - fields: assignee(), created(), id(), issuetype(), key(), priority(), project(), reporter(), self(), status(), summary(), updated()
- projects:
  - primary key: id
  - fields: id(), isPrivate(), key(), name(), projectTypeKey(), self(), simplified(), style()
- users:
  - primary key: accountId
  - fields: accountId(), accountType(), active(), displayName(), emailAddress(), self()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Jira Cloud API read of issue, project, and user data
- approval: none; read-only, no reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect jira
```

### Inspect as structured JSON

```bash
pm connectors inspect jira --json
```

## Agent Rules

- Run pm connectors inspect jira before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
