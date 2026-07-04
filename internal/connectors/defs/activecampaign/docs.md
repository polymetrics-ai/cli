# Overview

ActiveCampaign reads ActiveCampaign contacts, lists, deals, campaigns, tags, automations, custom
fields, accounts, users, deal stages, and deal tasks through the ActiveCampaign v3 REST API. This
bundle originally targeted capability parity with `internal/connectors/activecampaign` (the
hand-written connector it migrates; the legacy package stays registered and unchanged until wave6's
registry flip) and was expanded in Pass B to cover the full documented core-CRM read surface
(`api_surface.json`).

## Auth setup

Provide an ActiveCampaign API key via the `api_key` secret; it is sent as the `Api-Token` request
header (`api_key_header` auth mode, no value prefix) and is never logged, matching legacy's
`connsdk.APIKeyHeader(acAuthHeader, secret, "")` (`activecampaign.go:250`).

Provide `base_url` directly (e.g. `https://youraccount.api-us1.com/api/3`) — see Known limits for
why this bundle requires it rather than deriving it from an account subdomain the way legacy does.

## Streams notes

10 of the 11 streams (`contacts`, `lists`, `deals`, `campaigns`, `tags`, `automations`, `fields`,
`accounts`, `deal_stages`, `deal_tasks`) share the same shape: `GET` against the ActiveCampaign v3
list endpoint, records wrapped under a resource-named top-level key (`{"contacts": [...]}`,
`{"tags": [...]}`, `{"dealStages": [...]}`, etc.), pagination follows ActiveCampaign's limit/offset
convention (`pagination.type: offset_limit`, `limit_param: limit`, `offset_param: offset`,
`page_size: 20` — legacy's own `acDefaultPageSize` for the original 4; the same convention holds for
every new Pass B stream too, confirmed from each resource's own reference page). A page shorter than
20 records signals the end, matching legacy's `len(records) < pageSize` stop condition
(`activecampaign.go:179`) exactly for the original streams, and the engine's `OffsetPaginator.Next`'s
identical `recordCount < PageSize` rule for the new ones.

`users` is the one exception: ActiveCampaign's own `/users` reference page documents no
`limit`/`offset` query parameters at all (a small, unpaginated admin-scoped list), so its stream
declares a `pagination: {"type": "none"}` override rather than inheriting the base offset/limit
spec.

None of the streams declare an `incremental` block: legacy's own `InitialState` always seeds an
empty cursor and its `harvest` loop has no server-side incremental filter parameter for the original
4 streams, and none of ActiveCampaign's list endpoints (old or newly added) document a server-side
`updated_since`-style filter parameter either — full refresh only, across every stream in this
bundle. Each schema names its own natural soft cursor field (`udate`/`cdate`/`mdate`/
`updatedTimestamp`) confirmed from that resource's own documented record shape, for downstream
informational use only; `users` declares no cursor field at all since ActiveCampaign's user records
carry no timestamp fields.

`deal_stages` (`GET /dealStages`) surfaces each pipeline stage's own record (including its parent
pipeline's id via the `group` field); the parent `dealGroups` (pipeline) container objects
themselves are not modeled as a separate stream (see Known limits).

## Write actions & risks

None; `capabilities.write` is `false` and this bundle ships no `writes.json`. This is unchanged from
the wave2 migration, but for a different reason now: legacy's own `Write` method unconditionally
returned `connectors.ErrUnsupportedOperation` because writes were out of scope for wave2, whereas
Pass B's full-surface review found a genuine `ENGINE_GAP` blocking every one of ActiveCampaign's
documented write endpoints. Every create/update endpoint across every resource in this API
(contacts, lists, deals, campaigns, tags, automations-adjacent fields, contactTags, accounts,
dealTasks, notes, fieldValues, webhooks) wraps its request body in a resource-named top-level JSON
envelope — confirmed directly from each endpoint's own reference-page example request body, e.g.:

```
POST /contacts   -> {"contact": {"email": "...", "firstName": "...", "fieldValues": [...]}}
POST /deals      -> {"deal": {"contact": "...", "title": "...", "fields": [...]}}
POST /tags       -> {"tag": {"tag": "...", "tagType": "..."}}
POST /contactTags -> {"contactTag": {"contact": "...", "tag": "..."}}
```

This dialect's write body construction (`docs/migration/conventions.md` §3, "Write body
construction") always sends every record field flat at the top level (`body_type: json`'s default
body = every record field except `path_fields`, or an explicit `body_fields` allow-list — neither
supports wrapping the whole payload under a single named key). Implementing any of these as a
declarative write action would either send the wrong wire shape (a flat body ActiveCampaign's API
would not recognize as the expected resource) or require inventing an undocumented dialect feature —
both forbidden by the parity-deviation meta-rule (§5). This matches the sanctioned precedent already
recorded for the identical envelope-body gap class in `veeqo`/`statuspage`/
`solarwinds-service-desk`'s docs.md. See `api_surface.json` for the per-endpoint ENGINE_GAP
citations.

## Known limits

- **All create/update/delete mutations are blocked by the request-body-envelope `ENGINE_GAP`** — see
  Write actions & risks above. This affects every resource in the API, not just the ones this bundle
  reads.
- **`base_url` is required, not derived from `account_username`.** Legacy derives its base URL as
  `https://<account_username>.api-us1.com/api/3` when `base_url` is unset
  (`activecampaign.go:266-277`, with a conservative subdomain-safe charset validator on
  `account_username`). The engine's `spec.json` `"default"` materialization mechanism only fills in
  a FIXED literal default, not a value computed from another config key — there is no
  computed/derived-default primitive in the dialect. This bundle takes the documented option:
  `base_url` is `required` in `spec.json`, and `account_username`-based derivation is dropped. This
  is a config-surface narrowing (an operator must now supply the full URL instead of just the
  account subdomain), not a data/behavior change for any request this bundle actually issues.
- `dealGroups` (pipeline container objects) is not modeled as a separate stream: `deal_stages`
  already surfaces each stage's parent pipeline id via its `group` field, and pipelines are a
  low-cardinality admin-configuration list rarely queried independently of their stages.
- Per-parent sub-resources with no top-level, unfiltered list endpoint of their own
  (contactAutomations, contactTags-by-contact, deal notes, fieldValues) are out of scope: covering
  them would require per-parent `fan_out` (one preliminary list request per contact/deal id) that
  this pass does not scope, matching the same breadth-vs-cost triage this migration wave applies
  elsewhere.
- Bulk/async CSV contact import (`POST /import/bulk_import` + status polling) is out of scope: a
  batch-file upload workflow, not a per-record declarative read or write.
- E-commerce sync data (`connections`, `ecomOrders`, `ecomOrderProducts`) requires a configured
  storefront connection and is out of scope; `segments` (saved-segment definitions) is a
  deeply-nested admin-configuration list deferred pending real demand.
- Legacy's fixture-mode-only fields (`activecampaign.go`'s `readFixture`, e.g. `previous_cursor`
  echoing a prior sync cursor) are not modeled — they only ever appeared in legacy's own
  credential-free fixture path, never in a live record; this bundle's schemas target the live
  record shape only, and the engine's own conformance/fixture-replay harness is the credential-free
  test affordance for this bundle.
