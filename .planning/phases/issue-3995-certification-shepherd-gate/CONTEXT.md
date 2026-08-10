# CONTEXT — issue #3995 shared connector-certification Shepherd gate

## Phase mapping

GitHub child issue #3995 maps to this issue-named GSD phase under certification parent #3988.
It is implemented on `feat/3995-certification-shepherd-gate`, branched from the current
`origin/feat/3988-github-certification`; its eventual child PR base is exactly that parent branch.
The parent draft PR is #4018. #3984's generated-input delivery is present through merged PR #3999
at `815dc1ab65380e03f6e0c078ba36030baaec21ea`. #3985's concrete canon is present through merged
PR #4003 at `da7747a796049601a179a97c025bfb05f011f1e8`, while #3985 remains formally open.

## Locked decisions

- `.agents/agentic-delivery/canonical/delivery-contract.json` is the one canonical source. The
  contract defines a versioned, read-only certification gate, the accepted transitions, all input
  and verdict fields, and every generated harness projection.
- The evaluator lives under `internal/agentcontract`. It reads only the generated capability,
  flow, status, and accepted-evidence JSON created by #3984. It does not edit evidence, call a
  provider, inspect credentials, or use `cmd/connectorgen/certification*.go`.
- A certification gate binds every applicable capability, workflow, sync-mode primitive, and
  relevant flow pair for the requested connector. Each binding needs declaration, implementation,
  fixture testing, live testing, and a validated live-evidence pointer; the connector status must
  also be certified. File presence, route reachability, and an `implemented` flag alone never pass.
- The evaluator emits deterministic `PROCEED`, `RETRY`, or `HALT` verdicts. `RETRY` carries stable
  cell/evidence IDs such as `capability/github/capability:check/live_evidence`. Malformed,
  unsupported, absent, or adapter-incomplete contracts and inputs fail closed as `HALT`.
- The projected instruction block is rendered once from the canonical contract and embedded
  identically for Claude, Codex, Pi, and OpenCode. OpenCode projections are registered in the
  canonical projection list, not emitted by a special-case adapter.
- The current GitHub matrix is intentionally uncertified. General contract validation remains
  structural and does not fail merely because zero connectors are certified; transition evaluation
  rejects GitHub until valid live proof exists.

## Scope fences

- Do not alter connector bundles, provider behavior, proof-runtime work, live-provider tests,
  transport implementation, or `cmd/connectorgen/certification*.go`.
- Do not hand-author certification evidence or copy certification prose into connector definitions.
- Do not call providers, configure credentials, or mutate production/provider state.
- #3989 may revise the proof schema. This slice recognizes only the canonical schema version it
  declares and fails closed on any other version; it will report an unresolved #3989 integration
  dependency instead of inventing fields.

## GSD execution note

- `scripts/gsd doctor` and `go run ./cmd/agentcontractgen check` passed before edits.
- `scripts/gsd sources discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and
  `code-review` resolved successfully. Prompts were generated for this issue-named phase.
- `gsd-sdk query init.phase-op issue-3995-certification-shepherd-gate` returned
  `phase_found:false`: this issue is not an official numbered roadmap phase. The documented
  inline/manual GSD fallback is therefore used without weakening discussion, TDD, verification,
  gap handling, or review. This directory is the authoritative fallback record.
