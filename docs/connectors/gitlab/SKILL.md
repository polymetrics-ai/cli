---
name: pm-gitlab
description: GitLab connector knowledge and safe action guide.
---

# pm-gitlab

## Purpose

Reads GitLab projects, groups, users, and issues through the GitLab REST API v4.

## Icon

- asset: icons/gitlab.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.gitlab.com/ee/api/rest/deprecations.html

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- page_size
- start_date
- access_token (secret)

## ETL Streams

- projects:
  - primary key: id
  - cursor: last_activity_at
  - fields: archived(), created_at(), default_branch(), description(), forks_count(), id(), last_activity_at(), name(), open_issues_count(), path(), path_with_namespace(), star_count(), visibility(), web_url()
- groups:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), description(), full_name(), full_path(), id(), name(), parent_id(), path(), visibility(), web_url()
- users:
  - primary key: id
  - cursor: created_at
  - fields: bot(), created_at(), id(), is_admin(), name(), state(), username(), web_url()
- issues:
  - primary key: id
  - cursor: updated_at
  - fields: author_id(), closed_at(), created_at(), downvotes(), id(), iid(), project_id(), state(), title(), updated_at(), upvotes(), user_notes_count(), web_url()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external GitLab API read of projects, groups, users, and issues
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect gitlab
```

### Inspect as structured JSON

```bash
pm connectors inspect gitlab --json
```

## Agent Rules

- Run pm connectors inspect gitlab before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
