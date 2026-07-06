# Overview

Reads Bitly organizations, groups, campaigns, channels, bitlinks, branded short domains, webhooks,
QR codes, and group tags, and writes bitlink/campaign/group/channel/webhook/custom-bitlink/QR-code
mutations, through the Bitly v4 REST API.

Readable streams: `organizations`, `groups`, `campaigns`, `channels`, `bsds`, `webhooks`,
`qr_codes`, `group_tags`, `bitlinks`.

Write actions: `create_bitlink`, `update_bitlink`, `delete_bitlink`, `update_bitlink_tags`,
`delete_bitlink_tags`, `create_campaign`, `update_campaign`, `update_group`,
`update_group_preferences`, `create_channel`, `update_channel`, `create_webhook`, `update_webhook`,
`delete_webhook`, `create_custom_bitlink`, `update_custom_bitlink`, `create_qr_code`,
`update_qr_code`, `delete_qr_code`.

Service API documentation: https://dev.bitly.com/api-reference/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Bitly OAuth access token, sent as a Bearer token
  (Authorization: Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://api-ssl.bitly.com/v4`; format `uri`; Bitly API
  base URL override for tests or proxies.
- `group_guid` (optional, string); Bitly group guid the 'bitlinks' stream is scoped to (required for
  that stream; substituted into the group-scoped bitlinks path).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api-ssl.bitly.com/v4`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/groups`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: next_url: `bitlinks`; none: `organizations`, `groups`, `campaigns`,
`channels`, `bsds`, `webhooks`, `qr_codes`, `group_tags`.

- `organizations`: GET `/organizations` - records path `organizations`.
- `groups`: GET `/groups` - records path `groups`.
- `campaigns`: GET `/campaigns` - records path `campaigns`.
- `channels`: GET `/channels` - records path `channels`.
- `bsds`: GET `/bsds` - records at response root; computed output fields `account`.
- `webhooks`: GET `/webhooks` - records path `webhooks`.
- `qr_codes`: GET `/qr-codes` - records path `qr_codes`.
- `group_tags`: GET `/groups/{{ config.group_guid }}/tags` - records at response root; computed
  output fields `group_guid`.
- `bitlinks`: GET `/groups/{{ config.group_guid }}/bitlinks` - records path `links`; query
  `size`=`50`; follows a next-page URL from the response body; URL path `pagination.next`; next URLs
  stay on the configured API host.

## Write actions & risks

Overall write risk: external mutation of bitlinks, campaigns, groups, channels, webhooks, custom
bitlinks, and QR codes; create_webhook/update_webhook register or repoint an outbound event delivery
URL of the caller's choosing and warrant review before use, every write ships with an explicit
per-action risk string.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_bitlink`: POST `/bitlinks` - kind `create`; body type `json`; required record fields
  `long_url`; accepted fields `deeplinks`, `domain`, `group_guid`, `long_url`, `tags`, `title`;
  risk: creates a new publicly-resolvable short link; low-risk external mutation, no approval
  required.
- `update_bitlink`: PATCH `/bitlinks/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `archived`, `deeplinks`, `id`, `long_url`,
  `tags`, `title`; risk: mutates an existing bitlink's metadata or redirect destination; changing
  long_url redirects all future traffic on that short link and consumes an encode-limit unit.
- `delete_bitlink`: DELETE `/bitlinks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: permanently removes a bitlink; any traffic still hitting the short URL
  starts failing to resolve.
- `update_bitlink_tags`: PATCH `/bitlinks/{{ record.id }}/tags` - kind `update`; body type `json`;
  path fields `id`; body fields `tags`; required record fields `id`, `tags`; accepted fields `id`,
  `tags`; risk: replaces the full tag set on a bitlink; overwrites any tags not included in the
  submitted list.
- `delete_bitlink_tags`: DELETE `/bitlinks/{{ record.id }}/tags` - kind `delete`; body type `json`;
  path fields `id`; body fields `tags`; required record fields `id`, `tags`; accepted fields `id`,
  `tags`; missing records treated as success for status `404`; risk: removes the named tags from a
  bitlink; irreversible without re-adding them via update_bitlink_tags.
- `create_campaign`: POST `/campaigns` - kind `create`; body type `json`; required record fields
  `group_guid`; accepted fields `channel_guids`, `description`, `group_guid`, `name`; risk: creates
  a new campaign container in the target group; low-risk external mutation, no approval required.
- `update_campaign`: PATCH `/campaigns/{{ record.guid }}` - kind `update`; body type `json`; path
  fields `guid`; required record fields `guid`; accepted fields `channel_guids`, `description`,
  `guid`, `name`; risk: mutates an existing campaign's name, description, or associated channels.
- `update_group`: PATCH `/groups/{{ record.guid }}` - kind `update`; body type `json`; path fields
  `guid`; required record fields `guid`; accepted fields `guid`, `name`, `organization_guid`; risk:
  renames or re-parents an existing group; a visible change for every member of that group.
- `update_group_preferences`: PATCH `/groups/{{ record.group_guid }}/preferences` - kind `update`;
  body type `json`; path fields `group_guid`; body fields `domain_preference`; required record
  fields `group_guid`, `domain_preference`; accepted fields `domain_preference`, `group_guid`; risk:
  changes the default branded short domain new bitlinks in this group are created with.
- `create_channel`: POST `/channels` - kind `create`; body type `json`; accepted fields
  `campaign_guid`, `group_guid`, `name`; risk: creates a new channel container; low-risk external
  mutation, no approval required.
- `update_channel`: PATCH `/channels/{{ record.guid }}` - kind `update`; body type `json`; path
  fields `guid`; required record fields `guid`; accepted fields `campaign_guid`, `guid`, `name`;
  risk: mutates an existing channel's name or campaign association.
- `create_webhook`: POST `/webhooks` - kind `create`; body type `json`; required record fields
  `group_guid`, `event`, `url`; accepted fields `campaign_guid`, `event`, `group_guid`, `is_active`,
  `url`; risk: registers a new outbound webhook that will POST live event data (clicks/scans) to an
  external URL of the caller's choosing; verify the target endpoint before enabling.
- `update_webhook`: PATCH `/webhooks/{{ record.guid }}` - kind `update`; body type `json`; path
  fields `guid`; required record fields `guid`; accepted fields `event`, `guid`, `is_active`, `url`;
  risk: mutates an existing webhook's target URL, event type, or active state; a changed url
  redirects future event deliveries to a different endpoint.
- `delete_webhook`: DELETE `/webhooks/{{ record.guid }}` - kind `delete`; body type `none`; path
  fields `guid`; required record fields `guid`; accepted fields `guid`; missing records treated as
  success for status `404`; risk: permanently removes a webhook subscription; event delivery to its
  target URL stops immediately.
- `create_custom_bitlink`: POST `/custom_bitlinks` - kind `create`; body type `json`; required
  record fields `custom_bitlink`, `bitlink_id`; accepted fields `bitlink_id`, `custom_bitlink`;
  risk: claims a custom keyword/back-half on a branded short domain and points it at a bitlink;
  consumes a finite custom-bitlink allocation on the domain.
- `update_custom_bitlink`: PATCH `/custom_bitlinks/{{ record.custom_bitlink }}` - kind `update`;
  body type `json`; path fields `custom_bitlink`; body fields `bitlink_id`; required record fields
  `custom_bitlink`, `bitlink_id`; accepted fields `bitlink_id`, `custom_bitlink`; risk: re-points an
  existing custom keyword at a different bitlink; redirects all future traffic hitting that custom
  URL to the new destination.
- `create_qr_code`: POST `/qr-codes` - kind `create`; body type `json`; required record fields
  `group_guid`, `destination`; accepted fields `destination`, `group_guid`, `title`; risk: creates a
  new QR code resource pointed at a bitlink or long_url; low-risk external mutation, no approval
  required.
- `update_qr_code`: PATCH `/qr-codes/{{ record.qrcode_id }}` - kind `update`; body type `json`; path
  fields `qrcode_id`; required record fields `qrcode_id`; accepted fields `destination`,
  `qrcode_id`, `title`; risk: mutates an existing QR code's title or destination; changing
  destination redirects anyone scanning an already-printed/distributed code.
- `delete_qr_code`: DELETE `/qr-codes/{{ record.qrcode_id }}` - kind `delete`; body type `none`;
  path fields `qrcode_id`; required record fields `qrcode_id`; accepted fields `qrcode_id`; missing
  records treated as success for status `404`; risk: permanently removes a QR code resource; any
  already-printed/distributed copy of the code stops resolving.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 9 stream-backed endpoint group(s), 19 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=2, duplicate_of=18, non_data_endpoint=4, out_of_scope=33,
  requires_elevated_scope=2.
