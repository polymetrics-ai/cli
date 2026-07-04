# Overview

Qualaroo is a declarative HTTP bundle for the Qualaroo API v1. It keeps the two read streams that
the legacy Go connector emits (`nudges`, `responses`) and adds the current documented REST
Reporting API response stream (`survey_responses`) from
`https://help.qualaroo.com/the-rest-reporting-api`. The old metadata URL
`/hc/en-us/articles/201969438-The-Qualaroo-API` now returns a 404, so the bundle's `docs_url` points
at the live Reporting API article.

The legacy package under `internal/connectors/qualaroo` remains read-only and registered until the
wave6 registry flip. This bundle also remains read-only: the current public Reporting API
documentation describes response retrieval, not REST writes.

## Auth setup

Provide `api_key` as a secret for the legacy `nudges` and `responses` streams. The bundle sends it
with the same header shape as legacy:

```text
Authorization: Token token="<api_key>"
```

For the documented Reporting API stream, also provide `api_secret`. When `api_secret` is present,
the engine selects HTTP Basic auth with `api_key` as the username and `api_secret` as the password,
matching the Reporting API documentation. Both values are marked `x-secret: true`.

`base_url` defaults to `https://api.qualaroo.com/api/v1` and may be overridden for tests or local
proxies. `survey_id` is used only by `survey_responses` and is interpolated into
`/nudges/{survey_id}/responses.json`.

## Streams notes

- `nudges` reads `GET /nudges`, extracting records from the `nudges` response key. Its field
  projection intentionally matches legacy's `nudgeRecord` mapper: `id`, `name`, `status`,
  `created_at`, and `updated_at`.
- `responses` reads `GET /responses`, extracting records from the `responses` response key. Its
  field projection intentionally matches legacy's `responseRecord` mapper: `id`, `nudge_id`,
  `email`, `created_at`, and `updated_at`.
- `survey_responses` reads the documented Reporting API endpoint
  `GET /nudges/{survey_id}/responses.json`. Qualaroo documents a root JSON array of response
  objects containing respondent metadata, answered questions, and custom properties, so the stream
  uses root-array extraction (`records.path: ""`) and `projection: "passthrough"` to preserve the
  full response object.

The two legacy streams inherit the base `page_number` paginator (`page`/`per_page`, page size 100),
matching the legacy connector's default page size. `nudges` has a 100-record first fixture page and
a short second page so conformance proves the page-number paginator terminates.

`survey_responses` overrides pagination with `offset_limit` (`offset`/`limit`, page size 500),
matching the Reporting API documentation's offset and limit parameters. Optional date and order
parameters are not declared in this pass because the engine's conformance runtime synthesizes values
for every declared spec property; adding those optional filters would either send invalid synthetic
Reporting API values in fixtures or require a hook solely for fixture-only sanitization.

## Write actions & risks

None. Qualaroo's legacy connector implements no writes (`Write` returns
`connectors.ErrUnsupportedOperation`), and the current Reporting API documentation does not publish
REST write endpoints. `capabilities.write` is `false` and no `writes.json` is shipped.

## Known limits

- **Auth is selected bundle-wide.** The engine chooses one auth candidate for the whole runtime, not
  per stream. Supplying `api_secret` selects the documented Reporting API Basic auth, which is the
  intended mode for `survey_responses`; omitting `api_secret` keeps the legacy token-header mode for
  `nudges` and `responses`.
- **Check remains legacy-shaped.** `base.check` still calls `GET /nudges` because that is the
  stable legacy health check. A future engine feature for stream-specific checks could let
  `survey_responses` use a Reporting API check without changing the legacy check path.
- **Fallback field names are not modeled.** Legacy's mappers accept alternate keys (`title` for a
  nudge name, `response_id` for response id, and `survey_id` for response nudge id). The declarative
  engine does not have a coalesce/fallback expression, so schemas model the first key in each
  fallback chain, matching the primary legacy output shape.
- **`pagination.next_page` is not read for legacy streams.** Legacy also looks at
  `pagination.next_page`; the engine's page-number paginator stops on a short page. In the observed
  Qualaroo shape, a short final page and an empty next-page value co-occur, so this can at most cause
  one harmless extra request if Qualaroo ever returns a full final page with no next page.
