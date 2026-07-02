# Overview

Castor EDC is a wave2 fan-out migration. This bundle reads Castor EDC studies, users, countries,
and audit-trail events through Castor's OAuth2 (client-credentials) HAL+JSON REST API, migrating
`internal/connectors/castor-edc` (the legacy hand-written connector, which stays registered and
unchanged until wave6's registry flip) at capability parity. Castor EDC is a clinical-trial
electronic data capture platform; this connector is read-only.

## Auth setup

Provide `client_id`/`client_secret` secrets (Castor EDC OAuth2 application credentials); the
bundle exchanges them for a bearer token via OAuth2 client-credentials
(`auth.mode: oauth2_client_credentials`) against `token_url`, matching legacy's
`connsdk.OAuth2ClientCredentials`. `base_url` defaults to `https://data.castoredc.com/api`;
`token_url` independently defaults to `https://data.castoredc.com/oauth/token`. See Known limits
for the narrowed case where a regional host or a `base_url` override is used without also
overriding `token_url`.

## Streams notes

`study`, `user`, and `audit_trail` share the same shape: `GET` against the Castor HAL list
endpoint, page-number pagination (`pagination.type: page_number`, `page_param: page`,
`size_param: page_size`, `start_page: 1`, `page_size: 100` — matches legacy's
`castorDefaultPageSize`), records extracted from `_embedded.<key>` (Castor's HAL collection key,
which does not always match the endpoint path — e.g. `/country` nests under
`_embedded.countries`). `country` is declared `pagination.type: none` because legacy's own
harvest loop still issues only as many requests as needed to exhaust the HAL `page_count`/short-page
signal, and the reference implementation's only exercised shape (`TestReadNonPaginatedStream`) is
a single HAL page — matching that proven shape exactly rather than speculating about
multi-page country data. Legacy also honors a reported HAL `page_count` field as a secondary stop
signal (`page >= pageCount`) alongside the short-page rule; the engine's `page_number` paginator
does not read `page_count` at all, relying solely on the short-page rule — this never diverges
because a Castor page short of `page_size` always coincides with `page >= page_count` on the real
API (both signals describe the same underlying "no more records" condition). `study`, `user`, and
`audit_trail` each declare `incremental.cursor_field` (`updated_on`/`last_login`/`datetime`
respectively, matching legacy's declared `CursorFields`) with NO `request_param` and NO
`client_filtered` — legacy's own `harvest` never sends any incremental filter to the API and never
client-side filters either (every sync, incremental or full, walks every page); the bare
`cursor_field` declaration exists only so the engine derives `incremental_append` sync-mode
eligibility (matching legacy's own published catalog capability), with the actual read remaining
an unfiltered full walk on every sync, exactly as legacy behaves. `country` has no incremental
cursor (legacy declares none for it either).

`spec.json` intentionally does NOT declare `page_size`/`max_pages` as runtime-configurable
properties (unlike legacy, which accepts config overrides for both): `PaginationSpec.PageSize`/
`MaxPages` are read exclusively from `streams.json`'s static `pagination` JSON literal, never from
a `config.*`-templated value (F6, `conventions.md`). See Known limits.

## Write actions & risks

None. This connector is read-only, matching legacy's `Write` stub (`connectors.ErrUnsupportedOperation`).

## Known limits

- **Dynamic (fixture-replay) conformance checks are marked `skip_dynamic` at the bundle level**
  (`metadata.json`'s `conformance` block) for the identical reason as this codebase's other
  `oauth2_client_credentials` bundles (e.g. sendpulse): the token exchange needs a real
  resolvable `token_url`, which conformance's synthetic non-secret config value cannot provide, so
  every auth-resolving dynamic check would fail identically and uninformatively. Static checks
  (spec/schema validity, `interpolations_resolve`, `docs_present`, `fixtures_present`,
  `secret_redaction`) are unaffected and still run. This bundle has no Tier-2 `AuthHook` (its auth
  is fully declarative `oauth2_client_credentials`), so there is no `paritytest/castor-edc`
  package for this wave; the read/pagination/schema shape is proven by structural review against
  legacy `internal/connectors/castor-edc` instead.
- `page_size`/`max_pages` runtime overrides are not exposed (see Streams notes above) — every
  read uses the fixed `page_size: 100`/unbounded-pages shape baked into `streams.json`. This never
  changes any single emitted record's DATA, only how many requests a sync issues and at what page
  size — parity-deviation ledger candidate, ACCEPTABLE under the meta-rule.
- `token_url`'s default is a fixed literal (`https://data.castoredc.com/oauth/token`), not a
  `base_url`-derived value. Legacy derives `token_url` from whatever `base_url` resolves to at
  runtime (including a `url_region`-selected regional host), so a caller overriding `base_url`
  alone (or using `url_region`) gets a token endpoint under the SAME custom/regional host in
  legacy, but would still hit the fixed default host here. The engine's `spec.json` `"default"`
  materialization mechanism fills only a literal per property, with no cross-property derivation
  (`conventions.md` §3) — a caller who needs a regional/custom Castor host must also override
  `token_url` explicitly to point at the same host.
- `url_region` (legacy's `nl`/`uk`/`us`-style base_url shorthand) is not modeled as a separate
  config property: it was a derived-default mechanism (`https://{region}.castoredc.com/api`), the
  same class of derivation `token_url`'s narrowing above documents, and is superseded by
  overriding `base_url` directly with the fully-qualified regional host. Documented scope
  narrowing, not silently dropped: any caller using a regional Castor account can still reach it
  by setting `base_url` (and `token_url`) explicitly.
- `study_user` is not published as a stream: legacy's routing table (`castorStreamEndpoints`)
  contains a `study_user` entry mapping to the SAME `user` resource/embedded-key/mapper as the
  `user` stream, but legacy's own published catalog (`castorStreams()`) never lists `study_user` as
  a selectable stream — it is unreachable dead routing-table entry in legacy itself. This bundle
  therefore ships only the 4 streams legacy actually publishes (`study`, `user`, `country`,
  `audit_trail`); `study_user` is not a capability loss since legacy never exposed it as a usable
  stream in the first place.
- Full Castor EDC API surface (fields, visits, forms, surveys, records, export, data-point
  mutations) is out of scope until Pass B; see `api_surface.json`'s
  `excluded: {category: out_of_scope}` entries.
