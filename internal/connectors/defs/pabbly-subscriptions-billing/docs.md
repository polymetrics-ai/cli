# Overview

Pabbly Subscriptions Billing is a read-only declarative migration of
`internal/connectors/pabbly-subscriptions-billing` (legacy Go connector). It reads customers,
subscriptions, plans, and invoices from the Pabbly Subscriptions Billing HTTP API. This bundle is
capability-parity with legacy; legacy stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Pabbly account `username` (config) and `password` (secret); they are sent via HTTP Basic
auth (`auth: [{"mode": "basic", "username": "{{ config.username }}", "password": "{{
secrets.password }}"}]`) exactly like legacy's `connsdk.Basic(username, password)`. Legacy
hard-errors when either value is unset (`pabbly-subscriptions-billing connector requires config
username and secret password`), matching this bundle's `required: ["username", "password"]`.

## Streams notes

All 4 streams (`customers`, `subscriptions`, `plans`, `invoices`) share the identical shape: `GET`
against the Pabbly list endpoint, records at `data`, primary key `["id"]`. `customers`,
`subscriptions`, and `invoices` declare `x-cursor-field: created_at` on their schemas (matching
legacy's catalog-only `CursorFields` declaration on those 3 streams); `plans` declares none
(legacy's `plans` stream carries no `CursorFields` either). No stream declares a `streams.json`
`incremental` block: legacy's `Read`/`harvest` never applies a server-side or client-side
incremental filter for any of the 4 streams — the `CursorFields` catalog metadata is descriptive
only, never wired into an actual request parameter — so this bundle matches that real behavior
exactly rather than inventing an incremental filter legacy never had.

Pagination follows legacy's own `next_page`-in-body convention: the response body's `next_page`
field carries the literal value of the NEXT `page` query parameter to send — modeled as
`pagination.type: cursor` with `token_path: next_page` and `cursor_param: page`. Pagination stops
when `next_page` is absent or empty, identical to legacy's `strings.TrimSpace(next) == ""` check;
no `stop_path` is declared since legacy has no separate boolean stop signal beyond the token
itself. `per_page` is sent on every request from the `page_size` config value (default `100`,
matching legacy's `defaultPageSize`) via each stream's `query.per_page` object-form entry
(`default: "100"`).

## Write actions & risks

None. Legacy `Write` always returns `connectors.ErrUnsupportedOperation`; `metadata.json` declares
`capabilities.write: false` and no `writes.json` file exists, matching legacy exactly.

## Known limits

- `page_size`/`max_pages` config validation (legacy's numeric-range and `all`/`unlimited` keyword
  parsing) is not reproduced at the bundle-config level; the engine treats `page_size` as an opaque
  string substituted directly into the `per_page` query param. This never changes emitted record
  DATA for any legacy-valid input; it only narrows client-side input validation, out of scope for
  wave2 fan-out (Pass B).
- No `incremental` block is declared on any stream, matching legacy's real (lack of) incremental
  filtering behavior exactly — not a narrowing. `x-cursor-field` remains declared on 3 of the 4
  schemas purely as catalog/candidate-cursor metadata, mirroring legacy's own `CursorFields` field.
