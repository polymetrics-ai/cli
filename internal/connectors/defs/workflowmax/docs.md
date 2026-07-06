# Overview

Reads and writes WorkflowMax jobs, clients, and client contacts through the real WorkflowMax API v2
(api.workflowmax2.com/v2).

Readable streams: `jobs`, `clients`.

Write actions: `create_client`, `update_client`, `delete_client`, `create_job`, `delete_job`,
`create_client_contact`, `update_client_contact`, `delete_client_contact`.

Service API documentation: https://api-docs.workflowmax.com/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); WorkflowMax API v2 OAuth2 access token, sent as a
  Bearer token on every request. Never logged.
- `account_id` (required, string); WorkflowMax/Xero organisation identifier, sent as the required
  account-id header on every v2 request (api-docs.workflowmax.com Authentication).
- `base_url` (optional, string); default `https://api.workflowmax2.com`; format `uri`; WorkflowMax
  API base URL. Defaults to the production v2 endpoint; override for test proxies.
- `mode` (optional, string).
- `updated_since` (optional, string); Optional YYYY-MM-DD lower bound for the jobs/clients
  updatedSince filter.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.workflowmax2.com`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v2/jobs` with query `pageSize`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `pageSize`; starts
at 1; page size 100; maximum 1 page(s).

- `jobs`: GET `/v2/jobs` - records path `data`; query `updatedSince` from template `{{
  config.updated_since }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `pageSize`; starts at 1; page size 100; maximum 1 page(s); emits passthrough records.
- `clients`: GET `/v2/clients` - records path `data`; query `updatedSince` from template `{{
  config.updated_since }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `pageSize`; starts at 1; page size 100; maximum 1 page(s); emits passthrough records.

## Write actions & risks

Overall write risk: external mutation of WorkflowMax jobs, clients, and client contacts
(create/update/delete); approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_client`: POST `/v2/clients` - kind `create`; body type `json`; accepted fields
  `clientManagerUuid`, `email`, `exportCode`, `favorite`, `firstName`, `jobManagerUuid`, `lastName`,
  `name`, `phone`, `referralSource`, `website`; risk: creates a WorkflowMax client record; approval
  required.
- `update_client`: PUT `/v2/clients/{{ record.uuid }}` - kind `update`; body type `json`; path
  fields `uuid`; required record fields `uuid`; accepted fields `clientManagerUUID`, `email`,
  `exportCode`, `favorite`, `firstName`, `jobManagerUUID`, `lastName`, `name`, `phone`,
  `referralSource`, `uuid`, `website`; risk: updates a WorkflowMax client record; approval required.
- `delete_client`: DELETE `/v2/clients/{{ record.uuid }}` - kind `delete`; body type `none`; path
  fields `uuid`; required record fields `uuid`; accepted fields `uuid`; risk: permanently deletes a
  WorkflowMax client record; approval required.
- `create_job`: POST `/v2/jobs` - kind `create`; body type `json`; required record fields
  `clientUUID`, `jobName`, `statusUUID`, `startDate`, `dueDate`, `priority`; accepted fields
  `budget`, `clientManagerUUID`, `clientOrderNumber`, `clientUUID`, `contactUUID`, `description`,
  `dueDate`, `jobCategoryUUID`, `jobManagerUUID`, `jobName`, `jobNumber`, `priority`, `startDate`,
  `statusUUID`; risk: creates a WorkflowMax job; approval required.
- `delete_job`: DELETE `/v2/jobs/{{ record.uuid }}` - kind `delete`; body type `none`; path fields
  `uuid`; required record fields `uuid`; accepted fields `uuid`; risk: permanently deletes a
  WorkflowMax job; approval required.
- `create_client_contact`: POST `/v2/clients/contacts` - kind `create`; body type `json`; required
  record fields `firstName`; accepted fields `email`, `favourite`, `firstName`, `lastName`,
  `mobile`, `phone`, `salutation`; risk: creates a WorkflowMax client-contact record (not attached
  to any client until linked); approval required.
- `update_client_contact`: PUT `/v2/clients/contacts/{{ record.uuid }}` - kind `update`; body type
  `json`; path fields `uuid`; required record fields `uuid`; accepted fields `email`, `firstName`,
  `lastName`, `mobile`, `phone`, `uuid`; risk: updates a WorkflowMax client-contact record; approval
  required.
- `delete_client_contact`: DELETE `/v2/clients/contacts/{{ record.uuid }}` - kind `delete`; body
  type `none`; path fields `uuid`; required record fields `uuid`; accepted fields `uuid`; risk:
  permanently deletes a WorkflowMax client-contact record; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s), 8 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=8, duplicate_of=6, non_data_endpoint=3, out_of_scope=83.
