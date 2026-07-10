# TDD Ledger: Zendesk Sensitive/Admin Policy

## Cycle 1 — reverse-ETL write coverage

- Red target: Zendesk bundle exposes 295 non-deprecated write actions/commands, including 85 destructive confirmations.
- Red evidence: `go test ./internal/connectors/engine -run 'ZendeskWrite' -count=1` failed with `Zendesk write actions = 0, want 295`.
- Green implementation: generated 295 non-deprecated Zendesk reverse-ETL write actions/commands from the official OAS, with 85 DELETE actions requiring `confirm: destructive`; deprecated mutating operations remain blocked.

## Manual GSD fallback

`scripts/gsd prompt programming-loop init --phase issue-156-zendesk-complete-implementation --dry-run` failed with `unknown GSD command: programming-loop`; proceeding with manual red/green/refactor loop per AGENTS.
