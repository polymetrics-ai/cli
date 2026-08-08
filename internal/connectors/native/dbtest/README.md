# Database container harness maintainer guide

`dbtest` is opt-in test support for native database connectors. Add an engine through `Config`; do
not copy the harness or add a standalone Podman command path.

## Running the MySQL proof

The reference test uses pinned `docker.io/library/mysql:8.4.11`, a dynamically allocated loopback
port, and deterministic multi-page data. It checks connector reachability, catalog discovery, full
and incremental reads, every TLS mode, and the internal MySQL row-event reader.

```bash
POLYMETRICS_DATABASE_INTEGRATION=1 \
  POLYMETRICS_PODMAN_ENDPOINT=unix:///run/user/1000/podman/podman.sock \
  go test -tags=databaseintegration -count=1 -v ./internal/connectors/native/mysql
```

`POLYMETRICS_PODMAN_ENDPOINT` is a direct local Unix Podman API endpoint. Named connections and
remote endpoints are refused. Every invocation uses that endpoint directly, so the global Podman
default is never read or changed.

## Capacity and cleanup

Before startup and before every command directed at the daemon, `dbtest` reads identity and image-store data from
the selected endpoint. The endpoint's reported socket must match the configured socket, and its
reported image-store path must be measurable locally. An unprovable endpoint or capacity fails
before the pull, including when the source image is cached. An absent source image needs three times
`ExpectedImageBytes` free.

The harness owns only its generated container, volume, and run-image reference. The source image is
always retained. Cleanup is unconditional and idempotent, including failure and interrupt paths, and
the interrupt handler remains armed until the final generated-resource removal returns.

The MySQL TLS proof copies the container-generated CA certificate through the harness's scoped
container copy operation, then runs a live `verify-ca` session and checks the server's negotiated
`Ssl_cipher`.

## Adding an engine

Define a pinned image, container port, data-volume path, expected image bytes, and direct local
Podman endpoint plus engine and container arguments in one `Config`. Keep the test tagged
`databaseintegration`, create a non-default loopback endpoint, seed data that exceeds one page, and
defer `Close` immediately after `New`. Engines run sequentially unless a bounded
`SetMaxConcurrentEngines` value is deliberately selected.
