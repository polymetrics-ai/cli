# Overview

KYVE exposes public Cosmos-style REST query endpoints for its network's pools, stakers, funders,
and staking validators. This bundle is a pure Tier-1 declarative migration of
`internal/connectors/kyve` (the hand-written legacy connector): legacy is a thin
`connsdk.Requester`-based GET+paginate+map connector with no signature auth, no async
jobs, no compound writes, and no protocol beyond plain HTTP+JSON — every behavior it implements
(base-URL validation, cursor pagination via `pagination.next_key`, page-size/max-pages config, and
per-stream field projection) maps directly onto `streams.json`'s declarative dialect, so no
`hooks/kyve/` package or native component split is warranted. The legacy package stays registered
and unchanged until wave6's registry flip; the catalog inventory's `runtime_kind: "native_go"`
label reflects only that legacy happens to be a hand-written Go package, not that it needs one —
reading `internal/connectors/kyve/kyve.go` shows every code path is declarative-HTTP-shaped.

## Auth setup

None. KYVE's REST query endpoints are public and require no credentials, matching legacy exactly
(`base.auth: [{"mode": "none"}]`). `base_url` defaults to the public Korellia network endpoint
(`https://api.korellia.kyve.network`, legacy's own `defaultBaseURL`) and may be overridden for a
different KYVE network or a test proxy.

## Streams notes

All 4 streams share KYVE's `pagination.key`/`pagination.next_key` cursor-token pagination
convention (`pagination.type: cursor` with `token_path: "pagination.next_key"`,
`cursor_param: "pagination.key"`): the next page's `pagination.key` request parameter is read
verbatim from the previous response's `pagination.next_key` body field, and pagination stops when
that field is empty — matching legacy `harvest`'s exact `key == ""` stop condition. No `stop_path`
is declared since legacy has no secondary boolean stop signal (unlike Stripe's `has_more`). Every
request also sends `pagination.limit` from the `page_size` config value (default 100, legacy's
`defaultPageSize`), matching legacy's `boundedInt` bound of 1-1000 (`maxPageSize`) — the bound
itself is documentation only here (the engine dialect has no config-value numeric-range validator);
see Known limits.

- `pools` (`GET /kyve/query/v1beta1/pools`, records at `pools`): emits `id`/`name`/`runtime`
  directly from the top-level record, matching legacy `poolRecord` exactly (`item["id"]`,
  `item["name"]`, `item["runtime"]`).
- `stakers` (`GET /kyve/query/v1beta1/stakers`, records at `stakers`) and `funders`
  (`GET /kyve/query/v1beta1/funders`, records at `funders`) both use legacy's `accountRecord`
  shape: `address`/`amount` are read from the primary raw keys with legacy's defensive fallbacks
  (`account`, `balance`) via `coalesce`.
- `validators` (`GET /cosmos/staking/v1beta1/validators`, records at `validators`): emits
  `operator_address`/`status` directly and `moniker` from the nested `description.moniker` path,
  matching legacy `validatorRecord` exactly (`nested(item, "description", "moniker")`).

## Write actions & risks

None. KYVE's REST query surface is read-only from an off-chain client's perspective (state changes
are on-chain transactions, not something this connector or its legacy predecessor ever wrote);
`capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's
`Write` stub (`connectors.ErrUnsupportedOperation`).

## Known limits

- **`stakers` endpoint is not implemented on the live public Korellia network** (verified
  2026-07-03: `GET /kyve/query/v1beta1/stakers` returns a gRPC-gateway `{"code":12,"message":"Not
  Implemented"}` body, not a 404). This is a live-network-state fact, not a bundle defect — the
  stream is declared and fixture-tested identically to the other 3 (conformance and parity run
  entirely against the fixture, never the live network), and will start returning real data if/when
  the network re-enables the endpoint, with no bundle change required.
- `page_size`'s legacy-enforced numeric bound (1-1000) and `max_pages`'s `all`/`unlimited`
  string-literal acceptance are not separately validated by the engine dialect (no numeric-range or
  enum validator on a `spec.json` string property) — an out-of-range value is sent to the API
  verbatim rather than rejected client-side as legacy's `boundedInt`/`maxPages` would. This changes
  only the client-side error surface for an already-invalid config value, never accepted-input
  behavior for any value legacy itself would accept.
