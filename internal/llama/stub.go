// +build !cgo

package llama

import (
	"fmt"
	"sync"
)

// Stub implementations for builds without CGO/llama.cpp

// Backend represents the llama.cpp backend (stub)
type Backend struct {
	initialized bool
	mutex       sync.Mutex
}

// Model represents a loaded llama.cpp model (stub)
type Model struct {
	path   string
	params ModelParams
}

// Context represents a llama.cpp context (stub)
type Context struct {
	model  *Model
	params ContextParams
}

// ModelParams represents model loading parameters (stub)
type ModelParams struct {
	UseMemoryMap  bool
	UseMemoryLock bool
	VocabOnly     bool
	GPULayers     int
	MainGPU       int
	TensorSplit   []float32
}

// ContextParams represents context parameters (stub)
type ContextParams struct {
	ContextSize   int
	BatchSize     int
	Threads       int
	RopeFreqBase  float32
	RopeFreqScale float32
}

// Token represents a llama token (stub)
type Token int32

// Initialize initializes the llama.cpp backend (stub)
func Initialize() error {
	return fmt.Errorf("llama.cpp not available: build with CGO enabled and llama.cpp library")
}

// LoadModel loads a model from file (stub)
func LoadModel(path string, params ModelParams) (*Model, error) {
	return nil, fmt.Errorf("llama.cpp not available: build with CGO enabled and llama.cpp library")
}

// NewContext creates a new context for the model (stub)
func (m *Model) NewContext(params ContextParams) (*Context, error) {
	return nil, fmt.Errorf("llama.cpp not available: build with CGO enabled and llama.cpp library")
}

// Tokenize converts text to tokens (stub)
func (c *Context) Tokenize(text string, addBOS bool) ([]Token, error) {
	return nil, fmt.Errorf("llama.cpp not available: build with CGO enabled and llama.cpp library")
}

// Detokenize converts tokens to text (stub)
func (c *Context) Detokenize(tokens []Token) (string, error) {
	return "", fmt.Errorf("llama.cpp not available: build with CGO enabled and llama.cpp library")
}

// Eval evaluates tokens through the model (stub)
func (c *Context) Eval(tokens []Token, nPast int) error {
	return fmt.Errorf("llama.cpp not available: build with CGO enabled and llama.cpp library")
}

// Sample samples the next token (stub)
func (c *Context) Sample(temperature float32, topP float32, topK int) (Token, error) {
	return 0, fmt.Errorf("llama.cpp not available: build with CGO enabled and llama.cpp library")
}

// GetVocabSize returns the vocabulary size (stub)
func (m *Model) GetVocabSize() int {
	return 0
}

// GetContextSize returns the context size (stub)
func (c *Context) GetContextSize() int {
	return 0
}

// Free methods (stub)
func (m *Model) Free() {
	// No-op for stub
}

func (c *Context) Free() {
	// No-op for stub
}
