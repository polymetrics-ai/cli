# Summary — Stripe connector parity (#3005)

## What changed

- Appended the captain-policy addendum via `gh-axi` to #3005 and #3006-#3012. Each issue now states that DELETE/destructive operations are in scope when implemented with typed destructive confirmation and the plan → preview → explicit approval → execute path, without changing or fabricating counts.
- Refreshed `internal/connectors/defs/stripe/api_surface.json` against official Stripe OpenAPI spec3 `2026-07-29.dahlia`.
  - 589 official HTTP operations represented exactly once.
  - Method counts: `GET=263`, `POST=294`, `DELETE=32`.
  - Official lane counts: `etl_read=242`, `reverse_etl_write=316`, `direct_read_query_search=9`, `binary_file=7`, `cdc_changefeed=7`, `excluded_not_applicable=8`.
  - Existing executable coverage is 8 rows: 5 streams, `create_customer`, `update_customer`, and new `delete_customer`.
  - Remaining 581 rows are blocked/planned operation metadata, not silent gaps.
- Added fixture-backed typed destructive `delete_customer` write action with `confirm: "destructive"` and idempotent 404 handling.
- Hardened customer create/update/delete schemas to reject additional fields, empty effective mutable fields, and blank/non-`cus_...` customer IDs before planning or execution.
- Restored Stripe `base_url` as a defaulted config override and updated shared write-default materialization so previews/execution preserve test/proxy overrides.
- Added Stripe-owned `cli_surface.json` for implemented stream commands and customer create/update/delete plan commands.
- Regenerated Stripe connector manual/skill docs and CLI golden transcripts affected by the new Stripe provider command surface.
- Updated Stripe docs to record exact blocked dependencies for complex form writes, provider search/query (#2985), CDC (#2986/#2988), and binary/file surfaces.

## Safety

- No live Stripe credentials, provider calls, provider writes, live certification, VPS/Thaalam work, merges, or new dependencies.
- Shared engine write-default materialization was updated; no CLI Go files edited.
- Unimplemented official operations remain truthfully blocked/planned or fixture-only/uncertified.

## Verification

Targeted credential-free gates passed:

- Official operation inventory comparison: 589/589 rows, missing 0, extra 0, duplicates 0.
- `go run ./cmd/connectorgen validate internal/connectors/defs` — 548 connectors, 0 findings.
- Stripe-only temp defs-root validation — 1 connector, 0 findings.
- `go test ./internal/connectors/conformance -run 'TestConformance/stripe' -count=1`.
- `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`.
- `go vet ./internal/connectors/... ./internal/cli/...`.
- `go build ./cmd/pm`.
- `make connector-boundary`.
- `git diff --check`.
- `pm help stripe`, `pm stripe`, `pm stripe customers --help`, `pm stripe customers delete --help`.

Full `make verify` was not run in this worker turn; no-mistakes/final pipeline should own the full gate.

