package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// HuggingFaceRegistry handles interactions with Hugging Face Hub
type HuggingFaceRegistry struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

// ModelInfo represents model information from Hugging Face Hub
type ModelInfo struct {
	ID             string    `json:"id"`
	Author         string    `json:"author"`
	LastModified   time.Time `json:"lastModified"`
	Private        bool      `json:"private"`
	Downloads      int       `json:"downloads"`
	Likes          int       `json:"likes"`
	Tags           []string  `json:"tags"`
	ModelIndex     map[string]interface{} `json:"modelIndex"`
	Siblings       []FileInfo `json:"siblings"`
	GatedMode      string    `json:"gatedMode"`
	LibraryName    string    `json:"library_name"`
	PipelineTag    string    `json:"pipeline_tag"`
	CreatedAt      time.Time `json:"createdAt"`
}

// FileInfo represents file information in a model repository
type FileInfo struct {
	RFileName string `json:"rfilename"`
	Size      int64  `json:"size"`
	BlobID    string `json:"blob_id"`
	LfsOID    string `json:"lfs_oid,omitempty"`
}

// SearchResult represents search results from Hugging Face Hub
type SearchResult struct {
	Models      []ModelInfo `json:"models"`
	NumItems    int         `json:"numItemsOnPage"`
	NumPages    int         `json:"numPages"`
	PageIndex   int         `json:"pageIndex"`
	TotalItems  int         `json:"totalItems"`
}

// DownloadProgress represents download progress information
type DownloadProgress struct {
	ModelID      string
	FileName     string
	Downloaded   int64
	Total        int64
	Speed        int64 // bytes per second
	ETA          time.Duration
	Status       string
}

// ProgressCallback is called during downloads to report progress
type ProgressCallback func(progress DownloadProgress) error

// NewHuggingFaceRegistry creates a new Hugging Face registry client
func NewHuggingFaceRegistry(token string) *HuggingFaceRegistry {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &HuggingFaceRegistry{
		BaseURL: "https://huggingface.co",
		Token:   token,
		Client:  client,
	}
}

// SearchModels searches for models on Hugging Face Hub
func (r *HuggingFaceRegistry) SearchModels(query string, options SearchOptions) (*SearchResult, error) {
	// Build search URL
	searchURL := fmt.Sprintf("%s/api/models", r.BaseURL)
	
	params := url.Values{}
	if query != "" {
		params.Add("search", query)
	}
	if options.Filter != "" {
		params.Add("filter", options.Filter)
	}
	if options.Sort != "" {
		params.Add("sort", options.Sort)
	}
	if options.Direction != "" {
		params.Add("direction", options.Direction)
	}
	if options.Limit > 0 {
		params.Add("limit", strconv.Itoa(options.Limit))
	}
	
	// Add model type filters for LLMs
	params.Add("pipeline_tag", "text-generation")
	params.Add("library", "transformers")
	
	if len(params) > 0 {
		searchURL += "?" + params.Encode()
	}
	
	// Create request
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create search request: %w", err)
	}
	
	// Add authorization header if token is provided
	if r.Token != "" {
		req.Header.Set("Authorization", "Bearer "+r.Token)
	}
	
	// Make request
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	var models []ModelInfo
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}
	
	// Filter for GGUF models
	var filteredModels []ModelInfo
	for _, model := range models {
		if r.hasGGUFFiles(model) {
			filteredModels = append(filteredModels, model)
		}
	}
	
	return &SearchResult{
		Models:     filteredModels,
		NumItems:   len(filteredModels),
		TotalItems: len(filteredModels),
	}, nil
}

// GetModelInfo retrieves detailed information about a specific model
func (r *HuggingFaceRegistry) GetModelInfo(modelID string) (*ModelInfo, error) {
	url := fmt.Sprintf("%s/api/models/%s", r.BaseURL, modelID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	if r.Token != "" {
		req.Header.Set("Authorization", "Bearer "+r.Token)
	}
	
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("model not found or access denied: %s", modelID)
	}
	
	var model ModelInfo
	if err := json.NewDecoder(resp.Body).Decode(&model); err != nil {
		return nil, fmt.Errorf("failed to parse model info: %w", err)
	}
	
	return &model, nil
}

// ListGGUFFiles lists available GGUF files for a model
func (r *HuggingFaceRegistry) ListGGUFFiles(modelID string) ([]FileInfo, error) {
	model, err := r.GetModelInfo(modelID)
	if err != nil {
		return nil, err
	}
	
	var ggufFiles []FileInfo
	for _, file := range model.Siblings {
		if strings.HasSuffix(strings.ToLower(file.RFileName), ".gguf") {
			ggufFiles = append(ggufFiles, file)
		}
	}
	
	return ggufFiles, nil
}

// DownloadModel downloads a specific file from a model repository
func (r *HuggingFaceRegistry) DownloadModel(modelID, fileName, outputPath string, callback ProgressCallback) error {
	// Get file information
	files, err := r.ListGGUFFiles(modelID)
	if err != nil {
		return fmt.Errorf("failed to list model files: %w", err)
	}
	
	// Find the specific file
	var targetFile *FileInfo
	for _, file := range files {
		if file.RFileName == fileName {
			targetFile = &file
			break
		}
	}
	
	if targetFile == nil {
		return fmt.Errorf("file not found: %s", fileName)
	}
	
	// Build download URL
	downloadURL := fmt.Sprintf("%s/%s/resolve/main/%s", r.BaseURL, modelID, fileName)
	
	// Create request
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}
	
	if r.Token != "" {
		req.Header.Set("Authorization", "Bearer "+r.Token)
	}
	
	// Make request
	resp, err := r.Client.Do(req)
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}
	
	// Create output directory
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()
	
	// Download with progress reporting
	return r.downloadWithProgress(resp.Body, outFile, targetFile.Size, modelID, fileName, callback)
}

// DownloadBestGGUF downloads the best GGUF variant for a model
func (r *HuggingFaceRegistry) DownloadBestGGUF(modelID, outputPath string, callback ProgressCallback) (string, error) {
	files, err := r.ListGGUFFiles(modelID)
	if err != nil {
		return "", err
	}
	
	if len(files) == 0 {
		return "", fmt.Errorf("no GGUF files found for model %s", modelID)
	}
	
	// Select best file (prefer Q4_K_M quantization)
	bestFile := r.selectBestGGUF(files)
	
	// Determine output filename
	outputFile := filepath.Join(outputPath, bestFile.RFileName)
	
	logrus.Infof("Selected GGUF file: %s (%.1f MB)", bestFile.RFileName, float64(bestFile.Size)/(1024*1024))
	
	// Download the file
	err = r.DownloadModel(modelID, bestFile.RFileName, outputFile, callback)
	if err != nil {
		return "", err
	}
	
	return outputFile, nil
}

// Helper methods

func (r *HuggingFaceRegistry) hasGGUFFiles(model ModelInfo) bool {
	for _, file := range model.Siblings {
		if strings.HasSuffix(strings.ToLower(file.RFileName), ".gguf") {
			return true
		}
	}
	return false
}

func (r *HuggingFaceRegistry) selectBestGGUF(files []FileInfo) FileInfo {
	// Preference order: Q4_K_M > Q5_K_M > Q4_K_S > Q8_0 > others
	preferences := []string{
		"q4_k_m", "q5_k_m", "q4_k_s", "q8_0", "q4_0", "q5_0", "q6_k", "q2_k",
	}
	
	for _, pref := range preferences {
		for _, file := range files {
			if strings.Contains(strings.ToLower(file.RFileName), pref) {
				return file
			}
		}
	}
	
	// If no preferred quantization found, return the first file
	return files[0]
}

func (r *HuggingFaceRegistry) downloadWithProgress(reader io.Reader, writer io.Writer, totalSize int64, modelID, fileName string, callback ProgressCallback) error {
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
			if callback != nil && now.Sub(lastUpdate) >= time.Second {
				elapsed := now.Sub(startTime)
				speed := int64(float64(downloaded) / elapsed.Seconds())
				
				var eta time.Duration
				if speed > 0 && totalSize > 0 {
					remaining := totalSize - downloaded
					eta = time.Duration(float64(remaining)/float64(speed)) * time.Second
				}
				
				progress := DownloadProgress{
					ModelID:    modelID,
					FileName:   fileName,
					Downloaded: downloaded,
					Total:      totalSize,
					Speed:      speed,
					ETA:        eta,
					Status:     "downloading",
				}
				
				if err := callback(progress); err != nil {
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
	if callback != nil {
		progress := DownloadProgress{
			ModelID:    modelID,
			FileName:   fileName,
			Downloaded: downloaded,
			Total:      totalSize,
			Speed:      0,
			ETA:        0,
			Status:     "completed",
		}
		callback(progress)
	}
	
	return nil
}

// SearchOptions represents options for searching models
type SearchOptions struct {
	Filter    string // e.g., "text-generation"
	Sort      string // e.g., "downloads", "created", "updated"
	Direction string // "asc" or "desc"
	Limit     int    // max results to return
}
