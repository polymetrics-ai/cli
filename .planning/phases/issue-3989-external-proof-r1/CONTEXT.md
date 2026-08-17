# #3989 external-binary certification proof context

## Residual delivery addendum — live and OS-level evidence

## Task Delivery Header

- Issue: Refs #3989 — Certification: add external-binary proof capture and ephemeral fingerprint-first credentials.
- Base branch: `integration/4015-mvp-flat-r1` (refreshed and rebased at `4967fa2a0`).
- Merges into: `integration/4015-mvp-flat-r1` → `main`.
- Delivery: A non-draft PR from `fm/cli-3989-live-proof-residual-r1` is open against the stated base, API-read back after creation, with the residual tests, live smoke, generators, and scoped repository gates recorded.
- Working branch: `fm/cli-3989-live-proof-residual-r1`.
- Task: Retain an actually executed disposable-identity GitHub smoke and add the missing opaque-body, OS process-list/temporary-artifact, and single-run fingerprint-semantics evidence without rebuilding #4159's delivered proof boundary.
- Verification: Run the focused certify and CLI packages; execute the opt-in GitHub smoke with only a disposable token supplied through its environment variable; run `go test -timeout 20m ./cmd/connectorgen`, `go vet ./...`, `go build ./cmd/pm`, all individual `make verify` gates, the docs generator twice for byte stability, and the final diff check.

## Residual Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Live GitHub smoke actually executes | live | The fresh external binary makes at least one authorized GitHub request, creates exactly one sanitized proof, attests its exact process streams, and has no vault or raw token in the project tree. |
| Opaque request and response canaries are substituted | fake HTTPS provider | A credential canary placed in non-JSON request and response bodies appears only as its same root-salted fingerprint in both serialized proof locations. A local TLS provider is necessary to produce deterministic opaque bytes. |
| OS-visible command line and temporary paths contain no raw secret | fake HTTPS provider + local OS inspection | While the fresh child is held at a confirmed HTTPS request, `ps` command output and the task's project/build temporary paths contain no token bytes. |
| One run proves fingerprint semantics | deterministic proof writer | Credential A repeated in several locations uses one marker, credential B uses another marker, and the root salt itself is absent from the serialized proof. |

## Original merged-slice header — historical

- Issue: Closes #3989 — Certification: add external-binary proof capture and ephemeral fingerprint-first credentials.
- Base branch: `integration/4015-mvp-flat-r1`; delivery PR targets that branch.
- Working branch: `fm/cli-3989-external-proof-r1`.
- Scope: the shared certification runner, its proof boundary, and focused tests/artifacts only.
- Non-goals: #4125, #4136, #4090, #4154; connector feature work; changing ordinary persisted credential behavior.

## Manual-GSD fallback

`scripts/gsd doctor`, command-source resolution, and all five generated prompts succeeded. `gsd-sdk query init.phase-op 3989` reports `phase_found: false`, so the installed adapter cannot execute a numbered phase. Runtime role spawning is also unavailable for this task. This directory is the inline/manual fallback required by the delivery contract; it preserves discuss → plan --tdd → execute → verify-work → code-review evidence without weakening any gate.

## Locked decisions

- Certification credentials are resolved only from declared environment or stdin sources; raw values are fingerprinted immediately and never sent to the vault, a profile file, command arguments, reports, proof artifacts, or logs.
- The runner builds and invokes a fresh external `pm` binary. The existing in-process harness remains fixture-test infrastructure only and is not accepted as live evidence.
- Every observed HTTP exchange retains structural data and non-credential payload values, but substitutes every exact prepared credential occurrence before any serialization. Missing observations, a failed child invocation, non-full-parity attestation, or an incomplete substitution reject evidence.
- Bodies have an explicit size ceiling and truncation metadata. A proof needing omitted bytes is rejected. Redirect, retry, provider-error, JSON, opaque, and binary paths use the same observer/substitution boundary.
- Real-GitHub smoke is opt-in and must not request, display, or store credentials. The local HTTPS-provider test is the deterministic acceptance proof; a live smoke is executed only when its preconfigured secret reference is already available.

## Canonical references

- `docs/architecture/connector-certification-design.md` — certification stage and evidence contract.
- `cmd/connectorgen/certificationproof.go` — existing HMAC proof model and writer to move/reuse.
- `internal/connectors/certify/stages_source.go` — persisted/in-process runner path being replaced.
- `internal/connectors/certify/record.go` — test-only recorder to supersede for accepted live evidence.
- `internal/connectors/certify/cliharness.go` — fixture-only harness boundary.
- `internal/cli/certify_cli.go` — public certification command wiring.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-mvp-verify-certification-r1/report.md` §3989 — MVP audit and residual scope.

## Code context

- `certify.Runner` creates one temporary root and owns the sequential stage lifecycle.
- `cmd/connectorgen/certificationproof.go` has the all-or-nothing completed-evidence validator but no production caller.
- `connsdk.Requester` already centralizes HTTP retries and redirect policy; observer integration must preserve its existing response semantics.
