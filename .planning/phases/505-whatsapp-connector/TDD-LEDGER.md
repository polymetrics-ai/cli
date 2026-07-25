# Phase 505 — TDD ledger

The connector-engine gates are the executable test bar for a `defs/` bundle: they fail before the
bundle exists/complies and pass after. Each was driven red→green.

| # | Gate (test) | Red (before) | Green (after) |
|---|---|---|---|
| 1 | `go run ./cmd/connectorgen validate internal/connectors/defs` | `whatsapp: missing required file metadata.json` (bundle absent/incomplete) | `548 connector(s) checked, 0 findings` |
| 2 | `go test ./internal/connectors/conformance/ -run TestConformance/whatsapp` | `read_fixture_nonempty:*` 404 (fixtures did not match interpolated request paths) | `ok` (all static + dynamic replay checks pass) |
| 3 | `go test ./internal/connectors/bundleregistry/` | `bundle count = 548, want 547` (new connector auto-registers via `engine.LoadAll(defs.FS)`) | `ok` after count 547→548 |
| 4 | `go test ./internal/cli/` (`TestGoldenTranscripts`, `catalog_cli_test`) | catalog `"count": 551` and manual count text stale | `ok` after count 551→552 / 547→548 and golden regen |
| 5 | `make docs-check` (`pm docs validate`) | `connector catalog json has 551 entries, want 552` | `Validated connector docs in docs/connectors` |

## Iterative validator findings resolved (gate #1)

- `metadata.json` `rate_limit.strategy` — engine `RateLimitSpec` only accepts `requests_per_minute`
  (strict decode); reduced to `{requests_per_minute: 80}`.
- `operations.json` `stream_etl` op — loader requires `stream_etl`/`composite` kinds to carry a
  `composite` execution block; added `composite.steps` to `whatsapp.web_sync`.
- `cli_surface.json` analytics — direct-read operations mapping flags to `body.*` require a POST
  `rest_read` op; modeled the four analytics as typed POST read-queries (matches #513).
- `cli_surface.json` `config`-intent command — only `etl`/`reverse_etl`/`direct_read`/
  `local_workflow` are valid for `implemented`; dropped the standalone `auth status` command
  (auth surface covered by help topics + `pm connectors inspect`).

## Conformance fixture design (gate #2)

Replay synthesizes config as `synthetic-conformance-value` (secrets `synthetic-conformance-secret`,
`start_date` `2020-01-01T00:00:00Z`) and replaces the base URL. Stream fixtures therefore encode the
interpolated request path (`/synthetic-conformance-value/phone_numbers`) and the `limit` query the
stream sends; `phone_numbers` ships 2 pages to exercise `pagination_terminates`. Write fixtures are
omitted (all `write_request_shape:*` checks Skip), consistent with conformance policy; live sends /
media payloads are human-gated.
