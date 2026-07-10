---
description: Autonomous connector implementation — research the full provider API surface, then deliver a full all-ops CLI-parity bundle (thin connector-mode alias over pm-auto-loop)
argument-hint: "<connector name / provider prompt>"
---

# Polymetrics Connector Auto-Loop

Connector to implement:

$@

This is the connector specialization of the general autonomous loop. It runs the same stage machine
as `/pm-auto-loop` with `problem_type = connector`, which activates the RESEARCH stage and the
7-slice CLI-parity decomposition. Follow `.pi/prompts/pm-auto-loop.md` and
`.agents/agentic-delivery/workflows/pi-autonomous-orchestration-loop.md` exactly.

Connector-specific expectations (everything else is inherited from pm-auto-loop):

1. **RESEARCH first (mandatory for connectors).** Dispatch `pm-web-researcher` to produce the full
   API-surface doc per `.agents/agentic-delivery/contracts/connector-research-doc-template.md`. Do not
   start PARENT_PLAN until its coverage self-check reports `unclassified_endpoints: 0`,
   `all_source_urls_present: true`, and `complete: true`.
2. **All ops.** Every read endpoint in the research doc becomes an ETL stream (or `direct_read`),
   every write verb becomes a reverse-ETL action; each is classified in `api_surface.json` with a
   `source_url` and `execution_model`. Zero `partial`/`planned`/`unsupported_api` rows for
   API-backed endpoints.
3. **Standard decomposition.** Parent roadmap issue + the 7 parity sub-issues (surface-metadata,
   help-renderer, stream-runner, operation-ledger, direct-read, advanced-query/binary,
   sensitive/admin) per `.agents/agentic-delivery/contracts/parent-issue-roadmap-template.md`. Seed
   each worker from `.agents/connector-migration/templates/connector-rollout-prompt.md`; for the
   api_surface/writes-heavy slices, work like
   `.agents/connector-migration/agents/implementation/passb-expander.agent.yaml`.
4. **Shared-file discipline.** `api_surface.json`, `cli_surface.json`, `docs.md` are cross-slice —
   serialize those slices (stacked, integrate-before-next) and coordinator-reconcile at INTEGRATE;
   the loop records `not_spawned_write_scope_collision` rather than editing a shared file in parallel.
5. **Done is asserted** by the connector suite in `.agents/connector-migration/rollout-checklist.md`
   and `validation-gates.md`: `make connectorgen-validate` (0 findings/0 warnings), `make verify`,
   `make smoke`, `pm connectors certify <name>`, and website `pnpm run gen:website-data` idempotency.
   Reach the human gate only when all pass and every sub-issue is integrated into the parent branch.

Env: set `SEARXNG_BASE` (and its token if proxied) before launching so `pm-web-researcher` can query
the audited `searxng` connector.
