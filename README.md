# Colossus CLI

A powerful Go-based alternative to Ollama for running large language models locally.

## Features

- 🚀 **Fast & Lightweight**: Built in Go for optimal performance
- 🔌 **Compatible API**: Drop-in replacement for Ollama API endpoints  
- 🤖 **Model Management**: Download, list, and manage models easily
- 💬 **Interactive Chat**: Built-in chat interface for testing models
- 🌊 **Streaming Support**: Real-time streaming responses
- ⚙️ **Configurable**: Flexible configuration options
- 🎯 **Real Inference**: Support for llama.cpp with actual model loading
- 🚀 **GPU Acceleration**: CUDA and ROCm support for faster inference
- 🔍 **GPU Detection**: Automatic GPU detection and optimal configuration
- 🎚️ **Performance Tuning**: Automatic optimization based on hardware

## Quick Start

### Installation

```bash
# Clone the repository
git clone <your-repo-url>
cd colossus-cli

# Build the application
go build -o colossus
```

### Basic Usage

1. **Check GPU acceleration**:
```bash
./colossus gpu info
```

2. **Start the server**:
```bash
# With GPU acceleration (if available)
COLOSSUS_INFERENCE_ENGINE=llamacpp ./colossus serve

# Or CPU only
./colossus serve
```

3. **Pull a model** (in another terminal):
```bash
./colossus models pull tinyllama
```

4. **Start chatting**:
```bash
./colossus chat tinyllama
```

## API Endpoints

The Colossus API is compatible with Ollama's REST API:

### Chat Completions
```bash
POST /api/chat
{
  "model": "tinyllama",
  "messages": [
    {"role": "user", "content": "Hello!"}
  ],
  "stream": true
}
```

### Text Generation
```bash
POST /api/generate
{
  "model": "tinyllama",
  "prompt": "The capital of France is",
  "stream": false
}
```

### Model Management
```bash
# List models
GET /api/tags

# Pull a model
POST /api/pull
{"name": "tinyllama"}

# Delete a model
DELETE /api/delete
{"name": "tinyllama"}
```

## CLI Commands

### Server
```bash
# Start the API server
colossus serve --host 0.0.0.0 --port 11434

# Start with verbose logging
colossus serve --verbose
```

### Model Management
```bash
# List installed models
colossus models list

# Download a model
colossus models pull tinyllama

# Remove a model
colossus models rm tinyllama
```

### GPU Management
```bash
# Check GPU acceleration status
colossus gpu status

# Get detailed GPU information
colossus gpu info

# Get GPU info in JSON format
colossus gpu info --json
```

### Interactive Chat
```bash
# Start chat session
colossus chat tinyllama

# In chat, type '/bye' to exit
```

## Configuration

Colossus can be configured via:

1. **Command line flags**
2. **Environment variables**
3. **Config file** (`~/.colossus.yaml`)

### Example config file:
```yaml
host: "127.0.0.1"
port: 11434
models_path: "~/.colossus/models"
verbose: false
```

### Environment Variables:
```bash
# Server configuration
export COLOSSUS_HOST=0.0.0.0
export COLOSSUS_PORT=11434
export COLOSSUS_MODELS_PATH=/path/to/models

# Inference engine configuration
export COLOSSUS_INFERENCE_ENGINE=llamacpp  # or 'simulated'
export COLOSSUS_GPU_LAYERS=32              # Number of layers to offload to GPU
export COLOSSUS_FORCE_LLAMACPP=true        # Force llama.cpp even if not detected

# GPU configuration (auto-detected, but can be overridden)
export CUDA_VISIBLE_DEVICES=0,1            # NVIDIA GPUs to use
export ROCR_VISIBLE_DEVICES=0              # AMD GPUs to use
```

## Development

### Building from Source
```bash
# Install dependencies
go mod download

# Build
go build -o colossus

# Run tests
go test ./...
```

### Architecture

- **CLI Layer**: Cobra-based command interface
- **API Layer**: Gin-based REST API server
- **Model Manager**: Handles model downloading and storage
- **Inference Engine**: Manages model loading and text generation
- **Configuration**: Viper-based configuration management

## Extending Colossus

### Adding New Model Sources
To add support for new model registries, modify `internal/model/manager.go`:

```go
func (m *Manager) getModelURL(name string) string {
    // Add your custom model registry logic here
}
```

### Custom Inference Backends
To integrate with different inference engines (like llama.cpp, ONNX, etc.), 
implement the interface in `internal/inference/engine.go`.

## Compatibility
- ✅ Chat completions API
- ✅ Text generation API  
- ✅ Model management API
- ✅ Streaming responses
- ✅ Model pulling/pushing
- 🚧 Custom model formats (in progress)

