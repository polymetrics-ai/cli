# TDD Ledger — Issue #183

## Red target

Add `TestFreshchatImplementedETLCommandsHaveReplayFixtures` in `internal/connectors/conformance`.

Expected initial failure: missing replay fixture pages for Freshchat ETL streams that are declared in `cli_surface.json` but lack `fixtures/streams/<stream>/page_1.json`.

## Green target

Add replay fixtures for every implemented Freshchat ETL command stream and prove each emits at least one record through the real engine read path.

## Verification ledger

Pending.
