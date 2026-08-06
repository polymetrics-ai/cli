# #3771 Command-runner runtime content

## Scope

Deliver parent issue #3771 and its dependent sub-issues #3782, #3784, #3786, and #3790 in one
branch. The shared command runner must stop mutating or forwarding connector-declared
`redact_fields`. Declarations remain load-compatible; connector bundles are not rewritten.

## Binding decisions

- Runtime connector-command request, response, record, error, and preview content is complete.
  No command-runner masking path remains.
- Approval tokens remain time-bounded and omitted from JSON. Destructive write confirmation,
  preview/digest, expiry, and single-use behavior remain unchanged.
- Generic source-table reverse-ETL output handling is a separate application path and is not
  changed by this foundation.
- No live provider calls, credentials, migrations, engine changes, or capability declarations.
- #3771 ownership is limited to `BuildWriteCommand`, `Run`, `runDirectRead`,
  `runOperationDirectRead`, `runBinaryDownload`, and the redaction helper block in
  `internal/connectors/commandrunner/runner.go` plus their tests. In particular, do not edit the
  #3775 flag-materialization functions or the #3769 direct-read validation function.

## Evidence and provenance

- Parent issue: <https://github.com/polymetrics-ai/cli/issues/3771>
- Child issues: #3782, #3784, #3786, #3790, read in dependency order on 2026-08-06.
- The required historical ownership check found `origin/feat/204-crisp-all-ops` still associated
  with open PR #256 on 2026-08-06. This branch is not modified.
- #3739's historical decision file is absent from this checkout and all refs. Its retained plan
  and merged PR #3739 record the same captain decision for direct writes:
  `.planning/phases/cli-engine-rest-write-executor-r1/PLAN.md` and
  `.planning/phases/cli-engine-rest-write-executor-r1/TDD-LEDGER.md`.

## Required skills and workflow

Required-skills routing was read first, followed by `golang-how-to`, Go design, structures,
errors, security, safety, testing, CLI, documentation, and Vercel React best-practices skills.
The connector migration handoff, conventions, architecture design, CLI parity guidance, delivery
contract, and GSD Pi adapter guidance were also read.

`scripts/gsd doctor`, `scripts/gsd sources` for discuss/plan/execute/verify/review, and
`go run ./cmd/agentcontractgen check` passed on 2026-08-06. The generated GSD commands target
`issue-3771-commandrunner-runtime-content`, but `gsd-sdk query init.phase-op` reports that no
such roadmap phase exists. Therefore this branch uses the documented inline/manual GSD fallback:
these context, discussion, plan, TDD-ledger, and verification artifacts preserve the required
discuss → plan → execute → verify → review evidence without inventing a roadmap phase. No
sub-agents are used because this bounded shared-code task is directly assigned to this worker.

## Authoritative references

- `AGENTS.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/canonical/delivery-contract.json`
