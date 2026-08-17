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
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

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
  user_token (secret)

ETL STREAMS
  transactions:
    primary key: id
    cursor: posted_at_date
    fields: amount(), card_id(), description(), id(), initiated_at_date(), posted_at_date(), type()
  users:
    primary key: id
    fields: department_id(), email(), first_name(), id(), last_name(), manager_id(), status()
  expenses:
    primary key: id
    cursor: purchased_at
    fields: category(), department_id(), id(), location_id(), memo(), merchant_id(), original_amount(), purchased_at(), status(), updated_at(), user_id()
  vendors:
    primary key: id
    fields: company_name(), email(), id(), payment_accounts(), phone()
  budgets:
    primary key: budget_id
    fields: account_id(), budget_id(), creator_user_id(), description(), limit(), name(), parent_budget_id(), period_type(), status()
  departments:
    primary key: id
    fields: description(), id(), name()
  locations:
    primary key: id
    fields: description(), id(), name()
  titles:
    primary key: id
    fields: id(), name()
  legal_entities:
    primary key: id
    fields: billingAddress(), createdAt(), displayName(), id(), isDefault(), status()
  cards:
    primary key: id
    fields: billing_address(), budget_id(), card_name(), card_type(), expiration_date(), has_been_transferred(), id(), last_four(), limit_type(), mailing_address(), owner(), spend_controls(), status()
  accounts_card:
    primary key: id
    fields: account_limit(), available_balance(), current_balance(), current_statement_period(), id(), status()
  accounts_cash:
    primary key: id
    fields: account_number(), available_balance(), current_balance(), id(), name(), primary(), routing_number(), status()
  card_statements:
    primary key: id
    fields: end_balance(), id(), period(), start_balance()
  linked_accounts:
    primary key: id
    fields: available_balance(), bank_details(), brex_account_id(), current_balance(), id(), last_four()
  transfers:
    primary key: id
    fields: amount(), cancellation_reason(), counterparty(), creator_user_id(), description(), estimated_delivery_date(), id(), originating_account(), payment_type(), process_date(), status()
  webhooks:
    primary key: id
    fields: event_types(), group_id(), id(), status(), url()

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
    risk: creates a new organizational department; low-risk external mutation, no approval required
  create_location:
    endpoint: POST /v2/locations
    risk: creates a new organizational location; low-risk external mutation, no approval required
  create_title:
    endpoint: POST /v2/titles
    risk: creates a new job title; low-risk external mutation, no approval required
  create_user:
    endpoint: POST /v2/users
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
    required fields: id
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
