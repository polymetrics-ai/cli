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

## CI handoff repair review

Scope: the PR #4198 repair to the OS process-list/temporary-artifact proof and
its updated TDD/verification evidence.

Mode: inline/manual review after the failed CI run. The review checked the
request lifecycle rather than accepting a longer completion wait: `sync.Once`
holds only the first authenticated provider call; later declared certification
calls receive ordinary TLS responses and cannot block on the one-shot signal.
The first response is written, flushed, and observed to finish before the test
waits for the child report — a child-side state transition that follows
provider-response consumption. Cleanup releases the handler before closing the
test server, then cancels and reaps the external child after report persistence.
All observed values are status/error/path states; no credential is added to a
log, diagnostic, artifact, or assertion message.

Result: no Critical, Warning, or Info findings. Focused repeat coverage,
`go test -timeout 20m ./internal/cli`, `go test -timeout 20m ./cmd/connectorgen`,
and `go vet ./...` pass.

## CI complete-state settlement review

Scope: the second PR #4198 CI failure at the now-moved report-persistence
assertion and the resulting redesign of the OS process-list/temporary-artifact
proof.

Mode: inline/manual completion of the generated `code-review` prompt. The
phase is not registered in the roadmap and this single-worker task may not
spawn a review role.

The review enumerated the test's actual post-release observables: the held
provider response is flushed, the held handler has returned, the fresh child
has exited, `recurly.json` is present and parseable as a completed passing
report, and the test-owned `pm-certify-external-*` build directories are gone.
The revised test checks each condition independently in one bounded poll and
keeps the existing five-second handler and thirty-second settlement bounds. It
does not cancel a merely progressing child or lengthen any timeout. A child
error is observed concurrently with report persistence rather than hidden by a
file-only wait; diagnostics contain only booleans and fixed state names.

Result: no Critical, Warning, or Info findings. The focused test passed three
repeat runs under `-parallel=4`; the earlier Redis connection-refused logs are
confirmed deliberate unreachable-endpoint fixture coverage outside this test.
