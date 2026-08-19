# pm connectors inspect chift

```text
NAME
  pm connectors inspect chift - Chift connector manual

SYNOPSIS
  pm connectors inspect chift
  pm connectors inspect chift --json
  pm credentials add <name> --connector chift [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Chift consumers, connections, syncs, integrations, datastores, and webhook event definitions through the Chift REST API using a session-token (client credentials) exchange.

ICON
  id: chift
  asset: icons/chift.svg
  source: official
  review_status: official_verified
  review_url: https://docs.chift.eu/docs/introduction/welcome

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  page_size
  account_id (secret) (required)
  client_id (secret) (required)
  client_secret (secret) (required)

ETL STREAMS
  consumers:
    primary key: consumerid
    fields: active(boolean), consumerid(string), created_on(string), email(string), name(string), phone(string), redirect_url(string)
  connections:
    primary key: connectionid
    fields: api(string), connectionid(string), consumerid(string), created_on(string), name(string), status(string)
  syncs:
    primary key: syncid
    fields: consumerid(string), created_on(string), name(string), status(string), syncid(string), updated_on(string)
  integrations:
    primary key: integrationid
    fields: api(string), description(string), icon_url(string), integrationid(integer), local_agent(boolean), logo_url(string), name(string), status(string)
  datastores:
    primary key: id
    fields: id(string), name(string), status(string)
  webhook_definitions:
    primary key: event, api
    fields: api(string), event(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  create_consumer:
    endpoint: POST /consumers
    required fields: name
    risk: external mutation; approval required
  update_consumer:
    endpoint: PATCH /consumers/{{ record.consumerid }}
    required fields: consumerid
    risk: external mutation; approval required
  delete_consumer:
    endpoint: DELETE /consumers/{{ record.consumerid }}
    required fields: consumerid
    risk: irreversible external deletion; approval required

SECURITY
  read risk: external Chift API read of consumer/connection/sync/integration/datastore/webhook-definition metadata
  write risk: external mutation of Chift consumer records (create/update/delete); approval required
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect chift

  # Inspect as structured JSON
  pm connectors inspect chift --json

AGENT WORKFLOW
  - Run pm connectors inspect chift before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
