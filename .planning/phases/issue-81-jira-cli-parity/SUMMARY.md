# Summary: Jira CLI Parity Parent

Status: draft parent PR #129 opened; #104 selected as first local critical-path lane.

## Completed

- Read required issue-first, parent-orchestrator, stacked PR, GSD, CodeRabbit, automated-review, connector architecture, CLI parity, and skill-routing references.
- Ran GSD adapter health checks.
- Confirmed parent PR for `feat/81-jira-cli-parity` was missing at kickoff.
- Confirmed Jira baseline through metadata-only `pm connectors inspect jira --json` after reading connector help.
- Created parent planning artifacts and orchestration state.
- Committed and pushed parent seed commit `982fa4c1`.
- Opened draft parent PR #129: https://github.com/polymetrics-ai/cli/pull/129.

## Current Blockers

- `scripts/gsd prompt programming-loop ...` is unavailable (`unknown GSD command: programming-loop`); manual GSD fallback is active and recorded.
- No Pi `subagent` tool is exposed in this harness; mutating workers are not spawned. #104 begins locally as `local_critical_path`.

## Next

1. Add and run the red `TestBundleLoadEmbeddedJiraCLISurface` test.
2. Add `internal/connectors/defs/jira/cli_surface.json`.
3. Run targeted tests and connector validation.
4. Commit/push coherent green slice and update parent PR.
