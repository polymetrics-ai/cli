---
name: pm-clockodo
description: Clockodo connector knowledge and safe action guide.
---

# pm-clockodo

## Purpose

Reads Clockodo customers, projects, services, users, time entries, absences, teams, surcharges, lump-sum services, nonbusiness groups/days, holiday/overtime carryovers, target hours, and current-user settings, and writes customers/projects/services/teams/lump-sum services through the Clockodo REST API.

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

- absences_year
- base_url
- email_address
- entries_time_since
- entries_time_until
- external_application
- language
- nonbusinessdays_year
- api_key (secret)

## ETL Streams

- customers:
  - primary key: id
  - fields: active(), billable_default(), color(), id(), name(), note(), number()
- projects:
  - primary key: id
  - fields: active(), billable_default(), budget_is_hours(), budget_money(), completed(), customers_id(), deadline(), id(), name(), note(), number()
- services:
  - primary key: id
  - fields: active(), id(), name(), note(), number()
- users:
  - primary key: id
  - fields: active(), email(), id(), language(), name(), number(), role(), teams_id(), timezone()
- current_user_settings:
  - fields: company(), user(), workTimeRegulation()
- teams:
  - primary key: id
  - fields: id(), leader(), name()
- surcharges:
  - primary key: id
  - fields: accumulation(), id(), name(), night(), night_increased(), nonbusiness(), nonbusiness_special(), saturday(), sunday()
- lumpsum_services:
  - primary key: id
  - fields: active(), id(), name(), note(), number(), price(), unit()
- nonbusiness_groups:
  - primary key: id
  - fields: id(), name()
- nonbusiness_days:
  - primary key: id, date
  - fields: date(), half_day(), id(), name(), nonbusinessgroups_id()
- holidays_carry:
  - primary key: users_id, year
  - fields: count(), note(), users_id(), year()
- holidays_quota:
  - primary key: id
  - fields: count(), id(), users_id(), year_since(), year_until()
- overtime_carry:
  - primary key: users_id, year
  - fields: hours(), note(), users_id(), year()
- target_hours:
  - primary key: id
  - fields: absence_fixed_credit(), compensation_daily(), compensation_monthly(), date_since(), date_until(), friday(), id(), monday(), monthly_target(), saturday(), sunday(), thursday(), tuesday(), type(), users_id(), wednesday()
- absences:
  - primary key: id
  - fields: approved_by(), count_days(), count_hours(), date_approved(), date_enquired(), date_since(), date_until(), id(), note(), sick_note(), status(), type(), users_id()
- entries:
  - primary key: id
  - fields: billable(), clocked(), clocked_offline(), customers_id(), duration(), hourly_rate(), id(), lumpsum(), lumpsum_services_amount(), lumpsum_services_id(), lumpsum_services_price(), offset(), projects_id(), services_id(), texts_id(), time_clocked_since(), time_insert(), time_last_change(), time_last_change_worktime(), time_since(), time_until(), type(), users_id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_customer:
  - endpoint: POST /v2/customers
  - risk: external mutation; creates a live Clockodo customer; approval required
- update_customer:
  - endpoint: PUT /v2/customers/{{ record.id }}
  - required fields: id
  - risk: external mutation; overwrites a live Clockodo customer's fields; approval required
- delete_customer:
  - endpoint: DELETE /v2/customers/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live Clockodo customer; approval required
- create_project:
  - endpoint: POST /v2/projects
  - risk: external mutation; creates a live Clockodo project; approval required
- update_project:
  - endpoint: PUT /v2/projects/{{ record.id }}
  - required fields: id
  - risk: external mutation; overwrites a live Clockodo project's fields; approval required
- delete_project:
  - endpoint: DELETE /v2/projects/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly removes a live Clockodo project; approval required
- create_service:
  - endpoint: POST /v2/services
  - risk: external mutation; creates a live Clockodo service; approval required
- update_service:
  - endpoint: PUT /v2/services/{{ record.id }}
  - required fields: id
  - risk: external mutation; overwrites a live Clockodo service's fields; approval required
- delete_service:
  - endpoint: DELETE /v2/services/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live Clockodo service; approval required
- create_team:
  - endpoint: POST /v2/teams
  - risk: external mutation; creates a live Clockodo team; approval required
- update_team:
  - endpoint: PUT /v2/teams/{{ record.id }}
  - required fields: id
  - risk: external mutation; overwrites a live Clockodo team's fields; approval required
- delete_team:
  - endpoint: DELETE /v2/teams/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live Clockodo team; approval required
- create_lumpsum_service:
  - endpoint: POST /v2/lumpsumservices
  - risk: external mutation; creates a live Clockodo lump-sum service; approval required
- update_lumpsum_service:
  - endpoint: PUT /v2/lumpsumservices/{{ record.id }}
  - required fields: id
  - risk: external mutation; overwrites a live Clockodo lump-sum service's fields; approval required
- delete_lumpsum_service:
  - endpoint: DELETE /v2/lumpsumservices/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live Clockodo lump-sum service; approval required

## Security

- read risk: external Clockodo API read of customer, project, service, user, time-entry, absence, and workspace-configuration data
- write risk: external mutation; creates/updates/deletes live Clockodo customers, projects, services, teams, and lump-sum services
- approval: required for all write actions; reads remain none
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect clockodo
```

### Inspect as structured JSON

```bash
pm connectors inspect clockodo --json
```

## Agent Rules

- Run pm connectors inspect clockodo before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
