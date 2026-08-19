# Issue #4193 discussion log

## Inputs considered

- Captain priority in `data/production-parity-shared-context.md`: all commands
  must be user-friendly through a scalable implementation.
- Issue #4193: leaf help must render before project, credentials, and required flags.
- Existing router: `legacyLeafManualTopic` hardcodes only `etl run` and
  `connections create`.
- Built-binary baseline: many leaf `--help` requests either open `.polymetrics`,
  reject required flags, or execute a command instead of rendering documentation.

## Resolution

Use a single wrapper-level help interception path. Keep raw approval-carrier
validation before it because that validation protects a closed stdin-only approval
protocol and must not be masked by help. Treat unknown commands as usage errors;
do not let help make invalid command paths executable. Keep dynamic connector help
on its existing declaration-owned path.

No product decision is outstanding and no external credentialed action is required.
