---
phase: issue-4083-dbtest-docker-runtime-r1
issue: 4083
coverage:
  - id: D1
    description: Explicit Docker and Podman runtime selection pins every client invocation to the configured Unix endpoint.
    requirement: RUNTIME-SELECT
    verification:
      - kind: unit
        ref: internal/connectors/native/dbtest/harness_test.go: TestRuntimeCommandPinsTheConfiguredEndpoint
        status: pass
      - kind: unit
        ref: internal/connectors/native/dbtest/harness_test.go: TestNewRequiresAnExplicitContainerRuntime
        status: pass
    human_judgment: false
  - id: D2
    description: Unsafe, remote, named, and incomplete enabled runtime configuration fails rather than silently using a default or skipping.
    requirement: FAIL-CLOSED
    verification:
      - kind: unit
        ref: internal/connectors/native/dbtest/harness_test.go: TestNewRejectsUnsafeEndpoints
        status: pass
      - kind: integration
        ref: POLYMETRICS_DATABASE_INTEGRATION=1 tagged MySQL selector
        status: pass
    human_judgment: false
  - id: D3
    description: Docker target identity and locally measurable image-store capacity are re-proven before mutations.
    requirement: TARGET-PROOF
    verification:
      - kind: unit
        ref: internal/connectors/native/dbtest/harness_test.go: TestStartUsesDockerTargetIdentityAndCapacity
        status: pass
      - kind: unit
        ref: go test -race -timeout 20m ./internal/connectors/native/dbtest
        status: pass
    human_judgment: false
  - id: D4
    description: Live Docker and Podman reference proof.
    requirement: LIVE-PROOF
    verification:
      - kind: integration
        ref: TDD-LEDGER.md Live Docker and Live Podman rows
        status: unknown
    human_judgment: false
---

# Summary — Issue #4083 dbtest Docker runtime

The shared `dbtest` harness now accepts an explicit `RuntimeDocker` or
`RuntimePodman` value in its existing `Config`. Docker commands use the
configured socket with `--host`; Podman commands use it with `--url`. No
runtime is inferred, and named/remote socket forms remain rejected.

The tagged MySQL reference proof now requires
`POLYMETRICS_CONTAINER_RUNTIME` plus `POLYMETRICS_CONTAINER_ENDPOINT` whenever
the opt-in is enabled. Its former enabled-but-missing-endpoint skip is now an
actionable test failure. README and the durable harness guidance document the
two exact variables and existing safety constraints.

Docker target proof binds the daemon ID and `DockerRootDir` to the configured
socket and measures that root before every daemon command. An inaccessible or
unmeasurable Docker root fails closed before a pull, preserving the existing
capacity guard. Podman target and machine-forward behavior is retained.

## Live-runtime limitation

Neither live container proof was observed in this worktree. Docker's explicit
`unix:///Users/karthiksivadas/.docker/run/docker.sock` did not report a running
daemon identity/image-store path; Podman had no discovered direct local socket.
The exact commands and output-class evidence are recorded in `TDD-LEDGER.md`.
This is an environment limitation, not a passed live test.

## Lifecycle and review

The generated `scripts/gsd` prompts were executed inline because the canonical
contract forbids lifecycle role spawning. Manual verify-work and standard code
review found no unresolved code, safety, scope, or documentation issue. The
remaining delivery gate is no-mistakes PR/CI validation; no merge is authorized.
