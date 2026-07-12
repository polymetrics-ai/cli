# Verification: Shepherd Supervision

Status: complete and independently reviewed; ready for the stacked PR checkpoint.

```bash
bash -n scripts/pi-shepherd-loop.sh scripts/tests/pi-shepherd-supervision.sh
shellcheck --severity=warning scripts/pi-shepherd-loop.sh scripts/tests/pi-shepherd-supervision.sh
bash scripts/tests/pi-shepherd-supervision.sh
bash scripts/tests/auto-loop-control.sh
make agent-loop-test
make verify
git diff --check fix/323-auto-loop-hardening...HEAD
```

Tests use temporary repositories, fake Pi processes, synthetic identifiers, and exact process
groups. They must never inspect credentials, invoke a model/provider, or signal unrelated PIDs.

## Focused results — 2026-07-12

- Bash 3.2 syntax and ShellCheck warning gate: pass.
- Nineteen-scenario supervision harness: pass, including 32-way contention, inherited descendant
  fencing, both role deadlines, signal/HALT failure paths, state aliases and isolated resume
  invariants, stale-verdict retirement, terminal ratification, and exact model policy.
- Phase 0 safety/control regression: pass.
- `make agent-loop-test`: pass, including `go test -race ./internal/agentloop/...`.
- The validator-only model policy is asserted through fake local model discovery; no provider call
  is made. Orchestrator/worker model configuration remains unchanged.
- Final independent review: no remaining in-scope P0/P1 finding.

Enable blockers deliberately deferred by contract: same-epoch snapshot rollback requires #327's
per-transition journal predecessor/version; authenticated takeover and escaped-descendant
containment require #339/#342. The Phase 0 production fuse remains closed until those dependent
phases and later canary gates land.

## Full result — 2026-07-12

- `make verify`: pass, including the 414-second connector certification package, all Go tests,
  vet/build, documentation validation, smoke flow, lint, 547 connector-definition checks,
  agent-loop race tests, Phase 0 controls, and the 19-scenario supervision harness.
- `git diff --check`: pass.

Pending delivery actions: commit, push, stacked PR creation, and automated PR review coverage.
