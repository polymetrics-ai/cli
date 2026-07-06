# Overview

Reads LinkedIn organization (company page) profile, follower statistics, share statistics, and total
follower count through the LinkedIn Community Management REST API.

Readable streams: `follower_statistics`, `share_statistics`, `organizations`,
`total_follower_count`.

This connector is read-only; no write actions are declared.

Service API documentation:
https://learn.microsoft.com/en-us/linkedin/marketing/community-management/organizations/organization-lookup-api.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); LinkedIn member OAuth2 access token (long-lived). The
  refresh_token exchange, when needed, is performed by the operator/agent layer before this
  connector runs; only the resolved bearer token is used here.
- `base_url` (optional, string); default `https://api.linkedin.com/rest`; format `uri`; LinkedIn
  Community Management API base URL override for tests or proxies.
- `linkedin_version` (optional, string); default `202601`; LinkedIn-Version header value (YYYYMM).
  LinkedIn versions are monthly.
- `mode` (optional, string).
- `org_id` (required, string); LinkedIn organization id (company page). Scopes every request and is
  stamped onto every emitted record. Not a credential (a bare numeric organization identifier);
  declared as ordinary config, not x-secret, so it is available to computed_fields (secrets.* is
  never wired into computed_fields templating, see conventions.md 3) and to the request path/query.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.linkedin.com/rest`, `linkedin_version=202601`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/organizations/{{ config.org_id }}`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `start`; limit parameter `count`; page
size 100.

Pagination by stream: none: `organizations`, `total_follower_count`; offset_limit:
`follower_statistics`, `share_statistics`.

- `follower_statistics`: GET `/organizationalEntityFollowerStatistics` - records path `elements`;
  query `organizationalEntity`=`urn:li:organization:{{ config.org_id }}`;
  `q`=`organizationalEntity`; offset/limit pagination; offset parameter `start`; limit parameter
  `count`; page size 100; computed output fields `org_id`.
- `share_statistics`: GET `/organizationalEntityShareStatistics` - records path `elements`; query
  `organizationalEntity`=`urn:li:organization:{{ config.org_id }}`; `q`=`organizationalEntity`;
  offset/limit pagination; offset parameter `start`; limit parameter `count`; page size 100;
  computed output fields `org_id`.
- `organizations`: GET `/organizations/{{ config.org_id }}` - records path `.`; computed output
  fields `localized_name`, `localized_website`, `org_id`, `organization_type`,
  `primary_organization_type`, `staff_count_range`, `urn`, `vanity_name`, `version_tag`.
- `total_follower_count`: GET `/networkSizes/urn:li:organization:{{ config.org_id }}` - records path
  `.`; query `edgeType`=`COMPANY_FOLLOWED_BY_MEMBER`; computed output fields `first_degree_size`,
  `org_id`.

## Write actions & risks

This connector is read-only. Read behavior: external LinkedIn Community Management API read of
company page profile and lifetime statistics.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5.
