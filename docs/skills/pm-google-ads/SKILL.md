---
name: pm-google-ads
description: Google Ads connector knowledge and safe action guide.
---

# pm-google-ads

## Purpose

Declarative Google Ads connector for v22 customer, campaign, ad group, direct-read, and limited guarded reverse-ETL API surfaces.

## Icon

- id: google-adwords
- asset: icons/google-adwords.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.google.com/google-ads/api/docs/release-notes

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- customer_id
- login_customer_id
- max_pages
- mode
- page_size
- access_token (secret) (required)
- developer_token (secret) (required)

## ETL Streams

- accessible_customers:
  - primary key: customer_id
  - fields: customer_id(string), resource_name(string)
- campaigns:
  - primary key: id
  - fields: id(string), name(string), resource_name(string), status(string)
- ad_groups:
  - primary key: id
  - fields: id(string), name(string), resource_name(string), status(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- mutate_ad_group_criterion_customizers:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/AdGroupCriterionCustomizers:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.AdGroupCriterionCustomizers.mutate against the configured customer. Review provider-side effects before approval.
- mutate_campaign_goal_configs:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/CampaignGoalConfigs:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.CampaignGoalConfigs.mutate against the configured customer. Review provider-side effects before approval.
- mutate_customer_customizers:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/CustomerCustomizers:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.CustomerCustomizers.mutate against the configured customer. Review provider-side effects before approval.
- mutate_goals:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/Goals:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.Goals.mutate against the configured customer. Review provider-side effects before approval.
- mutate_account_budget_proposals:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/accountBudgetProposals:mutate
  - risk: Executes Google Ads API v22 method customers.accountBudgetProposals.mutate against the configured customer. Review provider-side effects before approval.
- create_account_links:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/accountLinks:create
  - risk: Executes Google Ads API v22 method customers.accountLinks.create against the configured customer. Review provider-side effects before approval.
- mutate_account_links:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/accountLinks:mutate
  - risk: Executes Google Ads API v22 method customers.accountLinks.mutate against the configured customer. Review provider-side effects before approval.
- mutate_ad_group_ad_labels:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/adGroupAdLabels:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.adGroupAdLabels.mutate against the configured customer. Review provider-side effects before approval.
- mutate_ad_group_ads:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/adGroupAds:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.adGroupAds.mutate against the configured customer. Review provider-side effects before approval.
- mutate_ad_group_asset_sets:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/adGroupAssetSets:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.adGroupAssetSets.mutate against the configured customer. Review provider-side effects before approval.
- mutate_ad_group_assets:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/adGroupAssets:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.adGroupAssets.mutate against the configured customer. Review provider-side effects before approval.
- mutate_ad_group_bid_modifiers:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/adGroupBidModifiers:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.adGroupBidModifiers.mutate against the configured customer. Review provider-side effects before approval.
- mutate_ad_group_criteria:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/adGroupCriteria:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.adGroupCriteria.mutate against the configured customer. Review provider-side effects before approval.
- mutate_ad_group_criterion_labels:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/adGroupCriterionLabels:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.adGroupCriterionLabels.mutate against the configured customer. Review provider-side effects before approval.
- mutate_ad_group_customizers:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/adGroupCustomizers:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.adGroupCustomizers.mutate against the configured customer. Review provider-side effects before approval.
- mutate_ad_group_labels:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/adGroupLabels:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.adGroupLabels.mutate against the configured customer. Review provider-side effects before approval.
- mutate_ad_groups:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/adGroups:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.adGroups.mutate against the configured customer. Review provider-side effects before approval.
- mutate_ad_parameters:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/adParameters:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.adParameters.mutate against the configured customer. Review provider-side effects before approval.
- mutate_ads:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/ads:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.ads.mutate against the configured customer. Review provider-side effects before approval.
- mutate_asset_group_assets:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/assetGroupAssets:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.assetGroupAssets.mutate against the configured customer. Review provider-side effects before approval.
- mutate_asset_group_listing_group_filters:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/assetGroupListingGroupFilters:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.assetGroupListingGroupFilters.mutate against the configured customer. Review provider-side effects before approval.
- mutate_asset_group_signals:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/assetGroupSignals:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.assetGroupSignals.mutate against the configured customer. Review provider-side effects before approval.
- mutate_asset_groups:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/assetGroups:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.assetGroups.mutate against the configured customer. Review provider-side effects before approval.
- mutate_asset_set_assets:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/assetSetAssets:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.assetSetAssets.mutate against the configured customer. Review provider-side effects before approval.
- mutate_asset_sets:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/assetSets:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.assetSets.mutate against the configured customer. Review provider-side effects before approval.
- mutate_assets:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/assets:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.assets.mutate against the configured customer. Review provider-side effects before approval.
- mutate_audiences:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/audiences:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.audiences.mutate against the configured customer. Review provider-side effects before approval.
- mutate_batch_jobs:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/batchJobs:mutate
  - risk: Executes Google Ads API v22 method customers.batchJobs.mutate against the configured customer. Review provider-side effects before approval.
- mutate_bidding_data_exclusions:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/biddingDataExclusions:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.biddingDataExclusions.mutate against the configured customer. Review provider-side effects before approval.
- mutate_bidding_seasonality_adjustments:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/biddingSeasonalityAdjustments:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.biddingSeasonalityAdjustments.mutate against the configured customer. Review provider-side effects before approval.
- mutate_bidding_strategies:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/biddingStrategies:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.biddingStrategies.mutate against the configured customer. Review provider-side effects before approval.
- mutate_billing_setups:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/billingSetups:mutate
  - risk: Executes Google Ads API v22 method customers.billingSetups.mutate against the configured customer. Review provider-side effects before approval.
- mutate_campaign_asset_sets:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/campaignAssetSets:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.campaignAssetSets.mutate against the configured customer. Review provider-side effects before approval.
- mutate_campaign_assets:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/campaignAssets:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.campaignAssets.mutate against the configured customer. Review provider-side effects before approval.
- mutate_campaign_bid_modifiers:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/campaignBidModifiers:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.campaignBidModifiers.mutate against the configured customer. Review provider-side effects before approval.
- mutate_campaign_budgets:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/campaignBudgets:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.campaignBudgets.mutate against the configured customer. Review provider-side effects before approval.
- mutate_campaign_conversion_goals:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/campaignConversionGoals:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.campaignConversionGoals.mutate against the configured customer. Review provider-side effects before approval.
- mutate_campaign_criteria:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/campaignCriteria:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.campaignCriteria.mutate against the configured customer. Review provider-side effects before approval.
- mutate_campaign_customizers:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/campaignCustomizers:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.campaignCustomizers.mutate against the configured customer. Review provider-side effects before approval.
- mutate_campaign_drafts:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/campaignDrafts:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.campaignDrafts.mutate against the configured customer. Review provider-side effects before approval.
- mutate_campaign_groups:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/campaignGroups:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.campaignGroups.mutate against the configured customer. Review provider-side effects before approval.
- mutate_campaign_labels:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/campaignLabels:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.campaignLabels.mutate against the configured customer. Review provider-side effects before approval.
- configure_campaign_lifecycle_goals_campaign_lifecycle_goal:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/campaignLifecycleGoal:configureCampaignLifecycleGoals
  - risk: Executes Google Ads API v22 method customers.campaignLifecycleGoal.configureCampaignLifecycleGoals against the configured customer. Review provider-side effects before approval.
- mutate_campaign_shared_sets:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/campaignSharedSets:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.campaignSharedSets.mutate against the configured customer. Review provider-side effects before approval.
- enable_p_max_brand_guidelines_campaigns:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/campaigns:enablePMaxBrandGuidelines
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.campaigns.enablePMaxBrandGuidelines against the configured customer. Review provider-side effects before approval.
- mutate_campaigns:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/campaigns:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.campaigns.mutate against the configured customer. Review provider-side effects before approval.
- mutate_conversion_actions:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/conversionActions:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.conversionActions.mutate against the configured customer. Review provider-side effects before approval.
- mutate_conversion_custom_variables:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/conversionCustomVariables:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.conversionCustomVariables.mutate against the configured customer. Review provider-side effects before approval.
- mutate_conversion_goal_campaign_configs:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/conversionGoalCampaignConfigs:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.conversionGoalCampaignConfigs.mutate against the configured customer. Review provider-side effects before approval.
- mutate_conversion_value_rule_sets:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/conversionValueRuleSets:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.conversionValueRuleSets.mutate against the configured customer. Review provider-side effects before approval.
- mutate_conversion_value_rules:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/conversionValueRules:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.conversionValueRules.mutate against the configured customer. Review provider-side effects before approval.
- create_customer_client_customers:
  - endpoint: POST /v22/customers/{{ config.customer_id }}:createCustomerClient
  - risk: Executes Google Ads API v22 method customers.createCustomerClient against the configured customer. Review provider-side effects before approval.
- mutate_custom_audiences:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/customAudiences:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.customAudiences.mutate against the configured customer. Review provider-side effects before approval.
- mutate_custom_conversion_goals:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/customConversionGoals:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.customConversionGoals.mutate against the configured customer. Review provider-side effects before approval.
- mutate_custom_interests:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/customInterests:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.customInterests.mutate against the configured customer. Review provider-side effects before approval.
- mutate_customer_asset_sets:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/customerAssetSets:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.customerAssetSets.mutate against the configured customer. Review provider-side effects before approval.
- mutate_customer_assets:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/customerAssets:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.customerAssets.mutate against the configured customer. Review provider-side effects before approval.
- mutate_customer_client_links:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/customerClientLinks:mutate
  - risk: Executes Google Ads API v22 method customers.customerClientLinks.mutate against the configured customer. Review provider-side effects before approval.
- mutate_customer_conversion_goals:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/customerConversionGoals:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.customerConversionGoals.mutate against the configured customer. Review provider-side effects before approval.
- mutate_customer_labels:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/customerLabels:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.customerLabels.mutate against the configured customer. Review provider-side effects before approval.
- configure_customer_lifecycle_goals_customer_lifecycle_goal:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/customerLifecycleGoal:configureCustomerLifecycleGoals
  - risk: Executes Google Ads API v22 method customers.customerLifecycleGoal.configureCustomerLifecycleGoals against the configured customer. Review provider-side effects before approval.
- move_manager_link_customer_manager_links:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/customerManagerLinks:moveManagerLink
  - risk: Executes Google Ads API v22 method customers.customerManagerLinks.moveManagerLink against the configured customer. Review provider-side effects before approval.
- mutate_customer_manager_links:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/customerManagerLinks:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.customerManagerLinks.mutate against the configured customer. Review provider-side effects before approval.
- mutate_customer_negative_criteria:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/customerNegativeCriteria:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.customerNegativeCriteria.mutate against the configured customer. Review provider-side effects before approval.
- mutate_customer_sk_ad_network_conversion_value_schemas:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/customerSkAdNetworkConversionValueSchemas:mutate
  - risk: Executes Google Ads API v22 method customers.customerSkAdNetworkConversionValueSchemas.mutate against the configured customer. Review provider-side effects before approval.
- mutate_customer_user_access_invitations:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/customerUserAccessInvitations:mutate
  - risk: Executes Google Ads API v22 method customers.customerUserAccessInvitations.mutate against the configured customer. Review provider-side effects before approval.
- mutate_customer_user_accesses:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/customerUserAccesses:mutate
  - risk: Executes Google Ads API v22 method customers.customerUserAccesses.mutate against the configured customer. Review provider-side effects before approval.
- mutate_customizer_attributes:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/customizerAttributes:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.customizerAttributes.mutate against the configured customer. Review provider-side effects before approval.
- create_data_links:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/dataLinks:create
  - risk: Executes Google Ads API v22 method customers.dataLinks.create against the configured customer. Review provider-side effects before approval.
- remove_data_links:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/dataLinks:remove
  - required fields: resourceName
  - risk: Executes Google Ads API v22 method customers.dataLinks.remove against the configured customer. Review provider-side effects before approval.
- update_data_links:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/dataLinks:update
  - required fields: resourceName
  - risk: Executes Google Ads API v22 method customers.dataLinks.update against the configured customer. Review provider-side effects before approval.
- mutate_experiment_arms:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/experimentArms:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.experimentArms.mutate against the configured customer. Review provider-side effects before approval.
- mutate_experiments:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/experiments:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.experiments.mutate against the configured customer. Review provider-side effects before approval.
- mutate_google_ads:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/googleAds:mutate
  - risk: Executes Google Ads API v22 method customers.googleAds.mutate against the configured customer. Review provider-side effects before approval.
- mutate_keyword_plan_ad_group_keywords:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/keywordPlanAdGroupKeywords:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.keywordPlanAdGroupKeywords.mutate against the configured customer. Review provider-side effects before approval.
- mutate_keyword_plan_ad_groups:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/keywordPlanAdGroups:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.keywordPlanAdGroups.mutate against the configured customer. Review provider-side effects before approval.
- mutate_keyword_plan_campaign_keywords:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/keywordPlanCampaignKeywords:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.keywordPlanCampaignKeywords.mutate against the configured customer. Review provider-side effects before approval.
- mutate_keyword_plan_campaigns:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/keywordPlanCampaigns:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.keywordPlanCampaigns.mutate against the configured customer. Review provider-side effects before approval.
- mutate_keyword_plans:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/keywordPlans:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.keywordPlans.mutate against the configured customer. Review provider-side effects before approval.
- mutate_labels:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/labels:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.labels.mutate against the configured customer. Review provider-side effects before approval.
- append_lead_conversation_local_services:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/localServices:appendLeadConversation
  - risk: Executes Google Ads API v22 method customers.localServices.appendLeadConversation against the configured customer. Review provider-side effects before approval.
- resolve_multi_party_auth_review:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/multiPartyAuthReview:resolve
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.multiPartyAuthReview.resolve against the configured customer. Review provider-side effects before approval.
- mutate_customers:
  - endpoint: POST /v22/customers/{{ config.customer_id }}:mutate
  - risk: Executes Google Ads API v22 method customers.mutate against the configured customer. Review provider-side effects before approval.
- create_offline_user_data_jobs:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/offlineUserDataJobs:create
  - risk: Executes Google Ads API v22 method customers.offlineUserDataJobs.create against the configured customer. Review provider-side effects before approval.
- create_product_link_invitations:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/productLinkInvitations:create
  - risk: Executes Google Ads API v22 method customers.productLinkInvitations.create against the configured customer. Review provider-side effects before approval.
- remove_product_link_invitations:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/productLinkInvitations:remove
  - required fields: resourceName
  - risk: Executes Google Ads API v22 method customers.productLinkInvitations.remove against the configured customer. Review provider-side effects before approval.
- update_product_link_invitations:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/productLinkInvitations:update
  - required fields: resourceName
  - risk: Executes Google Ads API v22 method customers.productLinkInvitations.update against the configured customer. Review provider-side effects before approval.
- create_product_links:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/productLinks:create
  - risk: Executes Google Ads API v22 method customers.productLinks.create against the configured customer. Review provider-side effects before approval.
- remove_product_links:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/productLinks:remove
  - required fields: resourceName
  - risk: Executes Google Ads API v22 method customers.productLinks.remove against the configured customer. Review provider-side effects before approval.
- mutate_recommendation_subscription_recommendation_subscriptions:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/recommendationSubscriptions:mutateRecommendationSubscription
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.recommendationSubscriptions.mutateRecommendationSubscription against the configured customer. Review provider-side effects before approval.
- apply_recommendations:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/recommendations:apply
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.recommendations.apply against the configured customer. Review provider-side effects before approval.
- dismiss_recommendations:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/recommendations:dismiss
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.recommendations.dismiss against the configured customer. Review provider-side effects before approval.
- mutate_remarketing_actions:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/remarketingActions:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.remarketingActions.mutate against the configured customer. Review provider-side effects before approval.
- remove_campaign_automatically_created_asset_customers:
  - endpoint: POST /v22/customers/{{ config.customer_id }}:removeCampaignAutomaticallyCreatedAsset
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.removeCampaignAutomaticallyCreatedAsset against the configured customer. Review provider-side effects before approval.
- mutate_shared_criteria:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/sharedCriteria:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.sharedCriteria.mutate against the configured customer. Review provider-side effects before approval.
- mutate_shared_sets:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/sharedSets:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.sharedSets.mutate against the configured customer. Review provider-side effects before approval.
- mutate_smart_campaign_settings:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/smartCampaignSettings:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.smartCampaignSettings.mutate against the configured customer. Review provider-side effects before approval.
- start_identity_verification_customers:
  - endpoint: POST /v22/customers/{{ config.customer_id }}:startIdentityVerification
  - risk: Executes Google Ads API v22 method customers.startIdentityVerification against the configured customer. Review provider-side effects before approval.
- upload_call_conversions_customers:
  - endpoint: POST /v22/customers/{{ config.customer_id }}:uploadCallConversions
  - risk: Executes Google Ads API v22 method customers.uploadCallConversions against the configured customer. Review provider-side effects before approval.
- upload_click_conversions_customers:
  - endpoint: POST /v22/customers/{{ config.customer_id }}:uploadClickConversions
  - risk: Executes Google Ads API v22 method customers.uploadClickConversions against the configured customer. Review provider-side effects before approval.
- upload_conversion_adjustments_customers:
  - endpoint: POST /v22/customers/{{ config.customer_id }}:uploadConversionAdjustments
  - risk: Executes Google Ads API v22 method customers.uploadConversionAdjustments against the configured customer. Review provider-side effects before approval.
- upload_user_data_customers:
  - endpoint: POST /v22/customers/{{ config.customer_id }}:uploadUserData
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.uploadUserData against the configured customer. Review provider-side effects before approval.
- mutate_user_list_customer_types:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/userListCustomerTypes:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.userListCustomerTypes.mutate against the configured customer. Review provider-side effects before approval.
- mutate_user_lists:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/userLists:mutate
  - required fields: operations
  - risk: Executes Google Ads API v22 method customers.userLists.mutate against the configured customer. Review provider-side effects before approval.

## Security

- read risk: external Google Ads API reads of customer, campaign, ad-group, and bounded direct-read metadata
- write risk: limited guarded Google Ads API reverse/write actions with closed record schemas; destructive/admin actions require explicit approval
- approval: reads require no approval; writes remain gated by plan -> preview -> explicit approval -> execute
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Google Ads v22 fixed direct reads and limited guarded reverse/write actions.
- Usage: pm google-ads <resource> <operation> [flags]
- Other Commands
  - audience-insights list-insights-eligible-dates - Read Google Ads audienceInsights.listInsightsEligibleDates. [intent=direct_read availability=implemented operation=google_ads.audience.insights.list.insights.eligible.dates]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --insights-application-info, --page, --page-cursor
  - customers asset-generations generate-images - Read Google Ads customers.assetGenerations.generateImages. [intent=direct_read availability=implemented operation=google_ads.customers.asset.generations.generate.images]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --advertising-channel-type, --asset-field-types, --final-url-generation, --freeform-generation, --product-recontext-generation, --page, --page-cursor
  - customers asset-generations generate-text - Read Google Ads customers.assetGenerations.generateText. [intent=direct_read availability=implemented operation=google_ads.customers.asset.generations.generate.text]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --advertising-channel-type, --asset-field-types (required), --existing-generation-context, --final-url, --freeform-prompt, --keywords, --page, --page-cursor
  - customers generate-ad-group-themes - Read Google Ads customers.generateAdGroupThemes. [intent=direct_read availability=implemented operation=google_ads.customers.generate.ad.group.themes]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --ad-groups (required), --keywords (required), --page, --page-cursor
  - customers generate-audience-composition-insights - Read Google Ads customers.generateAudienceCompositionInsights. [intent=direct_read availability=implemented operation=google_ads.customers.generate.audience.composition.insights]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --audience (required), --baseline-audience, --customer-insights-group, --data-month, --dimensions (required), --insights-application-info, --page, --page-cursor
  - customers generate-audience-overlap-insights - Read Google Ads customers.generateAudienceOverlapInsights. [intent=direct_read availability=implemented operation=google_ads.customers.generate.audience.overlap.insights]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --country-location (required), --customer-insights-group, --dimensions (required), --insights-application-info, --primary-attribute (required), --page, --page-cursor
  - customers generate-creator-insights - Read Google Ads customers.generateCreatorInsights. [intent=direct_read availability=implemented operation=google_ads.customers.generate.creator.insights]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --country-locations (required), --customer-insights-group (required), --insights-application-info, --search-attributes, --search-brand, --search-channels, --sub-country-locations, --page, --page-cursor
  - customers generate-insights-finder-report - Read Google Ads customers.generateInsightsFinderReport. [intent=direct_read availability=implemented operation=google_ads.customers.generate.insights.finder.report]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --baseline-audience (required), --customer-insights-group, --insights-application-info, --specific-audience (required), --page, --page-cursor
  - customers generate-keyword-forecast-metrics - Read Google Ads customers.generateKeywordForecastMetrics. [intent=direct_read availability=implemented operation=google_ads.customers.generate.keyword.forecast.metrics]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --campaign (required), --currency-code, --forecast-period, --page, --page-cursor
  - customers generate-keyword-historical-metrics - Read Google Ads customers.generateKeywordHistoricalMetrics. [intent=direct_read availability=implemented operation=google_ads.customers.generate.keyword.historical.metrics]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --aggregate-metrics, --geo-target-constants, --historical-metrics-options, --include-adult-keywords, --keyword-plan-network, --keywords, --language, --page, --page-cursor
  - customers generate-keyword-ideas - Read Google Ads customers.generateKeywordIdeas. [intent=direct_read availability=implemented operation=google_ads.customers.generate.keyword.ideas]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --aggregate-metrics, --geo-target-constants, --historical-metrics-options, --include-adult-keywords, --keyword-and-url-seed, --keyword-annotation, --keyword-plan-network, --keyword-seed, --language, --page-size, --site-seed, --url-seed, --page, --page-cursor
  - customers generate-reach-forecast - Read Google Ads customers.generateReachForecast. [intent=direct_read availability=implemented operation=google_ads.customers.generate.reach.forecast]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --campaign-duration (required), --cookie-frequency-cap, --cookie-frequency-cap-setting, --currency-code, --customer-reach-group, --effective-frequency-limit, --forecast-metric-options, --min-effective-frequency, --planned-products (required), --reach-application-info, --targeting, --page, --page-cursor
  - customers generate-shareable-previews - Read Google Ads customers.generateShareablePreviews. [intent=direct_read availability=implemented operation=google_ads.customers.generate.shareable.previews]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --shareable-previews (required), --page, --page-cursor
  - customers generate-suggested-targeting-insights - Read Google Ads customers.generateSuggestedTargetingInsights. [intent=direct_read availability=implemented operation=google_ads.customers.generate.suggested.targeting.insights]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --audience-definition, --audience-description, --customer-insights-group, --insights-application-info, --page, --page-cursor
  - customers generate-targeting-suggestion-metrics - Read Google Ads customers.generateTargetingSuggestionMetrics. [intent=direct_read availability=implemented operation=google_ads.customers.generate.targeting.suggestion.metrics]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --audiences (required), --customer-insights-group, --insights-application-info, --page, --page-cursor
  - customers generate-trending-insights - Read Google Ads customers.generateTrendingInsights. [intent=direct_read availability=implemented operation=google_ads.customers.generate.trending.insights]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --country-location (required), --customer-insights-group (required), --insights-application-info, --search-audience, --search-topics, --page, --page-cursor
  - customers get-identity-verification - Read Google Ads customers.getIdentityVerification. [intent=direct_read availability=implemented operation=google_ads.customers.get.identity.verification]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --page, --page-cursor
  - customers invoices list - Read Google Ads customers.invoices.list. [intent=direct_read availability=implemented operation=google_ads.customers.invoices.list]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --page, --page-cursor
  - customers payments-accounts list - Read Google Ads customers.paymentsAccounts.list. [intent=direct_read availability=implemented operation=google_ads.customers.payments.accounts.list]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --page, --page-cursor
  - customers recommendations generate - Read Google Ads customers.recommendations.generate. [intent=direct_read availability=implemented operation=google_ads.customers.recommendations.generate]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --ad-group-info, --advertising-channel-type (required), --asset-group-info, --bidding-info, --budget-info, --campaign-call-asset-count, --campaign-image-asset-count, --campaign-sitelink-count, --conversion-tracking-status, --country-codes, --language-codes, --merchant-center-account-id, --negative-locations-ids, --positive-locations-ids, --recommendation-types (required), --seed-info, --target-content-network, --target-partner-search-network, --page, --page-cursor
  - customers search-audience-insights-attributes - Read Google Ads customers.searchAudienceInsightsAttributes. [intent=direct_read availability=implemented operation=google_ads.customers.search.audience.insights.attributes]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --customer-insights-group, --dimensions (required), --insights-application-info, --location-country-filters, --query-text (required), --youtube-reach-location, --page, --page-cursor
  - customers suggest-brands - Read Google Ads customers.suggestBrands. [intent=direct_read availability=implemented operation=google_ads.customers.suggest.brands]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --brand-prefix (required), --selected-brands, --page, --page-cursor
  - customers suggest-keyword-themes - Read Google Ads customers.suggestKeywordThemes. [intent=direct_read availability=implemented operation=google_ads.customers.suggest.keyword.themes]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --suggestion-info (required), --page, --page-cursor
  - customers suggest-smart-campaign-ad - Read Google Ads customers.suggestSmartCampaignAd. [intent=direct_read availability=implemented operation=google_ads.customers.suggest.smart.campaign.ad]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --suggestion-info (required), --page, --page-cursor
  - customers suggest-smart-campaign-budget-options - Read Google Ads customers.suggestSmartCampaignBudgetOptions. [intent=direct_read availability=implemented operation=google_ads.customers.suggest.smart.campaign.budget.options]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --campaign (required), --suggestion-info (required), --page, --page-cursor
  - customers suggest-travel-assets - Read Google Ads customers.suggestTravelAssets. [intent=direct_read availability=implemented operation=google_ads.customers.suggest.travel.assets]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --language-option (required), --place-ids, --page, --page-cursor
  - geo-target-constants suggest - Read Google Ads geoTargetConstants.suggest. [intent=direct_read availability=implemented operation=google_ads.geo.target.constants.suggest]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --country-code, --geo-targets, --locale, --location-names, --page, --page-cursor
  - keyword-theme-constants suggest - Read Google Ads keywordThemeConstants.suggest. [intent=direct_read availability=implemented operation=google_ads.keyword.theme.constants.suggest]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --country-code, --language-code, --query-text, --page, --page-cursor
  - v22 generate-conversion-rates - Read Google Ads v22.generateConversionRates. [intent=direct_read availability=implemented operation=google_ads.v22.generate.conversion.rates]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --customer-id (required), --customer-reach-group, --reach-application-info, --page, --page-cursor
  - v22 list-plannable-locations - Read Google Ads v22.listPlannableLocations. [intent=direct_read availability=implemented operation=google_ads.v22.list.plannable.locations]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --reach-application-info, --page, --page-cursor
  - v22 list-plannable-products - Read Google Ads v22.listPlannableProducts. [intent=direct_read availability=implemented operation=google_ads.v22.list.plannable.products]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --plannable-location-id (required), --reach-application-info, --page, --page-cursor
  - v22 list-plannable-user-interests - Read Google Ads v22.listPlannableUserInterests. [intent=direct_read availability=implemented operation=google_ads.v22.list.plannable.user.interests]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --customer-id (required), --name-query, --path-query, --reach-application-info, --user-interest-taxonomy-types, --page, --page-cursor
  - v22 list-plannable-user-lists - Read Google Ads v22.listPlannableUserLists. [intent=direct_read availability=implemented operation=google_ads.v22.list.plannable.user.lists]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --customer-id (required), --customer-reach-group, --reach-application-info, --page, --page-cursor

## Commands

### Inspect as a manual

```bash
pm connectors inspect google-ads
```

### Inspect as structured JSON

```bash
pm connectors inspect google-ads --json
```

## Agent Rules

- Run pm connectors inspect google-ads before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
