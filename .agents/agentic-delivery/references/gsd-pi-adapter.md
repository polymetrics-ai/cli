# GSD Core Pi Adapter Reference

This repository uses official GSD Core workflows through a project-local Pi adapter.

## Source of truth

- Official docs snapshot: `.gsd/official-docs/`
- Source lock: `.gsd/upstream.lock.json`
- Command registry: `.gsd/commands.json`
- Shell adapter: `scripts/gsd`
- Pi extension: `.pi/extensions/gsd/index.ts`
- Pi skill: `.pi/skills/gsd-core/SKILL.md`
- Pi prompt fallback: `.pi/prompts/gsd.md`

The official GSD docs do not currently list Pi as an upstream runtime. Treat Pi support as repo-local adapter behavior.

## Required command path

Before implementation or behavior-changing work:

```bash
scripts/gsd doctor
scripts/gsd prompt programming-loop init --phase <phase-or-issue> --dry-run
```

Before planning/roadmap/codebase work:

```bash
scripts/gsd doctor
scripts/gsd prompt map-codebase --fast
scripts/gsd prompt new-project --from-existing --non-interactive
scripts/gsd prompt plan-phase <phase> --skip-research
```

In Pi after project trust/reload, use the interactive equivalents:

```text
/gsd doctor
/gsd map-codebase --fast
/gsd new-project --from-existing --non-interactive
/gsd plan-phase <phase> --skip-research
/gsd-programming-loop init --phase <phase-or-issue> --dry-run
/gsd-code-review <phase-or-issue>
```

## Agent requirements

- Agents and subagents must prefer `.pi` GSD commands when running inside Pi.
- For CLI command, flag, output, connector surface, help-topic, manual, or website-doc changes, agents must also follow `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.
- Non-interactive or non-Pi runners must use `scripts/gsd prompt <command> [args...]` and then execute the generated prompt with their local tools.
- Manual-GSD fallback is allowed only when `scripts/gsd doctor` fails or Pi project resources are unavailable; record the fallback in the phase, planning trace, handoff, or PR body.
- Do not copy raw upstream `agents/` or `commands/` files into this repo as runtime commands; use adapter-generated prompts and registry entries.

## Safety overlay

- No secrets in prompts, logs, artifacts, or handoffs.
- No new dependencies without human approval.
- No credentialed connector checks unless explicitly requested.
- No reverse ETL execution without plan, preview, approval, execute.
- No generic shell, generic HTTP write, or generic SQL write tools.
- Stop for destructive/admin/elevated actions, auth-scope changes, production deploys, quality-gate reductions, and merges to `main`.
