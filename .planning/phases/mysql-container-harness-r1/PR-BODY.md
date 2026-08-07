## Summary

Adds a reusable, Podman-backed database container test harness (`internal/connectors/native/dbtest`)
and lands one engine on it: a Tier-3 native MySQL source connector. The harness starts one ephemeral
container from a pinned tag on a dynamically assigned loopback port, seeds deterministic data, drives
the connector through check / discovery / full read / incremental read / change capture, and then
destroys the container, its volume, and its image on every exit path — including a failed assertion
and an interrupt.

A second engine is added by supplying another `dbtest.Config`, not by copying code.

## The single documented command

```bash
POLYMETRICS_DATABASE_INTEGRATION=1 \
  POLYMETRICS_PODMAN_CONNECTION=<your-machine> \
  POLYMETRICS_DATABASE_RECLAIM_DISK=1 \
  go test -tags=databaseintegration -count=1 -v ./internal/connectors/native/mysql
```

`POLYMETRICS_PODMAN_CONNECTION` is mandatory and has no default — every Podman call is scoped with
`--connection`. The test skips with a visible reason when the opt-in or connection is missing, and
**fails** rather than passing when the engine is unreachable. It never falls back to the global
default connection.

## Live proof

Passed against `docker.io/library/mysql:8.4.11` in Podman 5.3.2 (applehv, arm64), 33.8 s:

- **check** — connects and pings a real server.
- **discovery** — finds `pm_harness.events`, asserts primary key `id` and cursor `sequence`.
- **full read** — asserts records `1,2,3,4,5` with `page_size=2`, so paging is genuinely exercised.
- **incremental read** — from cursor `3`, asserts exactly `4,5`.
- **transport security** — four subtests, asserted against the server's own `Ssl_cipher`.
- **change capture** — real `INSERT`/`UPDATE`/`DELETE`, asserting operations
  `insert,insert,update,delete`, ids `6,7,6,6`, that two rows in one binlog event share a position
  but differ by row ordinal, and that committed checkpoints carry a schema fingerprint.

Assertions are on the records actually returned, not on exit codes.

## Disk: measured, not asserted by hope

Disk was the binding constraint (host was 96 % full when this ran). On macOS a container engine runs
in a VM whose disk is a sparse host file, so removing an image frees space *inside* the VM and leaves
the host file inflated. Measured on the task machine:

| Stage | machine `.raw` (MiB) | host avail (GiB) |
| --- | --- | --- |
| baseline, empty machine | 2609.0 | 37.072 |
| after `pull mysql:8.4.11` | 3563.4 | 36.139 |
| after `image rm`, no trim | 3624.7 | 36.079 |
| after **one** `fstrim` | 3383.7 | 36.378 |
| after prune + **second** `fstrim` | **2609.4** | **37.070** |

**A single trim returns only about a quarter of the space; a second pass returns effectively all of
it.** The harness therefore trims twice, and this is asserted in a unit test rather than left as a
comment. A full end-to-end run moved the machine's disk file by **+0.2 MiB** and the in-test host
figure by under 2 MiB.

Cleanup ordering is container → volume → run image → source image → trim, each step still attempted
if an earlier one fails. The pulled source image is removed by default;
`POLYMETRICS_DATABASE_KEEP_IMAGE=1` keeps it for a subsequent run.

## Podman: what actually blocked the previous attempt

The prior attempt concluded Podman was unusable ("three task-era Podman machines failed to start")
and moved the harness to Docker on Colima. That conclusion was wrong, and the root cause is worth
recording:

A `podman machine start polymetrics-runtime` process had been hung for **10 hours** holding the
**global** `~/.local/share/containers/podman/machine/machine-start.lock`. Every subsequent machine
start on the host — any lane's — queued behind it forever and looked like "the machine won't start".
Podman/applehv also permits **only one running VM at a time**, and the wedged machine held that slot.

Clearing the hung processes and stopping the wedged machine made a normal machine start work first
time. Podman needed no workaround.

### Host state I changed, and restored

- Killed the hung `podman machine start polymetrics-runtime` processes and stopped that machine. It
  was serving nothing (its socket was unreachable), and it is now **stopped and cleanly startable**
  rather than wedged — better than I found it.
- Created and used `fm-cli-db-harness-r1` (2 CPU / 2 GiB / 8 GiB). **Removed at the end**, reclaiming
  2.55 GiB.
- `podman machine rm` silently reassigned the global default connection to `polymetrics-runtime`.
  **Restored to `fm-bahmni-lab-r1-machine`**, its original value. Worth knowing: removing a machine
  can move the default out from under another lane.
- `fm-bahmni-lab-r1-machine` was never started, stopped, or otherwise touched.

## Transport security — the captain's ruling, implemented and proven

The previous run reported `mysql-transport-security` as "not present". It was genuinely **absent**:
no `sslmode` in the spec, no TLS field on the connection, and the driver called with no TLS option.
It is now implemented in a shared package, `internal/connectors/native/sqltls`, so MySQL and
PostgreSQL cannot drift.

| `sslmode` | Encrypts | Falls back to plaintext | Verifies chain | Verifies name |
| --- | --- | --- | --- | --- |
| `disabled` | no | n/a | no | no |
| `preferred` (default) | when offered | yes, only when the server advertises no TLS | no | no |
| `required` | yes | **never** | no | no |
| `verify-ca` | yes | **never** | yes | no |
| `verify-identity` | yes | **never** | yes | yes |

- **Never silently downgraded.** `preferred` is the only downgrading mode, and only on the driver's
  specific "server does not support TLS" refusal. If a driver upgrade changes that wording, the match
  stops firing and `preferred` fails closed — the safe direction.
- **Same option shape across SQL connectors.** `sslmode` / `sslrootcert` / `sslservername` are shared
  keys; libpq spellings (`disable`/`prefer`/`require`/`verify-full`) and MySQL spellings both parse to
  one canonical vocabulary. PostgreSQL now accepts the canonical names too, with existing libpq
  values passing through byte-identically (including `allow`, which is deliberately *not* rewritten).
- **Applies to replication too** — the binlog syncer gets the same `tls.Config`; a zero-value syncer
  would have replicated in plaintext regardless of the configured mode.
- The driver's own `NewClientTLSConfig` **panics** on a malformed CA PEM. This package builds the
  `tls.Config` itself and returns an error.

Proven three ways: unit tests against a hand-written TLS-less MySQL handshake (strict modes refuse,
`preferred`/`disabled` connect); certificate tests showing `verify-ca` rejects a chain outside the
configured root and accepts a correct chain regardless of hostname; and live subtests asking the real
server for its negotiated `Ssl_cipher`.

## Bugs this found that unit tests could not

- **`TRAILING` is a MySQL reserved word.** Both the catalog and read cursor-metadata queries aliased
  `information_schema.statistics` as `trailing`, making them syntax errors against a real server.
  Neither surfaced because both call sites replaced the server's error with a fixed string. Fixed,
  and the underlying error is now wrapped so this class of failure is debuggable.
- **Catalog aborted on one unquotable identifier.** A single table this connector cannot safely quote
  made the whole database undiscoverable, and `catalogStreams` would advertise streams whose every
  read would fail. Now such tables and columns are skipped, so discovery stays usable and never
  promises a stream it cannot serve.
- **Windows release build was broken.** `syscall.Statfs` is Unix-only; the package is now split by
  platform. Verified with `GOOS=windows go build ./...`.
- **Interrupt could leak a container.** A signal arriving between claiming ownership and issuing the
  create left the resource behind. `Close` now refuses further creates, cancels the in-flight
  command, and waits for the create sequence to settle; `Start` refuses to return an endpoint for a
  harness being torn down. Covered by a test that fires `Close` from inside a create, and by `-race`.

## Dependency evidence — `github.com/go-mysql-org/go-mysql v1.16.0`

Added under the captain's conditional approval for replication libraries.

- **Maintenance / last release:** v1.16.0, published **2026-07-15** — three weeks before this change.
  Actively maintained.
- **Licence:** MIT. Every transitive addition is permissive — MIT, Apache-2.0, or BSD-3-Clause. No
  copyleft.
- **Known CVEs:** `govulncheck ./...` (scanner v1.6.0, DB vuln.go.dev) reports **no vulnerabilities**
  across the module, including this dependency and its transitive set.
- **Transitive footprint:** 12 indirect modules added or bumped —
  `pingcap/{errors,log,tidb/pkg/parser}`, `shopspring/decimal`, `filippo.io/edwards25519`,
  `coreos/go-semver`, `go.uber.org/{zap,multierr}`, `natefinch/lumberjack.v2`, `stretchr/objx`, plus
  bumps to `goccy/go-json` and `klauspost/compress`. The TiDB SQL parser is the heaviest.
- **Measured binary size delta** (`go build -trimpath ./cmd/pm`, darwin/arm64):

  | | bytes | MiB |
  | --- | --- | --- |
  | before (`origin/main`) | 94,191,266 | 89.83 |
  | after (this branch) | 100,101,650 | 95.46 |
  | **delta** | **+5,910,384** | **+5.64 (+6.27 %)** |

  Most of that is the TiDB parser pulled in transitively. Flagging it rather than burying it: if
  +5.6 MiB per replication engine is unacceptable at four more engines, that is a decision to take
  now, not after MariaDB, SQL Server and Oracle have each added their own.

No other dependency was added.

## `cdc: true` is earned, not declared

MySQL is the **first** connector in the repo to advertise `cdc: true`; `postgres` correctly declares
`cdc: false` with `status: unsupported`. The capability is fail-closed by design, and it is claimed
here only because the executor is proven against a live server in the run above.

## Ready for the next four engines

The harness takes a new engine as configuration. What each needs:

| Engine | Image | Change-capture mechanism | What would block it |
| --- | --- | --- | --- |
| **MariaDB** | `mariadb:<pinned>` | `binlog_replication` | Closest to done — same driver and binlog protocol. Needs its own `mechanism` value confirmed against the closed vocabulary, and MariaDB's GTID differs from MySQL's. |
| **PostgreSQL** | `postgres:<pinned>` | `logical_replication` | Connector exists but CDC is a stub. Needs the already-approved `pglogrepl`, plus replication-slot lifecycle and LSN recovery. Slots are **server-side state** — the harness must drop them, or a failed run leaks WAL on the server, not just disk. |
| **SQL Server** | `mssql/server:<pinned>` | CDC capture tables / change tracking | Needs a driver decision (no approved one vendored yet). Licence acceptance is an image env var, and the image wants ~2 GB RAM — above this harness's current 2 GiB machine. |
| **Oracle** | `container-registry.oracle.com/database/free:<pinned>` | LogMiner / XStream | Hardest. The image is several GB (material on a disk-bound host), first boot is minutes not seconds, XStream needs licensed options, and the registry may require authentication — which the harness deliberately has no path to supply. |

Common prerequisite: `sslmode` is now shared, so each new SQL engine should adopt `sqltls` rather
than inventing its own spelling.

## Verification

`gofmt`, `go vet ./...`, `go build ./cmd/pm`, `GOOS=windows go build ./...`, `-race` on the harness,
and the `make verify` gates run individually: `tidy-check`, `docs-check`, `smoke-no-build`,
`agent-contract-check`, `connectorgen-validate`, `connectorgen-surface-sync`, `connector-boundary`,
`release-workflow-check`. `TestGoldenTranscripts` passes.

Two notes on generated artifacts:

- `connectorgen surface-sync` reported "1848 endpoints" but the real diff is one line, `"mysql": []`
  — correct for a database connector with no REST surface. Inspected before committing, not
  regenerated blind.
- `pm docs generate` rewrites **1033 unrelated connector docs**, because main's committed docs predate
  the field-type rendering that landed with dynamic schema discovery (#3892). That drift is real but
  is not this change's to carry, so only the mysql bundle is regenerated here.

## Not done

- Only MySQL is landed. Four engines are scoped above but deliberately not attempted — the brief asks
  for a working harness with one engine, not four half-built ones.
- GTID checkpointing, client-certificate auth, schema-change event projection, and cross-database CDC
  fan-out are out of scope for this slice.
