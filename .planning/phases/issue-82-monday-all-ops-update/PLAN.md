# Plan — issue #82 Monday all-ops update

## Objective

Update Monday parity from a safe executable subset to full official operation coverage per the updated `PI_CONNECTOR_PROMPT.md`: every canonical Monday GraphQL operation must have an app-level mapping instead of remaining blocked only because it was not modeled in the prior slice.

## GSD mode

- `scripts/gsd prompt plan-phase issue-82-monday-all-ops-update --skip-research` generated the plan prompt.
- `programming-loop` remains unavailable in this registry; manual GSD/TDD fallback active.

## Safety interpretation

Full coverage does not permit raw GraphQL, raw HTTP, generic JSON writes, credentialed live checks, or unapproved destructive execution. The green target is:

- all 367 official operations are mapped to streams, direct reads, or named reverse-ETL write actions;
- live mutation dispatch remains gated by reverse ETL plan/preview/approval/execute and the Monday hook blocks unverified generic mutation execution until per-action bodies are hardened;
- no secrets or live Monday calls are introduced.

## Slice

1. Red tests: assert no Monday api_surface row remains blocked merely as unmodeled metadata; all 87 queries are stream/direct-read covered and all 280 mutations have named write actions.
2. Green: generate full coverage mappings, write actions, command metadata, and a Monday WriteHook safety gate.
3. Verify connectorgen validation and targeted tests, then rerun local gates if changes are substantial.

## Verification

```bash
go test ./cmd/connectorgen -run 'TestMondayFullSurfaceAllOpsCovered' -count=1
go test ./internal/connectors/hooks/monday -run 'TestMondayWriteHookBlocksModeledMutations' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```
