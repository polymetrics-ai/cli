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

## Independent audit R1 gap-closure plan

1. **F1 — shared v3 evidence identity (red first).** Add a v3 operation-evidence workspace which transforms the checked-in GitHub operation lock into legal document-owned v3 inventory, including REST and GraphQL rows. Prove the pre-fix reader loses the v3 REST document rows; once fixed, it must use `parseSourceImportLock` for ordinary locks and preserve the same row set, all six classification lanes, and a declared plus deferred row. Preserve skipped/dynamic provider-absence projection as a deliberately separate evidence shape.
2. **F2 — bounded rendered citation (red first).** Extend rendered-reference fixtures so a generic same-origin publication URL fails. Define the smallest explicit capture/hash/location binding that can prove a non-fragment citation names the locked operation; reject a binding that differs from the document capture or operation location. Keep citation URLs provenance-only and do not fetch them. Read-only impact audit of `origin/fm/cli-map-batch67-r1` at `18248d233e6abd9d7ec03075a225cf35ee2f5399` found 861 generic rendered-operation citations across braintree (73), braze (95), copper (87), freshdesk (170), help-scout (144), salesloft (211), service-now (6), and zoho-bigin (75). Those locks must be repaired by their owners with fragments or bindings before source import/evidence will admit them; this foundation change does not rewrite or re-pin them.
3. **F3 — response evidence (red first).** Add retain HTTP fixtures carrying MIME headers and plausible-size HTML login/error bodies. Before identity verification, reject a structured lock delivered as HTML or an invalid/bad MIME response, and reject a generic HTML login/error signature. Keep a genuine expected HTML documentation fixture green and preserve byte/canonical verification for credible content.
4. Refactor the retain fetch path minimally to return raw bytes plus response media type only to retain classification. Existing hermetic byte-only fetchers remain supported without a default MIME claim; source import remains byte-only and strict.
5. Run focused red tests before production changes. Implement F1–F3, then run the focused package suite, the operation-evidence fixed-100 check, source projection/check gates, and regenerated artifacts in a clean tracked worktree. Record exact commands and outcomes in `TDD-LEDGER.md` and `VERIFICATION.md`.
6. Commit the coherent repair, push the existing PR branch, re-read its `main` base through the GitHub API, and request a fresh independent audit. Do not merge.

## Security and correctness boundaries

- Source URLs remain connector-owned, HTTPS/public-address validated, query-bounded, and never accepted from command arguments.
- Only a source response that passes wrong-source classification and the selected lock identity can be persisted.
- Retention validates raw byte size/digest for byte identity and a key-sorted JSON SHA-256 for canonical JSON identity. The raw digest/count of every stored file remains recorded even when it was not the selected comparison identity.
- File writes remain non-symlink, atomic, fsynced, content-addressed operations. Existing valid files are never replaced.
- Operation evidence is an offline evidence consumer: its normal lock decoder must not add fetching, credential use, parity-lock admission, or a second v3 identity model.
- Citation specificity is a source-lock admission condition. A generic rendered publication cannot become an exact operation claim merely because it shares an origin with a captured reference.
- MIME/body classification may only turn a response into wrong-source before persistence; it must never make a mismatch look verified or weaken the locked byte/canonical identity check.
