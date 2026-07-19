# Phase 437 Prompts

## Kickoff snapshot

Task: implement issue #437, final serialized Phase 9 unit under #397/#407, from exact `6c038bb4ab4a5497fca28a0cab42d0a7fa4eb22b` on isolated `refactor/437-connectors-certify-native-cobra`; Sol/high; no dependencies/services/credentials/PR/review.

Identity: `issue-437-pi-sol-high-20260719T095145Z`.

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt plan-phase 437 --skip-research
scripts/gsd prompt programming-loop init --phase 437 --dry-run
```

Doctor/list and plan generation passed. `programming-loop` is absent from the adapter registry, so the manual universal-runtime-loop fallback applied.

Execution decision: `local_critical_path` — final assigned serialized namespace; central router scope collides; no subagent tool exposed; user bounded delivery to implementation/commit/push.

Required skills: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-spf13-cobra`.

Safety prompt: preserve `cli.Run`, dynamic connector dispatch, exits 0/1/2/3, context/events/telemetry, stdout/stderr, and credential-value exclusion. Use only fixture/replay/local tests and temp roots. Do not touch defs, call live credentials, perform external writes, add dependencies/services, or create PR/review.

Downstream artifact: native connectors/certify subtree, declared current flags and runtime seam, compatibility normalization, contextual direct/trailing help, focused parity tests, and canonical manual/docs/website updates.

Verification result: superseded by the accepted correction cycle below; current phase verification is pending.

## Accepted correction kickoff snapshot

Task: at exact HEAD `0d1792cec3ea829ceb6228fc600b6dc7bbd90eee`, accept all five findings in `/tmp/pm-397-review-437.log`; update artifacts and commit/push them, then add/commit RED tests before production. Fail closed on unsupported safety/mode flags; restore legacy single span/validation/options and batch file/parallel/error precedence; exact-only connectors help; accurate pre-report versus completed-report exits across canonical/generated/website docs. Sol/high; fixture/temp only; no dependencies/services/credentials/writes/PR/review.

Commands: `scripts/gsd doctor` passed; both `scripts/gsd prompt programming-loop ...` forms returned `unknown GSD command`, so the manual universal runtime loop applies. Execution decision: `local_critical_path` (bounded correction, isolated worktree, no subagent tool).

Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `golang-spf13-cobra`.

Downstream artifact: complete at implementation head `a67d2ff9de84a2fabcd3b66097bf49518c1fa124`; terminal verification artifact `2987f21b` is pushed and this final delivery marker closes the phase.

Verification result: pass. GREEN evidence: focused `3.004s`; native/certify/telemetry `108.532s`; base/current differential 5/5 byte-identical; focused race `29.046s`; ×10 `24.991s`; certify redaction/replay/concurrency race `349.263s`; exit focus `21.618s`; local sample exit 0/pass/redacted; docs/golden `24.275s`; website regeneration hash-stable; full CLI `435.572s`; certify `338.846s`; validation 547/0; final `make verify` exit 0 in `468.36s`. No live credentials/writes, services, dependencies, PR, or review.
