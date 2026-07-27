# Copyright (c) 2026 Polymetrics AI
# SPDX-License-Identifier: AGPL-3.0-only

[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$Version,

    [Parameter(Mandatory = $true)]
    [ValidateSet('amd64', 'arm64')]
    [string]$GoArch,

    [Parameter(Mandatory = $true)]
    [string]$Out
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$OutputPath = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($Out)
$OutputDir = Split-Path -Parent $OutputPath
New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

$machineByArch = @{
    amd64 = 'X64'
    arm64 = 'ARM64'
}
$machine = $machineByArch[$GoArch]

Push-Location $RepoRoot
try {
    & go run ./build/windowsversion -version $Version -goarch $GoArch -out $OutputPath
    if ($LASTEXITCODE -ne 0) {
        throw "windowsversion generator failed with exit code $LASTEXITCODE"
    }
} finally {
    Pop-Location
}

Write-Host "generated VERSIONINFO resource: $OutputPath ($GoArch/$machine)"
