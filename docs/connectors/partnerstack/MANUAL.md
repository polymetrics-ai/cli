# pm connectors inspect partnerstack

```text
NAME
  pm connectors inspect partnerstack - PartnerStack connector manual

SYNOPSIS
  pm connectors inspect partnerstack
  pm connectors inspect partnerstack --json
  pm credentials add <name> --connector partnerstack [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads PartnerStack partnerships, customers, transactions, and groups through the REST API.

ICON
  asset: icons/partnerstack.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.partnerstack.com/docs/api-overview

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  limit
  max_pages
  mode
  api_key (secret)

ETL STREAMS
  partnerships:
    primary key: id
    cursor: created_at
    fields: created_at(), email(), id(), status()
  customers:
    primary key: id
    cursor: created_at
    fields: created_at(), email(), id(), name()
  transactions:
    primary key: id
    cursor: created_at
    fields: amount(), created_at(), currency(), customer_id(), id()
  groups:
    primary key: id
    cursor: created_at
    fields: created_at(), id(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external PartnerStack API read of partnership and referral-customer data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect partnerstack

  # Inspect as structured JSON
  pm connectors inspect partnerstack --json

AGENT WORKFLOW
  - Run pm connectors inspect partnerstack before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
