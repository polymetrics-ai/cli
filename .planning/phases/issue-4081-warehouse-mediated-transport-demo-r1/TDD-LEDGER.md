# #4081 — TDD ledger

**Status:** planning checkpoint ready; base admission is squash-aware content
identity at `e7d2b2963fc1dd164f63b31fccb8a3bab8084bec`.

| Slice | Red evidence required | Green evidence required | Refactor/preservation | Status |
| --- | --- | --- | --- | --- |
| Production construction | `App.Open` has unavailable verifier/empty registry/nil stage and rejects before GitHub I/O | Explicit non-nil stage/verifier/exact GitHub roles admit only declared route | Defaults and typed-nil rejection stay fail closed | Planned |
| Durable handle/reopen | Raw page/in-memory records can reach destination or tampered owner/hash can reopen | Stage returns only after durable WAL/DuckDB/Parquet receipt; reopen revalidates and reconstructs bounded copies | Structural owner paths, #4079 copies, no caller paths | Planned |
| Ack/checkpoint order | Destination failure advances checkpoint; checkpoint persistence failure loses/replaces workset | Receipt + independent read-back precede CAS; failed CAS replays identical handle/idempotency key | Existing CAS/source identity semantics unchanged | Planned |
| Exact binary demo | Existing built binary cannot execute closed GitHub path or cleanup is not proven | Fresh binary runs faithful-server path with sanitized evidence/read-back/inverse/zero residue | Bounded resources, no secrets/raw payloads, cleanup repeat succeeds | Planned |

## Committed checkpoints required after base admission

1. **Plan:** phase artifacts only.
2. **Red:** tests plus phase evidence only; focused command must fail for dormant
   construction/durable reopen/ordering behavior, not compilation.
3. **Green:** smallest construction, stage/reopen adapter, exact GitHub adapters,
   and demo harness; targeted tests and exact binary proof pass.
4. **Refactor/review:** formatting, review dispositions, and in-scope pipeline
   fixes only.

No RED or GREEN commit exists yet. The planning checkpoint must commit before
the first honest reproduction on the admitted combined head.
