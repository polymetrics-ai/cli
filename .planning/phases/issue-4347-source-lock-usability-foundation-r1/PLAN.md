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

## Independent exact-SHA audit R2 repair plan

1. **Red first — v3 inventory cannot become absence.** Add an operation-evidence fixture with the full legal GitHub document-owned v3 inventory plus a contradictory `state: dynamic`. The current evidence reader must demonstrate the bad absence projection; after the repair it must enter the strict v3 decoder and reject the contradictory wire shape rather than omit REST operations. Keep v3 skipped/dynamic records with no `source_documents` on their narrow absence path.
2. **Red first — canonical JSON MIME expectation.** Add a canonical-JSON retain fixture with no form pin and an HTTP `text/html` response containing otherwise parseable, mismatched JSON. It must report `wrong source`, never canonical identity drift; byte/generic HTML behavior stays unchanged.
3. **Red first — citation fragment and binding agree.** Add rendered-reference fixtures for an unrelated fragment and for a matching fragment carrying a contradictory capture binding. Both must fail source-lock parsing. A matching source-location fragment without a binding, and a no-fragment exact binding, remain valid.
4. **Red first — duplicate JSON members are ambiguous.** Add a canonical-JSON retain fixture with a duplicate key and prove it does not fall through to a canonical digest or ordinary refresh/drift message. Reject duplicate keys at every JSON object nesting level before canonical hashing; source retain reports wrong-source and offline identity verification reports invalid/ambiguous canonical input.
5. **Red first — canonical raw bytes are provenance.** Lock formatted JSON and serve its minified equivalent with a different byte count and raw SHA-256. `canonical_json` must retain and offline-read that file by the canonical digest while recording the served raw evidence; a byte-identity fixture remains strict.
6. Run the five focused tests before production edits, then make the smallest shared decoder/classifier changes. Re-run source retain, source import, and operation-evidence focused suites plus a clean tracked snapshot. Preserve all ignored retained artifacts, particularly Zendesk.

## Independent exact-SHA audit R3 help-contract repair plan

1. Reproduce the exact-head public-help regression before production edits:
   `TestSourceRetainHelpAndMigrationDocumentationDescribeIdentityAndWrongSource`
   must fail because its operator-facing phrase is split across a rendered
   newline.
2. Reflow only `sourceRetainUsage` so the existing, documentation-backed
   phrase `wrong source` is contiguous in `connectorgen source-retain --help`.
   Do not alter source identity, fetch classification, lock bytes, generated
   connector evidence, or Batch 6–7 citations.
3. Run the focused contract test, the complete changed `cmd/connectorgen`
   package, `go vet`, source-generation check commands, and a whitespace diff
   check. Record the exact repaired commit/remote head and request a fresh
   independent exact-head audit; no merge.

## Independent exact-SHA audit R4 duplicate-member repair plan

1. **Red first — retain lock acquisition is duplicate-free.** Add table-driven
   source-retain command fixtures for top-level, REST, GraphQL, and identity
   duplicate members, with the conflicting values in both orders. Each case
   must reject the lock as a duplicate-member ambiguity before the fetch seam
   is called; unknown forward-compatible fields remain tolerated.
2. **Red first — v3 inventory absence is duplicate-free.** Add both orders of
   a schema-v3 `rest.source_documents` duplicate around a dynamic-state
   absence fixture. Both must reject at the source-lock boundary with a
   duplicate-member error, never return `input.Absence`, and never rely on
   last-member-wins decoding.
3. Reuse the existing recursive JSON token validator through its permissive
   (unknown-field tolerant) decode path before either partial decoder makes an
   acquisition or provider-absence decision. Keep source import's stricter
   unknown-field and inventory validation unchanged.
4. Run the targeted red command before production edits. Implement the two
   fail-closed decode changes, then rerun focused source-retain and
   operation-evidence tests, the serial `cmd/connectorgen` package suite,
   generator checks, and a clean tracked snapshot where local ignored retained
   artifacts would otherwise affect generated-state checks. Record every
   command/result, commit, push, read the PR base back, and request a fresh
   exact-head audit; no merge.

## Security and correctness boundaries

- Source URLs remain connector-owned, HTTPS/public-address validated, query-bounded, and never accepted from command arguments.
- Only a source response that passes wrong-source classification and the selected lock identity can be persisted.
- Retention validates raw byte size/digest for byte identity and a key-sorted JSON SHA-256 for canonical JSON identity. The raw digest/count of every stored file remains recorded even when it was not the selected comparison identity.
- File writes remain non-symlink, atomic, fsynced, content-addressed operations. Existing valid files are never replaced.
- Operation evidence is an offline evidence consumer: its normal lock decoder must not add fetching, credential use, parity-lock admission, or a second v3 identity model.
- Citation specificity is a source-lock admission condition. A generic rendered publication cannot become an exact operation claim merely because it shares an origin with a captured reference.
- MIME/body classification may only turn a response into wrong-source before persistence; it must never make a mismatch look verified or weaken the locked byte/canonical identity check.
- A v3 provider-absence projection may not coexist with a non-empty document-owned inventory. Canonical JSON identity is a structured-JSON expectation for response MIME classification. A rendered citation fragment and any supplied capture binding must both agree with the locked operation evidence.
- Canonical JSON identity rejects duplicate object member names at every nesting level; it never assigns implementation-dependent last-member semantics to a source lock.
- Canonical JSON identity compares only strict, unambiguous canonical JSON. Its
  locked and retained raw digest/count are provenance; byte-identity locks
  remain exact raw-byte comparisons.
- Every partial lock reader must reject duplicate JSON members before it derives
  a fetch URL, selected identity, or provider-absence shape. Duplicate-member
  ambiguity is malformed input, not an order-dependent policy choice.
