# Colossus CLI Build Script for Windows
# PowerShell script to build Colossus with llama.cpp support

param(
    [string]$BuildType = "cpu",
    [string]$CudaPath = "C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v12.3",
    [string]$OutputDir = "bin",
    [switch]$Verbose,
    [switch]$Clean,
    [switch]$Help
)

# Colors for output
$ErrorColor = "Red"
$WarningColor = "Yellow" 
$InfoColor = "Green"
$VerboseColor = "Cyan"

function Write-Info($message) {
    Write-Host $message -ForegroundColor $InfoColor
}

function Write-Warning($message) {
    Write-Host $message -ForegroundColor $WarningColor
}

function Write-Error($message) {
    Write-Host $message -ForegroundColor $ErrorColor
}

function Write-Verbose($message) {
    if ($Verbose) {
        Write-Host $message -ForegroundColor $VerboseColor
    }
}

function Show-Help {
    Write-Host @"
Colossus CLI Build Script for Windows

USAGE:
    .\build.ps1 [OPTIONS]

OPTIONS:
    -BuildType <type>    Build type: cpu, cuda, rocm (default: cpu)
    -CudaPath <path>     Path to CUDA installation (default: C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v12.3)
    -OutputDir <dir>     Output directory for binaries (default: bin)
    -Verbose             Enable verbose output
    -Clean               Clean build artifacts before building
    -Help                Show this help message

EXAMPLES:
    .\build.ps1                           # Build CPU version
    .\build.ps1 -BuildType cuda           # Build with CUDA support
    .\build.ps1 -BuildType cuda -Verbose  # Build with CUDA and verbose output
    .\build.ps1 -Clean                    # Clean and build CPU version

REQUIREMENTS:
    - Go 1.21 or later
    - Git
    - Visual Studio Build Tools or Visual Studio (for C++ compilation)
    - CUDA Toolkit (for CUDA builds)
    - CMake (recommended)

"@
}

if ($Help) {
    Show-Help
    exit 0
}

# Validate build type
$ValidBuildTypes = @("cpu", "cuda", "rocm")
if ($BuildType -notin $ValidBuildTypes) {
    Write-Error "Invalid build type: $BuildType. Valid types: $($ValidBuildTypes -join ', ')"
    exit 1
}

Write-Info "🚀 Building Colossus CLI for Windows"
Write-Info "Build Type: $BuildType"
Write-Info "Output Directory: $OutputDir"

# Clean if requested
if ($Clean) {
    Write-Info "🧹 Cleaning build artifacts..."
    if (Test-Path $OutputDir) {
        Remove-Item -Recurse -Force $OutputDir
    }
    if (Test-Path "third_party/llama.cpp/build") {
        Remove-Item -Recurse -Force "third_party/llama.cpp/build"
    }
    Write-Info "✅ Clean complete"
}

# Create output directory
if (!(Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

# Check prerequisites
Write-Info "🔍 Checking prerequisites..."

# Check Go
try {
    $goVersion = & go version 2>$null
    Write-Verbose "Go version: $goVersion"
} catch {
    Write-Error "Go is not installed or not in PATH"
    exit 1
}

# Check Git
try {
    $gitVersion = & git --version 2>$null
    Write-Verbose "Git version: $gitVersion"
} catch {
    Write-Error "Git is not installed or not in PATH"
    exit 1
}

# Check if llama.cpp submodule is initialized
if (!(Test-Path "third_party/llama.cpp/CMakeLists.txt")) {
    Write-Info "📦 Initializing llama.cpp submodule..."
    & git submodule update --init --recursive
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to initialize llama.cpp submodule"
        exit 1
    }
}

# Build llama.cpp
Write-Info "🔨 Building llama.cpp..."
Push-Location "third_party/llama.cpp"

try {
    # Create build directory
    if (!(Test-Path "build")) {
        New-Item -ItemType Directory -Path "build" | Out-Null
    }
    
    Push-Location "build"
    
    # Configure CMake based on build type
    $cmakeArgs = @("-DCMAKE_BUILD_TYPE=Release")
    
    switch ($BuildType) {
        "cuda" {
            if (!(Test-Path $CudaPath)) {
                Write-Error "CUDA not found at: $CudaPath"
                Write-Info "Please install CUDA Toolkit or specify correct path with -CudaPath"
                exit 1
            }
            
            Write-Info "🎯 Building with CUDA support..."
            $cmakeArgs += "-DLLAMA_CUBLAS=ON"
            $cmakeArgs += "-DCUDAToolkit_ROOT=`"$CudaPath`""
            
            # Set CUDA environment
            $env:CUDA_PATH = $CudaPath
            $env:PATH = "$CudaPath\bin;$env:PATH"
        }
        "rocm" {
            Write-Warning "ROCm support on Windows is experimental"
            $cmakeArgs += "-DLLAMA_HIPBLAS=ON"
        }
        default {
            Write-Info "🖥️ Building CPU-only version..."
        }
    }
    
    # Run CMake configure
    Write-Verbose "CMake configure arguments: $($cmakeArgs -join ' ')"
    & cmake .. @cmakeArgs
    if ($LASTEXITCODE -ne 0) {
        Write-Error "CMake configuration failed"
        exit 1
    }
    
    # Build
    & cmake --build . --config Release --parallel
    if ($LASTEXITCODE -ne 0) {
        Write-Error "llama.cpp build failed"
        exit 1
    }
    
    Write-Info "✅ llama.cpp build complete"
} finally {
    Pop-Location
    Pop-Location
}

# Build Colossus
Write-Info "🔨 Building Colossus..."

# Set CGO environment variables
$env:CGO_ENABLED = "1"

# Set build-specific flags
$buildTags = @()
$cgoFlags = @()
$ldFlags = @()

switch ($BuildType) {
    "cuda" {
        $buildTags += "cuda"
        $cgoFlags += "-I$(Resolve-Path 'third_party/llama.cpp')"
        $cgoFlags += "-I`"$CudaPath\include`""
        $cgoFlags += "-DGGML_USE_CUBLAS"
        
        $ldFlags += "-L$(Resolve-Path 'third_party/llama.cpp/build/Release')"
        $ldFlags += "-L`"$CudaPath\lib\x64`""
        $ldFlags += "-lllama"
        $ldFlags += "-lcublas"
        $ldFlags += "-lcudart"
    }
    "rocm" {
        $buildTags += "rocm"
        Write-Warning "ROCm build on Windows is experimental"
    }
    default {
        # CPU build - check if we have the llama.cpp library
        $llamaLib = "third_party/llama.cpp/build/Release/llama.lib"
        if (Test-Path $llamaLib) {
            $cgoFlags += "-I$(Resolve-Path 'third_party/llama.cpp')"
            $ldFlags += "-L$(Resolve-Path 'third_party/llama.cpp/build/Release')"
            $ldFlags += "-lllama"
        } else {
            Write-Warning "llama.cpp library not found, building without real inference"
        }
    }
}

# Set environment variables for CGO
if ($cgoFlags) {
    $env:CGO_CFLAGS = $cgoFlags -join " "
    Write-Verbose "CGO_CFLAGS: $env:CGO_CFLAGS"
}

if ($ldFlags) {
    $env:CGO_LDFLAGS = $ldFlags -join " "
    Write-Verbose "CGO_LDFLAGS: $env:CGO_LDFLAGS"
}

# Build Go binary
$goBuildArgs = @("build")
if ($buildTags) {
    $goBuildArgs += "-tags"
    $goBuildArgs += ($buildTags -join ",")
}

$goBuildArgs += "-ldflags"
$goBuildArgs += "-s -w"  # Strip debug info for smaller binary
$goBuildArgs += "-o"
$goBuildArgs += "$OutputDir/colossus.exe"
$goBuildArgs += "."

Write-Verbose "Go build command: go $($goBuildArgs -join ' ')"

& go @goBuildArgs
if ($LASTEXITCODE -ne 0) {
    Write-Error "Go build failed"
    exit 1
}

Write-Info "✅ Colossus build complete!"

# Test the build
Write-Info "🧪 Testing build..."
try {
    $version = & "$OutputDir/colossus.exe" --help 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Info "✅ Build test successful"
    } else {
        Write-Warning "Build test failed"
    }
} catch {
    Write-Warning "Could not test build"
}

# Show GPU info if available
try {
    Write-Info "📊 GPU Information:"
    & "$OutputDir/colossus.exe" gpu info
} catch {
    Write-Verbose "Could not get GPU info"
}

Write-Info "🎉 Build complete!"
Write-Info "Binary location: $OutputDir/colossus.exe"

if ($BuildType -eq "cuda") {
    Write-Info ""
    Write-Info "CUDA build complete! To use:"
    Write-Info "  $env:COLOSSUS_INFERENCE_ENGINE=llamacpp"
    Write-Info "  .\$OutputDir\colossus.exe serve"
}
