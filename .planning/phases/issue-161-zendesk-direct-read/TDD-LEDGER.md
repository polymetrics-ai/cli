# TDD Ledger: Zendesk Direct Read

## Cycle 1 — generic JSON direct-read policy

- Red target: commandrunner and engine reject `json` direct-read output policy before this slice.
- Red evidence: `go test ./internal/connectors/engine -run 'JSONPolicy|ZendeskDirectRead' -count=1` failed with `direct read output policy "json" is not supported`.
- Green implementation: added the generic bounded `json` output policy in `engine.DirectRead`, commandrunner, validator schema, and tests.

## Cycle 2 — Zendesk direct-read command coverage

- Red target: Zendesk bundle exposes 282 implemented direct-read commands with api_surface coverage.
- Red evidence: same red test run failed with `implemented Zendesk direct-read commands = 0, want 282`.
- Green implementation: generated 282 implemented Zendesk direct-read commands from the official OAS with path/query flag allow-lists and `api_surface` direct-read coverage.

## Manual GSD fallback

`scripts/gsd prompt programming-loop init --phase issue-156-zendesk-complete-implementation --dry-run` failed with `unknown GSD command: programming-loop`; proceeding with manual red/green/refactor loop per AGENTS.
