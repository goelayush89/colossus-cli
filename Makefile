# Colossus CLI Makefile

.PHONY: build clean test run install lint help deps build-llamacpp build-llamacpp-cuda build-llamacpp-rocm

# Variables
BINARY_NAME=colossus
BUILD_DIR=bin
MAIN_PACKAGE=.
VERSION?=dev
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# Build type detection
BUILD_TYPE?=cpu
CUDA_PATH?=/usr/local/cuda
ROCM_PATH?=/opt/rocm
LLAMA_CPP_DIR=third_party/llama.cpp

# Default target
all: build

# Setup dependencies
deps:
	@echo "Setting up dependencies..."
	git submodule update --init --recursive
	@echo "Dependencies ready"

# Build llama.cpp (CPU only)
build-llamacpp:
	@echo "Building llama.cpp (CPU)..."
	@if [ ! -d "$(LLAMA_CPP_DIR)" ]; then \
		echo "Error: llama.cpp not found. Run 'make setup-llamacpp' first"; \
		exit 1; \
	fi
	cd $(LLAMA_CPP_DIR) && make clean && make
	@echo "llama.cpp (CPU) build complete"

# Build llama.cpp with CUDA support
build-llamacpp-cuda:
	@echo "Building llama.cpp with CUDA support..."
	@if [ ! -d "$(CUDA_PATH)" ]; then \
		echo "Error: CUDA not found at $(CUDA_PATH)"; \
		exit 1; \
	fi
	@if [ ! -d "$(LLAMA_CPP_DIR)" ]; then \
		echo "Error: llama.cpp not found. Run 'make setup-llamacpp' first"; \
		exit 1; \
	fi
	cd $(LLAMA_CPP_DIR) && make clean && make LLAMA_CUBLAS=1
	@echo "llama.cpp (CUDA) build complete"

# Build llama.cpp with ROCm support
build-llamacpp-rocm:
	@echo "Building llama.cpp with ROCm support..."
	@if [ ! -d "$(ROCM_PATH)" ]; then \
		echo "Error: ROCm not found at $(ROCM_PATH)"; \
		exit 1; \
	fi
	@if [ ! -d "$(LLAMA_CPP_DIR)" ]; then \
		echo "Error: llama.cpp not found. Run 'make setup-llamacpp' first"; \
		exit 1; \
	fi
	cd $(LLAMA_CPP_DIR) && \
	CC=$(ROCM_PATH)/llvm/bin/clang CXX=$(ROCM_PATH)/llvm/bin/clang++ \
	make clean && make LLAMA_HIPBLAS=1
	@echo "llama.cpp (ROCm) build complete"

# Setup llama.cpp submodule
setup-llamacpp:
	@echo "Setting up llama.cpp..."
	@if [ ! -d "$(LLAMA_CPP_DIR)" ]; then \
		git submodule add https://github.com/ggerganov/llama.cpp $(LLAMA_CPP_DIR); \
	fi
	git submodule update --init --recursive
	@echo "llama.cpp setup complete"

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(MAKE) build-$(BUILD_TYPE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build with CPU support
build-cpu:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

# Build with CUDA support
build-cuda:
	@if [ ! -f "$(LLAMA_CPP_DIR)/libllama.a" ]; then \
		echo "llama.cpp library not found. Run 'make build-llamacpp-cuda' first"; \
		exit 1; \
	fi
	CGO_CFLAGS="-I$(LLAMA_CPP_DIR) -I$(CUDA_PATH)/include -DGGML_USE_CUBLAS" \
	CGO_LDFLAGS="-L$(LLAMA_CPP_DIR) -L$(CUDA_PATH)/lib64 -lllama -lcublas -lcudart -lcurand -lcublasLt" \
	go build $(LDFLAGS) -tags cuda -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

# Build with ROCm support  
build-rocm:
	@if [ ! -f "$(LLAMA_CPP_DIR)/libllama.a" ]; then \
		echo "llama.cpp library not found. Run 'make build-llamacpp-rocm' first"; \
		exit 1; \
	fi
	CC=$(ROCM_PATH)/llvm/bin/clang CXX=$(ROCM_PATH)/llvm/bin/clang++ \
	CGO_CFLAGS="-I$(LLAMA_CPP_DIR) -I$(ROCM_PATH)/include -DGGML_USE_HIPBLAS" \
	CGO_LDFLAGS="-L$(LLAMA_CPP_DIR) -L$(ROCM_PATH)/lib -lllama -lhipblas -lrocblas -lamdhip64" \
	go build $(LDFLAGS) -tags rocm -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run the application in development mode
run:
	@echo "Running $(BINARY_NAME) in development mode..."
	go run $(MAIN_PACKAGE) serve --verbose

# Install the binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) $(MAIN_PACKAGE)

# Lint the code
lint:
	@echo "Running linter..."
	golangci-lint run

# Format the code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	# Linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	# macOS
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	# Windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
	@echo "Multi-platform build complete"

# Start the server
serve:
	@echo "Starting Colossus server..."
	$(BUILD_DIR)/$(BINARY_NAME) serve

# Chat with a model (requires model name)
chat:
	@echo "Starting chat session..."
	$(BUILD_DIR)/$(BINARY_NAME) chat $(MODEL)

# List models
models:
	@echo "Listing models..."
	$(BUILD_DIR)/$(BINARY_NAME) models list

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	go mod download
	go mod tidy
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi

# Quick demo setup
demo: build
	@echo "Setting up demo..."
	@echo "1. Starting server in background..."
	$(BUILD_DIR)/$(BINARY_NAME) serve &
	@sleep 2
	@echo "2. Server started on http://localhost:11434"
	@echo "3. You can now:"
	@echo "   - Pull a model: $(BUILD_DIR)/$(BINARY_NAME) models pull tinyllama"
	@echo "   - Start chat: $(BUILD_DIR)/$(BINARY_NAME) chat tinyllama"
	@echo "   - Test API: curl http://localhost:11434/api/tags"

# Help
help:
	@echo "Available commands:"
	@echo ""
	@echo "Building:"
	@echo "  build                - Build the binary (CPU only by default)"
	@echo "  build-cpu            - Build with CPU support only"
	@echo "  build-cuda           - Build with CUDA GPU support"
	@echo "  build-rocm           - Build with ROCm GPU support"
	@echo "  BUILD_TYPE=cuda make build - Build with specified type"
	@echo ""
	@echo "Dependencies:"
	@echo "  setup-llamacpp       - Setup llama.cpp submodule"
	@echo "  build-llamacpp       - Build llama.cpp (CPU)"
	@echo "  build-llamacpp-cuda  - Build llama.cpp with CUDA"
	@echo "  build-llamacpp-rocm  - Build llama.cpp with ROCm"
	@echo "  deps                 - Setup all dependencies"
	@echo ""
	@echo "Development:"
	@echo "  clean                - Clean build artifacts"
	@echo "  test                 - Run tests"
	@echo "  run                  - Run in development mode"
	@echo "  lint                 - Run linter"
	@echo "  fmt                  - Format code"
	@echo "  install              - Install to GOPATH/bin"
	@echo ""
	@echo "Utilities:"
	@echo "  build-all            - Build for multiple platforms"
	@echo "  serve                - Start the server (requires build)"
	@echo "  chat                 - Start chat session (requires MODEL=name)"
	@echo "  models               - List models"
	@echo "  dev-setup            - Setup development environment"
	@echo "  demo                 - Quick demo setup"
	@echo ""
	@echo "Examples:"
	@echo "  make setup-llamacpp && make build-llamacpp-cuda && make build-cuda"
	@echo "  BUILD_TYPE=rocm make build"
	@echo "  CUDA_PATH=/usr/local/cuda-12 make build-cuda"
