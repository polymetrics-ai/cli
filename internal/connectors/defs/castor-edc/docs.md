# Overview

Castor EDC is a wave2 fan-out migration, expanded in Pass B to the full documented API surface.
This bundle reads Castor EDC studies, users, countries, audit-trail events, and (once a `study_id`
is configured) the full study-scoped clinical-trial data model — records, fields, forms/reports,
sites, roles, surveys, survey packages/instances, phases, steps, queries, verifications, and
record-progress — through Castor's OAuth2 (client-credentials) HAL+JSON REST API. It also writes 6
study-scoped mutations (record/site/role creation, survey-package-instance dispatch, report-instance
creation, and record randomization). Legacy `internal/connectors/castor-edc` (which stays
registered and unchanged until wave6's registry flip) is read-only and covers only 4 streams;
this bundle now covers substantially more of Castor's real surface than legacy ever did, since
Pass B's scope is full-surface expansion, not legacy parity alone. Castor EDC is a clinical-trial
electronic data capture (EDC/CDMS) platform — every write here mutates real clinical-trial study
data and is gated behind approval (see Write actions & risks).

Endpoint research source: Castor's interactive Swagger UI at `https://data.castoredc.com/api`
requires an authenticated study session and cannot be crawled anonymously; the endpoint inventory
below is instead derived from Castor's own officially-linked open-source Python client
(`reiniervlinschoten/castoredc_api`, referenced from Castor's help-center API article), which
implements every endpoint exercised here 1:1 against the real API and is the closest available
authoritative source of truth. See `api_surface.json`'s `scope` field.

## Auth setup

Provide `client_id`/`client_secret` secrets (Castor EDC OAuth2 application credentials); the
bundle exchanges them for a bearer token via OAuth2 client-credentials
(`auth.mode: oauth2_client_credentials`) against `token_url`, matching legacy's
`connsdk.OAuth2ClientCredentials`. `base_url` defaults to `https://data.castoredc.com/api`;
`token_url` independently defaults to `https://data.castoredc.com/oauth/token`. See Known limits
for the narrowed case where a regional host or a `base_url` override is used without also
overriding `token_url`. Every study-scoped stream/write (everything except `study`, `user`,
`country`, `audit_trail`) additionally requires the `study_id` config value — the study to scope
requests to, matching legacy `castoredc_api`'s `CastorClient.link_study(study_id)` session model.
`study_id` is declared optional in `spec.json` (the 4 account-level streams work without it) but a
study-scoped stream/write with no `study_id` configured hard-errors on the unresolved
`{{ config.study_id }}` path segment — an ordinary required-config error, not a silent omission.

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

**Pass B study-scoped streams** (all new): `records`, `fields`, `field_dependencies`,
`field_optiongroups`, `field_validations`, `sites`, `study_metadata`, `metadata_types`, `phases`,
`queries`, `reports`, `report_instances`, `steps`, `surveys`, `survey_packages`,
`survey_package_instances`, `verifications`, and `record_progress` share `study`/`user`/
`audit_trail`'s page-number pagination shape (`page`/`page_size`, `page_size: 100`), matching
`castoredc_api`'s own `retrieve_multiple_pages` page-number walk. `roles` and `study_users` declare
`pagination.type: none`, matching `castoredc_api`'s own `all_roles`/`all_users_study` helpers,
which read the HAL `_embedded` collection directly with no page loop at all (Castor's role/study-user
lists are bounded by study membership size, never paginated in the reference client). None of the
new streams declare an `incremental.cursor_field` — Castor's own client exposes no
`created_on`/`updated_on`-shaped field consistently across this set (several, e.g. `roles`/`steps`/
`field_dependencies`, have no timestamp field in their documented response shape at all), so every
sync of these streams is a full walk, matching the one real precedent (legacy's `study`/`user`/
`audit_trail` full-walk-always behavior) rather than inventing an incremental capability Castor's
API surface doesn't support for these resources. `fields` sends a static `include:
metadata,validations,optiongroup` query parameter, matching `castoredc_api.all_fields`'s own fixed
`params={"include": "metadata,validations,optiongroup"}` (Castor's field endpoint requires this to
return validation/optiongroup detail inline rather than as separate lookups). `report_instances`
sends a static `archived: 0` query parameter (non-archived only), matching
`castoredc_api.all_report_instances`'s own default `archived=0`.

## Write actions & risks

Six study-scoped write actions were added in Pass B, all requiring `study_id` and all gated by
approval (`metadata.json`'s `capabilities.write: true`, `risk.write`). None of these existed in
legacy, which is entirely read-only (`Write` stub returning
`connectors.ErrUnsupportedOperation`) — these are genuinely new capability, not a parity port, so
every action is conservatively scoped to non-destructive creation/dispatch mutations only (no
update/delete write was added; see `api_surface.json` for the destructive/elevated-scope
mutations deliberately excluded):

- `create_record` (`POST /study/{study_id}/record`) — creates a new clinical-trial study
  participant record under an institute/site, given an `institute_id` and `email_address`.
- `create_site` (`POST /study/{study_id}/site`) — creates a new study site/institute.
- `create_role` (`POST /study/{study_id}/role`) — creates a new study access-control role with an
  explicit permissions object (matching Castor's documented boolean permission-flag shape:
  add/view/edit/delete/lock/query/export/randomization_read/sign/email_addresses/
  randomization_write/sdv/survey_send/survey_view).
- `create_survey_package_instance` (`POST /study/{study_id}/surveypackageinstance`) — dispatches a
  survey package invitation to a study participant record by email.
- `create_report_instance` (`POST /study/{study_id}/record/{record_id}/report-instance`) — creates
  a new report-form instance for a study participant record; `record_id` is a `path_fields` entry
  (carried in the URL, not the body).
- `create_randomization` (`POST /study/{study_id}/record/{record_id}/randomization`) — randomizes
  a study participant record into a trial arm; `body_type: none` (Castor's randomization endpoint
  takes an empty POST body — the randomization itself is driven entirely by the study's own
  randomization configuration, not caller-supplied data). This is the single irreversible action in
  this bundle's write surface (a randomized record cannot be un-randomized through the API) and is
  documented as such in its `risk` string.

Every other documented Castor mutation (data-point value writes, econsent, device tokens, user
invitation/permission management, survey-package-instance lock/unlock, bulk report-instance
creation) is excluded from `writes.json` — see `api_surface.json` for the specific
category+reason per endpoint (mostly `out_of_scope` for study-form-defined dynamic body shapes this
dialect cannot express as a fixed `record_schema`, and `requires_elevated_scope`/
`destructive_admin` for study user-provisioning actions).

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
- **Study-form-defined data-point endpoints are out of scope** (`data-point`/
  `data-point-collection` under `/record/{id}/...`): every one of these resources' record shape is
  entirely defined by that study's own dynamically-configured form fields (field_id/field_value
  pairs whose real type varies per study, per field), which this bundle cannot express as a fixed
  `schemas/<stream>.json`/`record_schema` — the genuine `SCHEMA_AMBIGUOUS` shape this dialect has no
  per-study `dynamic_schema` capability to cover. See `api_surface.json` for the itemized list.
- **CSV export endpoints are out of scope** (`/export/data`, `/export/structure`,
  `/export/optiongroups`): these return `Content-Type: text/csv`, not a JSON list envelope this
  bundle's declarative HTTP dialect can decode/paginate/project (`binary_payload` category).
- **`/statistics` is excluded as a non-data endpoint**: it returns aggregate study-level counters,
  not enumerable per-record rows with a stable primary key.
- User-provisioning mutations under `/study/{study_id}/user/{user_id}` (invite/update-permissions/
  delete) are excluded (`requires_elevated_scope`/`destructive_admin`): these manage study team
  membership and permissions, a materially different risk class from the clinical-data creation
  writes this bundle does implement.
- Device-token and econsent endpoints are excluded as `out_of_scope`: narrow, separate-lifecycle
  sub-features (mobile app pairing, electronic consent tracking) with no read/write parity
  precedent and low value relative to the core clinical-data surface covered here.
- `single_*`-by-id detail endpoints (single study/user/country/record/field/site/role/etc.) are
  excluded as `duplicate_of` their corresponding list stream: every field the detail endpoint
  returns is already present on the list stream's records, so a dedicated single-record stream
  would add a second, redundant HTTP shape for the exact same data.
