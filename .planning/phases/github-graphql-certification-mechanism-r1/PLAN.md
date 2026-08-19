# Plan — connector-neutral GraphQL certification

Commands resolved: `scripts/gsd prompt discuss-phase github-graphql-certification-mechanism-r1`; `scripts/gsd prompt plan-phase github-graphql-certification-mechanism-r1 --tdd`; execution is inline/manual because the direct-PR task forbids spawned roles.

## Boundaries

- Target connector: `github` only.
- Shared code may read generic GraphQL certification data from an engine-owned connector profile. It must contain no GitHub names, endpoints, schema fragments, query documents, or assertions.
- Connector definition files own the schema source, operation documents, expected-value assertions, live-run selection, and all non-pass classifications.
- No new dependency and no new GraphQL execution abstraction. Reuse the existing operation transport and certification stages.

## Slices

1. **Discovery and accounting.** Trace declaration → engine model → existing execution → certification results. Derive the 305 command population and record a mutually exclusive split: schema-conformant, needs-live, fixture-bound. Determine whether the schema is bundled or must be fetched. No production edits.
2. **Red test.** Add tests covering connector-neutral loading and stage failure. The bad case uses a valid compiled schema but changes a declared produced-value assertion, so compilation still succeeds and the test fails only if assertion execution is missing or wrong. Record the red output.
3. **Green implementation.** Add the smallest definition-driven stage and GitHub data to evaluate schema conformance plus connector-owned assertions. Ensure every absent live run produces a concrete non-pass reason, including `unexecutable` when a declared capability cannot run.
4. **Live probe and audit.** Run two small serial read-only queries with the approved disposable identity through the shipped binary. They must assert provider-produced values; remaining schema-only rows remain explicitly non-pass if no provider value was executed. Reconcile the counts to 305.
5. **Verification/review.** Run affected packages, their consumers, the repository gates, generated-file checks, boundary check, and website generator twice. Run GSD verify-work and code-review inline; address any gaps using the required GSD gap flow.

## TDD checkpoints

- Red: schema-valid record with a deliberately false assertion fails the new stage test.
- Green: restoring the assertion passes; malformed schema/document and unexecutable operation remain non-pass with their named reason.
- Refactor: retain generic data flow only; run `connectorgen boundary` to prove shared files contain no connector identifiers.

## Verification scope

`go test -timeout 20m ./internal/connectors/certify`; `go test -timeout 20m ./internal/connectors/engine`; `go test -timeout 20m ./cmd/connectorgen`; `go test -timeout 20m ./internal/cli`; plus `go vet ./...`, `go build ./cmd/pm`, `go run ./cmd/connectorgen validate`, `go run ./cmd/connectorgen surface-sync --check`, `go run ./cmd/connectorgen boundary`, project generated-file checks, lint/docs checks, and two identical `pnpm --dir website run gen:docs` runs.
