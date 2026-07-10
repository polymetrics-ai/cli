# TDD Ledger: Zendesk Help Renderer

## Cycle 1 — connector namespace help

- Red target: `pm help zendesk`, bare `pm zendesk`, and `pm zendesk read list-tickets --help` render help without credentials.
- Red evidence: `go test ./internal/cli -run 'Zendesk.*Help' -count=1` failed because `help zendesk` was missing, bare `zendesk` returned missing command path, and command `--help` attempted project/credential resolution.
- Green implementation: Added dynamic connector manual fallback for `pm help <connector>`/`pm <connector> --help`, bare connector namespace help, and command-specific help rendering from `cli_surface.json`; `go test ./internal/cli -run 'Zendesk.*Help' -count=1` passes.

## Cycle 2 — full-suite conformance fixture

- Red evidence: the first full gate run failed at `go test ./...` in `internal/connectors/conformance` because the executable Zendesk streams had no `fixtures/` directory and the mandatory first stream `list_activities` had zero fixture pages.
- Green implementation: Added a synthetic, non-secret `fixtures/streams/list_activities/page_1.json` replay page; `go test ./internal/connectors/conformance -run 'TestConformance/zendesk$' -count=1` passes.

## Manual GSD fallback

`scripts/gsd prompt programming-loop init --phase issue-156-zendesk-complete-implementation --dry-run` failed with `unknown GSD command: programming-loop`; proceeding with manual red/green/refactor loop per AGENTS.
