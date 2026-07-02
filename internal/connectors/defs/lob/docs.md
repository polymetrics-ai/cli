# Lob

## Overview

Reads Lob addresses, postcards, letters, checks, and bank accounts through the Lob print & mail
REST API (`https://api.lob.com/v1`). Read-only: Lob's print/mail API has no safe reverse-ETL write
surface for pm, matching the legacy `internal/connectors/lob` package.

## Auth setup

An API key (`secrets.api_key`) is sent as the HTTP Basic auth username with a blank password
(`Authorization: Basic base64(api_key:)`), matching legacy's `connsdk.Basic(secret, "")` usage
exactly.

## Streams notes

All five streams (`addresses`, `postcards`, `letters`, `checks`, `bank_accounts`) list from their
respective Lob resource endpoints, emit records from the top-level `data` array, and share Lob's
`next_url` cursor pagination: each list response's `next_url` field carries the full absolute URL
(including Lob's own host) for the next page, or `null`/absent when exhausted. This bundle declares
`pagination.type: next_url` with `next_url_path: next_url`, reading that field directly — legacy's
own `afterCursor` helper (which manually parses the `after` query param back out of `next_url`
rather than following the URL directly) produces an equivalent next request in effect, since Lob's
`next_url` already fully encodes the next page's `limit`/`after` query state on the same resource
path. Every object exposes a string `id` and an ISO-8601 `date_created` timestamp; `date_created` is
each schema's declared cursor field, matching legacy. Legacy performs no automated incremental
filtering (a full sync always re-lists from page 1); the cursor field is declared for downstream
state-tracking purposes only, mirroring legacy's own no-op `InitialState`.

`postcards`, `letters`, and `checks` share an identical mailpiece record shape
(`url`/`carrier`/`status`/`send_date`/`expected_delivery_date`), matching legacy's shared
`lobMailpieceRecord` mapper.

## Write actions & risks

None. Lob is read-only in pm (`capabilities.write: false`), matching legacy.

## Known limits

- **`next_url` pagination ships single-page fixtures for every stream** (conventions.md §4's
  sanctioned exception): a `next_url` stream's next-page URL must be the fixture replay server's own
  ephemeral address, which cannot be embedded in a static fixture file authored ahead of time. Every
  stream in this bundle uses `next_url` pagination, so there is no non-paginated sibling stream (the
  bitly/calendly pattern) to pick for `pagination_terminates`'s dynamic 2-page proof; it runs against
  `addresses`'s single-page fixture instead, which still passes (one fixture page, one request, clean
  termination) but does not exercise a genuine second-page follow.
- **No live `paritytest/lob` 2-page correctness test**: the conventions' sanctioned single-page
  fixture exception pairs it with a live `paritytest/<name>` test driving a real `httptest.Server` to
  prove actual 2-page follow behavior. This wave's fan-out migration scope is JSON + docs.md only (no
  Go files); that live parity proof is not created here and is left for a follow-up wave with Go
  authoring scope. The `next_url` pagination type itself, its same-host SSRF guard, and its loop
  guard are all pre-existing, already-tested engine primitives (see `internal/connectors/engine/
  paginate.go`'s `nextURL`); only this specific bundle's live 2-page exercise is deferred.
- Lob's real API confirms `next_url`/`previous_url` are always fully-qualified absolute URLs
  (`https://api.lob.com/v1/...`), same-host as the request — legacy's own `readFixture` helper (a
  fixture-mode-only synthetic data generator, never exercised against the real API) happened to use a
  relative-path placeholder string; that is a test-double simplification, not evidence of a different
  real wire shape, and is not what this bundle's fixtures model.
- Pass B (mailpiece creation/send writes, templates, campaigns, bank account verification, address
  autocompletion) is out of scope; see `api_surface.json`.
