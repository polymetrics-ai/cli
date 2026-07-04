# PandaDoc

## Overview

This bundle covers the current official PandaDoc public API reference from https://developers.pandadoc.com/reference/about. The three legacy streams keep their exact projected fields; Pass B adds documented JSON reads across documents, templates, contacts, content library items, forms, logs, members, SMS opt-outs, webhook events/subscriptions, workspaces, document settings/audit trails, notary resources, and product catalog resources.

## Auth setup

Provide an API key as `secrets.api_key`. It is sent as `Authorization: API-Key <key>`. The default `base_url` is now `https://api.pandadoc.com` so versioned `/public/v1`, `/public/v2`, and `/public/beta` paths can coexist in one bundle. The legacy streams still resolve to the same full URLs they used before.

## Streams notes

Legacy `documents`, `templates`, and `contacts` preserve their schema projection and cursor-field declarations. Newly added streams use passthrough projection with permissive schemas because PandaDoc responses are resource-specific and the declarative engine can safely emit their JSON records as returned. List streams read from `results` and use PandaDoc's `next` URL pagination when present; detail/status streams are single-object reads with pagination disabled.

## Write actions & risks

`writes.json` exposes documented JSON-body and no-body mutations, including document lifecycle actions, contact/template/folder/webhook/catalog/notary/admin operations, and destructive deletes. Destructive delete actions are marked with destructive confirmation metadata. Sending, reminder, status-change, ownership, and admin actions mutate live PandaDoc workflow state and require the normal reverse-ETL plan, preview, approval, execute flow.

## Known limits

- Multipart upload endpoints are excluded as `binary_payload` because the write dialect cannot build multipart file bodies.
- Download endpoints are excluded as `binary_payload` because they return files rather than JSON records.
- OAuth token exchange, member-token creation, API-key generation, and webhook shared-key rotation are excluded because they manage credential or secret material rather than ordinary business records.
- Some list streams rely on PandaDoc's default paging parameters until a `next` URL is returned; the legacy streams continue to send `count` and `page=1` exactly as before.
- Legacy accepts a runtime `max_pages` cap, but the declarative engine only supports fixed bundle-authored `pagination.max_pages` integers. This bundle intentionally does not declare an ignored `max_pages` `spec.json` property.
