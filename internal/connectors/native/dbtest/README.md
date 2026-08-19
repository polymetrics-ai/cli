# Database container harness maintainer guide

`dbtest` is opt-in test support for native database connectors. Add an engine through `Config`; do
not copy the harness or add a standalone container-runtime command path.

## Running the MySQL proof

The reference test uses pinned `docker.io/library/mysql:8.4.11`, a dynamically allocated loopback
port, and deterministic multi-page data. It checks connector reachability, catalog discovery, full
and incremental reads, every TLS mode, and the internal MySQL row-event reader.

```bash
POLYMETRICS_DATABASE_INTEGRATION=1 \
  POLYMETRICS_CONTAINER_RUNTIME=docker \
  POLYMETRICS_CONTAINER_ENDPOINT=unix:///var/run/docker.sock \
  go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/mysql
```

Or select Podman explicitly:

```bash
POLYMETRICS_DATABASE_INTEGRATION=1 \
  POLYMETRICS_CONTAINER_RUNTIME=podman \
  POLYMETRICS_CONTAINER_ENDPOINT=unix:///run/user/1000/podman/podman.sock \
  go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/mysql
```

`POLYMETRICS_CONTAINER_RUNTIME` must be exactly `docker` or `podman`.
`POLYMETRICS_CONTAINER_ENDPOINT` must be a direct local Unix API endpoint for that selected
runtime. Named connections and remote endpoints are refused. Every invocation passes that endpoint
directly to Docker (`--host`) or Podman (`--url`), so neither global runtime default is read or
changed. Leaving the opt-in unset skips the test; setting the opt-in without both runtime inputs is
a test failure, not a skip.

## Recorded live evidence

Docker was live-proved through Colima's explicit direct socket with this command:

```bash
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -run '^TestMySQLContainerHarness$' -v ./internal/connectors/native/mysql
```

It passed catalog discovery, full and incremental reads, TLS, and CDC. Podman remains not
live-proved because no local Podman VM exposed an explicit direct Unix API socket; no global default
was read.

## Capacity and cleanup

Before startup and before every command directed at the daemon, `dbtest` proves the selected
target's identity and image-store path. Before it invokes the database-image pull, it also proves
the target image-store capacity. A direct Docker daemon must report a stable daemon ID and a
measurable `DockerRootDir`; a direct Podman daemon must report the configured socket and a locally
measurable image-store path. When Docker's safe root path is inside a local VM such as Colima and
is therefore not measurable from the client host, the engine's `Config` may name a pinned,
pre-cached `DockerCapacityProbeImage`. The harness first proves that image is already present,
resolves its local immutable image ID (it never pulls it), and runs that ID with `--pull=never` in
a uniquely named ephemeral probe through the same explicit endpoint. The probe has no network, a
read-only filesystem, dropped capabilities, no-new-privileges, a PID bound, and a read-only bind of
the proven daemon root; it accepts only the expected strict two-line POSIX `df -P -B1` schema and
its numeric `Available` byte count. A missing probe image, an unmeasurable path without one, or
malformed probe output fails before the database image is pulled. Podman 5.3 machine forwards are
also accepted when both sides report safe Unix sockets and the daemon reports numeric
`GraphRootAllocated`/`GraphRootUsed` capacity; the harness uses their difference rather than a host
path inside the VM. Any other socket mismatch, remote scheme, or unprovable capacity fails before
the pull, including when the source image is cached. An absent source image needs three times
`ExpectedImageBytes` free.

The MySQL reference config uses `docker.io/library/busybox:1.37.0` as that Docker VM probe. On a
fresh Colima image, bootstrap that one probe explicitly before running the test; the command remains
pinned to the same direct Unix endpoint and does not use a Docker default:

```bash
docker --host unix:///Users/you/.colima/default/docker.sock \
  pull docker.io/library/busybox:1.37.0
```

The harness owns only its generated database container, its container-bound anonymous data volume,
run-image reference, and the ephemeral capacity probe when that VM path is selected. The anonymous
volume is removed through the verified immutable database-container ID with `container rm --volumes`;
the harness never addresses a volume name. The source and probe images, and the generated run-image
reference, are retained. Cleanup of container-bound resources is unconditional and idempotent,
including failure and interrupt paths, and the interrupt handler remains armed until the final
generated-resource removal returns.

The MySQL TLS proof copies the container-generated CA certificate through the harness's scoped
container copy operation, then runs a live `verify-ca` session and checks the server's negotiated
`Ssl_cipher`.

## Adding an engine

Define an explicit `ContainerRuntime`, a pinned image, container port, data-volume path, expected
image bytes, and direct local endpoint plus engine and container arguments in one `Config`. Keep
the test tagged `databaseintegration`, create a non-default loopback endpoint, seed data that
exceeds one page, and defer `Close` immediately after `New`. Engines run sequentially unless a bounded
`SetMaxConcurrentEngines` value is deliberately selected.
