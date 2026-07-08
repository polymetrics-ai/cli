# Stack

**Generated via:** `scripts/gsd prompt map-codebase --fast` through the official GSD Core Pi adapter
**Upstream GSD Core:** `open-gsd/gsd-core@20297a8ff941378b8615a5d3e8629e52c10a0f9d`

## Language and Runtime

- Go CLI monolith.
- Node.js is used for repo-local planning/tooling adapter `scripts/gsd` and Pi extension resources.
- Optional runtime-backed execution uses project runtime scripts; runtime services are not required for issue #122.

## Primary Product Surface

- CLI binary: `pm`.
- Main package: `cmd/pm`.
- Product domains: ETL, reverse ETL, connector inspection, credential management, local warehouse queries, scheduling, flow execution, and optional runtime-backed execution.

## Connector Architecture

- Declarative connector bundles: `internal/connectors/defs/<connector>/`.
- Runtime engine: `internal/connectors/engine/`.
- Hook escape hatches: `internal/connectors/hooks/`.
- Native connectors: `internal/connectors/native/`.
- Conformance/certification: `internal/connectors/conformance/`, `internal/connectors/certify/`.

## Planning and Agent Runtime

- Official GSD docs snapshot: `.gsd/official-docs/`.
- Official command registry: `.gsd/commands.json`.
- Source lock: `.gsd/upstream.lock.json`.
- Shell adapter: `scripts/gsd`.
- Pi settings/extension/prompt/skill: `.pi/`.
- Agent specs/contracts: `.agents/`.
- Active planning artifacts: `.planning/`.

## Verification Stack

Local gates from `AGENTS.md`:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

Planning-only issue #122 uses adapter/docs verification instead of Go gates unless Go source changes.

---
*Stack refreshed: 2026-07-08; phases intentionally unchanged.*
