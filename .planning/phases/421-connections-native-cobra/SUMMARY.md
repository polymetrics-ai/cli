# Phase 421 Summary

Status: PR #450 open against `feat/cli-architecture-v2`; accepted website ETL credential-shape review fix implemented and locally verified; commit/push checkpoint pending at this artifact point.

## Current state

- Worker branch: `refactor/421-connections-native-cobra`; sub-PR: https://github.com/polymetrics-ai/cli/pull/450.
- GSD adapter doctor passed; `programming-loop` prompt command missing, so manual GSD fallback recorded.
- Required reading and skills loaded. Repo-specific `.pi/skills/go-implementation/SKILL.md` is missing; global Go skills loaded.
- Scope stayed inside `connections` CLI/router/tests, website ETL docs/generated data for the accepted review fix, and issue-local planning artifacts.

## Delivered

- Replaced top-level `pm connections` legacy Cobra wrapper with a native Cobra subtree.
- Added native `connections create` and `connections list` actions with declared `StringArray` flags, `NoOptDefVal="true"`, unknown-flag whitelist, optional-value normalization, and docs-map help/usage.
- Removed the `connections` namespace `parseFlags` call site.
- Preserved legacy flag semantics: space/equals forms, repeated singleton last-wins, repeated `--primary-key`, repeated configs, bare bool sentinels, unknown flag tolerance, extra arg tolerance, and late `--root`/`--json` globals.
- Added no-op connection-name completion seam returning `ShellCompDirectiveNoFileComp`; Phase 15 completion implementation deferred.
- Added focused tests for native metadata, flag-form behavior, list tolerance, invalid action usage, and completion seam.

## Review-fix disposition

Accepted finding: `website/content/docs/etl.mdx` used `<credential>:<credential-name>` for `connections create --source/--destination`, while parser/docs require `<connector>:<credential>`. Fixed the website source and regenerated `website/lib/docs.generated.ts`. No secrets, credential values, help-tree behavior, new dependencies, or unrelated namespaces.

## Verification

- `go test ./internal/cli/... -run 'Connections|CobraRouterShell|Golden' -count=1` passed.
- `go test ./internal/cli/ -run Certify -count=1` passed.
- `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, and `make verify` passed.
- Runtime help parity checked: `./pm help connections`, `./pm connections`, `./pm connections --help`, `./pm connections --json`, and invalid action JSON usage error.
- Docs/website/golden diff empty; docs generate/validate and `npm run gen:docs --prefix website` passed with no tracked diff.
- `git diff --check origin/feat/cli-architecture-v2...HEAD` passed; `go.mod`/`go.sum` diff empty.
- Review-fix gates passed: `node website/scripts/gen-docs-data.mjs`, `gofmt -w cmd internal`, `go test ./internal/cli/... -run 'Connections|Golden' -count=1`, `go vet ./...`, `go build ./cmd/pm`, `npm --prefix website run gen:docs`, `make verify`, stale placeholder grep, diff check, and go.mod/go.sum guard. `website/node_modules` absent, so website typecheck/lint/test:unit were skipped without installing dependencies; CI authoritative.

## Safety

No secrets requested or printed. No credentialed checks. No runtime services started. No dependency changes. No parent/shared orchestration edits. No merge. `make verify` ran only the repository's local temp-dir smoke flow, including local reverse run to temp outbox.
