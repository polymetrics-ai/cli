# Remote Reproducibility

This page distinguishes a clean clone's deterministic checks from live
connector proof. A fresh machine can verify the canon and structural runtime
preflight without provider credentials; it cannot honestly manufacture live
certification.

## What a clean environment needs

- A checked-out repository including the `data/` source reports and
  `docs/connector-canon/` archive/index added by this change.
- Go **1.25** or the version/toolchain selected by `go.mod`; `GOTOOLCHAIN=auto`
  is supported by the Makefile.
- CGO enabled and a working C toolchain. DuckDB is the embedded query engine and
  Parquet implementation, so `CGO_ENABLED=0` cannot build a usable warehouse
  binary.
- Git, standard POSIX shell utilities, and `shasum` for the source-pin check.
- `golangci-lint` for the lint gate. Node is needed for the repository's GSD
  adapter and website toolchain when those checks are run.

No provider credential, database service, Podman endpoint, or runtime service
is needed for the source-pin, definition, command-preflight, documentation, or
focused unit-test checks.

## Clean-clone baseline

```bash
git clone <repository-url> polymetrics-cli
cd polymetrics-cli
scripts/gsd doctor
go run ./cmd/agentcontractgen check
make connector-canon-check
make connector-runtime-preflight
go run ./cmd/connectorgen validate internal/connectors/defs
go run ./cmd/connectorgen surface-sync --check
go build ./cmd/pm
./pm docs validate --connectors-dir docs/connectors
```

For a connector change, run its focused tests and `internal/cli` separately
with `-timeout 20m`; CI remains the full-suite authority. See
[the implementation procedure](IMPLEMENTATION-PROCEDURE.md) for the full
change checklist.

## What currently prevents full remote live proof

1. **There are zero accepted live-certification artifacts.** A clean clone
   cannot prove a live provider interaction from fixtures or filenames.
2. **Provider environments and scoped credentials are external inputs.** A
   worker needs captain-approved test scope, a sandbox or authorized account,
   non-secret credential injection, bounded data, and a cleanup/receipt record.
   None belongs in this repository.
3. **Database and CDC live checks need infrastructure.** Native database tests
   are opt-in and require the [dbtest maintainer guide's](../../internal/connectors/native/dbtest/README.md)
   explicit Docker-or-Podman runtime and direct local Unix endpoint. Runtime-backed
   checks need their documented local services; they are not part of the default
   local path.
4. **Parity lanes are still independent evidence streams.** PostgreSQL (#3972,
   including warehouse-flow/mode gate #3987) and GitHub parity must be
   evaluated on their own reviewed branches. The current GitHub source lock is
   generated evidence; the archived wrong-branch gap map is not.
5. **Some archived reports preserve non-portable historical references.** Their
   content is retained for audit, but old absolute worktree paths are not a
   reproducible command interface.

## Result

A clean environment can reproduce the repository's current canon, derivation,
and “declared but unexecutable” structural guard. It cannot claim a connector
is live-certified until the external proof above exists and is accepted. That
limitation is intentional and must remain visible in connector documentation.
