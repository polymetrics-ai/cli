## Task Delivery Header

- Issue: Refs #4283 — chore(connectors): pin and declare daily-use public API cohort
- Base branch: main
- Merges into: main
- Delivery: One PR against `main` containing sequential declaration-only increments of 10–20 connectors. This lane makes no live-certification claim and uses no provider credentials; each connector remains live-certification-pending.
- Working branch: fm/cli-top100-declaration-batch-r1
- Task: Pin only publicly retrievable provider API descriptions and derive connector-local, source-backed declarations. Exclude `github` and `zoom`; do not use provider credentials or change foundation code. Increment 1 targets Docker Hub, GitLab, Jira, Vercel, Notion, Stripe, Bitbucket, CircleCI, Sentry, and Asana.
- Verification: `scripts/gsd doctor`; `go run ./cmd/agentcontractgen check`; per-connector foundation checks; `connectorgen validate`; `surface-sync`; credential-free fixture/contract proofs; generated docs checks.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| The planned source is public and reproducibly pinned | green | `SOURCE-LOCK-VERIFICATION.json` verifies byte count and SHA-256 for all 10 locally retained public downloads against their source locks. |
| Every declared operation has a corresponding API surface entry | green | `PROGRESS-LEDGER.json` records 4,378 / 4,378 source method/path bindings (100%) and `connectorgen validate` passes with zero findings. |
| Live certification remains pending without credentials | green | Generated `certification-sweep.json` artifacts byte-check successfully for all 10 bundles and the ledger explicitly records `live_certification: pending`. |

## Scope Interpretation

Captain clarification on 2026-08-19 defines this as a parity-declaration lane, not a certification lane: the one-connector certification rule does not constrain its declaration increments. This work makes no live certification claim, changes no engine/foundation code, and records every unavailable executor or unsafe/elevated operation as disabled with evidence. No files under `defs/github/` or `defs/zoom/` may change.

## Increment 1 Plan

1. Retrieve each provider's public machine-readable API description without credentials and pin source URL, retrieval time, SHA-256, byte count, method counts, and operation inventory under the connector's `sources/` directory.
2. Reconcile the pinned method/path inventory with the existing `api_surface.json`; retain every existing covered or blocked disposition and record any source/api-surface drift as an explicit rejection or foundation gap.
3. Materialize missing empty operation/write ledgers only when the bundle has no executable operation/action of that kind; never create a fake request or response schema. Declare source transport only when the existing streams have the registered declarative source executor; declare destination transport only when an eligible existing typed action and acknowledgement contract exist.
4. Update per-connector progress, rejection, and foundation-gap records; live certification remains `pending`, never passed or failed.
5. Run red/green artifact checks, connector validation, surface synchronization, source/api-surface inventory checks, conformance/fixture tests where present, generated docs checks, and the repository gates appropriate to these JSON changes. Commit the increment before reporting its measured elapsed time and file count.

## Foundation Check

| Need | Evidence required | Increment-1 disposition |
| --- | --- | --- |
| Declarative stream source | Registered `declarative_stream_source` executor and existing fixture-backed stream | Declare only where the bundle already has stream fixtures; validate structurally. |
| Reverse destination | Existing typed action plus durable acknowledgement and registered destination executor | Do not invent a transport. Record `foundation-gap` when only raw REST writes exist. |
| Live certification | Credentialed provider interaction, bounded cleanup/receipt, accepted artifact | Explicitly pending; prohibited by task scope. |

The concrete registration gap was confirmed for this cohort. No `sync_transport.json` was emitted because the generic source/destination adapter is not registered; `FOUNDATION-GAPS.md` links #4093 and exact engine/test evidence.

Transport gate follow-up: Captain review required the omission to be explicit before increment 2. `TRANSPORT-GAP.md` records the path-(b) decision, the runtime evidence, the smallest safe recovery, and ten recoverable rejection entries. No new descriptor may use GitHub's evidence or issue-label action contract for another connector.

## Increment 2 Plan

1. Retrieve and pin the confirmed public provider artifacts for Gitea, Grafana, Trello, Slack, n8n, Google Calendar, Gmail, Twilio, Amazon SQS, and Elasticsearch. Preserve the source artifact's actual document format: OpenAPI/Swagger where published, Google Discovery metadata for the Google APIs, and the AWS-owned SQS service model for the native Query API.
2. Mechanically reconcile every documented method/path with its bundle's `api_surface.json`. Preserve an existing enabled binding only when it is the same documented method/path; add every remaining source operation as a disabled declaration with the fixed-vocabulary reason and source evidence. Do not create request, response, pagination, or body schemas.
3. Retain existing typed operations and writes, add empty ledgers only where they are absent, and ensure every source DELETE is either represented by an existing delete action or explicitly disabled. Do not manufacture `sync_transport.json`: add a recoverable `foundation-gap` record referencing #4093 for both transport directions unless the declared adapter, evidence constants, action binding, and acknowledgement contract all already exist.
4. Extend the source-lock verification, progress ledger, rejection list, foundation-gap log, TDD/verification evidence, connector certification sweeps, and generated documentation evidence. No live provider calls are permitted; live certification stays `pending`.
5. Run the declaration and generated-artifact gates for the ten affected bundles, then commit the increment and report its elapsed time and file count before starting increment 3.

## Lifecycle Record

- Inline/manual GSD fallback: this execution environment does not provide the compatible Pi runtime and the canonical contract forbids spawning GSD roles. Generated and reviewed: `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, `code-review` for issue `4283`.
- Skills loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing`.
- CLI parity reference reviewed. It becomes applicable after a connector-local declaration changes a generated command surface.
