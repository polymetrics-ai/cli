# Copyright (c) 2026 Polymetrics AI
# SPDX-License-Identifier: AGPL-3.0-only

[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$Version,

    [string]$OutputRoot = 'dist\windows'
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$OutputRootPath = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($OutputRoot)
$VersionInfoPattern = '^v?([0-9]+)\.([0-9]+)\.([0-9]+)(?:[-+].*)?$'
if ($Version -notmatch $VersionInfoPattern) {
    throw "Version '$Version' must look like vMAJOR.MINOR.PATCH"
}
$MsiVersion = "{0}.{1}.{2}" -f $Matches[1], $Matches[2], $Matches[3]

$Architectures = @(
    [pscustomobject]@{
        GoArch = 'amd64'
        WixArch = 'x64'
        UpgradeCode = '34C3F556-5634-5381-AE18-E1668FDECFA7'
    },
    [pscustomobject]@{
        GoArch = 'arm64'
        WixArch = 'arm64'
        UpgradeCode = 'EFEAFAA1-4276-509D-945A-D4F9BF7DBA30'
    }
)

New-Item -ItemType Directory -Force -Path $OutputRootPath | Out-Null

$wix = Get-Command wix -ErrorAction SilentlyContinue
if (-not $wix) {
    throw 'wix was not found on PATH. Install the WiX .NET tool before running this script.'
}

$oldGoos = $env:GOOS
$oldGoarch = $env:GOARCH
$oldCgoEnabled = $env:CGO_ENABLED
try {
    foreach ($arch in $Architectures) {
        $archDir = Join-Path $OutputRootPath $arch.GoArch
        New-Item -ItemType Directory -Force -Path $archDir | Out-Null

        $sysoPath = Join-Path $RepoRoot ("cmd\pm\pm_windows_{0}.syso" -f $arch.GoArch)
        Remove-Item -Force -ErrorAction SilentlyContinue (Join-Path $RepoRoot 'cmd\pm\pm_windows_*.syso')

        & (Join-Path $PSScriptRoot 'windows-versioninfo.ps1') -Version $Version -GoArch $arch.GoArch -Out $sysoPath
        if ($LASTEXITCODE -ne 0) {
            throw "windows-versioninfo.ps1 failed for $($arch.GoArch) with exit code $LASTEXITCODE"
        }

        $exePath = Join-Path $archDir 'pm.exe'
        Push-Location $RepoRoot
        try {
            $env:GOOS = 'windows'
            $env:GOARCH = $arch.GoArch
            $env:CGO_ENABLED = '0'
            & go build -trimpath -ldflags "-s -w -X polymetrics.ai/internal/cli.version=$MsiVersion" -o $exePath ./cmd/pm
            if ($LASTEXITCODE -ne 0) {
                throw "go build failed for windows/$($arch.GoArch) with exit code $LASTEXITCODE"
            }
        } finally {
            Pop-Location
            Remove-Item -Force -ErrorAction SilentlyContinue $sysoPath
        }

        $msiPath = Join-Path $archDir ("pm_{0}_windows_{1}.msi" -f $MsiVersion, $arch.GoArch)
        & wix build `
            -arch $arch.WixArch `
            -d "ProductVersion=$MsiVersion" `
            -d "SourceExe=$exePath" `
            -d "UpgradeCode=$($arch.UpgradeCode)" `
            -o $msiPath `
            (Join-Path $RepoRoot 'packaging\windows\pm.wxs')
        if ($LASTEXITCODE -ne 0) {
            throw "wix build failed for $($arch.WixArch) with exit code $LASTEXITCODE"
        }

        Write-Host "built unsigned Windows package snapshot: $msiPath"
    }
} finally {
    $env:GOOS = $oldGoos
    $env:GOARCH = $oldGoarch
    $env:CGO_ENABLED = $oldCgoEnabled
    Remove-Item -Force -ErrorAction SilentlyContinue (Join-Path $RepoRoot 'cmd\pm\pm_windows_*.syso')
}
