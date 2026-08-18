# Release verification

This page documents the narrow release-trust evidence added for future PM
GitHub releases. It does not describe Apple Developer ID/notarization, Windows
Authenticode, WinGet, or signed APT/RPM repositories; those remain separate
workstreams under polymetrics-ai/cli#550.

Published `v0.1.0` assets were produced before this evidence existed. Do not
expect these commands to succeed for that release.

## Homebrew formula notification rollout

After a future PM release's tag, assets, checksum manifest, Cosign bundles, GitHub
SLSA provenance, and package checks pass in `.github/workflows/release.yml`, the
release workflow requests the tap-owned `polymetrics-ai/homebrew-tap` workflow
`.github/workflows/pm-formula-update.yml` with `dry_run=true`. The notification
uses only the `pm-homebrew-pr-bot` GitHub App credentials named
`PM_HOMEBREW_PR_APP_ID` and `PM_HOMEBREW_PR_PRIVATE_KEY`; missing credentials
fail the notification job rather than falling back to a PAT, `GITHUB_TOKEN`, or
ambient `gh` authentication.

The current CLI-side rollout is dry-run only. It passes schema
`pm-homebrew-formula/v1`, source repo `polymetrics-ai/cli`, the verified `vX.Y.Z`
tag, release id, the source release workflow run id for assets signed by the
current workflow run, and
`target_commitish_policy=ignore`. Enabling `dry_run=false` would allow the tap's
App-owned formula branch/PR path and requires a later explicit code and operator
documentation change; do not activate it as part of release validation.

If the release already has the exact verified asset set from an earlier trusted
run, the notification omits `source_run_id` rather than attributing old
provenance to the current run. Website deployment is independent and does not
trigger this Homebrew path.

The PM release job also filters release-worthy commits before Release Please
generates the PM release PR: commits that touch only `website/**`,
`.github/workflows/website.yml`, or `.gitlab-ci.yml` are omitted from the
isolated Release Please history, while PM CLI and release workflow changes still
participate.

## Future release asset set

For a future tag such as `v0.2.0`, the GitHub release is expected to contain:

- generic archives:
  - `pm_0.2.0_darwin_amd64.tar.gz`
  - `pm_0.2.0_darwin_arm64.tar.gz`
  - `pm_0.2.0_linux_amd64.tar.gz`
  - `pm_0.2.0_linux_arm64.tar.gz`
- native standalone Linux packages:
  - `pm_0.2.0_linux_amd64.deb`
  - `pm_0.2.0_linux_arm64.deb`
  - `pm_0.2.0_linux_x86_64.rpm`
  - `pm_0.2.0_linux_aarch64.rpm`
- `checksums.txt`
- one Cosign bundle beside every archive, native package, and `checksums.txt`,
  named `<asset>.sigstore.json`.

The checksum file covers the archives and packages. The Cosign bundles are not
included in `checksums.txt`; they are verification evidence for the final bytes.

## Verify one downloaded asset

Set the tag and asset name, then download the asset, its bundle, and the
checksum manifest:

```bash
tag=v0.2.0
asset=pm_0.2.0_linux_amd64.tar.gz

mkdir -p pm-release-verify
cd pm-release-verify
gh release download "$tag" \
  --repo polymetrics-ai/cli \
  --pattern "$asset" \
  --pattern "$asset.sigstore.json" \
  --pattern checksums.txt
```

First verify integrity against the release checksum manifest:

```bash
grep "  $asset$" checksums.txt | sha256sum --check
```

A checksum match only proves the local bytes match the release manifest. It does
not prove who produced the asset. Verify the keyless Cosign bundle with the
repository, workflow, and exact GitHub ref constrained. Normal Release
Please-created releases mint evidence from the protected `main` workflow run
that checks out the release tag, so use `refs/heads/main`. If a release was
minted from a `release: published` tag workflow run, use `refs/tags/$tag`.

```bash
cert_ref=refs/heads/main
certificate_identity="https://github.com/polymetrics-ai/cli/.github/workflows/release.yml@$cert_ref"

cosign verify-blob "$asset" \
  --bundle "$asset.sigstore.json" \
  --certificate-identity "$certificate_identity" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com"
```

Then verify the GitHub artifact attestation. The GitHub CLI computes the digest
of the local file and fails if no attested subject with that digest and identity
exists:

```bash
gh attestation verify "$asset" \
  --repo polymetrics-ai/cli \
  --cert-identity "$certificate_identity" \
  --cert-oidc-issuer "https://token.actions.githubusercontent.com" \
  --predicate-type https://slsa.dev/provenance/v1 \
  --deny-self-hosted-runners \
  --format json
```

Repeat the same checksum, Cosign, and GitHub attestation commands for
`checksums.txt` itself by setting `asset=checksums.txt` and downloading
`checksums.txt.sigstore.json`.

## Standalone Linux packages

The `.deb` and `.rpm` files are standalone packages. They install `pm` to
`/usr/bin/pm` and include `LICENSE` and `NOTICE` under `/usr/share/doc/pm/`.
Their package metadata uses:

- package name: `pm`
- maintainer/vendor: `Polymetrics AI`
- homepage: `https://cli.polymetrics.ai`
- license: `AGPL-3.0-only`
- conflicts: none declared; this repository has no earlier native package name to replace
- Debian architectures: `amd64`, `arm64`
- RPM architectures: `x86_64`, `aarch64`
- upgrade/uninstall behavior: package managers replace the same `pm` package name on upgrade and remove `/usr/bin/pm` plus package-owned doc files on uninstall; there are no maintainer scripts or background services

Release CI exports `SOURCE_DATE_EPOCH` from the checked-out release commit so
repeated archive and nFPM assembly of the same commit produce stable package
bytes.

These packages are not signed APT or RPM repositories. A signed package
repository would require repository metadata such as APT `InRelease` or RPM
repository signatures; that belongs to issue #553 and is not claimed here.
