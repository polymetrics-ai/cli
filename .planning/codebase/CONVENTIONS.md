# Conventions

**Generated via:** official GSD Core Pi adapter command path.

## GSD Command Conventions

- Prefer Pi interactive commands after project trust/reload:
  - `/gsd <command> [args...]`
  - generated aliases such as `/gsd-plan-phase`, `/gsd-map-codebase`, `/gsd-programming-loop`.
- Prefer shell prompt generation for deterministic traces:
  - `scripts/gsd prompt <command> [args...]`.
- Run `scripts/gsd doctor` before relying on the adapter in a new environment.
- Use `scripts/gsd sources <command>` when recording provenance.
- Record manual-GSD fallback only when the adapter is unavailable.

## CLI Help / Docs / Website Parity Conventions

For CLI command, subcommand, flag, output, connector surface, or help-topic changes:

- Read `.agents/agentic-delivery/references/cli-help-docs-website-parity.md` before implementation.
- Namespace command groups with no action selected, such as `pm connectors`, should render contextual help/subcommand summary and exit successfully.
- Invalid actions should still return usage errors.
- Update runtime help, `docs/cli/**`, website docs under `website/**`, generated help/manual artifacts, and tests together.
- Record parity evidence in the GSD plan, TDD ledger, verification summary, worker handoff, and PR body.

## Issue and PR Conventions

- One primary issue per implementation PR.
- PR titles use Conventional Commits.
- Default-branch completion PRs use `Closes #N`; stacked/incremental PRs use `Refs #N`.
- Commit and push coherent green slices after local gates.
- Never push to `main`; merge to `main` is human-gated.

## Implementation Conventions

- Plan before production edits.
- For behavior changes, follow GSD/TDD: plan, red test, green implementation, refactor, verification.
- Keep GSD plan, TDD ledger, and verification checklist current.
- No new dependencies without explicit human approval.

## Connector Conventions

- Declarative-first connector bundles under `internal/connectors/defs/<connector>/`.
- Hooks and native implementations only when justified by migration conventions.
- One upstream operation maps to exactly one primary classification.
- Multi-surface connector planning must include REST, GraphQL, XML/SOAP, CSV/NDJSON, binary, file/object, SQL/CDC, queues/events/webhooks, native protocols, direct-read, and writes.

## Safety Conventions

- Never request, print, summarize, or store secrets.
- Add credentials only from environment variables or stdin when explicitly required.
- Reverse ETL must follow plan, preview, approval, execute.
- Do not expose generic shell, generic HTTP write, or generic SQL write tools.
- Treat command arguments as untrusted.

---
*Conventions refreshed: 2026-07-08 via repo-local official GSD Core Pi adapter.*
