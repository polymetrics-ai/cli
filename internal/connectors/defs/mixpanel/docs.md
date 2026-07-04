# Overview

Mixpanel is a Tier-2 declarative-bundle-plus-hook migration. It reads the legacy Mixpanel Query API
cohorts, annotations, and engage (profile) records through `/api/2.0`, and selected current
Mixpanel Query/Annotations API list or detail endpoints that fit the same Basic-auth model. This
bundle is read-only and keeps the hand-written connector's record contract for the original three
streams; the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Auth is HTTP Basic, resolved by `hooks/mixpanel/hooks.go`'s `AuthHook` rather than a declarative
`basic` `auth` candidate list, because legacy's `mixpanelCredentials` resolves username and
password through TWO fully INDEPENDENT fallback chains at once:

- username: the non-secret `username` config value if set, else the `username_secret` secret if
  set, else empty.
- password: the `password` secret if set, else the `api_secret` secret if set, else empty.

The engine's declarative `when`-gated auth-candidate-list dialect (conventions.md §3's
dual-auth-ordering pattern) can express a single ordered fallback chain, but `when` supports only
ONE truthiness/equality/membership check per candidate — it has no AND/OR combinator. Reproducing
two independent 2-way fallbacks as an ordered candidate list would require the candidate gating
itself to depend on which SPECIFIC username source AND which SPECIFIC password source are set
together, which `when`'s single-condition grammar cannot express without risking an incorrect
candidate firing first for some legacy-valid combination (e.g. `username` unset,
`password` set, `api_secret` also set — `when: {{ secrets.password }}` alone cannot also confirm
`username` is the RIGHT tier without also checking username's own presence, and vice versa for
every other combination). Rather than accept a documented behavior change for some real,
legacy-accepted credential combination, this bundle resolves both fallbacks in Go, exactly
mirroring `mixpanelCredentials` (`mixpanel/mixpanel.go`), field-for-field. `base_url` defaults to
`https://mixpanel.com/api/2.0` and may be overridden for tests/proxies. None of
`username_secret`/`password`/`api_secret` is ever logged.

## Streams notes

Original legacy-parity streams: `cohorts` (`GET /api/2.0/cohorts/list`, records at `cohorts`),
`annotations` (`GET /api/2.0/annotations`, records at `annotations`), and `engage`
(`GET /api/2.0/engage`, records at `results`). Every original stream sends
`limit={{ config.page_size }}` (default 1000, matching legacy's
`mixpanelDefaultPageSize`), and pagination follows Mixpanel's own `page`/`next` cursor convention
(`pagination.type: cursor`, `cursor_param: page`, `token_path: next`) — the next page's `page`
value is read verbatim from the response body's `next` field, and pagination stops when `next` is
empty or absent, exactly matching legacy's `harvest` loop.

`cohorts` and `annotations` are flat, ordinary-shaped record streams — every field legacy's
`mixpanelCohortRecord`/`mixpanelAnnotationRecord` emits (`id`/`name`/`count`,
`id`/`date`/`description`) is a direct, same-named raw-API field, so plain schema-mode projection
reproduces legacy exactly with no `computed_fields`/hook involvement; `id` keeps its real numeric
wire type (`"integer"`, matching the engine's typed schema-projection passthrough — a plain
schema-projected field, like a bare `{{ record.<path> }}` computed field, copies the raw JSON
value's native type verbatim).

`engage` needs its own `RecordHook` (the SAME `hooks/mixpanel/hooks.go` package, a second hook
interface — still well under the Tier-2 cap of 2): legacy's `mixpanelProfileRecord` resolves
THREE fields through independent multi-source fallback chains — `distinct_id` (`$distinct_id`,
else `distinct_id`), `email` (`$properties.$email`, else top-level `$email`, else top-level
`email`), `created` (`$properties.$created`, else top-level `created`) — the exact same
multi-source-fallback shape `computed_fields`' plain `{{ }}` templating cannot express (no
coalesce/first-of filter exists in this dialect), so it is handled in Go instead, field-for-field
identical to legacy's `first(...)` helper.

None of the 3 streams exposes a legacy-recognized incremental cursor field — Mixpanel's Query API
surfaces cohort/annotation/profile metadata, not an event stream; legacy's own catalog publishes
no `CursorFields` for any stream. All 3 streams are full-refresh only.

Additional read streams cover current documented object-record GET endpoints:

- `saved_funnels` reads `GET https://mixpanel.com/api/query/funnels/list` (records are the response
  array; primary key `funnel_id`).
- `activity_stream` reads `GET https://mixpanel.com/api/query/stream/query` with caller-supplied
  `distinct_ids`, `from_date`, and `to_date`, and emits `results.events`.
- `top_events` reads `GET https://mixpanel.com/api/query/events/top` and emits the `events` array.
- `event_property_names` reads `GET https://mixpanel.com/api/query/events/properties/top`; the
  keyed-object response is flattened into `{name, count}` records.
- `project_annotations`, `project_annotation`, and `annotation_tags` read the current Annotations
  API under `https://mixpanel.com/api/app/projects/{project_id}/annotations...`.

These current-doc streams use fixed standard US hosts because the engine has one connector-wide
base URL and the original `base_url` remains reserved for legacy `/api/2.0` parity. EU/India
regional variants are therefore not modeled for the additional current-doc streams in this bundle.

## Write actions & risks

None. `capabilities.write` is `false`; no `writes.json` is shipped. Legacy itself implements no
write path for Mixpanel (`Write` returns `connectors.ErrUnsupportedOperation`).

## Known limits

- `api_surface.json` enumerates Mixpanel's current OpenAPI surfaces. Mutating/admin products
  (ingestion, identity/profile updates, Lexicon schemas, service accounts, warehouse connector
  management, feature flag management, experiments, GDPR jobs, and data pipelines), raw export
  downloads, scalar-array Query endpoints, and aggregate report endpoints are excluded with typed
  reasons rather than exposed through this read-only legacy Query connector.
- Current Query API `POST /cohorts/list` and `POST /engage` are documented successors to the
  legacy `/api/2.0` read endpoints, but non-excluded POST endpoints are treated by conformance as
  write surface. The legacy-compatible GET streams remain the implemented source of cohort/profile
  record data for this connector.
- Additional current-doc streams use fixed standard-region `mixpanel.com` hosts. This avoids
  changing the meaning of the legacy `base_url` override that still controls the original
  `/api/2.0` reads.
- **Dynamic conformance checks are skipped at the bundle level** (`metadata.json`'s
  `conformance.skip_dynamic`): the sole `auth` candidate is `mode: custom` with no when-gated
  non-custom fallback, so conformance's synthetic non-secret config (which never populates a real
  username/password/api_secret combination) cannot exercise auth resolution at all — every
  auth-dependent dynamic check would otherwise fail identically and uninformatively, exactly
  mirroring gmail's identical precedent. Static checks (spec/schema validity,
  `interpolations_resolve`, docs/fixtures presence, secret redaction) still run and pass. Parity
  for both the credential-fallback AuthHook and the `engage` RecordHook is proven by
  `paritytest/mixpanel` (drives both connectors live against the same `httptest.Server`, asserting
  raw `connectors.Record` equality) and `hooks/mixpanel/hooks_test.go`'s unit coverage.
- Fixtures still ship for all 3 streams (`fixtures/streams/**`, `fixtures/check.json`) satisfying
  the static `fixtures_present` check and documenting the intended wire shape, even though the
  dynamic replay checks that would otherwise consume them are skipped bundle-wide.
