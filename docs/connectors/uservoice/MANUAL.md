# pm connectors inspect uservoice

```text
NAME
  pm connectors inspect uservoice - UserVoice connector manual

SYNOPSIS
  pm connectors inspect uservoice
  pm connectors inspect uservoice --json
  pm credentials add <name> --connector uservoice [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads suggestions, forums, users, categories, statuses, labels, comments, notes, and teams from the UserVoice Admin API, and writes suggestion/comment/label/note lifecycle mutations.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  start_date
  api_key (secret)

ETL STREAMS
  suggestions:
    primary key: id
    cursor: created_at
    fields: created_at(), id(), state(), title()
  forums:
    primary key: id
    cursor: updated_at
    fields: categories_count(), created_at(), id(), is_default(), is_open(), is_private(), is_public(), moderation_enabled(), name(), open_suggestions_count(), suggestions_count(), updated_at()
  users:
    primary key: id
    cursor: updated_at
    fields: avatar_url(), created_at(), email_address(), guid(), id(), is_admin(), is_owner(), name(), state(), updated_at()
  categories:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), name(), open_suggestions_count(), position(), suggestions_count(), updated_at()
  statuses:
    primary key: id
    cursor: updated_at
    fields: allow_comments(), created_at(), hex_color(), id(), is_default(), is_open(), name(), position(), updated_at()
  labels:
    primary key: id
    cursor: updated_at
    fields: can_recommend(), created_at(), full_name(), id(), level(), name(), open_suggestions_count(), updated_at()
  comments:
    primary key: id
    cursor: updated_at
    fields: body(), body_mime_type(), created_at(), id(), inappropriate_flags_count(), is_admin_comment(), state(), updated_at()
  notes:
    primary key: id
    cursor: updated_at
    fields: body(), body_mime_type(), created_at(), id(), reply_count(), updated_at()
  teams:
    primary key: id
    fields: id(), members_count(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_suggestion:
    endpoint: POST /api/v2/admin/suggestions
    risk: creates a new customer suggestion (idea); low-risk external mutation, no approval required
  update_suggestion:
    endpoint: PUT /api/v2/admin/suggestions/{{ record.id }}
    required fields: id
    risk: updates an existing suggestion's title/body; external mutation, no approval required
  approve_suggestion:
    endpoint: PUT /api/v2/admin/suggestions/{{ record.id }}/approve
    required fields: id
    risk: approves (publishes) a pending suggestion, making it publicly visible; no approval required
  delete_suggestion:
    endpoint: PUT /api/v2/admin/suggestions/{{ record.id }}/delete
    required fields: id
    risk: soft-deletes (moderates) a suggestion; UserVoice's own API keeps a matching restore endpoint (not modeled here) so this is a reversible moderation action, not permanent data loss, but is still marked destructive-shaped for operator awareness
  create_comment:
    endpoint: POST /api/v2/admin/comments
    risk: posts a new comment on an existing suggestion; low-risk external mutation, no approval required
  create_label:
    endpoint: POST /api/v2/admin/labels
    risk: creates a new label for tagging suggestions; low-risk external mutation, no approval required
  update_label:
    endpoint: PUT /api/v2/admin/labels/{{ record.id }}
    required fields: id
    risk: updates an existing label's name/settings; external mutation, no approval required
  create_note:
    endpoint: POST /api/v2/admin/notes
    risk: creates an internal (non-public) note on a suggestion; low-risk external mutation, no approval required

SECURITY
  read risk: external UserVoice API read of customer suggestion, forum, user, category, status, label, comment, note, and team data
  write risk: external mutation of UserVoice suggestions (create/update/approve/delete), comments, labels, and internal notes; suggestion delete is a soft moderation action, not permanent data loss
  approval: none required; delete_suggestion is UserVoice's own soft-delete/moderation action (reversible via restore_suggestion, not modeled), not an irreversible destructive delete
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect uservoice

  # Inspect as structured JSON
  pm connectors inspect uservoice --json

AGENT WORKFLOW
  - Run pm connectors inspect uservoice before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
