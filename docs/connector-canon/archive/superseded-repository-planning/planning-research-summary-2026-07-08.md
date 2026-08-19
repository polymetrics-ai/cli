# Research Summary

**Generated via:** official GSD Core Pi adapter command path
**Commands:** `scripts/gsd prompt map-codebase --fast`, `scripts/gsd prompt docs-update .planning AGENTS.md .agents --planning-only`, `scripts/gsd prompt health --context`

## Findings

1. The repository has a completed Connector Architecture v2 baseline with declarative connector bundles, hooks, native implementations, conformance, and certification scaffolding.
2. Connector parity cannot be represented as REST-only. The plan must account for REST/JSON, GraphQL, XML/SOAP, CSV/NDJSON/report exports, binary transfer, file/object storage, SQL/CDC, queues/events/webhooks/audit logs, native protocols, direct-read, and reverse ETL writes.
3. Quick map counts show 547 connector definition directories, 547 API surface files, 7159 stream definition files, 5699 write definition files, 78 hook directories, and 37 native connector directories.
4. Quick map counts are not authoritative parity claims. Phase 1 must regenerate and reconcile inventory before connector fanout.
5. Official GSD Core command docs are pinned and surfaced through a repo-local Pi adapter because Pi is not listed as an upstream GSD runtime.
6. Agent/subagent instructions need durable guidance to use `/gsd`, generated `/gsd-*` aliases, or `scripts/gsd prompt` rather than stale runtime-specific assumptions.

## Current Risks

- Duplicate operation work from docs aliases or generated API references.
- Wrong abstraction for non-REST, binary, native, or direct-read surfaces.
- Unsafe write exposure if reverse ETL boundaries are bypassed.
- Agent drift if YAML specs and contracts do not mention the Pi adapter.
- Planning drift if `.planning/phases/**` is accidentally regenerated during this refresh.

## Recommended Next Steps

1. Keep issue #122 scoped to planning, GSD/Pi runtime, and agent guidance.
2. Keep `.planning/phases/**` unchanged in this refresh.
3. Require future implementation agents to run official GSD command prompts before production edits.
4. Start connector fanout only after inventory reconciliation is generated and reviewed.
5. Route PR review through Claude automatic review, with Copilot fallback only under documented blocker conditions.

---
*Research refreshed: 2026-07-08 via repo-local official GSD Core Pi adapter.*
