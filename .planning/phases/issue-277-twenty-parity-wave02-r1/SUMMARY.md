# Summary: Twenty CRM connector parity wave02 r1

Status: implementation complete; commit pending.

This phase restores the Twenty CRM connector parity bundle on `fm/cli-twenty-parity-wave02-r1`, using manual GSD fallback because the repo-local adapter does not expose `scripts/gsd prompt programming-loop` in this checkout.

Delivered:

- `internal/connectors/defs/twenty/**` with 28 streams/schemas, 112 write actions, 168 API ledger rows, CLI surface metadata, docs, and fixture coverage.
- Destructive DELETE coverage for 28 objects gated with `confirm: "destructive"`, typed schemas, no-body fixtures, and reverse-ETL plan -> preview -> approval -> execute safety docs.
- `docs/connectors/twenty/{MANUAL.md,SKILL.md}`, generated connector catalogs, website connector data, CLI golden/count updates, and Twenty conformance tests.
- Idempotent captain-policy addendum applied to issues #277-#284 with marker count 1 on each issue.

Verification highlights:

- `make verify` passed.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` passed with 550 connectors and 0 findings/warnings.
- `go test ./internal/connectors/conformance -run 'TestTwenty|TestConformance/twenty' -count=1` passed.
- `go run ./cmd/pm docs validate --dir docs/cli --connectors-dir docs/connectors --website-dir website/content/docs` passed.
- `go run ./cmd/pm help twenty`, `go run ./cmd/pm connectors`, and `go run ./cmd/pm connectors inspect twenty --json` passed without credentials or live provider calls.

Notes:

- Exact `go test ./...` was attempted and hit the default 10m package timeout in slow `internal/cli`; the repository's `make verify` uses `go test -timeout 20m ./...` and passed.
- `fm-ensure-agents-md.sh .` was run and reported an existing repo conflict because both `AGENTS.md` and `CLAUDE.md` are real files; no instruction-file changes were made in this connector slice.
- No no-mistakes pipeline, push, PR, merge, live certification, credentialed call, or external destructive action was run.
