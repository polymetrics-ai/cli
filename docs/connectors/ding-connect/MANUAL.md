# pm connectors inspect ding-connect

```text
NAME
  pm connectors inspect ding-connect - Ding Connect connector manual

SYNOPSIS
  pm connectors inspect ding-connect
  pm connectors inspect ding-connect --json
  pm credentials add <name> --connector ding-connect [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads DingConnect reference catalogs (countries, currencies, regions, providers, products, product descriptions, promotions, provider status, error code descriptions, account balance) through the DingConnect REST API, and sends real-money mobile top-up transfers.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  x_correlation_id
  api_key (secret)

ETL STREAMS
  countries:
    primary key: uuid
    fields: CountryIso(), CountryName(), InternationalDialingInformation(), RegionCodes(), uuid()
  currencies:
    primary key: uuid
    fields: CurrencyIso(), CurrencyName(), uuid()
  regions:
    primary key: uuid
    fields: CountryIso(), RegionCode(), RegionName(), uuid()
  providers:
    primary key: uuid
    fields: CountryIso(), CustomerCareNumber(), LogoUrl(), Name(), PaymentTypes(), ProviderCode(), RegionCodes(), ValidationRegex(), uuid()
  products:
    primary key: uuid
    fields: Benefits(), CommissionRate(), DefaultDisplayText(), LocalizationKey(), Maximum(), Minimum(), PaymentTypes(), ProcessingMode(), ProviderCode(), RedemptionMechanism(), RegionCode(), SkuCode(), ValidityPeriodIso(), uuid()
  product_descriptions:
    primary key: uuid
    fields: DescriptionMarkdown(), DisplayText(), LanguageCode(), LocalizationKey(), ReadMoreMarkdown(), uuid()
  promotions:
    primary key: uuid
    fields: CurrencyIso(), EndUtc(), LocalizationKey(), MinimumSendAmount(), ProviderCode(), StartUtc(), ValidityPeriodIso(), uuid()
  provider_status:
    primary key: uuid
    fields: IsProcessingTransfers(), Message(), ProviderCode(), uuid()
  error_code_descriptions:
    primary key: uuid
    fields: Code(), Message(), uuid()
  balance:
    primary key: uuid
    fields: Balance(), CurrencyIso(), uuid()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  send_transfer:
    endpoint: POST /api/V1/SendTransfer
    risk: external mutation; sends a real-money mobile top-up/airtime transfer to a live account and deducts the distributor's DingConnect balance unless ValidateOnly is set; approval required

SECURITY
  read risk: external DingConnect API read of reference/catalog data and distributor account balance
  write risk: external mutation; sends a real-money mobile top-up/airtime transfer and deducts the distributor's live DingConnect balance
  approval: required for the send_transfer write action; read streams remain unapproved
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect ding-connect

  # Inspect as structured JSON
  pm connectors inspect ding-connect --json

AGENT WORKFLOW
  - Run pm connectors inspect ding-connect before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
