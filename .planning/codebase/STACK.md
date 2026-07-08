# Stack

**Generated via:** `scripts/gsd prompt map-codebase --fast` through the official GSD Core Pi adapter
**Upstream GSD Core:** `open-gsd/gsd-core@20297a8ff941378b8615a5d3e8629e52c10a0f9d`

## Language and Runtime

- Go CLI monolith.
- Node.js is used for repo-local planning/tooling adapter `scripts/gsd`, Pi extension resources, and the website.
- Optional runtime-backed execution uses project runtime scripts; runtime services are not required for issue #122.
- Runtime/RLM/Pi-agent integration knowledge is summarized in `.agents/agentic-delivery/references/runtime-rlm-website-integration.md` and sourced from `docs/architecture/runtime-dependencies.md` plus `docs/runtime/SETUP.md`.

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

## Optional Runtime Services

- Podman-first local orchestration with Docker Compose fallback.
- PostgreSQL for durable control-plane data, run ledgers, plans/approvals/audit events, checkpoints, and integration-test tables.
- DragonflyDB / Redis-compatible layer for leases, retry coordination, rate counters, workflow hints, batch pointers, ephemeral agent locks, and cheap caches.
- Temporal for durable workflows, long-running ETL/reverse ETL, approval waits, retries, cancellation/resume, per-connector worker isolation, and RLM agent mode.
- Runtime endpoints: PostgreSQL `localhost:15433`, DragonflyDB `localhost:6379`, Temporal `localhost:7233`, Temporal UI `http://localhost:8080`.

## Website Stack

- Next.js 16, React 19, Fumadocs, Radix UI, Lucide icons, Tailwind CSS v4 tooling.
- Website docs live under `website/content/docs/**`.
- Website generated data scripts live in `website/package.json`.
- Runtime/RLM website contract appears in `website/content/docs/architecture.mdx` and `website/content/docs/cli-reference.mdx`.

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
