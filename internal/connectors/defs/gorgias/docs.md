# Overview

Gorgias connector parity for the official Gorgias REST API reference. The bundle is definition-owned and covers the complete 114-operation official inventory from `https://developers.gorgias.com/llms.txt` and linked `reference/*.md` pages.

Official lane representation:

- 38 ETL read operations plus 4 CDC/changefeed-labelled reads are exposed as 42 stream-backed read surfaces.
- 63 official reverse-ETL write operations are represented as typed approval-gated write actions.
- 3 provider-query/search operations are represented as 2 bounded direct reads plus 1 blocked non-mutating PUT search row.
- 6 binary/file/voice-recording operations are represented as blocked file/binary operation metadata until the shared bounded binary executor exists.

Executable surfaces in this bundle are 42 streams, 63 write actions, and 2 bounded direct reads. The remaining 7 operations are represented as planned/blocked rows because this connector cannot safely execute them without a shared binary-download/file-upload or provider-search foundation.

## Auth setup

Connection fields:

- `base_url` (required): full Gorgias API root such as `https://example.gorgias.com/api`.
- `username` (required): Gorgias account email for HTTP Basic authentication.
- `password` (required secret): Gorgias API key used as the Basic auth password.
- `page_size` (optional, default `100`): cursor-paginated stream page size.
- Scoped detail streams such as customer, ticket, view, macro, or voice-call detail reads require the corresponding connector config path parameter documented in `spec.json`.

No fixture, docs, or CLI metadata stores live credentials.

## Streams notes

Streams use connector-relative paths because `base_url` includes `/api`. List streams use cursor pagination with `cursor` and `meta.next_cursor`; detail streams are single-object reads. The first stream (`tickets`) carries a two-page fixture to prove cursor termination; every declared stream has a sanitized fixture page.

Implemented stream names:

- `tickets` -> GET `/tickets`
- `account` -> GET `/account`
- `account_settings` -> GET `/account/settings`
- `custom_fields` -> GET `/custom-fields`
- `custom_field` -> GET `/custom-fields/{{ config.custom_field_id }}`
- `customers` -> GET `/customers`
- `customer_custom_fields_values` -> GET `/customers/{{ config.customer_id }}/custom-fields`
- `customer` -> GET `/customers/{{ config.customer_id }}`
- `events` -> GET `/events`
- `event` -> GET `/events/{{ config.event_id }}`
- `integrations` -> GET `/integrations`
- `integration` -> GET `/integrations/{{ config.integration_id }}`
- `jobs` -> GET `/jobs`
- `job` -> GET `/jobs/{{ config.job_id }}`
- `macros` -> GET `/macros`
- `macro` -> GET `/macros/{{ config.macro_id }}`
- `messages` -> GET `/messages`
- `metric_card` -> GET `/metric-cards/{{ config.slug }}`
- `voice_call_events` -> GET `/phone/voice-call-events`
- `voice_call_event` -> GET `/phone/voice-call-events/{{ config.voice_call_event_id }}`
- `voice_calls` -> GET `/phone/voice-calls`
- `voice_call` -> GET `/phone/voice-calls/{{ config.voice_call_id }}`
- `rules` -> GET `/rules`
- `rule` -> GET `/rules/{{ config.rule_id }}`
- `satisfaction_surveys` -> GET `/satisfaction-surveys`
- `satisfaction_survey` -> GET `/satisfaction-surveys/{{ config.satisfaction_survey_id }}`
- `tags` -> GET `/tags`
- `tag` -> GET `/tags/{{ config.tag_id }}`
- `teams` -> GET `/teams`
- `team` -> GET `/teams/{{ config.team_id }}`
- `ticket` -> GET `/tickets/{{ config.ticket_id }}`
- `ticket_custom_fields` -> GET `/tickets/{{ config.ticket_id }}/custom-fields`
- `ticket_messages` -> GET `/tickets/{{ config.ticket_id }}/messages`
- `ticket_message` -> GET `/tickets/{{ config.ticket_id }}/messages/{{ config.ticket_message_id }}`
- `ticket_tags` -> GET `/tickets/{{ config.ticket_id }}/tags`
- `users` -> GET `/users`
- `user` -> GET `/users/{{ config.user_id }}`
- `views` -> GET `/views`
- `view` -> GET `/views/{{ config.view_id }}`
- `view_items` -> GET `/views/{{ config.view_id }}/items`
- `widgets` -> GET `/widgets`
- `widget` -> GET `/widgets/{{ config.widget_id }}`

CDC/changefeed-labelled official operations (`events`, `event`, `voice_call_events`, `voice_call_event`) are exposed as ordinary bounded ETL streams only. `metadata.capabilities.cdc` remains `false` because there is no live CDC subscription reader in this connector.

## Write actions & risks

Every documented official reverse-ETL mutation lane is a named write action in `writes.json`, including POST statistic retrieval endpoints that remain approval-gated because the official ledger classifies them in the reverse-ETL lane. The engine executes write actions only through the existing reverse ETL path: plan -> preview -> explicit approval -> execute. DELETE, merge, archive, cancellation, and similar destructive/admin-style actions declare `confirm: destructive`; idempotent DELETE actions treat `404` as missing-ok where safe.

No raw method/path/body, arbitrary GraphQL, shell, generic SQL write, or generic HTTP write command is exposed. Binary/file operations are not promoted to generic write actions in this revision.

## Known limits

- Fixture/conformance evidence is not live provider certification; `certification.json` records fixture-only status.
- `PUT /views/{view_id}/items` is documented by Gorgias as search-like, but non-mutating PUT provider-query execution is blocked until foundation #2985 provides a safe contract.
- `POST /upload`, `GET /{file_type}/download/{domain_hash}/{resource_name}`, `POST /stats/{name}/download`, and voice-call recording binary operations remain blocked rows until bounded binary/file executors and output-file/payload policies are available.
- No live provider calls, credential checks, or certification runs were performed for issue #196.
