# GitHub Direct Read

Issue: #38
Branch: `feat/38-github-direct-read`
Parent issue: #44

## Goal

Add a constrained direct-read execution path for fixed GitHub API operations that are useful CLI
commands but not natural ETL streams.

## Pilot Scope

- Implement `repo read-file` and `repo read-dir` as fixed direct reads against
  `GET /repos/{owner}/{repo}/contents/{path}`.
- Require the command to reference an existing `api_surface` operation row.
- Execute only `intent=direct_read`, `availability=implemented`, `method=GET`,
  `operation.model=direct_read` rows.
- Reject arbitrary full URLs, mutation methods, missing path variables, path traversal, and oversized
  responses.
- Return JSON output by default and keep human output as formatted JSON.

## Safety Rules

- No raw API escape hatch.
- No arbitrary host or absolute URL.
- No request body.
- No mutation method.
- Path parameters come only from connector config or mapped command flags.
- Direct reads have a bounded response size and timeout.
- HTTP errors and error text must be redacted.
- Reverse ETL writes remain plan, preview, approval, execute.

## Red/Green Plan

1. Add red commandrunner tests for fixed direct-read commands and rejection cases.
2. Add red engine tests for httptest direct read, full URL rejection, mutation rejection, missing
   path variables before network I/O, path traversal, and oversized response handling.
3. Add red CLI tests for `pm github repo read-file --path README.md --json`.
4. Implement narrow direct-read interfaces and engine execution.
5. Wire command metadata for GitHub `repo read-file` and `repo read-dir`.
6. Validate JSON, connectorgen, targeted Go tests, vet, and build.

## Human Gates

- Any generic raw API command.
- Any direct write or mutation execution.
- Any auth scope refresh or secret access outside credential resolution.
- Parent PR merge into `main`.
