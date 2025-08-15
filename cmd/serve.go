package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"colossus-cli/internal/api"
	"colossus-cli/internal/config"
	"colossus-cli/internal/model"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Colossus API server",
	Long:  "Start the HTTP API server to handle model inference requests",
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	// Initialize configuration
	cfg := config.Load()
	
	// Setup logging
	if viper.GetBool("verbose") {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// Initialize model manager
	modelManager := model.NewManager(cfg.ModelsPath)

	// Setup API server
	server := api.NewServer(cfg, modelManager)
	
	// Start server
	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	logrus.Infof("Starting Colossus server on %s", address)

	srv := &http.Server{
		Addr:    address,
		Handler: server.Router(),
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Fatalf("Server forced to shutdown: %v", err)
	}

	logrus.Info("Server exited")
	return nil
}
