---
name: pm-judge-me-reviews
description: Judge.me Reviews connector knowledge and safe action guide.
---

# pm-judge-me-reviews

## Purpose

Reads and writes Judge.me reviews, widgets, reviewers, webhooks, shop metadata, settings, replies, and legacy product/widget resources through the Judge.me REST API.

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
- page_size
- product_external_id
- product_handle
- product_id
- rating
- review_id
- reviewer_email
- reviewer_external_id
- reviewer_id
- setting_keys
- shop_domain (required)
- webhook_id
- widget_page
- widget_per_page
- widget_review_type
- api_key (secret) (required)

## ETL Streams

- reviews:
  - primary key: id
  - cursor: created_at
  - fields: body(string), created_at(string), curated(string), hidden(boolean), id(integer), product_external_id(string), published(boolean), rating(integer), reviewer_email(string), reviewer_id(integer), reviewer_name(string), source(string), title(string), updated_at(string), verified(string)
- products:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), external_id(string), handle(string), id(integer), title(string), updated_at(string), url(string)
- widgets:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), id(integer), name(string), status(string), updated_at(string), widget_type(string)
- product_review_widget:
  - primary key: id
  - fields: badge(string), count(integer), html(string), id(string), message(string), rating(integer), settings(string), value(integer), widget(string)
- preview_badge_widget:
  - primary key: id
  - fields: badge(string), count(integer), html(string), id(string), message(string), rating(integer), settings(string), value(integer), widget(string)
- featured_carousel_widget:
  - primary key: id
  - fields: badge(string), count(integer), html(string), id(string), message(string), rating(integer), settings(string), value(integer), widget(string)
- reviews_tab_widget:
  - primary key: id
  - fields: badge(string), count(integer), html(string), id(string), message(string), rating(integer), settings(string), value(integer), widget(string)
- all_reviews_page_widget:
  - primary key: id
  - fields: badge(string), count(integer), html(string), id(string), message(string), rating(integer), settings(string), value(integer), widget(string)
- verified_badge_widget:
  - primary key: id
  - fields: badge(string), count(integer), html(string), id(string), message(string), rating(integer), settings(string), value(integer), widget(string)
- all_reviews_count_widget:
  - primary key: id
  - fields: badge(string), count(integer), html(string), id(string), message(string), rating(integer), settings(string), value(integer), widget(string)
- all_reviews_rating_widget:
  - primary key: id
  - fields: badge(string), count(integer), html(string), id(string), message(string), rating(integer), settings(string), value(integer), widget(string)
- shop_reviews_count_widget:
  - primary key: id
  - fields: badge(string), count(integer), html(string), id(string), message(string), rating(integer), settings(string), value(integer), widget(string)
- shop_reviews_rating_widget:
  - primary key: id
  - fields: badge(string), count(integer), html(string), id(string), message(string), rating(integer), settings(string), value(integer), widget(string)
- widget_settings:
  - primary key: id
  - fields: badge(string), count(integer), html(string), id(string), message(string), rating(integer), settings(string), value(integer), widget(string)
- html_miracle_widget:
  - primary key: id
  - fields: badge(string), count(integer), html(string), id(string), message(string), rating(integer), settings(string), value(integer), widget(string)
- checkout_comments_widget:
  - primary key: id
  - fields: badge(string), count(integer), html(string), id(string), message(string), rating(integer), settings(string), value(integer), widget(string)
- reviews_count:
  - primary key: id
  - fields: badge(string), count(integer), html(string), id(string), message(string), rating(integer), settings(string), value(integer), widget(string)
- review:
  - primary key: id
  - fields: body(string), created_at(string), curated(string), has_published_pictures(boolean), has_published_videos(boolean), hidden(boolean), id(integer), pictures(array), product_external_id(string), product_handle(string), product_title(string), rating(integer), reviewer(object), reviewer_email(string), reviewer_id(integer), reviewer_name(string), source(string), title(string), updated_at(string), verified(string)
- reviewer:
  - primary key: id
  - fields: accepts_marketing(boolean), email(string), external_id(integer), id(integer), name(string), phone(string), tags(array), unsubscribed_at(string)
- webhooks:
  - primary key: id
  - fields: failure_count(integer), id(integer), key(string), url(string)
- webhook:
  - primary key: id
  - fields: failure_count(integer), id(integer), key(string), url(string)
- shop_info:
  - primary key: id
  - fields: awesome(boolean), country(string), created_at(string), currency(string), custom_domain(string), domain(string), email(string), id(integer), name(string), owner(string), phone(string), plan(string), platform(string), timezone(string), updated_at(string), widget_version(string)
- settings:
  - primary key: id
  - fields: badge(string), count(integer), html(string), id(string), message(string), rating(integer), settings(string), value(integer), widget(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_review:
  - endpoint: POST /reviews
  - required fields: shop_domain, platform, name, email, rating, body
  - risk: creates a public web review in Judge.me; approval required
- update_review:
  - endpoint: PUT /reviews/{{ record.id }}
  - required fields: id, curated
  - risk: publishes or hides a Judge.me review by changing curated status; approval required
- update_reviewer:
  - endpoint: PUT /reviewers/{{ record.id }}
  - required fields: id, reviewer
  - risk: creates or updates reviewer identity fields in Judge.me; approval required
- request_reviewer_data:
  - endpoint: POST /reviewers/data_request
  - required fields: customer
  - risk: submits a Judge.me reviewer data request; approval required
- delete_webhook:
  - endpoint: DELETE /webhooks
  - required fields: key, url
  - risk: deletes a Judge.me webhook subscription; approval required
- create_webhook:
  - endpoint: POST /webhooks
  - required fields: webhook
  - risk: creates a Judge.me webhook subscription; approval required
- update_webhook:
  - endpoint: PUT /webhooks/{{ record.id }}
  - required fields: id, webhook
  - risk: updates a Judge.me webhook subscription; approval required
- bulk_create_webhooks:
  - endpoint: POST /webhooks/bulk_create
  - required fields: webhooks
  - risk: creates multiple Judge.me webhook subscriptions; approval required
- update_shop:
  - endpoint: PUT /shops
  - risk: updates Judge.me shop profile fields; approval required
- uninstall_shop:
  - endpoint: DELETE /shops
  - risk: uninstalls the shop from Judge.me; destructive approval required
- create_checkout_comment:
  - endpoint: POST /shops
  - required fields: content, external_product_id, create_from, customer
  - risk: creates a checkout comment in Judge.me Checkout Comments; approval required
- create_reply:
  - endpoint: POST /replies
  - required fields: review_id, reply
  - risk: creates a public reply on a Judge.me review; approval required
- create_private_reply:
  - endpoint: POST /private_replies
  - required fields: review_id, private_reply
  - risk: creates a private email reply for a Judge.me review; approval required

## Security

- read risk: external Judge.me API read of Shopify shop reviews, widgets, reviewers, webhooks, shop metadata, settings, and legacy product/widget resources
- write risk: external Judge.me API mutations can create reviews, update moderation/reviewer/shop/webhook state, create replies/comments, and uninstall a shop
- approval: reverse ETL writes require plan, preview, approval, execute
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect judge-me-reviews
```

### Inspect as structured JSON

```bash
pm connectors inspect judge-me-reviews --json
```

## Agent Rules

- Run pm connectors inspect judge-me-reviews before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
