# VERIFICATION — issue #3853 engine content preservation

Status: planned; not yet executed.

## Checklist

- [ ] Preview warnings retain top-level, nested, and config-secret interpolation content.
- [ ] Existing `redact_fields` declarations remain load-compatible without rewriting bundles.
- [ ] Direct-read, operation-direct-read, and binary-download failure messages retain bounded
  provider diagnostic content.
- [ ] Error-map class/hint, request bounds, redirect protections, preview digest, typed
  confirmation, approval evidence, and no-retry behavior remain unchanged.
- [ ] No #3771 command-runner, #3852 enum, successful-output policy, binary-record, generic
  source-table, provider, credential, or reverse-execution scope leakage occurs.
- [ ] Runtime help, manual, golden transcript, and website documentation agree about the
  connector-engine complete-content boundary and approval-token handling.
- [ ] Targeted tests, package tests, static/build checks, individual repository gates, and inline
  GSD verification/review are recorded with exact command output.

## Intentionally excluded

- Whole-repository `go test ./...` and `make verify` monolith: CI-owned under the repository's
  connector-suite timeout guidance.
- Live provider calls, credentialed checks, and reverse-ETL execution: prohibited by issue scope.
