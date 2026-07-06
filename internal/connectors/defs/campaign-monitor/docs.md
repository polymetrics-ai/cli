# Overview

Reads and writes Campaign Monitor clients, campaigns, subscriber lists, subscribers, segments, and
templates through the createsend.com v3.3 REST API.

Readable streams: `clients`, `campaigns`, `lists`, `suppressionlist`, `segments`, `templates`,
`list_custom_fields`, `list_webhooks`, `active_subscribers`, `unconfirmed_subscribers`,
`unsubscribed_subscribers`, `bounced_subscribers`, `deleted_subscribers`.

Write actions: `create_list`, `update_list`, `delete_list`, `add_subscriber`, `update_subscriber`,
`unsubscribe_subscriber`, `delete_subscriber`, `create_segment`, `update_segment`, `delete_segment`,
`create_campaign`, `send_campaign`, `unschedule_campaign`, `delete_campaign`.

Service API documentation: https://www.campaignmonitor.com/api/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.createsend.com/api/v3.3`; format `uri`;
  Campaign Monitor API base URL override for tests or proxies.
- `client_id` (optional, string); Campaign Monitor client id, required for every client-scoped
  stream ('campaigns', 'lists', 'suppressionlist', 'segments', 'templates') and every list-scoped
  fan-out stream ('list_custom_fields', 'list_webhooks', 'active_subscribers',
  'unconfirmed_subscribers', 'unsubscribed_subscribers', 'bounced_subscribers',
  'deleted_subscribers'), which fan out over this client's own lists.
- `password` (optional, secret, string); HTTP Basic auth password. Campaign Monitor allows this to
  be blank or a dummy value alongside the API key username; never logged.
- `username` (required, string); Campaign Monitor API key, sent as the HTTP Basic auth username.

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `base_url=https://api.createsend.com/api/v3.3`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password` when `{{ secrets.password
  }}`.
- HTTP Basic authentication using `config.username`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/clients.json`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `pagesize`; starts
at 1; page size 100.

Pagination by stream: none: `clients`, `lists`, `segments`, `templates`, `list_custom_fields`,
`list_webhooks`; page_number: `campaigns`, `suppressionlist`, `active_subscribers`,
`unconfirmed_subscribers`, `unsubscribed_subscribers`, `bounced_subscribers`, `deleted_subscribers`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `clients`: GET `/clients.json` - records at response root.
- `campaigns`: GET `/clients/{{ config.client_id }}/campaigns.json` - records path `Results`;
  page-number pagination; page parameter `page`; size parameter `pagesize`; starts at 1; page size
  100; incremental cursor `SentDate`; formatted as `rfc3339`.
- `lists`: GET `/clients/{{ config.client_id }}/lists.json` - records at response root.
- `suppressionlist`: GET `/clients/{{ config.client_id }}/suppressionlist.json` - records path
  `Results`; page-number pagination; page parameter `page`; size parameter `pagesize`; starts at 1;
  page size 100; incremental cursor `Date`; formatted as `rfc3339`.
- `segments`: GET `/clients/{{ fanout.id }}/segments.json` - records at response root; fan-out; ids
  from request `/clients.json`; id field `ClientID`; id inserted into the request path; stamps
  `OwningClientID`.
- `templates`: GET `/clients/{{ fanout.id }}/templates.json` - records at response root; fan-out;
  ids from request `/clients.json`; id field `ClientID`; id inserted into the request path; stamps
  `OwningClientID`.
- `list_custom_fields`: GET `/lists/{{ fanout.id }}/customfields.json` - records at response root;
  fan-out; ids from request `/clients/{{ config.client_id }}/lists.json`; id field `ListID`; id
  inserted into the request path; stamps `ListID`.
- `list_webhooks`: GET `/lists/{{ fanout.id }}/webhooks.json` - records at response root; fan-out;
  ids from request `/clients/{{ config.client_id }}/lists.json`; id field `ListID`; id inserted into
  the request path; stamps `ListID`.
- `active_subscribers`: GET `/lists/{{ fanout.id }}/active.json` - records path `Results`;
  page-number pagination; page parameter `page`; size parameter `pagesize`; starts at 1; page size
  1000; fan-out; ids from request `/clients/{{ config.client_id }}/lists.json`; id field `ListID`;
  id inserted into the request path; stamps `ListID`.
- `unconfirmed_subscribers`: GET `/lists/{{ fanout.id }}/unconfirmed.json` - records path `Results`;
  page-number pagination; page parameter `page`; size parameter `pagesize`; starts at 1; page size
  1000; fan-out; ids from request `/clients/{{ config.client_id }}/lists.json`; id field `ListID`;
  id inserted into the request path; stamps `ListID`.
- `unsubscribed_subscribers`: GET `/lists/{{ fanout.id }}/unsubscribed.json` - records path
  `Results`; page-number pagination; page parameter `page`; size parameter `pagesize`; starts at 1;
  page size 1000; fan-out; ids from request `/clients/{{ config.client_id }}/lists.json`; id field
  `ListID`; id inserted into the request path; stamps `ListID`.
- `bounced_subscribers`: GET `/lists/{{ fanout.id }}/bounced.json` - records path `Results`;
  page-number pagination; page parameter `page`; size parameter `pagesize`; starts at 1; page size
  1000; fan-out; ids from request `/clients/{{ config.client_id }}/lists.json`; id field `ListID`;
  id inserted into the request path; stamps `ListID`.
- `deleted_subscribers`: GET `/lists/{{ fanout.id }}/deleted.json` - records path `Results`;
  page-number pagination; page parameter `page`; size parameter `pagesize`; starts at 1; page size
  1000; fan-out; ids from request `/clients/{{ config.client_id }}/lists.json`; id field `ListID`;
  id inserted into the request path; stamps `ListID`.

## Write actions & risks

Overall write risk: external mutation of Campaign Monitor lists, subscribers
(add/update/unsubscribe/delete), segments, and draft campaigns; sending a campaign delivers real
email to real recipients.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_list`: POST `/lists/{{ config.client_id }}.json` - kind `create`; body type `json`;
  required record fields `Title`; accepted fields `ConfirmationSuccessPage`, `ConfirmedOptIn`,
  `Title`, `UnsubscribePage`, `UnsubscribeSetting`; risk: creates a new subscriber list under the
  configured client; low-risk external mutation, no approval required.
- `update_list`: PUT `/lists/{{ record.ListID }}.json` - kind `update`; body type `json`; path
  fields `ListID`; required record fields `ListID`; accepted fields `AddUnsubscribesToSuppList`,
  `ConfirmationSuccessPage`, `ConfirmedOptIn`, `ListID`, `ScrubActiveWithSuppList`, `Title`,
  `UnsubscribePage`, `UnsubscribeSetting`; risk: updates list settings; enabling
  ScrubActiveWithSuppList/AddUnsubscribesToSuppList changes subscriber state for existing contacts;
  low-risk external mutation, no approval required.
- `delete_list`: DELETE `/lists/{{ record.ListID }}.json` - kind `delete`; body type `none`; path
  fields `ListID`; required record fields `ListID`; accepted fields `ListID`; missing records
  treated as success for status `404`; risk: permanently removes a subscriber list and all of its
  subscribers/segments; irreversible, approval required.
- `add_subscriber`: POST `/subscribers/{{ record.ListID }}.json` - kind `create`; body type `json`;
  path fields `ListID`; required record fields `ListID`, `EmailAddress`; accepted fields
  `ConsentToSendSms`, `ConsentToTrack`, `CustomFields`, `EmailAddress`, `ListID`, `MobileNumber`,
  `Name`, `RestartSubscriptionBasedAutoresponders`, `Resubscribe`; risk: adds a new subscriber to a
  list; low-risk external mutation, no approval required.
- `update_subscriber`: PUT `/subscribers/{{ record.ListID }}.json?email={{
  record.CurrentEmailAddress | urlencode }}` - kind `update`; body type `json`; path fields
  `ListID`, `CurrentEmailAddress`; required record fields `ListID`, `CurrentEmailAddress`,
  `EmailAddress`; accepted fields `ConsentToSendSms`, `ConsentToTrack`, `CurrentEmailAddress`,
  `CustomFields`, `EmailAddress`, `ListID`, `MobileNumber`, `Name`,
  `RestartSubscriptionBasedAutoresponders`, `Resubscribe`; risk: updates an existing subscriber's
  profile/consent fields on a list, identified by their current email (CurrentEmailAddress, kept out
  of the body via path_fields since the API takes it as a query param, not a body field); low-risk
  external mutation, no approval required.
- `unsubscribe_subscriber`: POST `/subscribers/{{ record.ListID }}/unsubscribe.json` - kind
  `update`; body type `json`; path fields `ListID`; required record fields `ListID`, `EmailAddress`;
  accepted fields `EmailAddress`, `ListID`; risk: unsubscribes a contact from a list; low-risk
  external mutation, no approval required.
- `delete_subscriber`: DELETE `/subscribers/{{ record.ListID }}.json?email={{ record.EmailAddress |
  urlencode }}` - kind `delete`; body type `none`; path fields `ListID`, `EmailAddress`; required
  record fields `ListID`, `EmailAddress`; accepted fields `EmailAddress`, `ListID`; missing records
  treated as success for status `404`; risk: permanently removes a subscriber's record from a list
  (distinct from unsubscribing - this deletes the record entirely); irreversible, approval
  recommended.
- `create_segment`: POST `/segments/{{ record.ListID }}.json` - kind `create`; body type `json`;
  path fields `ListID`; required record fields `ListID`, `Title`, `RuleGroups`; accepted fields
  `ListID`, `RuleGroups`, `Title`; risk: creates a new subscriber segment (a saved rule-based
  filter) on a list; low-risk external mutation, no approval required.
- `update_segment`: PUT `/segments/{{ record.SegmentID }}.json` - kind `update`; body type `json`;
  path fields `SegmentID`; required record fields `SegmentID`, `Title`, `RuleGroups`; accepted
  fields `RuleGroups`, `SegmentID`, `Title`; risk: replaces a segment's name and full rule set;
  low-risk external mutation, no approval required.
- `delete_segment`: DELETE `/segments/{{ record.SegmentID }}.json` - kind `delete`; body type
  `none`; path fields `SegmentID`; required record fields `SegmentID`; accepted fields `SegmentID`;
  missing records treated as success for status `404`; risk: permanently removes a segment; any
  campaign scheduled to send to it loses that targeting; irreversible, low-risk, no approval
  required.
- `create_campaign`: POST `/campaigns/{{ config.client_id }}.json` - kind `create`; body type
  `json`; required record fields `Name`, `Subject`, `FromName`, `FromEmail`, `ReplyTo`, `HtmlUrl`,
  `ListIDs`; accepted fields `FromEmail`, `FromName`, `HtmlUrl`, `InlineCss`, `ListIDs`, `Name`,
  `ReplyTo`, `SegmentIDs`, `Subject`; risk: creates a new DRAFT campaign under the configured
  client; drafts are not sent until send_campaign is separately invoked, so this alone has no
  delivery side effect; low-risk, no approval required.
- `send_campaign`: POST `/campaigns/{{ record.CampaignID }}/send.json` - kind `update`; body type
  `json`; path fields `CampaignID`; required record fields `CampaignID`, `ConfirmationEmail`,
  `SendDate`; accepted fields `CampaignID`, `ConfirmationEmail`, `SendDate`; risk: delivers a real
  email campaign to every subscriber on its targeted lists/segments; irreversible once sent,
  approval required.
- `unschedule_campaign`: POST `/campaigns/{{ record.CampaignID }}/unschedule.json` - kind `update`;
  body type `none`; path fields `CampaignID`; required record fields `CampaignID`; accepted fields
  `CampaignID`; risk: cancels a scheduled-but-not-yet-sent campaign, reverting it to a draft;
  low-risk, no approval required.
- `delete_campaign`: DELETE `/campaigns/{{ record.CampaignID }}.json` - kind `delete`; body type
  `none`; path fields `CampaignID`; required record fields `CampaignID`; accepted fields
  `CampaignID`; missing records treated as success for status `404`; risk: permanently removes a
  draft or sent campaign's record from Campaign Monitor; irreversible, approval recommended.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 13 stream-backed endpoint group(s), 14 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, duplicate_of=13, non_data_endpoint=7, out_of_scope=54,
  requires_elevated_scope=4.
