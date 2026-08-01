# VERIFICATION — issue-3013-shopify-parity

Focused gates run locally on branch `fm/cli-shopify-parity-wave01-r1`:

```bash
# focused Shopify validation through temp defs root
# connectorgen validate expects a parent directory containing connector dirs
tmp=$(mktemp -d)
cp -R internal/connectors/defs/shopify "$tmp/shopify"
go run ./cmd/connectorgen validate "$tmp"
rm -rf "$tmp"

# full definition validation
go run ./cmd/connectorgen validate internal/connectors/defs

go test ./internal/connectors/conformance -run 'TestConformance/shopify' -count=1
go run ./cmd/pm docs validate --connectors-dir docs/connectors
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1
go vet ./internal/connectors/... ./internal/cli/...
go build ./cmd/pm
make connector-boundary
git diff --check

# public documentation parity audit (no credentials/provider API calls)
# re-fetches Shopify Admin GraphQL full index and REST latest resource pages,
# then compares fresh operation keys to internal/connectors/defs/shopify/api_surface.json
# Result: officialDocsTotal=1166, ledgerTotal=1166, missing=0, stale=0,
# deleteRows=44, coveredDelete=42, blockedDelete=2, noncanonicalDeleteRows=0
```

Results:

- `no-mistakes doctor`: passed with existing daemon running.
- Focused temp-root validation: `connectorgen validate: 1 connector(s) checked, 0 findings`.
- Full defs validation: `connectorgen validate: 550 connector(s) checked, 0 findings`.
- Conformance: `ok polymetrics.ai/internal/connectors/conformance` for `TestConformance/shopify`.
- Connector docs validation: `Validated connector docs in docs/connectors`.
- CLI/docs/golden: `ok polymetrics.ai/internal/cli` for `Connector|Dynamic|Golden` (247.690s with an isolated temporary `GOCACHE` to avoid concurrent shared-cache races).
- `go vet ./internal/connectors/... ./internal/cli/...`: passed.
- `go build ./cmd/pm`: passed.
- `make connector-boundary`: clean, with only pre-existing documented exceptions.
- `git diff --check`: passed.
- Public-doc parity re-audit: GraphQL full index and 67 REST resource pages matched the ledger exactly: 1166 official rows, 1166 ledger rows, 0 missing, 0 stale, 0 non-canonical DELETE paths.
- CLI parity smoke: `./pm help connectors`, `./pm connectors`, `./pm shopify --help`, and `./pm connectors inspect shopify --json` passed. Shopify help exposes `ledger`, `shop`, `delete`, and `graphql` groups; inspect shows stream `shop` and 42 write actions.
- Website parity grep: no textual website connector-count docs required updates; website connector pages use dynamic `pm connectors inspect ${connector.slug}` examples.

Not run: live Shopify provider calls, credentials, writes, certification, merges, or broad `go test ./...`/`make verify`.
