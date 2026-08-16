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

## Residual review

Mode: inline/manual completion of the generated `code-review` prompt. The phase remains absent from the roadmap and this single-worker task may not spawn a review role.

Scope: the residual test additions and their GSD evidence records. The review traced the real token path from `--from-env` into the fresh child environment, then checked the test's only OS inspection is `ps` command output (not the intentionally secret-bearing child environment). The held TLS request confirms the listed child is real; the scanner finds its command while requiring the canary absent, and walks only test-owned project and `TMPDIR` paths with bounded file windows.

The opaque proof test computes expected HMAC markers from the proof's root salt only in test memory, asserts A repetition/B separation across both opaque bodies, and asserts neither raw credential nor salt occurs in the artifact. The live smoke parses a passing GitHub report plus a sanitized observed 2xx response; it never renders either credential or proof body.

Result: no Critical, Warning, or Info findings. `git diff --check`, a final `go vet ./internal/cli`, and a diff secret-pattern check are clean. No production source, connector definition, declaration stage, direct-read candidate, or live-write path changed.
