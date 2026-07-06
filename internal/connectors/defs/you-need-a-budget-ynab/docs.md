# Overview

Reads YNAB budgets, accounts, categories, payees, months, transactions, and scheduled transactions,
and writes transaction/account/category/payee/scheduled-transaction mutations through the YNAB REST
API.

Readable streams: `budgets`, `accounts`, `transactions`, `categories`, `payees`, `months`,
`scheduled_transactions`.

Write actions: `create_transaction`, `update_transaction`, `delete_transaction`, `create_account`,
`create_category`, `update_category`, `update_month_category`, `create_payee`, `update_payee`,
`create_scheduled_transaction`, `delete_scheduled_transaction`.

Service API documentation: https://api.ynab.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); YNAB personal access token, sent as a Bearer token
  (Authorization: Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.ynab.com/v1`; format `uri`; YNAB API base URL
  override for tests or proxies.
- `budget_id` (optional, string); default `last-used`; YNAB budget ID scoping the 'accounts' and
  'transactions' streams; defaults to YNAB's 'last-used' budget alias.
- `limit` (optional, string); Optional 'limit' query param passed through verbatim on
  'accounts'/'transactions' reads.
- `mode` (optional, string).
- `month` (optional, string); YYYY-MM-01 month identifier (or the 'current' alias) required by the
  'months' stream's per-month detail path; also accepted by scheduled-transaction reads'
  since_date-equivalent narrowing is not applicable here (months has no since_date param).
- `since_date` (optional, string); format `date`; Optional YYYY-MM-DD lower bound passed as the
  'since_date' query param on 'accounts'/'transactions' reads.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.ynab.com/v1`, `budget_id=last-used`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/budgets`.

## Streams notes

Default pagination: single request; no pagination.

- `budgets`: GET `/budgets` - records path `data.budgets`; query `limit` from template `{{
  config.limit }}`, omitted when absent; `since_date` from template `{{ config.since_date }}`,
  omitted when absent; computed output fields `updated_at`; emits passthrough records.
- `accounts`: GET `/budgets/{{ config.budget_id }}/accounts` - records path `data.accounts`; query
  `limit` from template `{{ config.limit }}`, omitted when absent; `since_date` from template `{{
  config.since_date }}`, omitted when absent; computed output fields `updated_at`; emits passthrough
  records.
- `transactions`: GET `/budgets/{{ config.budget_id }}/transactions` - records path
  `data.transactions`; query `limit` from template `{{ config.limit }}`, omitted when absent;
  `since_date` from template `{{ config.since_date }}`, omitted when absent; computed output fields
  `name`, `updated_at`; emits passthrough records.
- `categories`: GET `/budgets/{{ config.budget_id }}/categories` - records path
  `data.category_groups`; computed output fields `updated_at`; emits passthrough records.
- `payees`: GET `/budgets/{{ config.budget_id }}/payees` - records path `data.payees`; computed
  output fields `updated_at`; emits passthrough records.
- `months`: GET `/budgets/{{ config.budget_id }}/months` - records path `data.months`; computed
  output fields `id`, `updated_at`; emits passthrough records.
- `scheduled_transactions`: GET `/budgets/{{ config.budget_id }}/scheduled_transactions` - records
  path `data.scheduled_transactions`; computed output fields `name`, `updated_at`; emits passthrough
  records.

## Write actions & risks

Overall write risk: external mutation: creates/updates/deletes budget transactions, creates
accounts/categories/payees/scheduled transactions, updates category names/goals and month-category
budgeted amounts.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_transaction`: POST `/budgets/{{ config.budget_id }}/transactions` - kind `create`; body
  type `json`; required record fields `transaction`; accepted fields `transaction`; risk: external
  mutation; creates a new budget transaction; approval required. Body is wrapped under a top-level
  "transaction" key (YNAB's own POST /budgets/{budget_id}/transactions convention) - the record
  itself carries that wrapper, since the engine's write dialect sends record fields verbatim as the
  JSON body with no nested-wrapper construction primitive (see teamwork/bitly precedent).
- `update_transaction`: PUT `/budgets/{{ config.budget_id }}/transactions/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`, `transaction`; accepted
  fields `id`, `transaction`; risk: external mutation; updates an existing budget transaction
  (amount, category, memo, cleared/approved status); approval required.
- `delete_transaction`: DELETE `/budgets/{{ config.budget_id }}/transactions/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; risk: irreversible external deletion; deletes
  a budget transaction (YNAB marks it deleted rather than purging, but it disappears from active
  budget totals); approval required.
- `create_account`: POST `/budgets/{{ config.budget_id }}/accounts` - kind `create`; body type
  `json`; required record fields `account`; accepted fields `account`; risk: external mutation;
  creates a new budget account with an opening balance; approval required. This action cannot be
  undone via the API (YNAB has no delete-account endpoint).
- `create_category`: POST `/budgets/{{ config.budget_id }}/categories` - kind `create`; body type
  `json`; required record fields `category`; accepted fields `category`; risk: external mutation;
  creates a new budget category within a category group; approval required.
- `update_category`: PATCH `/budgets/{{ config.budget_id }}/categories/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`, `category`; accepted
  fields `category`, `id`; risk: external mutation; renames/re-notes/re-goals an existing budget
  category; approval required.
- `update_month_category`: PATCH `/budgets/{{ config.budget_id }}/months/{{ record.month
  }}/categories/{{ record.category_id }}` - kind `update`; body type `json`; path fields `month`,
  `category_id`; required record fields `month`, `category_id`, `category`; accepted fields
  `category`, `category_id`, `month`; risk: external mutation; reassigns (budgets) an amount to a
  category for a specific month; approval required.
- `create_payee`: POST `/budgets/{{ config.budget_id }}/payees` - kind `create`; body type `json`;
  required record fields `payee`; accepted fields `payee`; risk: external mutation; creates a new
  payee; approval required.
- `update_payee`: PATCH `/budgets/{{ config.budget_id }}/payees/{{ record.id }}` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`, `payee`; accepted fields `id`,
  `payee`; risk: external mutation; renames an existing payee (also renames the corresponding
  transactions and shared payee history); approval required.
- `create_scheduled_transaction`: POST `/budgets/{{ config.budget_id }}/scheduled_transactions` -
  kind `create`; body type `json`; required record fields `scheduled_transaction`; accepted fields
  `scheduled_transaction`; risk: external mutation; creates a new recurring scheduled transaction
  that will auto-post future budget transactions; approval required.
- `delete_scheduled_transaction`: DELETE `/budgets/{{ config.budget_id }}/scheduled_transactions/{{
  record.id }}` - kind `delete`; body type `none`; path fields `id`; required record fields `id`;
  accepted fields `id`; missing records treated as success for status `404`; risk: irreversible
  external deletion; removes a recurring scheduled transaction; approval required.

## Known limits

- API coverage includes 7 stream-backed endpoint group(s), 11 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=17, non_data_endpoint=2, out_of_scope=7.
