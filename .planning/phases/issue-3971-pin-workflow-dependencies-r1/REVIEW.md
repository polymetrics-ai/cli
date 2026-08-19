# Manual code review: pin build dependencies

## Scope reviewed

- Eleven workflow files containing external actions.
- `website/Dockerfile` and the literal PostgreSQL service image.
- The new `scripts/tests/pinned-build-dependencies.sh` regression gate and its `Makefile` integration.
- GSD/TDD evidence for issue #3986 under parent #3971.

## Findings

No actionable findings.

1. Each `uses:` change replaces only the original mutable ref with the exact current commit resolution, retaining the original version in a trailing comment. A post-edit `git ls-remote` read confirmed all fourteen captured refs still resolve to the recorded commits.
2. Node and PostgreSQL retain their readable tags and add their current multi-architecture manifest-list digests. `docker buildx imagetools inspect` and `docker build --check` accepted both pins.
3. The regression gate covers every workflow `uses:` reference, literal workflow image, and Dockerfile `FROM` image; it rejects missing 40-character SHAs, version comments, or digests, then parses every workflow YAML file. The baseline failed before edits and the final repository passes.
4. The diff changes no triggers, permissions, step inputs, commands, action/image versions, or application source. The `Makefile` integration makes the immutable-reference gate part of the existing `release-workflow-check` path used by `make verify`.

## Disposition

Approved for commit. Alert #135 is still open against the pre-change default-branch analysis and must be rechecked only after merge and a new Scorecard analysis.
