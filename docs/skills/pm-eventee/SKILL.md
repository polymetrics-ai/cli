---
name: pm-eventee
description: Eventee connector knowledge and safe action guide.
---

# pm-eventee

## Purpose

Reads Eventee event agenda, attendee, registration, group, review, and partner data; writes documented Eventee agenda, attendee, registration, partner, speaker, and track mutations through the public REST API.

## Icon

- id: pm-sample
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
- api_token (secret) (required)

## ETL Streams

- lectures:
  - primary key: id
  - fields: available(boolean), booked(integer), capacity(integer), code(string), created_at(string), description(string), end(string), event_day_id(integer), event_id(integer), hall_id(integer), id(integer), name(string), start(string), type(integer), updated_at(string)
- speakers:
  - primary key: id
  - fields: bio(string), company(string), country(string), email(string), event_id(integer), id(integer), language(string), name(string), order(integer), phone(string), position(string), web(string)
- days:
  - primary key: id
  - fields: content_url(string), date(string), event_id(integer), id(integer)
- halls:
  - primary key: id
  - fields: created_at(string), event_id(integer), id(integer), name(string), order(integer), updated_at(string)
- tracks:
  - primary key: id
  - fields: color(string), created_at(string), id(integer), name(string), order(integer), updated_at(string)
- workshops:
  - primary key: id
  - fields: available(boolean), booked(integer), capacity(integer), code(string), created_at(string), description(string), end(string), event_day_id(integer), event_id(integer), hall_id(integer), id(integer), name(string), start(string), type(integer), updated_at(string)
- pauses:
  - primary key: id
  - fields: created_at(string), description(string), end(string), id(integer), name(string), start(string), updated_at(string)
- partners:
  - primary key: id
  - fields: address(string), code(string), company(string), created_at(string), description(string), email(string), exhibitor(boolean), id(integer), phone(string), sponsor(boolean), updated_at(string), web(string)
- reviews:
  - primary key: id
  - fields: OS(string), comment(string), created_at(string), device(string), id(integer), lecture(object), lecture_id(integer), stars(integer), updated_at(string), user_id(integer), username(string), userphoto(string)
- groups:
  - primary key: id
  - fields: agenda(boolean), color(string), emoji(string), gamification(boolean), id(integer), is_default(boolean), name(string), networking(boolean), newsfeed(boolean), public_name(string), session_ratings(boolean), social_wall(boolean), ticket_names(array)
- participants:
  - primary key: id
  - fields: checked_at(string), company(string), email(string), first_name(string), group_id(integer), id(integer), last_name(string), name(string), position(string), registered_at(string), role(string)
- registrations:
  - primary key: id
  - fields: bio(string), company(string), email(string), email_valid(boolean), facebook_link(string), first_name(string), group_id(integer), id(integer), last_name(string), linked_in_link(string), phone(string), photo(string), position(string), send_email(boolean), status(integer), twitter_link(string), web(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- clear_test_content:
  - endpoint: DELETE /test/content
  - risk: deletes all tracks, pauses, speakers, workshops, lectures, and halls from the configured test event
- create_hall:
  - endpoint: POST /hall
  - required fields: name
  - risk: creates a hall in the configured event
- update_hall:
  - endpoint: PATCH /hall/{{ record.id }}
  - required fields: id, name
  - risk: updates a hall in the configured event
- delete_hall:
  - endpoint: DELETE /hall/{{ record.id }}
  - required fields: id
  - risk: deletes a hall from the configured event
- create_lecture:
  - endpoint: POST /lecture
  - required fields: name, start, end, hall_id, speakers, type, tracks
  - risk: creates a lecture or session in the configured event
- update_lecture:
  - endpoint: PATCH /lecture/{{ record.id }}
  - required fields: id, name, start, end, hall_id, speakers, type, tracks
  - risk: updates an existing lecture or session in the configured event
- delete_lecture:
  - endpoint: DELETE /lecture/{{ record.id }}
  - required fields: id
  - risk: deletes a lecture or session from the configured event
- invite_attendees:
  - endpoint: PUT /attendee/invite
  - required fields: users
  - risk: invites one or more attendees to the configured event
- update_attendee_checkin:
  - endpoint: PUT /attendee/{{ record.id }}/checkin
  - required fields: id, checkin
  - risk: sets the check-in state for an attendee
- remove_attendee:
  - endpoint: DELETE /attendee
  - required fields: email
  - risk: removes an invited attendee and may remove their access and event-linked information
- create_partner:
  - endpoint: POST /partner
  - required fields: company
  - risk: creates a partner, sponsor, or exhibitor profile in the configured event
- update_partner:
  - endpoint: PATCH /partner/{{ record.id }}
  - required fields: id, company
  - risk: updates an existing partner, sponsor, or exhibitor profile in the configured event
- delete_partner:
  - endpoint: DELETE /partner/{{ record.id }}
  - required fields: id
  - risk: deletes a partner, sponsor, or exhibitor profile from the configured event
- create_pause:
  - endpoint: POST /pause
  - required fields: name, start, end
  - risk: creates a pause or break in the configured event agenda
- update_pause:
  - endpoint: PATCH /pause/{{ record.id }}
  - required fields: id, name, start, end
  - risk: updates an existing pause or break in the configured event agenda
- delete_pause:
  - endpoint: DELETE /pause/{{ record.id }}
  - required fields: id
  - risk: deletes a pause or break from the configured event agenda
- invite_registrations:
  - endpoint: PUT /registration/invite
  - required fields: registrations
  - risk: invites one or more registrants to the configured event
- remove_registration:
  - endpoint: DELETE /registration
  - required fields: email
  - risk: removes an invited registrant from the configured event
- create_speaker:
  - endpoint: POST /speaker
  - required fields: name, phone
  - risk: creates a speaker profile in the configured event
- update_speaker:
  - endpoint: PATCH /speaker/{{ record.id }}
  - required fields: id, name, phone
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
