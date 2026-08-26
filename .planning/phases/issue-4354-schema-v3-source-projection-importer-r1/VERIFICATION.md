# Schema-v3 source projection/importer foundation — verification checklist

## Required local evidence

- [ ] Focused Red/Green importer and projection Go tests with exact commands.
- [ ] Scoped `connectorgen source-import <proof> --check` passes without
  provider access and reports truthfully source-backed operations.
- [ ] `connectorgen validate`, `surface-sync --check`, and
  `operation-evidence --check` pass after only necessary generated evidence is
  refreshed.
- [ ] Six-lane counts and all `missing_foundation` rows are inspected from the
  generated evidence.
- [ ] Any new command is proven with a built `pm` binary in a fresh
  credential-free project to stop at `missing --credential`; otherwise report
  usable-surface delta `0`.
- [ ] `gofmt`, scoped tests, `go vet`, `go build`, `git diff --check`, and
  applicable docs/lint/generator checks are recorded.
- [ ] Full-suite and aggregate gates not safely runnable locally are listed as
  CI authority, not claimed as passing.
- [ ] Independent final-SHA audit records source → descriptor → lane mapping,
  generic-escape review, usable-surface delta, and the exact checked commit.
- [ ] PR base is read back from the GitHub API and equals `main`.
