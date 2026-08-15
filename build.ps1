# build.ps1
# Aryntra Aayam build script — produces cross-platform release binaries.
#
# Usage:
#   .\build.ps1              # build all targets
#   .\build.ps1 -Version 1.0.0

param(
    [string]$Version = "1.0.0"
)

$ErrorActionPreference = "Stop"

$Module   = "aayam"
$OutDir   = "dist"
$LdFlags  = "-s -w -X main.version=$Version"

$Targets = @(
    @{ OS = "windows"; Arch = "amd64"; Ext = ".exe" },
    @{ OS = "linux";   Arch = "amd64"; Ext = ""     },
    @{ OS = "darwin";  Arch = "arm64"; Ext = ""     }
)

Write-Host "Aryntra Aayam build script" -ForegroundColor Cyan
Write-Host "Version : $Version"
Write-Host "Output  : $OutDir"
Write-Host ""

# Verify the project builds cleanly before producing artifacts.
Write-Host "Verifying build..." -ForegroundColor Yellow
go build ./...
if ($LASTEXITCODE -ne 0) {
    Write-Error "go build ./... failed. Aborting."
    exit 1
}

Write-Host "Running tests..." -ForegroundColor Yellow
go test ./...
if ($LASTEXITCODE -ne 0) {
    Write-Error "go test ./... failed. Aborting."
    exit 1
}

Write-Host "Running vet..." -ForegroundColor Yellow
go vet ./...
if ($LASTEXITCODE -ne 0) {
    Write-Error "go vet ./... failed. Aborting."
    exit 1
}

# Produce release binaries.
if (-not (Test-Path $OutDir)) {
    New-Item -ItemType Directory -Path $OutDir | Out-Null
}

foreach ($target in $Targets) {
    $os   = $target.OS
    $arch = $target.Arch
    $ext  = $target.Ext
    $name = "$Module-$os-$arch$ext"
    $out  = Join-Path $OutDir $name

    Write-Host "Building $name ..." -ForegroundColor Green

    $env:GOOS   = $os
    $env:GOARCH = $arch

    go build -ldflags $LdFlags -o $out .
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Build failed for $name"
        exit 1
    }
}

# Restore host environment.
Remove-Item Env:\GOOS   -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "Build complete." -ForegroundColor Cyan
Write-Host ""
Get-ChildItem $OutDir | Select-Object Name, Length | Format-Table -AutoSize
