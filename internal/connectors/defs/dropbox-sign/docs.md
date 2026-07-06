# Overview

Reads Dropbox Sign (HelloSign) signature requests, templates, team members, and account details, and
writes signature-request/template/team/account lifecycle mutations, through the Dropbox Sign REST
API.

Readable streams: `signature_requests`, `templates`, `team_members`, `account`.

Write actions: `update_signature_request`, `cancel_signature_request`, `remind_signature_request`,
`release_hold_signature_request`, `remove_signature_request`, `delete_template`,
`add_template_user`, `remove_template_user`, `create_team`, `update_team`, `add_team_member`,
`remove_team_member`, `update_account`.

Service API documentation: https://developers.hellosign.com/api/reference/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Dropbox Sign API key. Sent as the HTTP Basic username with a
  blank password; never logged.
- `base_url` (optional, string); default `https://api.hellosign.com/v3`; format `uri`; Dropbox Sign
  API base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.hellosign.com/v3`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/account`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `page_size`;
starts at 1; page size 100.

Pagination by stream: none: `account`; page_number: `signature_requests`, `templates`,
`team_members`.

- `signature_requests`: GET `/signature_request/list` - records path `signature_requests`;
  page-number pagination; page parameter `page`; size parameter `page_size`; starts at 1; page size
  100.
- `templates`: GET `/template/list` - records path `templates`; page-number pagination; page
  parameter `page`; size parameter `page_size`; starts at 1; page size 100.
- `team_members`: GET `/team/members` - records path `team_members`; page-number pagination; page
  parameter `page`; size parameter `page_size`; starts at 1; page size 100.
- `account`: GET `/account` - records path `account`.

## Write actions & risks

Overall write risk: external mutation of signature requests
(update/cancel/remind/release_hold/remove), templates (delete/add_user/remove_user), teams
(create/update/add_member/remove_member), and account settings; several actions are destructive/not
reversible (cancel_signature_request, remove_signature_request, delete_template, remove_team_member)
and require approval.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `update_signature_request`: POST `/signature_request/update/{{ record.signature_request_id }}` -
  kind `update`; body type `json`; path fields `signature_request_id`; required record fields
  `signature_request_id`, `signature_id`; accepted fields `email_address`, `name`, `signature_id`,
  `signature_request_id`; risk: external mutation; changes a signer's email address or name on an
  in-progress signature request, redirecting where the next request/reminder is delivered; approval
  required.
- `cancel_signature_request`: POST `/signature_request/cancel/{{ record.signature_request_id }}` -
  kind `update`; body type `none`; path fields `signature_request_id`; required record fields
  `signature_request_id`; accepted fields `signature_request_id`; confirmation `destructive`; risk:
  destructive external mutation; cancels an incomplete signature request, this action is not
  reversible; approval required.
- `remind_signature_request`: POST `/signature_request/remind/{{ record.signature_request_id }}` -
  kind `custom`; body type `json`; path fields `signature_request_id`; required record fields
  `signature_request_id`, `email_address`; accepted fields `email_address`, `name`,
  `signature_request_id`; risk: external mutation; sends an email reminder to a signer; cannot be
  sent again within 1 hour of the last reminder (manual or automatic).
- `release_hold_signature_request`: POST `/signature_request/release_hold/{{
  record.signature_request_id }}` - kind `update`; body type `none`; path fields
  `signature_request_id`; required record fields `signature_request_id`; accepted fields
  `signature_request_id`; risk: external mutation; releases a held signature request created from an
  UnclaimedDraft, immediately sending requests to all signers; approval required.
- `remove_signature_request`: POST `/signature_request/remove/{{ record.signature_request_id }}` -
  kind `delete`; body type `none`; path fields `signature_request_id`; required record fields
  `signature_request_id`; accepted fields `signature_request_id`; confirmation `destructive`; risk:
  destructive external mutation; removes the caller's access to a completed signature request from
  the account's list view, this action is not reversible; approval required.
- `delete_template`: POST `/template/delete/{{ record.template_id }}` - kind `delete`; body type
  `none`; path fields `template_id`; required record fields `template_id`; accepted fields
  `template_id`; confirmation `destructive`; risk: destructive external mutation; completely deletes
  a template from the account, this action is not reversible; approval required.
- `add_template_user`: POST `/template/add_user/{{ record.template_id }}` - kind `custom`; body type
  `json`; path fields `template_id`; required record fields `template_id`; accepted fields
  `account_id`, `email_address`, `skip_notification`, `template_id`; risk: external mutation; grants
  the specified account (which must already be a Team member) access to a template.
- `remove_template_user`: POST `/template/remove_user/{{ record.template_id }}` - kind `custom`;
  body type `json`; path fields `template_id`; required record fields `template_id`; accepted fields
  `account_id`, `email_address`, `template_id`; risk: external mutation; revokes the specified
  account's access to a template.
- `create_team`: POST `/team/create` - kind `create`; body type `json`; accepted fields `name`;
  risk: external mutation; creates a new Team and makes the calling account its member; fails if the
  caller already belongs to a Team.
- `update_team`: PUT `/team` - kind `update`; body type `json`; required record fields `name`;
  accepted fields `name`; risk: external mutation; renames the caller's own Team.
- `add_team_member`: PUT `/team/add_member` - kind `custom`; body type `json`; accepted fields
  `account_id`, `email_address`, `role`; risk: external mutation; invites or moves a user onto the
  caller's Team, creating a new Dropbox Sign account for the invited email if one does not already
  exist.
- `remove_team_member`: POST `/team/remove_member` - kind `custom`; body type `json`; accepted
  fields `account_id`, `email_address`, `new_owner_email_address`, `new_role`, `new_team_id`;
  confirmation `destructive`; risk: destructive external mutation; removes a user from the caller's
  Team; optionally transfers the removed account's documents to another account (Enterprise plans
  only), which is not reversible; approval required.
- `update_account`: PUT `/account` - kind `update`; body type `json`; accepted fields `account_id`,
  `callback_url`, `locale`; risk: external mutation; updates the caller's account settings
  (currently limited to the event callback URL and locale).

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s), 13 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=25, destructive_admin=4, duplicate_of=4, non_data_endpoint=5, out_of_scope=19,
  requires_elevated_scope=1.
