# Overview

Reads and writes Mantle Core API resources through the heymantle.com REST API.

Readable streams: `customers`, `subscriptions`, `affiliate_commissions`, `affiliate_commissions_id`,
`affiliate_payouts`, `affiliate_payouts_id`, `affiliate_programs`, `affiliate_programs_id`,
`affiliate_referrals`, `affiliate_referrals_id`, `affiliates`, `affiliates_id`, `agents`,
`ai_agents_agent_id_runs_run_id`, `api_core_v1_metrics_active_installs`,
`api_core_v1_metrics_active_subscriptions`, `api_core_v1_metrics_arpu`, `api_core_v1_metrics_arr`,
`api_core_v1_metrics_logo_churn`, `api_core_v1_metrics_mrr`, `api_core_v1_metrics_net_installs`,
`api_core_v1_metrics_net_revenue`, `api_core_v1_metrics_net_revenue_retention`,
`api_core_v1_metrics_payout`, `api_core_v1_metrics_predicted_ltv`,
`api_core_v1_metrics_revenue_churn`, `api_core_v1_metrics_revenue_retention`,
`api_core_v1_metrics_subscription_churn`, `api_core_v1_metrics_usage_event`,
`api_core_v1_metrics_usage_metric`, `apps`, `apps_app_id_checklists`,
`apps_app_id_checklists_checklist_id`, `apps_app_id_plans_features`,
`apps_app_id_plans_features_feature_id`, `apps_id`, `apps_id_app_events`,
`apps_id_app_events_app_event_id`, `apps_id_plans`, `apps_id_plans_plan_id`, `apps_id_reviews`,
`apps_id_reviews_review_id`, `apps_id_skills_skill_id`, `apps_id_usage_metrics`,
`apps_id_usage_metrics_usage_metric_id`, `assistant_conversations_id`, `channels`, `charges`,
`companies`, `companies_id`, `contacts`, `contacts_id`, `custom_data`, `customer_segments`,
`customer_segments_id`, `customers_custom_fields`, `customers_custom_fields_id`, `customers_id`,
`customers_id_account_owners`, `customers_id_timeline`, `deal_activities`, `deal_activities_id`,
`deal_flows`, `deal_flows_id`, `deals`, `deals_id`, `deals_id_events`, `deals_id_timeline`,
`docs_collections`, `docs_groups`, `docs_pages`, `docs_pages_generate_status_job_key`,
`docs_pages_page_id`, `docs_repositories`, `docs_repositories_id`, `docs_sites`, `docs_sites_id`,
`docs_sites_id_redirects`, `docs_sites_id_repositories`, `docs_tree`, `email_campaigns`,
`email_campaigns_id`, `email_campaigns_id_preview`, `email_deliveries`, `email_deliveries_id`,
`email_layouts`, `email_senders`, `email_unsubscribe_groups`, `email_unsubscribe_groups_id_members`,
`entities`, `flow_extensions_actions`, `flows`, `flows_id`, `journal_entries`, `journal_entries_id`,
`lists`, `lists_id`, `meetings`, `meetings_id`, `meetings_id_permissions`,
`meetings_id_recording_url`, `meetings_id_transcribe`, `metrics_sales`, `notification_preferences`,
`organization`, `subscriptions_id`, `synced_emails`, `synced_emails_id`,
`synced_emails_id_messages`, `tasks`, `tasks_id`, `tasks_id_comments`, `tasks_id_todo_items`,
`tasks_id_todo_items_item_id`, `tickets`, `tickets_id`, `tickets_id_events`, `tickets_id_loops`,
`tickets_id_loops_loop_id`, `tickets_id_messages`, `tickets_id_messages_message_id`,
`tickets_saved_filters`, `tickets_saved_filters_filter_id`, `tickets_saved_replies`,
`tickets_saved_replies_reply_id`, `timeline_comments`, `timeline_comments_id`, `transactions`,
`transactions_id`, `usage_events`, `users`, `users_id`, `webhooks`.

Write actions: `create_agents`, `create_ai_agents_agent_id_runs`, `create_apps_app_id_checklists`,
`create_apps_app_id_plans_features`, `create_apps_id_app_events`, `create_apps_id_plans`,
`create_apps_id_usage_metrics`, `create_attachments`, `create_channels`, `create_companies`,
`create_customers`, `create_customers_custom_fields`, `create_deal_activities`, `create_deal_flows`,
`create_deals`, `create_deals_id_events`, `create_docs_collections`, `create_docs_groups`,
`create_docs_pages`, `create_docs_sites`, `create_docs_sites_id_redirects`,
`create_email_campaigns`, `create_flow_extensions_actions`, `create_flows`,
`create_journal_entries`, `create_lists`, `create_meetings`, `create_meetings_id_transcribe_upload`,
`create_tasks`, `create_tasks_id_comments`, `create_tasks_id_todo_items`, `create_tickets`,
`create_tickets_id_events`, `create_tickets_id_messages`, `create_tickets_saved_filters`,
`create_tickets_saved_replies`, `create_timeline_comments`, `create_usage_events`,
`create_webhooks`, `delete_apps_app_id_checklists_checklist_id`,
`delete_apps_app_id_plans_features_feature_id`, `delete_apps_id_usage_metrics_usage_metric_id`,
`delete_companies_id`, `delete_contacts_id`, `delete_customers_custom_fields_id`,
`delete_customers_id_account_owners_owner_id`, `delete_deal_activities_id`, `delete_deal_flows_id`,
`delete_deals_id`, `delete_docs_collections_collection_id`, `delete_docs_groups_group_id`,
`delete_docs_pages_page_id`, `delete_docs_pages_page_id_archive`,
`delete_docs_pages_page_id_publish`, `delete_docs_sites_id`,
`delete_docs_sites_id_redirects_redirect_id`, `delete_docs_sites_id_repositories`,
`delete_email_campaigns_id`, `delete_email_unsubscribe_groups_id_members`,
`delete_email_unsubscribe_groups_id_members_member_id`, `delete_flow_extensions_actions_id`,
`delete_flow_extensions_triggers_handle`, `delete_flows_id`, `delete_journal_entries_id`,
`delete_lists_id`, `delete_meetings_id`, `delete_meetings_id_permissions`,
`delete_synced_emails_id`, `delete_tasks_id`, `delete_tasks_id_comments_comment_id`,
`delete_tasks_id_todo_items_item_id`, `delete_tickets_saved_filters_filter_id`,
`delete_tickets_saved_replies_reply_id`, `delete_timeline_comments_id`, `delete_webhooks_id`,
`execute_affiliates_id_add_tags`, `execute_affiliates_id_remove_tags`, `execute_apps_id_analyze`,
`execute_apps_id_skills_skill_id`, `execute_contacts_id_add_tags`, and 68 more.

Service API documentation: https://coreapi.heymantle.dev/.

## Auth setup

Connection fields:

- `agent_id` (optional, string); Mantle path parameter agentId.
- `api_key` (required, secret, string); Mantle API key for Bearer authentication. Never logged.
- `app_event_id` (optional, string); Mantle path parameter appEventId.
- `app_id` (optional, string); Mantle path parameter appId.
- `base_url` (optional, string); default `https://api.heymantle.com`; format `uri`; Mantle API base
  URL override for tests or proxies.
- `checklist_id` (optional, string); Mantle path parameter checklistId.
- `collection_id` (optional, string); Mantle required query parameter collectionId.
- `event_name` (optional, string); Mantle required query parameter eventName.
- `feature_id` (optional, string); Mantle path parameter featureId.
- `filter_id` (optional, string); Mantle path parameter filter_id.
- `id` (optional, string); Mantle path parameter id.
- `item_id` (optional, string); Mantle path parameter itemId.
- `job_key` (optional, string); Mantle path parameter jobKey.
- `loop_id` (optional, string); Mantle path parameter loopId.
- `message_id` (optional, string); Mantle path parameter messageId.
- `page_id` (optional, string); Mantle path parameter page_id.
- `plan_id` (optional, string); Mantle path parameter planId.
- `reply_id` (optional, string); Mantle path parameter reply_id.
- `repository_id` (optional, string); Mantle required query parameter repositoryId.
- `resource_id` (optional, string); Mantle required query parameter resourceId.
- `resource_type` (optional, string); Mantle required query parameter resourceType.
- `review_id` (optional, string); Mantle path parameter reviewId.
- `run_id` (optional, string); Mantle path parameter runId.
- `skill_id` (optional, string); Mantle path parameter skillId.
- `usage_metric_id` (optional, string); Mantle required query parameter usageMetricId.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.heymantle.com`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/customers`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `customers`, `subscriptions`, `affiliate_commissions`,
`affiliate_payouts`, `affiliate_programs`, `affiliate_referrals`, `affiliates`,
`apps_id_app_events`, `apps_id_plans`, `apps_id_reviews`, `companies`, `contacts`,
`customers_id_timeline`, `deal_activities`, `deal_flows`, `deals`, `email_campaigns`,
`email_deliveries`, `entities`, `flows`, `lists`, `meetings`, `synced_emails`, `tasks`, `tickets`,
`tickets_id_events`, `tickets_id_messages`, `timeline_comments`, `transactions`, `usage_events`;
none: `affiliate_commissions_id`, `affiliate_payouts_id`, `affiliate_programs_id`,
`affiliate_referrals_id`, `affiliates_id`, `agents`, `ai_agents_agent_id_runs_run_id`,
`api_core_v1_metrics_active_installs`, `api_core_v1_metrics_active_subscriptions`,
`api_core_v1_metrics_arpu`, `api_core_v1_metrics_arr`, `api_core_v1_metrics_logo_churn`,
`api_core_v1_metrics_mrr`, `api_core_v1_metrics_net_installs`, `api_core_v1_metrics_net_revenue`,
`api_core_v1_metrics_net_revenue_retention`, `api_core_v1_metrics_payout`,
`api_core_v1_metrics_predicted_ltv`, `api_core_v1_metrics_revenue_churn`,
`api_core_v1_metrics_revenue_retention`, `api_core_v1_metrics_subscription_churn`,
`api_core_v1_metrics_usage_event`, `api_core_v1_metrics_usage_metric`, `apps`,
`apps_app_id_checklists`, `apps_app_id_checklists_checklist_id`, `apps_app_id_plans_features`,
`apps_app_id_plans_features_feature_id`, `apps_id`, `apps_id_app_events_app_event_id`,
`apps_id_plans_plan_id`, `apps_id_reviews_review_id`, `apps_id_skills_skill_id`,
`apps_id_usage_metrics`, `apps_id_usage_metrics_usage_metric_id`, `assistant_conversations_id`,
`channels`, `charges`, `companies_id`, `contacts_id`, `custom_data`, `customer_segments`,
`customer_segments_id`, `customers_custom_fields`, `customers_custom_fields_id`, `customers_id`,
`customers_id_account_owners`, `deal_activities_id`, `deal_flows_id`, `deals_id`, `deals_id_events`,
`deals_id_timeline`, `docs_collections`, `docs_groups`, `docs_pages`,
`docs_pages_generate_status_job_key`, `docs_pages_page_id`, `docs_repositories`,
`docs_repositories_id`, `docs_sites`, and 43 more.

- `customers`: GET `/v1/customers` - records path `customers`; query `take`=`500`; cursor
  pagination; cursor parameter `cursor`; next token from `cursor`; stop flag `hasNextPage`.
- `subscriptions`: GET `/v1/subscriptions` - records path `subscriptions`; query `take`=`500`;
  cursor pagination; cursor parameter `cursor`; next token from `cursor`; stop flag `hasNextPage`.
- `affiliate_commissions`: GET `/v1/affiliate_commissions` - records path `affiliateCommissions`;
  cursor pagination; cursor parameter `cursor`; next token from `cursor`; stop flag `hasNextPage`;
  emits passthrough records.
- `affiliate_commissions_id`: GET `/v1/affiliate_commissions/{{ config.id }}` - records path
  `affiliateCommission`; emits passthrough records.
- `affiliate_payouts`: GET `/v1/affiliate_payouts` - records path `affiliatePayouts`; cursor
  pagination; cursor parameter `cursor`; next token from `cursor`; stop flag `hasNextPage`; emits
  passthrough records.
- `affiliate_payouts_id`: GET `/v1/affiliate_payouts/{{ config.id }}` - records path
  `affiliatePayout`; emits passthrough records.
- `affiliate_programs`: GET `/v1/affiliate_programs` - records path `affiliatePrograms`; cursor
  pagination; cursor parameter `cursor`; next token from `cursor`; stop flag `hasNextPage`; emits
  passthrough records.
- `affiliate_programs_id`: GET `/v1/affiliate_programs/{{ config.id }}` - records path
  `affiliateProgram`; emits passthrough records.
- `affiliate_referrals`: GET `/v1/affiliate_referrals` - records path `affiliateReferrals`; cursor
  pagination; cursor parameter `cursor`; next token from `cursor`; stop flag `hasNextPage`; emits
  passthrough records.
- `affiliate_referrals_id`: GET `/v1/affiliate_referrals/{{ config.id }}` - records path
  `affiliateReferral`; emits passthrough records.
- `affiliates`: GET `/v1/affiliates` - records path `affiliates`; cursor pagination; cursor
  parameter `cursor`; next token from `nextCursor`; stop flag `hasNextPage`; emits passthrough
  records.
- `affiliates_id`: GET `/v1/affiliates/{{ config.id }}` - records at response root; emits
  passthrough records.
- `agents`: GET `/v1/agents` - records path `agents`; emits passthrough records.
- `ai_agents_agent_id_runs_run_id`: GET `/v1/ai/agents/{{ config.agent_id }}/runs/{{ config.run_id
  }}` - records path `run`; emits passthrough records.
- `api_core_v1_metrics_active_installs`: GET `/v1/api/core/v1/metrics/activeInstalls` - records at
  response root; emits passthrough records.
- `api_core_v1_metrics_active_subscriptions`: GET `/v1/api/core/v1/metrics/activeSubscriptions` -
  records at response root; emits passthrough records.
- `api_core_v1_metrics_arpu`: GET `/v1/api/core/v1/metrics/arpu` - records at response root; emits
  passthrough records.
- `api_core_v1_metrics_arr`: GET `/v1/api/core/v1/metrics/arr` - records at response root; emits
  passthrough records.
- `api_core_v1_metrics_logo_churn`: GET `/v1/api/core/v1/metrics/logoChurn` - records at response
  root; emits passthrough records.
- `api_core_v1_metrics_mrr`: GET `/v1/api/core/v1/metrics/mrr` - records at response root; emits
  passthrough records.
- `api_core_v1_metrics_net_installs`: GET `/v1/api/core/v1/metrics/netInstalls` - records at
  response root; emits passthrough records.
- `api_core_v1_metrics_net_revenue`: GET `/v1/api/core/v1/metrics/netRevenue` - records at response
  root; emits passthrough records.
- `api_core_v1_metrics_net_revenue_retention`: GET `/v1/api/core/v1/metrics/netRevenueRetention` -
  records at response root; emits passthrough records.
- `api_core_v1_metrics_payout`: GET `/v1/api/core/v1/metrics/payout` - records at response root;
  emits passthrough records.
- `api_core_v1_metrics_predicted_ltv`: GET `/v1/api/core/v1/metrics/predictedLtv` - records at
  response root; emits passthrough records.
- `api_core_v1_metrics_revenue_churn`: GET `/v1/api/core/v1/metrics/revenueChurn` - records at
  response root; emits passthrough records.
- `api_core_v1_metrics_revenue_retention`: GET `/v1/api/core/v1/metrics/revenueRetention` - records
  at response root; emits passthrough records.
- `api_core_v1_metrics_subscription_churn`: GET `/v1/api/core/v1/metrics/subscriptionChurn` -
  records at response root; emits passthrough records.
- `api_core_v1_metrics_usage_event`: GET `/v1/api/core/v1/metrics/usageEvent` - records at response
  root; query `eventName`=`{{ config.event_name }}`; emits passthrough records.
- `api_core_v1_metrics_usage_metric`: GET `/v1/api/core/v1/metrics/usageMetric` - records at
  response root; query `usageMetricId`=`{{ config.usage_metric_id }}`; emits passthrough records.
- `apps`: GET `/v1/apps` - records path `apps`; emits passthrough records.
- `apps_app_id_checklists`: GET `/v1/apps/{{ config.app_id }}/checklists` - records path
  `checklists`; emits passthrough records.
- `apps_app_id_checklists_checklist_id`: GET `/v1/apps/{{ config.app_id }}/checklists/{{
  config.checklist_id }}` - records path `checklist`; emits passthrough records.
- `apps_app_id_plans_features`: GET `/v1/apps/{{ config.app_id }}/plans/features` - records path
  `features`; emits passthrough records.
- `apps_app_id_plans_features_feature_id`: GET `/v1/apps/{{ config.app_id }}/plans/features/{{
  config.feature_id }}` - records path `feature`; emits passthrough records.
- `apps_id`: GET `/v1/apps/{{ config.id }}` - records path `app`; emits passthrough records.
- `apps_id_app_events`: GET `/v1/apps/{{ config.id }}/app_events` - records path `appEvents`; cursor
  pagination; cursor parameter `cursor`; next token from `cursor`; stop flag `hasNextPage`; emits
  passthrough records.
- `apps_id_app_events_app_event_id`: GET `/v1/apps/{{ config.id }}/app_events/{{ config.app_event_id
  }}` - records path `appEvent`; emits passthrough records.
- `apps_id_plans`: GET `/v1/apps/{{ config.id }}/plans` - records path `plans`; query `take`=`100`;
  cursor pagination; cursor parameter `cursor`; next token from `nextCursor`; stop flag
  `hasNextPage`; emits passthrough records.
- `apps_id_plans_plan_id`: GET `/v1/apps/{{ config.id }}/plans/{{ config.plan_id }}` - records path
  `plan`; emits passthrough records.
- `apps_id_reviews`: GET `/v1/apps/{{ config.id }}/reviews` - records path `reviews`; cursor
  pagination; cursor parameter `cursor`; next token from `nextCursor`; stop flag `hasNextPage`;
  emits passthrough records.
- `apps_id_reviews_review_id`: GET `/v1/apps/{{ config.id }}/reviews/{{ config.review_id }}` -
  records path `review`; emits passthrough records.
- `apps_id_skills_skill_id`: GET `/v1/apps/{{ config.id }}/skills/{{ config.skill_id }}` - records
  at response root; emits passthrough records.
- `apps_id_usage_metrics`: GET `/v1/apps/{{ config.id }}/usage_metrics` - records path
  `usageMetrics`; emits passthrough records.
- `apps_id_usage_metrics_usage_metric_id`: GET `/v1/apps/{{ config.id }}/usage_metrics/{{
  config.usage_metric_id }}` - records path `usageMetric`; emits passthrough records.
- `assistant_conversations_id`: GET `/v1/assistant/conversations/{{ config.id }}` - records path
  `conversation`; emits passthrough records.
- `channels`: GET `/v1/channels` - records path `channels`; emits passthrough records.
- `charges`: GET `/v1/charges` - records at response root; emits passthrough records.
- `companies`: GET `/v1/companies` - records path `companies`; cursor pagination; cursor parameter
  `cursor`; next token from `nextCursor`; stop flag `hasNextPage`; emits passthrough records.
- `companies_id`: GET `/v1/companies/{{ config.id }}` - records path `company`; emits passthrough
  records.
- `contacts`: GET `/v1/contacts` - records path `contacts`; cursor pagination; cursor parameter
  `cursor`; next token from `cursor`; stop flag `hasNextPage`; emits passthrough records.
- `contacts_id`: GET `/v1/contacts/{{ config.id }}` - records path `contact`; emits passthrough
  records.
- `custom_data`: GET `/v1/custom_data` - records at response root; query `resourceId`=`{{
  config.resource_id }}`; `resourceType`=`{{ config.resource_type }}`; emits passthrough records.
- `customer_segments`: GET `/v1/customer_segments` - records path `customerSegments`; emits
  passthrough records.
- `customer_segments_id`: GET `/v1/customer_segments/{{ config.id }}` - records path `customField`;
  emits passthrough records.
- `customers_custom_fields`: GET `/v1/customers/custom_fields` - records path `customFields`; emits
  passthrough records.
- `customers_custom_fields_id`: GET `/v1/customers/custom_fields/{{ config.id }}` - records path
  `customField`; emits passthrough records.
- `customers_id`: GET `/v1/customers/{{ config.id }}` - records path `customer`; emits passthrough
  records.
- `customers_id_account_owners`: GET `/v1/customers/{{ config.id }}/account_owners` - records at
  response root; emits passthrough records.
- `customers_id_timeline`: GET `/v1/customers/{{ config.id }}/timeline` - records path `events`;
  cursor pagination; cursor parameter `cursor`; next token from `cursor`; stop flag `hasNextPage`;
  emits passthrough records.
- `deal_activities`: GET `/v1/deal_activities` - records path `dealActivities`; cursor pagination;
  cursor parameter `cursor`; next token from `cursor`; stop flag `hasNextPage`; emits passthrough
  records.
- `deal_activities_id`: GET `/v1/deal_activities/{{ config.id }}` - records path `dealActivity`;
  emits passthrough records.
- `deal_flows`: GET `/v1/deal_flows` - records path `dealFlows`; cursor pagination; cursor parameter
  `cursor`; next token from `cursor`; stop flag `hasNextPage`; emits passthrough records.
- `deal_flows_id`: GET `/v1/deal_flows/{{ config.id }}` - records path `dealFlow`; emits passthrough
  records.
- `deals`: GET `/v1/deals` - records path `deals`; cursor pagination; cursor parameter `cursor`;
  next token from `cursor`; stop flag `hasNextPage`; emits passthrough records.
- `deals_id`: GET `/v1/deals/{{ config.id }}` - records path `deal`; emits passthrough records.
- `deals_id_events`: GET `/v1/deals/{{ config.id }}/events` - records path `events`; query
  `take`=`100`; emits passthrough records.
- `deals_id_timeline`: GET `/v1/deals/{{ config.id }}/timeline` - records path `events`; query
  `take`=`100`; emits passthrough records.
- `docs_collections`: GET `/v1/docs/collections` - records path `collections`; query
  `repositoryId`=`{{ config.repository_id }}`; emits passthrough records.
- `docs_groups`: GET `/v1/docs/groups` - records path `groups`; query `collectionId`=`{{
  config.collection_id }}`; emits passthrough records.
- `docs_pages`: GET `/v1/docs/pages` - records path `pages`; emits passthrough records.
- `docs_pages_generate_status_job_key`: GET `/v1/docs/pages/generate/status/{{ config.job_key }}` -
  records at response root; emits passthrough records.
- `docs_pages_page_id`: GET `/v1/docs/pages/{{ config.page_id }}` - records at response root; emits
  passthrough records.
- `docs_repositories`: GET `/v1/docs/repositories` - records path `repositories`; emits passthrough
  records.
- `docs_repositories_id`: GET `/v1/docs/repositories/{{ config.id }}` - records path `repository`;
  emits passthrough records.
- `docs_sites`: GET `/v1/docs/sites` - records path `sites`; emits passthrough records.
- `docs_sites_id`: GET `/v1/docs/sites/{{ config.id }}` - records path `site`; emits passthrough
  records.
- `docs_sites_id_redirects`: GET `/v1/docs/sites/{{ config.id }}/redirects` - records path
  `redirects`; emits passthrough records.
- `docs_sites_id_repositories`: GET `/v1/docs/sites/{{ config.id }}/repositories` - records path
  `attachments`; emits passthrough records.
- `docs_tree`: GET `/v1/docs/tree` - records at response root; query `repositoryId`=`{{
  config.repository_id }}`; emits passthrough records.
- `email_campaigns`: GET `/v1/email/campaigns` - records path `campaigns`; cursor pagination; cursor
  parameter `cursor`; next token from `cursor`; stop flag `hasNextPage`; emits passthrough records.
- `email_campaigns_id`: GET `/v1/email/campaigns/{{ config.id }}` - records path `campaign`; emits
  passthrough records.
- `email_campaigns_id_preview`: GET `/v1/email/campaigns/{{ config.id }}/preview` - records at
  response root; emits passthrough records.
- `email_deliveries`: GET `/v1/email/deliveries` - records path `deliveries`; cursor pagination;
  cursor parameter `cursor`; next token from `cursor`; stop flag `hasNextPage`; emits passthrough
  records.
- `email_deliveries_id`: GET `/v1/email/deliveries/{{ config.id }}` - records path `delivery`; emits
  passthrough records.
- `email_layouts`: GET `/v1/email/layouts` - records path `layouts`; emits passthrough records.
- `email_senders`: GET `/v1/email/senders` - records path `senders`; emits passthrough records.
- `email_unsubscribe_groups`: GET `/v1/email/unsubscribe_groups` - records path `unsubscribeGroups`;
  emits passthrough records.
- `email_unsubscribe_groups_id_members`: GET `/v1/email/unsubscribe_groups/{{ config.id }}/members`
  - records path `members`; emits passthrough records.
- `entities`: GET `/v1/entities` - records path `entities`; cursor pagination; cursor parameter
  `cursor`; next token from `cursor`; stop flag `hasNextPage`; emits passthrough records.
- `flow_extensions_actions`: GET `/v1/flow/extensions/actions` - records path `actions`; emits
  passthrough records.
- `flows`: GET `/v1/flows` - records path `flows`; cursor pagination; cursor parameter `cursor`;
  next token from `cursor`; stop flag `hasNextPage`; emits passthrough records.
- `flows_id`: GET `/v1/flows/{{ config.id }}` - records path `flow`; emits passthrough records.
- `journal_entries`: GET `/v1/journal_entries` - records path `entries`; emits passthrough records.
- `journal_entries_id`: GET `/v1/journal_entries/{{ config.id }}` - records path `entry`; emits
  passthrough records.
- `lists`: GET `/v1/lists` - records path `lists`; cursor pagination; cursor parameter `cursor`;
  next token from `cursor`; stop flag `hasNextPage`; emits passthrough records.
- `lists_id`: GET `/v1/lists/{{ config.id }}` - records path `entities`; emits passthrough records.
- `meetings`: GET `/v1/meetings` - records path `meetings`; cursor pagination; cursor parameter
  `cursor`; next token from `cursor`; stop flag `hasNextPage`; emits passthrough records.
- `meetings_id`: GET `/v1/meetings/{{ config.id }}` - records at response root; emits passthrough
  records.
- `meetings_id_permissions`: GET `/v1/meetings/{{ config.id }}/permissions` - records path
  `permissions`; emits passthrough records.
- `meetings_id_recording_url`: GET `/v1/meetings/{{ config.id }}/recording-url` - records at
  response root; emits passthrough records.
- `meetings_id_transcribe`: GET `/v1/meetings/{{ config.id }}/transcribe` - records path
  `transcript`; emits passthrough records.
- `metrics_sales`: GET `/v1/metrics/sales` - records at response root; emits passthrough records.
- `notification_preferences`: GET `/v1/notification_preferences` - records at response root; emits
  passthrough records.
- `organization`: GET `/v1/organization` - records at response root; emits passthrough records.
- `subscriptions_id`: GET `/v1/subscriptions/{{ config.id }}` - records path `subscription`; emits
  passthrough records.
- `synced_emails`: GET `/v1/synced_emails` - records path `syncedEmails`; cursor pagination; cursor
  parameter `cursor`; next token from `cursor`; stop flag `hasNextPage`; emits passthrough records.
- `synced_emails_id`: GET `/v1/synced_emails/{{ config.id }}` - records at response root; emits
  passthrough records.
- `synced_emails_id_messages`: GET `/v1/synced_emails/{{ config.id }}/messages` - records at
  response root; emits passthrough records.
- `tasks`: GET `/v1/tasks` - records path `tasks`; query `take`=`100`; cursor pagination; cursor
  parameter `cursor`; next token from `cursor`; stop flag `hasNextPage`; emits passthrough records.
- `tasks_id`: GET `/v1/tasks/{{ config.id }}` - records path `task`; emits passthrough records.
- `tasks_id_comments`: GET `/v1/tasks/{{ config.id }}/comments` - records path `comments`; emits
  passthrough records.
- `tasks_id_todo_items`: GET `/v1/tasks/{{ config.id }}/todo-items` - records path `items`; emits
  passthrough records.
- `tasks_id_todo_items_item_id`: GET `/v1/tasks/{{ config.id }}/todo-items/{{ config.item_id }}` -
  records path `item`; emits passthrough records.
- `tickets`: GET `/v1/tickets` - records path `tickets`; cursor pagination; cursor parameter
  `cursor`; next token from `cursor`; stop flag `hasNextPage`; emits passthrough records.
- `tickets_id`: GET `/v1/tickets/{{ config.id }}` - records path `ticket`; emits passthrough
  records.
- `tickets_id_events`: GET `/v1/tickets/{{ config.id }}/events` - records path `events`; cursor
  pagination; cursor parameter `cursor`; next token from `cursor`; stop flag `hasNextPage`; emits
  passthrough records.
- `tickets_id_loops`: GET `/v1/tickets/{{ config.id }}/loops` - records path `loops`; emits
  passthrough records.
- `tickets_id_loops_loop_id`: GET `/v1/tickets/{{ config.id }}/loops/{{ config.loop_id }}` - records
  path `loop`; emits passthrough records.
- `tickets_id_messages`: GET `/v1/tickets/{{ config.id }}/messages` - records path `messages`;
  cursor pagination; cursor parameter `cursor`; next token from `cursor`; stop flag `hasNextPage`;
  emits passthrough records.
- `tickets_id_messages_message_id`: GET `/v1/tickets/{{ config.id }}/messages/{{ config.message_id
  }}` - records path `message`; emits passthrough records.
- `tickets_saved_filters`: GET `/v1/tickets/saved_filters` - records path `savedTicketFilters`;
  emits passthrough records.
- `tickets_saved_filters_filter_id`: GET `/v1/tickets/saved_filters/{{ config.filter_id }}` -
  records path `savedTicketFilter`; emits passthrough records.
- `tickets_saved_replies`: GET `/v1/tickets/saved_replies` - records path `savedReplies`; emits
  passthrough records.
- `tickets_saved_replies_reply_id`: GET `/v1/tickets/saved_replies/{{ config.reply_id }}` - records
  path `savedReply`; emits passthrough records.
- `timeline_comments`: GET `/v1/timeline_comments` - records path `timelineComments`; cursor
  pagination; cursor parameter `cursor`; next token from `cursor`; stop flag `hasNextPage`; emits
  passthrough records.
- `timeline_comments_id`: GET `/v1/timeline_comments/{{ config.id }}` - records path
  `timelineComment`; emits passthrough records.
- `transactions`: GET `/v1/transactions` - records path `transactions`; cursor pagination; cursor
  parameter `cursor`; next token from `cursor`; stop flag `hasNextPage`; emits passthrough records.
- `transactions_id`: GET `/v1/transactions/{{ config.id }}` - records at response root; emits
  passthrough records.
- `usage_events`: GET `/v1/usage_events` - records path `events`; cursor pagination; cursor
  parameter `cursor`; next token from `cursor`; stop flag `hasNextPage`; emits passthrough records.
- `users`: GET `/v1/users` - records path `users`; emits passthrough records.
- `users_id`: GET `/v1/users/{{ config.id }}` - records path `user`; emits passthrough records.
- `webhooks`: GET `/v1/webhooks` - records path `webhooks`; emits passthrough records.

## Write actions & risks

Overall write risk: external Mantle API mutation of customer, billing, CRM, docs, email, helpdesk,
webhook, and workflow resources; approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_agents`: POST `/v1/agents` - kind `create`; body type `json`; required record fields
  `email`; accepted fields `email`, `name`; risk: medium: external Mantle mutation; approval
  required.
- `create_ai_agents_agent_id_runs`: POST `/v1/ai/agents/{{ record.agent_id }}/runs` - kind `create`;
  body type `json`; path fields `agent_id`; required record fields `agent_id`, `prompt`; accepted
  fields `agent_id`, `imageUrls`, `options`, `outputSchema`, `prompt`, `sourceImageUrl`; risk:
  medium: external Mantle side effect; approval required.
- `create_apps_app_id_checklists`: POST `/v1/apps/{{ record.app_id }}/checklists` - kind `create`;
  body type `json`; path fields `app_id`; required record fields `app_id`, `name`; accepted fields
  `app_id`, `customerSegmentId`, `description`, `handle`, `name`, `status`, `steps`, `title`; risk:
  medium: external Mantle mutation; approval required.
- `create_apps_app_id_plans_features`: POST `/v1/apps/{{ record.app_id }}/plans/features` - kind
  `create`; body type `json`; path fields `app_id`; required record fields `app_id`; accepted fields
  `allowedValues`, `app_id`, `defaultValue`, `description`, `key`, `name`, `type`, `usageMetricId`,
  `visible`; risk: medium: external Mantle mutation; approval required.
- `create_apps_id_app_events`: POST `/v1/apps/{{ record.id }}/app_events` - kind `create`; body type
  `json`; path fields `id`; required record fields `id`, `type`, `customerId`; accepted fields
  `appReviewId`, `charge`, `customerId`, `externalId`, `id`, `invoiceId`, `metadata`, `occurredAt`,
  `plan`, `platformEventId`, `subscription`, `transactionId`, `type`; risk: medium: external Mantle
  mutation; approval required.
- `create_apps_id_plans`: POST `/v1/apps/{{ record.id }}/plans` - kind `create`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `amount`, `autoUpgradeToPlanId`,
  `currencyCode`, `customFields`, `customer`, `customerExcludeTags`, `customerTags`, `description`,
  `features`, `flexBilling`, `flexBillingTerms`, `id`, `interval`, `name`, `onUsageLimitReached`,
  `planUsageCharges`, `public`, `rollOverPendingUsage`, and 5 more; risk: medium: external Mantle
  mutation; approval required.
- `create_apps_id_usage_metrics`: POST `/v1/apps/{{ record.id }}/usage_metrics` - kind `create`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `calculation`,
  `eventName`, `id`, `name`, `params`; risk: medium: external Mantle mutation; approval required.
- `create_attachments`: POST `/v1/attachments` - kind `create`; body type `json`; required record
  fields `filename`; accepted fields `contentType`, `filename`; risk: medium: external Mantle
  mutation; approval required.
- `create_channels`: POST `/v1/channels` - kind `create`; body type `json`; required record fields
  `type`, `name`; accepted fields `name`, `type`; risk: medium: external Mantle mutation; approval
  required.
- `create_companies`: POST `/v1/companies` - kind `create`; body type `json`; accepted fields
  `name`, `parentCustomerId`; risk: medium: external Mantle mutation; approval required.
- `create_customers`: POST `/v1/customers` - kind `create`; body type `json`; accepted fields
  `appInstallations`, `companyId`, `countryCode`, `customFields`, `description`, `domain`, `email`,
  `industry`, `name`, `preferredCurrency`, `shopifyDomain`, `shopifyShopId`, `tags`, `url`; risk:
  medium: external Mantle mutation; approval required.
- `create_customers_custom_fields`: POST `/v1/customers/custom_fields` - kind `create`; body type
  `json`; accepted fields `appId`, `appLevel`, `defaultValue`, `filterable`, `name`, `options`,
  `private`, `showOnCustomerDetail`, `type`; risk: medium: external Mantle mutation; approval
  required.
- `create_deal_activities`: POST `/v1/deal_activities` - kind `create`; body type `json`; required
  record fields `name`; accepted fields `defaultWeight`, `description`, `icon`, `name`,
  `timelineDescriptionTemplate`, `timelineTitleTemplate`; risk: medium: external Mantle mutation;
  approval required.
- `create_deal_flows`: POST `/v1/deal_flows` - kind `create`; body type `json`; required record
  fields `name`, `dealStages`; accepted fields `dealStages`, `defaultAcquisitionChannel`,
  `defaultAcquisitionSource`, `defaultDealOwnerId`, `description`, `isDefaultDealFlow`, `name`;
  risk: medium: external Mantle side effect; approval required.
- `create_deals`: POST `/v1/deals` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `acquisitionChannel`, `acquisitionSource`, `affiliateId`, `amount`,
  `amountCurrencyCode`, `appId`, `closedAt`, `closingAt`, `companyId`, `contactIds`, `contacts`,
  `customer`, `customerId`, `dealFlowId`, `dealStageId`, `domain`, `firstInteractionAt`, `name`, and
  5 more; risk: medium: external Mantle mutation; approval required.
- `create_deals_id_events`: POST `/v1/deals/{{ record.id }}/events` - kind `create`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `appEventId`,
  `customActivityName`, `dealActivityId`, `dealStageId`, `id`, `notes`, `occurredAt`, `taskId`,
  `userId`; risk: medium: external Mantle mutation; approval required.
- `create_docs_collections`: POST `/v1/docs/collections` - kind `create`; body type `json`; required
  record fields `repositoryId`, `handle`, `title`; accepted fields `description`, `displayOrder`,
  `handle`, `icon`, `locale`, `repositoryId`, `status`, `title`; risk: medium: external Mantle
  mutation; approval required.
- `create_docs_groups`: POST `/v1/docs/groups` - kind `create`; body type `json`; required record
  fields `repositoryId`, `collectionId`, `handle`, `title`; accepted fields `collectionId`,
  `displayOrder`, `handle`, `icon`, `locale`, `repositoryId`, `title`; risk: medium: external Mantle
  mutation; approval required.
- `create_docs_pages`: POST `/v1/docs/pages` - kind `create`; body type `json`; required record
  fields `repositoryId`, `groupId`, `handle`, `title`; accepted fields `content`, `displayOrder`,
  `groupId`, `handle`, `locale`, `parentPageId`, `repositoryId`, `seoDescription`, `seoTitle`,
  `summary`, `title`; risk: medium: external Mantle mutation; approval required.
- `create_docs_sites`: POST `/v1/docs/sites` - kind `create`; body type `json`; required record
  fields `handle`, `title`; accepted fields `customDomain`, `defaultLocale`, `handle`,
  `repositoryIds`, `shortDescription`, `title`, `useCustomDomain`, `visibility`; risk: medium:
  external Mantle mutation; approval required.
- `create_docs_sites_id_redirects`: POST `/v1/docs/sites/{{ record.id }}/redirects` - kind `create`;
  body type `json`; path fields `id`; required record fields `id`, `redirects`; accepted fields
  `id`, `redirects`, `strict`; risk: medium: external Mantle mutation; approval required.
- `create_email_campaigns`: POST `/v1/email/campaigns` - kind `create`; body type `json`; required
  record fields `name`; accepted fields `appId`, `criteria`, `html`, `layoutId`, `name`,
  `plainText`, `previewText`, `senderId`, `subject`, `type`, `unsubscribeGroupId`; risk: medium:
  external Mantle side effect; approval required.
- `create_flow_extensions_actions`: POST `/v1/flow/extensions/actions` - kind `create`; body type
  `json`; accepted fields `description`, `handle`, `name`, `settingsSchema`, `url`; risk: medium:
  external Mantle side effect; approval required.
- `create_flows`: POST `/v1/flows` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `allowRepeatRuns`, `blockRepeatsTimeUnit`, `blockRepeatsTimeValue`, `name`; risk:
  medium: external Mantle side effect; approval required.
- `create_journal_entries`: POST `/v1/journal_entries` - kind `create`; body type `json`; required
  record fields `date`, `description`; accepted fields `appId`, `date`, `description`, `emoji`,
  `tags`, `title`, `url`; risk: medium: external Mantle mutation; approval required.
- `create_lists`: POST `/v1/lists` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `description`, `name`; risk: medium: external Mantle mutation; approval required.
- `create_meetings`: POST `/v1/meetings` - kind `create`; body type `json`; accepted fields
  `attendees`, `meetingData`, `transcript`; risk: medium: external Mantle mutation; approval
  required.
- `create_meetings_id_transcribe_upload`: POST `/v1/meetings/{{ record.id }}/transcribe/upload` -
  kind `create`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `filename`, `id`; risk: medium: external Mantle mutation; approval required.
- `create_tasks`: POST `/v1/tasks` - kind `create`; body type `json`; required record fields
  `title`; accepted fields `appInstallationId`, `assigneeId`, `contactId`, `customerId`,
  `dealActivityId`, `dealId`, `description`, `descriptionHtml`, `dueDate`, `priority`, `status`,
  `tags`, `title`, `todoItems`; risk: medium: external Mantle mutation; approval required.
- `create_tasks_id_comments`: POST `/v1/tasks/{{ record.id }}/comments` - kind `create`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `comment`, `commentHtml`,
  `id`, `taggedUsers`; risk: medium: external Mantle mutation; approval required.
- `create_tasks_id_todo_items`: POST `/v1/tasks/{{ record.id }}/todo-items` - kind `create`; body
  type `json`; path fields `id`; required record fields `id`, `content`; accepted fields
  `completed`, `content`, `displayOrder`, `id`; risk: medium: external Mantle mutation; approval
  required.
- `create_tickets`: POST `/v1/tickets` - kind `create`; body type `json`; required record fields
  `subject`; accepted fields `appId`, `assignedToId`, `channelId`, `contact`, `contactId`,
  `contactIds`, `createdAt`, `customerId`, `cxEmailAddressId`, `inboxId`, `lastMessageAt`,
  `managedBy`, `message`, `priority`, `readOnly`, `send`, `status`, `subject`, and 2 more; risk:
  medium: external Mantle mutation; approval required.
- `create_tickets_id_events`: POST `/v1/tickets/{{ record.id }}/events` - kind `create`; body type
  `json`; path fields `id`; required record fields `id`, `type`, `actorType`; accepted fields
  `actorType`, `agentId`, `contactId`, `id`, `newValue`, `occurredAt`, `oldValue`,
  `threadMessageId`, `type`; risk: medium: external Mantle mutation; approval required.
- `create_tickets_id_messages`: POST `/v1/tickets/{{ record.id }}/messages` - kind `create`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `actorType`,
  `agentId`, `attachments`, `contact`, `contactId`, `content`, `contentType`, `fullContent`, `id`,
  `inReplyToId`, `isInternal`, `messageId`, `occurredAt`, `referencesIds`, `send`; risk: medium:
  external Mantle mutation; approval required.
- `create_tickets_saved_filters`: POST `/v1/tickets/saved_filters` - kind `create`; body type
  `json`; required record fields `name`; accepted fields `displayOrder`, `filters`, `name`; risk:
  medium: external Mantle mutation; approval required.
- `create_tickets_saved_replies`: POST `/v1/tickets/saved_replies` - kind `create`; body type
  `json`; required record fields `title`; accepted fields `appId`, `categoryId`, `content`,
  `defaultLocale`, `handle`, `isPrivate`, `roles`, `status`, `tags`, `title`; risk: medium: external
  Mantle mutation; approval required.
- `create_timeline_comments`: POST `/v1/timeline_comments` - kind `create`; body type `json`;
  required record fields `commentHtml`; accepted fields `appInstallationId`, `attachments`,
  `comment`, `commentHtml`, `customerId`, `dealId`, `taggedUsers`; risk: medium: external Mantle
  mutation; approval required.
- `create_usage_events`: POST `/v1/usage_events` - kind `create`; body type `json`; accepted fields
  `appId`, `customerId`, `eventId`, `eventName`, `event_id`, `event_name`, `events`, `private`,
  `properties`, `timestamp`; risk: medium: external Mantle mutation; approval required.
- `create_webhooks`: POST `/v1/webhooks` - kind `create`; body type `json`; accepted fields
  `address`, `appIds`, `filter`, `topic`; risk: medium: external Mantle side effect; approval
  required.
- `delete_apps_app_id_checklists_checklist_id`: DELETE `/v1/apps/{{ record.app_id }}/checklists/{{
  record.checklist_id }}` - kind `delete`; body type `none`; path fields `app_id`, `checklist_id`;
  required record fields `app_id`, `checklist_id`; accepted fields `app_id`, `checklist_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: high: external
  Mantle mutation or side effect; approval required.
- `delete_apps_app_id_plans_features_feature_id`: DELETE `/v1/apps/{{ record.app_id
  }}/plans/features/{{ record.feature_id }}` - kind `delete`; body type `none`; path fields
  `app_id`, `feature_id`; required record fields `app_id`, `feature_id`; accepted fields `app_id`,
  `feature_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: high: external Mantle mutation or side effect; approval required.
- `delete_apps_id_usage_metrics_usage_metric_id`: DELETE `/v1/apps/{{ record.id }}/usage_metrics/{{
  record.usage_metric_id }}` - kind `delete`; body type `none`; path fields `id`, `usage_metric_id`;
  required record fields `id`, `usage_metric_id`; accepted fields `id`, `usage_metric_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: high: external
  Mantle mutation or side effect; approval required.
- `delete_companies_id`: DELETE `/v1/companies/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: high: external Mantle mutation or side
  effect; approval required.
- `delete_contacts_id`: DELETE `/v1/contacts/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: high: external Mantle mutation or side
  effect; approval required.
- `delete_customers_custom_fields_id`: DELETE `/v1/customers/custom_fields/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: high:
  external Mantle mutation or side effect; approval required.
- `delete_customers_id_account_owners_owner_id`: DELETE `/v1/customers/{{ record.id
  }}/account_owners/{{ record.owner_id }}` - kind `delete`; body type `none`; path fields `id`,
  `owner_id`; required record fields `id`, `owner_id`; accepted fields `id`, `owner_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: high: external
  Mantle mutation or side effect; approval required.
- `delete_deal_activities_id`: DELETE `/v1/deal_activities/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Mantle
  mutation or side effect; approval required.
- `delete_deal_flows_id`: DELETE `/v1/deal_flows/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: high: external Mantle mutation or side
  effect; approval required.
- `delete_deals_id`: DELETE `/v1/deals/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: high: external Mantle mutation or side effect;
  approval required.
- `delete_docs_collections_collection_id`: DELETE `/v1/docs/collections/{{ record.collection_id }}`
  - kind `delete`; body type `none`; path fields `collection_id`; required record fields
  `collection_id`; accepted fields `collection_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: high: external Mantle mutation or side effect; approval
  required.
- `delete_docs_groups_group_id`: DELETE `/v1/docs/groups/{{ record.group_id }}` - kind `delete`;
  body type `none`; path fields `group_id`; required record fields `group_id`; accepted fields
  `group_id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  high: external Mantle mutation or side effect; approval required.
- `delete_docs_pages_page_id`: DELETE `/v1/docs/pages/{{ record.page_id }}` - kind `delete`; body
  type `none`; path fields `page_id`; required record fields `page_id`; accepted fields `page_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: high:
  external Mantle mutation or side effect; approval required.
- `delete_docs_pages_page_id_archive`: DELETE `/v1/docs/pages/{{ record.page_id }}/archive` - kind
  `delete`; body type `none`; path fields `page_id`; required record fields `page_id`; accepted
  fields `page_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: high: external Mantle mutation or side effect; approval required.
- `delete_docs_pages_page_id_publish`: DELETE `/v1/docs/pages/{{ record.page_id }}/publish` - kind
  `delete`; body type `none`; path fields `page_id`; required record fields `page_id`; accepted
  fields `page_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: high: external Mantle mutation or side effect; approval required.
- `delete_docs_sites_id`: DELETE `/v1/docs/sites/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: high: external Mantle mutation or side
  effect; approval required.
- `delete_docs_sites_id_redirects_redirect_id`: DELETE `/v1/docs/sites/{{ record.id }}/redirects/{{
  record.redirect_id }}` - kind `delete`; body type `none`; path fields `id`, `redirect_id`;
  required record fields `id`, `redirect_id`; accepted fields `id`, `redirect_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Mantle
  mutation or side effect; approval required.
- `delete_docs_sites_id_repositories`: DELETE `/v1/docs/sites/{{ record.id }}/repositories` - kind
  `delete`; body type `json`; path fields `id`; required record fields `id`, `repositoryId`;
  accepted fields `id`, `repositoryId`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: high: external Mantle mutation or side effect; approval
  required.
- `delete_email_campaigns_id`: DELETE `/v1/email/campaigns/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Mantle
  mutation or side effect; approval required.
- `delete_email_unsubscribe_groups_id_members`: DELETE `/v1/email/unsubscribe_groups/{{ record.id
  }}/members` - kind `delete`; body type `none`; path fields `id`; required record fields `id`;
  accepted fields `id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Mantle mutation or side effect; approval required.
- `delete_email_unsubscribe_groups_id_members_member_id`: DELETE `/v1/email/unsubscribe_groups/{{
  record.id }}/members/{{ record.member_id }}` - kind `delete`; body type `none`; path fields `id`,
  `member_id`; required record fields `id`, `member_id`; accepted fields `id`, `member_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: high: external
  Mantle mutation or side effect; approval required.
- `delete_flow_extensions_actions_id`: DELETE `/v1/flow/extensions/actions/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: high:
  external Mantle mutation or side effect; approval required.
- `delete_flow_extensions_triggers_handle`: DELETE `/v1/flow/extensions/triggers/{{ record.handle
  }}` - kind `delete`; body type `none`; path fields `handle`; required record fields `handle`;
  accepted fields `handle`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Mantle mutation or side effect; approval required.
- `delete_flows_id`: DELETE `/v1/flows/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: high: external Mantle mutation or side effect;
  approval required.
- `delete_journal_entries_id`: DELETE `/v1/journal_entries/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Mantle
  mutation or side effect; approval required.
- `delete_lists_id`: DELETE `/v1/lists/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: high: external Mantle mutation or side effect;
  approval required.
- `delete_meetings_id`: DELETE `/v1/meetings/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: high: external Mantle mutation or side
  effect; approval required.
- `delete_meetings_id_permissions`: DELETE `/v1/meetings/{{ record.id }}/permissions` - kind
  `delete`; body type `json`; path fields `id`; required record fields `id`, `userId`; accepted
  fields `id`, `userId`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Mantle mutation or side effect; approval required.
- `delete_synced_emails_id`: DELETE `/v1/synced_emails/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Mantle
  mutation or side effect; approval required.
- `delete_tasks_id`: DELETE `/v1/tasks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: high: external Mantle mutation or side effect;
  approval required.
- `delete_tasks_id_comments_comment_id`: DELETE `/v1/tasks/{{ record.id }}/comments/{{
  record.comment_id }}` - kind `delete`; body type `none`; path fields `id`, `comment_id`; required
  record fields `id`, `comment_id`; accepted fields `comment_id`, `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: high: external Mantle mutation or side
  effect; approval required.
- `delete_tasks_id_todo_items_item_id`: DELETE `/v1/tasks/{{ record.id }}/todo-items/{{
  record.item_id }}` - kind `delete`; body type `none`; path fields `id`, `item_id`; required record
  fields `id`, `item_id`; accepted fields `id`, `item_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: high: external Mantle mutation or side effect;
  approval required.
- `delete_tickets_saved_filters_filter_id`: DELETE `/v1/tickets/saved_filters/{{ record.filter_id
  }}` - kind `delete`; body type `none`; path fields `filter_id`; required record fields
  `filter_id`; accepted fields `filter_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: high: external Mantle mutation or side effect; approval
  required.
- `delete_tickets_saved_replies_reply_id`: DELETE `/v1/tickets/saved_replies/{{ record.reply_id }}`
  - kind `delete`; body type `none`; path fields `reply_id`; required record fields `reply_id`;
  accepted fields `reply_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Mantle mutation or side effect; approval required.
- `delete_timeline_comments_id`: DELETE `/v1/timeline_comments/{{ record.id }}` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: high: external
  Mantle mutation or side effect; approval required.
- `delete_webhooks_id`: DELETE `/v1/webhooks/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: high: external Mantle mutation or side
  effect; approval required.
- `execute_affiliates_id_add_tags`: POST `/v1/affiliates/{{ record.id }}/addTags` - kind `custom`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `id`, `tags`;
  risk: medium: external Mantle mutation; approval required.
- `execute_affiliates_id_remove_tags`: POST `/v1/affiliates/{{ record.id }}/removeTags` - kind
  `custom`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`,
  `tags`; risk: high: external Mantle mutation or side effect; approval required.
- `execute_apps_id_analyze`: POST `/v1/apps/{{ record.id }}/analyze` - kind `custom`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `appUrl`, `archiveKey`,
  `codebaseType`, `id`, `localAnalysis`; risk: medium: external Mantle mutation; approval required.
- `execute_apps_id_skills_skill_id`: POST `/v1/apps/{{ record.id }}/skills/{{ record.skill_id }}` -
  kind `custom`; body type `json`; path fields `id`, `skill_id`; required record fields `id`,
  `skill_id`, `targets`; accepted fields `id`, `skill_id`, `targets`; risk: medium: external Mantle
  mutation; approval required.
- `execute_contacts_id_add_tags`: POST `/v1/contacts/{{ record.id }}/addTags` - kind `custom`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `id`, `tags`; risk:
  medium: external Mantle mutation; approval required.
- `execute_contacts_id_remove_tags`: POST `/v1/contacts/{{ record.id }}/removeTags` - kind `custom`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `id`, `tags`;
  risk: high: external Mantle mutation or side effect; approval required.
- `execute_customers_id_account_owners`: POST `/v1/customers/{{ record.id }}/account_owners` - kind
  `custom`; body type `json`; path fields `id`; required record fields `id`, `userId`, `type`;
  accepted fields `appCommissions`, `commissionEndsAt`, `commissionPercentage`, `hasCommission`,
  `id`, `managementFeePercentage`, `type`, `userId`; risk: medium: external Mantle mutation;
  approval required.
- `execute_customers_id_add_tags`: POST `/v1/customers/{{ record.id }}/addTags` - kind `custom`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `id`, `tags`;
  risk: medium: external Mantle mutation; approval required.
- `execute_customers_id_remove_tags`: POST `/v1/customers/{{ record.id }}/removeTags` - kind
  `custom`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`,
  `tags`; risk: high: external Mantle mutation or side effect; approval required.
- `execute_devices`: POST `/v1/devices` - kind `custom`; body type `json`; required record fields
  `token`, `platform`; accepted fields `appVersion`, `platform`, `token`; risk: medium: external
  Mantle mutation; approval required.
- `execute_docs_pages_generate`: POST `/v1/docs/pages/generate` - kind `custom`; body type `json`;
  required record fields `repositoryId`, `prompt`; accepted fields `collectionId`, `groupId`,
  `prompt`, `repositoryId`; risk: medium: external Mantle side effect; approval required.
- `execute_docs_pages_page_id_archive`: POST `/v1/docs/pages/{{ record.page_id }}/archive` - kind
  `custom`; body type `none`; path fields `page_id`; required record fields `page_id`; accepted
  fields `page_id`; risk: high: external Mantle mutation or side effect; approval required.
- `execute_docs_pages_page_id_generate`: POST `/v1/docs/pages/{{ record.page_id }}/generate` - kind
  `custom`; body type `json`; path fields `page_id`; required record fields `page_id`, `prompt`;
  accepted fields `fields`, `page_id`, `prompt`; risk: medium: external Mantle side effect; approval
  required.
- `execute_docs_pages_page_id_publish`: POST `/v1/docs/pages/{{ record.page_id }}/publish` - kind
  `custom`; body type `json`; path fields `page_id`; required record fields `page_id`,
  `repositoryId`; accepted fields `page_id`, `repositoryId`; risk: high: external Mantle mutation or
  side effect; approval required.
- `execute_docs_sites_id_repositories`: POST `/v1/docs/sites/{{ record.id }}/repositories` - kind
  `custom`; body type `json`; path fields `id`; required record fields `id`, `repositoryId`;
  accepted fields `id`, `repositoryId`; risk: medium: external Mantle mutation; approval required.
- `execute_email_campaigns_id_cancel`: POST `/v1/email/campaigns/{{ record.id }}/cancel` - kind
  `custom`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: high: external Mantle mutation or side effect; approval required.
- `execute_email_campaigns_id_deliver`: POST `/v1/email/campaigns/{{ record.id }}/deliver` - kind
  `custom`; body type `json`; path fields `id`; required record fields `id`, `customerId`; accepted
  fields `customerId`, `email`, `id`, `idempotencyKey`, `metadata`; risk: high: external Mantle
  mutation or side effect; approval required.
- `execute_email_campaigns_id_send`: POST `/v1/email/campaigns/{{ record.id }}/send` - kind
  `custom`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `anyApp`, `audienceIds`, `criteria`, `id`, `scheduledAt`, `triggerIdempotencyKey`; risk: high:
  external Mantle mutation or side effect; approval required.
- `execute_email_campaigns_id_test`: POST `/v1/email/campaigns/{{ record.id }}/test` - kind
  `custom`; body type `json`; path fields `id`; required record fields `id`, `email`; accepted
  fields `email`, `id`, `recipient`; risk: high: external Mantle mutation or side effect; approval
  required.
- `execute_email_unsubscribe_groups_id_members`: POST `/v1/email/unsubscribe_groups/{{ record.id
  }}/members` - kind `custom`; body type `json`; path fields `id`; required record fields `id`,
  `emails`; accepted fields `emails`, `id`; risk: medium: external Mantle side effect; approval
  required.
- `execute_lists_id_add`: POST `/v1/lists/{{ record.id }}/add` - kind `custom`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `contactIds`, `customerIds`, `id`;
  risk: medium: external Mantle mutation; approval required.
- `execute_lists_id_remove`: POST `/v1/lists/{{ record.id }}/remove` - kind `custom`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `contactIds`,
  `customerIds`, `id`; risk: high: external Mantle mutation or side effect; approval required.
- `execute_meetings_id_permissions`: POST `/v1/meetings/{{ record.id }}/permissions` - kind
  `custom`; body type `json`; path fields `id`; required record fields `id`, `userId`; accepted
  fields `id`, `userId`; risk: medium: external Mantle mutation; approval required.
- `execute_meetings_id_task_suggestions_suggestion_id_accept`: POST `/v1/meetings/{{ record.id
  }}/task-suggestions/{{ record.suggestion_id }}/accept` - kind `custom`; body type `json`; path
  fields `id`, `suggestion_id`; required record fields `id`, `suggestion_id`; accepted fields
  `assigneeId`, `description`, `dueDate`, `id`, `priority`, `suggestion_id`, `title`; risk: medium:
  external Mantle side effect; approval required.
- `execute_meetings_id_task_suggestions_suggestion_id_dismiss`: POST `/v1/meetings/{{ record.id
  }}/task-suggestions/{{ record.suggestion_id }}/dismiss` - kind `custom`; body type `none`; path
  fields `id`, `suggestion_id`; required record fields `id`, `suggestion_id`; accepted fields `id`,
  `suggestion_id`; risk: medium: external Mantle side effect; approval required.
- `execute_meetings_id_transcribe`: POST `/v1/meetings/{{ record.id }}/transcribe` - kind `custom`;
  body type `json`; path fields `id`; required record fields `id`, `recordingKey`; accepted fields
  `id`, `recordingKey`; risk: medium: external Mantle mutation; approval required.
- `execute_synced_emails`: POST `/v1/synced_emails` - kind `custom`; body type `json`; accepted
  fields `emailData`, `messages`; risk: medium: external Mantle side effect; approval required.
- `execute_synced_emails_id_messages`: POST `/v1/synced_emails/{{ record.id }}/messages` - kind
  `custom`; body type `json`; path fields `id`; required record fields `id`, `messages`; accepted
  fields `id`, `messages`; risk: medium: external Mantle side effect; approval required.
- `execute_tickets_id_ai_replies`: POST `/v1/tickets/{{ record.id }}/ai-replies` - kind `custom`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `adjustment`,
  `draft`, `id`; risk: medium: external Mantle side effect; approval required.
- `update_apps_app_id_checklists_checklist_id`: PUT `/v1/apps/{{ record.app_id }}/checklists/{{
  record.checklist_id }}` - kind `update`; body type `json`; path fields `app_id`, `checklist_id`;
  required record fields `app_id`, `checklist_id`; accepted fields `app_id`, `checklist_id`,
  `customerSegmentId`, `description`, `handle`, `name`, `status`, `steps`, `title`; risk: medium:
  external Mantle mutation; approval required.
- `update_apps_app_id_plans_features_feature_id`: PUT `/v1/apps/{{ record.app_id
  }}/plans/features/{{ record.feature_id }}` - kind `update`; body type `json`; path fields
  `app_id`, `feature_id`; required record fields `app_id`, `feature_id`; accepted fields
  `allowedValues`, `app_id`, `createdAt`, `defaultValue`, `description`, `feature_id`, `id`, `key`,
  `name`, `type`, `updatedAt`, `usageMetric`; risk: medium: external Mantle mutation; approval
  required.
- `update_apps_id_plans_plan_id`: PUT `/v1/apps/{{ record.id }}/plans/{{ record.plan_id }}` - kind
  `update`; body type `json`; path fields `id`, `plan_id`; required record fields `id`, `plan_id`;
  accepted fields `amount`, `autoUpgradeToPlanId`, `basePlanIds`, `customFields`, `customer`,
  `customerExcludeTags`, `customerTags`, `deprecated`, `description`, `features`, `flexBilling`,
  `flexBillingTerms`, `id`, `interval`, `name`, `needsReview`, `onUsageLimitReached`,
  `planUsageCharges`, and 10 more; risk: medium: external Mantle mutation; approval required.
- `update_apps_id_plans_plan_id_archive`: PUT `/v1/apps/{{ record.id }}/plans/{{ record.plan_id
  }}/archive` - kind `update`; body type `none`; path fields `id`, `plan_id`; required record fields
  `id`, `plan_id`; accepted fields `id`, `plan_id`; risk: high: external Mantle mutation or side
  effect; approval required.
- `update_apps_id_plans_plan_id_unarchive`: PUT `/v1/apps/{{ record.id }}/plans/{{ record.plan_id
  }}/unarchive` - kind `update`; body type `none`; path fields `id`, `plan_id`; required record
  fields `id`, `plan_id`; accepted fields `id`, `plan_id`; risk: high: external Mantle mutation or
  side effect; approval required.
- `update_apps_id_usage_metrics_usage_metric_id`: PUT `/v1/apps/{{ record.id }}/usage_metrics/{{
  record.usage_metric_id }}` - kind `update`; body type `json`; path fields `id`, `usage_metric_id`;
  required record fields `id`, `usage_metric_id`; accepted fields `calculation`, `eventName`, `id`,
  `name`, `params`, `usage_metric_id`; risk: medium: external Mantle mutation; approval required.
- `update_companies_id`: PUT `/v1/companies/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `parentCustomerId`; accepted fields `id`, `name`,
  `parentCustomerId`; risk: medium: external Mantle mutation; approval required.
- `update_contacts_id`: PUT `/v1/contacts/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `customers`, `email`, `id`, `jobTitle`,
  `name`, `notes`, `phone`, `secondaryEmails`, `socialProfiles`, `tags`; risk: medium: external
  Mantle mutation; approval required.
- `update_custom_data`: PUT `/v1/custom_data` - kind `update`; body type `json`; required record
  fields `resourceId`, `resourceType`, `key`, `value`; accepted fields `key`, `resourceId`,
  `resourceType`, `value`; risk: medium: external Mantle mutation; approval required.
- `update_customers_custom_fields_id`: PUT `/v1/customers/custom_fields/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `appId`, `appLevel`, `defaultValue`, `filterable`, `id`, `options`, `private`,
  `showOnCustomerDetail`; risk: medium: external Mantle mutation; approval required.
- `update_customers_id`: PUT `/v1/customers/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `appInstallations`, `companyId`,
  `countryCode`, `customFields`, `description`, `domain`, `email`, `id`, `industry`, `name`,
  `preferredCurrency`, `shopifyDomain`, `shopifyShopId`, `tags`, `url`; risk: medium: external
  Mantle mutation; approval required.
- `update_customers_id_account_owners_owner_id`: PUT `/v1/customers/{{ record.id
  }}/account_owners/{{ record.owner_id }}` - kind `update`; body type `json`; path fields `id`,
  `owner_id`; required record fields `id`, `owner_id`; accepted fields `appCommissions`,
  `commissionEndsAt`, `commissionPercentage`, `hasCommission`, `id`, `managementFeePercentage`,
  `owner_id`, `type`; risk: medium: external Mantle mutation; approval required.
- `update_deal_activities_id`: PUT `/v1/deal_activities/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `defaultWeight`,
  `description`, `icon`, `id`, `name`, `timelineDescriptionTemplate`, `timelineTitleTemplate`; risk:
  medium: external Mantle mutation; approval required.
- `update_deal_flows_id`: PUT `/v1/deal_flows/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `name`, `dealStages`; accepted fields `dealStages`,
  `defaultAcquisitionChannel`, `defaultAcquisitionSource`, `defaultDealOwnerId`, `description`,
  `id`, `isDefaultDealFlow`, `name`; risk: medium: external Mantle side effect; approval required.
- `update_deals_id`: PUT `/v1/deals/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `acquisitionChannel`, `acquisitionSource`,
  `affiliateId`, `amount`, `amountCurrencyCode`, `appId`, `closedAt`, `closingAt`, `companyId`,
  `contactIds`, `contacts`, `customer`, `customerId`, `dealFlowId`, `dealStageId`, `domain`,
  `firstInteractionAt`, `id`, and 6 more; risk: medium: external Mantle mutation; approval required.
- `update_docs_collections_collection_id`: PUT `/v1/docs/collections/{{ record.collection_id }}` -
  kind `update`; body type `json`; path fields `collection_id`; required record fields
  `collection_id`; accepted fields `collection_id`, `description`, `displayOrder`, `handle`, `icon`,
  `locale`, `status`, `title`; risk: medium: external Mantle mutation; approval required.
- `update_docs_groups_group_id`: PUT `/v1/docs/groups/{{ record.group_id }}` - kind `update`; body
  type `json`; path fields `group_id`; required record fields `group_id`; accepted fields
  `displayOrder`, `group_id`, `handle`, `icon`, `locale`, `title`; risk: medium: external Mantle
  mutation; approval required.
- `update_docs_pages_page_id`: PUT `/v1/docs/pages/{{ record.page_id }}` - kind `update`; body type
  `json`; path fields `page_id`; required record fields `page_id`; accepted fields `content`,
  `displayOrder`, `groupId`, `handle`, `locale`, `openGraphImage`, `page_id`, `parentPageId`,
  `seoDescription`, `seoTitle`, `title`; risk: medium: external Mantle mutation; approval required.
- `update_docs_repositories_id`: PUT `/v1/docs/repositories/{{ record.id }}` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `autoTranslate`,
  `config`, `customDomain`, `defaultLocale`, `id`, `locale`, `shortDescription`, `supportedLocales`,
  `title`, `useCustomDomain`, `visibility`; risk: medium: external Mantle mutation; approval
  required.
- `update_docs_sites_id`: PUT `/v1/docs/sites/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `config`, `customDomain`,
  `defaultLocale`, `handle`, `id`, `locale`, `shortDescription`, `supportedLocales`, `title`,
  `useCustomDomain`, `visibility`, `widgetId`; risk: medium: external Mantle mutation; approval
  required.
- `update_docs_sites_id_redirects_redirect_id`: PUT `/v1/docs/sites/{{ record.id }}/redirects/{{
  record.redirect_id }}` - kind `update`; body type `json`; path fields `id`, `redirect_id`;
  required record fields `id`, `redirect_id`; accepted fields `fromPath`, `id`, `redirect_id`,
  `toPath`; risk: medium: external Mantle mutation; approval required.
- `update_docs_sites_id_repositories`: PUT `/v1/docs/sites/{{ record.id }}/repositories` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`, `attachments`; accepted
  fields `attachments`, `id`; risk: medium: external Mantle mutation; approval required.
- `update_email_campaigns_id`: PUT `/v1/email/campaigns/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `html`, `id`, `layoutId`,
  `name`, `plainText`, `previewText`, `senderId`, `status`, `subject`, `unsubscribeGroupId`; risk:
  medium: external Mantle side effect; approval required.
- `update_flow_actions_runs_id`: PUT `/v1/flow/actions/runs/{{ record.id }}` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `id`, `status`; risk:
  medium: external Mantle side effect; approval required.
- `update_flow_extensions_actions_id`: PUT `/v1/flow/extensions/actions/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `description`, `id`, `name`, `settingsSchema`, `url`; risk: medium: external Mantle mutation;
  approval required.
- `update_flows_id`: PATCH `/v1/flows/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `allowRepeatRuns`,
  `blockRepeatsTimeUnit`, `blockRepeatsTimeValue`, `id`, `name`; risk: medium: external Mantle side
  effect; approval required.
- `update_journal_entries_id`: PUT `/v1/journal_entries/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `appId`, `date`,
  `description`, `emoji`, `id`, `tags`, `title`, `url`; risk: medium: external Mantle mutation;
  approval required.
- `update_lists_id`: PUT `/v1/lists/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `description`, `id`, `name`; risk: medium:
  external Mantle mutation; approval required.
- `update_meetings_id`: PUT `/v1/meetings/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `customerId`, `dealId`, `duration`,
  `endTime`, `id`, `meetingUrl`, `platform`, `platformMeetingId`, `recordingStatus`, `recordingUrl`,
  `startTime`, `summary`, `title`; risk: medium: external Mantle mutation; approval required.
- `update_meetings_id_attendees_attendee_id`: PUT `/v1/meetings/{{ record.id }}/attendees/{{
  record.attendee_id }}` - kind `update`; body type `json`; path fields `id`, `attendee_id`;
  required record fields `id`, `attendee_id`; accepted fields `attendee_id`, `contactId`, `email`,
  `id`, `name`, `userId`; risk: medium: external Mantle mutation; approval required.
- `update_meetings_id_visibility`: PUT `/v1/meetings/{{ record.id }}/visibility` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`, `visibility`; accepted fields
  `id`, `visibility`; risk: medium: external Mantle mutation; approval required.
- `update_notification_preferences`: PUT `/v1/notification_preferences` - kind `update`; body type
  `json`; required record fields `helpDeskNotificationPreferences`; accepted fields
  `helpDeskNotificationPreferences`; risk: medium: external Mantle mutation; approval required.
- `update_synced_emails_id`: PUT `/v1/synced_emails/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `contactId`, `customerId`,
  `dealId`, `id`, `snippet`, `source`, `subject`; risk: medium: external Mantle side effect;
  approval required.
- `update_tasks_id`: PUT `/v1/tasks/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `assigneeId`, `dealActivityId`, `description`,
  `descriptionHtml`, `dueDate`, `id`, `priority`, `reminders`, `status`, `tags`, `timezone`,
  `title`, `todoItems`; risk: medium: external Mantle mutation; approval required.
- `update_tasks_id_comments_comment_id`: PUT `/v1/tasks/{{ record.id }}/comments/{{
  record.comment_id }}` - kind `update`; body type `json`; path fields `id`, `comment_id`; required
  record fields `id`, `comment_id`; accepted fields `comment`, `commentHtml`, `comment_id`, `id`,
  `taggedUsers`; risk: medium: external Mantle mutation; approval required.
- `update_tasks_id_todo_items_item_id`: PUT `/v1/tasks/{{ record.id }}/todo-items/{{ record.item_id
  }}` - kind `update`; body type `json`; path fields `id`, `item_id`; required record fields `id`,
  `item_id`; accepted fields `completed`, `content`, `displayOrder`, `id`, `item_id`; risk: medium:
  external Mantle mutation; approval required.
- `update_tickets_id`: PUT `/v1/tickets/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `appId`, `assignedToId`, `channelId`,
  `contact`, `contactId`, `customerId`, `id`, `lastMessageAt`, `managedBy`, `priority`, `readOnly`,
  `status`, `subject`, `tags`, `updatedAt`; risk: medium: external Mantle mutation; approval
  required.
- `update_tickets_id_events`: PUT `/v1/tickets/{{ record.id }}/events` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: medium:
  external Mantle mutation; approval required.
- `update_tickets_id_messages`: PUT `/v1/tickets/{{ record.id }}/messages` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: medium:
  external Mantle mutation; approval required.
- `update_tickets_saved_filters_filter_id`: PUT `/v1/tickets/saved_filters/{{ record.filter_id }}` -
  kind `update`; body type `json`; path fields `filter_id`; required record fields `filter_id`;
  accepted fields `displayOrder`, `filter_id`, `filters`, `name`; risk: medium: external Mantle
  mutation; approval required.
- `update_tickets_saved_replies_reply_id`: PUT `/v1/tickets/saved_replies/{{ record.reply_id }}` -
  kind `update`; body type `json`; path fields `reply_id`; required record fields `reply_id`;
  accepted fields `appId`, `categoryId`, `content`, `defaultLocale`, `handle`, `isPrivate`,
  `reply_id`, `roles`, `status`, `tags`, `title`; risk: medium: external Mantle mutation; approval
  required.
- `update_timeline_comments_id`: PUT `/v1/timeline_comments/{{ record.id }}` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`, `commentHtml`; accepted fields
  `appEventId`, `commentHtml`, `dealEventId`, `id`; risk: medium: external Mantle mutation; approval
  required.
- `update_webhooks_id`: PUT `/v1/webhooks/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `address`, `appIds`, `filter`, `id`,
  `topic`; risk: medium: external Mantle side effect; approval required.
- `upsert_contacts`: POST `/v1/contacts` - kind `upsert`; body type `json`; accepted fields
  `customers`, `email`, `jobTitle`, `name`, `notes`, `phone`, `secondaryEmails`, `socialProfiles`,
  `tags`; risk: medium: external Mantle mutation; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 133 stream-backed endpoint group(s), 148 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  deprecated=4, non_data_endpoint=40.
