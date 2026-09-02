# pm connectors inspect tally-prime

```text
NAME
  pm connectors inspect tally-prime - TallyPrime connector manual

SYNOPSIS
  pm connectors inspect tally-prime
  pm connectors inspect tally-prime --json
  pm credentials add <name> --connector tally-prime [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads TallyPrime accounting data (companies, ledgers, groups, stock items, vouchers) via TDL Export/Collection envelope requests POSTed to a locally-running TallyPrime Gateway Server. Read-only source; schema is discovered dynamically since TallyPrime has no static REST resource surface.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: accounting

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  company (required)
  envelope_format
  from_date
  gateway_url (required)
  http_timeout_seconds
  mode
  to_date

SECURITY
  read risk: local TallyPrime Gateway Server read of accounting masters (companies/ledgers/groups/stock items) and transactional vouchers via TDL Export/Collection envelopes over HTTP POST to a locally-running TallyPrime instance (no public network egress)
  write risk: n/a (read-only source)
  approval: none required for read-only sync
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect tally-prime

  # Inspect as structured JSON
  pm connectors inspect tally-prime --json

AGENT WORKFLOW
  - Run pm connectors inspect tally-prime before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
