# Colossus CLI - Development Build with Real Inference
# Quick build script that enables real llama.cpp inference if available

param(
    [switch]$Force,
    [switch]$Verbose
)

$InfoColor = "Green"
$WarnColor = "Yellow"
$ErrorColor = "Red"

function Write-Info($msg) { Write-Host $msg -ForegroundColor $InfoColor }
function Write-Warn($msg) { Write-Host $msg -ForegroundColor $WarnColor }
function Write-Error($msg) { Write-Host $msg -ForegroundColor $ErrorColor }

Write-Info "🛠️ Building Colossus CLI (Development with Real Inference)"
Write-Info "=========================================================="

# Check if llama.cpp is available
$llamaCppAvailable = $false

if (Test-Path "third_party/llama.cpp") {
    Write-Info "✅ llama.cpp submodule found"
    $llamaCppAvailable = $true
} else {
    Write-Warn "⚠️ llama.cpp submodule not found"
    if ($Force) {
        Write-Info "📥 Initializing llama.cpp submodule..."
        git submodule update --init --recursive third_party/llama.cpp
        if ($LASTEXITCODE -eq 0) {
            $llamaCppAvailable = $true
            Write-Info "✅ llama.cpp submodule initialized"
        }
    }
}

# Create windows-binary directory
New-Item -ItemType Directory -Path "windows-binary" -Force | Out-Null

# Build based on availability
if ($llamaCppAvailable -and (Test-Path "third_party/llama.cpp/libllama.dll" -or $Force)) {
    Write-Info "🔨 Building with real llama.cpp inference..."
    
    # Build with CGO enabled
    $env:CGO_ENABLED = "1"
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    $env:CGO_CFLAGS = "-I./third_party/llama.cpp/include"
    $env:CGO_LDFLAGS = "-L./third_party/llama.cpp"
    
    go build -tags="llamacpp_cgo" -ldflags="-s -w" -o "windows-binary/colossus.exe" .
    
    if ($LASTEXITCODE -eq 0) {
        Write-Info "✅ Built with real inference support!"
        Write-Info "🎯 To use real inference: Set COLOSSUS_INFERENCE_ENGINE=llamacpp"
    } else {
        Write-Warn "⚠️ CGO build failed, falling back to simulation..."
        $llamaCppAvailable = $false
    }
}

if (!$llamaCppAvailable -or $LASTEXITCODE -ne 0) {
    Write-Info "🔨 Building with simulation (development mode)..."
    
    # Build without CGO
    $env:CGO_ENABLED = "0"
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    
    go build -ldflags="-s -w" -o "windows-binary/colossus.exe" .
    
    if ($LASTEXITCODE -eq 0) {
        Write-Info "✅ Built with simulation mode!"
        Write-Info "🎯 This version uses simulated responses for testing"
    } else {
        Write-Error "❌ Build failed completely"
        exit 1
    }
}

# Copy batch files and docs
Copy-Item "windows-binary/*.bat" "windows-binary/" -Force -ErrorAction SilentlyContinue
Copy-Item "windows-binary/README.txt" "windows-binary/" -Force -ErrorAction SilentlyContinue

$binarySize = [math]::Round((Get-Item "windows-binary/colossus.exe").Length / 1MB, 1)

Write-Info ""
Write-Info "🎉 Development Build Complete!"
Write-Info "=============================="
Write-Info "Binary: windows-binary/colossus.exe (${binarySize}MB)"

if ($llamaCppAvailable) {
    Write-Info "Mode: Real Inference Available"
    Write-Info "🚀 To use real AI:"
    Write-Info "   $env:COLOSSUS_INFERENCE_ENGINE='llamacpp'"
    Write-Info "   .\colossus.exe serve"
} else {
    Write-Info "Mode: Simulation Only"
    Write-Info "🚀 To enable real inference:"
    Write-Info "   Run: .\build-production.ps1"
}

Write-Info ""
Write-Info "📖 Quick Start:"
Write-Info "   .\windows-binary\colossus.exe serve"
Write-Info "   .\windows-binary\colossus.exe models pull qwen"
Write-Info "   .\windows-binary\colossus.exe chat qwen"
