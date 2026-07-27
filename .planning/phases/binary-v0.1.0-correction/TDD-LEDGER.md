# TDD ledger - Binary v0.1.0 correction

## Red / current defects

- Superseded branch `ci/release-publishing` has reachable history that introduced a revoked `website-release` job and docs claiming PM binary releases publish the website.
- Superseded branch history included website source/tests/workflow docs and issue-guard repair traces introduced only for the revoked coupling/closed PR.
- Superseded verification evidence pointed at a stale `Release-As: 0.1.0` commit SHA that is not auditable for the replacement PR.
- Existing release docs described a sticky config `release-as` option even though the corrected requirement is one-shot and non-persistent.

## Green expectations

- `ci/pm-v0.1.0-release` starts from `origin/main` and contains no website-coupling commit ancestry.
- `fm/cli-release-and-connector-issues-r1` is rejected by branch-name validation.
- No `website-release` job or PM-binary-to-website publication claim remains.
- No website source/test/workflow doc changes remain in this branch diff.
- Verification evidence names the clean candidate commit that carries `Release-As: 0.1.0`.
- Release docs state the one-shot `Release-As: 0.1.0` footer and connector patch/minor model truthfully.
- GoReleaser snapshot emits the six expected archives plus `checksums.txt` and verifier passes.
