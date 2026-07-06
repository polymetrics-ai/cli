# Overview

Reads Harness NextGen organizations, projects, services, connectors, and pipelines through the
Harness platform REST API.

Readable streams: `organizations`, `projects`, `services`, `connectors`, `pipelines`.

This connector is read-only; no write actions are declared.

Service API documentation:
https://developer.harness.io/docs/platform/automation/api/api-quickstart/.

## Auth setup

Connection fields:

- `account_id` (required, string); Harness account identifier; scopes every NextGen read via the
  accountIdentifier query parameter.
- `api_key` (required, secret, string); Harness NextGen API key, sent as the x-api-key header. Never
  logged.
- `base_url` (optional, string); default `https://app.harness.io`; format `uri`; Harness platform
  API base URL override for tests, self-managed instances, or proxies.
- `mode` (optional, string).
- `page_size` (optional, string); default `50`; Records per page (1-100).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://app.harness.io`, `page_size=50`.

Authentication behavior:

- API key authentication in `x-api-key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/ng/api/organizations`.

## Streams notes

Default pagination: page-number pagination; page parameter `pageIndex`; size parameter `pageSize`;
starts at 0; page size 50.

- `organizations`: GET `/ng/api/organizations` - records path `data.content`; query
  `accountIdentifier`=`{{ config.account_id }}`; page-number pagination; page parameter `pageIndex`;
  size parameter `pageSize`; starts at 0; page size 50; computed output fields `account_identifier`,
  `description`, `identifier`, `name`.
- `projects`: GET `/ng/api/projects` - records path `data.content`; query `accountIdentifier`=`{{
  config.account_id }}`; page-number pagination; page parameter `pageIndex`; size parameter
  `pageSize`; starts at 0; page size 50; computed output fields `account_identifier`, `color`,
  `description`, `identifier`, `modules`, `name`, `org_identifier`.
- `services`: GET `/ng/api/servicesV2` - records path `data.content`; query `accountIdentifier`=`{{
  config.account_id }}`; page-number pagination; page parameter `pageIndex`; size parameter
  `pageSize`; starts at 0; page size 50; computed output fields `account_identifier`, `deleted`,
  `description`, `identifier`, `name`, `org_identifier`, `project_identifier`.
- `connectors`: GET `/ng/api/connectors` - records path `data.content`; query
  `accountIdentifier`=`{{ config.account_id }}`; page-number pagination; page parameter `pageIndex`;
  size parameter `pageSize`; starts at 0; page size 50; computed output fields `description`,
  `identifier`, `name`, `org_identifier`, `project_identifier`, `type`.
- `pipelines`: GET `/pipeline/api/pipelines/list` - records path `data.content`; query
  `accountIdentifier`=`{{ config.account_id }}`; page-number pagination; page parameter `pageIndex`;
  size parameter `pageSize`; starts at 0; page size 50; computed output fields `description`,
  `identifier`, `name`, `org_identifier`, `project_identifier`, `stage_count`.

## Write actions & risks

This connector is read-only. Read behavior: external Harness NextGen platform API read of
organization/project/service/connector/pipeline metadata.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=2.
