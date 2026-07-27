# TDD ledger - Binary v0.1.0 correction

## Red / current defects

- Branch contains a rejected `fm/*` branch-name exception.
- Branch contains a revoked `website-release` job and docs claiming PM binary releases dispatch/deploy the website.
- Branch includes website source/tests/workflow/deploy-doc changes and issue-guard repair traces introduced only for the revoked coupling/closed PR.
- Existing release docs describe a sticky config `release-as` option even though the corrected requirement is one-shot and non-persistent.

## Green expectations

- `ci/release-publishing` passes branch-name validation through the normal pattern.
- `fm/cli-release-and-connector-issues-r1` is rejected by branch-name validation.
- No `website-release` job or PM-binary-to-website dispatch claim remains.
- No website source/test/workflow/deploy doc changes remain in this branch diff.
- Release docs state the one-shot `Release-As: 0.1.0` footer and connector patch/minor model truthfully.
- GoReleaser snapshot emits the six expected archives plus `checksums.txt` and verifier passes.
