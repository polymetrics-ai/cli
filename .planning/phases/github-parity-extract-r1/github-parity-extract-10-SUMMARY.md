---
phase: github-parity-extract-r1
plan: "10"
coverage:
  - id: D1
    description: Typed, source-pinned GraphQL root/type inventory is generated without a raw GraphQL runtime escape hatch.
    verification:
      - kind: unit
        ref: scripts/tests/github-combined-operation-ledger.test.mjs
        status: pass
      - kind: other
        ref: node scripts/github-combined-operation-ledger.mjs --check
        status: pass
    human_judgment: false
  - id: D2
    description: The PM-only lab boundary and deploy-key readback safety gate remain green while provider activity is paused.
    verification:
      - kind: unit
        ref: scripts/tests/github-live-lab.test.mjs
        status: pass
      - kind: other
        ref: node scripts/github-live-lab.mjs --check-boundary
        status: pass
    human_judgment: false
---

# Summary — typed GitHub GraphQL source import

The combined GitHub ledger now has a version-2, source-bearing GraphQL type model.  It retains the
pinned official SDL provenance and all 305 root operations, while adding typed root arguments and
return values plus compact input/object/interface/union/enum/scalar facts needed by the next
generated-command slice.

`createEnterpriseOrganization` is now a typed inventory canary.  `node` and `nodes` distinguish
the 282 source-declared possible object types from the fixed documents' currently supported
projections.  This does not claim those roots are implemented or live-proven.

No provider fixture, credential, `pm` command, browser, or GitHub API operation ran in this slice.
The only external read was the public official source artifact used to reproduce the already-pinned
hash before mechanical regeneration.

The canonical single-worker lane required inline/manual GSD execution; the generated
`discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` prompts,
adapter health, source resolution, and `agentcontractgen check` are recorded in the plan/TDD ledger.
