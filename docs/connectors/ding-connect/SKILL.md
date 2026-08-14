---
name: pm-ding-connect
description: Ding Connect connector knowledge and safe action guide.
---

# pm-ding-connect

## Purpose

Reads DingConnect reference catalogs (countries, currencies, regions, providers, products, product descriptions, promotions, provider status, error code descriptions, account balance) through the DingConnect REST API, and sends real-money mobile top-up transfers.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- x_correlation_id
- api_key (secret) (required)

## ETL Streams

- countries:
  - primary key: uuid
  - fields: CountryIso(string), CountryName(string), InternationalDialingInformation(object), RegionCodes(array), uuid(string)
- currencies:
  - primary key: uuid
  - fields: CurrencyIso(string), CurrencyName(string), uuid(string)
- regions:
  - primary key: uuid
  - fields: CountryIso(string), RegionCode(string), RegionName(string), uuid(string)
- providers:
  - primary key: uuid
  - fields: CountryIso(string), CustomerCareNumber(string), LogoUrl(string), Name(string), PaymentTypes(array), ProviderCode(string), RegionCodes(array), ValidationRegex(string), uuid(string)
- products:
  - primary key: uuid
  - fields: Benefits(array), CommissionRate(number), DefaultDisplayText(string), LocalizationKey(string), Maximum(object), Minimum(object), PaymentTypes(array), ProcessingMode(string), ProviderCode(string), RedemptionMechanism(string), RegionCode(string), SkuCode(string), ValidityPeriodIso(string), uuid(string)
- product_descriptions:
  - primary key: uuid
  - fields: DescriptionMarkdown(string), DisplayText(string), LanguageCode(string), LocalizationKey(string), ReadMoreMarkdown(string), uuid(string)
- promotions:
  - primary key: uuid
  - fields: CurrencyIso(string), EndUtc(string), LocalizationKey(string), MinimumSendAmount(number), ProviderCode(string), StartUtc(string), ValidityPeriodIso(string), uuid(string)
- provider_status:
  - primary key: uuid
  - fields: IsProcessingTransfers(boolean), Message(string), ProviderCode(string), uuid(string)
- error_code_descriptions:
  - primary key: uuid
  - fields: Code(string), Message(string), uuid(string)
- balance:
  - primary key: uuid
  - fields: Balance(number), CurrencyIso(string), uuid(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- send_transfer:
  - endpoint: POST /api/V1/SendTransfer
  - required fields: SkuCode, SendValue, AccountNumber, DistributorRef
  - risk: external mutation; sends a real-money mobile top-up/airtime transfer to a live account and deducts the distributor's DingConnect balance unless ValidateOnly is set; approval required

## Security

- read risk: external DingConnect API read of reference/catalog data and distributor account balance
- write risk: external mutation; sends a real-money mobile top-up/airtime transfer and deducts the distributor's live DingConnect balance
- approval: required for the send_transfer write action; read streams remain unapproved
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect ding-connect
```

### Inspect as structured JSON

```bash
pm connectors inspect ding-connect --json
```

## Agent Rules

- Run pm connectors inspect ding-connect before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
