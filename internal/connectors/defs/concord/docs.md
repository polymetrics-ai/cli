# Overview

Concord is a contract lifecycle management platform. This bundle reads and writes Concord contract
lifecycle management data through the Concord REST API: the 5 legacy-parity streams (agreements,
user organizations, folders, reports, tags) plus, as of this Pass B full-surface expansion, 20
additional read streams (organization/folder/report/clause/approval detail lookups, clauses,
approvals, groups, members, events, subscription, branding, automated templates, the authenticated
user's own profile/preferences/webhook-integrations, and 7 agreement sub-resources fanned out from
the agreements list) and 13 write actions (create/update/delete for folders, reports, and clauses;
create/update/delete for company approvals; create for groups). It migrates
`internal/connectors/concord` (the hand-written legacy connector), which stays registered and
unchanged until wave6's registry flip, at parity for the original 5 streams; every Pass B addition
is new coverage with no legacy counterpart to match, verified directly against Concord's published
OpenAPI 3.1 spec (`https://api.doc.concordnow.com/concord-openapi-bundled.yaml`, rendered at the
`docs_url` above).

This bundle was UNBLOCKED from `docs/migration/quarantine.json` once the engine's `page_number`
paginator gained an explicit 0-indexed `start_page: 0` (S4 engine mini-wave item 1) — legacy's
`harvest` deliberately starts at page 0 (`for page := 0; ...; page++`), which the pre-increment
paginator could not express (a zero `start_page` was indistinguishable from an unset one).

## Auth setup

Provide a Concord API key via the `api_key` secret; it is sent as the `X-API-KEY` header
(`api_key_header` auth mode) and is never logged, matching legacy's `connsdk.APIKeyHeader`
wiring exactly. `base_url` defaults to the production host
(`https://api.concordnow.com/api/rest/1`); override it for the UAT environment
(`https://uat.concordnow.com/api/rest/1`) or a test/proxy URL.

## Streams notes

All 5 streams share Concord's page-increment pagination (`pagination.type: page_number`,
`page_param: page`, `start_page: 0`, no `size_param` — legacy sends the page size as `limit`,
which this bundle sends via each stream's own templated `query.limit` entry instead of the
paginator's size-param mechanism): the first request sends `page=0`, matching legacy's
`harvest` loop and `concord_test.go`'s pagination assertions exactly; pagination stops on a
short/empty page (fewer than the effective page size), identical to legacy's
`len(records) < pageSize` check.

`limit` is sent as `{{ config.page_size }}` (default `"100"`, matching legacy's
`concordDefaultPage`) via an opt-in optional-query entry with a `default`, since
`PaginationSpec.PageSize`/`SizeParam` are static (non-templated) fields and cannot carry a
config-driven page size directly — the paginator's own static `page_size: 100` governs only the
short-page stop-detection threshold (mirrors legacy's compile-time `concordDefaultPage` for that
purpose), while the actual `limit` query value sent on the wire is fully config-driven, exactly
matching legacy's runtime-overridable page size end to end.

Three streams are **organization-scoped** (`agreements`, `folders`, `reports`): their `path`
templates `{{ config.organization_id }}` directly. `organization_id` is intentionally NOT in
`spec.json`'s `required[]` (it is only required for these 3 of 5 streams, matching legacy's own
conditional `concordOrgID` check that never runs for `user_organizations`/`tags`) — reading an
org-scoped stream without `organization_id` set hard-errors at path-resolution time with an
unresolved-key error naming `organization_id`, the same practical outcome as legacy's dedicated
`errors.New("concord config organization_id is required...")` check, just raised at a different
point in the call stack.

`user_organizations` reads `user/me/organizations` (records at `organizations`); `tags` reads the
flat `tags` endpoint (records at `tags`); `reports` reads `organizations/{id}/reports` (records at
`reports`); `agreements` and `folders` return a bare top-level array (`records.path: ""`) —
matching each stream's `recordsPath` in legacy's `concordStreamEndpoints` routing table exactly.

None of the 5 streams exposes a legacy-recognized incremental cursor field — Concord only supports
full-refresh sync upstream (legacy's own catalog publishes no `CursorFields` for any stream); all
5 are full-refresh only.

`check` issues a single bounded `GET user/me/organizations`, mirroring legacy's `Check`
implementation exactly (listing the authenticated user's organizations confirms auth and
connectivity without mutating anything and without needing an org id).

### Pass B streams (20 new, all `pagination: {"type": "none"}` overrides unless noted)

Detail lookups (config-driven id, one request each): `organization`
(`GET /organizations/{organization_id}`), `folder`/`folder_agreements`
(`GET .../folders/{folder_id}` and `.../folders/{folder_id}/agreements`, both needing
`config.folder_id`), `report` (`GET .../reports/{report_id}`), `clause`
(`GET .../clauses/{clause_id}`), `approval` (`GET .../approvals/{approval_id}`), `agreement`
(`GET .../agreements/{agreement_uid}`).

List endpoints: `clauses` (`GET .../clauses`, real `offset`/`limit` pagination — modeled as
`offset_limit`, records at `organizationClauses`), `members` (`GET .../members`, real `start`/
`limit` pagination — modeled as `offset_limit` with `offset_param: start`, records at `members`),
`approvals`/`groups` (unpaginated, records at `approvals`/`groups`), `events` (unpaginated,
requires `config.events_start_date`/`events_end_date`, a `yyyy-MM-dd` range Concord caps at 7
days — this bundle sends both as plain required query params and lets Concord's own validation
enforce the range; records at `events`), `subscription`/`branding`/`automated_templates`
(unpaginated org-level singletons/lists), `user_me`/`user_preferences` (the authenticated user's
own profile/settings, no organization scoping), `webhooks_integrations`
(`GET /users/me/integrations/webhooks`, the authenticated user's own webhook integrations).

7 agreement sub-resources use `fan_out.ids_from.request` against
`/organizations/{organization_id}/agreements` (`id_field: "uid"`, matching this bundle's own
pre-existing `agreements` stream schema field — see the discrepancy note under Known limits) and
`into.path_var` to substitute `{{ fanout.id }}` into each sub-resource's path:
`agreement_metadata`, `agreement_summary` (single-object sub-resources, `stamp_field:
agreement_uid`), `agreement_comments` (a keyed-object response — `records.keyed_object: true`,
`key_field: comment_uuid`, `stamp_field: agreement_id`), `agreement_activities`/
`agreement_members`/`agreement_versions`/`agreement_attachments` (array sub-resources,
`stamp_field: agreement_id`). `agreement_members`' real per-record shape has no independent unique
id, so `member_id` is derived via a bare-single-reference `computed_fields` typed extraction
(`{{ record.user.id }}`) and `x-primary-key` is the composite `["agreement_id", "member_id"]`.
Because each sub-resource stream's own `pagination` override (`"type": "none"`) also governs its
fan_out id-listing preliminary request (`fanOutIDsFromRequest` reads `stream.Pagination`, falling
back to `b.HTTP.Pagination` only when the stream declares none at all), the preliminary agreements
listing for these 7 streams issues exactly one unpaginated request, not Concord's real page-by-page
listing — acceptable here since the goal is enumerating every agreement uid to fan out over, not
reproducing the base `agreements` stream's own pagination contract a second time.

## Write actions & risks

13 actions across 5 resource families, all templating `{{ config.organization_id }}` into the
path (`capabilities.write` is now `true`; `metadata.json`'s `risk.write` summarizes the shared
external-mutation exposure):

- **Folders** (`create_folder`/`update_folder`/`delete_folder`): `POST .../folders`,
  `PUT .../folders/{{ record.id }}`, `DELETE .../folders/{{ record.id }}`. `delete_folder` declares
  `delete.missing_ok_status: [404]` (idempotent delete).
- **Reports** (`create_report`/`update_report`/`delete_report`): `update_report`'s body mirrors
  Concord's `ReportDto` PUT contract, which requires `id`/`name`/`description`/`filters` together
  (a full-replace PUT, not a partial patch) — `record_schema` marks all four `required`, matching
  the real API's own requirement, not just this bundle's convenience.
- **Clauses** (`create_clause`/`update_clause`/`delete_clause`): `create_clause` requires
  `title`+`content` (Concord's own `PostOrganizationClauseDto` contract); `update_clause` accepts
  an optional `version` field for optimistic-concurrency, matching `PutOrganizationClauseDto`.
- **Groups** (`create_group` only): Concord's API exposes no single-group DELETE at all (only a
  bulk `PATCH .../groups` membership-update endpoint, out of scope this pass — see
  `api_surface.json`) and no single-group PUT/replace either, so `create_group` is the only
  write action modeled for this resource.
- **Approvals** (`create_approval`/`update_approval`/`delete_approval`): `update_approval` is a
  `POST` (not PUT), matching Concord's own "Update a Company Approval" endpoint shape exactly
  (`POST .../approvals/{{ record.id }}`) — this is Concord's real wire convention, not a bundle
  authoring inconsistency.

Every action's risk mutates shared organization-level CLM configuration (folder structure, saved
reports, reusable clause templates, user groups, approval workflows) but none touches agreement
CONTENT itself (no create/sign/renew-agreement write is modeled — see Known limits).

## Known limits

- **Pre-existing discrepancy, not introduced or fixed this pass**: this bundle's pre-existing
  `agreements` stream reads `GET /organizations/{organization_id}/agreements` with
  `records.path: ""` (a bare top-level array) and a schema field named `uid`. Concord's currently
  published OpenAPI spec documents NO GET method at all on that exact path (only POST, for
  creating an agreement) — the real list-agreements GET lives at
  `/user/me/organizations/{organizationId}/agreements`, returns an `AgreementListDto` envelope
  (`{items: [...], numberOfItems, page, total, ...}`, not a bare array), and each item's real id
  field is `uuid`, not `uid`. This predates Pass B (it is this bundle's original wave1/wave2
  legacy-parity choice, `api_surface.json`'s pre-existing `covered_by` entry) and is left
  UNCHANGED here — Pass B's mandate is full-surface EXPANSION, not an unreviewed behavior change to
  an already-shipped, already-tested stream. The 7 new `fan_out`-driven agreement sub-resource
  streams deliberately reuse this bundle's own `agreements`-stream field convention
  (`id_field: "uid"`) for consistency with what this bundle already ships, not the live API's
  actual `uuid` field — if the base `agreements` stream is ever corrected to the real path/shape in
  a future pass, these 7 fan_out `id_field` declarations should be revisited in the same change.
  `folders`' pre-existing stream has an analogous discrepancy (real folders list returns a
  `FolderTreeDto`, not a bare array) — also left unchanged for the same reason.
- Every remaining documented Concord endpoint not implemented this pass is excluded with a
  specific, closed-vocabulary reason in `api_surface.json` — see that file rather than duplicating
  the full list here. The largest excluded category is agreement-CONTENT mutation (create/patch
  agreement, clause/field/version/member/comment/attachment/signature/renewal actions on a live
  agreement): Concord's `AgreementCreate` schema alone is a deeply nested templating/
  source-agreement/parameter-replacement shape not safely reducible to a flat JSON write action,
  and every signature/renewal/negotiation-membership action is legally consequential — both
  judged out of scope for this pass rather than risking a silent mis-mapping of binding contract
  data. Binary-payload endpoints (PDF/DOCX export, attachment/logo upload) and
  organization-admin-only surfaces (branding, permission groups, integrations, proxy profiles,
  member/invitation management) are excluded as `binary_payload`/`requires_elevated_scope`.
- Legacy's config-overridable `max_pages` (a per-read hard page-count cap, defaulting to
  unlimited) has no engine-dialect equivalent: `PaginationSpec.MaxPages` is a static integer
  declared in `streams.json`, not a templated/config-driven value (the same limitation documented
  for `page_size`'s pagination-block threshold, above — only the per-request `limit` query value is
  config-driven, not the paginator's own construction-time fields). This bundle declares no
  `max_pages` in `streams.json` (absent = unbounded), matching legacy's own default (unlimited)
  behavior for every caller that never overrides it; `max_pages` is not declared in `spec.json`
  either, since a declared-but-unwireable key is worse than an absent one (F6, matching
  cisco-meraki's identical precedent).
- Legacy's `env` config knob (`"uat"` or `"api"`, selecting the host prefix) has no dialect
  equivalent as a derived default (`base_url`'s `spec.json` default is a fixed literal, not a
  function of another config value — see `docs/migration/conventions.md`'s note on derived
  defaults). This bundle instead declares `base_url` with a fixed default of the production host
  and documents overriding it to the full UAT URL directly; `env` itself is not declared in
  `spec.json` (a declared-but-unwireable key is worse than an absent one, per the F6 dead-config
  rule).
- `fixtures/streams/agreements/page_1.json` ships a FULL 100-record page (matching the static
  `page_size: 100` pagination threshold) plus a 1-record `page_2.json`, satisfying
  `conformance`'s `pagination_terminates` 2-page requirement; the other 4 legacy-parity streams
  ship a single-page fixture each (pagination is already proven by `agreements`). All 20 Pass B
  streams declare `pagination: {"type": "none"}` (single-request detail/list lookups), so none
  need a 2-page fixture under the §4 rule; the 7 `fan_out` sub-resource streams each ship exactly
  2 fixture pages regardless — one for the preliminary agreements id-listing request, one for the
  per-agreement sub-resource request itself — which is a fan_out-shape necessity (both requests
  hit the SAME replay server), not a pagination proof.
