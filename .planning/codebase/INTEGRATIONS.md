# Integrations

**Generated via:** `scripts/gsd prompt map-codebase --fast` and `scripts/gsd prompt docs-update .planning AGENTS.md .agents --planning-only`.

## Internal Product Integrations

- Connector registry and runtime integrate declarative bundles, hooks, native connectors, conformance, certification, ETL, and reverse ETL.
- Local warehouse/query flows integrate source reads with local storage and SQL-style querying.
- Reverse ETL integrates warehouse rows with product-specific write actions using plan, preview, approval, execute.
- Runtime-backed execution is optional and controlled by `scripts/runtime.sh`.

## Optional Runtime / RLM / Pi Agent Integrations

Canonical docs:

- `docs/architecture/runtime-dependencies.md`
- `docs/runtime/SETUP.md`
- `docs/cli/runtime.md`
- `docs/cli/rlm.md`
- `docs/cli/perf.md`
- `docs/cli/agent.md`
- `.agents/agentic-delivery/references/runtime-rlm-website-integration.md`

The default CLI path remains dependency-free. Runtime-backed services are opt-in for integration tests, runtime perf comparisons, RLM agent mode, and external orchestration experiments.

| Integration | Role | Default local endpoint |
|---|---|---|
| Podman / Docker Compose | Local runtime service orchestration. Prefer `podman compose`, then `podman-compose`, then Docker Compose fallbacks. | local machine |
| PostgreSQL | Durable control-plane data, run ledger, plans/approvals/audit events, checkpoints, integration-test tables. | `localhost:15433` |
| DragonflyDB / Redis-compatible layer | Short-lived leases, retry coordination, rate counters, workflow hints, batch pointers, ephemeral agent locks, cheap caches. | `localhost:6379` |
| Temporal | Durable workflows for long ETL/reverse ETL, approval waits, retries, cancellation/resume, per-connector worker isolation, RLM agent mode. | `localhost:7233` |
| Temporal UI | Local workflow visibility. | `http://localhost:8080` |

RLM modes:

- `deterministic` and `fixture` are dependency-free.
- `model` is opt-in model-backed behavior.
- `agent` is opt-in runtime-backed behavior using Temporal and a Podman-managed agent image.

Relevant commands:

```bash
pm runtime doctor --json
pm perf compare --iterations 25 --runtime --json
pm rlm run --spec spec.json --in customers --out scored_customers --mode agent --request "score leads" --json
pm agent image build --json
pm agent image pull --json
pm agent image ensure --json
pm worker status --json
pm worker serve --json
```

The worker serves typed RLM Temporal workflows. It is not a generic remote command runner.

## Website Integration

Canonical docs:

- `website/content/docs/architecture.mdx`
- `website/content/docs/cli-reference.mdx`
- `website/package.json`

Website stack:

- Next.js 16 and React 19.
- Fumadocs for MDX documentation.
- Radix UI, Lucide icons, Tailwind CSS v4 tooling.
- Generated data scripts for docs, connector bundles, connector catalog, and connector pages.

Relevant checks when website files change:

```bash
cd website
npm run gen:website-data
npm run typecheck
npm run test:unit
npm run test:e2e
npm run build
```

Use only relevant website checks for the issue. Do not add frontend dependencies without human approval.

## External Service Boundaries

- Live connector credentials are not used for issue #122.
- Missing live credentials in certification should be `uncertified`, not failure.
- Destructive/admin/elevated external actions are human-gated.
- No secrets are requested, printed, summarized, or stored in planning artifacts.

## GSD / Pi Integration

- Official GSD source: `.gsd/upstream.lock.json`.
- Official docs snapshot: `.gsd/official-docs/`.
- Command registry: `.gsd/commands.json`.
- Shell adapter: `scripts/gsd`.
- Pi extension: `.pi/extensions/gsd/index.ts`.
- Pi skill: `.pi/skills/gsd-core/SKILL.md`.
- Pi prompt fallback: `.pi/prompts/gsd.md`.

Interactive Pi command examples:

```text
/gsd doctor
/gsd list
/gsd map-codebase --fast
/gsd plan-phase 1 --skip-research
/gsd-programming-loop init --phase <phase> --dry-run
```

Shell/non-interactive equivalents:

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt map-codebase --fast
scripts/gsd prompt plan-phase 1 --skip-research
scripts/gsd prompt programming-loop init --phase <phase> --dry-run
```

## Review Integrations

- CodeRabbit automatic review is primary for PR review.
- Copilot review is fallback-only when CodeRabbit is blocked, skipped, unavailable, disabled, paused, or rate-limited.
- Human approval remains required for parent PR merge to `main` and other human gates.

---
*Integrations refreshed: 2026-07-08 via repo-local official GSD Core Pi adapter.*
