# Overview

Reads Freshdesk tickets, contacts, companies, agents, and groups through the Freshdesk REST API v2.

Readable streams: `tickets`, `contacts`, `companies`, `agents`, `groups`.

This connector is currently read-only; no write actions are declared. The full Freshdesk API reference is inventoried in `api_surface.json` for follow-up stream, direct-read, binary/export, and reverse-ETL lanes. Unimplemented operations are blocked by default and are not exposed as raw HTTP commands.

Service API documentation: https://developers.freshdesk.com/api/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Freshdesk API key. Sent as the HTTP Basic username (password
  is the literal X); never logged.
- `base_url` (required, string); format `uri`; Freshdesk API base URL, e.g. https://<domain>/api/v2
  (domain example: acme.freshdesk.com).
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; for the tickets stream
  only, objects updated at or after this time are read (updated_since).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `max_pages=0`, `page_size=100`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/agents` with query `per_page`=`1`.

## Streams notes

Default pagination: follows RFC 5988 Link headers with rel=next.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `tickets`: GET `/tickets` - records path `.`; follows RFC 5988 Link headers with rel=next;
  incremental cursor `updated_at`; sent as `updated_since`; formatted as `rfc3339`; initial lower
  bound from `start_date`.
- `contacts`: GET `/contacts` - records path `.`; follows RFC 5988 Link headers with rel=next.
- `companies`: GET `/companies` - records path `.`; follows RFC 5988 Link headers with rel=next.
- `agents`: GET `/agents` - records path `.`; follows RFC 5988 Link headers with rel=next.
- `groups`: GET `/groups` - records path `.`; follows RFC 5988 Link headers with rel=next.

## Write actions & risks

This connector currently declares no executable write actions. The operation ledger records Freshdesk create, update, admin, export/import, and delete candidates as blocked until named reverse-ETL actions provide schemas, redaction, plan previews, approval tokens, and typed confirmation for destructive/admin operations.

Read behavior: external Freshdesk API read of support tickets, contacts, companies, and team routing records.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage inventory includes 170 Freshdesk documentation rows from the public REST API reference (117 GET, 10 POST, 10 PUT, 33 DELETE).
- Implemented coverage remains limited to 5 stream-backed endpoint groups: `tickets`, `contacts`, `companies`, `agents`, and `groups`.
- All other inventoried operations are blocked by default in metadata for later #175-#179 lanes; they are not exposed as raw HTTP reads or writes.
