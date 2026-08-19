Refs #4193

## Summary

Every documented legacy leaf now resolves help through one shared wrapper path
before opening project state, resolving credentials, validating required flags,
or executing a handler. The previous `legacyLeafManualTopic` two-command
switch is removed.

The shared path derives valid command prefixes from the wrapper manual—the same
source used for runtime help and generated `docs/cli` pages—so a new documented
leaf enters the behavior automatically. An unknown top-level or leaf command
still reports a usage error even if a help flag is present.

I deliberately did not broaden this into a full Cobra migration. Core commands
still use the legacy parser under Cobra's wrapper because flag parsing and
approval-carrier semantics are established there; changing that parser would
be materially wider than this defect. The shared pre-handler fixes the required
ordering with a smaller, testable change.

The one contextual resolver left is closed ETL transport help. Its selected
transport name and security contract are dynamic, so it remains an explicit
resolver field on the ETL wrapper rather than a router switch. It is not part
of the generic legacy leaf mechanism and keeps its dedicated manual.

## Tests

Happy paths:

- Built binary: 63 legacy paths × `--help`/`-h` × empty/initialized project =
  252 requests, all exit 0 with `NAME` output and no stderr.
- Built binary: 36 dynamic connector roots × the same four conditions = 144
  successful renders.
- Programmatic generated-surface sweep: 8,900 connector leaves × `--help`/`-h`
  = 17,800 pre-dispatch manual renders.

Bad paths:

- Unknown top-level and unknown leaf commands with `--help` return usage exit
  2 rather than a manual.
- A valued `--approval-token-stdin` remains a usage error and help does not
  mask it.

Edge paths:

- `-h`, help mixed with other flags, and missing required positionals all
  render their manual.
- The exhaustive legacy sweep runs from both an empty directory and a directory
  initialized with `pm init`.
- ETL transport retains its exact contextual manual.

## Verification

- `go test -timeout 20m ./internal/cli -count=1` — PASS (513.833s)
- `go test -timeout 20m ./cmd/connectorgen` — PASS (168.031s; required because
  this change updates CLI doc source/generated artifacts)
- `go vet ./...`, `go build ./cmd/pm`, `make tidy-check`, `make lint`,
  `make docs-check`, `make smoke-no-build` — PASS
- `go run ./cmd/agentcontractgen check`, `go run ./cmd/connectorgen validate`,
  `go run ./cmd/connectorgen surface-sync --check`, `make connector-boundary`,
  `scripts/tests/release-target-parity.sh` — PASS
- CLI docs regenerated and transcript fixture updated; website docs generator
  ran twice with no website diff.
- `git diff --check` — PASS

Repository guidance prohibits a monolithic `go test ./...` / `make verify` in
this per-command environment; CI owns that aggregate. The changed package and
the generated-artifact consumer were both run in full, and remaining
`make verify` gates were run individually.

## CLI help/manual/website parity

- `pm help <topic>`, bare legacy namespaces, leaf help, `-h`, and `pm version`
  were covered by the built-binary and regression sweeps.
- Generated `docs/cli/{extract,help,init,man,worker}.md` are tracked and pass
  `TestGoldenDocsGenerateMatchesTrackedCLIManuals`.
- Website source did not need a prose change: command spellings and flags did
  not change. `pnpm --dir website run gen:docs` was byte-stable across two runs.

## Delivery evidence

The issue-first GSD lifecycle was completed inline because the available
environment does not provide a compatible isolated Pi worker and the canonical
single-worker contract prohibits role spawning. See this phase's `PLAN.md`,
`TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, and `REVIEW.md`.

Required skills used: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`,
`golang-documentation`, and `golang-spf13-cobra`.
