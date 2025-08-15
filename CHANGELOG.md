# Changelog

All notable changes to Colossus CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- GitHub Pages deployment with binary downloads
- Cross-platform build automation via GitHub Actions
- Automatic release creation with checksums
- Bug report and feature request templates
- Contributing guidelines and architecture documentation

### Changed
- Enhanced factory pattern for inference engine selection
- Improved build scripts for Windows development

### Security
- Added checksum verification for release binaries

## [1.0.0] - 2025-08-16

### Added
- 🚀 **Initial Release** - Complete Ollama alternative in Go
- ⚡ **Real Inference Engine** - llama.cpp integration with CGO bindings
- 🤖 **Model Management** - Download, list, and manage models locally
- 🤗 **Hugging Face Integration** - Automatic model discovery and GGUF download
- 🔥 **GPU Acceleration** - CUDA, ROCm, and Metal support with auto-detection
- 🌐 **Ollama API Compatibility** - Full REST API compatibility
- 💬 **Interactive Chat** - Terminal-based chat interface
- 📊 **Progress Tracking** - Real-time download progress reporting
- 🛡️ **Model Validation** - Automatic format detection and validation
- 🎯 **Smart Defaults** - Intelligent configuration and GPU layer optimization

### Core Commands
- `colossus serve` - Start the API server
- `colossus models` - Model management (pull, list, remove)
- `colossus chat` - Interactive chat sessions
- `colossus gpu` - GPU information and status

### API Endpoints
- `POST /api/generate` - Text generation
- `POST /api/chat` - Chat completions
- `GET /api/tags` - List local models
- `POST /api/pull` - Download models
- `DELETE /api/delete` - Remove models

### Supported Platforms
- **Windows** (x64) - Development and production builds
- **macOS** (Intel & Apple Silicon) - Full support with Metal acceleration
- **Linux** (x64 & ARM64) - CUDA and ROCm support

### Inference Engines
- **Simulated Engine** - Fast development and testing
- **LlamaCpp Engine** - Real inference with llama.cpp
- **Factory Pattern** - Automatic engine selection based on environment

### Model Sources
- **Hugging Face Hub** - Automatic GGUF model discovery
- **Direct URLs** - Custom model downloads
- **Local Files** - Import existing models

### GPU Features
- **Automatic Detection** - NVIDIA, AMD, and Apple GPU detection
- **Optimal Configuration** - Smart GPU layer calculation
- **Fallback Support** - Graceful CPU fallback
- **Memory Management** - Efficient GPU memory usage

### Configuration
- **CLI Flags** - Command-line configuration
- **Environment Variables** - Flexible environment setup
- **Config Files** - YAML configuration support
- **Smart Defaults** - Zero-configuration experience

### Documentation
- **Build Guide** - Comprehensive building instructions
- **GPU Acceleration** - Detailed GPU setup guide
- **Integration Guide** - API integration examples
- **Deployment Guide** - Production deployment recommendations

## [0.1.0] - 2025-08-15

### Added
- Initial project structure
- Basic CLI framework with Cobra
- Simulated inference engine for development
- Model management foundation
- Configuration system with Viper
- Basic API server with Gin

### Developer Notes
- Established project architecture
- Set up development workflow
- Created build automation scripts
