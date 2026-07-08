# Connector Migration Agents

These agents support connector-architecture-v2 migration and review work. They use the shared
agent metadata shape from `.agents/agentic-delivery/schemas/agent-spec.schema.yaml` instead of a
runner-specific format.

## Layout

- `agents/implementation/passb-expander.agent.yaml`: expands one connector definition bundle to its
  documented API surface.
- `agents/review/connector-reviewer.agent.yaml`: read-only adversarial review for migrated or
  expanded connector bundles.
- `rollout-checklist.md`: the end-to-end checklist every connector rollout slice must satisfy.
- `templates/connector-rollout-prompt.md`: the per-connector worker prompt template (connector-neutral;
  replace the bracketed variables before dispatch).
- `validation-gates.md`: mandatory gates (JSON parse, connectorgen validate, secret scan, source
  links, operation classification, build/test, website idempotency, review).
- `ownership-rules.md`: coordinator-owned vs worker-owned files to prevent shared-file collisions.
- `next-batches.md`: sequenced candidate connectors (GitLab, Slack, Stripe, Jira, Salesforce, …) for
  rolling out the GitHub pilot's CLI parity shape.

## Rules

- Assign exactly one connector per implementation agent.
- Keep writes scoped to the connector paths declared in the issue or handoff.
- Do not edit shared/generated files unless the issue explicitly authorizes it.
- Do not commit from migration agents; the coordinator owns commits and merge validation.
- Stop for new dependencies, auth scope changes, secrets, destructive external actions, or quality
  gate reductions.
