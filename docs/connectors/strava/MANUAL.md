# pm connectors inspect strava

```text
NAME
  pm connectors inspect strava - Strava connector manual

SYNOPSIS
  pm connectors inspect strava
  pm connectors inspect strava --json
  pm credentials add <name> --connector strava [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads the authenticated Strava athlete's profile, activities, lifetime stats, and clubs through the Strava v3 REST API.

ICON
  id: strava
  asset: icons/strava.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.strava.com/docs/reference/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  athlete_id
  base_url
  client_id
  token_url
  client_secret (secret)
  refresh_token (secret)

ETL STREAMS
  activities:
    primary key: id
    cursor: start_date
    fields: achievement_count(integer), average_speed(number), distance(number), elapsed_time(integer), id(integer), kudos_count(integer), max_speed(number), moving_time(integer), name(string), sport_type(string), start_date(string), start_date_local(string), timezone(string), total_elevation_gain(number), type(string)
  athlete:
    primary key: id
    fields: city(string), country(string), created_at(string), firstname(string), id(integer), lastname(string), sex(string), state(string), updated_at(string), username(string), weight(number)
  athlete_stats:
    primary key: id
    fields: all_ride_totals(object), all_run_totals(object), all_swim_totals(object), biggest_climb_elevation_gain(number), biggest_ride_distance(number), id(integer), recent_ride_totals(object), recent_run_totals(object), recent_swim_totals(object), ytd_ride_totals(object), ytd_run_totals(object), ytd_swim_totals(object)
  clubs:
    primary key: id
    fields: city(string), country(string), id(integer), member_count(integer), membership(string), name(string), private(boolean), sport_type(string), state(string), url(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Strava API read of the authenticated athlete's profile, activity, stats, and club data
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Strava's declared streams and reverse-ETL actions.
  Usage: pm strava <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    activities list - Run the activities ETL stream [intent=etl availability=implemented stream=activities]
    api get activities id - Documented GET /activities/{id} (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.activities-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get activities id comments - Documented GET /activities/{id}/comments (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.activities-id-comments]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get activities id kudos - Documented GET /activities/{id}/kudos (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.activities-id-kudos]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get activities id laps - Documented GET /activities/{id}/laps (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.activities-id-laps]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get activities id streams - Documented GET /activities/{id}/streams (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.activities-id-streams]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get activities id zones - Documented GET /activities/{id}/zones (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.activities-id-zones]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get athlete zones - Documented GET /athlete/zones (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.athlete-zones]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get athletes id routes - Documented GET /athletes/{id}/routes (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.athletes-id-routes]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get clubs id - Documented GET /clubs/{id} (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.clubs-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get clubs id activities - Documented GET /clubs/{id}/activities (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.clubs-id-activities]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get clubs id admins - Documented GET /clubs/{id}/admins (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.clubs-id-admins]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get clubs id members - Documented GET /clubs/{id}/members (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.clubs-id-members]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get gear id - Documented GET /gear/{id} (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.gear-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get routes id - Documented GET /routes/{id} (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.routes-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get routes id export-gpx - Documented GET /routes/{id}/export_gpx (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.routes-id-export-gpx]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get routes id export-tcx - Documented GET /routes/{id}/export_tcx (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.routes-id-export-tcx]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get routes id streams - Documented GET /routes/{id}/streams (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.routes-id-streams]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get segment-efforts - Documented GET /segment_efforts (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.segment-efforts]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get segment-efforts id - Documented GET /segment_efforts/{id} (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.segment-efforts-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get segment-efforts id streams - Documented GET /segment_efforts/{id}/streams (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.segment-efforts-id-streams]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get segments explore - Documented GET /segments/explore (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.segments-explore]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get segments id - Documented GET /segments/{id} (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.segments-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get segments id streams - Documented GET /segments/{id}/streams (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.segments-id-streams]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get segments starred - Documented GET /segments/starred (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.segments-starred]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get uploads uploadid - Documented GET /uploads/{uploadId} (not implemented) [intent=direct_read availability=not_implemented operation=strava.get.uploads-uploadid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post activities - Documented POST /activities (not implemented) [intent=direct_write availability=not_implemented operation=strava.post.activities]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post push-subscriptions - Documented POST /push_subscriptions (not implemented) [intent=direct_write availability=not_implemented operation=strava.post.push-subscriptions]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post uploads - Documented POST /uploads (not implemented) [intent=direct_write availability=not_implemented operation=strava.post.uploads]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put activities id - Documented PUT /activities/{id} (not implemented) [intent=direct_write availability=not_implemented operation=strava.put.activities-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put athlete - Documented PUT /athlete (not implemented) [intent=direct_write availability=not_implemented operation=strava.put.athlete]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put segments id starred - Documented PUT /segments/{id}/starred (not implemented) [intent=direct_write availability=not_implemented operation=strava.put.segments-id-starred]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    athlete list - Run the athlete ETL stream [intent=etl availability=implemented stream=athlete]
    athlete stats list - Run the athlete stats ETL stream [intent=etl availability=implemented stream=athlete_stats]
    clubs list - Run the clubs ETL stream [intent=etl availability=implemented stream=clubs]

EXAMPLES
  # Inspect as a manual
  pm connectors inspect strava

  # Inspect as structured JSON
  pm connectors inspect strava --json

AGENT WORKFLOW
  - Run pm connectors inspect strava before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
