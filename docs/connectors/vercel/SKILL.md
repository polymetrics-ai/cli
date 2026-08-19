---
name: pm-vercel
description: Vercel connector knowledge and safe action guide.
---

# pm-vercel

## Purpose

Reads deployments, projects, teams, domains, aliases, webhooks, log drains, and edge configs from the Vercel REST API, and writes projects, deployments, domains, project environment variables, webhooks, log drains, edge configs, and alias removal.

## Icon

- id: simple-icons-vercel
- asset: icons/simple-icons/vercel.svg
- title: Vercel
- simple_icon_slug: vercel
- simple_icon_hex: 000000
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Vercel
- match: exact-name-or-slug
- matched_by: vercel

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- start_date
- team_id
- access_token (secret) (required)

## ETL Streams

- deployments:
  - primary key: id
  - cursor: created
  - fields: created(integer), id(string), name(string), state(string)
- projects:
  - primary key: id
  - fields: accountId(string), createdAt(integer), framework(string), id(string), name(string), updatedAt(integer)
- teams:
  - primary key: id
  - fields: id(string), name(string), slug(string)
- domains:
  - primary key: id
  - fields: createdAt(integer), id(string), name(string), teamId(string), verified(boolean)
- project_env_vars:
  - primary key: id
  - fields: createdAt(integer), id(string), key(string), project_id(string), target(array), type(string), updatedAt(integer)
- aliases:
  - primary key: uid
  - fields: alias(string), created(string), createdAt(integer), deployment(object), deploymentId(string), projectId(string), uid(string)
- webhooks:
  - primary key: id
  - fields: createdAt(integer), events(array), id(string), ownerId(string), projectIds(array), updatedAt(integer), url(string)
- log_drains:
  - primary key: id
  - fields: createdAt(integer), deliveryFormat(string), environments(array), id(string), name(string), projectIds(array), samplingRate(number), sources(array), updatedAt(integer), url(string)
- edge_configs:
  - primary key: id
  - fields: createdAt(integer), digest(string), id(string), itemCount(integer), ownerId(string), sizeInBytes(integer), slug(string), updatedAt(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_project:
  - endpoint: POST /v11/projects
  - required fields: name
  - risk: external mutation; approval required
- update_project:
  - endpoint: PATCH /v9/projects/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_project:
  - endpoint: DELETE /v9/projects/{{ record.id }}
  - required fields: id
  - risk: destructive external mutation; approval required
- create_deployment:
  - endpoint: POST /v13/deployments
  - required fields: name
  - risk: external mutation; approval required
- cancel_deployment:
  - endpoint: PATCH /v12/deployments/{{ record.id }}/cancel
  - required fields: id
  - risk: external mutation; approval required
- delete_deployment:
  - endpoint: DELETE /v13/deployments/{{ record.id }}
  - required fields: id
  - risk: destructive external mutation; approval required
- add_project_domain:
  - endpoint: POST /v10/projects/{{ record.project_id }}/domains
  - required fields: project_id, name
  - risk: external mutation; approval required
- remove_project_domain:
  - endpoint: DELETE /v9/projects/{{ record.project_id }}/domains/{{ record.domain }}
  - required fields: project_id, domain
  - risk: destructive external mutation; approval required
- create_project_env_var:
  - endpoint: POST /v10/projects/{{ record.project_id }}/env
  - required fields: project_id, key, value, type
  - risk: external mutation; approval required
- delete_project_env_var:
  - endpoint: DELETE /v9/projects/{{ record.project_id }}/env/{{ record.id }}
  - required fields: project_id, id
  - risk: destructive external mutation; approval required
- create_webhook:
  - endpoint: POST /v1/webhooks
  - required fields: url, events
  - risk: external mutation; approval required
- delete_webhook:
  - endpoint: DELETE /v1/webhooks/{{ record.id }}
  - required fields: id
  - risk: destructive external mutation; approval required
- create_log_drain:
  - endpoint: POST /v1/log-drains
  - required fields: deliveryFormat, url, sources
  - risk: external mutation; approval required
- delete_log_drain:
  - endpoint: DELETE /v1/log-drains/{{ record.id }}
  - required fields: id
  - risk: destructive external mutation; approval required
- create_edge_config:
  - endpoint: POST /v1/edge-config
  - required fields: slug
  - risk: external mutation; approval required
- update_edge_config:
  - endpoint: PUT /v1/edge-config/{{ record.id }}
  - required fields: id, slug
  - risk: external mutation; approval required
- delete_edge_config:
  - endpoint: DELETE /v1/edge-config/{{ record.id }}
  - required fields: id
  - risk: destructive external mutation; approval required
- delete_alias:
  - endpoint: DELETE /v2/aliases/{{ record.uid }}
  - required fields: uid
  - risk: destructive external mutation (removes a deployment alias); approval required

## Security

- read risk: external Vercel API read of deployment, project, team, domain, alias, webhook, log-drain, and edge-config data
- write risk: external mutation of Vercel projects, deployments, domains, environment variables, webhooks, log drains, edge configs, and aliases; approval required
- approval: read: none; write: required for every action
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Read Vercel streams and plan declared project updates through reverse ETL.
- Usage: pm vercel <command> [flags]
- Source CLI: Vercel REST API (PATCH /v9/projects/{id})
- Global flags:
  - --credential (string): Credential name to use for the Vercel request.
  - --connection (string): Credential alias used only when --credential is omitted.
  - --config (string_array): Connector config override as key=value; never pass secret values here.
  - --json (boolean): Emit machine-readable JSON output.
  - --plan (string): Execute an approved reverse-ETL plan by id.
  - --preview (boolean): Preview a reverse-ETL write command without making a network mutation.
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
- Projects
  - projects list - Read Vercel projects through the declared ETL stream. [intent=etl availability=implemented stream=projects]
  - projects update - Plan an update to a Vercel project through reverse ETL. [intent=reverse_etl availability=implemented write=update_project]; approval: Plan first, inspect preview output, then run only with the generated approval token.; risk: Updates a Vercel project; requires reverse-ETL plan, preview, explicit approval, then execute.; flags: --id (required), --name, --build-command
- Help topics:
  - execution-model - Reverse ETL uses plan, preview, approval, and execute; provider-live certification remains pending.

## Sync Transport

- Source transport: declared
- Destination transport: declared
- A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.
- Source executor: declarative_api/declarative_stream_source
- Destination executor: declarative_api/declarative_typed_destination

## Commands

### Inspect as a manual

```bash
pm connectors inspect vercel
```

### Inspect as structured JSON

```bash
pm connectors inspect vercel --json
```

## Agent Rules

- Run pm connectors inspect vercel before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
