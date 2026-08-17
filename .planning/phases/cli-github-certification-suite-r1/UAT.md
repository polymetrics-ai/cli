# UAT — GitHub certification suite r1

Autonomous `--auto` verification used directly observed, reproducible
outcomes; no human judgment was required.

| Deliverable | Evidence | Result |
| --- | --- | --- |
| D1: surface-derived exhaustive accounting | `TestCertificationSweepForGitHubIsSurfaceDerivedAndExhaustive`; generated artifact check | Pass — 1,571 unique declared command paths, each with one non-pass status and reason. |
| D2: provider/product distinction | `TestCertificationSweepSeparatesProductDefectsAndProviderRefusals`; `FINDINGS.md` | Pass — `releases assets view` is a product defect and the measured HTTP 422 is a provider refusal. |
| D3: produced-value certification proof | post-schema scratch mismatch followed by restored compiled `pm connectors certify github --direct-read-only` | Pass — the named stage turned red on its `/response/name` assertion, then all 23 declaration-owned stages passed after restoration. |
| D4: artifact safety | strict generated-artifact validation, `go vet`, lint, connector-boundary, and Makefile gate | Pass — no unknown artifact fields, connector-specific shared-Go code, or unresolved review findings. |
