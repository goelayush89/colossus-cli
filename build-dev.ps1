# Quick Development Build Script for Colossus CLI
# Builds Colossus with simulated inference for development and testing

param(
    [string]$OutputDir = "bin",
    [switch]$Verbose,
    [switch]$Clean
)

$InfoColor = "Green"
$WarningColor = "Yellow"

function Write-Info($message) {
    Write-Host $message -ForegroundColor $InfoColor
}

function Write-Warning($message) {
    Write-Host $message -ForegroundColor $WarningColor
}

Write-Info "🚀 Quick Development Build for Colossus CLI"

# Clean if requested
if ($Clean) {
    Write-Info "🧹 Cleaning..."
    if (Test-Path $OutputDir) {
        Remove-Item -Recurse -Force $OutputDir
    }
}

# Create output directory
if (!(Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

# Build with simulated inference (no CGO required)
Write-Info "🔨 Building development version..."

$env:CGO_ENABLED = "0"  # Disable CGO for faster builds

& go build -ldflags "-s -w" -o "$OutputDir/colossus.exe" .

if ($LASTEXITCODE -eq 0) {
    Write-Info "✅ Development build complete!"
    Write-Info "Binary: $OutputDir/colossus.exe"
    Write-Info ""
    Write-Info "This build uses simulated inference for development."
    Write-Info "For real inference, use: .\build.ps1 -BuildType cuda"
    Write-Info ""
    Write-Info "Quick test:"
    Write-Info "  .\$OutputDir\colossus.exe gpu info"
    Write-Info "  .\$OutputDir\colossus.exe serve"
} else {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}
