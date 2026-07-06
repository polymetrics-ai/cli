---
name: pm-piwik
description: Piwik / Matomo connector knowledge and safe action guide.
---

# pm-piwik

## Purpose

Reads Piwik/Matomo sites, recent visits, configured goals, and documented analytics reports through the Reporting API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- custom_dimension_id
- date
- last_minutes
- mode
- period
- site_id
- token_auth (secret)

## ETL Streams

- sites:
  - primary key: site_id
  - fields: main_url(), name(), site_id()
- visits:
  - primary key: visit_id
  - cursor: last_action_at
  - fields: last_action_at(), visit_id(), visitor_id()
- actions:
  - primary key: label
  - fields: hits(), label(), visits()
- goals:
  - primary key: goal_id
  - fields: active(), goal_id(), name()
- report_metadata:
  - primary key: record_id
  - fields: action(), category(), dimension(), documentation(), metrics(), metricsDocumentation(), module(), name(), processedMetrics(), record_id(), subcategory(), unique_id()
- live_counters:
  - primary key: report
  - fields: actions(), converted(), report(), visitors(), visits()
- site:
  - primary key: site_id
  - fields: currency(), main_url(), name(), site_id(), timezone(), type()
- sites_manager_sites_with_at_least_view_access:
  - primary key: site_id
  - fields: currency(), main_url(), name(), site_id(), timezone(), type()
- sites_manager_sites_with_view_access:
  - primary key: site_id
  - fields: currency(), main_url(), name(), site_id(), timezone(), type()
- sites_manager_sites_with_admin_access:
  - primary key: site_id
  - fields: currency(), main_url(), name(), site_id(), timezone(), type()
- actions_summary:
  - primary key: report
  - fields: avg_time_generation(), hits(), nb_downloads(), nb_keywords(), nb_outlinks(), nb_pageviews(), nb_searches(), nb_uniq_downloads(), nb_uniq_outlinks(), nb_uniq_pageviews(), report()
- actions_downloads:
  - primary key: record_id
  - fields: code(), idsubdatatable(), label(), nb_hits(), nb_visits(), record_id(), report(), segment()
- actions_entry_page_titles:
  - primary key: record_id
  - fields: avg_time_generation(), avg_time_on_page(), bounce_rate(), code(), entry_bounce_count(), entry_nb_visits(), exit_rate(), idsubdatatable(), label(), nb_conversions(), nb_conversions_entry(), nb_conversions_entry_rate(), record_id(), report(), revenue_entry(), revenue_per_entry(), revenue_per_visit(), segment()
- actions_entry_page_urls:
  - primary key: record_id
  - fields: avg_time_generation(), avg_time_on_page(), bounce_rate(), code(), entry_bounce_count(), entry_nb_visits(), exit_rate(), idsubdatatable(), label(), nb_conversions(), nb_conversions_entry(), nb_conversions_entry_rate(), record_id(), report(), revenue_entry(), revenue_per_entry(), revenue_per_visit(), segment()
- actions_exit_page_titles:
  - primary key: record_id
  - fields: avg_time_generation(), avg_time_on_page(), bounce_rate(), code(), exit_nb_visits(), exit_rate(), idsubdatatable(), label(), nb_visits(), record_id(), report(), segment()
- actions_exit_page_urls:
  - primary key: record_id
  - fields: avg_time_generation(), avg_time_on_page(), bounce_rate(), code(), exit_nb_visits(), exit_rate(), idsubdatatable(), label(), nb_visits(), record_id(), report(), segment()
- actions_outlinks:
  - primary key: record_id
  - fields: code(), idsubdatatable(), label(), nb_hits(), nb_visits(), record_id(), report(), segment()
- actions_page_titles:
  - primary key: record_id
  - fields: avg_time_generation(), avg_time_on_page(), bounce_rate(), code(), exit_rate(), idsubdatatable(), label(), nb_conversions(), nb_conversions_attrib(), nb_conversions_page_rate(), nb_hits(), nb_visits(), record_id(), report(), revenue_attrib(), revenue_per_visit(), segment()
- actions_page_titles_following_site_search:
  - primary key: record_id
  - fields: avg_time_generation(), avg_time_on_page(), bounce_rate(), code(), exit_rate(), idsubdatatable(), label(), nb_hits(), nb_hits_following_search(), record_id(), report(), segment()
- actions_page_urls_following_site_search:
  - primary key: record_id
  - fields: avg_time_generation(), avg_time_on_page(), bounce_rate(), code(), exit_rate(), idsubdatatable(), label(), nb_hits(), nb_hits_following_search(), record_id(), report(), segment()
- actions_site_search_categories:
  - primary key: record_id
  - fields: code(), exit_rate(), idsubdatatable(), label(), nb_pages_per_search(), nb_visits(), record_id(), report(), segment()
- actions_site_search_keywords:
  - primary key: record_id
  - fields: avg_time_generation(), avg_time_on_page(), bounce_rate(), code(), exit_rate(), idsubdatatable(), label(), nb_pages_per_search(), nb_visits(), record_id(), report(), segment()
- actions_site_search_no_result_keywords:
  - primary key: record_id
  - fields: avg_time_generation(), avg_time_on_page(), bounce_rate(), code(), exit_rate(), idsubdatatable(), label(), nb_visits(), record_id(), report(), segment()
- contents_content_names:
  - primary key: record_id
  - fields: code(), idsubdatatable(), interaction_rate(), label(), nb_impressions(), nb_interactions(), record_id(), report(), segment()
- contents_content_pieces:
  - primary key: record_id
  - fields: code(), idsubdatatable(), interaction_rate(), label(), nb_impressions(), nb_interactions(), record_id(), report(), segment()
- custom_dimensions_custom_dimension:
  - primary key: record_id
  - fields: avg_time_generation(), avg_time_on_dimension(), avg_time_on_site(), bounce_rate(), code(), exit_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_hits(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), segment()
- custom_variables_custom_variables:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- device_plugins_plugin:
  - primary key: record_id
  - fields: code(), idsubdatatable(), label(), nb_visits(), nb_visits_percentage(), record_id(), report(), segment()
- devices_detection_brand:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- devices_detection_browser_engines:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), segment()
- devices_detection_browser_versions:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), segment()
- devices_detection_browsers:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- devices_detection_model:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- devices_detection_os_families:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), segment()
- devices_detection_os_versions:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), segment()
- devices_detection_type:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- events_action:
  - primary key: record_id
  - fields: avg_event_value(), code(), idsubdatatable(), label(), max_event_value(), min_event_value(), nb_events(), nb_events_with_value(), nb_uniq_visitors(), nb_visits(), record_id(), report(), segment(), sum_event_value()
- events_category:
  - primary key: record_id
  - fields: avg_event_value(), code(), idsubdatatable(), label(), max_event_value(), min_event_value(), nb_events(), nb_events_with_value(), nb_uniq_visitors(), nb_visits(), record_id(), report(), segment(), sum_event_value()
- events_name:
  - primary key: record_id
  - fields: avg_event_value(), code(), idsubdatatable(), label(), max_event_value(), min_event_value(), nb_events(), nb_events_with_value(), nb_uniq_visitors(), nb_visits(), record_id(), report(), segment(), sum_event_value()
- goals_summary:
  - primary key: report
  - fields: avg_order_revenue(), conversion_rate(), items(), nb_conversions(), nb_visits_converted(), report(), revenue(), revenue_discount(), revenue_shipping(), revenue_subtotal(), revenue_tax()
- goals_days_to_conversion:
  - primary key: record_id
  - fields: code(), idsubdatatable(), label(), nb_conversions(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- goals_items_category:
  - primary key: record_id
  - fields: avg_price(), avg_quantity(), code(), conversion_rate(), idsubdatatable(), label(), nb_visits(), orders(), quantity(), record_id(), report(), revenue(), segment()
- goals_items_name:
  - primary key: record_id
  - fields: avg_price(), avg_quantity(), code(), conversion_rate(), idsubdatatable(), label(), nb_visits(), orders(), quantity(), record_id(), report(), revenue(), segment()
- goals_items_sku:
  - primary key: record_id
  - fields: avg_price(), avg_quantity(), code(), conversion_rate(), idsubdatatable(), label(), nb_visits(), orders(), quantity(), record_id(), report(), revenue(), segment()
- goals_visits_until_conversion:
  - primary key: record_id
  - fields: code(), idsubdatatable(), label(), nb_conversions(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- marketing_campaigns_reporting_content:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- marketing_campaigns_reporting_group:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- marketing_campaigns_reporting_id:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- marketing_campaigns_reporting_keyword:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- marketing_campaigns_reporting_medium:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- marketing_campaigns_reporting_name:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- marketing_campaigns_reporting_placement:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- marketing_campaigns_reporting_source:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- marketing_campaigns_reporting_source_medium:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- multi_channel_conversion_attribution_channel_attribution:
  - primary key: record_id
  - fields: code(), idsubdatatable(), label(), nb_attribution_conversions_firstInteraction(), nb_attribution_conversions_lastInteraction(), nb_attribution_conversions_lastNonDirect(), nb_attribution_conversions_linear(), nb_attribution_conversions_positionBased(), nb_attribution_conversions_timeDecay(), nb_attribution_revenue_firstInteraction(), nb_attribution_revenue_lastInteraction(), nb_attribution_revenue_lastNonDirect(), nb_attribution_revenue_linear(), nb_attribution_revenue_positionBased(), nb_attribution_revenue_timeDecay(), record_id(), report(), segment()
- multi_sites_all:
  - primary key: record_id
  - fields: actions_evolution(), ai_chatbots_requests(), ai_chatbots_requests_evolution(), code(), ecommerce_revenue(), ecommerce_revenue_evolution(), hits(), hits_evolution(), idsubdatatable(), label(), nb_actions(), nb_conversions(), nb_conversions_evolution(), nb_pageviews(), nb_visits(), orders(), orders_evolution(), pageviews_evolution(), record_id(), report(), revenue(), revenue_evolution(), segment(), visits_evolution()
- multi_sites_one:
  - primary key: record_id
  - fields: actions_evolution(), ai_chatbots_requests(), ai_chatbots_requests_evolution(), code(), ecommerce_revenue(), ecommerce_revenue_evolution(), hits(), hits_evolution(), idsubdatatable(), label(), nb_actions(), nb_conversions(), nb_conversions_evolution(), nb_pageviews(), nb_visits(), orders(), orders_evolution(), pageviews_evolution(), record_id(), report(), revenue(), revenue_evolution(), segment(), visits_evolution()
- page_performance:
  - primary key: report
  - fields: avg_page_load_time(), avg_time_dom_completion(), avg_time_dom_processing(), avg_time_network(), avg_time_on_load(), avg_time_server(), avg_time_transfer(), report()
- referrers_summary:
  - primary key: report
  - fields: Referrers_distinctAIAssistants(), Referrers_distinctCampaigns(), Referrers_distinctKeywords(), Referrers_distinctSearchEngines(), Referrers_distinctSocialNetworks(), Referrers_distinctWebsites(), Referrers_visitorsFromAIAssistants(), Referrers_visitorsFromAIAssistants_percent(), Referrers_visitorsFromCampaigns(), Referrers_visitorsFromCampaigns_percent(), Referrers_visitorsFromDirectEntry(), Referrers_visitorsFromDirectEntry_percent(), Referrers_visitorsFromSearchEngines(), Referrers_visitorsFromSearchEngines_percent(), Referrers_visitorsFromSocialNetworks(), Referrers_visitorsFromSocialNetworks_percent(), Referrers_visitorsFromWebsites(), Referrers_visitorsFromWebsites_percent(), report()
- referrers_ai_assistants:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- referrers_all:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), segment()
- referrers_keywords:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- referrers_referrer_type:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- referrers_search_engines:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- referrers_socials:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- referrers_websites:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- resolution_configuration:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), segment()
- resolution_resolution:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), segment()
- search_engine_keywords_performance_crawling_overview_bing:
  - primary key: report
  - fields: report()
- search_engine_keywords_performance_keywords:
  - primary key: record_id
  - fields: code(), idsubdatatable(), label(), nb_visits(), record_id(), report(), segment()
- search_engine_keywords_performance_keywords_bing:
  - primary key: record_id
  - fields: code(), ctr(), idsubdatatable(), label(), nb_clicks(), nb_impressions(), position(), record_id(), report(), segment()
- search_engine_keywords_performance_keywords_google_image:
  - primary key: record_id
  - fields: code(), ctr(), idsubdatatable(), label(), nb_clicks(), nb_impressions(), position(), record_id(), report(), segment()
- search_engine_keywords_performance_keywords_google_video:
  - primary key: record_id
  - fields: code(), ctr(), idsubdatatable(), label(), nb_clicks(), nb_impressions(), position(), record_id(), report(), segment()
- search_engine_keywords_performance_keywords_google_web:
  - primary key: record_id
  - fields: code(), ctr(), idsubdatatable(), label(), nb_clicks(), nb_impressions(), position(), record_id(), report(), segment()
- search_engine_keywords_performance_keywords_imported:
  - primary key: record_id
  - fields: code(), ctr(), idsubdatatable(), label(), nb_clicks(), nb_impressions(), position(), record_id(), report(), segment()
- user_country_city:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- user_country_continent:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- user_country_country:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- user_country_region:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- user_id_users:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_visits(), nb_visits_converted(), record_id(), report(), segment()
- user_language_language:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), segment()
- user_language_language_code:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), segment()
- users_flow_users_flow_pretty:
  - primary key: record_id
  - fields: code(), exit_rate(), idsubdatatable(), label(), nb_exits(), nb_proceeded(), nb_visits(), proceeded_rate(), record_id(), report(), segment()
- visit_frequency:
  - primary key: report
  - fields: avg_time_on_site_new(), avg_time_on_site_returning(), bounce_rate_new(), bounce_rate_returning(), max_actions_new(), max_actions_returning(), nb_actions_new(), nb_actions_per_visit_new(), nb_actions_per_visit_returning(), nb_actions_returning(), nb_uniq_visitors_new(), nb_uniq_visitors_returning(), nb_users_new(), nb_users_returning(), nb_visits_new(), nb_visits_returning(), report()
- visit_time_by_day_of_week:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), segment()
- visit_time_visit_information_per_local_time:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), segment()
- visit_time_visit_information_per_server_time:
  - primary key: record_id
  - fields: avg_time_on_site(), bounce_rate(), code(), conversion_rate(), idsubdatatable(), label(), nb_actions(), nb_actions_per_visit(), nb_conversions(), nb_uniq_visitors(), nb_users(), nb_visits(), record_id(), report(), revenue(), revenue_per_visit(), segment()
- visitor_interest_number_of_visits_by_days_since_last:
  - primary key: record_id
  - fields: code(), idsubdatatable(), label(), nb_visits(), record_id(), report(), segment()
- visitor_interest_number_of_visits_by_visit_count:
  - primary key: record_id
  - fields: code(), idsubdatatable(), label(), nb_visits(), nb_visits_percentage(), record_id(), report(), segment()
- visitor_interest_number_of_visits_per_page:
  - primary key: record_id
  - fields: code(), idsubdatatable(), label(), nb_visits(), record_id(), report(), segment()
- visitor_interest_number_of_visits_per_visit_duration:
  - primary key: record_id
  - fields: code(), idsubdatatable(), label(), nb_visits(), record_id(), report(), segment()
- visits_summary:
  - primary key: report
  - fields: avg_time_on_site(), bounce_rate(), max_actions(), nb_actions(), nb_actions_per_visit(), nb_uniq_visitors(), nb_users(), nb_visits(), report()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Piwik/Matomo Reporting API read of site analytics, site metadata, and report metadata
- approval: none; read-only analytics sync
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect piwik
```

### Inspect as structured JSON

```bash
pm connectors inspect piwik --json
```

## Agent Rules

- Run pm connectors inspect piwik before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
