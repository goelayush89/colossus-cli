package api

import (
	"encoding/json"
	"net/http"

	"colossus-cli/internal/config"
	"colossus-cli/internal/inference"
	"colossus-cli/internal/model"
	"colossus-cli/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Server represents the API server
type Server struct {
	config        *config.Config
	modelManager  *model.Manager
	engine        inference.InferenceEngine
	engineType    inference.EngineType
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, modelManager *model.Manager) *Server {
	engineType := inference.GetEngineTypeFromEnv()
	
	return &Server{
		config:       cfg,
		modelManager: modelManager,
		engine:       inference.NewEngine(engineType),
		engineType:   engineType,
	}
}

// Router returns the configured gin router
func (s *Server) Router() *gin.Engine {
	if !s.config.Verbose {
		gin.SetMode(gin.ReleaseMode)
	}
	
	r := gin.Default()
	
	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		
		c.Next()
	})
	
	// API routes
	api := r.Group("/api")
	{
		api.GET("/tags", s.listModels)
		api.POST("/pull", s.pullModel)
		api.DELETE("/delete", s.deleteModel)
		api.POST("/generate", s.generate)
		api.POST("/chat", s.chat)
	}
	
	// Health check
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Colossus API Server",
			"status":  "running",
		})
	})
	
	return r
}

// listModels handles GET /api/tags
func (s *Server) listModels(c *gin.Context) {
	models, err := s.modelManager.ListModels()
	if err != nil {
		logrus.Errorf("Failed to list models: %v", err)
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to list models",
		})
		return
	}
	
	c.JSON(http.StatusOK, types.ModelsResponse{
		Models: models,
	})
}

// pullModel handles POST /api/pull
func (s *Server) pullModel(c *gin.Context) {
	var req types.PullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid request",
		})
		return
	}
	
	// Set response headers for streaming
	c.Header("Content-Type", "application/x-ndjson")
	c.Header("Transfer-Encoding", "chunked")
	
	// Send progress updates
	encoder := json.NewEncoder(c.Writer)
	
	// Send initial status
	encoder.Encode(types.PullResponse{
		Status: "pulling manifest",
	})
	c.Writer.Flush()
	
	// Pull the model
	if err := s.modelManager.PullModel(req.Name); err != nil {
		encoder.Encode(types.PullResponse{
			Status: "error: " + err.Error(),
		})
		return
	}
	
	// Send completion status
	encoder.Encode(types.PullResponse{
		Status: "success",
	})
}

// deleteModel handles DELETE /api/delete
func (s *Server) deleteModel(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid request",
		})
		return
	}
	
	if err := s.modelManager.RemoveModel(req.Name); err != nil {
		c.JSON(http.StatusNotFound, types.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Model deleted successfully"})
}

// generate handles POST /api/generate
func (s *Server) generate(c *gin.Context) {
	var req types.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid request",
		})
		return
	}
	
	// Ensure model is loaded
	if err := s.ensureModelLoaded(req.Model); err != nil {
		c.JSON(http.StatusNotFound, types.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	
	if req.Stream {
		s.streamGenerate(c, &req)
	} else {
		s.simpleGenerate(c, &req)
	}
}

// chat handles POST /api/chat
func (s *Server) chat(c *gin.Context) {
	var req types.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid request",
		})
		return
	}
	
	// Ensure model is loaded
	if err := s.ensureModelLoaded(req.Model); err != nil {
		c.JSON(http.StatusNotFound, types.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	
	if req.Stream {
		s.streamChat(c, &req)
	} else {
		s.simpleChat(c, &req)
	}
}

// ensureModelLoaded loads a model if it's not already loaded
func (s *Server) ensureModelLoaded(modelName string) error {
	if s.engine.IsModelLoaded(modelName) {
		return nil
	}
	
	modelPath, err := s.modelManager.GetModelPath(modelName)
	if err != nil {
		return err
	}
	
	// Get appropriate options for the engine type
	options := inference.GetDefaultModelOptions(s.engineType)
	
	return s.engine.LoadModel(modelName, modelPath, options)
}

// simpleGenerate handles non-streaming generation
func (s *Server) simpleGenerate(c *gin.Context, req *types.GenerateRequest) {
	resp, err := s.engine.Generate(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, resp)
}

// streamGenerate handles streaming generation
func (s *Server) streamGenerate(c *gin.Context, req *types.GenerateRequest) {
	c.Header("Content-Type", "application/x-ndjson")
	c.Header("Transfer-Encoding", "chunked")
	
	encoder := json.NewEncoder(c.Writer)
	
	// Use the engine's streaming capability
	err := s.engine.GenerateStream(req, func(resp *types.GenerateResponse) error {
		if err := encoder.Encode(resp); err != nil {
			return err
		}
		c.Writer.Flush()
		return nil
	})
	
	if err != nil {
		encoder.Encode(types.ErrorResponse{Error: err.Error()})
	}
}

// simpleChat handles non-streaming chat
func (s *Server) simpleChat(c *gin.Context, req *types.ChatRequest) {
	resp, err := s.engine.Chat(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, resp)
}

// streamChat handles streaming chat
func (s *Server) streamChat(c *gin.Context, req *types.ChatRequest) {
	c.Header("Content-Type", "application/x-ndjson")
	c.Header("Transfer-Encoding", "chunked")
	
	encoder := json.NewEncoder(c.Writer)
	
	// Use the engine's streaming capability
	err := s.engine.ChatStream(req, func(resp *types.ChatResponse) error {
		if err := encoder.Encode(resp); err != nil {
			return err
		}
		c.Writer.Flush()
		return nil
	})
	
	if err != nil {
		encoder.Encode(types.ErrorResponse{Error: err.Error()})
	}
}


