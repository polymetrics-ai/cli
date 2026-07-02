# Overview

ActiveCampaign is a wave2 fan-out declarative-HTTP migration. It reads ActiveCampaign contacts,
lists, deals, and campaigns through the ActiveCampaign v3 REST API. This bundle targets capability
parity with `internal/connectors/activecampaign` (the hand-written connector it migrates); the
legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide an ActiveCampaign API key via the `api_key` secret; it is sent as the `Api-Token` request
header (`api_key_header` auth mode, no value prefix) and is never logged, matching legacy's
`connsdk.APIKeyHeader(acAuthHeader, secret, "")` (`activecampaign.go:250`).

Provide `base_url` directly (e.g. `https://youraccount.api-us1.com/api/3`) — see Known limits for
why this bundle requires it rather than deriving it from an account subdomain the way legacy does.

## Streams notes

All 4 streams (`contacts`, `lists`, `deals`, `campaigns`) share the same shape: `GET` against the
ActiveCampaign v3 list endpoint, records wrapped under a resource-named top-level key
(`{"contacts": [...]}`, etc. — matches legacy's per-endpoint `recordsKey`), primary key `["id"]`.
None of the streams declare an `incremental` block: legacy's own `InitialState` always seeds an
empty cursor and its `harvest` loop has no server-side incremental filter parameter — full refresh
only, matching legacy exactly. Each schema names its own natural soft cursor field
(`udate`/`cdate`/`mdate`) per legacy's `streams.go` catalog, for downstream informational use only.

Pagination follows ActiveCampaign's limit/offset convention (`pagination.type: offset_limit`,
`limit_param: limit`, `offset_param: offset`, `page_size: 20` — legacy's own `acDefaultPageSize`).
A page shorter than 20 records signals the end, matching legacy's `len(records) < pageSize` stop
condition (`activecampaign.go:179`) exactly — the engine's `OffsetPaginator.Next` uses the
identical `recordCount < PageSize` rule.

## Write actions & risks

None. ActiveCampaign is read-only both in legacy and here: legacy's own package doc/`Write` method
returns `connectors.ErrUnsupportedOperation` unconditionally. `capabilities.write` is `false` and
this bundle ships no `writes.json`.

## Known limits

- **`base_url` is required, not derived from `account_username`.** Legacy derives its base URL as
  `https://<account_username>.api-us1.com/api/3` when `base_url` is unset
  (`activecampaign.go:266-277`, with a conservative subdomain-safe charset validator on
  `account_username`). The engine's `spec.json` `"default"` materialization mechanism only fills in
  a FIXED literal default, not a value computed from another config key — there is no
  computed/derived-default primitive in the dialect (conventions.md §3's `spec.json` `"default"`
  paragraph explicitly calls this out: "For a DERIVED default... this mechanism alone is not
  enough; either require `base_url` and drop the derivation... or express the derivation as a
  computed_fields-style template if/when the dialect grows one — do not invent ad hoc Go for it").
  This bundle takes the documented option: `base_url` is `required` in `spec.json`, and
  `account_username`-based derivation is dropped. This is a config-surface narrowing (an operator
  must now supply the full URL instead of just the account subdomain), not a data/behavior change
  for any request this bundle actually issues.
- Full ActiveCampaign v3 API surface (automations, tags, custom fields, deal tasks/notes, contact
  writes) is out of scope for wave2; see `api_surface.json`'s `excluded: {category: out_of_scope,
  reason: "Pass B capability expansion"}` entries. Only the 4 legacy-parity read streams are
  implemented.
- Legacy's fixture-mode-only fields (`activecampaign.go`'s `readFixture`, e.g. `previous_cursor`
  echoing a prior sync cursor) are not modeled — they only ever appeared in legacy's own
  credential-free fixture path, never in a live record; this bundle's schemas target the live
  record shape only, and the engine's own conformance/fixture-replay harness is the credential-free
  test affordance for this bundle.
