# Tracker: Post-GSD Merge Certify Full Sweep

**Created:** 2026-07-08  
**Purpose:** Track PR #123 merge, then rebase and resume the GitHub connector full-certificate branch from `SESSION_HANDOFF.md` using the repo-local GSD/Pi workflow.

## Current Gate

PR #123 has merged and the certificate workstream has been resumed on `feat/certify-full-sweep`.

```bash
gh pr view 123 --json state,mergedAt,mergeCommit,url
```

Observed states:

```text
Tracker creation: PR #123 OPEN, auto-merge enabled, mergeStateStatus=BLOCKED, mergedAt=null.
Resume checkpoint: PR #123 MERGED at merge commit 8d2ddf41; feat/certify-full-sweep rebased on origin/main and pushed.
```

## Secret Handling Notice

A GitHub token was pasted into chat during tracker creation. Do not copy, store, commit, echo, summarize, or use that value. Treat it as compromised and rotate/revoke it immediately.

Credentialed tests must receive secrets only through local environment variables or a secret manager, for example:

```bash
export PM_GITHUB_DEV_TOKEN=... # local shell only; never commit, paste, or log
```

Do not put credentials in `.planning`, PR bodies, issue comments, logs, test fixtures, or command history examples.

## Branch / Worktree To Resume

Source: local `SESSION_HANDOFF.md` (untracked; not part of PR #123).

Expected certificate branch/worktree from handoff:

```text
worktree: /Users/karthiksivadas/Development/polymetrics-cli-agents/wt-gated-principle
branch: feat/certify-full-sweep
parent branch: feat/44-github-cli-parity
active issue: #121
parent issue: #44
```

Verify before any rebase:

```bash
cd /Users/karthiksivadas/Development/polymetrics-cli-agents/wt-gated-principle
git status --short --branch
git log --oneline -5
gh issue view 121 --json number,title,state,url
```

## Post-Merge Rebase Plan

After PR #123 is merged:

```bash
# Confirm GSD PR landed
gh pr view 123 --json state,mergedAt,mergeCommit,url

git fetch origin main feat/44-github-cli-parity

# Resume certificate branch worktree
cd /Users/karthiksivadas/Development/polymetrics-cli-agents/wt-gated-principle
git status --short --branch

# Ensure branch is the certificate branch
git switch feat/certify-full-sweep

# Rebase on the correct target after deciding with parent branch owner:
# Option A: if #121 remains stacked under parent PR #49
git rebase origin/feat/44-github-cli-parity

# Option B: if parent branch has merged and #121 should target main
git rebase origin/main
```

Stop for humans if:

- the branch has uncommitted changes;
- the parent branch no longer exists;
- issue #121 scope changed;
- rebase conflicts touch shared generated files or unrelated connector surfaces;
- auth scopes, dependencies, destructive actions, or quality gates change.

## Required GSD / Pi Workflow

Use the repo-local official GSD adapter before implementation:

```bash
scripts/gsd doctor
scripts/gsd prompt programming-loop init --phase issue-121-certify-full-sweep --dry-run
scripts/gsd prompt plan-phase 121 --skip-research
```

In Pi after project trust/reload:

```text
/gsd doctor
/gsd-programming-loop init --phase issue-121-certify-full-sweep --dry-run
/gsd plan-phase 121 --skip-research
```

Required references:

- `AGENTS.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md` if CLI/help/docs/website changes are made
- `.agents/agentic-delivery/references/runtime-rlm-website-integration.md` for RLM/Pi-agent/runtime work
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`

Required skills for #121-style work:

- `golang-how-to`
- `golang-testing`
- `golang-security`
- `golang-safety`
- `golang-error-handling`
- `golang-cli` for CLI surfaces
- `golang-documentation` for CLI/docs/website updates
- `golang-context` and `golang-concurrency` for worker/runtime orchestration
- `golang-database` if PostgreSQL paths are changed
- design skills when `website/**` changes

## Implementation Scope From Handoff

Issue #121 goal from session handoff:

```text
Full certificate - all-streams read sweep + binary + direct-read + flow/schedule
```

Planned slices:

1. all-streams read sweep;
2. binary sweep;
3. direct-read sweep;
4. flow/schedule per stream;
5. live write lifecycle when credentials are available;
6. typed-confirmation gates / destructive/admin gates / secret transforms;
7. full certificate emission + merge gate.

Start with all-streams read sweep unless issue #121 has changed.

## Test-First Requirements

Every implemented behavior needs tests before production changes.

Minimum local gates for non-credentialed slices:

```bash
go test ./internal/connectors/certify/ -run TestFullSweepReadsAllStreams -count=1 -v
go test ./internal/connectors/certify/ -count=1
go test ./internal/connectors/... -count=1
git diff --check
```

Broader gates before PR handoff:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

Do not weaken gates to make credential-dependent tests pass. Split fixture/replay tests from live tests.

## Credentialed / Live Test Gate

Live credentialed checks are optional and must be explicitly gated.

Use environment variables only:

```bash
PM_GITHUB_DEV_TOKEN=... POLYMETRICS_INTEGRATION=1 go test ./internal/connectors/certify/ -run TestGitHubFullCertificateLive -count=1 -v
```

Expected behavior without credentials:

- fixture/replay tests still run;
- live stages report `uncertified`, skipped, or blocked as designed;
- missing credentials are not treated as product failures;
- reports must not print token values.

## RLM / Bus-Factor / Future Pi-Agent Test Plan

Future RLM work should use `.agents/agentic-delivery/references/runtime-rlm-website-integration.md`.

Non-credentialed first:

```bash
pm rlm run --spec <fixture-spec.json> --out bus_factor_scores --mode deterministic --json
pm rlm run --spec <fixture-spec.json> --out bus_factor_scores --mode fixture --json
```

Runtime-backed/Pi-agent future gate, only when explicitly in scope:

```bash
scripts/runtime.sh doctor
scripts/runtime.sh up
pm runtime doctor --json
pm agent image ensure --json
pm worker status --json
pm rlm run --spec <bus-factor-spec.json> --in <stream-table> --out bus_factor_scores --mode agent --request "score bus factor risk for this stream" --json
scripts/runtime.sh down
```

Future Pi-agent behavior to track:

- Pi agent chooses the RLM scoring problem based on the stream metadata.
- Stream metadata determines whether bus-factor scoring is relevant.
- The RLM problem selection is typed and auditable, not a generic free-form remote command.
- Agent mode remains data-only and does not mutate external systems.
- Runtime-backed tests are separated from dependency-free tests.

## CLI / Docs / Website Parity Gate

If #121 adds CLI commands for binary/direct-read/certification surfaces, update and verify:

```bash
pm help <topic>
pm <namespace>
pm <command> --help
rg -n "<command>|<flag>|<topic>" docs/cli website
```

Follow `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.

## Review / PR Routing

- If #121 remains stacked, target the parent branch and use `Refs #121`.
- If targeting `main` directly and completing #121, use `Closes #121`.
- Wait for CodeRabbit automatic review on non-draft PRs to `main`.
- Do not post redundant `@coderabbitai review` commands unless fallback conditions apply.
- Copilot review is fallback-only when CodeRabbit is blocked.

## All-Streams Read Sweep Slice Evidence

Status: first implementation slice in progress and locally green for the certify package.

Behavior added:

- `Options.Full` now iterates every stream returned by the live catalog for source read stages.
- Per-stream live, capture, file credential, and incremental connection names avoid collisions.
- Catalog primary key and cursor fields are reused per stream, with legacy `id` / `updated_at` defaults.
- Flow roundtrip now uses the stream-scoped capture credential and catalog-derived primary/cursor fields after a full sweep.

Test-first evidence:

```text
Red: go test ./internal/connectors/certify/ -run 'TestCatalogStreamSpecsFromStreams|TestFullSweepNamesAreStreamScoped|TestFullSweepStreamSpecsFallbackToSelectedStream' -count=1 -v
     failed on undefined full-sweep helpers before production edits.
Green: go test ./internal/connectors/certify/ -run TestFullSweepSourceStagesAgainstSample -count=1 -v
Green: go test ./internal/connectors/certify/ -count=1 -timeout=10m
Green: go test ./internal/connectors/... -count=1
```

No credentialed/live GitHub checks were run. The compromised pasted token remains unused and must be revoked/rotated before any future credentialed testing.

## Flow/Schedule Per-Stream Slice Evidence

Status: full mode now runs glue stages per catalog stream.

Behavior added:

- `Options.Full` runs `flow_roundtrip` and `schedule_roundtrip` after each stream’s read/capture stages instead of only once for the final stream.
- Flow names, flow tables, flow connection names, query tables, and schedule names are stream-scoped during full sweeps to avoid collisions.
- Non-full behavior keeps the original single flow/schedule tail stages.

Test-first evidence:

```text
Red: go test ./internal/connectors/certify/ -run TestFullSweepFlowAndScheduleNamesAreStreamScoped -count=1 -v
     failed on undefined stream-scoped glue helpers before production edits.
Green: go test ./internal/connectors/certify/ -run 'TestFullSweepFlowAndScheduleNamesAreStreamScoped|TestFullSweepSourceStagesAgainstSample' -count=1 -v
Green: go test ./internal/connectors/certify/ -count=1 -timeout=10m
Green: go test ./internal/connectors/... -count=1
```

## Binary Download Sweep Slice Evidence

Status: binary-download certification safety gate added for full mode.

Behavior added:

- `Options.Full` now runs a binary-download certification stage when a curated connector candidate exists.
- GitHub uses the declared `release download` command path and verifies operation-backed binary executors remain safely blocked instead of writing files without an explicit bounded file policy.
- Connectors without curated binary candidates record a documented non-failing skip and a `capabilities.binary` result.
- The stage scans binary command output for secret leaks before reporting `blocked`/pass.

Test-first evidence:

```text
Red: go test ./internal/connectors/certify/ -run 'TestBinaryDownloadCandidateForGitHub|TestBinaryDownloadCandidateForUnknownConnector' -count=1 -v
     failed on undefined binary candidate helper before production edits.
Green: go test ./internal/connectors/certify/ -run 'TestBinaryDownloadCandidateForGitHub|TestBinaryDownloadCandidateForUnknownConnector|TestFullSweepSourceStagesAgainstSample' -count=1 -v
Green: go test ./internal/connectors/certify/ -count=1 -timeout=10m
Green: go test ./internal/connectors/... -count=1
```

No binary bytes are downloaded in dependency-free tests. A real bounded binary executor remains a future implementation gate; this slice certifies that the current binary surface is explicit and safely blocked.

## Direct-Read Sweep Slice Evidence

Status: curated direct-read certification stage added for full mode.

Behavior added:

- `Options.Full` now runs a direct-read certification stage when a curated connector candidate exists.
- GitHub uses the implemented `repo read-file` command with credential resolution, bounded `--max-bytes`, and content/download URL redaction from the commandrunner path.
- Connectors without curated direct-read candidates record a documented non-failing skip and a `capabilities.direct_read` result.
- The stage scans direct-read output for secret leaks before reporting pass.

Test-first evidence:

```text
Red: go test ./internal/connectors/certify/ -run 'TestDirectReadCandidateForGitHub|TestDirectReadCandidateForUnknownConnector' -count=1 -v
     failed on undefined direct-read candidate helper before production edits.
Green: go test ./internal/connectors/certify/ -run 'TestDirectReadCandidateForGitHub|TestDirectReadCandidateForUnknownConnector|TestFullSweepSourceStagesAgainstSample' -count=1 -v
Green: go test ./internal/connectors/certify/ -count=1 -timeout=10m
Green: go test ./internal/connectors/... -count=1
```

No credentialed/live GitHub checks were run. The stage is ready for future env-gated live runs with rotated credentials only.

## Verification Checkpoint

Latest local gates for the pushed implementation slices:

```text
PASS: git diff --check
PASS: go vet ./...
PASS: go build ./cmd/pm
PASS: go test ./internal/connectors/certify/ -count=1 -timeout=10m
PASS: go test ./internal/connectors/... -count=1
PASS: PM_CRONTAB_FILE=$(mktemp) go test ./internal/cli -run TestScheduleCLI_Remove -count=1 -timeout=2m -v
CAVEAT: go test ./... timed out in internal/cli TestScheduleCLI_Remove when it invoked the host crontab without PM_CRONTAB_FILE; the isolated crontab-seam rerun passed.
```

## Tracker Checklist

- [x] PR #123 merged.
- [x] Certificate branch/worktree verified.
- [x] Target rebase base confirmed (`main` for the resumed branch checkpoint).
- [x] Rebase completed cleanly or blockers recorded.
- [x] GSD programming loop prompt generated for #121.
- [x] Required Go/runtime/design skills loaded and recorded.
- [x] Red tests added for first resumed slice.
- [x] Fixture/non-credentialed tests green for the first all-streams read sweep slice.
- [x] Credentialed tests gated behind env vars only.
- [x] Binary sweep safety gate implemented and tested (executor remains future gated work).
- [x] Direct-read sweep implemented and tested.
- [x] Flow/schedule per stream implemented and tested.
- [ ] RLM bus-factor fixture plan created.
- [ ] Future Pi-agent stream-based RLM problem selection tracked.
- [ ] CLI/docs/website parity verified for CLI-visible changes.
- [ ] PR opened/updated with verification and safety notes.
