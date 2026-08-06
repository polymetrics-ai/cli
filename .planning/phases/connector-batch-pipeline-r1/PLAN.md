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
  `golang-security`, `golang-safety`, `golang-lint`, `golang-design-patterns`,
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

## Slice D — v2 artifact materialization and first five (RED → GREEN)

The #3869 provenance contract merged after the original control-plane slice.
This slice is intentionally limited to `cmd/connectorgen` and the five
manifest-selected bundle directories.

1. RED: add a local-artifact test showing that a selected bundle cannot gain a
   v2 inventory, command surface, or empty operation-executor catalog without
   every cited OpenAPI endpoint being classified and every existing executable
   stream/write remaining covered.
2. GREEN: add `connectorgen batch materialize`, which fetches (or reads a
   supplied public-artifact cache) only manifest URLs, records the new
   retrieval date and SHA-256 through the v2 artifact table, generates
   provenance for every artifact endpoint, writes an explicit empty
   `operations.json` when no non-redacting direct executor is promotable, and
   derives `cli_surface.json` commands only from existing executable
   streams/writes.
3. Author the five selected directories from their cited artifacts. Any
   artifact/bundle mismatch becomes a named materialization or gate drop; it
   is never filled with an invented endpoint or implicit capability.
4. Regenerate connector manuals/catalog data and validate each candidate with
   `connectorgen validate`, `surface-sync --check`, and real
   `commandrunner.Preflight` before the final batch report.

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

## Review remediation — fail-closed artifact and bundle ownership boundaries

1. Reject artifact URLs with userinfo, query, or fragment components; resolve
   every hostname before dialing, reject non-public destinations, disable proxy
   routing, and apply the same guard to every redirect.
2. Parse OpenAPI/Swagger inventories through a strict path-item walker that
   resolves local references, includes TRACE, and fails with a concrete unknown
   inventory reason for unsupported operation containers such as webhooks.
3. Treat a successful CLI-surface load with zero implemented commands as a
   runtime-preflight failure rather than an included candidate.
4. Read an existing bundle only from an explicit source root and materialize a
   copied, new destination bundle. A pre-existing destination is a named
   collision; cleanup may remove only a destination created by that invocation.
5. Require an HTTP 200 artifact response without `Content-Range`; classify a
   partial or ranged response as an unknown inventory before parsing it.
6. Make a bare `connectorgen batch` invocation render its contextual usage to
   stdout and succeed, while invalid subcommands retain usage-error behavior.
7. Retain executable coverage only for exact cited method/path identities; do
   not infer trailing-slash, method-case, or other endpoint equivalence.
8. Keep `TRACE` and `OPTIONS` in the inventory while reporting them as
   method-specific protocol-metadata exclusions through the existing ledger
   vocabulary.
9. Enforce protocol-metadata exclusions before `covered_by` at every batch
   classification boundary, including direct standalone gate runs.
10. Admit a gate candidate only with complete v2 provenance whose artifact and
    endpoint citation URLs exactly match its manifest artifact; count every
    refusal in the aggregate batch report.

## Commit checkpoints

1. Plan/context/TDD checkpoint.
2. RED test checkpoint with observed failures retained in `TDD-LEDGER.md`.
3. GREEN command plus generated batch manifest checkpoint.
4. Verification/review documentation checkpoint.
