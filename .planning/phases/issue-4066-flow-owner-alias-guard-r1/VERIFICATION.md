# #4066 Verification Checklist

**Status:** Local verification green; external delivery remains pending

## Manual-GSD fallback

`#4066` is a final issue-first correction rather than a numbered active
roadmap phase. The repository's single-worker contract forbids the normal GSD
role spawning, so this records the required lifecycle inline. The resolved
adapter commands were `discuss-phase`, `plan-phase`, `execute-phase`,
`verify-work`, and `code-review`; the context, plan, and TDD ledger were
written before the production edit and this verifier records their outcome.

## Acceptance evidence

- [x] A flow with an omitted connection rejects a generated owner alias through
      the existing typed `records` ambiguity and flow remedy.
- [x] A legitimate real table that collides with a generated alias remains
      accessible to generic SQL but is unavailable to an unscoped flow.
- [x] ASCII-case variants of generated aliases and bare names fail closed for
      unscoped flows in quoted and unquoted SQL while generic SQL preserves the
      real table.
- [x] Explicit flow selectors, `_unattributed`, `SELECT 1`, and action-source
      reads remain covered by the focused race matrix.
- [x] The #4063 generated metadata remains the single scalar
      `flow_cli.go:20` to `flow_cli.go:21`; no later correction changes it.

## Local commands

- [x] `go test -race -timeout 20m ./internal/cli -run
      '^(TestFlowOmittedConnectionRejectsGeneratedOwnerAliases|TestFlowGeneratedOwnerAliasCollision|TestFlowGeneratedOwnerAliasCaseVariantCollision|TestFlowCaseVariantBareTableAmbiguity|TestFlowSourceConnectionSelectorsReadOnlyOwningRows|TestFlowSourceConnectionSelectorRefusesOmissionAndAcceptsUnattributed|TestFlowActionSourceReadsAllSelectedConnectionRows)$' -count=1`
- [x] `go test -timeout 20m ./cmd/connectorgen -run
      '^(TestCertificationDiscoverFunctionKindsFromRuntimeSource|TestCertificationGeneratedArtifactDriftFails)$' -count=1`
- [x] `go test -timeout 20m ./cmd/connectorgen -count=1`
- [x] `go test -timeout 20m ./internal/flow ./internal/app ./internal/warehouse
      ./internal/cli -count=1`
- [x] `go run ./cmd/connectorgen certification-matrix --check`
- [x] `go vet ./cmd/connectorgen ./internal/flow ./internal/app
      ./internal/warehouse ./internal/cli`
- [x] `make lint`
- [x] `go run ./cmd/agentcontractgen check`
- [x] `scripts/verify-gsd-workflow origin/feat/3988-github-certification`

## No-mistakes / Shepherd record

Run `01KZR72KN8ZKF8V8M5Q19FPJJ5` passed at
`804ca41a6407d4bb7cdaf8141b69d7dc47011048`. Push, PR, and CI stages were
intentionally skipped because the tool cannot bind this stacked integration
base. Its four review fixes are dispositioned in `REVIEW.md`; custody was
returned with `no-mistakes axi sync --recover` and no pipeline commit was
dropped.

## Known inherited formatting boundary

`git diff --check 5c92888c996319b41eec6e86ca99fcda4cb365f9` still reports the
seven older Markdown hard-break lines under
`.planning/phases/issue-3897-flow-connection-scope-r1/` that predate the exact
starting head. This correction removes its own six introduced hard-break
findings and does not expand scope to alter those inherited records.
