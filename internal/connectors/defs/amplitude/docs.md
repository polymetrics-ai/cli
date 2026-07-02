# Overview

Amplitude is a read-only declarative-HTTP connector for the Amplitude Analytics REST API. It reads
Amplitude behavioral cohorts, chart annotations, and event types. This bundle migrates
`internal/connectors/amplitude` (the hand-written connector); the legacy package stays registered
and unchanged until wave6's registry flip.

## Auth setup

Provide an Amplitude project API key via the `api_key` secret and its paired secret key via the
`secret_key` secret. Both flow only into HTTP Basic auth (`api_key` as username, `secret_key` as
password) and are never logged.

## Streams notes

All 3 streams (`cohorts`, `annotations`, `events_list`) are full-refresh list endpoints with no
pagination and no incremental cursor, matching legacy exactly — each Amplitude list endpoint
returns its full collection in one response. `cohorts` records live at the body's `cohorts` key;
`annotations` and `events_list` records live at `data`. Field names are copied verbatim from the
raw API response (Amplitude's own camelCase field names, e.g. `lastComputed`, `non_active`,
preserved as-is) via schema projection, with no `computed_fields` renames needed since legacy's own
record mappers pass every field straight through unchanged.

## Write actions & risks

Not applicable — this connector is read-only (`capabilities.write: false`), matching legacy: an
analytics API with no safe reverse-ETL write surface.

## Known limits

- EU-residency projects: legacy derives `https://analytics.eu.amplitude.com` automatically from a
  `data_region` config value containing "eu". The engine's `spec.json` `"default"` mechanism only
  materializes a fixed literal default, not one derived from another config key's value (see
  `docs/migration/conventions.md` §3's `default` materialization note — this is the same
  base-URL-derivation gap documented for sentry/chargebee). This bundle narrows the config surface:
  `data_region` is no longer a config option; EU-residency users must set `base_url` to
  `https://analytics.eu.amplitude.com` explicitly. Documented scope narrowing, not a silent
  behavior change — every request legacy would route to the EU host still reaches it, just via an
  explicit `base_url` instead of an inferred one.
- Full Amplitude API surface (cohort membership export, raw event export, taxonomy management,
  dashboards, event ingestion) is out of scope for this wave; see `api_surface.json`'s
  `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}` entries.
