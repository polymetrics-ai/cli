# Overview

Reads Piwik/Matomo sites, recent visits, configured goals, and documented analytics reports through
the Reporting API.

Readable streams: `sites`, `visits`, `actions`, `goals`, `report_metadata`, `live_counters`, `site`,
`sites_manager_sites_with_at_least_view_access`, `sites_manager_sites_with_view_access`,
`sites_manager_sites_with_admin_access`, `actions_summary`, `actions_downloads`,
`actions_entry_page_titles`, `actions_entry_page_urls`, `actions_exit_page_titles`,
`actions_exit_page_urls`, `actions_outlinks`, `actions_page_titles`,
`actions_page_titles_following_site_search`, `actions_page_urls_following_site_search`,
`actions_site_search_categories`, `actions_site_search_keywords`,
`actions_site_search_no_result_keywords`, `contents_content_names`, `contents_content_pieces`,
`custom_dimensions_custom_dimension`, `custom_variables_custom_variables`, `device_plugins_plugin`,
`devices_detection_brand`, `devices_detection_browser_engines`,
`devices_detection_browser_versions`, `devices_detection_browsers`, `devices_detection_model`,
`devices_detection_os_families`, `devices_detection_os_versions`, `devices_detection_type`,
`events_action`, `events_category`, `events_name`, `goals_summary`, `goals_days_to_conversion`,
`goals_items_category`, `goals_items_name`, `goals_items_sku`, `goals_visits_until_conversion`,
`marketing_campaigns_reporting_content`, `marketing_campaigns_reporting_group`,
`marketing_campaigns_reporting_id`, `marketing_campaigns_reporting_keyword`,
`marketing_campaigns_reporting_medium`, `marketing_campaigns_reporting_name`,
`marketing_campaigns_reporting_placement`, `marketing_campaigns_reporting_source`,
`marketing_campaigns_reporting_source_medium`,
`multi_channel_conversion_attribution_channel_attribution`, `multi_sites_all`, `multi_sites_one`,
`page_performance`, `referrers_summary`, `referrers_ai_assistants`, `referrers_all`,
`referrers_keywords`, `referrers_referrer_type`, `referrers_search_engines`, `referrers_socials`,
`referrers_websites`, `resolution_configuration`, `resolution_resolution`,
`search_engine_keywords_performance_crawling_overview_bing`,
`search_engine_keywords_performance_keywords`, `search_engine_keywords_performance_keywords_bing`,
`search_engine_keywords_performance_keywords_google_image`,
`search_engine_keywords_performance_keywords_google_video`,
`search_engine_keywords_performance_keywords_google_web`,
`search_engine_keywords_performance_keywords_imported`, `user_country_city`,
`user_country_continent`, `user_country_country`, `user_country_region`, `user_id_users`,
`user_language_language`, `user_language_language_code`, `users_flow_users_flow_pretty`,
`visit_frequency`, `visit_time_by_day_of_week`, `visit_time_visit_information_per_local_time`,
`visit_time_visit_information_per_server_time`,
`visitor_interest_number_of_visits_by_days_since_last`,
`visitor_interest_number_of_visits_by_visit_count`, `visitor_interest_number_of_visits_per_page`,
`visitor_interest_number_of_visits_per_visit_duration`, `visits_summary`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.matomo.org/api-reference/reporting-api.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://matomo.example.com`; format `uri`; Piwik/Matomo
  instance base URL override for tests or proxies.
- `custom_dimension_id` (optional, string); Matomo idDimension for the
  custom_dimensions_custom_dimension stream.
- `date` (optional, string); default `today`; Matomo reporting date or date range for analytics
  report streams.
- `last_minutes` (optional, string); default `30`; Lookback window, in minutes, for the
  live_counters stream.
- `mode` (optional, string).
- `period` (optional, string); default `day`; Matomo reporting period (day, week, month, year,
  range) for analytics report streams.
- `site_id` (optional, string); Matomo idSite; required for site-scoped analytics and detail
  streams.
- `token_auth` (required, secret, string); Piwik/Matomo API auth token, sent as the token_auth query
  parameter. Never logged.

Secret fields are redacted in logs and write previews: `token_auth`.

Default configuration values: `base_url=https://matomo.example.com`, `date=today`,
`last_minutes=30`, `period=day`.

Authentication behavior:

- API key authentication in query parameter `token_auth` using `secrets.token_auth`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/index.php` with query `format`=`JSON`;
`method`=`SitesManager.getAllSites`; `module`=`API`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `sites`, `goals`, `report_metadata`, `live_counters`, `site`,
`sites_manager_sites_with_at_least_view_access`, `sites_manager_sites_with_view_access`,
`sites_manager_sites_with_admin_access`; offset_limit: `visits`, `actions`, `actions_summary`,
`actions_downloads`, `actions_entry_page_titles`, `actions_entry_page_urls`,
`actions_exit_page_titles`, `actions_exit_page_urls`, `actions_outlinks`, `actions_page_titles`,
`actions_page_titles_following_site_search`, `actions_page_urls_following_site_search`,
`actions_site_search_categories`, `actions_site_search_keywords`,
`actions_site_search_no_result_keywords`, `contents_content_names`, `contents_content_pieces`,
`custom_dimensions_custom_dimension`, `custom_variables_custom_variables`, `device_plugins_plugin`,
`devices_detection_brand`, `devices_detection_browser_engines`,
`devices_detection_browser_versions`, `devices_detection_browsers`, `devices_detection_model`,
`devices_detection_os_families`, `devices_detection_os_versions`, `devices_detection_type`,
`events_action`, `events_category`, `events_name`, `goals_summary`, `goals_days_to_conversion`,
`goals_items_category`, `goals_items_name`, `goals_items_sku`, `goals_visits_until_conversion`,
`marketing_campaigns_reporting_content`, `marketing_campaigns_reporting_group`,
`marketing_campaigns_reporting_id`, `marketing_campaigns_reporting_keyword`,
`marketing_campaigns_reporting_medium`, `marketing_campaigns_reporting_name`,
`marketing_campaigns_reporting_placement`, `marketing_campaigns_reporting_source`,
`marketing_campaigns_reporting_source_medium`,
`multi_channel_conversion_attribution_channel_attribution`, `multi_sites_all`, `multi_sites_one`,
`page_performance`, `referrers_summary`, `referrers_ai_assistants`, `referrers_all`,
`referrers_keywords`, `referrers_referrer_type`, `referrers_search_engines`, `referrers_socials`,
`referrers_websites`, `resolution_configuration`, `resolution_resolution`, and 24 more.

- `sites`: GET `/index.php` - records at response root; query `filter_limit`=`100`;
  `filter_offset`=`0`; `format`=`JSON`; `method`=`SitesManager.getAllSites`; `module`=`API`;
  computed output fields `main_url`, `site_id`.
- `visits`: GET `/index.php` - records at response root; query `date`=`{{ config.date }}`;
  `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`Live.getLastVisitsDetails`;
  `module`=`API`; `period`=`{{ config.period }}`; offset/limit pagination; offset parameter
  `filter_offset`; limit parameter `filter_limit`; page size 100; maximum 3 page(s); computed output
  fields `last_action_at`, `visit_id`, `visitor_id`.
- `actions`: GET `/index.php` - records at response root; query `date`=`{{ config.date }}`;
  `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`Actions.getPageUrls`; `module`=`API`;
  `period`=`{{ config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit
  parameter `filter_limit`; page size 100; maximum 3 page(s); computed output fields `hits`,
  `visits`.
- `goals`: GET `/index.php` - records at response root; query `date`=`{{ config.date }}`;
  `filter_limit`=`100`; `filter_offset`=`0`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`Goals.getGoals`; `module`=`API`; `period`=`{{ config.period }}`; computed output fields
  `goal_id`.
- `report_metadata`: GET `/index.php` - records at response root; query `date`=`{{ config.date }}`;
  `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`API.getReportMetadata`;
  `module`=`API`; `period`=`{{ config.period }}`; computed output fields `record_id`, `unique_id`.
- `live_counters`: GET `/index.php` - records at response root; query `format`=`JSON`; `idSite`=`{{
  config.site_id }}`; `lastMinutes`=`{{ config.last_minutes }}`; `method`=`Live.getCounters`;
  `module`=`API`; computed output fields `report`.
- `site`: GET `/index.php` - records at response root; query `format`=`JSON`; `idSite`=`{{
  config.site_id }}`; `method`=`SitesManager.getSiteFromId`; `module`=`API`; computed output fields
  `main_url`, `site_id`.
- `sites_manager_sites_with_at_least_view_access`: GET `/index.php` - records at response root;
  query `filter_limit`=`100`; `filter_offset`=`0`; `format`=`JSON`;
  `method`=`SitesManager.getSitesWithAtLeastViewAccess`; `module`=`API`; computed output fields
  `main_url`, `site_id`.
- `sites_manager_sites_with_view_access`: GET `/index.php` - records at response root; query
  `filter_limit`=`100`; `filter_offset`=`0`; `format`=`JSON`;
  `method`=`SitesManager.getSitesWithViewAccess`; `module`=`API`; computed output fields `main_url`,
  `site_id`.
- `sites_manager_sites_with_admin_access`: GET `/index.php` - records at response root; query
  `filter_limit`=`100`; `filter_offset`=`0`; `format`=`JSON`;
  `method`=`SitesManager.getSitesWithAdminAccess`; `module`=`API`; computed output fields
  `main_url`, `site_id`.
- `actions_summary`: GET `/index.php` - records at response root; query `date`=`{{ config.date }}`;
  `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`Actions.get`; `module`=`API`;
  `period`=`{{ config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit
  parameter `filter_limit`; page size 100; computed output fields `report`.
- `actions_downloads`: GET `/index.php` - records at response root; query `date`=`{{ config.date
  }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`Actions.getDownloads`;
  `module`=`API`; `period`=`{{ config.period }}`; offset/limit pagination; offset parameter
  `filter_offset`; limit parameter `filter_limit`; page size 100; computed output fields
  `record_id`, `report`.
- `actions_entry_page_titles`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`Actions.getEntryPageTitles`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `actions_entry_page_urls`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`Actions.getEntryPageUrls`; `module`=`API`; `period`=`{{ config.period }}`; offset/limit
  pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page size 100;
  computed output fields `record_id`, `report`.
- `actions_exit_page_titles`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`Actions.getExitPageTitles`; `module`=`API`; `period`=`{{ config.period }}`; offset/limit
  pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page size 100;
  computed output fields `record_id`, `report`.
- `actions_exit_page_urls`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`Actions.getExitPageUrls`; `module`=`API`; `period`=`{{ config.period }}`; offset/limit
  pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page size 100;
  computed output fields `record_id`, `report`.
- `actions_outlinks`: GET `/index.php` - records at response root; query `date`=`{{ config.date }}`;
  `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`Actions.getOutlinks`; `module`=`API`;
  `period`=`{{ config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit
  parameter `filter_limit`; page size 100; computed output fields `record_id`, `report`.
- `actions_page_titles`: GET `/index.php` - records at response root; query `date`=`{{ config.date
  }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`Actions.getPageTitles`;
  `module`=`API`; `period`=`{{ config.period }}`; offset/limit pagination; offset parameter
  `filter_offset`; limit parameter `filter_limit`; page size 100; computed output fields
  `record_id`, `report`.
- `actions_page_titles_following_site_search`: GET `/index.php` - records at response root; query
  `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`Actions.getPageTitlesFollowingSiteSearch`; `module`=`API`; `period`=`{{ config.period
  }}`; offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`;
  page size 100; computed output fields `record_id`, `report`.
- `actions_page_urls_following_site_search`: GET `/index.php` - records at response root; query
  `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`Actions.getPageUrlsFollowingSiteSearch`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `actions_site_search_categories`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`Actions.getSiteSearchCategories`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `actions_site_search_keywords`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`Actions.getSiteSearchKeywords`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `actions_site_search_no_result_keywords`: GET `/index.php` - records at response root; query
  `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`Actions.getSiteSearchNoResultKeywords`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `contents_content_names`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`Contents.getContentNames`; `module`=`API`; `period`=`{{ config.period }}`; offset/limit
  pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page size 100;
  computed output fields `record_id`, `report`.
- `contents_content_pieces`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`Contents.getContentPieces`; `module`=`API`; `period`=`{{ config.period }}`; offset/limit
  pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page size 100;
  computed output fields `record_id`, `report`.
- `custom_dimensions_custom_dimension`: GET `/index.php` - records at response root; query
  `date`=`{{ config.date }}`; `format`=`JSON`; `idDimension`=`{{ config.custom_dimension_id }}`;
  `idSite`=`{{ config.site_id }}`; `method`=`CustomDimensions.getCustomDimension`; `module`=`API`;
  `period`=`{{ config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit
  parameter `filter_limit`; page size 100; computed output fields `record_id`, `report`.
- `custom_variables_custom_variables`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`CustomVariables.getCustomVariables`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `device_plugins_plugin`: GET `/index.php` - records at response root; query `date`=`{{ config.date
  }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`DevicePlugins.getPlugin`;
  `module`=`API`; `period`=`{{ config.period }}`; offset/limit pagination; offset parameter
  `filter_offset`; limit parameter `filter_limit`; page size 100; computed output fields
  `record_id`, `report`.
- `devices_detection_brand`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`DevicesDetection.getBrand`; `module`=`API`; `period`=`{{ config.period }}`; offset/limit
  pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page size 100;
  computed output fields `record_id`, `report`.
- `devices_detection_browser_engines`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`DevicesDetection.getBrowserEngines`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `devices_detection_browser_versions`: GET `/index.php` - records at response root; query
  `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`DevicesDetection.getBrowserVersions`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `devices_detection_browsers`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`DevicesDetection.getBrowsers`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `devices_detection_model`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`DevicesDetection.getModel`; `module`=`API`; `period`=`{{ config.period }}`; offset/limit
  pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page size 100;
  computed output fields `record_id`, `report`.
- `devices_detection_os_families`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`DevicesDetection.getOsFamilies`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `devices_detection_os_versions`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`DevicesDetection.getOsVersions`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `devices_detection_type`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`DevicesDetection.getType`; `module`=`API`; `period`=`{{ config.period }}`; offset/limit
  pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page size 100;
  computed output fields `record_id`, `report`.
- `events_action`: GET `/index.php` - records at response root; query `date`=`{{ config.date }}`;
  `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`Events.getAction`; `module`=`API`;
  `period`=`{{ config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit
  parameter `filter_limit`; page size 100; computed output fields `record_id`, `report`.
- `events_category`: GET `/index.php` - records at response root; query `date`=`{{ config.date }}`;
  `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`Events.getCategory`; `module`=`API`;
  `period`=`{{ config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit
  parameter `filter_limit`; page size 100; computed output fields `record_id`, `report`.
- `events_name`: GET `/index.php` - records at response root; query `date`=`{{ config.date }}`;
  `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`Events.getName`; `module`=`API`;
  `period`=`{{ config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit
  parameter `filter_limit`; page size 100; computed output fields `record_id`, `report`.
- `goals_summary`: GET `/index.php` - records at response root; query `date`=`{{ config.date }}`;
  `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`Goals.get`; `module`=`API`;
  `period`=`{{ config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit
  parameter `filter_limit`; page size 100; computed output fields `report`.
- `goals_days_to_conversion`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`Goals.getDaysToConversion`; `module`=`API`; `period`=`{{ config.period }}`; offset/limit
  pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page size 100;
  computed output fields `record_id`, `report`.
- `goals_items_category`: GET `/index.php` - records at response root; query `date`=`{{ config.date
  }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`Goals.getItemsCategory`;
  `module`=`API`; `period`=`{{ config.period }}`; offset/limit pagination; offset parameter
  `filter_offset`; limit parameter `filter_limit`; page size 100; computed output fields
  `record_id`, `report`.
- `goals_items_name`: GET `/index.php` - records at response root; query `date`=`{{ config.date }}`;
  `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`Goals.getItemsName`; `module`=`API`;
  `period`=`{{ config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit
  parameter `filter_limit`; page size 100; computed output fields `record_id`, `report`.
- `goals_items_sku`: GET `/index.php` - records at response root; query `date`=`{{ config.date }}`;
  `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`Goals.getItemsSku`; `module`=`API`;
  `period`=`{{ config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit
  parameter `filter_limit`; page size 100; computed output fields `record_id`, `report`.
- `goals_visits_until_conversion`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`Goals.getVisitsUntilConversion`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `marketing_campaigns_reporting_content`: GET `/index.php` - records at response root; query
  `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`MarketingCampaignsReporting.getContent`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `marketing_campaigns_reporting_group`: GET `/index.php` - records at response root; query
  `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`MarketingCampaignsReporting.getGroup`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `marketing_campaigns_reporting_id`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`MarketingCampaignsReporting.getId`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `marketing_campaigns_reporting_keyword`: GET `/index.php` - records at response root; query
  `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`MarketingCampaignsReporting.getKeyword`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `marketing_campaigns_reporting_medium`: GET `/index.php` - records at response root; query
  `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`MarketingCampaignsReporting.getMedium`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `marketing_campaigns_reporting_name`: GET `/index.php` - records at response root; query
  `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`MarketingCampaignsReporting.getName`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `marketing_campaigns_reporting_placement`: GET `/index.php` - records at response root; query
  `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`MarketingCampaignsReporting.getPlacement`; `module`=`API`; `period`=`{{ config.period
  }}`; offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`;
  page size 100; computed output fields `record_id`, `report`.
- `marketing_campaigns_reporting_source`: GET `/index.php` - records at response root; query
  `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`MarketingCampaignsReporting.getSource`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `marketing_campaigns_reporting_source_medium`: GET `/index.php` - records at response root; query
  `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`MarketingCampaignsReporting.getSourceMedium`; `module`=`API`; `period`=`{{ config.period
  }}`; offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`;
  page size 100; computed output fields `record_id`, `report`.
- `multi_channel_conversion_attribution_channel_attribution`: GET `/index.php` - records at response
  root; query `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`MultiChannelConversionAttribution.getChannelAttribution`; `module`=`API`; `period`=`{{
  config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit parameter
  `filter_limit`; page size 100; computed output fields `record_id`, `report`.
- `multi_sites_all`: GET `/index.php` - records at response root; query `date`=`{{ config.date }}`;
  `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`MultiSites.getAll`; `module`=`API`;
  `period`=`{{ config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit
  parameter `filter_limit`; page size 100; computed output fields `record_id`, `report`.
- `multi_sites_one`: GET `/index.php` - records at response root; query `date`=`{{ config.date }}`;
  `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`MultiSites.getOne`; `module`=`API`;
  `period`=`{{ config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit
  parameter `filter_limit`; page size 100; computed output fields `record_id`, `report`.
- `page_performance`: GET `/index.php` - records at response root; query `date`=`{{ config.date }}`;
  `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`PagePerformance.get`; `module`=`API`;
  `period`=`{{ config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit
  parameter `filter_limit`; page size 100; computed output fields `report`.
- `referrers_summary`: GET `/index.php` - records at response root; query `date`=`{{ config.date
  }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`Referrers.get`; `module`=`API`;
  `period`=`{{ config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit
  parameter `filter_limit`; page size 100; computed output fields `report`.
- `referrers_ai_assistants`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`Referrers.getAIAssistants`; `module`=`API`; `period`=`{{ config.period }}`; offset/limit
  pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page size 100;
  computed output fields `record_id`, `report`.
- `referrers_all`: GET `/index.php` - records at response root; query `date`=`{{ config.date }}`;
  `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`Referrers.getAll`; `module`=`API`;
  `period`=`{{ config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit
  parameter `filter_limit`; page size 100; computed output fields `record_id`, `report`.
- `referrers_keywords`: GET `/index.php` - records at response root; query `date`=`{{ config.date
  }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`Referrers.getKeywords`;
  `module`=`API`; `period`=`{{ config.period }}`; offset/limit pagination; offset parameter
  `filter_offset`; limit parameter `filter_limit`; page size 100; computed output fields
  `record_id`, `report`.
- `referrers_referrer_type`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`Referrers.getReferrerType`; `module`=`API`; `period`=`{{ config.period }}`; offset/limit
  pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page size 100;
  computed output fields `record_id`, `report`.
- `referrers_search_engines`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`Referrers.getSearchEngines`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `referrers_socials`: GET `/index.php` - records at response root; query `date`=`{{ config.date
  }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`Referrers.getSocials`;
  `module`=`API`; `period`=`{{ config.period }}`; offset/limit pagination; offset parameter
  `filter_offset`; limit parameter `filter_limit`; page size 100; computed output fields
  `record_id`, `report`.
- `referrers_websites`: GET `/index.php` - records at response root; query `date`=`{{ config.date
  }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`Referrers.getWebsites`;
  `module`=`API`; `period`=`{{ config.period }}`; offset/limit pagination; offset parameter
  `filter_offset`; limit parameter `filter_limit`; page size 100; computed output fields
  `record_id`, `report`.
- `resolution_configuration`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`Resolution.getConfiguration`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `resolution_resolution`: GET `/index.php` - records at response root; query `date`=`{{ config.date
  }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`Resolution.getResolution`;
  `module`=`API`; `period`=`{{ config.period }}`; offset/limit pagination; offset parameter
  `filter_offset`; limit parameter `filter_limit`; page size 100; computed output fields
  `record_id`, `report`.
- `search_engine_keywords_performance_crawling_overview_bing`: GET `/index.php` - records at
  response root; query `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`SearchEngineKeywordsPerformance.getCrawlingOverviewBing`; `module`=`API`; `period`=`{{
  config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit parameter
  `filter_limit`; page size 100; computed output fields `report`.
- `search_engine_keywords_performance_keywords`: GET `/index.php` - records at response root; query
  `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`SearchEngineKeywordsPerformance.getKeywords`; `module`=`API`; `period`=`{{ config.period
  }}`; offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`;
  page size 100; computed output fields `record_id`, `report`.
- `search_engine_keywords_performance_keywords_bing`: GET `/index.php` - records at response root;
  query `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`SearchEngineKeywordsPerformance.getKeywordsBing`; `module`=`API`; `period`=`{{
  config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit parameter
  `filter_limit`; page size 100; computed output fields `record_id`, `report`.
- `search_engine_keywords_performance_keywords_google_image`: GET `/index.php` - records at response
  root; query `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`SearchEngineKeywordsPerformance.getKeywordsGoogleImage`; `module`=`API`; `period`=`{{
  config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit parameter
  `filter_limit`; page size 100; computed output fields `record_id`, `report`.
- `search_engine_keywords_performance_keywords_google_video`: GET `/index.php` - records at response
  root; query `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`SearchEngineKeywordsPerformance.getKeywordsGoogleVideo`; `module`=`API`; `period`=`{{
  config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit parameter
  `filter_limit`; page size 100; computed output fields `record_id`, `report`.
- `search_engine_keywords_performance_keywords_google_web`: GET `/index.php` - records at response
  root; query `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`SearchEngineKeywordsPerformance.getKeywordsGoogleWeb`; `module`=`API`; `period`=`{{
  config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit parameter
  `filter_limit`; page size 100; computed output fields `record_id`, `report`.
- `search_engine_keywords_performance_keywords_imported`: GET `/index.php` - records at response
  root; query `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`SearchEngineKeywordsPerformance.getKeywordsImported`; `module`=`API`; `period`=`{{
  config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit parameter
  `filter_limit`; page size 100; computed output fields `record_id`, `report`.
- `user_country_city`: GET `/index.php` - records at response root; query `date`=`{{ config.date
  }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`UserCountry.getCity`;
  `module`=`API`; `period`=`{{ config.period }}`; offset/limit pagination; offset parameter
  `filter_offset`; limit parameter `filter_limit`; page size 100; computed output fields
  `record_id`, `report`.
- `user_country_continent`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`UserCountry.getContinent`; `module`=`API`; `period`=`{{ config.period }}`; offset/limit
  pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page size 100;
  computed output fields `record_id`, `report`.
- `user_country_country`: GET `/index.php` - records at response root; query `date`=`{{ config.date
  }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`UserCountry.getCountry`;
  `module`=`API`; `period`=`{{ config.period }}`; offset/limit pagination; offset parameter
  `filter_offset`; limit parameter `filter_limit`; page size 100; computed output fields
  `record_id`, `report`.
- `user_country_region`: GET `/index.php` - records at response root; query `date`=`{{ config.date
  }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`UserCountry.getRegion`;
  `module`=`API`; `period`=`{{ config.period }}`; offset/limit pagination; offset parameter
  `filter_offset`; limit parameter `filter_limit`; page size 100; computed output fields
  `record_id`, `report`.
- `user_id_users`: GET `/index.php` - records at response root; query `date`=`{{ config.date }}`;
  `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`UserId.getUsers`; `module`=`API`;
  `period`=`{{ config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit
  parameter `filter_limit`; page size 100; computed output fields `record_id`, `report`.
- `user_language_language`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`UserLanguage.getLanguage`; `module`=`API`; `period`=`{{ config.period }}`; offset/limit
  pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page size 100;
  computed output fields `record_id`, `report`.
- `user_language_language_code`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`UserLanguage.getLanguageCode`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `users_flow_users_flow_pretty`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`UsersFlow.getUsersFlowPretty`; `module`=`API`; `period`=`{{ config.period }}`;
  offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page
  size 100; computed output fields `record_id`, `report`.
- `visit_frequency`: GET `/index.php` - records at response root; query `date`=`{{ config.date }}`;
  `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`VisitFrequency.get`; `module`=`API`;
  `period`=`{{ config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit
  parameter `filter_limit`; page size 100; computed output fields `report`.
- `visit_time_by_day_of_week`: GET `/index.php` - records at response root; query `date`=`{{
  config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`VisitTime.getByDayOfWeek`; `module`=`API`; `period`=`{{ config.period }}`; offset/limit
  pagination; offset parameter `filter_offset`; limit parameter `filter_limit`; page size 100;
  computed output fields `record_id`, `report`.
- `visit_time_visit_information_per_local_time`: GET `/index.php` - records at response root; query
  `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`VisitTime.getVisitInformationPerLocalTime`; `module`=`API`; `period`=`{{ config.period
  }}`; offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`;
  page size 100; computed output fields `record_id`, `report`.
- `visit_time_visit_information_per_server_time`: GET `/index.php` - records at response root; query
  `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`VisitTime.getVisitInformationPerServerTime`; `module`=`API`; `period`=`{{ config.period
  }}`; offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`;
  page size 100; computed output fields `record_id`, `report`.
- `visitor_interest_number_of_visits_by_days_since_last`: GET `/index.php` - records at response
  root; query `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`VisitorInterest.getNumberOfVisitsByDaysSinceLast`; `module`=`API`; `period`=`{{
  config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit parameter
  `filter_limit`; page size 100; computed output fields `record_id`, `report`.
- `visitor_interest_number_of_visits_by_visit_count`: GET `/index.php` - records at response root;
  query `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`VisitorInterest.getNumberOfVisitsByVisitCount`; `module`=`API`; `period`=`{{
  config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit parameter
  `filter_limit`; page size 100; computed output fields `record_id`, `report`.
- `visitor_interest_number_of_visits_per_page`: GET `/index.php` - records at response root; query
  `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`VisitorInterest.getNumberOfVisitsPerPage`; `module`=`API`; `period`=`{{ config.period
  }}`; offset/limit pagination; offset parameter `filter_offset`; limit parameter `filter_limit`;
  page size 100; computed output fields `record_id`, `report`.
- `visitor_interest_number_of_visits_per_visit_duration`: GET `/index.php` - records at response
  root; query `date`=`{{ config.date }}`; `format`=`JSON`; `idSite`=`{{ config.site_id }}`;
  `method`=`VisitorInterest.getNumberOfVisitsPerVisitDuration`; `module`=`API`; `period`=`{{
  config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit parameter
  `filter_limit`; page size 100; computed output fields `record_id`, `report`.
- `visits_summary`: GET `/index.php` - records at response root; query `date`=`{{ config.date }}`;
  `format`=`JSON`; `idSite`=`{{ config.site_id }}`; `method`=`VisitsSummary.get`; `module`=`API`;
  `period`=`{{ config.period }}`; offset/limit pagination; offset parameter `filter_offset`; limit
  parameter `filter_limit`; page size 100; computed output fields `report`.

## Write actions & risks

This connector is read-only. Read behavior: external Piwik/Matomo Reporting API read of site
analytics, site metadata, and report metadata.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 92 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2, destructive_admin=17, duplicate_of=8, non_data_endpoint=6,
  requires_elevated_scope=14.
