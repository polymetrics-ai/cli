# Docker Hub live E2E follow-up — code review

## Method

Inline standard review, because the canonical parent-worker contract forbids
spawning a separate GSD reviewer role. The review followed the generated
`/gsd-code-review dockerhub-parity-sweep-r1` prompt after resolving sources with
`scripts/gsd sources code-review`.

## Reviewed scope

- `internal/connectors/connsdk/http.go`
- `internal/connectors/connsdk/http_test.go`
- `internal/connectors/defs/dockerhub/writes.json`
- `internal/connectors/defs/dockerhub/write_paths_test.go`
- `.planning/phases/dockerhub-parity-sweep-r1/*`

## Findings

No Critical, Warning, or Info findings.

The ALPN change clones a caller-provided TLS config before setting only
`NextProtos`, retaining certificate and other caller settings while preventing an
HTTP/2 server from negotiating a protocol incompatible with the strict HTTP/1
one-shot writer. The focused TLS/HTTP2 regression and race run pass.

Every Docker Hub write action is covered by the bundle-level path regression: no
provider-rooted `/v2` prefix remains, no raw OpenAPI `{parameter}` remains, and each
record-backed path field is declared. `connectorgen validate`, surface sync, command
preflight, Docker Hub conformance, and the rebuilt binary's 54-route help sweep pass.

The only incomplete user acceptance item is external: the supplied Docker Hub PAT
returns HTTP 403 for private-repository creation and account/organization reads, and
the SCIM-only credential returns HTTP 401. See `VERIFICATION.md` for the complete
redacted command matrix and named dependencies.
