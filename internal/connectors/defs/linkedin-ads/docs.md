# Overview

Reads LinkedIn Ads accounts, campaign groups, campaigns, and creatives through the LinkedIn
Marketing REST API.

Readable streams: `accounts`, `campaign_groups`, `campaigns`, `creatives`.

This connector is read-only; no write actions are declared.

Service API documentation:
https://learn.microsoft.com/en-us/linkedin/marketing/integrations/recent-changes?view=li-lms-2024-10.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Sent as a Bearer token; never logged. The refresh_token
  exchange is performed outside this connector.
- `base_url` (optional, string); default `https://api.linkedin.com/rest`; format `uri`; LinkedIn
  Marketing API base URL override for tests or proxies.
- `linkedin_version` (optional, string); default `202601`; LinkedIn-Version header value (monthly,
  YYYYMM). Defaults to the connector's built-in version.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (count parameter), 1-1000.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.linkedin.com/rest`, `linkedin_version=202601`,
`max_pages=0`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/adAccounts`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `start`; limit parameter `count`; page
size 100.

- `accounts`: GET `/adAccounts` - records path `elements`; query `q`=`search`; offset/limit
  pagination; offset parameter `start`; limit parameter `count`; page size 100; computed output
  fields `created_at`, `last_modified`.
- `campaign_groups`: GET `/adCampaignGroups` - records path `elements`; query `q`=`search`;
  offset/limit pagination; offset parameter `start`; limit parameter `count`; page size 100;
  computed output fields `created_at`, `last_modified`, `run_schedule`, `total_budget`.
- `campaigns`: GET `/adCampaigns` - records path `elements`; query `q`=`search`; offset/limit
  pagination; offset parameter `start`; limit parameter `count`; page size 100; computed output
  fields `campaign_group`, `cost_type`, `created_at`, `daily_budget`, `last_modified`,
  `objective_type`, `run_schedule`, `unit_cost`.
- `creatives`: GET `/creatives` - records path `elements`; query `q`=`search`; offset/limit
  pagination; offset parameter `start`; limit parameter `count`; page size 100; computed output
  fields `created_at`, `intended_status`, `is_serving`, `last_modified`, `review_status`.

## Write actions & risks

This connector is read-only. Read behavior: external LinkedIn Marketing API read of ad account,
campaign, and creative data.

## Known limits

- Published rate limit metadata: requests_per_minute=100.
- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=6.
