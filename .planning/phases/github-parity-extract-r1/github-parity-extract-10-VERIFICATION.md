# Verification — plan 10 typed source import

All checks were local after the source artifact hash verification.  No provider write occurred.

| Gate | Result |
| --- | --- |
| `node --test scripts/tests/github-combined-operation-ledger.test.mjs` | PASS — 8 tests |
| `node scripts/github-combined-operation-ledger.mjs --check` | PASS — REST 1220, Query 31, Mutation 274, total 1525 |
| `node --test scripts/tests/github-live-lab.test.mjs` | PASS — 23 tests |
| `node scripts/github-live-lab.mjs --check-boundary ...` | PASS — one exact allowed target |
| `go test -timeout 20m ./internal/connectors/engine -count=1` | PASS |
| `go test -timeout 20m ./internal/connectors/commandrunner -count=1` | PASS |
| `go run ./cmd/connectorgen validate internal/connectors/defs` | PASS — 551 connectors, 0 findings |
| `go run ./cmd/connectorgen surface-sync --check` | PASS — no drift |
| `git diff --check` | PASS |

The generated lock retains the source’s 1,546,421-byte SHA-256
`c09aba9911b08d2aa8a022578edaf256aa040f38d7fb7196656356ea236c249d` without embedding raw SDL.
