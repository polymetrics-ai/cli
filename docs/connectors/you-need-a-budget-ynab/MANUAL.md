# pm connectors inspect you-need-a-budget-ynab

```text
NAME
  pm connectors inspect you-need-a-budget-ynab - You Need A Budget (YNAB) connector manual

SYNOPSIS
  pm connectors inspect you-need-a-budget-ynab
  pm connectors inspect you-need-a-budget-ynab --json
  pm credentials add <name> --connector you-need-a-budget-ynab [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads YNAB budgets, accounts, categories, payees, months, transactions, and scheduled transactions, and writes transaction/account/category/payee/scheduled-transaction mutations through the YNAB REST API.

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
  budget_id
  limit
  mode
  month
  since_date
  api_key (secret)

ETL STREAMS
  budgets:
    primary key: id
    cursor: updated_at
    fields: currency_format(), date_format(), first_month(), id(), last_modified_on(), last_month(), name(), updated_at()
  accounts:
    primary key: id
    cursor: updated_at
    fields: balance(), cleared_balance(), closed(), deleted(), id(), last_reconciled_at(), name(), on_budget(), type(), uncleared_balance(), updated_at()
  transactions:
    primary key: id
    cursor: updated_at
    fields: account_id(), amount(), approved(), category_id(), category_name(), cleared(), date(), deleted(), id(), memo(), name(), payee_id(), payee_name(), updated_at()
  categories:
    primary key: id
    cursor: updated_at
    fields: categories(), deleted(), hidden(), id(), internal(), name(), updated_at()
  payees:
    primary key: id
    cursor: updated_at
    fields: deleted(), id(), name(), transfer_account_id(), updated_at()
  months:
    primary key: id
    cursor: updated_at
    fields: activity(), age_of_money(), budgeted(), deleted(), id(), income(), month(), note(), to_be_budgeted(), updated_at()
  scheduled_transactions:
    primary key: id
    cursor: updated_at
    fields: account_id(), amount(), category_id(), category_name(), date_first(), date_next(), deleted(), flag_color(), frequency(), id(), memo(), name(), payee_id(), payee_name(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_transaction:
    endpoint: POST /budgets/{{ config.budget_id }}/transactions
    risk: external mutation; creates a new budget transaction; approval required. Body is wrapped under a top-level "transaction" key (YNAB's own POST /budgets/{budget_id}/transactions convention) — the record itself carries that wrapper, since the engine's write dialect sends record fields verbatim as the JSON body with no nested-wrapper construction primitive (see teamwork/bitly precedent).
  update_transaction:
    endpoint: PUT /budgets/{{ config.budget_id }}/transactions/{{ record.id }}
    required fields: id
    risk: external mutation; updates an existing budget transaction (amount, category, memo, cleared/approved status); approval required
  delete_transaction:
    endpoint: DELETE /budgets/{{ config.budget_id }}/transactions/{{ record.id }}
    required fields: id
    risk: irreversible external deletion; deletes a budget transaction (YNAB marks it deleted rather than purging, but it disappears from active budget totals); approval required
  create_account:
    endpoint: POST /budgets/{{ config.budget_id }}/accounts
    risk: external mutation; creates a new budget account with an opening balance; approval required. This action cannot be undone via the API (YNAB has no delete-account endpoint).
  create_category:
    endpoint: POST /budgets/{{ config.budget_id }}/categories
    risk: external mutation; creates a new budget category within a category group; approval required
  update_category:
    endpoint: PATCH /budgets/{{ config.budget_id }}/categories/{{ record.id }}
    required fields: id
    risk: external mutation; renames/re-notes/re-goals an existing budget category; approval required
  update_month_category:
    endpoint: PATCH /budgets/{{ config.budget_id }}/months/{{ record.month }}/categories/{{ record.category_id }}
    required fields: month, category_id
    risk: external mutation; reassigns (budgets) an amount to a category for a specific month; approval required
  create_payee:
    endpoint: POST /budgets/{{ config.budget_id }}/payees
    risk: external mutation; creates a new payee; approval required
  update_payee:
    endpoint: PATCH /budgets/{{ config.budget_id }}/payees/{{ record.id }}
    required fields: id
    risk: external mutation; renames an existing payee (also renames the corresponding transactions and shared payee history); approval required
  create_scheduled_transaction:
    endpoint: POST /budgets/{{ config.budget_id }}/scheduled_transactions
    risk: external mutation; creates a new recurring scheduled transaction that will auto-post future budget transactions; approval required
  delete_scheduled_transaction:
    endpoint: DELETE /budgets/{{ config.budget_id }}/scheduled_transactions/{{ record.id }}
    required fields: id
    risk: irreversible external deletion; removes a recurring scheduled transaction; approval required

SECURITY
  read risk: external YNAB API read of budget, account, category, payee, month, transaction, and scheduled-transaction data
  write risk: external mutation: creates/updates/deletes budget transactions, creates accounts/categories/payees/scheduled transactions, updates category names/goals and month-category budgeted amounts
  approval: required for all write actions; reads require none
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect you-need-a-budget-ynab

  # Inspect as structured JSON
  pm connectors inspect you-need-a-budget-ynab --json

AGENT WORKFLOW
  - Run pm connectors inspect you-need-a-budget-ynab before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
