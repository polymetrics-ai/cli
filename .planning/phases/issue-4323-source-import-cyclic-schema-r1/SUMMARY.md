# Summary — issue 4323

The shared source importer now treats an OpenAPI schema-reference cycle as
declared missing-foundation evidence. It keeps the source declaration and
operation provenance, adds `cli-recursive-schema-foundation-r1` with the
canonical JSON Pointer, and marks affected operations merge-blocked. Unused
recursive component schemas use the existing top-level descriptor `gaps` path.

The behavior is covered through the real source-import path for direct,
mutual, and deeply nested cycles, with a finite-schema control. A public Grafana
artifact import retained 314 operations and reported 52 recursive-schema gaps.
No v3 source-lock fields or provenance fields changed.

Local validation: focused importer tests, `go vet ./...`, `go build ./cmd/pm`,
frozen GitHub byte/digest measurement, `git diff --check`, and full `make verify`
all passed.
