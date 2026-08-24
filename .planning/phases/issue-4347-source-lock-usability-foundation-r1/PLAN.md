# Plan — issue 4347 source-lock usability

## Scope and ownership

Shared `cmd/connectorgen` source-lock retention foundation for issue #4347. Owned production paths are `cmd/connectorgen/sourceartifact.go`, `cmd/connectorgen/sourceartifact_test.go`, the narrow shared source-artifact identity helpers in `cmd/connectorgen/sourceimport.go` and their tests, plus applicable `docs/migration/**` command guidance. No individual connector bundle or ignored `sources/` lock is modified.

## TDD execution plan

1. Add red tests covering the acceptance table:
   - a v3 source-retain fixture with an absent `artifact.openapi` / aggregate inventory retains successfully and leaves the lock byte-for-byte unchanged;
   - a `*-parity-source-lock.json` fixture retains without being parsed as an operation-import lock;
   - two reordered JSON documents share a canonical digest but differ in byte digest, and a `canonical_json` lock accepts both while a byte lock rejects the second;
   - HTTP 403 and a dramatically undersized page produce a `wrong source` condition rather than a refresh/drift condition.
2. Run the focused tests and save the failing output as the `Red:` evidence in `TDD-LEDGER.md`.
3. Implement a retain-only lock loader and source-identity verification. Keep `parseSourceImportLock` and import inventory/form checks intact. Make raw byte origin durable in the retained manifest and use it for content-addressed paths.
4. Add source-response classification before identity comparison. Keep redirect refusal and public-destination checks; carry their failure as an actionable wrong-source error. Do not add network calls to import/verification.
5. Update `source-retain` help and migration conventions with identity modes, form/version recording, wrong-source versus drift, and explicit re-pin semantics. Update focused help/docs assertions.
6. Re-run all targeted tests until green, record `Green:` evidence, then run the built binary against the eight exact locks one connector at a time. Preserve the untracked source locks and record command results outside the tracked code change.
7. Complete the focused verification gates, inspect generated/snapshot drift, execute inline `verify-work` and `code-review`, fix any findings, then commit/push/open a `main`-targeted PR.

## Security and correctness boundaries

- Source URLs remain connector-owned, HTTPS/public-address validated, query-bounded, and never accepted from command arguments.
- Only a source response that passes wrong-source classification and the selected lock identity can be persisted.
- Retention validates raw byte size/digest for byte identity and a key-sorted JSON SHA-256 for canonical JSON identity. The raw digest/count of every stored file remains recorded even when it was not the selected comparison identity.
- File writes remain non-symlink, atomic, fsynced, content-addressed operations. Existing valid files are never replaced.
