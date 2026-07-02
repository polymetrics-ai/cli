# Overview

Dropbox Sign (formerly HelloSign) is an e-signature platform. This bundle reads Dropbox Sign
signature requests, templates, team members, and account details through the Dropbox Sign REST
API (`https://api.hellosign.com/v3`) using HTTP Basic auth. It is read-only, migrated from
`internal/connectors/dropbox-sign` (the hand-written connector this bundle replaces at capability
parity); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Dropbox Sign API key via the `api_key` secret; it is sent as the HTTP Basic username with
a blank password (`Authorization: Basic base64(api_key:)`), matching legacy's
`connsdk.Basic(secret, "")`, and is never logged.

## Streams notes

Four streams: `signature_requests` (`GET /signature_request/list`, records at
`signature_requests`), `templates` (`GET /template/list`, records at `templates`), `team_members`
(`GET /team/members`, records at `team_members`) share the base-level `page_number` pagination
(`page`/`page_size` query params, 100 records per page, stopping on a short page) — matches
legacy's `list_info.num_pages` page-increment loop for the common case where every page except the
last is exactly `page_size` long. `account` (`GET /account`, a single-object endpoint with no
`list_info`) overrides pagination to `none` at the stream level and takes no pagination query
params at all, matching legacy's `paginated := endpoint.recordsPath != "account"` special case.

Primary keys: `signature_request_id` (signature_requests), `template_id` (templates), `account_id`
(team_members, account). Incremental cursor fields (`created_at`/`updated_at`) are declared on the
schemas that have them, matching legacy's `CursorFields`, but no stream declares an `incremental`
block: Dropbox Sign's list endpoints accept no server-side `updated_since`-style filter parameter,
matching legacy's own read path (full refresh only, no incremental request param sent).

## Write actions & risks

None. Dropbox Sign is read-only in both legacy and this bundle (`capabilities.write: false`); no
`writes.json` file is shipped. Legacy's own comment explains why: signature requests are created
via file uploads and signer flows that do not map cleanly onto reverse-ETL record writes.

## Known limits

- Full Dropbox Sign API surface (sending signature requests, template creation, embedded signing,
  bulk send, unclaimed drafts) is out of scope for this wave; see `api_surface.json`'s
  `excluded: {category: out_of_scope}` entries. The document-download endpoint is separately
  excluded as `binary_payload` (a PDF, not a syncable record stream).
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting for
  Dropbox Sign, so none is added here either (matches legacy's real, lack-of, throttling behavior).
- Legacy's `pageCount` reads `list_info.num_pages` and stops once `page >= numPages` (or on an
  empty page); this bundle's `page_number` paginator instead stops purely on a short page
  (`recordCount < page_size`). Both converge on the identical final record set for every real
  Dropbox Sign response shape (every page except the last is exactly `page_size` long); only a
  final page that happens to be exactly `page_size` long causes legacy to issue one additional
  (empty) request that this bundle does not. Documented as a benign pagination-loop-count
  difference, not an emitted-data deviation.
