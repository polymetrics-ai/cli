# Local review — #4352

## Scope and source-bound review

- No arbitrary HTTP method, URL, path, header, request body, shell, or curl escape hatch was added.
- The command source ID must equal the selected declaration's source ID; engine preflight also compares exact GET method and connector-relative path before dispatch.
- A source-bound direct read remains on the fixed `OperationDirectRead` executor and reaches the existing credential check only after no-network preflight.
- A source-bound collection cannot claim ETL solely because it is a GET: it needs a unique declared stream ETL composite, matching route, records path, schema, and pagination.
- Existing ETL commands do not require the new source binding. This was verified after a full runner regression caught and corrected the overly broad check.

## Findings

No unresolved code finding. The two repository-wide commands noted in `VERIFICATION.md` exceeded this local runner's limit and are intentionally left to PR/CI rather than treated as passing.

## Required external review

Open a PR to `main` and let the repository's configured Claude review run. Confirm it covers the final commit range; disposition any actionable findings before human merge. The PR remains unmerged by this task.
