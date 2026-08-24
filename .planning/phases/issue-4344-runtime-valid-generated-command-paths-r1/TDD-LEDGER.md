# TDD ledger — #4344

| Slice | Red | Green | Refactor / verification | Status |
| --- | --- | --- | --- | --- |
| Generated parameterized command identity | Add behavioral projection/preflight tests for raw `{parameter}` source identity, collision pairs, and flags | Implement operation-derived injective identity and invalid-legacy migration | Re-run focused tests, then force the old function to show the regression test fails | planned |
| Checked-in surface reconciliation | `surface-sync --check` reports projection drift when an affected descriptor is available | Regenerate only derived artifacts | Validate / operation-evidence / command preflight checks | planned |
| Credential-free binary reachability | Built binary rejects the raw legacy path before credentials | Built binary reaches `missing --credential` for generated commands | Isolated Bitbucket 50-command sweep and legacy-name comparison | planned |
