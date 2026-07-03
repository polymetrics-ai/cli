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
  shape: `address` and `amount` read directly off the raw record. Legacy's real mapper is
  `first(item, "address", "account")`/`first(item, "amount", "balance")` — a defensive two-key
  fallback the engine's `computed_fields` dialect cannot express (no coalesce/first-of filter
  exists). This bundle wires only the PRIMARY key of each pair (`address`, `amount`); see Known
  limits.
- `validators` (`GET /cosmos/staking/v1beta1/validators`, records at `validators`): emits
  `operator_address`/`status` directly and `moniker` from the nested `description.moniker` path,
  matching legacy `validatorRecord` exactly (`nested(item, "description", "moniker")`).

## Write actions & risks

None. KYVE's REST query surface is read-only from an off-chain client's perspective (state changes
are on-chain transactions, not something this connector or its legacy predecessor ever wrote);
`capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's
`Write` stub (`connectors.ErrUnsupportedOperation`).

## Known limits

- **`stakers`/`funders`' secondary field-name fallback is not modeled.** Legacy's `accountRecord`
  reads `address` OR `account` (whichever is present) and `amount` OR `balance` (whichever is
  present) per record. The engine's `computed_fields` dialect resolves a single fixed reference path
  per output field with no coalesce/first-of mechanism, so only the first (primary) key name of
  each pair is wired (`address`, `amount`). A live record whose ONLY populated key is the secondary
  name (`account` with no `address`, or `balance` with no `amount`) would silently emit `null` for
  that field here, whereas legacy would emit the secondary value. Live verification against the
  public Korellia `funders` endpoint (2026-07-03) shows real records use `address` (not `account`)
  as the top-level field and nest their real balance under `stats.total_used_funds[].amount` (not a
  flat `balance` field at all) — legacy's own `balance` fallback branch appears to already be dead
  against the current live wire shape, and neither key name is a nested-path lookup legacy itself
  ever performed, so this bundle intentionally does not attempt to model the nested balance shape
  either (out of scope; a `computed_fields` deviation, not a correctness regression relative to
  legacy's own — already partially stale — mapping). Documented here rather than filed as an
  `ENGINE_GAP` because the fallback is a defensive convenience, not a behavior any currently-passing
  legacy test exercises (`kyve_test.go` only exercises `pools`).
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
