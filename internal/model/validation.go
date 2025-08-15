package model

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ModelFormat represents different model file formats
type ModelFormat int

const (
	FormatUnknown ModelFormat = iota
	FormatGGUF
	FormatGGML
	FormatSafeTensors
	FormatPyTorch
	FormatONNX
)

// String returns the string representation of the model format
func (f ModelFormat) String() string {
	switch f {
	case FormatGGUF:
		return "GGUF"
	case FormatGGML:
		return "GGML"
	case FormatSafeTensors:
		return "SafeTensors"
	case FormatPyTorch:
		return "PyTorch"
	case FormatONNX:
		return "ONNX"
	default:
		return "Unknown"
	}
}

// ModelInfo represents information about a model file
type ModelInfo struct {
	Format      ModelFormat
	Version     string
	Architecture string
	Parameters  int64
	ContextSize int
	VocabSize   int
	Valid       bool
	Error       string
}

// GGUF magic number and constants
const (
	GGUFMagic    = 0x46554747 // "GGUF"
	GGMLMagic    = 0x67676d6c // "ggml"
	GGUFVersion2 = 2
	GGUFVersion3 = 3
)

// GGUF metadata value types
const (
	GGUFTypeUint8   = 0
	GGUFTypeInt8    = 1
	GGUFTypeUint16  = 2
	GGUFTypeInt16   = 3
	GGUFTypeUint32  = 4
	GGUFTypeInt32   = 5
	GGUFTypeFloat32 = 6
	GGUFTypeBool    = 7
	GGUFTypeString  = 8
	GGUFTypeArray   = 9
	GGUFTypeUint64  = 10
	GGUFTypeInt64   = 11
	GGUFTypeFloat64 = 12
)

// ValidateModel validates a model file and returns information about it
func ValidateModel(path string) (*ModelInfo, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &ModelInfo{
			Format: FormatUnknown,
			Valid:  false,
			Error:  "File not found",
		}, nil
	}

	// Open file for reading
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Detect format based on file extension and magic numbers
	format := detectFormat(path, file)
	
	switch format {
	case FormatGGUF:
		return validateGGUF(file)
	case FormatGGML:
		return validateGGML(file)
	case FormatSafeTensors:
		return validateSafeTensors(file)
	case FormatPyTorch:
		return validatePyTorch(file)
	case FormatONNX:
		return validateONNX(file)
	default:
		return &ModelInfo{
			Format: FormatUnknown,
			Valid:  false,
			Error:  "Unsupported model format",
		}, nil
	}
}

// detectFormat detects the model format based on file extension and magic numbers
func detectFormat(path string, file *os.File) ModelFormat {
	ext := strings.ToLower(filepath.Ext(path))
	
	// Check file extension first
	switch ext {
	case ".gguf":
		return FormatGGUF
	case ".ggml", ".bin":
		// Could be GGML or PyTorch, check magic number
		if magic := readMagic(file); magic == GGMLMagic {
			return FormatGGML
		}
		// Check for PyTorch format
		if isPyTorchFile(file) {
			return FormatPyTorch
		}
		// Default to GGML for .bin files if not PyTorch
		if ext == ".bin" {
			return FormatGGML
		}
	case ".safetensors":
		return FormatSafeTensors
	case ".onnx":
		return FormatONNX
	case ".pt", ".pth":
		return FormatPyTorch
	}
	
	// If extension doesn't match, try to detect by magic number
	magic := readMagic(file)
	switch magic {
	case GGUFMagic:
		return FormatGGUF
	case GGMLMagic:
		return FormatGGML
	}
	
	return FormatUnknown
}

// readMagic reads the first 4 bytes as a magic number
func readMagic(file *os.File) uint32 {
	file.Seek(0, 0) // Seek to beginning
	var magic uint32
	binary.Read(file, binary.LittleEndian, &magic)
	return magic
}

// validateGGUF validates a GGUF format model
func validateGGUF(file *os.File) (*ModelInfo, error) {
	file.Seek(0, 0)
	
	info := &ModelInfo{
		Format: FormatGGUF,
		Valid:  true,
	}
	
	// Read GGUF header
	var magic uint32
	var version uint32
	var tensorCount uint64
	var metadataKVCount uint64
	
	if err := binary.Read(file, binary.LittleEndian, &magic); err != nil {
		info.Valid = false
		info.Error = "Failed to read magic number"
		return info, nil
	}
	
	if magic != GGUFMagic {
		info.Valid = false
		info.Error = "Invalid GGUF magic number"
		return info, nil
	}
	
	if err := binary.Read(file, binary.LittleEndian, &version); err != nil {
		info.Valid = false
		info.Error = "Failed to read version"
		return info, nil
	}
	
	if version != GGUFVersion2 && version != GGUFVersion3 {
		info.Valid = false
		info.Error = fmt.Sprintf("Unsupported GGUF version: %d", version)
		return info, nil
	}
	
	info.Version = fmt.Sprintf("v%d", version)
	
	if err := binary.Read(file, binary.LittleEndian, &tensorCount); err != nil {
		info.Valid = false
		info.Error = "Failed to read tensor count"
		return info, nil
	}
	
	if err := binary.Read(file, binary.LittleEndian, &metadataKVCount); err != nil {
		info.Valid = false
		info.Error = "Failed to read metadata count"
		return info, nil
	}
	
	// Parse metadata to extract model information
	metadata, err := parseGGUFMetadata(file, metadataKVCount)
	if err != nil {
		info.Valid = false
		info.Error = fmt.Sprintf("Failed to parse metadata: %v", err)
		return info, nil
	}
	
	// Extract model information from metadata
	if arch, ok := metadata["general.architecture"].(string); ok {
		info.Architecture = arch
	}
	
	if contextLength, ok := metadata[info.Architecture+".context_length"].(uint64); ok {
		info.ContextSize = int(contextLength)
	}
	
	if vocabSize, ok := metadata[info.Architecture+".vocab_size"].(uint64); ok {
		info.VocabSize = int(vocabSize)
	}
	
	// Estimate parameters from tensor count and model architecture
	info.Parameters = estimateParametersFromTensors(int64(tensorCount), info.Architecture)
	
	return info, nil
}

// validateGGML validates a GGML format model
func validateGGML(file *os.File) (*ModelInfo, error) {
	file.Seek(0, 0)
	
	info := &ModelInfo{
		Format: FormatGGML,
		Valid:  true,
	}
	
	// Read GGML header (simplified)
	var magic uint32
	if err := binary.Read(file, binary.LittleEndian, &magic); err != nil {
		info.Valid = false
		info.Error = "Failed to read magic number"
		return info, nil
	}
	
	if magic != GGMLMagic {
		info.Valid = false
		info.Error = "Invalid GGML magic number"
		return info, nil
	}
	
	// GGML validation is more complex and depends on the specific variant
	// This is a simplified check
	info.Architecture = "unknown"
	info.Parameters = 7000000000 // Default estimate
	
	return info, nil
}

// validateSafeTensors validates a SafeTensors format model
func validateSafeTensors(file *os.File) (*ModelInfo, error) {
	file.Seek(0, 0)
	
	info := &ModelInfo{
		Format: FormatSafeTensors,
		Valid:  true,
	}
	
	// Read SafeTensors header length
	var headerLength uint64
	if err := binary.Read(file, binary.LittleEndian, &headerLength); err != nil {
		info.Valid = false
		info.Error = "Failed to read header length"
		return info, nil
	}
	
	// Basic validation - header length should be reasonable
	if headerLength > 100*1024*1024 { // 100MB seems excessive for a header
		info.Valid = false
		info.Error = "Header length too large"
		return info, nil
	}
	
	info.Architecture = "transformer"
	
	return info, nil
}

// validatePyTorch validates a PyTorch format model
func validatePyTorch(file *os.File) (*ModelInfo, error) {
	info := &ModelInfo{
		Format: FormatPyTorch,
		Valid:  true,
	}
	
	// PyTorch files are pickled Python objects
	// Basic validation would require unpickling, which is complex
	// For now, just check if it looks like a valid pickle file
	
	file.Seek(0, 0)
	header := make([]byte, 10)
	file.Read(header)
	
	// Check for pickle protocol markers
	if len(header) > 0 && (header[0] == 0x80 || header[0] == ']' || header[0] == '(') {
		info.Architecture = "transformer"
		info.Parameters = 7000000000 // Default estimate
	} else {
		info.Valid = false
		info.Error = "Not a valid PyTorch file"
	}
	
	return info, nil
}

// validateONNX validates an ONNX format model
func validateONNX(file *os.File) (*ModelInfo, error) {
	info := &ModelInfo{
		Format: FormatONNX,
		Valid:  true,
	}
	
	// ONNX files are protobuf format
	// Basic validation would require protobuf parsing
	// For now, just check file size and basic structure
	
	stat, err := file.Stat()
	if err != nil {
		info.Valid = false
		info.Error = "Failed to get file info"
		return info, nil
	}
	
	if stat.Size() < 1024 { // Very small files are likely not valid models
		info.Valid = false
		info.Error = "File too small to be a valid ONNX model"
		return info, nil
	}
	
	info.Architecture = "onnx"
	
	return info, nil
}

// Helper functions

func parseGGUFMetadata(file *os.File, kvCount uint64) (map[string]interface{}, error) {
	metadata := make(map[string]interface{})
	
	for i := uint64(0); i < kvCount; i++ {
		key, err := readGGUFString(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read metadata key: %w", err)
		}
		
		value, err := readGGUFValue(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read metadata value for key %s: %w", key, err)
		}
		
		metadata[key] = value
	}
	
	return metadata, nil
}

func readGGUFString(file *os.File) (string, error) {
	var length uint64
	if err := binary.Read(file, binary.LittleEndian, &length); err != nil {
		return "", err
	}
	
	if length > 1024*1024 { // 1MB limit for strings
		return "", fmt.Errorf("string too long: %d bytes", length)
	}
	
	data := make([]byte, length)
	if _, err := io.ReadFull(file, data); err != nil {
		return "", err
	}
	
	return string(data), nil
}

func readGGUFValue(file *os.File) (interface{}, error) {
	var valueType uint32
	if err := binary.Read(file, binary.LittleEndian, &valueType); err != nil {
		return nil, err
	}
	
	switch valueType {
	case GGUFTypeUint8:
		var value uint8
		binary.Read(file, binary.LittleEndian, &value)
		return value, nil
	case GGUFTypeInt8:
		var value int8
		binary.Read(file, binary.LittleEndian, &value)
		return value, nil
	case GGUFTypeUint32:
		var value uint32
		binary.Read(file, binary.LittleEndian, &value)
		return value, nil
	case GGUFTypeInt32:
		var value int32
		binary.Read(file, binary.LittleEndian, &value)
		return value, nil
	case GGUFTypeUint64:
		var value uint64
		binary.Read(file, binary.LittleEndian, &value)
		return value, nil
	case GGUFTypeInt64:
		var value int64
		binary.Read(file, binary.LittleEndian, &value)
		return value, nil
	case GGUFTypeFloat32:
		var value float32
		binary.Read(file, binary.LittleEndian, &value)
		return value, nil
	case GGUFTypeString:
		return readGGUFString(file)
	case GGUFTypeBool:
		var value uint8
		binary.Read(file, binary.LittleEndian, &value)
		return value != 0, nil
	default:
		// Skip unknown types
		return nil, fmt.Errorf("unsupported value type: %d", valueType)
	}
}

func isPyTorchFile(file *os.File) bool {
	file.Seek(0, 0)
	header := make([]byte, 10)
	n, _ := file.Read(header)
	
	if n < 1 {
		return false
	}
	
	// Check for pickle protocol markers
	return header[0] == 0x80 || header[0] == ']' || header[0] == '('
}

func estimateParametersFromTensors(tensorCount int64, architecture string) int64 {
	// Rough estimation based on tensor count and architecture
	switch strings.ToLower(architecture) {
	case "llama":
		// LLaMA models have roughly this many tensors per billion parameters
		return tensorCount * 150000000 // Very rough estimate
	case "gpt2":
		return tensorCount * 100000000
	case "bert":
		return tensorCount * 80000000
	default:
		return tensorCount * 100000000 // Default estimate
	}
}

// IsValidModelFormat checks if a file extension is supported
func IsValidModelFormat(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	supportedExtensions := []string{".gguf", ".ggml", ".bin", ".safetensors", ".onnx", ".pt", ".pth"}
	
	for _, supportedExt := range supportedExtensions {
		if ext == supportedExt {
			return true
		}
	}
	
	return false
}

// GetSupportedFormats returns a list of supported model formats
func GetSupportedFormats() []string {
	return []string{"GGUF", "GGML", "SafeTensors", "PyTorch", "ONNX"}
}
