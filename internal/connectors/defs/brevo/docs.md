# Overview

Brevo (formerly Sendinblue) is a wave2 fan-out declarative-HTTP migration. It reads Brevo contacts,
email campaigns, contact lists, and senders through the Brevo REST API v3
(`GET https://api.brevo.com/v3/...`). This bundle migrates `internal/connectors/brevo` (the
hand-written connector); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide a Brevo API key via the `api_key` secret; it is sent as the `api-key` header
(`api_key_header` auth mode, matching legacy's `connsdk.APIKeyHeader("api-key", secret, "")`) and
is never logged. `base_url` defaults to `https://api.brevo.com/v3` and may be overridden for
tests/proxies.

## Streams notes

`contacts` and `email_campaigns` (Brevo's `/emailCampaigns` endpoint; stream renamed to snake_case
per this repo's naming convention, §2) share the same shape: `GET` list endpoints paginated with Brevo's
offset/limit convention (`pagination.type: offset_limit`, `limit_param: limit`, `offset_param:
offset`, `page_size: 100` — matches legacy's `brevoDefaultPageSize`); records live at `contacts`/
`campaigns` respectively (legacy's `recordsPath`). Both support Brevo's `modifiedSince` incremental
filter (`incremental.request_param: modifiedSince`, `param_format: rfc3339` — sent verbatim as the
persisted cursor or, on a fresh sync, the RFC3339 `start_date` config value, identical to legacy's
`incrementalLowerBound`). `contacts_lists` (`GET /contacts/lists`, records at `lists`) is also
offset/limit paginated but has no incremental filter, matching legacy (`supportsModifiedSince:
false`). `senders` (`GET /senders`, records at `senders`) is a single-request, non-paginated
endpoint (legacy's `paginated: false`) — no `pagination` block is declared for it, matching the
engine's `none` default.

## Write actions & risks

None. Brevo is read-only in this connector (legacy's own `Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes `page_size` (1-1000,
  default 100) and `max_pages` (0/all/unlimited or a positive integer) as config-driven overrides
  read fresh on every `Read` call (`brevoPageSize`/`brevoMaxPages`). The engine's `offset_limit`
  paginator reads its page size only from `pagination.page_size` (a fixed literal baked into
  `streams.json`, sourced once at bundle-load time), with no `config.*`-templated override
  mechanism — matching the exact limitation documented in this repo's `bitly` bundle for its
  `next_url` paginator. This bundle declares `page_size: 100` to match legacy's real default
  exactly; a caller cannot raise or lower it at read time as legacy allowed. `spec.json` still
  declares `page_size`/`max_pages` for documentation continuity with legacy's config surface, but
  neither is wired into any template (F6-adjacent: these are the same "legacy had this knob, the
  engine's chosen pagination type cannot express it" shape as bitly's `page_size`, not silently
  dropped dead config).
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance) stamps a
  `previous_cursor` field onto fixture-mode records when a prior cursor is set. This bundle's
  schemas and parity target the live wire shape only; the engine's own conformance/fixture-replay
  harness supersedes legacy's in-code fixture-mode path.
- The `contacts` stream's 2-page conformance fixture (`fixtures/streams/contacts/{page_1,
  page_2}.json`) uses a full 100-record page 1 (matching the real `page_size: 100`) followed by a
  1-record short page 2, so pagination truly terminates on Brevo's own short-page signal rather than
  an artificially-lowered page size — this keeps the fixture's wire shape identical to a real
  production page count instead of trading fixture verbosity for behavioral accuracy.
