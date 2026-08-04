# Test plan — destructive-write confirmation foundation

All tests use in-memory records, temporary project state, `httptest`, or bundle fixtures. No live
provider calls or credential values are permitted.

## Strict TDD slices

1. **Plan/declaration** — red proves the typed confirmation object is unsupported and destructive
   intent is not normalized for future operations; green adds the closed schema and normalized
   target policy.
2. **Preview** — red proves a destructive stored plan can execute without a prior preview; green
   persists the engine preview digest and refuses execution before preview.
3. **Explicit approval** — red proves direct destructive engine execution accepts no typed approval
   evidence; green requires plan/hash/approval timestamp and typed confirmation.
4. **Execute seam** — red proves no generic pre-dispatch wrapper exists for a future `rest_write`
   executor; green demonstrates the same wrapper blocks the callback until evidence is complete and
   then invokes it exactly once.
5. **Bypass resistance** — red proves the generic reverse path/direct engine path can bypass the
   full lifecycle; green covers connector-command, bulk reverse ETL, state tampering, digest
   mismatch, and `batchable: false` preservation.

## Verification

Focused packages: `internal/connectors/engine`, `internal/connectors/commandrunner`, `internal/app`,
`internal/cli`, and `internal/connectors/conformance`. Then run the repository's separated local
gates from `AGENTS.md`; CI carries the full 550+ connector suite.
