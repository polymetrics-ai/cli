# Concerns and Watchpoints

**Generated via:** official GSD Core Pi adapter command path.

## Active Concerns

1. **Inventory freshness** — quick counts exist, but Phase 1 must regenerate authoritative inventory before connector fanout.
2. **Surface duplication** — duplicated docs pages, aliases, generated references, and product guides can create duplicate operation work unless canonical identity is enforced.
3. **REST bias** — connector parity must not force GraphQL, XML/SOAP, CSV/NDJSON, binary, file/object, SQL/CDC, queue/event/webhook, native, or direct-read surfaces into REST stream shapes.
4. **Write safety** — product-safe mutations must map to reverse ETL actions; destructive/admin/elevated writes remain human-gated or excluded.
5. **Credential safety** — no planning artifact should contain secrets or prompt for credentials.
6. **Agent drift** — reusable agents and subagents must use the repo-local official GSD Pi adapter instead of stale Claude-local or manual-only command assumptions.
7. **Planning scope creep** — issue #122 must not edit `cmd/`, `internal/`, or `.planning/phases/**` in this refresh.
8. **CLI docs drift** — CLI feature work can leave runtime help, bare namespace behavior, `docs/cli/**`, website docs, and generated help/manual artifacts out of sync unless parity is required in the GSD plan and PR verification.
9. **Review coverage** — Claude skip/rate-limit/disabled states are blockers or require documented fallback, not approvals.

## Human Gates

- New dependencies.
- Auth scope changes.
- Secrets or credential access.
- Destructive/admin external actions.
- Production deploys.
- Quality-gate reductions.
- Reverse ETL execution.
- Parent PR merge to `main`.

## Mitigations

- Use `scripts/gsd doctor` and `.pi/skills/gsd-core/SKILL.md` as the first-line GSD runtime check.
- Record `scripts/gsd prompt ...` commands in planning traces for deterministic evidence.
- Keep connector fanout blocked until inventory review is complete.
- Keep generated claims tied to conformance/certification outputs.
- For CLI-visible work, require `.agents/agentic-delivery/references/cli-help-docs-website-parity.md` evidence before handoff.

---
*Concerns refreshed: 2026-07-08 via repo-local official GSD Core Pi adapter.*
