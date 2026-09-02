# pm connectors inspect faker

```text
NAME
  pm connectors inspect faker - Sample Data connector manual

SYNOPSIS
  pm connectors inspect faker
  pm connectors inspect faker --json
  pm credentials add <name> --connector faker [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Generates deterministic sample users, purchases, and products without network access.

ICON
  id: simple-icons-faker
  asset: icons/simple-icons/faker.svg
  title: Faker
  simple_icon_slug: faker
  simple_icon_hex: 779B2E
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Faker
  match: exact-name-or-slug
  matched_by: faker

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  count
  seed

SECURITY
  read risk: none; in-process synthetic data generation, no network access
  write risk: n/a (read-only source)
  approval: none required; no external data is read or written
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect faker

  # Inspect as structured JSON
  pm connectors inspect faker --json

AGENT WORKFLOW
  - Run pm connectors inspect faker before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
