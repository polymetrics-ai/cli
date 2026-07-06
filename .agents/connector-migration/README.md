# Connector Migration Agents

These agents support connector-architecture-v2 migration and review work. They use the shared
agent metadata shape from `.agents/agentic-delivery/schemas/agent-spec.schema.yaml` instead of a
runner-specific format.

## Layout

- `agents/implementation/passb-expander.agent.yaml`: expands one connector definition bundle to its
  documented API surface.
- `agents/review/connector-reviewer.agent.yaml`: read-only adversarial review for migrated or
  expanded connector bundles.

## Rules

- Assign exactly one connector per implementation agent.
- Keep writes scoped to the connector paths declared in the issue or handoff.
- Do not edit shared/generated files unless the issue explicitly authorizes it.
- Do not commit from migration agents; the coordinator owns commits and merge validation.
- Stop for new dependencies, auth scope changes, secrets, destructive external actions, or quality
  gate reductions.
