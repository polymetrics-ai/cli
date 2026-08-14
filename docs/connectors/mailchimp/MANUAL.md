# pm connectors inspect mailchimp

```text
NAME
  pm connectors inspect mailchimp - Mailchimp connector manual

SYNOPSIS
  pm connectors inspect mailchimp
  pm connectors inspect mailchimp --json
  pm credentials add <name> --connector mailchimp [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Mailchimp Marketing API audiences, members, campaigns, reports, automations, templates, files, batches, webhooks, ecommerce, reporting, and related resources; exposes typed approval-gated Mailchimp mutations where the declarative engine can model the documented operation safely.

ICON
  id: mailchimp
  asset: icons/mailchimp.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://mailchimp.com/developer/release-notes/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  app_id
  audience_id
  batch_id
  batch_webhook_id
  campaign_id
  cart_id
  connected_site_id
  contact_id
  conversation_id
  customer_id
  data_center (required)
  domain_name
  export_id
  feedback_id
  file_id
  folder_id
  image_id
  interest_category_id
  interest_id
  journey_id
  line_id
  link_id
  list_id
  merge_id
  message_id
  mode
  month
  note_id
  order_id
  outreach_id
  page_id
  product_id
  promo_code_id
  promo_rule_id
  question_id
  report_id
  response_id
  search_query
  segment_id
  sms_campaign_id
  start_date
  step_id
  store_id
  subscriber_hash
  survey_id
  template_id
  variant_id
  webhook_id
  workflow_email_id
  workflow_id
  access_token (secret)
  api_key (secret)

ETL STREAMS
  account_exports_exports:
    primary key: id
    fields: _links(object), download_url(string), export_id(integer), finished(string), id(string), size_in_bytes(integer), started(string)
  activity_feed_chimp_chatter:
    primary key: campaign_id
    fields: campaign_id(string), list_id(string), message(string), title(string), type(string), update_time(string), url(string)
  audiences:
    primary key: id
    fields: enabled_channels(array), id(string), name(string), stats(object)
  audiences_contacts:
    primary key: id
    fields: audience_id(string), created_at(string), email_channel(object), id(string), language(string), last_updated_at(string), merge_fields(object), sms_channel(object), source(object), status(string), tags(array)
  authorized_apps_apps:
    primary key: id
    fields: _links(object), description(string), id(integer), name(string), users(array)
  automations:
    primary key: id
    cursor: create_time
    fields: _links(object), create_time(string), emails_sent(integer), id(string), recipients(object), report_summary(object), settings(object), start_time(string), status(string), tracking(object), trigger_settings(object)
  automations_emails:
    primary key: id
    cursor: create_time
    fields: _links(object), archive_url(string), content_type(string), create_time(string), delay(object), emails_sent(integer), has_logo_merge_tag(boolean), id(string), needs_block_refresh(boolean), position(integer), recipients(object), report_summary(object), send_time(string), settings(object), social_card(object), start_time(string), status(string), tracking(object), trigger_settings(object), web_id(integer), workflow_id(string)
  automations_emails_queue:
    primary key: id
    fields: _links(object), email_address(string), email_id(string), id(string), list_id(string), next_send(string), workflow_email_id(string), workflow_id(string)
  automations_removed_subscribers_subscribers:
    primary key: id
    fields: _links(object), email_address(string), id(string), list_id(string), workflow_id(string)
  batch_webhooks_webhooks:
    primary key: id
    fields: _links(object), enabled(boolean), id(string), url(string)
  batches:
    primary key: id
    fields: _links(object), completed_at(string), errored_operations(integer), finished_operations(integer), id(string), response_body_url(string), status(string), submitted_at(string), total_operations(integer)
  campaign_folders_folders:
    primary key: id
    fields: _links(object), count(integer), id(string), name(string)
  campaigns:
    primary key: id
    cursor: create_time
    fields: _links(object), ab_split_opts(object), archive_url(string), content_type(string), create_time(string), delivery_status(object), emails_sent(integer), id(string), long_archive_url(string), needs_block_refresh(boolean), parent_campaign_id(string), recipients(object), report_summary(object), resend_shortcut_eligibility(object), resend_shortcut_usage(object), resendable(boolean), rss_opts(object), send_time(string), settings(object), social_card(object), status(object), tracking(object), type(object), variate_settings(object), web_id(integer)
  campaigns_feedback:
    primary key: campaign_id
    cursor: updated_at
    fields: _links(object), block_id(integer), campaign_id(string), created_at(string), created_by(string), feedback_id(integer), is_complete(boolean), message(string), parent_id(integer), source(string), updated_at(string)
  connected_sites_sites:
    primary key: store_id
    cursor: updated_at
    fields: _links(object), created_at(string), domain(string), foreign_id(string), is_pixel_enabled(boolean), platform(string), site_script(object), store_id(string), updated_at(string)
  conversations:
    primary key: id
    fields: _links(object), campaign_id(string), from_email(string), from_label(string), id(string), last_message(object), list_id(string), message_count(integer), subject(string), unread_messages(integer)
  conversations_messages_conversation_messages:
    primary key: id
    fields: _links(object), conversation_id(string), from_email(string), from_label(string), id(string), list_id(integer), message(string), read(boolean), subject(string), timestamp(string)
  ecommerce_orders:
    primary key: id
    fields: _links(object), billing_address(object), campaign_id(string), cancelled_at_foreign(string), cart_id(string), currency_code(string), customer(object), discount_total(number), financial_status(string), fulfillment_status(string), id(string), landing_site(string), lines(array), order_total(number), order_url(string), outreach(object), processed_at_foreign(string), promos(array), shipping_address(object), shipping_total(number), store_id(string), tax_total(number), tracking_carrier(string), tracking_code(string), tracking_number(string), tracking_url(string), updated_at_foreign(string)
  ecommerce_stores:
    primary key: id
    cursor: updated_at
    fields: _links(object), address(object), automations(object), connected_site(object), created_at(string), currency_code(string), domain(string), email_address(string), id(string), is_syncing(boolean), list_id(string), list_is_active(boolean), money_format(string), name(string), phone(string), platform(string), primary_locale(string), timezone(string), updated_at(string)
  ecommerce_stores_carts:
    primary key: id
    cursor: updated_at
    fields: _links(object), campaign_id(string), checkout_url(string), created_at(string), currency_code(string), customer(object), id(string), lines(array), order_total(number), store_id(string), tax_total(number), updated_at(string)
  ecommerce_stores_carts_lines:
    primary key: id
    fields: _links(object), cart_id(string), id(string), price(number), product_id(string), product_title(string), product_variant_id(string), product_variant_title(string), quantity(integer), store_id(string)
  ecommerce_stores_customers:
    primary key: id
    cursor: updated_at
    fields: _links(object), address(object), company(string), created_at(string), email_address(string), first_name(string), id(string), last_name(string), opt_in_status(boolean), orders_count(integer), sms_phone_number(string), store_id(string), total_spent(number), updated_at(string)
  ecommerce_stores_orders:
    primary key: id
    fields: _links(object), billing_address(object), campaign_id(string), cancelled_at_foreign(string), cart_id(string), currency_code(string), customer(object), discount_total(number), financial_status(string), fulfillment_status(string), id(string), landing_site(string), lines(array), order_total(number), order_url(string), outreach(object), processed_at_foreign(string), promos(array), shipping_address(object), shipping_total(number), store_id(string), tax_total(number), tracking_carrier(string), tracking_code(string), tracking_number(string), tracking_url(string), updated_at_foreign(string)
  ecommerce_stores_orders_lines:
    primary key: id
    fields: _links(object), discount(number), id(string), image_url(string), order_id(string), price(number), product_id(string), product_title(string), product_variant_id(string), product_variant_title(string), quantity(integer), store_id(string)
  ecommerce_stores_products:
    primary key: id
    fields: _links(object), currency_code(string), description(string), handle(string), id(string), image_url(string), images(array), published_at_foreign(string), store_id(string), title(string), type(string), url(string), variants(array), vendor(string)
  ecommerce_stores_products_images:
    primary key: id
    fields: _links(object), id(string), product_id(string), store_id(string), url(string), variant_ids(array)
  ecommerce_stores_products_variants:
    primary key: id
    cursor: updated_at
    fields: _links(object), backorders(string), created_at(string), id(string), image_url(string), inventory_quantity(integer), price(number), product_id(string), sku(string), store_id(string), title(string), updated_at(string), url(string), visibility(string)
  ecommerce_stores_promo_rules:
    primary key: id
    fields: _links(object), amount(number), created_at_foreign(string), description(string), enabled(boolean), ends_at(string), id(string), starts_at(string), store_id(string), target(string), title(string), type(string), updated_at_foreign(string)
  ecommerce_stores_promo_rules_promo_codes:
    primary key: id
    fields: _links(object), code(string), created_at_foreign(string), enabled(boolean), id(string), promo_rule_id(string), redemption_url(string), store_id(string), updated_at_foreign(string), usage_count(integer)
  facebook_ads:
    primary key: id
    fields: id(string)
  file_manager_files:
    primary key: id
    fields: _links(object), created_at(string), created_by(string), folder_id(integer), full_size_url(string), height(integer), id(integer), name(string), size(integer), thumbnail_url(string), type(string), width(integer)
  file_manager_folders:
    primary key: id
    fields: _links(object), created_at(string), created_by(string), file_count(integer), id(integer), name(string)
  file_manager_folders_files:
    primary key: id
    fields: _links(object), created_at(string), created_by(string), folder_id(string), full_size_url(string), height(integer), id(integer), name(string), size(integer), thumbnail_url(string), type(string), width(integer)
  landing_pages:
    primary key: id
    cursor: updated_at
    fields: _links(object), created_at(string), created_by_source(string), description(string), id(string), list_id(string), name(string), published_at(string), status(string), store_id(string), template_id(integer), title(string), tracking(object), unpublished_at(string), updated_at(string), url(string), web_id(integer)
  lists:
    primary key: id
    cursor: date_created
    fields: _links(object), beamer_address(string), campaign_defaults(object), contact(object), date_created(string), double_optin(boolean), email_type_option(boolean), has_welcome(boolean), id(string), list_rating(integer), marketing_permissions(boolean), modules(array), name(string), notify_on_subscribe(string), notify_on_unsubscribe(string), permission_reminder(string), stats(object), subscribe_url_long(string), subscribe_url_short(string), use_archive_bar(boolean), visibility(string), web_id(integer)
  lists_abuse_reports:
    primary key: id
    fields: _links(object), campaign_id(string), date(string), email_address(string), email_id(string), id(integer), list_id(string), merge_fields(object), vip(boolean)
  lists_activity:
    primary key: list_id
    fields: _links(object), day(string), emails_sent(integer), hard_bounce(integer), list_id(string), other_adds(integer), other_removes(integer), recipient_clicks(integer), soft_bounce(integer), subs(integer), unique_opens(integer), unsubs(integer)
  lists_clients:
    primary key: list_id
    fields: client(string), list_id(string), members(integer)
  lists_growth_history_history:
    primary key: list_id
    fields: _links(object), cleaned(integer), deleted(integer), existing(integer), imports(integer), list_id(string), month(string), optins(integer), pending(integer), reconfirm(integer), subscribed(integer), transactional(integer), unsubscribed(integer)
  lists_interest_categories_categories:
    primary key: id
    fields: _links(object), display_order(integer), id(string), list_id(string), title(string), type(string)
  lists_interest_categories_interests:
    primary key: id
    fields: _links(object), category_id(string), display_order(integer), id(string), interest_category_id(string), list_id(string), name(string), subscriber_count(string)
  lists_locations:
    primary key: list_id
    fields: cc(string), country(string), list_id(string), percent(number), total(integer)
  lists_members:
    primary key: id
    cursor: last_changed
    fields: _links(object), consents_to_one_to_one_messaging(boolean), contact_id(string), email_address(string), email_client(string), email_type(string), full_name(string), id(string), interests(object), ip_opt(string), ip_signup(string), language(string), last_changed(string), last_note(object), list_id(string), location(object), marketing_permissions(array), member_rating(integer), merge_fields(object), sms_phone_number(string), sms_subscription_last_updated(string), sms_subscription_status(string), source(string), stats(object), status(string), tags(array), tags_count(integer), timestamp_opt(string), timestamp_signup(string), unique_email_id(string), unsubscribe_reason(string), vip(boolean), web_id(integer)
  lists_members_activity:
    primary key: campaign_id
    fields: action(string), campaign_id(string), list_id(string), parent_campaign(string), subscriber_hash(string), timestamp(string), title(string), type(string), url(string)
  lists_members_events:
    primary key: list_id
    fields: list_id(string), name(string), occurred_at(string), properties(object), subscriber_hash(string)
  lists_members_goals:
    primary key: list_id
    fields: data(string), event(string), goal_id(integer), last_visited_at(string), list_id(string), subscriber_hash(string)
  lists_members_notes:
    primary key: id
    cursor: updated_at
    fields: _links(object), contact_id(string), created_at(string), created_by(string), email_id(string), id(integer), list_id(string), note(string), subscriber_hash(string), updated_at(string)
  lists_members_tags:
    primary key: id
    fields: date_added(string), id(integer), list_id(string), name(string), subscriber_hash(string)
  lists_merge_fields:
    primary key: list_id
    fields: _links(object), default_value(string), display_order(integer), help_text(string), list_id(string), merge_field_limit(integer), merge_id(integer), name(string), options(object), public(boolean), required(boolean), tag(string), total_items(integer), type(string)
  lists_segments:
    primary key: id
    cursor: updated_at
    fields: _links(object), created_at(string), id(integer), list_id(string), member_count(integer), name(string), options(object), type(string), updated_at(string)
  lists_segments_members:
    primary key: id
    cursor: last_changed
    fields: _links(object), email_address(string), email_client(string), email_type(string), full_name(string), id(string), interests(object), ip_opt(string), ip_signup(string), language(string), last_changed(string), last_note(object), list_id(string), location(object), member_rating(integer), merge_fields(object), segment_id(string), stats(object), status(string), timestamp_opt(string), timestamp_signup(string), unique_email_id(string), vip(boolean)
  lists_signup_forms:
    primary key: list_id
    fields: _links(object), contents(array), header(object), list_id(string), signup_form_url(string), styles(array)
  lists_surveys:
    primary key: id
    cursor: updated_at
    fields: _links(object), created_at(string), hosted_url(string), id(string), is_piped_to_inbox(boolean), list_id(string), published_at(string), question_count(integer), questions(array), response_count(integer), sections(array), status(string), title(string), updated_at(string), web_id(string)
  lists_webhooks:
    primary key: id
    fields: _links(object), events(object), id(string), list_id(string), sources(object), url(string)
  reporting_facebook_ads:
    primary key: id
    fields: id(string)
  reporting_facebook_ads_ecommerce_product_activity_products:
    primary key: outreach_id
    fields: currency_code(string), image_url(string), outreach_id(string), recommendation_purchased(integer), recommendation_total(integer), sku(string), title(string), total_purchased(number), total_revenue(number)
  reporting_landing_pages:
    primary key: id
    fields: _links(object), clicks(integer), conversion_rate(number), ecommerce(object), id(string), list_id(string), list_name(string), name(string), published_at(string), signup_tags(array), status(string), subscribes(integer), timeseries(object), title(string), unique_visits(integer), unpublished_at(string), url(string), visits(integer), web_id(integer)
  reporting_surveys:
    primary key: id
    cursor: updated_at
    fields: created_at(string), id(string), list_id(string), list_name(string), published_at(string), status(string), title(string), total_responses(integer), updated_at(string), url(string), web_id(integer)
  reporting_surveys_questions:
    primary key: id
    fields: average_rating(number), contact_counts(object), has_other(boolean), id(string), is_required(boolean), merge_field(object), options(array), other_label(string), placeholder_label(string), query(string), range_high_label(string), range_low_label(string), subscribe_checkbox_enabled(boolean), subscribe_checkbox_label(string), survey_id(string), total_responses(integer), type(string)
  reporting_surveys_questions_answers:
    primary key: id
    fields: contact(object), id(string), is_new_contact(boolean), question_id(string), response_id(string), submitted_at(string), survey_id(string), value(string)
  reporting_surveys_responses:
    primary key: survey_id
    fields: contact(object), is_new_contact(boolean), response_id(string), submitted_at(string), survey_id(string)
  reports:
    primary key: id
    cursor: send_time
    fields: _links(object), ab_split(object), abuse_reports(integer), bounces(object), campaign_title(string), clicks(object), delivery_status(object), ecommerce(object), emails_sent(integer), facebook_likes(object), forwards(object), id(string), industry_stats(object), list_id(string), list_is_active(boolean), list_name(string), list_stats(object), opens(object), preview_text(string), rss_last_send(string), send_time(string), share_report(object), subject_line(string), timeseries(array), timewarp(array), type(string), unsubscribed(integer)
  reports_abuse_reports:
    primary key: id
    fields: _links(object), campaign_id(string), date(string), email_address(string), email_id(string), id(integer), list_id(string), list_is_active(boolean), merge_fields(object), vip(boolean)
  reports_advice:
    primary key: campaign_id
    fields: _links(object), campaign_id(string), message(string), type(string)
  reports_click_details_urls_clicked:
    primary key: id
    fields: _links(object), ab_split(object), campaign_id(string), click_percentage(number), id(string), last_click(string), total_clicks(integer), unique_click_percentage(number), unique_clicks(integer), url(string)
  reports_click_details_members:
    primary key: email_id
    fields: _links(object), campaign_id(string), clicks(integer), contact_status(string), email_address(string), email_id(string), link_id(string), list_id(string), list_is_active(boolean), merge_fields(object), url_id(string), vip(boolean)
  reports_domain_performance_domains:
    primary key: campaign_id
    fields: bounces(integer), bounces_pct(number), campaign_id(string), clicks(integer), clicks_pct(number), delivered(integer), domain(string), emails_pct(number), emails_sent(integer), opens(integer), opens_pct(number), unsubs(integer), unsubs_pct(number)
  reports_ecommerce_product_activity_products:
    primary key: campaign_id
    fields: campaign_id(string), currency_code(string), image_url(string), recommendation_purchased(integer), recommendation_total(integer), sku(string), title(string), total_purchased(number), total_revenue(number)
  reports_eepurl_referrers:
    primary key: campaign_id
    fields: campaign_id(string), clicks(integer), first_click(string), last_click(string), referrer(string)
  reports_email_activity_emails:
    primary key: email_id
    fields: _links(object), activity(array), campaign_id(string), email_address(string), email_id(string), list_id(string), list_is_active(boolean)
  reports_locations:
    primary key: campaign_id
    fields: campaign_id(string), country_code(string), opens(integer), proxy_excluded_opens(integer), region(string), region_name(string)
  reports_open_details_members:
    primary key: email_id
    fields: _links(object), campaign_id(string), contact_status(string), email_address(string), email_id(string), list_id(string), list_is_active(boolean), merge_fields(object), opens(array), opens_count(integer), proxy_excluded_opens_count(integer), vip(boolean)
  reports_sent_to:
    primary key: email_id
    fields: _links(object), absplit_group(string), campaign_id(string), email_address(string), email_id(string), gmt_offset(integer), last_open(string), list_id(string), list_is_active(boolean), merge_fields(object), open_count(integer), status(string), vip(boolean)
  reports_sub_reports:
    primary key: id
    cursor: send_time
    fields: _links(object), ab_split(object), abuse_reports(integer), bounces(object), campaign_id(string), campaign_title(string), clicks(object), delivery_status(object), ecommerce(object), emails_sent(integer), facebook_likes(object), forwards(object), id(string), industry_stats(object), list_id(string), list_is_active(boolean), list_name(string), list_stats(object), opens(object), preview_text(string), rss_last_send(string), send_time(string), share_report(object), subject_line(string), timeseries(array), timewarp(array), type(string), unsubscribed(integer)
  reports_unsubscribed_unsubscribes:
    primary key: email_id
    fields: _links(object), campaign_id(string), email_address(string), email_id(string), list_id(string), list_is_active(boolean), merge_fields(object), reason(string), timestamp(string), vip(boolean)
  sms_campaigns:
    primary key: id
    cursor: create_time
    fields: _links(object), channel(string), create_time(string), excluded_segments(array), expire_time(string), folder_id(string), id(string), is_send_now(boolean), list_id(integer), name(string), recipient_count(integer), segments(array), send_time(string), status(string), updated_at(string), web_id(string)
  template_folders_folders:
    primary key: id
    fields: _links(object), count(integer), id(string), name(string)
  templates:
    primary key: id
    cursor: date_created
    fields: _links(object), active(boolean), category(string), content_type(string), created_by(string), date_created(string), date_edited(string), drag_and_drop(boolean), edited_by(string), folder_id(string), id(integer), name(string), responsive(boolean), share_url(string), thumbnail(string), type(string)
  verified_domains_domains:
    primary key: domain
    fields: authenticated(boolean), domain(string), is_free_email_provider(boolean), status(string), verification_email(string), verification_sent(string), verified(boolean)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  post_account_exports:
    endpoint: POST /account-exports
    optional fields: include_stages, since_timestamp
    risk: Externally visible Mailchimp mutation: Add export. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_audiences_contacts:
    endpoint: POST /audiences/{{ record.audience_id }}/contacts
    required fields: audience_id
    optional fields: language, email_channel, sms_channel, merge_fields, tags, update_existing
    risk: Externally visible Mailchimp mutation: Add Contact. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_audiences_contacts:
    endpoint: PATCH /audiences/{{ record.audience_id }}/contacts/{{ record.contact_id }}
    required fields: audience_id, contact_id
    optional fields: language, email_channel, sms_channel, merge_fields, tags
    risk: Externally visible Mailchimp mutation: Update Contact. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  archive_contacts:
    endpoint: POST /audiences/{{ record.audience_id }}/contacts/{{ record.contact_id }}/actions/archive
    required fields: audience_id, contact_id
    risk: Destructive or externally visible Mailchimp mutation: Archive Contact. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  forget_contacts:
    endpoint: POST /audiences/{{ record.audience_id }}/contacts/{{ record.contact_id }}/actions/forget
    required fields: audience_id, contact_id
    risk: Destructive or externally visible Mailchimp mutation: Forget Contact. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_automations:
    endpoint: POST /automations
    optional fields: recipients, settings, trigger_settings
    risk: Externally visible Mailchimp mutation: Add automation. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  archive_automations:
    endpoint: POST /automations/{{ record.workflow_id }}/actions/archive
    required fields: workflow_id
    risk: Destructive or externally visible Mailchimp mutation: Archive automation. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  pause_all_emails_automations:
    endpoint: POST /automations/{{ record.workflow_id }}/actions/pause-all-emails
    required fields: workflow_id
    risk: Destructive or externally visible Mailchimp mutation: Pause automation emails. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  start_all_emails_automations:
    endpoint: POST /automations/{{ record.workflow_id }}/actions/start-all-emails
    required fields: workflow_id
    risk: Destructive or externally visible Mailchimp mutation: Start automation emails. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_automations_emails:
    endpoint: DELETE /automations/{{ record.workflow_id }}/emails/{{ record.workflow_email_id }}
    required fields: workflow_id, workflow_email_id
    risk: Destructive or externally visible Mailchimp mutation: Delete workflow email. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_automations_emails:
    endpoint: PATCH /automations/{{ record.workflow_id }}/emails/{{ record.workflow_email_id }}
    required fields: workflow_id, workflow_email_id
    optional fields: settings, delay
    risk: Externally visible Mailchimp mutation: Update workflow email. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  pause_emails:
    endpoint: POST /automations/{{ record.workflow_id }}/emails/{{ record.workflow_email_id }}/actions/pause
    required fields: workflow_id, workflow_email_id
    risk: Destructive or externally visible Mailchimp mutation: Pause automated email. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  start_emails:
    endpoint: POST /automations/{{ record.workflow_id }}/emails/{{ record.workflow_email_id }}/actions/start
    required fields: workflow_id, workflow_email_id
    risk: Destructive or externally visible Mailchimp mutation: Start automated email. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_automations_emails_queue:
    endpoint: POST /automations/{{ record.workflow_id }}/emails/{{ record.workflow_email_id }}/queue
    required fields: workflow_id, workflow_email_id
    optional fields: email_address
    risk: Externally visible Mailchimp mutation: Add subscriber to workflow email. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_automations_removed_subscribers:
    endpoint: POST /automations/{{ record.workflow_id }}/removed-subscribers
    required fields: workflow_id
    optional fields: email_address
    risk: Externally visible Mailchimp mutation: Remove subscriber from workflow. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_batch_webhooks:
    endpoint: POST /batch-webhooks
    optional fields: url, enabled
    risk: Externally visible Mailchimp mutation: Add batch webhook. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_batch_webhooks:
    endpoint: DELETE /batch-webhooks/{{ record.batch_webhook_id }}
    required fields: batch_webhook_id
    risk: Destructive or externally visible Mailchimp mutation: Delete batch webhook. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_batch_webhooks:
    endpoint: PATCH /batch-webhooks/{{ record.batch_webhook_id }}
    required fields: batch_webhook_id
    optional fields: url, enabled
    risk: Externally visible Mailchimp mutation: Update batch webhook. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_batches:
    endpoint: DELETE /batches/{{ record.batch_id }}
    required fields: batch_id
    risk: Destructive or externally visible Mailchimp mutation: Delete batch request. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_campaign_folders:
    endpoint: POST /campaign-folders
    optional fields: name
    risk: Externally visible Mailchimp mutation: Add campaign folder. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_campaign_folders:
    endpoint: DELETE /campaign-folders/{{ record.folder_id }}
    required fields: folder_id
    risk: Destructive or externally visible Mailchimp mutation: Delete campaign folder. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_campaign_folders:
    endpoint: PATCH /campaign-folders/{{ record.folder_id }}
    required fields: folder_id
    optional fields: name
    risk: Externally visible Mailchimp mutation: Update campaign folder. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_campaigns:
    endpoint: POST /campaigns
    optional fields: type, recipients, settings, variate_settings, tracking, rss_opts, social_card, content_type
    risk: Externally visible Mailchimp mutation: Add campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_campaigns:
    endpoint: DELETE /campaigns/{{ record.campaign_id }}
    required fields: campaign_id
    risk: Destructive or externally visible Mailchimp mutation: Delete campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_campaigns:
    endpoint: PATCH /campaigns/{{ record.campaign_id }}
    required fields: campaign_id
    optional fields: recipients, settings, variate_settings, tracking, rss_opts, social_card
    risk: Externally visible Mailchimp mutation: Update campaign settings. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  cancel_send_campaigns:
    endpoint: POST /campaigns/{{ record.campaign_id }}/actions/cancel-send
    required fields: campaign_id
    risk: Destructive or externally visible Mailchimp mutation: Cancel campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  create_resend_campaigns:
    endpoint: POST /campaigns/{{ record.campaign_id }}/actions/create-resend
    required fields: campaign_id
    risk: Destructive or externally visible Mailchimp mutation: Resend campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  pause_campaigns:
    endpoint: POST /campaigns/{{ record.campaign_id }}/actions/pause
    required fields: campaign_id
    risk: Destructive or externally visible Mailchimp mutation: Pause rss campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  replicate_campaigns:
    endpoint: POST /campaigns/{{ record.campaign_id }}/actions/replicate
    required fields: campaign_id
    risk: Externally visible Mailchimp mutation: Replicate campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  resume_campaigns:
    endpoint: POST /campaigns/{{ record.campaign_id }}/actions/resume
    required fields: campaign_id
    risk: Externally visible Mailchimp mutation: Resume rss campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  schedule_campaigns:
    endpoint: POST /campaigns/{{ record.campaign_id }}/actions/schedule
    required fields: campaign_id
    risk: Externally visible Mailchimp mutation: Schedule campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  send_campaigns:
    endpoint: POST /campaigns/{{ record.campaign_id }}/actions/send
    required fields: campaign_id
    risk: Destructive or externally visible Mailchimp mutation: Send campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  test_campaigns:
    endpoint: POST /campaigns/{{ record.campaign_id }}/actions/test
    required fields: campaign_id
    risk: Destructive or externally visible Mailchimp mutation: Send test email. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  unschedule_campaigns:
    endpoint: POST /campaigns/{{ record.campaign_id }}/actions/unschedule
    required fields: campaign_id
    risk: Externally visible Mailchimp mutation: Unschedule campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  put_campaigns_content:
    endpoint: PUT /campaigns/{{ record.campaign_id }}/content
    required fields: campaign_id
    optional fields: plain_text, html, url, template, archive, variate_contents
    risk: Externally visible Mailchimp mutation: Set campaign content. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_campaigns_feedback:
    endpoint: POST /campaigns/{{ record.campaign_id }}/feedback
    required fields: campaign_id
    optional fields: block_id, message, is_complete
    risk: Externally visible Mailchimp mutation: Add campaign feedback. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_campaigns_feedback:
    endpoint: DELETE /campaigns/{{ record.campaign_id }}/feedback/{{ record.feedback_id }}
    required fields: campaign_id, feedback_id
    risk: Destructive or externally visible Mailchimp mutation: Delete campaign feedback message. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_campaigns_feedback:
    endpoint: PATCH /campaigns/{{ record.campaign_id }}/feedback/{{ record.feedback_id }}
    required fields: campaign_id, feedback_id
    optional fields: block_id, message, is_complete
    risk: Externally visible Mailchimp mutation: Update campaign feedback message. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_connected_sites:
    endpoint: POST /connected-sites
    optional fields: foreign_id, domain
    risk: Externally visible Mailchimp mutation: Add connected site. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_connected_sites:
    endpoint: DELETE /connected-sites/{{ record.connected_site_id }}
    required fields: connected_site_id
    risk: Destructive or externally visible Mailchimp mutation: Delete connected site. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  disable_pixel_connected_sites:
    endpoint: POST /connected-sites/{{ record.connected_site_id }}/actions/disable-pixel
    required fields: connected_site_id
    risk: Externally visible Mailchimp mutation: Disable pixel for connected site. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  enable_pixel_connected_sites:
    endpoint: POST /connected-sites/{{ record.connected_site_id }}/actions/enable-pixel
    required fields: connected_site_id
    risk: Externally visible Mailchimp mutation: Enable pixel for connected site. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  verify_script_installation_connected_sites:
    endpoint: POST /connected-sites/{{ record.connected_site_id }}/actions/verify-script-installation
    required fields: connected_site_id
    risk: Destructive or externally visible Mailchimp mutation: Verify connected site script. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  trigger_steps:
    endpoint: POST /customer-journeys/journeys/{{ record.journey_id }}/steps/{{ record.step_id }}/actions/trigger
    required fields: journey_id, step_id
    risk: Destructive or externally visible Mailchimp mutation: Customer Journeys API trigger for a contact. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_ecommerce_stores:
    endpoint: POST /ecommerce/stores
    optional fields: id, list_id, name, platform, domain, is_syncing, email_address, currency_code, money_format, primary_locale, timezone, phone, address
    risk: Externally visible Mailchimp mutation: Add store. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_ecommerce_stores:
    endpoint: DELETE /ecommerce/stores/{{ record.store_id }}
    required fields: store_id
    risk: Destructive or externally visible Mailchimp mutation: Delete store. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_ecommerce_stores:
    endpoint: PATCH /ecommerce/stores/{{ record.store_id }}
    required fields: store_id
    optional fields: name, platform, domain, is_syncing, email_address, currency_code, money_format, primary_locale, timezone, phone, address
    risk: Externally visible Mailchimp mutation: Update store. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_ecommerce_stores_carts:
    endpoint: POST /ecommerce/stores/{{ record.store_id }}/carts
    required fields: store_id
    optional fields: id, customer, campaign_id, checkout_url, currency_code, order_total, tax_total, lines
    risk: Externally visible Mailchimp mutation: Add cart. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_ecommerce_stores_carts:
    endpoint: DELETE /ecommerce/stores/{{ record.store_id }}/carts/{{ record.cart_id }}
    required fields: store_id, cart_id
    risk: Destructive or externally visible Mailchimp mutation: Delete cart. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_ecommerce_stores_carts:
    endpoint: PATCH /ecommerce/stores/{{ record.store_id }}/carts/{{ record.cart_id }}
    required fields: store_id, cart_id
    optional fields: customer, campaign_id, checkout_url, currency_code, order_total, tax_total, lines
    risk: Externally visible Mailchimp mutation: Update cart. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_ecommerce_stores_carts_lines:
    endpoint: POST /ecommerce/stores/{{ record.store_id }}/carts/{{ record.cart_id }}/lines
    required fields: store_id, cart_id
    optional fields: id, product_id, product_variant_id, quantity, price
    risk: Externally visible Mailchimp mutation: Add cart line item. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_ecommerce_stores_carts_lines:
    endpoint: DELETE /ecommerce/stores/{{ record.store_id }}/carts/{{ record.cart_id }}/lines/{{ record.line_id }}
    required fields: store_id, cart_id, line_id
    risk: Destructive or externally visible Mailchimp mutation: Delete cart line item. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_ecommerce_stores_carts_lines:
    endpoint: PATCH /ecommerce/stores/{{ record.store_id }}/carts/{{ record.cart_id }}/lines/{{ record.line_id }}
    required fields: store_id, cart_id, line_id
    optional fields: product_id, product_variant_id, quantity, price
    risk: Externally visible Mailchimp mutation: Update cart line item. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_ecommerce_stores_customers:
    endpoint: POST /ecommerce/stores/{{ record.store_id }}/customers
    required fields: store_id
    optional fields: id, email_address, sms_phone_number, opt_in_status, company, first_name, last_name, address
    risk: Externally visible Mailchimp mutation: Add customer. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_ecommerce_stores_customers:
    endpoint: DELETE /ecommerce/stores/{{ record.store_id }}/customers/{{ record.customer_id }}
    required fields: store_id, customer_id
    risk: Destructive or externally visible Mailchimp mutation: Delete customer. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_ecommerce_stores_customers:
    endpoint: PATCH /ecommerce/stores/{{ record.store_id }}/customers/{{ record.customer_id }}
    required fields: store_id, customer_id
    optional fields: opt_in_status, company, first_name, last_name, address
    risk: Externally visible Mailchimp mutation: Update customer. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  put_ecommerce_stores_customers:
    endpoint: PUT /ecommerce/stores/{{ record.store_id }}/customers/{{ record.customer_id }}
    required fields: store_id, customer_id
    optional fields: id, email_address, sms_phone_number, opt_in_status, company, first_name, last_name, address
    risk: Externally visible Mailchimp mutation: Add or update customer. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_ecommerce_stores_orders:
    endpoint: POST /ecommerce/stores/{{ record.store_id }}/orders
    required fields: store_id
    optional fields: id, customer, campaign_id, cart_id, landing_site, financial_status, fulfillment_status, currency_code, order_total, order_url, discount_total, tax_total, shipping_total, tracking_code, processed_at_foreign, cancelled_at_foreign, updated_at_foreign, shipping_address, billing_address, promos, lines, outreach, tracking_number, tracking_carrier, tracking_url
    risk: Externally visible Mailchimp mutation: Add order. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_ecommerce_stores_orders:
    endpoint: DELETE /ecommerce/stores/{{ record.store_id }}/orders/{{ record.order_id }}
    required fields: store_id, order_id
    risk: Destructive or externally visible Mailchimp mutation: Delete order. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_ecommerce_stores_orders:
    endpoint: PATCH /ecommerce/stores/{{ record.store_id }}/orders/{{ record.order_id }}
    required fields: store_id, order_id
    optional fields: customer, campaign_id, cart_id, landing_site, financial_status, fulfillment_status, currency_code, order_total, order_url, discount_total, tax_total, shipping_total, tracking_code, processed_at_foreign, cancelled_at_foreign, updated_at_foreign, shipping_address, billing_address, promos, lines, outreach, tracking_number, tracking_carrier, tracking_url
    risk: Externally visible Mailchimp mutation: Update order. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  put_ecommerce_stores_orders:
    endpoint: PUT /ecommerce/stores/{{ record.store_id }}/orders/{{ record.order_id }}
    required fields: store_id, order_id
    optional fields: id, customer, campaign_id, cart_id, landing_site, financial_status, fulfillment_status, currency_code, order_total, order_url, discount_total, tax_total, shipping_total, tracking_code, processed_at_foreign, cancelled_at_foreign, updated_at_foreign, shipping_address, billing_address, promos, lines, outreach, tracking_number, tracking_carrier, tracking_url
    risk: Externally visible Mailchimp mutation: Add or update order. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_ecommerce_stores_orders_lines:
    endpoint: POST /ecommerce/stores/{{ record.store_id }}/orders/{{ record.order_id }}/lines
    required fields: store_id, order_id
    optional fields: id, product_id, product_variant_id, quantity, price, discount
    risk: Externally visible Mailchimp mutation: Add order line item. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_ecommerce_stores_orders_lines:
    endpoint: DELETE /ecommerce/stores/{{ record.store_id }}/orders/{{ record.order_id }}/lines/{{ record.line_id }}
    required fields: store_id, order_id, line_id
    risk: Destructive or externally visible Mailchimp mutation: Delete order line item. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_ecommerce_stores_orders_lines:
    endpoint: PATCH /ecommerce/stores/{{ record.store_id }}/orders/{{ record.order_id }}/lines/{{ record.line_id }}
    required fields: store_id, order_id, line_id
    optional fields: product_id, product_variant_id, quantity, price, discount
    risk: Externally visible Mailchimp mutation: Update order line item. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_ecommerce_stores_products:
    endpoint: POST /ecommerce/stores/{{ record.store_id }}/products
    required fields: store_id
    optional fields: id, title, handle, url, description, type, vendor, image_url, variants, images, published_at_foreign
    risk: Externally visible Mailchimp mutation: Add product. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_ecommerce_stores_products:
    endpoint: DELETE /ecommerce/stores/{{ record.store_id }}/products/{{ record.product_id }}
    required fields: store_id, product_id
    risk: Destructive or externally visible Mailchimp mutation: Delete product. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_ecommerce_stores_products:
    endpoint: PATCH /ecommerce/stores/{{ record.store_id }}/products/{{ record.product_id }}
    required fields: store_id, product_id
    optional fields: title, handle, url, description, type, vendor, image_url, variants, images, published_at_foreign
    risk: Externally visible Mailchimp mutation: Update product. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  put_ecommerce_stores_products:
    endpoint: PUT /ecommerce/stores/{{ record.store_id }}/products/{{ record.product_id }}
    required fields: store_id, product_id
    optional fields: id, title, handle, url, description, type, vendor, image_url, variants, images, published_at_foreign
    risk: Externally visible Mailchimp mutation: Create or update product. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_ecommerce_stores_products_images:
    endpoint: POST /ecommerce/stores/{{ record.store_id }}/products/{{ record.product_id }}/images
    required fields: store_id, product_id
    optional fields: id, url, variant_ids
    risk: Externally visible Mailchimp mutation: Add product image. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_ecommerce_stores_products_images:
    endpoint: DELETE /ecommerce/stores/{{ record.store_id }}/products/{{ record.product_id }}/images/{{ record.image_id }}
    required fields: store_id, product_id, image_id
    risk: Destructive or externally visible Mailchimp mutation: Delete product image. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_ecommerce_stores_products_images:
    endpoint: PATCH /ecommerce/stores/{{ record.store_id }}/products/{{ record.product_id }}/images/{{ record.image_id }}
    required fields: store_id, product_id, image_id
    optional fields: id, url, variant_ids
    risk: Externally visible Mailchimp mutation: Update product image. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_ecommerce_stores_products_variants:
    endpoint: POST /ecommerce/stores/{{ record.store_id }}/products/{{ record.product_id }}/variants
    required fields: store_id, product_id
    optional fields: id, title, url, sku, price, inventory_quantity, image_url, backorders, visibility
    risk: Externally visible Mailchimp mutation: Add product variant. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_ecommerce_stores_products_variants:
    endpoint: DELETE /ecommerce/stores/{{ record.store_id }}/products/{{ record.product_id }}/variants/{{ record.variant_id }}
    required fields: store_id, product_id, variant_id
    risk: Destructive or externally visible Mailchimp mutation: Delete product variant. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_ecommerce_stores_products_variants:
    endpoint: PATCH /ecommerce/stores/{{ record.store_id }}/products/{{ record.product_id }}/variants/{{ record.variant_id }}
    required fields: store_id, product_id, variant_id
    optional fields: title, url, sku, price, inventory_quantity, image_url, backorders, visibility
    risk: Externally visible Mailchimp mutation: Update product variant. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  put_ecommerce_stores_products_variants:
    endpoint: PUT /ecommerce/stores/{{ record.store_id }}/products/{{ record.product_id }}/variants/{{ record.variant_id }}
    required fields: store_id, product_id, variant_id
    optional fields: id, title, url, sku, price, inventory_quantity, image_url, backorders, visibility
    risk: Externally visible Mailchimp mutation: Add or update product variant. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_ecommerce_stores_promo_rules:
    endpoint: POST /ecommerce/stores/{{ record.store_id }}/promo-rules
    required fields: store_id
    optional fields: id, title, description, starts_at, ends_at, amount, type, target, enabled, created_at_foreign, updated_at_foreign
    risk: Externally visible Mailchimp mutation: Add promo rule. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_ecommerce_stores_promo_rules:
    endpoint: DELETE /ecommerce/stores/{{ record.store_id }}/promo-rules/{{ record.promo_rule_id }}
    required fields: store_id, promo_rule_id
    risk: Destructive or externally visible Mailchimp mutation: Delete promo rule. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_ecommerce_stores_promo_rules:
    endpoint: PATCH /ecommerce/stores/{{ record.store_id }}/promo-rules/{{ record.promo_rule_id }}
    required fields: store_id, promo_rule_id
    optional fields: title, description, starts_at, ends_at, amount, type, target, enabled, created_at_foreign, updated_at_foreign
    risk: Externally visible Mailchimp mutation: Update promo rule. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_ecommerce_stores_promo_rules_promo_codes:
    endpoint: POST /ecommerce/stores/{{ record.store_id }}/promo-rules/{{ record.promo_rule_id }}/promo-codes
    required fields: store_id, promo_rule_id
    optional fields: id, code, redemption_url, usage_count, enabled, created_at_foreign, updated_at_foreign
    risk: Externally visible Mailchimp mutation: Add promo code. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_ecommerce_stores_promo_rules_promo_codes:
    endpoint: DELETE /ecommerce/stores/{{ record.store_id }}/promo-rules/{{ record.promo_rule_id }}/promo-codes/{{ record.promo_code_id }}
    required fields: store_id, promo_rule_id, promo_code_id
    risk: Destructive or externally visible Mailchimp mutation: Delete promo code. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_ecommerce_stores_promo_rules_promo_codes:
    endpoint: PATCH /ecommerce/stores/{{ record.store_id }}/promo-rules/{{ record.promo_rule_id }}/promo-codes/{{ record.promo_code_id }}
    required fields: store_id, promo_rule_id, promo_code_id
    optional fields: code, redemption_url, usage_count, enabled, created_at_foreign, updated_at_foreign
    risk: Externally visible Mailchimp mutation: Update promo code. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_file_manager_files:
    endpoint: POST /file-manager/files
    optional fields: folder_id, name, file_data
    risk: Externally visible Mailchimp mutation: Add file. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_file_manager_files:
    endpoint: DELETE /file-manager/files/{{ record.file_id }}
    required fields: file_id
    risk: Destructive or externally visible Mailchimp mutation: Delete file. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_file_manager_files:
    endpoint: PATCH /file-manager/files/{{ record.file_id }}
    required fields: file_id
    optional fields: folder_id, name
    risk: Externally visible Mailchimp mutation: Update file. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_file_manager_folders:
    endpoint: POST /file-manager/folders
    optional fields: name
    risk: Externally visible Mailchimp mutation: Add folder. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_file_manager_folders:
    endpoint: DELETE /file-manager/folders/{{ record.folder_id }}
    required fields: folder_id
    risk: Destructive or externally visible Mailchimp mutation: Delete folder. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_file_manager_folders:
    endpoint: PATCH /file-manager/folders/{{ record.folder_id }}
    required fields: folder_id
    optional fields: name
    risk: Externally visible Mailchimp mutation: Update folder. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_landing_pages:
    endpoint: POST /landing-pages
    optional fields: name, title, description, store_id, list_id, type, template_id, tracking
    risk: Externally visible Mailchimp mutation: Add landing page. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_landing_pages:
    endpoint: DELETE /landing-pages/{{ record.page_id }}
    required fields: page_id
    risk: Destructive or externally visible Mailchimp mutation: Delete landing page. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_landing_pages:
    endpoint: PATCH /landing-pages/{{ record.page_id }}
    required fields: page_id
    optional fields: name, title, description, store_id, list_id, tracking
    risk: Externally visible Mailchimp mutation: Update landing page. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  publish_landing_pages:
    endpoint: POST /landing-pages/{{ record.page_id }}/actions/publish
    required fields: page_id
    risk: Destructive or externally visible Mailchimp mutation: Publish landing page. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  unpublish_landing_pages:
    endpoint: POST /landing-pages/{{ record.page_id }}/actions/unpublish
    required fields: page_id
    risk: Destructive or externally visible Mailchimp mutation: Unpublish landing page. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_lists:
    endpoint: POST /lists
    optional fields: name, contact, permission_reminder, use_archive_bar, campaign_defaults, notify_on_subscribe, notify_on_unsubscribe, email_type_option, double_optin, marketing_permissions
    risk: Externally visible Mailchimp mutation: Add list. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_lists:
    endpoint: DELETE /lists/{{ record.list_id }}
    required fields: list_id
    risk: Destructive or externally visible Mailchimp mutation: Delete list. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_lists:
    endpoint: PATCH /lists/{{ record.list_id }}
    required fields: list_id
    optional fields: name, contact, permission_reminder, use_archive_bar, campaign_defaults, notify_on_subscribe, notify_on_unsubscribe, email_type_option, double_optin, marketing_permissions
    risk: Externally visible Mailchimp mutation: Update lists. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_lists_2:
    endpoint: POST /lists/{{ record.list_id }}
    required fields: list_id
    optional fields: members, sync_tags, update_existing
    risk: Externally visible Mailchimp mutation: Batch subscribe or unsubscribe. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_lists_interest_categories:
    endpoint: POST /lists/{{ record.list_id }}/interest-categories
    required fields: list_id
    optional fields: title, display_order, type
    risk: Externally visible Mailchimp mutation: Add interest category. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_lists_interest_categories:
    endpoint: DELETE /lists/{{ record.list_id }}/interest-categories/{{ record.interest_category_id }}
    required fields: list_id, interest_category_id
    risk: Destructive or externally visible Mailchimp mutation: Delete interest category. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_lists_interest_categories:
    endpoint: PATCH /lists/{{ record.list_id }}/interest-categories/{{ record.interest_category_id }}
    required fields: list_id, interest_category_id
    optional fields: title, display_order, type
    risk: Externally visible Mailchimp mutation: Update interest category. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_lists_interest_categories_interests:
    endpoint: POST /lists/{{ record.list_id }}/interest-categories/{{ record.interest_category_id }}/interests
    required fields: list_id, interest_category_id
    optional fields: name, display_order
    risk: Externally visible Mailchimp mutation: Add interest in category. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_lists_interest_categories_interests:
    endpoint: DELETE /lists/{{ record.list_id }}/interest-categories/{{ record.interest_category_id }}/interests/{{ record.interest_id }}
    required fields: list_id, interest_category_id, interest_id
    risk: Destructive or externally visible Mailchimp mutation: Delete interest in category. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_lists_interest_categories_interests:
    endpoint: PATCH /lists/{{ record.list_id }}/interest-categories/{{ record.interest_category_id }}/interests/{{ record.interest_id }}
    required fields: list_id, interest_category_id, interest_id
    optional fields: name, display_order
    risk: Externally visible Mailchimp mutation: Update interest in category. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_lists_members:
    endpoint: POST /lists/{{ record.list_id }}/members
    required fields: list_id
    optional fields: email_address, email_type, status, merge_fields, interests, language, vip, location, marketing_permissions, ip_signup, timestamp_signup, ip_opt, timestamp_opt, tags
    risk: Externally visible Mailchimp mutation: Add member to list. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_lists_members:
    endpoint: DELETE /lists/{{ record.list_id }}/members/{{ record.subscriber_hash }}
    required fields: list_id, subscriber_hash
    risk: Destructive or externally visible Mailchimp mutation: Archive list member. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_lists_members:
    endpoint: PATCH /lists/{{ record.list_id }}/members/{{ record.subscriber_hash }}
    required fields: list_id, subscriber_hash
    optional fields: email_address, email_type, status, merge_fields, interests, language, vip, location, marketing_permissions, ip_signup, timestamp_signup, ip_opt, timestamp_opt
    risk: Externally visible Mailchimp mutation: Update list member. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  put_lists_members:
    endpoint: PUT /lists/{{ record.list_id }}/members/{{ record.subscriber_hash }}
    required fields: list_id, subscriber_hash
    optional fields: email_address, status_if_new, email_type, status, merge_fields, interests, language, vip, location, marketing_permissions, ip_signup, timestamp_signup, ip_opt, timestamp_opt
    risk: Externally visible Mailchimp mutation: Add or update list member. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_permanent_members:
    endpoint: POST /lists/{{ record.list_id }}/members/{{ record.subscriber_hash }}/actions/delete-permanent
    required fields: list_id, subscriber_hash
    risk: Destructive or externally visible Mailchimp mutation: Delete list member. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_lists_members_events:
    endpoint: POST /lists/{{ record.list_id }}/members/{{ record.subscriber_hash }}/events
    required fields: list_id, subscriber_hash
    optional fields: name, properties, is_syncing, occurred_at
    risk: Externally visible Mailchimp mutation: Add event. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_lists_members_notes:
    endpoint: POST /lists/{{ record.list_id }}/members/{{ record.subscriber_hash }}/notes
    required fields: list_id, subscriber_hash
    optional fields: note
    risk: Externally visible Mailchimp mutation: Add member note. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_lists_members_notes:
    endpoint: DELETE /lists/{{ record.list_id }}/members/{{ record.subscriber_hash }}/notes/{{ record.note_id }}
    required fields: list_id, subscriber_hash, note_id
    risk: Destructive or externally visible Mailchimp mutation: Delete note. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_lists_members_notes:
    endpoint: PATCH /lists/{{ record.list_id }}/members/{{ record.subscriber_hash }}/notes/{{ record.note_id }}
    required fields: list_id, subscriber_hash, note_id
    optional fields: note
    risk: Externally visible Mailchimp mutation: Update note. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_lists_members_tags:
    endpoint: POST /lists/{{ record.list_id }}/members/{{ record.subscriber_hash }}/tags
    required fields: list_id, subscriber_hash
    optional fields: tags, is_syncing
    risk: Externally visible Mailchimp mutation: Add or remove member tags. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_lists_merge_fields:
    endpoint: POST /lists/{{ record.list_id }}/merge-fields
    required fields: list_id
    optional fields: tag, name, type, required, default_value, public, display_order, options, help_text
    risk: Externally visible Mailchimp mutation: Add merge field. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_lists_merge_fields:
    endpoint: DELETE /lists/{{ record.list_id }}/merge-fields/{{ record.merge_id }}
    required fields: list_id, merge_id
    risk: Destructive or externally visible Mailchimp mutation: Delete merge field. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_lists_merge_fields:
    endpoint: PATCH /lists/{{ record.list_id }}/merge-fields/{{ record.merge_id }}
    required fields: list_id, merge_id
    optional fields: tag, name, required, default_value, public, display_order, options, help_text
    risk: Externally visible Mailchimp mutation: Update merge field. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_lists_segments:
    endpoint: POST /lists/{{ record.list_id }}/segments
    required fields: list_id
    optional fields: name, static_segment, options
    risk: Externally visible Mailchimp mutation: Add segment. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_lists_segments:
    endpoint: DELETE /lists/{{ record.list_id }}/segments/{{ record.segment_id }}
    required fields: list_id, segment_id
    risk: Destructive or externally visible Mailchimp mutation: Delete segment. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_lists_segments:
    endpoint: PATCH /lists/{{ record.list_id }}/segments/{{ record.segment_id }}
    required fields: list_id, segment_id
    optional fields: name, static_segment, options
    risk: Externally visible Mailchimp mutation: Update segment. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_lists_segments_2:
    endpoint: POST /lists/{{ record.list_id }}/segments/{{ record.segment_id }}
    required fields: list_id, segment_id
    optional fields: members_to_add, members_to_remove
    risk: Externally visible Mailchimp mutation: Batch add or remove members. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_lists_segments_members:
    endpoint: POST /lists/{{ record.list_id }}/segments/{{ record.segment_id }}/members
    required fields: list_id, segment_id
    optional fields: email_address
    risk: Externally visible Mailchimp mutation: Add member to segment. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_lists_segments_members:
    endpoint: DELETE /lists/{{ record.list_id }}/segments/{{ record.segment_id }}/members/{{ record.subscriber_hash }}
    required fields: list_id, segment_id, subscriber_hash
    risk: Destructive or externally visible Mailchimp mutation: Remove list member from segment. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_lists_signup_forms:
    endpoint: POST /lists/{{ record.list_id }}/signup-forms
    required fields: list_id
    optional fields: header, contents, styles
    risk: Externally visible Mailchimp mutation: Customize signup form. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_lists_surveys:
    endpoint: POST /lists/{{ record.list_id }}/surveys
    required fields: list_id
    optional fields: title, sections
    risk: Externally visible Mailchimp mutation: Create survey. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_lists_surveys:
    endpoint: DELETE /lists/{{ record.list_id }}/surveys/{{ record.survey_id }}
    required fields: list_id, survey_id
    risk: Destructive or externally visible Mailchimp mutation: Delete survey. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_lists_surveys:
    endpoint: PATCH /lists/{{ record.list_id }}/surveys/{{ record.survey_id }}
    required fields: list_id, survey_id
    optional fields: title, is_piped_to_inbox, sections
    risk: Externally visible Mailchimp mutation: Update survey. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  create_email_surveys:
    endpoint: POST /lists/{{ record.list_id }}/surveys/{{ record.survey_id }}/actions/create-email
    required fields: list_id, survey_id
    risk: Externally visible Mailchimp mutation: Create a Survey Campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  publish_surveys:
    endpoint: POST /lists/{{ record.list_id }}/surveys/{{ record.survey_id }}/actions/publish
    required fields: list_id, survey_id
    risk: Destructive or externally visible Mailchimp mutation: Publish a Survey. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  replicate_surveys:
    endpoint: POST /lists/{{ record.list_id }}/surveys/{{ record.survey_id }}/actions/replicate
    required fields: list_id, survey_id
    risk: Externally visible Mailchimp mutation: Replicate survey. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  unpublish_surveys:
    endpoint: POST /lists/{{ record.list_id }}/surveys/{{ record.survey_id }}/actions/unpublish
    required fields: list_id, survey_id
    risk: Destructive or externally visible Mailchimp mutation: Unpublish a Survey. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_lists_webhooks:
    endpoint: POST /lists/{{ record.list_id }}/webhooks
    required fields: list_id
    optional fields: url, events, sources
    risk: Externally visible Mailchimp mutation: Add webhook. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_lists_webhooks:
    endpoint: DELETE /lists/{{ record.list_id }}/webhooks/{{ record.webhook_id }}
    required fields: list_id, webhook_id
    risk: Destructive or externally visible Mailchimp mutation: Delete webhook. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_lists_webhooks:
    endpoint: PATCH /lists/{{ record.list_id }}/webhooks/{{ record.webhook_id }}
    required fields: list_id, webhook_id
    optional fields: url, events, sources
    risk: Externally visible Mailchimp mutation: Update webhook. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_sms_campaigns:
    endpoint: POST /sms-campaigns
    optional fields: name, list_id, folder_id, segments, excluded_segments
    risk: Externally visible Mailchimp mutation: Add SMS campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_sms_campaigns:
    endpoint: DELETE /sms-campaigns/{{ record.sms_campaign_id }}
    required fields: sms_campaign_id
    risk: Destructive or externally visible Mailchimp mutation: Delete SMS campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_sms_campaigns:
    endpoint: PATCH /sms-campaigns/{{ record.sms_campaign_id }}
    required fields: sms_campaign_id
    optional fields: name, folder_id, segments, excluded_segments
    risk: Externally visible Mailchimp mutation: Update SMS campaign settings. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  cancel_send_sms_campaigns:
    endpoint: POST /sms-campaigns/{{ record.sms_campaign_id }}/actions/cancel-send
    required fields: sms_campaign_id
    risk: Destructive or externally visible Mailchimp mutation: Cancel SMS campaign send. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  schedule_sms_campaigns:
    endpoint: POST /sms-campaigns/{{ record.sms_campaign_id }}/actions/schedule
    required fields: sms_campaign_id
    risk: Externally visible Mailchimp mutation: Schedule SMS campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  send_sms_campaigns:
    endpoint: POST /sms-campaigns/{{ record.sms_campaign_id }}/actions/send
    required fields: sms_campaign_id
    risk: Destructive or externally visible Mailchimp mutation: Send SMS campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  put_sms_campaigns_content:
    endpoint: PUT /sms-campaigns/{{ record.sms_campaign_id }}/content
    required fields: sms_campaign_id
    optional fields: message_body, media
    risk: Externally visible Mailchimp mutation: Set SMS campaign content. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_template_folders:
    endpoint: POST /template-folders
    optional fields: name
    risk: Externally visible Mailchimp mutation: Add template folder. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_template_folders:
    endpoint: DELETE /template-folders/{{ record.folder_id }}
    required fields: folder_id
    risk: Destructive or externally visible Mailchimp mutation: Delete template folder. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_template_folders:
    endpoint: PATCH /template-folders/{{ record.folder_id }}
    required fields: folder_id
    optional fields: name
    risk: Externally visible Mailchimp mutation: Update template folder. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_templates:
    endpoint: POST /templates
    optional fields: name, folder_id, html
    risk: Externally visible Mailchimp mutation: Add template. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_templates:
    endpoint: DELETE /templates/{{ record.template_id }}
    required fields: template_id
    risk: Destructive or externally visible Mailchimp mutation: Delete template. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  patch_templates:
    endpoint: PATCH /templates/{{ record.template_id }}
    required fields: template_id
    optional fields: name, folder_id, html
    risk: Externally visible Mailchimp mutation: Update template. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  post_verified_domains:
    endpoint: POST /verified-domains
    optional fields: verification_email
    risk: Externally visible Mailchimp mutation: Add domain to account. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  delete_verified_domains:
    endpoint: DELETE /verified-domains/{{ record.domain_name }}
    required fields: domain_name
    risk: Destructive or externally visible Mailchimp mutation: Delete domain. Reverse ETL must plan, preview, receive explicit approval, and then execute.
  verify_verified_domains:
    endpoint: POST /verified-domains/{{ record.domain_name }}/actions/verify
    required fields: domain_name
    risk: Destructive or externally visible Mailchimp mutation: Verify domain. Reverse ETL must plan, preview, receive explicit approval, and then execute.

SECURITY
  read risk: external Mailchimp Marketing API reads using configured datacenter credentials; nested streams may require operation-specific identifier config
  write risk: approval-gated Mailchimp mutations against audiences, campaigns, automations, templates, files, webhooks, ecommerce, and related resources
  approval: reverse ETL writes require plan, preview, explicit approval, and destructive confirmation where declared
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Inspect, read, search, and safely plan typed Mailchimp Marketing API operations.
  Usage: pm mailchimp <command> [flags]
  Source CLI: Mailchimp Marketing API (Official Swagger 2.0 schema 3.0.91)
  Global flags:
    --credential (string): Credential name to use for the Mailchimp request.
    --connection (string): Alias for --credential.
    --config (string_array): Connector config override as key=value.
    --json (boolean): Emit machine-readable JSON output.
    --limit (integer): Maximum ETL records to emit.
    --max-bytes (integer): Maximum direct-read response bytes; operations are capped by their metadata.
    --plan (string): Execute an approved reverse-ETL plan by id.
    --preview (boolean): Preview a reverse-ETL write command without making a network mutation.
    --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
    --confirm (string): Typed confirmation challenge for destructive reverse-ETL writes.
  ETL streams
    lists list - Get lists info as ETL records. [intent=etl availability=implemented stream=lists]
    lists get - Get list info [intent=direct_read availability=implemented operation=mailchimp.get_lists_id]; flags: --list-id, --page, --page-cursor
    lists abuse-reports get - Get abuse report [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_abuse_reports_id]; flags: --list-id, --report-id, --page, --page-cursor
    lists growth-history get - Get growth history by month [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_growth_history_id]; flags: --list-id, --month, --page, --page-cursor
    lists interest-categories get - Get interest category info [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_interest_categories_id]; flags: --list-id, --interest-category-id, --page, --page-cursor
    lists interest-categories interests get - Get interest in category [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_interest_categories_id_interests_id]; flags: --list-id, --interest-category-id, --interest-id, --page, --page-cursor
    lists members get - Get member info [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_members_id]; flags: --list-id, --subscriber-hash, --page, --page-cursor
    lists members activity-feed get - View recent activity [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_members_id_activity_feed]; flags: --list-id, --subscriber-hash, --page, --page-cursor
    lists members notes get - Get member note [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_members_id_notes_id]; flags: --list-id, --subscriber-hash, --note-id, --page, --page-cursor
    lists merge-fields get - Get merge field [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_merge_fields_id]; flags: --list-id, --merge-id, --page, --page-cursor
    lists segments get - Get segment info [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_segments_id]; flags: --list-id, --segment-id, --page, --page-cursor
    lists surveys get - Get survey [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_surveys_id]; flags: --list-id, --survey-id, --page, --page-cursor
    lists tag-search get - Search for tags on a list by name. [intent=direct_read availability=implemented operation=mailchimp.search_tags_by_name]; flags: --list-id, --query, --page, --page-cursor
    lists webhooks get - Get webhook info [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_webhooks_id]; flags: --list-id, --webhook-id, --page, --page-cursor
    campaigns list - List campaigns as ETL records. [intent=etl availability=implemented stream=campaigns]
    campaigns get - Get campaign info [intent=direct_read availability=implemented operation=mailchimp.get_campaigns_id]; flags: --campaign-id, --page, --page-cursor
    campaigns content get - Get campaign content [intent=direct_read availability=implemented operation=mailchimp.get_campaigns_id_content]; flags: --campaign-id, --page, --page-cursor
    campaigns feedback get - Get campaign feedback message [intent=direct_read availability=implemented operation=mailchimp.get_campaigns_id_feedback_id]; flags: --campaign-id, --feedback-id, --page, --page-cursor
    campaigns send-checklist get - Get campaign send checklist [intent=direct_read availability=implemented operation=mailchimp.get_campaigns_id_send_checklist]; flags: --campaign-id, --page, --page-cursor
    reports list - List campaign reports as ETL records. [intent=etl availability=implemented stream=reports]
    reports get - Get campaign report [intent=direct_read availability=implemented operation=mailchimp.get_reports_id]; flags: --campaign-id, --page, --page-cursor
    reports abuse-reports get - Get abuse report [intent=direct_read availability=implemented operation=mailchimp.get_reports_id_abuse_reports_id_id]; flags: --campaign-id, --report-id, --page, --page-cursor
    reports click-details get - Get campaign link details [intent=direct_read availability=implemented operation=mailchimp.get_reports_id_click_details_id]; flags: --campaign-id, --link-id, --page, --page-cursor
    reports click-details members get - Get clicked link subscriber [intent=direct_read availability=implemented operation=mailchimp.get_reports_id_click_details_id_members_id]; flags: --campaign-id, --link-id, --subscriber-hash, --page, --page-cursor
    reports email-activity get - Get subscriber email activity [intent=direct_read availability=implemented operation=mailchimp.get_reports_id_email_activity_id]; flags: --campaign-id, --subscriber-hash, --page, --page-cursor
    reports open-details get - Get opened campaign subscriber [intent=direct_read availability=implemented operation=mailchimp.get_reports_id_open_details_id_members_id]; flags: --campaign-id, --subscriber-hash, --page, --page-cursor
    reports sent-to get - Get campaign recipient info [intent=direct_read availability=implemented operation=mailchimp.get_reports_id_sent_to_id]; flags: --campaign-id, --subscriber-hash, --page, --page-cursor
    reports unsubscribed get - Get unsubscribed member [intent=direct_read availability=implemented operation=mailchimp.get_reports_id_unsubscribed_id]; flags: --campaign-id, --subscriber-hash, --page, --page-cursor
    automations list - List automations as ETL records. [intent=etl availability=implemented stream=automations]
    automations get - Get automation info [intent=direct_read availability=implemented operation=mailchimp.get_automations_id]; flags: --workflow-id, --page, --page-cursor
    automations emails get - Get workflow email info [intent=direct_read availability=implemented operation=mailchimp.get_automations_id_emails_id]; flags: --workflow-id, --workflow-email-id, --page, --page-cursor
    automations emails queue get - Get automated email subscriber [intent=direct_read availability=implemented operation=mailchimp.get_automations_id_emails_id_queue_id]; flags: --workflow-id, --workflow-email-id, --subscriber-hash, --page, --page-cursor
    automations removed-subscribers get - Get subscriber removed from workflow [intent=direct_read availability=implemented operation=mailchimp.get_automations_id_removed_subscribers_id]; flags: --workflow-id, --subscriber-hash, --page, --page-cursor
    ecommerce stores get - Get store info [intent=direct_read availability=implemented operation=mailchimp.get_ecommerce_stores_id]; flags: --store-id, --page, --page-cursor
    ecommerce stores carts get - Get cart info [intent=direct_read availability=implemented operation=mailchimp.get_ecommerce_stores_id_carts_id]; flags: --store-id, --cart-id, --page, --page-cursor
    ecommerce stores carts lines get - Get cart line item [intent=direct_read availability=implemented operation=mailchimp.get_ecommerce_stores_id_carts_id_lines_id]; flags: --store-id, --cart-id, --line-id, --page, --page-cursor
    ecommerce stores customers get - Get customer info [intent=direct_read availability=implemented operation=mailchimp.get_ecommerce_stores_id_customers_id]; flags: --store-id, --customer-id, --page, --page-cursor
    ecommerce stores orders get - Get order info [intent=direct_read availability=implemented operation=mailchimp.get_ecommerce_stores_id_orders_id]; flags: --store-id, --order-id, --page, --page-cursor
    ecommerce stores orders lines get - Get order line item [intent=direct_read availability=implemented operation=mailchimp.get_ecommerce_stores_id_orders_id_lines_id]; flags: --store-id, --order-id, --line-id, --page, --page-cursor
    ecommerce stores products get - Get product info [intent=direct_read availability=implemented operation=mailchimp.get_ecommerce_stores_id_products_id]; flags: --store-id, --product-id, --page, --page-cursor
    ecommerce stores products images get - Get product image info [intent=direct_read availability=implemented operation=mailchimp.get_ecommerce_stores_id_products_id_images_id]; flags: --store-id, --product-id, --image-id, --page, --page-cursor
    ecommerce stores products variants get - Get product variant info [intent=direct_read availability=implemented operation=mailchimp.get_ecommerce_stores_id_products_id_variants_id]; flags: --store-id, --product-id, --variant-id, --page, --page-cursor
    ecommerce stores promo-rules get - Get promo rule [intent=direct_read availability=implemented operation=mailchimp.get_ecommerce_stores_id_promorules_id]; flags: --store-id, --promo-rule-id, --page, --page-cursor
    ecommerce stores promo-rules promo-codes get - Get promo code [intent=direct_read availability=implemented operation=mailchimp.get_ecommerce_stores_id_promocodes_id]; flags: --store-id, --promo-rule-id, --promo-code-id, --page, --page-cursor
    templates list - List templates as ETL records. [intent=etl availability=implemented stream=templates]
    templates get - Get template info [intent=direct_read availability=implemented operation=mailchimp.get_templates_id]; flags: --template-id, --page, --page-cursor
    templates default-content get - View default content [intent=direct_read availability=implemented operation=mailchimp.get_templates_id_default_content]; flags: --template-id, --page, --page-cursor
    file-manager files get - Get file [intent=direct_read availability=implemented operation=mailchimp.get_file_manager_files_id]; flags: --file-id, --page, --page-cursor
    file-manager folders get - Get folder [intent=direct_read availability=implemented operation=mailchimp.get_file_manager_folders_id]; flags: --folder-id, --page, --page-cursor
  Typed direct reads and search
    search-members get - Search members [intent=direct_read availability=implemented operation=mailchimp.get_search_members]; flags: --query, --page, --page-cursor
    search-campaigns get - Search campaigns [intent=direct_read availability=implemented operation=mailchimp.get_search_campaigns]; flags: --query, --page, --page-cursor
    reports list - List campaign reports as ETL records. [intent=etl availability=implemented stream=reports]
    reports get - Get campaign report [intent=direct_read availability=implemented operation=mailchimp.get_reports_id]; flags: --campaign-id, --page, --page-cursor
    reports abuse-reports get - Get abuse report [intent=direct_read availability=implemented operation=mailchimp.get_reports_id_abuse_reports_id_id]; flags: --campaign-id, --report-id, --page, --page-cursor
    reports click-details get - Get campaign link details [intent=direct_read availability=implemented operation=mailchimp.get_reports_id_click_details_id]; flags: --campaign-id, --link-id, --page, --page-cursor
    reports click-details members get - Get clicked link subscriber [intent=direct_read availability=implemented operation=mailchimp.get_reports_id_click_details_id_members_id]; flags: --campaign-id, --link-id, --subscriber-hash, --page, --page-cursor
    reports email-activity get - Get subscriber email activity [intent=direct_read availability=implemented operation=mailchimp.get_reports_id_email_activity_id]; flags: --campaign-id, --subscriber-hash, --page, --page-cursor
    reports open-details get - Get opened campaign subscriber [intent=direct_read availability=implemented operation=mailchimp.get_reports_id_open_details_id_members_id]; flags: --campaign-id, --subscriber-hash, --page, --page-cursor
    reports sent-to get - Get campaign recipient info [intent=direct_read availability=implemented operation=mailchimp.get_reports_id_sent_to_id]; flags: --campaign-id, --subscriber-hash, --page, --page-cursor
    reports unsubscribed get - Get unsubscribed member [intent=direct_read availability=implemented operation=mailchimp.get_reports_id_unsubscribed_id]; flags: --campaign-id, --subscriber-hash, --page, --page-cursor
    campaigns list - List campaigns as ETL records. [intent=etl availability=implemented stream=campaigns]
    campaigns get - Get campaign info [intent=direct_read availability=implemented operation=mailchimp.get_campaigns_id]; flags: --campaign-id, --page, --page-cursor
    campaigns content get - Get campaign content [intent=direct_read availability=implemented operation=mailchimp.get_campaigns_id_content]; flags: --campaign-id, --page, --page-cursor
    campaigns feedback get - Get campaign feedback message [intent=direct_read availability=implemented operation=mailchimp.get_campaigns_id_feedback_id]; flags: --campaign-id, --feedback-id, --page, --page-cursor
    campaigns send-checklist get - Get campaign send checklist [intent=direct_read availability=implemented operation=mailchimp.get_campaigns_id_send_checklist]; flags: --campaign-id, --page, --page-cursor
    lists list - Get lists info as ETL records. [intent=etl availability=implemented stream=lists]
    lists get - Get list info [intent=direct_read availability=implemented operation=mailchimp.get_lists_id]; flags: --list-id, --page, --page-cursor
    lists abuse-reports get - Get abuse report [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_abuse_reports_id]; flags: --list-id, --report-id, --page, --page-cursor
    lists growth-history get - Get growth history by month [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_growth_history_id]; flags: --list-id, --month, --page, --page-cursor
    lists interest-categories get - Get interest category info [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_interest_categories_id]; flags: --list-id, --interest-category-id, --page, --page-cursor
    lists interest-categories interests get - Get interest in category [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_interest_categories_id_interests_id]; flags: --list-id, --interest-category-id, --interest-id, --page, --page-cursor
    lists members get - Get member info [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_members_id]; flags: --list-id, --subscriber-hash, --page, --page-cursor
    lists members activity-feed get - View recent activity [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_members_id_activity_feed]; flags: --list-id, --subscriber-hash, --page, --page-cursor
    lists members notes get - Get member note [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_members_id_notes_id]; flags: --list-id, --subscriber-hash, --note-id, --page, --page-cursor
    lists merge-fields get - Get merge field [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_merge_fields_id]; flags: --list-id, --merge-id, --page, --page-cursor
    lists segments get - Get segment info [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_segments_id]; flags: --list-id, --segment-id, --page, --page-cursor
    lists surveys get - Get survey [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_surveys_id]; flags: --list-id, --survey-id, --page, --page-cursor
    lists tag-search get - Search for tags on a list by name. [intent=direct_read availability=implemented operation=mailchimp.search_tags_by_name]; flags: --list-id, --query, --page, --page-cursor
    lists webhooks get - Get webhook info [intent=direct_read availability=implemented operation=mailchimp.get_lists_id_webhooks_id]; flags: --list-id, --webhook-id, --page, --page-cursor
  Approval-gated reverse ETL actions
    actions post-account-exports - Add export [intent=reverse_etl availability=implemented write=post_account_exports]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add export. Reverse ETL must plan, preview, receive explicit approval, and then execute.
    actions post-audiences-contacts - Add Contact [intent=reverse_etl availability=implemented write=post_audiences_contacts]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add Contact. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --audience-id
    actions patch-audiences-contacts - Update Contact [intent=reverse_etl availability=implemented write=patch_audiences_contacts]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update Contact. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --audience-id, --contact-id
    actions archive-contacts - Archive Contact [intent=reverse_etl availability=implemented write=archive_contacts]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Archive Contact. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --audience-id, --contact-id
    actions forget-contacts - Forget Contact [intent=reverse_etl availability=implemented write=forget_contacts]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Forget Contact. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --audience-id, --contact-id
    actions post-automations - Add automation [intent=reverse_etl availability=implemented write=post_automations]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add automation. Reverse ETL must plan, preview, receive explicit approval, and then execute.
    actions archive-automations - Archive automation [intent=reverse_etl availability=implemented write=archive_automations]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Archive automation. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --workflow-id
    actions pause-all-emails-automations - Pause automation emails [intent=reverse_etl availability=implemented write=pause_all_emails_automations]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Pause automation emails. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --workflow-id
    actions start-all-emails-automations - Start automation emails [intent=reverse_etl availability=implemented write=start_all_emails_automations]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Start automation emails. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --workflow-id
    actions delete-automations-emails - Delete workflow email [intent=reverse_etl availability=implemented write=delete_automations_emails]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete workflow email. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --workflow-id, --workflow-email-id
    actions patch-automations-emails - Update workflow email [intent=reverse_etl availability=implemented write=patch_automations_emails]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update workflow email. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --workflow-id, --workflow-email-id
    actions pause-emails - Pause automated email [intent=reverse_etl availability=implemented write=pause_emails]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Pause automated email. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --workflow-id, --workflow-email-id
    actions start-emails - Start automated email [intent=reverse_etl availability=implemented write=start_emails]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Start automated email. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --workflow-id, --workflow-email-id
    actions post-automations-emails-queue - Add subscriber to workflow email [intent=reverse_etl availability=implemented write=post_automations_emails_queue]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add subscriber to workflow email. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --workflow-id, --workflow-email-id
    actions post-automations-removed-subscribers - Remove subscriber from workflow [intent=reverse_etl availability=implemented write=post_automations_removed_subscribers]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Remove subscriber from workflow. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --workflow-id
    actions post-batch-webhooks - Add batch webhook [intent=reverse_etl availability=implemented write=post_batch_webhooks]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add batch webhook. Reverse ETL must plan, preview, receive explicit approval, and then execute.
    actions delete-batch-webhooks - Delete batch webhook [intent=reverse_etl availability=implemented write=delete_batch_webhooks]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete batch webhook. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --batch-webhook-id
    actions patch-batch-webhooks - Update batch webhook [intent=reverse_etl availability=implemented write=patch_batch_webhooks]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update batch webhook. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --batch-webhook-id
    actions delete-batches - Delete batch request [intent=reverse_etl availability=implemented write=delete_batches]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete batch request. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --batch-id
    actions post-campaign-folders - Add campaign folder [intent=reverse_etl availability=implemented write=post_campaign_folders]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add campaign folder. Reverse ETL must plan, preview, receive explicit approval, and then execute.
    actions delete-campaign-folders - Delete campaign folder [intent=reverse_etl availability=implemented write=delete_campaign_folders]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete campaign folder. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --folder-id
    actions patch-campaign-folders - Update campaign folder [intent=reverse_etl availability=implemented write=patch_campaign_folders]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update campaign folder. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --folder-id
    actions post-campaigns - Add campaign [intent=reverse_etl availability=implemented write=post_campaigns]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.
    actions delete-campaigns - Delete campaign [intent=reverse_etl availability=implemented write=delete_campaigns]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --campaign-id
    actions patch-campaigns - Update campaign settings [intent=reverse_etl availability=implemented write=patch_campaigns]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update campaign settings. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --campaign-id
    actions cancel-send-campaigns - Cancel campaign [intent=reverse_etl availability=implemented write=cancel_send_campaigns]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Cancel campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --campaign-id
    actions create-resend-campaigns - Resend campaign [intent=reverse_etl availability=implemented write=create_resend_campaigns]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Resend campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --campaign-id
    actions pause-campaigns - Pause rss campaign [intent=reverse_etl availability=implemented write=pause_campaigns]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Pause rss campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --campaign-id
    actions replicate-campaigns - Replicate campaign [intent=reverse_etl availability=implemented write=replicate_campaigns]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Replicate campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --campaign-id
    actions resume-campaigns - Resume rss campaign [intent=reverse_etl availability=implemented write=resume_campaigns]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Resume rss campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --campaign-id
    actions schedule-campaigns - Schedule campaign [intent=reverse_etl availability=implemented write=schedule_campaigns]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Schedule campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --campaign-id
    actions send-campaigns - Send campaign [intent=reverse_etl availability=implemented write=send_campaigns]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Send campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --campaign-id
    actions test-campaigns - Send test email [intent=reverse_etl availability=implemented write=test_campaigns]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Send test email. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --campaign-id
    actions unschedule-campaigns - Unschedule campaign [intent=reverse_etl availability=implemented write=unschedule_campaigns]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Unschedule campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --campaign-id
    actions put-campaigns-content - Set campaign content [intent=reverse_etl availability=implemented write=put_campaigns_content]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Set campaign content. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --campaign-id
    actions post-campaigns-feedback - Add campaign feedback [intent=reverse_etl availability=implemented write=post_campaigns_feedback]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add campaign feedback. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --campaign-id
    actions delete-campaigns-feedback - Delete campaign feedback message [intent=reverse_etl availability=implemented write=delete_campaigns_feedback]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete campaign feedback message. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --campaign-id, --feedback-id
    actions patch-campaigns-feedback - Update campaign feedback message [intent=reverse_etl availability=implemented write=patch_campaigns_feedback]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update campaign feedback message. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --campaign-id, --feedback-id
    actions post-connected-sites - Add connected site [intent=reverse_etl availability=implemented write=post_connected_sites]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add connected site. Reverse ETL must plan, preview, receive explicit approval, and then execute.
    actions delete-connected-sites - Delete connected site [intent=reverse_etl availability=implemented write=delete_connected_sites]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete connected site. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --connected-site-id
    actions disable-pixel-connected-sites - Disable pixel for connected site [intent=reverse_etl availability=implemented write=disable_pixel_connected_sites]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Disable pixel for connected site. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --connected-site-id
    actions enable-pixel-connected-sites - Enable pixel for connected site [intent=reverse_etl availability=implemented write=enable_pixel_connected_sites]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Enable pixel for connected site. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --connected-site-id
    actions verify-script-installation-connected-sites - Verify connected site script [intent=reverse_etl availability=implemented write=verify_script_installation_connected_sites]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Verify connected site script. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --connected-site-id
    actions trigger-steps - Customer Journeys API trigger for a contact [intent=reverse_etl availability=implemented write=trigger_steps]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Customer Journeys API trigger for a contact. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --journey-id, --step-id
    actions post-ecommerce-stores - Add store [intent=reverse_etl availability=implemented write=post_ecommerce_stores]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add store. Reverse ETL must plan, preview, receive explicit approval, and then execute.
    actions delete-ecommerce-stores - Delete store [intent=reverse_etl availability=implemented write=delete_ecommerce_stores]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete store. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id
    actions patch-ecommerce-stores - Update store [intent=reverse_etl availability=implemented write=patch_ecommerce_stores]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update store. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id
    actions post-ecommerce-stores-carts - Add cart [intent=reverse_etl availability=implemented write=post_ecommerce_stores_carts]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add cart. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id
    actions delete-ecommerce-stores-carts - Delete cart [intent=reverse_etl availability=implemented write=delete_ecommerce_stores_carts]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete cart. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --cart-id
    actions patch-ecommerce-stores-carts - Update cart [intent=reverse_etl availability=implemented write=patch_ecommerce_stores_carts]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update cart. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --cart-id
    actions post-ecommerce-stores-carts-lines - Add cart line item [intent=reverse_etl availability=implemented write=post_ecommerce_stores_carts_lines]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add cart line item. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --cart-id
    actions delete-ecommerce-stores-carts-lines - Delete cart line item [intent=reverse_etl availability=implemented write=delete_ecommerce_stores_carts_lines]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete cart line item. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --cart-id, --line-id
    actions patch-ecommerce-stores-carts-lines - Update cart line item [intent=reverse_etl availability=implemented write=patch_ecommerce_stores_carts_lines]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update cart line item. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --cart-id, --line-id
    actions post-ecommerce-stores-customers - Add customer [intent=reverse_etl availability=implemented write=post_ecommerce_stores_customers]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add customer. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id
    actions delete-ecommerce-stores-customers - Delete customer [intent=reverse_etl availability=implemented write=delete_ecommerce_stores_customers]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete customer. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --customer-id
    actions patch-ecommerce-stores-customers - Update customer [intent=reverse_etl availability=implemented write=patch_ecommerce_stores_customers]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update customer. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --customer-id
    actions put-ecommerce-stores-customers - Add or update customer [intent=reverse_etl availability=implemented write=put_ecommerce_stores_customers]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add or update customer. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --customer-id
    actions post-ecommerce-stores-orders - Add order [intent=reverse_etl availability=implemented write=post_ecommerce_stores_orders]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add order. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id
    actions delete-ecommerce-stores-orders - Delete order [intent=reverse_etl availability=implemented write=delete_ecommerce_stores_orders]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete order. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --order-id
    actions patch-ecommerce-stores-orders - Update order [intent=reverse_etl availability=implemented write=patch_ecommerce_stores_orders]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update order. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --order-id
    actions put-ecommerce-stores-orders - Add or update order [intent=reverse_etl availability=implemented write=put_ecommerce_stores_orders]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add or update order. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --order-id
    actions post-ecommerce-stores-orders-lines - Add order line item [intent=reverse_etl availability=implemented write=post_ecommerce_stores_orders_lines]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add order line item. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --order-id
    actions delete-ecommerce-stores-orders-lines - Delete order line item [intent=reverse_etl availability=implemented write=delete_ecommerce_stores_orders_lines]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete order line item. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --order-id, --line-id
    actions patch-ecommerce-stores-orders-lines - Update order line item [intent=reverse_etl availability=implemented write=patch_ecommerce_stores_orders_lines]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update order line item. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --order-id, --line-id
    actions post-ecommerce-stores-products - Add product [intent=reverse_etl availability=implemented write=post_ecommerce_stores_products]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add product. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id
    actions delete-ecommerce-stores-products - Delete product [intent=reverse_etl availability=implemented write=delete_ecommerce_stores_products]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete product. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --product-id
    actions patch-ecommerce-stores-products - Update product [intent=reverse_etl availability=implemented write=patch_ecommerce_stores_products]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update product. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --product-id
    actions put-ecommerce-stores-products - Create or update product [intent=reverse_etl availability=implemented write=put_ecommerce_stores_products]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Create or update product. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --product-id
    actions post-ecommerce-stores-products-images - Add product image [intent=reverse_etl availability=implemented write=post_ecommerce_stores_products_images]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add product image. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --product-id
    actions delete-ecommerce-stores-products-images - Delete product image [intent=reverse_etl availability=implemented write=delete_ecommerce_stores_products_images]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete product image. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --product-id, --image-id
    actions patch-ecommerce-stores-products-images - Update product image [intent=reverse_etl availability=implemented write=patch_ecommerce_stores_products_images]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update product image. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --product-id, --image-id
    actions post-ecommerce-stores-products-variants - Add product variant [intent=reverse_etl availability=implemented write=post_ecommerce_stores_products_variants]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add product variant. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --product-id
    actions delete-ecommerce-stores-products-variants - Delete product variant [intent=reverse_etl availability=implemented write=delete_ecommerce_stores_products_variants]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete product variant. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --product-id, --variant-id
    actions patch-ecommerce-stores-products-variants - Update product variant [intent=reverse_etl availability=implemented write=patch_ecommerce_stores_products_variants]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update product variant. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --product-id, --variant-id
    actions put-ecommerce-stores-products-variants - Add or update product variant [intent=reverse_etl availability=implemented write=put_ecommerce_stores_products_variants]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add or update product variant. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --product-id, --variant-id
    actions post-ecommerce-stores-promo-rules - Add promo rule [intent=reverse_etl availability=implemented write=post_ecommerce_stores_promo_rules]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add promo rule. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id
    actions delete-ecommerce-stores-promo-rules - Delete promo rule [intent=reverse_etl availability=implemented write=delete_ecommerce_stores_promo_rules]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete promo rule. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --promo-rule-id
    actions patch-ecommerce-stores-promo-rules - Update promo rule [intent=reverse_etl availability=implemented write=patch_ecommerce_stores_promo_rules]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update promo rule. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --promo-rule-id
    actions post-ecommerce-stores-promo-rules-promo-codes - Add promo code [intent=reverse_etl availability=implemented write=post_ecommerce_stores_promo_rules_promo_codes]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add promo code. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --promo-rule-id
    actions delete-ecommerce-stores-promo-rules-promo-codes - Delete promo code [intent=reverse_etl availability=implemented write=delete_ecommerce_stores_promo_rules_promo_codes]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete promo code. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --promo-rule-id, --promo-code-id
    actions patch-ecommerce-stores-promo-rules-promo-codes - Update promo code [intent=reverse_etl availability=implemented write=patch_ecommerce_stores_promo_rules_promo_codes]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update promo code. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --store-id, --promo-rule-id, --promo-code-id
    actions post-file-manager-files - Add file [intent=reverse_etl availability=implemented write=post_file_manager_files]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add file. Reverse ETL must plan, preview, receive explicit approval, and then execute.
    actions delete-file-manager-files - Delete file [intent=reverse_etl availability=implemented write=delete_file_manager_files]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete file. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --file-id
    actions patch-file-manager-files - Update file [intent=reverse_etl availability=implemented write=patch_file_manager_files]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update file. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --file-id
    actions post-file-manager-folders - Add folder [intent=reverse_etl availability=implemented write=post_file_manager_folders]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add folder. Reverse ETL must plan, preview, receive explicit approval, and then execute.
    actions delete-file-manager-folders - Delete folder [intent=reverse_etl availability=implemented write=delete_file_manager_folders]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete folder. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --folder-id
    actions patch-file-manager-folders - Update folder [intent=reverse_etl availability=implemented write=patch_file_manager_folders]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update folder. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --folder-id
    actions post-landing-pages - Add landing page [intent=reverse_etl availability=implemented write=post_landing_pages]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add landing page. Reverse ETL must plan, preview, receive explicit approval, and then execute.
    actions delete-landing-pages - Delete landing page [intent=reverse_etl availability=implemented write=delete_landing_pages]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete landing page. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --page-id
    actions patch-landing-pages - Update landing page [intent=reverse_etl availability=implemented write=patch_landing_pages]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update landing page. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --page-id
    actions publish-landing-pages - Publish landing page [intent=reverse_etl availability=implemented write=publish_landing_pages]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Publish landing page. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --page-id
    actions unpublish-landing-pages - Unpublish landing page [intent=reverse_etl availability=implemented write=unpublish_landing_pages]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Unpublish landing page. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --page-id
    actions post-lists - Add list [intent=reverse_etl availability=implemented write=post_lists]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add list. Reverse ETL must plan, preview, receive explicit approval, and then execute.
    actions delete-lists - Delete list [intent=reverse_etl availability=implemented write=delete_lists]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete list. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id
    actions patch-lists - Update lists [intent=reverse_etl availability=implemented write=patch_lists]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update lists. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id
    actions post-lists-2 - Batch subscribe or unsubscribe [intent=reverse_etl availability=implemented write=post_lists_2]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Batch subscribe or unsubscribe. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id
    actions post-lists-interest-categories - Add interest category [intent=reverse_etl availability=implemented write=post_lists_interest_categories]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add interest category. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id
    actions delete-lists-interest-categories - Delete interest category [intent=reverse_etl availability=implemented write=delete_lists_interest_categories]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete interest category. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --interest-category-id
    actions patch-lists-interest-categories - Update interest category [intent=reverse_etl availability=implemented write=patch_lists_interest_categories]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update interest category. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --interest-category-id
    actions post-lists-interest-categories-interests - Add interest in category [intent=reverse_etl availability=implemented write=post_lists_interest_categories_interests]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add interest in category. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --interest-category-id
    actions delete-lists-interest-categories-interests - Delete interest in category [intent=reverse_etl availability=implemented write=delete_lists_interest_categories_interests]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete interest in category. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --interest-category-id, --interest-id
    actions patch-lists-interest-categories-interests - Update interest in category [intent=reverse_etl availability=implemented write=patch_lists_interest_categories_interests]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update interest in category. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --interest-category-id, --interest-id
    actions post-lists-members - Add member to list [intent=reverse_etl availability=implemented write=post_lists_members]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add member to list. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id
    actions delete-lists-members - Archive list member [intent=reverse_etl availability=implemented write=delete_lists_members]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Archive list member. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --subscriber-hash
    actions patch-lists-members - Update list member [intent=reverse_etl availability=implemented write=patch_lists_members]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update list member. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --subscriber-hash
    actions put-lists-members - Add or update list member [intent=reverse_etl availability=implemented write=put_lists_members]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add or update list member. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --subscriber-hash
    actions delete-permanent-members - Delete list member [intent=reverse_etl availability=implemented write=delete_permanent_members]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete list member. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --subscriber-hash
    actions post-lists-members-events - Add event [intent=reverse_etl availability=implemented write=post_lists_members_events]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add event. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --subscriber-hash
    actions post-lists-members-notes - Add member note [intent=reverse_etl availability=implemented write=post_lists_members_notes]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add member note. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --subscriber-hash
    actions delete-lists-members-notes - Delete note [intent=reverse_etl availability=implemented write=delete_lists_members_notes]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete note. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --subscriber-hash, --note-id
    actions patch-lists-members-notes - Update note [intent=reverse_etl availability=implemented write=patch_lists_members_notes]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update note. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --subscriber-hash, --note-id
    actions post-lists-members-tags - Add or remove member tags [intent=reverse_etl availability=implemented write=post_lists_members_tags]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add or remove member tags. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --subscriber-hash
    actions post-lists-merge-fields - Add merge field [intent=reverse_etl availability=implemented write=post_lists_merge_fields]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add merge field. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id
    actions delete-lists-merge-fields - Delete merge field [intent=reverse_etl availability=implemented write=delete_lists_merge_fields]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete merge field. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --merge-id
    actions patch-lists-merge-fields - Update merge field [intent=reverse_etl availability=implemented write=patch_lists_merge_fields]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update merge field. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --merge-id
    actions post-lists-segments - Add segment [intent=reverse_etl availability=implemented write=post_lists_segments]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add segment. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id
    actions delete-lists-segments - Delete segment [intent=reverse_etl availability=implemented write=delete_lists_segments]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete segment. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --segment-id
    actions patch-lists-segments - Update segment [intent=reverse_etl availability=implemented write=patch_lists_segments]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update segment. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --segment-id
    actions post-lists-segments-2 - Batch add or remove members [intent=reverse_etl availability=implemented write=post_lists_segments_2]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Batch add or remove members. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --segment-id
    actions post-lists-segments-members - Add member to segment [intent=reverse_etl availability=implemented write=post_lists_segments_members]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add member to segment. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --segment-id
    actions delete-lists-segments-members - Remove list member from segment [intent=reverse_etl availability=implemented write=delete_lists_segments_members]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Remove list member from segment. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --segment-id, --subscriber-hash
    actions post-lists-signup-forms - Customize signup form [intent=reverse_etl availability=implemented write=post_lists_signup_forms]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Customize signup form. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id
    actions post-lists-surveys - Create survey [intent=reverse_etl availability=implemented write=post_lists_surveys]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Create survey. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id
    actions delete-lists-surveys - Delete survey [intent=reverse_etl availability=implemented write=delete_lists_surveys]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete survey. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --survey-id
    actions patch-lists-surveys - Update survey [intent=reverse_etl availability=implemented write=patch_lists_surveys]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update survey. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --survey-id
    actions create-email-surveys - Create a Survey Campaign [intent=reverse_etl availability=implemented write=create_email_surveys]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Create a Survey Campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --survey-id
    actions publish-surveys - Publish a Survey [intent=reverse_etl availability=implemented write=publish_surveys]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Publish a Survey. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --survey-id
    actions replicate-surveys - Replicate survey [intent=reverse_etl availability=implemented write=replicate_surveys]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Replicate survey. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --survey-id
    actions unpublish-surveys - Unpublish a Survey [intent=reverse_etl availability=implemented write=unpublish_surveys]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Unpublish a Survey. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --survey-id
    actions post-lists-webhooks - Add webhook [intent=reverse_etl availability=implemented write=post_lists_webhooks]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add webhook. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id
    actions delete-lists-webhooks - Delete webhook [intent=reverse_etl availability=implemented write=delete_lists_webhooks]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete webhook. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --webhook-id
    actions patch-lists-webhooks - Update webhook [intent=reverse_etl availability=implemented write=patch_lists_webhooks]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update webhook. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --list-id, --webhook-id
    actions post-sms-campaigns - Add SMS campaign [intent=reverse_etl availability=implemented write=post_sms_campaigns]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add SMS campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.
    actions delete-sms-campaigns - Delete SMS campaign [intent=reverse_etl availability=implemented write=delete_sms_campaigns]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete SMS campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --sms-campaign-id
    actions patch-sms-campaigns - Update SMS campaign settings [intent=reverse_etl availability=implemented write=patch_sms_campaigns]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update SMS campaign settings. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --sms-campaign-id
    actions cancel-send-sms-campaigns - Cancel SMS campaign send [intent=reverse_etl availability=implemented write=cancel_send_sms_campaigns]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Cancel SMS campaign send. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --sms-campaign-id
    actions schedule-sms-campaigns - Schedule SMS campaign [intent=reverse_etl availability=implemented write=schedule_sms_campaigns]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Schedule SMS campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --sms-campaign-id
    actions send-sms-campaigns - Send SMS campaign [intent=reverse_etl availability=implemented write=send_sms_campaigns]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Send SMS campaign. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --sms-campaign-id
    actions put-sms-campaigns-content - Set SMS campaign content [intent=reverse_etl availability=implemented write=put_sms_campaigns_content]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Set SMS campaign content. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --sms-campaign-id
    actions post-template-folders - Add template folder [intent=reverse_etl availability=implemented write=post_template_folders]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add template folder. Reverse ETL must plan, preview, receive explicit approval, and then execute.
    actions delete-template-folders - Delete template folder [intent=reverse_etl availability=implemented write=delete_template_folders]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete template folder. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --folder-id
    actions patch-template-folders - Update template folder [intent=reverse_etl availability=implemented write=patch_template_folders]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update template folder. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --folder-id
    actions post-templates - Add template [intent=reverse_etl availability=implemented write=post_templates]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add template. Reverse ETL must plan, preview, receive explicit approval, and then execute.
    actions delete-templates - Delete template [intent=reverse_etl availability=implemented write=delete_templates]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete template. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --template-id
    actions patch-templates - Update template [intent=reverse_etl availability=implemented write=patch_templates]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Update template. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --template-id
    actions post-verified-domains - Add domain to account [intent=reverse_etl availability=implemented write=post_verified_domains]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: Externally visible Mailchimp mutation: Add domain to account. Reverse ETL must plan, preview, receive explicit approval, and then execute.
    actions delete-verified-domains - Delete domain [intent=reverse_etl availability=implemented write=delete_verified_domains]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Delete domain. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --domain-name
    actions verify-verified-domains - Verify domain [intent=reverse_etl availability=implemented write=verify_verified_domains]; approval: reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation; risk: Destructive or externally visible Mailchimp mutation: Verify domain. Reverse ETL must plan, preview, receive explicit approval, and then execute.; flags: --domain-name
  Other Commands
    account-exports-exports list - List account exports as ETL records. [intent=etl availability=implemented stream=account_exports_exports]
    account-exports get - Get account export info [intent=direct_read availability=implemented operation=mailchimp.get_account_export_id]; flags: --export-id, --page, --page-cursor
    activity-feed-chimp-chatter list - Get latest chimp chatter as ETL records. [intent=etl availability=implemented stream=activity_feed_chimp_chatter]
    audiences list - Get a list of audiences as ETL records. [intent=etl availability=implemented stream=audiences]
    audiences get - Get audience info [intent=direct_read availability=implemented operation=mailchimp.get_audience_id]; flags: --audience-id, --page, --page-cursor
    audiences-contacts list - Get Contacts as ETL records. [intent=etl availability=implemented stream=audiences_contacts]
    audiences contacts get - Get Contact [intent=direct_read availability=implemented operation=mailchimp.get_audience_contact]; flags: --audience-id, --contact-id, --page, --page-cursor
    authorized-apps-apps list - List authorized apps as ETL records. [intent=etl availability=implemented stream=authorized_apps_apps]
    authorized-apps get - Get authorized app info [intent=direct_read availability=implemented operation=mailchimp.get_authorized_apps_id]; flags: --app-id, --page, --page-cursor
    automations-emails list - List automated emails as ETL records. [intent=etl availability=implemented stream=automations_emails]
    automations-emails-queue list - List automated email subscribers as ETL records. [intent=etl availability=implemented stream=automations_emails_queue]
    automations-removed-subscribers-subscribers list - List subscribers removed from workflow as ETL records. [intent=etl availability=implemented stream=automations_removed_subscribers_subscribers]
    batch-webhooks-webhooks list - List batch webhooks as ETL records. [intent=etl availability=implemented stream=batch_webhooks_webhooks]
    batch-webhooks get - Get batch webhook info [intent=direct_read availability=implemented operation=mailchimp.get_batch_webhook]; flags: --batch-webhook-id, --page, --page-cursor
    batches list - List batch requests as ETL records. [intent=etl availability=implemented stream=batches]
    batches get - Get batch operation status [intent=direct_read availability=implemented operation=mailchimp.get_batches_id]; flags: --batch-id, --page, --page-cursor
    campaign-folders-folders list - List campaign folders as ETL records. [intent=etl availability=implemented stream=campaign_folders_folders]
    campaign-folders get - Get campaign folder [intent=direct_read availability=implemented operation=mailchimp.get_campaign_folders_id]; flags: --folder-id, --page, --page-cursor
    campaigns-feedback list - List campaign feedback as ETL records. [intent=etl availability=implemented stream=campaigns_feedback]
    connected-sites-sites list - List connected sites as ETL records. [intent=etl availability=implemented stream=connected_sites_sites]
    connected-sites get - Get connected site [intent=direct_read availability=implemented operation=mailchimp.get_connected_sites_id]; flags: --connected-site-id, --page, --page-cursor
    conversations list - List conversations as ETL records. [intent=etl availability=implemented stream=conversations]
    conversations get - Get conversation [intent=direct_read availability=implemented operation=mailchimp.get_conversations_id]; flags: --conversation-id, --page, --page-cursor
    conversations-messages-conversation-messages list - List messages as ETL records. [intent=etl availability=implemented stream=conversations_messages_conversation_messages]
    conversations messages get - Get message [intent=direct_read availability=implemented operation=mailchimp.get_conversations_id_messages_id]; flags: --conversation-id, --message-id, --page, --page-cursor
    ecommerce-orders list - List account orders as ETL records. [intent=etl availability=implemented stream=ecommerce_orders]
    ecommerce-stores list - List stores as ETL records. [intent=etl availability=implemented stream=ecommerce_stores]
    ecommerce-stores-carts list - List carts as ETL records. [intent=etl availability=implemented stream=ecommerce_stores_carts]
    ecommerce-stores-carts-lines list - List cart line items as ETL records. [intent=etl availability=implemented stream=ecommerce_stores_carts_lines]
    ecommerce-stores-customers list - List customers as ETL records. [intent=etl availability=implemented stream=ecommerce_stores_customers]
    ecommerce-stores-orders list - List orders as ETL records. [intent=etl availability=implemented stream=ecommerce_stores_orders]
    ecommerce-stores-orders-lines list - List order line items as ETL records. [intent=etl availability=implemented stream=ecommerce_stores_orders_lines]
    ecommerce-stores-products list - List product as ETL records. [intent=etl availability=implemented stream=ecommerce_stores_products]
    ecommerce-stores-products-images list - List product images as ETL records. [intent=etl availability=implemented stream=ecommerce_stores_products_images]
    ecommerce-stores-products-variants list - List product variants as ETL records. [intent=etl availability=implemented stream=ecommerce_stores_products_variants]
    ecommerce-stores-promo-rules list - List promo rules as ETL records. [intent=etl availability=implemented stream=ecommerce_stores_promo_rules]
    ecommerce-stores-promo-rules-promo-codes list - List promo codes as ETL records. [intent=etl availability=implemented stream=ecommerce_stores_promo_rules_promo_codes]
    facebook-ads list - List facebook ads as ETL records. [intent=etl availability=implemented stream=facebook_ads]
    facebook-ads get - Get facebook ad info [intent=direct_read availability=implemented operation=mailchimp.get_facebook_ads_id]; flags: --outreach-id, --page, --page-cursor
    file-manager-files list - List stored files as ETL records. [intent=etl availability=implemented stream=file_manager_files]
    file-manager-folders list - List folders as ETL records. [intent=etl availability=implemented stream=file_manager_folders]
    file-manager-folders-files list - List stored files as ETL records. [intent=etl availability=implemented stream=file_manager_folders_files]
    landing-pages list - List landing pages as ETL records. [intent=etl availability=implemented stream=landing_pages]
    landing-pages get - Get landing page info [intent=direct_read availability=implemented operation=mailchimp.get_landing_page_id]; flags: --page-id, --page, --page-cursor
    landing-pages content get - Get landing page content [intent=direct_read availability=implemented operation=mailchimp.get_landing_page_id_content]; flags: --page-id, --page, --page-cursor
    lists-abuse-reports list - List abuse reports as ETL records. [intent=etl availability=implemented stream=lists_abuse_reports]
    lists-activity list - List recent activity as ETL records. [intent=etl availability=implemented stream=lists_activity]
    lists-clients list - List top email clients as ETL records. [intent=etl availability=implemented stream=lists_clients]
    lists-growth-history-history list - List growth history data as ETL records. [intent=etl availability=implemented stream=lists_growth_history_history]
    lists-interest-categories-categories list - List interest categories as ETL records. [intent=etl availability=implemented stream=lists_interest_categories_categories]
    lists-interest-categories-interests list - List interests in category as ETL records. [intent=etl availability=implemented stream=lists_interest_categories_interests]
    lists-locations list - List locations as ETL records. [intent=etl availability=implemented stream=lists_locations]
    lists-members list - List members info as ETL records. [intent=etl availability=implemented stream=lists_members]
    lists-members-activity list - View recent activity 50 as ETL records. [intent=etl availability=implemented stream=lists_members_activity]
    lists-members-events list - List member events as ETL records. [intent=etl availability=implemented stream=lists_members_events]
    lists-members-goals list - List member goal events as ETL records. [intent=etl availability=implemented stream=lists_members_goals]
    lists-members-notes list - List recent member notes as ETL records. [intent=etl availability=implemented stream=lists_members_notes]
    lists-members-tags list - List member tags as ETL records. [intent=etl availability=implemented stream=lists_members_tags]
    lists-merge-fields list - List merge fields as ETL records. [intent=etl availability=implemented stream=lists_merge_fields]
    lists-segments list - List segments as ETL records. [intent=etl availability=implemented stream=lists_segments]
    lists-segments-members list - List members in segment as ETL records. [intent=etl availability=implemented stream=lists_segments_members]
    lists-signup-forms list - List signup forms as ETL records. [intent=etl availability=implemented stream=lists_signup_forms]
    lists-surveys list - Get information about all surveys for a list as ETL records. [intent=etl availability=implemented stream=lists_surveys]
    lists-webhooks list - List webhooks as ETL records. [intent=etl availability=implemented stream=lists_webhooks]
    reporting-facebook-ads list - List facebook ads reports as ETL records. [intent=etl availability=implemented stream=reporting_facebook_ads]
    reporting facebook-ads get - Get facebook ad report [intent=direct_read availability=implemented operation=mailchimp.get_reporting_facebook_ads_id]; flags: --outreach-id, --page, --page-cursor
    reporting-facebook-ads-ecommerce-product-activity-products list - List facebook ecommerce report as ETL records. [intent=etl availability=implemented stream=reporting_facebook_ads_ecommerce_product_activity_products]
    reporting-landing-pages list - List landing pages reports as ETL records. [intent=etl availability=implemented stream=reporting_landing_pages]
    reporting landing-pages get - Get landing page report [intent=direct_read availability=implemented operation=mailchimp.get_reporting_landing_pages_id]; flags: --outreach-id, --page, --page-cursor
    reporting-surveys list - List survey reports as ETL records. [intent=etl availability=implemented stream=reporting_surveys]
    reporting surveys get - Get survey report [intent=direct_read availability=implemented operation=mailchimp.get_reporting_surveys_id]; flags: --survey-id, --page, --page-cursor
    reporting-surveys-questions list - List survey question reports as ETL records. [intent=etl availability=implemented stream=reporting_surveys_questions]
    reporting surveys questions get - Get survey question report [intent=direct_read availability=implemented operation=mailchimp.get_reporting_surveys_id_questions_id]; flags: --survey-id, --question-id, --page, --page-cursor
    reporting-surveys-questions-answers list - List answers for question as ETL records. [intent=etl availability=implemented stream=reporting_surveys_questions_answers]
    reporting-surveys-responses list - List survey responses as ETL records. [intent=etl availability=implemented stream=reporting_surveys_responses]
    reporting surveys responses get - Get survey response [intent=direct_read availability=implemented operation=mailchimp.get_reporting_surveys_id_responses_id]; flags: --survey-id, --response-id, --page, --page-cursor
    reports-abuse-reports list - List abuse reports as ETL records. [intent=etl availability=implemented stream=reports_abuse_reports]
    reports-advice list - List campaign feedback as ETL records. [intent=etl availability=implemented stream=reports_advice]
    reports-click-details-urls-clicked list - List campaign details as ETL records. [intent=etl availability=implemented stream=reports_click_details_urls_clicked]
    reports-click-details-members list - List clicked link subscribers as ETL records. [intent=etl availability=implemented stream=reports_click_details_members]
    reports-domain-performance-domains list - List domain performance stats as ETL records. [intent=etl availability=implemented stream=reports_domain_performance_domains]
    reports-ecommerce-product-activity-products list - List campaign product activity as ETL records. [intent=etl availability=implemented stream=reports_ecommerce_product_activity_products]
    reports-eepurl-referrers list - List EepURL activity as ETL records. [intent=etl availability=implemented stream=reports_eepurl_referrers]
    reports-email-activity-emails list - List email activity as ETL records. [intent=etl availability=implemented stream=reports_email_activity_emails]
    reports-locations list - List top open activities as ETL records. [intent=etl availability=implemented stream=reports_locations]
    reports-open-details-members list - List campaign open details as ETL records. [intent=etl availability=implemented stream=reports_open_details_members]
    reports-sent-to list - List campaign recipients as ETL records. [intent=etl availability=implemented stream=reports_sent_to]
    reports-sub-reports list - List child campaign reports as ETL records. [intent=etl availability=implemented stream=reports_sub_reports]
    reports-unsubscribed-unsubscribes list - List unsubscribed members as ETL records. [intent=etl availability=implemented stream=reports_unsubscribed_unsubscribes]
    sms-campaigns list - List SMS campaigns as ETL records. [intent=etl availability=implemented stream=sms_campaigns]
    sms-campaigns get - Get SMS campaign info [intent=direct_read availability=implemented operation=mailchimp.get_sms_campaigns_id]; flags: --sms-campaign-id, --page, --page-cursor
    sms-campaigns content get - Get SMS campaign content [intent=direct_read availability=implemented operation=mailchimp.get_sms_campaigns_id_content]; flags: --sms-campaign-id, --page, --page-cursor
    template-folders-folders list - List template folders as ETL records. [intent=etl availability=implemented stream=template_folders_folders]
    template-folders get - Get template folder [intent=direct_read availability=implemented operation=mailchimp.get_template_folders_id]; flags: --folder-id, --page, --page-cursor
    verified-domains-domains list - List sending domains as ETL records. [intent=etl availability=implemented stream=verified_domains_domains]
    verified-domains get - Get domain info [intent=direct_read availability=implemented operation=mailchimp.get_verified_domain]; flags: --domain-name, --page, --page-cursor
  Help topics:
    auth - Use access_token or api_key from environment/stdin; never paste secrets into prompts or shell history.
    safety - Writes are reverse-ETL only: plan, preview, explicit approval, execute; destructive actions require confirmation.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect mailchimp

  # Inspect as structured JSON
  pm connectors inspect mailchimp --json

AGENT WORKFLOW
  - Run pm connectors inspect mailchimp before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
