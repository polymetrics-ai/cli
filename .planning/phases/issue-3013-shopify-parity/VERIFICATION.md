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
# Historical result below is superseded by the 2026-08-06 published-reference
# rebuild checklist; do not use the former 1,166-row aggregate as evidence.
```

Historical results before the current published-reference rebuild (not evidence for the regenerated bundle unless repeated below):

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
- Historical public-doc parity re-audit: superseded. The 2026-08-06 fresh Markdown inventory found 1,098 current published rows, not 1,166.
- Historical CLI parity smoke (pre-rebuild): `./pm help connectors`, `./pm connectors`, `./pm shopify --help`, and `./pm connectors inspect shopify --json` passed. The former `42 write actions` observation predates the 33-operation rebuild and is superseded; it is not current implementation evidence.
- Website parity grep: no textual website connector-count docs required updates; website connector pages use dynamic `pm connectors inspect ${connector.slug}` examples.

Not run: live Shopify provider calls, credentials, writes, certification, merges, or broad `go test ./...`/`make verify`.

## Resume host-restriction checklist (2026-08-06)

- [x] Red: `go test ./internal/connectors/defs/shopify -run TestShopDomainUsesCanonicalAdminHost -count=1` failed for `fixture-shop.invalid.example` before the pattern change.
- [x] Green (connector-local): the Shopify validation test now accepts `fixture-shop.myshopify.com` and rejects a non-`myshopify.com` host without echoing it; the app-level rejection row and canonical-host acceptance test are present.
- [ ] `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors` regenerates Shopify manual/skill documentation without unrelated doc drift.
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs` passes.
- [ ] `go test ./internal/connectors/conformance -run 'TestConformance/shopify' -count=1` passes.
- [ ] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`, `go vet ./internal/connectors/... ./internal/cli/...`, `go build ./cmd/pm`, `make connector-boundary`, and `git diff --check` pass or have a recorded blocker.

## Historical icon-generator evidence (2026-08-06; now owned by #3809)

`PM_ICON_REGISTRY_SOURCE` was not set in this worktree. Following the documented generator contract, the public Airbyte OSS registry was resolved from Airbyte's registry documentation and invoked without printing any credentials:

```bash
PM_ICON_REGISTRY_SOURCE='https://connectors.airbyte.com/files/registries/v0/oss_registry.json' make icons-generate
```

It stopped before writing generated output:

```text
iconregistrygen: ambiguous source/destination icon collapse for "customer-io": conflicting source URLs
```

The public source currently supplies distinct source and destination `customer-io` asset URLs that canonicalize to the same bare connector key. No registry JSON or icon output was hand-edited. Because Shopify still lacks its required generated registry row, app initialization panics at icon-coverage validation and both the app-level credential-boundary test and generated-doc command remain blocked. #3809 now owns the shared generator repair; this lane makes no further generator invocation, workaround, or registry edit.

## Published-reference rebuild checklist (2026-08-06)

- [x] Red/preflight: fetch Shopify's current Admin GraphQL full-index Markdown and the 67 Admin REST latest resource-page Markdown artifacts from the public sitemap. The extraction yields 287 queries, 518 mutations, 152 GET, 73 POST, 35 PUT, and 33 DELETE rows: 1,098 total. The AccessScope resource page already includes the access-scope GET endpoint and is counted once.
- [x] Rechecked the public source set after the rebuild without credentials: the current GraphQL full index still exposes 287 query and 518 mutation links, the sitemap still resolves 67 REST resource pages, and each of the 67 recorded REST Markdown artifacts returned HTTP 200.
- [x] Rebuilt `api_surface.json`, `source_inventory.json`, `operations.json`, and `cli_surface.json` from those artifacts: 1,098 rows, one citation and `2026-08-06` retrieval date per row, and no stale/duplicate AccessScope count.
- [x] Validated 33 current typed destructive REST DELETE declarations, including inventory-level identifiers, through connector-local metadata/fixture tests. Each has `confirmation.kind: "destructive"`, `mutation_class: "destructive"`, `batchable: false`, and a 1 MiB bound.
- [x] Restored all 136 still-current GraphQL mutation rows formerly classified as destructive to explicit `destructive_action` ledger rows with per-row `typed_destructive_confirmation` requirements and planned static commands. They remain blocked until individual fixed GraphQL document and typed-input contracts are available; no generic or redacting fallback was declared.
- [x] Hardened connector-local provenance regression coverage: every static command, blocked `api_surface` operation, and typed destructive operation must match the source inventory's exact per-row citation. GraphQL query/mutation citations must be canonical operation pages, and REST citations must point to an anchored section of the retrieved resource artifact.
- [x] Focused connector-local gates passed: `go test ./internal/connectors/defs/shopify -count=1`; `go run ./cmd/connectorgen validate internal/connectors/defs` (`551 connector(s) checked, 0 findings`); `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check`; `go test ./internal/connectors/conformance -run 'TestConformance/shopify' -count=1`; `go vet ./internal/connectors/defs/shopify`; and `git diff --check`.
- [x] Attempted the icon-dependent global preflight once after connector-local completion. `go test ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight -count=1` panicked at `internal/connectors/icons.go:250` because Shopify has no generated icon entry. Per the user direction, #3809 owns that shared generator/registry repair; this lane did not rerun `icons-generate` or hand-edit registry output.

## Remaining shared dependencies

- #3852 owns `internal/connectors/engine/schema/cli_surface.schema.json`, which excludes the captain-required non-redacting `json` direct-write `output_policy` even though the declared `rest_write` executor and command runner support it. The 33 fixed typed delete commands are therefore deliberately `planned` with exact static `source_cli_path` mappings. Promoting them to `implemented` requires #3852 to admit `json` and then running `connectorgen surface-sync`; no shared schema/runtime file was edited and no redacting fallback was declared in this lane.
- #3809 must make `make icons-generate` succeed and produce Shopify's explicit registry row before app initialization, generated manuals/catalogs, and the global command-runner preflight can load the bundle.
