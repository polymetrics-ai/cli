# Overview

Zendesk Support inventories 625 official Support API operation(s) from the Zendesk OAS 2.0.0 ledger. The executable fixture-backed bundle currently reads 33 stream(s), and writes through 84 action(s).

Readable streams: `tickets`, `users`, `organizations`, `groups`, `satisfaction_ratings`,
`deleted_tickets`, `account_attributes`, `attribute_definitions`, `brands`, `custom_roles`,
`schedules`, `sla_policies`, `tags`, `ticket_fields`, `ticket_forms`, `topics`, `user_fields`,
`automations`, `categories`, `sections`, `articles`, `group_memberships`, `macros`,
`organization_fields`, `organization_memberships`, `posts`, `ticket_activities`, `ticket_audits`,
`ticket_metric_events`, `ticket_events`, `ticket_skips`, `triggers`, `views`.

Write actions: 84 named actions (43 `create`, 32 `update`, 9 `delete`) covering tickets, users,
organizations, groups, macros, triggers and trigger categories, automations, views, ticket fields,
ticket forms and form statuses, custom statuses, custom objects (records, fields, access rules,
object triggers), IT asset management, brands, workspaces, saved searches, bookmarks, approval
requests, task lists and templates, group/organization memberships, deletion schedules, ticket
imports, attachments, and account/current-user settings. `writes.json` is the authoritative
per-action contract; `pm connectors inspect zendesk-support` and the generated
`docs/connectors/zendesk-support/SKILL.md` render every action's endpoint, required record fields,
and risk note.

Service API documentation: https://developer.zendesk.com/api-reference/ticketing/introduction/.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); OAuth2 access token. Sent as Authorization: Bearer
  <access_token>.
- `api_token` (optional, secret, string); Zendesk API token (Admin Center > Apps and integrations >
  APIs > Zendesk API). Sent via HTTP Basic as '<email>/token:<api_token>'. Requires email.
- `base_url` (required, string); format `uri`; Your Zendesk Support account root, e.g.
  https://acme.zendesk.com for subdomain 'acme'. The engine appends /api/v2 to every request; do not
  include /api/v2 yourself. Also usable as a base URL override for tests/proxies.
- `email` (optional, secret, string); Zendesk agent email address paired with api_token for
  API-token Basic auth (the '<email>/token' username half).

Secret fields are redacted in logs and write previews: `access_token`, `api_token`, `email`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token` when `{{ secrets.access_token }}`.
- HTTP Basic authentication using `secrets.email`, `secrets.api_token` when `{{ secrets.api_token
  }}`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/v2/groups`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page[after]`; next token from
`meta.after_cursor`; stop flag `meta.has_more`.

Pagination by stream: cursor: `tickets`, `users`, `organizations`, `groups`, `satisfaction_ratings`;
next_url: `deleted_tickets`, `account_attributes`, `attribute_definitions`, `brands`,
`custom_roles`, `schedules`, `sla_policies`, `tags`, `ticket_fields`, `ticket_forms`, `topics`,
`user_fields`, `automations`, `categories`, `sections`, `articles`, `group_memberships`, `macros`,
`organization_fields`, `organization_memberships`, `posts`, `ticket_activities`, `ticket_audits`,
`ticket_metric_events`, `ticket_events`, `ticket_skips`, `triggers`, `views`.

- `tickets`: GET `/api/v2/tickets` - records path `tickets`; query `page[size]`=`100`; cursor
  pagination; cursor parameter `page[after]`; next token from `meta.after_cursor`; stop flag
  `meta.has_more`.
- `users`: GET `/api/v2/users` - records path `users`; query `page[size]`=`100`; cursor pagination;
  cursor parameter `page[after]`; next token from `meta.after_cursor`; stop flag `meta.has_more`.
- `organizations`: GET `/api/v2/organizations` - records path `organizations`; query
  `page[size]`=`100`; cursor pagination; cursor parameter `page[after]`; next token from
  `meta.after_cursor`; stop flag `meta.has_more`.
- `groups`: GET `/api/v2/groups` - records path `groups`; query `page[size]`=`100`; cursor
  pagination; cursor parameter `page[after]`; next token from `meta.after_cursor`; stop flag
  `meta.has_more`.
- `satisfaction_ratings`: GET `/api/v2/satisfaction_ratings` - records path `satisfaction_ratings`;
  query `page[size]`=`100`; cursor pagination; cursor parameter `page[after]`; next token from
  `meta.after_cursor`; stop flag `meta.has_more`.
- `deleted_tickets`: GET `/api/v2/deleted_tickets` - records path `deleted_tickets`; query
  `page[size]`=`100`; follows a next-page URL from the response body; URL path `links.next`; next
  URLs stay on the configured API host.
- `account_attributes`: GET `/api/v2/routing/attributes` - records path `attributes`; query
  `per_page`=`100`; follows a next-page URL from the response body; URL path `next_page`; next URLs
  stay on the configured API host.
- `attribute_definitions`: GET `/api/v2/routing/attributes/definitions` - records path `attributes`;
  query `per_page`=`100`; follows a next-page URL from the response body; URL path `next_page`; next
  URLs stay on the configured API host.
- `brands`: GET `/api/v2/brands` - records path `brands`; query `page[size]`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host.
- `custom_roles`: GET `/api/v2/custom_roles` - records path `custom_roles`; query `per_page`=`100`;
  follows a next-page URL from the response body; URL path `next_page`; next URLs stay on the
  configured API host.
- `schedules`: GET `/api/v2/business_hours/schedules.json` - records path `schedules`; query
  `page[size]`=`100`; follows a next-page URL from the response body; URL path `next_page`; next
  URLs stay on the configured API host.
- `sla_policies`: GET `/api/v2/slas/policies.json` - records path `sla_policies`; query
  `per_page`=`100`; follows a next-page URL from the response body; URL path `next_page`; next URLs
  stay on the configured API host.
- `tags`: GET `/api/v2/tags` - records path `tags`; query `page[size]`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host.
- `ticket_fields`: GET `/api/v2/ticket_fields` - records path `ticket_fields`; query
  `page[size]`=`100`; follows a next-page URL from the response body; URL path `links.next`; next
  URLs stay on the configured API host.
- `ticket_forms`: GET `/api/v2/ticket_forms` - records path `ticket_forms`; query `per_page`=`100`;
  follows a next-page URL from the response body; URL path `next_page`; next URLs stay on the
  configured API host.
- `topics`: GET `/api/v2/community/topics` - records path `topics`; query `page[size]`=`100`;
  follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host.
- `user_fields`: GET `/api/v2/user_fields` - records path `user_fields`; query `per_page`=`100`;
  follows a next-page URL from the response body; URL path `next_page`; next URLs stay on the
  configured API host.
- `automations`: GET `/api/v2/automations` - records path `automations`; query `page[size]`=`100`;
  follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host.
- `categories`: GET `/api/v2/help_center/categories` - records path `categories`; query
  `page[size]`=`100`; follows a next-page URL from the response body; URL path `links.next`; next
  URLs stay on the configured API host.
- `sections`: GET `/api/v2/help_center/sections` - records path `sections`; query
  `page[size]`=`100`; follows a next-page URL from the response body; URL path `links.next`; next
  URLs stay on the configured API host.
- `articles`: GET `/api/v2/help_center/incremental/articles` - records path `articles`; query
  `per_page`=`100`; follows a next-page URL from the response body; URL path `next_page`; next URLs
  stay on the configured API host.
- `group_memberships`: GET `/api/v2/group_memberships` - records path `group_memberships`; query
  `page[size]`=`100`; follows a next-page URL from the response body; URL path `links.next`; next
  URLs stay on the configured API host.
- `macros`: GET `/api/v2/macros` - records path `macros`; query `page[size]`=`100`;
  `sort_by`=`created_at`; `sort_order`=`asc`; follows a next-page URL from the response body; URL
  path `links.next`; next URLs stay on the configured API host.
- `organization_fields`: GET `/api/v2/organization_fields` - records path `organization_fields`;
  query `per_page`=`100`; follows a next-page URL from the response body; URL path `next_page`; next
  URLs stay on the configured API host.
- `organization_memberships`: GET `/api/v2/organization_memberships` - records path
  `organization_memberships`; query `page[size]`=`100`; follows a next-page URL from the response
  body; URL path `links.next`; next URLs stay on the configured API host.
- `posts`: GET `/api/v2/community/posts` - records path `posts`; query `page[size]`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host.
- `ticket_activities`: GET `/api/v2/activities` - records path `activities`; query
  `page[size]`=`100`; `sort`=`created_at`; `sort_by`=`created_at`; `sort_order`=`asc`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host.
- `ticket_audits`: GET `/api/v2/ticket_audits` - records path `audits`; query `limit`=`100`;
  `sort_by`=`created_at`; `sort_order`=`desc`; follows a next-page URL from the response body; URL
  path `before_url`; next URLs stay on the configured API host.
- `ticket_metric_events`: GET `/api/v2/incremental/ticket_metric_events` - records path
  `ticket_metric_events`; query `per_page`=`100`; follows a next-page URL from the response body;
  URL path `next_page`; next URLs stay on the configured API host.
- `ticket_events`: GET `/api/v2/incremental/ticket_events.json` - records path `ticket_events`;
  query `per_page`=`100`; follows a next-page URL from the response body; URL path `next_page`; next
  URLs stay on the configured API host.
- `ticket_skips`: GET `/api/v2/skips.json` - records path `skips`; query `page[size]`=`100`;
  `sort_order`=`desc`; follows a next-page URL from the response body; URL path `links.next`; next
  URLs stay on the configured API host.
- `triggers`: GET `/api/v2/triggers` - records path `triggers`; query `per_page`=`100`; follows a
  next-page URL from the response body; URL path `next_page`; next URLs stay on the configured API
  host.
- `views`: GET `/api/v2/views` - records path `views`; query `page[size]`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host.

## Write actions & risks

Overall write risk: allow-listed Zendesk Support mutations across the ticketing, people, routing,
custom-object, IT-asset-management, and account-configuration surfaces; destructive deletes require
approval plus typed destructive confirmation.

Reverse ETL writes must be planned, previewed, approved, and then executed. `writes.json` is the
authoritative per-action contract (endpoint, bounded record schema, required/accepted fields, path
fields, idempotency and confirmation notes); read it with `pm connectors inspect zendesk-support` or
in the generated `docs/connectors/zendesk-support/SKILL.md`. This file does not restate per-action
fields.

Risk classes:

- Ten `delete` actions (`delete_ticket`, `delete_user`, `delete_organization`, `delete_group`,
  `delete_macro`, `delete_trigger`, `delete_automation`, `delete_api_token`, `delete_view`, `delete_ticket_field`) carry
  the legacy `confirm: "destructive"` declaration normalized by the shared typed gate and treat
  status `404` as success. They are the only destructive actions. The shared gate makes the 88
  remaining `destructive_action` rows technically bindable, but they stay unbound pending
  connector-local typed action schemas, canonical command mappings, and fixtures;
  `reverse_etl_execute_test.go`'s `TestDestructiveOperationsStayBlocked` fails if that count moves
  without that authoring work.
- Bulk and import actions (`tickets_create_many`, `ticket_import`, `ticket_bulk_import`,
  `group_membership_bulk_create`, `update_many_macros`, `update_many_triggers`,
  `update_many_object_triggers`, `bulk_update_default_custom_status`,
  `bulk_recover_suspended_tickets`, `custom_object_record_bulk_jobs`, `itam_asset_bulk_jobs`,
  `batch_operate_trigger_categories`) mutate many provider records per request; their record schemas
  bound the array element shape rather than accepting a free-form payload.
- Account- and tenant-wide settings actions (`update_account_email_settings`,
  `update_current_user_settings`, `reorder_workspaces`) change configuration for the whole Zendesk
  account rather than a single record.
- The 27 pre-existing `writes <action> plan` commands stay at `availability: "partial"` in
  `cli_surface.json`: this foundation supplies the shared gate but deliberately does not promote or
  bind connector commands; that remains connector-authoring work.

## Known limits

- Official ledger source: Zendesk Support OAS 2.0.0 from `https://developer.zendesk.com/zendesk/oas.yaml`. The 2026-08-01 parity checkpoint re-fetched the OpenAPI 3.0.3 document (`info.version` 2.0.0) and counted 434 paths / 625 unique operations: GET 325, POST 111, PUT 89, PATCH 14, DELETE 86. `api_surface.json` inventories all 625 official operations exactly once with 0 missing, 0 stale, and 0 duplicate official endpoint keys, plus 6 supplemental `covered_by` row(s) for existing fixture-backed bundle surfaces that are not present in that Support OAS.
- Executable fixture-backed surfaces are 33 streams and 84 reverse-ETL write actions (117 `covered_by` rows). `operations.json` and planned `cli_surface.json` entries are typed, connector-owned metadata only; they do not enable a raw API escape hatch or claim live certification.
- Every write action's record schema is derived from the pinned OpenAPI source above, never inferred from response schemas. 105 mutation rows are therefore still blocked on purpose rather than shipped as guessed contracts: 98 because the pinned source declares no request body at all, and 7 because it declares an unbounded or bulk free-form payload region that a closed record schema cannot represent usefully from the command surface. `reverse_etl_execute_test.go` pins that arithmetic (57 promoted + 98 + 7 = the 162 rows originally carrying the blocked reverse-ETL reason), so rewording a blocked reason cannot hide the shortfall.
- Envelope and resource levels of each record schema are closed with `additionalProperties: false`; deeply nested provider-defined regions stay `type: object` with `additionalProperties: true`, which is the bundle's bounded-but-not-exhaustive convention.
- Destructive and `DELETE` operations are in scope. Existing delete write actions use `confirm: "destructive"` and idempotent 404 handling. Remaining destructive official operations are blocked by default until a connector-local typed action declares schema, risk, redaction/idempotency where applicable, and the plan -> preview -> explicit approval -> execute path with typed destructive confirmation.
- Credential, password, and token operation rows are separately blocked as sensitive reverse ETL metadata with non-inline input where required, credential redaction, and typed sensitive confirmation before any future execution; credential-returning direct reads stay blocked until a bounded redacted output policy is reviewed.
- Direct/provider-search/query operations are blocked pending the bounded provider command foundation (#2985); CDC/changefeed-style rows are blocked pending CDC truthfulness/state foundations (#2986/#2988); binary downloads and file-upload rows are blocked pending bounded binary/file executor support. File-upload operation rows retain connector-local upload direction, path, and `max_bytes` bounds; the source-backed method, required query parameters, content types, and multipart/body fields are recorded in operation descriptions until a shared file-operation schema/executor can validate them without exposing inline raw bytes.
- The current draft-07 subset cannot express the `access_token` OR (`api_token` AND `email`) credential pairing. The schema documents the required pairing, and invalid combinations fail at the credential-free check request rather than reading or printing secret values.
- Fixture-only evidence remains uncertified (`certified=0`). This bundle does not run live Zendesk credentials, live provider calls, live writes, VPS/Thaalam work, or release certification.
