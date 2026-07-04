# Overview

Akeneo is a Tier-2 (AuthHook) migration of the legacy hand-written `internal/connectors/akeneo`
connector, expanded in Pass B to the full documented Akeneo PIM REST API surface
(api.akeneo.com/api-reference.html). It reads Akeneo PIM `products`, `categories`, `families`,
`attributes`, `channels`, `product_models`, `family_variants`, `attribute_groups`,
`association_types`, `locales`, `currencies`, and `measure_families`, and writes create-or-update
upserts for the 9 catalog-structure resources, via the Akeneo REST API v1, authenticating with
Akeneo's **OAuth2 password grant**: HTTP Basic `client_id:secret` plus a JSON body
`{grant_type:password, username, password}` POSTed to `/api/oauth/v1/token`, exchanged for a
short-lived Bearer access token used on every subsequent request. This bundle originally was
engine-vs-legacy parity-tested against `internal/connectors/akeneo` (the hand-written connector it
migrates, itself read-only) at exact capability parity; the legacy package stays registered and
unchanged until wave6's registry flip, so this bundle's write surface and 7 new streams are a
genuine capability expansion beyond legacy, not a parity port —
`internal/connectors/paritytest/akeneo/parity_test.go` was updated accordingly to assert the new,
wider surface instead of legacy's original 5-stream/read-only shape (see that file's own comments,
and Known limits below).

This connector was previously quarantined (`docs/migration/quarantine.json`, blocker
`AUTH_COMPLEX`): "the engine's declarative auth dialect only supports oauth2_client_credentials
(client_credentials grant)... [not] password grant." That blocker is resolved by expressing the
password-grant exchange as a Tier-2 `hooks/akeneo` `AuthHook` instead of forcing it into a
declarative `oauth2_client_credentials` mode that cannot represent it — the same escape hatch
gmail (OAuth2 refresh-token grant) and jamf-pro (Basic-credential token exchange) already use for
non-`oauth2_client_credentials`-shaped token exchanges.

## Auth setup

Provide `base_url` (the Akeneo PIM host, e.g. `https://example.trial.akeneo.cloud`), `client_id`
and `api_username` (both non-secret config, matching legacy's own `cfg.Config["client_id"]`/
`cfg.Config["api_username"]` classification even though both flow into the token exchange), and
two secrets: `secret` (the OAuth2 client secret) and `password` (the API user's password) — never
logged. `hooks/akeneo/hooks.go` implements `AuthHook`, mirroring legacy
`akeneo.go`'s `passwordGrantAuth` almost verbatim: it POSTs `{"grant_type":"password",
"username":<api_username>,"password":<password>}` (JSON body) with an `Authorization: Basic
base64(client_id:secret)` header to `{{ config.base_url }}/api/oauth/v1/token`, caches the
resulting access token until 60 seconds before its declared `expires_in`, and sets
`Authorization: Bearer <access_token>` on every request.

`base_url` must be an absolute `http://` or `https://` URL with a host (matching legacy's own
`akeneoBaseURL` tolerance for both schemes — legacy's own test suite exercises a plain-`http`
`httptest.Server` token/resource endpoint, so the hook does not narrow this to https-only the way
gmail's hook does for the unrelated Google OAuth endpoint); any other scheme, or a host-less URL,
fails closed before any request is attempted.

The bundle's `base.auth` declares exactly one candidate: `{"mode": "custom", "hook": "akeneo",
...}` — legacy has no alternate auth path (no static API key, no public/no-auth fallback), so
there is no `when`-gated bypass to declare, matching gmail's identical single-candidate shape.

## Streams notes

Twelve streams, all primary-keyed on `id`: the original 5 (`products`, `categories`, `families`,
`attributes`, `channels`) plus 7 Pass B additions (`product_models`, `family_variants`,
`attribute_groups`, `association_types`, `locales`, `currencies`, `measure_families`). Every stream
reads `GET /api/rest/v1/<resource>` with `records.path: "_embedded.items"` (Akeneo's HAL list
envelope) and sends `limit={{ config.page_size }}` (default 100, matching legacy's
`akeneoDefaultPageSize`/`akeneoMaxPageSize` of 100). Pagination is declared once at
`base.pagination` (`type: next_url`, `next_url_path: "_links.next.href"`) since Akeneo's HAL
convention is identical across every resource: the next page is fetched by following
`_links.next.href` verbatim (an absolute URL), matching legacy `harvest`'s own loop
(`akeneo.go:141-181`) exactly — the engine's `next_url` paginator re-merges the stream's static
`limit` query param onto the absolute next URL (`engine/read.go`'s `mergeQuery` +
`connsdk.Requester`'s `resolveURL` Del+Add re-apply), which is a no-op replace since Akeneo's own
`_links.next.href` already carries the identical `limit` value the engine re-applies (see bitly's
docs.md for the identical, already-accepted divergence pattern with legacy connectors that reset
their query to empty once following an absolute next URL — legacy akeneo's own `harvest` does NOT
reset the query to nil in that sense, it sets `query = nil` and relies on the followed URL alone,
so this bundle's behavior of re-sending `limit` is a strict superset, never a different value,
since Akeneo's HAL `next` link is generated by the server FROM the current `limit`).

`computed_fields` derive each stream's `id` with the same fallback chain legacy's shared
`akeneoCode` helper uses: `{{ coalesce record.code record.identifier record.uuid }}`. Products
normally resolve through `identifier`; the other original streams normally resolve through `code`;
the `uuid` fallback is preserved for malformed-but-decodable product records exactly as legacy did.

`product_models`/`family_variants`/`attribute_groups`/`association_types`/`locales`/`currencies`/
`measure_families` are all real, practical GET list endpoints in Akeneo's documented surface with
the identical HAL envelope/pagination shape as the original 5 — no new dialect feature was needed
to add them. Reference Entities and Asset Manager sub-resources (Enterprise Edition features,
parent-scoped under a reference-entity/asset-family code) are NOT modeled as streams: this
connector's `spec.json` has no config surface enumerating which reference-entity/asset-family codes
to sync, and a fan-out over "every reference-entity/asset-family code" would need its own parent
list-all endpoint request first — out of scope for this pass (`api_surface.json`'s `out_of_scope`
entries).

## Write actions & risks

Pass B flips `capabilities.write` to `true` (a genuine capability expansion beyond legacy, which was
fully read-only). 9 actions in `writes.json`, ALL a single-record `PATCH /api/rest/v1/<resource>/
{{ record.id }}` (`path_fields: ["id"]`, `kind: "upsert"`) — Akeneo's own documented API convention:
this single idempotent PATCH creates the resource (`201 Created`) if `id` doesn't yet exist, or
updates it (`204 No Content`) if it does, with no separate POST-then-PATCH split needed:

- `create_or_update_product`, `create_or_update_category`, `create_or_update_family`,
  `create_or_update_attribute`, `create_or_update_channel` (the original 5 streams' write
  counterparts).
- `create_or_update_product_model`, `create_or_update_family_variant`,
  `create_or_update_attribute_group`, `create_or_update_association_type` (Pass B streams' write
  counterparts — `locales`/`currencies`/`measure_families` have no per-record write endpoint in
  Akeneo's API at all, see Known limits, so those 3 streams are read-only).

Every action uses `body_type: "json"` with `path_fields: ["id"]` excluding the id already in the
path. No action needs a hook: every one of these operations is a single JSON HTTP request using the
SAME already-resolved `custom` AuthHook bearer token every read stream uses, no additional auth
complexity, multipart body, or compound follow-up call. Bulk multi-record PATCH forms (newline-
delimited JSON, several records per call) and POST-only identifier-less create forms are excluded
as `duplicate_of` in `api_surface.json` — the single-record PATCH already covers the same reachable
outcome, matching the engine's one-request-per-record write semantics (conventions.md §3). Product/
product-model/family delete, and every Enterprise-Edition-gated write (reference entities, asset
manager, rule definitions, catalogs, workflows, UI extensions), are excluded per-endpoint in
`api_surface.json` (`destructive_admin`/`requires_elevated_scope`/`binary_payload`/`out_of_scope`).

## Known limits

- **`base_url` is now required** (no config-time default): legacy accepted either `config.host` or
  `config.base_url`, with `host` as an earlier-checked alias. This bundle only wires `base_url`
  (the `spec.json` field name every other bundle's host-configuration property uses) — `host` is
  not modeled as a spec.json alias. Every other bundle in this repo (stripe, bitly, gmail, ...)
  names this property `base_url` only; carrying an `host`-vs-`base_url` alias forward would be a
  one-off inconsistency with no corresponding engine dialect support for property aliasing. This is
  a config-surface narrowing (a caller must use `base_url`), not an emitted-record-data change.
- **`max_pages` is not modeled as a config-driven override.** Legacy exposes a `max_pages` request
  cap, but the declarative `next_url` paginator's `max_pages` value is fixed in `streams.json`, not
  templated from `config.*`. The bundle therefore keeps legacy's default unbounded page walk and
  does not declare a dead `max_pages` config property.
- **`TestConformance/akeneo`'s dynamic (fixture-replay) checks are genuinely `skip_dynamic`'d,
  including the new `write_request_shape:<action>` checks Pass B adds** for the identical reason
  gmail's are (per the bundle-level `skip_dynamic` marker's own documented widened scope,
  `conformance/dynamic.go`'s `runDynamicChecks`, gmail Pass B precedent): the sole auth candidate is
  `mode: custom`, and conformance's synthetic non-secret config (`"synthetic-conformance-value"` for
  every non-x-secret property, a fixed placeholder for every x-secret property) can never resolve to
  a working Akeneo token exchange — every auth-resolving dynamic check, read or write, would fail
  identically and uninformatively regardless of hook wiring. All 9 write actions still ship a
  `fixtures/writes/<action>.json` fixture (matching gmail's own precedent of shipping full fixtures
  despite the skip) for documentation and future-proofing if the marker is ever lifted.
  `paritytest/akeneo` (which wires the real `AuthHook` via `engine.HooksFor("akeneo")`, matching
  gmail/monday's precedent) is the authoritative parity/correctness bar for this connector's auth,
  pagination, AND write-request-construction paths.
- **`pagination_terminates`'s dynamic check uses a genuine, single-page conformance fixture per
  stream** (the sanctioned `next_url` exception, `docs/migration/conventions.md` §4): a
  `next_url` stream's next-page URL is the replay server's own runtime address, unknown to a
  static fixture file, so every stream's fixture here declares no `_links.next` (an empty/absent
  next link, matching Akeneo's own last-page shape) and stops after one page. Real 2-page
  `next_url` correctness is proven live by `paritytest/akeneo`'s
  `TestParityAkeneo_ProductsStreamPaginatesAcrossHALPages`, which drives an actual
  `httptest.Server` serving a genuine `_links.next.href` absolute URL on page 1.
- **`locales`/`currencies`/`measure_families` are read-only streams.** Akeneo's documented API has
  no per-record write endpoint for any of the three: locales/currencies have no write endpoint at
  all (enable/disable happens only via each channel's `currencies`/`locales` array, already covered
  by `create_or_update_channel`), and measure families expose only a bulk multi-family PATCH with no
  single-code form the engine's one-request-per-record write semantics can target.
- **Reference Entities and Asset Manager (Enterprise Edition features) are not modeled at all** —
  neither as streams nor writes. Every sub-resource is parent-scoped under a reference-entity/
  asset-family code, and this connector's `spec.json` has no config surface enumerating which codes
  to sync/write; see `api_surface.json`'s `out_of_scope` entries for the full sub-resource list.
