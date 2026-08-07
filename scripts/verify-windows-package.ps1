# Copyright (c) 2026 Polymetrics AI
# SPDX-License-Identifier: AGPL-3.0-only

[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$Version,

    [string]$DistDir = 'dist\windows',

    [switch]$InstallX64
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$VersionPattern = '^v?([0-9]+)\.([0-9]+)\.([0-9]+)(?:[-+].*)?$'
if ($Version -notmatch $VersionPattern) {
    throw "Version '$Version' must look like vMAJOR.MINOR.PATCH"
}
$MsiVersion = "{0}.{1}.{2}" -f $Matches[1], $Matches[2], $Matches[3]
$FileVersion = "{0}.0" -f $MsiVersion
$DistPath = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($DistDir)

# arm64 is deliberately absent: go-duckdb ships no windows/arm64 library, so no
# arm64 pm.exe exists to package. Its reserved UpgradeCode is
# {EFEAFAA1-4276-509D-945A-D4F9BF7DBA30}, kept here so the same MSI upgrade
# identity is reused if the target ever comes back.
$Architectures = @(
    [pscustomobject]@{
        GoArch = 'amd64'
        WingetArch = 'x64'
        UpgradeCode = '{34C3F556-5634-5381-AE18-E1668FDECFA7}'
    }
)

function Assert-Equal {
    param(
        [Parameter(Mandatory = $true)] [string]$Name,
        [AllowNull()] [object]$Got,
        [AllowNull()] [object]$Want
    )
    if ($Got -ne $Want) {
        throw "$Name mismatch: got '$Got', want '$Want'"
    }
}

function Open-MsiDatabase {
    param([Parameter(Mandatory = $true)][string]$Path)

    $installer = New-Object -ComObject WindowsInstaller.Installer
    $database = $installer.OpenDatabase($Path, 0)
    return [pscustomobject]@{
        Installer = $installer
        Database = $database
    }
}

function Invoke-MsiScalar {
    param(
        [Parameter(Mandatory = $true)]$Database,
        [Parameter(Mandatory = $true)][string]$Query
    )

    $view = $Database.OpenView($Query)
    try {
        [void]$view.Execute()
        $record = $view.Fetch()
        if (-not $record) {
            return $null
        }
        return $record.StringData(1)
    } finally {
        [void]$view.Close()
    }
}

function Normalize-MsiScalar {
    param([AllowNull()][object]$Value)

    if ($null -eq $Value) {
        return $null
    }
    return ([string]$Value).Trim()
}

function Get-MsiProperty {
    param(
        [Parameter(Mandatory = $true)]$Database,
        [Parameter(Mandatory = $true)][string]$Property
    )
    $value = Invoke-MsiScalar -Database $Database -Query "SELECT ``Value`` FROM ``Property`` WHERE ``Property``='$Property'"
    return (Normalize-MsiScalar -Value $value)
}

function Assert-ExecutableVersionInfo {
    param([Parameter(Mandatory = $true)][string]$Path)

    if (-not (Test-Path $Path)) {
        throw "missing executable: $Path"
    }

    $info = (Get-Item $Path).VersionInfo
    Assert-Equal -Name "$Path CompanyName" -Got $info.CompanyName -Want 'Polymetrics AI'
    Assert-Equal -Name "$Path FileDescription" -Got $info.FileDescription -Want 'Polymetrics CLI'
    Assert-Equal -Name "$Path FileVersion" -Got $info.FileVersion -Want $FileVersion
    Assert-Equal -Name "$Path ProductName" -Got $info.ProductName -Want 'Polymetrics CLI'
    Assert-Equal -Name "$Path ProductVersion" -Got $info.ProductVersion -Want $FileVersion
    Assert-Equal -Name "$Path OriginalFilename" -Got $info.OriginalFilename -Want 'pm.exe'
}

function Assert-MsiStructure {
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][string]$UpgradeCode
    )

    if (-not (Test-Path $Path)) {
        throw "missing MSI: $Path"
    }

    $opened = Open-MsiDatabase -Path $Path
    $db = $opened.Database
    Assert-Equal -Name "$Path ProductName" -Got (Get-MsiProperty -Database $db -Property 'ProductName') -Want 'Polymetrics CLI'
    Assert-Equal -Name "$Path Manufacturer" -Got (Get-MsiProperty -Database $db -Property 'Manufacturer') -Want 'Polymetrics AI'
    Assert-Equal -Name "$Path ProductVersion" -Got (Get-MsiProperty -Database $db -Property 'ProductVersion') -Want $MsiVersion
    Assert-Equal -Name "$Path UpgradeCode" -Got (Get-MsiProperty -Database $db -Property 'UpgradeCode') -Want $UpgradeCode

    $fileName = Invoke-MsiScalar -Database $db -Query "SELECT ``FileName`` FROM ``File`` WHERE ``File``='PmExecutable'"
    if ($fileName -notmatch 'pm\.exe') {
        throw "$Path File table does not install pm.exe; got '$fileName'"
    }

    $pathRow = Invoke-MsiScalar -Database $db -Query "SELECT ``Value`` FROM ``Environment`` WHERE ``Environment``='AddPmToMachinePath'"
    if ($pathRow -notmatch '\[INSTALLFOLDER\]') {
        throw "$Path Environment table does not add INSTALLFOLDER to PATH; got '$pathRow'"
    }
}

function Invoke-Msiexec {
    param([Parameter(Mandatory = $true)][string[]]$Arguments)

    $process = Start-Process -FilePath msiexec.exe -ArgumentList $Arguments -Wait -PassThru -NoNewWindow
    if ($process.ExitCode -ne 0) {
        throw "msiexec $($Arguments -join ' ') failed with exit code $($process.ExitCode)"
    }
}

foreach ($arch in $Architectures) {
    $archDir = Join-Path $DistPath $arch.GoArch
    $exe = Join-Path $archDir 'pm.exe'
    $msi = Join-Path $archDir ("pm_{0}_windows_{1}.msi" -f $MsiVersion, $arch.GoArch)

    Assert-ExecutableVersionInfo -Path $exe
    Assert-MsiStructure -Path $msi -UpgradeCode $arch.UpgradeCode
    Write-Host "verified unsigned package structure for $($arch.GoArch): $msi"
}

if ($InstallX64) {
    $msi = Join-Path (Join-Path $DistPath 'amd64') ("pm_{0}_windows_amd64.msi" -f $MsiVersion)
    $opened = Open-MsiDatabase -Path $msi
    $productCode = Get-MsiProperty -Database $opened.Database -Property 'ProductCode'
    if (-not $productCode) {
        throw "could not read ProductCode from $msi"
    }

    $installDir = Join-Path $env:ProgramFiles 'Polymetrics\CLI'
    Invoke-Msiexec -Arguments @('/i', $msi, '/qn', '/norestart')
    try {
        $installedExe = Join-Path $installDir 'pm.exe'
        Assert-ExecutableVersionInfo -Path $installedExe
        & $installedExe version | Out-Host
        if ($LASTEXITCODE -ne 0) {
            throw "installed pm.exe version failed with exit code $LASTEXITCODE"
        }

        $machinePath = [Environment]::GetEnvironmentVariable('Path', 'Machine')
        if ($machinePath -notlike "*$installDir*") {
            throw "machine PATH does not include $installDir after install"
        }
    } finally {
        Invoke-Msiexec -Arguments @('/x', $productCode, '/qn', '/norestart')
    }

    if (Test-Path (Join-Path $installDir 'pm.exe')) {
        throw "pm.exe still exists after uninstall: $installDir"
    }
    $machinePathAfter = [Environment]::GetEnvironmentVariable('Path', 'Machine')
    if ($machinePathAfter -like "*$installDir*") {
        throw "machine PATH still includes $installDir after uninstall"
    }

    Write-Host "verified unsigned x64 MSI install/run/uninstall"
}
