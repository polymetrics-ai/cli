# TDD ledger — issue 4338

| Slice | Red evidence | Green evidence | Refactor / scope evidence |
| --- | --- | --- | --- |
| Asana absent action | **Red (2026-08-23):** the focused command failed to compile because the source-cited mutation-disposition model and application path did not exist. | **Green (2026-08-23):** the real source-descriptor marshal → source-projection → executable-coverage path accepts the separate `asana.create_access_request` `POST /access_requests` fixture with no action only after it carries a cited runtime gap; writes and CLI bytes remain unchanged. | Separate fixture; no count-only proof and no Asana definition edit. |
| Jira incomplete contract | **Red (2026-08-23):** the same focused command failed before any implementation. | **Green (2026-08-23):** the separate real Jira bulk-edit shape, `POST /rest/api/3/bulk/issues/fields`, retains its existing `partial` command and incomplete action contract; source projection synthesizes neither missing request fields nor a new command, and executable coverage accepts only the source-cited runtime gap. | Separate fixture; no Jira definition edit and no working command downgraded to `partial`. |
| Sentry mutation variants | **Red (2026-08-23):** source-cited mutation support did not exist before the original focused red run. | **Green (2026-08-23):** separate source-projection and executable-coverage fixtures retain cited gaps for the handoff's SCIM group `PATCH` and organization dashboard `POST`, without an action or command. | Sentry is not folded into read-only and no Sentry definition is edited. |
| Vercel-sized mutation batch | **Red (2026-08-23):** source-cited mutation support did not exist before the original focused red run. | **Green (2026-08-23):** a 159-operation batch spanning actual Vercel bulk-redirect PATCH, bulk-restore POST, nested file-write POST, and destructive cache POST shapes gives every operation a cited runtime gap; real source projection and executable coverage accept the batch without fabricating a command/action. | Behavioral per-operation loop, not a declaration-count oracle or connector waiver. |
| Closed safety boundary | **Red (2026-08-23):** the test could not compile because neither disposition nor mutation/read-only guard existed. | **Green (2026-08-23):** a fully materialized complete action is refused while retaining `availability: implemented`; an incomplete action with an `implemented` command claim is likewise refused; a DELETE with the read-only foundation fails both projection and executable coverage rather than suppressing the mutation finding. | Test actual projection/validation results rather than model shape. |
| Regression control | **Red (2026-08-23):** covered by the focused suite introduced before implementation. | **Green (2026-08-23):** GitHub's installed projection reports no drift and `writes.json`, `cli_surface.json`, and `api_surface.json` remain byte-identical. | No definition files are edited. |

Focused command, using an isolated Go build cache because another lane invalidated
the shared cache during compilation:

```sh
GOCACHE=<isolated-cache> go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestSourceProjectionSourceCitedNonExecutableMutationDispositionsCoverAbsentAndIncompleteActions|TestSourceProjectionSourceCitedNonExecutableMutationDispositionRejectsCompleteAction|TestSourceProjectionSourceCitedNonExecutableMutationDispositionRejectsImplementedIncompleteActionClaim|TestSourceProjectionSourceCitedNonExecutableMutationDispositionScalesAcrossVercelMutationShapes|TestSourceProjectionReadOnlyFoundationCannotSatisfyMutationCoverage|TestSourceProjectionSourceCitedMutationDispositionRejectsPOSTGraphQLQuery|TestSourceProjectionSourceCitedMutationDispositionLeavesExistingProjectionByteIdentical)$'
```

Red: build failed with undefined source-cited disposition symbols before production
implementation. Green: passed (`ok polymetrics.ai/cmd/connectorgen`).
