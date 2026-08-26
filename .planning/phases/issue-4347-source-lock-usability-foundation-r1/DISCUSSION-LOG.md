# Discussion log — issue 4347 source-lock usability

## Inputs resolved

- The launch brief required `connector-lane-build-order` and `data/SOURCE-LOCK-STATUS.md`; Firstmate clarified in inbox `001.msg` that these live in its environment and supplied the complete live status inline. The unavailable skill is not loaded by this worker.
- The existing code retains only `<connector>-operation-source-lock.json` through `loadConnectorSourceImportLock`. It runs the full import parser, including v3 aggregate OpenAPI validation, before it has fetched a document. That creates the retain/import circular dependency.
- The local worktree contains all 21 source locks. Twenty are ignored `<connector>-parity-source-lock.json` files with a minimal `rest.{source_url,sha256,bytes}` record; GitHub has the operation lock. The eight exact locks named in the brief are among those files.
- Firstmate reproduced Google Calendar Discovery JSON with unequal SHA-256 values and equal parsed JSON. The selected behaviour is canonical key-sorted JSON identity for locks that expressly select it; byte identity remains the default for static source artifacts.

## Decisions carried into the plan

1. Keep import validation strict and move only retention to a minimal, retain-specific loader. A retain lock needs connector ownership, bounded public URL, a positive byte observation, a valid digest/identity declaration, and discoverable artifacts; it must not require parseable OpenAPI, declared form, operation inventory, or aggregate version coverage.
2. Treat `identity: "byte"` as the legacy/default mode and add the explicit `identity: "canonical_json"` plus canonical digest to select semantic JSON verification. The manifest records source form/version when detectable, otherwise `undetermined`, and records fetched raw byte digest/count independently.
3. Classify HTTP denial, redirect, login-like HTML, and a drastic size collapse as `wrong source` before identity comparison. Byte/canonical mismatch after a credible response remains drift.
4. Retain only writes digest-addressed raw bytes and an append-safe provenance manifest. It never edits a source lock or silently repins it.

## Independent audit R1 — 2026-08-26

The independent audit failed the already-green PR at `f9d237840ed49737ef92667b7e2173465dcc000d`. Its three findings are accepted as correctness gaps, not review style preferences:

1. `operation-evidence` had its own permissive raw-lock model and read `rest.operations`, so it could silently omit the document-owned `rest.source_documents[].operations` inventory that strict v3 import uses. Normal operation locks must now decode through `parseSourceImportLock`; provider-evidenced skipped/dynamic locks retain their deliberately separate absence projection.
2. A same-origin rendered publication URL was accepted as an exact operation citation even where it pointed only at the generic reference. A rendered operation now needs either a non-empty URL fragment or an explicit binding to the hash-pinned capture and the operation's locked extraction location.
3. The retain transport discarded HTTP `Content-Type` before source classification. A plausible-size HTML login/error page could therefore reach digest comparison and be described as drift. The retain-only transport result retains MIME/body evidence until classification; known structured sources reject HTML/bad MIME, and generic HTML is rejected only for credible login/error signatures so legitimate documentation remains retainable.

This is an inline/manual GSD gap closure: compatible role spawning remains prohibited. The red/green rows are added before production changes.

## Independent exact-SHA audit — R2, 2026-08-26

The exact-SHA audit reopened the foundation with three coupled admission and
classification bypasses. They are handled as one red-first repair:

1. A v3 lock with a populated document-owned inventory and a legacy-looking
   `state: dynamic` could be projected as provider absence before its strict
   decoder ran, silently suppressing REST rows. Only an absence-shaped lock
   with no v3 document inventory may use the evidence-only absence projection;
   a populated v3 lock must proceed into strict parsing and be rejected if its
   wire shape is contradictory.
2. `canonical_json` itself proves that the expected artifact is JSON. A
   `text/html` HTTP response must therefore be wrong-source before canonical
   comparison, even when an older lock has no OpenAPI form pin or v3 document
   content type.
3. A rendered citation fragment is evidence only when it names the locked
   extraction location. Any supplied capture binding is a constraint, not an
   optional fallback: it must be validated even when the URL has a fragment.

The repair does not fetch citations, relax source-import admission, re-pin a
lock, or touch ignored retained artifact paths.

4. Canonical JSON cannot be defined by a decoder that silently chooses the last
   duplicate object member. Duplicate-member JSON is ambiguous source content,
   not a canonical match and not ordinary document drift. The canonical parser
   must reject it recursively before hashing; retain reports it as bad source
   before comparison, while offline import reports an invalid canonical
   artifact rather than a refresh/re-pin request.
5. For `canonical_json`, strict unambiguous JSON plus its canonical digest is
   the identity. Locked raw byte count/SHA-256 and a retained file's raw
   byte count/SHA-256 remain observable provenance only: whitespace or
   minification may change them without turning a canonical match into drift.
   Byte-identity locks retain their exact raw-size and raw-digest requirement.
