package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"colossus-cli/internal/gpu"

	"github.com/spf13/cobra"
)

var gpuCmd = &cobra.Command{
	Use:   "gpu",
	Short: "GPU acceleration information and configuration",
	Long:  "Commands for checking GPU acceleration availability and configuration",
}

var gpuInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display GPU acceleration information",
	RunE:  runGPUInfo,
}

var gpuStatusCmd = &cobra.Command{
	Use:   "status", 
	Short: "Check GPU acceleration status",
	RunE:  runGPUStatus,
}

func init() {
	rootCmd.AddCommand(gpuCmd)
	gpuCmd.AddCommand(gpuInfoCmd)
	gpuCmd.AddCommand(gpuStatusCmd)
	
	// Add flags for output format
	gpuInfoCmd.Flags().Bool("json", false, "Output in JSON format")
	gpuStatusCmd.Flags().Bool("json", false, "Output in JSON format")
}

func runGPUInfo(cmd *cobra.Command, args []string) error {
	gpuInfo := gpu.DetectGPUs()
	
	jsonOutput, _ := cmd.Flags().GetBool("json")
	
	if jsonOutput {
		jsonData, err := json.MarshalIndent(gpuInfo, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal GPU info: %w", err)
		}
		fmt.Println(string(jsonData))
		return nil
	}
	
	// Human-readable output
	fmt.Printf("GPU Acceleration Status: ")
	if gpuInfo.Available {
		fmt.Printf("✓ Available (%s)\n", gpuInfo.Type)
	} else {
		fmt.Printf("✗ Not Available\n")
	}
	
	if gpuInfo.Available {
		fmt.Printf("GPU Type: %s\n", gpuInfo.Type)
		fmt.Printf("Device Count: %d\n", gpuInfo.DeviceCount)
		
		if gpuInfo.DriverVersion != "" {
			fmt.Printf("Driver Version: %s\n", gpuInfo.DriverVersion)
		}
		
		if len(gpuInfo.Devices) > 0 {
			fmt.Println("\nGPU Devices:")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tMEMORY\tUTILIZATION\tTEMPERATURE\tAVAILABLE")
			
			for _, device := range gpuInfo.Devices {
				memory := formatMemory(device.Memory)
				utilization := formatPercent(device.Utilization)
				temperature := formatTemperature(device.Temperature)
				available := formatBool(device.Available)
				
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
					device.ID, device.Name, memory, utilization, temperature, available)
			}
			
			w.Flush()
		}
		
		// Show optimal configuration
		optimalLayers := gpu.GetOptimalGPULayers(gpuInfo, 7000000000) // 7B model
		fmt.Printf("\nRecommended Configuration:\n")
		fmt.Printf("  GPU Layers: %d (for 7B model)\n", optimalLayers)
		fmt.Printf("  Environment: COLOSSUS_INFERENCE_ENGINE=llamacpp\n")
		fmt.Printf("  Environment: COLOSSUS_GPU_LAYERS=%d\n", optimalLayers)
	} else {
		fmt.Println("\nTo enable GPU acceleration:")
		fmt.Println("  1. Install CUDA Toolkit (NVIDIA) or ROCm (AMD)")
		fmt.Println("  2. Ensure drivers are properly installed")
		fmt.Println("  3. Set COLOSSUS_INFERENCE_ENGINE=llamacpp")
		fmt.Println("  4. Restart Colossus server")
	}
	
	return nil
}

func runGPUStatus(cmd *cobra.Command, args []string) error {
	gpuInfo := gpu.DetectGPUs()
	
	jsonOutput, _ := cmd.Flags().GetBool("json")
	
	if jsonOutput {
		status := map[string]interface{}{
			"gpu_available": gpuInfo.Available,
			"gpu_type":      gpuInfo.Type,
			"device_count":  gpuInfo.DeviceCount,
		}
		jsonData, _ := json.MarshalIndent(status, "", "  ")
		fmt.Println(string(jsonData))
		return nil
	}
	
	if gpuInfo.Available {
		fmt.Printf("GPU acceleration is AVAILABLE (%s with %d device(s))\n", 
			gpuInfo.Type, gpuInfo.DeviceCount)
		return nil
	} else {
		fmt.Println("GPU acceleration is NOT AVAILABLE")
		return nil
	}
}

// Helper functions for formatting

func formatMemory(memoryMB int64) string {
	if memoryMB == 0 {
		return "N/A"
	}
	
	if memoryMB >= 1024 {
		return fmt.Sprintf("%.1f GB", float64(memoryMB)/1024)
	}
	
	return fmt.Sprintf("%d MB", memoryMB)
}

func formatPercent(percent int) string {
	if percent < 0 {
		return "N/A"
	}
	return fmt.Sprintf("%d%%", percent)
}

func formatTemperature(temp int) string {
	if temp <= 0 {
		return "N/A"
	}
	return fmt.Sprintf("%d°C", temp)
}

func formatBool(b bool) string {
	if b {
		return "✓"
	}
	return "✗"
}
