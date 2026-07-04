# Overview

SurveyMonkey is a Tier-1 declarative HTTP bundle for the official SurveyMonkey developer reference.
The Pass B surface was expanded from the legacy 3-stream parity bundle to the full rendered reference
served by the official developer portal's embedded reference.

The bundle now covers every documented operation in that reference: 82 GET endpoints as streams and
53 POST/PUT/PATCH endpoints as write actions. The documented surface includes both REST v3 paths
under `/v3` and SCIM v2 paths under `/scim/v2`, so `base_url` now defaults to the host root
`https://api.surveymonkey.com` and each stream/action path carries its API prefix explicitly.

Legacy Go under `internal/connectors/surveymonkey` remains read-only and registered until the wave 6
cutover. The original legacy-parity streams (`surveys`, `collectors`, `survey_responses`) keep
schema projection to preserve their old emitted field set. Newly-added streams use passthrough
projection so live SurveyMonkey fields are preserved rather than being narrowed to a guessed schema.

## Auth setup

Provide a SurveyMonkey OAuth access token via the `access_token` secret. It is sent as bearer auth
for REST v3 and SCIM v2 requests and is never logged or copied into records. `base_url` defaults to
`https://api.surveymonkey.com`; override it only for tests or an API-compatible proxy.

## Streams notes

REST v3 list streams use SurveyMonkey's standard `data` response envelope and `links.next`
pagination, with `per_page=100` on the first request. REST v3 detail, statistics, rollup, trend,
benchmark analysis, profile, organization, and error-detail streams read a single root object.

SCIM list streams read the `Resources` array with `count=1000`; SCIM detail streams read a single
root object. SCIM endpoints are included because they are documented in the same SurveyMonkey API
reference and represent real list/detail resources.

Many detail streams require the corresponding optional config identifier (`survey_id`,
`collector_id`, `workgroup_id`, `scim_user_id`, etc.). Missing identifiers fail during path
interpolation before any request is sent.

## Write actions & risks

Writes cover every documented POST/PUT/PATCH operation from the fetched reference. Actions are
specific to their SurveyMonkey endpoint, for example `create_survey`, `update_collector_message`,
`send_collector_message`, `create_workgroup_share`, and `update_scim_user`; there is no generic HTTP
write action.

Every write executes a live SurveyMonkey mutation only after reverse-ETL preview and approval.
Collector message sends, survey sharing, workgroup sharing, organization changes, membership changes,
and SCIM provisioning are high-impact actions because they can notify real recipients or change
account/team access.

## Known limits

- `page_size` and `max_pages` remain fixed in the declarative bundle. Legacy exposed config-driven
  overrides, but the `next_url` paginator has no runtime-configurable page-size or max-page field.
- SCIM endpoints document `application/scim+json`; the engine only supports base-wide headers, so
  this bundle uses the same JSON request machinery for REST and SCIM instead of per-stream media
  types.
- Newly-added stream schemas are minimal catalog schemas with passthrough projection. They provide a
  stable primary-key field and common top-level fields, but they do not attempt to exhaustively
  mirror every SurveyMonkey response schema in the rendered docs.
- Fixture replay is retained for the original legacy-parity streams only. Additional streams and
  writes are statically validated and surface-covered; no synthetic replay pages are fabricated for
  endpoints that were not present in the legacy connector.
