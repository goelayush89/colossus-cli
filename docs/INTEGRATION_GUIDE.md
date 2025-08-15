# Integration Guide: Real Inference with llama.cpp

This guide explains how to complete the integration with actual llama.cpp Go bindings for real model inference.

## Current Implementation Status

✅ **Completed:**
- Inference engine interface design
- GPU detection and configuration
- CUDA/ROCm support architecture
- API integration with inference engines
- Simulated engine for testing
- CLI commands for GPU management
- Automatic hardware optimization

🚧 **Remaining Work:**
- Complete llama.cpp Go bindings integration
- Replace simulated inference with real calls
- Model format validation and loading
- Memory management optimizations

## Integration Steps

### 1. Choose llama.cpp Go Bindings

**Recommended: go-skynet/go-llama.cpp**

```bash
# Add as Git submodule
git submodule add https://github.com/go-skynet/go-llama.cpp third_party/go-llama.cpp
cd third_party/go-llama.cpp
git submodule update --init --recursive
```

**Alternative: Create custom CGO bindings**

If you prefer direct control, create CGO bindings to llama.cpp:

```bash
# Clone llama.cpp
git submodule add https://github.com/ggerganov/llama.cpp third_party/llama.cpp
```

### 2. Update Build System

**Update Makefile:**

```makefile
# Add llama.cpp build targets
build-llamacpp:
	cd third_party/llama.cpp && make

build-llamacpp-cuda:
	cd third_party/llama.cpp && make LLAMA_CUBLAS=1

build-llamacpp-rocm:
	cd third_party/llama.cpp && make LLAMA_HIPBLAS=1

# Update main build to include llama.cpp
build: build-llamacpp
	CGO_CFLAGS="-I./third_party/llama.cpp" \
	CGO_LDFLAGS="-L./third_party/llama.cpp -lllama" \
	go build -o bin/colossus
```

**Update go.mod:**

```go
require (
    // ... existing dependencies
    github.com/go-skynet/go-llama.cpp v0.0.0-latest
)

replace github.com/go-skynet/go-llama.cpp => ./third_party/go-llama.cpp
```

### 3. Implement Real Inference

**Update internal/inference/llamacpp_engine.go:**

```go
package inference

import (
    // Add llama.cpp bindings
    llama "github.com/go-skynet/go-llama.cpp"
)

type LlamaCppModel struct {
    Name       string
    Path       string
    LoadedAt   time.Time
    Info       *ModelInfo
    Options    *ModelOptions
    
    // Real llama.cpp objects
    model   *llama.LLamaModel
    context *llama.LLamaContext
    mutex   sync.Mutex
}

func (e *LlamaCppEngine) LoadModel(name, path string, options *ModelOptions) error {
    // Real implementation
    modelParams := llama.DefaultModelParams()
    modelParams.NGpuLayers = int32(options.GPULayers)
    modelParams.UseMemoryMap = options.UseMemoryMap
    modelParams.UseMemoryLock = options.UseMemoryLock
    
    model := llama.LoadModelFromFile(path, modelParams)
    if model == nil {
        return fmt.Errorf("failed to load model from %s", path)
    }
    
    contextParams := llama.DefaultContextParams()
    contextParams.NCtx = int32(options.ContextSize)
    contextParams.NBatch = int32(options.BatchSize)
    contextParams.NThreads = int32(options.Threads)
    
    context := llama.NewContextWithModel(model, contextParams)
    if context == nil {
        model.Free()
        return fmt.Errorf("failed to create context for model %s", name)
    }
    
    // Store the loaded model
    e.models[name] = &LlamaCppModel{
        Name:     name,
        Path:     path,
        LoadedAt: time.Now(),
        model:    model,
        context:  context,
        // ... populate Info from actual model
    }
    
    return nil
}

func (e *LlamaCppEngine) Generate(req *types.GenerateRequest) (*types.GenerateResponse, error) {
    model, err := e.getModel(req.Model)
    if err != nil {
        return nil, err
    }
    
    model.mutex.Lock()
    defer model.mutex.Unlock()
    
    // Tokenize prompt
    tokens := model.context.Tokenize(req.Prompt, true)
    
    // Set generation parameters
    params := llama.DefaultGenerateParams()
    if req.Options != nil {
        params.Temperature = float32(req.Options.Temperature)
        params.TopP = float32(req.Options.TopP)
        params.TopK = int32(req.Options.TopK)
        // ... other parameters
    }
    
    // Generate response
    result := model.context.Generate(tokens, params)
    
    return &types.GenerateResponse{
        Model:     req.Model,
        CreatedAt: time.Now(),
        Response:  result,
        Done:      true,
    }, nil
}
```

### 4. Update Model Loading

**Update internal/model/manager.go:**

```go
func (m *Manager) PullModel(name string) error {
    // Add model format validation
    if !m.isValidModelFormat(modelPath) {
        return fmt.Errorf("unsupported model format")
    }
    
    // Download with progress reporting
    return m.downloadWithProgress(modelURL, modelPath)
}

func (m *Manager) isValidModelFormat(path string) bool {
    // Check for supported formats
    validExtensions := []string{".gguf", ".ggml", ".bin"}
    ext := strings.ToLower(filepath.Ext(path))
    
    for _, validExt := range validExtensions {
        if ext == validExt {
            return true
        }
    }
    return false
}
```

### 5. Add Model Registry Integration

**Create internal/registry/huggingface.go:**

```go
package registry

type HuggingFaceRegistry struct {
    BaseURL string
    Token   string
}

func (r *HuggingFaceRegistry) SearchModels(query string) ([]ModelInfo, error) {
    // Implement HF Hub API integration
}

func (r *HuggingFaceRegistry) DownloadModel(modelID string) error {
    // Implement model downloading from HF Hub
}
```

### 6. Testing Integration

**Create tests for real inference:**

```go
// internal/inference/llamacpp_test.go
func TestLlamaCppEngineLoad(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping llama.cpp integration test")
    }
    
    engine := NewLlamaCppEngine()
    
    // Test with a small model
    modelPath := "testdata/tiny-model.gguf"
    options := DefaultModelOptions()
    
    err := engine.LoadModel("test-model", modelPath, options)
    assert.NoError(t, err)
    
    // Test inference
    req := &types.GenerateRequest{
        Model:  "test-model",
        Prompt: "Hello world",
    }
    
    resp, err := engine.Generate(req)
    assert.NoError(t, err)
    assert.NotEmpty(t, resp.Response)
}
```

### 7. Performance Optimization

**Add memory pooling:**

```go
type MemoryPool struct {
    tokenBuffers sync.Pool
    responseBuffers sync.Pool
}

func (p *MemoryPool) GetTokenBuffer() []int32 {
    if buf := p.tokenBuffers.Get(); buf != nil {
        return buf.([]int32)[:0]
    }
    return make([]int32, 0, 512)
}
```

**Add batch processing:**

```go
func (e *LlamaCppEngine) GenerateBatch(requests []*types.GenerateRequest) ([]*types.GenerateResponse, error) {
    // Implement batch inference for better GPU utilization
}
```

## Build Instructions

### With CUDA Support

```bash
# Install CUDA dependencies
export CUDA_HOME=/usr/local/cuda
export PATH=$CUDA_HOME/bin:$PATH

# Build llama.cpp with CUDA
cd third_party/llama.cpp
make LLAMA_CUBLAS=1

# Build Colossus
cd ../..
CGO_CFLAGS="-I./third_party/llama.cpp -I$CUDA_HOME/include" \
CGO_LDFLAGS="-L./third_party/llama.cpp -L$CUDA_HOME/lib64 -lllama -lcublas -lcudart" \
go build -tags cuda -o colossus
```

### With ROCm Support

```bash
# Install ROCm dependencies
export ROCM_PATH=/opt/rocm
export PATH=$ROCM_PATH/bin:$PATH

# Build llama.cpp with ROCm
cd third_party/llama.cpp
CC=/opt/rocm/llvm/bin/clang CXX=/opt/rocm/llvm/bin/clang++ \
make LLAMA_HIPBLAS=1

# Build Colossus
cd ../..
CC=/opt/rocm/llvm/bin/clang CXX=/opt/rocm/llvm/bin/clang++ \
CGO_CFLAGS="-I./third_party/llama.cpp -I$ROCM_PATH/include" \
CGO_LDFLAGS="-L./third_party/llama.cpp -L$ROCM_PATH/lib -lllama -lhipblas -lrocblas" \
go build -tags rocm -o colossus
```

## Deployment

### Docker Integration

**Dockerfile.cuda:**
```dockerfile
FROM nvidia/cuda:12.3-devel-ubuntu22.04

# Install Go
RUN apt-get update && apt-get install -y golang-go git

# Copy source
COPY . /app
WORKDIR /app

# Build with CUDA support
RUN make build-llamacpp-cuda && make build

EXPOSE 11434
CMD ["./bin/colossus", "serve"]
```

**docker-compose.yml:**
```yaml
version: '3.8'
services:
  colossus:
    build:
      context: .
      dockerfile: Dockerfile.cuda
    ports:
      - "11434:11434"
    environment:
      - COLOSSUS_INFERENCE_ENGINE=llamacpp
      - CUDA_VISIBLE_DEVICES=0
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]
```

## Next Steps

1. **Choose Integration Approach**: Decide between go-skynet bindings or custom CGO
2. **Setup Build Environment**: Install CUDA/ROCm and configure build system
3. **Implement Core Functions**: Start with LoadModel and Generate methods
4. **Add Tests**: Create comprehensive test suite for real inference
5. **Optimize Performance**: Add memory pooling and batch processing
6. **Document**: Update API documentation and examples

## Resources

- [llama.cpp Documentation](https://github.com/ggerganov/llama.cpp)
- [go-skynet/go-llama.cpp](https://github.com/go-skynet/go-llama.cpp)
- [CUDA Installation Guide](https://docs.nvidia.com/cuda/cuda-installation-guide-linux/)
- [ROCm Installation Guide](https://rocmdocs.amd.com/en/latest/deploy/linux/quick_start.html)
- [CGO Documentation](https://pkg.go.dev/cmd/cgo)

This integration will provide production-ready inference capabilities with GPU acceleration support.
