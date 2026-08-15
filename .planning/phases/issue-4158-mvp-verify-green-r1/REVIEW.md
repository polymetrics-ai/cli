# Code Review — fixture migration

`code-review --depth=deep` was performed inline after local verification; the
task contract forbids lifecycle-role spawning.

## Reviewed paths

- `internal/cli/github_flow_roundtrip_test.go`
- `internal/cli/flow_cli_test.go`
- GSD/TDD evidence under this phase directory

## Findings

No blocking correctness, authorization, secret-handling, or test-contract
findings. The change uses the reverse-plan ID that the acceptance test creates
and approves, leaves `read_back_stream` as the only action-local configuration,
and adds exact typed pre-I/O refusals for missing, revoked, and stale jobs.

## External review route

The direct PR targets non-default branch `integration/4015-mvp-flat-r1`.
Primary route is `claude_auto` on PR creation; no manual Claude command is
posted. Review coverage is pending GitHub's automatic record after the PR is
opened. Copilot is fallback-only if Claude becomes unavailable.
