# pm connectors inspect oura

```text
NAME
  pm connectors inspect oura - Oura connector manual

SYNOPSIS
  pm connectors inspect oura
  pm connectors inspect oura --json
  pm credentials add <name> --connector oura [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Oura API v2 usercollection profile, daily summary, time-series, sleep, tag, workout, session, and device-configuration data.

ICON
  asset: icons/oura.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://cloud.ouraring.com/docs/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  document_id
  end_date
  end_datetime
  latest
  start_date
  start_datetime
  api_key (secret)

ETL STREAMS
  personal_info:
    primary key: id
    fields: age(), biological_sex(), email(), height(), id(), weight()
  daily_sleep:
    primary key: id
    cursor: day
    fields: day(), id(), score(), timestamp()
  daily_activity:
    primary key: id
    cursor: day
    fields: day(), id(), score(), timestamp()
  daily_readiness:
    primary key: id
    cursor: day
    fields: day(), id(), score(), timestamp()
  daily_cardiovascular_age:
    primary key: id
    fields: day(), id(), pulse_wave_velocity(), vascular_age()
  daily_resilience:
    primary key: id
    fields: contributors(), day(), id(), level()
  daily_spo2:
    primary key: id
    fields: breathing_disturbance_index(), day(), id(), spo2_percentage()
  daily_stress:
    primary key: id
    fields: day(), day_summary(), id(), recovery_high(), stress_high()
  enhanced_tag:
    primary key: id
    fields: comment(), custom_name(), end_day(), end_time(), id(), start_day(), start_time(), tag_type_code()
  rest_mode_period:
    primary key: id
    fields: end_day(), end_time(), episodes(), id(), start_day(), start_time()
  ring_configuration:
    primary key: id
    fields: color(), design(), firmware_version(), hardware_type(), id(), set_up_at(), size()
  session:
    primary key: id
    fields: day(), end_datetime(), heart_rate(), heart_rate_variability(), id(), mood(), motion_count(), start_datetime(), type()
  sleep:
    primary key: id
    fields: app_sleep_phase_5_min(), average_breath(), average_heart_rate(), average_hrv(), awake_time(), bedtime_end(), bedtime_start(), day(), deep_sleep_duration(), efficiency(), heart_rate(), hrv(), id(), latency(), light_sleep_duration(), low_battery_alert(), lowest_heart_rate(), movement_30_sec(), period(), readiness(), readiness_score_delta(), rem_sleep_duration(), restless_periods(), ring_id(), sleep_algorithm_version(), sleep_analysis_reason(), sleep_phase_30_sec(), sleep_phase_5_min(), sleep_score_delta(), time_in_bed(), total_sleep_duration(), type()
  sleep_time:
    primary key: id
    fields: day(), id(), optimal_bedtime(), recommendation(), status()
  tag:
    primary key: id
    fields: day(), id(), tags(), text(), timestamp()
  vo2_max:
    primary key: id
    fields: day(), id(), timestamp(), vo2_max()
  workout:
    primary key: id
    fields: activity(), calories(), day(), distance(), end_datetime(), id(), intensity(), label(), source(), start_datetime()
  heartrate:
    primary key: timestamp
    fields: bpm(), source(), timestamp(), timestamp_unix()
  ring_battery_level:
    primary key: timestamp
    fields: charging(), in_charger(), level(), timestamp(), timestamp_unix()
  daily_sleep_detail:
    primary key: id
    fields: contributors(), day(), id(), score(), timestamp()
  daily_activity_detail:
    primary key: id
    fields: active_calories(), average_met_minutes(), class_5_min(), contributors(), day(), equivalent_walking_distance(), high_activity_met_minutes(), high_activity_time(), id(), inactivity_alerts(), low_activity_met_minutes(), low_activity_time(), medium_activity_met_minutes(), medium_activity_time(), met(), meters_to_target(), non_wear_time(), resting_time(), score(), sedentary_met_minutes(), sedentary_time(), steps(), target_calories(), target_meters(), timestamp(), total_calories()
  daily_readiness_detail:
    primary key: id
    fields: contributors(), day(), id(), score(), temperature_deviation(), temperature_trend_deviation(), timestamp()
  daily_cardiovascular_age_detail:
    primary key: id
    fields: day(), id(), pulse_wave_velocity(), vascular_age()
  daily_resilience_detail:
    primary key: id
    fields: contributors(), day(), id(), level()
  daily_spo2_detail:
    primary key: id
    fields: breathing_disturbance_index(), day(), id(), spo2_percentage()
  daily_stress_detail:
    primary key: id
    fields: day(), day_summary(), id(), recovery_high(), stress_high()
  enhanced_tag_detail:
    primary key: id
    fields: comment(), custom_name(), end_day(), end_time(), id(), start_day(), start_time(), tag_type_code()
  rest_mode_period_detail:
    primary key: id
    fields: end_day(), end_time(), episodes(), id(), start_day(), start_time()
  ring_configuration_detail:
    primary key: id
    fields: color(), design(), firmware_version(), hardware_type(), id(), set_up_at(), size()
  session_detail:
    primary key: id
    fields: day(), end_datetime(), heart_rate(), heart_rate_variability(), id(), mood(), motion_count(), start_datetime(), type()
  sleep_detail:
    primary key: id
    fields: app_sleep_phase_5_min(), average_breath(), average_heart_rate(), average_hrv(), awake_time(), bedtime_end(), bedtime_start(), day(), deep_sleep_duration(), efficiency(), heart_rate(), hrv(), id(), latency(), light_sleep_duration(), low_battery_alert(), lowest_heart_rate(), movement_30_sec(), period(), readiness(), readiness_score_delta(), rem_sleep_duration(), restless_periods(), ring_id(), sleep_algorithm_version(), sleep_analysis_reason(), sleep_phase_30_sec(), sleep_phase_5_min(), sleep_score_delta(), time_in_bed(), total_sleep_duration(), type()
  sleep_time_detail:
    primary key: id
    fields: day(), id(), optimal_bedtime(), recommendation(), status()
  tag_detail:
    primary key: id
    fields: day(), id(), tags(), text(), timestamp()
  vo2_max_detail:
    primary key: id
    fields: day(), id(), timestamp(), vo2_max()
  workout_detail:
    primary key: id
    fields: activity(), calories(), day(), distance(), end_datetime(), id(), intensity(), label(), source(), start_datetime()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Oura API read of personal wellness and health data, including profile, sleep, activity, readiness, heart-rate, tags, workouts, sessions, and device configuration
  approval: none; this bundle is read-only and excludes app-level webhook subscription mutations
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect oura

  # Inspect as structured JSON
  pm connectors inspect oura --json

AGENT WORKFLOW
  - Run pm connectors inspect oura before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
