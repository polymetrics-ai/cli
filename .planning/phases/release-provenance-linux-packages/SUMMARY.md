# Summary — release/provenance-linux-packages

Status: implemented and locally verified.

## Delivered

- Added GoReleaser v2/nFPM `.deb` and `.rpm` packages for Linux `amd64`/`arm64` with repository-grounded metadata, `/usr/bin/pm`, and `/usr/share/doc/pm/{LICENSE,NOTICE}`.
- Extended `scripts/verify-release-assets.sh` to enforce 10 release assets, checksum coverage, archive contents, package metadata/contents/architectures, optional Cosign bundle verification, and optional GitHub attestation verification.
- Added offline trust-evidence fixtures for PR snapshot validation without production signing services.
- Added Docker-based standalone package install/reinstall/remove checks for Ubuntu and Fedora-family environments on the available Docker architecture.
- Added release-only GitHub OIDC provenance and Cosign keyless bundle generation before upload, with final verification gating upload and no overwrite/clobber path for existing PM release assets.
- Added narrow release verification documentation and linked it from `docs/GUIDE.md`.
- Repaired PR #560 CI by upgrading the existing indirect gRPC module to the GO-2026-6061 fixed floor and teaching `prissueguard` to recognize explicit issue-first delivery/parent references already present in the PR body.
- Repaired the external `security/snyk` check by updating independent website dependency metadata and lockfiles for vulnerable `next`, `js-yaml`, `postcss`, and `sharp` resolutions without changing release workflow behavior or website application source.

## Deferred explicitly to separate issues

- Apple Developer ID signing/notarization.
- Windows Authenticode and WinGet.
- Signed APT/RPM repositories and repository-signing key custody (issue #553).
- Broad public release runbooks (issue #558).

## Verification

See `VERIFICATION.md` for command results. Full `make verify` passed locally.
