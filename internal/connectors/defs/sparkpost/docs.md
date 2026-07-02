# Overview

SparkPost is a wave2 fan-out declarative-HTTP migration. It reads recipient lists, templates,
sending domains, transmissions, and suppression-list records through the SparkPost API v1 (`GET
{{ config.base_url }}/...`). This bundle is migrated from `internal/connectors/sparkpost` (the
hand-written connector it replaces); the legacy package stays registered and unchanged until
wave6's registry flip. Read-only (`capabilities.write` is `false`, matching legacy's `Write`
returning `connectors.ErrUnsupportedOperation`).

## Auth setup

Provide a SparkPost API key via the `api_key` secret; it is sent verbatim as the `Authorization`
header value with no prefix (`mode: api_key_header`, empty `prefix`), matching legacy's
`connsdk.APIKeyHeader("Authorization", key, "")`. `base_url` defaults to
`https://api.sparkpost.com/api/v1` (legacy's US-region default) and should be set explicitly to
`https://api.eu.sparkpost.com/api/v1` for EU-region accounts — see Known limits for how this
bundle represents legacy's region selection.

## Streams notes

All 5 streams share the identical shape: `GET`, records at `results`, and two optional query
params — `from`/`to`, sourced from `start_date`/`end_date` config (`omit_when_absent: true`, so
neither param is sent on a full, unbounded read) — matching legacy's `copyConfig(q, cfg,
"start_date", "from")`/`copyConfig(q, cfg, "end_date", "to")`, which only sets each query param
when the corresponding config value is non-empty. No pagination is declared for any stream,
matching legacy's `readRecords`, which reads the `results` array from a single response with no
paging loop. Primary keys match each endpoint's natural identifier: `id` for `recipient_lists`/
`templates`/`transmissions`, `domain` for `sending_domains`, `recipient` for `suppression_list`.
None declare an incremental cursor — legacy's `from`/`to` are report-style bounding filters passed
straight through per-read, not a persisted-state cursor the connector advances itself.

## Write actions & risks

None. Legacy `sparkpost.go`'s `Write` returns `connectors.ErrUnsupportedOperation`
unconditionally; `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **Legacy's `api_prefix` config knob (`api` vs `api.eu`, validated to exactly those two values)
  is not modeled as a separate spec property.** Legacy derives the base URL as
  `https://<api_prefix>.sparkpost.com/api/v1` when `base_url` itself is unset, defaulting
  `api_prefix` to `api` and rejecting any other value. This bundle instead exposes a single
  `base_url` property (`default: "https://api.sparkpost.com/api/v1"`), documented to be set
  explicitly to the full EU URL for EU accounts — functionally equivalent (the same two possible
  base URLs are reachable) but expressed as one directly-settable URL rather than a
  validated-enum-then-derived-URL. This is a config-surface narrowing (fewer invalid input shapes
  are even representable, not fewer reachable base URLs), not an emitted-data change.
- **`base_url` scheme/host validation (`http`/`https` only, non-empty host) is enforced by legacy
  in Go** (`validatedBaseURL`) with a specific error message; the engine has no equivalent
  declarative URL-shape validator, so a malformed `base_url` here surfaces as a generic
  request-construction or connection error rather than legacy's dedicated `"config base_url must
  use http or https"`/`"must include a host"` messages. This never changes behavior for any valid
  `base_url`.
- The full SparkPost API surface (transmission sends, webhooks, subaccounts, message events) is
  out of scope for this wave; see `api_surface.json`'s `excluded` entries.
