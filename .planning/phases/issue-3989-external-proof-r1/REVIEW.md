# #3989 standard code review

Scope: `git diff origin/integration/4015-mvp-flat-r1...HEAD` for the external
proof, ephemeral-credential, CLI/help, documentation, and GSD evidence paths.

Mode: inline/manual fallback. The repository's GSD runtime cannot execute the
numbered phase or spawn a compatible review worker; the generated
`scripts/gsd prompt code-review 3989 --depth=standard` workflow was completed
against the changed source directly.

## Result

No Critical, Warning, or Info findings. The review traced the credential value
from environment/stdin intake through child argv replacement, in-memory app
resolution, observer capture, stream relay, and artifact sanitization. It also
checked that missing/truncated/over-limit proof inputs fail before an artifact
write and that stdout/stderr are matched via their HMAC fingerprints.

The provided PostgreSQL container lane independently reproduces the existing
#4158 managed-target durable-acknowledgement failure. PostgreSQL paths are not
part of this issue and were not changed.
