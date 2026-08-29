# Build script for procscope on Windows
# Usage: .\build.ps1 [command]
# Commands: build, build-all, clean, help

param(
    [string]$Command = "build"
)

$BinDir = "bin"

function Build {
    Write-Host "Building procscope for Windows..."
    if (!(Test-Path $BinDir)) {
        New-Item -ItemType Directory -Path $BinDir | Out-Null
    }
    & go build -o "$BinDir/procscope.exe" .
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Build complete: $BinDir/procscope.exe"
    } else {
        Write-Host "Build failed!"
        exit 1
    }
}

function BuildAll {
    Write-Host "Building procscope for all platforms..."
    if (!(Test-Path $BinDir)) {
        New-Item -ItemType Directory -Path $BinDir | Out-Null
    }

    Write-Host "Building for Windows..."
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    & go build -o "$BinDir/procscope.exe" .
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Windows build failed!"
        exit 1
    }

    Write-Host "Building for Linux..."
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    & go build -o "$BinDir/procscope-linux" .
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Linux build failed!"
        exit 1
    }

    Write-Host "Building for macOS..."
    $env:GOOS = "darwin"
    $env:GOARCH = "amd64"
    & go build -o "$BinDir/procscope-mac" .
    if ($LASTEXITCODE -ne 0) {
        Write-Host "macOS build failed!"
        exit 1
    }

    # Reset environment
    $env:GOOS = ""
    $env:GOARCH = ""

    Write-Host "All builds complete in $BinDir/"
}

function Clean {
    Write-Host "Cleaning build artifacts..."
    if (Test-Path $BinDir) {
        Remove-Item -Recurse -Force $BinDir
    }
    Write-Host "Clean complete"
}

function ShowHelp {
    Write-Host "procscope build script"
    Write-Host ""
    Write-Host "Usage: .\build.ps1 [command]"
    Write-Host ""
    Write-Host "Commands:"
    Write-Host "  build      - Build for Windows (default)"
    Write-Host "  build-all  - Build for Windows, Linux, and macOS"
    Write-Host "  clean      - Remove build artifacts"
    Write-Host "  help       - Show this help message"
}

switch ($Command) {
    "build" { Build }
    "build-all" { BuildAll }
    "clean" { Clean }
    "help" { ShowHelp }
    default { 
        Write-Host "Unknown command: $Command"
        ShowHelp
        exit 1
    }
}
