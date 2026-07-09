# Summary — Issue #97 Linear CLI surface metadata

Status: verified green slice integrated directly on parent branch; draft parent PR #131 opened.

Delivered:

- Added `internal/connectors/defs/linear/cli_surface.json`.
- Added `TestLinearCLISurfaceMapsImplementedStreams` to prove embedded Linear metadata maps:
  - `issue list` → `issues`;
  - `team list` → `teams`;
  - `project list` → `projects`;
  - `user list` → `users`.
- Regenerated website connector data so Linear CLI metadata appears in generated website data.
- Kept only existing stream-backed commands executable. Direct reads, writes, admin/sensitive actions, and raw GraphQL remain planned, unsupported, or unsafe/disallowed.

Verification:

```bash
go test ./internal/connectors/engine -run TestLinearCLISurfaceMapsImplementedStreams -count=1
go test ./internal/connectors/engine ./internal/connectors/commandrunner -count=1
go test ./internal/connectors/conformance -run 'TestConformance/linear' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
npm --prefix website run gen:website-data
go vet ./...
go test ./...
go build ./cmd/pm
./pm docs validate --connectors-dir docs/connectors
make verify
git diff --check
```

All above passed. `npm --prefix website run test:unit` is blocked in this checkout because local website dependencies are not installed (`vitest: command not found`).
