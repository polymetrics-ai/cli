---
name: pm-pingdom
description: Pingdom connector knowledge and safe action guide.
---

# pm-pingdom

## Purpose

Reads Pingdom checks, probes, actions, maintenance windows/occurrences, alerting contacts/teams, credits, transaction checks, and reference data, and writes check/contact/team/maintenance mutations through API 3.1.

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
- api_key (secret)

## ETL Streams

- checks:
  - primary key: id
  - fields: hostname(), id(), lasterrortime(), lastresponsetime(), lasttesttime(), name(), resolution(), status(), tags(), type()
- probes:
  - primary key: id
  - fields: active(), city(), country(), hostname(), id(), ip(), name()
- actions:
  - primary key: id
  - fields: checkid(), checkname(), contactname(), id(), status(), time(), via()
- maintenance:
  - primary key: id
  - fields: description(), effectiveto(), from(), id(), recurrencetype(), repeatevery(), to()
- reference:
  - fields: checktypes(), probes(), regions()
- alerting_contacts:
  - primary key: id
  - fields: id(), name(), notification_targets(), owner(), paused(), teams(), type()
- alerting_teams:
  - primary key: id
  - fields: id(), members(), name()
- maintenance_occurrences:
  - primary key: id
  - fields: from(), id(), maintenanceid(), to()
- credits:
  - fields: autofillsms(), autofillsms_amount(), autofillsms_when_left(), availablechecks(), availablerumsites(), availablesms(), availablesmstests(), checklimit(), max_sms_overage(), useddefault(), usedtransaction()
- tms_checks:
  - primary key: id
  - fields: active(), created_at(), id(), interval(), last_downtime_end(), last_downtime_start(), modified_at(), name(), region(), status(), tags(), type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_check:
  - endpoint: POST /checks
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
  - risk: creates a new alerting contact with email/SMS notification targets; low-risk external mutation, no approval required
- update_contact:
  - endpoint: PUT /alerting/contacts/{{ record.id }}
  - required fields: id
  - risk: updates an existing alerting contact's name/paused state/notification targets (Pingdom's PUT is a full replacement, requiring name/paused/notification_targets together); external mutation, approval required
- delete_contact:
  - endpoint: DELETE /alerting/contacts/{{ record.id }}
  - required fields: id
  - risk: permanently deletes an alerting contact and its notification targets; destructive external mutation, approval required
- create_team:
  - endpoint: POST /alerting/teams
  - risk: creates a new alerting team from a list of contact ids; low-risk external mutation, no approval required
- update_team:
  - endpoint: PUT /alerting/teams/{{ record.id }}
  - required fields: id
  - risk: updates an existing alerting team's name/member list; external mutation, approval required
- delete_team:
  - endpoint: DELETE /alerting/teams/{{ record.id }}
  - required fields: id
  - risk: permanently deletes an alerting team; destructive external mutation, approval required
- create_maintenance:
  - endpoint: POST /maintenance
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
