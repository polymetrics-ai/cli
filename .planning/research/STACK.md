# Research: Stack

**Generated via:** official GSD Core Pi adapter command path.

## Implementation Stack

- Go CLI monolith.
- Embedded JSON connector bundles.
- Local warehouse/query and ETL/reverse ETL services.
- Optional runtime-backed execution.

## Planning Stack

- Official GSD Core docs snapshot pinned in `.gsd/official-docs/`.
- Repo-local command registry in `.gsd/commands.json`.
- Repo-local shell adapter in `scripts/gsd`.
- Pi extension/prompt/skill resources in `.pi/`.
- Agent-neutral contracts and YAML specs in `.agents/`.

## Verification Stack

- Go gates for source changes.
- `scripts/gsd doctor/version/list/verify-pi` for GSD adapter health.
- JSON/YAML parse checks for planning and agent resources.
- `git diff --name-only -- cmd internal` and `git diff --name-only -- .planning/phases` scope guards for issue #122.

---
*Stack research refreshed: 2026-07-08.*
