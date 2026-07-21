# PRD Coverage — Issue #475

## Cycle 5 Diagnostic

Issue #475 remains a narrow Pi AgentSession runtime slice under parent issue #471. The repository
program PRD is connector-focused, so this phase's accepted issue contract, exact-head review
findings, and existing PLAN define the phase-equivalent coverage gate.

| Required outcome | Artifact / test boundary | Status before Cycle 5 execution |
|---|---|---|
| No referenced deadline timer from rejected duplicate/concurrent reservation | `agent-session-runtime.test.ts`; reservation ordering in `agent-session-runtime.ts` | expected RED captured |
| Nested flow values cannot hide later sensitive siblings | direct, prompt, tool-output, and handoff tests | expected RED captured |
| Unmatched apostrophe prose cannot hide the next structured assignment | direct, prompt, tool-output, and handoff tests | expected RED captured |
| Ordinary braces and flow-shaped comments cannot mutate harmless prose | byte-identical direct controls | expected RED captured |
| Prior structured, multiline, block, Bearer, flow, and spaced redaction remains intact | existing focused regression suite | retained |
| Scanner remains bounded and single-pass | explicit line/flow lexical state machine; monotonic cursors and balanced delimiters | implementation pending |
| Declared phase verification | focused/full Shepherd tests, pinned Pi 0.80.6 strict TypeScript, offline RPC, diff/base/scope | pending |

No dependency, CLI/help/docs/website, Go, connector, runtime-backed service, live credential, or
external mutation work is required. Parent orchestration owns fresh exact-head review and
integration after this worker returns a clean pushed head.
