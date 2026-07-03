# Overview

Box is a read-only source connector. It reads enterprise users, groups, collections, and
folder-scoped items through the Box REST API (`https://api.box.com/2.0`) using the OAuth2
client-credentials grant (Server Authentication). This bundle migrates `internal/connectors/box`
(the hand-written legacy connector); the legacy package stays registered and unchanged until
wave6's registry flip.

## Auth setup

Provide `client_id`/`client_secret` (both `x-secret`) for a Box Server Authentication with
Client Credentials Grant app. The engine's `oauth2_client_credentials` auth mode exchanges them
at `token_url` (`https://api.box.com/oauth2/token` by default) for a short-lived bearer token,
scoped by two Box-specific token-request form params sent via `auth[].extra_params`
(`box_subject_type`, `box_subject_id`) — matching legacy's `authenticator`/`boxSubject` exactly.
`box_subject_type` defaults to `enterprise` (the application service account); set it to `user` to
scope the token to a specific user instead. `box_subject_id` is the enterprise id or user id being
scoped to. `extra_params` values always hard-error on an unresolved config key (never silently
omitted, unlike `stream.Query`'s opt-in tolerance) — see the parity-deviation note below for how
this bundle satisfies that requirement when `box_subject_id` is left unset.

## Streams notes

All 4 streams (`users`, `groups`, `collections`, `folder_items`) share the same shape: `GET`
against a Box list endpoint returning the `{entries:[...], offset, limit, total_count}` envelope
(`records.path: "entries"`), primary key `["id"]`. Pagination is `offset_limit`
(`limit_param: limit`, `offset_param: offset`, `page_size: 100`), stopping on a short/empty page —
identical to legacy's `harvest` loop's `len(records) < pageSize` short-page stop. `folder_items`
reads `/folders/{{ config.folder_id }}/items`, where `folder_id` defaults to `"0"` (the root
folder), matching legacy's `resolveResource`/`boxRootFolderID` default exactly. `users`, `groups`,
and `folder_items` expose `modified_at` as a schema cursor candidate (matching legacy's
`CursorFields`), but Box's list endpoints have no server-side `modified_at` filter parameter in
legacy's own implementation, so no `incremental` block is declared (per §8 rule 2, `x-cursor-field`
stays schema-only) — full refresh only, matching legacy's actual behavior. `collections` has no
cursor field at all (legacy declares `CursorFields: []string{}`).

`check` issues a single bounded `GET /users?limit=1`, mirroring legacy's `Check` implementation
exactly (a 1-record probe of the `users` endpoint confirms auth and connectivity without mutating
anything).

## Write actions & risks

None. Box is read-only (`capabilities.write: false`); legacy's own `Write` always returns
`connectors.ErrUnsupportedOperation` — the upstream manifest source has no reverse-ETL write
target.

## Known limits

- **Conformance dynamic checks are skipped** (`metadata.json`'s `conformance.skip_dynamic`):
  `oauth2_client_credentials` auth's `token_url` is a separate declared `config.token_url`
  property; conformance's replay-server rewiring (`withReplayURL`) only overrides the bundle's
  base request URL used for stream/check paths, never `RuntimeConfig.Config["token_url"]` itself,
  so the token exchange always targets the synthetic non-secret placeholder value
  (`"synthetic-conformance-value"`, not a real URL) and fails before any declarative stream/check
  request is issued — every auth-resolving dynamic check would otherwise fail identically and
  uninformatively. Static checks (spec/schema validity, `interpolations_resolve`, docs/fixtures
  presence, secret redaction) still run and pass. Box has no Tier-2 `AuthHook` (auth is fully
  declarative `oauth2_client_credentials`), so there is no `paritytest/box` package for this wave;
  the read/pagination/schema-projection shape is proven by structural review against legacy
  `internal/connectors/box` instead. Matches `clazar`/`sendpulse`/`kyriba`'s identical documented
  precedent.
- Only the 4 legacy-parity read streams are implemented; see `api_surface.json`. Box's full
  documented API surface (files, folders content, webhooks, tasks, etc.) is out of scope until
  Pass B.
- Documented parity deviation: legacy only sets the `box_subject_id` token-request form param when
  non-empty (`if subjectID != "" { extra.Set(...) }`); the engine's `auth[].extra_params` dialect
  hard-errors on an unresolved config key rather than supporting `stream.Query`'s
  `omit_when_absent` tolerance, so `box_subject_id` declares `spec.json`'s `default`-materialization
  mechanism with no explicit default value set (`""` when the caller never configures it) instead.
  The practical difference is the token request always carries a `box_subject_id=` key (empty
  string) rather than omitting the key entirely when unset — this only affects the OAuth2 token
  exchange request shape, never emitted record data, and Box's token endpoint accepts an empty
  `box_subject_id` in combination with `box_subject_type=enterprise` identically to an absent one
  in practice (the enterprise service account is the default subject either way). See
  `docs/migration/conventions.md`'s parity-deviation ledger.
- `folder_id` must be a Box numeric folder id; legacy's own `validFolderID` digit-only validation
  is not re-implemented as a spec.json pattern constraint (draft-07 `pattern` would work but was
  judged unnecessary extra surface for a config value the API itself will reject loudly if
  malformed) — an invalid `folder_id` surfaces as a Box API error rather than a local validation
  error.
- `users`/`groups`/`folder_items` are full-refresh only (no server-side incremental filter wired in
  legacy), even though `modified_at` is a schema-declared cursor candidate.
