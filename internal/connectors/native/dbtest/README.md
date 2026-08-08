# Database container harness maintainer guide

`dbtest` is opt-in test support for native database connectors. Add an engine through `Config`; do
not copy the harness or add a standalone Podman command path.

## Running the MySQL proof

The reference test uses pinned `docker.io/library/mysql:8.4.11`, a dynamically allocated loopback
port, and deterministic multi-page data. It checks connector reachability, catalog discovery, full
and incremental reads, every TLS mode, and the internal MySQL row-event reader.

```bash
POLYMETRICS_DATABASE_INTEGRATION=1 \
  POLYMETRICS_DATABASE_OWN_MACHINE=1 \
  go test -tags=databaseintegration -count=1 -v ./internal/connectors/native/mysql
```

Choose a target mode: `POLYMETRICS_DATABASE_OWN_MACHINE=1` creates a task-owned machine; otherwise
`POLYMETRICS_PODMAN_CONNECTION` names an existing connection. The latter may use
`POLYMETRICS_PODMAN_MACHINE` when its local machine name differs from the connection name. There is
no default connection fallback. Container commands use the configured connection; machine commands
use the exact configured or generated machine name.

`NewMachine` uses `--update-connection=false`, records the global default before initialization,
and restores that recorded default only when initialization changed it and it remains unchanged by
another process. A failed or cancelled initialization is still removed on a detached deadline.

## Capacity and cleanup

For an absent source image, `Start` reads the target connection's image-store path and measures free
space on the matching local Podman machine before it calls `pull`. It requires three times
`ExpectedImageBytes`. Any remote, shared, or otherwise target whose image store cannot be measured
is refused before the pull; host `statfs` is reporting evidence only and never substitutes for
target capacity.

The harness owns its generated container, volume, and run image reference. A task-owned machine
also removes the source image reference and its machine. A caller-supplied or remote machine leaves
the source image in place unless `RemoveSharedSourceImage` is explicitly set. Every cleanup stage
runs after earlier failures. Only a task-owned machine is trimmed, twice, after image removal;
`HostDiskReleasedBytesEstimate` is the whole-run host free-space delta, not image reclaimability.

The MySQL TLS proof copies the container-generated CA certificate through the harness's scoped
container copy operation, then runs a live `verify-ca` session and checks the server's negotiated
`Ssl_cipher`.

## Adding an engine

Define a pinned image, container port, data-volume path, expected image bytes, explicit connection
and machine, plus engine and container arguments in one `Config`. Keep the test tagged
`databaseintegration`, create a non-default loopback endpoint, seed data that exceeds one page, and
defer `Close` immediately after `New`. Engines run sequentially unless a bounded
`SetMaxConcurrentEngines` value is deliberately selected.
