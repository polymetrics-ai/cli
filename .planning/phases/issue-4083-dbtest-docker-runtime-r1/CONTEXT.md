# Context — Issue #4083 dbtest Docker runtime

## Decision record

`internal/connectors/native/dbtest` remains a single `Config`-driven owner of
ephemeral database containers. The caller must declare both a runtime
(`docker` or `podman`) and a direct local Unix endpoint; the harness must not
infer a runtime from the endpoint, environment, or either CLI's global
context/default.

The current MySQL test's enabled-without-endpoint `Skip` is a false-pass risk.
It will stay opt-in when `POLYMETRICS_DATABASE_INTEGRATION` is unset, but once
enabled it must fail clearly unless the runtime and matching local endpoint
are supplied.

An externally supplied PostgreSQL endpoint is intentionally excluded. It
would relinquish the harness's proof of the server image/version, daemon
settings, extensions, storage durability, test cleanliness, resource
ownership, and unconditional cleanup. It does not solve container runtime
selection and would be a separate test category if a future issue needs one.

## Safety boundary

Named connections, remote schemes, and implicit/default runtime contexts stay
refused. Every runtime command receives the configured endpoint explicitly.
The existing identity and capacity proof remains mandatory: Docker may proceed
only when its chosen daemon can prove an identity and an image-store free-space
measurement. A Docker VM path not visible to the host needs an explicitly
configured, pre-cached, locked-down daemon-side probe; otherwise it still fails
closed rather than weakening the pre-pull guard.

## Lifecycle fallback

The canonical delivery contract prohibits spawning lifecycle roles in this
worktree. `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review` are executed inline, with their generated
prompts resolved through `scripts/gsd`.
