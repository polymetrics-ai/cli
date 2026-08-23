# Summary — Issue 4331

Implemented the source-lock v3 document contract for `rendered_reference`, `bundle`, and explicit `unavailable` documents while retaining absent-kind OpenAPI behavior.

- Rendered references require captured evidence, normalized media type, coverage confidence, and same-origin operation citations.
- Bundles use their archive integrity and the already declared operation inventory; extraction is intentionally deferred to a separate parser foundation.
- Unavailable evidence is reported as a blocking source-projection finding.
- Source import never fetches cited URLs.

The inline/manual GSD fallback was used because compatible Pi workers are unavailable. Red/green/refactor evidence, package tests, and repository gates are recorded in the phase ledger and verification files.
