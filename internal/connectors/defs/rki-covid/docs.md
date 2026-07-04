# Overview

RKI COVID reads public Germany COVID-19 metrics derived from Robert Koch-Institut reports via the
corona-zahlen.org JSON API. It is credential-free, read-only, and pure Tier-1 declarative HTTP:
there is no auth, no pagination, and no write surface. Pass B expanded the bundle from the original
five legacy-parity streams to the tabular JSON data surface documented under
`https://api.corona-zahlen.org/docs/`.

The legacy connector remains the record-shape ground truth for the original `germany`, `states`,
`districts`, `cases_history`, and `deaths_history` streams: every stream uses
`projection: "passthrough"` and stamps the legacy `stream` marker so vendor fields are preserved.
The two original history streams now use the documented `/germany/history/...` paths; the emitted
records are the same `data[]` history objects legacy reads from the older alias paths.

## Auth setup

None. `streams.json` declares `base.auth: [{"mode":"none"}]`, matching legacy's credential-free
`connsdk.Requester`. `base_url` defaults to `https://api.corona-zahlen.org` and may be overridden for
tests or proxies.

## Streams notes

The bundle declares 28 streams:

- Legacy/current summaries: `germany`, `states`, `districts`.
- Germany histories: `cases_history`, `deaths_history`, `germany_incidence_history`,
  `germany_recovered_history`, `germany_r_value_history`, `germany_hospitalization_history`,
  `germany_frozen_incidence_history`.
- Age-group resources: `germany_age_groups`, `states_age_groups`, `districts_age_groups`.
- State histories: `states_cases_history`, `states_deaths_history`, `states_incidence_history`,
  `states_recovered_history`, `states_frozen_incidence_history`, `states_hospitalization_history`.
- District histories: `districts_cases_history`, `districts_deaths_history`,
  `districts_incidence_history`, `districts_recovered_history`,
  `districts_frozen_incidence_history`.
- Testing and vaccinations: `testing_history`, `vaccinations`, `vaccinations_states`,
  `vaccinations_history`.

Object-keyed expanded responses use `records.keyed_object` with the documented key stamped onto
`abbreviation`, `ags`, or `age_group`. The legacy `states` and `districts` streams intentionally
keep the original `connsdk.RecordsAt(..., "data")` behavior: each emits the whole keyed object as
one passthrough record with `id` set to the stream name. History-array responses use the documented
`data` or `data.history` arrays and compute `id` from `date`. State/district history collection
endpoints emit one record per state or district with the raw `history` array preserved, because the
current dialect does not need to fan out inner history rows to satisfy a distinct documented
collection resource.

The optional `days` config value is still sent as a `days` query parameter on every stream when set.
That preserves legacy behavior exactly: legacy builds one shared query map before calling each
endpoint, so it sends `days` even to endpoints where the docs do not mention that query parameter.
The documented `/{days}` and `/{weeks}` path variants are treated as narrower aliases in
`api_surface.json`, not separate streams.

## Write actions & risks

None. The public API is read-only for this connector, and legacy's `Write` returns
`connectors.ErrUnsupportedOperation`. `capabilities.write` remains `false` and no `writes.json` is
present.

## Known limits

- Path-parameter-limited aliases such as `/states/{state}`, `/districts/{ags}`,
  `/germany/history/cases/{days}`, and `/testing/history/{weeks}` are excluded as `duplicate_of`.
  The broader collection streams cover the same record shapes without requiring global synthetic
  config values for each narrow alias.
- Section/default redirect paths such as `/germany/history`, `/states/history`,
  `/districts/history`, and `/map` are excluded as `duplicate_of`; their concrete targets are listed
  separately in `api_surface.json`.
- Map and legend endpoints are excluded as `non_data_endpoint` because they are visualization
  payloads for choropleth rendering, not tabular ETL records.
- `page_size` remains intentionally absent. Legacy validates `config.page_size` but never uses the
  value in a request; carrying a no-op config key into the declarative spec would be misleading.
