# Inline code review: Issue #4328

## Scope reviewed

- `cmd/connectorgen/validate.go`
- `cmd/connectorgen/sourceprojection.go`
- Regression tests and generated HubSpot/certification artifacts

## Result

No actionable findings.

The validator now obtains request sensitivity only from schema `x-secret` markers (including an enclosing JSON field that contains a declared secret) and request-side `SensitivePolicy` declarations with a non-empty `InputMode`. This excludes response-only redaction policies. Existing GraphQL query compatibility remains closed; GraphQL mutations, REST parameters, nested REST body paths, write records, and source-projected fields use the same declaration-owned rule.

The source projector preserves `x-secret`, emits `env_only` for secret-bearing fields, and preserves the pre-existing bare-string behavior on its hook and concrete-variant paths. `surface-sync --check` is clean after the review correction.

## Safety and generated artifacts

- No change under `cmd/connectorgen/sourceimport.go`.
- No change to `internal/connectors/defs/github/rate_limits.json`.
- GitHub source/descriptor bytes and SHA-256 values remain exact; see `VERIFICATION.md`.
- Generated HubSpot manual/skill files and the certification subject were refreshed using their repository-owned generators.

## Automated review route

The direct, non-draft PR will use `claude_auto` on open. Its status is `pending` until GitHub records the automatic review. No manual Claude or Copilot request was made before that trigger.
