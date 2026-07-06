# Overview

Reads Facebook Marketing ad accounts, campaigns, ads, ad sets, ad creatives, custom audiences, and
performance insights, and creates/updates campaigns and ad sets, through the Graph API.

Readable streams: `ad_accounts`, `campaigns`, `ads`, `ad_sets`, `ad_creatives`, `custom_audiences`,
`insights`.

Write actions: `create_campaign`, `update_campaign`, `create_ad_set`.

Service API documentation: https://developers.facebook.com/docs/marketing-api/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Facebook Graph API access token, sent as a Bearer
  token. Used for every request; never logged.
- `ad_account_id` (optional, string); Facebook ad account id, required for the account-scoped
  'campaigns' and 'ads' streams. Must include the 'act_' prefix (e.g.
- `base_url` (optional, string); default `https://graph.facebook.com/v20.0`; format `uri`; Facebook
  Graph API base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://graph.facebook.com/v20.0`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/me/adaccounts` with query `fields`=`id,account_id,name`; `limit`=`1`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `paging.next`; next
URLs stay on the configured API host.

- `ad_accounts`: GET `/me/adaccounts` - records path `data`; query
  `fields`=`id,account_id,name,account_status,currency,timezone_name`; `limit`=`100`; follows a
  next-page URL from the response body; URL path `paging.next`; next URLs stay on the configured API
  host.
- `campaigns`: GET `/{{ config.ad_account_id }}/campaigns` - records path `data`; query
  `fields`=`id,name,status,effective_status,objective,created_time,updated_time`; `limit`=`100`;
  follows a next-page URL from the response body; URL path `paging.next`; next URLs stay on the
  configured API host.
- `ads`: GET `/{{ config.ad_account_id }}/ads` - records path `data`; query
  `fields`=`id,name,status,effective_status,created_time,updated_time`; `limit`=`100`; follows a
  next-page URL from the response body; URL path `paging.next`; next URLs stay on the configured API
  host.
- `ad_sets`: GET `/{{ config.ad_account_id }}/adsets` - records path `data`; query
  `fields`=`id,name,campaign_id,status,effective_status,daily_budget,lifetime_budget,billing_event,optimization_goal,bid_amount,start_time,end_time,created_time,updated_time`;
  `limit`=`100`; follows a next-page URL from the response body; URL path `paging.next`; next URLs
  stay on the configured API host.
- `ad_creatives`: GET `/{{ config.ad_account_id }}/adcreatives` - records path `data`; query
  `fields`=`id,name,object_story_id,object_type,thumbnail_url,status`; `limit`=`100`; follows a
  next-page URL from the response body; URL path `paging.next`; next URLs stay on the configured API
  host.
- `custom_audiences`: GET `/{{ config.ad_account_id }}/customaudiences` - records path `data`; query
  `fields`=`id,name,subtype,description,approximate_count_lower_bound,approximate_count_upper_bound,operation_status,time_created,time_updated`;
  `limit`=`100`; follows a next-page URL from the response body; URL path `paging.next`; next URLs
  stay on the configured API host.
- `insights`: GET `/{{ config.ad_account_id }}/insights` - records path `data`; query
  `date_preset`=`last_30d`;
  `fields`=`campaign_id,campaign_name,adset_id,adset_name,ad_id,ad_name,impressions,clicks,spend,reach,frequency,cpc,cpm,ctr,date_start,date_stop`;
  `level`=`ad`; `limit`=`100`; follows a next-page URL from the response body; URL path
  `paging.next`; next URLs stay on the configured API host; computed output fields `id`.

## Write actions & risks

Overall write risk: external mutation of a live Facebook ad account; creating/updating campaigns and
ad sets can incur real ad spend once ads are attached and the campaign/ad set is active.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_campaign`: POST `/{{ config.ad_account_id }}/campaigns` - kind `create`; body type `form`;
  required record fields `name`, `objective`, `status`, `special_ad_categories`; accepted fields
  `daily_budget`, `lifetime_budget`, `name`, `objective`, `special_ad_categories`, `status`; risk:
  external mutation on a live Facebook ad account; creates a campaign that can incur ad spend once
  ads are attached; approval required.
- `update_campaign`: POST `/{{ record.id }}` - kind `update`; body type `form`; path fields `id`;
  required record fields `id`; accepted fields `daily_budget`, `id`, `lifetime_budget`, `name`,
  `status`; risk: external mutation on a live Facebook ad account (e.g. pausing/resuming spend);
  approval required.
- `create_ad_set`: POST `/{{ config.ad_account_id }}/adsets` - kind `create`; body type `form`;
  required record fields `name`, `campaign_id`, `billing_event`, `optimization_goal`, `targeting`,
  `status`; accepted fields `bid_amount`, `billing_event`, `campaign_id`, `daily_budget`,
  `lifetime_budget`, `name`, `optimization_goal`, `status`, `targeting`; risk: external mutation on
  a live Facebook ad account; creates an ad set that can incur ad spend once ads are attached;
  approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 7 stream-backed endpoint group(s), 3 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=2, duplicate_of=3, out_of_scope=20, requires_elevated_scope=1.
