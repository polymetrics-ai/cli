# Verification — GitHub mutation certification slice 4 writes-d

## Pending live run

- [x] Surface finding resolved by fleet ruling: connector commands use their authorized local `--approve` token path; stdin-token support is transport-only. See `TDD-LEDGER.md`.
- [ ] Each manifest command has exactly one terminal bucket: `certified`, `no_object`, `wrong_credential`, `entitlement`, `not_implemented`, `product_defect`, or `escape_needs_captain`.
- [ ] Every `certified` record contains an observable effect assertion and an independently proven cleanup absence assertion.
- [ ] No credential value is present in the worktree, evidence, command output, PR body, or status line.
- [ ] `go run ./cmd/connectorgen certification-matrix --check` is run after each eligible evidence record and at handoff.
- [ ] Repository verification and `scripts/verify-gsd-workflow` are run before the PR.
- [ ] The opened PR base is read from GitHub's API and equals `integration/4015-mvp-flat-r1`.
