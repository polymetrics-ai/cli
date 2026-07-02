# PandaDoc

## Overview

Reads PandaDoc documents, templates, and contacts through the PandaDoc public REST API
(`https://api.pandadoc.com/public/v1`). Read-only, matching the legacy
`internal/connectors/pandadoc` package this bundle migrates; the legacy package stays registered
and unchanged until wave6's registry flip.

## Auth setup

An API key (`secrets.api_key`) is sent as the `Authorization` header with an `API-Key ` prefix
(`Authorization: API-Key <key>`), matching legacy's
`connsdk.APIKeyHeader("Authorization", apiKey, "API-Key ")` usage exactly (`pandadoc.go:188-192`).
`api_key` is required; there is no unauthenticated fallback.

## Streams notes

All three streams (`documents`, `templates`, `contacts`) list from their respective PandaDoc
resource endpoints (`GET /documents`, `GET /templates`, `GET /contacts`), sending `count` (page
size, default 100, matching legacy's `pandaDocDefaultCount`/`pandaDocMaxCount` of 100) and a fixed
`page=1` as the initial query, and emit records from the top-level `results` array — matching
legacy's `connsdk.RecordsAt(resp.Body, "results")`.

Pagination follows PandaDoc's own `next` field, which carries the full absolute next-page URL (or
`null`/absent when exhausted) — this bundle declares `pagination.type: next_url` with
`next_url_path: next`, reading that field directly, matching legacy's `harvestNextURL` loop
(`pandadoc.go:132-160`), which follows `next` verbatim and stops when it is empty.

Each schema's fields are projected exactly from legacy's `mapRecord`, which copies only the
declared field set onto every record (`id`/`name`/`status`/`date_created`/`date_modified` for
documents; `id`/`name`/`date_created`/`date_modified` for templates;
`id`/`email`/`first_name`/`last_name`/`created_date` for contacts). `date_created`
(`documents`/`templates`) and `created_date` (`contacts`) are declared as each schema's cursor
field, matching legacy's `cursorFields`; legacy performs no automated incremental filtering (a full
sync always re-lists from page 1 with no date-range query parameter), so the cursor field is
declared for downstream state-tracking purposes only, mirroring legacy's own no-op `InitialState`
(`pandadoc.go:95-100`, which returns an empty cursor unconditionally).

`max_pages` accepts an integer, `all`, or `unlimited` in legacy (`pandaDocMaxPages`,
`pandadoc.go:247-260`); this bundle's `spec.json` keeps the identical string-typed `max_pages`
config surface with the same default (`0`, meaning unlimited).

## Write actions & risks

None. PandaDoc is read-only in pm (`capabilities.write: false`); legacy's own `Write` is a stub
returning `connectors.ErrUnsupportedOperation` (`pandadoc.go:179-181`), and this bundle ships no
`writes.json`.

## Known limits

- **`next_url` pagination ships single-page fixtures for every stream** (`docs/migration/
  conventions.md` §4's sanctioned exception, the same pattern used by `lob`): a `next_url` stream's
  next-page URL must be the fixture replay server's own ephemeral address, which cannot be embedded
  in a static fixture file authored ahead of time. Every stream in this bundle uses `next_url`
  pagination, so there is no non-paginated sibling stream to pick for `pagination_terminates`'s
  dynamic proof; it runs against `documents`'s single-page fixture instead (`next: null`), which
  still passes (one fixture page, one request, clean termination) but does not exercise a genuine
  second-page follow.
- **No live `paritytest/pandadoc` 2-page correctness test**: the conventions' sanctioned
  single-page fixture exception pairs it with a live `paritytest/<name>` test driving a real
  `httptest.Server` to prove actual 2-page follow behavior. This migration's scope is JSON +
  docs.md only (no Go files); that live parity proof is not created here and is left for a
  follow-up wave with Go authoring scope. The `next_url` pagination type itself, its same-host SSRF
  guard, and its loop guard are all pre-existing, already-tested engine primitives (see
  `internal/connectors/engine/paginate.go`'s `nextURL`); only this specific bundle's live 2-page
  exercise is deferred.
- Pass B (document creation/send/status-change writes, template folders, forms, webhooks, API
  usage introspection) is out of scope; see `api_surface.json`.
