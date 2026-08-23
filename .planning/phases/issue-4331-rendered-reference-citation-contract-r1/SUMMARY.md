# Summary — Issue 4331

Implemented the source-lock v3 document contract for `rendered_reference`, `bundle`, and explicit `unavailable` documents while retaining absent-kind OpenAPI behavior.

- Rendered references require captured evidence, normalized media type, coverage confidence, and same-origin operation citations. Their bytes must not be a standalone OpenAPI/Swagger description, while structured JSON/YAML and OpenAPI path fragments remain valid.
- Bundles use ZIP or gzip archive integrity and the already declared operation inventory; extraction is intentionally deferred to a separate parser foundation.
- Hash-pinned rendered pages and bundles may retain an evidence-only document with no operations; every operation still takes the existing identity, route, deduplication, count, and (for rendered pages) citation checks.
- Unavailable declarations require a source-traced reason and coverage confidence, produce a blocking source-projection finding, and never trigger an artifact fetch. When an unavailable response is captured, its complete bytes and hash remain validated; absence of a usable provider document does not require fabricating such bytes.
- Source import never fetches cited URLs.

The inline/manual GSD fallback was used because compatible Pi workers are unavailable. Red/green/refactor evidence, package tests, and repository gates are recorded in the phase ledger and verification files.
