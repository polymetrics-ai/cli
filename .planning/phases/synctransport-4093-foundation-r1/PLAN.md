## Task Delivery Header

- Issue: Refs #4093 — feat(synctransport): load definition-owned transports, register production adapters, and retire transient stages.
- Base branch: main
- Merges into: main.
- Delivery: A pull request from `fm/cli-synctransport-4093-foundation-r1` to `main`, with the required evidence, tests, generated checks, build, lint, and review complete.
- Working branch: fm/cli-synctransport-4093-foundation-r1
- Task: Replace the GitHub-shaped App transport composition with definition-owned neutral source and typed destination adapters; preserve GitHub behavior, make a synthetic second definition compose without App/dispatch changes, and make transient staging recoverable and bounded.
- Verification: Focused red/green Go tests for bundle loading, neutral composition, role validation, reconciliation, and stage cleanup; `go test -timeout 20m` for affected packages and `internal/cli`; `go run ./cmd/connectorgen validate`; `surface-sync --check`; connector-boundary; generated/snapshot checks; lint; build; and repository GSD/contract checks.

## GSD execution record

This direct-PR worker used the required inline/manual fallback because the local non-Pi runtime cannot provide compatible isolated GSD workers and the repository's single-worker contract forbids role spawning. Resolved adapter path: `scripts/gsd doctor`, `scripts/gsd sources {discuss-phase,plan-phase,execute-phase,verify-work,code-review}`, and `scripts/gsd prompt` for all five commands. The inline record is maintained in this phase directory.

Required skills loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, `golang-documentation`, and `golang-lint`.

## Discovery facts

- Exactly two current bundles contain `sync_transport.json`: `github` and `postgres`.
- `engine.loadSyncTransport` already schema-validates, strictly decodes, verifies schema version 1, validates, and clones the descriptor. `Bundle.SyncTransport` already projects it.
- `synctransport.RegisterDeclaredTransports` already rejects absent factories and evidence mismatches before executor construction or registry mutation; retain that property.
- `internal/app/issue_label_warehouse_transport.go` is the remaining composition seam: it owns the declarative source's fixed evidence and the closed GitHub issue-label destination, including App coupling.

## Plan

1. **Red — neutrality and closed admission.** Add a shared-adapter evidence test and a synthetic declarative second bundle. Assert that its descriptor selects its own evidence, the reusable production source is registered once, and no App/orchestrator/dispatch selection code changes are needed. Retain malformed member and unknown executor/evidence refusal coverage.
2. **Green — definition-selected source factory.** Extend `synctransport.DefinitionFactory` to admit the exact evidence values collected from definitions using a shared reference. Replace the fixed GitHub evidence constants with descriptor collection. Make the declarative source select the connector in its request rather than retaining the first registered connector.
3. **Red/green — typed action destination.** Keep the existing closed issue-label action adapter and have it select the preflighted destination for plan/apply/read-back. Preserve its explicit source binding, acknowledgement, approval, and strategy checks; do not add a generic HTTP/SQL writer. Reject `change_capture` as a destination declaration and prove the refusal is before I/O.
4. **Parity — durable transient-stage lifecycle.** Verify the existing receipt acknowledgement/read-back reconciliation and connection-owned bounded cleanup against the neutral path. No stage implementation change is permitted unless the regression evidence exposes a gap.
5. **Refactor/document.** Retire the fixed App factory wiring only after GitHub parity tests pass. Add `docs/sync-transport-definition.md` as the mechanical authoring contract. Run generated validation, repository gates, verification, and review; record every command/result in `VERIFICATION.md` and the firstmate status file.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Unknown or malformed transport members fail closed | live | Loader/registration tests assert an error and that source, destination, and I/O counters remain zero. |
| Bundle round-trip preserves descriptor exactly | live | A bundled definition is decoded then compared to the expected full descriptor value. |
| Registration has no connector-name branch | live | A synthetic differently named bundle registers the same generic adapters and is discovered/composed through its descriptor. |
| Evidence admission is definition-selected | live | Two definitions with distinct evidence values use the same factory; a substituted evidence value is refused before construction. |
| Synthetic second connector works without App/dispatch edits | live | A test registers the synthetic bundle and runs its source-to-destination route, asserting read/stage/apply/ack effects. |
| History and CDC refusals precede I/O | live | Route tests assert the typed refusal and zero source/stage/destination calls. |
| Kill-after-commit does not duplicate effects | live | A controlled unknown commit result is reconciled by persisted acknowledgement/read-back and the apply counter stays one. |
| Owned stage cleanup is safe and bounded | live | Quota/GC tests assert only matching owned, expired stages are deleted and retained stages do not exceed the configured bound. |
| GitHub behavior survives retirement | live | Existing GitHub source/destination composition tests pass through the neutral adapters with their prior observable effects. |
| Connector authoring recipe is usable | live | `docs/sync-transport-definition.md` names required descriptor fields, evidence, typed action/acknowledgement contract, and per-mode strategies. |

## Scope and safety

- Foundation implementation may edit `internal/connectors/engine`, `internal/connectors`, `internal/synctransport`, and App wiring only to remove the retired composition; new connector declarations remain under `internal/connectors/defs/<connector>/`.
- No new dependency, generic HTTP write executor, generic SQL write executor, credentialed check, or reverse-ETL execution is authorized.
- CLI command/help/website parity is not applicable: no CLI grammar, flag, help, or website documentation changes are planned. The dedicated definition authoring document is required.
