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
  api_key (secret) (required)
  username (secret) (required)

ETL STREAMS
  card_tokens:
    primary key: key
    fields: cardHolderName(string), cardType(string), createdDate(string), currency(string), expirationMonth(string), expirationYear(string), key(string), lastFour(string)
  recipients:
    primary key: recipientId
    fields: createdDate(string), currency(string), email(string), name(string), recipientId(string), status(string), updatedDate(string)
  spendbacks:
    primary key: id
    fields: amount(number), createdDate(string), currency(string), id(string), recipientId(string), status(string)
  payment_types:
    primary key: id
    fields: displayName(string), enabled(boolean), id(string), name(string)
  terminal_list:
    primary key: terminalId
    fields: merchantId(string), name(string), status(string), terminalId(string)
  user:
    primary key: accountId
    fields: accountId(string), email(string), merchantId(string), role(string), username(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

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
