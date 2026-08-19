# pm connectors inspect mantle

```text
NAME
  pm connectors inspect mantle - Mantle connector manual

SYNOPSIS
  pm connectors inspect mantle
  pm connectors inspect mantle --json
  pm credentials add <name> --connector mantle [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Mantle Core API resources through the heymantle.com REST API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  agent_id
  app_event_id
  app_id
  base_url
  checklist_id
  collection_id
  event_name
  feature_id
  filter_id
  id
  item_id
  job_key
  loop_id
  message_id
  page_id
  plan_id
  reply_id
  repository_id
  resource_id
  resource_type
  review_id
  run_id
  skill_id
  usage_metric_id
  api_key (secret) (required)

ETL STREAMS
  customers:
    primary key: id
    cursor: updatedAt
    fields: averageMonthlyRevenue(number), countryCode(string), createdAt(string), domain(string), email(string), firstInteractionAt(string), id(string), industry(string), last30Revenue(number), lifetimeValue(number), name(string), shopifyDomain(string), shopifyShopId(string), tags(array), test(boolean), updatedAt(string)
  subscriptions:
    primary key: id
    cursor: createdAt
    fields: activatedAt(string), active(boolean), billingCycleAnchor(string), canceledAt(string), createdAt(string), currentPeriodEnd(string), currentPeriodStart(string), frozenAt(string), id(string), lineItems(array), plan(object), presentmentSubtotal(number), presentmentTotal(number), subtotal(number), total(number), trialExpiresAt(string), trialStartsAt(string)
  affiliate_commissions:
    primary key: id
    fields: affiliateId(string), affiliateProgramId(string), affiliateReferralId(string), amount(number), appInstallationId(string), cancelReason(string), cancelled(boolean), createdAt(string), date(string), id(string), name(string), notes(string), payoutId(string), transactionId(string), type(string), updatedAt(string)
  affiliate_commissions_id:
    primary key: id
    fields: affiliateId(string), affiliateProgramId(string), affiliateReferralId(string), amount(number), appInstallationId(string), cancelReason(string), cancelled(boolean), createdAt(string), date(string), id(string), name(string), notes(string), payoutId(string), transactionId(string), type(string), updatedAt(string)
  affiliate_payouts:
    primary key: id
    fields: affiliateId(string), affiliateProgramId(string), amount(number), amountPaid(number), createdAt(string), id(string), notes(string), number(integer), paidAt(string), paymentMethod(string), paymentMethodData(object), paymentRequestedAt(string), periodEnd(string), periodStart(string), status(string), updatedAt(string)
  affiliate_payouts_id:
    primary key: id
    fields: affiliateId(string), affiliateProgramId(string), amount(number), amountPaid(number), createdAt(string), id(string), notes(string), number(integer), paidAt(string), paymentMethod(string), paymentMethodData(object), paymentRequestedAt(string), periodEnd(string), periodStart(string), status(string), updatedAt(string)
  affiliate_programs:
    primary key: id
    fields: affiliatesCount(integer), allowAffiliateDashboardAccess(boolean), appId(string), createdAt(string), customAffiliateLink(string), groups(array), id(string), name(string), removeOnUninstallDays(integer), requireApprovalToJoin(boolean), requireTermsAcceptance(boolean), requiredAffiliateFields(object), rules(object), showInMarketplace(boolean), signupLink(string), startCountingCommissionFrom(string), termsUrl(string), updatedAt(string)
  affiliate_programs_id:
    primary key: id
    fields: affiliatesCount(integer), allowAffiliateDashboardAccess(boolean), appId(string), createdAt(string), customAffiliateLink(string), groups(array), id(string), name(string), removeOnUninstallDays(integer), requireApprovalToJoin(boolean), requireTermsAcceptance(boolean), requiredAffiliateFields(object), rules(object), showInMarketplace(boolean), signupLink(string), startCountingCommissionFrom(string), termsUrl(string), updatedAt(string)
  affiliate_referrals:
    primary key: id
    fields: affiliateId(string), affiliateProgramId(string), appId(string), appInstallationId(string), appListingPageViewId(string), createdAt(string), customerId(string), date(string), id(string), rules(object), updatedAt(string)
  affiliate_referrals_id:
    primary key: id
    fields: affiliateId(string), affiliateProgramId(string), appId(string), appInstallationId(string), appListingPageViewId(string), createdAt(string), customerId(string), date(string), id(string), rules(object), updatedAt(string)
  affiliates:
    primary key: id
    fields: agreedToTermsAt(string), createdAt(string), email(string), id(string), lastAppInstallationDate(string), lastTransactionDate(string), memberships(array), name(string), paypalEmail(string), startCountingCommissionFrom(string), tags(array), updatedAt(string), user(object)
  affiliates_id:
    fields: affiliate(object), attributions(array)
  agents:
    primary key: id
    fields: createdAt(string), email(string), id(string), name(string), organizationId(string), updatedAt(string), userId(string)
  ai_agents_agent_id_runs_run_id:
    primary key: id
    fields: agentId(string), completedAt(string), createdAt(string), error(string), id(string), response(string), status(string), structuredResponse(object), tokenUsage(object)
  api_core_v1_metrics_active_installs:
    fields: format(string), period(string), periodEnd(string), periodStart(string), timeSeries(array), timeSeriesInterval(string), value(number)
  api_core_v1_metrics_active_subscriptions:
    fields: format(string), period(string), periodEnd(string), periodStart(string), timeSeries(array), timeSeriesInterval(string), value(number)
  api_core_v1_metrics_arpu:
    fields: format(string), period(string), periodEnd(string), periodStart(string), timeSeries(array), timeSeriesInterval(string), value(number)
  api_core_v1_metrics_arr:
    fields: format(string), period(string), periodEnd(string), periodStart(string), timeSeries(array), timeSeriesInterval(string), value(number)
  api_core_v1_metrics_logo_churn:
    fields: format(string), period(string), periodEnd(string), periodStart(string), timeSeries(array), timeSeriesInterval(string), value(number)
  api_core_v1_metrics_mrr:
    fields: format(string), period(string), periodEnd(string), periodStart(string), timeSeries(array), timeSeriesInterval(string), value(number)
  api_core_v1_metrics_net_installs:
    fields: format(string), period(string), periodEnd(string), periodStart(string), timeSeries(array), timeSeriesInterval(string), value(number)
  api_core_v1_metrics_net_revenue:
    fields: format(string), period(string), periodEnd(string), periodStart(string), timeSeries(array), timeSeriesInterval(string), value(number)
  api_core_v1_metrics_net_revenue_retention:
    fields: format(string), period(string), periodEnd(string), periodStart(string), timeSeries(array), timeSeriesInterval(string), value(number)
  api_core_v1_metrics_payout:
    fields: billedUsageTotal(number), credits(number), dueTotal(number), fees(number), format(string), grossAmount(number), netAmount(number), period(string), periodEnd(string), periodStart(string), refunds(number), taxes(number), timeSeries(array), timeSeriesInterval(string), upcomingTotal(number)
  api_core_v1_metrics_predicted_ltv:
    fields: format(string), period(string), periodEnd(string), periodStart(string), timeSeries(array), timeSeriesInterval(string), value(number)
  api_core_v1_metrics_revenue_churn:
    fields: format(string), period(string), periodEnd(string), periodStart(string), timeSeries(array), timeSeriesInterval(string), value(number)
  api_core_v1_metrics_revenue_retention:
    fields: format(string), period(string), periodEnd(string), periodStart(string), timeSeries(array), timeSeriesInterval(string), value(number)
  api_core_v1_metrics_subscription_churn:
    fields: format(string), period(string), periodEnd(string), periodStart(string), timeSeries(array), timeSeriesInterval(string), value(number)
  api_core_v1_metrics_usage_event:
    fields: format(string), period(string), periodEnd(string), periodStart(string), timeSeries(array), timeSeriesInterval(string), value(number)
  api_core_v1_metrics_usage_metric:
    fields: format(string), method(string), period(string), periodEnd(string), periodStart(string), timeSeries(array), timeSeriesInterval(string), value(number)
  apps:
    primary key: id
    fields: apiClientId(string), createdAt(string), development(boolean), displayName(string), iconUrl(string), id(string), name(string), platform(string), slug(string)
  apps_app_id_checklists:
    primary key: id
    fields: createdAt(string), description(string), handle(string), id(string), name(string), status(string), steps(array), title(string), updatedAt(string)
  apps_app_id_checklists_checklist_id:
    primary key: id
    fields: createdAt(string), description(string), handle(string), id(string), name(string), status(string), steps(array), title(string), updatedAt(string)
  apps_app_id_plans_features:
    primary key: id
    fields: allowedValues(array), createdAt(string), defaultValue(string), description(string), id(string), key(string), name(string), type(string), updatedAt(string), usageMetric(object)
  apps_app_id_plans_features_feature_id:
    primary key: id
    fields: allowedValues(array), createdAt(string), defaultValue(string), description(string), id(string), key(string), name(string), type(string), updatedAt(string), usageMetric(object)
  apps_id:
    primary key: id
    fields: apiClientId(string), createdAt(string), development(boolean), displayName(string), iconUrl(string), id(string), name(string), platform(string), slug(string)
  apps_id_app_events:
    primary key: id
    fields: appId(string), appInstallationId(string), createdAt(string), customerId(string), id(string), occurredAt(string), previousSubscriptionId(string), subscriptionId(string), transactionId(string), type(string), updatedAt(string)
  apps_id_app_events_app_event_id:
    primary key: id
    fields: appId(string), appInstallationId(string), createdAt(string), customerId(string), id(string), occurredAt(string), previousSubscriptionId(string), subscriptionId(string), transactionId(string), type(string), updatedAt(string)
  apps_id_plans:
    primary key: id
    fields: activeTrialCount(number), amount(number), averageLifetimeValue(number), createdAt(string), currencyCode(string), customFields(object), customerExcludeTags(array), customerId(string), customerTags(array), deprecatedAt(string), description(string), features(object), flexBilling(boolean), flexBillingTerms(string), id(string), interval(string), lifetimeValue(number), monthlyRevenue(number), name(string), public(boolean), shopifyPlans(array), subscriberCount(number), trialDays(number), type(string), updatedAt(string), visible(boolean)
  apps_id_plans_plan_id:
    primary key: id
    fields: activeTrialCount(number), amount(number), averageLifetimeValue(number), createdAt(string), currencyCode(string), customFields(object), customerExcludeTags(array), customerId(string), customerTags(array), deprecatedAt(string), description(string), features(object), flexBilling(boolean), flexBillingTerms(string), id(string), interval(string), lifetimeValue(number), monthlyRevenue(number), name(string), public(boolean), shopifyPlans(array), subscriberCount(number), trialDays(number), type(string), updatedAt(string), visible(boolean)
  apps_id_reviews:
    primary key: id
    fields: archivedAt(string), content(string), createdAt(string), date(string), id(string), location(string), rating(number), updatedAt(string)
  apps_id_reviews_review_id:
    primary key: id
    fields: archivedAt(string), content(string), createdAt(string), date(string), id(string), location(string), rating(number), updatedAt(string)
  apps_id_skills_skill_id:
    fields: id(string), value(string)
  apps_id_usage_metrics:
    primary key: id
    fields: calculation(string), eventName(string), id(string), name(string), params(object)
  apps_id_usage_metrics_usage_metric_id:
    primary key: id
    fields: calculation(string), eventName(string), id(string), name(string), params(object)
  assistant_conversations_id:
    primary key: id
    fields: agent(object), createdAt(string), customData(object), id(string), messages(array), sessionId(string), summary(string), updatedAt(string), user(object)
  channels:
    primary key: id
    fields: createdAt(string), enabled(boolean), id(string), name(string), type(string), updatedAt(string)
  charges:
    primary key: id
    fields: amount(number), appName(string), billedOn(string), currencyCode(string), customerDomain(string), customerName(string), id(string), name(string), paidOn(string), status(string), test(boolean), type(string)
  companies:
    primary key: id
    fields: createdAt(string), customerIds(array), id(string), name(string), parentCustomerId(string), updatedAt(string)
  companies_id:
    primary key: id
    fields: createdAt(string), customerIds(array), id(string), name(string), parentCustomerId(string), updatedAt(string)
  contacts:
    primary key: id
    fields: createdAt(string), customers(array), email(string), id(string), jobTitle(string), label(string), name(string), notes(string), phone(string), secondaryEmails(array), socialProfiles(array), tags(array), updatedAt(string)
  contacts_id:
    primary key: id
    fields: createdAt(string), customers(array), email(string), id(string), jobTitle(string), label(string), name(string), notes(string), phone(string), secondaryEmails(array), socialProfiles(array), tags(array), updatedAt(string)
  custom_data:
    fields: customData(array)
  customer_segments:
    primary key: id
    fields: filters(array), id(string), name(string)
  customer_segments_id:
    primary key: id
    fields: filters(array), id(string), name(string)
  customers_custom_fields:
    primary key: id
    fields: appId(string), appLevel(boolean), defaultValue(), filterable(boolean), id(string), name(string), options(array), private(boolean), showOnCustomerDetail(boolean), type(string)
  customers_custom_fields_id:
    primary key: id
    fields: appId(string), appLevel(boolean), defaultValue(), filterable(boolean), id(string), name(string), options(array), private(boolean), showOnCustomerDetail(boolean), type(string)
  customers_id:
    primary key: id
    fields: accountOwners(array), appInstallations(array), billingAddress(object), companyId(string), contacts(array), createdAt(string), customFields(object), deals(array), domain(string), email(string), firstInteractionAt(string), id(string), industry(string), name(string), shopifyDomain(string), tags(array), updatedAt(string)
  customers_id_account_owners:
    fields: accountOwners(array)
  customers_id_timeline:
    primary key: id
    fields: affiliateAttribution(object), appCharge(object), appInstallationId(string), createdAt(string), customerId(string), email(object), id(string), metadata(object), netChange(number), occurredAt(string), planAmount(number), planChanges(object), planCurrencyCode(string), planInterval(string), previousSubscription(object), source(string), subscription(object), type(string), uninstallEvent(object), updatedAt(string), utm(object)
  deal_activities:
    primary key: id
    fields: createdAt(string), defaultWeight(number), description(string), icon(string), id(string), name(string), timelineDescriptionTemplate(string), timelineTitleTemplate(string), updatedAt(string)
  deal_activities_id:
    primary key: id
    fields: createdAt(string), defaultWeight(number), description(string), icon(string), id(string), name(string), timelineDescriptionTemplate(string), timelineTitleTemplate(string), updatedAt(string)
  deal_flows:
    primary key: id
    fields: acquirer(object), affiliate(object), createdAt(string), dealStages(array), defaultAcquisitionChannel(string), defaultAcquisitionSource(), defaultDealOwner(object), deletedAt(string), description(string), displayOrder(integer), id(string), name(string), partnership(object), updatedAt(string)
  deal_flows_id:
    primary key: id
    fields: acquirer(object), affiliate(object), createdAt(string), dealStages(array), defaultAcquisitionChannel(string), defaultAcquisitionSource(), defaultDealOwner(object), deletedAt(string), description(string), displayOrder(integer), id(string), name(string), partnership(object), updatedAt(string)
  deals:
    primary key: id
    fields: acquirer(object), acquisitionChannel(string), acquisitionSource(string), affiliate(object), amount(number), amountCurrencyCode(string), app(object), archivedAt(string), closedAt(string), closingAt(string), contacts(array), createdAt(string), currentAmount(number), customData(array), customer(object), dealFlow(object), dealStage(object), firstInteractionAt(string), id(string), name(string), notes(string), owners(array), partnership(object), plan(object), updatedAt(string)
  deals_id:
    primary key: id
    fields: acquirer(object), acquisitionChannel(string), acquisitionSource(string), affiliate(object), amount(number), amountCurrencyCode(string), app(object), archivedAt(string), closedAt(string), closingAt(string), contacts(array), createdAt(string), currentAmount(number), customData(array), customer(object), dealFlow(object), dealStage(object), firstInteractionAt(string), id(string), name(string), notes(string), owners(array), partnership(object), plan(object), updatedAt(string)
  deals_id_events:
    primary key: id
    fields: createdAt(string), dealActivity(object), dealStage(object), id(string), notes(string), occurredAt(string), previousDealStage(object), task(object), timelineComment(), type(string), user(object)
  deals_id_timeline:
    primary key: id
    fields: createdAt(string), dealActivity(object), dealStage(object), id(string), notes(string), occurredAt(string), previousDealStage(object), task(object), timelineComment(), type(string), user(object)
  docs_collections:
    primary key: id
    fields: createdAt(string), description(string), displayOrder(integer), groups(array), handles(array), id(string), status(string), title(string), updatedAt(string)
  docs_groups:
    primary key: id
    fields: createdAt(string), displayOrder(integer), handle(string), id(string), pages(array), status(string), title(string), updatedAt(string)
  docs_pages:
    primary key: id
    fields: children(array), depth(integer), displayOrder(integer), handle(string), id(string), path(string), publishedAt(string), title(string)
  docs_pages_generate_status_job_key:
    fields: id(string), status(string)
  docs_pages_page_id:
    primary key: id
    fields: allLocales(array), children(array), collection(object), content(string), createdAt(string), depth(integer), displayOrder(integer), faqs(array), files(array), group(object), handle(string), id(string), locale(string), openGraphImage(string), path(string), publishedAt(string), seoDescription(string), seoTitle(string), title(string), updatedAt(string)
  docs_repositories:
    primary key: id
    fields: collections(array), config(object), createdAt(string), customDomain(string), defaultLocale(string), handle(string), id(string), locales(array), shortDescription(string), supportedLocales(array), title(string), updatedAt(string), url(string), useCustomDomain(boolean), visibility(string)
  docs_repositories_id:
    primary key: id
    fields: collections(array), config(object), createdAt(string), customDomain(string), defaultLocale(string), handle(string), id(string), locales(array), shortDescription(string), supportedLocales(array), title(string), updatedAt(string), url(string), useCustomDomain(boolean), visibility(string)
  docs_sites:
    primary key: id
    fields: config(object), createdAt(string), customDomain(string), customDomainVerifiedAt(string), defaultLocale(string), handle(string), id(string), locales(array), repositories(array), shortDescription(string), supportedLocales(array), title(string), updatedAt(string), url(string), useCustomDomain(boolean), visibility(string), widgetId(string)
  docs_sites_id:
    primary key: id
    fields: config(object), createdAt(string), customDomain(string), customDomainVerifiedAt(string), defaultLocale(string), handle(string), id(string), locales(array), repositories(array), shortDescription(string), supportedLocales(array), title(string), updatedAt(string), url(string), useCustomDomain(boolean), visibility(string), widgetId(string)
  docs_sites_id_redirects:
    primary key: id
    fields: createdAt(string), enabled(boolean), fromPath(string), id(string), notes(string), redirectType(string), toPath(string), updatedAt(string)
  docs_sites_id_repositories:
    primary key: id
    fields: createdAt(string), id(string), position(integer), repository(object), repositoryId(string)
  docs_tree:
    primary key: id
    fields: collections(array), config(object), createdAt(string), customDomain(string), defaultLocale(string), handle(string), id(string), locales(array), shortDescription(string), supportedLocales(array), title(string), updatedAt(string), url(string), useCustomDomain(boolean), visibility(string)
  email_campaigns:
    primary key: id
    fields: appId(string), clickCount(integer), createdAt(string), deliveredCount(integer), format(string), html(string), id(string), layout(object), layoutId(string), name(string), openCount(integer), plainText(string), previewText(string), sender(object), senderId(string), sentCount(integer), status(string), subject(string), thumbnail(string), type(string), unsubscribeGroup(object), unsubscribeGroupId(string), updatedAt(string)
  email_campaigns_id:
    primary key: id
    fields: appId(string), clickCount(integer), createdAt(string), deliveredCount(integer), format(string), html(string), id(string), layout(object), layoutId(string), name(string), openCount(integer), plainText(string), previewText(string), sender(object), senderId(string), sentCount(integer), status(string), subject(string), thumbnail(string), type(string), unsubscribeGroup(object), unsubscribeGroupId(string), updatedAt(string)
  email_campaigns_id_preview:
    fields: customerId(string), emailId(string), html(string), previewText(string), subject(string), warning(string)
  email_deliveries:
    primary key: id
    fields: audienceSize(integer), createdAt(string), email(object), emailId(string), id(string), stats(object), status(string)
  email_deliveries_id:
    primary key: id
    fields: audienceCriteria(object), audienceSize(integer), createdAt(string), email(object), emailId(string), error(string), id(string), recentDeliveries(array), recentEvents(array), stats(object), status(string), updatedAt(string)
  email_layouts:
    primary key: id
    fields: appId(string), createdAt(string), defaultLayout(boolean), format(string), id(string), name(string), thumbnail(string), updatedAt(string)
  email_senders:
    primary key: id
    fields: address(object), appId(string), createdAt(string), from(string), id(string), name(string), replyTo(string), updatedAt(string)
  email_unsubscribe_groups:
    primary key: id
    fields: createdAt(string), description(string), id(string), name(string)
  email_unsubscribe_groups_id_members:
    primary key: id
    fields: dateAdded(string), email(string), id(string)
  entities:
  flow_extensions_actions:
    primary key: id
    fields: address(string), createdAt(string), description(string), handle(string), id(string), name(string), organizationId(string), settingsSchema(object), updatedAt(string)
  flows:
    primary key: id
    fields: allowRepeatRuns(boolean), blockRepeatsTimeUnit(string), blockRepeatsTimeValue(integer), createdAt(string), id(string), name(string), status(string), stepsCount(integer), updatedAt(string)
  flows_id:
    primary key: id
    fields: allowRepeatRuns(boolean), blockRepeatsTimeUnit(string), blockRepeatsTimeValue(integer), createdAt(string), firstStepId(string), id(string), name(string), status(string), steps(array), stepsCount(integer), updatedAt(string)
  journal_entries:
    primary key: id
    fields: app(object), appId(string), createdAt(string), date(string), description(string), emoji(string), files(array), id(string), tags(array), title(string), updatedAt(string), url(string)
  journal_entries_id:
    primary key: id
    fields: app(object), appId(string), createdAt(string), date(string), description(string), emoji(string), files(array), id(string), tags(array), title(string), updatedAt(string), url(string)
  lists:
    primary key: id
    fields: contactCount(integer), createdAt(string), customerCount(integer), description(string), id(string), name(string), updatedAt(string)
  lists_id:
  meetings:
    primary key: id
    fields: aiDealInsights(object), aiDecisions(array), aiEnrichedAt(string), aiEnrichmentStatus(string), aiKeyPoints(array), aiOpenQuestions(array), aiSummary(string), aiTopics(array), attendees(array), createdAt(string), createdBy(object), customer(object), deal(object), duration(integer), endTime(string), externalId(string), id(string), meetingUrl(string), overallSentiment(number), platform(string), platformMeetingId(string), recordingStatus(string), recordingUrl(string), sentimentTrend(array), startTime(string), summary(string), taskSuggestions(array), title(string), transcript(object), updatedAt(string)
  meetings_id:
    primary key: id
    fields: aiDealInsights(object), aiDecisions(array), aiEnrichedAt(string), aiEnrichmentStatus(string), aiKeyPoints(array), aiOpenQuestions(array), aiSummary(string), aiTopics(array), attendees(array), createdAt(string), createdBy(object), customer(object), deal(object), duration(integer), endTime(string), externalId(string), id(string), meetingUrl(string), overallSentiment(number), platform(string), platformMeetingId(string), recordingStatus(string), recordingUrl(string), sentimentTrend(array), startTime(string), summary(string), taskSuggestions(array), title(string), transcript(object), updatedAt(string)
  meetings_id_permissions:
    primary key: id
    fields: createdAt(string), grantedBy(object), id(string), user(object), userId(string)
  meetings_id_recording_url:
    fields: expiresIn(integer), recordingUrl(string)
  meetings_id_transcribe:
    primary key: id
    fields: externalId(), id(), status()
  metrics_sales:
    fields: id(string), value(number)
  notification_preferences:
    fields: configured(boolean), id(string)
  organization:
    primary key: id
    fields: contactTags(array), currentInstallation(object), customerTags(array), id(string), name(string)
  subscriptions_id:
    primary key: id
    fields: activatedAt(string), active(boolean), appliedDiscount(object), billingCycleAnchor(string), cancelOn(string), canceledAt(string), confirmationUrl(string), createdAt(string), currentPeriodEnd(string), currentPeriodStart(string), features(array), featuresOrder(array), frozenAt(string), id(string), lineItems(array), plan(object), presentmentSubtotal(number), presentmentTotal(number), shopifySubscription(object), subtotal(number), total(number), trialExpiresAt(string), trialStartsAt(string), usageBalanceUsed(number), usageCappedAmount(number)
  synced_emails:
    primary key: id
    fields: contact(object), createdAt(string), customer(object), deal(object), externalId(string), id(string), lastMessageAt(string), messageCount(integer), messages(array), snippet(string), source(string), subject(string), syncedBy(object), updatedAt(string)
  synced_emails_id:
    primary key: id
    fields: contact(object), createdAt(string), customer(object), deal(object), externalId(string), id(string), lastMessageAt(string), messageCount(integer), messages(array), snippet(string), source(string), subject(string), syncedBy(object), updatedAt(string)
  synced_emails_id_messages:
    fields: messages(array)
  tasks:
    primary key: id
    fields: assignee(object), completedAt(string), contact(object), contactId(string), createdAt(string), createdBy(object), customer(object), customerId(string), deal(object), dealActivity(object), dealActivityId(string), dealId(string), description(string), descriptionHtml(string), dueDate(string), hasComments(boolean), id(string), priority(string), status(string), tags(array), title(string), todoItems(array), updatedAt(string)
  tasks_id:
    primary key: id
    fields: assignee(object), completedAt(string), contact(object), contactId(string), createdAt(string), createdBy(object), customer(object), customerId(string), deal(object), dealActivity(object), dealActivityId(string), dealId(string), description(string), descriptionHtml(string), dueDate(string), hasComments(boolean), id(string), priority(string), status(string), tags(array), title(string), todoItems(array), updatedAt(string)
  tasks_id_comments:
    primary key: id
    fields: comment(string), commentHtml(string), createdAt(string), createdBy(object), id(string), taggedUsers(array), taskId(string), updatedAt(string)
  tasks_id_todo_items:
    primary key: id
    fields: completed(boolean), completedAt(string), content(string), displayOrder(integer), id(string)
  tasks_id_todo_items_item_id:
    primary key: id
    fields: completed(boolean), completedAt(string), content(string), displayOrder(integer), id(string)
  tickets:
    primary key: id
    fields: app(object), assignedTo(object), channel(object), closedAt(string), contact(object), createdAt(string), customer(object), cxEmailAddressId(string), firstResponseAt(string), id(string), inboxId(string), lastMessage(object), lastMessageAt(string), messageCount(integer), priority(string), resolvedAt(string), sourceId(string), sourceType(string), status(string), subject(string), tags(array), ticketNumber(string), updatedAt(string)
  tickets_id:
    primary key: id
    fields: app(object), assignedTo(object), channel(object), closedAt(string), contact(object), createdAt(string), customer(object), cxEmailAddressId(string), firstResponseAt(string), id(string), inboxId(string), lastMessage(object), lastMessageAt(string), messageCount(integer), priority(string), resolvedAt(string), sourceId(string), sourceType(string), status(string), subject(string), tags(array), ticketNumber(string), updatedAt(string)
  tickets_id_events:
    primary key: id
    fields: actorType(string), agent(object), contact(object), createdAt(string), id(string), newValue(object), occurredAt(string), oldValue(object), threadMessageId(string), type(string)
  tickets_id_loops:
    primary key: id
    fields: category(object), categoryId(string), closedAt(string), closedByMessageId(string), closedReason(string), id(string), openedAt(string), openedByMessageId(string), question(string), status(string)
  tickets_id_loops_loop_id:
    primary key: id
    fields: category(object), categoryId(string), closedAt(string), closedByMessageId(string), closedReason(string), id(string), openedAt(string), openedByMessageId(string), question(string), status(string)
  tickets_id_messages:
    primary key: id
    fields: actorType(string), agent(object), attachments(array), contact(object), content(string), contentType(string), createdAt(string), fullContent(string), id(string), inReplyToId(string), isInternal(boolean), messageId(string), occurredAt(string), referencesIds(array)
  tickets_id_messages_message_id:
    primary key: id
    fields: actorType(string), agent(object), attachments(array), contact(object), content(string), contentType(string), createdAt(string), fullContent(string), id(string), inReplyToId(string), isInternal(boolean), messageId(string), occurredAt(string), referencesIds(array)
  tickets_saved_filters:
    primary key: id
    fields: createdAt(string), displayOrder(integer), filters(object), id(string), name(string), updatedAt(string), userId(string)
  tickets_saved_filters_filter_id:
    primary key: id
    fields: createdAt(string), displayOrder(integer), filters(object), id(string), name(string), updatedAt(string), userId(string)
  tickets_saved_replies:
    primary key: id
    fields: appId(string), category(object), categoryId(string), content(object), createdAt(string), defaultLocale(string), handle(string), id(string), roles(array), status(string), tags(array), title(string), updatedAt(string), user(object), userId(string)
  tickets_saved_replies_reply_id:
    primary key: id
    fields: appId(string), category(object), categoryId(string), content(object), createdAt(string), defaultLocale(string), handle(string), id(string), roles(array), status(string), tags(array), title(string), updatedAt(string), user(object), userId(string)
  timeline_comments:
    primary key: id
    fields: appInstallationId(string), attachments(array), comment(string), commentHtml(string), createdAt(string), customerId(string), dealId(string), id(string), originalCommentId(string), taggedUsers(array), updatedAt(string), user(object), userId(string)
  timeline_comments_id:
    primary key: id
    fields: appInstallationId(string), attachments(array), comment(string), commentHtml(string), createdAt(string), customerId(string), dealId(string), id(string), originalCommentId(string), taggedUsers(array), updatedAt(string), user(object), userId(string)
  transactions:
    primary key: id
    fields: appId(string), appInstallationId(string), billingProvider(string), billingProviderId(string), createdAt(string), customerId(string), date(string), grossAmount(number), grossAmountCurrencyCode(string), id(string), netAmount(number), netAmountCurrencyCode(string), processingFee(number), processingFeeCurrencyCode(string), subscriptionId(string), type(string), updatedAt(string)
  transactions_id:
    primary key: id
    fields: appId(string), appInstallationId(string), billingProvider(string), billingProviderId(string), createdAt(string), customerId(string), date(string), grossAmount(number), grossAmountCurrencyCode(string), id(string), netAmount(number), netAmountCurrencyCode(string), processingFee(number), processingFeeCurrencyCode(string), subscriptionId(string), type(string), updatedAt(string)
  usage_events:
    fields: eventId(string), eventName(string), properties(object), timestamp(string)
  users:
    primary key: id
    fields: email(string), id(string), jobTitle(string), name(string), roles(array)
  users_id:
    primary key: id
    fields: email(string), id(string), jobTitle(string), name(string), roles(array)
  webhooks:
    primary key: id
    fields: address(string), appIds(array), createdAt(string), filter(object), id(string), topic(string), updatedAt(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_agents:
    endpoint: POST /v1/agents
    required fields: email
    risk: medium: external Mantle mutation; approval required
  create_ai_agents_agent_id_runs:
    endpoint: POST /v1/ai/agents/{{ record.agent_id }}/runs
    required fields: agent_id, prompt
    risk: medium: external Mantle side effect; approval required
  create_apps_app_id_checklists:
    endpoint: POST /v1/apps/{{ record.app_id }}/checklists
    required fields: app_id, name
    risk: medium: external Mantle mutation; approval required
  create_apps_app_id_plans_features:
    endpoint: POST /v1/apps/{{ record.app_id }}/plans/features
    required fields: app_id
    risk: medium: external Mantle mutation; approval required
  create_apps_id_app_events:
    endpoint: POST /v1/apps/{{ record.id }}/app_events
    required fields: id, type, customerId
    risk: medium: external Mantle mutation; approval required
  create_apps_id_plans:
    endpoint: POST /v1/apps/{{ record.id }}/plans
    required fields: id
    risk: medium: external Mantle mutation; approval required
  create_apps_id_usage_metrics:
    endpoint: POST /v1/apps/{{ record.id }}/usage_metrics
    required fields: id
    risk: medium: external Mantle mutation; approval required
  create_attachments:
    endpoint: POST /v1/attachments
    required fields: filename
    risk: medium: external Mantle mutation; approval required
  create_channels:
    endpoint: POST /v1/channels
    required fields: type, name
    risk: medium: external Mantle mutation; approval required
  create_companies:
    endpoint: POST /v1/companies
    risk: medium: external Mantle mutation; approval required
  create_customers:
    endpoint: POST /v1/customers
    risk: medium: external Mantle mutation; approval required
  create_customers_custom_fields:
    endpoint: POST /v1/customers/custom_fields
    risk: medium: external Mantle mutation; approval required
  create_deal_activities:
    endpoint: POST /v1/deal_activities
    required fields: name
    risk: medium: external Mantle mutation; approval required
  create_deal_flows:
    endpoint: POST /v1/deal_flows
    required fields: name, dealStages
    risk: medium: external Mantle side effect; approval required
  create_deals:
    endpoint: POST /v1/deals
    required fields: name
    risk: medium: external Mantle mutation; approval required
  create_deals_id_events:
    endpoint: POST /v1/deals/{{ record.id }}/events
    required fields: id
    risk: medium: external Mantle mutation; approval required
  create_docs_collections:
    endpoint: POST /v1/docs/collections
    required fields: repositoryId, handle, title
    risk: medium: external Mantle mutation; approval required
  create_docs_groups:
    endpoint: POST /v1/docs/groups
    required fields: repositoryId, collectionId, handle, title
    risk: medium: external Mantle mutation; approval required
  create_docs_pages:
    endpoint: POST /v1/docs/pages
    required fields: repositoryId, groupId, handle, title
    risk: medium: external Mantle mutation; approval required
  create_docs_sites:
    endpoint: POST /v1/docs/sites
    required fields: handle, title
    risk: medium: external Mantle mutation; approval required
  create_docs_sites_id_redirects:
    endpoint: POST /v1/docs/sites/{{ record.id }}/redirects
    required fields: id, redirects
    risk: medium: external Mantle mutation; approval required
  create_email_campaigns:
    endpoint: POST /v1/email/campaigns
    required fields: name
    risk: medium: external Mantle side effect; approval required
  create_flow_extensions_actions:
    endpoint: POST /v1/flow/extensions/actions
    risk: medium: external Mantle side effect; approval required
  create_flows:
    endpoint: POST /v1/flows
    required fields: name
    risk: medium: external Mantle side effect; approval required
  create_journal_entries:
    endpoint: POST /v1/journal_entries
    required fields: date, description
    risk: medium: external Mantle mutation; approval required
  create_lists:
    endpoint: POST /v1/lists
    required fields: name
    risk: medium: external Mantle mutation; approval required
  create_meetings:
    endpoint: POST /v1/meetings
    risk: medium: external Mantle mutation; approval required
  create_meetings_id_transcribe_upload:
    endpoint: POST /v1/meetings/{{ record.id }}/transcribe/upload
    required fields: id
    risk: medium: external Mantle mutation; approval required
  create_tasks:
    endpoint: POST /v1/tasks
    required fields: title
    risk: medium: external Mantle mutation; approval required
  create_tasks_id_comments:
    endpoint: POST /v1/tasks/{{ record.id }}/comments
    required fields: id
    risk: medium: external Mantle mutation; approval required
  create_tasks_id_todo_items:
    endpoint: POST /v1/tasks/{{ record.id }}/todo-items
    required fields: id, content
    risk: medium: external Mantle mutation; approval required
  create_tickets:
    endpoint: POST /v1/tickets
    required fields: subject
    risk: medium: external Mantle mutation; approval required
  create_tickets_id_events:
    endpoint: POST /v1/tickets/{{ record.id }}/events
    required fields: id, type, actorType
    risk: medium: external Mantle mutation; approval required
  create_tickets_id_messages:
    endpoint: POST /v1/tickets/{{ record.id }}/messages
    required fields: id
    risk: medium: external Mantle mutation; approval required
  create_tickets_saved_filters:
    endpoint: POST /v1/tickets/saved_filters
    required fields: name
    risk: medium: external Mantle mutation; approval required
  create_tickets_saved_replies:
    endpoint: POST /v1/tickets/saved_replies
    required fields: title
    risk: medium: external Mantle mutation; approval required
  create_timeline_comments:
    endpoint: POST /v1/timeline_comments
    required fields: commentHtml
    risk: medium: external Mantle mutation; approval required
  create_usage_events:
    endpoint: POST /v1/usage_events
    risk: medium: external Mantle mutation; approval required
  create_webhooks:
    endpoint: POST /v1/webhooks
    risk: medium: external Mantle side effect; approval required
  delete_apps_app_id_checklists_checklist_id:
    endpoint: DELETE /v1/apps/{{ record.app_id }}/checklists/{{ record.checklist_id }}
    required fields: app_id, checklist_id
    risk: high: external Mantle mutation or side effect; approval required
  delete_apps_app_id_plans_features_feature_id:
    endpoint: DELETE /v1/apps/{{ record.app_id }}/plans/features/{{ record.feature_id }}
    required fields: app_id, feature_id
    risk: high: external Mantle mutation or side effect; approval required
  delete_apps_id_usage_metrics_usage_metric_id:
    endpoint: DELETE /v1/apps/{{ record.id }}/usage_metrics/{{ record.usage_metric_id }}
    required fields: id, usage_metric_id
    risk: high: external Mantle mutation or side effect; approval required
  delete_companies_id:
    endpoint: DELETE /v1/companies/{{ record.id }}
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  delete_contacts_id:
    endpoint: DELETE /v1/contacts/{{ record.id }}
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  delete_customers_custom_fields_id:
    endpoint: DELETE /v1/customers/custom_fields/{{ record.id }}
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  delete_customers_id_account_owners_owner_id:
    endpoint: DELETE /v1/customers/{{ record.id }}/account_owners/{{ record.owner_id }}
    required fields: id, owner_id
    risk: high: external Mantle mutation or side effect; approval required
  delete_deal_activities_id:
    endpoint: DELETE /v1/deal_activities/{{ record.id }}
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  delete_deal_flows_id:
    endpoint: DELETE /v1/deal_flows/{{ record.id }}
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  delete_deals_id:
    endpoint: DELETE /v1/deals/{{ record.id }}
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  delete_docs_collections_collection_id:
    endpoint: DELETE /v1/docs/collections/{{ record.collection_id }}
    required fields: collection_id
    risk: high: external Mantle mutation or side effect; approval required
  delete_docs_groups_group_id:
    endpoint: DELETE /v1/docs/groups/{{ record.group_id }}
    required fields: group_id
    risk: high: external Mantle mutation or side effect; approval required
  delete_docs_pages_page_id:
    endpoint: DELETE /v1/docs/pages/{{ record.page_id }}
    required fields: page_id
    risk: high: external Mantle mutation or side effect; approval required
  delete_docs_pages_page_id_archive:
    endpoint: DELETE /v1/docs/pages/{{ record.page_id }}/archive
    required fields: page_id
    risk: high: external Mantle mutation or side effect; approval required
  delete_docs_pages_page_id_publish:
    endpoint: DELETE /v1/docs/pages/{{ record.page_id }}/publish
    required fields: page_id
    risk: high: external Mantle mutation or side effect; approval required
  delete_docs_sites_id:
    endpoint: DELETE /v1/docs/sites/{{ record.id }}
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  delete_docs_sites_id_redirects_redirect_id:
    endpoint: DELETE /v1/docs/sites/{{ record.id }}/redirects/{{ record.redirect_id }}
    required fields: id, redirect_id
    risk: high: external Mantle mutation or side effect; approval required
  delete_docs_sites_id_repositories:
    endpoint: DELETE /v1/docs/sites/{{ record.id }}/repositories
    required fields: id, repositoryId
    risk: high: external Mantle mutation or side effect; approval required
  delete_email_campaigns_id:
    endpoint: DELETE /v1/email/campaigns/{{ record.id }}
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  delete_email_unsubscribe_groups_id_members:
    endpoint: DELETE /v1/email/unsubscribe_groups/{{ record.id }}/members
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  delete_email_unsubscribe_groups_id_members_member_id:
    endpoint: DELETE /v1/email/unsubscribe_groups/{{ record.id }}/members/{{ record.member_id }}
    required fields: id, member_id
    risk: high: external Mantle mutation or side effect; approval required
  delete_flow_extensions_actions_id:
    endpoint: DELETE /v1/flow/extensions/actions/{{ record.id }}
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  delete_flow_extensions_triggers_handle:
    endpoint: DELETE /v1/flow/extensions/triggers/{{ record.handle }}
    required fields: handle
    risk: high: external Mantle mutation or side effect; approval required
  delete_flows_id:
    endpoint: DELETE /v1/flows/{{ record.id }}
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  delete_journal_entries_id:
    endpoint: DELETE /v1/journal_entries/{{ record.id }}
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  delete_lists_id:
    endpoint: DELETE /v1/lists/{{ record.id }}
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  delete_meetings_id:
    endpoint: DELETE /v1/meetings/{{ record.id }}
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  delete_meetings_id_permissions:
    endpoint: DELETE /v1/meetings/{{ record.id }}/permissions
    required fields: id, userId
    risk: high: external Mantle mutation or side effect; approval required
  delete_synced_emails_id:
    endpoint: DELETE /v1/synced_emails/{{ record.id }}
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  delete_tasks_id:
    endpoint: DELETE /v1/tasks/{{ record.id }}
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  delete_tasks_id_comments_comment_id:
    endpoint: DELETE /v1/tasks/{{ record.id }}/comments/{{ record.comment_id }}
    required fields: id, comment_id
    risk: high: external Mantle mutation or side effect; approval required
  delete_tasks_id_todo_items_item_id:
    endpoint: DELETE /v1/tasks/{{ record.id }}/todo-items/{{ record.item_id }}
    required fields: id, item_id
    risk: high: external Mantle mutation or side effect; approval required
  delete_tickets_saved_filters_filter_id:
    endpoint: DELETE /v1/tickets/saved_filters/{{ record.filter_id }}
    required fields: filter_id
    risk: high: external Mantle mutation or side effect; approval required
  delete_tickets_saved_replies_reply_id:
    endpoint: DELETE /v1/tickets/saved_replies/{{ record.reply_id }}
    required fields: reply_id
    risk: high: external Mantle mutation or side effect; approval required
  delete_timeline_comments_id:
    endpoint: DELETE /v1/timeline_comments/{{ record.id }}
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  delete_webhooks_id:
    endpoint: DELETE /v1/webhooks/{{ record.id }}
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  execute_affiliates_id_add_tags:
    endpoint: POST /v1/affiliates/{{ record.id }}/addTags
    required fields: id
    risk: medium: external Mantle mutation; approval required
  execute_affiliates_id_remove_tags:
    endpoint: POST /v1/affiliates/{{ record.id }}/removeTags
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  execute_apps_id_analyze:
    endpoint: POST /v1/apps/{{ record.id }}/analyze
    required fields: id
    risk: medium: external Mantle mutation; approval required
  execute_apps_id_skills_skill_id:
    endpoint: POST /v1/apps/{{ record.id }}/skills/{{ record.skill_id }}
    required fields: id, skill_id, targets
    risk: medium: external Mantle mutation; approval required
  execute_contacts_id_add_tags:
    endpoint: POST /v1/contacts/{{ record.id }}/addTags
    required fields: id
    risk: medium: external Mantle mutation; approval required
  execute_contacts_id_remove_tags:
    endpoint: POST /v1/contacts/{{ record.id }}/removeTags
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  execute_customers_id_account_owners:
    endpoint: POST /v1/customers/{{ record.id }}/account_owners
    required fields: id, userId, type
    risk: medium: external Mantle mutation; approval required
  execute_customers_id_add_tags:
    endpoint: POST /v1/customers/{{ record.id }}/addTags
    required fields: id
    risk: medium: external Mantle mutation; approval required
  execute_customers_id_remove_tags:
    endpoint: POST /v1/customers/{{ record.id }}/removeTags
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  execute_devices:
    endpoint: POST /v1/devices
    required fields: token, platform
    risk: medium: external Mantle mutation; approval required
  execute_docs_pages_generate:
    endpoint: POST /v1/docs/pages/generate
    required fields: repositoryId, prompt
    risk: medium: external Mantle side effect; approval required
  execute_docs_pages_page_id_archive:
    endpoint: POST /v1/docs/pages/{{ record.page_id }}/archive
    required fields: page_id
    risk: high: external Mantle mutation or side effect; approval required
  execute_docs_pages_page_id_generate:
    endpoint: POST /v1/docs/pages/{{ record.page_id }}/generate
    required fields: page_id, prompt
    risk: medium: external Mantle side effect; approval required
  execute_docs_pages_page_id_publish:
    endpoint: POST /v1/docs/pages/{{ record.page_id }}/publish
    required fields: page_id, repositoryId
    risk: high: external Mantle mutation or side effect; approval required
  execute_docs_sites_id_repositories:
    endpoint: POST /v1/docs/sites/{{ record.id }}/repositories
    required fields: id, repositoryId
    risk: medium: external Mantle mutation; approval required
  execute_email_campaigns_id_cancel:
    endpoint: POST /v1/email/campaigns/{{ record.id }}/cancel
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  execute_email_campaigns_id_deliver:
    endpoint: POST /v1/email/campaigns/{{ record.id }}/deliver
    required fields: id, customerId
    risk: high: external Mantle mutation or side effect; approval required
  execute_email_campaigns_id_send:
    endpoint: POST /v1/email/campaigns/{{ record.id }}/send
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  execute_email_campaigns_id_test:
    endpoint: POST /v1/email/campaigns/{{ record.id }}/test
    required fields: id, email
    risk: high: external Mantle mutation or side effect; approval required
  execute_email_unsubscribe_groups_id_members:
    endpoint: POST /v1/email/unsubscribe_groups/{{ record.id }}/members
    required fields: id, emails
    risk: medium: external Mantle side effect; approval required
  execute_lists_id_add:
    endpoint: POST /v1/lists/{{ record.id }}/add
    required fields: id
    risk: medium: external Mantle mutation; approval required
  execute_lists_id_remove:
    endpoint: POST /v1/lists/{{ record.id }}/remove
    required fields: id
    risk: high: external Mantle mutation or side effect; approval required
  execute_meetings_id_permissions:
    endpoint: POST /v1/meetings/{{ record.id }}/permissions
    required fields: id, userId
    risk: medium: external Mantle mutation; approval required
  execute_meetings_id_task_suggestions_suggestion_id_accept:
    endpoint: POST /v1/meetings/{{ record.id }}/task-suggestions/{{ record.suggestion_id }}/accept
    required fields: id, suggestion_id
    risk: medium: external Mantle side effect; approval required
  execute_meetings_id_task_suggestions_suggestion_id_dismiss:
    endpoint: POST /v1/meetings/{{ record.id }}/task-suggestions/{{ record.suggestion_id }}/dismiss
    required fields: id, suggestion_id
    risk: medium: external Mantle side effect; approval required
  execute_meetings_id_transcribe:
    endpoint: POST /v1/meetings/{{ record.id }}/transcribe
    required fields: id, recordingKey
    risk: medium: external Mantle mutation; approval required
  execute_synced_emails:
    endpoint: POST /v1/synced_emails
    risk: medium: external Mantle side effect; approval required
  execute_synced_emails_id_messages:
    endpoint: POST /v1/synced_emails/{{ record.id }}/messages
    required fields: id, messages
    risk: medium: external Mantle side effect; approval required
  execute_tickets_id_ai_replies:
    endpoint: POST /v1/tickets/{{ record.id }}/ai-replies
    required fields: id
    risk: medium: external Mantle side effect; approval required
  update_apps_app_id_checklists_checklist_id:
    endpoint: PUT /v1/apps/{{ record.app_id }}/checklists/{{ record.checklist_id }}
    required fields: app_id, checklist_id
    risk: medium: external Mantle mutation; approval required
  update_apps_app_id_plans_features_feature_id:
    endpoint: PUT /v1/apps/{{ record.app_id }}/plans/features/{{ record.feature_id }}
    required fields: app_id, feature_id
    risk: medium: external Mantle mutation; approval required
  update_apps_id_plans_plan_id:
    endpoint: PUT /v1/apps/{{ record.id }}/plans/{{ record.plan_id }}
    required fields: id, plan_id
    risk: medium: external Mantle mutation; approval required
  update_apps_id_plans_plan_id_archive:
    endpoint: PUT /v1/apps/{{ record.id }}/plans/{{ record.plan_id }}/archive
    required fields: id, plan_id
    risk: high: external Mantle mutation or side effect; approval required
  update_apps_id_plans_plan_id_unarchive:
    endpoint: PUT /v1/apps/{{ record.id }}/plans/{{ record.plan_id }}/unarchive
    required fields: id, plan_id
    risk: high: external Mantle mutation or side effect; approval required
  update_apps_id_usage_metrics_usage_metric_id:
    endpoint: PUT /v1/apps/{{ record.id }}/usage_metrics/{{ record.usage_metric_id }}
    required fields: id, usage_metric_id
    risk: medium: external Mantle mutation; approval required
  update_companies_id:
    endpoint: PUT /v1/companies/{{ record.id }}
    required fields: id, parentCustomerId
    risk: medium: external Mantle mutation; approval required
  update_contacts_id:
    endpoint: PUT /v1/contacts/{{ record.id }}
    required fields: id
    risk: medium: external Mantle mutation; approval required
  update_custom_data:
    endpoint: PUT /v1/custom_data
    required fields: resourceId, resourceType, key, value
    risk: medium: external Mantle mutation; approval required
  update_customers_custom_fields_id:
    endpoint: PUT /v1/customers/custom_fields/{{ record.id }}
    required fields: id
    risk: medium: external Mantle mutation; approval required
  update_customers_id:
    endpoint: PUT /v1/customers/{{ record.id }}
    required fields: id
    risk: medium: external Mantle mutation; approval required
  update_customers_id_account_owners_owner_id:
    endpoint: PUT /v1/customers/{{ record.id }}/account_owners/{{ record.owner_id }}
    required fields: id, owner_id
    risk: medium: external Mantle mutation; approval required
  update_deal_activities_id:
    endpoint: PUT /v1/deal_activities/{{ record.id }}
    required fields: id
    risk: medium: external Mantle mutation; approval required
  update_deal_flows_id:
    endpoint: PUT /v1/deal_flows/{{ record.id }}
    required fields: id, name, dealStages
    risk: medium: external Mantle side effect; approval required
  update_deals_id:
    endpoint: PUT /v1/deals/{{ record.id }}
    required fields: id
    risk: medium: external Mantle mutation; approval required
  update_docs_collections_collection_id:
    endpoint: PUT /v1/docs/collections/{{ record.collection_id }}
    required fields: collection_id
    risk: medium: external Mantle mutation; approval required
  update_docs_groups_group_id:
    endpoint: PUT /v1/docs/groups/{{ record.group_id }}
    required fields: group_id
    risk: medium: external Mantle mutation; approval required
  update_docs_pages_page_id:
    endpoint: PUT /v1/docs/pages/{{ record.page_id }}
    required fields: page_id
    risk: medium: external Mantle mutation; approval required
  update_docs_repositories_id:
    endpoint: PUT /v1/docs/repositories/{{ record.id }}
    required fields: id
    risk: medium: external Mantle mutation; approval required
  update_docs_sites_id:
    endpoint: PUT /v1/docs/sites/{{ record.id }}
    required fields: id
    risk: medium: external Mantle mutation; approval required
  update_docs_sites_id_redirects_redirect_id:
    endpoint: PUT /v1/docs/sites/{{ record.id }}/redirects/{{ record.redirect_id }}
    required fields: id, redirect_id
    risk: medium: external Mantle mutation; approval required
  update_docs_sites_id_repositories:
    endpoint: PUT /v1/docs/sites/{{ record.id }}/repositories
    required fields: id, attachments
    risk: medium: external Mantle mutation; approval required
  update_email_campaigns_id:
    endpoint: PUT /v1/email/campaigns/{{ record.id }}
    required fields: id
    risk: medium: external Mantle side effect; approval required
  update_flow_actions_runs_id:
    endpoint: PUT /v1/flow/actions/runs/{{ record.id }}
    required fields: id
    risk: medium: external Mantle side effect; approval required
  update_flow_extensions_actions_id:
    endpoint: PUT /v1/flow/extensions/actions/{{ record.id }}
    required fields: id
    risk: medium: external Mantle mutation; approval required
  update_flows_id:
    endpoint: PATCH /v1/flows/{{ record.id }}
    required fields: id
    risk: medium: external Mantle side effect; approval required
  update_journal_entries_id:
    endpoint: PUT /v1/journal_entries/{{ record.id }}
    required fields: id
    risk: medium: external Mantle mutation; approval required
  update_lists_id:
    endpoint: PUT /v1/lists/{{ record.id }}
    required fields: id
    risk: medium: external Mantle mutation; approval required
  update_meetings_id:
    endpoint: PUT /v1/meetings/{{ record.id }}
    required fields: id
    risk: medium: external Mantle mutation; approval required
  update_meetings_id_attendees_attendee_id:
    endpoint: PUT /v1/meetings/{{ record.id }}/attendees/{{ record.attendee_id }}
    required fields: id, attendee_id
    risk: medium: external Mantle mutation; approval required
  update_meetings_id_visibility:
    endpoint: PUT /v1/meetings/{{ record.id }}/visibility
    required fields: id, visibility
    risk: medium: external Mantle mutation; approval required
  update_notification_preferences:
    endpoint: PUT /v1/notification_preferences
    required fields: helpDeskNotificationPreferences
    risk: medium: external Mantle mutation; approval required
  update_synced_emails_id:
    endpoint: PUT /v1/synced_emails/{{ record.id }}
    required fields: id
    risk: medium: external Mantle side effect; approval required
  update_tasks_id:
    endpoint: PUT /v1/tasks/{{ record.id }}
    required fields: id
    risk: medium: external Mantle mutation; approval required
  update_tasks_id_comments_comment_id:
    endpoint: PUT /v1/tasks/{{ record.id }}/comments/{{ record.comment_id }}
    required fields: id, comment_id
    risk: medium: external Mantle mutation; approval required
  update_tasks_id_todo_items_item_id:
    endpoint: PUT /v1/tasks/{{ record.id }}/todo-items/{{ record.item_id }}
    required fields: id, item_id
    risk: medium: external Mantle mutation; approval required
  update_tickets_id:
    endpoint: PUT /v1/tickets/{{ record.id }}
    required fields: id
    risk: medium: external Mantle mutation; approval required
  update_tickets_id_events:
    endpoint: PUT /v1/tickets/{{ record.id }}/events
    required fields: id
    risk: medium: external Mantle mutation; approval required
  update_tickets_id_messages:
    endpoint: PUT /v1/tickets/{{ record.id }}/messages
    required fields: id
    risk: medium: external Mantle mutation; approval required
  update_tickets_saved_filters_filter_id:
    endpoint: PUT /v1/tickets/saved_filters/{{ record.filter_id }}
    required fields: filter_id
    risk: medium: external Mantle mutation; approval required
  update_tickets_saved_replies_reply_id:
    endpoint: PUT /v1/tickets/saved_replies/{{ record.reply_id }}
    required fields: reply_id
    risk: medium: external Mantle mutation; approval required
  update_timeline_comments_id:
    endpoint: PUT /v1/timeline_comments/{{ record.id }}
    required fields: id, commentHtml
    risk: medium: external Mantle mutation; approval required
  update_webhooks_id:
    endpoint: PUT /v1/webhooks/{{ record.id }}
    required fields: id
    risk: medium: external Mantle side effect; approval required
  upsert_contacts:
    endpoint: POST /v1/contacts
    risk: medium: external Mantle mutation; approval required

SECURITY
  read risk: external Mantle API read of customer, billing, CRM, docs, email, helpdesk, and workflow data
  write risk: external Mantle API mutation of customer, billing, CRM, docs, email, helpdesk, webhook, and workflow resources; approval required
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect mantle

  # Inspect as structured JSON
  pm connectors inspect mantle --json

AGENT WORKFLOW
  - Run pm connectors inspect mantle before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
