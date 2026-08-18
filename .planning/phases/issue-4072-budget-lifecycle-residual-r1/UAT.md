# #4072 residual automated UAT

Named-issue manual fallback: this behavior has no visual or judgment-only
step. The `coverage` entries in `SUMMARY.md` are automated and pass.

| Deliverable | Source | Result |
| --- | --- | --- |
| Granted token lifecycle | `TestGitHubAppAuthBudgetLifecycleGrantFinishesExactlyOnce` | passed |
| Refused token lifecycle | `TestGitHubAppAuthBudgetLifecycleRefusalDoesNotFinishOrSend` | passed |
| Earlier admission/no-retry contract | `TestGitHubAppAuth*` regression suite | passed |
