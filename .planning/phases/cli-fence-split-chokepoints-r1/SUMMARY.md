---
coverage:
  - id: D1
    description: GitHub and PostgreSQL certification is connector-local, allowlisted, and reconstructed only in memory.
    verification:
      - kind: unit
        ref: cmd/connectorgen TestCertificationShardsRoundTripGeneratedMatrices
        status: pass
      - kind: other
        ref: go run ./cmd/connectorgen certification-matrix --check
        status: pass
    human_judgment: false
  - id: D2
    description: A connector-scoped generation changes no other connector shard or shared status projection.
    verification:
      - kind: unit
        ref: cmd/connectorgen TestCertificationScopedGenerationLeavesOtherShardByteIdentical
        status: pass
    human_judgment: false
  - id: D3
    description: Generated source anchors are symbols and remain stable across line insertions.
    verification:
      - kind: unit
        ref: cmd/connectorgen TestCertificationSourceAnchorsUseSymbols
        status: pass
      - kind: other
        ref: post-change binary_download insertion SHA-256 comparison
        status: pass
    human_judgment: false
  - id: D4
    description: App composition/mode dispatch and PostgreSQL CDC capability state are mechanically preserved.
    verification:
      - kind: unit
        ref: go test -count=1 ./internal/app ./internal/connectors/native/postgres
        status: pass
    human_judgment: false
---

# Summary — Fence split chokepoints r1

## Manual GSD verify-work fallback

This task is not a numbered roadmap phase, so Pi cannot initialise the official phase command.
The coverage record above maps every deliverable to a passing automated test or repeatable command;
no human-judgment deliverable remains. `VERIFICATION.md` records the complete local gate set and
the source-insertion measurement.

## Manual GSD code review fallback

Reviewed the generator command parser, scoped source/evidence loading, shard validation and
reconstruction, status fallback, generated outputs, app moves, PostgreSQL capability row table, and
documentation against the zero-behaviour-change constraints. No critical or warning finding remains.
The review specifically checked that a scoped run writes its requested shard only, that the drift
gate reconstructs all allowlisted shards, and that unknown connectors retain the former status error.

## CLI help/docs parity

The product `pm` command surface is unchanged. `connectorgen` is repository developer tooling; its
existing top-level usage now documents the new scoped invocation. No `pm` help topic, manual,
website page, generated manual, JSON output, credential flow, or completion surface applies. The
architecture and connector-authoring documentation were updated, and `make docs-check` passed.
