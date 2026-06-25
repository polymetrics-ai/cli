# Test Plan

- Unit test all catalog entries have a native port plan.
- Unit test Postgres, MySQL, and MongoDB CDC classification.
- Unit test enabled GitHub remains wave 0 and references existing native connector.
- CLI test `pm connectors port-plan --all --json`.
- CLI test `pm connectors port-plan source-postgres` includes CDC and conformance sections.
- Docs validation test ensures catalog manuals include native port plan sections.
- Full local verification with `go test ./...`, `go vet ./...`, `go build`, docs validation, and `make verify`.
