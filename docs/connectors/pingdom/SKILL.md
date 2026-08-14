---
name: pm-pingdom
description: Pingdom connector knowledge and safe action guide.
---

# pm-pingdom

## Purpose

Reads Pingdom checks, probes, actions, maintenance windows/occurrences, alerting contacts/teams, credits, transaction checks, and reference data, and writes check/contact/team/maintenance mutations through API 3.1.

## Icon

- id: simple-icons-pingdom
- asset: icons/simple-icons/pingdom.svg
- title: Pingdom
- simple_icon_slug: pingdom
- simple_icon_hex: FFF000
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Pingdom
- match: exact-name-or-slug
- matched_by: pingdom

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret) (required)

## ETL Streams

- checks:
  - primary key: id
  - fields: hostname(string), id(integer), lasterrortime(integer), lastresponsetime(integer), lasttesttime(integer), name(string), resolution(integer), status(string), tags(array), type(string)
- probes:
  - primary key: id
  - fields: active(boolean), city(string), country(string), hostname(string), id(integer), ip(string), name(string)
- actions:
  - primary key: id
  - fields: checkid(integer), checkname(string), contactname(string), id(integer), status(string), time(integer), via(integer)
- maintenance:
  - primary key: id
  - fields: description(string), effectiveto(integer), from(integer), id(integer), recurrencetype(string), repeatevery(integer), to(integer)
- reference:
  - fields: checktypes(object), probes(object), regions(object)
- alerting_contacts:
  - primary key: id
  - fields: id(integer), name(string), notification_targets(object), owner(boolean), paused(boolean), teams(array), type(string)
- alerting_teams:
  - primary key: id
  - fields: id(integer), members(array), name(string)
- maintenance_occurrences:
  - primary key: id
  - fields: from(integer), id(integer), maintenanceid(integer), to(integer)
- credits:
  - fields: autofillsms(boolean), autofillsms_amount(integer), autofillsms_when_left(integer), availablechecks(integer), availablerumsites(integer), availablesms(integer), availablesmstests(integer), checklimit(integer), max_sms_overage(integer), useddefault(integer), usedtransaction(integer)
- tms_checks:
  - primary key: id
  - fields: active(boolean), created_at(integer), id(integer), interval(integer), last_downtime_end(integer), last_downtime_start(integer), modified_at(integer), name(string), region(string), status(string), tags(array), type(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_check:
  - endpoint: POST /checks
  - required fields: name, host, type
  - risk: creates a new Pingdom uptime check (this action models the common HTTP-type check shape; Pingdom's other 8 check types share the same name/host/type/paused/resolution/notification fields plus type-specific attributes not modeled here, see docs.md Known limits); low-risk external mutation, no approval required
- update_check:
  - endpoint: PUT /checks/{{ record.id }}
  - required fields: id
  - risk: updates an existing check's settings (name/host/paused/resolution/tags); external mutation, approval required
- delete_check:
  - endpoint: DELETE /checks/{{ record.id }}
  - required fields: id
  - risk: permanently deletes an uptime check and its historical results; destructive external mutation, approval required
- create_contact:
  - endpoint: POST /alerting/contacts
  - required fields: name, notification_targets
  - risk: creates a new alerting contact with email/SMS notification targets; low-risk external mutation, no approval required
- update_contact:
  - endpoint: PUT /alerting/contacts/{{ record.id }}
  - required fields: id, name, paused, notification_targets
  - risk: updates an existing alerting contact's name/paused state/notification targets (Pingdom's PUT is a full replacement, requiring name/paused/notification_targets together); external mutation, approval required
- delete_contact:
  - endpoint: DELETE /alerting/contacts/{{ record.id }}
  - required fields: id
  - risk: permanently deletes an alerting contact and its notification targets; destructive external mutation, approval required
- create_team:
  - endpoint: POST /alerting/teams
  - required fields: name, member_ids
  - risk: creates a new alerting team from a list of contact ids; low-risk external mutation, no approval required
- update_team:
  - endpoint: PUT /alerting/teams/{{ record.id }}
  - required fields: id, name, member_ids
  - risk: updates an existing alerting team's name/member list; external mutation, approval required
- delete_team:
  - endpoint: DELETE /alerting/teams/{{ record.id }}
  - required fields: id
  - risk: permanently deletes an alerting team; destructive external mutation, approval required
- create_maintenance:
  - endpoint: POST /maintenance
  - required fields: description, from, to
  - risk: creates a new maintenance window that suppresses alerting for the assigned checks during the scheduled period; low-risk external mutation, no approval required
- delete_maintenance:
  - endpoint: DELETE /maintenance/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a maintenance window, immediately resuming alerting for its assigned checks; destructive external mutation, approval required

## Security

- read risk: external Pingdom API read of uptime/transaction monitoring configuration, alerting configuration, account credits, and event data
- write risk: creates/updates/deletes uptime checks, alerting contacts and teams, and maintenance windows
- approval: required for update_check/update_contact/update_team/delete_check/delete_contact/delete_team/delete_maintenance; create_check/create_contact/create_team/create_maintenance require no approval (low-risk, non-destructive)
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect pingdom
```

### Inspect as structured JSON

```bash
pm connectors inspect pingdom --json
```

## Agent Rules

- Run pm connectors inspect pingdom before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
