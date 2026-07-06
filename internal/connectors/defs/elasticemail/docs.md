# Overview

Reads and writes Elastic Email contacts, campaigns, lists, segments, templates, webhooks, domains,
inbound routes, suppressions, and account statistics through the Elastic Email v4 REST API.

Readable streams: `contacts`, `campaigns`, `lists`, `segments`, `templates`, `domains`,
`suppressions`, `suppressions_bounces`, `suppressions_complaints`, `suppressions_unsubscribes`,
`webhooks`, `files`, `inbound_routes`, `sub_accounts`, `statistics_campaigns`,
`statistics_channels`.

Write actions: `create_contact`, `update_contact`, `delete_contact`, `create_list`, `update_list`,
`delete_list`, `add_list_contacts`, `create_segment`, `update_segment`, `delete_segment`,
`create_template`, `update_template`, `delete_template`, `create_campaign`, `update_campaign`,
`pause_campaign`, `delete_campaign`, `create_webhook`, `update_webhook`, `delete_webhook`,
`create_domain`, `delete_domain`, `create_inbound_route`, `update_inbound_route`,
`delete_inbound_route`.

Service API documentation: https://elasticemail.com/developers/api-documentation/.

## Auth setup

Connection fields:

- `api_key` (optional, secret, string); Elastic Email API key, sent as the X-ElasticEmail-ApiKey
  header. Never logged.
- `base_url` (optional, string); default `https://api.elasticemail.com/v4`; format `uri`; Elastic
  Email v4 API base URL. Defaults to the production endpoint; override for test proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.elasticemail.com/v4`.

Authentication behavior:

- API key authentication in `X-ElasticEmail-ApiKey` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/lists`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

Pagination by stream: none: `domains`, `inbound_routes`; offset_limit: `contacts`, `campaigns`,
`lists`, `segments`, `templates`, `suppressions`, `suppressions_bounces`, `suppressions_complaints`,
`suppressions_unsubscribes`, `webhooks`, `files`, `sub_accounts`, `statistics_campaigns`,
`statistics_channels`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `contacts`: GET `/contacts` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; incremental cursor `DateUpdated`; formatted as
  `rfc3339`.
- `campaigns`: GET `/campaigns` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `lists`: GET `/lists` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `segments`: GET `/segments` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `templates`: GET `/templates` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `domains`: GET `/domains` - records at response root.
- `suppressions`: GET `/suppressions` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `suppressions_bounces`: GET `/suppressions/bounces` - records at response root; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `suppressions_complaints`: GET `/suppressions/complaints` - records at response root; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `suppressions_unsubscribes`: GET `/suppressions/unsubscribes` - records at response root;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `webhooks`: GET `/webhook` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `files`: GET `/files` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `inbound_routes`: GET `/inboundroute` - records at response root.
- `sub_accounts`: GET `/subaccounts` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `statistics_campaigns`: GET `/statistics/campaigns` - records at response root; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `statistics_channels`: GET `/statistics/channels` - records at response root; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100.

## Write actions & risks

Overall write risk: external Elastic Email API mutations covering
contact/list/segment/template/campaign/webhook/domain/inbound-route lifecycle management;
create_campaign and pause_campaign can affect a live email send to real recipients, and
webhook/inbound-route writes register caller-controlled external destinations for live event/mail
forwarding.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_contact`: POST `/contacts` - kind `create`; body type `json`; required record fields
  `Email`; accepted fields `Consent`, `CustomFields`, `Email`, `FirstName`, `LastName`, `Status`;
  risk: adds a new contact to the account's overall recipient list; low-risk external mutation, no
  approval required.
- `update_contact`: PUT `/contacts/{{ record.Email }}` - kind `update`; body type `json`; path
  fields `Email`; required record fields `Email`; accepted fields `Consent`, `CustomFields`,
  `Email`, `FirstName`, `LastName`, `Status`; risk: mutates an existing contact's
  status/name/custom-field data; a Status change (e.g. to Unsubscribed) changes future campaign
  eligibility for this recipient.
- `delete_contact`: DELETE `/contacts/{{ record.Email }}` - kind `delete`; body type `none`; path
  fields `Email`; required record fields `Email`; accepted fields `Email`; missing records treated
  as success for status `404`; risk: permanently removes a contact and its activity/consent history
  from the account.
- `create_list`: POST `/lists` - kind `create`; body type `json`; required record fields `ListName`;
  accepted fields `AllowUnsubscribe`, `Emails`, `ListName`; risk: creates a new contact list,
  optionally seeding it from existing contact emails; low-risk external mutation, no approval
  required.
- `update_list`: PUT `/lists/{{ record.ListName }}` - kind `update`; body type `json`; path fields
  `ListName`; body fields `NewListName`, `AllowUnsubscribe`; required record fields `ListName`;
  accepted fields `AllowUnsubscribe`, `ListName`, `NewListName`; risk: renames an existing list or
  changes its unsubscribe-allowed setting; a rename changes the identifier campaigns/segments
  reference this list by.
- `delete_list`: DELETE `/lists/{{ record.ListName }}` - kind `delete`; body type `none`; path
  fields `ListName`; required record fields `ListName`; accepted fields `ListName`; missing records
  treated as success for status `404`; risk: permanently removes a contact list; any campaign still
  targeting this list by name will fail to resolve its recipients.
- `add_list_contacts`: POST `/lists/{{ record.ListName }}/contacts` - kind `create`; body type
  `json`; path fields `ListName`; body fields `Emails`, `Status`; required record fields `ListName`,
  `Emails`; accepted fields `Emails`, `ListName`, `Status`; risk: adds existing contacts to a list,
  making them eligible recipients for any campaign targeting that list.
- `create_segment`: POST `/segments` - kind `create`; body type `json`; required record fields
  `Name`, `Rule`; accepted fields `Name`, `Rule`; risk: creates a new dynamic contact segment from a
  SQL-like rule; low-risk external mutation, no approval required.
- `update_segment`: PUT `/segments/{{ record.Name }}` - kind `update`; body type `json`; path fields
  `Name`; required record fields `Name`, `Rule`; accepted fields `Name`, `Rule`; risk: changes the
  membership rule of an existing segment; immediately changes which contacts any campaign targeting
  this segment will reach.
- `delete_segment`: DELETE `/segments/{{ record.Name }}` - kind `delete`; body type `none`; path
  fields `Name`; required record fields `Name`; accepted fields `Name`; missing records treated as
  success for status `404`; risk: permanently removes a segment; any campaign still targeting this
  segment by name will fail to resolve its recipients.
- `create_template`: POST `/templates` - kind `create`; body type `json`; required record fields
  `Name`; accepted fields `Body`, `Name`, `Subject`, `TemplateScope`; risk: creates a new email
  template; low-risk external mutation, no approval required.
- `update_template`: PUT `/templates/{{ record.Name }}` - kind `update`; body type `json`; path
  fields `Name`; required record fields `Name`; accepted fields `Body`, `Name`, `Subject`,
  `TemplateScope`; risk: overwrites the subject/body of an existing template; any campaign
  referencing this template by name sends the new content on its next send.
- `delete_template`: DELETE `/templates/{{ record.Name }}` - kind `delete`; body type `none`; path
  fields `Name`; required record fields `Name`; accepted fields `Name`; missing records treated as
  success for status `404`; risk: permanently removes a template; any campaign still referencing
  this template by name will fail to build its content.
- `create_campaign`: POST `/campaigns` - kind `create`; body type `json`; required record fields
  `Name`, `Recipients`; accepted fields `Content`, `Name`, `Options`, `Recipients`; risk: creates a
  new campaign targeting the given lists/segments; depending on Options this may schedule a live
  send to real recipients, not a preview-only action.
- `update_campaign`: PUT `/campaigns/{{ record.Name }}` - kind `update`; body type `json`; path
  fields `Name`; required record fields `Name`; accepted fields `Content`, `Name`, `Options`,
  `Recipients`; risk: mutates an existing campaign's content, recipients, or send options; a
  campaign already in progress may not accept every field change.
- `pause_campaign`: PUT `/campaigns/{{ record.Name }}/pause` - kind `update`; body type `none`; path
  fields `Name`; required record fields `Name`; accepted fields `Name`; risk: pauses an in-progress
  campaign send; recipients not yet reached will not receive the email until the campaign is
  resumed.
- `delete_campaign`: DELETE `/campaigns/{{ record.Name }}` - kind `delete`; body type `none`; path
  fields `Name`; required record fields `Name`; accepted fields `Name`; missing records treated as
  success for status `404`; risk: permanently removes a campaign; if it has not finished sending,
  any remaining scheduled deliveries are cancelled.
- `create_webhook`: POST `/webhook` - kind `create`; body type `json`; required record fields
  `Name`, `URL`; accepted fields `Name`, `NotificationForAbuseReport`, `NotificationForClicked`,
  `NotificationForError`, `NotificationForOpened`, `NotificationForSent`,
  `NotificationForUnsubscribed`, `NotifyOncePerEmail`, `URL`; risk: registers a new outbound webhook
  that will POST live event data (sent/opened/clicked/bounced) to an external URL of the caller's
  choosing; verify the target endpoint before enabling.
- `update_webhook`: PUT `/webhook/{{ record.WebhookID }}` - kind `update`; body type `json`; path
  fields `WebhookID`; required record fields `WebhookID`; accepted fields
  `NotificationForAbuseReport`, `NotificationForClicked`, `NotificationForError`,
  `NotificationForOpened`, `NotificationForSent`, `NotificationForUnsubscribed`,
  `NotifyOncePerEmail`, `URL`, `WebhookID`; risk: mutates an existing webhook's target URL or event
  subscriptions; a changed URL redirects future event deliveries to a different endpoint.
- `delete_webhook`: DELETE `/webhook/{{ record.WebhookID }}` - kind `delete`; body type `none`; path
  fields `WebhookID`; required record fields `WebhookID`; accepted fields `WebhookID`; missing
  records treated as success for status `404`; risk: permanently removes a webhook subscription;
  event delivery to its target URL stops immediately.
- `create_domain`: POST `/domains` - kind `create`; body type `json`; required record fields
  `Domain`; accepted fields `Domain`, `SetAsDefault`; risk: registers a new sending domain pending
  DNS verification; low-risk external mutation, no approval required.
- `delete_domain`: DELETE `/domains/{{ record.Domain }}` - kind `delete`; body type `none`; path
  fields `Domain`; required record fields `Domain`; accepted fields `Domain`; missing records
  treated as success for status `404`; risk: permanently removes a verified sending domain; any
  campaign configured to send from this domain will fail until reconfigured.
- `create_inbound_route`: POST `/inboundroute` - kind `create`; body type `json`; required record
  fields `Name`, `Filter`, `FilterType`, `ActionType`; accepted fields `ActionType`, `EmailAddress`,
  `Filter`, `FilterType`, `HttpAddress`, `Name`; risk: creates a new inbound-mail routing rule that
  forwards matching inbound email to an external address or webhook URL of the caller's choosing.
- `update_inbound_route`: PUT `/inboundroute/{{ record.PublicId }}` - kind `update`; body type
  `json`; path fields `PublicId`; required record fields `PublicId`; accepted fields `ActionType`,
  `EmailAddress`, `Filter`, `FilterType`, `HttpAddress`, `Name`, `PublicId`; risk: mutates an
  existing inbound route's match filter or forwarding destination; redirects future matching inbound
  mail to a different address/URL.
- `delete_inbound_route`: DELETE `/inboundroute/{{ record.PublicId }}` - kind `delete`; body type
  `none`; path fields `PublicId`; required record fields `PublicId`; accepted fields `PublicId`;
  missing records treated as success for status `404`; risk: permanently removes an inbound-mail
  routing rule; matching inbound mail is no longer forwarded once removed.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 16 stream-backed endpoint group(s), 25 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=5, destructive_admin=3, duplicate_of=17, non_data_endpoint=3, out_of_scope=26,
  requires_elevated_scope=17.
