# Plan: Bitbucket CLI Surface Metadata

Sub-issue: #90
Parent issue: #79
Parent branch: `feat/79-bitbucket-cli-parity`
Nominal sub-issue branch: `feat/bitbucket-cli-surface`
Execution mode for this worktree: `local_critical_path` (no subagent tool available; #90 creates the seed Bitbucket bundle and should not run in parallel with #93).

## GSD command path

- `scripts/gsd prompt plan-phase issue-90-bitbucket-cli-surface --skip-research --tdd` — prompt generated and followed.
- `scripts/gsd prompt programming-loop init --phase issue-90-bitbucket-cli-surface --dry-run` — unavailable (`scripts/gsd: unknown GSD command: programming-loop`).
- Manual fallback: `.pi/prompts/pm-gsd-loop.md` + `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-documentation`
- `golang-context`
- `golang-concurrency`
- `golang-graphql`
- `golang-lint`
- `golang-spf13-cobra`

## Objective

Create a validated Bitbucket `cli_surface.json` and the minimal safe connector definition bundle needed for `connectorgen validate` to load it. This slice is metadata-only: it must not enable credentialed checks, live connector calls, direct-read execution, generic raw HTTP writes, local git operations, binary downloads, or reverse ETL execution.

## Source inputs

- Official Bitbucket Swagger: `https://api.bitbucket.org/swagger.json`.
- Official Bitbucket API docs: `https://developer.atlassian.com/cloud/bitbucket/rest/`.
- GitHub pilot shape: `internal/connectors/defs/github/cli_surface.json`.
- Validator/schema shape: `internal/connectors/engine/schema/cli_surface.schema.json`, `cmd/connectorgen/validate.go`.

## Slice boundaries

In scope:

1. Add issue-scoped red tests proving Bitbucket CLI surface metadata must exist, parse, and remain safe.
2. Add `internal/connectors/defs/bitbucket/` seed bundle files required by the loader:
   - `metadata.json`
   - `spec.json`
   - `streams.json` with an intentionally empty stream list for this metadata slice
   - `api_surface.json` with an operation-ledger seed row, not the final 331-operation ledger
   - `cli_surface.json`
   - `docs.md`
3. Keep all executable command availability blocked, planned, partial, unsupported, or unsafe until later lanes implement streams/writes/direct reads.
4. Map Bitbucket-like app intents: repositories, pull requests, issues, pipelines, deployments, downloads, snippets, workspaces/projects, webhooks, branch restrictions, and admin settings.
5. Mark local git clone/download flows as `local_workflow` / `unsupported_local` or operation-backed blocked rows; no binary executor in this slice.

Out of scope:

- Full 331-operation ledger (#93).
- Stream-backed execution (#92).
- Direct-read execution/output policies (#94).
- GraphQL/advanced body support (#95).
- Sensitive/admin policy engine work (#96).
- CLI renderer/docs behavior (#91).

## Test-first plan

Red test target:

```bash
go test ./cmd/connectorgen -run TestBitbucketCLISurfaceMetadata -count=1
```

Expected red before production edits: failure reading `../../internal/connectors/defs/bitbucket/cli_surface.json` or missing bundle metadata.

Green criteria:

- Test passes after seed bundle + `cli_surface.json` are added.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` reports no findings for the full defs tree.
- JSON files parse with `jq .`.

## Safety policy for metadata

- No real credentials, tokens, access keys, Authorization headers, private keys, or secret-shaped examples.
- Examples may use `--json`, `--workspace`, `--repo`, and synthetic non-secret values only.
- Reverse ETL commands in metadata are `planned`, `partial`, `unsupported_api`, or `unsafe_or_disallowed` unless a declared write action exists (none in this slice).
- Generic raw API command is explicitly disallowed.
- Binary/local download commands remain blocked/unsupported until a bounded executor and destination policy exist.

## CLI help/docs/website parity

This slice adds metadata only. #91 owns help rendering and connector docs updates. Record exemptions:

- Runtime `pm bitbucket --help`: not applicable yet; no dispatcher behavior added.
- `pm help bitbucket`: not applicable yet; no runtime help topic added.
- `docs/cli/**`: not applicable yet; no user-facing command behavior added.
- `website/**`: not applicable unless generated connector data changes during validation.

## Verification checklist

```bash
jq . internal/connectors/defs/bitbucket/*.json
go test ./cmd/connectorgen -run TestBitbucketCLISurfaceMetadata -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
go test ./cmd/connectorgen -count=1
go build ./cmd/pm
```

Broader gates after a green slice as time permits:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
make verify
```
