package inference

import (
	"fmt"
	"strings"
	"time"

	"colossus-cli/internal/types"

	"github.com/sirupsen/logrus"
)

// SimulatedEngine handles simulated model inference (for demo/testing)
type SimulatedEngine struct {
	models map[string]*LoadedModel
}

// LoadedModel represents a model loaded in memory
type LoadedModel struct {
	Name       string
	Path       string
	LoadedAt   time.Time
	Info       *ModelInfo
	// In a real implementation, this would contain the actual model data
	// For this demo, we'll simulate responses
}

// NewSimulatedEngine creates a new simulated inference engine
func NewSimulatedEngine() *SimulatedEngine {
	return &SimulatedEngine{
		models: make(map[string]*LoadedModel),
	}
}

// LoadModel loads a model into memory with options
func (e *SimulatedEngine) LoadModel(name, path string, options *ModelOptions) error {
	logrus.Infof("Loading model: %s from %s", name, path)
	
	if options == nil {
		options = DefaultModelOptions()
	}
	
	// For demo purposes, we simulate loading
	e.models[name] = &LoadedModel{
		Name:     name,
		Path:     path,
		LoadedAt: time.Now(),
		Info: &ModelInfo{
			Name:        name,
			Path:        path,
			ContextSize: options.ContextSize,
			VocabSize:   32000, // Simulated
			Parameters:  7000000000, // 7B parameters
			GPULayers:   options.GPULayers,
			MemoryUsed:  4000000000, // 4GB simulated
		},
	}
	
	logrus.Infof("Model %s loaded successfully", name)
	return nil
}

// UnloadModel removes a model from memory
func (e *SimulatedEngine) UnloadModel(name string) error {
	if _, exists := e.models[name]; !exists {
		return fmt.Errorf("model not loaded: %s", name)
	}
	
	delete(e.models, name)
	logrus.Infof("Model %s unloaded", name)
	return nil
}

// IsModelLoaded checks if a model is loaded
func (e *SimulatedEngine) IsModelLoaded(name string) bool {
	_, exists := e.models[name]
	return exists
}

// Generate generates text using a loaded model
func (e *SimulatedEngine) Generate(req *types.GenerateRequest) (*types.GenerateResponse, error) {
	if !e.IsModelLoaded(req.Model) {
		return nil, fmt.Errorf("model not loaded: %s", req.Model)
	}
	
	// For demo purposes, we simulate a response
	response := simulateResponse(req.Prompt)
	
	return &types.GenerateResponse{
		Model:     req.Model,
		CreatedAt: time.Now(),
		Response:  response,
		Done:      true,
	}, nil
}

// Chat handles chat completion using a loaded model
func (e *SimulatedEngine) Chat(req *types.ChatRequest) (*types.ChatResponse, error) {
	if !e.IsModelLoaded(req.Model) {
		return nil, fmt.Errorf("model not loaded: %s", req.Model)
	}
	
	// Convert chat messages to prompt
	prompt := e.formatChatPrompt(req.Messages)
	
	// Generate response
	response := simulateResponse(prompt)
	
	return &types.ChatResponse{
		Model:     req.Model,
		CreatedAt: time.Now(),
		Message: types.Message{
			Role:    "assistant",
			Content: response,
		},
		Done: true,
	}, nil
}

// simulateResponse generates a simulated response (for demo purposes)
func simulateResponse(prompt string) string {
	// Enhanced simulation with more realistic responses
	// Note: This is simulated inference - for real inference, compile with llama.cpp
	
	lowerPrompt := strings.ToLower(prompt)
	
	// Greeting responses
	if strings.Contains(lowerPrompt, "hi") || strings.Contains(lowerPrompt, "hello") {
		return "Hello! I'm a simulated AI assistant. How can I help you today? (Note: This is simulated - for real inference, enable llama.cpp)"
	}
	
	// Question responses
	if strings.Contains(lowerPrompt, "how are you") {
		return "I'm doing well, thank you for asking! I'm currently running in simulation mode. What can I help you with?"
	}
	
	if strings.Contains(lowerPrompt, "what") && strings.Contains(lowerPrompt, "name") {
		return "I'm Colossus CLI's simulated assistant. In real mode, I would be the qwen model you downloaded."
	}
	
	if strings.Contains(lowerPrompt, "tell me about") {
		return "I'd be happy to help! However, I'm currently running in simulation mode. For real AI responses, please compile Colossus with llama.cpp support."
	}
	
	// Programming/coding questions
	if strings.Contains(lowerPrompt, "code") || strings.Contains(lowerPrompt, "programming") {
		return "I can help with coding questions! Though I'm in simulation mode right now, the real qwen model you downloaded is great for programming tasks."
	}
	
	// Math questions
	if strings.Contains(lowerPrompt, "math") || strings.Contains(lowerPrompt, "calculate") {
		return "I can help with math problems! In real inference mode, I'd be able to solve complex calculations for you."
	}
	
	// Help/assistance
	if strings.Contains(lowerPrompt, "help") {
		return "I'm here to help! Currently running in simulation mode. To get real AI responses, compile Colossus with llama.cpp integration."
	}
	
	// Goodbye
	if strings.Contains(lowerPrompt, "bye") || strings.Contains(lowerPrompt, "goodbye") {
		return "Goodbye! Thanks for trying Colossus CLI. Enable llama.cpp for real AI conversations!"
	}
	
	// Default responses based on prompt characteristics
	if len(prompt) > 100 {
		return "That's quite a detailed message! I'm currently in simulation mode, but the real qwen model would provide a thoughtful response to your query."
	}
	
	if strings.Contains(lowerPrompt, "?") {
		return "That's a great question! In real inference mode, I'd analyze your question carefully and provide a detailed answer."
	}
	
	// Fallback responses that acknowledge simulation
	fallbacks := []string{
		"I hear you! I'm currently running in simulation mode. For real AI responses, enable llama.cpp integration.",
		"Interesting! While I'm simulating responses right now, the qwen model you downloaded would handle this well in real mode.",
		"Thanks for your message! I'm in simulation mode - the real qwen model would provide much better responses.",
		"I understand! Currently simulating responses, but your downloaded qwen model is ready for real inference.",
	}
	
	// Simple hash for consistent responses
	hash := 0
	for _, c := range prompt {
		hash += int(c)
	}
	
	return fallbacks[hash%len(fallbacks)]
}

// GenerateStream generates text with streaming support
func (e *SimulatedEngine) GenerateStream(req *types.GenerateRequest, callback func(*types.GenerateResponse) error) error {
	if !e.IsModelLoaded(req.Model) {
		return fmt.Errorf("model not loaded: %s", req.Model)
	}
	
	response := simulateResponse(req.Prompt)
	words := splitIntoWords(response)
	
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
	}
	
	return nil
}

// ChatStream handles chat completion with streaming support
func (e *SimulatedEngine) ChatStream(req *types.ChatRequest, callback func(*types.ChatResponse) error) error {
	if !e.IsModelLoaded(req.Model) {
		return fmt.Errorf("model not loaded: %s", req.Model)
	}
	
	prompt := e.formatChatPrompt(req.Messages)
	response := simulateResponse(prompt)
	words := splitIntoWords(response)
	
	for i, word := range words {
		resp := &types.ChatResponse{
			Model:     req.Model,
			CreatedAt: time.Now(),
			Message: types.Message{
				Role:    "assistant",
				Content: word,
			},
			Done: i == len(words)-1,
		}
		
		if err := callback(resp); err != nil {
			return err
		}
	}
	
	return nil
}

// GetModelInfo returns information about a loaded model
func (e *SimulatedEngine) GetModelInfo(name string) (*ModelInfo, error) {
	model, exists := e.models[name]
	if !exists {
		return nil, fmt.Errorf("model not loaded: %s", name)
	}
	
	return model.Info, nil
}

// Shutdown gracefully shuts down the inference engine
func (e *SimulatedEngine) Shutdown() error {
	logrus.Info("Shutting down simulated inference engine")
	
	// Unload all models
	for name := range e.models {
		e.UnloadModel(name)
	}
	
	return nil
}

// splitIntoWords splits text into words for streaming simulation
func splitIntoWords(text string) []string {
	words := []string{}
	current := ""
	
	for _, char := range text {
		current += string(char)
		if char == ' ' || char == '\n' {
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

// formatChatPrompt converts chat messages to a single prompt
func (e *SimulatedEngine) formatChatPrompt(messages []types.Message) string {
	var parts []string
	
	for _, msg := range messages {
		switch msg.Role {
		case "user":
			parts = append(parts, "User: "+msg.Content)
		case "assistant":
			parts = append(parts, "Assistant: "+msg.Content)
		case "system":
			parts = append(parts, "System: "+msg.Content)
		}
	}
	
	return strings.Join(parts, "\n")
}
