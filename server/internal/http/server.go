package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/journal"
	"github.com/JaimeStill/persistent-context/internal/types"
	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	server *http.Server
	config *config.HTTPConfig
	engine *gin.Engine
	deps   *Dependencies
}

// HealthChecker interface for checking service health
type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// Dependencies holds all the dependencies for the HTTP server
type Dependencies struct {
	VectorDBHealth HealthChecker
	LLMHealth      HealthChecker
	Journal        journal.Journal
}

// NewServer creates a new HTTP server using Gin
func NewServer(cfg *config.HTTPConfig, deps *Dependencies) *Server {
	// Set Gin mode - for now just use release mode
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()

	// Add middleware
	engine.Use(gin.Recovery())
	engine.Use(gin.Logger())

	s := &Server{
		config: cfg,
		engine: engine,
		deps:   deps,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%s", cfg.Port),
			Handler:      engine,
			ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		},
	}

	// Register routes
	s.registerRoutes()

	return s
}

// registerRoutes sets up HTTP routes
func (s *Server) registerRoutes() {
	// Health and monitoring endpoints
	s.engine.GET("/health", s.handleHealth)
	s.engine.GET("/ready", s.handleReady)
	s.engine.GET("/metrics", s.handleMetrics)

	// API routes group  
	api := s.engine.Group("/api/v1")
	{
		// Journal endpoints
		api.POST("/journal", s.handleCaptureMemory)
		api.GET("/journal", s.handleGetMemories)
		api.GET("/journal/:id", s.handleGetMemoryByID)
		api.POST("/journal/search", s.handleSearchMemories)
		api.POST("/journal/consolidate", s.handleConsolidateMemories)
		api.GET("/journal/stats", s.handleGetMemoryStats)

		// Persona endpoints (placeholders for Session 9)
		api.GET("/personas", s.handleGetPersonas)
		api.POST("/personas/export", s.handleExportPersona)
	}
}

// handleHealth returns a simple health check
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "persistent-context",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// handleReady checks if the service is ready with all dependencies
func (s *Server) handleReady(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		// Check dependencies
		vectordbStatus := "healthy"
		llmStatus := "healthy"

		if s.deps.VectorDBHealth != nil {
			if err := s.deps.VectorDBHealth.HealthCheck(ctx); err != nil {
				vectordbStatus = "unhealthy"
			}
		} else {
			vectordbStatus = "unknown"
		}

		if s.deps.LLMHealth != nil {
			if err := s.deps.LLMHealth.HealthCheck(ctx); err != nil {
				llmStatus = "unhealthy"
			}
		} else {
			llmStatus = "unknown"
		}

		// Determine overall readiness
		ready := vectordbStatus == "healthy" && llmStatus == "healthy"
		status := http.StatusOK
		if !ready {
			status = http.StatusServiceUnavailable
		}

		c.JSON(status, gin.H{
			"status": map[bool]string{true: "ready", false: "not_ready"}[ready],
			"ready":  ready,
			"dependencies": gin.H{
				"vectordb": vectordbStatus,
				"llm":      llmStatus,
			},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
}

// handleMetrics returns basic metrics (placeholder for now)
func (s *Server) handleMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"metrics": gin.H{
			"uptime":               time.Since(time.Now()).String(), // Placeholder
			"memory_entries":       0,
			"consolidation_cycles": 0,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Journal endpoint handlers

// handleCaptureMemory handles POST /api/v1/journal
func (s *Server) handleCaptureMemory(c *gin.Context) {
		var req CaptureMemoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_request",
				Message: err.Error(),
			})
			return
		}

		ctx := c.Request.Context()
		entry, err := s.deps.Journal.CaptureContext(ctx, req.Source, req.Content, req.Metadata)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "capture_failed",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, CaptureMemoryResponse{
			ID:      entry.ID,
			Message: "Memory captured successfully",
		})
}

// handleGetMemories handles GET /api/v1/journal
func (s *Server) handleGetMemories(c *gin.Context) {
		var req GetMemoriesRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_request",
				Message: err.Error(),
			})
			return
		}

	ctx := c.Request.Context()
	memories, err := s.deps.Journal.GetMemories(ctx, req.Limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "retrieval_failed",
				Message: err.Error(),
			})
			return
		}

	c.JSON(http.StatusOK, GetMemoriesResponse{
		Memories: memories,
		Count:    len(memories),
		Limit:    req.Limit,
	})
}

// handleGetMemoryByID handles GET /api/v1/journal/:id
func (s *Server) handleGetMemoryByID(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_request",
				Message: "Memory ID is required",
			})
			return
		}

	ctx := c.Request.Context()
	memory, err := s.deps.Journal.GetMemoryByID(ctx, id)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "memory_not_found",
				Message: err.Error(),
			})
			return
		}

	c.JSON(http.StatusOK, memory)
}

// handleSearchMemories handles POST /api/v1/journal/search
func (s *Server) handleSearchMemories(c *gin.Context) {
		var req SearchMemoriesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_request",
				Message: err.Error(),
			})
			return
		}

		// Default to episodic memory type if not specified
		memType := types.TypeEpisodic
		if req.MemoryType != "" {
			memType = types.MemoryType(req.MemoryType)
		}

		// Default limit if not specified
		if req.Limit == 0 {
			req.Limit = 10
		}

	ctx := c.Request.Context()
	memories, err := s.deps.Journal.QuerySimilarMemories(ctx, req.Content, memType, req.Limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "search_failed",
				Message: err.Error(),
			})
			return
		}

	c.JSON(http.StatusOK, SearchMemoriesResponse{
		Memories: memories,
		Query:    req.Content,
		Count:    len(memories),
		Limit:    req.Limit,
	})
}

// handleConsolidateMemories handles POST /api/v1/journal/consolidate
func (s *Server) handleConsolidateMemories(c *gin.Context) {
		var req ConsolidateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_request",
				Message: err.Error(),
			})
			return
		}

		ctx := c.Request.Context()
		
	// Retrieve memories by IDs
	var memories []*types.MemoryEntry
	for _, id := range req.MemoryIDs {
		memory, err := s.deps.Journal.GetMemoryByID(ctx, id)
			if err != nil {
				c.JSON(http.StatusNotFound, ErrorResponse{
					Error:   "memory_not_found",
					Message: fmt.Sprintf("Memory with ID %s not found: %v", id, err),
				})
				return
			}
			memories = append(memories, memory)
		}

	// Consolidate memories
	err := s.deps.Journal.ConsolidateMemories(ctx, memories)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "consolidation_failed",
				Message: err.Error(),
			})
			return
		}

	c.JSON(http.StatusOK, ConsolidateResponse{
		Message:        "Memories consolidated successfully",
		ProcessedCount: len(memories),
	})
}

// handleGetMemoryStats handles GET /api/v1/journal/stats
func (s *Server) handleGetMemoryStats(c *gin.Context) {
	ctx := c.Request.Context()
	stats, err := s.deps.Journal.GetMemoryStats(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "stats_failed",
				Message: err.Error(),
			})
			return
		}

	c.JSON(http.StatusOK, StatsResponse{
		Stats: stats,
	})
}

func (s *Server) handleGetPersonas(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "Persona listing endpoint - to be implemented in Session 3",
	})
}

func (s *Server) handleExportPersona(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "Persona export endpoint - to be implemented in Session 3",
	})
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
