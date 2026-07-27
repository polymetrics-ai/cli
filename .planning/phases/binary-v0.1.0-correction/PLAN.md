# Binary v0.1.0 correction

Branch: `ci/release-publishing`
Replacement PR title: `ci(release): prepare pm v0.1.0 binary release`
Supersedes: closed unmerged PR #528
Issue link: `Refs #67`

## Required skills loaded

- `gsd-core`
- `no-mistakes`
- `golang-how-to`
- `golang-continuous-integration`
- `golang-documentation`

## GSD command path

- `scripts/gsd doctor` passed earlier in this recovery.
- `scripts/gsd prompt programming-loop init --phase pr-528-branch-convention-correction --dry-run` returned `unknown GSD command: programming-loop`; manual-GSD fallback remains active and is recorded here.

## Corrected scope

The website and PM binary are independent products. This branch must prepare the PM binary release path only:

1. Keep compliant branch `ci/release-publishing`; never recreate or exempt `fm/*`.
2. Remove release-to-website coupling and all claims that PM binary releases dispatch/deploy the website.
3. Exclude website source/tests/workflows/deploy docs and unrelated issue-guard repairs.
4. Use Release Please's one-shot `Release-As: 0.1.0` commit footer, not persistent config, to make the next PM release `v0.1.0`.
5. Preserve GoReleaser binary assets: macOS/Linux/Windows amd64+arm64 plus `checksums.txt`.
6. Document embedded connector release truth: connector definitions ship inside the PM binary; fixes ship in PM patch releases, features normally in pre-1.0 minor releases; do not claim unmerged WhatsApp is included.
7. State `v0.1.0` must wait for mandatory Bahmni corrective PR and green release commit; captain owns merge/release decisions.

## Plan

- Revert out-of-scope website and issue-guard files to `origin/main`.
- Remove `fm/*` branch-name exception from conventions and docs.
- Remove `website-release` job from `.github/workflows/release.yml`.
- Rewrite `docs/release-and-connectors.md` around binary-only release preparation and connector issue structure.
- Verify YAML parsing, branch-name behavior, GoReleaser snapshot assets, release asset verifier, and Release Please one-shot override evidence.
- Commit with `Release-As: 0.1.0` in the commit body.
- Open replacement PR through `gh-axi` with the exact requested title and `Refs #67`.
- Run no-mistakes validation without `--yes`; do not merge.
