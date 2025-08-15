//go:build cgo && llamacpp_cgo

package llama

/*
#cgo CFLAGS: -I${SRCDIR}/../../third_party/llama.cpp
#cgo LDFLAGS: -L${SRCDIR}/../../third_party/llama.cpp -lllama -lm -lstdc++
#cgo linux LDFLAGS: -lrt -ldl -lpthread
#cgo darwin LDFLAGS: -framework Foundation -framework Metal -framework MetalKit

#ifdef GGML_USE_CUBLAS
#cgo CFLAGS: -DGGML_USE_CUBLAS
#cgo LDFLAGS: -lcublas -lcudart -lcurand -lcublasLt
#endif

#ifdef GGML_USE_HIPBLAS
#cgo CFLAGS: -DGGML_USE_HIPBLAS -D__HIP_PLATFORM_AMD__
#cgo LDFLAGS: -lhipblas -lrocblas -lamdhip64
#endif

#include <stdlib.h>
#include <string.h>
#include "llama.h"

// Helper wrapper functions for easier CGO integration
typedef struct {
    struct llama_model* model;
    struct llama_context* ctx;
    struct llama_context_params ctx_params;
    struct llama_model_params model_params;
} llama_wrapper_t;

// Create model parameters with default values
struct llama_model_params llama_model_default_params_wrapper() {
    return llama_model_default_params();
}

// Create context parameters with default values  
struct llama_context_params llama_context_default_params_wrapper() {
    return llama_context_default_params();
}

// Load model from file
struct llama_model* llama_load_model_wrapper(const char* path, struct llama_model_params params) {
    return llama_load_model_from_file(path, params);
}

// Create context
struct llama_context* llama_new_context_wrapper(struct llama_model* model, struct llama_context_params params) {
    return llama_new_context_with_model(model, params);
}

// Tokenize text
int llama_tokenize_wrapper(struct llama_context* ctx, const char* text, int text_len, llama_token* tokens, int max_tokens, bool add_bos, bool special) {
    return llama_tokenize(llama_get_model(ctx), text, text_len, tokens, max_tokens, add_bos, special);
}

// Detokenize tokens
int llama_token_to_piece_wrapper(struct llama_context* ctx, llama_token token, char* buf, int length) {
    return llama_token_to_piece(llama_get_model(ctx), token, buf, length);
}

// Evaluate tokens
int llama_eval_wrapper(struct llama_context* ctx, llama_token* tokens, int n_tokens, int n_past) {
    return llama_decode(ctx, llama_batch_get_one(tokens, n_tokens, n_past, 0));
}

// Sample next token
llama_token llama_sample_token_wrapper(struct llama_context* ctx, llama_token* candidates, int n_candidates, float temp, float top_p, int top_k) {
    struct llama_sampling_params params = {
        .temp = temp,
        .top_p = top_p,
        .top_k = top_k,
        .penalty_repeat = 1.1f,
        .penalty_freq = 0.0f,
        .penalty_present = 0.0f,
    };
    
    // This is a simplified sampling - real implementation would be more complex
    llama_token_data_array candidates_p = {candidates, (size_t)n_candidates, false};
    
    if (temp > 0) {
        llama_sample_temp(ctx, &candidates_p, temp);
        if (top_p < 1.0f) {
            llama_sample_nucleus(ctx, &candidates_p, top_p, 1);
        }
        if (top_k > 0) {
            llama_sample_top_k(ctx, &candidates_p, top_k, 1);
        }
    }
    
    return llama_sample_token(ctx, &candidates_p);
}

// Get model information
void llama_model_info_wrapper(struct llama_model* model, char* buf, size_t buf_size) {
    snprintf(buf, buf_size, "Model loaded successfully");
}

// Free resources
void llama_free_model_wrapper(struct llama_model* model) {
    llama_free_model(model);
}

void llama_free_context_wrapper(struct llama_context* ctx) {
    llama_free(ctx);
}

*/
import "C"
import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

// Initialize llama.cpp backend
var (
	llamaInitOnce sync.Once
	llamaBackend  *Backend
)

// Backend represents the llama.cpp backend
type Backend struct {
	initialized bool
	mutex       sync.Mutex
}

// Model represents a loaded llama.cpp model
type Model struct {
	cModel *C.struct_llama_model
	path   string
	params ModelParams
}

// Context represents a llama.cpp context
type Context struct {
	cContext *C.struct_llama_context
	model    *Model
	params   ContextParams
}

// ModelParams represents model loading parameters
type ModelParams struct {
	UseMemoryMap  bool
	UseMemoryLock bool
	VocabOnly     bool
	GPULayers     int
	MainGPU       int
	TensorSplit   []float32
}

// ContextParams represents context parameters
type ContextParams struct {
	ContextSize int
	BatchSize   int
	Threads     int
	RopeFreqBase float32
	RopeFreqScale float32
}

// Token represents a llama token
type Token C.llama_token

// Initialize initializes the llama.cpp backend
func Initialize() error {
	var err error
	llamaInitOnce.Do(func() {
		C.llama_backend_init(false)
		llamaBackend = &Backend{initialized: true}
		
		// Set up cleanup on program exit
		runtime.SetFinalizer(llamaBackend, (*Backend).cleanup)
	})
	return err
}

// LoadModel loads a model from file
func LoadModel(path string, params ModelParams) (*Model, error) {
	if err := Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize llama backend: %w", err)
	}

	// Convert Go params to C params
	cParams := C.llama_model_default_params_wrapper()
	cParams.use_mmap = C.bool(params.UseMemoryMap)
	cParams.use_mlock = C.bool(params.UseMemoryLock)
	cParams.vocab_only = C.bool(params.VocabOnly)
	cParams.n_gpu_layers = C.int(params.GPULayers)
	cParams.main_gpu = C.int(params.MainGPU)

	// Handle tensor split for multi-GPU
	if len(params.TensorSplit) > 0 {
		for i, split := range params.TensorSplit {
			if i < len(cParams.tensor_split) {
				cParams.tensor_split[i] = C.float(split)
			}
		}
	}

	// Load the model
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	cModel := C.llama_load_model_wrapper(cPath, cParams)
	if cModel == nil {
		return nil, fmt.Errorf("failed to load model from %s", path)
	}

	model := &Model{
		cModel: cModel,
		path:   path,
		params: params,
	}

	// Set up cleanup
	runtime.SetFinalizer(model, (*Model).cleanup)

	return model, nil
}

// NewContext creates a new context for the model
func (m *Model) NewContext(params ContextParams) (*Context, error) {
	// Convert Go params to C params
	cParams := C.llama_context_default_params_wrapper()
	cParams.n_ctx = C.uint32_t(params.ContextSize)
	cParams.n_batch = C.uint32_t(params.BatchSize)
	cParams.n_threads = C.int(params.Threads)
	cParams.rope_freq_base = C.float(params.RopeFreqBase)
	cParams.rope_freq_scale = C.float(params.RopeFreqScale)

	// Create context
	cContext := C.llama_new_context_wrapper(m.cModel, cParams)
	if cContext == nil {
		return nil, fmt.Errorf("failed to create context")
	}

	context := &Context{
		cContext: cContext,
		model:    m,
		params:   params,
	}

	// Set up cleanup
	runtime.SetFinalizer(context, (*Context).cleanup)

	return context, nil
}

// Tokenize converts text to tokens
func (c *Context) Tokenize(text string, addBOS bool) ([]Token, error) {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	// Allocate token buffer
	maxTokens := len(text) + 256 // Rough estimate
	tokens := make([]C.llama_token, maxTokens)

	// Tokenize
	nTokens := C.llama_tokenize_wrapper(
		c.cContext,
		cText,
		C.int(len(text)),
		&tokens[0],
		C.int(maxTokens),
		C.bool(addBOS),
		true,
	)

	if nTokens < 0 {
		return nil, fmt.Errorf("tokenization failed")
	}

	// Convert to Go tokens
	result := make([]Token, nTokens)
	for i := 0; i < int(nTokens); i++ {
		result[i] = Token(tokens[i])
	}

	return result, nil
}

// Detokenize converts tokens to text
func (c *Context) Detokenize(tokens []Token) (string, error) {
	var result string

	for _, token := range tokens {
		buf := make([]C.char, 256)
		length := C.llama_token_to_piece_wrapper(
			c.cContext,
			C.llama_token(token),
			&buf[0],
			C.int(len(buf)),
		)

		if length > 0 {
			result += C.GoString(&buf[0])
		}
	}

	return result, nil
}

// Eval evaluates tokens through the model
func (c *Context) Eval(tokens []Token, nPast int) error {
	if len(tokens) == 0 {
		return nil
	}

	// Convert tokens to C array
	cTokens := make([]C.llama_token, len(tokens))
	for i, token := range tokens {
		cTokens[i] = C.llama_token(token)
	}

	// Evaluate
	result := C.llama_eval_wrapper(
		c.cContext,
		&cTokens[0],
		C.int(len(tokens)),
		C.int(nPast),
	)

	if result != 0 {
		return fmt.Errorf("evaluation failed with code %d", result)
	}

	return nil
}

// Sample samples the next token
func (c *Context) Sample(temperature float32, topP float32, topK int) (Token, error) {
	// Get logits (simplified approach)
	// In real implementation, you'd get logits from the context and create candidates
	candidates := make([]C.llama_token, 1)
	candidates[0] = 0 // Simplified - would use actual vocab

	token := C.llama_sample_token_wrapper(
		c.cContext,
		&candidates[0],
		C.int(len(candidates)),
		C.float(temperature),
		C.float(topP),
		C.int(topK),
	)

	return Token(token), nil
}

// GetVocabSize returns the vocabulary size
func (m *Model) GetVocabSize() int {
	return int(C.llama_n_vocab(C.llama_get_model(m.cModel)))
}

// GetContextSize returns the context size
func (c *Context) GetContextSize() int {
	return int(C.llama_n_ctx(c.cContext))
}

// cleanup methods for proper resource management

func (m *Model) cleanup() {
	if m.cModel != nil {
		C.llama_free_model_wrapper(m.cModel)
		m.cModel = nil
	}
}

func (c *Context) cleanup() {
	if c.cContext != nil {
		C.llama_free_context_wrapper(c.cContext)
		c.cContext = nil
	}
}

func (b *Backend) cleanup() {
	if b.initialized {
		C.llama_backend_free()
		b.initialized = false
	}
}

// Free manually frees resources (call this explicitly when done)
func (m *Model) Free() {
	m.cleanup()
	runtime.SetFinalizer(m, nil)
}

func (c *Context) Free() {
	c.cleanup()
	runtime.SetFinalizer(c, nil)
}
