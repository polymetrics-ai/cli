---
name: pm-datadog
description: Datadog connector knowledge and safe action guide.
---

# pm-datadog

## Purpose

Reads Datadog monitors, dashboards, dashboard lists, users, SLOs, SLO corrections, scheduled downtimes, notebooks, organizations, hosts, Synthetics tests/locations/variables, and API/application keys, and writes monitor/dashboard/downtime/notebook/SLO/user/event/Synthetics-test/API-key mutations, through the Datadog v1 REST API.

## Icon

- id: datadog
- asset: icons/datadog.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.datadoghq.com/api/latest/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret) (required)
- application_key (secret) (required)

## ETL Streams

- monitors:
  - primary key: id
  - cursor: modified
  - fields: created(string), id(integer), message(string), modified(string), name(string), overall_state(string), priority(integer), query(string), type(string)
- dashboards:
  - primary key: id
  - fields: author_handle(string), created_at(string), description(string), id(string), is_read_only(boolean), layout_type(string), modified_at(string), title(string), url(string)
- users:
  - primary key: id
  - fields: created_at(string), disabled(boolean), email(string), handle(string), id(string), name(string), status(string), type(string), verified(boolean)
- slo:
  - primary key: id
  - fields: created_at(integer), description(string), id(string), modified_at(integer), name(string), type(string)
- downtimes:
  - primary key: id
  - fields: active(boolean), disabled(boolean), end(integer), id(integer), message(string), monitor_id(integer), scope(string), start(integer)
- dashboard_lists:
  - primary key: id
  - cursor: modified
  - fields: created(string), dashboard_count(integer), id(integer), is_favorite(boolean), modified(string), name(string), type(string)
- notebooks:
  - primary key: id
  - cursor: modified
  - fields: author_handle(string), created(string), id(integer), modified(string), name(string), type(string)
- organizations:
  - primary key: public_id
  - fields: created(string), description(string), name(string), public_id(string), trial(boolean)
- hosts:
  - primary key: id
  - cursor: last_reported_time
  - fields: aliases(array), apps(array), aws_name(string), host_name(string), id(integer), is_muted(boolean), last_reported_time(integer), mute_timeout(integer), name(string), sources(array), up(boolean)
- slo_corrections:
  - primary key: id
  - cursor: modified_at
  - fields: category(string), created_at(integer), description(string), duration(integer), end(integer), id(string), modified_at(integer), slo_id(string), start(integer), timezone(string), type(string)
- synthetics_tests:
  - primary key: public_id
  - fields: locations(array), message(string), monitor_id(integer), name(string), public_id(string), status(string), subtype(string), tags(array), type(string)
- synthetics_locations:
  - primary key: id
  - fields: id(string), name(string)
- synthetics_variables:
  - primary key: id
  - fields: description(string), id(string), is_fido(boolean), is_totp(boolean), name(string), parse_test_public_id(string), tags(array)
- api_keys:
  - primary key: key
  - cursor: created
  - fields: created(string), created_by(string), key(string), name(string)
- application_keys:
  - primary key: hash
  - fields: hash(string), name(string), owner(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_monitor:
  - endpoint: POST /api/v1/monitor
  - required fields: name, type, query, message
  - risk: creates a new alerting monitor; low-risk external mutation, no approval required
- update_monitor:
  - endpoint: PUT /api/v1/monitor/{{ record.id }}
  - required fields: id
  - risk: mutates an existing monitor's alert condition/notification message; a changed query/threshold affects live alerting behavior, approval required
- delete_monitor:
  - endpoint: DELETE /api/v1/monitor/{{ record.id }}
  - required fields: id
  - risk: irreversibly removes a monitor and its alerting history reference; approval required
- create_dashboard:
  - endpoint: POST /api/v1/dashboard
  - required fields: title, layout_type, widgets
  - risk: creates a new dashboard; low-risk external mutation, no approval required
- update_dashboard:
  - endpoint: PUT /api/v1/dashboard/{{ record.id }}
  - required fields: id, title, layout_type, widgets
  - risk: replaces an existing dashboard's full widget layout; external mutation, approval required
- delete_dashboard:
  - endpoint: DELETE /api/v1/dashboard/{{ record.id }}
  - required fields: id
  - risk: irreversibly removes a dashboard; approval required
- create_dashboard_list:
  - endpoint: POST /api/v1/dashboard/lists/manual
  - required fields: name
  - risk: creates a new dashboard list (folder); low-risk external mutation, no approval required
- update_dashboard_list:
  - endpoint: PUT /api/v1/dashboard/lists/manual/{{ record.id }}
  - required fields: id, name
  - risk: renames an existing dashboard list; external mutation, approval required
- delete_dashboard_list:
  - endpoint: DELETE /api/v1/dashboard/lists/manual/{{ record.id }}
  - required fields: id
  - risk: irreversibly removes a dashboard list (folder); the dashboards themselves are unaffected, approval required
- create_downtime:
  - endpoint: POST /api/v1/downtime
  - required fields: scope
  - risk: schedules a downtime that silences monitor alerts for the given scope; suppresses real alerting during the window, approval required
- update_downtime:
  - endpoint: PUT /api/v1/downtime/{{ record.id }}
  - required fields: id
  - risk: mutates an existing downtime's window/scope; changes which alerts are currently suppressed, approval required
- cancel_downtime:
  - endpoint: DELETE /api/v1/downtime/{{ record.id }}
  - required fields: id
  - risk: cancels a scheduled/active downtime; alerting resumes immediately for its scope, approval required
- create_notebook:
  - endpoint: POST /api/v1/notebooks
  - required fields: name, cells, time
  - risk: creates a new notebook; low-risk external mutation, no approval required
- update_notebook:
  - endpoint: PUT /api/v1/notebooks/{{ record.id }}
  - required fields: id, name, cells, time
  - risk: replaces an existing notebook's content; external mutation, approval required
- delete_notebook:
  - endpoint: DELETE /api/v1/notebooks/{{ record.id }}
  - required fields: id
  - risk: irreversibly removes a notebook; approval required
- create_slo:
  - endpoint: POST /api/v1/slo
  - required fields: name, type, thresholds
  - risk: creates a new SLO target; low-risk external mutation, no approval required
- update_slo:
  - endpoint: PUT /api/v1/slo/{{ record.id }}
  - required fields: id
  - risk: mutates an existing SLO's target thresholds; affects SLO burn-rate alerting, approval required
- delete_slo:
  - endpoint: DELETE /api/v1/slo/{{ record.id }}
  - required fields: id
  - risk: irreversibly removes an SLO and its historical error-budget tracking; approval required
- create_user:
  - endpoint: POST /api/v1/user
  - required fields: email
  - risk: invites a new user into the Datadog organization with the given role; approval required
- update_user:
  - endpoint: PUT /api/v1/user/{{ record.handle }}
  - required fields: handle
  - risk: mutates an existing user's role/profile; a changed access_role directly changes that user's permissions, approval required
- disable_user:
  - endpoint: DELETE /api/v1/user/{{ record.handle }}
  - required fields: handle
  - risk: disables a user's access to the Datadog organization; approval required
- create_event:
  - endpoint: POST /api/v1/events
  - required fields: title, text
  - risk: posts a custom event into the Datadog event stream; low-risk external mutation, no approval required
- create_synthetics_api_test:
  - endpoint: POST /api/v1/synthetics/tests/api
  - required fields: name, type, config, locations
  - risk: creates a new Synthetics API test that begins actively probing the configured URL/host on a schedule; low-risk external mutation, no approval required
- update_synthetics_api_test:
  - endpoint: PUT /api/v1/synthetics/tests/api/{{ record.public_id }}
  - required fields: public_id
  - risk: mutates an existing Synthetics API test's request target/assertions; changes what is actively probed, approval required
- create_api_key:
  - endpoint: POST /api/v1/api_key
  - required fields: name
  - risk: creates a new organization API key with full agent-submission scope; a newly-minted long-lived credential, approval required
- update_api_key:
  - endpoint: PUT /api/v1/api_key/{{ record.key }}
  - required fields: key, name
  - risk: renames an existing API key; low-risk external mutation, no approval required
- delete_api_key:
  - endpoint: DELETE /api/v1/api_key/{{ record.key }}
  - required fields: key
  - risk: irreversibly revokes an organization API key; every agent/integration still using it immediately loses ingest access, approval required

## Security

- read risk: external Datadog API read of monitor, dashboard, SLO, downtime, notebook, organization, host, Synthetics, and API/application key configuration data
- write risk: external mutation of Datadog monitors, dashboards, downtimes, notebooks, SLOs, users, events, Synthetics API tests, and API keys; create_downtime/update_downtime suppress real alerting for their scope, update_monitor/update_slo change live alerting/burn-rate thresholds, and delete_api_key immediately revokes ingest access for anything still using it, so every write ships an explicit per-action risk string
- approval: required for every delete_*/cancel_downtime action (irreversible or alerting-suppressing) and for update_monitor/update_downtime/update_slo/update_user/update_synthetics_api_test/create_downtime/create_user (live alerting or access-control side effects); create_monitor/create_dashboard/create_dashboard_list/create_notebook/create_slo/create_event/create_synthetics_api_test/update_dashboard_list/update_api_key are low-risk
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Datadog's declared typed write actions.
- Usage: pm datadog <command> [flags]
- Global flags:
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
- Reverse ETL writes
- Other Commands
  - cancel downtime apply - Typed action cancel_downtime [intent=reverse_etl availability=partial write=cancel_downtime]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/downtime/{id} disagrees with covered api_surface path /api/v1/downtime/{downtime_id}.; risk: cancels a scheduled/active downtime; alerting resumes immediately for its scope, approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/downtime/{id} disagrees with covered api_surface path /api/v1/downtime/{downtime_id}.; flags: --id (required)
  - create api key apply - POST /api/v1/api_key (create_api_key) [intent=reverse_etl availability=implemented write=create_api_key]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: creates a new organization API key with full agent-submission scope; a newly-minted long-lived credential, approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --name (required)
  - create dashboard apply - POST /api/v1/dashboard (create_dashboard) [intent=reverse_etl availability=implemented write=create_dashboard]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: creates a new dashboard; low-risk external mutation, no approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --layout-type (required), --title (required), --widgets (required)
  - create dashboard list apply - POST /api/v1/dashboard/lists/manual (create_dashboard_list) [intent=reverse_etl availability=implemented write=create_dashboard_list]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: creates a new dashboard list (folder); low-risk external mutation, no approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --name (required)
  - create downtime apply - POST /api/v1/downtime (create_downtime) [intent=reverse_etl availability=implemented write=create_downtime]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: schedules a downtime that silences monitor alerts for the given scope; suppresses real alerting during the window, approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --scope (required)
  - create event apply - POST /api/v1/events (create_event) [intent=reverse_etl availability=implemented write=create_event]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: posts a custom event into the Datadog event stream; low-risk external mutation, no approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --text (required), --title (required)
  - create monitor apply - POST /api/v1/monitor (create_monitor) [intent=reverse_etl availability=implemented write=create_monitor]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: creates a new alerting monitor; low-risk external mutation, no approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --message (required), --name (required), --query (required), --type (required)
  - create notebook apply - POST /api/v1/notebooks (create_notebook) [intent=reverse_etl availability=implemented write=create_notebook]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: creates a new notebook; low-risk external mutation, no approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --cells (required), --name (required), --time (required)
  - create slo apply - POST /api/v1/slo (create_slo) [intent=reverse_etl availability=implemented write=create_slo]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: creates a new SLO target; low-risk external mutation, no approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --name (required), --thresholds (required), --type (required)
  - create synthetics api test apply - POST /api/v1/synthetics/tests/api (create_synthetics_api_test) [intent=reverse_etl availability=implemented write=create_synthetics_api_test]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: creates a new Synthetics API test that begins actively probing the configured URL/host on a schedule; low-risk external mutation, no approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --config (required), --locations (required), --name (required), --type (required)
  - create user apply - POST /api/v1/user (create_user) [intent=reverse_etl availability=implemented write=create_user]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: invites a new user into the Datadog organization with the given role; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --email (required)
  - delete api key apply - DELETE /api/v1/api_key/{key} (delete_api_key) [intent=reverse_etl availability=implemented write=delete_api_key]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: irreversibly revokes an organization API key; every agent/integration still using it immediately loses ingest access, approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --key (required)
  - delete dashboard apply - Typed action delete_dashboard [intent=reverse_etl availability=partial write=delete_dashboard]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/dashboard/{id} disagrees with covered api_surface path /api/v1/dashboard/{dashboard_id}.; risk: irreversibly removes a dashboard; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/dashboard/{id} disagrees with covered api_surface path /api/v1/dashboard/{dashboard_id}.; flags: --id (required)
  - delete dashboard list apply - Typed action delete_dashboard_list [intent=reverse_etl availability=partial write=delete_dashboard_list]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/dashboard/lists/manual/{id} disagrees with covered api_surface path /api/v1/dashboard/lists/manual/{list_id}.; risk: irreversibly removes a dashboard list (folder); the dashboards themselves are unaffected, approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/dashboard/lists/manual/{id} disagrees with covered api_surface path /api/v1/dashboard/lists/manual/{list_id}.; flags: --id (required)
  - delete monitor apply - Typed action delete_monitor [intent=reverse_etl availability=partial write=delete_monitor]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/monitor/{id} disagrees with covered api_surface path /api/v1/monitor/{monitor_id}.; risk: irreversibly removes a monitor and its alerting history reference; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/monitor/{id} disagrees with covered api_surface path /api/v1/monitor/{monitor_id}.; flags: --id (required)
  - delete notebook apply - Typed action delete_notebook [intent=reverse_etl availability=partial write=delete_notebook]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/notebooks/{id} disagrees with covered api_surface path /api/v1/notebooks/{notebook_id}.; risk: irreversibly removes a notebook; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/notebooks/{id} disagrees with covered api_surface path /api/v1/notebooks/{notebook_id}.; flags: --id (required)
  - delete slo apply - Typed action delete_slo [intent=reverse_etl availability=partial write=delete_slo]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/slo/{id} disagrees with covered api_surface path /api/v1/slo/{slo_id}.; risk: irreversibly removes an SLO and its historical error-budget tracking; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/slo/{id} disagrees with covered api_surface path /api/v1/slo/{slo_id}.; flags: --id (required)
  - disable user apply - Typed action disable_user [intent=reverse_etl availability=partial write=disable_user]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/user/{handle} disagrees with covered api_surface path /api/v1/user/{user_handle}.; risk: disables a user's access to the Datadog organization; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/user/{handle} disagrees with covered api_surface path /api/v1/user/{user_handle}.; flags: --handle (required)
  - update api key apply - PUT /api/v1/api_key/{key} (update_api_key) [intent=reverse_etl availability=implemented write=update_api_key]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: renames an existing API key; low-risk external mutation, no approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --key (required), --name (required)
  - update dashboard apply - Typed action update_dashboard [intent=reverse_etl availability=partial write=update_dashboard]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/dashboard/{id} disagrees with covered api_surface path /api/v1/dashboard/{dashboard_id}.; risk: replaces an existing dashboard's full widget layout; external mutation, approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/dashboard/{id} disagrees with covered api_surface path /api/v1/dashboard/{dashboard_id}.; flags: --id (required), --layout-type (required), --title (required), --widgets (required)
  - update dashboard list apply - Typed action update_dashboard_list [intent=reverse_etl availability=partial write=update_dashboard_list]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/dashboard/lists/manual/{id} disagrees with covered api_surface path /api/v1/dashboard/lists/manual/{list_id}.; risk: renames an existing dashboard list; external mutation, approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/dashboard/lists/manual/{id} disagrees with covered api_surface path /api/v1/dashboard/lists/manual/{list_id}.; flags: --id (required), --name (required)
  - update downtime apply - Typed action update_downtime [intent=reverse_etl availability=partial write=update_downtime]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/downtime/{id} disagrees with covered api_surface path /api/v1/downtime/{downtime_id}.; risk: mutates an existing downtime's window/scope; changes which alerts are currently suppressed, approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/downtime/{id} disagrees with covered api_surface path /api/v1/downtime/{downtime_id}.; flags: --id (required)
  - update monitor apply - Typed action update_monitor [intent=reverse_etl availability=partial write=update_monitor]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/monitor/{id} disagrees with covered api_surface path /api/v1/monitor/{monitor_id}.; risk: mutates an existing monitor's alert condition/notification message; a changed query/threshold affects live alerting behavior, approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/monitor/{id} disagrees with covered api_surface path /api/v1/monitor/{monitor_id}.; flags: --id (required)
  - update notebook apply - Typed action update_notebook [intent=reverse_etl availability=partial write=update_notebook]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/notebooks/{id} disagrees with covered api_surface path /api/v1/notebooks/{notebook_id}.; risk: replaces an existing notebook's content; external mutation, approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/notebooks/{id} disagrees with covered api_surface path /api/v1/notebooks/{notebook_id}.; flags: --cells (required), --id (required), --name (required), --time (required)
  - update slo apply - Typed action update_slo [intent=reverse_etl availability=partial write=update_slo]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/slo/{id} disagrees with covered api_surface path /api/v1/slo/{slo_id}.; risk: mutates an existing SLO's target thresholds; affects SLO burn-rate alerting, approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/slo/{id} disagrees with covered api_surface path /api/v1/slo/{slo_id}.; flags: --id (required)
  - update synthetics api test apply - PUT /api/v1/synthetics/tests/api/{public_id} (update_synthetics_api_test) [intent=reverse_etl availability=implemented write=update_synthetics_api_test]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: mutates an existing Synthetics API test's request target/assertions; changes what is actively probed, approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --public-id (required)
  - update user apply - Typed action update_user [intent=reverse_etl availability=partial write=update_user]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/user/{handle} disagrees with covered api_surface path /api/v1/user/{user_handle}.; risk: mutates an existing user's role/profile; a changed access_role directly changes that user's permissions, approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/user/{handle} disagrees with covered api_surface path /api/v1/user/{user_handle}.; flags: --handle (required)

## Sync Transport

- Source transport: declared
- Destination transport: unsupported
- A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.
- Source executor: declarative_api/declarative_stream_source

## Commands

### Inspect as a manual

```bash
pm connectors inspect datadog
```

### Inspect as structured JSON

```bash
pm connectors inspect datadog --json
```

## Agent Rules

- Run pm connectors inspect datadog before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
