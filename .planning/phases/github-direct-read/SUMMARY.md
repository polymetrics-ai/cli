# GitHub Direct Read Summary

Issue: #38
Branch: `feat/38-github-direct-read`

## Delivered

- Added a narrow direct-read connector contract and result envelope.
- Added engine-backed direct reads for fixed GET operations only.
- Added safe path-variable resolution for connector config and mapped command flags.
- Added response-size cap, timeout, JSON parsing, and redacted HTTP error wrapping.
- Wired `pm github repo read-file` and `pm github repo read-dir` to the fixed GitHub contents
  endpoint.
- Kept `api_surface.json` as the authoring validator and used validated `cli_surface.json` refs for
  runtime because production embeds intentionally omit api-surface ledgers.

## Safety

- Arbitrary full URLs are rejected.
- Mutation methods are rejected.
- Missing path variables fail before network I/O.
- Repository content paths reject traversal and absolute paths.
- Direct reads send no request body.
- Reverse ETL writes remain separate and gated.

## Source

- GitHub REST repository contents endpoint:
  https://docs.github.com/en/rest/repos/contents#get-repository-content
