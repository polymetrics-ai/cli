# Phase 437 Summary

Status: accepted safety-critical correction complete, verified, committed, and pushed.

## Identity

- Session: `issue-437-pi-sol-high-20260719T095145Z`
- Profile: Sol/high
- Branch: `refactor/437-connectors-certify-native-cobra`
- Exact start/base: `6c038bb4ab4a5497fca28a0cab42d0a7fa4eb22b` / `feat/cli-architecture-v2`
- Implementation head: `445ad06c`
- Parent #397, umbrella #407, draft parent PR #438

## Delivered

`connectors` is now a native Cobra subtree with declared `list`, `catalog`, `inspect`, `man`, `docs`, hidden positional help, and nested `certify`. Certify declares the complete current command-contract flag set and adapts single, batch, and sweep execution through an invocation-local runtime interface. The existing in-process runner, report rendering, telemetry spans, context propagation, bounded batch workers, events, and exit 0/1/2/3 contract remain intact.

Compatibility normalization preserves repeated/bare/assigned/space flags, legacy ignored operands and unknown flags, literal separators, malformed command heads, action/operand ownership, and invocation globals. Direct and trailing action help now intentionally renders the canonical connectors manual without running an operation. Invalid actions still exit 2. Only connectors/certify parser calls and connectors' legacy wrapper were removed; `internal/cli/parse.go` and both sanctioned dynamic `pm <connector> <path>` parser call sites are unchanged.

The canonical manual, generated `docs/cli/connectors.md`, golden transcript manual content, website CLI reference, and generated website data now document certify modes, JSON envelopes, credential-reference safety, and certification exits.

## GSD / TDD

Doctor/list and plan-phase prompt generation passed. The adapter lacks `programming-loop`, so the manual universal loop was used. Six artifacts preceded production. Initial RED failed on absent native constructors. Two later focused RED checkpoints exposed operation output/usage on trailing help and connector-name capture on direct inspect help; both corrections were test-first and committed separately.

## Verification

Final focused native tests passed (`3.989s`), repeated ×10 (`34.833s`), race (`40.842s`), router/golden/certify/telemetry (`111.919s`), and certify concurrency/event race (`2.395s`). The exact-start operational differential matched 21/21 unchanged cases; contextual action help is intentional. Runtime topic/bare/direct/positional/trailing text and JSON help passed. Connector validation checked 547 bundles with 0 findings. The explicit sample certify smoke returned exit 0, `ConnectorCertification`, pass, with empty stderr.

Final `make verify` exited 0: CLI `431.305s`, certify `337.280s`, full vet/tests/build, docs validation, ordered local smoke, lint 0, and connector validation green. Website docs generation is drift-free. Website typecheck was not applicable because no existing TypeScript tool installation is present; no dependency install was attempted.

## Safety / handoff

Verification used fixture/replay/local sample paths and temporary roots only. No live credential checks, external writes/sweeps, connector definitions, dependencies, services, real credential values, generic tools, destructive/admin/production operations, PR, or review. Parent-orchestrator integration/review coverage remains intentionally pending.

## Correction in progress

At exact HEAD `0d1792cec3ea829ceb6228fc600b6dc7bbd90eee`, all five findings from `/tmp/pm-397-review-437.log` were accepted. Implementation head `a67d2ff9de84a2fabcd3b66097bf49518c1fa124` hides and rejects six unsupported controls before runner invocation, so replay is credential-free by refusal and cannot enter live/write stages. It restores exactly one single-certify span with connector validation before option parsing, credential-file load before batch parallel parsing with legacy error bytes/wrappers, and exact-only connectors help normalization. Canonical/generated/website docs now separate CLI pre-report exits (usage 2, validation 3, setup/runtime 1) from completed report outcomes (pass 0, failure 2, leaks 3).

RED reproduced all findings. GREEN passed focused `3.004s`, native/certify/telemetry `108.532s`, base/current differential 5/5 byte-identical, race `29.046s`, ×10 `24.991s`, certify redaction/replay/concurrency race `349.263s`, exit focus `21.618s`, local sample exit 0/pass/redacted, docs/golden `24.275s`, and hash-stable website generation. Full CLI passed `435.572s`, certify `338.846s`, connector validation 547/0, and final `make verify` exited 0 in `468.36s` with CLI `444.436s`, certify `346.018s`, docs, smoke, lint, build, vet, and validation green. No live credentials/writes, services, dependencies, PR, or review.
