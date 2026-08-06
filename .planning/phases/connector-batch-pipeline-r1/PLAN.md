# PLAN — connector batch pipeline r1

Branch: `fm/cli-bulk-connector-pipeline-r1`.
Tracker: task/branch name; no parent issue number was supplied.

## GSD path

- `scripts/gsd doctor`: passed.
- `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review`: passed.
- `go run ./cmd/agentcontractgen check`: passed.
- Discuss prompt: `scripts/gsd prompt discuss-phase connector-batch-pipeline-r1 --auto`;
  executed inline and captured in `CONTEXT.md` and `DISCUSSION-LOG.md`.
- Plan prompt: `scripts/gsd prompt plan-phase connector-batch-pipeline-r1 --tdd`;
  executed inline in this plan.
- Execute, verify, and review prompts were resolved and will be run inline.
- Inline/manual fallback: this task has no numeric roadmap phase or supplied
  parent issue and the single-worker contract forbids role spawning. The
  fallback does not weaken TDD, verification, review, or human gates.

## Required skills loaded

- `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
  `golang-security`, `golang-safety`, `golang-design-patterns`,
  `golang-structs-interfaces`, and `golang-documentation`.

## Slice A — ledger intake and deterministic manifest (RED → GREEN)

1. RED: add focused `cmd/connectorgen` tests proving missing ledger evidence
   (including artifact URL/version/retrieval date) is a hard planning error;
   a selected row must preserve all measured counts and evidence instead of
   estimating them.
2. GREEN: add `connectorgen batch plan --ledger <path> --out <path>` with
   bounded size, explicit selection, deterministic output, and a report of
   excluded candidates/reasons. It does not contact providers or write a
   connector bundle.
3. Generate and commit a five-connector manifest from the live ledger. Its
   candidate rationale is evidence-quality-first and records each cited artifact
   and ledger retrieval date.

## Slice B — per-connector gate and explicit drop protocol (RED → GREEN)

1. RED: add a test containing one valid and one invalid temporary bundle. The
   invalid one must be reported as dropped while validation continues for the
   valid one.
2. GREEN: add `connectorgen batch gate --manifest <path> --defs-root <path>`.
   For each included connector it runs `validatePath`, `syncBundle(..., true)`,
   and `commandrunner.Preflight` against a freshly loaded engine bundle. The
   latter is the runtime entry point, not a copied generator rule.
3. Emit a machine-readable batch report with included, dropped, declared
   operation counts, and executable/blocked/excluded split. A nonzero exit is
   returned if any selected connector is dropped, but all candidates are
   evaluated first.

## Slice C — documentation and staged handoff

1. Document the full ledger → artifact → operation ledger → bundle → gate path,
   command ownership, no-redaction rule, and plan/preview/approval/execute
   requirements for writes.
2. State the precise external gate: no `api_surface.json` v2 emission before
   #3869. Rebase and reverify against the #3870 and #3868 foundation commits
   once they land; do not invent a replacement provenance contract.
3. Provide the exact next-run commands and batch-drop mechanics so a worker can
   begin authoring as soon as the dependencies merge.

## Deliberately out of scope until foundations land

- No connector directory, schema, fixture, CLI surface, operation declaration,
  provider request, credential, or live API execution.
- No shared schema/engine/runner changes.
- No output redaction or `redact_fields` declaration.
- No artifact parser that emits a competing v1/v2 provenance representation.

## Verification plan

- Red and green focused tests: `go test ./cmd/connectorgen -run '^TestBatch'`.
- Package regression: `go test ./cmd/connectorgen`.
- Runtime-guard proof through a temporary bundle: the batch gate invokes the
  exported `commandrunner.Preflight`; after actual connectors are authored, run
  `go test ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$'`.
- Static/build: `gofmt`, `go vet ./cmd/connectorgen`, `go build ./cmd/connectorgen`.
- Repository gates individually as applicable: `make tidy-check`, `make lint`,
  `make docs-check-no-build`, `make smoke-no-build`, `make agent-contract-check`,
  `make connectorgen-validate`, `make connectorgen-surface-sync`,
  `make connector-boundary`, and `make release-workflow-check`.
- `connectorgen` is a developer tool, not a `pm` public command; website/public
  CLI manual parity is N/A. Its own usage text, tests, and migration guide are
  the applicable documentation contract.

## Commit checkpoints

1. Plan/context/TDD checkpoint.
2. RED test checkpoint with observed failures retained in `TDD-LEDGER.md`.
3. GREEN command plus generated batch manifest checkpoint.
4. Verification/review documentation checkpoint.
