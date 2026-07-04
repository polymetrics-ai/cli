# Overview

Freshservice is a declarative HTTP connector for the official Freshservice REST API v2. This Pass B bundle preserves the five legacy record projections for `tickets`, `agents`, `requesters`, `assets`, and `problems`, then adds documented JSON list/detail endpoints as passthrough streams plus documented mutation endpoints as write actions.

## Auth setup

Provide a Freshservice API key via the `api_key` secret. The connector sends it as the HTTP Basic username with the literal password `X`, matching the legacy implementation. `domain_name` is the Freshservice account host, for example `acme.freshservice.com`, and the connector appends `/api/v2`.

## Streams notes

The legacy streams remain first and keep schema projection so their emitted records match `internal/connectors/freshservice`: `tickets`, `agents`, `requesters`, `assets`, and `problems`. Newly added streams use `projection: passthrough` with permissive schemas because the legacy connector never shaped those records. Streams for detail and singleton endpoints disable pagination; list and filter endpoints use Freshservice page-number pagination with `page` and `per_page` at the documented page size of 100.

Path-scoped and filter streams require the corresponding config key named in `spec.json` (for example `ticket_id`, `project_id`, or a stream-specific query key). These keys are intentionally optional globally because they are only needed when the matching stream is selected.

## Write actions & risks

The bundle declares 263 write actions for documented POST, PUT, PATCH, and DELETE endpoints that the Tier-1 dialect can express as one HTTP request per record. Record schemas require path parameters and allow documented JSON body fields to pass through. DELETE actions send no request body, treat 404 as idempotent missing-ok, and are marked `confirm: destructive`. Reverse ETL must still follow plan, preview, approval, and execute.

## Known limits

- Download and export helpers, including attachment downloads and on-call calendar/PIR/audit exports, are excluded as `binary_payload` or `non_data_endpoint` in `api_surface.json`; they are not JSON list/detail streams or record mutations.
- Newly added stream schemas are permissive passthrough schemas derived from documented response wrappers, not hand-curated warehouse schemas. The existing five legacy streams remain narrow to preserve emitted-record parity.
- The declarative write dialect does not model compound workflows or multipart upload bodies. Actions here cover the direct JSON/form-compatible request shape only; callers must provide body fields accepted by Freshservice for the selected endpoint.
- `page_size` and `max_pages` remain documented legacy config fields, but this engine bundle uses the fixed declarative pagination size of 100 and does not expose runtime page-size/max-page overrides.
