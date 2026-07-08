# Integrations

**Generated via:** `scripts/gsd prompt map-codebase --fast` and `scripts/gsd prompt docs-update .planning AGENTS.md .agents --planning-only`.

## Internal Product Integrations

- Connector registry and runtime integrate declarative bundles, hooks, native connectors, conformance, certification, ETL, and reverse ETL.
- Local warehouse/query flows integrate source reads with local storage and SQL-style querying.
- Reverse ETL integrates warehouse rows with product-specific write actions using plan, preview, approval, execute.
- Runtime-backed execution is optional and controlled by `scripts/runtime.sh`.

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
