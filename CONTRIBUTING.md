# Contributing to Colossus CLI

Thank you for your interest in contributing to Colossus CLI! This document provides guidelines and information for contributors.

## 🚀 Quick Start

1. **Fork the repository**
2. **Clone your fork**:
   ```bash
   git clone https://github.com/yourusername/colossus-cli.git
   cd colossus-cli
   ```
3. **Set up the development environment**:
   ```bash
   git submodule update --init --recursive
   go mod download
   ```
4. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## 🛠️ Development Setup

### Prerequisites
- Go 1.21 or later
- Git with submodule support
- C compiler (for llama.cpp integration):
  - Linux: `gcc` or `clang`
  - macOS: Xcode Command Line Tools
  - Windows: Visual Studio Build Tools

### Building
```bash
# Development build (simulated inference)
go build -o bin/colossus

# Production build (with llama.cpp)
make build-cpu           # CPU only
make build-cuda          # NVIDIA GPU support
make build-rocm          # AMD GPU support
```

### Running Tests
```bash
go test ./...
```

## 📝 Contributing Guidelines

### Code Style
- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and small

### Commit Messages
Use conventional commits format:
```
type(scope): description

[optional body]

[optional footer]
```

Examples:
- `feat(api): add streaming support for chat endpoint`
- `fix(model): handle GGUF file validation errors`
- `docs(readme): update installation instructions`

Types:
- `feat`: New features
- `fix`: Bug fixes
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance tasks

### Pull Request Process

1. **Update documentation** if needed
2. **Add tests** for new functionality
3. **Ensure all tests pass**: `go test ./...`
4. **Update CHANGELOG.md** if applicable
5. **Create detailed PR description**:
   - What changes were made
   - Why the changes were needed
   - How to test the changes
   - Any breaking changes

### Issue Guidelines

#### Bug Reports
- Use the bug report template
- Include steps to reproduce
- Provide system information
- Add error messages and logs

#### Feature Requests
- Use the feature request template
- Describe the use case
- Explain the expected behavior
- Consider alternative solutions

## 🏗️ Architecture Overview

```
colossus-cli/
├── cmd/                    # CLI commands (Cobra)
├── internal/
│   ├── api/               # REST API server (Gin)
│   ├── config/            # Configuration management
│   ├── inference/         # Inference engines
│   ├── llama/             # llama.cpp bindings
│   ├── model/             # Model management
│   ├── registry/          # Model registries (HF Hub)
│   └── types/             # Shared data types
├── docs/                  # Documentation
├── scripts/               # Build and utility scripts
└── third_party/           # External dependencies
    └── llama.cpp/         # Git submodule
```

### Key Components

1. **Inference Engine Interface**: Abstraction for different inference backends
2. **Model Manager**: Handles model downloads, validation, and storage
3. **API Server**: Ollama-compatible REST API
4. **Registry Integration**: Hugging Face Hub client
5. **GPU Detection**: Automatic hardware detection and optimization

## 🧪 Testing

### Unit Tests
```bash
go test ./internal/...
```

### Integration Tests
```bash
# Start test server
./bin/colossus serve --port 11435 &

# Run API tests
go test ./tests/integration/...
```

### Manual Testing
```bash
# Test model download
./bin/colossus models pull tinyllama

# Test chat
./bin/colossus chat tinyllama

# Test API
curl -X POST http://localhost:11434/api/generate \
  -H "Content-Type: application/json" \
  -d '{"model": "tinyllama", "prompt": "Hello"}'
```

## 📚 Documentation

### Code Documentation
- Add Go doc comments for public functions
- Include examples in complex functions
- Document configuration options

### User Documentation
- Update README.md for user-facing changes
- Add examples for new features
- Update build guides for new dependencies

## 🔧 Debugging

### Enable Debug Logging
```bash
export COLOSSUS_LOG_LEVEL=debug
./bin/colossus serve --verbose
```

### Common Issues

#### Build Errors
1. **CGO errors**: Ensure C compiler is installed
2. **Submodule errors**: Run `git submodule update --init --recursive`
3. **Dependency errors**: Run `go mod tidy`

#### Runtime Errors
1. **Model loading**: Check file permissions and formats
2. **API errors**: Verify server is running and ports are available
3. **GPU errors**: Check GPU drivers and CUDA/ROCm installation

## 🎯 Areas for Contribution

### High Priority
- [ ] Performance optimizations
- [ ] Additional model format support
- [ ] Better error handling and recovery
- [ ] Comprehensive testing

### Medium Priority
- [ ] WebUI interface
- [ ] Docker container optimization
- [ ] Additional inference backends
- [ ] Monitoring and metrics

### Documentation
- [ ] API documentation
- [ ] Architecture deep-dive
- [ ] Performance tuning guide
- [ ] Troubleshooting guide

## 📞 Getting Help

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Code Review**: All PRs receive thorough review

## 📄 License

By contributing to Colossus CLI, you agree that your contributions will be licensed under the same license as the project.

## 🙏 Recognition

Contributors are recognized in:
- CHANGELOG.md for significant contributions
- GitHub contributors page
- Release notes for major features

Thank you for contributing to Colossus CLI! 🚀
