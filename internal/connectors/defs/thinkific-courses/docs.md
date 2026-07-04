# Overview

Thinkific Courses reads the Thinkific Admin API v1 at `https://api.thinkific.com/api/public/v1`. The bundle keeps the four legacy streams (`courses`, `chapters`, `lessons`, `enrollments`) and adds documented Admin API v1 streams for bundles, categories, coupons, contents, course detail/chapters, custom profile fields, groups, instructors, orders, product publish requests, products, promotions, course reviews, users, and site scripts.

Pass B also adds declarative write actions for documented JSON Admin API mutations that do not require write-time query parameters.

## Auth setup

Provide `api_key` as a secret and `subdomain` as config. The API key is sent in `X-Auth-API-Key`; the subdomain is sent in `X-Auth-Subdomain`, matching legacy. `base_url` defaults to `https://api.thinkific.com/api/public/v1`.

Detail and nested streams use explicit config IDs such as `bundle_id`, `course_id`, `collection_id`, `group_id`, `user_id`, and `site_script_id`.

## Streams notes

Thinkific collection endpoints use the `items` response envelope and page-number pagination with `page`/`limit`. Detail endpoints are single-object streams. The legacy `chapters` and `lessons` list streams remain because the Go connector emits them today, even though the current OpenAPI focuses on `/chapters/{id}`, `/chapters/{id}/contents`, and `/courses/{id}/chapters`.

The stream projections remain passthrough because the legacy connector emits raw records directly from Thinkific without a mapRecord transformation.

## Write actions & risks

Write actions require the standard reverse-ETL plan, preview, approval, execute flow. Covered actions create and update bundle enrollments, categories, collection memberships, enrollments, external orders, groups, group analysts/users, instructors, product publish decisions, promotions, users, and site scripts.

Deletes are marked destructive and allow 404 as an idempotent missing result. Required path fields are removed from the JSON body; remaining record fields become the request body.

## Known limits

- `POST /coupons`, `POST /coupons/bulk_create`, and `POST /course_reviews` are excluded because they require query parameters and `writes.json` has no query parameter field.
- OAuth authorization endpoints are excluded because they use the site-subdomain base URL rather than the Admin API v1 base.
- Thinkific Webhooks API endpoints are excluded because they use `https://api.thinkific.com/api/v2` and a different Authorization header surface.
- Config aliases accepted by legacy for `X-Auth-Subdomain`/`x_auth_subdomain` are narrowed to `subdomain` in this declarative bundle.
