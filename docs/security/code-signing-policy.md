# Code signing policy

PM's Windows code-signing work is being prepared for a future release. Current published Windows artifacts before that release may still show Windows as **Unknown Publisher**. A PM artifact is not considered a signed Windows release unless this policy, the release notes, and the release checks identify it as signed and verified.

## Planned provider and publisher display

PM's preferred open-source signing route is:

> Free code signing provided by [SignPath.io](https://about.signpath.io), certificate by [SignPath Foundation](https://signpath.org).

When that route is active, Windows Authenticode dialogs may display **SignPath Foundation** as the certificate publisher because the Foundation issues the certificate in its own name for eligible open-source projects. PM product metadata, installer metadata, repository ownership, and package manager metadata identify the software as **Polymetrics CLI** by **Polymetrics AI**.

Provider enrollment, identity evidence, provider projects, signing policies, certificates, secrets, and protected release environments are maintained outside Git. They must not be stored in issues, pull requests, logs, release assets, or agent-readable material.

## Repository roles

PM uses repository-based release roles rather than private identity lists in public documentation:

- **Authors:** contributors whose changes are merged through the `polymetrics-ai/cli` repository.
- **Reviewers:** maintainers or delegated reviewers who review pull requests and automated-review findings before release.
- **Approvers:** repository owners or release maintainers who can approve protected release/signing environments and provider signing requests.

All release-signing approvers must use multi-factor authentication for GitHub and provider access.

## Release signing rules

- Production signing is allowed only from protected release/tag workflows for source that has passed the repository's required CI and review gates.
- Pull requests, forks, local snapshots, and unsigned packaging validation workflows must not access signing credentials or call the signing provider.
- Signing keys must remain hardware-backed or cloud-HSM-backed with the approved provider. PM must not place a PFX/private key, certificate password, recovery secret, or identity evidence in GitHub.
- Windows executables are signed with SHA-256 Authenticode before they are packaged into the installer.
- Windows installers are built from those signed executable bytes, then the installers are signed with SHA-256 Authenticode.
- Release signing must include an RFC 3161 timestamp or the provider-managed timestamping equivalent, and release verification must confirm the timestamp.
- Checksums, provenance attestations, archive uploads, and package-manager hashes are generated only after signing and verification have completed.
- If signing, timestamping, or verification fails, the release fails closed. PM must not upload an unsigned Windows fallback under a signed-artifact name.

## Verification expectations

Future signed Windows releases must verify all of the following before upload:

- Authenticode publisher/certificate chain is trusted for the selected provider route.
- Digest algorithm is SHA-256.
- Timestamp is present and valid.
- Installer integrity and nested `pm.exe` integrity are valid after extraction or install.
- Clean Windows install, `pm version`, upgrade, and uninstall checks pass for each architecture that is published through package managers.

WinGet `InstallerSha256` values pin final installer bytes for download integrity. They do not replace Authenticode publisher trust or timestamp verification.

## User-directed network behavior and privacy

PM is a local-first CLI. It stores configuration and credentials locally unless the user or operator configures another location. PM does not need a hosted Polymetrics control plane for the first useful run.

PM can transfer information to networked systems when the user or operator asks it to do so, including but not limited to:

- adding or checking credentials for a configured connector;
- extracting data from a configured API, database, file/object store, queue, or other connector source;
- running reverse ETL after the required plan, preview, approval, and execute sequence;
- using optional runtime-backed services that the operator explicitly starts or configures;
- downloading release/package metadata through package managers such as WinGet.

Do not put secret values in GitHub issues, pull requests, support logs, release artifacts, or bug reports. Report vulnerabilities through the private process in [`SECURITY.md`](../../SECURITY.md).

## Incident response and revocation

If a signing credential, provider account, signed artifact, or release workflow is suspected to be compromised:

1. Disable the affected signing workflow, GitHub environment secret, or provider credential.
2. Pause provider signing policies while impact is investigated.
3. Open a private security incident using the process in [`SECURITY.md`](../../SECURITY.md).
4. Preserve audit logs, signing request IDs, workflow run IDs, artifact digests, and release metadata.
5. Coordinate with the signing provider on certificate revocation, including whether revocation should be immediate or retroactive and how timestamps affect already-published signatures.
6. Publish a corrected signed release and package-manager update when needed. Do not silently replace immutable release assets.

## Current implementation status

This policy is preparatory. Production Windows signing is not active until provider enrollment, protected release environments, signing policies, and Windows release verification are completed by authorized maintainers.
