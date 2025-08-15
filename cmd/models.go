package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"colossus-cli/internal/config"
	"colossus-cli/internal/model"

	"github.com/spf13/cobra"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Manage models",
	Long:  "Commands for managing language models",
}

var listModelsCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed models",
	RunE:  runListModels,
}

var pullModelCmd = &cobra.Command{
	Use:   "pull [MODEL_NAME]",
	Short: "Download a model",
	Args:  cobra.ExactArgs(1),
	RunE:  runPullModel,
}

var removeModelCmd = &cobra.Command{
	Use:   "rm [MODEL_NAME]",
	Short: "Remove a model",
	Args:  cobra.ExactArgs(1),
	RunE:  runRemoveModel,
}

func init() {
	rootCmd.AddCommand(modelsCmd)
	modelsCmd.AddCommand(listModelsCmd)
	modelsCmd.AddCommand(pullModelCmd)
	modelsCmd.AddCommand(removeModelCmd)
}

func runListModels(cmd *cobra.Command, args []string) error {
	cfg := config.Load()
	manager := model.NewManager(cfg.ModelsPath)
	
	models, err := manager.ListModels()
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	if len(models) == 0 {
		fmt.Println("No models found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSIZE\tMODIFIED")
	
	for _, model := range models {
		fmt.Fprintf(w, "%s\t%s\t%s\n", 
			model.Name, 
			formatSize(model.Size), 
			model.ModifiedAt.Format("2006-01-02 15:04:05"))
	}
	
	return w.Flush()
}

func runPullModel(cmd *cobra.Command, args []string) error {
	cfg := config.Load()
	manager := model.NewManager(cfg.ModelsPath)
	
	modelName := args[0]
	fmt.Printf("Pulling model '%s'...\n", modelName)
	
	// Create progress callback with visual progress bar
	progressCallback := func(progress model.DownloadProgress) error {
		showProgressBar(progress)
		return nil
	}
	
	if err := manager.PullModelWithProgress(modelName, progressCallback); err != nil {
		fmt.Println() // New line after progress bar
		return fmt.Errorf("failed to pull model: %w", err)
	}
	
	fmt.Println() // New line after progress bar
	fmt.Printf("✅ Successfully pulled model '%s'\n", modelName)
	return nil
}

func runRemoveModel(cmd *cobra.Command, args []string) error {
	cfg := config.Load()
	manager := model.NewManager(cfg.ModelsPath)
	
	modelName := args[0]
	
	if err := manager.RemoveModel(modelName); err != nil {
		return fmt.Errorf("failed to remove model: %w", err)
	}
	
	fmt.Printf("Successfully removed model '%s'\n", modelName)
	return nil
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// showProgressBar displays a visual progress bar for downloads
func showProgressBar(progress model.DownloadProgress) {
	// Calculate percentage
	percentage := progress.Percentage
	if percentage > 100 {
		percentage = 100
	}
	
	// Progress bar configuration
	barWidth := 40
	filledWidth := int(percentage * float64(barWidth) / 100)
	
	// Create progress bar
	bar := strings.Repeat("█", filledWidth) + strings.Repeat("░", barWidth-filledWidth)
	
	// Format download info
	downloaded := formatSize(progress.Downloaded)
	total := formatSize(progress.Total)
	speed := formatSpeed(progress.Speed)
	
	// Format ETA
	eta := formatDuration(progress.ETA)
	
	// Build progress line
	progressLine := fmt.Sprintf("\r📥 [%s] %.1f%% (%s/%s) %s ETA: %s", 
		bar, percentage, downloaded, total, speed, eta)
	
	// Clear the line and print progress
	fmt.Print("\033[2K") // Clear current line
	fmt.Print(progressLine)
}

// formatSpeed formats download speed in human-readable format
func formatSpeed(bytesPerSecond int64) string {
	if bytesPerSecond == 0 {
		return "0 B/s"
	}
	
	const unit = 1024
	if bytesPerSecond < unit {
		return fmt.Sprintf("%d B/s", bytesPerSecond)
	}
	
	div, exp := int64(unit), 0
	for n := bytesPerSecond / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB/s", float64(bytesPerSecond)/float64(div), "KMGTPE"[exp])
}

// formatDuration formats duration in human-readable format
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "calculating..."
	}
	
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
	
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh%dm", hours, minutes)
}
