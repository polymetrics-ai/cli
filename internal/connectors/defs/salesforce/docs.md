# Overview

Reads Salesforce object metadata and allow-listed Account, Contact, and Lead SOQL queries through
the REST API. Read-only.

Readable streams: `sobjects`, `accounts`, `contacts`, `leads`.

This connector is read-only; no write actions are declared.

Service API documentation:
https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/intro_rest.htm.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Salesforce OAuth access token, used as the Bearer
  credential. Never logged.
- `api_version` (optional, string); default `v60.0`; Salesforce REST API version, e.g. v60.0 or 60.0
  (a leading v is added if missing).
- `instance_url` (required, string); format `uri`; Salesforce instance/org base URL (e.g.
  https://yourinstance.my.salesforce.com).
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `api_version=v60.0`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `instance_url` value after applying defaults.

Connection checks call GET `/services/data/{{ config.api_version }}/`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: next_url: `accounts`, `contacts`, `leads`; none: `sobjects`.

- `sobjects`: GET `/services/data/{{ config.api_version }}/sobjects` - records path `sobjects`;
  computed output fields `label`, `qualified_api_name`.
- `accounts`: GET `/services/data/{{ config.api_version }}/query` - records path `records`; query
  `q`=`SELECT Id, Name, LastModifiedDate FROM Account ORDER BY LastModifiedDate ASC`; follows a
  next-page URL from the response body; URL path `nextRecordsUrl`; next URLs stay on the configured
  API host; computed output fields `id`, `last_modified_date`, `name`.
- `contacts`: GET `/services/data/{{ config.api_version }}/query` - records path `records`; query
  `q`=`SELECT Id, Name, Email, LastModifiedDate FROM Contact ORDER BY LastModifiedDate ASC`; follows
  a next-page URL from the response body; URL path `nextRecordsUrl`; next URLs stay on the
  configured API host; computed output fields `email`, `id`, `last_modified_date`, `name`.
- `leads`: GET `/services/data/{{ config.api_version }}/query` - records path `records`; query
  `q`=`SELECT Id, Name, Email, LastModifiedDate FROM Lead ORDER BY LastModifiedDate ASC`; follows a
  next-page URL from the response body; URL path `nextRecordsUrl`; next URLs stay on the configured
  API host; computed output fields `email`, `id`, `last_modified_date`, `name`.

## Write actions & risks

This connector is read-only. Read behavior: external Salesforce API read of object metadata,
Account, Contact, and Lead records.

## Known limits

- Batch defaults: read_page_size=2000.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=6.
