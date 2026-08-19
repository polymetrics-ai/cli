Refs #4289

## Task Delivery Header

- Issue: Refs #4289 — complete six-class parity maps for batches 2 and 3.
- Base branch: `fm/cli-reverse-etl-destination-r1`.
- Merges into: `fm/cli-reverse-etl-destination-r1` → `main`.
- Delivery: existing PR #4300 publishes `fm/cli-map-batch23-r1`; merge remains human-gated.
- Working branch: `fm/cli-map-batch23-repair-r1` (publishes to the existing PR head).
- Task: reconcile every documented operation in the nineteen owned connectors across seven surfaces while preserving the corrected Batch-1 row schema and provider-source confidence.
- Verification: source-map and seven-surface integrity verifiers; connector generation/surface-sync; scoped conformance/runtime-preflight/CLI tests; build, vet, generated checks, and detached connector-boundary scan.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Complete provider input rather than the prior API-surface ceiling | live | `SOURCE-LOCK-VERIFICATION.json` records each connector's `acb85dc03` old count, regenerated execution-surface count, root source total, and complete confidence basis; the verifier reports 19 connectors / 5,127 documented operations. |
| Every documented operation has the corrected six-class ledger row | live | `verify-parity-maps.mjs` checks exact source operation identity, required row fields, class totals, API-surface binding, delete accounting, and declaration-pending reason. |
| Direct writes and reverse ETL are not conflated | live | The verifier rejects a `reverse_etl` endpoint class, requires an explicit action-level disposition for every typed write, and records App/CLI dispatch plus exact action-selection as foundation holds. |
| Existing executable connector surfaces remain valid | live | `connectorgen validate --json`, `surface-sync --check`, target conformance, and runtime preflight pass against the generated bundles. |

## What Changed

- Re-derived every owned source inventory from public provider material and regenerated the disposition maps. The report contains each connector's old/new `api_surface` count, `operations_found`, `counts.total`, per-kind/method counts, coverage confidence, and basis.
- Expanded the key understated surfaces: Facebook Marketing 36→1,446 API-surface entries from 1,445 named SDK declarations; LinkedIn Ads 10→287 source-backed entries from a complete sitemap crawl; Grafana 11→315; Trello 7→263; Slack 10→176; Elasticsearch 5→818; Amazon Selling Partner 353→370.
- Used complete rendered references for Aircall and LinkedIn, the GoCardless public OpenAPI payload, all Amazon SP-API models, and the Facebook Business SDK code-generation archive. Facebook's node/edge identifiers are explicitly recorded as instance-dependent rather than fabricated fixed identifiers.
- Repaired the held PayPal input: the former two-route Transaction Search pin is replaced by all thirteen official PayPal REST OpenAPI documents, yielding 115 source operations and eight documented deletes. Every batch-3 lock now records the same root `counts.total` shape as batch 2, and the verifier enforces it.
- Retained existing typed stream/action bindings as executable contracts while adding every absent source operation; prior `api_surface.json` content is not used as the inventory boundary.

## Classification and Safety

- Unauthored operations are `declaration-pending`, not `foundation-gap`.
- Typed writes remain enabled `direct_write`. All 621 existing schema-backed write actions are semantically eligible and individually representable by the closed typed destination. Reverse-ETL eligibility remains foundation-pending on persisted App/CLI dispatch and exact multi-action selection; no transport binding, selector, or action is invented.
- ETL is declaration-pending until a connector authors `sync_transport.json`; no engine code, credential, provider call, schema, runtime contract, or executable command changed.
- All provider artifacts were retrieved credential-free from public documentation/source URLs. No live provider operation was executed.

## GSD / TDD / Review Evidence

- The phase plan, TDD ledger, verification checklist, run state, and review record are in `.planning/phases/issue-4289-parity-map-batches-2-3-r1/`.
- Red: selected bundles lacked the source lock and corrected parity ledger; source coverage also lacked total counts and confidence accounting.
- Green: generated inventories map all 5,127 source rows with required provenance and classification invariants.
- `scripts/gsd sources code-review` and `scripts/gsd prompt code-review` were resolved. The Pi workflow runtime was unavailable in this shell, so the documented inline/manual review fallback was recorded in `REVIEW.md`; no finding remains.
- Required skills used: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-safety`, and `golang-security`. CLI help/manual/website parity is not applicable: this PR changes only connector definition data and generated parity evidence, not CLI parsing, flags, or help content.

## Verification

- `node .planning/phases/issue-4289-parity-map-batches-2-3-r1/traces/verify-parity-maps.mjs` → PASS (19 connectors / 5,127 operations).
- `go run ./cmd/connectorgen validate --json` → PASS (552 checked, 0 findings).
- `go run ./cmd/connectorgen surface-sync --check` → PASS (552 scanned, 0 drift corrections).
- Held-PR PayPal/root-count repair reran `connectorgen validate`, `surface-sync --check`, targeted conformance, commandrunner runtime preflight, `internal/cli`, generated/snapshot checks, detached `connector-boundary`, build, vet, and `git diff --check` → PASS; the boundary trace is `traces/connector-boundary-held-pr-repair.stdout`.
- Targeted connector conformance, implemented-command runtime preflight, and `go test -timeout 20m ./internal/cli` → PASS.
- `go build ./cmd/pm`, `go vet ./...`, `git diff --check` → PASS.
- `make tidy-check lint docs-check smoke-no-build agent-contract-check github-parity-artifacts-check connectorgen-certification-matrix connectorgen-certification-candidates connectorgen-certification-sweep connector-canon-check release-workflow-check` → PASS.
- Detached `make connector-boundary` → PASS; captured at `traces/connector-boundary-after-source-remediation.stdout`.

## Checkpoints and Review Route

- Commits: planning, batch 2, batch 3, verification evidence, source-lock remediation, and manual review record — all carry `Refs #4289`.
- Automated review: Claude automatic review is expected on PR open. No Copilot request was made; it is fallback-only if Claude is unavailable. There are no actionable automated-review findings at PR creation.
