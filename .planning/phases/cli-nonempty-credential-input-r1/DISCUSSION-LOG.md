# Discussion log — non-empty credential input foundation

The Firstmate brief supplies the product decision: empty credential material is
invalid after the one documented stdin transport normalization and must not be
persisted or emitted. No provider-specific policy or escape hatch is allowed.

Resolved implementation choices:

1. Preserve a legitimate secret's bytes by removing one trailing terminal
   `\n` or `\r\n`, rather than using a broad trim.
2. Place the persistence invariant in the App boundary and the output
   invariant in the shared declarative authentication boundary; CLI input is
   an earlier, user-actionable guard.
3. Treat an absent optional auth selection as valid; only a selected
   credential-bearing declaration is required to carry non-empty material.
4. Keep the Twenty CRM bundle untouched. Twenty PR #4298 will consume the
   exact foundation commit recorded in `SUMMARY.md` after this branch commits.
