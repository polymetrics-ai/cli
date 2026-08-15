---
name: pm-clockodo
description: Clockodo connector knowledge and safe action guide.
---

# pm-clockodo

## Purpose

Reads Clockodo customers, projects, services, users, time entries, absences, teams, surcharges, lump-sum services, nonbusiness groups/days, holiday/overtime carryovers, target hours, and current-user settings, and writes customers/projects/services/teams/lump-sum services through the Clockodo REST API.

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

- absences_year
- base_url
- email_address (required)
- entries_time_since
- entries_time_until
- external_application (required)
- language
- nonbusinessdays_year
- api_key (secret) (required)

## ETL Streams

- customers:
  - primary key: id
  - fields: active(boolean), billable_default(boolean), color(integer), id(integer), name(string), note(string), number(string)
- projects:
  - primary key: id
  - fields: active(boolean), billable_default(boolean), budget_is_hours(boolean), budget_money(number), completed(boolean), customers_id(integer), deadline(string), id(integer), name(string), note(string), number(string)
- services:
  - primary key: id
  - fields: active(boolean), id(integer), name(string), note(string), number(string)
- users:
  - primary key: id
  - fields: active(boolean), email(string), id(integer), language(string), name(string), number(string), role(string), teams_id(integer), timezone(string)
- current_user_settings:
  - fields: company(object), user(object), workTimeRegulation(object)
- teams:
  - primary key: id
  - fields: id(integer), leader(integer), name(string)
- surcharges:
  - primary key: id
  - fields: accumulation(string), id(integer), name(string), night(number), night_increased(number), nonbusiness(number), nonbusiness_special(number), saturday(number), sunday(number)
- lumpsum_services:
  - primary key: id
  - fields: active(boolean), id(integer), name(string), note(string), number(string), price(number), unit(string)
- nonbusiness_groups:
  - primary key: id
  - fields: id(integer), name(string)
- nonbusiness_days:
  - primary key: id, date
  - fields: date(string), half_day(integer), id(integer), name(string), nonbusinessgroups_id(integer)
- holidays_carry:
  - primary key: users_id, year
  - fields: count(number), note(string), users_id(integer), year(integer)
- holidays_quota:
  - primary key: id
  - fields: count(number), id(integer), users_id(integer), year_since(integer), year_until(integer)
- overtime_carry:
  - primary key: users_id, year
  - fields: hours(number), note(string), users_id(integer), year(integer)
- target_hours:
  - primary key: id
  - fields: absence_fixed_credit(boolean), compensation_daily(number), compensation_monthly(number), date_since(string), date_until(string), friday(number), id(integer), monday(number), monthly_target(number), saturday(number), sunday(number), thursday(number), tuesday(number), type(string), users_id(integer), wednesday(number)
- absences:
  - primary key: id
  - fields: approved_by(integer), count_days(number), count_hours(number), date_approved(string), date_enquired(string), date_since(string), date_until(string), id(integer), note(string), sick_note(boolean), status(integer), type(integer), users_id(integer)
- entries:
  - primary key: id
  - fields: billable(integer), clocked(boolean), clocked_offline(boolean), customers_id(integer), duration(integer), hourly_rate(number), id(integer), lumpsum(number), lumpsum_services_amount(number), lumpsum_services_id(integer), lumpsum_services_price(number), offset(integer), projects_id(integer), services_id(integer), texts_id(integer), time_clocked_since(string), time_insert(string), time_last_change(string), time_last_change_worktime(string), time_since(string), time_until(string), type(integer), users_id(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_customer:
  - endpoint: POST /v2/customers
  - required fields: name
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
  - required fields: name, customers_id
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
  - required fields: name
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
  - required fields: name
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
  - required fields: name, price
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
