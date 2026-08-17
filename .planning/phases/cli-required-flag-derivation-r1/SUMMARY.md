# Summary — CLI required-flag derivation r1

## Delivered

- `surface-sync` now derives CLI `required: true` from a matching required
  REST path parameter for every supported REST operation and every connector.
- GitHub regeneration changes 104 flag fields in 92 commands, closing the
  sweep's 92 product defects without a GitHub-specific code path.
- The all-bundle invariant now finds zero optional CLI flags mapped to required
  REST path parameters across all 552 connector bundles. No other connector
  required a generated-file change.
- Missing required flags are a typed pre-I/O usage refusal in the CLI.
- The not-applicable audit verifies all 50 GitHub entries: 26 `unsupported_api`
  declarations are provider-supported contradictions, one remains unsupported,
  and all 23 `unsupported_local` declarations remain correct. This PR does not
  silently reclassify any of them.

## Delivery evidence

See `TDD-LEDGER.md`, `VERIFICATION.md`, `NOT-APPLICABLE-AUDIT.md`, and
`REVIEW.md`. GSD prompts were executed inline under the recorded contract
fallback because this environment forbids role spawning.
