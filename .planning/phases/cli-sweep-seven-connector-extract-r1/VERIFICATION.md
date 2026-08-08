# Verification checklist — seven connector extraction r1

## Bundle and generated-surface checks

- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs` succeeds.
- [ ] `go run ./cmd/connectorgen surface-sync --check` succeeds after regeneration.
- [ ] The source-derived counts match the seven-row table in `PLAN.md`.
- [ ] `operation_endpoint_ledger.json` delta names only the seven connectors.
- [ ] `git diff --name-only` passes the scope fence in `PLAN.md`.

## Tests and binary checks

- [ ] Focused seven connector `cmd/connectorgen` tests pass.
- [ ] `go test -timeout 20m ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight` passes.
- [ ] `go build -o /tmp/pm-cli-sweep-seven ./cmd/pm` succeeds.
- [ ] Every implemented command in the seven generated CLI surfaces routes to its own real-binary
  help `NAME` line; totals are 911, 584, 139, 127, 100, 63, and 60.
- [ ] `pm help <connector>` and bare `pm <connector>` succeed for all seven.

## Docs and website checks

- [ ] `pm docs generate --dir docs/cli --connectors-dir docs/connectors` regenerated connector docs.
- [ ] `pm docs validate --dir docs/cli --connectors-dir docs/connectors --website-dir website/content/docs` passes.
- [ ] `cd website && pnpm run gen:website-data` regenerated website data.
- [ ] Generated website data reflects all seven connectors and no excluded connector regression.

## Final local gates

- [ ] Focused affected package tests, `go vet` for affected packages, `make tidy-check`, `make lint`,
  `make agent-contract-check`, `make connector-boundary`, `make release-workflow-check`, and
  `scripts/verify-gsd-workflow origin/main` pass as applicable.
- [ ] GSD `verify-work` and `code-review` prompts were generated and executed inline; findings are
  recorded and resolved or explicitly handed off.

## PR handoff requirement

State verbatim that workday-rest, jira, help-scout, greenhouse, chatwoot, gmail, and lever-hiring
are implemented but **not certified**, have **never been exercised against their live services**,
and no credentials were held or used for them.

