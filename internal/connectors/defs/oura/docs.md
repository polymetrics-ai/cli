# Overview

Reads Oura API v2 usercollection profile, daily summary, time-series, sleep, tag, workout, session,
and device-configuration data.

Readable streams: `personal_info`, `daily_sleep`, `daily_activity`, `daily_readiness`,
`daily_cardiovascular_age`, `daily_resilience`, `daily_spo2`, `daily_stress`, `enhanced_tag`,
`rest_mode_period`, `ring_configuration`, `session`, `sleep`, `sleep_time`, `tag`, `vo2_max`,
`workout`, `heartrate`, `ring_battery_level`, `daily_sleep_detail`, `daily_activity_detail`,
`daily_readiness_detail`, `daily_cardiovascular_age_detail`, `daily_resilience_detail`,
`daily_spo2_detail`, `daily_stress_detail`, `enhanced_tag_detail`, `rest_mode_period_detail`,
`ring_configuration_detail`, `session_detail`, `sleep_detail`, `sleep_time_detail`, `tag_detail`,
`vo2_max_detail`, `workout_detail`.

This connector is read-only; no write actions are declared.

Service API documentation: https://cloud.ouraring.com/v2/docs.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Sent as Authorization: Bearer <api_key>.
- `base_url` (optional, string); default `https://api.ouraring.com/v2`; format `uri`; Oura API v2
  root URL override for tests or proxies.
- `document_id` (optional, string); Document id used only by *_detail streams.
- `end_date` (optional, string); Optional YYYY-MM-DD upper-bound query value for date-based
  usercollection list streams.
- `end_datetime` (optional, string); Optional RFC3339 upper-bound query value for time-series
  usercollection streams.
- `latest` (optional, string); Optional Oura latest query flag for time-series streams; use the API
  documented value.
- `start_date` (optional, string); Optional YYYY-MM-DD lower-bound query value for date-based
  usercollection list streams.
- `start_datetime` (optional, string); Optional RFC3339 lower-bound query value for time-series
  usercollection streams.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.ouraring.com/v2`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/usercollection/personal_info`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `daily_sleep`, `daily_activity`, `daily_readiness`,
`daily_cardiovascular_age`, `daily_resilience`, `daily_spo2`, `daily_stress`, `enhanced_tag`,
`rest_mode_period`, `ring_configuration`, `session`, `sleep`, `sleep_time`, `tag`, `vo2_max`,
`workout`, `heartrate`, `ring_battery_level`; none: `personal_info`, `daily_sleep_detail`,
`daily_activity_detail`, `daily_readiness_detail`, `daily_cardiovascular_age_detail`,
`daily_resilience_detail`, `daily_spo2_detail`, `daily_stress_detail`, `enhanced_tag_detail`,
`rest_mode_period_detail`, `ring_configuration_detail`, `session_detail`, `sleep_detail`,
`sleep_time_detail`, `tag_detail`, `vo2_max_detail`, `workout_detail`.

- `personal_info`: GET `/usercollection/personal_info` - records path `.`.
- `daily_sleep`: GET `/usercollection/daily_sleep` - records path `data`; query `end_date` from
  template `{{ config.end_date }}`, omitted when absent; `start_date` from template `{{
  config.start_date }}`, omitted when absent; cursor pagination; cursor parameter `next_token`; next
  token from `next_token`.
- `daily_activity`: GET `/usercollection/daily_activity` - records path `data`; query `end_date`
  from template `{{ config.end_date }}`, omitted when absent; `start_date` from template `{{
  config.start_date }}`, omitted when absent; cursor pagination; cursor parameter `next_token`; next
  token from `next_token`.
- `daily_readiness`: GET `/usercollection/daily_readiness` - records path `data`; query `end_date`
  from template `{{ config.end_date }}`, omitted when absent; `start_date` from template `{{
  config.start_date }}`, omitted when absent; cursor pagination; cursor parameter `next_token`; next
  token from `next_token`.
- `daily_cardiovascular_age`: GET `/usercollection/daily_cardiovascular_age` - records path `data`;
  query `end_date` from template `{{ config.end_date }}`, omitted when absent; `start_date` from
  template `{{ config.start_date }}`, omitted when absent; cursor pagination; cursor parameter
  `next_token`; next token from `next_token`.
- `daily_resilience`: GET `/usercollection/daily_resilience` - records path `data`; query `end_date`
  from template `{{ config.end_date }}`, omitted when absent; `start_date` from template `{{
  config.start_date }}`, omitted when absent; cursor pagination; cursor parameter `next_token`; next
  token from `next_token`.
- `daily_spo2`: GET `/usercollection/daily_spo2` - records path `data`; query `end_date` from
  template `{{ config.end_date }}`, omitted when absent; `start_date` from template `{{
  config.start_date }}`, omitted when absent; cursor pagination; cursor parameter `next_token`; next
  token from `next_token`.
- `daily_stress`: GET `/usercollection/daily_stress` - records path `data`; query `end_date` from
  template `{{ config.end_date }}`, omitted when absent; `start_date` from template `{{
  config.start_date }}`, omitted when absent; cursor pagination; cursor parameter `next_token`; next
  token from `next_token`.
- `enhanced_tag`: GET `/usercollection/enhanced_tag` - records path `data`; query `end_date` from
  template `{{ config.end_date }}`, omitted when absent; `start_date` from template `{{
  config.start_date }}`, omitted when absent; cursor pagination; cursor parameter `next_token`; next
  token from `next_token`.
- `rest_mode_period`: GET `/usercollection/rest_mode_period` - records path `data`; query `end_date`
  from template `{{ config.end_date }}`, omitted when absent; `start_date` from template `{{
  config.start_date }}`, omitted when absent; cursor pagination; cursor parameter `next_token`; next
  token from `next_token`.
- `ring_configuration`: GET `/usercollection/ring_configuration` - records path `data`; cursor
  pagination; cursor parameter `next_token`; next token from `next_token`.
- `session`: GET `/usercollection/session` - records path `data`; query `end_date` from template `{{
  config.end_date }}`, omitted when absent; `start_date` from template `{{ config.start_date }}`,
  omitted when absent; cursor pagination; cursor parameter `next_token`; next token from
  `next_token`.
- `sleep`: GET `/usercollection/sleep` - records path `data`; query `end_date` from template `{{
  config.end_date }}`, omitted when absent; `start_date` from template `{{ config.start_date }}`,
  omitted when absent; cursor pagination; cursor parameter `next_token`; next token from
  `next_token`.
- `sleep_time`: GET `/usercollection/sleep_time` - records path `data`; query `end_date` from
  template `{{ config.end_date }}`, omitted when absent; `start_date` from template `{{
  config.start_date }}`, omitted when absent; cursor pagination; cursor parameter `next_token`; next
  token from `next_token`.
- `tag`: GET `/usercollection/tag` - records path `data`; query `end_date` from template `{{
  config.end_date }}`, omitted when absent; `start_date` from template `{{ config.start_date }}`,
  omitted when absent; cursor pagination; cursor parameter `next_token`; next token from
  `next_token`.
- `vo2_max`: GET `/usercollection/vO2_max` - records path `data`; query `end_date` from template `{{
  config.end_date }}`, omitted when absent; `start_date` from template `{{ config.start_date }}`,
  omitted when absent; cursor pagination; cursor parameter `next_token`; next token from
  `next_token`.
- `workout`: GET `/usercollection/workout` - records path `data`; query `end_date` from template `{{
  config.end_date }}`, omitted when absent; `start_date` from template `{{ config.start_date }}`,
  omitted when absent; cursor pagination; cursor parameter `next_token`; next token from
  `next_token`.
- `heartrate`: GET `/usercollection/heartrate` - records path `data`; query `end_datetime` from
  template `{{ config.end_datetime }}`, omitted when absent; `latest` from template `{{
  config.latest }}`, omitted when absent; `start_datetime` from template `{{ config.start_datetime
  }}`, omitted when absent; cursor pagination; cursor parameter `next_token`; next token from
  `next_token`.
- `ring_battery_level`: GET `/usercollection/ring_battery_level` - records path `data`; query
  `end_datetime` from template `{{ config.end_datetime }}`, omitted when absent; `latest` from
  template `{{ config.latest }}`, omitted when absent; `start_datetime` from template `{{
  config.start_datetime }}`, omitted when absent; cursor pagination; cursor parameter `next_token`;
  next token from `next_token`.
- `daily_sleep_detail`: GET `/usercollection/daily_sleep/{{ config.document_id }}` - records path
  `.`.
- `daily_activity_detail`: GET `/usercollection/daily_activity/{{ config.document_id }}` - records
  path `.`.
- `daily_readiness_detail`: GET `/usercollection/daily_readiness/{{ config.document_id }}` - records
  path `.`.
- `daily_cardiovascular_age_detail`: GET `/usercollection/daily_cardiovascular_age/{{
  config.document_id }}` - records path `.`.
- `daily_resilience_detail`: GET `/usercollection/daily_resilience/{{ config.document_id }}` -
  records path `.`.
- `daily_spo2_detail`: GET `/usercollection/daily_spo2/{{ config.document_id }}` - records path `.`.
- `daily_stress_detail`: GET `/usercollection/daily_stress/{{ config.document_id }}` - records path
  `.`.
- `enhanced_tag_detail`: GET `/usercollection/enhanced_tag/{{ config.document_id }}` - records path
  `.`.
- `rest_mode_period_detail`: GET `/usercollection/rest_mode_period/{{ config.document_id }}` -
  records path `.`.
- `ring_configuration_detail`: GET `/usercollection/ring_configuration/{{ config.document_id }}` -
  records path `.`.
- `session_detail`: GET `/usercollection/session/{{ config.document_id }}` - records path `.`.
- `sleep_detail`: GET `/usercollection/sleep/{{ config.document_id }}` - records path `.`.
- `sleep_time_detail`: GET `/usercollection/sleep_time/{{ config.document_id }}` - records path `.`.
- `tag_detail`: GET `/usercollection/tag/{{ config.document_id }}` - records path `.`.
- `vo2_max_detail`: GET `/usercollection/vO2_max/{{ config.document_id }}` - records path `.`.
- `workout_detail`: GET `/usercollection/workout/{{ config.document_id }}` - records path `.`.

## Write actions & risks

This connector is read-only. Read behavior: external Oura API read of personal wellness and health
data, including profile, sleep, activity, readiness, heart-rate, tags, workouts, sessions, and
device configuration.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 35 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=34, non_data_endpoint=2, requires_elevated_scope=6.
