# Verification checklist — seven connector extraction r1

## Bundle and generated-surface checks

- [x] Focused plural-write validator tests recorded a red state then passed after the
  captain-authorized `covered_by.writes` engine/schema/validator foundation was added.
- [x] The post-foundation all-bundle validator checked all 551 connectors with zero findings,
  proving no existing connector behavior changed.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` checked 551 connectors with 0 findings.
- [x] `go run ./cmd/connectorgen surface-sync --check` scanned 551 connectors with no drift.
- [x] The source-derived counts match the seven-row table in `PLAN.md`.
- [x] `operation_endpoint_ledger.json` delta is confined to chatwoot, jira, lever-hiring, and
  workday-rest, a subset of the seven-connector allowlist.
- [x] Final staged-diff scope audit found 162 branch-delta paths, 0 unexpected paths, and 0
  github/zendesk-support paths.

## Tests and binary checks

- [x] Focused seven connector `cmd/connectorgen` tests pass.
- [x] `go test -timeout 20m ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight` passes.
- [x] `go build -o /tmp/pm-cli-sweep-seven ./cmd/pm` succeeds.
- [x] Every implemented command in the seven generated CLI surfaces routes to its own real-binary
  help `NAME` line; totals are 911, 584, 139, 127, 100, 63, and 60 (1,984 total).
- [x] `pm help <connector>` and bare `pm <connector>` succeed for all seven.

## Docs and website checks

- [x] `pm docs generate --dir docs/cli --connectors-dir docs/connectors` regenerated connector docs.
- [x] `pm docs validate --dir docs/cli --connectors-dir docs/connectors --website-dir website/content/docs` passes.
- [x] `cd website && pnpm run gen:website-data` regenerated website data.
- [x] Generated website data reflects all seven connector counts; no excluded connector input changed.

## Post-review corrections and re-verification

Inline review found three classes of defect this checklist had marked verified. See TDD-LEDGER.md
"Red — inline review findings against the imported tree" for the reproduced failures.

- [x] **The verified package set was incomplete, and that is why the defects shipped.**
  `internal/connectors/conformance` and `internal/connectors/certify` both consume the api_surface
  coverage rule but were absent from the gate list below, so a `covered_by.writes` migration that
  reached only one of four consumers passed every gate this phase ran. Both packages are now gated.
- [x] Plural-only coverage is regression-tested rather than incidentally covered:
  `certify.TestSurfaceInventoryCountsPluralOnlyWriteCoverage` (classification **and** write counts
  292/252), `certify.TestSurfaceInventoryPluralOnlyBundlesUseNoSingularWrite`, and
  `connectorgen.TestBatchMaterializePluralOnlyWriteCoverage` (end-to-end `batch materialize`,
  including a two-element `writes` array over one endpoint).
- [x] Twelve help-scout stream paths carried unbound `{name}` placeholders that no gate could see.
  Paths now interpolate declared config keys, and `engine.ResolveCheckRequestPath` enforces the
  invariant for both `connectorgen validate` and `conformance`; `writes.json` stays exempt because
  `path_fields` legitimately binds 165 shipped write paths.
- [x] help-scout, jira and workday-rest published read-only risk/description/docs text while
  shipping 65, 292 and 252 write actions. Corrected and regenerated through docs/website generators.
- [x] Re-run on the post-fix tree: `internal/connectors/engine`, `internal/connectors/conformance`,
  `internal/connectors/certify` and `cmd/connectorgen` all pass;
  `connectorgen validate` reports 551 connectors / 0 findings; `surface-sync --check` reports no
  drift; `TestEveryImplementedCommandPassesRuntimePreflight` passes; `pm docs validate` passes.
- [ ] The 1,984-command real-binary NAME sweep was **not** re-run after these corrections. No
  correction adds, removes or renames a command — the changes are stream path templates, spec keys,
  risk/description text and docs prose — and the seven implemented/documented counts are unchanged.

## Final local gates

- [x] Focused affected package tests, full `cmd/connectorgen`, `internal/connectors/engine`,
  `internal/connectors/conformance`, `internal/connectors/certify`,
  `internal/connectors/commandrunner`, and full `internal/cli` tests pass; affected-package `go vet`
  and `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`,
  `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`,
  `make connector-boundary`, and `make release-workflow-check` pass.
- [x] `scripts/verify-gsd-workflow origin/main` reports GSD/TDD evidence for every implementation
  file changed by this phase.
- [x] GSD `verify-work` and `code-review` prompts were generated and executed inline under the
  contract's no-role-spawning fallback; automated PR review remains a firstmate handoff.

## PR handoff requirement

State verbatim that workday-rest, jira, help-scout, greenhouse, chatwoot, gmail, and lever-hiring
are implemented but **not certified**, have **never been exercised against their live services**,
and no credentials were held or used for them.
