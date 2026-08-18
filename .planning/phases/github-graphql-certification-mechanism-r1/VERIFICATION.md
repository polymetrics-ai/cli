# Verification checklist — GraphQL certification mechanism

- [x] Required base branch fetched and branch verified at its exact current SHA before edits.
- [x] Required GSD command sources resolved; inline/manual fallback documented because spawned roles are forbidden for this direct-PR delivery.
- [x] Required Go skills loaded and recorded.
- [x] 305 GraphQL command classification is mutually exclusive and sums exactly to 305: 29 schema-conformant, 2 live-required, 274 fixture-bound.
- [x] Schema compilation and a deliberately broken post-compilation assertion produce a recorded red test.
- [x] Restored assertion and declared-unexecutable capability are tested; every unexecuted GraphQL record is concrete and non-pass.
- [x] Two bounded read-only product-path probes assert provider-produced values; the report records both as pass while the other 303 rows remain non-pass.
- [x] `go test -timeout 20m ./internal/connectors/certify` passes.
- [x] `go test -timeout 20m ./internal/connectors/engine` passes.
- [x] `go test -timeout 20m ./cmd/connectorgen` passes.
- [x] `go test -timeout 20m ./internal/cli` passes.
- [x] `go vet ./...`, `go build ./cmd/pm`, `connectorgen validate`, `surface-sync --check`, and `boundary` pass.
- [x] Repository generated-file, lint, docs, and smoke gates pass; website docs generator is byte-stable across two runs.
- [x] GSD verify-work and inline code-review complete; `REVIEW.md` records no findings. PR will name happy, bad, and edge cases and API-readback will confirm the base ref.
