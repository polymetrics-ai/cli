# VERIFICATION — Crisp provider parity Wave 1

## Result

PASS. The branch was rebased onto `origin/main` at `bf54aa4d4` before final verification. The inventory-only commit remains separate (`5e212f869`), followed by this Wave 1 bundle change.

## Provider ledger and scope

- `internal/connectors/defs/crisp/api_surface.json` has 234 cited provider operations: 105 reads and 129 writes.
- Exactly 21 documented Conversations/Conversation `GET` operations are stream-covered and executable in this wave.
- The other 213 rows have an explicit cited blocked disposition. Of those, 16 name a shared foundation (14 typed `HEAD` reads, RTM result correlation, or RTM plus bounded signed upload); 197 are intentionally deferred connector-local later-wave work.
- No live Crisp requests, credentials, or writes were used.

## Executability and complete content

- `go build -o ./pm ./cmd/pm` succeeded.
- `./pm help crisp` and bare `./pm crisp` rendered contextual help successfully.
- Every one of the 21 `./pm crisp <command> --help` paths resolved successfully.
- `TestCrispListCommandPreservesFixtureContent` uses a local `httptest` replay of the Crisp fixture and asserts that `fixture-visible-crisp-content` reaches the `commandrunner.Run` emitter unchanged. It revalidates Crisp against the content-preservation fixes merged in #3868 and #3872.

## Local gates

- `go test ./internal/connectors/defs/crisp -count=1`
- `go test -race ./internal/connectors/defs/crisp -run TestCrispListCommandPreservesFixtureContent -count=1`
- `go test ./internal/connectors/commandrunner -count=1`
- `go test ./internal/connectors/engine -count=1`
- `go test ./internal/connectors/conformance -count=1`
- `go test ./internal/connectors -count=1`
- `go test ./internal/cli -count=1`
- Targeted `go vet` for those packages
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go run ./cmd/connectorgen surface-sync --check`
- `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connector-boundary`, and `make release-workflow-check`
- `pnpm run test:scripts && pnpm run typecheck` in `website/`
- `git diff --check`

## Documentation and website parity

Fresh connector documentation generated into an isolated temporary directory matches the Crisp `MANUAL.md` and `SKILL.md` tracked in `docs/connectors/crisp/`. The generated connector catalog and website catalog both list Crisp with its 21 Wave 1 streams; the website catalog's mapped CLI surface contains all 21 commands.

## GSD delivery trace

The required GSD sources and prompts for discuss, TDD planning, execution, verification, and review were resolved before implementation. This single-connector edit used the repository's documented inline fallback because there is no matching numbered GSD phase and the canonical connector worker contract forbids role-agent spawning. The plan, TDD ledger, verification, and review evidence are recorded in this directory.
