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

Verification result: complete at implementation head `445ad06c`. Initial and two help-correction RED checkpoints preceded final production states. Focused/repeated/race, router/golden, full CLI/certify, exact-start operational differential 21/21, context/concurrency/events/telemetry/nondisclosure, runtime help, docs/website generation, connector validation 547/0, local certify smoke, gofmt/vet/build, and final `make verify` all pass. No defs, dependency, dynamic parser, live credential/write, service, PR, or review change.
