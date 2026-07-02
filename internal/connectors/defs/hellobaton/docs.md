# Overview

Hellobaton is a wave2 fan-out declarative-HTTP migration. It reads Hellobaton projects,
milestones, tasks, phases, companies, and users through the Hellobaton REST API
(`GET https://<company>.hellobaton.com/api/...`). This bundle targets capability parity with
`internal/connectors/hellobaton` (the hand-written connector it migrates); the legacy package
stays registered and unchanged until wave6's registry flip. Hellobaton's own public API reference
is not consistently reachable (per-customer-subdomain documentation), so the legacy Go source is
this bundle's ground truth, per the migration conventions' legacy-over-docs precedence rule.

## Auth setup

Provide a Hellobaton API key via the `api_key` secret; it is sent as the `api_key` query parameter
on every request (`api_key_query` auth mode), matching legacy's `connsdk.APIKeyQuery("api_key",
secret)` (`hellobaton.go:231`). It is never logged (the engine's `api_key_query` auth mode routes
the value through the same secret-redaction path as every other auth mode; `DryRunWrite` previews
and structured logs never surface it).

`base_url` is **required** in this bundle (unlike legacy, which derives
`https://<company>.hellobaton.com/api` from a bare `company` config value when `base_url` is
unset — see Known limits). Provide the full API base URL directly, e.g.
`https://acme.hellobaton.com/api`.

## Streams notes

All 6 streams (`projects`, `milestones`, `tasks`, `phases`, `companies`, `users`) share the
identical shape: `GET` against the Hellobaton list endpoint, records at the DRF-style `results`
key, primary key `["id"]`, advisory cursor field `modified` (full refresh only — see below).
Pagination follows Hellobaton's Django-REST-Framework convention (`pagination.type: next_url`,
`next_url_path: "next"`): the response envelope is `{count, next, previous, results:[...]}` where
`next` is the ABSOLUTE URL of the following page (or `null`/absent when exhausted), matching
legacy's `harvest` function exactly (`hellobaton.go:135-174`). `page_size=100` (legacy's own
`hellobatonDefaultPageSize`) is declared as a static per-stream `query` value and is, per the
engine's `readDeclarative` (`mergeQuery`), re-sent on every page including next-URL-addressed
pages — unlike legacy, which explicitly resets to an empty `url.Values{}` once it follows an
absolute next-page URL (`hellobaton.go:170-171`), sending `page_size` only on the first request.
This is a wire-request-shape divergence from legacy, verified benign in DATA terms only because
DRF's own paginator embeds the CURRENT page's `page_size` in its `next` URL, so the engine's
re-applied `page_size=100` value is idempotent with what the next URL already carries (same
pattern as bitly's identical `size=50` divergence, documented in `docs/migration/conventions.md`
§5/bitly's own `docs.md`). If a Hellobaton `next` URL ever omitted or diverged from `page_size`,
this bundle's request would differ from legacy's; today it does not.

None of the 6 endpoints expose a server-side incremental filter parameter (legacy's own package
doc: "Only full-refresh reads are supported upstream"); `modified` is declared as an advisory
`x-cursor-field` on every schema (matching legacy's advisory `CursorFields`, `hellobaton/
streams.go:35-72`) but no stream declares an `incremental` block — full refresh only, exactly
matching legacy.

## Write actions & risks

None. Hellobaton has no obviously-safe reverse-ETL writes in the legacy connector
(`Capabilities: Write: false`); this bundle ships no `writes.json`, matching legacy's `Write`
returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`base_url` is required; legacy's `company`-subdomain derivation is not modeled.** Legacy
  derives a default base URL from a bare `company` config value
  (`https://<company>.hellobaton.com/api`, `hellobaton.go:248-259`, with a DNS-label-charset
  validator on the subdomain) when `base_url` is unset. The engine's `spec.json` `"default"`
  materialization mechanism only fills in a FIXED literal default, not a value derived from
  another config field — this is exactly the documented "derived default" gap in
  `docs/migration/conventions.md` §3 (sentry's `hostname`-based URL, chargebee's `site`-based URL
  are the same shape). Per that section's guidance, this bundle requires `base_url` directly and
  drops the `company`-derivation convenience rather than inventing ad hoc Go for it — a documented
  config-surface narrowing, not a silent behavior change: any caller that already knows its
  Hellobaton subdomain supplies the equivalent full URL instead.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) stamps `connector: "hellobaton"` and `fixture: true` markers
  plus a conditional `previous_cursor` field onto every fixture-mode record
  (`hellobaton.go:178-214`). None of these are part of the LIVE record shape; this bundle's schemas
  target the live path only. The engine's own conformance/fixture-replay harness supplies the
  credential-free test affordance this bundle needs.
- **Sanctioned single-page fixture exception (conventions.md §4).** Every stream in this bundle
  uses `next_url` pagination, whose next-page URL is the replay server's own address — unknown
  until the harness picks a port at runtime, so a static fixture file cannot embed the correct
  absolute URL for a second page. Per the sanctioned exception, each stream ships a single-page
  conformance fixture (satisfies `fixtures_present`/`read_fixture_nonempty`); this is honest
  test-harness scoping, not a fixture-authoring shortcut, matching bitly's and calendly's identical
  precedent.
