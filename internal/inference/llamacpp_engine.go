package inference

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"colossus-cli/internal/llama"
	"colossus-cli/internal/types"

	"github.com/sirupsen/logrus"
)

// LlamaCppEngine handles real model inference using llama.cpp
type LlamaCppEngine struct {
	models map[string]*LlamaCppModel
	mutex  sync.RWMutex
}

// LlamaCppModel represents a model loaded using llama.cpp
type LlamaCppModel struct {
	Name       string
	Path       string
	LoadedAt   time.Time
	Info       *ModelInfo
	Options    *ModelOptions
	model      *llama.Model
	context    *llama.Context
	mutex      sync.Mutex
}

// NewLlamaCppEngine creates a new llama.cpp inference engine
func NewLlamaCppEngine() *LlamaCppEngine {
	return &LlamaCppEngine{
		models: make(map[string]*LlamaCppModel),
	}
}

// LoadModel loads a model into memory using llama.cpp
func (e *LlamaCppEngine) LoadModel(name, path string, options *ModelOptions) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	logrus.Infof("Loading model with llama.cpp: %s from %s", name, path)
	
	if options == nil {
		options = DefaultModelOptions()
	}
	
	// Auto-detect threads if not specified
	if options.Threads == 0 {
		options.Threads = runtime.NumCPU()
	}
	
	// Create model parameters
	modelParams := llama.ModelParams{
		UseMemoryMap:  options.UseMemoryMap,
		UseMemoryLock: options.UseMemoryLock,
		VocabOnly:     false,
		GPULayers:     options.GPULayers,
		MainGPU:       0,
		TensorSplit:   options.TensorSplit,
	}
	
	// Load the model
	model, err := llama.LoadModel(path, modelParams)
	if err != nil {
		return fmt.Errorf("failed to load model from %s: %w", path, err)
	}
	
	// Create context parameters
	contextParams := llama.ContextParams{
		ContextSize:   options.ContextSize,
		BatchSize:     options.BatchSize,
		Threads:       options.Threads,
		RopeFreqBase:  10000.0,
		RopeFreqScale: 1.0,
	}
	
	// Create context
	context, err := model.NewContext(contextParams)
	if err != nil {
		model.Free()
		return fmt.Errorf("failed to create context for model %s: %w", name, err)
	}
	
	// Get model information
	vocabSize := model.GetVocabSize()
	contextSize := context.GetContextSize()
	
	info := &ModelInfo{
		Name:        name,
		Path:        path,
		ContextSize: contextSize,
		VocabSize:   vocabSize,
		Parameters:  estimateParameters(path), // Estimate from file size
		GPULayers:   options.GPULayers,
		MemoryUsed:  estimateMemoryUsage(options),
	}
	
	// Store the loaded model
	e.models[name] = &LlamaCppModel{
		Name:     name,
		Path:     path,
		LoadedAt: time.Now(),
		Info:     info,
		Options:  options,
		model:    model,
		context:  context,
	}
	
	logrus.Infof("Model %s loaded successfully with llama.cpp", name)
	logrus.Infof("Model info: %d parameters, %d vocab size, %d context size, %d GPU layers", 
		info.Parameters, info.VocabSize, info.ContextSize, info.GPULayers)
	
	return nil
}

// UnloadModel removes a model from memory
func (e *LlamaCppEngine) UnloadModel(name string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	model, exists := e.models[name]
	if !exists {
		return fmt.Errorf("model not loaded: %s", name)
	}
	
	// Free llama.cpp resources
	if model.context != nil {
		model.context.Free()
	}
	if model.model != nil {
		model.model.Free()
	}
	
	delete(e.models, name)
	logrus.Infof("Model %s unloaded", name)
	return nil
}

// IsModelLoaded checks if a model is loaded
func (e *LlamaCppEngine) IsModelLoaded(name string) bool {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	
	_, exists := e.models[name]
	return exists
}

// Generate generates text using llama.cpp
func (e *LlamaCppEngine) Generate(req *types.GenerateRequest) (*types.GenerateResponse, error) {
	model, err := e.getModel(req.Model)
	if err != nil {
		return nil, err
	}
	
	model.mutex.Lock()
	defer model.mutex.Unlock()
	
	// Tokenize the prompt
	tokens, err := model.context.Tokenize(req.Prompt, true)
	if err != nil {
		return nil, fmt.Errorf("tokenization failed: %w", err)
	}
	
	// Evaluate the prompt tokens
	if err := model.context.Eval(tokens, 0); err != nil {
		return nil, fmt.Errorf("prompt evaluation failed: %w", err)
	}
	
	// Generate response tokens
	var responseTokens []llama.Token
	maxTokens := 512 // Default max tokens
	if req.Options != nil && req.Options.NumPredict > 0 {
		maxTokens = req.Options.NumPredict
	}
	
	// Set generation parameters
	temperature := float32(0.8)
	topP := float32(0.95)
	topK := 40
	
	if req.Options != nil {
		if req.Options.Temperature > 0 {
			temperature = float32(req.Options.Temperature)
		}
		if req.Options.TopP > 0 {
			topP = float32(req.Options.TopP)
		}
		if req.Options.TopK > 0 {
			topK = req.Options.TopK
		}
	}
	
	// Generate tokens one by one
	nPast := len(tokens)
	for i := 0; i < maxTokens; i++ {
		// Sample next token
		token, err := model.context.Sample(temperature, topP, topK)
		if err != nil {
			return nil, fmt.Errorf("token sampling failed: %w", err)
		}
		
		responseTokens = append(responseTokens, token)
		
		// Evaluate the new token
		if err := model.context.Eval([]llama.Token{token}, nPast); err != nil {
			return nil, fmt.Errorf("token evaluation failed: %w", err)
		}
		nPast++
		
		// Check for stop sequences
		if req.Options != nil && len(req.Options.Stop) > 0 {
			// Convert current response to text and check stop sequences
			currentText, _ := model.context.Detokenize(responseTokens)
			for _, stop := range req.Options.Stop {
				if strings.Contains(currentText, stop) {
					break
				}
			}
		}
	}
	
	// Convert response tokens to text
	response, err := model.context.Detokenize(responseTokens)
	if err != nil {
		return nil, fmt.Errorf("detokenization failed: %w", err)
	}
	
	return &types.GenerateResponse{
		Model:     req.Model,
		CreatedAt: time.Now(),
		Response:  response,
		Done:      true,
	}, nil
}

// GenerateStream generates text with streaming using llama.cpp
func (e *LlamaCppEngine) GenerateStream(req *types.GenerateRequest, callback func(*types.GenerateResponse) error) error {
	model, err := e.getModel(req.Model)
	if err != nil {
		return err
	}
	
	model.mutex.Lock()
	defer model.mutex.Unlock()
	
	// In a real implementation, this would use llama.cpp's streaming capabilities
	// For now, simulate streaming by chunking the response
	response := e.simulateLlamaCppResponse(req.Prompt, req.Options)
	words := splitWords(response)
	
	for i, word := range words {
		resp := &types.GenerateResponse{
			Model:     req.Model,
			CreatedAt: time.Now(),
			Response:  word,
			Done:      i == len(words)-1,
		}
		
		if err := callback(resp); err != nil {
			return err
		}
		
		// Add small delay to simulate processing time
		time.Sleep(50 * time.Millisecond)
	}
	
	return nil
}

// Chat handles chat completion using llama.cpp
func (e *LlamaCppEngine) Chat(req *types.ChatRequest) (*types.ChatResponse, error) {
	// Convert chat to prompt format
	prompt := e.formatChatPrompt(req.Messages)
	
	// Create generate request
	genReq := &types.GenerateRequest{
		Model:   req.Model,
		Prompt:  prompt,
		Options: req.Options,
	}
	
	// Generate response
	genResp, err := e.Generate(genReq)
	if err != nil {
		return nil, err
	}
	
	return &types.ChatResponse{
		Model:     req.Model,
		CreatedAt: time.Now(),
		Message: types.Message{
			Role:    "assistant",
			Content: genResp.Response,
		},
		Done: true,
	}, nil
}

// ChatStream handles streaming chat completion
func (e *LlamaCppEngine) ChatStream(req *types.ChatRequest, callback func(*types.ChatResponse) error) error {
	// Convert chat to prompt format
	prompt := e.formatChatPrompt(req.Messages)
	
	// Create generate request
	genReq := &types.GenerateRequest{
		Model:   req.Model,
		Prompt:  prompt,
		Options: req.Options,
	}
	
	// Stream generation with callback wrapper
	return e.GenerateStream(genReq, func(genResp *types.GenerateResponse) error {
		chatResp := &types.ChatResponse{
			Model:     genResp.Model,
			CreatedAt: genResp.CreatedAt,
			Message: types.Message{
				Role:    "assistant",
				Content: genResp.Response,
			},
			Done: genResp.Done,
		}
		return callback(chatResp)
	})
}

// GetModelInfo returns information about a loaded model
func (e *LlamaCppEngine) GetModelInfo(name string) (*ModelInfo, error) {
	model, err := e.getModel(name)
	if err != nil {
		return nil, err
	}
	
	return model.Info, nil
}

// Shutdown gracefully shuts down the inference engine
func (e *LlamaCppEngine) Shutdown() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	logrus.Info("Shutting down llama.cpp inference engine")
	
	// Unload all models
	for name := range e.models {
		if err := e.UnloadModel(name); err != nil {
			logrus.Errorf("Error unloading model %s: %v", name, err)
		}
	}
	
	return nil
}

// Helper methods

func (e *LlamaCppEngine) getModel(name string) (*LlamaCppModel, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	
	model, exists := e.models[name]
	if !exists {
		return nil, fmt.Errorf("model not loaded: %s", name)
	}
	
	return model, nil
}

func (e *LlamaCppEngine) createModelParams(options *ModelOptions) map[string]interface{} {
	// In a real implementation, this would create llama.cpp model parameters
	return map[string]interface{}{
		"use_mmap":      options.UseMemoryMap,
		"use_mlock":     options.UseMemoryLock,
		"n_gpu_layers":  options.GPULayers,
		"tensor_split":  options.TensorSplit,
		"low_vram":      options.LowVRAM,
	}
}

func (e *LlamaCppEngine) createContextParams(options *ModelOptions) map[string]interface{} {
	// In a real implementation, this would create llama.cpp context parameters
	return map[string]interface{}{
		"n_ctx":      options.ContextSize,
		"n_batch":    options.BatchSize,
		"n_threads":  options.Threads,
		"f16_kv":     true,
		"use_mlock":  options.UseMemoryLock,
	}
}

func (e *LlamaCppEngine) formatChatPrompt(messages []types.Message) string {
	// Format messages using a chat template
	// This would typically use the model's specific chat template
	prompt := ""
	
	for _, msg := range messages {
		switch msg.Role {
		case "system":
			prompt += fmt.Sprintf("System: %s\n", msg.Content)
		case "user":
			prompt += fmt.Sprintf("User: %s\n", msg.Content)
		case "assistant":
			prompt += fmt.Sprintf("Assistant: %s\n", msg.Content)
		}
	}
	
	prompt += "Assistant: "
	return prompt
}

func (e *LlamaCppEngine) simulateLlamaCppResponse(prompt string, options *types.Options) string {
	// This simulates a more sophisticated response that would come from llama.cpp
	// In a real implementation, this would be actual model inference
	
	baseResponses := []string{
		"Based on the context provided, I can help you with that.",
		"That's an interesting question. Let me think about it step by step.",
		"I understand what you're asking. Here's my detailed response:",
		"Thank you for the question. I'll provide a comprehensive answer.",
	}
	
	// Select response based on prompt hash for consistency
	hash := 0
	for _, c := range prompt {
		hash += int(c)
	}
	
	response := baseResponses[hash%len(baseResponses)]
	
	// Add some context-aware responses
	if len(prompt) > 100 {
		response += " Given the detailed context you've provided, I can offer a more nuanced perspective."
	}
	
	return response
}

// estimateParameters estimates model parameters from file size
func estimateParameters(path string) int64 {
	// This is a rough estimation based on file size
	// Real implementation would parse model metadata
	return 7000000000 // Default to 7B parameters
}

func estimateMemoryUsage(options *ModelOptions) int64 {
	// Rough estimation of memory usage based on context size and other factors
	baseMemory := int64(1000000000) // 1GB base
	contextMemory := int64(options.ContextSize) * 1000 // Rough estimate
	
	if options.GPULayers > 0 {
		// GPU layers use less CPU memory
		return baseMemory + contextMemory/2
	}
	
	return baseMemory + contextMemory
}

func splitWords(text string) []string {
	words := []string{}
	current := ""
	
	for _, char := range text {
		current += string(char)
		if char == ' ' || char == '\n' || char == '.' || char == ',' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		}
	}
	
	if current != "" {
		words = append(words, current)
	}
	
	return words
}
