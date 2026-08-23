---
name: pm-you-need-a-budget-ynab
description: You Need A Budget (YNAB) connector knowledge and safe action guide.
---

# pm-you-need-a-budget-ynab

## Purpose

Reads YNAB budgets, accounts, categories, payees, months, transactions, and scheduled transactions, and writes transaction/account/category/payee/scheduled-transaction mutations through the YNAB REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- budget_id
- limit
- mode
- month
- since_date
- api_key (secret) (required)

## ETL Streams

- budgets:
  - primary key: id
  - cursor: updated_at
  - fields: currency_format(object), date_format(object), first_month(string), id(string), last_modified_on(string), last_month(string), name(string), updated_at(string)
- accounts:
  - primary key: id
  - cursor: updated_at
  - fields: balance(integer), cleared_balance(integer), closed(boolean), deleted(boolean), id(string), last_reconciled_at(string), name(string), on_budget(boolean), type(string), uncleared_balance(integer), updated_at(string)
- transactions:
  - primary key: id
  - cursor: updated_at
  - fields: account_id(string), amount(integer), approved(boolean), category_id(string), category_name(string), cleared(string), date(string), deleted(boolean), id(string), memo(string), name(string), payee_id(string), payee_name(string), updated_at(string)
- categories:
  - primary key: id
  - cursor: updated_at
  - fields: categories(array), deleted(boolean), hidden(boolean), id(string), internal(boolean), name(string), updated_at(string)
- payees:
  - primary key: id
  - cursor: updated_at
  - fields: deleted(boolean), id(string), name(string), transfer_account_id(string), updated_at(string)
- months:
  - primary key: id
  - cursor: updated_at
  - fields: activity(integer), age_of_money(integer), budgeted(integer), deleted(boolean), id(string), income(integer), month(string), note(string), to_be_budgeted(integer), updated_at(string)
- scheduled_transactions:
  - primary key: id
  - cursor: updated_at
  - fields: account_id(string), amount(integer), category_id(string), category_name(string), date_first(string), date_next(string), deleted(boolean), flag_color(string), frequency(string), id(string), memo(string), name(string), payee_id(string), payee_name(string), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_transaction:
  - endpoint: POST /budgets/{{ config.budget_id }}/transactions
  - required fields: transaction
  - risk: external mutation; creates a new budget transaction; approval required. Body is wrapped under a top-level "transaction" key (YNAB's own POST /budgets/{budget_id}/transactions convention) — the record itself carries that wrapper, since the engine's write dialect sends record fields verbatim as the JSON body with no nested-wrapper construction primitive (see teamwork/bitly precedent).
- update_transaction:
  - endpoint: PUT /budgets/{{ config.budget_id }}/transactions/{{ record.id }}
  - required fields: id, transaction
  - risk: external mutation; updates an existing budget transaction (amount, category, memo, cleared/approved status); approval required
- delete_transaction:
  - endpoint: DELETE /budgets/{{ config.budget_id }}/transactions/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion; deletes a budget transaction (YNAB marks it deleted rather than purging, but it disappears from active budget totals); approval required
- create_account:
  - endpoint: POST /budgets/{{ config.budget_id }}/accounts
  - required fields: account
  - risk: external mutation; creates a new budget account with an opening balance; approval required. This action cannot be undone via the API (YNAB has no delete-account endpoint).
- create_category:
  - endpoint: POST /budgets/{{ config.budget_id }}/categories
  - required fields: category
  - risk: external mutation; creates a new budget category within a category group; approval required
- update_category:
  - endpoint: PATCH /budgets/{{ config.budget_id }}/categories/{{ record.id }}
  - required fields: id, category
  - risk: external mutation; renames/re-notes/re-goals an existing budget category; approval required
- update_month_category:
  - endpoint: PATCH /budgets/{{ config.budget_id }}/months/{{ record.month }}/categories/{{ record.category_id }}
  - required fields: month, category_id, category
  - risk: external mutation; reassigns (budgets) an amount to a category for a specific month; approval required
- create_payee:
  - endpoint: POST /budgets/{{ config.budget_id }}/payees
  - required fields: payee
  - risk: external mutation; creates a new payee; approval required
- update_payee:
  - endpoint: PATCH /budgets/{{ config.budget_id }}/payees/{{ record.id }}
  - required fields: id, payee
  - risk: external mutation; renames an existing payee (also renames the corresponding transactions and shared payee history); approval required
- create_scheduled_transaction:
  - endpoint: POST /budgets/{{ config.budget_id }}/scheduled_transactions
  - required fields: scheduled_transaction
  - risk: external mutation; creates a new recurring scheduled transaction that will auto-post future budget transactions; approval required
- delete_scheduled_transaction:
  - endpoint: DELETE /budgets/{{ config.budget_id }}/scheduled_transactions/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion; removes a recurring scheduled transaction; approval required

## Security

- read risk: external YNAB API read of budget, account, category, payee, month, transaction, and scheduled-transaction data
- write risk: external mutation: creates/updates/deletes budget transactions, creates accounts/categories/payees/scheduled transactions, updates category names/goals and month-category budgeted amounts
- approval: required for all write actions; reads require none
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect you-need-a-budget-ynab
```

### Inspect as structured JSON

```bash
pm connectors inspect you-need-a-budget-ynab --json
```

## Agent Rules

- Run pm connectors inspect you-need-a-budget-ynab before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
