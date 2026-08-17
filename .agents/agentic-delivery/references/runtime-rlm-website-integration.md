# Runtime, RLM, Pi Agent, and Website Integration Knowledge

Use this reference for planning, implementation, review, and documentation work that touches optional runtime services, RLM/agent mode, local integration tests, or website documentation.

## Canonical source docs

Read these before changing runtime/RLM/website behavior or planning around it:

- `docs/architecture/runtime-dependencies.md`
- `docs/runtime/SETUP.md`
- `docs/cli/runtime.md`
- `docs/cli/rlm.md`
- `docs/cli/perf.md`
- `docs/cli/agent.md`
- `website/content/docs/architecture.mdx`
- `website/content/docs/cli-reference.mdx`
- `website/package.json`

## Runtime dependency topology

The default `pm` path remains dependency-free and local-first. Runtime-backed services are opt-in for integration tests, performance comparisons, RLM agent mode, and external orchestration experiments.

Runtime services:

| Service | Role | Default local endpoint |
|---|---|---|
| PostgreSQL | Durable control-plane data, run ledgers, reverse ETL plans/approvals/audit events, checkpoints, integration-test tables | `localhost:15433` |
| DragonflyDB / Redis-compatible layer | Short-lived leases, rate-limit counters, retry coordination, workflow hints, batch pointers, ephemeral agent locks, cheap caches | `localhost:6379` |
| Temporal | Durable orchestration for long-running ETL, reverse ETL approval waits, retries, cancellation/resume, per-connector worker isolation, RLM agent mode | `localhost:7233` |
| Temporal UI | Local workflow visibility | `http://localhost:8080` |

Runtime selection is Podman-first:

1. `podman compose`
2. `podman-compose`
3. `docker compose`
4. `docker-compose`

On macOS, Podman uses the `polymetrics-runtime` machine created by `scripts/setup-runtime-macos.sh`.

## RLM and Pi agent integration

RLM materializes scored records to the local warehouse.

Modes:

- `deterministic` — dependency-free.
- `fixture` — dependency-free fixture-backed behavior.
- `model` — opt-in model-backed behavior.
- `agent` — opt-in runtime-backed mode using Temporal and a Podman-managed agent image.

Relevant commands:

```bash
pm rlm run --spec spec.json --in customers --out scored_customers --mode deterministic --json
pm rlm run --spec spec.json --out scored_customers --mode fixture --json
pm rlm run --spec spec.json --in customers --out scored_customers --mode agent --request "score leads" --json
pm agent image build --json
pm agent image pull --json
pm agent image ensure --json
pm worker status --json
pm worker serve --json
```

The worker serves typed RLM Temporal workflows. It is not a generic remote command runner.

## Runtime commands and tests

Runtime control:

```bash
scripts/runtime.sh doctor
scripts/runtime.sh up
scripts/runtime.sh ps
scripts/runtime.sh logs
scripts/runtime.sh down
scripts/runtime.sh reset
```

Runtime CLI:

```bash
pm runtime doctor --json
pm perf compare --iterations 25 --json
pm perf compare --iterations 25 --runtime --json
```

Runtime integration tests are optional and explicitly gated:

```bash
scripts/runtime.sh up
POLYMETRICS_INTEGRATION=1 go test ./...
scripts/runtime.sh down
```

Do not make runtime services mandatory for default unit tests or dependency-free CLI workflows.

## Safety boundaries

- Never store plaintext credentials in PostgreSQL, DragonflyDB, Temporal workflow history, docs, logs, prompts, or PR bodies.
- DragonflyDB is coordination/cache only; durable truth belongs in PostgreSQL/local state, not Redis-compatible ephemeral state.
- Temporal workflow inputs and history must stay bounded and serializable; do not place secrets, raw row payloads, approval tokens, or large batches in workflow history.
- RLM output is data only. It does not send messages or mutate external systems.
- Runtime-backed checks are optional and should be skipped or marked not-run unless explicitly requested.
- New dependencies, production runtime changes, credentialed checks, and deployment changes remain human-gated.

## Website integration knowledge

Website stack:

- Next.js 16.
- React 19.
- Fumadocs (`fumadocs-core`, `fumadocs-mdx`, `fumadocs-ui`).
- Radix UI, Lucide icons, Tailwind CSS v4 tooling.
- Generated docs/data scripts in `website/package.json`.

Website commands:

```bash
cd website
npm run gen:website-data
npm run typecheck
npm run test:unit
npm run test:e2e
npm run build
```

Use only the relevant checks for a change. Do not add frontend dependencies without human approval.

Website parity:

- Runtime/RLM changes should update `website/content/docs/architecture.mdx` and `website/content/docs/cli-reference.mdx` when user-facing behavior changes.
- CLI docs parity still follows `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.
- Website UI/design work should load `frontend-design`, `web-design-guidelines`, `vercel-react-best-practices`, and `vercel-composition-patterns` as applicable.

## Planning rule

Do not copy all runtime docs into every prompt. Keep `.planning` to a concise integration summary and link to canonical docs. Add runtime/RLM/website details to GSD plans only when the issue touches those areas.
