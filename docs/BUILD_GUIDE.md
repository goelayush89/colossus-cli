# Build Guide: Real Inference with GPU Support

This guide provides step-by-step instructions for building Colossus CLI with real llama.cpp inference and GPU acceleration.

## Prerequisites

### System Requirements

**For CPU-only builds:**
- Go 1.21 or later
- GCC/Clang compiler
- Make
- Git

**For CUDA builds (NVIDIA GPUs):**
- NVIDIA GPU with CUDA compute capability 3.5+
- CUDA Toolkit 11.0 or later (12.x recommended)
- NVIDIA drivers 450.80.02 or later
- All CPU requirements

**For ROCm builds (AMD GPUs):**
- AMD GPU with ROCm support
- ROCm 5.0 or later
- Linux only (Ubuntu 20.04/22.04, RHEL 8/9)
- All CPU requirements

## Quick Start

### 1. Clone and Setup

```bash
# Clone the repository
git clone https://github.com/your-org/colossus-cli.git
cd colossus-cli

# Setup llama.cpp submodule
make setup-llamacpp
```

### 2. Choose Your Build Type

**CPU Only (Default):**
```bash
make build
```

**CUDA GPU Acceleration:**
```bash
# Build llama.cpp with CUDA support
make build-llamacpp-cuda

# Build Colossus with CUDA
make build-cuda
```

**ROCm GPU Acceleration:**
```bash
# Build llama.cpp with ROCm support  
make build-llamacpp-rocm

# Build Colossus with ROCm
make build-rocm
```

### 3. Test the Build

```bash
# Check GPU detection
./bin/colossus gpu info

# Start the server
./bin/colossus serve --verbose

# Download and test a model
./bin/colossus models pull microsoft/DialoGPT-medium
./bin/colossus chat microsoft/DialoGPT-medium
```

## Detailed Build Instructions

### Step 1: Environment Setup

**Install Go:**
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install golang-go git build-essential

# macOS
brew install go git

# Verify installation
go version
```

**Install CUDA (for NVIDIA GPUs):**
```bash
# Ubuntu 22.04
wget https://developer.download.nvidia.com/compute/cuda/12.3.1/local_installers/cuda_12.3.1_545.23.08_linux.run
sudo sh cuda_12.3.1_545.23.08_linux.run

# Add to PATH
export CUDA_HOME=/usr/local/cuda
export PATH=$CUDA_HOME/bin:$PATH
export LD_LIBRARY_PATH=$CUDA_HOME/lib64:$LD_LIBRARY_PATH

# Verify installation
nvidia-smi
nvcc --version
```

**Install ROCm (for AMD GPUs):**
```bash
# Ubuntu 22.04
wget https://repo.radeon.com/amdgpu-install/22.40.5/ubuntu/jammy/amdgpu-install_5.4.50405-1_all.deb
sudo dpkg -i amdgpu-install_5.4.50405-1_all.deb
sudo amdgpu-install --usecase=rocm

# Add to PATH
export ROCM_PATH=/opt/rocm
export PATH=$ROCM_PATH/bin:$PATH
export LD_LIBRARY_PATH=$ROCM_PATH/lib:$LD_LIBRARY_PATH

# Verify installation
rocm-smi
hipcc --version
```

### Step 2: Setup Dependencies

```bash
# Clone the repository
git clone https://github.com/your-org/colossus-cli.git
cd colossus-cli

# Setup all dependencies including llama.cpp
make deps
make setup-llamacpp
```

### Step 3: Build llama.cpp Library

Choose the appropriate build based on your hardware:

**CPU Only:**
```bash
make build-llamacpp
```

**CUDA (NVIDIA):**
```bash
make build-llamacpp-cuda

# Or with custom CUDA path
CUDA_PATH=/usr/local/cuda-12 make build-llamacpp-cuda
```

**ROCm (AMD):**
```bash
make build-llamacpp-rocm

# Or with custom ROCm path
ROCM_PATH=/opt/rocm make build-llamacpp-rocm
```

### Step 4: Build Colossus

**CPU Only:**
```bash
make build-cpu
```

**CUDA:**
```bash
make build-cuda

# Or using BUILD_TYPE
BUILD_TYPE=cuda make build
```

**ROCm:**
```bash
make build-rocm

# Or using BUILD_TYPE  
BUILD_TYPE=rocm make build
```

### Step 5: Verify Build

```bash
# Check the binary
./bin/colossus --help

# Test GPU detection
./bin/colossus gpu info

# Expected output for CUDA:
# GPU Acceleration Status: ✓ Available (cuda)
# GPU Type: cuda
# Device Count: 1
# Driver Version: 525.60.11
```

## Advanced Build Options

### Custom Paths

```bash
# Custom CUDA installation
CUDA_PATH=/opt/cuda make build-cuda

# Custom ROCm installation
ROCM_PATH=/opt/rocm-5.7 make build-rocm

# Custom llama.cpp location
LLAMA_CPP_DIR=./my-llama-cpp make build
```

### Cross-Platform Builds

```bash
# Build for multiple platforms
make build-all

# Manual cross-compilation
GOOS=linux GOARCH=amd64 make build-cpu
GOOS=windows GOARCH=amd64 make build-cpu
GOOS=darwin GOARCH=arm64 make build-cpu
```

### Development Builds

```bash
# Debug build with verbose output
CGO_CFLAGS="-g -O0" make build-cuda

# Build with custom version
VERSION=v1.0.0-beta make build

# Development build with all features
make dev-setup
make build-cuda
```

## Configuration and Usage

### Environment Variables

```bash
# Inference engine selection
export COLOSSUS_INFERENCE_ENGINE=llamacpp

# GPU configuration
export COLOSSUS_GPU_LAYERS=32
export CUDA_VISIBLE_DEVICES=0,1
export ROCR_VISIBLE_DEVICES=0

# Hugging Face token for model downloads
export HUGGINGFACE_TOKEN=your_token_here

# Model storage
export COLOSSUS_MODELS_PATH=/path/to/models
```

### Running with GPU Acceleration

```bash
# Check GPU status
./bin/colossus gpu info

# Start server with GPU acceleration
COLOSSUS_INFERENCE_ENGINE=llamacpp ./bin/colossus serve --verbose

# Download a model from Hugging Face
./bin/colossus models pull microsoft/DialoGPT-small

# Start interactive chat
./bin/colossus chat microsoft/DialoGPT-small
```

### API Usage

```bash
# Test the API
curl http://localhost:11434/api/tags

# Chat completion with GPU acceleration
curl -X POST http://localhost:11434/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "model": "microsoft/DialoGPT-small",
    "messages": [{"role": "user", "content": "Hello!"}],
    "stream": true
  }'
```

## Troubleshooting

### Common Build Issues

**CUDA not found:**
```bash
# Check CUDA installation
nvidia-smi
echo $CUDA_HOME

# Fix PATH issues
export CUDA_HOME=/usr/local/cuda
export PATH=$CUDA_HOME/bin:$PATH
export LD_LIBRARY_PATH=$CUDA_HOME/lib64:$LD_LIBRARY_PATH
```

**ROCm not found:**
```bash
# Check ROCm installation
rocm-smi
echo $ROCM_PATH

# Fix PATH issues
export ROCM_PATH=/opt/rocm
export PATH=$ROCM_PATH/bin:$PATH
export LD_LIBRARY_PATH=$ROCM_PATH/lib:$LD_LIBRARY_PATH
```

**llama.cpp build failures:**
```bash
# Clean and rebuild
cd third_party/llama.cpp
make clean
make LLAMA_CUBLAS=1  # or LLAMA_HIPBLAS=1 for ROCm

# Check for missing dependencies
sudo apt install build-essential cmake
```

**Go build failures:**
```bash
# Update Go modules
go mod tidy
go mod download

# Check CGO settings
go env CGO_ENABLED  # should be "1"

# Clean Go cache
go clean -cache
go clean -modcache
```

### Runtime Issues

**GPU not detected:**
```bash
# Check drivers
nvidia-smi  # for CUDA
rocm-smi    # for ROCm

# Check environment
./bin/colossus gpu info

# Force GPU engine
export COLOSSUS_FORCE_LLAMACPP=true
```

**Out of memory errors:**
```bash
# Reduce GPU layers
export COLOSSUS_GPU_LAYERS=16

# Check GPU memory
nvidia-smi  # look at memory usage
```

**Model download failures:**
```bash
# Check internet connection
curl -I https://huggingface.co

# Set Hugging Face token
export HUGGINGFACE_TOKEN=your_token

# Check disk space
df -h
```

## Performance Optimization

### GPU Memory Management

```bash
# Monitor GPU usage
watch -n 1 nvidia-smi

# Optimize GPU layers based on available memory
./bin/colossus gpu info  # shows recommended settings
```

### Model Selection

```bash
# List available model formats
./bin/colossus models list

# Download optimized quantized models
./bin/colossus models pull TheBloke/Llama-2-7B-Chat-GGUF
```

## Docker Support

### CUDA Docker

```dockerfile
FROM nvidia/cuda:12.3-devel-ubuntu22.04

RUN apt-get update && apt-get install -y \
    golang-go git build-essential

COPY . /app
WORKDIR /app

RUN make setup-llamacpp && \
    make build-llamacpp-cuda && \
    make build-cuda

EXPOSE 11434
CMD ["./bin/colossus", "serve"]
```

### ROCm Docker

```dockerfile
FROM rocm/dev-ubuntu-22.04:5.7

RUN apt-get update && apt-get install -y \
    golang-go git

COPY . /app
WORKDIR /app

RUN make setup-llamacpp && \
    make build-llamacpp-rocm && \
    make build-rocm

EXPOSE 11434
CMD ["./bin/colossus", "serve"]
```

### Building Docker Images

```bash
# CUDA image
docker build -f Dockerfile.cuda -t colossus:cuda .

# ROCm image  
docker build -f Dockerfile.rocm -t colossus:rocm .

# Run with GPU support
docker run --gpus all -p 11434:11434 colossus:cuda
```

## Development and Testing

### Running Tests

```bash
# Run all tests
make test

# Run with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./...
```

### Development Workflow

```bash
# Setup development environment
make dev-setup

# Format code
make fmt

# Run linter
make lint

# Live development
make run
```

### Contributing

```bash
# Setup pre-commit hooks
make dev-setup

# Run full test suite
make test

# Build all variants
make build-all

# Generate documentation
go doc ./...
```

This comprehensive build guide should help you successfully build and deploy Colossus CLI with real llama.cpp inference and GPU acceleration!
