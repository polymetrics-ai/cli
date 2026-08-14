# Plan: issue #4087 legacy sync-mode bypass

**Mode:** TDD

## Scope

Close only the two non-canonical execution bypasses for the existing public compatibility names. Do not rename, remove, deprecate, or document new connector-specific behavior.

## Slice 1 — red contract and execution regression tests

1. Add a table-driven test for both affected aliases, through both direct and persisted-legacy parsing.
2. Assert each parsed result has its expected non-empty canonical `ContractMode`, is admitted as typed, and preserves its public normalized name.
3. Run each alias through the ordinary source-to-warehouse fixture. Assert `*synccontract.ModeNotExecutableError` before the source `Read` method is called. This is the current typed pre-I/O outcome in the absence of a registered closed transport.
4. Add an explicit control table for every closed canonical mode name. Assert the current source, destination, contract, and typed-admission values remain unchanged.

**Red:** the affected-alias test fails on the base because each alias has an empty `ContractMode` and reaches legacy execution.

## Slice 2 — single-source generic mapping

1. Replace duplicated normal/persisted-legacy mode construction with one connector-neutral mapping authority in `internal/app/sync_modes.go`.
2. Set the two deduped legacy aliases to `ModeFullOverwrite` and `ModeIncrementalDedupe`, respectively, and mark only those aliases for typed admission.
3. Keep the other legacy adapters and all canonical closed mode entries behaviorally identical.

**Green:** the new contract and pre-I/O tests pass, alongside existing sync and transport tests.

## Slice 3 — parity and verification

1. Build `pm` and confirm relevant help text still exposes the unchanged compatibility names where the current surface does so.
2. Run the generated-surface check; regenerate only if its generator reports a changed artifact.
3. Run focused package tests, formatting, vet, build, and the individual repository verification gates prescribed by `AGENTS.md`.
4. Complete the inline/manual GSD verification and code review records. Commit the plan checkpoint and the green implementation slice separately.

## Guardrails

- No connector name/literal outside `internal/connectors/defs/<connector>/` is introduced.
- No credentialed checks, external writes, dependencies, generated-file hand edits, or documentation churn.
- The public spellings and their normalized names remain stable.
- No surrounding ETL/transport refactor.
