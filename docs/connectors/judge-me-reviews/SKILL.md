---
name: pm-judge-me-reviews
description: Judge.me Reviews connector knowledge and safe action guide.
---

# pm-judge-me-reviews

## Purpose

Reads and writes Judge.me reviews, widgets, reviewers, webhooks, shop metadata, settings, replies, and legacy product/widget resources through the Judge.me REST API.

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
- shop_domain
- webhook_id
- widget_page
- widget_per_page
- widget_review_type
- api_key (secret)

## ETL Streams

- reviews:
  - primary key: id
  - cursor: created_at
  - fields: body(), created_at(), curated(), hidden(), id(), product_external_id(), published(), rating(), reviewer_email(), reviewer_id(), reviewer_name(), source(), title(), updated_at(), verified()
- products:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), external_id(), handle(), id(), title(), updated_at(), url()
- widgets:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), id(), name(), status(), updated_at(), widget_type()
- product_review_widget:
  - primary key: id
  - fields: badge(), count(), html(), id(), message(), rating(), settings(), value(), widget()
- preview_badge_widget:
  - primary key: id
  - fields: badge(), count(), html(), id(), message(), rating(), settings(), value(), widget()
- featured_carousel_widget:
  - primary key: id
  - fields: badge(), count(), html(), id(), message(), rating(), settings(), value(), widget()
- reviews_tab_widget:
  - primary key: id
  - fields: badge(), count(), html(), id(), message(), rating(), settings(), value(), widget()
- all_reviews_page_widget:
  - primary key: id
  - fields: badge(), count(), html(), id(), message(), rating(), settings(), value(), widget()
- verified_badge_widget:
  - primary key: id
  - fields: badge(), count(), html(), id(), message(), rating(), settings(), value(), widget()
- all_reviews_count_widget:
  - primary key: id
  - fields: badge(), count(), html(), id(), message(), rating(), settings(), value(), widget()
- all_reviews_rating_widget:
  - primary key: id
  - fields: badge(), count(), html(), id(), message(), rating(), settings(), value(), widget()
- shop_reviews_count_widget:
  - primary key: id
  - fields: badge(), count(), html(), id(), message(), rating(), settings(), value(), widget()
- shop_reviews_rating_widget:
  - primary key: id
  - fields: badge(), count(), html(), id(), message(), rating(), settings(), value(), widget()
- widget_settings:
  - primary key: id
  - fields: badge(), count(), html(), id(), message(), rating(), settings(), value(), widget()
- html_miracle_widget:
  - primary key: id
  - fields: badge(), count(), html(), id(), message(), rating(), settings(), value(), widget()
- checkout_comments_widget:
  - primary key: id
  - fields: badge(), count(), html(), id(), message(), rating(), settings(), value(), widget()
- reviews_count:
  - primary key: id
  - fields: badge(), count(), html(), id(), message(), rating(), settings(), value(), widget()
- review:
  - primary key: id
  - fields: body(), created_at(), curated(), has_published_pictures(), has_published_videos(), hidden(), id(), pictures(), product_external_id(), product_handle(), product_title(), rating(), reviewer(), reviewer_email(), reviewer_id(), reviewer_name(), source(), title(), updated_at(), verified()
- reviewer:
  - primary key: id
  - fields: accepts_marketing(), email(), external_id(), id(), name(), phone(), tags(), unsubscribed_at()
- webhooks:
  - primary key: id
  - fields: failure_count(), id(), key(), url()
- webhook:
  - primary key: id
  - fields: failure_count(), id(), key(), url()
- shop_info:
  - primary key: id
  - fields: awesome(), country(), created_at(), currency(), custom_domain(), domain(), email(), id(), name(), owner(), phone(), plan(), platform(), timezone(), updated_at(), widget_version()
- settings:
  - primary key: id
  - fields: badge(), count(), html(), id(), message(), rating(), settings(), value(), widget()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_review:
  - endpoint: POST /reviews
  - risk: creates a public web review in Judge.me; approval required
- update_review:
  - endpoint: PUT /reviews/{{ record.id }}
  - required fields: id
  - risk: publishes or hides a Judge.me review by changing curated status; approval required
- update_reviewer:
  - endpoint: PUT /reviewers/{{ record.id }}
  - required fields: id
  - risk: creates or updates reviewer identity fields in Judge.me; approval required
- request_reviewer_data:
  - endpoint: POST /reviewers/data_request
  - risk: submits a Judge.me reviewer data request; approval required
- delete_webhook:
  - endpoint: DELETE /webhooks
  - optional fields: key, url
  - risk: deletes a Judge.me webhook subscription; approval required
- create_webhook:
  - endpoint: POST /webhooks
  - risk: creates a Judge.me webhook subscription; approval required
- update_webhook:
  - endpoint: PUT /webhooks/{{ record.id }}
  - required fields: id
  - risk: updates a Judge.me webhook subscription; approval required
- bulk_create_webhooks:
  - endpoint: POST /webhooks/bulk_create
  - risk: creates multiple Judge.me webhook subscriptions; approval required
- update_shop:
  - endpoint: PUT /shops
  - risk: updates Judge.me shop profile fields; approval required
- uninstall_shop:
  - endpoint: DELETE /shops
  - risk: uninstalls the shop from Judge.me; destructive approval required
- create_checkout_comment:
  - endpoint: POST /shops
  - risk: creates a checkout comment in Judge.me Checkout Comments; approval required
- create_reply:
  - endpoint: POST /replies
  - risk: creates a public reply on a Judge.me review; approval required
- create_private_reply:
  - endpoint: POST /private_replies
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
