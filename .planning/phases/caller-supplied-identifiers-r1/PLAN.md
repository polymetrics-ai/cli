# PLAN — caller-supplied identifier sets

Branch: `fm/cli-found-caller-supplied-identifiers-r1`.

## GSD path

- `scripts/gsd doctor`, every required `scripts/gsd sources` lookup, and
  `go run ./cmd/agentcontractgen check` passed.
- Generated inline prompts: `discuss-phase caller-supplied-identifiers-r1 --auto`,
  `plan-phase caller-supplied-identifiers-r1 --tdd`,
  `execute-phase caller-supplied-identifiers-r1 --interactive`,
  `verify-work caller-supplied-identifiers-r1 --auto`, and `code-review` with
  the changed-file scope.
- Manual fallback is required because this firstmate task has no numbered GSD
  roadmap phase and compatible isolated role execution is unavailable. No
  lifecycle gate is waived.

## Required skills loaded

`golang-how-to`, `golang-cli`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-safety`,
`golang-security`, `golang-testing`, and `golang-documentation`.

## TDD slices

1. **Red — closed declaration and command binding.** Add test-only operation
   bundles and tests that fail because the schema, loader, and command runner
   lack `identifier_set.<name>` support. Cover each wire encoding, max/min
   bounds, malformed `chain_address`, and absent vs explicit blank values.
2. **Green — parse and validate.** Add the typed operation declaration,
   schema rules, loader semantic validation, command-to-operation binding, and
   engine pre-wire validation with value-free diagnostics.
3. **Green — wire shaping.** Transform validated sets to comma query, repeated
   query, JSON body array, or the one-item path segment. Assert server request
   shape and zero network hits on every rejection.
4. **Guardrails and docs.** Require exactly one matching string-array flag for
   every declared set; preserve the #3870 output-policy schema/runtime drift
   guard; do not add `covered_by` inference. Document authoring and the
   nested-batch exclusion in migration conventions.

## Verification plan

- Focused engine, commandrunner, and connectorgen tests with `-count=1`.
- Package tests for `internal/connectors/engine`,
  `internal/connectors/commandrunner`, and `cmd/connectorgen`; runtime
  preflight sweep; `go build ./cmd/pm`; scoped `go vet`; formatting; diff
  check.
- Individual repository gates from `make verify` excluding its timeout-prone
  full-suite wrapper: tidy, lint, docs, agent contract, connectorgen validate,
  surface sync, connector boundary, and release workflow checks.
- No credentialed provider call, production connector declaration, new
  dependency, reverse-ETL execution, generic request tool, or full monolithic
  `go test ./...` locally.
