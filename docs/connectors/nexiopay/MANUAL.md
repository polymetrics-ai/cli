# pm connectors inspect nexiopay

```text
NAME
  pm connectors inspect nexiopay - Nexio Pay connector manual

SYNOPSIS
  pm connectors inspect nexiopay
  pm connectors inspect nexiopay --json
  pm credentials add <name> --connector nexiopay [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Nexio Pay card tokens, payout recipients, spendbacks, payment types, terminals, and the API user via the Nexio REST API.

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
  username (secret)

ETL STREAMS
  card_tokens:
    primary key: key
    fields: cardHolderName(), cardType(), createdDate(), currency(), expirationMonth(), expirationYear(), key(), lastFour()
  recipients:
    primary key: recipientId
    fields: createdDate(), currency(), email(), name(), recipientId(), status(), updatedDate()
  spendbacks:
    primary key: id
    fields: amount(), createdDate(), currency(), id(), recipientId(), status()
  payment_types:
    primary key: id
    fields: displayName(), enabled(), id(), name()
  terminal_list:
    primary key: terminalId
    fields: merchantId(), name(), status(), terminalId()
  user:
    primary key: accountId
    fields: accountId(), email(), merchantId(), role(), username()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Nexio Pay API read of card tokens, payout, and account data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect nexiopay

  # Inspect as structured JSON
  pm connectors inspect nexiopay --json

AGENT WORKFLOW
  - Run pm connectors inspect nexiopay before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
