# TDD LEDGER — issue #180 Freshchat parity wave02 r1

## Red / pre-change evidence

- `go run ./cmd/connectorgen validate internal/connectors/defs/freshchat` failed because the command treats `fixtures/` and `schemas/` as connector roots when pointed at a single connector directory. Recorded as an issue checklist mismatch; canonical root validation `go run ./cmd/connectorgen validate internal/connectors/defs` had 0 findings.
- `go test ./internal/connectors/conformance -run 'TestConformance/freshchat' -count=1` passed before edits.
- Custom Freshchat captain parity red-check failed with 6 expected gaps: 3 legacy `excluded` api_surface rows, missing `cli_surface.json`, missing `certification.json`, and missing `confirm=destructive` on `create_agent`, `update_agent`, `update_agent_status`.

## Green / implementation evidence

- Converted the 3 legacy excluded Freshchat API rows into explicit blocked operation-ledger entries with source URLs and shared-foundation blockers.
- Added connector-local `cli_surface.json` with 34 typed command entries: implemented ETL/read commands, reverse-ETL write planning commands, and planned/blocked provider-search/binary commands.
- Added fixture-only `certification.json`; no live provider certification or write pairing is claimed.
- Updated Freshchat schemas/streams with documented bounded filters for agents, users, messages, groups, and channels.
- Added typed destructive confirmation/redaction metadata for destructive/admin Freshchat write actions.
- Updated Freshchat docs and regenerated the CLI golden transcript entry that lists provider-style connector commands.
- Freshchat parity green-check passed: 34 official operations represented, 31 covered operations, 3 blocked foundation operations, 0 exclusions, 13 write actions, and destructive/admin write confirmations present.

## Refactor / review evidence

- `go run ./cmd/connectorgen validate internal/connectors/defs` passed with 0 findings.
- `go test ./internal/connectors/conformance -run 'TestConformance/freshchat' -count=1` passed.
- `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` passed.
- `go vet ./...`, `go test -timeout 30m ./...`, `go build ./cmd/pm`, `make connector-boundary`, `git diff --check`, and `make verify` passed.
- Appended captain-policy addendum to issues #180-#187 with `gh-axi` and verified exactly one marker per issue.
- No live Freshchat provider calls, credentials, writes, VPS, merges, pushes, or PR actions were performed.
