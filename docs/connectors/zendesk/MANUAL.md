# pm connectors inspect zendesk

```text
NAME
  pm connectors inspect zendesk - Zendesk connector manual

SYNOPSIS
  pm connectors inspect zendesk
  pm connectors inspect zendesk --json
  pm credentials add <name> --connector zendesk [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Metadata inventory for the official Zendesk Support API surface; executable streams, bounded direct reads, and reverse-ETL writes are mapped by the Zendesk CLI parity sub-issues.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=false write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  access_token (secret)
  api_token (secret)
  email (secret)

SECURITY
  read risk: metadata-only bundle in issue #157; later lanes enable safe ETL streams and bounded direct reads from the official Zendesk OAS
  write risk: metadata-only bundle in issue #157; Zendesk mutations remain blocked until typed reverse-ETL schemas, approval text, redaction, and destructive confirmation are added
  approval: all external writes remain plan → preview → approval → execute and blocked in this slice
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Work with Zendesk Support API metadata from the command line.
  Usage: pm zendesk <command> [flags]
  Source CLI: Zendesk API reference (https://developer.zendesk.com/zendesk/oas.yaml)
  Global flags:
    --json (boolean): Write machine-readable JSON output.
    --connection (string): Use a saved Zendesk connector credential/account scope.: maps_to=connection
  Read Candidates
  Write Candidates
  Blocked Metadata
  Other Commands
    read candidates - Review Zendesk read candidates from the official OAS [intent=direct_read availability=planned]; notes: Metadata inventory only. api_surface.json tracks 282 structured direct-read candidates and later lanes map durable collections to ETL streams or bounded direct reads.
    binary candidates - Review Zendesk binary/file read candidates from the official OAS [intent=direct_read availability=planned]; notes: Metadata inventory only. api_surface.json tracks 37 binary/file read candidates; #162 must define bounded size/path/download policy before execution.
    write candidates - Review Zendesk sensitive/admin write candidates from the official OAS [intent=reverse_etl availability=planned]; approval: reverse ETL execution remains plan → preview → approval → execute.; risk: Zendesk mutation candidates remain blocked until typed reverse-ETL schemas, risk text, approval text, redaction, and policy gates are added.; notes: Metadata inventory only. api_surface.json tracks 210 sensitive/admin reverse-ETL candidates.
    destructive candidates - Review Zendesk destructive write candidates from the official OAS [intent=reverse_etl availability=planned]; approval: reverse ETL execution remains plan → preview → approval → execute with typed confirmation for destructive actions.; risk: Zendesk DELETE operations are destructive and remain blocked until typed reverse-ETL schemas, risk text, approval text, redaction, and destructive confirmation are added.; notes: Metadata inventory only. api_surface.json tracks 85 destructive-action candidates.
    deprecated operations - Review deprecated Zendesk operations from the official OAS [intent=docs_only availability=unsafe_or_disallowed]; notes: Metadata inventory only. api_surface.json tracks 3 deprecated operation rows; #160 must confirm replacements or blockers.
  Help topics:
    zendesk-auth - Zendesk authentication uses OAuth bearer tokens or email/API-token Basic auth with secrets stored in credentials.
    zendesk-safety - Zendesk writes remain reverse-ETL gated: plan, preview, approval, execute, with destructive confirmation where required.
    zendesk-operation-ledger - The initial operation ledger blocks all official OAS operations until later lanes map exact executable surfaces.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect zendesk

  # Inspect as structured JSON
  pm connectors inspect zendesk --json

AGENT WORKFLOW
  - Run pm connectors inspect zendesk before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
