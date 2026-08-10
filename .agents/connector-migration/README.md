# Connector Migration Agents

These legacy role specifications are retained for audit and their owning migration waves. They are
not an active delivery path: use the generated canonical worker and the
[connector delivery canon](../../docs/connector-canon/INDEX.md) for current work. They use the
shared agent metadata shape from `.agents/agentic-delivery/schemas/agent-spec.schema.yaml` instead
of a runner-specific format.

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

## GSD Runtime

All connector migration agents use the repo-local official GSD Core Pi adapter. Read
`.agents/agentic-delivery/references/gsd-pi-adapter.md` before GSD work; it owns the current Pi and
shell command paths, command-resolution checks, and inline/manual fallback rules. Do not maintain a
second command list here.

## CLI Help / Docs / Website Parity

When connector migration work adds or changes a CLI-visible connector surface, command, flag, help topic, or generated docs metadata, follow `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`. Update runtime help, bare namespace behavior, `docs/cli/**`, website docs, generated help/manual artifacts, and tests together, or record explicit not-applicable notes.

## Rules

- Assign exactly one target connector per implementation agent.
- Follow `ownership-rules.md` for coordinator-owned and worker-owned paths; connector lanes stay inside the declared target connector scope.
- If connector implementation needs shared runtime/tooling, schema, generated-index, or unrelated connector changes, stop and create/link a separate foundation issue/PR before proceeding.
- Record ownership guard evidence, changed-path compliance, target connector scope, and any foundation PR path in the worker handoff and PR body.
- Do not commit from migration agents; the coordinator owns commits and merge validation.
- Stop for new dependencies, auth scope changes, secrets, destructive external actions, or quality
  gate reductions.
