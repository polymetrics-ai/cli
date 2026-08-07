# pm connectors inspect typeform

```text
NAME
  pm connectors inspect typeform - Typeform connector manual

SYNOPSIS
  pm connectors inspect typeform
  pm connectors inspect typeform --json
  pm credentials add <name> --connector typeform [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Typeform forms, workspaces, themes, and images through the Typeform REST API.

ICON
  id: typeform
  asset: icons/typeform.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://www.typeform.com/developers/changelog/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  form_ids
  mode
  page_size
  access_token (secret)

ETL STREAMS
  forms:
    primary key: id
    cursor: last_updated_at
    fields: created_at(string), id(string), is_public(boolean), last_updated_at(string), self_href(string), theme_href(string), title(string), type(string)
  responses:
    primary key: response_id
    cursor: submitted_at
    fields: answers(array), calculated(object), form_id(string), hidden(object), landed_at(string), landing_id(string), metadata(object), response_id(string), submitted_at(string), token(string)
  workspaces:
    primary key: id
    fields: account_id(string), default(boolean), id(string), name(string), self_href(string), shared(boolean)
  themes:
    primary key: id
    fields: background(object), colors(object), font(string), id(string), name(string), visibility(string)
  images:
    primary key: id
    fields: file_name(string), has_alpha(boolean), height(integer), id(string), media_type(string), src(string), width(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Typeform API read of form, workspace, theme, and image metadata
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect typeform

  # Inspect as structured JSON
  pm connectors inspect typeform --json

AGENT WORKFLOW
  - Run pm connectors inspect typeform before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
