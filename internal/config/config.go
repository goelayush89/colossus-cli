package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	ModelsPath string `mapstructure:"models_path"`
	Verbose    bool   `mapstructure:"verbose"`
}

// Load loads the configuration from various sources
func Load() *Config {
	// Set defaults
	viper.SetDefault("host", "127.0.0.1")
	viper.SetDefault("port", 11434)
	viper.SetDefault("verbose", false)
	
	// Set default models path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	defaultModelsPath := filepath.Join(homeDir, ".colossus", "models")
	viper.SetDefault("models_path", defaultModelsPath)
	
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		// If unmarshaling fails, use defaults
		cfg = Config{
			Host:       viper.GetString("host"),
			Port:       viper.GetInt("port"),
			ModelsPath: viper.GetString("models_path"),
			Verbose:    viper.GetBool("verbose"),
		}
	}
	
	// Ensure models directory exists
	if err := os.MkdirAll(cfg.ModelsPath, 0755); err != nil {
		// If we can't create the directory, use current directory
		cfg.ModelsPath = "./models"
		os.MkdirAll(cfg.ModelsPath, 0755)
	}
	
	return &cfg
}
