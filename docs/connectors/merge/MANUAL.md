# pm connectors inspect merge

```text
NAME
  pm connectors inspect merge - Merge connector manual

SYNOPSIS
  pm connectors inspect merge
  pm connectors inspect merge --json
  pm credentials add <name> --connector merge [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Merge ATS common-model objects (candidates, applications, jobs, offers, departments, users) through the Merge unified REST API.

ICON
  id: merge
  asset: icons/merge.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.merge.dev/api-reference/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  page_size
  start_date
  account_token (secret) (required)
  api_token (secret) (required)

ETL STREAMS
  candidates:
    primary key: id
    cursor: modified_at
    fields: can_email(boolean), company(string), first_name(string), id(string), is_private(boolean), last_interaction_at(string), last_name(string), modified_at(string), remote_created_at(string), remote_id(string), remote_updated_at(string), remote_was_deleted(boolean), title(string)
  applications:
    primary key: id
    cursor: modified_at
    fields: applied_at(string), candidate(string), credited_to(string), current_stage(string), id(string), job(string), modified_at(string), reject_reason(string), rejected_at(string), remote_id(string), remote_was_deleted(boolean), source(string)
  jobs:
    primary key: id
    cursor: modified_at
    fields: code(string), confidential(boolean), description(string), id(string), modified_at(string), name(string), remote_created_at(string), remote_id(string), remote_updated_at(string), remote_was_deleted(boolean), status(string), type(string)
  offers:
    primary key: id
    cursor: modified_at
    fields: application(string), closed_at(string), creator(string), id(string), modified_at(string), remote_created_at(string), remote_id(string), remote_was_deleted(boolean), sent_at(string), start_date(string), status(string)
  departments:
    primary key: id
    cursor: modified_at
    fields: id(string), modified_at(string), name(string), remote_id(string), remote_was_deleted(boolean)
  users:
    primary key: id
    cursor: modified_at
    fields: access_role(string), disabled(boolean), email(string), first_name(string), id(string), last_name(string), modified_at(string), remote_created_at(string), remote_id(string), remote_was_deleted(boolean)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Merge unified API read of ATS candidate and hiring data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect merge

  # Inspect as structured JSON
  pm connectors inspect merge --json

AGENT WORKFLOW
  - Run pm connectors inspect merge before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
