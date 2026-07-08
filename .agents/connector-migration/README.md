# Connector Migration Agents

These agents support connector-architecture-v2 migration and review work. They use the shared
agent metadata shape from `.agents/agentic-delivery/schemas/agent-spec.schema.yaml` instead of a
runner-specific format.

## Layout

- `agents/implementation/passb-expander.agent.yaml`: expands one connector definition bundle to its
  documented API surface.
- `agents/review/connector-reviewer.agent.yaml`: read-only adversarial review for migrated or
  expanded connector bundles.

## GSD Runtime

All connector migration agents use the repo-local official GSD Core Pi adapter:

- In Pi, use `/gsd <command>` or generated aliases such as `/gsd-programming-loop` and
  `/gsd-code-review`.
- In shell/non-interactive runners, use `scripts/gsd prompt <command> [args...]` and execute the
  generated prompt.
- Read `.agents/agentic-delivery/references/gsd-pi-adapter.md` before GSD work.
- Record manual-GSD fallback only when the adapter is unavailable.

## Rules

- Assign exactly one connector per implementation agent.
- Keep writes scoped to the connector paths declared in the issue or handoff.
- Do not edit shared/generated files unless the issue explicitly authorizes it.
- Do not commit from migration agents; the coordinator owns commits and merge validation.
- Stop for new dependencies, auth scope changes, secrets, destructive external actions, or quality
  gate reductions.
