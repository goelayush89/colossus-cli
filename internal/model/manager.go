package model

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"colossus-cli/internal/registry"
	"colossus-cli/internal/types"

	"github.com/sirupsen/logrus"
)

// Manager handles model operations
type Manager struct {
	modelsPath string
	hfRegistry *registry.HuggingFaceRegistry
}

// ProgressCallback is called during downloads to report progress
type ProgressCallback func(progress DownloadProgress) error

// DownloadProgress represents download progress information
type DownloadProgress struct {
	ModelName    string
	FileName     string
	Downloaded   int64
	Total        int64
	Speed        int64 // bytes per second
	ETA          time.Duration
	Status       string
	Percentage   float64
}

// NewManager creates a new model manager
func NewManager(modelsPath string) *Manager {
	// Initialize Hugging Face registry
	hfToken := os.Getenv("HUGGINGFACE_TOKEN")
	hfRegistry := registry.NewHuggingFaceRegistry(hfToken)
	
	return &Manager{
		modelsPath: modelsPath,
		hfRegistry: hfRegistry,
	}
}

// ListModels returns a list of installed models
func (m *Manager) ListModels() ([]types.ModelInfo, error) {
	var models []types.ModelInfo
	
	err := filepath.Walk(m.modelsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Check for supported model formats
		if IsValidModelFormat(info.Name()) {
			relPath, _ := filepath.Rel(m.modelsPath, path)
			
			// Validate the model file
			modelInfo, err := ValidateModel(path)
			if err != nil {
				logrus.Warnf("Failed to validate model %s: %v", relPath, err)
			}
			
			model := types.ModelInfo{
				Name:       strings.TrimSuffix(relPath, filepath.Ext(relPath)),
				Size:       info.Size(),
				ModifiedAt: info.ModTime(),
			}
			
			// Add validation information if available
			if modelInfo != nil && modelInfo.Valid {
				model.Digest = fmt.Sprintf("%s-%s", modelInfo.Format.String(), modelInfo.Version)
			}
			
			models = append(models, model)
		}
		
		return nil
	})
	
	return models, err
}

// PullModel downloads a model from a registry or URL
func (m *Manager) PullModel(name string) error {
	return m.PullModelWithProgress(name, nil)
}

// PullModelWithProgress downloads a model with progress reporting
func (m *Manager) PullModelWithProgress(name string, progressCallback ProgressCallback) error {
	logrus.Infof("Pulling model: %s", name)
	
	// First, try to download from Hugging Face Hub
	if strings.Contains(name, "/") {
		// Model name contains "/" so it's likely a Hugging Face model ID
		return m.downloadFromHuggingFace(name, progressCallback)
	}
	
	// Try predefined model URLs
	modelURL := m.getModelURL(name)
	if modelURL != "" {
		modelPath := filepath.Join(m.modelsPath, name+".gguf")
		return m.downloadFileWithProgress(modelURL, modelPath, name, progressCallback)
	}
	
	// Try searching Hugging Face for the model
	searchResults, err := m.hfRegistry.SearchModels(name, registry.SearchOptions{
		Limit: 5,
		Sort:  "downloads",
		Direction: "desc",
	})
	if err != nil {
		return fmt.Errorf("failed to search for model: %w", err)
	}
	
	if len(searchResults.Models) == 0 {
		return fmt.Errorf("model not found: %s", name)
	}
	
	// Use the first (most downloaded) result
	bestMatch := searchResults.Models[0]
	logrus.Infof("Found model: %s (downloads: %d)", bestMatch.ID, bestMatch.Downloads)
	
	return m.downloadFromHuggingFace(bestMatch.ID, progressCallback)
}

// RemoveModel removes a model from local storage
func (m *Manager) RemoveModel(name string) error {
	// Find the model file
	modelPath := filepath.Join(m.modelsPath, name+".gguf")
	
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		// Try with .bin extension
		modelPath = filepath.Join(m.modelsPath, name+".bin")
		if _, err := os.Stat(modelPath); os.IsNotExist(err) {
			return fmt.Errorf("model not found: %s", name)
		}
	}
	
	return os.Remove(modelPath)
}

// GetModelPath returns the path to a model file
func (m *Manager) GetModelPath(name string) (string, error) {
	// Try different extensions
	extensions := []string{".gguf", ".bin"}
	
	for _, ext := range extensions {
		modelPath := filepath.Join(m.modelsPath, name+ext)
		if _, err := os.Stat(modelPath); err == nil {
			return modelPath, nil
		}
	}
	
	return "", fmt.Errorf("model not found: %s", name)
}

// getModelURL returns the download URL for a model
// This is a simplified implementation - in practice you'd have a registry
func (m *Manager) getModelURL(name string) string {
	// Map of model names to download URLs
	// This is a demo - in reality you'd query model registries
	models := map[string]string{
		"tinyllama": "https://huggingface.co/TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF/resolve/main/tinyllama-1.1b-chat-v1.0.q4_k_m.gguf",
		"phi-2":     "https://huggingface.co/microsoft/phi-2/resolve/main/pytorch_model.bin",
	}
	
	return models[name]
}

// downloadFromHuggingFace downloads a model from Hugging Face Hub
func (m *Manager) downloadFromHuggingFace(modelID string, progressCallback ProgressCallback) error {
	// Create model directory
	modelDir := filepath.Join(m.modelsPath, strings.ReplaceAll(modelID, "/", "_"))
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return fmt.Errorf("failed to create model directory: %w", err)
	}
	
	// Convert progress callback
	var hfCallback registry.ProgressCallback
	if progressCallback != nil {
		hfCallback = func(progress registry.DownloadProgress) error {
			localProgress := DownloadProgress{
				ModelName:  modelID,
				FileName:   progress.FileName,
				Downloaded: progress.Downloaded,
				Total:      progress.Total,
				Speed:      progress.Speed,
				ETA:        progress.ETA,
				Status:     progress.Status,
			}
			
			if progress.Total > 0 {
				localProgress.Percentage = float64(progress.Downloaded) / float64(progress.Total) * 100
			}
			
			return progressCallback(localProgress)
		}
	}
	
	// Download best GGUF variant
	modelPath, err := m.hfRegistry.DownloadBestGGUF(modelID, modelDir, hfCallback)
	if err != nil {
		return fmt.Errorf("failed to download from Hugging Face: %w", err)
	}
	
	// Validate the downloaded model
	validation, err := ValidateModel(modelPath)
	if err != nil {
		logrus.Warnf("Failed to validate downloaded model: %v", err)
	} else if !validation.Valid {
		return fmt.Errorf("downloaded model failed validation: %s", validation.Error)
	} else {
		logrus.Infof("Model validated successfully: %s %s", validation.Format, validation.Architecture)
	}
	
	logrus.Infof("Successfully downloaded model %s to %s", modelID, modelPath)
	return nil
}

// downloadFileWithProgress downloads a file with progress reporting
func (m *Manager) downloadFileWithProgress(url, filepath, modelName string, progressCallback ProgressCallback) error {
	logrus.Infof("Downloading from: %s", url)
	
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: %s", resp.Status)
	}
	
	// Get content length for progress tracking
	contentLength := resp.ContentLength
	
	// Download with progress reporting
	if progressCallback != nil && contentLength > 0 {
		return m.copyWithProgress(resp.Body, out, contentLength, modelName, filepath, progressCallback)
	}
	
	// Simple copy without progress
	_, err = io.Copy(out, resp.Body)
	return err
}

// copyWithProgress copies data with progress reporting
func (m *Manager) copyWithProgress(reader io.Reader, writer io.Writer, totalSize int64, modelName, fileName string, progressCallback ProgressCallback) error {
	buffer := make([]byte, 32*1024) // 32KB buffer
	var downloaded int64
	startTime := time.Now()
	lastUpdate := startTime
	
	for {
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("read error: %w", err)
		}
		
		if n > 0 {
			if _, writeErr := writer.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("write error: %w", writeErr)
			}
			
			downloaded += int64(n)
			
			// Report progress every second
			now := time.Now()
			if now.Sub(lastUpdate) >= time.Second {
				elapsed := now.Sub(startTime)
				speed := int64(float64(downloaded) / elapsed.Seconds())
				
				var eta time.Duration
				var percentage float64
				if speed > 0 && totalSize > 0 {
					remaining := totalSize - downloaded
					eta = time.Duration(float64(remaining)/float64(speed)) * time.Second
					percentage = float64(downloaded) / float64(totalSize) * 100
				}
				
				progress := DownloadProgress{
					ModelName:  modelName,
					FileName:   fileName,
					Downloaded: downloaded,
					Total:      totalSize,
					Speed:      speed,
					ETA:        eta,
					Status:     "downloading",
					Percentage: percentage,
				}
				
				if err := progressCallback(progress); err != nil {
					return fmt.Errorf("progress callback error: %w", err)
				}
				
				lastUpdate = now
			}
		}
		
		if err == io.EOF {
			break
		}
	}
	
	// Final progress update
	if progressCallback != nil {
		progress := DownloadProgress{
			ModelName:  modelName,
			FileName:   fileName,
			Downloaded: downloaded,
			Total:      totalSize,
			Speed:      0,
			ETA:        0,
			Status:     "completed",
			Percentage: 100,
		}
		progressCallback(progress)
	}
	
	return nil
}

// downloadFile downloads a file from a URL (legacy method)
func (m *Manager) downloadFile(url, filepath string) error {
	return m.downloadFileWithProgress(url, filepath, "unknown", nil)
}
