# GSD Command / Workflow Log — Issue #122

## Official source

This repo now uses a project-local Pi adapter based on the official GSD Core documentation:

```text
https://github.com/open-gsd/gsd-core/blob/next/docs/README.md
```

Pinned upstream source:

```json
{
  "source": "github.com/open-gsd/gsd-core",
  "ref": "next",
  "commit": "20297a8ff941378b8615a5d3e8629e52c10a0f9d",
  "runtime_adapter": "pi-project-local"
}
```

The official docs do not currently list Pi as a supported runtime in `docs/how-to/install-on-your-runtime.md`, so this repository provides a project-local Pi adapter rather than claiming upstream Pi support.

## Pi adapter

Repo-local files:

- `scripts/gsd` — shell/runtime-neutral adapter.
- `.gsd/upstream.lock.json` — official source lock.
- `.gsd/commands.json` — generated command registry from official `docs/COMMANDS.md`.
- `.gsd/official-docs/` — official docs snapshot used by the adapter.
- `.gsd/prompts/issue-122-rebootstrap.md` — canonical repo-specific onboarding prompt.
- `.pi/settings.json` — project-local Pi resource loading.
- `.pi/extensions/gsd/index.ts` — Pi slash-command adapter (`/gsd`, `/gsd-plan-phase`, etc.).
- `.pi/prompts/gsd.md` — prompt-template fallback.
- `.pi/skills/gsd-core/SKILL.md` — project-local GSD skill for default planning/implementation behavior.

## Commands run

Initial red/preflight evidence:

```bash
git status --short --branch
node "$HOME/.claude/get-shit-done/bin/gsd-tools.cjs" init new-project
node "$HOME/.claude/get-shit-done/bin/gsd-tools.cjs" init map-codebase
find .planning/phases -maxdepth 1 -mindepth 1 -type d | wc -l
test -d .planning/codebase || echo NO_CODEBASE_MAP
git diff --name-only -- cmd internal
```

Official adapter setup/verification:

```bash
scripts/gsd sync-upstream
scripts/gsd doctor
scripts/gsd version
scripts/gsd list
scripts/gsd sources plan-phase
scripts/gsd sources issue-122-rebootstrap
scripts/gsd prompt issue-122-rebootstrap > .planning/traces/issue-122-gsd-onboarding-prompt.md
scripts/gsd prompt issue-122-rebootstrap >/tmp/issue122-regenerated.md
diff -u .planning/traces/issue-122-gsd-onboarding-prompt.md /tmp/issue122-regenerated.md
scripts/gsd prompt plan-phase 1 --skip-research >/tmp/gsd-plan-phase.md
scripts/gsd verify-pi
```

Equivalent official GSD flow encoded by `scripts/gsd prompt issue-122-rebootstrap`:

```text
/gsd-onboard or /gsd-map-codebase + /gsd-new-project for brownfield onboarding
/gsd-plan-phase 1 --skip-research
/gsd-programming-loop init --phase 01-inventory-reconciliation --dry-run
```

## Archive evidence

The previous active `.planning/` tree was archived outside active `.planning/` before replacement:

```text
../planning-archives/polymetrics-cli-issue-122-pre-rebootstrap-20260708173641.tar.gz
sha256 e0959e4c8eba6e8610255a0cd9a98b39267902ba19600515abfdab726bfd57f5
```

## Initial red evidence

- Existing custom `.planning/phases/` contained 26 phase directories.
- `.planning/codebase/` was absent.
- `gsd-tools init new-project` reported `project_exists: true`, `is_brownfield: true`, and `needs_codebase_map: true` before the rebootstrap.
- `git diff --name-only -- cmd internal` produced no output before planning edits.

## Scope guard

Issue #122 is planning/tooling-only. No `cmd/` or `internal/` source edits are allowed. The final source guard remains:

```bash
git diff --name-only -- cmd internal
```

Expected output: none.
