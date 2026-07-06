# pm connectors inspect aircall

```text
NAME
  pm connectors inspect aircall - Aircall connector manual

SYNOPSIS
  pm connectors inspect aircall
  pm connectors inspect aircall --json
  pm credentials add <name> --connector aircall [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Aircall calls, users, contacts, numbers, teams, tags, and webhooks, and writes user/team/contact/tag/webhook mutations plus call archive/comment/tag actions, through the Aircall REST API.

ICON
  asset: icons/aircall.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.aircall.io/api-references/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  start_date
  api_id (secret)
  api_token (secret)

ETL STREAMS
  calls:
    primary key: id
    cursor: started_at
    fields: answered_at(), archived(), direction(), duration(), ended_at(), id(), missed_call_reason(), raw_digits(), recording(), sid(), started_at(), status(), voicemail()
  users:
    primary key: id
    cursor: created_at
    fields: availability_status(), available(), created_at(), email(), id(), language(), name(), time_zone(), wrap_up_time()
  contacts:
    primary key: id
    cursor: created_at
    fields: company_name(), created_at(), first_name(), id(), information(), is_shared(), last_name(), updated_at()
  numbers:
    primary key: id
    fields: country(), created_at(), digits(), id(), is_ivr(), live_recording_activated(), name(), open(), time_zone()
  teams:
    primary key: id
    fields: created_at(), id(), name()
  tags:
    primary key: id
    fields: color(), description(), id(), name()
  webhooks:
    primary key: id
    fields: active(), events(), id(), url()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_user:
    endpoint: POST /users
    risk: creates a new Aircall agent seat, which may consume a billable license; external mutation, approval required
  update_user:
    endpoint: PUT /users/{{ record.id }}
    required fields: id
    risk: mutates an existing agent's profile/availability; a visible change for that agent's call routing
  delete_user:
    endpoint: DELETE /users/{{ record.id }}
    required fields: id
    risk: permanently removes an Aircall agent seat; irreversible, frees the associated license; approval required
  create_team:
    endpoint: POST /teams
    risk: creates a new team container; low-risk external mutation, no approval required
  delete_team:
    endpoint: DELETE /teams/{{ record.id }}
    required fields: id
    risk: permanently removes a team; agents assigned to it lose that team's routing/membership immediately
  add_user_to_team:
    endpoint: POST /teams/{{ record.team_id }}/users/{{ record.user_id }}
    required fields: team_id, user_id
    risk: adds an agent to a team, changing that agent's call routing/membership; low-risk, reversible via remove_user_from_team
  remove_user_from_team:
    endpoint: DELETE /teams/{{ record.team_id }}/users/{{ record.user_id }}
    required fields: team_id, user_id
    risk: removes an agent from a team, changing that agent's call routing/membership immediately
  create_contact:
    endpoint: POST /contacts
    risk: creates a new shared/personal directory contact; low-risk external mutation, no approval required
  update_contact:
    endpoint: PUT /contacts/{{ record.id }}
    required fields: id
    risk: mutates an existing contact's directory record, including its full phone_numbers/emails arrays (a partial array here replaces the previous set)
  delete_contact:
    endpoint: DELETE /contacts/{{ record.id }}
    required fields: id
    risk: permanently removes a directory contact; irreversible
  create_tag:
    endpoint: POST /tags
    risk: creates a new call-tagging label; low-risk external mutation, no approval required
  update_tag:
    endpoint: PUT /tags/{{ record.id }}
    required fields: id
    risk: renames or recolors an existing tag; a visible change everywhere the tag is already applied
  delete_tag:
    endpoint: DELETE /tags/{{ record.id }}
    required fields: id
    risk: permanently removes a tag; it is un-applied from every call that previously carried it
  create_webhook:
    endpoint: POST /webhooks
    risk: registers a new outbound webhook that will POST live call/event data to an external URL of the caller's choosing; verify the target endpoint before enabling
  update_webhook:
    endpoint: PUT /webhooks/{{ record.id }}
    required fields: id
    risk: mutates an existing webhook's target URL, subscribed events, or active state; a changed url redirects future event deliveries to a different endpoint
  delete_webhook:
    endpoint: DELETE /webhooks/{{ record.id }}
    required fields: id
    risk: permanently removes a webhook subscription; event delivery to its target URL stops immediately
  archive_call:
    endpoint: PUT /calls/{{ record.id }}/archive
    required fields: id
    risk: marks a call as archived, hiding it from default call-list views; reversible via unarchive_call
  unarchive_call:
    endpoint: PUT /calls/{{ record.id }}/unarchive
    required fields: id
    risk: restores a previously archived call to default call-list views
  comment_call:
    endpoint: POST /calls/{{ record.id }}/comments
    required fields: id
    optional fields: content
    risk: adds an internal comment note to a call record; visible to other agents with call access, no external side effect
  tag_call:
    endpoint: POST /calls/{{ record.id }}/tags
    required fields: id
    optional fields: tag_ids
    risk: applies the given tags to a call; additive, does not remove tags already present

SECURITY
  read risk: external Aircall API read of call, contact, and directory data
  write risk: external Aircall API mutation of agents, teams, contacts, tags, webhooks, and call archive/comment/tag state; approval required for user/team/contact/webhook create-update-delete, low-risk for additive call tagging/commenting
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect aircall

  # Inspect as structured JSON
  pm connectors inspect aircall --json

AGENT WORKFLOW
  - Run pm connectors inspect aircall before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
