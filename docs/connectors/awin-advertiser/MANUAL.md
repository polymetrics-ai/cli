# pm connectors inspect awin-advertiser

```text
NAME
  pm connectors inspect awin-advertiser - Awin Advertiser connector manual

SYNOPSIS
  pm connectors inspect awin-advertiser
  pm connectors inspect awin-advertiser --json
  pm credentials add <name> --connector awin-advertiser [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Awin advertiser transactions, publisher-aggregated performance reports, publisher relationships, and publisher performance reports, and creates advertiser promotion/voucher offers, through the Awin Advertiser REST API.

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
  advertiserId (required)
  base_url
  mode
  publisher_id
  report_region
  report_start_date
  start_date
  transaction_status
  api_key (secret) (required)

ETL STREAMS
  transactions:
    primary key: id
    cursor: transactionDate
    fields: advertiserId(integer), clickDate(string), clickRefs(object), commissionAmount(object), commissionSharingPublisherId(integer), customParameters(object), id(integer), publisherId(integer), saleAmount(object), siteName(string), transactionDate(string), transactionStatus(string), type(string), url(string), validationDate(string)
  campaign_performance:
    primary key: publisherId
    fields: advertiserId(integer), clicks(integer), confirmedNo(integer), currency(string), declinedNo(integer), impressions(integer), pendingNo(integer), publisherId(integer), publisherName(string), region(string), totalComm(number), totalNo(integer), totalSaleAmount(object)
  publishers:
    primary key: id
    fields: displayUrl(string), id(integer), kind(string), name(string), status(string)
  publisher_performance:
    primary key: publisherId
    fields: advertiserId(integer), advertiserName(string), bonusComm(number), bonusNo(integer), bonusValue(number), clicks(integer), confirmedComm(number), confirmedNo(integer), confirmedValue(number), currency(string), declinedComm(number), declinedNo(integer), declinedValue(number), impressions(integer), pendingComm(number), pendingNo(integer), pendingValue(number), publisherId(integer), publisherName(string), region(string), tags(array), totalComm(number), totalNo(integer), totalValue(number)
  creative_performance:
    primary key: creativeId, publisherId
    fields: advertiserId(integer), advertiserName(string), bonusComm(number), bonusNo(integer), bonusValue(number), clicks(integer), confirmedComm(number), confirmedNo(integer), confirmedValue(number), creativeId(integer), creativeName(string), currency(string), declinedComm(number), declinedNo(integer), declinedValue(number), impressions(integer), pendingComm(number), pendingNo(integer), pendingValue(number), publisherId(integer), publisherName(string), region(string), tagName(string), totalComm(number), totalNo(integer), totalValue(number)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_offer:
    endpoint: POST /promotion/advertiser/{{ config.advertiserId }}
    required fields: title, description, terms, type, url, startDate, endDate, appliesToAllRegions, promotionCategories
    risk: creates a new promotion or voucher offer in the advertiser's MyOffers system, visible to publishers immediately; external mutation, approval required

SECURITY
  read risk: external Awin API read of advertiser commission transactions and publisher performance data
  write risk: creates a new promotion or voucher offer in the advertiser's MyOffers system, immediately visible to publishers; external mutation, approval required
  approval: required for create_offer
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect awin-advertiser

  # Inspect as structured JSON
  pm connectors inspect awin-advertiser --json

AGENT WORKFLOW
  - Run pm connectors inspect awin-advertiser before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
