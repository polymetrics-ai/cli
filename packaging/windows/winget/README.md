# WinGet manifest templates for PM

These templates document the approved Windows Package Manager identity for the first future signed PM MSI release:

```text
PolymetricsAI.PolymetricsCLI
```

They are **not** submission manifests. Do not submit them to `microsoft/winget-pkgs` until a public immutable release contains final Authenticode-signed MSI installers.

Before the first external PR:

1. Build the final Windows executable bytes.
2. Sign and RFC 3161 timestamp the executables through the approved HSM-backed provider.
3. Build the x64 and arm64 WiX MSIs from those signed executable bytes.
4. Sign and timestamp the MSI files.
5. Verify signatures, publisher/chain/timestamp, install, upgrade, and uninstall on Windows.
6. Compute `InstallerSha256` from the final signed MSI bytes only after signing is complete.
7. Replace placeholders in these templates under the `winget-pkgs` path:
   `manifests/p/PolymetricsAI/PolymetricsCLI/<VERSION>/`.

WinGet hashes pin installer bytes for repository validation and download integrity. They do not replace Authenticode signing or timestamp verification.

Arm64 should be included in the first WinGet PR only after native Windows ARM64 install/run/uninstall validation passes.
