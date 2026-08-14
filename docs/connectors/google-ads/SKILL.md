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

- move_manager_link_customer_manager_links:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/customerManagerLinks:moveManagerLink
  - risk: Executes Google Ads API v22 method customers.customerManagerLinks.moveManagerLink against the configured customer. Review provider-side effects before approval.
- remove_data_links:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/dataLinks:remove
  - required fields: resourceName
  - risk: Executes Google Ads API v22 method customers.dataLinks.remove against the configured customer. Review provider-side effects before approval.
- update_data_links:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/dataLinks:update
  - required fields: resourceName
  - risk: Executes Google Ads API v22 method customers.dataLinks.update against the configured customer. Review provider-side effects before approval.
- remove_product_link_invitations:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/productLinkInvitations:remove
  - required fields: resourceName
  - risk: Executes Google Ads API v22 method customers.productLinkInvitations.remove against the configured customer. Review provider-side effects before approval.
- update_product_link_invitations:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/productLinkInvitations:update
  - required fields: resourceName
  - risk: Executes Google Ads API v22 method customers.productLinkInvitations.update against the configured customer. Review provider-side effects before approval.
- remove_product_links:
  - endpoint: POST /v22/customers/{{ config.customer_id }}/productLinks:remove
  - required fields: resourceName
  - risk: Executes Google Ads API v22 method customers.productLinks.remove against the configured customer. Review provider-side effects before approval.
- start_identity_verification_customers:
  - endpoint: POST /v22/customers/{{ config.customer_id }}:startIdentityVerification
  - risk: Executes Google Ads API v22 method customers.startIdentityVerification against the configured customer. Review provider-side effects before approval.

## Security

- read risk: external Google Ads API reads of customer, campaign, ad-group, and bounded direct-read metadata
- write risk: limited guarded Google Ads API reverse/write actions with closed record schemas; destructive/admin actions require explicit approval
- approval: reads require no approval; writes remain gated by plan -> preview -> explicit approval -> execute
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Google Ads v22 fixed direct reads and limited guarded reverse/write actions.
- Usage: pm google-ads <resource> <operation> [flags]
- Source CLI: Google Ads API v22 REST discovery (https://developers.google.com/google-ads/api/reference/rpc/v22/overview)
- Other Commands
  - audience-insights list-insights-eligible-dates - Read Google Ads audienceInsights.listInsightsEligibleDates. [intent=direct_read availability=implemented operation=google_ads.audience.insights.list.insights.eligible.dates]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --page, --page-cursor
  - customers asset-generations generate-images - Read Google Ads customers.assetGenerations.generateImages. [intent=direct_read availability=implemented operation=google_ads.customers.asset.generations.generate.images]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --advertising-channel-type, --asset-field-types, --page, --page-cursor
  - customers asset-generations generate-text - Read Google Ads customers.assetGenerations.generateText. [intent=direct_read availability=implemented operation=google_ads.customers.asset.generations.generate.text]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --advertising-channel-type, --asset-field-types (required), --final-url, --freeform-prompt, --keywords, --page, --page-cursor
  - customers generate-ad-group-themes - Read Google Ads customers.generateAdGroupThemes. [intent=direct_read availability=implemented operation=google_ads.customers.generate.ad.group.themes]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --ad-groups (required), --keywords (required), --page, --page-cursor
  - customers generate-keyword-historical-metrics - Read Google Ads customers.generateKeywordHistoricalMetrics. [intent=direct_read availability=implemented operation=google_ads.customers.generate.keyword.historical.metrics]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --geo-target-constants, --include-adult-keywords, --keyword-plan-network, --keywords, --language, --page, --page-cursor
  - customers generate-keyword-ideas - Read Google Ads customers.generateKeywordIdeas. [intent=direct_read availability=implemented operation=google_ads.customers.generate.keyword.ideas]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --geo-target-constants, --include-adult-keywords, --keyword-annotation, --keyword-plan-network, --language, --page-size, --page, --page-cursor
  - customers generate-suggested-targeting-insights - Read Google Ads customers.generateSuggestedTargetingInsights. [intent=direct_read availability=implemented operation=google_ads.customers.generate.suggested.targeting.insights]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --customer-insights-group, --page, --page-cursor
  - customers get-identity-verification - Read Google Ads customers.getIdentityVerification. [intent=direct_read availability=implemented operation=google_ads.customers.get.identity.verification]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --page, --page-cursor
  - customers invoices list - Read Google Ads customers.invoices.list. [intent=direct_read availability=implemented operation=google_ads.customers.invoices.list]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --billing-setup (required), --issue-month (required), --issue-year (required), --page, --page-cursor
  - customers payments-accounts list - Read Google Ads customers.paymentsAccounts.list. [intent=direct_read availability=implemented operation=google_ads.customers.payments.accounts.list]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --page, --page-cursor
  - customers recommendations generate - Read Google Ads customers.recommendations.generate. [intent=direct_read availability=implemented operation=google_ads.customers.recommendations.generate]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --advertising-channel-type (required), --campaign-call-asset-count, --campaign-image-asset-count, --campaign-sitelink-count, --conversion-tracking-status, --country-codes, --language-codes, --merchant-center-account-id, --negative-locations-ids, --positive-locations-ids, --recommendation-types (required), --target-content-network, --target-partner-search-network, --page, --page-cursor
  - customers search-audience-insights-attributes - Read Google Ads customers.searchAudienceInsightsAttributes. [intent=direct_read availability=implemented operation=google_ads.customers.search.audience.insights.attributes]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --customer-insights-group, --dimensions (required), --query-text (required), --page, --page-cursor
  - customers suggest-brands - Read Google Ads customers.suggestBrands. [intent=direct_read availability=implemented operation=google_ads.customers.suggest.brands]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --brand-prefix (required), --selected-brands, --page, --page-cursor
  - customers suggest-travel-assets - Read Google Ads customers.suggestTravelAssets. [intent=direct_read availability=implemented operation=google_ads.customers.suggest.travel.assets]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --language-option (required), --place-ids, --page, --page-cursor
  - geo-target-constants suggest - Read Google Ads geoTargetConstants.suggest. [intent=direct_read availability=implemented operation=google_ads.geo.target.constants.suggest]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --country-code, --locale, --page, --page-cursor
  - keyword-theme-constants suggest - Read Google Ads keywordThemeConstants.suggest. [intent=direct_read availability=implemented operation=google_ads.keyword.theme.constants.suggest]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --country-code, --language-code, --query-text, --page, --page-cursor
  - v22 generate-conversion-rates - Read Google Ads v22.generateConversionRates. [intent=direct_read availability=implemented operation=google_ads.v22.generate.conversion.rates]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --customer-id (required), --customer-reach-group, --page, --page-cursor
  - v22 list-plannable-locations - Read Google Ads v22.listPlannableLocations. [intent=direct_read availability=implemented operation=google_ads.v22.list.plannable.locations]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --page, --page-cursor
  - v22 list-plannable-products - Read Google Ads v22.listPlannableProducts. [intent=direct_read availability=implemented operation=google_ads.v22.list.plannable.products]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --plannable-location-id (required), --page, --page-cursor
  - v22 list-plannable-user-interests - Read Google Ads v22.listPlannableUserInterests. [intent=direct_read availability=implemented operation=google_ads.v22.list.plannable.user.interests]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --customer-id (required), --name-query, --path-query, --user-interest-taxonomy-types, --page, --page-cursor
  - v22 list-plannable-user-lists - Read Google Ads v22.listPlannableUserLists. [intent=direct_read availability=implemented operation=google_ads.v22.list.plannable.user.lists]; approval: none; risk: Bounded JSON direct read; response fields with secret-like names are redacted.; flags: --customer-id (required), --customer-reach-group, --page, --page-cursor

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
