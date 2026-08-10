# #3864 verification checklist

## Status: planned

- [ ] TDD RED outputs are recorded before production code.
- [ ] Focused `internal/connectors`, `internal/synctransport`, `internal/app`, and
  `internal/cli` tests pass with `-timeout 20m`.
- [ ] Transport package race test and cancellation regression pass.
- [ ] `go vet` and build pass.
- [ ] Required non-suite `make verify` components pass individually.
- [ ] `make connector-runtime-preflight`, connector canon, connectorgen validation, and
  surface-sync checks pass.
- [ ] `pm connectors`, `pm help connectors`, `pm connectors --help`, and
  `pm connectors inspect sample --json` are checked in an initialized project without
  credentials; docs/website parity is checked.
- [ ] Manual `verify-work` outcome, gap count, code-review findings/dispositions,
  no-mistakes result, supervisor-compatible evidence, and child-local check state are
  recorded.

## Explicit limits

This verification can prove only fake-backed dispatch and metadata surfaces. It cannot
truthfully assert executable #3810 conformance, a real API/database transport, a live
provider flow, automatic Shepherd certification, or a green GitHub CI/review state until
those gates actually run.
