# Overview

Mailtrap (`mailtrap.io/api`) is an account-management REST API for email-testing inboxes and
sending domains. This bundle migrates `internal/connectors/mailtrap` (the hand-written legacy
connector) to a declarative Tier-1 bundle at capability parity: 4 read-only streams, no writes.

## Auth setup

Provide a Mailtrap API token via the `api_token` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_token>`) and is never logged. The account-scoped streams (`inboxes`,
`projects`, `sending_domains`) additionally require the `account_id` config value, substituted into
each stream's account-scoped path (`/accounts/{{ config.account_id }}/...`) and stamped onto every
emitted record of those 3 streams via `computed_fields` (matching legacy's `stampAccount` helper).
`accounts` itself is root-scoped and does not carry an `account_id` field, matching legacy exactly.

## Streams notes

- `accounts` (root-scoped, `GET /accounts`): records at the response body root (`records.path: "."`
  — Mailtrap's bare top-level JSON array).
- `inboxes` / `projects` (account-scoped, bare top-level array, `records.path: "."`).
- `sending_domains` (account-scoped, `{"data":[...]}` envelope, `records.path: "data"`) — the one
  stream whose legacy `recordsPath` differs (`"data"` vs. `""` for the other three), matching
  legacy's `mailtrapStreamEndpoints` table exactly.

All 4 streams paginate with `pagination.type: page_number` (`page_param: page`, `size_param:
per_page`, `start_page: 1`), matching legacy's `harvest` loop, which advances a `page` query param
and stops on a short page (`len(records) < pageSize`). **`page_size` is not exposed as config** —
`PaginationSpec.PageSize` is a plain JSON int resolved once at bundle load, with no template/config-
driven override in this engine version (the same static-pagination-field limitation documented in
the auth0/searxng goldens, `docs/migration/conventions.md`). `streams.json`'s `pagination.page_size`
is set to **100**, matching legacy's `mailtrapDefaultPageSize` constant exactly — this static field is
sent as `per_page` on every request, fixture-replayed or live, so it must reflect legacy's real
default rather than a small fixture-only value (a prior revision of this bundle shipped
`page_size: 2`, which would have sent `per_page=2` on every live request too; corrected per the
wave2 sweep's C3 finding). The required 2-page `pagination_terminates` proof (`accounts` stream)
therefore ships a 100-record page 1 (a full page, matching `page_size`) followed by a 1-record short
page 2 — mirroring the auth0/aviationstack goldens' same page-size-realism-over-fixture-brevity
tradeoff, not a shortcut. Legacy's own `mailtrapPageSize` config-driven default (100) is honored by
this static value; the dedupe-by-`id` defensive guard (protecting against an upstream that ignores
`page` entirely and repeats the same page) is not reproduced: the short-page stop condition alone is
what this bundle relies on, which is legacy's PRIMARY stop signal and the one every legacy test
actually exercises (`TestReadInboxesPaginates`) — the id-dedupe guard is a defensive fallback for a
misbehaving upstream, never exercised by any legacy test with a real (non-repeating) fixture.

No stream in this bundle is incremental — Mailtrap account-management objects carry no
last-modified timestamp in legacy's own record mapping, matching `mailtrapStreams()`'s empty
`CursorFields` for every stream.

Legacy enforces no client-side rate limiting, so this bundle declares no `streams.json`
`base.rate_limit` either, matching that (lack of) behavior exactly.

## Write actions & risks

None. Mailtrap is a read-only source in both legacy and this bundle (`capabilities.write: false`) —
legacy's own `Write` method is an unconditional `ErrUnsupportedOperation` stub with no reverse-ETL
surface.

## Known limits

- `page_size` is not config-driven (see Streams notes) — the engine's `page_number` pagination type
  has no mechanism to read a page size from `RuntimeConfig.Config` at request time; only the
  static, bundle-declared `pagination.page_size` (100, matching legacy's default) is ever sent or
  used as the stop threshold.
- Legacy's id-based dedupe guard (defensive against an upstream that ignores the `page` param and
  repeats the same page indefinitely) is not reproduced — the short-page stop condition
  (`len(records) < page_size`) is legacy's actual behavior-driving stop signal and is preserved
  exactly; the dedupe guard only ever fires against a non-conformant upstream response shape no
  legacy test exercises.
- The full Mailtrap API surface (sending messages, message search, inbox cleaning, permissions) is
  out of scope for this wave; see `api_surface.json`'s `excluded` entries.
