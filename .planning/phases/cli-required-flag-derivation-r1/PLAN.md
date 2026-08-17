# Plan — CLI required-flag derivation r1

## Goal

Make required REST path parameters and CLI required flags one derived contract for every connector, then make the parity and unsupported classifications auditable.

## Scope and ownership

- Production ownership: `cmd/connectorgen/**`, generated connector surfaces, and narrowly necessary generic validation tests.
- Target connector for the parity measurement: GitHub. The implementation is a generator foundation slice mandated by P1, not a connector-local workaround; it contains no provider identifiers.
- Exclusions: GraphQL certification, fixture/entitlement work, evidence emission, broker/MCP/UI work, and any reclassification of unsupported commands.

## TDD slices

1. **Red:** Add a programmatic all-bundle invariant that reports every optional CLI flag mapped to a required REST path parameter. Add a focused surface-sync fixture covering a required path parameter and an optional query parameter. Run it to demonstrate the known failures before implementation.
2. **Green:** Change surface derivation so parameter-owned flags receive `required` from their matching REST parameter, including an already-derived flag that must be updated on regeneration. Regenerate through the repository generator; never edit emitted JSON by hand.
3. **Red/green runtime protection:** Add/extend the command-runner test for a missing generated path flag, asserting the concrete typed usage error and zero requester calls. Confirm a supplied path flag reaches credential preflight instead of a late path interpolation failure.
4. **Parity evidence:** Run the GitHub surface sweep and repository-wide counterpart before and after regeneration; record GitHub zero and the other-connector delta.
5. **Unsupported audit:** Enumerate the 27 `unsupported_api` and 23 `unsupported_local` declarations through their command/operation/provider-surface metadata. Emit an evidence report listing all 50 and any contradiction, without changing their classification.
6. **Verification and review:** Run required local gates, generator and website docs twice for deterministic output, execute the `verify-work` checklist, then perform code review and record findings/dispositions.

## CLI help/docs/website parity

- Runtime help: inspect a changed GitHub command's `--help`; the path flag must render as required.
- `pm help`, bare namespaces: not behaviorally changed beyond generated flag metadata; run relevant help checks and record an explicit exemption for unrelated namespace behavior.
- `docs/cli`, website, generated manuals: source prose is unchanged; run the generated-doc command twice and verify no uncommitted second-run diff so the metadata change propagates wherever generated.

## Commands planned

`go test -timeout 20m ./cmd/connectorgen`; focused commandrunner and CLI tests; `go run ./cmd/connectorgen validate internal/connectors/defs`; `go run ./cmd/connectorgen surface-sync` twice; `go run ./cmd/connectorgen surface-sync --check`; `go run ./cmd/connectorgen boundary`; `make connector-runtime-preflight`; `pnpm --dir website run gen:docs` twice; `git diff --exit-code -- website`; applicable repository gates individually; `go build ./cmd/pm`.
