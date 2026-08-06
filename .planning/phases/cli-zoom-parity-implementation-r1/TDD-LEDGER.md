# TDD Ledger — Zoom provider-owned inventory parity, R1

| Slice | Red evidence | Green evidence | Status |
|---|---|---|---|
| Wave 0 ledger | Not a behavior change: validate the report JSON and assert its inventory/disposition totals before replacing the ledger. | `connectorgen validate internal/connectors/defs/zoom` and conformance see the same 1,913-row surface with no unclassified operation. | green — `46eff2585`; 2026-08-05 |
| Wave 1 command reachability | `go test ./internal/connectors/defs/zoom -run TestCoveredStreamsHaveReachableCommands -count=1` failed on 2026-08-05 because `CommandSurface()` was nil. | The same test now proves three routes, their exact streams/API rows, real `commandrunner.Preflight` success, and fixture-backed bounded execution with the `--user-id` config override. | green — `go test ./internal/connectors/defs/zoom -count=1`; 2026-08-05 |
| Wave 1 runtime proof | Before the surface is present, `pm zoom` cannot synthesize provider command help. | Built `./cmd/pm`; `pm zoom`, `pm help zoom`, and the three documented command-help routes all resolve with exit 0 and no credential lookup. | green — 2026-08-05 |
| Generated docs/catalog | Before regeneration, the generated Zoom catalog lacks the command-surface representation and stale docs describe only seven ledger rows. | `pm docs generate --dir docs/cli` and `pnpm run gen:website-data` generated the Zoom manual, skill, and website catalog. The regenerated entries expose all three commands and honest ledger counts. | green — 2026-08-05 |

No live Zoom request is a test input. All runtime tests use existing sanitized fixtures or help/preflight paths that never resolve credentials.
