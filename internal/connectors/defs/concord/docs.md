# Overview

Reads and writes Concord contract lifecycle management data: agreements (and their
metadata/summary/comments/activities/members/versions/attachments sub-resources), organizations,
folders, reports, tags, clauses, approvals, groups, members, events, subscription, branding, and
automated templates through the Concord REST API.

Readable streams: `agreements`, `user_organizations`, `folders`, `reports`, `tags`, `organization`,
`folder`, `folder_agreements`, `report`, `clauses`, `clause`, `approvals`, `approval`, `groups`,
`members`, `events`, `subscription`, `branding`, `automated_templates`, `user_me`,
`user_preferences`, `webhooks_integrations`, `agreement`, `agreement_metadata`, `agreement_summary`,
`agreement_comments`, `agreement_activities`, `agreement_members`, `agreement_versions`,
`agreement_attachments`.

Write actions: `create_folder`, `update_folder`, `delete_folder`, `create_report`, `update_report`,
`delete_report`, `create_clause`, `update_clause`, `delete_clause`, `create_group`,
`create_approval`, `update_approval`, `delete_approval`.

Service API documentation: https://api.concordnow.com/api/rest/1/docs.

## Auth setup

Connection fields:

- `agreement_uid` (optional, string); Concord agreement uid; required for the agreement detail
  stream and its sub-resource streams (agreement_metadata, agreement_summary, agreement_comments,
  agreement_activities, agreement_members, agreement_versions, agreement_attachments) when read
  standalone rather than via fan_out from the agreements stream.
- `api_key` (required, secret, string); Concord API key, sent as the X-API-KEY header. Never logged.
- `approval_id` (optional, string); Concord approval id; required for the approval detail stream.
- `base_url` (optional, string); default `https://api.concordnow.com/api/rest/1`; format `uri`;
  Concord API base URL. Defaults to the production environment; override with
  https://uat.concordnow.com/api/rest/1 for UAT, or a test/proxy URL.
- `clause_id` (optional, string); Concord clause id; required for the clause detail stream.
- `clauses_page_size` (optional, string); default `100`; Records per page (limit query param) for
  the clauses stream's offset/limit pagination.
- `events_end_date` (optional, string); Events stream date-range end (yyyy-MM-dd); required for the
  events stream. Concord requires start/end to span at most 7 days.
- `events_start_date` (optional, string); Events stream date-range start (yyyy-MM-dd); required for
  the events stream. Concord requires start/end to span at most 7 days.
- `folder_id` (optional, string); Concord folder id; required for the folder and folder_agreements
  detail streams.
- `mode` (optional, string).
- `organization_id` (optional, string); Concord organization id; required for every
  organization-scoped stream (agreements, folders, reports, organization, clauses, approvals,
  groups, members, events, subscription, branding, automated_templates, folder, folder_agreements,
  report, clause, approval, agreement and its sub-resource streams).
- `page_size` (optional, string); default `100`; Records per page (limit query param; 1-1000).
- `report_id` (optional, string); Concord report id (or sample report enum value); required for the
  report detail stream.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.concordnow.com/api/rest/1`,
`clauses_page_size=100`, `page_size=100`.

Authentication behavior:

- API key authentication in `X-API-KEY` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/user/me/organizations`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; no page-size parameter; starts at
0; page size 100.

Pagination by stream: none: `organization`, `folder`, `folder_agreements`, `report`, `clause`,
`approvals`, `approval`, `groups`, `events`, `subscription`, `branding`, `automated_templates`,
`user_me`, `user_preferences`, `webhooks_integrations`, `agreement`, `agreement_metadata`,
`agreement_summary`, `agreement_comments`, `agreement_activities`, `agreement_members`,
`agreement_versions`, `agreement_attachments`; offset_limit: `clauses`, `members`; page_number:
`agreements`, `user_organizations`, `folders`, `reports`, `tags`.

- `agreements`: GET `/organizations/{{ config.organization_id }}/agreements` - records at response
  root; query `limit` from template `{{ config.page_size }}`, default `100`; page-number pagination;
  page parameter `page`; no page-size parameter; starts at 0; page size 100.
- `user_organizations`: GET `/user/me/organizations` - records path `organizations`; query `limit`
  from template `{{ config.page_size }}`, default `100`; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 0; page size 100.
- `folders`: GET `/organizations/{{ config.organization_id }}/folders` - records at response root;
  query `limit` from template `{{ config.page_size }}`, default `100`; page-number pagination; page
  parameter `page`; no page-size parameter; starts at 0; page size 100.
- `reports`: GET `/organizations/{{ config.organization_id }}/reports` - records path `reports`;
  query `limit` from template `{{ config.page_size }}`, default `100`; page-number pagination; page
  parameter `page`; no page-size parameter; starts at 0; page size 100.
- `tags`: GET `/tags` - records path `tags`; query `limit` from template `{{ config.page_size }}`,
  default `100`; page-number pagination; page parameter `page`; no page-size parameter; starts at 0;
  page size 100.
- `organization`: GET `/organizations/{{ config.organization_id }}` - records at response root.
- `folder`: GET `/organizations/{{ config.organization_id }}/folders/{{ config.folder_id }}` -
  records at response root.
- `folder_agreements`: GET `/organizations/{{ config.organization_id }}/folders/{{ config.folder_id
  }}/agreements` - records at response root.
- `report`: GET `/organizations/{{ config.organization_id }}/reports/{{ config.report_id }}` -
  records at response root.
- `clauses`: GET `/organizations/{{ config.organization_id }}/clauses` - records path
  `organizationClauses`; query `limit` from template `{{ config.clauses_page_size }}`, default
  `100`; offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `clause`: GET `/organizations/{{ config.organization_id }}/clauses/{{ config.clause_id }}` -
  records at response root.
- `approvals`: GET `/organizations/{{ config.organization_id }}/approvals` - records path
  `approvals`.
- `approval`: GET `/organizations/{{ config.organization_id }}/approvals/{{ config.approval_id }}` -
  records at response root.
- `groups`: GET `/organizations/{{ config.organization_id }}/groups` - records path `groups`.
- `members`: GET `/organizations/{{ config.organization_id }}/members` - records path `members`;
  offset/limit pagination; offset parameter `start`; limit parameter `limit`; page size 100.
- `events`: GET `/organizations/{{ config.organization_id }}/events` - records path `events`; query
  `end`=`{{ config.events_end_date }}`; `start`=`{{ config.events_start_date }}`.
- `subscription`: GET `/organizations/{{ config.organization_id }}/subscription` - records at
  response root.
- `branding`: GET `/organizations/{{ config.organization_id }}/branding` - records at response root.
- `automated_templates`: GET `/organizations/{{ config.organization_id }}/auto` - records at
  response root.
- `user_me`: GET `/user/me` - records at response root.
- `user_preferences`: GET `/user/me/preferences` - records at response root.
- `webhooks_integrations`: GET `/users/me/integrations/webhooks` - records at response root.
- `agreement`: GET `/organizations/{{ config.organization_id }}/agreements/{{ config.agreement_uid
  }}` - records at response root.
- `agreement_metadata`: GET `/organizations/{{ config.organization_id }}/agreements/{{ fanout.id
  }}/metadata` - records at response root; fan-out; ids from request `/organizations/{{
  config.organization_id }}/agreements`; id field `uid`; id inserted into the request path; stamps
  `agreement_uid`.
- `agreement_summary`: GET `/organizations/{{ config.organization_id }}/agreements/{{ fanout.id
  }}/summary` - records at response root; fan-out; ids from request `/organizations/{{
  config.organization_id }}/agreements`; id field `uid`; id inserted into the request path; stamps
  `agreement_uid`.
- `agreement_comments`: GET `/organizations/{{ config.organization_id }}/agreements/{{ fanout.id
  }}/comments` - records at response root; flattens keyed objects; key field `comment_uuid`;
  fan-out; ids from request `/organizations/{{ config.organization_id }}/agreements`; id field
  `uid`; id inserted into the request path; stamps `agreement_id`.
- `agreement_activities`: GET `/organizations/{{ config.organization_id }}/agreements/{{ fanout.id
  }}/activities` - records path `activities`; fan-out; ids from request `/organizations/{{
  config.organization_id }}/agreements`; id field `uid`; id inserted into the request path; stamps
  `agreement_id`.
- `agreement_members`: GET `/organizations/{{ config.organization_id }}/agreements/{{ fanout.id
  }}/members` - records at response root; computed output fields `member_id`; fan-out; ids from
  request `/organizations/{{ config.organization_id }}/agreements`; id field `uid`; id inserted into
  the request path; stamps `agreement_id`.
- `agreement_versions`: GET `/organizations/{{ config.organization_id }}/agreements/{{ fanout.id
  }}/versions` - records at response root; fan-out; ids from request `/organizations/{{
  config.organization_id }}/agreements`; id field `uid`; id inserted into the request path; stamps
  `agreement_id`.
- `agreement_attachments`: GET `/organizations/{{ config.organization_id }}/agreements/{{ fanout.id
  }}/attachments` - records path `attachments`; fan-out; ids from request `/organizations/{{
  config.organization_id }}/agreements`; id field `uid`; id inserted into the request path; stamps
  `agreement_id`.

## Write actions & risks

Overall write risk: external mutation of Concord folders, reports, clauses, groups, and company
approval workflows (create/update/delete); does not create, sign, or modify agreements themselves.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_folder`: POST `/organizations/{{ config.organization_id }}/folders` - kind `create`; body
  type `json`; required record fields `name`, `parentId`; accepted fields `name`, `parentId`; risk:
  creates a new Concord folder within the configured organization; low risk, no data destruction.
- `update_folder`: PUT `/organizations/{{ config.organization_id }}/folders/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`,
  `name`, `parentId`; risk: renames/moves an existing Concord folder; may change document
  organization visible to other users.
- `delete_folder`: DELETE `/organizations/{{ config.organization_id }}/folders/{{ record.id }}` -
  kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; missing records treated as success for status `404`; risk: permanently deletes a Concord
  folder; destructive, external mutation; approval required.
- `create_report`: POST `/organizations/{{ config.organization_id }}/reports` - kind `create`; body
  type `json`; accepted fields `description`, `filters`, `name`, `sampleId`; risk: creates a new
  saved Concord report within the configured organization; low risk.
- `update_report`: PUT `/organizations/{{ config.organization_id }}/reports/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`, `name`, `description`,
  `filters`; accepted fields `description`, `filters`, `id`, `name`; risk: replaces an existing
  Concord saved report's definition; may change what other users see when they run it.
- `delete_report`: DELETE `/organizations/{{ config.organization_id }}/reports/{{ record.id }}` -
  kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; missing records treated as success for status `404`; risk: permanently deletes a Concord
  saved report; destructive, external mutation; approval required.
- `create_clause`: POST `/organizations/{{ config.organization_id }}/clauses` - kind `create`; body
  type `json`; required record fields `title`, `content`; accepted fields `content`, `description`,
  `title`; risk: creates a new reusable Concord clause template within the configured organization;
  low risk.
- `update_clause`: PUT `/organizations/{{ config.organization_id }}/clauses/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`, `title`, `content`;
  accepted fields `content`, `description`, `id`, `title`, `version`; risk: updates an existing
  Concord clause template; may affect future agreements linked to this clause.
- `delete_clause`: DELETE `/organizations/{{ config.organization_id }}/clauses/{{ record.id }}` -
  kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; missing records treated as success for status `404`; risk: permanently deletes a Concord
  clause template; destructive, external mutation; approval required.
- `create_group`: POST `/organizations/{{ config.organization_id }}/groups` - kind `create`; body
  type `json`; required record fields `name`; accepted fields `description`, `name`; risk: creates a
  new Concord user group within the configured organization; low risk.
- `create_approval`: POST `/organizations/{{ config.organization_id }}/approvals` - kind `create`;
  body type `json`; required record fields `title`, `description`, `blockThirdPartySignature`;
  accepted fields `blockThirdPartySignature`, `description`, `rules`, `title`; risk: creates a new
  Concord company approval workflow within the configured organization; affects future agreement
  signature routing.
- `update_approval`: POST `/organizations/{{ config.organization_id }}/approvals/{{ record.id }}` -
  kind `update`; body type `json`; path fields `id`; required record fields `id`, `title`,
  `description`, `blockThirdPartySignature`; accepted fields `blockThirdPartySignature`,
  `description`, `id`, `rules`, `title`; risk: replaces an existing Concord company approval
  workflow; affects agreements already routed through it.
- `delete_approval`: DELETE `/organizations/{{ config.organization_id }}/approvals/{{ record.id }}`
  - kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; missing records treated as success for status `404`; risk: permanently deletes a Concord
  company approval workflow; destructive, external mutation; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 30 stream-backed endpoint group(s), 13 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=7, destructive_admin=2, duplicate_of=8, out_of_scope=50, requires_elevated_scope=8.
