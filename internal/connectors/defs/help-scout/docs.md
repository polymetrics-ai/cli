# Overview

Help Scout is a wave2 fan-out declarative-HTTP migration. It reads Help Scout conversations,
customers, mailboxes, and users through the Mailbox API v2
(`GET https://api.helpscout.net/v2/...`), authenticating with OAuth2 client-credentials. This
bundle targets capability parity with `internal/connectors/help-scout` (the hand-written connector
it migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Help Scout OAuth2 application's `client_id` and `client_secret` secrets; the engine's
`oauth2_client_credentials` auth mode exchanges them against `token_url` (default
`https://api.helpscout.net/v2/oauth2/token`) for a bearer access token, matching legacy's
`connsdk.OAuth2ClientCredentials` (`help_scout.go:266-271`) exactly (`grant_type=client_credentials`
form POST). Both secrets are never logged. `base_url` defaults to `https://api.helpscout.net/v2`;
both `base_url` and `token_url` may be overridden for tests/proxies.

## Streams notes

All 4 streams (`conversations`, `customers`, `mailboxes`, `users`) share the identical shape: `GET`
against the Mailbox API's HAL+JSON list endpoint, records at `_embedded.<resource>` (a dotted
`records.path`), primary key `["id"]`. Pagination follows Help Scout's HAL page-envelope convention
(`pagination.type: page_number`, `page_param: page`, `size_param: size`, `page_size: 50`) —
legacy's own `harvest` loop (`help_scout.go:145-200`) walks `page=1..totalPages` reading
`page.totalPages` from the body, with a defensive short-page fallback when the envelope is missing.
The engine's `page_number` paginator stops purely on a short page (fewer than `page_size` records
returned), which is data-equivalent to legacy's primary `totalPages`-driven stop for any real Help
Scout dataset: Help Scout always returns exactly `size` records per page except the last, so "a
page returned fewer than `page_size` records" and "this was the last page per `page.totalPages`"
never disagree in practice; only legacy's own defensive fallback path (used only when the page
envelope itself is malformed/missing) is the one legacy-side branch this pagination shape does not
individually reproduce, and that branch is itself just the same short-page stop rule legacy applies
when the primary signal is unavailable.

Both `conversations` and `customers` (and, per legacy's blanket `harvest`, every stream) send
`sortField=modifiedAt&sortOrder=asc` and, when `start_date` is configured, `modifiedSince=<value>`
verbatim (`help_scout.go:150-154`). This bundle expresses `modifiedSince` as an **optional query
param** (`{"template": "{{ config.start_date }}", "omit_when_absent": true}`), not an `incremental`
block: legacy never reads a persisted sync cursor back into this filter (`harvest` reads only
`req.Config.Config["start_date"]`, never `req.State["cursor"]`) — it always resends the exact same
raw `start_date` config value on every sync, with no forward advancement. Declaring an `incremental`
block instead (with `start_config_key: start_date`) would change this: the engine persists and
re-reads a genuinely advancing state cursor between syncs, which is not what legacy does. Modeling
`modifiedSince` as a plain optional per-sync filter, not an advancing incremental cursor, is the
literal parity-preserving representation of legacy's actual (non-cursor-advancing) behavior.
`userUpdatedAt`/`updatedAt` are still declared as advisory `x-cursor-field` values on every schema
(matching legacy's advisory `CursorFields`), but no stream declares an `incremental` block.

The `check` request (`GET /mailboxes`) omits legacy's `page=1` query param: the engine's `check`
dispatch (`engine/read.go`'s `Check`) never attaches a query to the check request at all
(`RequestSpec` has no query field), so this bundle's `check` intentionally matches that
engine-wide constraint rather than declaring an unusable field. This has no data-parity
consequence — `check` only confirms auth/connectivity, never emits or counts records.

## Write actions & risks

None. Help Scout has no obviously-safe reverse-ETL writes in the legacy connector (`Capabilities:
Write: false`); this bundle ships no `writes.json`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`.

## Known limits

- **Dynamic (fixture-replay) conformance checks are skipped bundle-wide** (`metadata.json`'s
  `conformance.skip_dynamic`). `oauth2_client_credentials` auth requires a real, resolvable
  `token_url`; conformance's synthetic non-secret config value (`"synthetic-conformance-value"`)
  is not a resolvable URL, so the token exchange fails before any declarative stream/check request
  is ever issued — every auth-resolving dynamic check would otherwise fail identically and
  uninformatively for that reason alone, not because of anything else in this bundle's shape.
  Static checks (spec/schema validity, `interpolations_resolve`, docs/fixtures presence, secret
  redaction) still run and pass. This bundle has no Tier-2 hook (auth is fully declarative
  `oauth2_client_credentials`), so there is no `paritytest/help-scout` package for this wave; the
  read/pagination/schema-projection shape is proven by structural review against legacy
  `internal/connectors/help-scout` instead — matches sendpulse's and dwolla's identical documented
  precedent (`docs/migration/conventions.md` §4).
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) stamps a fixed, deterministic 2-record set across every stream
  and appends `previous_cursor` when `req.State["cursor"]` happens to be set
  (`help_scout.go:205-248`). None of these are part of the LIVE record shape; this bundle's schemas
  target the live path only. The engine's own conformance/fixture-replay harness supplies the
  credential-free test affordance this bundle needs.
- **The `check` request's `page=1` query param is not modeled** (see Streams notes above) — the
  engine's declarative check dispatch never sends a query on the check request at all, for any
  bundle. This has no data-parity impact since `check` only verifies auth/connectivity.
