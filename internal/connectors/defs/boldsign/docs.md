# Overview

Reads BoldSign documents, templates, teams, contacts, brands, users, contact groups, and sender
identities, and writes team/contact-group/document-lifecycle/user-lifecycle mutations, through the
BoldSign REST API.

Readable streams: `documents`, `templates`, `teams`, `contacts`, `brands`, `users`,
`contact_groups`, `sender_identities`.

Write actions: `create_team`, `update_team`, `update_contact`, `delete_contact`,
`create_contact_group`, `update_contact_group`, `delete_contact_group`, `revoke_document`,
`remind_document`, `delete_document`, `add_document_tags`, `delete_document_tags`, `update_user`,
`change_user_team`.

Service API documentation: https://developers.boldsign.com/api-reference.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); BoldSign API key, sent as the X-API-KEY header. Never
  logged.
- `base_url` (optional, string); default `https://api.boldsign.com`; format `uri`; BoldSign API base
  URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.boldsign.com`.

Authentication behavior:

- API key authentication in `X-API-KEY` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/document/list` with query `Page`=`1`; `PageSize`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `Page`; size parameter `PageSize`; starts
at 1; page size 50.

- `documents`: GET `/v1/document/list` - records path `result`; page-number pagination; page
  parameter `Page`; size parameter `PageSize`; starts at 1; page size 50; computed output fields
  `created_date`, `document_id`, `enable_signing_order`, `expiry_date`, `is_deleted`,
  `message_title`, `sender_detail`, `sender_email`, `signer_details`.
- `templates`: GET `/v1/template/list` - records path `result`; page-number pagination; page
  parameter `Page`; size parameter `PageSize`; starts at 1; page size 50; computed output fields
  `created_date`, `document_id`, `is_shared_template`, `sender_email`, `template_description`,
  `template_name`.
- `teams`: GET `/v1/teams/list` - records path `results`; page-number pagination; page parameter
  `Page`; size parameter `PageSize`; starts at 1; page size 50; computed output fields
  `created_date`, `team_id`, `team_name`.
- `contacts`: GET `/v1/contacts/list` - records path `result`; page-number pagination; page
  parameter `Page`; size parameter `PageSize`; starts at 1; page size 50; computed output fields
  `company_name`, `id`, `phone_number`.
- `brands`: GET `/v1/brand/list` - records path `result`; page-number pagination; page parameter
  `Page`; size parameter `PageSize`; starts at 1; page size 50; computed output fields
  `background_color`, `brand_id`, `brand_name`, `button_color`, `is_default`.
- `users`: GET `/v1/users/list` - records path `result`; page-number pagination; page parameter
  `Page`; size parameter `PageSize`; starts at 1; page size 50; computed output fields
  `created_date`, `first_name`, `last_name`, `meta_data`, `modified_date`, `team_id`, `team_name`,
  `user_id`, `user_status`.
- `contact_groups`: GET `/v1/contactGroups/list` - records path `result`; page-number pagination;
  page parameter `Page`; size parameter `PageSize`; starts at 1; page size 50; computed output
  fields `group_id`, `group_name`.
- `sender_identities`: GET `/v1/senderIdentities/list` - records path `result`; page-number
  pagination; page parameter `Page`; size parameter `PageSize`; starts at 1; page size 50; computed
  output fields `approved_date`, `brand_id`, `created_by`, `meta_data`, `notification_settings`,
  `redirect_url`.

## Write actions & risks

Overall write risk: external mutation of BoldSign teams, contacts, contact groups, document
lifecycle state (revoke/remind/delete/tags), and user role/team/status; includes 2 destructive
(irreversible-effect) actions (delete_contact, delete_contact_group, delete_document,
revoke_document).

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_team`: POST `/v1/teams/create` - kind `create`; body type `json`; required record fields
  `teamName`; accepted fields `teamName`; risk: external mutation; creates a new BoldSign team;
  approval required.
- `update_team`: PUT `/v1/teams/update` - kind `update`; body type `json`; required record fields
  `teamId`, `teamName`; accepted fields `teamId`, `teamName`; risk: external mutation; renames an
  existing BoldSign team; approval required.
- `update_contact`: PUT `/v1/contacts/update?id={{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `email`, `name`; accepted fields `companyName`,
  `email`, `id`, `jobTitle`, `name`, `phoneNumber`; risk: external mutation; overwrites an existing
  BoldSign contact's details; approval required.
- `delete_contact`: DELETE `/v1/contacts/delete?id={{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: destructive external
  mutation; permanently deletes a BoldSign contact; approval required.
- `create_contact_group`: POST `/v1/contactGroups/create` - kind `create`; body type `json`;
  required record fields `groupName`; accepted fields `contacts`, `directories`, `groupName`; risk:
  external mutation; creates a new BoldSign contact group; approval required.
- `update_contact_group`: PUT `/v1/contactGroups/update?groupId={{ record.groupId }}` - kind
  `update`; body type `json`; path fields `groupId`; required record fields `groupId`, `groupName`;
  accepted fields `contacts`, `directories`, `groupId`, `groupName`; risk: external mutation;
  overwrites an existing BoldSign contact group's members/name; approval required.
- `delete_contact_group`: DELETE `/v1/contactGroups/delete?groupId={{ record.groupId }}` - kind
  `delete`; body type `none`; path fields `groupId`; required record fields `groupId`; accepted
  fields `groupId`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: destructive external mutation; permanently deletes a BoldSign contact group; approval
  required.
- `revoke_document`: POST `/v1/document/revoke?documentId={{ record.documentId }}` - kind `update`;
  body type `json`; path fields `documentId`; required record fields `documentId`, `message`;
  accepted fields `documentId`, `message`, `onBehalfOf`; confirmation `destructive`; risk:
  destructive external mutation; revokes a BoldSign document, permanently ending its signature
  request; approval required.
- `remind_document`: POST `/v1/document/remind?documentId={{ record.documentId }}` - kind `custom`;
  body type `json`; path fields `documentId`; required record fields `documentId`; accepted fields
  `documentId`, `message`, `onBehalfOf`, `reminderPhoneNumbers`; risk: external mutation; sends an
  email/SMS reminder to a document's pending signers; approval required.
- `delete_document`: DELETE `/v1/document/delete?documentId={{ record.documentId
  }}&deletePermanently={{ record.deletePermanently }}` - kind `delete`; body type `none`; path
  fields `documentId`, `deletePermanently`; required record fields `documentId`,
  `deletePermanently`; accepted fields `deletePermanently`, `documentId`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: destructive external mutation; moves a
  BoldSign document to trash (or permanently deletes it when deletePermanently=true); approval
  required.
- `add_document_tags`: PATCH `/v1/document/addTags` - kind `update`; body type `json`; required
  record fields `documentId`, `tags`; accepted fields `documentId`, `tags`; risk: external mutation;
  adds label tags to a BoldSign document; approval required.
- `delete_document_tags`: DELETE `/v1/document/deleteTags` - kind `update`; body type `json`;
  required record fields `documentId`, `tags`; accepted fields `documentId`, `tags`; risk: external
  mutation; removes label tags from a BoldSign document; approval required.
- `update_user`: PUT `/v1/users/update` - kind `update`; body type `json`; required record fields
  `userId`; accepted fields `toUserId`, `userId`, `userRole`, `userStatus`; risk: external mutation;
  changes a BoldSign user's role or active/deactivated status; approval required.
- `change_user_team`: PUT `/v1/users/changeTeam?userId={{ record.userId }}` - kind `update`; body
  type `json`; path fields `userId`; required record fields `userId`, `toTeamId`; accepted fields
  `toTeamId`, `transferDocumentsToUserId`, `userId`; risk: external mutation; moves a BoldSign user
  to a different team; approval required.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 8 stream-backed endpoint group(s), 14 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=6, destructive_admin=2, duplicate_of=10, non_data_endpoint=11, out_of_scope=34.
