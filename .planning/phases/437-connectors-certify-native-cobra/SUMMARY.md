# Phase 437 Summary

Status: complete and verified; terminal artifact commit/push pending.

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
