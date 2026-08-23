---
name: pm-oura
description: Oura connector knowledge and safe action guide.
---

# pm-oura

## Purpose

Reads Oura API v2 usercollection profile, daily summary, time-series, sleep, tag, workout, session, and device-configuration data.

## Icon

- id: oura
- asset: icons/oura.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://cloud.ouraring.com/docs/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- document_id
- end_date
- end_datetime
- latest
- start_date
- start_datetime
- api_key (secret) (required)

## ETL Streams

- personal_info:
  - primary key: id
  - fields: age(integer), biological_sex(string), email(string), height(number), id(string), weight(number)
- daily_sleep:
  - primary key: id
  - cursor: day
  - fields: day(string), id(string), score(integer), timestamp(string)
- daily_activity:
  - primary key: id
  - cursor: day
  - fields: day(string), id(string), score(integer), timestamp(string)
- daily_readiness:
  - primary key: id
  - cursor: day
  - fields: day(string), id(string), score(integer), timestamp(string)
- daily_cardiovascular_age:
  - primary key: id
  - fields: day(string), id(string), pulse_wave_velocity(number), vascular_age(integer)
- daily_resilience:
  - primary key: id
  - fields: contributors(object), day(string), id(string), level(string)
- daily_spo2:
  - primary key: id
  - fields: breathing_disturbance_index(integer), day(string), id(string), spo2_percentage(object)
- daily_stress:
  - primary key: id
  - fields: day(string), day_summary(string), id(string), recovery_high(integer), stress_high(integer)
- enhanced_tag:
  - primary key: id
  - fields: comment(string), custom_name(string), end_day(string), end_time(string), id(string), start_day(string), start_time(string), tag_type_code(string)
- rest_mode_period:
  - primary key: id
  - fields: end_day(string), end_time(string), episodes(array), id(string), start_day(string), start_time(string)
- ring_configuration:
  - primary key: id
  - fields: color(string), design(string), firmware_version(string), hardware_type(string), id(string), set_up_at(string), size(integer)
- session:
  - primary key: id
  - fields: day(string), end_datetime(string), heart_rate(object), heart_rate_variability(object), id(string), mood(string), motion_count(object), start_datetime(string), type(string)
- sleep:
  - primary key: id
  - fields: app_sleep_phase_5_min(string), average_breath(number), average_heart_rate(number), average_hrv(integer), awake_time(integer), bedtime_end(string), bedtime_start(string), day(string), deep_sleep_duration(integer), efficiency(integer), heart_rate(object), hrv(object), id(string), latency(integer), light_sleep_duration(integer), low_battery_alert(boolean), lowest_heart_rate(integer), movement_30_sec(string), period(integer), readiness(object), readiness_score_delta(integer), rem_sleep_duration(integer), restless_periods(integer), ring_id(string), sleep_algorithm_version(string), sleep_analysis_reason(string), sleep_phase_30_sec(string), sleep_phase_5_min(string), sleep_score_delta(integer), time_in_bed(integer), total_sleep_duration(integer), type(string)
- sleep_time:
  - primary key: id
  - fields: day(string), id(string), optimal_bedtime(object), recommendation(string), status(string)
- tag:
  - primary key: id
  - fields: day(string), id(string), tags(array), text(string), timestamp(string)
- vo2_max:
  - primary key: id
  - fields: day(string), id(string), timestamp(string), vo2_max(integer)
- workout:
  - primary key: id
  - fields: activity(string), calories(number), day(string), distance(number), end_datetime(string), id(string), intensity(string), label(string), source(string), start_datetime(string)
- heartrate:
  - primary key: timestamp
  - fields: bpm(integer), source(string), timestamp(string), timestamp_unix(integer)
- ring_battery_level:
  - primary key: timestamp
  - fields: charging(boolean), in_charger(boolean), level(integer), timestamp(string), timestamp_unix(integer)
- daily_sleep_detail:
  - primary key: id
  - fields: contributors(object), day(string), id(string), score(integer), timestamp(string)
- daily_activity_detail:
  - primary key: id
  - fields: active_calories(integer), average_met_minutes(number), class_5_min(string), contributors(object), day(string), equivalent_walking_distance(integer), high_activity_met_minutes(integer), high_activity_time(integer), id(string), inactivity_alerts(integer), low_activity_met_minutes(integer), low_activity_time(integer), medium_activity_met_minutes(integer), medium_activity_time(integer), met(object), meters_to_target(integer), non_wear_time(integer), resting_time(integer), score(integer), sedentary_met_minutes(integer), sedentary_time(integer), steps(integer), target_calories(integer), target_meters(integer), timestamp(string), total_calories(integer)
- daily_readiness_detail:
  - primary key: id
  - fields: contributors(object), day(string), id(string), score(integer), temperature_deviation(number), temperature_trend_deviation(number), timestamp(string)
- daily_cardiovascular_age_detail:
  - primary key: id
  - fields: day(string), id(string), pulse_wave_velocity(number), vascular_age(integer)
- daily_resilience_detail:
  - primary key: id
  - fields: contributors(object), day(string), id(string), level(string)
- daily_spo2_detail:
  - primary key: id
  - fields: breathing_disturbance_index(integer), day(string), id(string), spo2_percentage(object)
- daily_stress_detail:
  - primary key: id
  - fields: day(string), day_summary(string), id(string), recovery_high(integer), stress_high(integer)
- enhanced_tag_detail:
  - primary key: id
  - fields: comment(string), custom_name(string), end_day(string), end_time(string), id(string), start_day(string), start_time(string), tag_type_code(string)
- rest_mode_period_detail:
  - primary key: id
  - fields: end_day(string), end_time(string), episodes(array), id(string), start_day(string), start_time(string)
- ring_configuration_detail:
  - primary key: id
  - fields: color(string), design(string), firmware_version(string), hardware_type(string), id(string), set_up_at(string), size(integer)
- session_detail:
  - primary key: id
  - fields: day(string), end_datetime(string), heart_rate(object), heart_rate_variability(object), id(string), mood(string), motion_count(object), start_datetime(string), type(string)
- sleep_detail:
  - primary key: id
  - fields: app_sleep_phase_5_min(string), average_breath(number), average_heart_rate(number), average_hrv(integer), awake_time(integer), bedtime_end(string), bedtime_start(string), day(string), deep_sleep_duration(integer), efficiency(integer), heart_rate(object), hrv(object), id(string), latency(integer), light_sleep_duration(integer), low_battery_alert(boolean), lowest_heart_rate(integer), movement_30_sec(string), period(integer), readiness(object), readiness_score_delta(integer), rem_sleep_duration(integer), restless_periods(integer), ring_id(string), sleep_algorithm_version(string), sleep_analysis_reason(string), sleep_phase_30_sec(string), sleep_phase_5_min(string), sleep_score_delta(integer), time_in_bed(integer), total_sleep_duration(integer), type(string)
- sleep_time_detail:
  - primary key: id
  - fields: day(string), id(string), optimal_bedtime(object), recommendation(string), status(string)
- tag_detail:
  - primary key: id
  - fields: day(string), id(string), tags(array), text(string), timestamp(string)
- vo2_max_detail:
  - primary key: id
  - fields: day(string), id(string), timestamp(string), vo2_max(integer)
- workout_detail:
  - primary key: id
  - fields: activity(string), calories(number), day(string), distance(number), end_datetime(string), id(string), intensity(string), label(string), source(string), start_datetime(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Oura API read of personal wellness and health data, including profile, sleep, activity, readiness, heart-rate, tags, workouts, sessions, and device configuration
- approval: none; this bundle is read-only and excludes app-level webhook subscription mutations
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect oura
```

### Inspect as structured JSON

```bash
pm connectors inspect oura --json
```

## Agent Rules

- Run pm connectors inspect oura before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
