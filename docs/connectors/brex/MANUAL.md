# pm connectors inspect brex

```text
NAME
  pm connectors inspect brex - Brex connector manual

SYNOPSIS
  pm connectors inspect brex
  pm connectors inspect brex --json
  pm credentials add <name> --connector brex [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Brex transactions, users, expenses, vendors, budgets, cards, accounts, statements, transfers, and webhooks through the Brex platform REST API.

ICON
  id: simple-icons-brex
  asset: icons/simple-icons/brex.svg
  title: Brex
  simple_icon_slug: brex
  simple_icon_hex: 212121
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Brex
  match: exact-name-or-slug
  matched_by: brex

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  mode
  page_size
  start_date
  user_token (secret) (required)

ETL STREAMS
  transactions:
    primary key: id
    cursor: posted_at_date
    fields: amount(object), card_id(string), description(string), id(string), initiated_at_date(string), posted_at_date(string), type(string)
  users:
    primary key: id
    fields: department_id(string), email(string), first_name(string), id(string), last_name(string), manager_id(string), status(string)
  expenses:
    primary key: id
    cursor: purchased_at
    fields: category(string), department_id(string), id(string), location_id(string), memo(string), merchant_id(string), original_amount(object), purchased_at(string), status(string), updated_at(string), user_id(string)
  vendors:
    primary key: id
    fields: company_name(string), email(string), id(string), payment_accounts(array), phone(string)
  budgets:
    primary key: budget_id
    fields: account_id(string), budget_id(string), creator_user_id(string), description(string), limit(object), name(string), parent_budget_id(string), period_type(string), status(string)
  departments:
    primary key: id
    fields: description(string), id(string), name(string)
  locations:
    primary key: id
    fields: description(string), id(string), name(string)
  titles:
    primary key: id
    fields: id(string), name(string)
  legal_entities:
    primary key: id
    fields: billingAddress(object), createdAt(string), displayName(string), id(string), isDefault(boolean), status(string)
  cards:
    primary key: id
    fields: billing_address(object), budget_id(string), card_name(string), card_type(string), expiration_date(object), has_been_transferred(boolean), id(string), last_four(string), limit_type(string), mailing_address(object), owner(object), spend_controls(object), status(string)
  accounts_card:
    primary key: id
    fields: account_limit(object), available_balance(object), current_balance(object), current_statement_period(object), id(string), status(string)
  accounts_cash:
    primary key: id
    fields: account_number(string), available_balance(object), current_balance(object), id(string), name(string), primary(boolean), routing_number(string), status(string)
  card_statements:
    primary key: id
    fields: end_balance(object), id(string), period(object), start_balance(object)
  linked_accounts:
    primary key: id
    fields: available_balance(object), bank_details(object), brex_account_id(string), current_balance(object), id(string), last_four(string)
  transfers:
    primary key: id
    fields: amount(object), cancellation_reason(string), counterparty(object), creator_user_id(string), description(string), estimated_delivery_date(string), id(string), originating_account(object), payment_type(string), process_date(string), status(string)
  webhooks:
    primary key: id
    fields: event_types(array), group_id(string), id(string), status(string), url(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  update_vendor:
    endpoint: PUT /v1/vendors/{{ record.id }}
    required fields: id
    risk: mutates an existing vendor's name, contact details, or payment accounts; affects future transfer counterparty resolution
  delete_vendor:
    endpoint: DELETE /v1/vendors/{{ record.id }}
    required fields: id
    risk: permanently removes a vendor record; any transfer still referencing it as counterparty will fail to resolve
  create_department:
    endpoint: POST /v2/departments
    required fields: name
    risk: creates a new organizational department; low-risk external mutation, no approval required
  create_location:
    endpoint: POST /v2/locations
    required fields: name
    risk: creates a new organizational location; low-risk external mutation, no approval required
  create_title:
    endpoint: POST /v2/titles
    required fields: name
    risk: creates a new job title; low-risk external mutation, no approval required
  create_user:
    endpoint: POST /v2/users
    required fields: email, first_name, last_name
    risk: invites a new user to the Brex account; sends a real invitation email to the target address
  update_user:
    endpoint: PUT /v2/users/{{ record.id }}
    required fields: id
    risk: mutates an existing user's status, manager, department, location, or title; setting status to a terminated/suspended state revokes account access
  update_card:
    endpoint: PUT /v2/cards/{{ record.id }}
    required fields: id
    risk: mutates an existing card's spend controls (limit amount/category/merchant restrictions); takes effect on the physical/virtual card immediately
  lock_card:
    endpoint: POST /v2/cards/{{ record.id }}/lock
    required fields: id
    risk: immediately blocks all new transactions on the card until unlocked; does not affect already-authorized/pending transactions
  unlock_card:
    endpoint: POST /v2/cards/{{ record.id }}/unlock
    required fields: id
    risk: immediately re-enables new transactions on a previously locked card
  terminate_card:
    endpoint: POST /v2/cards/{{ record.id }}/terminate
    required fields: id
    risk: permanently deactivates a card; irreversible, the card can never be unlocked or reused after termination
  update_expense:
    endpoint: PUT /v1/expenses/card/{{ record.expense_id }}
    required fields: expense_id
    risk: mutates an existing card expense's memo; low-risk metadata-only external mutation
  update_webhook:
    endpoint: PUT /v1/webhooks/{{ record.id }}
    required fields: id, url, event_types, status
    risk: re-points an already-registered webhook's delivery URL, event set, or active status; redirects live event delivery immediately
  delete_webhook:
    endpoint: DELETE /v1/webhooks/{{ record.id }}
    required fields: id
    risk: permanently removes a webhook subscription; irreversible

SECURITY
  read risk: external Brex API read of card transaction, user, expense, vendor, card, account, and transfer data
  write risk: external mutation of vendors, org directory (departments/locations/titles/users), card controls/lifecycle, expenses, and webhooks; card lock/unlock/terminate take effect on real payment instruments immediately
  approval: required for all write actions; each action's per-record risk string in writes.json is the authoritative summary
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect brex

  # Inspect as structured JSON
  pm connectors inspect brex --json

AGENT WORKFLOW
  - Run pm connectors inspect brex before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
