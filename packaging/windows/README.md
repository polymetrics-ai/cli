# Windows packaging foundation

This directory contains provider-inert Windows packaging sources for PM.

Approved future WinGet identifier: `PolymetricsAI.PolymetricsCLI`.

## What is safe in pull requests

Pull requests may:

- generate a temporary Windows VERSIONINFO resource from source;
- build unsigned `pm.exe` snapshots for `windows/amd64` and `windows/arm64`;
- build unsigned WiX MSI snapshots for x64 and arm64;
- verify metadata, MSI structure, and x64 install/run/uninstall behavior.

Pull requests must not:

- call SignPath or another signing provider;
- read signing credentials or identity evidence;
- create a self-signed or fake trusted signature;
- publish unsigned snapshots as release artifacts.

## Installer defaults

- Product: `Polymetrics CLI`
- Manufacturer: `Polymetrics AI`
- Install scope: machine
- Install path: `%ProgramFiles%\Polymetrics\CLI\pm.exe`
- PATH behavior: add the install directory to the machine `PATH`; remove it on uninstall
- Upgrade behavior: major upgrades are allowed; downgrades are blocked

Stable per-architecture MSI `UpgradeCode` values:

| Architecture | WiX `-arch` | UpgradeCode |
|---|---|---|
| amd64/x64 | `x64` | `{34C3F556-5634-5381-AE18-E1668FDECFA7}` |
| arm64 | `arm64` | `{EFEAFAA1-4276-509D-945A-D4F9BF7DBA30}` |

The final release workflow will build MSIs from already signed `pm.exe` bytes, sign/timestamp the MSI files, verify both layers, and only then compute checksums and WinGet hashes.
