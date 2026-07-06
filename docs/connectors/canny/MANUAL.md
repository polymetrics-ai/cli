# pm connectors inspect canny

```text
NAME
  pm connectors inspect canny - Canny connector manual

SYNOPSIS
  pm connectors inspect canny
  pm connectors inspect canny --json
  pm credentials add <name> --connector canny [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Canny boards, posts, comments, categories, and companies through the Canny REST API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  api_key (secret)

ETL STREAMS
  boards:
    primary key: id
    cursor: created
    fields: created(), id(), isPrivate(), name(), postCount(), url()
  posts:
    primary key: id
    cursor: created
    fields: commentCount(), created(), details(), eta(), id(), score(), status(), statusChangedAt(), title(), url()
  comments:
    primary key: id
    cursor: created
    fields: created(), id(), internal(), likeCount(), parentID(), private(), value()
  categories:
    primary key: id
    cursor: created
    fields: created(), id(), name(), parentID(), postCount(), url()
  companies:
    primary key: id
    cursor: created
    fields: created(), domain(), id(), memberCount(), monthlySpend(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Canny API reads performed by the legacy connector via a Tier-2 hook
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect canny

  # Inspect as structured JSON
  pm connectors inspect canny --json

AGENT WORKFLOW
  - Run pm connectors inspect canny before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
