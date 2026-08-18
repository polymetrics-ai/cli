# pm connectors inspect trello

```text
NAME
  pm connectors inspect trello - Trello connector manual

SYNOPSIS
  pm connectors inspect trello
  pm connectors inspect trello --json
  pm credentials add <name> --connector trello [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Trello boards, lists, and checklists through the Trello REST API. Cards and actions are blocked (see docs.md Known limits).

ICON
  id: trello
  asset: icons/trello.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.atlassian.com/cloud/trello/rest/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  board_ids
  key (secret) (required)
  token (secret) (required)

ETL STREAMS
  boards:
    primary key: id
    fields: closed(boolean), dateLastActivity(string), desc(string), id(string), idOrganization(string), name(string), shortUrl(string), url(string)
  lists:
    primary key: id
    fields: closed(boolean), id(string), idBoard(string), name(string), pos(number), subscribed(boolean)
  checklists:
    primary key: id
    fields: id(string), idBoard(string), idCard(string), name(string), pos(number)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Trello API read of board/list/checklist data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect trello

  # Inspect as structured JSON
  pm connectors inspect trello --json

AGENT WORKFLOW
  - Run pm connectors inspect trello before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
