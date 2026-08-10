# pm connectors inspect piwik

```text
NAME
  pm connectors inspect piwik - Piwik / Matomo connector manual

SYNOPSIS
  pm connectors inspect piwik
  pm connectors inspect piwik --json
  pm credentials add <name> --connector piwik [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Piwik/Matomo sites, recent visits, configured goals, and documented analytics reports through the Reporting API.

ICON
  id: simple-icons-matomo
  asset: icons/simple-icons/matomo.svg
  title: Matomo
  simple_icon_slug: matomo
  simple_icon_hex: 3152A0
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Matomo
  match: curated-alias
  matched_by: matomo

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  custom_dimension_id
  date
  last_minutes
  mode
  period
  site_id
  token_auth (secret) (required)

ETL STREAMS
  sites:
    primary key: site_id
    fields: main_url(string), name(string), site_id(string)
  visits:
    primary key: visit_id
    cursor: last_action_at
    fields: last_action_at(string), visit_id(string), visitor_id(string)
  actions:
    primary key: label
    fields: hits(number), label(string), visits(number)
  goals:
    primary key: goal_id
    fields: active(boolean), goal_id(string), name(string)
  report_metadata:
    primary key: record_id
    fields: action(string), category(string), dimension(string), documentation(string), metrics(object), metricsDocumentation(object), module(string), name(string), processedMetrics(object), record_id(string), subcategory(string), unique_id(string)
  live_counters:
    primary key: report
    fields: actions(number), converted(number), report(string), visitors(number), visits(number)
  site:
    primary key: site_id
    fields: currency(string), main_url(string), name(string), site_id(string), timezone(string), type(string)
  sites_manager_sites_with_at_least_view_access:
    primary key: site_id
    fields: currency(string), main_url(string), name(string), site_id(string), timezone(string), type(string)
  sites_manager_sites_with_view_access:
    primary key: site_id
    fields: currency(string), main_url(string), name(string), site_id(string), timezone(string), type(string)
  sites_manager_sites_with_admin_access:
    primary key: site_id
    fields: currency(string), main_url(string), name(string), site_id(string), timezone(string), type(string)
  actions_summary:
    primary key: report
    fields: avg_time_generation(number), hits(number), nb_downloads(number), nb_keywords(number), nb_outlinks(number), nb_pageviews(number), nb_searches(number), nb_uniq_downloads(number), nb_uniq_outlinks(number), nb_uniq_pageviews(number), report(string)
  actions_downloads:
    primary key: record_id
    fields: code(string), idsubdatatable(number), label(string), nb_hits(number), nb_visits(number), record_id(string), report(string), segment(string)
  actions_entry_page_titles:
    primary key: record_id
    fields: avg_time_generation(number), avg_time_on_page(number), bounce_rate(number), code(string), entry_bounce_count(number), entry_nb_visits(number), exit_rate(number), idsubdatatable(number), label(string), nb_conversions(number), nb_conversions_entry(number), nb_conversions_entry_rate(number), record_id(string), report(string), revenue_entry(number), revenue_per_entry(number), revenue_per_visit(number), segment(string)
  actions_entry_page_urls:
    primary key: record_id
    fields: avg_time_generation(number), avg_time_on_page(number), bounce_rate(number), code(string), entry_bounce_count(number), entry_nb_visits(number), exit_rate(number), idsubdatatable(number), label(string), nb_conversions(number), nb_conversions_entry(number), nb_conversions_entry_rate(number), record_id(string), report(string), revenue_entry(number), revenue_per_entry(number), revenue_per_visit(number), segment(string)
  actions_exit_page_titles:
    primary key: record_id
    fields: avg_time_generation(number), avg_time_on_page(number), bounce_rate(number), code(string), exit_nb_visits(number), exit_rate(number), idsubdatatable(number), label(string), nb_visits(number), record_id(string), report(string), segment(string)
  actions_exit_page_urls:
    primary key: record_id
    fields: avg_time_generation(number), avg_time_on_page(number), bounce_rate(number), code(string), exit_nb_visits(number), exit_rate(number), idsubdatatable(number), label(string), nb_visits(number), record_id(string), report(string), segment(string)
  actions_outlinks:
    primary key: record_id
    fields: code(string), idsubdatatable(number), label(string), nb_hits(number), nb_visits(number), record_id(string), report(string), segment(string)
  actions_page_titles:
    primary key: record_id
    fields: avg_time_generation(number), avg_time_on_page(number), bounce_rate(number), code(string), exit_rate(number), idsubdatatable(number), label(string), nb_conversions(number), nb_conversions_attrib(number), nb_conversions_page_rate(number), nb_hits(number), nb_visits(number), record_id(string), report(string), revenue_attrib(number), revenue_per_visit(number), segment(string)
  actions_page_titles_following_site_search:
    primary key: record_id
    fields: avg_time_generation(number), avg_time_on_page(number), bounce_rate(number), code(string), exit_rate(number), idsubdatatable(number), label(string), nb_hits(number), nb_hits_following_search(number), record_id(string), report(string), segment(string)
  actions_page_urls_following_site_search:
    primary key: record_id
    fields: avg_time_generation(number), avg_time_on_page(number), bounce_rate(number), code(string), exit_rate(number), idsubdatatable(number), label(string), nb_hits(number), nb_hits_following_search(number), record_id(string), report(string), segment(string)
  actions_site_search_categories:
    primary key: record_id
    fields: code(string), exit_rate(number), idsubdatatable(number), label(string), nb_pages_per_search(number), nb_visits(number), record_id(string), report(string), segment(string)
  actions_site_search_keywords:
    primary key: record_id
    fields: avg_time_generation(number), avg_time_on_page(number), bounce_rate(number), code(string), exit_rate(number), idsubdatatable(number), label(string), nb_pages_per_search(number), nb_visits(number), record_id(string), report(string), segment(string)
  actions_site_search_no_result_keywords:
    primary key: record_id
    fields: avg_time_generation(number), avg_time_on_page(number), bounce_rate(number), code(string), exit_rate(number), idsubdatatable(number), label(string), nb_visits(number), record_id(string), report(string), segment(string)
  contents_content_names:
    primary key: record_id
    fields: code(string), idsubdatatable(number), interaction_rate(number), label(string), nb_impressions(number), nb_interactions(number), record_id(string), report(string), segment(string)
  contents_content_pieces:
    primary key: record_id
    fields: code(string), idsubdatatable(number), interaction_rate(number), label(string), nb_impressions(number), nb_interactions(number), record_id(string), report(string), segment(string)
  custom_dimensions_custom_dimension:
    primary key: record_id
    fields: avg_time_generation(number), avg_time_on_dimension(number), avg_time_on_site(number), bounce_rate(number), code(string), exit_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_hits(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), segment(string)
  custom_variables_custom_variables:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  device_plugins_plugin:
    primary key: record_id
    fields: code(string), idsubdatatable(number), label(string), nb_visits(number), nb_visits_percentage(number), record_id(string), report(string), segment(string)
  devices_detection_brand:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  devices_detection_browser_engines:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), segment(string)
  devices_detection_browser_versions:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), segment(string)
  devices_detection_browsers:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  devices_detection_model:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  devices_detection_os_families:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), segment(string)
  devices_detection_os_versions:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), segment(string)
  devices_detection_type:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  events_action:
    primary key: record_id
    fields: avg_event_value(number), code(string), idsubdatatable(number), label(string), max_event_value(number), min_event_value(number), nb_events(number), nb_events_with_value(number), nb_uniq_visitors(number), nb_visits(number), record_id(string), report(string), segment(string), sum_event_value(number)
  events_category:
    primary key: record_id
    fields: avg_event_value(number), code(string), idsubdatatable(number), label(string), max_event_value(number), min_event_value(number), nb_events(number), nb_events_with_value(number), nb_uniq_visitors(number), nb_visits(number), record_id(string), report(string), segment(string), sum_event_value(number)
  events_name:
    primary key: record_id
    fields: avg_event_value(number), code(string), idsubdatatable(number), label(string), max_event_value(number), min_event_value(number), nb_events(number), nb_events_with_value(number), nb_uniq_visitors(number), nb_visits(number), record_id(string), report(string), segment(string), sum_event_value(number)
  goals_summary:
    primary key: report
    fields: avg_order_revenue(number), conversion_rate(number), items(number), nb_conversions(number), nb_visits_converted(number), report(string), revenue(number), revenue_discount(number), revenue_shipping(number), revenue_subtotal(number), revenue_tax(number)
  goals_days_to_conversion:
    primary key: record_id
    fields: code(string), idsubdatatable(number), label(string), nb_conversions(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  goals_items_category:
    primary key: record_id
    fields: avg_price(number), avg_quantity(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_visits(number), orders(number), quantity(number), record_id(string), report(string), revenue(number), segment(string)
  goals_items_name:
    primary key: record_id
    fields: avg_price(number), avg_quantity(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_visits(number), orders(number), quantity(number), record_id(string), report(string), revenue(number), segment(string)
  goals_items_sku:
    primary key: record_id
    fields: avg_price(number), avg_quantity(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_visits(number), orders(number), quantity(number), record_id(string), report(string), revenue(number), segment(string)
  goals_visits_until_conversion:
    primary key: record_id
    fields: code(string), idsubdatatable(number), label(string), nb_conversions(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  marketing_campaigns_reporting_content:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  marketing_campaigns_reporting_group:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  marketing_campaigns_reporting_id:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  marketing_campaigns_reporting_keyword:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  marketing_campaigns_reporting_medium:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  marketing_campaigns_reporting_name:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  marketing_campaigns_reporting_placement:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  marketing_campaigns_reporting_source:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  marketing_campaigns_reporting_source_medium:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  multi_channel_conversion_attribution_channel_attribution:
    primary key: record_id
    fields: code(string), idsubdatatable(number), label(string), nb_attribution_conversions_firstInteraction(number), nb_attribution_conversions_lastInteraction(number), nb_attribution_conversions_lastNonDirect(number), nb_attribution_conversions_linear(number), nb_attribution_conversions_positionBased(number), nb_attribution_conversions_timeDecay(number), nb_attribution_revenue_firstInteraction(number), nb_attribution_revenue_lastInteraction(number), nb_attribution_revenue_lastNonDirect(number), nb_attribution_revenue_linear(number), nb_attribution_revenue_positionBased(number), nb_attribution_revenue_timeDecay(number), record_id(string), report(string), segment(string)
  multi_sites_all:
    primary key: record_id
    fields: actions_evolution(number), ai_chatbots_requests(number), ai_chatbots_requests_evolution(number), code(string), ecommerce_revenue(number), ecommerce_revenue_evolution(number), hits(number), hits_evolution(number), idsubdatatable(number), label(string), nb_actions(number), nb_conversions(number), nb_conversions_evolution(number), nb_pageviews(number), nb_visits(number), orders(number), orders_evolution(number), pageviews_evolution(number), record_id(string), report(string), revenue(number), revenue_evolution(number), segment(string), visits_evolution(number)
  multi_sites_one:
    primary key: record_id
    fields: actions_evolution(number), ai_chatbots_requests(number), ai_chatbots_requests_evolution(number), code(string), ecommerce_revenue(number), ecommerce_revenue_evolution(number), hits(number), hits_evolution(number), idsubdatatable(number), label(string), nb_actions(number), nb_conversions(number), nb_conversions_evolution(number), nb_pageviews(number), nb_visits(number), orders(number), orders_evolution(number), pageviews_evolution(number), record_id(string), report(string), revenue(number), revenue_evolution(number), segment(string), visits_evolution(number)
  page_performance:
    primary key: report
    fields: avg_page_load_time(number), avg_time_dom_completion(number), avg_time_dom_processing(number), avg_time_network(number), avg_time_on_load(number), avg_time_server(number), avg_time_transfer(number), report(string)
  referrers_summary:
    primary key: report
    fields: Referrers_distinctAIAssistants(number), Referrers_distinctCampaigns(number), Referrers_distinctKeywords(number), Referrers_distinctSearchEngines(number), Referrers_distinctSocialNetworks(number), Referrers_distinctWebsites(number), Referrers_visitorsFromAIAssistants(number), Referrers_visitorsFromAIAssistants_percent(number), Referrers_visitorsFromCampaigns(number), Referrers_visitorsFromCampaigns_percent(number), Referrers_visitorsFromDirectEntry(number), Referrers_visitorsFromDirectEntry_percent(number), Referrers_visitorsFromSearchEngines(number), Referrers_visitorsFromSearchEngines_percent(number), Referrers_visitorsFromSocialNetworks(number), Referrers_visitorsFromSocialNetworks_percent(number), Referrers_visitorsFromWebsites(number), Referrers_visitorsFromWebsites_percent(number), report(string)
  referrers_ai_assistants:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  referrers_all:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), segment(string)
  referrers_keywords:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  referrers_referrer_type:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  referrers_search_engines:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  referrers_socials:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  referrers_websites:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  resolution_configuration:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), segment(string)
  resolution_resolution:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), segment(string)
  search_engine_keywords_performance_crawling_overview_bing:
    primary key: report
    fields: report(string)
  search_engine_keywords_performance_keywords:
    primary key: record_id
    fields: code(string), idsubdatatable(number), label(string), nb_visits(number), record_id(string), report(string), segment(string)
  search_engine_keywords_performance_keywords_bing:
    primary key: record_id
    fields: code(string), ctr(number), idsubdatatable(number), label(string), nb_clicks(number), nb_impressions(number), position(number), record_id(string), report(string), segment(string)
  search_engine_keywords_performance_keywords_google_image:
    primary key: record_id
    fields: code(string), ctr(number), idsubdatatable(number), label(string), nb_clicks(number), nb_impressions(number), position(number), record_id(string), report(string), segment(string)
  search_engine_keywords_performance_keywords_google_video:
    primary key: record_id
    fields: code(string), ctr(number), idsubdatatable(number), label(string), nb_clicks(number), nb_impressions(number), position(number), record_id(string), report(string), segment(string)
  search_engine_keywords_performance_keywords_google_web:
    primary key: record_id
    fields: code(string), ctr(number), idsubdatatable(number), label(string), nb_clicks(number), nb_impressions(number), position(number), record_id(string), report(string), segment(string)
  search_engine_keywords_performance_keywords_imported:
    primary key: record_id
    fields: code(string), ctr(number), idsubdatatable(number), label(string), nb_clicks(number), nb_impressions(number), position(number), record_id(string), report(string), segment(string)
  user_country_city:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  user_country_continent:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  user_country_country:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  user_country_region:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  user_id_users:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_visits(number), nb_visits_converted(number), record_id(string), report(string), segment(string)
  user_language_language:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), segment(string)
  user_language_language_code:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), segment(string)
  users_flow_users_flow_pretty:
    primary key: record_id
    fields: code(string), exit_rate(number), idsubdatatable(number), label(string), nb_exits(number), nb_proceeded(number), nb_visits(number), proceeded_rate(number), record_id(string), report(string), segment(string)
  visit_frequency:
    primary key: report
    fields: avg_time_on_site_new(number), avg_time_on_site_returning(number), bounce_rate_new(number), bounce_rate_returning(number), max_actions_new(number), max_actions_returning(number), nb_actions_new(number), nb_actions_per_visit_new(number), nb_actions_per_visit_returning(number), nb_actions_returning(number), nb_uniq_visitors_new(number), nb_uniq_visitors_returning(number), nb_users_new(number), nb_users_returning(number), nb_visits_new(number), nb_visits_returning(number), report(string)
  visit_time_by_day_of_week:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), segment(string)
  visit_time_visit_information_per_local_time:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), segment(string)
  visit_time_visit_information_per_server_time:
    primary key: record_id
    fields: avg_time_on_site(number), bounce_rate(number), code(string), conversion_rate(number), idsubdatatable(number), label(string), nb_actions(number), nb_actions_per_visit(number), nb_conversions(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), record_id(string), report(string), revenue(number), revenue_per_visit(number), segment(string)
  visitor_interest_number_of_visits_by_days_since_last:
    primary key: record_id
    fields: code(string), idsubdatatable(number), label(string), nb_visits(number), record_id(string), report(string), segment(string)
  visitor_interest_number_of_visits_by_visit_count:
    primary key: record_id
    fields: code(string), idsubdatatable(number), label(string), nb_visits(number), nb_visits_percentage(number), record_id(string), report(string), segment(string)
  visitor_interest_number_of_visits_per_page:
    primary key: record_id
    fields: code(string), idsubdatatable(number), label(string), nb_visits(number), record_id(string), report(string), segment(string)
  visitor_interest_number_of_visits_per_visit_duration:
    primary key: record_id
    fields: code(string), idsubdatatable(number), label(string), nb_visits(number), record_id(string), report(string), segment(string)
  visits_summary:
    primary key: report
    fields: avg_time_on_site(number), bounce_rate(number), max_actions(number), nb_actions(number), nb_actions_per_visit(number), nb_uniq_visitors(number), nb_users(number), nb_visits(number), report(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Piwik/Matomo Reporting API read of site analytics, site metadata, and report metadata
  approval: none; read-only analytics sync
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect piwik

  # Inspect as structured JSON
  pm connectors inspect piwik --json

AGENT WORKFLOW
  - Run pm connectors inspect piwik before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
