# Mutation-candidate lifecycle Slice 0 — plan

**Issue link:** Refs #4015

## Scope

Extend the landed `connectorgen certification-candidates` generic projection; do not create a parallel command or a second candidate model. The target connector definition is GitHub. Shared Go owns only declaration-independent candidate projection, classification data types, validation, and deterministic reporting. GitHub-owned membership, matching/evidence, and any exact manual override live in `internal/connectors/defs/github/certification.json`.

## TDD execution plan

1. **RED — mutation projection inventory.** Add generator tests that load the current GitHub declared surface and require exactly 856 fixture-required mutation candidates: 279 `direct_write` and 577 `reverse_etl`. Assert the projection derives command path, declaration reference, executor, input slots, required flags, credential flag, address, and fixture provenance.
2. **GREEN — generic mutation projection.** Generalize the landed direct-read candidate projection through intent-specific declaration adapters. Do not add provider literals or execute an executor. Regenerate the candidate/sweep artifact only through its canonical generator.
3. **RED — exhaustive fail-closed classification.** Add tests that require every generated mutation row to resolve to exactly one classification, total the buckets to 856, and reject missing, duplicate, broad, and unknown matches.
4. **GREEN — connector-owned classification declarations.** Add the GitHub classification catalog and a closed classification vocabulary. Classify `contained`, `real_money`, `real_people`, `public_visibility`, `third_party_scope`, or `unassessed`; code must use the latter for an unresolved declaration. Preserve concrete evidence in the emitted report.
5. **RED/GREEN — classifier refusal demonstration.** Use synthetic declarations in table-driven tests to verify paid seats, an outside invitation, public publication, and a third-party target cannot receive `contained`; separately prove a disposable-scope operation does receive it. This is a test-only failure demonstration, not a provider operation.
6. **RED/GREEN — derived CRUD fixture provenance.** For REST `cli_surface` addresses, derive a collection by removing only a terminal path parameter; derive sibling POST/PUT provisioners and order collections by URL depth. A shared GraphQL transport, a command without `api_surface`, or a collection without a POST/PUT must be a named exception, never an authored fixture fallback.
7. **Generated artifacts and accounting.** Regenerate candidate and certification sweep surfaces, run the generator twice, and record byte stability. Show `279 + 577 = 856`, the fixture-provenance split, and the complete sweep total of `1,571` in the PR body.
8. **Verification and review.** Run the targeted tests, their consumers, no-I/O generator checks, binary build/help checks, required individual `make verify` gates, full `make verify`, inline security/code review, then write verification and review evidence.

## Scope fences

- Do not execute any mutation, direct write, reverse plan/preview/approval/execute, fixture/container creation, evidence publication, or credentialed command.
- Do not hand-author generated candidates at scale or edit a generated artifact.
- Do not change #4215’s accepted-evidence contract or duplicate #4214’s read projection.
- Do not weaken existing tests or add connector-specific names to shared Go or a boundary allowlist.
- No additional dependency is authorized.

## Verification plan

- `go test -timeout 20m ./cmd/connectorgen`
- `go test -timeout 20m ./internal/connectors/certify ./internal/cli`
- `go build ./cmd/pm`
- canonical candidate/sweep generator and `--check`, each run twice with an empty second diff
- `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, and `make release-workflow-check`
- `make verify` when the local command budget permits; otherwise record the exact unrun command and reason before completion

## Dependency gate

Before code changes, read back PR #4214. If it is not merged into the required base, append the blocking status and stop; the only authorized work before it lands is this planning evidence.
