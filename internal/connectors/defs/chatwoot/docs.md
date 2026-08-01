# Chatwoot connector

## Overview

The Chatwoot connector reads account-scoped conversations, contacts, inboxes, labels, agents, teams, and messages from the official Chatwoot Application API. This wave also records every operation from the official application, platform, client, and other OpenAPI tag-group sources in `api_surface.json` and `operations.json`.

Official operation ledger counts: 145 total rows; 7 stream-covered rows; 73 reverse-ETL write-covered rows; 65 planned/blocked rows.

## Auth setup

Use `pm credentials add <name> --connector chatwoot` with `--from-env api_access_token=ENV` or `--value-stdin api_access_token`. Do not paste tokens into chat, shell history, docs, fixtures, or JSON examples. `base_url` is the Chatwoot server root URL (default `https://app.chatwoot.com`) and `account_id` scopes application/platform account endpoints.

The public client API documented by Chatwoot does not use the same user/platform token contract; those operations are represented as planned/blocked until a separate no-auth inbox/contact-safe connector contract exists.

## Streams notes

Implemented fixture-backed streams remain account-scoped and bounded: `conversations`, `contacts`, `inboxes`, `agents`, `teams`, `labels`, and fan-out `messages`. Additional documented GET/report/search/changefeed operations are present exactly once in the ledger and remain planned/blocked until typed stream/direct-read schemas and fixtures are authored.

The connector uses bounded page-number pagination for paginated streams and fixture replay only in local conformance. No live provider call is made by this bundle.

## Write actions & risks

This bundle declares 73 named reverse-ETL write actions for safely expressible Chatwoot application/platform mutations. Every write must follow plan -> preview -> explicit approval token -> execute. 49 DELETE/destructive/admin/elevated actions carry `confirm: destructive` and require the typed `--confirm destructive` challenge before execution. Delete actions are modeled as idempotent for 404 responses where the resource is already absent.

Public client writes and write endpoints requiring query-parameter execution support remain planned/blocked rather than exposed through a raw API escape hatch. There is no generic HTTP method/path/body tool.

## Known limits

- Live certification is not claimed; no credentials or provider writes were used.
- Public client API operations need a separate no-auth, inbox/contact-bounded contract before execution.
- `POST /api/v1/accounts/{account_id}/custom_filters` requires a query parameter on a write path; this connector-local slice records it as blocked because the shared write engine has no query-param write contract.
- Additional GET/report/changefeed operations beyond the seven fixture-backed streams are operation-ledger rows until typed stream or direct-read fixtures prove safe execution.
- Generated operation schemas are bounded metadata for planning and validation; provider-specific edge semantics still require live-safe certification before certification claims.
