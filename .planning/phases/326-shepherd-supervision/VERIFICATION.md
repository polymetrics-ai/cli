# Verification: Shepherd Supervision

Status: final private-capability correction passes local gates; exact-head independent review pending.

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
- Pre-stability independent review found no P0/P1 at that earlier head; later exact-head review
  blocked the filesystem authorization revision and drove the corrections below.

Enable blockers deliberately deferred by contract: same-epoch snapshot rollback requires #327's
per-transition journal predecessor/version; authenticated takeover and escaped-descendant
containment require #339/#342. The Phase 0 production fuse remains closed until those dependent
phases and later canary gates land.

## Full result — 2026-07-12

- `make verify`: pass, including the 414-second connector certification package, all Go tests,
  vet/build, documentation validation, smoke flow, lint, 547 connector-definition checks,
  agent-loop race tests, Phase 0 controls, and the 19-scenario supervision harness.
- `git diff --check`: pass.

## Stability correction result — 2026-07-12

- Filesystem role authorization and the intermediate `SIGSTOP`/`SIGCONT` handshake have both been
  removed. A mode-0600 FIFO is opened and unlinked before role spawn; the exact bound role can
  consume its one-use inherited GO capability only after the controller's durable bind returns.
- The role receives only the remaining shared turn budget and revalidates parent liveness,
  canonical binding, token shape, and deadline immediately before exec. A late validator cannot
  receive a fresh full timeout.
- Leader completion distinguishes explicit zombies from live/uncertain processes and reaps each
  leader at most once.
- A PID/PGID mismatch never signals the untrusted PID; it retains durable unresolved evidence for
  authenticated recovery.
- Invalid/duplicate/empty/unknown focused-test filters fail rather than reporting a zero-test green.
- Current-design repeated evidence: focused private-GO/deadline suite, a ten-way parallel risk
  matrix 10/10, two full 22-scenario suites, Bash syntax, ShellCheck warnings, Phase 0 controls,
  Go unit/race checks, and the fixed-snapshot `make agent-loop-test` all pass. Earlier
  failed-HALT/instant-validator 50/50 and shared-validator-deadline 10/10 evidence remains valid.
- Final `make verify` passes formatting/tidy, vet, all Go tests, build, docs validation, CLI smoke,
  lint, 547 connector definitions, race/control gates, and the expanded 22-scenario harness.
- The SIGKILL oracle directly validates the inherited canonical lock descriptor and waits for
  durable binding/process disappearance without acquiring or changing that lock. The private-GO
  oracle also sends the formerly exploitable external `SIGCONT` and proves it has no authority.

Pending delivery actions: exact-head independent local review, push, CI, and parent integration.
