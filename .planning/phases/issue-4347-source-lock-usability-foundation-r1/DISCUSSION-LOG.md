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
