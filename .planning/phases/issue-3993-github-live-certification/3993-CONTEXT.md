# Phase 3993: github-live-certification - Context

**Gathered:** 2026-08-11
**Status:** Ready for planning
**Source:** Captain ship brief; existing #3993 issue tree and current remote PR #4061.

<domain>
## Phase Boundary

Recover the existing GitHub certification delivery on PR #4061 without changing its
branch, base, draft state, or historical correction ledger. Establish a fresh,
bounded 1/5 lineage for current-SHA closure: prove only behavior exercised by one
freshly built `pm` process with the shared GitHub rate coordinator, an approved
full-parity GitHub App credential boundary, typed inverse cleanup, and sanitized
evidence. The GitHub -> connection-owned Parquet/DuckDB -> GitHub route remains
warehouse-mediated throughout.

</domain>

<decisions>
## Implementation Decisions

### Delivery and lineage

- **D-01:** #3993 is the sole existing owner for current-SHA credentialed live closure; do not create a duplicate sub-issue.
- **D-02:** The prior `1/#4020 -> 2/#4022 -> 3/#4027 -> 4/#4039 -> 5/#4050` correction ledger is immutable. This delivery starts a separately named fresh lineage at **1/5**, never “correction 6”.
- **D-03:** Work only on `test/3993-github-live-roundtrip-nm5` / PR #4061, whose required base remains `feat/3988-github-certification`; preserve draft status and never force-push, merge, retarget, reset, or create another PR.

### Provider safety and truthfulness

- **D-04:** Use only an approved full-parity GitHub App installation credential and an immutable, uniquely run-owned `Polymetrics-Cert` boundary. Inventory availability, type, scope, and fingerprint only; no secret value may enter output, artifacts, or repository files.
- **D-05:** Every provider mutation requires a declared typed inverse, independent provider read-back, idempotent cleanup, and final empty-residue proof. If a target is not explicitly safe under the existing issue contract, record `needs-decision` rather than substitute a personal or unrelated repository.
- **D-06:** A Node runner which starts one external `pm` process per operation is not evidence of one-process/shared-coordinator execution. It may remain synthetic or historical evidence only; any current live claim must come from a demonstrably in-process built-binary path.
- **D-07:** Fixture/static classification, old rate snapshots, old provider results, and a successful exit are never current-SHA certification evidence.

### Scope and dependencies

- **D-08:** Complete branch-local GitHub harness, credential-boundary, read-only provider, cleanup, and evidence work that is independent today. Do not copy or modify transport code from #4059, and perform only one final parent update after #4060 lands.
- **D-09:** A warehouse-to-GitHub flow action waits on #3994 and its underlying transport (#4059) when their executable contract is unavailable. A real persisted scheduled firing waits on #3992. Record each exact failed prerequisite and continue independent slices.
- **D-10:** The live proof must report direct-read returned record/pagination behavior, rate/admission observations, exact binary/surface/case hashes, warehouse row counts and query/read-back, flow identity, cleanup state, and redacted protocol evidence. It must not promote partial coverage to full connector certification.

### the agent's Discretion

- Choose the smallest branch-owned test-harness repair that makes one-process provenance mechanically unambiguous.
- Use focused package and Node tests first; run a heavy live provider certification only after announcing the requested validation window.

</decisions>

<specifics>
## Specific Ideas

- The existing built `pm connectors certify` route uses the real in-process `cli.Run` harness and is a possible subject for a bounded current-SHA proof. Its actual stage coverage must be measured, not presumed to satisfy #3993’s full-surface requirement.
- The legacy external-process sweep must not be silently serialized or called rate-coordinated merely because its parent uses a barrier.

</specifics>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Delivery contract and issue lineage

- `AGENTS.md` — required GSD/TDD lifecycle, GitHub safety, connector ownership, and verification gates.
- `.agents/agentic-delivery/canonical/delivery-contract.json` — canonical issue-first lifecycle.
- `.agents/agentic-delivery/contracts/issue-agent-contract.md` — delivery-record and PR requirements.
- `.agents/agentic-delivery/references/gsd-pi-adapter.md` — inline/manual GSD fallback rules.
- `.planning/phases/issue-3993-github-live-certification/PLAN.md` — immutable historical delivery and correction record.
- `.planning/phases/issue-3993-github-live-certification/TDD-LEDGER.md` — prior Red/Green evidence; new work must be a fresh lineage.

### Connector and certification architecture

- `docs/connector-canon/INDEX.md` — current connector delivery canon.
- `docs/connector-canon/IMPLEMENTATION-PROCEDURE.md` — foundation checks, warehouse mediation, and live-proof requirements.
- `docs/connector-canon/REMOTE-REPRODUCIBILITY.md` — structural versus live-proof boundary.
- `docs/migration/HANDOFF-CODEX.md` — current connector authoring entry point.
- `docs/migration/conventions.md` — declarative connector and direct-read authoring rules.
- `docs/architecture/connector-architecture-v2-design.md` — declarative engine and rate-limit architecture.
- `docs/architecture/connector-certification-design.md` — certification report, transcript, cleanup, flow, and schedule requirements.
- `docs/architecture/github-postgres-warehouse-certification.md` — GitHub/PostgreSQL certification obligations.

### Current evidence and dependencies

- `data/cli-github-certification-red-children-audit-r1/report.md` — exhausted historical ledger and external-runner limitation.
- `data/cli-github-live-connector-validation-r1/report.md` — historical live validation only; not current-SHA evidence.
- `data/cli-github-parity-live-coverage-r2/report.md` — prior coverage facts and limits.
- `data/no-mistakes-stacked-certification-resolution-r1/report.md` — safe local-only no-mistakes handling for this stacked branch.
- `internal/connectors/defs/github/certification.json` — definition-owned direct-read and typed write-pairing candidates.
- `internal/connectors/defs/github/rate_limits.json` — provider-cited policy and shared-coordinator scope.
- `internal/connectors/certify/` and `internal/cli/certify_cli.go` — current in-process certification path.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `internal/connectors/certify/cliharness.go`: drives `cli.Run` in-process and redacts known secret values from report argv.
- `internal/connectors/certify/stages_source.go`: serial certification stages, warehouse materialization, flow and schedule checks.
- `internal/connectors/defs/github/certification.json`: declares GitHub direct-read candidates and create/cleanup pairings.
- `scripts/github-live-proof-sweep.mjs`: existing external-process proof artifact with strict boundary validation, but `runProcess` starts a process per operation.

### Established Patterns

- `pm connectors certify github --full` is one outer built-binary process, while its harness invokes the CLI in-process.
- Provider-facing mutation evidence is valid only through `plan -> preview -> approval -> execute`, independent read-back, and inverse cleanup.

### Integration Points

- Any current-SHA live attempt must use a newly built `./pm`, the approved boundary/configuration, and a disposable local project root.
- PR #4060 can change parent-base behavior; #4059, #3994, and #3992 govern dependent outbound/flow/schedule slices.

</code_context>

<deferred>
## Deferred Ideas

- Full 1,521-command concurrent in-process provider execution remains #3993’s acceptance target but cannot be claimed from the existing external-process sweep.
- Transport fixes (#4059), connector flow actions (#3994), and real scheduled action firing (#3992) are outside this branch’s custody.

</deferred>

---

*Phase: 3993-github-live-certification*
*Context gathered: 2026-08-11 via captain-authorized non-interactive recovery*
