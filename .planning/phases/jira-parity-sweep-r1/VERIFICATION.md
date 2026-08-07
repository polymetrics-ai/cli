# Verification — jira parity sweep (state at captain-ordered pause, 2026-08-07)

## Run and green

| gate | result |
| --- | --- |
| `go test ./cmd/connectorgen/ -run TestJiraDocumentedSurfaceIsComplete` | **PASS** (red first, captured verbatim in `TDD-LEDGER.md`) |
| `go test ./cmd/connectorgen/` — the **whole** package (finding 5) | **PASS**, 12.3 s |
| `go run ./cmd/connectorgen validate` | **551 connector(s) checked, 0 findings** |
| `go run ./cmd/connectorgen surface-sync --check` | **clean**, 0 filled / 0 corrected |
| `go test ./internal/connectors/commandrunner/ -run TestEveryImplementedCommandPassesRuntimePreflight` | **PASS** |
| `gofmt -l cmd internal` | no output |
| `go vet ./cmd/connectorgen/ ./internal/connectors/...` | clean |
| `go build ./cmd/pm` | ok |

## Reachability — 590 / 590, measured by running the binary

Authoring a command is not evidence that it routes, and **exit status proves nothing**: a namespace
miss renders group help and exits 0 (finding 30).

```
go build -o /tmp/pm-jira ./cmd/pm
PM_BIN=/tmp/pm-jira PM_CONNECTOR=jira \
  xargs -P 12 -I{} tools/probe_reachability.sh "{}" < /tmp/sweep/jira-cmds.txt
→ 0 lines of output over 590 implemented/partial commands
```

The probe asserts the rendered `NAME` line, and it was proved to be capable of failing before its
zero was believed:

```
$ /tmp/pm-jira jira issue get-issue --help | head -2
NAME
  pm jira issue get-issue - Get issue.

$ probe_reachability.sh "issue no-such-command"
FAIL(unrouted): issue no-such-command ::   pm jira - Jira command surface
```

## Endpoint-ledger delta — confined to jira, inspected BY OBJECT

```
551 connectors before, 551 after · added: [] · removed: [] · changed: ['jira'] (0 → 22 entries)
```

22 = the 22 read-shaped POSTs modelled as operation-backed `rest_read`. The 270 plain direct reads
add nothing, which is finding 26's low-blast-radius shape working as intended.

## Known-unmet, carried — none caused by this change

Measured rather than claimed: `git stash push -u`, re-run, diff. The failing set is identical with
and without the jira slice; only durations differ.

1. **`TestGoldenTranscripts` — eleven subtests.** Pre-existing since before github. Discharged by
   the end-of-sweep regeneration, never per connector (finding F6).
2. **`internal/connectors/certify`** — `TestSurfaceInventoryForGitHubAccountsForAllReviewedEndpoints`
   and `TestGithubWriteActionInventoryAccountsForAllDeclaredActions`. Pre-existing on this branch,
   about **github**, and **not listed in the PROGRESS.md known-red section**. Recorded as finding 53.
3. **`internal/connectors/defs/zendesk-support`** — `TestReverseETLLedgerReconciles`,
   `TestDestructiveOperationsStayBlocked`, `TestReverseETLWriteActionsExecute`. Pre-existing on this
   branch, also unlisted. Recorded as finding 53.

## NOT RUN at the pause — state this rather than imply it

- **`go test -timeout 20m ./internal/cli/`.** Not run. The package takes ~13 minutes with a large
  connector and the pause order arrived first. It must be run before the sweep PR opens, **with
  `-timeout 20m`** — the bare form inherits Go's 600s default and dies mid-run looking exactly like a
  hang you caused (finding 36).
- **`make verify`'s other gates** (`tidy-check`, `lint`, `docs-check`, `smoke-no-build`,
  `agent-contract-check`, `connectorgen-validate`, `connectorgen-surface-sync`,
  `connector-boundary`, `release-workflow-check`). `connectorgen-validate` and
  `connectorgen-surface-sync` were run directly and are green; the rest were not run.
- **CLI help / docs / website parity** (`docs/cli/**`, `website/**`, generated help and manual
  artifacts) per `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`. `pm jira
  <cmd> --help` renders for all 590, and `pm jira` with no action renders the group help, but the
  **docs and website regeneration has not been done**. It is part of the end-of-sweep shared-artifact
  regeneration, which is deliberately run once at the end rather than per connector (finding F6).
- **`GSD_BASE_REF=origin/main ./scripts/verify-gsd-workflow`.** Not run at the pause; the evidence it
  checks for (`PLAN.md`, `RUN-STATE.json`, `TDD-LEDGER.md`, `SUMMARY.md`, `VERIFICATION.md`, plus the
  `.planning/traces/` lifecycle trace) is all present and committed.
