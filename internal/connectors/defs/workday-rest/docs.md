# Overview

Workday REST is a wave2 fan-out declarative-HTTP migration. It reads Workday HCM REST API
resources — workers, organizations, and job profiles — through a conservative endpoint subset
(`GET {base_url}/ccx/api/hcm/v1/{tenant}/...`). This bundle is engine-vs-legacy parity-tested
against `internal/connectors/workday-rest` (the hand-written connector it migrates); the legacy
package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Workday REST bearer access token via the `access_token` secret (required); it is sent as
`Authorization: Bearer <access_token>` (`connsdk.Bearer(token)`, `workday_rest.go:168-172`) and is
never logged. Legacy hard-errors when `access_token` is empty
(`"workday-rest connector requires secret access_token"`); this bundle's `spec.json` makes
`access_token` a required property, so the engine's own config-validation rejects a request before
auth would ever be attempted with a missing secret.

`tenant` (config, required) is substituted as a path segment into every stream's resource URL
(`ccx/api/hcm/v1/{{ config.tenant }}/workers`, matching legacy's `resolveResource`'s `{tenant}`
placeholder substitution and its `tenant == "" || strings.ContainsAny(tenant, "/?#")` rejection —
see workday's docs.md for the identical config-validation-parity note, which applies here
byte-for-byte).

`base_url` defaults to `https://wd2-impl-services1.workday.com` (legacy's own `defaultBaseURL`,
shared with the plain `workday` connector) and may be overridden per-tenant or for tests/proxies.

## Streams notes

All 3 streams (`workers`, `organizations`, `jobs`) are `GET` list endpoints returning records under
a top-level `data` key (`records.path: "data"`, matching legacy's `recordsPath: "data"`).
Pagination is `page_number` (`page_param: page`, `size_param: limit`, matching legacy's
`connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "limit", StartPage: 1}`), stopping on a
short page; `page_size` defaults to 100 (legacy's `defaultPageSize`) and `max_pages` defaults to 1
(legacy's own default when `max_pages` is unset).

`workers` and `jobs` declare `x-cursor-field: updated` for catalog/sync-mode-classification parity
with legacy's `CursorFields: []string{"updated"}` — matching legacy exactly, **no `incremental`
block is declared and no request actually filters or advances by this field**: legacy's `Read`
never references `req.State` or sends any date-range query parameter (verified: no `State`
reference anywhere in `workday_rest.go`). Every read is a full-refresh page-through of the entire
collection on both sides.

`organizations` does NOT declare `x-cursor-field`, even though legacy's own `streamEndpoints`
entry sets `cursorFields: []string{"updated"}` for it (`workday_rest.go:65`) — legacy's
`organizations` field list is `id`/`descriptor`/`type` only, with no `updated` field ever emitted
for this stream. This is a legacy catalog-metadata quirk (the same class as the plain `workday`
connector's `organizations`.docs.md note): `x-cursor-field` must name a property that actually
exists in the same schema (a hard `connectorgen validate` rule), so this bundle omits it here
rather than fabricate a field legacy never emits. No emitted record DATA differs from legacy.

All 3 streams declare `"projection": "passthrough"`. Legacy's `Read` emits the raw API record
verbatim (`return emit(connectors.Record(rec))`, `workday_rest.go:141-143`, fed by
`connsdk.Harvest`'s unfiltered `RecordsAt` decode) with no field-building/filtering —
`streamEndpoints[stream].fields` is consumed only by `Catalog` (`workday_rest.go:101-107`), never
by `Read`. Any real Workday field beyond each stream's narrow catalog schema (e.g. HCM's
richer worker/organization/job payload attributes) survives to the emitted record exactly as
legacy would emit it. Declaring the default `"schema"` projection mode here would silently narrow
every emitted record to the catalog schema's properties — a silent, undocumented parity deviation
from legacy's verbatim passthrough — so `passthrough` is required, matching
`docs/migration/conventions.md`'s projection rule (§3) and the post-wave2 §8 rule 1: legacy's raw
`emit(record)` with no `mapRecord` field-building is the mechanical signal to use `passthrough`.

## Write actions & risks

None. Legacy `Write` always returns `connectors.ErrUnsupportedOperation`; `capabilities.write` is
`false` and this bundle ships no `writes.json`.

## Known limits

- **`docs_url` is unreachable ("manual intervention needed").** Per
  `docs/migration/conventions.md` ("legacy is ground truth over any doc"), this bundle is authored
  entirely from the legacy Go package (`internal/connectors/workday-rest/workday_rest.go`), which
  is itself a conservative, docs-limited implementation. No behavior in this bundle depends on
  unreachable documentation.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`boundedInt`/`readMaxPages`, `workday_rest.go:212-240`, bounded 1-500 for
  `page_size`). The engine's `page_number` paginator's `PageSize`/`MaxPages` fields are plain
  integers with no template/config-driven override mechanism, so neither can be wired to a
  `spec.json` config value; both are fixed in `streams.json`'s `base.pagination` block instead
  (`page_size: 100`, `max_pages: 1`, both matching legacy's own defaults). Neither key is declared
  in `spec.json` (F6, REVIEW.md).
- **Legacy's fixture-mode-only fields (`connector`, `stream`, `fixture`) are not modeled.**
  Legacy's `readFixture` path (only reached when `config.mode == "fixture"`) stamps these 3 extra
  marker fields onto every fixture-mode record (`workday_rest.go:151`). This bundle's schemas and
  fixtures target the live path only; the engine's own conformance/fixture-replay harness provides
  the credential-free test affordance this bundle needs.
- **Single-page fixtures only, matching `max_pages: 1`'s real, always-enforced behavior.** Since
  `page_size`/`max_pages` cannot be config-driven (previous bullet), `max_pages: 1` is a hard,
  unconfigurable cap in this bundle exactly as it is in legacy's own unset-`max_pages` default —
  every stream genuinely only ever fetches one page in practice, so a second fixture page would
  describe a request the connector never issues. This bundle ships single-page fixtures for every
  stream (the identical precedent set by `defs/searxng`'s own `max_pages: 1` streams), rather than
  a misleading 2-page fixture that `pagination_terminates` could never actually reach.
