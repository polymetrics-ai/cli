# PRD Coverage — Issue #475

## Cycle 6 Diagnostic

Issue #475 remains a narrow Pi AgentSession runtime slice under parent issue #471. The repository
program PRD is connector-focused, so this phase's accepted issue contract, exact-head review
findings, and existing PLAN define the phase-equivalent coverage gate.

| Required outcome | Artifact / test boundary | Status before Cycle 6 execution |
|---|---|---|
| Multiline nested flow values cannot hide later same-line sensitive siblings | direct, prompt, `workspace_read`, typed capability, and handoff tests | focused GREEN |
| Apostrophes inside unquoted words cannot hide later sensitive siblings | same five consumer boundaries | focused GREEN |
| Safe `rock-'n-roll` values remain byte-identical | direct harmless control | pass |
| Line-end discovery is near-linear for many same-line assignments | deterministic scan metrics at approximately 25/50/100 KiB | focused GREEN; visits equal bytes |
| Prior lifecycle and redaction invariants remain intact | existing 36 focused regressions | pass |
| Typed lexer remains the single transformer architecture | value-local multiline closers, token-aware quotes, monotonic line cursor | implemented |
| Declared phase verification | focused/full Shepherd tests, pinned Pi 0.80.6 strict TypeScript, offline RPC, diff/base/scope | pending |

No dependency, CLI/help/docs/website, Go, connector, runtime-backed service, live credential, or
external mutation work is required. Parent orchestration owns fresh exact-head review and
integration after this worker returns a clean pushed head.
