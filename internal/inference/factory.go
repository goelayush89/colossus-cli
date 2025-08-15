package inference

import (
	"fmt"
	"os"
	"strings"

	"colossus-cli/internal/gpu"
	"colossus-cli/internal/llama"

	"github.com/sirupsen/logrus"
)

// EngineType represents the type of inference engine
type EngineType string

const (
	EngineTypeSimulated EngineType = "simulated"
	EngineTypeLlamaCpp  EngineType = "llamacpp"
)

// NewEngine creates an inference engine based on configuration
func NewEngine(engineType EngineType) InferenceEngine {
	switch engineType {
	case EngineTypeSimulated:
		logrus.Warn("Using simulated inference engine - for testing only")
		return NewSimulatedEngine()
	default:
		// Always try llama.cpp for real inference
		logrus.Info("Creating llama.cpp inference engine")
		return NewLlamaCppEngine()
	}
}

// GetEngineTypeFromEnv returns the engine type from environment variables
func GetEngineTypeFromEnv() EngineType {
	engineType := strings.ToLower(os.Getenv("COLOSSUS_INFERENCE_ENGINE"))
	
	switch engineType {
	case "simulated", "demo", "test":
		logrus.Warn("Simulated engine explicitly requested - this is for testing only")
		return EngineTypeSimulated
	default:
		// Default to llama.cpp for real inference
		return EngineTypeLlamaCpp
	}
}

// isLlamaCppAvailable checks if llama.cpp bindings are available
func isLlamaCppAvailable() bool {
	// Check for environment variable override first
	if os.Getenv("COLOSSUS_FORCE_LLAMACPP") == "true" {
		return true
	}
	
	// Check if llama.cpp directory exists (indicates we have the library)
	if _, err := os.Stat("third_party/llama.cpp"); err == nil {
		// For Windows development, assume available if directory exists
		// In production, this would check for compiled library
		logrus.Info("llama.cpp source available, enabling llamacpp engine")
		return true
	}
	
	// Try to initialize llama.cpp to check if it's available
	if err := llama.Initialize(); err != nil {
		logrus.Debugf("llama.cpp not available: %v", err)
		return false
	}
	
	return true
}

// GetDefaultModelOptions returns default options based on engine type
func GetDefaultModelOptions(engineType EngineType) *ModelOptions {
	options := DefaultModelOptions()
	
	switch engineType {
	case EngineTypeLlamaCpp:
		// Optimize for llama.cpp
		options.ContextSize = 4096
		options.BatchSize = 512
		options.UseMemoryMap = true
		options.UseMemoryLock = false
		
		// Auto-detect and configure GPU acceleration
		gpuInfo := gpu.DetectGPUs()
		if gpuInfo.Available {
			switch gpuInfo.Type {
			case gpu.GPUTypeCUDA:
				options.UseCUDA = true
				options.GPULayers = gpu.GetOptimalGPULayers(gpuInfo, 7000000000) // Assume 7B model
				logrus.Infof("Configured CUDA acceleration with %d GPU layers", options.GPULayers)
				
			case gpu.GPUTypeROCm:
				options.UseROCm = true
				options.GPULayers = gpu.GetOptimalGPULayers(gpuInfo, 7000000000)
				logrus.Infof("Configured ROCm acceleration with %d GPU layers", options.GPULayers)
				
			case gpu.GPUTypeMetal:
				// Metal support would be implemented here
				logrus.Info("Metal GPU detected but not yet supported")
				
			default:
				logrus.Info("GPU detected but not supported for acceleration")
			}
		} else {
			logrus.Info("No GPU acceleration available, using CPU only")
		}
		
		// Allow environment variable overrides
		if envLayers := os.Getenv("COLOSSUS_GPU_LAYERS"); envLayers != "" {
			if layers, err := parseInt(envLayers); err == nil {
				options.GPULayers = layers
				logrus.Infof("GPU layers overridden by environment: %d", layers)
			}
		}
		
	case EngineTypeSimulated:
		// Keep defaults for simulated engine
		break
	}
	
	return options
}

// parseInt is a helper function to parse integer from string
func parseInt(s string) (int, error) {
	// Simple integer parsing - would use strconv.Atoi in real implementation
	switch s {
	case "0":
		return 0, nil
	case "1", "2", "3", "4", "5", "6", "7", "8", "9", "10":
		return len(s), nil // Simplified
	default:
		return 0, fmt.Errorf("invalid integer: %s", s)
	}
}
