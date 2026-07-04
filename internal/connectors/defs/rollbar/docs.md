# Rollbar connector

## Overview

This bundle covers the documented Rollbar API v1 surface from the current
Rollbar API Reference. It preserves the legacy `items` and `projects` read
shape and adds documented streams and write actions that the declarative engine
can represent without carrying extra secret values in records.

## Auth setup

Authentication uses the Rollbar `X-Rollbar-Access-Token` header, matching the
legacy Go connector and Rollbar's API reference. Store the token in the
`access_token` secret. Optional config identifiers such as `item_id`,
`project_id`, `team_id`, `rule_id`, `service_link_id`, `version`,
`environment`, `session_id`, and `replay_id` are used only by detail streams
and path-scoped writes.

## Streams notes

Documented GET list/detail endpoints are exposed as streams unless the endpoint
returns credential material. The `item_by_uuid` stream uses the `uuid` config
query parameter; `version_items` requires `environment` and `event`.

Report endpoints that return arrays of arrays are emitted as single
response-record streams so the response data is not dropped by object-array
extraction.

## Write actions & risks

Writes cover object-body and no-body Rollbar mutations for item updates,
project/team lifecycle, team/user/project assignments, invitations,
notification settings and rule updates, occurrence/session replay deletes,
service links, and RQL job cancellation.

Destructive actions declare destructive confirmation and idempotent 404 delete
handling. All writes mutate external Rollbar state and require the normal
reverse-ETL preview and approval flow.

## Known limits

The bundle intentionally excludes project access token management because those
endpoints expose or mutate credential material. It also excludes binary upload
endpoints, root-array notification-rule create/replace endpoints,
query-parameter-only person deletion, deploy reporting that requires a token in
the JSON body, and POST-body analytics/RQL query submission. Those exclusions
are reflected in `api_surface.json`; dialect gaps are also recorded in
`docs/migration/quarantine.json`.

Fixtures use synthetic values only. They do not contain Rollbar tokens, project
access tokens, webhook URLs, PagerDuty service keys, or customer payload data.
