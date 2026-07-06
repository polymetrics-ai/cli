---
name: pm-vercel
description: Vercel connector knowledge and safe action guide.
---

# pm-vercel

## Purpose

Reads deployments, projects, teams, domains, aliases, webhooks, log drains, and edge configs from the Vercel REST API, and writes projects, deployments, domains, project environment variables, webhooks, log drains, edge configs, and alias removal.

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
- start_date
- team_id
- access_token (secret)

## ETL Streams

- deployments:
  - primary key: id
  - cursor: created
  - fields: created(), id(), name(), state()
- projects:
  - primary key: id
  - fields: accountId(), createdAt(), framework(), id(), name(), updatedAt()
- teams:
  - primary key: id
  - fields: id(), name(), slug()
- domains:
  - primary key: id
  - fields: createdAt(), id(), name(), teamId(), verified()
- project_env_vars:
  - primary key: id
  - fields: createdAt(), id(), key(), project_id(), target(), type(), updatedAt()
- aliases:
  - primary key: uid
  - fields: alias(), created(), createdAt(), deployment(), deploymentId(), projectId(), uid()
- webhooks:
  - primary key: id
  - fields: createdAt(), events(), id(), ownerId(), projectIds(), updatedAt(), url()
- log_drains:
  - primary key: id
  - fields: createdAt(), deliveryFormat(), environments(), id(), name(), projectIds(), samplingRate(), sources(), updatedAt(), url()
- edge_configs:
  - primary key: id
  - fields: createdAt(), digest(), id(), itemCount(), ownerId(), sizeInBytes(), slug(), updatedAt()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_project:
  - endpoint: POST /v11/projects
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
  - required fields: project_id
  - risk: external mutation; approval required
- remove_project_domain:
  - endpoint: DELETE /v9/projects/{{ record.project_id }}/domains/{{ record.domain }}
  - required fields: project_id, domain
  - risk: destructive external mutation; approval required
- create_project_env_var:
  - endpoint: POST /v10/projects/{{ record.project_id }}/env
  - required fields: project_id
  - risk: external mutation; approval required
- delete_project_env_var:
  - endpoint: DELETE /v9/projects/{{ record.project_id }}/env/{{ record.id }}
  - required fields: project_id, id
  - risk: destructive external mutation; approval required
- create_webhook:
  - endpoint: POST /v1/webhooks
  - risk: external mutation; approval required
- delete_webhook:
  - endpoint: DELETE /v1/webhooks/{{ record.id }}
  - required fields: id
  - risk: destructive external mutation; approval required
- create_log_drain:
  - endpoint: POST /v1/log-drains
  - risk: external mutation; approval required
- delete_log_drain:
  - endpoint: DELETE /v1/log-drains/{{ record.id }}
  - required fields: id
  - risk: destructive external mutation; approval required
- create_edge_config:
  - endpoint: POST /v1/edge-config
  - risk: external mutation; approval required
- update_edge_config:
  - endpoint: PUT /v1/edge-config/{{ record.id }}
  - required fields: id
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
