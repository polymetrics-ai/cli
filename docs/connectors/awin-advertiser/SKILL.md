---
name: pm-awin-advertiser
description: Awin Advertiser connector knowledge and safe action guide.
---

# pm-awin-advertiser

## Purpose

Reads Awin advertiser transactions, publisher-aggregated performance reports, publisher relationships, and publisher performance reports, and creates advertiser promotion/voucher offers, through the Awin Advertiser REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- advertiserId
- base_url
- mode
- publisher_id
- report_region
- report_start_date
- start_date
- transaction_status
- api_key (secret)

## ETL Streams

- transactions:
  - primary key: id
  - cursor: transactionDate
  - fields: advertiserId(), clickDate(), clickRefs(), commissionAmount(), commissionSharingPublisherId(), customParameters(), id(), publisherId(), saleAmount(), siteName(), transactionDate(), transactionStatus(), type(), url(), validationDate()
- campaign_performance:
  - primary key: publisherId
  - fields: advertiserId(), clicks(), confirmedNo(), currency(), declinedNo(), impressions(), pendingNo(), publisherId(), publisherName(), region(), totalComm(), totalNo(), totalSaleAmount()
- publishers:
  - primary key: id
  - fields: displayUrl(), id(), kind(), name(), status()
- publisher_performance:
  - primary key: publisherId
  - fields: advertiserId(), advertiserName(), bonusComm(), bonusNo(), bonusValue(), clicks(), confirmedComm(), confirmedNo(), confirmedValue(), currency(), declinedComm(), declinedNo(), declinedValue(), impressions(), pendingComm(), pendingNo(), pendingValue(), publisherId(), publisherName(), region(), tags(), totalComm(), totalNo(), totalValue()
- creative_performance:
  - primary key: creativeId, publisherId
  - fields: advertiserId(), advertiserName(), bonusComm(), bonusNo(), bonusValue(), clicks(), confirmedComm(), confirmedNo(), confirmedValue(), creativeId(), creativeName(), currency(), declinedComm(), declinedNo(), declinedValue(), impressions(), pendingComm(), pendingNo(), pendingValue(), publisherId(), publisherName(), region(), tagName(), totalComm(), totalNo(), totalValue()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_offer:
  - endpoint: POST /promotion/advertiser/{{ config.advertiserId }}
  - risk: creates a new promotion or voucher offer in the advertiser's MyOffers system, visible to publishers immediately; external mutation, approval required

## Security

- read risk: external Awin API read of advertiser commission transactions and publisher performance data
- write risk: creates a new promotion or voucher offer in the advertiser's MyOffers system, immediately visible to publishers; external mutation, approval required
- approval: required for create_offer
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect awin-advertiser
```

### Inspect as structured JSON

```bash
pm connectors inspect awin-advertiser --json
```

## Agent Rules

- Run pm connectors inspect awin-advertiser before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
