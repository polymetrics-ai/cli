# Review — connector batch pipeline r1

Manual inline code review completed after the focused and package tests.

## Findings and disposition

1. `validatePath` treats a directory without `metadata.json` as a defs root,
   which could let a malformed candidate evade the intended per-bundle stage.
   **Fixed:** `batch gate` now turns that exact case into a named `validate`
   drop before calling the existing validator.
2. A direct-write surface-sync test would require the currently redacting
   command policy on this base. **Fixed without weakening the test:** it uses a
   binary-download operation, whose command has no output policy, to prove
   derived path-flag drift is dropped. No redaction declaration was added.
3. Runtime conformance must not be duplicated in the generator. **Verified:**
   `batchRuntimePreflight` calls exported `commandrunner.Preflight` for every
   implemented command; it has no copied executor-shape rules.
4. Batch failure must not hide sibling results. **Verified:** the loop builds a
   complete report before returning nonzero; tests retain an included sibling
   plus a named drop.

No unresolved review finding remains. Shared schema, engine, and runner paths
were not edited.
