package inference

import "colossus-cli/internal/types"

// InferenceEngine defines the interface for model inference
type InferenceEngine interface {
	// LoadModel loads a model into memory
	LoadModel(name, path string, options *ModelOptions) error
	
	// UnloadModel removes a model from memory
	UnloadModel(name string) error
	
	// IsModelLoaded checks if a model is loaded
	IsModelLoaded(name string) bool
	
	// Generate generates text using a loaded model
	Generate(req *types.GenerateRequest) (*types.GenerateResponse, error)
	
	// GenerateStream generates text with streaming support
	GenerateStream(req *types.GenerateRequest, callback func(*types.GenerateResponse) error) error
	
	// Chat handles chat completion using a loaded model
	Chat(req *types.ChatRequest) (*types.ChatResponse, error)
	
	// ChatStream handles chat completion with streaming support
	ChatStream(req *types.ChatRequest, callback func(*types.ChatResponse) error) error
	
	// GetModelInfo returns information about a loaded model
	GetModelInfo(name string) (*ModelInfo, error)
	
	// Shutdown gracefully shuts down the inference engine
	Shutdown() error
}

// ModelOptions represents options for loading a model
type ModelOptions struct {
	// Context size
	ContextSize int `json:"context_size"`
	
	// GPU layers to offload
	GPULayers int `json:"gpu_layers"`
	
	// Number of threads
	Threads int `json:"threads"`
	
	// Batch size
	BatchSize int `json:"batch_size"`
	
	// Use memory mapping
	UseMemoryMap bool `json:"use_memory_map"`
	
	// Use memory lock
	UseMemoryLock bool `json:"use_memory_lock"`
	
	// Low VRAM mode
	LowVRAM bool `json:"low_vram"`
	
	// Tensor split for multi-GPU
	TensorSplit []float32 `json:"tensor_split"`
	
	// CUDA/ROCm specific options
	UseCUDA bool `json:"use_cuda"`
	UseROCm bool `json:"use_rocm"`
}

// ModelInfo represents information about a loaded model
type ModelInfo struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	ContextSize int    `json:"context_size"`
	VocabSize   int    `json:"vocab_size"`
	Parameters  int64  `json:"parameters"`
	GPULayers   int    `json:"gpu_layers"`
	MemoryUsed  int64  `json:"memory_used"`
}

// DefaultModelOptions returns default options for model loading
func DefaultModelOptions() *ModelOptions {
	return &ModelOptions{
		ContextSize:   2048,
		GPULayers:     0,
		Threads:       0, // Auto-detect
		BatchSize:     512,
		UseMemoryMap:  true,
		UseMemoryLock: false,
		LowVRAM:       false,
		UseCUDA:       false,
		UseROCm:       false,
	}
}
