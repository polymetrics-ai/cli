---
name: pm-aircall
description: Aircall connector knowledge and safe action guide.
---

# pm-aircall

## Purpose

Reads Aircall calls, users, contacts, numbers, teams, tags, and webhooks, and writes user/team/contact/tag/webhook mutations plus call archive/comment/tag actions, through the Aircall REST API.

## Icon

- id: aircall
- asset: icons/aircall.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.aircall.io/api-references/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- start_date
- api_id (secret) (required)
- api_token (secret) (required)

## ETL Streams

- calls:
  - primary key: id
  - cursor: started_at
  - fields: answered_at(integer), archived(boolean), direction(string), duration(integer), ended_at(integer), id(integer), missed_call_reason(string), raw_digits(string), recording(string), sid(string), started_at(integer), status(string), voicemail(string)
- users:
  - primary key: id
  - cursor: created_at
  - fields: availability_status(string), available(boolean), created_at(string), email(string), id(integer), language(string), name(string), time_zone(string), wrap_up_time(integer)
- contacts:
  - primary key: id
  - cursor: created_at
  - fields: company_name(string), created_at(string), first_name(string), id(integer), information(string), is_shared(boolean), last_name(string), updated_at(string)
- numbers:
  - primary key: id
  - fields: country(string), created_at(string), digits(string), id(integer), is_ivr(boolean), live_recording_activated(boolean), name(string), open(boolean), time_zone(string)
- teams:
  - primary key: id
  - fields: created_at(string), id(integer), name(string)
- tags:
  - primary key: id
  - fields: color(string), description(string), id(integer), name(string)
- webhooks:
  - primary key: id
  - fields: active(boolean), events(array), id(integer), url(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_user:
  - endpoint: POST /users
  - required fields: name, email
  - risk: creates a new Aircall agent seat, which may consume a billable license; external mutation, approval required
- update_user:
  - endpoint: PUT /users/{{ record.id }}
  - required fields: id
  - risk: mutates an existing agent's profile/availability; a visible change for that agent's call routing
- delete_user:
  - endpoint: DELETE /users/{{ record.id }}
  - required fields: id
  - risk: permanently removes an Aircall agent seat; irreversible, frees the associated license; approval required
- create_team:
  - endpoint: POST /teams
  - required fields: name
  - risk: creates a new team container; low-risk external mutation, no approval required
- delete_team:
  - endpoint: DELETE /teams/{{ record.id }}
  - required fields: id
  - risk: permanently removes a team; agents assigned to it lose that team's routing/membership immediately
- add_user_to_team:
  - endpoint: POST /teams/{{ record.team_id }}/users/{{ record.user_id }}
  - required fields: team_id, user_id
  - risk: adds an agent to a team, changing that agent's call routing/membership; low-risk, reversible via remove_user_from_team
- remove_user_from_team:
  - endpoint: DELETE /teams/{{ record.team_id }}/users/{{ record.user_id }}
  - required fields: team_id, user_id
  - risk: removes an agent from a team, changing that agent's call routing/membership immediately
- create_contact:
  - endpoint: POST /contacts
  - risk: creates a new shared/personal directory contact; low-risk external mutation, no approval required
- update_contact:
  - endpoint: PUT /contacts/{{ record.id }}
  - required fields: id
  - risk: mutates an existing contact's directory record, including its full phone_numbers/emails arrays (a partial array here replaces the previous set)
- delete_contact:
  - endpoint: DELETE /contacts/{{ record.id }}
  - required fields: id
  - risk: permanently removes a directory contact; irreversible
- create_tag:
  - endpoint: POST /tags
  - required fields: name, color
  - risk: creates a new call-tagging label; low-risk external mutation, no approval required
- update_tag:
  - endpoint: PUT /tags/{{ record.id }}
  - required fields: id
  - risk: renames or recolors an existing tag; a visible change everywhere the tag is already applied
- delete_tag:
  - endpoint: DELETE /tags/{{ record.id }}
  - required fields: id
  - risk: permanently removes a tag; it is un-applied from every call that previously carried it
- create_webhook:
  - endpoint: POST /webhooks
  - required fields: url, events
  - risk: registers a new outbound webhook that will POST live call/event data to an external URL of the caller's choosing; verify the target endpoint before enabling
- update_webhook:
  - endpoint: PUT /webhooks/{{ record.id }}
  - required fields: id
  - risk: mutates an existing webhook's target URL, subscribed events, or active state; a changed url redirects future event deliveries to a different endpoint
- delete_webhook:
  - endpoint: DELETE /webhooks/{{ record.id }}
  - required fields: id
  - risk: permanently removes a webhook subscription; event delivery to its target URL stops immediately
- archive_call:
  - endpoint: PUT /calls/{{ record.id }}/archive
  - required fields: id
  - risk: marks a call as archived, hiding it from default call-list views; reversible via unarchive_call
- unarchive_call:
  - endpoint: PUT /calls/{{ record.id }}/unarchive
  - required fields: id
  - risk: restores a previously archived call to default call-list views
- comment_call:
  - endpoint: POST /calls/{{ record.id }}/comments
  - required fields: id, content
  - risk: adds an internal comment note to a call record; visible to other agents with call access, no external side effect
- tag_call:
  - endpoint: POST /calls/{{ record.id }}/tags
  - required fields: id, tag_ids
  - risk: applies the given tags to a call; additive, does not remove tags already present

## Security

- read risk: external Aircall API read of call, contact, and directory data
- write risk: external Aircall API mutation of agents, teams, contacts, tags, webhooks, and call archive/comment/tag state; approval required for user/team/contact/webhook create-update-delete, low-risk for additive call tagging/commenting
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect aircall
```

### Inspect as structured JSON

```bash
pm connectors inspect aircall --json
```

## Agent Rules

- Run pm connectors inspect aircall before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
