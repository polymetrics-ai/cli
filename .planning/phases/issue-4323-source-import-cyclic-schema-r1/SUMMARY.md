# Summary — issue 4323

The shared source importer now treats an OpenAPI schema-reference cycle as
declared missing-foundation evidence. It keeps the source declaration and
operation provenance, adds `cli-recursive-schema-foundation-r1` with the
canonical JSON Pointer, and marks affected operations merge-blocked. Unused
recursive component schemas use the existing top-level descriptor `gaps` path.

It also retains bounded OpenAPI 3.0 reference siblings. `description`,
`summary`, schema `readOnly`, and schema `type` are the only added forms. An
equivalent `type` is ordinary retained source data; a non-equivalent `type` is
retained exactly but creates a source-traced,
`cli-openapi30-reference-sibling-foundation-r1` merge-blocking gap. Other
semantic siblings, such as a response `content` override, remain grammar
errors. Preflight-only component evidence is preserved in the same top-level
descriptor-gap path rather than silently disappearing.

The behavior is covered through the real source-import path for direct,
mutual, and deeply nested cycles, with a finite-schema control. A public Grafana
artifact import retained 314 operations and reported 52 recursive-schema gaps.
Asana's pinned public source imports and verifies all 249 operations; its two
non-equivalent type overlays are explicit gaps. GitLab live `master` has
drifted from its strict historical lock, and Docker Hub's pinned provider
artifact has a separate dangling schema-to-response reference; neither source
was changed or suppressed. No v3 source-lock fields or provenance fields
changed.

Local validation: focused importer tests, `go vet ./...`, `go build ./cmd/pm`,
frozen GitHub byte/digest measurement, `git diff --check`, and final full
`make verify` all passed. The final full suite ran with the unchanged shared
20-minute budget; `cmd/connectorgen`, `internal/cli`, and connector-boundary
passed in 189.173s, 528.045s, and 406.747s respectively.
