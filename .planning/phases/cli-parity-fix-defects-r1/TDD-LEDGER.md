# TDD ledger — GitHub command certification defect fixes

| Slice | Red | Green | Refactor / live proof |
| --- | --- | --- | --- |
| Exact integers | Persist a reverse plan with `9007199254740993`; after reload the fake provider must observe that exact path segment, never `9.007199254740992e+15` or the rounded decimal. | Decode interface-backed state numbers without a `float64` hop while preserving existing typed fields. | Exercise at least three real integer-ID commands and assert provider state, then cleanup. |
| Required bodies | Preflight or fake-provider assertions show blob, commit, tree, check-run, branch-protection, and commit-status cannot send their required payload today. | Closed record schemas expose typed flags and the fake provider receives exact required JSON. | Exercise at least three real commands, assert returned/read-back fields, and contain immutable Git objects in the fixture repository. |
| Provider paths | Bundle assertions reject Actions secret/variable endpoints containing `/agents/` and reject the defective project-draft declaration. | Operations, write actions, API surface, and CLI surface agree on provider-served endpoints. | Exercise three corrected Actions commands through `pm`, assert collection/object state, and delete/read back absence. |
| Write accounting | A declared idempotent DELETE returning 404 currently reports `RecordsWritten=1`; command run reports success. | The same response reports zero writes and an incomplete acknowledgement/error; a real successful 2xx still reports one. | Re-run org webhook create/delete plus another affected command with independent reads; never infer mutation from exit status alone. |

Red and Green command outputs are appended here before the production change and
after each green slice. A skipped test is not evidence.

## Red checkpoint — 2026-08-18

- Exact integer: `go test -timeout 20m ./internal/app -run '^TestGitHubConnectorCommandPreservesExactLargeIntegerPathAfterPlanReload$' -count=1` failed before the request with `reverse plan command payload changed since approval`; persistence had changed `9007199254740993`.
- Required bodies: `go test -timeout 20m ./internal/connectors/engine -run '^TestGitHubRequiredMutationBodiesReachTheWire$' -count=1` failed for all six commands because the required schema properties/CLI flags were absent.
- Wrong paths: `go test -timeout 20m ./internal/connectors/engine -run '^TestGitHubCorrectedCommandDeclarations$' -count=1` failed on all three Actions aliases and on the user-project numeric identifier contract.
- False success: `go test -timeout 20m ./internal/connectors/engine -run '^TestWriteDeleteMissingOkStatusDoesNotCountAsWritten$' -count=1` failed because the 404 returned no error and counted `RecordsWritten=1`. The app-level org webhook test has the same red behavior.

## Green checkpoint — 2026-08-18

- Exact integers: state decoding now preserves JSON number lexemes and interpolation renders numeric values in decimal form. `TestGitHubConnectorCommandPreservesExactLargeIntegerPathAfterPlanReload` proves `9007199254740993` survives plan persistence, and `TestInterpolatePathPreservesIntegerIDsWithoutScientificNotation` covers both that exact integer and the recorded `667233046` float regression.
- Required bodies: the six write schemas now expose their provider-required fields, their CLI commands map typed flags into those fields, and `TestGitHubRequiredMutationBodiesReachTheWire` proves the resulting JSON bodies at a provider double.
- Wrong paths: the three mislabelled Agent commands bind to the existing Actions actions and `/actions/` paths. The singular user-project command now requires a numeric `user_id`; `TestGitHubCorrectedCommandDeclarations` covers all four declarations.
- False success: allow-listed delete 404s now return `RecordsUnchanged=1`, never `RecordsWritten=1`. Command execution rejects that incomplete mutation acknowledgement, while the closed issue-label cleanup retains explicit already-absent idempotence. The focused engine/app tests and full connector conformance suite pass.
