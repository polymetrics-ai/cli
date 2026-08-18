# pm connectors inspect eventzilla

```text
NAME
  pm connectors inspect eventzilla - Eventzilla connector manual

SYNOPSIS
  pm connectors inspect eventzilla
  pm connectors inspect eventzilla --json
  pm credentials add <name> --connector eventzilla [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Eventzilla events, categories, users, attendees, ticket types, and transactions, and writes attendee check-in and event sales-page toggle mutations, through the Eventzilla v2 REST API.

ICON
  id: eventzilla
  asset: icons/eventzilla.svg
  source: official
  review_status: official_verified
  review_url: https://www.eventzilla.net/api/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  api_key (secret) (required)

ETL STREAMS
  events:
    primary key: id
    fields: categories(string), currency(string), end_date(string), end_time(string), id(integer), start_date(string), start_time(string), status(string), tickets_sold(integer), tickets_total(integer), time_zone(string), title(string), url(string), venue(string)
  categories:
    primary key: category
    fields: category(string)
  users:
    primary key: id
    fields: company(string), email(string), first_name(string), id(integer), last_name(string), last_seen(string), phone_primary(string), timezone(string), user_type(string), username(string)
  attendees:
    primary key: id
    fields: email(string), event_id(integer), first_name(string), id(integer), is_attended(string), last_name(string), refno(string), ticket_type(string), transaction_amount(number), transaction_date(string), transaction_status(string)
  tickets:
    primary key: id
    fields: event_id(integer), id(integer), is_visible(boolean), price(number), quantity_total(integer), sales_end_date(string), sales_start_date(string), ticket_type(string), title(string)
  transactions:
    primary key: checkout_id
    fields: buyer_first_name(string), buyer_last_name(string), checkout_id(integer), comments(string), email(string), event_date(string), event_id(integer), payment_type(string), promo_code(string), tickets_in_transaction(string), title(string), transaction_amount(string), transaction_date(string), transaction_ref(string), transaction_status(string), user_id(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  checkin_attendee:
    endpoint: POST /attendees/checkin
    required fields: barcode, eventcheckin
    risk: marks an attendee checked in or reverts check-in at the door; low-risk operational mutation, no approval required
  toggle_event_sales:
    endpoint: POST /events/togglesales
    required fields: eventid, status
    risk: publishes or unpublishes an event's public sales page; setting status false immediately stops new ticket sales for that event, approval required

SECURITY
  read risk: external Eventzilla API read of event, category, user, attendee, ticket, and transaction data
  write risk: external mutation of attendee check-in state and event sales-page publish status; every write ships with an explicit per-action risk string
  approval: required for toggle_event_sales (unpublishing stops new ticket sales immediately); checkin_attendee is a low-risk operational door-scan mutation, no approval required
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect eventzilla

  # Inspect as structured JSON
  pm connectors inspect eventzilla --json

AGENT WORKFLOW
  - Run pm connectors inspect eventzilla before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
