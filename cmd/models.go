package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

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
	
	if err := manager.PullModel(modelName); err != nil {
		return fmt.Errorf("failed to pull model: %w", err)
	}
	
	fmt.Printf("Successfully pulled model '%s'\n", modelName)
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
