# VERIFICATION — issue 3595 icon registry single-source foundation

## Required focused gates

```bash
gofmt -w cmd internal
go test ./internal/connectors ./internal/connectors/boundary ./cmd/iconregistrygen ./cmd/connectorgen
go test ./cmd/pm ./internal/cli
node --check website/scripts/gen-connector-bundles.mjs
node --check website/scripts/fetch-simple-icons.mjs
```

Add or adjust package-specific commands if implementation places tests elsewhere. Use existing package managers/tools only; do not add dependencies without approval.

Review repair focused gate:

```bash
go test ./internal/connectors ./internal/connectors/bundleregistry ./cmd/iconregistrygen && node --test website/scripts/icon-registry.test.mjs
```

Review repair round 2 focused gate:

```bash
node --test website/scripts/icon-registry.test.mjs && go test ./internal/cli -run TestValidateConnectorDocsRejectsStaleIconMetadata && go run ./cmd/pm docs validate --connectors-dir docs/connectors
```

Review repair round 3 focused gate:

```bash
go test ./internal/cli -run '^TestValidateConnectorDocsRejectsStaleIconMetadata$' && go run ./cmd/pm docs validate --connectors-dir docs/connectors
```

Review repair round 4 focused gate:

```bash
go test ./internal/connectors ./internal/cli ./cmd/iconregistrygen -run 'Test(ConnectorIconRegistryProjectsCompleteMetadata|ConnectorIconMetadataOmitsAbsentOptionalFields|ValidateConnectorDocsRejectsStaleIconMetadata|BuildIconEntriesPreservesCuratedAttribution)$' && go run ./cmd/pm docs validate --connectors-dir docs/connectors
```

Review repair round 5 focused gate:

```bash
node website/scripts/gen-connector-bundles.mjs
node website/scripts/gen-connector-catalog.mjs
node website/scripts/gen-connectors.mjs
# Assert the first run changes only derived icon values/deletes, rerun all three generators, and require byte-stable outputs.
go test ./internal/cli -run 'Test(GeneratedConnectorIconBlockRequiresExactUniqueHeading|ValidateConnectorDocsRejectsStaleIconMetadata)$'
node --test website/scripts/icon-registry.test.mjs
```

Review repair round 6 focused gate:

```bash
gofmt -w internal/cli/connector_docs.go internal/cli/connector_docs_test.go cmd/iconregistrygen/main.go cmd/iconregistrygen/main_test.go
go test ./internal/cli ./cmd/iconregistrygen -run 'Test(GeneratedConnectorIconBlockRequiresExactUniqueHeading|BuildIconEntriesRejectsDuplicateCuratedKeys|BuildIconEntriesRejectsSharedAssetPathSourceURLConflict|BuildIconEntriesAllowsSharedAssetPathWithIdenticalSourceURL)$'
node --test website/scripts/icon-registry.test.mjs
node website/scripts/gen-connector-bundles.mjs
# Hash bounded generated outputs, run the generator again, and require identical hashes plus a clean output diff.
```

Review repair round 7 (F15) focused gate:

```bash
go test ./cmd/iconregistrygen -v
python3 -c "import json; d=json.load(open('internal/connectors/icon_data.json')); print(len([e for e in d if not e.get('connector')]))"
```

Review repair round 8 (CodeQL fetch-simple-icons input validation) focused gate:

```bash
cd website
node --test scripts/icon-registry.test.mjs
node --check scripts/fetch-simple-icons.mjs
node --check scripts/lib/simple-icons.mjs
node --check scripts/lib/connector-icons.mjs
pnpm run lint
pnpm run typecheck
pnpm run gen:website-data   # must diff clean
```

Review repair round 9 (fetch-boundary shape guard, containment dedupe, icon-coverage memoization) focused gate:

```bash
gofmt -l internal/connectors
go test ./internal/connectors -run 'TestMustValidateIconCoverageRevalidatesAfterRegistration|TestRegistryIconCoverage'
node --test website/scripts/icon-registry.test.mjs
node --check website/scripts/fetch-simple-icons.mjs
node --check website/scripts/gen-connector-bundles.mjs
node --check website/scripts/lib/simple-icons.mjs
```

Review repair round 10 (CodeQL alert #93 fetched-content checksum pinning) focused gate:

```bash
cd website
node --test scripts/icon-registry.test.mjs
node --check scripts/fetch-simple-icons.mjs
node --check scripts/lib/simple-icons.mjs
pnpm run lint
pnpm run typecheck
pnpm run gen:website-data   # must diff clean
node scripts/fetch-simple-icons.mjs --update-lockfile   # real network; regenerates website/data/simple-icons.lock.json
node scripts/fetch-simple-icons.mjs                      # real network; verifies against the lockfile just written
git status --short                                       # only the lockfile + code changes; no docs/website SVG diff
```

## Repository gates before integration

```bash
go vet ./...
go test ./...
go build ./cmd/pm
make connector-boundary
make verify
```

If a gate is not applicable or blocked by environment, record the exact reason and do not claim it passed.

## GitHub / no-mistakes gates

- PR targets `fix/3579-connector-path-ownership-guardrails` and uses `Refs #3595` and `Refs #3579`.
- Required/current GitHub checks green before parent integration.
- Comprehensive native-Codex `gpt-5.6-sol` no-mistakes validation at `xhigh`, including full-diff comprehensive review/rereview of all material substantiated issues.
- Do not integrate PR #3590 from the prior no-mistakes run; #3590 needs fresh 5.6 SOL validation after this foundation lands and is reconciled.

## Current evidence

- `scripts/gsd doctor`: pass in `/Users/karthiksivadas/.treehouse/cli-83d592/5/worker-3595-icon-registry`.
- `scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1/workers/issue-3595 --dry-run`: failed with `unknown GSD command: programming-loop`; manual GSD fallback uses `.pi/prompts/pm-gsd-loop.md` and must be recorded in PR evidence.
- Pre-edit audit/proof commands completed without credentialed checks.
- RED `go test ./internal/connectors ./cmd/iconregistrygen`: failed before implementation on missing exact bare lookup/ownership/generator-collision support.
- RED `node --test website/scripts/icon-registry.test.mjs`: failed before implementation on prefixed registry keys, website override authority, and website script prefix handling.
- GREEN `go test ./internal/connectors ./cmd/iconregistrygen`: pass.
- GREEN `go test ./internal/connectors ./internal/connectors/boundary ./cmd/iconregistrygen ./cmd/connectorgen`: pass.
- GREEN `node --test website/scripts/icon-registry.test.mjs`; `node --check website/scripts/gen-connector-bundles.mjs`; `node --check website/scripts/fetch-simple-icons.mjs`: pass.
- GREEN `go test ./internal/cli ./cmd/pm`: pass (`internal/cli` took 365.247s).
- GREEN `go vet ./...`: pass.
- GREEN `go test ./...`: pass.
- GREEN `go build ./cmd/pm`: pass.
- GREEN `make connector-boundary`: clean boundary report.
- GREEN `make verify`: pass, including docs validation, smoke, lint, connectorgen validate, connector boundary, and Homebrew notification checks.
- GREEN review round 6 focused Go gate: `internal/cli` and `cmd/iconregistrygen` targeted F11/F13/F14 regressions passed.
- GREEN review round 6 Node gate: all 6 icon-registry tests passed, including F12 invalid unimplemented-row coverage.
- GREEN review round 6 deterministic generation: two consecutive bundle generations emitted 550 connectors and 334 icons with identical data/public-icon hashes and no checked-in derived-output diff.
- RED `go test ./cmd/iconregistrygen -run TestLoadCuratedIconEntries`: failed before the F15 fix (empty and `source-`/`destination-`-prefixed curated keys were silently accepted with no error).
- GREEN review round 7 (F15) focused Go gate: `go test ./cmd/iconregistrygen -v` — all 16 tests pass, including empty/prefixed curated-key rejection, bare curated-key preservation, and continued raw-upstream prefix collapse.
- GREEN F15 no-regression audit: the committed `internal/connectors/icon_data.json` (554 entries) contains zero empty/prefixed curated connector keys, so the default `--curated`-equals-`--out` regeneration path is unaffected by the stricter check.
- GREEN `Website checks` root cause: PR #3596's failing CI run targeted stale commit `10829e569`, predating the six pipeline review-fix commits; reproduced at pipeline head `8faa94cc6` with CI-matching pnpm 11.7.0 — `pnpm run gen:website-data` diffed clean, and `lint`/`typecheck`/`test:unit` (76/76)/`build` all passed locally.
- Review round 8 (no-mistakes run `01KZ2661QR8MTV33B5WDB7S1HV`) fixed all 8 review findings plus 1 follow-up CI path-filter finding; document gate approved as-is (doc-migration-dir-consolidation deferred to `cli-docs-migration-dir-consolidation-r1`); PR #3596 reached 27 passed / 0 failed / 8 skipped, `mergeable_state: clean`, before the two CodeQL threads below were raised.
- RED `node --test scripts/icon-registry.test.mjs` (from `website/`): failed before the CodeQL input-validation fix — `website/scripts/lib/simple-icons.mjs` did not exist.
- GREEN CodeQL fetch-simple-icons input validation gate: `node --test scripts/icon-registry.test.mjs` (8/8, equivalently `pnpm run test:scripts`) — `../`-traversal path rejected, nested `../` path rejected, absolute path rejected, slug containing `/` rejected before any fetch, slug containing a scheme rejected before any fetch, empty slug rejected, valid bare slug + in-tree path resolves unchanged.
- GREEN `pnpm run lint`, `pnpm run typecheck`, `pnpm run gen:website-data` (zero diff): pass after the CodeQL fix, confirming no regression to generated website data.
- Review round 9 (PR #3596 review commit `4b182b1ba`) hardened the round-8 fetch boundary so an in-tree non-icon or non-string registry path is rejected as well as an escaping one, deduped `assertInside` into `website/scripts/lib/connector-icons.mjs`, and memoized `Registry.MustValidateIconCoverage` with invalidation on `Register`. Its focused gate commands are listed above; the gate results are owned by the enclosing no-mistakes validation run for PR #3596 and are not restated here.
- CodeQL alert #93 traced to a distinct dataflow (fetched SVG response body → `writeFileSync`) that no prior round validated; decision key `pr3596-codeql-93-svg-content-checksum-20260803` required checksum-pinning keyed by connector.
- RED `node --test scripts/icon-registry.test.mjs` (from `website/`): failed before implementation — `readSimpleIconsLockfile`/`sha256Hex`/`verifyFetchedIconDigest`/`writeSimpleIconsLockfile` did not exist.
- GREEN review round 10 (CodeQL #93 checksum pinning) focused Node gate: 14/14, including matching/tampered/missing-entry digest verification, independent per-connector verification for two connectors sharing one digest, and sorted lockfile round-trip persistence.
- GREEN real end-to-end regeneration against the live Simple Icons CDN: `node scripts/fetch-simple-icons.mjs --update-lockfile` wrote 61 real connector entries to `website/data/simple-icons.lock.json`, confirming duplicate digests for `ebay-finance`/`ebay-fulfillment` and all 8 `zoho-*` connectors, not deduped; a second real run in default verify mode passed 61/61.
- GREEN real end-to-end tamper proof: corrupted one real lockfile entry's `sha256`, reran the real script — exited 1 naming the connector plus expected/received digests, and the real on-disk SVG (verified via MD5 sentinel) was left unchanged.
- GREEN `pnpm run lint`, `pnpm run typecheck`, `pnpm run gen:website-data` (zero diff): pass after the fix. `git status --short` after the real regeneration shows only the new lockfile and the code diff — every re-fetched SVG under `docs/connectors/icons/simple-icons/**` and its website public copy is byte-identical to what was already committed.
