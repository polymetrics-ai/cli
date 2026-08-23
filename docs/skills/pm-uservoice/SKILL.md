---
name: pm-uservoice
description: UserVoice connector knowledge and safe action guide.
---

# pm-uservoice

## Purpose

Reads suggestions, forums, users, categories, statuses, labels, comments, notes, and teams from the UserVoice Admin API, and writes suggestion/comment/label/note lifecycle mutations.

## Icon

- id: simple-icons-uservoice
- asset: icons/simple-icons/uservoice.svg
- title: UserVoice
- simple_icon_slug: uservoice
- simple_icon_hex: FF6720
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=UserVoice
- match: exact-name-or-slug
- matched_by: uservoice

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- start_date
- api_key (secret) (required)

## ETL Streams

- suggestions:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), id(string), state(string), title(string)
- forums:
  - primary key: id
  - cursor: updated_at
  - fields: categories_count(integer), created_at(string), id(integer), is_default(boolean), is_open(boolean), is_private(boolean), is_public(boolean), moderation_enabled(boolean), name(string), open_suggestions_count(integer), suggestions_count(integer), updated_at(string)
- users:
  - primary key: id
  - cursor: updated_at
  - fields: avatar_url(string), created_at(string), email_address(string), guid(string), id(integer), is_admin(boolean), is_owner(boolean), name(string), state(string), updated_at(string)
- categories:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(integer), name(string), open_suggestions_count(integer), position(integer), suggestions_count(integer), updated_at(string)
- statuses:
  - primary key: id
  - cursor: updated_at
  - fields: allow_comments(boolean), created_at(string), hex_color(string), id(integer), is_default(boolean), is_open(boolean), name(string), position(integer), updated_at(string)
- labels:
  - primary key: id
  - cursor: updated_at
  - fields: can_recommend(boolean), created_at(string), full_name(string), id(integer), level(integer), name(string), open_suggestions_count(integer), updated_at(string)
- comments:
  - primary key: id
  - cursor: updated_at
  - fields: body(string), body_mime_type(string), created_at(string), id(integer), inappropriate_flags_count(integer), is_admin_comment(boolean), state(string), updated_at(string)
- notes:
  - primary key: id
  - cursor: updated_at
  - fields: body(string), body_mime_type(string), created_at(string), id(integer), reply_count(integer), updated_at(string)
- teams:
  - primary key: id
  - fields: id(integer), members_count(integer), name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_suggestion:
  - endpoint: POST /api/v2/admin/suggestions
  - required fields: title, links
  - risk: creates a new customer suggestion (idea); low-risk external mutation, no approval required
- update_suggestion:
  - endpoint: PUT /api/v2/admin/suggestions/{{ record.id }}
  - required fields: id
  - risk: updates an existing suggestion's title/body; external mutation, no approval required
- approve_suggestion:
  - endpoint: PUT /api/v2/admin/suggestions/{{ record.id }}/approve
  - required fields: id
  - risk: approves (publishes) a pending suggestion, making it publicly visible; no approval required
- delete_suggestion:
  - endpoint: PUT /api/v2/admin/suggestions/{{ record.id }}/delete
  - required fields: id
  - risk: soft-deletes (moderates) a suggestion; UserVoice's own API keeps a matching restore endpoint (not modeled here) so this is a reversible moderation action, not permanent data loss, but is still marked destructive-shaped for operator awareness
- create_comment:
  - endpoint: POST /api/v2/admin/comments
  - required fields: body, links
  - risk: posts a new comment on an existing suggestion; low-risk external mutation, no approval required
- create_label:
  - endpoint: POST /api/v2/admin/labels
  - required fields: name
  - risk: creates a new label for tagging suggestions; low-risk external mutation, no approval required
- update_label:
  - endpoint: PUT /api/v2/admin/labels/{{ record.id }}
  - required fields: id
  - risk: updates an existing label's name/settings; external mutation, no approval required
- create_note:
  - endpoint: POST /api/v2/admin/notes
  - required fields: body, links
  - risk: creates an internal (non-public) note on a suggestion; low-risk external mutation, no approval required

## Security

- read risk: external UserVoice API read of customer suggestion, forum, user, category, status, label, comment, note, and team data
- write risk: external mutation of UserVoice suggestions (create/update/approve/delete), comments, labels, and internal notes; suggestion delete is a soft moderation action, not permanent data loss
- approval: none required; delete_suggestion is UserVoice's own soft-delete/moderation action (reversible via restore_suggestion, not modeled), not an irreversible destructive delete
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect uservoice
```

### Inspect as structured JSON

```bash
pm connectors inspect uservoice --json
```

## Agent Rules

- Run pm connectors inspect uservoice before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
