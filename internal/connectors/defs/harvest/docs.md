# Overview

Reads Harvest clients, contacts, company settings, projects, tasks, task assignments, users, time
entries, invoices, estimates, expenses, item categories, expense categories, and roles through the
Harvest v2 REST API.

Readable streams: `clients`, `projects`, `tasks`, `users`, `time_entries`, `contacts`, `company`,
`invoices`, `estimates`, `expenses`, `invoice_item_categories`, `estimate_item_categories`,
`expense_categories`, `roles`, `task_assignments`.

This connector is read-only; no write actions are declared.

Service API documentation: https://help.getharvest.com/api-v2/.

## Auth setup

Connection fields:

- `account_id` (required, string); Harvest account ID, sent as the Harvest-Account-Id header on
  every request.
- `api_token` (required, secret, string); Harvest personal access token, sent as a Bearer token.
  Never logged.
- `base_url` (optional, string); default `https://api.harvestapp.com/v2`; format `uri`; Harvest API
  base URL override for tests or proxies.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-2000).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only objects updated at
  or after this time are read on a fresh sync.

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.harvestapp.com/v2`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/clients`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page`; next token from `next_page`.

Pagination by stream: cursor: `clients`, `projects`, `tasks`, `users`, `time_entries`, `contacts`,
`invoices`, `estimates`, `expenses`, `invoice_item_categories`, `estimate_item_categories`,
`expense_categories`, `roles`, `task_assignments`; none: `company`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `clients`: GET `/clients` - records path `clients`; query `per_page`=`{{ config.page_size }}`;
  `updated_since` from template `{{ incremental.lower_bound }}`, omitted when absent; cursor
  pagination; cursor parameter `page`; next token from `next_page`; incremental cursor `updated_at`;
  sent as `updated_since`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `projects`: GET `/projects` - records path `projects`; query `per_page`=`{{ config.page_size }}`;
  `updated_since` from template `{{ incremental.lower_bound }}`, omitted when absent; cursor
  pagination; cursor parameter `page`; next token from `next_page`; incremental cursor `updated_at`;
  sent as `updated_since`; formatted as `rfc3339`; initial lower bound from `start_date`; computed
  output fields `client_id`, `client_name`.
- `tasks`: GET `/tasks` - records path `tasks`; query `per_page`=`{{ config.page_size }}`;
  `updated_since` from template `{{ incremental.lower_bound }}`, omitted when absent; cursor
  pagination; cursor parameter `page`; next token from `next_page`; incremental cursor `updated_at`;
  sent as `updated_since`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `users`: GET `/users` - records path `users`; query `per_page`=`{{ config.page_size }}`;
  `updated_since` from template `{{ incremental.lower_bound }}`, omitted when absent; cursor
  pagination; cursor parameter `page`; next token from `next_page`; incremental cursor `updated_at`;
  sent as `updated_since`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `time_entries`: GET `/time_entries` - records path `time_entries`; query `per_page`=`{{
  config.page_size }}`; `updated_since` from template `{{ incremental.lower_bound }}`, omitted when
  absent; cursor pagination; cursor parameter `page`; next token from `next_page`; incremental
  cursor `updated_at`; sent as `updated_since`; formatted as `rfc3339`; initial lower bound from
  `start_date`; computed output fields `client_id`, `project_id`, `task_id`, `user_id`.
- `contacts`: GET `/contacts` - records path `contacts`; query `per_page`=`{{ config.page_size }}`;
  `updated_since` from template `{{ incremental.lower_bound }}`, omitted when absent; cursor
  pagination; cursor parameter `page`; next token from `next_page`; incremental cursor `updated_at`;
  sent as `updated_since`; formatted as `rfc3339`; initial lower bound from `start_date`; computed
  output fields `client_id`, `client_name`.
- `company`: GET `/company` - single-object response; records path `.`.
- `invoices`: GET `/invoices` - records path `invoices`; query `per_page`=`{{ config.page_size }}`;
  `updated_since` from template `{{ incremental.lower_bound }}`, omitted when absent; cursor
  pagination; cursor parameter `page`; next token from `next_page`; incremental cursor `updated_at`;
  sent as `updated_since`; formatted as `rfc3339`; initial lower bound from `start_date`; computed
  output fields `client_id`, `client_name`, `creator_id`, `creator_name`.
- `estimates`: GET `/estimates` - records path `estimates`; query `per_page`=`{{ config.page_size
  }}`; `updated_since` from template `{{ incremental.lower_bound }}`, omitted when absent; cursor
  pagination; cursor parameter `page`; next token from `next_page`; incremental cursor `updated_at`;
  sent as `updated_since`; formatted as `rfc3339`; initial lower bound from `start_date`; computed
  output fields `client_id`, `client_name`, `creator_id`, `creator_name`.
- `expenses`: GET `/expenses` - records path `expenses`; query `per_page`=`{{ config.page_size }}`;
  `updated_since` from template `{{ incremental.lower_bound }}`, omitted when absent; cursor
  pagination; cursor parameter `page`; next token from `next_page`; incremental cursor `updated_at`;
  sent as `updated_since`; formatted as `rfc3339`; initial lower bound from `start_date`; computed
  output fields `client_id`, `expense_category_id`, `invoice_id`, `project_id`, `user_id`.
- `invoice_item_categories`: GET `/invoice_item_categories` - records path
  `invoice_item_categories`; query `per_page`=`{{ config.page_size }}`; `updated_since` from
  template `{{ incremental.lower_bound }}`, omitted when absent; cursor pagination; cursor parameter
  `page`; next token from `next_page`; incremental cursor `updated_at`; sent as `updated_since`;
  formatted as `rfc3339`; initial lower bound from `start_date`.
- `estimate_item_categories`: GET `/estimate_item_categories` - records path
  `estimate_item_categories`; query `per_page`=`{{ config.page_size }}`; `updated_since` from
  template `{{ incremental.lower_bound }}`, omitted when absent; cursor pagination; cursor parameter
  `page`; next token from `next_page`; incremental cursor `updated_at`; sent as `updated_since`;
  formatted as `rfc3339`; initial lower bound from `start_date`.
- `expense_categories`: GET `/expense_categories` - records path `expense_categories`; query
  `per_page`=`{{ config.page_size }}`; `updated_since` from template `{{ incremental.lower_bound
  }}`, omitted when absent; cursor pagination; cursor parameter `page`; next token from `next_page`;
  incremental cursor `updated_at`; sent as `updated_since`; formatted as `rfc3339`; initial lower
  bound from `start_date`.
- `roles`: GET `/roles` - records path `roles`; query `per_page`=`{{ config.page_size }}`;
  `updated_since` from template `{{ incremental.lower_bound }}`, omitted when absent; cursor
  pagination; cursor parameter `page`; next token from `next_page`; incremental cursor `updated_at`;
  sent as `updated_since`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `task_assignments`: GET `/task_assignments` - records path `task_assignments`; query
  `per_page`=`{{ config.page_size }}`; `updated_since` from template `{{ incremental.lower_bound
  }}`, omitted when absent; cursor pagination; cursor parameter `page`; next token from `next_page`;
  incremental cursor `updated_at`; sent as `updated_since`; formatted as `rfc3339`; initial lower
  bound from `start_date`; computed output fields `project_id`, `task_id`.

## Write actions & risks

This connector is read-only. Read behavior: external Harvest API read of Harvest business, time,
project, invoice, estimate, expense, role, and category metadata.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 15 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, duplicate_of=2, out_of_scope=4, requires_elevated_scope=3.
