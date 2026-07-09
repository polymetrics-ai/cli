# Summary: Jira CLI Parity Parent

Status: planning started; #104 selected as first local critical-path lane.

## Completed

- Read required issue-first, parent-orchestrator, stacked PR, GSD, CodeRabbit, automated-review, connector architecture, CLI parity, and skill-routing references.
- Ran GSD adapter health checks.
- Confirmed parent PR for `feat/81-jira-cli-parity` was missing at kickoff.
- Confirmed Jira baseline through metadata-only `pm connectors inspect jira --json` after reading connector help.
- Created parent planning artifacts and orchestration state.

## Current Blockers

- Parent PR is not open yet; create after the parent seed commit is pushed.
- `scripts/gsd prompt programming-loop ...` is unavailable (`unknown GSD command: programming-loop`); manual GSD fallback is active and recorded.
- No Pi `subagent` tool is exposed in this harness; mutating workers are not spawned. #104 begins locally as `local_critical_path`.

## Next

1. Create #104 phase artifacts.
2. Add and run the red `TestBundleLoadEmbeddedJiraCLISurface` test.
3. Add `internal/connectors/defs/jira/cli_surface.json`.
4. Run targeted tests and connector validation.
5. Commit/push coherent green slice and open/update parent PR.
