# piomx installer for Windows — downloads the latest release binary.
# Usage:
#   powershell -c "irm https://raw.githubusercontent.com/runtime-terror404/piomx/main/install.ps1 | iex"
#   .\install.ps1                # install or update
#   .\install.ps1 -Uninstall     # remove binary + optional config cleanup

param([switch]$Uninstall)

$Repo = "runtime-terror404/piomx"
$BinName = "piomx.exe"

# ---- helpers ----

function Get-InstallDir {
    $dir = "$env:USERPROFILE\.local\bin"
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
    }
    return $dir
}

function Add-ToPath {
    $installDir = Get-InstallDir
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -notlike "*$installDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installDir", "User")
        Write-Host "Added $installDir to user PATH."
    }
}

# ---- uninstall ----

function Do-Uninstall {
    $installDir = Get-InstallDir
    $binPath = Join-Path $installDir $BinName

    if (Test-Path $binPath) {
        Remove-Item $binPath -Force
        Write-Host "piomx uninstalled."
    } else {
        Write-Host "piomx is not installed at $binPath"
    }

    $configDir = "$env:APPDATA\piomx"
    if (Test-Path $configDir) {
        $answer = Read-Host "Remove config at $configDir? [y/N]"
        if ($answer -eq 'y' -or $answer -eq 'Y') {
            Remove-Item $configDir -Recurse -Force
            Write-Host "Config removed."
        } else {
            Write-Host "Config kept at $configDir"
        }
    }
    exit 0
}

# ---- install ----

function Do-Install {
    $arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
    $os = "windows"

    $installDir = Get-InstallDir
    $binPath = Join-Path $installDir $BinName
    $tmpDir = Join-Path $env:TEMP "piomx-install"
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

    try {
        # Fetch latest release.
        Write-Host "Fetching latest release for $os/$arch..."
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
        $asset = $release.assets | Where-Object { $_.name -like "piomx_${os}_${arch}*" } | Select-Object -First 1

        if (-not $asset) {
            Write-Error "Could not find release for $os/$arch"
            exit 1
        }

        Write-Host "Downloading $($asset.name)..."
        Invoke-WebRequest -Uri $asset.browser_download_url -OutFile "$tmpDir\$BinName"

        if (Test-Path $binPath) {
            Write-Host "Replacing existing installation at $binPath"
        }

        Move-Item "$tmpDir\$BinName" $binPath -Force
        Write-Host "Installed $BinName to $binPath"

        Add-ToPath

        Write-Host ""
        Write-Host "Done. Open a new terminal and run 'piomx' to get started."
    } finally {
        Remove-Item $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

# ---- main ----

if ($Uninstall) {
    Do-Uninstall
} else {
    Do-Install
}
