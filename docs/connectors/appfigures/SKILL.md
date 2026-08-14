---
name: pm-appfigures
description: Appfigures connector knowledge and safe action guide.
---

# pm-appfigures

## Purpose

Reads Appfigures app-store reviews, products, analytics reports (sales/ratings/revenue/subscriptions/ads/estimates), reference data (categories/countries/languages/currencies/stores/SDKs), release events, connected external accounts, account users, and account info through the Appfigures v2 REST API, and manages release events and review responses.

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
- end_date
- group_by
- search_store
- start_date
- api_key (secret) (required)

## ETL Streams

- reviews:
  - primary key: id
  - fields: author(string), date(string), has_response(boolean), id(string), iso(string), original_title(string), product(integer), review(string), stars(number), title(string), version(string), weight(integer)
- products:
  - primary key: id
  - fields: added(string), developer(string), id(integer), name(string), ref_no(string), sku(string), store(string), store_id(integer), updated(string), vendor_identifier(string)
- sales:
  - fields: date(string), downloads(integer), net_downloads(integer), promos(integer), returns(integer), revenue(string), updates(integer)
- ratings:
  - fields: average(number), breakdown(string), date(string), stars(number)
- categories:
  - primary key: id
  - fields: device(string), id(integer), name(string), store(string), subtype(string)
- revenue:
  - primary key: report
  - fields: ads(string), business(string), edu(string), gross_business(string), gross_edu(string), gross_iaps(string), gross_returns(string), gross_sales(string), gross_subscriptions(string), gross_total(string), iaps(string), report(string), returns(string), sales(string), subscriptions(string), total(string)
- subscriptions:
  - primary key: report
  - fields: active_free_trials(integer), active_subscriptions(integer), actual_revenue(string), cancellations(integer), cancelled_subscriptions(integer), churn(string), gross_mrr(string), gross_revenue(string), mrr(string), new_subscriptions(integer), new_trials(integer), reactivations(integer), renewals(integer), report(string), trial_conversion_rate(string), trial_conversions(integer)
- ads:
  - primary key: report
  - fields: clicks(integer), ctr(number), ecpm(string), fillrate(number), impressions(integer), report(string), requests(integer), requests_filled(integer), revenue(string)
- estimates:
  - primary key: report
  - fields: downloads(integer), report(string), revenue(string)
- events:
  - primary key: id
  - fields: caption(string), date(string), details(string), id(integer), origin(string), products(array)
- external_accounts:
  - primary key: id
  - fields: account_id(integer), auto_import(boolean), id(integer), metadata(object), nickname(string), store(string), store_id(integer), username(string)
- users:
  - primary key: id
  - fields: account(object), active(boolean), avatar_url(string), currency(string), date_format(string), email(string), entitlements(array), id(integer), is_owner(boolean), last_login(string), name(string), products(array), region(string), role(string), timezone(string)
- account_info:
  - primary key: user_id
  - fields: daily_limit(integer), daily_used(integer), sequence(integer), user_email(string), user_id(integer), user_name(string), version(string)
- data_countries:
  - primary key: iso
  - fields: apple_store_no(string), iso(string), name(string)
- data_languages:
  - primary key: code
  - fields: code(string), iso(string), name(string)
- data_stores:
  - primary key: store_key
  - fields: code(string), countries(array), id(integer), name(string), short_name(string), store_key(string), storefronts(array), supported_features(array), type(string)
- data_currencies:
  - primary key: Currency
  - fields: Currency(string), Symbol(string)
- data_sdks:
  - primary key: id
  - fields: active(boolean), code(string), description(string), developer(object), external_links(array), id(string), name(string), notes(string), release_date(string), started_tracking(string), tags(array), tracked_platforms(array)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- reply_to_review:
  - endpoint: POST /reviews/{{ record.id }}/response
  - required fields: id, content
  - risk: publishes a developer response to a customer review, visible on the public app store listing
- create_event:
  - endpoint: POST /events/
  - required fields: caption, date
  - risk: creates a release/marketing event marker overlaid on every Appfigures analytics chart
- update_event:
  - endpoint: PUT /events/{{ record.id }}
  - required fields: id
  - risk: mutates an existing release/marketing event marker overlaid on every Appfigures analytics chart
- delete_event:
  - endpoint: DELETE /events/{{ record.id }}
  - required fields: id
  - risk: permanently deletes an event marker from every Appfigures analytics chart

## Security

- read risk: external Appfigures API read of app-store review, analytics, and account data
- write risk: external Appfigures API mutation — publishes a public review response, and creates/edits/deletes release-event markers overlaid on analytics charts
- approval: reverse ETL plan approval required before writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect appfigures
```

### Inspect as structured JSON

```bash
pm connectors inspect appfigures --json
```

## Agent Rules

- Run pm connectors inspect appfigures before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
