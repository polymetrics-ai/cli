---
phase: github-parity-extract-r1
plan: "11"
coverage:
  - id: G1
    description: Every source-pinned GitHub GraphQL root has a fixed typed PM operation and command with exact POST /graphql coverage.
    verification:
      - kind: unit
        ref: scripts/tests/gen-github-graphql-parity.test.mjs
        status: pass
      - kind: other
        ref: node scripts/gen-github-graphql-parity.mjs --check
        status: pass
    human_judgment: false
  - id: G2
    description: GraphQL direct reads and writes remain fixed-document, redacted, and lifecycle-bound.
    verification:
      - kind: unit
        ref: internal/connectors/engine/graphql_operation_test.go
        status: pass
      - kind: unit
        ref: internal/connectors/commandrunner/runner_test.go
        status: pass
    human_judgment: false
  - id: G3
    description: The permanent source inventory and PM-only lab safety boundary remain green without a provider request.
    verification:
      - kind: other
        ref: make github-parity-artifacts-check
        status: pass
      - kind: unit
        ref: scripts/tests/github-live-lab.test.mjs
        status: pass
    human_judgment: false
---

# Summary — generated fixed GitHub GraphQL contracts

The pinned GitHub GraphQL source inventory is now executable through 305 fixed typed PM contracts:
31 queries and 274 mutations, all sharing one exact `POST /graphql` transport row bound by operation
ID. The generated surface leaves legacy GraphQL compatibility bindings intact while keeping their
old four-operation denominator out of new source progress reporting.

The combined inventory is `1525/1525` classified, with `1345/1525` exactly implemented and
`0/1525` live-proven at this current head. That last number is intentionally zero: local generation,
fixtures, help, preflight, and historic runs do not constitute live provider acceptance.

The checkpoint also closes three review safety gaps: fixed GraphQL HTTP errors are redacted; a
direct-read document must contain exactly one named query and cannot select an appended mutation;
and a POST GraphQL query is consistently accounted as an executable read for capability validation.

The canonical single-worker lane used inline/manual GSD execution because role spawning is prohibited
in this parent lane. Required GSD prompts, skills, red/green evidence, verification, UAT, and review
records are in the plan and TDD ledger. No provider request or write occurred.
