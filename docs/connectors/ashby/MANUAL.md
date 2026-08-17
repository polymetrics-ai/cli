# pm connectors inspect ashby

```text
NAME
  pm connectors inspect ashby - Ashby connector manual

SYNOPSIS
  pm connectors inspect ashby
  pm connectors inspect ashby --json
  pm credentials add <name> --connector ashby [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Ashby applicant-tracking data — candidates, jobs, applications, and users — through the Ashby REST API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

ICON
  asset: icons/ashby.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.ashbyhq.com/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  start_date
  api_key (secret)

ETL STREAMS
  candidates:
    primary key: id
    cursor: updatedAt
    fields: company(), createdAt(), id(), locationSummary(), name(), primaryEmailAddress(), primaryPhoneNumber(), timezone(), title(), updatedAt()
  jobs:
    primary key: id
    cursor: updatedAt
    fields: createdAt(), customFields(), defaultInterviewPlanId(), departmentId(), employmentType(), id(), locationId(), status(), title(), updatedAt()
  applications:
    primary key: id
    cursor: updatedAt
    fields: archiveReason(), candidateId(), createdAt(), currentInterviewStageId(), id(), jobId(), source(), status(), updatedAt()
  users:
    primary key: id
    cursor: updatedAt
    fields: email(), firstName(), globalRole(), id(), isEnabled(), lastName(), updatedAt()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Ashby API reads performed by the legacy connector via a Tier-2 hook
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect ashby

  # Inspect as structured JSON
  pm connectors inspect ashby --json

AGENT WORKFLOW
  - Run pm connectors inspect ashby before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
