package gpu

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// GPUInfo represents information about available GPUs
type GPUInfo struct {
	Type         GPUType `json:"type"`
	DeviceCount  int     `json:"device_count"`
	Devices      []GPU   `json:"devices"`
	DriverVersion string `json:"driver_version"`
	Available    bool    `json:"available"`
}

// GPUType represents the type of GPU acceleration
type GPUType string

const (
	GPUTypeNone   GPUType = "none"
	GPUTypeCUDA   GPUType = "cuda"
	GPUTypeROCm   GPUType = "rocm"
	GPUTypeMetal  GPUType = "metal"
	GPUTypeOpenCL GPUType = "opencl"
)

// GPU represents a single GPU device
type GPU struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Memory       int64  `json:"memory_mb"`
	Utilization  int    `json:"utilization_percent"`
	Temperature  int    `json:"temperature_c"`
	Available    bool   `json:"available"`
}

// DetectGPUs detects available GPU acceleration options
func DetectGPUs() *GPUInfo {
	info := &GPUInfo{
		Type:        GPUTypeNone,
		DeviceCount: 0,
		Devices:     []GPU{},
		Available:   false,
	}

	// Check CUDA first (NVIDIA)
	if cudaInfo := detectCUDA(); cudaInfo.Available {
		*info = *cudaInfo
		return info
	}

	// Check ROCm (AMD)
	if rocmInfo := detectROCm(); rocmInfo.Available {
		*info = *rocmInfo
		return info
	}

	// Check Metal (Apple Silicon)
	if runtime.GOOS == "darwin" {
		if metalInfo := detectMetal(); metalInfo.Available {
			*info = *metalInfo
			return info
		}
	}

	// Check OpenCL (fallback)
	if openclInfo := detectOpenCL(); openclInfo.Available {
		*info = *openclInfo
		return info
	}

	logrus.Info("No GPU acceleration detected, using CPU only")
	return info
}

// detectCUDA detects NVIDIA CUDA support
func detectCUDA() *GPUInfo {
	info := &GPUInfo{
		Type:      GPUTypeCUDA,
		Available: false,
	}

	// Check for CUDA environment variables
	cudaPath := os.Getenv("CUDA_PATH")
	cudaHome := os.Getenv("CUDA_HOME")
	cudaVisible := os.Getenv("CUDA_VISIBLE_DEVICES")

	if cudaPath == "" && cudaHome == "" {
		return info
	}

	// Try to run nvidia-smi to get GPU information
	cmd := exec.Command("nvidia-smi", "--query-gpu=index,name,memory.total,utilization.gpu,temperature.gpu", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		logrus.Debugf("nvidia-smi not available: %v", err)
		return info
	}

	// Parse nvidia-smi output
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, ", ")
		if len(parts) >= 5 {
			id, _ := strconv.Atoi(parts[0])
			name := parts[1]
			memory, _ := strconv.ParseInt(parts[2], 10, 64)
			utilization, _ := strconv.Atoi(parts[3])
			temperature, _ := strconv.Atoi(parts[4])

			// Check if device is visible
			deviceAvailable := true
			if cudaVisible != "" && !strings.Contains(cudaVisible, parts[0]) {
				deviceAvailable = false
			}

			info.Devices = append(info.Devices, GPU{
				ID:          id,
				Name:        name,
				Memory:      memory,
				Utilization: utilization,
				Temperature: temperature,
				Available:   deviceAvailable,
			})
		}
	}

	if len(info.Devices) > 0 {
		info.Available = true
		info.DeviceCount = len(info.Devices)

		// Get CUDA driver version
		cmd = exec.Command("nvidia-smi", "--query-gpu=driver_version", "--format=csv,noheader,nounits")
		if output, err := cmd.Output(); err == nil {
			info.DriverVersion = strings.TrimSpace(string(output))
		}

		logrus.Infof("Detected %d CUDA GPU(s)", info.DeviceCount)
	}

	return info
}

// detectROCm detects AMD ROCm support
func detectROCm() *GPUInfo {
	info := &GPUInfo{
		Type:      GPUTypeROCm,
		Available: false,
	}

	// Check for ROCm environment variables
	rocmPath := os.Getenv("ROCM_PATH")
	rocmVisible := os.Getenv("ROCR_VISIBLE_DEVICES")

	if rocmPath == "" {
		rocmPath = "/opt/rocm"
	}

	// Try to run rocm-smi to get GPU information
	cmd := exec.Command("rocm-smi", "--showid", "--showproductname", "--showmeminfo", "vram", "--showuse", "--showtemp")
	output, err := cmd.Output()
	if err != nil {
		logrus.Debugf("rocm-smi not available: %v", err)
		return info
	}

	// Parse rocm-smi output (simplified parsing)
	lines := strings.Split(string(output), "\n")
	deviceID := 0
	
	for _, line := range lines {
		if strings.Contains(line, "GPU") && strings.Contains(line, "ID") {
			// This is a simplified parser - real implementation would be more robust
			info.Devices = append(info.Devices, GPU{
				ID:        deviceID,
				Name:      "AMD GPU", // Would parse actual name
				Memory:    8192,      // Would parse actual memory
				Available: rocmVisible == "" || strings.Contains(rocmVisible, strconv.Itoa(deviceID)),
			})
			deviceID++
		}
	}

	if len(info.Devices) > 0 {
		info.Available = true
		info.DeviceCount = len(info.Devices)
		logrus.Infof("Detected %d ROCm GPU(s)", info.DeviceCount)
	}

	return info
}

// detectMetal detects Apple Metal support
func detectMetal() *GPUInfo {
	info := &GPUInfo{
		Type:      GPUTypeMetal,
		Available: false,
	}

	if runtime.GOOS != "darwin" {
		return info
	}

	// Check if we're on Apple Silicon
	cmd := exec.Command("sysctl", "-n", "hw.optional.arm64")
	output, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(output)) == "1" {
		// On Apple Silicon, Metal is available
		info.Available = true
		info.DeviceCount = 1
		info.Devices = append(info.Devices, GPU{
			ID:        0,
			Name:      "Apple GPU",
			Available: true,
		})
		logrus.Info("Detected Apple Metal GPU support")
	}

	return info
}

// detectOpenCL detects OpenCL support
func detectOpenCL() *GPUInfo {
	info := &GPUInfo{
		Type:      GPUTypeOpenCL,
		Available: false,
	}

	// Try to detect OpenCL devices
	// This is a simplified check - real implementation would use OpenCL libraries
	
	return info
}

// GetOptimalGPULayers returns the optimal number of layers to offload to GPU
func GetOptimalGPULayers(gpuInfo *GPUInfo, modelSize int64) int {
	if !gpuInfo.Available || len(gpuInfo.Devices) == 0 {
		return 0
	}

	// Calculate based on available GPU memory
	totalGPUMemory := int64(0)
	for _, device := range gpuInfo.Devices {
		if device.Available {
			totalGPUMemory += device.Memory * 1024 * 1024 // Convert MB to bytes
		}
	}

	// Rough estimation: each layer needs about 100MB for a 7B model
	layerMemory := int64(100 * 1024 * 1024)
	if modelSize > 7000000000 { // 13B+ models
		layerMemory = 200 * 1024 * 1024
	}

	// Leave 2GB for context and other GPU operations
	availableMemory := totalGPUMemory - (2 * 1024 * 1024 * 1024)
	if availableMemory <= 0 {
		return 0
	}

	maxLayers := int(availableMemory / layerMemory)
	
	// Cap at reasonable limits based on model type
	switch {
	case modelSize <= 3000000000: // Small models (3B)
		if maxLayers > 32 {
			maxLayers = 32
		}
	case modelSize <= 7000000000: // Medium models (7B)
		if maxLayers > 40 {
			maxLayers = 40
		}
	default: // Large models (13B+)
		if maxLayers > 80 {
			maxLayers = 80
		}
	}

	logrus.Infof("Optimal GPU layers: %d (GPU memory: %.1f GB)", maxLayers, float64(totalGPUMemory)/(1024*1024*1024))
	return maxLayers
}

// IsGPUAccelerationAvailable returns true if any GPU acceleration is available
func IsGPUAccelerationAvailable() bool {
	info := DetectGPUs()
	return info.Available
}
