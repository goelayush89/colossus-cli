# Colossus CLI - Production Build Script with Real llama.cpp Integration
# This script builds a production-ready Colossus CLI with actual llama.cpp inference

param(
    [ValidateSet("cpu", "cuda", "rocm")]
    [string]$Target = "cpu",
    [switch]$Clean,
    [switch]$Verbose
)

$InfoColor = "Green"
$WarnColor = "Yellow"
$ErrorColor = "Red"

function Write-Info($msg) { Write-Host $msg -ForegroundColor $InfoColor }
function Write-Warn($msg) { Write-Host $msg -ForegroundColor $WarnColor }
function Write-Error($msg) { Write-Host $msg -ForegroundColor $ErrorColor }

Write-Info "🚀 Building Production Colossus CLI with Real Inference"
Write-Info "========================================================"
Write-Info "Target: $Target"
Write-Info ""

# Check prerequisites
Write-Info "🔍 Checking prerequisites..."

# Check for Git
if (!(Get-Command git -ErrorAction SilentlyContinue)) {
    Write-Error "❌ Git not found. Please install Git first."
    exit 1
}

# Check for CMake (needed for llama.cpp)
if (!(Get-Command cmake -ErrorAction SilentlyContinue)) {
    Write-Warn "⚠️ CMake not found. Attempting to install via winget..."
    try {
        winget install Kitware.CMake
        Write-Info "✅ CMake installed successfully"
    } catch {
        Write-Error "❌ Failed to install CMake. Please install manually."
        exit 1
    }
}

# Check for Visual Studio Build Tools
$vsPath = "${env:ProgramFiles(x86)}\Microsoft Visual Studio\2019\BuildTools\Common7\Tools\VsDevCmd.bat"
if (!(Test-Path $vsPath)) {
    $vsPath = "${env:ProgramFiles}\Microsoft Visual Studio\2022\Community\Common7\Tools\VsDevCmd.bat"
    if (!(Test-Path $vsPath)) {
        Write-Error "❌ Visual Studio Build Tools not found. Please install Visual Studio Build Tools."
        exit 1
    }
}

Write-Info "✅ Prerequisites checked"

# Clean if requested
if ($Clean) {
    Write-Info "🧹 Cleaning previous builds..."
    if (Test-Path "third_party/llama.cpp/build") { Remove-Item -Recurse -Force "third_party/llama.cpp/build" }
    if (Test-Path "third_party/llama.cpp/libllama.a") { Remove-Item -Force "third_party/llama.cpp/libllama.a" }
    if (Test-Path "third_party/llama.cpp/llama.dll") { Remove-Item -Force "third_party/llama.cpp/llama.dll" }
    if (Test-Path "windows-binary") { Remove-Item -Recurse -Force "windows-binary" }
    Write-Info "✅ Cleaned"
}

# Initialize submodules
Write-Info "📥 Initializing llama.cpp submodule..."
git submodule update --init --recursive third_party/llama.cpp
if ($LASTEXITCODE -ne 0) {
    Write-Error "❌ Failed to initialize llama.cpp submodule"
    exit 1
}

# Build llama.cpp
Write-Info "🔨 Building llama.cpp library..."
Push-Location "third_party/llama.cpp"

try {
    # Create build directory
    if (!(Test-Path "build")) { New-Item -ItemType Directory -Path "build" | Out-Null }
    Push-Location "build"
    
    # Configure CMake based on target
    $cmakeArgs = @("-DCMAKE_BUILD_TYPE=Release", "-DBUILD_SHARED_LIBS=ON", "-DLLAMA_BUILD_EXAMPLES=OFF", "-DLLAMA_BUILD_TESTS=OFF")
    
    switch ($Target) {
        "cuda" {
            Write-Info "🚀 Configuring for NVIDIA CUDA acceleration..."
            $cmakeArgs += "-DLLAMA_CUBLAS=ON"
        }
        "rocm" {
            Write-Info "🔥 Configuring for AMD ROCm acceleration..."
            $cmakeArgs += "-DLLAMA_HIPBLAS=ON"
        }
        default {
            Write-Info "💻 Configuring for CPU-only inference..."
        }
    }
    
    # Generate build files
    cmake .. @cmakeArgs
    if ($LASTEXITCODE -ne 0) {
        Write-Error "❌ CMake configuration failed"
        exit 1
    }
    
    # Build the library
    cmake --build . --config Release --parallel
    if ($LASTEXITCODE -ne 0) {
        Write-Error "❌ llama.cpp build failed"
        exit 1
    }
    
    Write-Info "✅ llama.cpp built successfully"
    
} finally {
    Pop-Location  # Exit build directory
    Pop-Location  # Exit llama.cpp directory
}

# Copy llama.cpp artifacts to root for linking
Write-Info "📦 Copying llama.cpp artifacts..."
$llamaLib = "third_party/llama.cpp/build/Release/llama.dll"
$ggmlLib = "third_party/llama.cpp/build/Release/ggml.dll"

if (!(Test-Path $llamaLib)) {
    # Try different paths based on build system
    $llamaLib = "third_party/llama.cpp/build/libllama.dll"
    if (!(Test-Path $llamaLib)) {
        $llamaLib = "third_party/llama.cpp/build/src/Release/llama.dll"
    }
}

if (Test-Path $llamaLib) {
    Copy-Item $llamaLib "third_party/llama.cpp/" -Force
    Write-Info "✅ Copied llama.dll"
} else {
    Write-Error "❌ Could not find llama.dll after build"
    exit 1
}

# Create windows-binary directory
New-Item -ItemType Directory -Path "windows-binary" -Force | Out-Null

# Build Colossus with CGO enabled
Write-Info "🔨 Building Colossus CLI with real inference..."

# Set environment variables for CGO
$env:CGO_ENABLED = "1"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

# Set CGO flags for llama.cpp
$env:CGO_CFLAGS = "-I./third_party/llama.cpp/include -I./third_party/llama.cpp"
$env:CGO_LDFLAGS = "-L./third_party/llama.cpp -lllama"

# Add additional flags based on target
switch ($Target) {
    "cuda" {
        $env:CGO_CFLAGS += " -DGGML_USE_CUBLAS"
        $env:CGO_LDFLAGS += " -lcublas -lcudart"
    }
    "rocm" {
        $env:CGO_CFLAGS += " -DGGML_USE_HIPBLAS"
        $env:CGO_LDFLAGS += " -lhipblas -lrocblas"
    }
}

# Build the Go binary
Write-Info "Building Go binary with llama.cpp integration..."
go build -tags="llamacpp_cgo" -ldflags="-s -w" -o "windows-binary/colossus.exe" .

if ($LASTEXITCODE -eq 0) {
    Write-Info "✅ Colossus CLI built successfully with real inference!"
    
    # Copy required DLLs
    if (Test-Path "third_party/llama.cpp/llama.dll") {
        Copy-Item "third_party/llama.cpp/llama.dll" "windows-binary/" -Force
        Write-Info "✅ Copied llama.dll to binary directory"
    }
    
    # Copy batch files and documentation
    Copy-Item "windows-binary/*.bat" "windows-binary/" -Force -ErrorAction SilentlyContinue
    Copy-Item "windows-binary/README.txt" "windows-binary/" -Force -ErrorAction SilentlyContinue
    
} else {
    Write-Error "❌ Go build failed"
    exit 1
}

# Get binary info
$binarySize = [math]::Round((Get-Item "windows-binary/colossus.exe").Length / 1MB, 1)

Write-Info ""
Write-Info "🎉 Production Build Complete!"
Write-Info "=============================="
Write-Info "Target: $Target"
Write-Info "Binary: windows-binary/colossus.exe (${binarySize}MB)"
Write-Info "Features:"
Write-Info "  ✅ Real llama.cpp inference"
Write-Info "  ✅ GPU acceleration: $Target"
Write-Info "  ✅ Progress bars for downloads"
Write-Info "  ✅ Enhanced GGUF repositories"
Write-Info "  ✅ Production-ready performance"
Write-Info ""
Write-Info "🚀 Usage:"
Write-Info "  cd windows-binary"
Write-Info "  .\colossus.exe serve"
Write-Info "  .\colossus.exe models pull qwen"
Write-Info "  .\colossus.exe chat qwen"
Write-Info ""
Write-Info "🌟 Your Colossus CLI now has REAL AI inference!"
