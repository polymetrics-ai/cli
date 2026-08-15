## Intent

Refs #3982
Refs #3983
Refs #3979

This clubs the three audited residuals at their single definition-owned transport registration site: make PostgreSQL a managed destination, send real PostgreSQL worksets through the durable warehouse into that destination, and invoke gap-free bootstrap from the shipped App/CLI flow.

## What changed

- Added PostgreSQL's `sync_transport.json` destination declaration and connector-local production factory for all five managed driver modes. PostgreSQL's public `write` capability remains `false`; history and `change_capture` remain unadvertised.
- Added the closed warehouse-to-PostgreSQL adapter. It derives owner/target/schema/mapping from persisted connection identity, consumes the stage-owned Parquet workset, uses change-delivery for keyed upsert, reads durable ledger/baseline evidence back, and admits a checkpoint only after acknowledgement.
- Generalized PostgreSQL's declared source to typed dynamic relations and bridged first-run `BootstrapCDC` plus resumed `ReadCDC` into warehouse pages and transaction-bound checkpoints.
- Added sealed managed-target plan/preview/one-shot approval commands and production dispatch for declared PostgreSQL or API sources. Approval binds both endpoint revisions, source schema, connection/stream/mode, and destination identity; it is revalidated for expiry at every mutation boundary.
- Added typed API-schema projection for the real authenticated GitHub→PostgreSQL proof, while preserving legacy GitHub→warehouse dispatch.
- Added generated CLI/manual/skills/website/catalog/golden parity updates.

## Production call chains

**PostgreSQL or API → managed PostgreSQL:**

`cmd/pm` → `cli.Run` (`etl transport ... plan/preview`, then approved `etl run`) → `app.Open` → `composeTransportRegistry` → `RegisterDeclaredTransports` → definition-selected source/destination factories → `dispatchETLMode` / `runTransportETL` → `synctransport.Orchestrator` → registered source executor → connection-owned warehouse `Stage` / `Reopen` → `ManagedTargetTransportDestination` → `DeriveChangeDeliveryWorkset` or `DatabaseWriteExecutor` → native PostgreSQL `DatabaseDriver` → destination receipt/baseline read-back → acknowledged checkpoint/state commit.

**Bootstrap and resume:**

`pm etl run` → `SnapshotTransportSource.ReadTransport` → first run `readBootstrapTransport` → `BootstrapCDC` snapshot barrier → non-final snapshot pages staged/applied without checkpoint publication → final snapshot acknowledgement → committed pgoutput transaction pages → warehouse/destination/read-back/checkpoint; restart converts the persisted transport checkpoint and enters `ReadCDC` from the last acknowledged LSN.

## Edge-case coverage

| Required edge | Coverage through the changed production seam | Observable result / refusal evidence | Status |
| --- | --- | --- | --- |
| Cancellation mid-operation | Database write cancellation between batches plus orchestrator/App cancellation after apply. | Typed `context.Canceled`; failed batch rolls back rows/delivery and an acknowledged page retains only its already-durable checkpoint. The exact live native cancellation assertion is inside the pre-existing monolithic #4158 test and is not claimed separately. | Green; live subcase disclosed |
| Connection or process dies partway | Built `pm` bootstrap process is killed after the barrier, restarted, and fed another committed row; unknown-commit executor test injects disconnect. | Restart applies the next row once and advances beyond the prior LSN; unknown commit returns terminal unknown outcome with no fabricated acknowledgement or blind retry. | Green |
| Empty input | Built PostgreSQL binary route reads an empty relation. | Completed zero-row run changes only its durable stream checkpoint; target business counts remain unchanged and no empty managed relation is invented. | Green |
| Single-row input | Authenticated real GitHub issue #4163 runs through built `pm` in `full_overwrite`. | Exactly one live API row, structured JSON, target ledger receipt, warehouse artifacts, and checkpoint are observed. | Green |
| Large input | Built PostgreSQL route transfers 1,001 rows with batch size 1,000. | Exact 1,001 target rows across two pages, durable receipt/baseline, warehouse artifacts, and checkpoint. | Green |
| Duplicate delivery | A fresh approved binary replay sends the same source window. | Target row count and durable delivery ID remain byte-for-byte unchanged. | Green |
| Out-of-order delivery | Stable per-key baseline-window tests and order-fence tests refuse/regulate late source positions. | No checkpoint regression or duplicate row. The exact live managed-driver history/out-of-order polling assertion is the forbidden pre-existing #4158 failure; history is not advertised here. | Green at supported boundary; #4158 explicit |
| Schema drift | Source column is changed after preview and before binary execution. | Typed `ErrPostgresManagedTargetApprovalStale`; target rows/delivery and source checkpoint remain unchanged. | Green |
| Permission refusal | Built binary targets a restricted live PostgreSQL role. | Managed-target ownership proof refuses; zero target business/control mutation and no checkpoint advance. Typed live-driver permission classification exists in the #4158 monolith, so this independently passing binary assertion checks its stable error class plus no effects. | Green; typed subcase disclosed |
| Authentication refusal | Built binary targets a missing live PostgreSQL user. | `ErrManagedTargetTransportUnavailable` class reaches the CLI; zero target business/control mutation and no checkpoint advance. | Green |
| Concurrent runs racing same target | Managed-target provisioning lock tests, owner-isolated baseline tests, and App stale-writer/CAS race tests run under `-race`. | One owner/acknowledged state wins; losing/cancelled writers cannot cross owner scope or overwrite the winning checkpoint. A two-binary target-commit race is not separately runnable without the monolithic #4158 lane and is disclosed rather than silently omitted. | Green deterministic boundary; live race disclosed |
| Resume after interruption | Built bootstrap process kill/restart plus live failed-snapshot rebootstrap proof. | Resume starts after the acknowledged LSN; failed initial snapshot requires explicit rebootstrap and cannot advance a checkpoint. | Green |
| Replay of acknowledged item | Consumed token replay and fresh-plan data replay both run through built `pm`. | Consumed token is refused before source/target effects; fresh replay completes with unchanged rows and delivery identity. | Green |

Every independently runnable refusal above asserts the typed Go error at its component/App boundary and zero or unchanged rows, sends, ledger/baseline evidence, and checkpoint as applicable. The table names the exact live assertions that cannot be isolated from the captain-forbidden pre-existing #4158 failure.

## Red / Green / Refactor

- **Red:** App production preflight had no PostgreSQL destination factory; PostgreSQL source was fixed to a synthetic stream; App never called bootstrap; no built binary could reach PostgreSQL as a destination.
- **Green:** production `app.Open` composition preflights PostgreSQL→PostgreSQL and GitHub→PostgreSQL; both built-binary live proofs assert database, warehouse, receipt, and checkpoint state.
- **Review refactor:** removed unsafe paged PostgreSQL `full_overwrite` source advertising, moved overwrite proof to the exact one-row API source, revalidated approval expiry before every apply, preserved acknowledged results on finalization error, and enforced GitHub batch size from either endpoint.

## Testing

- `go test -timeout 20m ./internal/connectors ./internal/connectors/database ./internal/synctransport`
- `go test -timeout 20m ./internal/connectors/native/postgres`
- focused `internal/app` and `internal/cli` transport/approval/docs/golden tests
- focused `-race` for connectors/database/synctransport and App transport state
- focused `go vet`, `golangci-lint` (0 issues), `go build ./cmd/pm`, and `make smoke-no-build`
- live built binary PostgreSQL→warehouse→PostgreSQL: pass in 78.555s
- live authenticated GitHub→warehouse→PostgreSQL `full_overwrite`: pass in 39.378s; credential entered through `pm credentials add`, never printed or persisted in artifacts
- focused live PostgreSQL workset delivery, gap-free bootstrap, and failed-snapshot rebootstrap: pass
- final generation: docs, connector catalog, manuals/skills, website data, and golden transcripts in one pass
- drift gates: `surface-sync --check` (552 scanned, 0 corrections), docs/goldens, GitHub parity artifacts, agent contract, connector canon, release parity, tidy, and diff checks: pass

Per firstmate load control, no new full `go test ./...`, full parity/certification matrix, or runtime-preflight sweep was launched; CI owns those repository-wide runs. No no-mistakes command was run.

## GSD lifecycle and skills

The project-local adapter sources were resolved for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`. Their prompts were executed inline because the canonical issue worker contract forbids spawning lifecycle roles. The phase plan, TDD ledger, trace, verification, UAT, summary, and review record the red/green evidence.

Required skills used: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-database`, and `golang-documentation`.

## Safety and review

- No generic SQL/HTTP write surface, caller-selected target, dependency, or capability flip.
- Secrets remain environment/stdin-only and are excluded from plans, output, state, logs, and fixtures.
- #4125 and #4158 are unchanged.
- Inline code review: no unresolved findings. Automatic Claude review is expected on PR open; Copilot fallback was not requested.
