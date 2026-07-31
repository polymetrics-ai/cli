# Mailchimp parity wave03-r1 TDD ledger

## Red evidence

Captured before production connector edits in `traces/mailchimp-official-audit.{json,md}`.

```bash
node <<'NODE' # fetch official Swagger root + 181 path refs, compare current api_surface
# summarized output:
# operations_total=298 path_count=181 ref_count=181 current_rows=9 missing=290 extra=1
# by_method: GET=149 POST=75 PATCH=32 DELETE=35 PUT=7
NODE
```

Result: fail/red. The current Mailchimp `api_surface.json` had only 9 rows against 298 official operations, with 290 official operations missing and 1 stale row not present in the official path set.

## Green evidence

Fixture-backed implementation generated from the official Mailchimp Marketing Swagger root and 181 provider-owned path refs.

Final counts from `internal/connectors/defs/mailchimp`:

- `api_surface.json` operations: 298 total.
- Implemented/covered rows: 295 total = 79 ETL stream rows + 68 typed direct-read rows + 148 reverse-ETL write rows.
- Blocked/local-workflow rows: 3 = `GET /`, `GET /ping`, and `POST /batches`.
- Excluded/N/A rows: 0.
- Fixture coverage: 79 stream fixture directories, 148 write request-shape fixtures, plus `fixtures/check.json`.

Green fixture gates:

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/mailchimp
# connectorgen validate: 1 connector(s) checked, 0 findings

go test ./internal/connectors/conformance -run 'TestConformance/mailchimp' -count=1
# ok   polymetrics.ai/internal/connectors/conformance  3.423s

go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1
# ok   polymetrics.ai/internal/cli  21.456s
```

## Refactor evidence

Verification found two generic scalability/tooling gaps after the expanded bundle landed:

- `connectorgen validate internal/connectors/defs/mailchimp` originally treated `fixtures/` and `schemas/` as sibling bundles. `cmd/connectorgen` now detects a single bundle path by the presence of `metadata.json`, validates it through the existing bundle validator, and retains the historical multi-bundle root behavior. Added `TestValidate_AcceptsSingleBundlePath`.
- Repeated CLI/certify tests repeatedly decoded the immutable embedded connector bundle set. `internal/connectors/bundleregistry.New` now caches the parsed embedded bundles with `sync.Once` while still returning a fresh mutable registry per call. Focused certify batch test improved from ~68s package time to ~4s package time, and the full required CLI regex gate dropped to ~21s.

Full local gate evidence is recorded in `VERIFICATION.md` and trace logs under `traces/`.
