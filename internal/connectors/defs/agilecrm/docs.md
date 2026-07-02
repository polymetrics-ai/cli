# Overview

AgileCRM is a wave2 fan-out declarative-HTTP migration. It reads AgileCRM contacts, deals, tasks,
and milestone pipelines through the AgileCRM REST API (`GET https://<domain>.agilecrm.com/dev/api/...`).
This bundle targets capability parity with `internal/connectors/agilecrm` (the hand-written
connector it migrates); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide the AgileCRM account `email` (config) and `api_key` (secret); they are sent as HTTP Basic
auth (`email:api_key`, base64-encoded) and the secret is never logged, matching legacy's
`connsdk.Basic(email, secret)` (`agilecrm.go:254`). The account subdomain is provided via the
required `domain` config value and templated directly into the base URL
(`https://{{ config.domain }}.agilecrm.com/dev/api`), matching legacy's derived
`https://%s.agilecrm.com/dev/api` (`agilecrm.go:286`) — `domain` is restricted to the safe
alphanumeric+hyphen label charset legacy itself enforces (`validDomain`), expressed here as a
`spec.json` `pattern`.

## Streams notes

`contacts` and `deals` (`opportunity`) paginate via AgileCRM's last-record cursor convention: a
full page's LAST array element carries a `cursor` string field, resubmitted as `?cursor=...` to
fetch the next page; an absent `cursor` on the last element signals the end of the list. This is
expressed as `pagination.type: cursor` with `last_record_field: cursor` (no `stop_path` — the
engine's `lastRecordCursor` paginator already stops exactly when the last-record field is absent,
matching legacy's own stop condition with no additional signal needed). The raw `cursor` field is
never declared in either stream's schema, so schema projection drops it from every emitted record —
matching legacy's explicit `delete(item, "cursor")` before emitting (`agilecrm.go:178`).
`page_size=50` is sent as a static per-stream query value (AgileCRM's own default,
`agilecrmDefaultPageSize`); see Known limits for why it is not config-driven.

`tasks` and `milestone` (`milestone/pipelines`) return a single bounded top-level JSON array with no
pagination at all (`pagination.type: none`), matching legacy's `paginated: false` routing-table
entries exactly.

Every stream's records live at the top-level array root (`records.path: "."`).

`owner_id` is derived via a `computed_fields` rename from the nested `owner.id` object AgileCRM's
real API returns (`{"owner":{"id":...}}`), matching legacy's `ownerID` helper's PRIMARY branch
(`agilecrm.go:172-178`). See Known limits for legacy's secondary (top-level `owner_id`) fallback
branch, which is not modeled.

None of AgileCRM's core list endpoints expose an incremental cursor request parameter (legacy's own
package doc: read-only, full-refresh reads only); this bundle declares no `incremental` block for
any stream, matching legacy exactly. `created_time`/`updated_time` are still published as
`x-cursor-field` (contacts/deals/tasks use `created_time`) purely for downstream
`incremental_append_deduped` sync-mode eligibility, matching legacy's own published
`CursorFields`.

## Write actions & risks

None. AgileCRM's core list endpoints have no obviously-safe reverse-ETL writes (legacy's own
package doc: "no obviously safe reverse-ETL write actions"); `capabilities.write` is `false` and
this bundle ships no `writes.json`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`.

## Known limits

- **`base_url` override is not modeled.** Legacy accepts an explicit `base_url` config override
  that bypasses `domain` entirely (`agilecrmBaseURL`, `agilecrm.go:276-298`), used by legacy's own
  tests to point at an `httptest.Server`. This bundle's `streams.json` templates `base.url` directly
  from `{{ config.domain }}` (see Auth setup) — there is no second template form to express an
  alternate raw-URL path, so only the domain-derived form is modeled. This bundle's own
  fixture/conformance harness instead points `domain` at a value the replay server matches via
  `Host`-independent fixture request matching (method+path+query only, not host), so the override
  path is not needed for this bundle's own tests either.
- **`owner_id`'s legacy fallback branch is not modeled.** Legacy's `ownerID` helper falls back to a
  top-level `item["owner_id"]` field when the nested `owner` object is absent
  (`agilecrm.go:170-178`). The engine's `computed_fields` dialect has no fallback/coalesce
  mechanism — a single template resolves against exactly one record path, and an absent path is
  silently skipped for that record (conventions.md §3). This bundle's `owner_id` computed field
  therefore only reproduces legacy's PRIMARY branch (`record.owner.id`); a record whose real wire
  shape omits the nested `owner` object entirely and relies on the top-level fallback would emit no
  `owner_id` at all here, versus legacy's populated fallback value. Documented as an accepted,
  narrower parity approximation per conventions.md §5 — not a data-changing behavior for any record
  that carries the (documented, primary) nested `owner` shape.
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size` (1-100,
  default 50) and `max_pages` (0/all/unlimited or a positive integer cap) as config-driven
  overrides (`agilecrmPageSize`/`agilecrmMaxPages`). The engine's `cursor` (`last_record_field`)
  paginator has no config-driven page-size or request-count-cap knob at all (mirrors stripe's
  resolved ledger item 3 and this wave's own adobe-commerce-magento precedent). `page_size`/
  `max_pages` are therefore not declared in `spec.json`; this bundle sends AgileCRM's own default
  (`page_size=50`) as a static per-stream query literal.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (reached only
  when `config.mode == "fixture"`) stamps deterministic synthetic records including a `connector`
  marker, a `fixture: true` flag, and a per-stream synthetic shape unrelated to the live wire schema
  (`agilecrm.go:197-233`). This is a credential-free conformance-harness affordance with no
  live-path equivalent; this bundle's schemas and fixtures target the live record shape only, and
  the engine's own `internal/connectors/conformance` fixture-replay harness provides the
  credential-free test affordance this bundle needs.
