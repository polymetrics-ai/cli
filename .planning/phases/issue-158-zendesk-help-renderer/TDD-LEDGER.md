# TDD Ledger: Zendesk Help Renderer

## Cycle 1 — connector namespace help

- Red target: `pm help zendesk`, bare `pm zendesk`, and `pm zendesk read list-tickets --help` render help without credentials.
- Red evidence: `go test ./internal/cli -run 'Zendesk.*Help' -count=1` failed because `help zendesk` was missing, bare `zendesk` returned missing command path, and command `--help` attempted project/credential resolution.
- Green implementation: pending.

## Manual GSD fallback

`scripts/gsd prompt programming-loop init --phase issue-156-zendesk-complete-implementation --dry-run` failed with `unknown GSD command: programming-loop`; proceeding with manual red/green/refactor loop per AGENTS.
