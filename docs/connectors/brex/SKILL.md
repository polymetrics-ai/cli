---
name: pm-brex
description: Brex connector knowledge and safe action guide.
---

# pm-brex

## Purpose

Reads and writes Brex transactions, users, expenses, vendors, budgets, cards, accounts, statements, transfers, and webhooks through the Brex platform REST API.

## Icon

- id: simple-icons-brex
- asset: icons/simple-icons/brex.svg
- title: Brex
- simple_icon_slug: brex
- simple_icon_hex: 212121
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Brex
- match: exact-name-or-slug
- matched_by: brex

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- mode
- page_size
- start_date
- user_token (secret)

## ETL Streams

- transactions:
  - primary key: id
  - cursor: posted_at_date
  - fields: amount(object), card_id(string), description(string), id(string), initiated_at_date(string), posted_at_date(string), type(string)
- users:
  - primary key: id
  - fields: department_id(string), email(string), first_name(string), id(string), last_name(string), manager_id(string), status(string)
- expenses:
  - primary key: id
  - cursor: purchased_at
  - fields: category(string), department_id(string), id(string), location_id(string), memo(string), merchant_id(string), original_amount(object), purchased_at(string), status(string), updated_at(string), user_id(string)
- vendors:
  - primary key: id
  - fields: company_name(string), email(string), id(string), payment_accounts(array), phone(string)
- budgets:
  - primary key: budget_id
  - fields: account_id(string), budget_id(string), creator_user_id(string), description(string), limit(object), name(string), parent_budget_id(string), period_type(string), status(string)
- departments:
  - primary key: id
  - fields: description(string), id(string), name(string)
- locations:
  - primary key: id
  - fields: description(string), id(string), name(string)
- titles:
  - primary key: id
  - fields: id(string), name(string)
- legal_entities:
  - primary key: id
  - fields: billingAddress(object), createdAt(string), displayName(string), id(string), isDefault(boolean), status(string)
- cards:
  - primary key: id
  - fields: billing_address(object), budget_id(string), card_name(string), card_type(string), expiration_date(object), has_been_transferred(boolean), id(string), last_four(string), limit_type(string), mailing_address(object), owner(object), spend_controls(object), status(string)
- accounts_card:
  - primary key: id
  - fields: account_limit(object), available_balance(object), current_balance(object), current_statement_period(object), id(string), status(string)
- accounts_cash:
  - primary key: id
  - fields: account_number(string), available_balance(object), current_balance(object), id(string), name(string), primary(boolean), routing_number(string), status(string)
- card_statements:
  - primary key: id
  - fields: end_balance(object), id(string), period(object), start_balance(object)
- linked_accounts:
  - primary key: id
  - fields: available_balance(object), bank_details(object), brex_account_id(string), current_balance(object), id(string), last_four(string)
- transfers:
  - primary key: id
  - fields: amount(object), cancellation_reason(string), counterparty(object), creator_user_id(string), description(string), estimated_delivery_date(string), id(string), originating_account(object), payment_type(string), process_date(string), status(string)
- webhooks:
  - primary key: id
  - fields: event_types(array), group_id(string), id(string), status(string), url(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- update_vendor:
  - endpoint: PUT /v1/vendors/{{ record.id }}
  - required fields: id
  - risk: mutates an existing vendor's name, contact details, or payment accounts; affects future transfer counterparty resolution
- delete_vendor:
  - endpoint: DELETE /v1/vendors/{{ record.id }}
  - required fields: id
  - risk: permanently removes a vendor record; any transfer still referencing it as counterparty will fail to resolve
- create_department:
  - endpoint: POST /v2/departments
  - required fields: name
  - risk: creates a new organizational department; low-risk external mutation, no approval required
- create_location:
  - endpoint: POST /v2/locations
  - required fields: name
  - risk: creates a new organizational location; low-risk external mutation, no approval required
- create_title:
  - endpoint: POST /v2/titles
  - required fields: name
  - risk: creates a new job title; low-risk external mutation, no approval required
- create_user:
  - endpoint: POST /v2/users
  - required fields: email, first_name, last_name
  - risk: invites a new user to the Brex account; sends a real invitation email to the target address
- update_user:
  - endpoint: PUT /v2/users/{{ record.id }}
  - required fields: id
  - risk: mutates an existing user's status, manager, department, location, or title; setting status to a terminated/suspended state revokes account access
- update_card:
  - endpoint: PUT /v2/cards/{{ record.id }}
  - required fields: id
  - risk: mutates an existing card's spend controls (limit amount/category/merchant restrictions); takes effect on the physical/virtual card immediately
- lock_card:
  - endpoint: POST /v2/cards/{{ record.id }}/lock
  - required fields: id
  - risk: immediately blocks all new transactions on the card until unlocked; does not affect already-authorized/pending transactions
- unlock_card:
  - endpoint: POST /v2/cards/{{ record.id }}/unlock
  - required fields: id
  - risk: immediately re-enables new transactions on a previously locked card
- terminate_card:
  - endpoint: POST /v2/cards/{{ record.id }}/terminate
  - required fields: id
  - risk: permanently deactivates a card; irreversible, the card can never be unlocked or reused after termination
- update_expense:
  - endpoint: PUT /v1/expenses/card/{{ record.expense_id }}
  - required fields: expense_id
  - risk: mutates an existing card expense's memo; low-risk metadata-only external mutation
- update_webhook:
  - endpoint: PUT /v1/webhooks/{{ record.id }}
  - required fields: id, url, event_types, status
  - risk: re-points an already-registered webhook's delivery URL, event set, or active status; redirects live event delivery immediately
- delete_webhook:
  - endpoint: DELETE /v1/webhooks/{{ record.id }}
  - required fields: id
  - risk: permanently removes a webhook subscription; irreversible

## Security

- read risk: external Brex API read of card transaction, user, expense, vendor, card, account, and transfer data
- write risk: external mutation of vendors, org directory (departments/locations/titles/users), card controls/lifecycle, expenses, and webhooks; card lock/unlock/terminate take effect on real payment instruments immediately
- approval: required for all write actions; each action's per-record risk string in writes.json is the authoritative summary
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Brex's declared streams and reverse-ETL actions.
- Usage: pm brex <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - accounts card list - Run the accounts card ETL stream [intent=etl availability=implemented stream=accounts_card]
  - accounts cash list - Run the accounts cash ETL stream [intent=etl availability=implemented stream=accounts_cash]
  - api delete v1 fields field-id values - Documented DELETE /v1/fields/{field_id}/values (not implemented) [intent=direct_write availability=not_implemented operation=brex.delete.v1-fields-field-id-values]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v1 fields id - Documented DELETE /v1/fields/{id} (not implemented) [intent=direct_write availability=not_implemented operation=brex.delete.v1-fields-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v1 webhooks groups id - Documented DELETE /v1/webhooks/groups/{id} (not implemented) [intent=direct_write availability=not_implemented operation=brex.delete.v1-webhooks-groups-id]; approval: not implemented: the provider webhook-URL mutation has no dedicated security-reviewed execution contract; risk: high; notes: named_dependency=review.webhook_url_mutation: the provider webhook-URL mutation has no dedicated security-reviewed execution contract
  - api get v1 budget-programs - Documented GET /v1/budget_programs (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-budget-programs]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 budget-programs id - Documented GET /v1/budget_programs/{id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-budget-programs-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 budgets - Documented GET /v1/budgets (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-budgets]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 budgets id - Documented GET /v1/budgets/{id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-budgets-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 expenses - Documented GET /v1/expenses (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-expenses]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 expenses card expense-id - Documented GET /v1/expenses/card/{expense_id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-expenses-card-expense-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 expenses id - Documented GET /v1/expenses/{id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-expenses-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 fields - Documented GET /v1/fields (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-fields]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 fields field-id values - Documented GET /v1/fields/{field_id}/values (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-fields-field-id-values]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 fields field-id values brex-id - Documented GET /v1/fields/{field_id}/values/{brex_id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-fields-field-id-values-brex-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 fields id - Documented GET /v1/fields/{id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-fields-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 referrals - Documented GET /v1/referrals (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-referrals]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 referrals id - Documented GET /v1/referrals/{id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-referrals-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 transfers id - Documented GET /v1/transfers/{id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-transfers-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 trips - Documented GET /v1/trips (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-trips]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 trips trip-id - Documented GET /v1/trips/{trip_id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-trips-trip-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 trips trip-id bookings - Documented GET /v1/trips/{trip_id}/bookings (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-trips-trip-id-bookings]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 trips trip-id bookings booking-id - Documented GET /v1/trips/{trip_id}/bookings/{booking_id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-trips-trip-id-bookings-booking-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 vendors id - Documented GET /v1/vendors/{id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-vendors-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 webhooks groups - Documented GET /v1/webhooks/groups (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-webhooks-groups]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 webhooks groups id - Documented GET /v1/webhooks/groups/{id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-webhooks-groups-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 webhooks groups id members - Documented GET /v1/webhooks/groups/{id}/members (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-webhooks-groups-id-members]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 webhooks id - Documented GET /v1/webhooks/{id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-webhooks-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 webhooks secrets - Documented GET /v1/webhooks/secrets (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v1-webhooks-secrets]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 accounts cash id - Documented GET /v2/accounts/cash/{id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v2-accounts-cash-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 accounts cash id statements - Documented GET /v2/accounts/cash/{id}/statements (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v2-accounts-cash-id-statements]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 accounts cash primary - Documented GET /v2/accounts/cash/primary (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v2-accounts-cash-primary]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 budgets id - Documented GET /v2/budgets/{id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v2-budgets-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 cards id - Documented GET /v2/cards/{id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v2-cards-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 cards id pan - Documented GET /v2/cards/{id}/pan (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v2-cards-id-pan]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 company - Documented GET /v2/company (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v2-company]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 departments id - Documented GET /v2/departments/{id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v2-departments-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 legal-entities id - Documented GET /v2/legal_entities/{id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v2-legal-entities-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 locations id - Documented GET /v2/locations/{id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v2-locations-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 spend-limits - Documented GET /v2/spend_limits (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v2-spend-limits]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 spend-limits id - Documented GET /v2/spend_limits/{id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v2-spend-limits-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 titles id - Documented GET /v2/titles/{id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v2-titles-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 transactions cash id - Documented GET /v2/transactions/cash/{id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v2-transactions-cash-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 users id - Documented GET /v2/users/{id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v2-users-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 users id limit - Documented GET /v2/users/{id}/limit (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v2-users-id-limit]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 users me - Documented GET /v2/users/me (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v2-users-me]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 accounting records - Documented GET /v3/accounting/records (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v3-accounting-records]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 accounting records record-id - Documented GET /v3/accounting/records/{record_id} (not implemented) [intent=direct_read availability=not_implemented operation=brex.get.v3-accounting-records-record-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api post v1 budgets - Documented POST /v1/budgets (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v1-budgets]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 budgets id archive - Documented POST /v1/budgets/{id}/archive (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v1-budgets-id-archive]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 expenses card expense-id receipt-upload - Documented POST /v1/expenses/card/{expense_id}/receipt_upload (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v1-expenses-card-expense-id-receipt-upload]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 expenses card receipt-match - Documented POST /v1/expenses/card/receipt_match (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v1-expenses-card-receipt-match]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 fields - Documented POST /v1/fields (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v1-fields]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 fields field-id values - Documented POST /v1/fields/{field_id}/values (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v1-fields-field-id-values]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 incoming-transfers - Documented POST /v1/incoming_transfers (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v1-incoming-transfers]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 referrals - Documented POST /v1/referrals (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v1-referrals]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 referrals id document-upload - Documented POST /v1/referrals/{id}/document_upload (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v1-referrals-id-document-upload]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 referrals id process-ein-document - Documented POST /v1/referrals/{id}/process_ein_document (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v1-referrals-id-process-ein-document]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 transfers - Documented POST /v1/transfers (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v1-transfers]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 vendors - Documented POST /v1/vendors (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v1-vendors]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 webhooks - Documented POST /v1/webhooks (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v1-webhooks]; approval: not implemented: the provider webhook-URL mutation has no dedicated security-reviewed execution contract; risk: high; notes: named_dependency=review.webhook_url_mutation: the provider webhook-URL mutation has no dedicated security-reviewed execution contract
  - api post v1 webhooks groups - Documented POST /v1/webhooks/groups (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v1-webhooks-groups]; approval: not implemented: the provider webhook-URL mutation has no dedicated security-reviewed execution contract; risk: high; notes: named_dependency=review.webhook_url_mutation: the provider webhook-URL mutation has no dedicated security-reviewed execution contract
  - api post v1 webhooks groups id add-members - Documented POST /v1/webhooks/groups/{id}/add_members (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v1-webhooks-groups-id-add-members]; approval: not implemented: the provider webhook-URL mutation has no dedicated security-reviewed execution contract; risk: high; notes: named_dependency=review.webhook_url_mutation: the provider webhook-URL mutation has no dedicated security-reviewed execution contract
  - api post v1 webhooks groups id remove-members - Documented POST /v1/webhooks/groups/{id}/remove_members (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v1-webhooks-groups-id-remove-members]; approval: not implemented: the provider webhook-URL mutation has no dedicated security-reviewed execution contract; risk: high; notes: named_dependency=review.webhook_url_mutation: the provider webhook-URL mutation has no dedicated security-reviewed execution contract
  - api post v2 budgets - Documented POST /v2/budgets (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v2-budgets]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v2 budgets id archive - Documented POST /v2/budgets/{id}/archive (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v2-budgets-id-archive]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v2 cards - Documented POST /v2/cards (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v2-cards]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v2 cards id secure-email - Documented POST /v2/cards/{id}/secure_email (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v2-cards-id-secure-email]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v2 spend-limits - Documented POST /v2/spend_limits (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v2-spend-limits]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v2 spend-limits id archive - Documented POST /v2/spend_limits/{id}/archive (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v2-spend-limits-id-archive]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v2 users id limit - Documented POST /v2/users/{id}/limit (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v2-users-id-limit]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 accounting integration - Documented POST /v3/accounting/integration (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v3-accounting-integration]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 accounting integration integration-id disconnect - Documented POST /v3/accounting/integration/{integration_id}/disconnect (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v3-accounting-integration-integration-id-disconnect]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 accounting integration integration-id reactivate - Documented POST /v3/accounting/integration/{integration_id}/reactivate (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v3-accounting-integration-integration-id-reactivate]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 accounting records export-results - Documented POST /v3/accounting/records/export-results (not implemented) [intent=direct_write availability=not_implemented operation=brex.post.v3-accounting-records-export-results]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put v1 budgets id - Documented PUT /v1/budgets/{id} (not implemented) [intent=direct_write availability=not_implemented operation=brex.put.v1-budgets-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put v1 fields field-id values - Documented PUT /v1/fields/{field_id}/values (not implemented) [intent=direct_write availability=not_implemented operation=brex.put.v1-fields-field-id-values]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put v1 fields id - Documented PUT /v1/fields/{id} (not implemented) [intent=direct_write availability=not_implemented operation=brex.put.v1-fields-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put v2 budgets id - Documented PUT /v2/budgets/{id} (not implemented) [intent=direct_write availability=not_implemented operation=brex.put.v2-budgets-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put v2 spend-limits id - Documented PUT /v2/spend_limits/{id} (not implemented) [intent=direct_write availability=not_implemented operation=brex.put.v2-spend-limits-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - budgets list - Run the budgets ETL stream [intent=etl availability=implemented stream=budgets]
  - card statements list - Run the card statements ETL stream [intent=etl availability=implemented stream=card_statements]
  - cards list - Run the cards ETL stream [intent=etl availability=implemented stream=cards]
  - create department apply - Plan and execute the create department reverse-ETL action [intent=reverse_etl availability=implemented write=create_department]; approval: requires plan, preview, approval, and execute; risk: creates a new organizational department; low-risk external mutation, no approval required; flags: --name (required)
  - create location apply - Plan and execute the create location reverse-ETL action [intent=reverse_etl availability=implemented write=create_location]; approval: requires plan, preview, approval, and execute; risk: creates a new organizational location; low-risk external mutation, no approval required; flags: --name (required)
  - create title apply - Plan and execute the create title reverse-ETL action [intent=reverse_etl availability=implemented write=create_title]; approval: requires plan, preview, approval, and execute; risk: creates a new job title; low-risk external mutation, no approval required; flags: --name (required)
  - create user apply - Plan and execute the create user reverse-ETL action [intent=reverse_etl availability=implemented write=create_user]; approval: requires plan, preview, approval, and execute; risk: invites a new user to the Brex account; sends a real invitation email to the target address; flags: --email (required), --first_name (required), --last_name (required)
  - delete vendor apply - Plan and execute the delete vendor reverse-ETL action [intent=reverse_etl availability=implemented write=delete_vendor]; approval: requires plan, preview, approval, and execute; risk: permanently removes a vendor record; any transfer still referencing it as counterparty will fail to resolve; flags: --id (required)
  - delete webhook apply - Plan and execute the delete webhook reverse-ETL action [intent=reverse_etl availability=implemented write=delete_webhook]; approval: requires plan, preview, approval, and execute; risk: permanently removes a webhook subscription; irreversible; flags: --id (required)
  - departments list - Run the departments ETL stream [intent=etl availability=implemented stream=departments]
  - expenses list - Run the expenses ETL stream [intent=etl availability=implemented stream=expenses]
  - legal entities list - Run the legal entities ETL stream [intent=etl availability=implemented stream=legal_entities]
  - linked accounts list - Run the linked accounts ETL stream [intent=etl availability=implemented stream=linked_accounts]
  - locations list - Run the locations ETL stream [intent=etl availability=implemented stream=locations]
  - lock card apply - Plan and execute the lock card reverse-ETL action [intent=reverse_etl availability=implemented write=lock_card]; approval: requires plan, preview, approval, and execute; risk: immediately blocks all new transactions on the card until unlocked; does not affect already-authorized/pending transactions; flags: --id (required)
  - terminate card apply - Plan and execute the terminate card reverse-ETL action [intent=reverse_etl availability=implemented write=terminate_card]; approval: requires plan, preview, approval, and execute; risk: permanently deactivates a card; irreversible, the card can never be unlocked or reused after termination; flags: --id (required)
  - titles list - Run the titles ETL stream [intent=etl availability=implemented stream=titles]
  - transactions list - Run the transactions ETL stream [intent=etl availability=implemented stream=transactions]
  - transfers list - Run the transfers ETL stream [intent=etl availability=implemented stream=transfers]
  - unlock card apply - Plan and execute the unlock card reverse-ETL action [intent=reverse_etl availability=implemented write=unlock_card]; approval: requires plan, preview, approval, and execute; risk: immediately re-enables new transactions on a previously locked card; flags: --id (required)
  - update card apply - Plan and execute the update card reverse-ETL action [intent=reverse_etl availability=implemented write=update_card]; approval: requires plan, preview, approval, and execute; risk: mutates an existing card's spend controls (limit amount/category/merchant restrictions); takes effect on the physical/virtual card immediately; flags: --id (required)
  - update expense apply - Plan and execute the update expense reverse-ETL action [intent=reverse_etl availability=implemented write=update_expense]; approval: requires plan, preview, approval, and execute; risk: mutates an existing card expense's memo; low-risk metadata-only external mutation; flags: --expense_id (required)
  - update user apply - Plan and execute the update user reverse-ETL action [intent=reverse_etl availability=implemented write=update_user]; approval: requires plan, preview, approval, and execute; risk: mutates an existing user's status, manager, department, location, or title; setting status to a terminated/suspended state revokes account access; flags: --id (required)
  - update vendor apply - Plan and execute the update vendor reverse-ETL action [intent=reverse_etl availability=implemented write=update_vendor]; approval: requires plan, preview, approval, and execute; risk: mutates an existing vendor's name, contact details, or payment accounts; affects future transfer counterparty resolution; flags: --id (required)
  - update webhook apply - Plan and execute the update webhook reverse-ETL action [intent=reverse_etl availability=implemented write=update_webhook]; approval: requires plan, preview, approval, and execute; risk: re-points an already-registered webhook's delivery URL, event set, or active status; redirects live event delivery immediately; flags: --event_types (required), --id (required), --status (required), --url (required)
  - users list - Run the users ETL stream [intent=etl availability=implemented stream=users]
  - vendors list - Run the vendors ETL stream [intent=etl availability=implemented stream=vendors]
  - webhooks list - Run the webhooks ETL stream [intent=etl availability=implemented stream=webhooks]

## Commands

### Inspect as a manual

```bash
pm connectors inspect brex
```

### Inspect as structured JSON

```bash
pm connectors inspect brex --json
```

## Agent Rules

- Run pm connectors inspect brex before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
