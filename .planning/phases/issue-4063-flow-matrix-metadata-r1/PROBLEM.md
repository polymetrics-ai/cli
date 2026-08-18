# #4063: Refresh flow-authoring discovery metadata

**Parent:** #3897
**Program parent:** #3988
**Existing stacked PR:** #4060
**Branch:** feat/3897-flow-connection-scope-nm
**Required base:** feat/3988-github-certification
**Exact starting head:** 002ddf674a447bf0872486aa979efdaa078f602c

## Problem

The flow source annotation remains at
internal/cli/flow_cli.go:20 after func runFlow moved to line 21. The
generated flow matrix therefore fails its canonical drift check even though
the runtime source is correct.

## Correction reservation

The authoritative #3897 TDD ledger and RUN-STATE at the exact starting head
both record correction 3 / 5. Issue #4063 reserves correction 4 / 5 before
any generated-file mutation. One correction remains after this bounded fix.

## Scope boundary

Run the canonical certification-matrix generator and accept only the scalar
change in workflow_kinds[flow_authoring].discovery_source:

  internal/cli/flow_cli.go:20 -> internal/cli/flow_cli.go:21

Do not modify runtime behavior, tests, capability facts, connector definitions,
certification status, accepted evidence, or another generated artifact.

## Manual-GSD fallback

The GSD adapter and every required command resolve, but phase-op and
plan-phase reject this named issue because the active ROADMAP.md has no
numbered issue phase. The repository contract requires the lifecycle anyway,
so this named directory records the discuss, plan --tdd, execute,
verify-work, and code-review evidence inline. The canonical single-worker
contract forbids GSD-role delegation.
