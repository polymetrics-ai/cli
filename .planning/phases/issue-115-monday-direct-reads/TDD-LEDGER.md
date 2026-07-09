# TDD Ledger — issue #115 Monday direct reads

Manual GSD programming-loop fallback is active.

| Step | Evidence | Result |
| --- | --- | --- |
| Red | `go test ./cmd/connectorgen -run 'TestMondayDirectRead' -count=1` fails because Monday has no implemented direct-read commands (`missing "me view"`). | Captured |
| Green | Added fixed `me view` and `account view` direct-read commands, direct-read api_surface coverage, and real bundled query documents for `monday.me.get_me` / `monday.account.get_account`. Targeted tests pass. | Captured |
| Refactor | `go test ./cmd/connectorgen -run 'TestMonday(DirectRead|CLISurface|OperationLedger)' -count=1`, `go test ./internal/connectors/commandrunner -run 'TestRunMonday(DirectRead|BoardList)' -count=1`, `go test ./internal/connectors/engine -run 'TestBundleLoadEmbeddedMonday' -count=1`, and connectorgen validation pass. | Captured |
