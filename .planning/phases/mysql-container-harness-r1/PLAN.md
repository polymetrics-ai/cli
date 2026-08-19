# Plan — MySQL container harness R1

## Scope and ownership

This phase lands **MySQL only** on a reusable database-container test harness. Production changes
are the MySQL native connector/bundle and the shared SQL TLS helper. The internally proven
binary-log reader remains outside the public connector capability surface until a production runtime
entrypoint exists. The only PostgreSQL follow-up is its
connection configuration adapter so the captain-mandated TLS option shape is actually enforced by
Check, Catalog, and Read; it deliberately does not edit the incoming PostgreSQL write path.

The harness is test support under `internal/connectors/native/dbtest`. A second engine is supplied
as `dbtest.Config`; no copy of the harness is needed. Engines use a one-slot default semaphore;
bounded parallelism is explicit opt-in only.

## Design

1. `dbtest` accepts one pinned image, a direct local Unix Podman endpoint, engine data directory,
   and engine/container arguments. It allocates collision-resistant container/volume/run image
   names, publishes `127.0.0.1::<engine-port>`, proves endpoint identity and target-store capacity,
   records disk free before and after, and refuses default-host-port mappings.
2. `Start` claims ownership before each indeterminate create, and `Close` is idempotent. It cancels
   in-flight startup on interrupt, waits for that create sequence, and then attempts all cleanup in
   order: container, volume, run image. The source image is never removed, and a later cleanup action
   still runs after an earlier error.
3. MySQL validates identifiers, does dynamic catalog discovery, reads bounded keyset pages ordered
   by a primary key (or `(cursor, primary_key)`), and preserves an explicit empty cursor state. Its
   row-based/full-image binlog reader remains an internal proof rather than a declared executor.
4. The tagged integration test starts exactly one real MySQL 8.4.11 server, seeds five deterministic
   rows with `page_size=2`, asserts check/discovery/full/incremental records plus internal CDC
   events, and defers
   `Close` before any assertion. Its failure message retains only the harness's sanitized stage.
5. SQL transport security has the one shared shape: `sslmode`, `sslrootcert`, and
   `sslservername`. MySQL applies it to normal and replication clients; PostgreSQL feeds it into
   pgx's pool configuration for every normal database operation.

## TDD sequence and evidence

1. **Red:** missing scoped direct Podman endpoint, harness lifecycle, binlog declaration, and MySQL
   implementation contracts initially failed because no harness/native connector existed.
   **Green:** focused harness/native/engine tests passed after implementation.
2. **Red:** after rebase, PostgreSQL's embedded definition rejected canonical `disabled` before the
   native connection code could enforce it. **Green:** the shared modes and CA/server-name fields are
   accepted and `TestPostgresPoolConfigUsesSharedTransportSecurityOptions` proves strict pool setup.
3. **Green boundary proof:** `TestReadIsETLNotPagewiseDirectRead` proves #3902 page context is not
   being silently bypassed: MySQL has no direct-read surface; its ETL reader drains its own complete
   SQL pages and the live test proves more than one query page.
4. **Red:** pipeline-custody recovery found `connectorgen gen` blank-imported `dbtest` and the
   shared `sqltls` library into the production native set. **Green:**
   `TestGen_NativesetImportsRuntimePackagesAndExcludesSupportLibraries` now distinguishes runtime
   connector packages from support libraries, and the generated native set is regenerated from that
   source of truth.
5. **Red:** the `verify-ca` TLS configuration placed its manual chain verification in
   `VerifyPeerCertificate`, which Go skips for resumed sessions when built-in verification is disabled.
   **Green:** focused `sqltls` tests require `VerifyConnection` to reject an untrusted chain and
   accept a trusted chain without hostname verification, then prove it is called during a resumed
   TLS 1.2 session.
6. Run generated-surface regeneration, focused tests, CLI tests, vet/build, individual verify gates,
   a live direct-endpoint Podman proof, and an inline review. Record exact commands/results in
   `VERIFICATION.md` and PR-ready facts in `PR-BODY.md`.

## Dependency legitimacy gate

The sole added module is `github.com/go-mysql-org/go-mysql v1.16.0`, used for both client protocol
and binlog replication. Its upstream release and repository activity, MIT licence, complete module
delta, current vulnerability scan, and measured dependency-introduction binary delta are recorded
in `PR-BODY.md`. No other direct dependency is added.

## Acceptance facts

- The real server is never assumed at a default host port.
- Seed data exceeds one SQL page and supports exact incremental assertions.
- Internal change-capture proof asserts returned insert/update/delete records and acknowledgement-gated state.
- Failure and interrupt teardown share the idempotent cleanup path.
- A missing live engine cannot silently pass; an absent opt-in/endpoint emits a visible skip.
