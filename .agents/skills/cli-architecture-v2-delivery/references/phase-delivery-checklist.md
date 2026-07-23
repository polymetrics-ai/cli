# Phase delivery checklist

Use this checklist for one assigned CLI Architecture v2 slice. Domain skills provide implementation
details; this reference enforces delivery evidence.

## Before production edits

- [ ] Fetch and record default, parent, worker, PR-head, and merge-base identities.
- [ ] Confirm issue scope, dependencies, owner, isolated worktree, and disjoint write scope.
- [ ] Read current parent code and phase artifacts; do not reopen satisfied work from stale prose.
- [ ] Run `scripts/gsd doctor`, `scripts/gsd list`, and `scripts/gsd sources <command>`.
- [ ] Update PLAN, TDD ledger, verification checklist, prompts/run state where used, skills, and
      human gates.
- [ ] If `programming-loop` is unavailable, record the manual universal-loop fallback.
- [ ] Write and run the smallest RED test before behavior changes.

## Implementation loop

- [ ] GREEN implements only the assigned acceptance criterion.
- [ ] REFACTOR preserves green tests and existing machine/security behavior.
- [ ] Commit coherent plan, test, implementation, and review-fix checkpoints.
- [ ] Do not add a dependency without the exact approval required by the issue and ADR.
- [ ] Do not broaden into Connector Architecture v2 or unrelated connector bundles.

## Track-specific routing

### CLI, configuration, completion, help

Load Go CLI/Cobra/Viper/testing/security/documentation skills as applicable. Test a fresh command
tree, configuration precedence/isolation, completion no-file behavior, and output ownership. Verify
`pm help <topic>`, bare namespace behavior, `pm <command> --help`, generated help/manual,
`docs/cli/**`, website docs, completion metadata, and tests.

### Events and terminal UI

Load `bubble-tea-tui-design` plus context, concurrency, safety, security, testing, and documentation.
Use its complete gate matrix, including stdin+stdout TTY eligibility; stdin-piped and stdout-piped
fallbacks; CI, disabled, and dumb-terminal paths; JSON, plain, and no-input bypasses; resize;
no-color/ASCII/reduced-motion/accessibility; cancellation; final frames; sanitation; and race/leak
checks. Do not duplicate those mechanics here.

### Logging and OpenTelemetry

Load observability, context, security, error-handling, and testing. Add benchmarking and performance
skills when measuring or optimizing telemetry overhead. Preserve opt-in/disabled behavior,
redaction, bounded attributes, exporter shutdown, and stable stdout/stderr. Beta or additional
modules remain human-gated.

## Verification and handoff

- [ ] Run focused tests, race tests where concurrency changed, `go vet ./...`, `go test ./...`,
      `go build ./cmd/pm`, and `make verify` unless an applicable contract records a narrower gate.
- [ ] Record exact commands, outcomes, limitations, and any optional runtime checks not run.
- [ ] Re-fetch and verify the reviewed head has not drifted.
- [ ] Provide issue/base/head, commit range, files, tests, documentation parity, findings,
      dispositions, safety gates, and residual blockers to the parent orchestrator.
- [ ] Use `Refs` for stacked work. Never close default-branch delivery early or merge the parent.
