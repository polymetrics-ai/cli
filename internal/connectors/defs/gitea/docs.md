# Overview

Gitea's public REST API v1 is fully declared from the pinned Swagger document. This bundle intentionally exposes no executable read or write command until its provider-described contracts have been reviewed as connector-local schemas and fixtures.

## Auth setup

No Gitea operation is enabled in this declaration-only bundle. The public Swagger document describes token, basic-auth, and administrative sudo variants; a future enabled action must declare the least-privilege credential and scope required by that exact operation.

## Streams notes

No stream is enabled. Every documented GET and HEAD operation remains a blocked declaration pending a bounded record schema, pagination policy, sanitized fixture, and conformance evidence.

## Write actions & risks

No write action is enabled. Every documented POST, PUT, PATCH, and DELETE is declared in `api_surface.json`; destructive and administrative actions remain blocked until a named typed action supplies a bounded record schema, redaction, approval, and destructive confirmation where required.

## Known limits

This is a source-backed declaration surface, not live certification. Provider credentials are neither requested nor used; live certification remains pending.
