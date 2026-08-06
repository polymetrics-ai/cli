# PLAN — truthful changefeed contract and discovery

Issues: #3745, #3746. Parent: #2986. Branch: `fm/cli-found-changefeed-contract-r1`.

## GSD path

- `scripts/gsd doctor`: passed.
- `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review`: passed.
- `go run ./cmd/agentcontractgen check`: passed.
- Discuss prompt: `scripts/gsd prompt discuss-phase issue-3745-3746-changefeed-discovery-r1 --auto`.
- Plan prompt: `scripts/gsd prompt plan-phase issue-3745-3746-changefeed-discovery-r1 --tdd --skip-research`.
- Execute, verify, and code-review prompts will be generated and executed inline after the
  matching phase evidence exists. No GSD role is spawned.

## Required skills loaded

- `golang-how-to` — Go task routing.
- `golang-design-patterns`, `golang-structs-interfaces`, `golang-naming` — closed descriptor and
  minimal capability-interface design.
- `golang-cli`, `golang-documentation` — machine-readable catalog/inspect behavior and explicit
  #3748 docs parity deferral.
- `golang-testing` — test-first executable regression proof.
- `golang-error-handling`, `golang-safety`, `golang-security` — fail-closed optional interfaces,
  untrusted descriptor input, and no secret/live-provider exposure.
- `golang-database` — PostgreSQL CDC status and no-SQL/connection safety boundary.

## Owned implementation boundary

The one connector evidence target is `postgres`. The foundation is authorized by #3745/#3746 to
change the shared descriptor/loader/CLI projection seams required for that target. Expected paths
are limited to the changefeed types, bundle loading/definition derivation, descriptor schema or
loader tests, PostgreSQL's own defs directory, embed registration, and focused catalog/inspect
tests. `internal/connectors/commandrunner/runner.go`, command schemas, generation validation,
docs/website, and all non-PostgreSQL connector directories are excluded.

## TDD slices

### Slice 1 — #3745 RED proof and closed contract

1. Add executable tests that show current catalog CDC output includes PostgreSQL even though its
   bundle says `cdc: false`, and that the advertised route leads to
   `connectors.ErrUnsupportedOperation` from the PostgreSQL reader.
2. Add a focused descriptor-contract test covering unsupported output and an
   implemented-descriptor-without-executor mismatch. Run the focused tests and record the actual
   failing output before any production behavior change.
3. GREEN: add the closed `connectors` descriptor types and a small capability-provider interface
   distinct from `CDCReader`; add optional `changefeed.json` parsing to the embedded bundle
   contract. Preserve legacy CDC types unchanged.
4. GREEN: add PostgreSQL's evidence-backed `changefeed.json` with `status: unsupported`; it must
   have no executor and must state why the existing decoder/stub is not a real feed.
5. REFACTOR: make status/mechanism/executor matching deterministic and ensure zero/absent
   descriptor values cannot accidentally become implemented.

### Slice 2 — #3746 derived discovery and PostgreSQL correction

1. Derive public CDC availability from the loaded descriptor plus a matching registered executor;
   never from a `CDCReader` assertion. An `implemented` descriptor with absent/mismatched executor
   stays non-capable.
2. Project the descriptor in `Definition`/inspect JSON so PostgreSQL explains `unsupported`, its
   mechanism, source/reason, state/recovery, and guarantees rather than disappearing silently.
3. Update catalog filtering so `--capability cdc` exposes only true implemented descriptors.
4. GREEN verification must execute `pm connectors catalog --capability cdc --json` and
   `pm connectors inspect postgres --json`; neither command reads credentials.

## CLI parity disposition

This slice changes existing JSON fields, not command names, flags, help, or namespace routing.
The #3748 surface slice exclusively owns help/manual/docs/website generator changes. The plan
will still execute and record: `pm connectors`, `pm help connectors`, catalog/inspect JSON, and
searches of `docs/cli` and `website` to prove no unrelated surface was changed. The PR handoff
must label broader documentation parity as deferred to #3748, not complete.

## Verification plan

- Focused tests for `internal/connectors`, `internal/connectors/engine`,
  `internal/connectors/native/postgres`, and `internal/cli` as determined by the implementation.
- Execute the two required `go run ./cmd/pm connectors ... --json` commands without credentials.
- `gofmt` only on changed Go files; targeted `go vet` and `go build ./cmd/pm`.
- Independent applicable verify gates: `make tidy-check`, `make lint`, `make docs-check`,
  `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`,
  `make connectorgen-surface-sync`, `make connector-boundary`, and `make release-workflow-check`.
- Do not run the timeout-prone whole `go test ./...` / `make verify` monolith locally; CI owns it.
- `git diff --check`; changed-path and forbidden-lane audit.

## Commit checkpoints

1. Planning/context/TDD checkpoint.
2. RED test checkpoint retaining execution output in the ledger.
3. GREEN contract, PostgreSQL descriptor, and discovery projection.
4. Verification/review fix checkpoint.
