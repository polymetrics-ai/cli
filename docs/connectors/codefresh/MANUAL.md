# pm connectors inspect codefresh

```text
NAME
  pm connectors inspect codefresh - Codefresh connector manual

SYNOPSIS
  pm connectors inspect codefresh
  pm connectors inspect codefresh --json
  pm credentials add <name> --connector codefresh [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Codefresh projects, pipelines, builds, runner agents, shared contexts, container images, registries, triggers, and annotations, and can create/update/delete/run projects, pipelines, contexts, and agents through the Codefresh REST API.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  account_id
  base_url
  mode
  api_key (secret)

ETL STREAMS
  projects:
    primary key: id
    fields: favorite(), id(), pipelines_number(), project_name(), updated_at()
  pipelines:
    primary key: id
    fields: created_at(), id(), is_public(), name(), project(), updated_at()
  agents:
    primary key: id
    fields: created_at(), id(), name(), status(), version()
  contexts:
    primary key: id
    fields: id(), name(), owner(), type()
  builds:
    primary key: id
    cursor: created
    fields: branch_name(), commit_message(), committer(), created(), finished(), id(), pipeline_name(), progress(), project(), project_id(), provider(), repo_name(), repo_owner(), revision(), status(), trigger(), trigger_type(), triggered_by()
  images:
    primary key: id
    cursor: created
    fields: branch(), commit(), commit_url(), created(), id(), image_display_name(), image_name(), repo(), sha(), size()
  registries:
    primary key: id
    fields: behind_firewall(), default(), domain(), id(), internal(), kind(), name(), primary(), provider()
  triggers:
    primary key: event, pipeline
    fields: event(), event_description(), event_status(), event_type(), filter_tag(), pipeline()
  trigger_events:
    primary key: uri
    fields: account(), description(), endpoint(), kind(), status(), type(), uri()
  annotations:
    primary key: id
    fields: account_id(), entity_id(), entity_type(), id(), key(), type(), value()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_project:
    endpoint: POST /projects
    risk: external mutation; creates a new Codefresh project; approval required
  delete_project:
    endpoint: DELETE /projects/{{ record.id }}
    required fields: id
    risk: destructive; irreversible deletion of a Codefresh project; approval required
  create_pipeline:
    endpoint: POST /pipelines
    risk: external mutation; creates a new Codefresh pipeline; approval required
  update_pipeline:
    endpoint: PUT /pipelines/{{ record.name }}
    required fields: name
    risk: external mutation; replaces an existing Codefresh pipeline's spec; approval required
  delete_pipeline:
    endpoint: DELETE /pipelines/{{ record.name }}
    required fields: name
    risk: destructive; irreversible deletion of a Codefresh pipeline; approval required
  run_pipeline:
    endpoint: POST /pipelines/run/{{ record.name }}
    required fields: name
    risk: external mutation; triggers a real Codefresh pipeline run (build minutes/resources consumed); approval required
  create_context:
    endpoint: POST /contexts
    risk: external mutation; creates a new Codefresh shared context (may hold configuration values); approval required
  delete_context:
    endpoint: DELETE /contexts/{{ record.name }}
    required fields: name
    risk: destructive; irreversible deletion of a Codefresh shared context; approval required
  create_agent:
    endpoint: POST /agents
    risk: external mutation; registers a new Codefresh runner agent; approval required
  delete_agent:
    endpoint: DELETE /agent/{{ record.id }}
    required fields: id
    risk: destructive; irreversible deregistration of a Codefresh runner agent; approval required

SECURITY
  read risk: external Codefresh API read of projects, pipelines, builds, runner agents, shared contexts, container images, registries, triggers, and annotations
  write risk: external mutation of Codefresh projects, pipelines, contexts, and runner agents, including irreversible deletes and triggering real pipeline runs (consumes build minutes/resources)
  approval: required for all write actions; read is unrestricted
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect codefresh

  # Inspect as structured JSON
  pm connectors inspect codefresh --json

AGENT WORKFLOW
  - Run pm connectors inspect codefresh before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
