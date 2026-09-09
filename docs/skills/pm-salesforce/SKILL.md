---
name: pm-salesforce
description: Salesforce connector knowledge and safe action guide.
---

# pm-salesforce

## Purpose

Reads Salesforce object metadata and allow-listed Account, Contact, and Lead SOQL queries through the REST API. Read-only.

## Icon

- id: salesforce
- asset: icons/salesforce.svg
- source: official
- review_status: official_verified
- review_url: https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/intro_rest.htm

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- api_version
- instance_url (required)
- mode
- access_token (secret) (required)

## ETL Streams

- sobjects:
  - primary key: qualified_api_name
  - fields: label(string), qualified_api_name(string)
- accounts:
  - primary key: id
  - fields: email(string), id(string), last_modified_date(string), name(string)
- contacts:
  - primary key: id
  - fields: email(string), id(string), last_modified_date(string), name(string)
- leads:
  - primary key: id
  - fields: email(string), id(string), last_modified_date(string), name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Salesforce API read of object metadata, Account, Contact, and Lead records
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Declared salesforce API commands.
- Usage: pm salesforce <command> [flags]
- Global flags:
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
- Other Commands
  - operations get-services-data-version-sobjects - Declared etl: GET /services/data/{version}/sobjects. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations get-services-data-version-query-select-from-account - Declared etl: GET /services/data/{version}/query (SELECT ... FROM Account). [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations get-services-data-version-query-select-from-contact - Declared etl: GET /services/data/{version}/query (SELECT ... FROM Contact). [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations get-services-data-version-query-select-from-lead - Declared etl: GET /services/data/{version}/query (SELECT ... FROM Lead). [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations get-services-data-version-query-arbitrary-soql - Declared direct read: GET /services/data/{version}/query (arbitrary SOQL). [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-services-data-version-sobjects-s-object-describe - Declared direct read: GET /services/data/{version}/sobjects/{sObject}/describe. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-services-data-version-sobjects-s-object - Declared direct write: POST /services/data/{version}/sobjects/{sObject}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /services/data/{version}/sobjects/{sObject}.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations patch-services-data-version-sobjects-s-object-id - Declared direct write: PATCH /services/data/{version}/sobjects/{sObject}/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PATCH /services/data/{version}/sobjects/{sObject}/{id}.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations delete-services-data-version-sobjects-s-object-id - Declared direct write: DELETE /services/data/{version}/sobjects/{sObject}/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /services/data/{version}/sobjects/{sObject}/{id}.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-services-data-version-jobs-query-bulk-api - Declared direct read: GET /services/data/{version}/jobs/query (Bulk API). [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor

## Commands

### Inspect as a manual

```bash
pm connectors inspect salesforce
```

### Inspect as structured JSON

```bash
pm connectors inspect salesforce --json
```

## Agent Rules

- Run pm connectors inspect salesforce before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
