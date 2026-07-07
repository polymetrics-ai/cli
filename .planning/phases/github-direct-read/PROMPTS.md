# Prompt Snapshot

## Coordinator

Use GSD strict TDD with Codex `gpt-5.5` and `xhigh` reasoning. Implement issue #38 as a stacked
sub-PR against `feat/44-github-cli-parity`. Start from red tests and keep the first direct-read
implementation constrained to fixed GitHub repository contents reads.

## Backend

Add a narrow `DirectRead` interface, engine execution for fixed GET operation rows, commandrunner
routing for implemented direct-read commands, and CLI JSON output. Do not add raw API, generic HTTP,
mutation execution, or direct write support.

## Security

Reject absolute URLs, non-GET methods, missing path variables, path traversal, and oversized
responses. Keep credential resolution inside existing connector runtime config. Redact HTTP errors.

## Reviewer

Check that runtime execution is limited to validated command metadata, authoring validation still
links command refs to operation-ledger rows, and production embeds do not need full `api_surface`.
