# Context — issue 4323 source-import cyclic schema gaps

## Decided scope

- Preserve valid recursive OpenAPI references as source-bound gap evidence through
  the existing missing-foundation path.
- Preserve operations and provenance; a recursive operation must remain present
  but cannot look fully supported.
- Keep the v3 multi-document source-lock design intact. The change is limited to
  the shared importer and its behavioral tests.
- Cycles are neither flattened nor expanded to an arbitrary depth.

## Inline GSD fallback

The adapter's generated lifecycle prompts were resolved and followed inline.
Compatible isolated GSD role runtime is not available in this delivery session,
and repository policy requires a single inline delivery worker.
