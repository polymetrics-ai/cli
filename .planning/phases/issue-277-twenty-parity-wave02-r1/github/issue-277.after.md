# twenty (Twenty CRM) — connector all-ops CLI parity (parent)

Bring **Twenty CRM** to full all-ops CLI parity as a metadata-driven connector under
`internal/connectors/defs/twenty/`. Base `https://api.twenty.com` (`/rest` + `/graphql`, bearer auth).

Introspected surface (see `.planning/auto-loop/RESEARCH/twenty/RESEARCH.json`, `complete: true`, 0 unclassified):
**28 objects · 546 fields · 56 read ops** (28 `stream_read` list + 28 `direct_read` get) **· 112 write verbs** (84 `reverse_etl` create/update/batch + 28 `destructive_admin` delete) = **168 ops**.

## Decomposition (7 dependency-ordered parity slices)

| # | Slice | Deps | Endpoints |
| --- | --- | --- | --- |
| S1 | Foundation: metadata + spec + api_surface coverage manifest | — | all 168 classified |
| S2 | Object schemas (28 objects / 546 fields) | S1 | field schemas |
| S3 | Read streams (28 stream_read + 28 direct_read) | S1, S2 | 56 read |
| S4 | Reverse-ETL writes (create/update/batch) | S1, S2 | 84 reverse_etl |
| S5 | Destructive writes (28 delete, typed-confirm) | S4 | 28 destructive_admin |
| S6 | CLI surface + help/manual/website parity | S3, S4, S5 | gh-like over 168 |
| S7 | Fixtures + docs.md + conformance & `pm connectors certify` | S2-S6 | end-to-end |

Plan: `.planning/auto-loop/PLAN.md`. Research: `.planning/auto-loop/RESEARCH/twenty/RESEARCH.{md,json}`.

## Policy
- Parent branch from `main`; parent PR is DRAFT → `main`, `Refs #<this>`, human-gated merge only.
- Sub-PRs target the parent branch (`Refs #<sub>` + `Refs #<this>`, no closing keywords).
- Reverse-ETL follows plan → preview → approval → execute; deletes require typed confirmation.
- Never push to `main`; no secret stored/printed; no new deps or auth-scope changes without a human gate.

Orchestrated by the autonomous delivery loop (`.agents/agentic-delivery/workflows/pi-autonomous-orchestration-loop.md`); Codex `pm-gsd-worker` implements each sub-issue.

<!-- captain-policy-twenty-destructive-confirmation-v1 -->
## Captain policy addendum — destructive/admin parity safety

This issue's existing scope and operation counts are preserved. Documented Twenty CRM DELETE/destructive/admin operations remain in scope for the connector ledger instead of being blanket-excluded as unsafe. They may be executable only when represented by connector-owned typed schemas, bounded fixtures, `confirm: "destructive"` / typed destructive confirmation, and the existing reverse-ETL plan -> preview -> explicit approval -> execute path.

This addendum authorizes no live provider calls, no credentials, no generic raw write tools, no unsafe execution, and no count changes; it only records the captain policy that destructive operations are included with typed confirmation and safety gates.
