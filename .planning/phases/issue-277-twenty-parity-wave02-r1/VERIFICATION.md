# Verification Checklist: Twenty CRM connector parity wave02 r1

## Pre-edit / red gates

- [x] Worktree isolation checked.
- [x] Branch created: `fm/cli-twenty-parity-wave02-r1`.
- [x] `no-mistakes doctor` passed; daemon was not restarted.
- [x] GSD adapter health checked; programming-loop prompt unavailable, manual fallback recorded.
- [x] Missing-bundle targeted validation captured before production edits: connectorgen exited 1 for missing `internal/connectors/defs/twenty`; conformance had no matching subtest and exited 0.

## Connector-local gates

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs --json` — pass, `connectors_checked=550`, 0 findings/warnings. Note: direct `internal/connectors/defs/twenty` targeting is incompatible with the current validator because it expects the defs root and treats nested `fixtures`/`schemas` dirs as connector dirs.
- [x] `go test ./internal/connectors/conformance -run 'TestTwenty|TestConformance/twenty' -count=1` — pass.
- [x] `git diff --check` — pass.

## CLI help/manual/website parity gates

- [x] `go run ./cmd/pm help docs` — read before/around docs generation.
- [x] `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors` — pass; generated Twenty connector manual and catalog artifacts.
- [x] `go run ./cmd/pm docs validate --dir docs/cli --connectors-dir docs/connectors --website-dir website/content/docs` — pass.
- [x] `go run ./cmd/pm help twenty` — pass; Twenty CRM manual renders.
- [x] `go run ./cmd/pm connectors` — pass; bare namespace renders help.
- [x] `go run ./cmd/pm connectors inspect twenty --json` — pass; no credentials read, manifest includes 28 streams and 112 writes.
- [x] `cd website && pnpm run gen:website-data` — pass; generated 550 website connectors including slug `twenty`.
- [x] Golden/catalog parity tests updated and verified: `go test ./internal/cli -run 'TestConnectorCatalogCLIJSON|TestGoldenTranscripts' -count=1` — pass.

## Broad local gates

- [x] `gofmt -w cmd internal` — pass.
- [x] `go vet ./...` — pass.
- [x] `go test ./...` — attempted; default 10m package timeout hit slow `internal/cli` after this connector increased the generated catalog surface. The Makefile-supported timeout gate below passed.
- [x] `go test -timeout 20m ./...` — pass via `make verify`.
- [x] `go build ./cmd/pm` — pass.
- [x] `make verify` — pass.
- [x] `/Users/karthiksivadas/karthik-agent-workspace/bin/fm-ensure-agents-md.sh .` — attempted; exited 1 with existing repo conflict because both `AGENTS.md` and `CLAUDE.md` are real files.

## External/review gates intentionally not run in this worker turn

- [ ] no-mistakes pipeline — explicitly deferred by firstmate instructions until after implementation commit.
- [ ] Push/PR creation — explicitly deferred by firstmate instructions.
- [ ] Live Twenty credentialed certification — out of scope; no credentials or live calls.
- [ ] Merges — forbidden.

## Final result

Implementation and verification complete. Ready for the requested implementation commit, then stop for firstmate.
