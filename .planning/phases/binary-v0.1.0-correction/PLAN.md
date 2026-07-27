# Binary v0.1.0 correction

Branch: `ci/pm-v0.1.0-release`
Replacement PR title: `ci(release): prepare pm v0.1.0 binary release`
Supersedes: closed unmerged PR #528
Issue link: `Refs #67`

## Required skills loaded

- `gsd-core`
- `gsd-programming-loop`
- `no-mistakes`
- `golang-how-to`
- `golang-continuous-integration`
- `golang-documentation`
- `golang-error-handling`
- `golang-lint`
- `golang-safety`
- `golang-security`
- `golang-testing`

## GSD command path

- `scripts/gsd doctor` passed earlier in this recovery.
- `scripts/gsd prompt programming-loop init --phase binary-v0.1.0-correction --dry-run` returned `unknown GSD command: programming-loop`; manual-GSD fallback remains active and is recorded here.

## Corrected scope

The website and PM binary are independent products. This branch must prepare the PM binary release path only:

1. Create clean compliant replacement branch `ci/pm-v0.1.0-release` from `origin/main`; never rewrite, force-push, delete, or mutate `ci/release-publishing`.
2. Remove release-to-website coupling and all claims that PM binary releases publish the website.
3. Exclude website source/tests/workflows/docs and unrelated issue-guard repairs.
4. Use Release Please's one-shot `Release-As: 0.1.0` commit footer, not persistent config, to make the next PM release `v0.1.0`.
5. Preserve GoReleaser binary assets: macOS/Linux/Windows amd64+arm64 plus `checksums.txt`.
6. Document embedded connector release truth: connector definitions ship inside the PM binary; fixes ship in PM patch releases, features normally in pre-1.0 minor releases; do not claim unmerged WhatsApp is included.
7. State `v0.1.0` must wait for mandatory Bahmni corrective PR and green release commit; captain owns merge/release decisions.

## Plan

- Replay only the final binary-release docs/artifacts onto `origin/main`.
- Confirm no `fm/*` branch-name exception is present in conventions or docs.
- Confirm no `website-release` job is present in `.github/workflows/release.yml`.
- Rewrite `docs/release-and-connectors.md` around binary-only release preparation and connector issue structure.
- Verify YAML parsing, branch-name behavior, GoReleaser snapshot assets, release asset verifier, and Release Please one-shot override evidence.
- Commit with `Release-As: 0.1.0` in the commit body.
- Open replacement PR through `gh-axi` with the exact requested title and `Refs #67`.
- Run no-mistakes validation without `--yes`; do not merge.
