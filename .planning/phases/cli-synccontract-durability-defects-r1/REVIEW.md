# Code review — sync-contract durability defects

`scripts/gsd prompt code-review cli-synccontract-durability-defects-r1` was generated and
reviewed inline under the single-worker fallback.

## Scope review

- The state-load audit has exactly two `a.store.Load()` call sites. Both now route through the same
  normalizer; no raw `a.state` replacement remains.
- The directory creation change is confined to the two existing `MkdirAll` call sites and uses the
  existing `durability.SyncDirectory` primitive. No dependency or generic filesystem API was added.
- The acknowledgement remains after all directory syncs. Any failed parent sync returns an error
  before a pending stream checkpoint is constructed.
- Tests do not expose credentials, connection strings, or warehouse records.

## Findings

No blocking correctness, durability, security, or scope findings.

The scoped app linter reports two unrelated, pre-existing test-only `errcheck` findings in
`query_engine_helpers_test.go` and `reverse_approval_test.go`; neither is in this diff and both are
outside the two-defect scope.
