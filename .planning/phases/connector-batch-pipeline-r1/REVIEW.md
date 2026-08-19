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
5. The current main permits a non-redacting direct-write policy, but the old
   gate would also admit redacting declarations. **Fixed after the rebase:** a
   red-first test demonstrated that `json_redacted` was included; `batch gate`
   now records it as an `output_policy` drop and also rejects `redact_fields`
   and legacy repository-content policies.
6. A post-main re-gate request suggested promoting 21 direct reads and six
   document writes. **Verified without an unsafe promotion:** the actual
   runner policy table permits `json`/`none` for direct writes only; direct
   reads remain redacting-only. The cited DocuSeal artifact declares the six
   document requests as `application/json`, not `multipart/form-data`, so the
   #3871 multipart capability is not their executable contract. The
   materializer and gate were rerun with zero hand edits to `api_surface`
   reasons; the executable count remains 39.
7. The new reachable connector namespaces changed root help. **Fixed:** the
   existing golden transcript test first failed, then regenerated its expected
   transcript through its documented update mode and passed unchanged.
8. A generic 30-connector capacity claim would misrepresent the Top-50
   cohort, whose provider surfaces range from 11 to 1,913 operations.
   **Fixed:** the batch documentation now has a source-cited size-tier map:
   800+ and 500+ providers are dedicated whole-provider branches, while only
   the smallest complete surfaces are grouped into conventional multi-provider
   batches. It retains whole-artifact materialization rather than inventing
   partial-operation exclusions.

No unresolved review finding remains. Shared schema, engine, and runner paths
were not edited.
