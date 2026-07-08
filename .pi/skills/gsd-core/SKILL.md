---
name: gsd-core
description: Official GSD Core workflow adapter for Pi in this repo. Use for issue implementation, planning, milestones, roadmap work, phase discussion/planning/execution, TDD, verification, PR preparation, connector parity planning, and any task that mentions GSD.
---

# GSD Core for Pi

This repo uses a project-local Pi adapter for official GSD Core workflows.

## First steps

Before running GSD workflows in Pi:

```bash
scripts/gsd doctor
scripts/gsd list
```

Use the Pi slash command when interactive:

```text
/gsd plan-phase 1 --skip-research
/gsd-map-codebase --fast
/gsd-onboard --fast
```

Use shell prompt generation when non-interactive or when you need reproducible evidence:

```bash
scripts/gsd prompt plan-phase 1 --skip-research
scripts/gsd prompt issue-122-rebootstrap
```

## Required workflow for implementation work

For implementation or behavior-changing work:

1. Read the issue first.
2. Read `AGENTS.md`.
3. Run the relevant GSD command prompt through `scripts/gsd` or `/gsd`.
4. Create or update the GSD plan, TDD ledger, and verification checklist before production edits.
5. Capture red/green/refactor evidence for behavior changes.
6. Run targeted verification, then the issue's broader verification.
7. Commit coherent green checkpoints.
8. Open/update the PR with GSD evidence and `Closes #N` or `Refs #N` as appropriate.

## Safety rules

- Never request, print, store, or summarize secrets.
- Do not add dependencies without human approval.
- Do not run credentialed connector checks unless explicitly requested.
- Do not run reverse ETL execution without plan, preview, approval, execute.
- Do not expose generic shell, generic HTTP write, or generic SQL write tools.
- Destructive/admin/elevated-scope/dependency changes are human-gated.
- For planning-only work, do not edit `cmd/` or `internal/`.

## Official provenance

The adapter is based on official docs from:

```text
https://github.com/open-gsd/gsd-core/blob/next/docs/README.md
```

Use these commands to inspect the pinned upstream source:

```bash
scripts/gsd version
scripts/gsd sources plan-phase
```
