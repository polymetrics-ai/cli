# pm connectors inspect statuspage

```text
NAME
  pm connectors inspect statuspage - Statuspage connector manual

SYNOPSIS
  pm connectors inspect statuspage
  pm connectors inspect statuspage --json
  pm credentials add <name> --connector statuspage [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Statuspage pages, components, incidents, subscribers, component groups, metrics, metrics providers, page access groups/users, and incident templates through the Statuspage API.

ICON
  asset: icons/statuspage.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.statuspage.io/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  page_id
  api_key (secret)

ETL STREAMS
  pages:
    primary key: id
    fields: id(), name(), url()
  components:
    primary key: id
    fields: created_at(), id(), name(), status()
  incidents:
    primary key: id
    cursor: created_at
    fields: created_at(), id(), name(), status()
  subscribers:
    primary key: id
    fields: created_at(), id(), name(), status()
  component_groups:
    primary key: id
    fields: created_at(), description(), id(), name(), page_id(), position(), updated_at()
  metrics:
    primary key: id
    fields: backfilled(), created_at(), decimal_places(), display(), id(), last_fetched_at(), metric_identifier(), metrics_provider_id(), most_recent_data_at(), name(), suffix(), tooltip_description(), updated_at()
  metrics_providers:
    primary key: id
    fields: created_at(), disabled(), id(), last_revalidated_at(), metric_base_uri(), page_id(), type(), updated_at()
  page_access_groups:
    primary key: id
    fields: component_ids(), created_at(), external_identifier(), id(), metric_ids(), name(), page_access_user_ids(), page_id(), updated_at()
  page_access_users:
    primary key: id
    fields: created_at(), email(), external_login(), id(), page_access_group_id(), page_access_group_ids(), page_id(), updated_at()
  incident_templates:
    primary key: id
    fields: body(), group_id(), id(), name(), should_send_notifications(), should_tweet(), title(), update_status()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Statuspage API read of page, component, incident, subscriber, component group, metric, metrics provider, page access group/user, and incident template data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect statuspage

  # Inspect as structured JSON
  pm connectors inspect statuspage --json

AGENT WORKFLOW
  - Run pm connectors inspect statuspage before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
