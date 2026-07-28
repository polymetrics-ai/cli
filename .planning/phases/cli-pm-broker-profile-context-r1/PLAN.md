# CLI PM Broker profile/context foundation

Branch: `fm/cli-pm-broker-profile-context-r1`
Base: `origin/integration/pm-broker-production-program` at `4b0615835b70ec3809379c0cf3e0aa91aca876b8`
Sub-issue PR base: `integration/pm-broker-production-program`
Primary CLI issue: [#566](https://github.com/polymetrics-ai/cli/issues/566)
Parent PR: [#593](https://github.com/polymetrics-ai/cli/pull/593)
PM Broker contract references: [pm-broker#33](https://github.com/polymetrics-ai/pm-broker/pull/33), [pm-broker#35](https://github.com/polymetrics-ai/pm-broker/pull/35)

## GSD path

- `scripts/gsd doctor` passed.
- `scripts/gsd list` passed and confirmed the repo-local Pi adapter command registry.
- `scripts/gsd prompt plan-phase cli-pm-broker-profile-context-r1 --skip-research` generated the planning prompt.
- `scripts/gsd prompt programming-loop init --phase cli-pm-broker-profile-context-r1 --dry-run` failed with `scripts/gsd: unknown GSD command: programming-loop`; manual GSD fallback is active per `.agents/agentic-delivery/references/gsd-pi-adapter.md`.
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-documentation`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-safety`, `golang-naming`, `golang-spf13-cobra`, and `golang-spf13-viper`.

## Objective

Add the CLI-side PM Broker profile/context/domain foundation for `Organization`, `Workspace`, `Environment`, `BrokerProfile`, named context selection, runtime mode selection, and safe contract-version refusal without enabling live provider operations.

## Accepted scope

1. Add a narrow internal package seam for PM Broker-safe domain types, versioned user context state, context resolution, runtime mode validation, and contract-version errors.
2. Use user-facing names `Organization`, `Workspace`, and `Environment` while preserving immutable internal IDs (`org_`, `wks_`, `env_`, `bpf_`).
3. Store only safe user-level context metadata in the OS-standard PM config area; project `.polymetrics/config.yaml` may carry required/default context and runtime-mode expectations only.
4. Add CLI commands for context management and read-only cached metadata discovery:
   - `pm context create|use|show|list`
   - `pm organizations list|show`
   - `pm workspaces list|show`
   - `pm environments list|show`
5. Document and validate runtime modes `remote`, `local`, and policy-bound `hybrid`; default production context resolution to `remote`; refuse local fallback for production writes and scheduled production jobs.
6. Add exact, narrow error semantics for incompatible broker contract versions: HTTP status `426`, error code `incompatible_contract_version`, supported versions, and safe correlation metadata. If the fake-client lane is absent, keep this as a package seam with TODO references, not a duplicated client.
7. Keep all tests synthetic, network-free, credential-free, and deterministic.

## Non-goals and safety boundaries

- No live broker HTTP calls, provider operations, GCP/VPS resources, production resources, or credential creation.
- No movement of credentials and no legacy vault changes.
- No service-account JSON key support.
- No raw secret export, raw authenticated HTTP, generic JSON/body forwarding, SQL write, shell escape hatch, runtime plugin upload, or public stable auth registry claim.
- No reverse ETL execution outside the existing plan -> preview -> approval -> run boundary.
- No new Go module dependencies.

## Implementation slices

1. **Planning checkpoint** — create this plan, TDD ledger, verification checklist, prompt snapshot, summary, and run state before production edits.
2. **Red tests** — add failing unit/CLI tests for PM Broker IDs/context store, runtime-mode policy, contract-version refusal, config keys, and CLI help/JSON behavior.
3. **Domain/package seam** — implement `internal/pmbroker` with versioned safe state, ID/name validation, metadata list helpers, context resolver, runtime-mode policy, OS user config path helper, and narrow incompatible-version error types.
4. **Config and CLI integration** — extend typed config with safe broker defaults and wire legacy Cobra commands for `context`, `organizations`, `workspaces`, and `environments` without live network calls.
5. **Docs/help parity** — update embedded help, `docs/cli/**`, website CLI reference/generated docs, and examples. Explicitly mark live broker/provider execution as future/unsupported.
6. **Verification** — run gofmt, focused package/CLI/config tests, help transcript checks, `go test ./internal/pmbroker ./internal/config ./internal/cli`, `go test ./...`, `go vet ./...`, `go build ./cmd/pm`, and broader gates as time allows before commit.
7. **Delivery** — commit the green slice, push `fm/cli-pm-broker-profile-context-r1`, and open the PR against `integration/pm-broker-production-program` before no-mistakes when instructed.

## CLI help/docs/website parity plan

- `pm context`, `pm context --help`, and `pm help context` render the context manual and exit 0.
- `pm organizations`, `pm workspaces`, and `pm environments` render namespace help when no action is selected; `list`/`show` print cached authorized metadata only.
- Invalid actions still return usage errors.
- `--json` envelopes are stable and never contain secrets or raw credential references.
- `docs/cli/context.md`, `docs/cli/organizations.md`, `docs/cli/workspaces.md`, `docs/cli/environments.md`, `docs/cli/config.md`, and website docs are updated or regenerated.

## Orchestration decision

`local_critical_path`: the assigned work is one small, dependency-ordered CLI/domain slice in an already isolated worktree. No disjoint mutating worker is needed. Read-only recon was performed inline through repo docs, issue #566, parent PR #593, and PM Broker PRs #33/#35.

## Human gates

Stop before any request for secrets, credentialed broker/provider checks, production resources, new dependencies, auth scope changes, destructive actions, a `main` merge, or a public stable auth registry claim.
