# Overview

Freshchat is a customer-messaging product. This bundle covers the JSON-oriented Freshchat v2 API
surface: account configuration, users, conversations, messages, agents, groups, channels, roles,
outbound messages, report status, metrics, and business-hours checks. It also exposes plain JSON
write actions for users, conversations, agents, outbound WhatsApp messages, report extraction, and
CSAT ratings.

## Auth setup

Provide a Freshchat API key via the `api_key` secret; it is sent as `Authorization: Bearer
<api_key>` and is never logged.

## Streams notes

The legacy-parity streams (`agents`, `users`, `groups`, `channels`, `roles`) preserve the legacy
projected record fields. `agents`, `users`, and `channels` publish `updated_time` as a cursor field
but do not send a server-side incremental filter, matching legacy behavior.

Detail streams (`agent_details`, `user_details`, `conversation_detail`, `conversation_messages`,
`report_status`) are config-scoped because the Freshchat API exposes point lookup paths but no
universal account-wide conversation/report listing endpoint that the engine can use to enumerate
all ids safely. Set the corresponding id config only when reading those streams.

Metrics streams use Freshchat's query-parameter API. `historical_metrics` requires
`metrics_metric`, `metrics_start`, and `metrics_end`; `instant_metrics` requires `metrics_metric`.
Optional grouping, filtering, aggregator, interval, and summary query values are passed only when
their config keys are set.

## Write actions & risks

Write actions create, update, or delete Freshchat users and agents; create or update
conversations; send conversation messages; send outbound WhatsApp messages; request report
extraction; and create CSAT ratings. Deletes are marked destructive and treat 404 as idempotent
success when Freshchat reports the target is already absent. All reverse ETL writes still require
plan preview and approval.

## Known limits

- Multipart file and image uploads are excluded as `binary_payload`; they require multipart/binary
  transfer semantics that the declarative JSON/form write dialect does not model.
- `/users/fetch` is excluded because it is a read-like POST endpoint whose useful behavior depends
  on request-body search criteria. The current engine read path declares `stream.body` but does not
  send it on requests, so modeling it as a stream would silently issue the wrong request.
- **`base_url` is required config, not derived from `account_name`.** Legacy derives the base URL
  from an `account_name` config value when `base_url` is unset
  (`https://<account_name>.freshchat.com/v2`), validating that `account_name` is a bare subdomain
  with no `/:. ` characters. The engine's `spec.json` `"default"` materialization mechanism only
  supports a fixed literal default value, not one derived from another config value at read/check
  time. This bundle therefore requires the fully-formed `base_url` directly.
