# Overview

This bundle covers the Okta Admin Management API using Okta's published `management-minimal` OpenAPI spec. The legacy `users`, `groups`, and `system_logs` streams stay first and keep their existing schema projection; newly added streams use passthrough schemas generated from documented response record shapes.

## Auth setup

Set `base_url` to the Okta org URL, for example `https://your-org.okta.com`. Provide either `api_token` for SSWS auth or `access_token` for OAuth bearer auth. When both secrets are present, SSWS token auth is tried first, matching legacy precedence.

## Streams notes

The bundle declares 284 streams. Okta collection streams generally use Link-header pagination with `limit=200`; detail and singleton endpoints use no pagination. `system_logs` keeps the legacy `since` lower-bound behavior through optional `start_date` and emitted-state cursors.

## Write actions & risks

The bundle declares 429 write actions for JSON or bodyless POST, PUT, PATCH, and DELETE operations. These include admin lifecycle actions, app and policy management, credential and signing key operations, user and group operations, role/resource-set APIs, event hooks, realms, brands, and related deletes. All writes require approval; deletes are marked destructive and lifecycle/custom actions are high risk.

## Known limits

Excluded endpoints are limited to dialect gaps: scalar-array GET responses, XML SAML metadata reads, certificate/file/JWT payload writes, writes that require query parameters, and JSON mutations whose request body is a top-level array. Optional Okta filters are not exposed as config unless required by the endpoint. Private-key OAuth client authentication is still outside this bundle; use SSWS or bearer secrets.
