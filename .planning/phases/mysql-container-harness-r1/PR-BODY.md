## Summary

Adds the reusable, sequential MySQL container-test harness and one Tier-3 native MySQL source
connector. The MySQL bundle declares `integration_type: "database"` and a fail-closed
`binlog_replication` changefeed that is advertised only by the matching native executor. The
production registry now installs that native MySQL connector over the declarative bundle.

The harness was changed from Podman to Docker on Colima. Three independent Podman machines on
this host, including the task-owned machine, failed to start. Docker through Colima was instead
independently proven with MySQL 8.4, and the live proof below uses only an explicit Docker context.
Docker resource cleanup cannot shrink Colima's host VM disk, so the documented command opts into a
post-cleanup `colima delete` + `colima start` reset. This is deliberately destructive to the
disposable default profile and is never automatic.

No standalone GitHub issue was supplied for this firstmate task; attach the parent issue before
opening a PR.

## Live proof and cleanup

```bash
DOCKER_CONTEXT=colima POLYMETRICS_DATABASE_INTEGRATION=1 \
  POLYMETRICS_DATABASE_RESET_COLIMA=1 \
  go test -tags=databaseintegration -count=1 -v ./internal/connectors/native/mysql
```

The command passed against MySQL 8.4.11 in 53.48 seconds. It started the exact
`docker.io/library/mysql:8.4.11` tag on a Docker-assigned non-default loopback port, seeded five
deterministic rows, and asserted the connector's real check, dynamic discovery, paged full read,
incremental read, and insert/update/delete binlog CDC records. Deferred teardown removed the
generated container, volume, and run-pulled image, then reset Colima. The test reported
before=84802707456 and after=84784943104 free bytes with `colima_reset=true` (a 17.8 MB decrease,
within the harness's 128 MB ordinary-build noise allowance); a post-run explicit
Docker-context listing showed zero containers, volumes, and images.

The live test uncovered and fixed Docker argument placement, MySQL protocol metadata casing, and
MySQL 8.4's `SHOW BINARY LOG STATUS` replacement for the removed `SHOW MASTER STATUS` syntax.
The saved live run exercised the same native implementation directly; native registry installation
is now asserted separately by the focused bundle-registry test. Follow-up review makes cleanup of a
run-pulled image unconditional, preserves explicit empty and whitespace cursor state, normalizes
text result values, and rejects unsafe statement binlog events. A fresh opt-in Docker replay after
those final lifecycle, read, and CDC fixes belongs to the outer test phase.

## Dependency approval evidence

Only one new direct module was added: `github.com/go-mysql-org/go-mysql v1.16.0`. It provides both
the MySQL client and row-binlog replication protocol, so no second client dependency was added.

- Maintenance / release: v1.16.0 is a verified upstream tag released 2026-07-15; this was the
  current assessed release. [Upstream release](https://github.com/go-mysql-org/go-mysql/releases/tag/v1.16.0)
- Licence: MIT; the module also retains a BSD-3-Clause notice for imported Vitess code.
- Footprint: Go's pruned root graph adds nine indirect modules
  (`filippo.io/edwards25519`, `coreos/go-semver`, PingCAP errors/log/parser, decimal, multierr,
  zap, and lumberjack) and upgrades existing `goccy/go-json`, `stretchr/objx`, and
  `klauspost/compress`. The upstream module declares 22 non-tool transitive requirements; the
  native package's resolved build closure contains 17 external modules, several already present in
  this repository.
- CVEs/advisories: the first scan found Go advisory `GO-2026-5841` / `GHSA-259r-337f-4rfw` in the
  inherited `klauspost/compress v1.18.6` (its fixed version is v1.18.7). The root module pins
  v1.18.7 before this change lands. Final `go mod verify` and
  `go run golang.org/x/vuln/cmd/govulncheck@latest -show verbose
  ./internal/connectors/native/mysql` reported no vulnerabilities. [Go advisory](https://pkg.go.dev/vuln/GO-2026-5841)
- Binary size: clean `pm` build before dependency work: **93,727,298 bytes**; the earlier
  **93,727,346-byte** post-dependency build predates native MySQL registration and is not the final
  production binary measurement. The native factory is now linked by the production registry; the
  outer build phase must refresh the final binary-size evidence.

## Follow-on engine configurations

- MariaDB: ready for the same harness configuration and binlog mechanism; block on a live test of
  version-specific position/row-event behavior and account privileges.
- SQL Server: ready for a Docker image/data-volume configuration; block on Microsoft image licence
  acceptance, SQL Server CDC/log configuration, and the approved driver's security evidence.
- Oracle: ready for image/data-volume configuration; block on XE image licensing and ARM64
  availability, then the approved replication library's security evidence.
- PostgreSQL: ready for the same lifecycle harness; block on the existing connector's human-gated
  `pglogrepl` dependency, logical-slot lifecycle, and live executor proof.

## Runtime ownership

Created `fm-cli-db-harness-r1` while attempting the original Podman route. It initialized but never
started; after the runtime change instruction it was left stopped and untouched rather than spending
more time on broken Podman. It was **not removed** yet. No command touched
`fm-bahmni-lab-r1-machine`. The task's Docker/Colima test leaves the Colima default profile running
but empty after its reset.

## Validation

- GSD inline/manual fallback recorded in `.planning/phases/mysql-container-harness-r1/`; required
  Go skills: `golang-how-to`, design patterns, structs/interfaces, error handling, security,
  safety, testing, context, concurrency, database, documentation, dependency management, and
  pkg-go-dev.
- `go mod verify`
- `govulncheck -show verbose ./internal/connectors/native/mysql`
- `gofmt`, scoped `go vet`, focused harness/MySQL/bundle/command-runner/engine tests, relevant
  connector-catalog CLI tests, a full `go test ./internal/connectors/...` regression, and
  `go build ./cmd/pm`
- `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`,
  `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, and
  `make release-workflow-check`
- `make tidy-check` passed after the commit. Before the commit its intended `go.mod`/`go.sum`
  changes correctly made the clean-tree comparison fail.
- Inline GSD `verify-work` and code review evidence in `UAT.md` and `REVIEW.md`; compatible
  isolated GSD workers are unavailable/forbidden by the single-worker contract.
- `pnpm --dir website run test:scripts`

## CLI/docs parity

No new CLI verb or flag was needed, but the registered bundle count is now 551 and the MySQL bundle
is overridden by its registered native connector. Regenerated
connector catalog, CLI manual/help, golden transcript, connector manual/skill, icon registry, and
website catalog/icon artifacts are included. `pm help connectors`, `pm connectors`, and
`pm connectors inspect mysql --json` were exercised. The connector's bundled documentation
contains the developer-only integration command, cleanup semantics, runtime rationale, and reset
warning.
