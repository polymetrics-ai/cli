# Verification — Issue #4083

Status: Docker live proof passed locally; final PR/CI gate must be rerun for the
Docker VM-capacity follow-up.

## Required evidence

- [x] Foundation owner issue #4083 created rather than expanding #3976's
  PostgreSQL connector lane with shared test infrastructure.
- [x] GSD adapter and command sources resolved; inline/manual lifecycle
  fallback recorded because role spawning is forbidden.
- [x] Required skills loaded.
- [x] CI remediation red evidence recorded: the exact Go 1.25.12
  `govulncheck` command found seven reachable standard-library vulnerabilities
  fixed in Go 1.25.13.
- [x] CI remediation green evidence: `go.mod`, both security-job setup pins,
  and the `govulncheck` `GOTOOLCHAIN` pin agree on Go 1.25.13; the exact scan
  passed with zero reachable vulnerabilities.
- [x] Red test evidence recorded: focused dbtest selector exited 1 against the
  hard-wired Podman baseline because `New` accepted no explicit runtime.
- [x] Green focused unit evidence recorded: explicit runtime values, endpoint
  pinning, shared unsafe-endpoint refusals, Docker identity/capacity proof, and
  Docker parser rejection cases pass.
- [x] dbtest and MySQL non-live package regression evidence recorded.
- [x] Docker VM capacity red/green evidence recorded: a pre-cached, pinned,
  locked-down daemon-side probe is required when the selected Docker root is
  not measurable from the client host; it is never pulled by `dbtest`.
- [x] Review follow-ups passed: a pre-existing Docker capacity-probe name is
  refused before cleanup ownership; a probe that appears after the name check
  needs its per-run label and inspected immutable ID before removal; and raw
  runtime whitespace/control characters retain the same environment-specific
  configuration guidance.
- [x] Docker live proof passed through
  `unix:///Users/karthiksivadas/.colima/default/docker.sock`; the exact tagged
  command, PASS output, TLS subtests, and before/after capacity values are in
  `TDD-LEDGER.md`.
- [x] Podman live proof remains unavailable: no local Podman VM exposes an
  explicit direct Unix API endpoint, and no global default was queried.
- [x] README/AGENTS guidance checked for exact runtime/environment and safety
  wording.
- [x] Inline GSD verify-work and code-review evidence recorded in `UAT.md`,
  `SUMMARY.md`, and `REVIEW.md`; no unresolved finding.
- [ ] no-mistakes final PR/CI evidence recorded; no merge performed.

## External-endpoint verdict

Rejected for this issue. An external PostgreSQL server is neither a Docker nor
Podman test target and would weaken the harness's image/version, settings,
extension, storage, cleanliness, ownership, and cleanup guarantees. It needs
its own opt-in test contract if ever required.
