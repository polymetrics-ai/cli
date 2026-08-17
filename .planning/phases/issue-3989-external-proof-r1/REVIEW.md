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

Scope: the second PR #4198 CI failure at the then-current report-persistence
assertion and the intermediate redesign of the OS
process-list/temporary-artifact proof. This was superseded by the later
child-side observation inversion.

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

## CI child-side observation inversion review

Scope: the third PR #4198 CI failure, where every parent-side settlement
observable was false, and the decision to replace outside observation with a
child-side security snapshot.

Mode: inline/manual completion of the generated `code-review` prompt. The
phase is not registered in the roadmap and this single-worker task may not
spawn a review role.

The review rejected a fourth timing repair. The fresh external child now calls
`ps` for its own PID, captures its own argv, and scans the project root, runner
workdir, and fresh-binary directory immediately after it resolves credentials
into memory. It writes a JSON artifact only after redacting every prepared
value and checking the serialized result contains none. The parent reads that
artifact only after normal process completion. The test requires the process
command, safe argv, two `TMPDIR`-owned locations, and at least one scanned
temporary file, so a placeholder or post-cleanup empty scan cannot pass.

The artifact request is an integration-test evidence path, not a command-line
switch and never skips the proof. No timeout was added or widened. Result: no
Critical, Warning, or Info findings.

## Rebased Recurly vehicle review

Scope: PR #4198 after rebase onto `integration/4015-mvp-flat-r1` at
`4a0289bcc`, which made the one-route Recurly TLS fixture return exit 1.

The revised all-stage roll-up is correct: the fixture authenticates a real
`/accounts` call but cannot provider-certify Recurly's declared write surface.
The OS-boundary test does not claim otherwise. It now requires that exact
incomplete-parity result after the fresh child has captured its own process,
argv, and artifact state. This preserves the test's security subject without
weakening certification, adding a skip, or changing any timeout.

Result: no Critical, Warning, or Info findings.

## CLI package-capacity review

Scope: the fixed 20-minute `internal/cli` package deadline after CI displayed a
Bahmni test at timeout. The manual review first used a complete verbose package
timing run: the package passed in 706.417s; Bahmni was 39.140s, the two
external-child proofs were 118.770s and 118.210s, and the 17,800-case dynamic
leaf-help sweep was 22.500s. The displayed test was not diagnosed as the hang.

`TestEveryDynamicConnectorLeafHelpRendersWithoutDispatch` now constructs every
case before scheduling it. The shared registry and command surfaces are loaded
once and then read only; each parallel case holds an immutable path copy and
allocates its own arguments/manual output. No test counter, buffer, project
directory, credential, transport, or global state is shared between parallel
subtests. Test names now include connector, command, and help spelling, which
improves a failure's attribution.

The focused verbose run retains exactly 17,800 variants, the focused `-race`
run passes, and the full package passes in 694.432s without a timeout change,
test sampling, or coverage deletion. Result: no Critical, Warning, or Info
findings.

## Live diagnostic review

Scope: the post-rebase safe diagnostic path for the external GitHub smoke.

Mode: inline/manual review. The report and child-stream diagnostics use the
same in-memory HMAC marker shape as accepted proofs, while their salts are
neither persisted nor returned. `assertKind` is the common boundary for typed
CLI errors, so any source, flow, resume, or schedule mismatch retains the
non-secret category/code/message only after fingerprint redaction. The report
gate returns exactly the first non-passing stage before refusing proof
creation; it never serializes a failed proof or relays raw child stdout/stderr.

The live smoke's non-secret coordination account was supplied to satisfy—not
bypass—GitHub's declared rate policy. Full parity nevertheless ended at
different non-passing stages on two final runs, so no claim of a provider
success or accepted evidence was made. Planted-secret tests cover exact,
base64, and URL forms and `git diff --check`, full `internal/connectors/certify`,
full `internal/cli`, the required consumer package, vet, and build pass.

Result: no code-quality or credential-boundary findings. Live full parity is
an external acceptance blocker, not a condition this review can waive.
