# pm connectors inspect capsule-crm

```text
NAME
  pm connectors inspect capsule-crm - Capsule CRM connector manual

SYNOPSIS
  pm connectors inspect capsule-crm
  pm connectors inspect capsule-crm --json
  pm credentials add <name> --connector capsule-crm [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Capsule CRM parties, opportunities, cases, tasks, users, tags, custom field definitions, teams, pipelines, milestones, lost reasons, task categories, boards, and stages, and writes party/opportunity/case/task create, update, and delete actions, through the Capsule v2 REST API.

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
  base_url
  mode
  bearer_token (secret)

ETL STREAMS
  parties:
    primary key: id
    cursor: updated_at
    fields: about(), created_at(), first_name(), id(), job_title(), last_contacted_at(), last_name(), organisation_name(), owner(), title(), type(), updated_at()
  opportunities:
    primary key: id
    cursor: updated_at
    fields: closed_on(), created_at(), description(), expected_close_on(), id(), lost_reason(), milestone_id(), milestone_name(), name(), party_id(), probability(), updated_at(), value_amount(), value_currency()
  kases:
    primary key: id
    cursor: updated_at
    fields: closed_on(), created_at(), description(), id(), name(), owner(), party_id(), status(), updated_at()
  tasks:
    primary key: id
    cursor: updated_at
    fields: category_id(), created_at(), description(), detail(), due_on(), id(), kase_id(), opportunity_id(), party_id(), status(), updated_at()
  users:
    primary key: id
    cursor: updated_at
    fields: created_at(), email(), id(), name(), status(), updated_at(), username()
  tags:
    primary key: id
    fields: color(), id(), name()
  custom_fields:
    primary key: id
    fields: entity_type(), id(), name(), restricted_to_type(), tag(), type()
  teams:
    primary key: id
    fields: id(), name()
  pipelines:
    primary key: id
    fields: created_at(), default(), display_order(), id(), name(), updated_at()
  milestones:
    primary key: id
    fields: id(), name(), pipeline_id(), probability()
  lost_reasons:
    primary key: id
    fields: id(), name()
  categories:
    primary key: id
    fields: color(), id(), name()
  boards:
    primary key: id
    fields: created_at(), entity_type(), id(), name(), updated_at()
  stages:
    primary key: id
    fields: board_id(), display_order(), id(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_party:
    endpoint: POST /parties
    risk: external mutation; creates a live Capsule CRM contact; approval required. Body wraps the record under a top-level "party" key (Capsule's resource-envelope convention) — the record itself must carry that wrapper, since the engine's write dialect sends record fields verbatim as the JSON body with no nested-wrapper construction primitive.
  update_party:
    endpoint: PUT /parties/{{ record.id }}
    required fields: id
    optional fields: party
    risk: external mutation; updates a live Capsule CRM contact; approval required. Body wraps the record under a top-level "party" key; "id" is path-only (path_fields) and excluded from the body via body_fields.
  delete_party:
    endpoint: DELETE /parties/{{ record.id }}
    required fields: id
    risk: external mutation; irreversibly deletes a live Capsule CRM contact and its associated history; approval required
  create_opportunity:
    endpoint: POST /opportunities
    risk: external mutation; creates a live Capsule CRM sales opportunity; approval required. Body wraps the record under a top-level "opportunity" key.
  update_opportunity:
    endpoint: PUT /opportunities/{{ record.id }}
    required fields: id
    optional fields: opportunity
    risk: external mutation; updates a live Capsule CRM sales opportunity (including moving pipeline stage or closing/losing it); approval required
  delete_opportunity:
    endpoint: DELETE /opportunities/{{ record.id }}
    required fields: id
    risk: external mutation; irreversibly deletes a live Capsule CRM sales opportunity; approval required
  create_kase:
    endpoint: POST /kases
    risk: external mutation; creates a live Capsule CRM case/project; approval required. Body wraps the record under a top-level "kase" key (Capsule kept the "kase" spelling in the API after renaming Cases to Projects in the product UI, to avoid a breaking change; see docs.md).
  update_kase:
    endpoint: PUT /kases/{{ record.id }}
    required fields: id
    optional fields: kase
    risk: external mutation; updates a live Capsule CRM case/project, including closing it; approval required
  delete_kase:
    endpoint: DELETE /kases/{{ record.id }}
    required fields: id
    risk: external mutation; irreversibly deletes a live Capsule CRM case/project; approval required
  create_task:
    endpoint: POST /tasks
    risk: external mutation; creates a live Capsule CRM task/reminder; approval required. Body wraps the record under a top-level "task" key.
  update_task:
    endpoint: PUT /tasks/{{ record.id }}
    required fields: id
    optional fields: task
    risk: external mutation; updates a live Capsule CRM task, including marking it complete; approval required
  delete_task:
    endpoint: DELETE /tasks/{{ record.id }}
    required fields: id
    risk: external mutation; irreversibly deletes a live Capsule CRM task; approval required

SECURITY
  read risk: external Capsule CRM API read of CRM records and account configuration (tags, custom fields, pipelines)
  write risk: external mutation of live Capsule CRM parties, opportunities, cases, and tasks including irreversible deletes; approval required for every write action
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect capsule-crm

  # Inspect as structured JSON
  pm connectors inspect capsule-crm --json

AGENT WORKFLOW
  - Run pm connectors inspect capsule-crm before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
