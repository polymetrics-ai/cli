# Overview

Box Data Extract is a read-only source connector. It reads Box folder items through the Box REST
API (`https://api.box.com/2.0` by default) using the OAuth2 client-credentials grant. This bundle
migrates `internal/connectors/box-data-extract` (the hand-written legacy connector); the legacy
package stays registered and unchanged until wave6's registry flip. Legacy is pure
`connsdk`-HTTP with no signature auth and no writes, so it maps to a Tier-1 declarative bundle with
zero Go — matching `internal/connectors/defs/box`'s identical `oauth2_client_credentials` +
`extra_params` pattern (a distinct connector/catalog from the `box` bundle: different name,
different single stream).

## Auth setup

Provide `client_id`/`client_secret` (both `x-secret`) for a Box Server Authentication with Client
Credentials Grant app. The engine's `oauth2_client_credentials` auth mode exchanges them at
`token_url` (`https://api.box.com/oauth2/token` by default) for a short-lived bearer token, scoped
by two Box-specific token-request form params sent via `auth[].extra_params` (`box_subject_type`,
`box_subject_id`) — matching legacy's `requester`/`box_subject_type`/`box_subject_id` construction
exactly. `box_subject_type` defaults to `enterprise` (the application service account); set it to
`user` to scope the token to a specific user instead. `box_subject_id` is the enterprise id or user
id being scoped to.

## Streams notes

`files` reads `/folders/{{ config.box_folder_id }}/items` (`box_folder_id` defaults to `"0"`, the
root folder — matching legacy's `folderItemsResource`/root-folder default exactly), records at
`entries` (Box's `{entries:[...], offset, limit, total_count}` envelope), primary key `["id"]`.
Pagination is `offset_limit` (`limit_param: limit`, `offset_param: offset`, `page_size: 100`,
matching legacy's `defaultPageSize`), stopping on a short/empty page — identical to legacy's
`readOffset` loop's `len(records) < pageSize` short-page stop.

`check` issues a single bounded `GET /folders/{{ config.box_folder_id }}/items?limit=1`, mirroring
legacy's `Check` implementation exactly (a 1-item probe of the configured folder confirms auth and
connectivity without mutating anything).

## Write actions & risks

None. Box Data Extract is read-only (`capabilities.write: false`); legacy's own `Write` always
returns `connectors.ErrUnsupportedOperation`.

## Known limits

- **Conformance dynamic checks are skipped** (`metadata.json`'s `conformance.skip_dynamic`):
  `oauth2_client_credentials` auth's `token_url` is a separate declared `config.token_url`
  property; conformance's replay-server rewiring (`withReplayURL`) only overrides the bundle's base
  request URL used for stream/check paths, never `RuntimeConfig.Config["token_url"]` itself, so the
  token exchange always targets the synthetic non-secret placeholder value
  (`"synthetic-conformance-value"`, not a real URL) and fails before any declarative stream/check
  request is issued — every auth-resolving dynamic check would otherwise fail identically and
  uninformatively. Static checks (spec/schema validity, `interpolations_resolve`, docs/fixtures
  presence, secret redaction) still run and pass. This bundle has no Tier-2 `AuthHook` (auth is
  fully declarative `oauth2_client_credentials`), so there is no `paritytest/box-data-extract`
  package for this wave; the read/pagination/schema-projection shape is proven by structural review
  against legacy `internal/connectors/box-data-extract` instead. Matches `box`'s/`clazar`'s/
  `sendpulse`'s/`kyriba`'s identical documented precedent.
- **`file_text` is NOT migrated (documented scope narrowing, not an `ENGINE_GAP`).** Legacy's
  `Read` for `file_text` unconditionally returns an error outside fixture mode
  (`"box-data-extract file_text live read requires fixture mode or a safe extraction service"`) —
  there is no live Box endpoint being called for this stream at all in legacy; it exists only to
  satisfy `Catalog`'s 2-stream shape and fixture-mode conformance-style tests. Because legacy itself
  never emits a real `file_text` record outside fixture mode, omitting the stream from this bundle's
  `streams.json`/catalog changes no real accepted-input behavior for any live caller — there was
  never a working live path to preserve. This deviates from `discord`'s otherwise-similar
  `members`-omission precedent in one respect: `discord`'s omitted stream has a genuine (just
  hard-to-express) legacy implementation and is tracked as an `ENGINE_GAP`; `file_text` has no
  legacy implementation to express at all, so no blocker is filed for it. See
  `api_surface.json`'s excluded entry for this stream.
- Only the Box folder-items surface is implemented; Box's full documented API surface (file
  content download, webhooks, tasks, users/groups outside a folder scope) is out of scope until
  Pass B — see `internal/connectors/defs/box` for the sibling bundle covering users/groups/
  collections.
- `box_folder_id` must be a Box numeric folder id; legacy's own digit-only validation is not
  re-implemented as a `spec.json` pattern constraint (judged unnecessary extra surface for a config
  value the API itself will reject loudly if malformed) — an invalid `box_folder_id` surfaces as a
  Box API error rather than a local validation error.
- Documented parity deviation: legacy only sets the `box_subject_id` token-request form param when
  non-empty (`if id := ...; id != "" { extra.Set(...) }`); the engine's `auth[].extra_params`
  dialect hard-errors on an unresolved config key rather than supporting `stream.Query`'s
  `omit_when_absent` tolerance, so `box_subject_id` relies on `spec.json`'s
  `default`-materialization mechanism with no explicit default value set (`""` when the caller
  never configures it) instead. The practical difference is the token request always carries a
  `box_subject_id=` key (empty string) rather than omitting the key entirely when unset — this only
  affects the OAuth2 token exchange request shape, never emitted record data, and Box's token
  endpoint accepts an empty `box_subject_id` in combination with `box_subject_type=enterprise`
  identically to an absent one in practice (the enterprise service account is the default subject
  either way). Identical to `box`'s own documented deviation. See
  `docs/migration/conventions.md`'s parity-deviation ledger.
- `files` is full-refresh only; legacy declares no cursor field for this stream (`Catalog`'s
  `files` entry has no `CursorFields`), so no `x-cursor-field` is declared in `schemas/files.json`
  either.
