---
coverage:
  - id: D1
    description: Changefeed declarations use a closed, evidence-backed descriptor and legacy CDCReader alone cannot make a capability public.
    verification:
      - kind: unit
        ref: internal/connectors/changefeed_test.go (descriptor, executor, metadata, manifest, and catalog projections)
        status: pass
      - kind: unit
        ref: internal/connectors/engine/bundle_test.go (optional and invalid descriptor loading)
        status: pass
    human_judgment: false
  - id: D2
    description: PostgreSQL's unsupported logical-replication stub is not catalogued as implemented CDC and remains explainable through inspect JSON.
    verification:
      - kind: integration
        ref: internal/cli/changefeed_cli_test.go
        status: pass
      - kind: other
        ref: go run ./cmd/pm connectors catalog --capability cdc --json; go run ./cmd/pm connectors inspect postgres --json
        status: pass
    human_judgment: false
  - id: D3
    description: The foundation does not classify the remaining connector fleet or add a dependency, provider call, redaction path, command-runner change, conformance work, surfacing work, or generator enforcement.
    verification:
      - kind: other
        ref: changed-path audit, git diff --check, and connector-boundary gate
        status: pass
    human_judgment: false
---

# SUMMARY — issues #3745 and #3746 truthful changefeed discovery

## Delivered

- Added a closed optional `changefeed.json` contract with the researched status and mechanism
  vocabularies, provider source record, named executor, checkpoint/recovery properties, delivery
  guarantees, stream coverage, strict loading, and defensive projection.
- Made public CDC truth fail closed: an implemented descriptor must match a registered
  `ChangefeedExecutor`; legacy metadata and `CDCReader` presence are insufficient in list,
  manifest, catalog, and inspect JSON projections.
- Added PostgreSQL's evidence-backed `unsupported` logical-replication descriptor. Its legacy
  `CDCReader` remains an unsupported migration stub and cannot reach the implemented catalog.
- Retained red-before-green evidence and focused contract/loader/CLI regression tests.

## Explicit non-goals retained

- No provider fleet taxonomy mapping; every unclassified connector stays absent/non-capable.
- No PostgreSQL replication dependency, executor, provider connection, credential, or live call.
- No #3747 conformance fixtures, #3748 human manual/help/docs/website surfacing, or #3749
  generator enforcement.

## Verification

See `TDD-LEDGER.md` and `VERIFICATION.md` for the red evidence, full focused suite, executable
catalog/inspect output, and independent local gates.
