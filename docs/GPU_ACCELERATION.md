# GPU Acceleration Guide

This guide explains how to enable GPU acceleration in Colossus CLI for faster model inference.

## Quick Start

1. **Check GPU availability**:
   ```bash
   colossus gpu info
   ```

2. **Enable GPU acceleration**:
   ```bash
   export COLOSSUS_INFERENCE_ENGINE=llamacpp
   export COLOSSUS_GPU_LAYERS=32
   ```

3. **Start server with GPU acceleration**:
   ```bash
   colossus serve --verbose
   ```

## Supported GPU Types

### NVIDIA CUDA

**Requirements:**
- NVIDIA GPU with CUDA compute capability 3.5+
- CUDA Toolkit 11.0 or later
- NVIDIA drivers 450.80.02 or later

**Installation:**
1. Install CUDA Toolkit from [NVIDIA's website](https://developer.nvidia.com/cuda-toolkit)
2. Verify installation: `nvidia-smi`
3. Set environment variables:
   ```bash
   export CUDA_HOME=/usr/local/cuda
   export PATH=$CUDA_HOME/bin:$PATH
   export LD_LIBRARY_PATH=$CUDA_HOME/lib64:$LD_LIBRARY_PATH
   ```

**Build Colossus with CUDA support:**
```bash
# Build llama.cpp with CUDA
cd third_party/llama.cpp
make LLAMA_CUBLAS=1

# Build Colossus
cd ../..
CGO_LDFLAGS="-lcublas -lcudart -L/usr/local/cuda/lib64/" go build
```

### AMD ROCm

**Requirements:**
- AMD GPU with ROCm support
- ROCm 5.0 or later
- Linux only (Ubuntu 20.04/22.04, RHEL 8/9)

**Installation:**
1. Install ROCm from [AMD's documentation](https://rocmdocs.amd.com/en/latest/deploy/linux/quick_start.html)
2. Verify installation: `rocm-smi`
3. Set environment variables:
   ```bash
   export ROCM_PATH=/opt/rocm
   export PATH=$ROCM_PATH/bin:$PATH
   export LD_LIBRARY_PATH=$ROCM_PATH/lib:$LD_LIBRARY_PATH
   ```

**Build Colossus with ROCm support:**
```bash
# Build llama.cpp with ROCm
cd third_party/llama.cpp
make LLAMA_HIPBLAS=1

# Build Colossus
cd ../..
CC=/opt/rocm/llvm/bin/clang CXX=/opt/rocm/llvm/bin/clang++ \
CGO_LDFLAGS="-O3 --hip-link --rtlib=compiler-rt -unwindlib=libgcc -lrocblas -lhipblas" \
go build
```

### Apple Metal

**Requirements:**
- Apple Silicon Mac (M1, M1 Pro, M1 Max, M2, etc.)
- macOS 12.0 or later

**Status:** Experimental support (not yet implemented)

## Configuration

### Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `COLOSSUS_INFERENCE_ENGINE` | Inference engine type | `simulated` | `llamacpp` |
| `COLOSSUS_GPU_LAYERS` | Number of layers to offload to GPU | `0` | `32` |
| `CUDA_VISIBLE_DEVICES` | CUDA devices to use | All | `0,1` |
| `ROCR_VISIBLE_DEVICES` | ROCm devices to use | All | `0` |
| `COLOSSUS_FORCE_LLAMACPP` | Force llama.cpp engine | `false` | `true` |

### Model-specific Configuration

Different models require different GPU layer configurations:

| Model Size | Recommended GPU Layers | Memory Usage |
|------------|----------------------|--------------|
| 3B parameters | 20-32 layers | 3-4 GB |
| 7B parameters | 32-40 layers | 6-8 GB |
| 13B parameters | 40-60 layers | 12-16 GB |
| 30B+ parameters | 60-80 layers | 24+ GB |

### Automatic Configuration

Colossus automatically detects GPU capabilities and configures optimal settings:

```bash
# Check current configuration
colossus gpu info

# Example output:
# GPU Acceleration Status: ✓ Available (cuda)
# GPU Type: cuda
# Device Count: 1
# Driver Version: 525.60.11
# 
# GPU Devices:
# ID  NAME                MEMORY    UTILIZATION  TEMPERATURE  AVAILABLE
# 0   NVIDIA RTX 4090     24.0 GB   15%          45°C         ✓
# 
# Recommended Configuration:
#   GPU Layers: 40 (for 7B model)
#   Environment: COLOSSUS_INFERENCE_ENGINE=llamacpp
#   Environment: COLOSSUS_GPU_LAYERS=40
```

## Performance Optimization

### Memory Management

**GPU Memory:**
- Monitor GPU memory usage: `nvidia-smi` or `rocm-smi`
- Reduce context size if running out of memory
- Use `COLOSSUS_GPU_LAYERS` to control memory usage

**System Memory:**
- Enable memory mapping: automatically enabled
- Disable memory lock for large models if needed

### Multi-GPU Setup

**NVIDIA Multi-GPU:**
```bash
export CUDA_VISIBLE_DEVICES=0,1
export COLOSSUS_GPU_LAYERS=60  # Split across GPUs
```

**Tensor Parallel (Future):**
```bash
# Will be supported in future versions
export COLOSSUS_TENSOR_PARALLEL=2
```

### Performance Tuning

**Context Size:**
- Larger context = more memory, slower inference
- Recommended: 2048-4096 for most use cases

**Batch Size:**
- Larger batch = better GPU utilization
- Recommended: 512-1024

**Thread Count:**
- For CPU fallback operations
- Recommended: Number of CPU cores

## Troubleshooting

### Common Issues

**CUDA not detected:**
```bash
# Check CUDA installation
nvidia-smi
nvcc --version

# Check environment
echo $CUDA_HOME
echo $PATH
```

**ROCm not detected:**
```bash
# Check ROCm installation
rocm-smi
hipcc --version

# Check environment
echo $ROCM_PATH
```

**Out of memory errors:**
```bash
# Reduce GPU layers
export COLOSSUS_GPU_LAYERS=16

# Or reduce context size
# (modify model options in config)
```

**Slow performance:**
```bash
# Check GPU utilization
nvidia-smi

# Increase batch size
# (modify model options in config)
```

### Debug Mode

Enable verbose logging to debug GPU issues:

```bash
colossus serve --verbose
```

Look for messages like:
```
INFO[0000] Detected 1 CUDA GPU(s)
INFO[0000] Configured CUDA acceleration with 32 GPU layers
INFO[0000] Loading model with llama.cpp: llama2-7b from ./models/llama2-7b.gguf
```

## Building from Source

### Prerequisites

**For CUDA support:**
```bash
# Ubuntu/Debian
sudo apt install nvidia-cuda-toolkit

# Or download from NVIDIA
wget https://developer.download.nvidia.com/compute/cuda/12.3.1/local_installers/cuda_12.3.1_545.23.08_linux.run
sudo sh cuda_12.3.1_545.23.08_linux.run
```

**For ROCm support:**
```bash
# Ubuntu 22.04
wget https://repo.radeon.com/amdgpu-install/22.40.5/ubuntu/jammy/amdgpu-install_5.4.50405-1_all.deb
sudo dpkg -i amdgpu-install_5.4.50405-1_all.deb
sudo amdgpu-install --usecase=rocm
```

### Build Steps

1. **Clone with submodules:**
   ```bash
   git clone --recurse-submodules https://github.com/your-org/colossus-cli.git
   cd colossus-cli
   ```

2. **Build llama.cpp dependencies:**
   ```bash
   # For CUDA
   cd third_party/llama.cpp
   make LLAMA_CUBLAS=1
   
   # For ROCm
   cd third_party/llama.cpp
   make LLAMA_HIPBLAS=1
   
   # For CPU only
   cd third_party/llama.cpp
   make
   ```

3. **Build Colossus:**
   ```bash
   cd ../..
   
   # With CUDA
   BUILD_TYPE=cublas make build
   
   # With ROCm  
   BUILD_TYPE=hipblas make build
   
   # CPU only
   make build
   ```

### Docker Support

**CUDA Docker:**
```dockerfile
FROM nvidia/cuda:12.3-devel-ubuntu22.04
# ... build steps
```

**ROCm Docker:**
```dockerfile
FROM rocm/dev-ubuntu-22.04:5.7
# ... build steps
```

## API Usage

### Check GPU Status via API

```bash
# Add GPU info endpoint
curl http://localhost:11434/api/gpu/info

# Response:
{
  "type": "cuda",
  "device_count": 1,
  "devices": [
    {
      "id": 0,
      "name": "NVIDIA RTX 4090",
      "memory_mb": 24576,
      "available": true
    }
  ],
  "available": true
}
```

### Model Loading with GPU Options

```bash
curl -X POST http://localhost:11434/api/models/load \
  -H "Content-Type: application/json" \
  -d '{
    "name": "llama2-7b",
    "options": {
      "gpu_layers": 32,
      "context_size": 4096,
      "use_cuda": true
    }
  }'
```

## Performance Benchmarks

Typical performance improvements with GPU acceleration:

| Model | Hardware | CPU Only | GPU Accelerated | Speedup |
|-------|----------|----------|----------------|---------|
| 7B | RTX 4090 | 5 tokens/s | 45 tokens/s | 9x |
| 7B | RTX 3080 | 5 tokens/s | 35 tokens/s | 7x |
| 13B | RTX 4090 | 3 tokens/s | 25 tokens/s | 8x |
| 30B | A100 80GB | 1 token/s | 15 tokens/s | 15x |

*Benchmarks may vary based on model, prompt length, and system configuration.*
