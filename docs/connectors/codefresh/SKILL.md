---
name: pm-codefresh
description: Codefresh connector knowledge and safe action guide.
---

# pm-codefresh

## Purpose

Reads Codefresh projects, pipelines, builds, runner agents, shared contexts, container images, registries, triggers, and annotations, and can create/update/delete/run projects, pipelines, contexts, and agents through the Codefresh REST API.

## Icon

- id: simple-icons-codefresh
- asset: icons/simple-icons/codefresh.svg
- title: Codefresh
- simple_icon_slug: codefresh
- simple_icon_hex: 08B1AB
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Codefresh
- match: exact-name-or-slug
- matched_by: codefresh

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_id
- base_url
- mode
- api_key (secret) (required)

## ETL Streams

- projects:
  - primary key: id
  - fields: favorite(boolean), id(string), pipelines_number(integer), project_name(string), updated_at(string)
- pipelines:
  - primary key: id
  - fields: created_at(string), id(string), is_public(boolean), name(string), project(string), updated_at(string)
- agents:
  - primary key: id
  - fields: created_at(string), id(string), name(string), status(string), version(string)
- contexts:
  - primary key: id
  - fields: id(string), name(string), owner(string), type(string)
- builds:
  - primary key: id
  - cursor: created
  - fields: branch_name(string), commit_message(string), committer(string), created(string), finished(string), id(string), pipeline_name(string), progress(string), project(string), project_id(string), provider(string), repo_name(string), repo_owner(string), revision(string), status(string), trigger(string), trigger_type(string), triggered_by(string)
- images:
  - primary key: id
  - cursor: created
  - fields: branch(string), commit(string), commit_url(string), created(string), id(string), image_display_name(string), image_name(string), repo(string), sha(string), size(integer)
- registries:
  - primary key: id
  - fields: behind_firewall(boolean), default(boolean), domain(string), id(string), internal(boolean), kind(string), name(string), primary(boolean), provider(string)
- triggers:
  - primary key: event, pipeline
  - fields: event(string), event_description(string), event_status(string), event_type(string), filter_tag(string), pipeline(string)
- trigger_events:
  - primary key: uri
  - fields: account(string), description(string), endpoint(string), kind(string), status(string), type(string), uri(string)
- annotations:
  - primary key: id
  - fields: account_id(string), entity_id(string), entity_type(string), id(string), key(string), type(string), value(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_project:
  - endpoint: POST /projects
  - required fields: projectName
  - risk: external mutation; creates a new Codefresh project; approval required
- delete_project:
  - endpoint: DELETE /projects/{{ record.id }}
  - required fields: id
  - risk: destructive; irreversible deletion of a Codefresh project; approval required
- create_pipeline:
  - endpoint: POST /pipelines
  - required fields: metadata
  - risk: external mutation; creates a new Codefresh pipeline; approval required
- update_pipeline:
  - endpoint: PUT /pipelines/{{ record.name }}
  - required fields: name, metadata
  - risk: external mutation; replaces an existing Codefresh pipeline's spec; approval required
- delete_pipeline:
  - endpoint: DELETE /pipelines/{{ record.name }}
  - required fields: name
  - risk: destructive; irreversible deletion of a Codefresh pipeline; approval required
- run_pipeline:
  - endpoint: POST /pipelines/run/{{ record.name }}
  - required fields: name
  - risk: external mutation; triggers a real Codefresh pipeline run (build minutes/resources consumed); approval required
- create_context:
  - endpoint: POST /contexts
  - required fields: metadata, spec
  - risk: external mutation; creates a new Codefresh shared context (may hold configuration values); approval required
- delete_context:
  - endpoint: DELETE /contexts/{{ record.name }}
  - required fields: name
  - risk: destructive; irreversible deletion of a Codefresh shared context; approval required
- create_agent:
  - endpoint: POST /agents
  - required fields: name
  - risk: external mutation; registers a new Codefresh runner agent; approval required
- delete_agent:
  - endpoint: DELETE /agent/{{ record.id }}
  - required fields: id
  - risk: destructive; irreversible deregistration of a Codefresh runner agent; approval required

## Security

- read risk: external Codefresh API read of projects, pipelines, builds, runner agents, shared contexts, container images, registries, triggers, and annotations
- write risk: external mutation of Codefresh projects, pipelines, contexts, and runner agents, including irreversible deletes and triggering real pipeline runs (consumes build minutes/resources)
- approval: required for all write actions; read is unrestricted
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect codefresh
```

### Inspect as structured JSON

```bash
pm connectors inspect codefresh --json
```

## Agent Rules

- Run pm connectors inspect codefresh before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
