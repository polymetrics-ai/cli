# Summary: Chatwoot CLI Parity Parent Orchestration

Status: in progress.

## Done

- Read required repo, GSD, parent-orchestration, review-routing, CLI parity, connector migration, and Go skill references.
- Confirmed issue #148 and sub-issues #149-#155 are open.
- Confirmed parent PR for `feat/148-chatwoot-cli-parity` -> `main` is missing.
- Recorded manual GSD fallback because `scripts/gsd prompt programming-loop ...` is not available in the repo-local adapter registry.
- Created parent planning, TDD, verification, run-state, and orchestration-state artifacts.
- Recorded runtime fanout blocker: current Pi API tool surface lacks `subagent`, so issue #149 is local critical path.

## Next

1. Validate JSON planning artifacts and `git diff --check`.
2. Commit and push the parent planning seed.
3. Open draft parent PR with `Refs #148`.
4. Execute issue #149 local critical path: red official-surface count validation, refresh Chatwoot `api_surface.json`, add `cli_surface.json`, update docs/metadata, run targeted gates.

## Safety

No secrets requested or used. No credentialed connector checks. No dependency changes. No external writes. Parent PR merge to `main` remains human-gated.
