# Overview

Zendesk Support reads 33 stream(s), and writes through 27 action(s).

Readable streams: `tickets`, `users`, `organizations`, `groups`, `satisfaction_ratings`,
`deleted_tickets`, `account_attributes`, `attribute_definitions`, `brands`, `custom_roles`,
`schedules`, `sla_policies`, `tags`, `ticket_fields`, `ticket_forms`, `topics`, `user_fields`,
`automations`, `categories`, `sections`, `articles`, `group_memberships`, `macros`,
`organization_fields`, `organization_memberships`, `posts`, `ticket_activities`, `ticket_audits`,
`ticket_metric_events`, `ticket_events`, `ticket_skips`, `triggers`, `views`.

Write actions: `create_ticket`, `update_ticket`, `delete_ticket`, `create_user`, `update_user`,
`delete_user`, `create_organization`, `update_organization`, `delete_organization`, `create_group`,
`update_group`, `delete_group`, `create_macro`, `update_macro`, `delete_macro`, `create_trigger`,
`update_trigger`, `delete_trigger`, `create_automation`, `update_automation`, `delete_automation`,
`create_view`, `update_view`, `delete_view`, `create_ticket_field`, `update_ticket_field`,
`delete_ticket_field`.

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

Overall write risk: allow-listed Zendesk Support mutations for tickets, users, organizations,
groups, macros, triggers, automations, views, and ticket fields; destructive deletes require
approval.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_ticket`: POST `/api/v2/tickets` - kind `create`; body type `json`; body fields `ticket`;
  required record fields `ticket`; accepted fields `ticket`; risk: creates a Zendesk ticket record;
  external mutation requiring approval.
- `update_ticket`: PUT `/api/v2/tickets/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; body fields `ticket`; required record fields `id`, `ticket`; accepted fields `id`,
  `ticket`; risk: updates a Zendesk ticket record; external mutation requiring approval.
- `delete_ticket`: DELETE `/api/v2/tickets/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: deletes a Zendesk ticket record; destructive
  external mutation requiring approval.
- `create_user`: POST `/api/v2/users` - kind `create`; body type `json`; body fields `user`;
  required record fields `user`; accepted fields `user`; risk: creates a Zendesk user record;
  external mutation requiring approval.
- `update_user`: PUT `/api/v2/users/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; body fields `user`; required record fields `id`, `user`; accepted fields `id`, `user`; risk:
  updates a Zendesk user record; external mutation requiring approval.
- `delete_user`: DELETE `/api/v2/users/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: deletes a Zendesk user record; destructive
  external mutation requiring approval.
- `create_organization`: POST `/api/v2/organizations` - kind `create`; body type `json`; body fields
  `organization`; required record fields `organization`; accepted fields `organization`; risk:
  creates a Zendesk organization record; external mutation requiring approval.
- `update_organization`: PUT `/api/v2/organizations/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; body fields `organization`; required record fields `id`, `organization`;
  accepted fields `id`, `organization`; risk: updates a Zendesk organization record; external
  mutation requiring approval.
- `delete_organization`: DELETE `/api/v2/organizations/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: deletes a Zendesk
  organization record; destructive external mutation requiring approval.
- `create_group`: POST `/api/v2/groups` - kind `create`; body type `json`; body fields `group`;
  required record fields `group`; accepted fields `group`; risk: creates a Zendesk group record;
  external mutation requiring approval.
- `update_group`: PUT `/api/v2/groups/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; body fields `group`; required record fields `id`, `group`; accepted fields `group`,
  `id`; risk: updates a Zendesk group record; external mutation requiring approval.
- `delete_group`: DELETE `/api/v2/groups/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: deletes a Zendesk group record; destructive
  external mutation requiring approval.
- `create_macro`: POST `/api/v2/macros` - kind `create`; body type `json`; body fields `macro`;
  required record fields `macro`; accepted fields `macro`; risk: creates a Zendesk macro record;
  external mutation requiring approval.
- `update_macro`: PUT `/api/v2/macros/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; body fields `macro`; required record fields `id`, `macro`; accepted fields `id`,
  `macro`; risk: updates a Zendesk macro record; external mutation requiring approval.
- `delete_macro`: DELETE `/api/v2/macros/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: deletes a Zendesk macro record; destructive
  external mutation requiring approval.
- `create_trigger`: POST `/api/v2/triggers` - kind `create`; body type `json`; body fields
  `trigger`; required record fields `trigger`; accepted fields `trigger`; risk: creates a Zendesk
  trigger record; external mutation requiring approval.
- `update_trigger`: PUT `/api/v2/triggers/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; body fields `trigger`; required record fields `id`, `trigger`; accepted fields `id`,
  `trigger`; risk: updates a Zendesk trigger record; external mutation requiring approval.
- `delete_trigger`: DELETE `/api/v2/triggers/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: deletes a Zendesk trigger record;
  destructive external mutation requiring approval.
- `create_automation`: POST `/api/v2/automations` - kind `create`; body type `json`; body fields
  `automation`; required record fields `automation`; accepted fields `automation`; risk: creates a
  Zendesk automation record; external mutation requiring approval.
- `update_automation`: PUT `/api/v2/automations/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; body fields `automation`; required record fields `id`, `automation`; accepted
  fields `automation`, `id`; risk: updates a Zendesk automation record; external mutation requiring
  approval.
- `delete_automation`: DELETE `/api/v2/automations/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: deletes a Zendesk
  automation record; destructive external mutation requiring approval.
- `create_view`: POST `/api/v2/views` - kind `create`; body type `json`; body fields `view`;
  required record fields `view`; accepted fields `view`; risk: creates a Zendesk view record;
  external mutation requiring approval.
- `update_view`: PUT `/api/v2/views/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; body fields `view`; required record fields `id`, `view`; accepted fields `id`, `view`; risk:
  updates a Zendesk view record; external mutation requiring approval.
- `delete_view`: DELETE `/api/v2/views/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: deletes a Zendesk view record; destructive
  external mutation requiring approval.
- `create_ticket_field`: POST `/api/v2/ticket_fields` - kind `create`; body type `json`; body fields
  `ticket_field`; required record fields `ticket_field`; accepted fields `ticket_field`; risk:
  creates a Zendesk ticket field record; external mutation requiring approval.
- `update_ticket_field`: PUT `/api/v2/ticket_fields/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; body fields `ticket_field`; required record fields `id`, `ticket_field`;
  accepted fields `id`, `ticket_field`; risk: updates a Zendesk ticket field record; external
  mutation requiring approval.
- `delete_ticket_field`: DELETE `/api/v2/ticket_fields/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: deletes a Zendesk ticket
  field record; destructive external mutation requiring approval.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 33 stream-backed endpoint group(s), 27 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=1, duplicate_of=3, out_of_scope=11.
