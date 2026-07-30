# Issue 2985 Provider Search/Query Foundation Plan

Issue: #2985
Branch: `feat/connector-provider-query-2985`
Integration base: `feat/connector-wave-01` when available; implementation may start from fresh `origin/main` and must rebase before no-mistakes.

## GSD mode

Manual GSD fallback recorded. `scripts/gsd doctor` passed, but `scripts/gsd prompt programming-loop init --phase issue-2985 --dry-run` failed with `unknown GSD command: programming-loop`; the adapter registry in this checkout has no `programming-loop` command. This phase follows `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` inline.

Skills loaded: `gsd-core`, `no-mistakes`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-database`, `golang-documentation`, `golang-spf13-cobra`, plus CLI parity guidance.

## Revalidated contract

- Canonical issue #2985 asks for distinct `provider_search` / `provider_query` capability metadata, not a global `capabilities.query=true` flip.
- Open duplicate PR/branch search found no open PRs and no matching remote branch for #2985. `origin/feat/connector-wave-01` did not exist at planning time; `origin/main` matched local HEAD before branch creation.
- Existing bounded patterns to preserve: direct-read uses fixed connector-relative endpoints, output policies, max bytes, path/query validation, and no raw URL; GraphQL engine uses fixed documents and typed variables; `pm query` remains local warehouse SQL/table inspection.

## Slice boundaries

1. Add additive serialized capability fields `provider_search` and `provider_query` to connector metadata schemas/types and rendered catalog/help surfaces without changing `query` semantics.
2. Add typed provider operation contracts under `operations.json` for `provider_search` / `provider_query` with request schema, response schema, explicit bounds, pagination metadata, output policy, and fixture references.
3. Validate fail-closed safety:
   - metadata-only capability enablement is a hard finding;
   - provider operations require schemas and positive bounds;
   - provider query/search operations reject raw/generic SQL, GraphQL, HTTP path/method/url, body, or raw payload escape-hatch request fields;
   - command surfaces may reference provider operations only through matching `provider_search`/`provider_query` intents and remain unsupported by runtime dispatch until an executor is implemented.
4. Expose provider operations in `Definition`/`Manifest`/connector manual rendering so `pm connectors inspect` distinguishes ETL streams, direct reads, provider search/query, reverse ETL, and local warehouse query.
5. Update CLI/docs/website parity text for the semantic separation; do not add live connector credentials, live writes, or certification claims.

## Non-goals

- No provider-specific command or connector bundle fan-out.
- No live provider executor, credentialed check, or external API call.
- No raw SQL write/read, raw GraphQL document, raw HTTP method/path/body, shell, file, binary, or browser escape hatch.
- No CDC or certification status claim.
- No new dependencies.

## Expected files

- `internal/connectors/connectors.go`, `manifest.go`, `definition.go`, `guide.go`
- `internal/connectors/engine/bundle.go`, `connector.go`, schema JSON files
- `cmd/connectorgen/validate.go`, `cmd/connectorgen/main_test.go`
- CLI help/docs: `internal/cli/docs.go`, `docs/cli/connectors.md`, `docs/cli/query.md`
- Website docs/types/components where the capability matrix and query page mention connector capabilities.
- Optional architecture doc update: `docs/architecture/connector-operation-kernel.md`

## Orchestration decision

Cycle `plan`: `local_critical_path` — this worker owns one foundation slice in an isolated worktree; no sub-worker spawn because the change touches shared core contracts that must stay coherent.

## Human gates

Stop for new dependencies, auth scope changes, credentialed connector checks, live writes, reverse ETL execution, generic raw escape hatches, quality-gate reductions, or any request to target/push/merge `main`.
