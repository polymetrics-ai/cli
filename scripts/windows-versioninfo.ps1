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

function Find-WindowsSdkTool {
    param([Parameter(Mandatory = $true)][string]$ToolName)

    $kitsRoot = Join-Path ${env:ProgramFiles(x86)} 'Windows Kits\10\bin'
    if (Test-Path $kitsRoot) {
        $matches = @(Get-ChildItem -Path $kitsRoot -Filter $ToolName -Recurse -ErrorAction SilentlyContinue |
            Where-Object { $_.FullName -match '\\x64\\' } |
            Sort-Object FullName -Descending)
        if ($matches.Count -gt 0) {
            return $matches[0].FullName
        }
    }

    $fromPath = Get-Command $ToolName -ErrorAction SilentlyContinue
    if ($fromPath) {
        return $fromPath.Source
    }

    throw "Could not find $ToolName. Install the Windows SDK on the Windows runner."
}

function Find-Cvtres {
    $vswhere = Join-Path ${env:ProgramFiles(x86)} 'Microsoft Visual Studio\Installer\vswhere.exe'
    if (Test-Path $vswhere) {
        $installPath = & $vswhere -latest -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64 -property installationPath
        if ($LASTEXITCODE -eq 0 -and $installPath) {
            $matches = @(Get-ChildItem -Path (Join-Path $installPath 'VC\Tools\MSVC') -Filter cvtres.exe -Recurse -ErrorAction SilentlyContinue |
                Where-Object { $_.FullName -match '\\Hostx64\\x64\\' } |
                Sort-Object FullName -Descending)
            if ($matches.Count -gt 0) {
                return $matches[0].FullName
            }
        }
    }

    $fromPath = Get-Command cvtres.exe -ErrorAction SilentlyContinue
    if ($fromPath) {
        return $fromPath.Source
    }

    throw 'Could not find cvtres.exe. Install Visual Studio Build Tools with VC tools on the Windows runner.'
}

$machineByArch = @{
    amd64 = 'X64'
    arm64 = 'ARM64'
}
$machine = $machineByArch[$GoArch]

$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("pm-versioninfo-{0}" -f ([Guid]::NewGuid().ToString('N')))
New-Item -ItemType Directory -Force -Path $tempDir | Out-Null
try {
    $rcPath = Join-Path $tempDir 'pm.rc'
    $resPath = Join-Path $tempDir 'pm.res'

    Push-Location $RepoRoot
    try {
        & go run ./build/windowsversion -version $Version -out $rcPath
    } finally {
        Pop-Location
    }
    if ($LASTEXITCODE -ne 0) {
        throw "windowsversion generator failed with exit code $LASTEXITCODE"
    }

    $rcExe = Find-WindowsSdkTool -ToolName 'rc.exe'
    $cvtres = Find-Cvtres

    & $rcExe /nologo /fo $resPath $rcPath
    if ($LASTEXITCODE -ne 0) {
        throw "rc.exe failed with exit code $LASTEXITCODE"
    }

    & $cvtres /NOLOGO /MACHINE:$machine /OUT:$OutputPath $resPath
    if ($LASTEXITCODE -ne 0) {
        throw "cvtres.exe failed with exit code $LASTEXITCODE"
    }

    Write-Host "generated VERSIONINFO resource: $OutputPath ($GoArch/$machine)"
} finally {
    Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $tempDir
}
