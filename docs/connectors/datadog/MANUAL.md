# pm connectors inspect datadog

```text
NAME
  pm connectors inspect datadog - Datadog connector manual

SYNOPSIS
  pm connectors inspect datadog
  pm connectors inspect datadog --json
  pm credentials add <name> --connector datadog [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Datadog monitors, dashboards, dashboard lists, users, SLOs, SLO corrections, scheduled downtimes, notebooks, organizations, hosts, Synthetics tests/locations/variables, and API/application keys, and writes monitor/dashboard/downtime/notebook/SLO/user/event/Synthetics-test/API-key mutations, through the Datadog v1 REST API.

ICON
  asset: icons/datadog.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.datadoghq.com/api/latest/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret)
  application_key (secret)

ETL STREAMS
  monitors:
    primary key: id
    cursor: modified
    fields: created(), id(), message(), modified(), name(), overall_state(), priority(), query(), type()
  dashboards:
    primary key: id
    fields: author_handle(), created_at(), description(), id(), is_read_only(), layout_type(), modified_at(), title(), url()
  users:
    primary key: id
    fields: created_at(), disabled(), email(), handle(), id(), name(), status(), type(), verified()
  slo:
    primary key: id
    fields: created_at(), description(), id(), modified_at(), name(), type()
  downtimes:
    primary key: id
    fields: active(), disabled(), end(), id(), message(), monitor_id(), scope(), start()
  dashboard_lists:
    primary key: id
    cursor: modified
    fields: created(), dashboard_count(), id(), is_favorite(), modified(), name(), type()
  notebooks:
    primary key: id
    cursor: modified
    fields: author_handle(), created(), id(), modified(), name(), type()
  organizations:
    primary key: public_id
    fields: created(), description(), name(), public_id(), trial()
  hosts:
    primary key: id
    cursor: last_reported_time
    fields: aliases(), apps(), aws_name(), host_name(), id(), is_muted(), last_reported_time(), mute_timeout(), name(), sources(), up()
  slo_corrections:
    primary key: id
    cursor: modified_at
    fields: category(), created_at(), description(), duration(), end(), id(), modified_at(), slo_id(), start(), timezone(), type()
  synthetics_tests:
    primary key: public_id
    fields: locations(), message(), monitor_id(), name(), public_id(), status(), subtype(), tags(), type()
  synthetics_locations:
    primary key: id
    fields: id(), name()
  synthetics_variables:
    primary key: id
    fields: description(), id(), is_fido(), is_totp(), name(), parse_test_public_id(), tags()
  api_keys:
    primary key: key
    cursor: created
    fields: created(), created_by(), key(), name()
  application_keys:
    primary key: hash
    fields: hash(), name(), owner()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_monitor:
    endpoint: POST /api/v1/monitor
    risk: creates a new alerting monitor; low-risk external mutation, no approval required
  update_monitor:
    endpoint: PUT /api/v1/monitor/{{ record.id }}
    required fields: id
    risk: mutates an existing monitor's alert condition/notification message; a changed query/threshold affects live alerting behavior, approval required
  delete_monitor:
    endpoint: DELETE /api/v1/monitor/{{ record.id }}
    required fields: id
    risk: irreversibly removes a monitor and its alerting history reference; approval required
  create_dashboard:
    endpoint: POST /api/v1/dashboard
    risk: creates a new dashboard; low-risk external mutation, no approval required
  update_dashboard:
    endpoint: PUT /api/v1/dashboard/{{ record.id }}
    required fields: id
    risk: replaces an existing dashboard's full widget layout; external mutation, approval required
  delete_dashboard:
    endpoint: DELETE /api/v1/dashboard/{{ record.id }}
    required fields: id
    risk: irreversibly removes a dashboard; approval required
  create_dashboard_list:
    endpoint: POST /api/v1/dashboard/lists/manual
    risk: creates a new dashboard list (folder); low-risk external mutation, no approval required
  update_dashboard_list:
    endpoint: PUT /api/v1/dashboard/lists/manual/{{ record.id }}
    required fields: id
    risk: renames an existing dashboard list; external mutation, approval required
  delete_dashboard_list:
    endpoint: DELETE /api/v1/dashboard/lists/manual/{{ record.id }}
    required fields: id
    risk: irreversibly removes a dashboard list (folder); the dashboards themselves are unaffected, approval required
  create_downtime:
    endpoint: POST /api/v1/downtime
    risk: schedules a downtime that silences monitor alerts for the given scope; suppresses real alerting during the window, approval required
  update_downtime:
    endpoint: PUT /api/v1/downtime/{{ record.id }}
    required fields: id
    risk: mutates an existing downtime's window/scope; changes which alerts are currently suppressed, approval required
  cancel_downtime:
    endpoint: DELETE /api/v1/downtime/{{ record.id }}
    required fields: id
    risk: cancels a scheduled/active downtime; alerting resumes immediately for its scope, approval required
  create_notebook:
    endpoint: POST /api/v1/notebooks
    risk: creates a new notebook; low-risk external mutation, no approval required
  update_notebook:
    endpoint: PUT /api/v1/notebooks/{{ record.id }}
    required fields: id
    risk: replaces an existing notebook's content; external mutation, approval required
  delete_notebook:
    endpoint: DELETE /api/v1/notebooks/{{ record.id }}
    required fields: id
    risk: irreversibly removes a notebook; approval required
  create_slo:
    endpoint: POST /api/v1/slo
    risk: creates a new SLO target; low-risk external mutation, no approval required
  update_slo:
    endpoint: PUT /api/v1/slo/{{ record.id }}
    required fields: id
    risk: mutates an existing SLO's target thresholds; affects SLO burn-rate alerting, approval required
  delete_slo:
    endpoint: DELETE /api/v1/slo/{{ record.id }}
    required fields: id
    risk: irreversibly removes an SLO and its historical error-budget tracking; approval required
  create_user:
    endpoint: POST /api/v1/user
    risk: invites a new user into the Datadog organization with the given role; approval required
  update_user:
    endpoint: PUT /api/v1/user/{{ record.handle }}
    required fields: handle
    risk: mutates an existing user's role/profile; a changed access_role directly changes that user's permissions, approval required
  disable_user:
    endpoint: DELETE /api/v1/user/{{ record.handle }}
    required fields: handle
    risk: disables a user's access to the Datadog organization; approval required
  create_event:
    endpoint: POST /api/v1/events
    risk: posts a custom event into the Datadog event stream; low-risk external mutation, no approval required
  create_synthetics_api_test:
    endpoint: POST /api/v1/synthetics/tests/api
    risk: creates a new Synthetics API test that begins actively probing the configured URL/host on a schedule; low-risk external mutation, no approval required
  update_synthetics_api_test:
    endpoint: PUT /api/v1/synthetics/tests/api/{{ record.public_id }}
    required fields: public_id
    risk: mutates an existing Synthetics API test's request target/assertions; changes what is actively probed, approval required
  create_api_key:
    endpoint: POST /api/v1/api_key
    risk: creates a new organization API key with full agent-submission scope; a newly-minted long-lived credential, approval required
  update_api_key:
    endpoint: PUT /api/v1/api_key/{{ record.key }}
    required fields: key
    risk: renames an existing API key; low-risk external mutation, no approval required
  delete_api_key:
    endpoint: DELETE /api/v1/api_key/{{ record.key }}
    required fields: key
    risk: irreversibly revokes an organization API key; every agent/integration still using it immediately loses ingest access, approval required

SECURITY
  read risk: external Datadog API read of monitor, dashboard, SLO, downtime, notebook, organization, host, Synthetics, and API/application key configuration data
  write risk: external mutation of Datadog monitors, dashboards, downtimes, notebooks, SLOs, users, events, Synthetics API tests, and API keys; create_downtime/update_downtime suppress real alerting for their scope, update_monitor/update_slo change live alerting/burn-rate thresholds, and delete_api_key immediately revokes ingest access for anything still using it, so every write ships an explicit per-action risk string
  approval: required for every delete_*/cancel_downtime action (irreversible or alerting-suppressing) and for update_monitor/update_downtime/update_slo/update_user/update_synthetics_api_test/create_downtime/create_user (live alerting or access-control side effects); create_monitor/create_dashboard/create_dashboard_list/create_notebook/create_slo/create_event/create_synthetics_api_test/update_dashboard_list/update_api_key are low-risk
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect datadog

  # Inspect as structured JSON
  pm connectors inspect datadog --json

AGENT WORKFLOW
  - Run pm connectors inspect datadog before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
