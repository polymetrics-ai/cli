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
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  bearer_token (secret) (required)

ETL STREAMS
  parties:
    primary key: id
    cursor: updated_at
    fields: about(string), created_at(string), first_name(string), id(integer), job_title(string), last_contacted_at(string), last_name(string), organisation_name(string), owner(string), title(string), type(string), updated_at(string)
  opportunities:
    primary key: id
    cursor: updated_at
    fields: closed_on(string), created_at(string), description(string), expected_close_on(string), id(integer), lost_reason(string), milestone_id(integer), milestone_name(string), name(string), party_id(integer), probability(number), updated_at(string), value_amount(number), value_currency(string)
  kases:
    primary key: id
    cursor: updated_at
    fields: closed_on(string), created_at(string), description(string), id(integer), name(string), owner(string), party_id(integer), status(string), updated_at(string)
  tasks:
    primary key: id
    cursor: updated_at
    fields: category_id(integer), created_at(string), description(string), detail(string), due_on(string), id(integer), kase_id(integer), opportunity_id(integer), party_id(integer), status(string), updated_at(string)
  users:
    primary key: id
    cursor: updated_at
    fields: created_at(string), email(string), id(integer), name(string), status(string), updated_at(string), username(string)
  tags:
    primary key: id
    fields: color(string), id(integer), name(string)
  custom_fields:
    primary key: id
    fields: entity_type(string), id(integer), name(string), restricted_to_type(string), tag(string), type(string)
  teams:
    primary key: id
    fields: id(integer), name(string)
  pipelines:
    primary key: id
    fields: created_at(string), default(boolean), display_order(integer), id(integer), name(string), updated_at(string)
  milestones:
    primary key: id
    fields: id(integer), name(string), pipeline_id(integer), probability(number)
  lost_reasons:
    primary key: id
    fields: id(integer), name(string)
  categories:
    primary key: id
    fields: color(string), id(integer), name(string)
  boards:
    primary key: id
    fields: created_at(string), entity_type(string), id(integer), name(string), updated_at(string)
  stages:
    primary key: id
    fields: board_id(integer), display_order(integer), id(integer), name(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_party:
    endpoint: POST /parties
    required fields: party
    risk: external mutation; creates a live Capsule CRM contact; approval required. Body wraps the record under a top-level "party" key (Capsule's resource-envelope convention) — the record itself must carry that wrapper, since the engine's write dialect sends record fields verbatim as the JSON body with no nested-wrapper construction primitive.
  update_party:
    endpoint: PUT /parties/{{ record.id }}
    required fields: id, party
    risk: external mutation; updates a live Capsule CRM contact; approval required. Body wraps the record under a top-level "party" key; "id" is path-only (path_fields) and excluded from the body via body_fields.
  delete_party:
    endpoint: DELETE /parties/{{ record.id }}
    required fields: id
    risk: external mutation; irreversibly deletes a live Capsule CRM contact and its associated history; approval required
  create_opportunity:
    endpoint: POST /opportunities
    required fields: opportunity
    risk: external mutation; creates a live Capsule CRM sales opportunity; approval required. Body wraps the record under a top-level "opportunity" key.
  update_opportunity:
    endpoint: PUT /opportunities/{{ record.id }}
    required fields: id, opportunity
    risk: external mutation; updates a live Capsule CRM sales opportunity (including moving pipeline stage or closing/losing it); approval required
  delete_opportunity:
    endpoint: DELETE /opportunities/{{ record.id }}
    required fields: id
    risk: external mutation; irreversibly deletes a live Capsule CRM sales opportunity; approval required
  create_kase:
    endpoint: POST /kases
    required fields: kase
    risk: external mutation; creates a live Capsule CRM case/project; approval required. Body wraps the record under a top-level "kase" key (Capsule kept the "kase" spelling in the API after renaming Cases to Projects in the product UI, to avoid a breaking change; see docs.md).
  update_kase:
    endpoint: PUT /kases/{{ record.id }}
    required fields: id, kase
    risk: external mutation; updates a live Capsule CRM case/project, including closing it; approval required
  delete_kase:
    endpoint: DELETE /kases/{{ record.id }}
    required fields: id
    risk: external mutation; irreversibly deletes a live Capsule CRM case/project; approval required
  create_task:
    endpoint: POST /tasks
    required fields: task
    risk: external mutation; creates a live Capsule CRM task/reminder; approval required. Body wraps the record under a top-level "task" key.
  update_task:
    endpoint: PUT /tasks/{{ record.id }}
    required fields: id, task
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
