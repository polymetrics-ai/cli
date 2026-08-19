# Verification checklist — Twenty CRM

- [ ] GSD command resolution and inline fallback recorded.
- [ ] Recovery assessment precedes bundle import.
- [x] Targeted Twenty test red/green evidence recorded in `TDD-LEDGER.md`.
- [x] Focused structural generator and conformance checks: `go run ./cmd/connectorgen validate internal/connectors/defs/twenty`, `go run ./cmd/connectorgen surface-sync --check`, and `go test -timeout 20m ./internal/connectors/conformance -run 'TestConformance/twenty'`.
- [ ] Structural generator, runtime preflight, conformance, boundary, docs/golden,
  lint, build, and verification results recorded with exact commands.
- [ ] Built-binary live read, pagination, create, update, round-trip, delete,
  and post-delete absence evidence recorded without secret material.
- [ ] PR base is read back through GitHub API and equals `main`.
