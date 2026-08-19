# pm connectors inspect easypromos

```text
NAME
  pm connectors inspect easypromos - Easypromos connector manual

SYNOPSIS
  pm connectors inspect easypromos
  pm connectors inspect easypromos --json
  pm credentials add <name> --connector easypromos [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Easypromos promotions, organizing brands, stages, users, participations, and prizes through the Easypromos REST API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  promotion_id
  bearer_token (secret) (required)

ETL STREAMS
  promotions:
    primary key: id
    fields: created(string), default_language(string), description(string), end_date(string), id(string), organizing_brand_id(string), organizing_brand_name(string), promotion_type(string), start_date(string), status(string), timezone(string), title(string), url(string)
  organizing_brands:
    primary key: id
    fields: id(string), name(string)
  stages:
    primary key: id
    fields: end_date(string), id(string), name(string), start_date(string), type(string), visible(boolean)
  users:
    primary key: id
    fields: country(string), created(string), email(string), external_id(string), first_name(string), id(string), language(string), last_name(string), login_type(string), nickname(string), promotion_id(string), status(string)
  participations:
    primary key: id
    fields: created(string), id(string), ip(string), promotion_id(string), stage_id(string), user_agent(string), user_id(string)
  prizes:
    primary key: id
    fields: code(string), created(string), download_url(string), id(string), participation_id(string), prize_type_id(string), prize_type_name(string), redeem_url(string), stage_id(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Easypromos API read of promotion, user, participation, and prize data
  approval: none; read-only, no reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect easypromos

  # Inspect as structured JSON
  pm connectors inspect easypromos --json

AGENT WORKFLOW
  - Run pm connectors inspect easypromos before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
