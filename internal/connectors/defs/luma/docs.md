# Luma

## Overview

Luma uses the official OpenAPI spec at https://public-api.luma.com/openapi.json. This bundle covers the current public API on `https://public-api.luma.com`: events, calendars, guests, contacts, contact tags, event tags, coupons, ticket types, memberships, webhooks, and organization resources. The legacy stream names `events`, `event_guests`, and `event_hosts` remain available; `event_hosts` is now extracted from the documented event detail response because the old standalone host-list endpoint is no longer present in the OpenAPI spec.

## Auth setup

Create a Luma API key for the calendar or organization you want to manage and store it as `api_key`. Requests authenticate with the `x-luma-api-key` header. The default `base_url` is `https://public-api.luma.com`; override it only for tests or a trusted proxy.

## Streams notes

Cursor-paginated list streams use `pagination_cursor`, `has_more`, and `next_cursor`. Detail and lookup streams override pagination with `type: none`. Event-scoped streams use `event_id`; the legacy-compatible `event_guests` and `event_hosts` streams continue to read `event_api_id` so existing configs do not change. New Pass B streams use passthrough projection with OpenAPI-derived schemas.

## Write actions & risks

The bundle exposes 40 POST write actions for documented Luma mutations, including event create/update/cancel, guest status/invites/add, host changes, coupons, contact and event tags, calendar event approval/rejection, image upload URL creation, ticket types, memberships, webhooks, calendars, and organization event transfers. Delete/cancel/invite/reject/transfer style actions are marked for destructive confirmation because they can remove data, notify guests, reject submissions, or move ownership.

## Known limits

- The old `docs.lu.ma` URL is retired; metadata points to `docs.luma.com`, and `api_surface.json` points at the OpenAPI document.
- The current OpenAPI uses `id` and `calendar_id`; legacy stream schemas still expose `api_id` and `calendar_api_id` by computed fields that prefer legacy wire keys (`api_id`, `calendar_api_id`, `event_api_id`, `user_api_id`) and fall back to current OpenAPI record keys (`id`, `calendar_id`, `user_id`) where the dialect supports a record-to-record fallback.
- `event_hosts.access_level` is copied only when the source host object includes it. The earlier standalone host-list endpoint carried this field; the current event-detail hosts array may omit it, and the bundle does not fabricate a replacement value.
- Optional filter query parameters from the OpenAPI are not all surfaced as config options in this pass; streams cover the base list/detail resources and required identifiers.
- Luma's image flow creates an upload URL only. Uploading binary image bytes to that URL is outside this connector's JSON write dialect and is not modeled as a separate write action.
