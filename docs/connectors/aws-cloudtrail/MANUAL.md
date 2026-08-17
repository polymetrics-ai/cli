# pm connectors inspect aws-cloudtrail

```text
NAME
  pm connectors inspect aws-cloudtrail - AWS CloudTrail connector manual

SYNOPSIS
  pm connectors inspect aws-cloudtrail
  pm connectors inspect aws-cloudtrail --json
  pm credentials add <name> --connector aws-cloudtrail [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads AWS CloudTrail management events (last 90 days) via the LookupEvents API using AWS Signature V4 authentication. Read-only. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

ICON
  asset: icons/aws-cloudtrail.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  aws_region_name
  base_url
  lookup_attributes_filter
  mode
  start_date
  aws_key_id (secret)
  aws_secret_key (secret)

ETL STREAMS
  management_events:
    primary key: EventId
    cursor: EventTime
    fields: AccessKeyId(), CloudTrailEvent(), EventId(), EventName(), EventSource(), EventTime(), ReadOnly(), Resources(), Username()
  read_only_events:
    primary key: EventId
    cursor: EventTime
    fields: AccessKeyId(), CloudTrailEvent(), EventId(), EventName(), EventSource(), EventTime(), ReadOnly(), Resources(), Username()
  write_only_events:
    primary key: EventId
    cursor: EventTime
    fields: AccessKeyId(), CloudTrailEvent(), EventId(), EventName(), EventSource(), EventTime(), ReadOnly(), Resources(), Username()
  console_logins:
    primary key: EventId
    cursor: EventTime
    fields: AccessKeyId(), CloudTrailEvent(), EventId(), EventName(), EventSource(), EventTime(), ReadOnly(), Resources(), Username()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external AWS CloudTrail API reads performed by the legacy connector via a Tier-2 hook
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect aws-cloudtrail

  # Inspect as structured JSON
  pm connectors inspect aws-cloudtrail --json

AGENT WORKFLOW
  - Run pm connectors inspect aws-cloudtrail before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
