# Run state — issue 4336

State: red

- Issue #4336 opened.
- Isolated task worktree and branch confirmed.
- `origin/main` and branch head are `e338cd301`; the required SHA is an ancestor.
- GSD adapter, command-source resolution, and canonical contract checks passed.
- The #4327 source-gap precedent and its plan/test/review evidence were read.
- Red importer fixture run: `go test -timeout 20m ./cmd/connectorgen -run 'TestSourceImportProviderDialectContracts|TestSourceImportKeepsDepthBoundFiniteAfterProviderIncrease' -count=1` exited 1 at the intended Bitbucket, Notion, Stripe, Vercel, Docker Hub, and GitLab boundaries.
- Red schema compiler run: `go test -timeout 20m ./internal/connectors/engine -run 'TestSchemaCompileKeywordMatrix|TestSchemaValidateInstances' -count=1` exited 1 because `example` and `patternProperties` are unknown keywords.
- Targeted provider-contract tests are green after the finite-bound, source-gap, and schema-dialect implementation.
- Full pinned-provider evidence run was deliberately local-only (the temporary harness was removed): `go test -timeout 20m ./cmd/connectorgen -run TestScratchFullProviderDocuments -count=1 -v` exited 1 before each target because it exposed further request-contract foundations outside this lane: Bitbucket unbounded string path parameter, Notion ambiguous request `oneOf`, Stripe unbounded array parameter, Vercel unbounded string parameter, Docker Hub unbounded number parameter, and GitLab dynamic request object properties.
- State: blocked by the brief's explicit second-missing-foundation stop rule. Do not widen this lane; captain must decide whether to split or sequence those request-contract foundations before the full-document proof can proceed.
