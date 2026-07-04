# Overview

Zendesk Support reads Support REST API v2 resources using the existing OAuth bearer or API-token basic-auth candidates. Pass B keeps the five legacy streams byte-for-byte in request and emitted record shape, then adds directly expressible top-level streams from the public Airbyte `source-zendesk-support` manifest plus `views`, which was already listed in this bundle's pending surface.

## Auth setup

Provide `base_url` as the Zendesk account root, for example `https://acme.zendesk.com`. Authentication can use either a secret `access_token` bearer token, or secret `email` plus secret `api_token` for Zendesk API-token basic auth. Secrets are used only by the engine authenticator.

## Streams notes

Implemented streams: `tickets`, `users`, `organizations`, `groups`, `satisfaction_ratings`, `deleted_tickets`, `account_attributes`, `attribute_definitions`, `brands`, `custom_roles`, `schedules`, `sla_policies`, `tags`, `ticket_fields`, `ticket_forms`, `topics`, `user_fields`, `automations`, `categories`, `sections`, `articles`, `group_memberships`, `macros`, `organization_fields`, `organization_memberships`, `posts`, `ticket_activities`, `ticket_audits`, `ticket_metric_events`, `ticket_events`, `ticket_skips`, `triggers`, `views`.

The original `tickets`, `users`, `organizations`, `groups`, and `satisfaction_ratings` streams retain the legacy `/api/v2/<resource>` paths, cursor pagination settings, and field schemas. Added top-level streams use the Airbyte/Zendesk response key for extraction and a per-stream next-link paginator (`links.next`, `next_page`, `after_url`, or `before_url`) based on the manifest paginator.

Nested article/post votes and comments, side conversations, `ticket_metrics`, and dynamic interval exports are listed in `api_surface.json` as exclusions because they require parent partition routing, state-delegating stream orchestration, or dynamic interval query templates that this declarative bundle is not modeling in this pass.

## Write actions & risks

`writes.json` exposes allow-listed CRUD-style mutations for `ticket`, `user`, `organization`, `group`, `macro`, `trigger`, `automation`, `view`, and `ticket_field`. Create/update actions send Zendesk's documented wrapper object shape such as `{ "ticket": {...} }`; delete actions are marked destructive and idempotent for 404s. All writes require reverse-ETL approval.

Uploads, redactions, password/session/compliance operations, voice display helper endpoints, async imports, and bulk job endpoints are explicitly excluded from the surface because they are binary, destructive/admin-only, or async/compound operations rather than safe row-level reverse-ETL actions.

## Known limits

- This is not a blanket generic Zendesk HTTP client. Only named streams and named write actions are exposed.
- Several Zendesk endpoints are tenant-, plan-, or permission-dependent. Permission failures should surface through the configured error map rather than being hidden.
- True incremental exports are represented only where existing legacy behavior already supported them; dynamic Airbyte interval templates are not copied into ordinary collection streams.
