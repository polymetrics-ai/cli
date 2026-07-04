# Overview

Oura reads the Oura API v2 usercollection surface from `https://api.ouraring.com/v2`. The bundle covers the member profile endpoint, every documented production usercollection list endpoint, and every documented usercollection document-detail endpoint from the OpenAPI 1.35 spec. Legacy streams keep the narrow field projection emitted by `internal/connectors/oura`; newly added streams project the documented top-level API fields.

## Auth setup

Provide an Oura OAuth2 bearer access token via the `api_key` secret. The field name stays `api_key` for legacy compatibility, but Oura's v2 docs describe OAuth2 bearer access tokens as the supported access method. `base_url` defaults to `https://api.ouraring.com/v2` and may be overridden for tests, proxies, or Oura sandbox reads.

## Streams notes

List streams: daily_sleep, daily_activity, daily_readiness, daily_cardiovascular_age, daily_resilience, daily_spo2, daily_stress, enhanced_tag, rest_mode_period, ring_configuration, session, sleep, sleep_time, tag, vo2_max, workout, heartrate, ring_battery_level. Each list stream reads the response `data` array and uses Oura's `next_token` cursor pagination. Date-based streams accept optional `start_date` and `end_date` config values. Time-series streams `heartrate` and `ring_battery_level` accept optional `start_datetime`, `end_datetime`, and `latest` config values.

Detail streams: daily_sleep_detail, daily_activity_detail, daily_readiness_detail, daily_cardiovascular_age_detail, daily_resilience_detail, daily_spo2_detail, daily_stress_detail, enhanced_tag_detail, rest_mode_period_detail, ring_configuration_detail, session_detail, sleep_detail, sleep_time_detail, tag_detail, vo2_max_detail, workout_detail. Detail streams use the shared optional `document_id` config value in the request path and read the response object as a single record.

`personal_info` is a single-object stream. `daily_sleep`, `daily_activity`, and `daily_readiness` preserve legacy's emitted record shape: `id`, `day`, `score`, and `timestamp`.

## Write actions & risks

None. This bundle is read-only. Oura webhook subscription create/update/renew/delete endpoints are intentionally excluded from writes because they use app-level `x-client-id` plus `x-client-secret` credentials and manage webhook subscriptions rather than usercollection data records.

## Known limits

- `start_date` and `end_date` must be pre-formatted `YYYY-MM-DD` values. Legacy accepted `start_datetime`/`end_datetime` and truncated them for its three daily streams; the declarative bundle uses the documented date query names directly.
- `page_size` and `max_pages` are not runtime-configurable. Oura pagination is driven by `next_token`; the declarative engine stops when Oura omits or clears that token.
- Webhook subscription GET and mutation endpoints are documented in `api_surface.json` but excluded from this usercollection connector because their app-level header auth is distinct from Oura bearer user-data auth.
- Sandbox usercollection endpoints are excluded as duplicates. The same stream definitions can target sandbox URLs through `base_url` when a caller deliberately configures that environment.
