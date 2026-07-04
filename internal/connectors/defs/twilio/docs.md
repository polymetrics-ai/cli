# Overview

Twilio covers the public Twilio REST API v2010 (`api.twilio.com/2010-04-01`) plus the legacy
`internal/connectors/twilio` stream surface. Pass B expands the bundle from the five legacy read
streams to the full v2010 OpenAPI surface: every documented GET operation is represented as a
stream, and every documented POST/DELETE operation is represented as a typed write action.

The five legacy streams (`messages`, `calls`, `recordings`, `conferences`, `usage_records`) keep
their existing Tier-2 StreamHook path because legacy parity depends on Twilio's host-relative
`next_page_uri` behavior. New docs-derived streams are declarative and use Twilio's documented
`Page`/`PageSize` list parameters.

## Auth setup

Provide two secrets: `account_sid` and `auth_token`. The Account SID is used as the HTTP Basic
username and scopes account-relative paths such as `/Accounts/{AccountSid}/Messages.json`; the Auth
Token is used only as the HTTP Basic password. Both fields are declared `x-secret: true`.

`base.check` probes `GET /Accounts/{{ secrets.account_sid }}/Messages.json`, matching the legacy
non-mutating connectivity check.

## Streams notes

The legacy streams preserve their hand-written projection through `hooks/twilio`. Docs-derived
streams use `projection: "passthrough"` with minimal schemas so the engine preserves the live
Twilio response fields rather than pretending the generated catalog schema is exhaustive.

List streams use `page_number` pagination with `Page` starting at `0` and `PageSize=50`, which is
documented in Twilio's v2010 OpenAPI parameters. Detail streams use `records.single_object`.
Streams with path parameters other than `AccountSid` require the matching config key from
`spec.json` (`sid`, `call_sid`, `message_sid`, `country_code`, and similar).

## Write actions & risks

`writes.json` contains typed form-encoded actions for every documented v2010 POST and DELETE
operation. POST actions cover creates and Twilio's POST-as-update resources. DELETE actions send no
body, treat `404` as idempotent success, and are marked `confirm: "destructive"`.

All writes are mutating Twilio account operations and require reverse-ETL plan preview plus approval.

## Known limits

- This bundle is scoped to the v2010 public REST API file (`twilio_api_v2010.json`) and the legacy
  Twilio connector. Separate Twilio product APIs with distinct hosts or versions remain separate
  connector scope, not hidden exclusions in this bundle.
- The five legacy streams still carry per-stream `skip_dynamic` markers because their real runtime
  path is StreamHook-backed. The authoritative coverage remains `internal/connectors/paritytest/twilio`
  and `internal/connectors/hooks/twilio/hooks_test.go`.
- Docs-derived streams intentionally omit optional filter query parameters. They still cover the
  documented endpoint and preserve returned fields via passthrough projection.
