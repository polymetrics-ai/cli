---
name: pm-eventee
description: Eventee connector knowledge and safe action guide.
---

# pm-eventee

## Purpose

Reads Eventee event agenda, attendee, registration, group, review, and partner data; writes documented Eventee agenda, attendee, registration, partner, speaker, and track mutations through the public REST API.

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
- api_token (secret)

## ETL Streams

- lectures:
  - primary key: id
  - fields: available(), booked(), capacity(), code(), created_at(), description(), end(), event_day_id(), event_id(), hall_id(), id(), name(), start(), type(), updated_at()
- speakers:
  - primary key: id
  - fields: bio(), company(), country(), email(), event_id(), id(), language(), name(), order(), phone(), position(), web()
- days:
  - primary key: id
  - fields: content_url(), date(), event_id(), id()
- halls:
  - primary key: id
  - fields: created_at(), event_id(), id(), name(), order(), updated_at()
- tracks:
  - primary key: id
  - fields: color(), created_at(), id(), name(), order(), updated_at()
- workshops:
  - primary key: id
  - fields: available(), booked(), capacity(), code(), created_at(), description(), end(), event_day_id(), event_id(), hall_id(), id(), name(), start(), type(), updated_at()
- pauses:
  - primary key: id
  - fields: created_at(), description(), end(), id(), name(), start(), updated_at()
- partners:
  - primary key: id
  - fields: address(), code(), company(), created_at(), description(), email(), exhibitor(), id(), phone(), sponsor(), updated_at(), web()
- reviews:
  - primary key: id
  - fields: OS(), comment(), created_at(), device(), id(), lecture(), lecture_id(), stars(), updated_at(), user_id(), username(), userphoto()
- groups:
  - primary key: id
  - fields: agenda(), color(), emoji(), gamification(), id(), is_default(), name(), networking(), newsfeed(), public_name(), session_ratings(), social_wall(), ticket_names()
- participants:
  - primary key: id
  - fields: checked_at(), company(), email(), first_name(), group_id(), id(), last_name(), name(), position(), registered_at(), role()
- registrations:
  - primary key: id
  - fields: bio(), company(), email(), email_valid(), facebook_link(), first_name(), group_id(), id(), last_name(), linked_in_link(), phone(), photo(), position(), send_email(), status(), twitter_link(), web()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- clear_test_content:
  - endpoint: DELETE /test/content
  - risk: deletes all tracks, pauses, speakers, workshops, lectures, and halls from the configured test event
- create_hall:
  - endpoint: POST /hall
  - risk: creates a hall in the configured event
- update_hall:
  - endpoint: PATCH /hall/{{ record.id }}
  - required fields: id
  - risk: updates a hall in the configured event
- delete_hall:
  - endpoint: DELETE /hall/{{ record.id }}
  - required fields: id
  - risk: deletes a hall from the configured event
- create_lecture:
  - endpoint: POST /lecture
  - risk: creates a lecture or session in the configured event
- update_lecture:
  - endpoint: PATCH /lecture/{{ record.id }}
  - required fields: id
  - risk: updates an existing lecture or session in the configured event
- delete_lecture:
  - endpoint: DELETE /lecture/{{ record.id }}
  - required fields: id
  - risk: deletes a lecture or session from the configured event
- invite_attendees:
  - endpoint: PUT /attendee/invite
  - risk: invites one or more attendees to the configured event
- update_attendee_checkin:
  - endpoint: PUT /attendee/{{ record.id }}/checkin
  - required fields: id
  - risk: sets the check-in state for an attendee
- remove_attendee:
  - endpoint: DELETE /attendee
  - risk: removes an invited attendee and may remove their access and event-linked information
- create_partner:
  - endpoint: POST /partner
  - risk: creates a partner, sponsor, or exhibitor profile in the configured event
- update_partner:
  - endpoint: PATCH /partner/{{ record.id }}
  - required fields: id
  - risk: updates an existing partner, sponsor, or exhibitor profile in the configured event
- delete_partner:
  - endpoint: DELETE /partner/{{ record.id }}
  - required fields: id
  - risk: deletes a partner, sponsor, or exhibitor profile from the configured event
- create_pause:
  - endpoint: POST /pause
  - risk: creates a pause or break in the configured event agenda
- update_pause:
  - endpoint: PATCH /pause/{{ record.id }}
  - required fields: id
  - risk: updates an existing pause or break in the configured event agenda
- delete_pause:
  - endpoint: DELETE /pause/{{ record.id }}
  - required fields: id
  - risk: deletes a pause or break from the configured event agenda
- invite_registrations:
  - endpoint: PUT /registration/invite
  - risk: invites one or more registrants to the configured event
- remove_registration:
  - endpoint: DELETE /registration
  - risk: removes an invited registrant from the configured event
- create_speaker:
  - endpoint: POST /speaker
  - risk: creates a speaker profile in the configured event
- update_speaker:
  - endpoint: PATCH /speaker/{{ record.id }}
  - required fields: id
  - risk: updates an existing speaker profile in the configured event
- delete_speaker:
  - endpoint: DELETE /speaker/{{ record.id }}
  - required fields: id
  - risk: deletes a speaker profile from the configured event
- create_track:
  - endpoint: POST /label
  - risk: creates a track label in the configured event
- update_track:
  - endpoint: PATCH /label/{{ record.id }}
  - required fields: id
  - risk: updates an existing track label in the configured event
- delete_track:
  - endpoint: DELETE /label/{{ record.id }}
  - required fields: id
  - risk: deletes a track label from the configured event

## Security

- read risk: external Eventee API reads of event agenda, attendee, registration, group, review, and partner data
- write risk: creates, updates, invites, checks in, removes, or deletes Eventee event content and attendees/registrants; destructive deletes require approval
- approval: reverse ETL writes require plan preview and approval token; delete actions are marked destructive
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect eventee
```

### Inspect as structured JSON

```bash
pm connectors inspect eventee --json
```

## Agent Rules

- Run pm connectors inspect eventee before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
