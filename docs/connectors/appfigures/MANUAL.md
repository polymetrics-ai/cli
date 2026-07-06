# pm connectors inspect appfigures

```text
NAME
  pm connectors inspect appfigures - Appfigures connector manual

SYNOPSIS
  pm connectors inspect appfigures
  pm connectors inspect appfigures --json
  pm credentials add <name> --connector appfigures [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Appfigures app-store reviews, products, analytics reports (sales/ratings/revenue/subscriptions/ads/estimates), reference data (categories/countries/languages/currencies/stores/SDKs), release events, connected external accounts, account users, and account info through the Appfigures v2 REST API, and manages release events and review responses.

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
  end_date
  group_by
  search_store
  start_date
  api_key (secret)

ETL STREAMS
  reviews:
    primary key: id
    fields: author(), date(), has_response(), id(), iso(), original_title(), product(), review(), stars(), title(), version(), weight()
  products:
    primary key: id
    fields: added(), developer(), id(), name(), ref_no(), sku(), store(), store_id(), updated(), vendor_identifier()
  sales:
    fields: date(), downloads(), net_downloads(), promos(), returns(), revenue(), updates()
  ratings:
    fields: average(), breakdown(), date(), stars()
  categories:
    primary key: id
    fields: device(), id(), name(), store(), subtype()
  revenue:
    primary key: report
    fields: ads(), business(), edu(), gross_business(), gross_edu(), gross_iaps(), gross_returns(), gross_sales(), gross_subscriptions(), gross_total(), iaps(), report(), returns(), sales(), subscriptions(), total()
  subscriptions:
    primary key: report
    fields: active_free_trials(), active_subscriptions(), actual_revenue(), cancellations(), cancelled_subscriptions(), churn(), gross_mrr(), gross_revenue(), mrr(), new_subscriptions(), new_trials(), reactivations(), renewals(), report(), trial_conversion_rate(), trial_conversions()
  ads:
    primary key: report
    fields: clicks(), ctr(), ecpm(), fillrate(), impressions(), report(), requests(), requests_filled(), revenue()
  estimates:
    primary key: report
    fields: downloads(), report(), revenue()
  events:
    primary key: id
    fields: caption(), date(), details(), id(), origin(), products()
  external_accounts:
    primary key: id
    fields: account_id(), auto_import(), id(), metadata(), nickname(), store(), store_id(), username()
  users:
    primary key: id
    fields: account(), active(), avatar_url(), currency(), date_format(), email(), entitlements(), id(), is_owner(), last_login(), name(), products(), region(), role(), timezone()
  account_info:
    primary key: user_id
    fields: daily_limit(), daily_used(), sequence(), user_email(), user_id(), user_name(), version()
  data_countries:
    primary key: iso
    fields: apple_store_no(), iso(), name()
  data_languages:
    primary key: code
    fields: code(), iso(), name()
  data_stores:
    primary key: store_key
    fields: code(), countries(), id(), name(), short_name(), store_key(), storefronts(), supported_features(), type()
  data_currencies:
    primary key: Currency
    fields: Currency(), Symbol()
  data_sdks:
    primary key: id
    fields: active(), code(), description(), developer(), external_links(), id(), name(), notes(), release_date(), started_tracking(), tags(), tracked_platforms()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  reply_to_review:
    endpoint: POST /reviews/{{ record.id }}/response
    required fields: id
    optional fields: content
    risk: publishes a developer response to a customer review, visible on the public app store listing
  create_event:
    endpoint: POST /events/
    risk: creates a release/marketing event marker overlaid on every Appfigures analytics chart
  update_event:
    endpoint: PUT /events/{{ record.id }}
    required fields: id
    risk: mutates an existing release/marketing event marker overlaid on every Appfigures analytics chart
  delete_event:
    endpoint: DELETE /events/{{ record.id }}
    required fields: id
    risk: permanently deletes an event marker from every Appfigures analytics chart

SECURITY
  read risk: external Appfigures API read of app-store review, analytics, and account data
  write risk: external Appfigures API mutation — publishes a public review response, and creates/edits/deletes release-event markers overlaid on analytics charts
  approval: reverse ETL plan approval required before writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect appfigures

  # Inspect as structured JSON
  pm connectors inspect appfigures --json

AGENT WORKFLOW
  - Run pm connectors inspect appfigures before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
