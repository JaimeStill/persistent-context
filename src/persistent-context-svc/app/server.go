package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/JaimeStill/persistent-context/pkg/models"
	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	server *http.Server
	config *Config
	engine *gin.Engine
	deps   *Dependencies
}

// NewServer creates a new HTTP server using Gin
func NewServer(cfg *Config, deps *Dependencies) *Server {
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
			Addr:         fmt.Sprintf(":%d", cfg.HTTP.Port),
			Handler:      engine,
			ReadTimeout:  time.Duration(cfg.HTTP.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.HTTP.WriteTimeout) * time.Second,
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
	
	// Admin endpoints
	s.engine.POST("/admin/init", s.handleInitialize)

	// API routes group
	api := s.engine.Group("/api/v1")
	{
		// Journal endpoints
		api.POST("/journal", s.handleCaptureMemory)
		api.GET("/journal", s.handleGetMemories)
		api.POST("/journal/search", s.handleSearchMemories)
		api.POST("/journal/consolidate", s.handleConsolidation)
		api.GET("/journal/stats", s.handleGetMemoryStats)
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

// handleInitialize handles POST /admin/init - initializes VectorDB collections
func (s *Server) handleInitialize(c *gin.Context) {
	ctx := c.Request.Context()
	
	// Initialize VectorDB collections
	if err := s.deps.VectorDB.Initialize(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "initialization_failed",
			Message: fmt.Sprintf("failed to initialize vector database: %v", err),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":    "initialized",
		"message":   "VectorDB collections initialized successfully",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Journal endpoint handlers

// handleCaptureMemory handles POST /api/v1/journal
func (s *Server) handleCaptureMemory(c *gin.Context) {
	var req models.CaptureMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	entry, err := s.deps.Journal.CaptureContext(ctx, req.Source, req.Content, req.Metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "capture_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.CaptureMemoryResponse{
		ID:      entry.ID,
		Message: "Memory captured successfully",
	})
}

// handleGetMemories handles GET /api/v1/journal
func (s *Server) handleGetMemories(c *gin.Context) {
	var req models.GetMemoriesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Apply default limit if not specified
	limit := req.Limit
	if limit == 0 {
		limit = 100 // Default limit
	}

	ctx := c.Request.Context()
	memories, err := s.deps.Journal.GetMemories(ctx, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "retrieval_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.GetMemoriesResponse{
		Memories: memories,
		Count:    len(memories),
		Limit:    req.Limit,
	})
}


// handleSearchMemories handles POST /api/v1/journal/search
func (s *Server) handleSearchMemories(c *gin.Context) {
	var req models.SearchMemoriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Apply default memory type if not specified
	memType := models.TypeEpisodic
	if req.MemoryType != "" {
		memType = models.MemoryType(req.MemoryType)
	}

	// Apply default limit if not specified
	limit := req.Limit
	if limit == 0 {
		limit = 10 // Default limit
	}

	ctx := c.Request.Context()
	memories, err := s.deps.Journal.QuerySimilarMemories(ctx, req.Content, memType, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "search_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SearchMemoriesResponse{
		Memories: memories,
		Query:    req.Content,
		Count:    len(memories),
		Limit:    limit,
	})
}

// handleConsolidation handles POST /api/v1/journal/consolidate
func (s *Server) handleConsolidation(c *gin.Context) {
	ctx := c.Request.Context()

	// Get recent episodic memories for intelligent consolidation
	memories, err := s.deps.Journal.GetMemories(ctx, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "retrieval_failed",
			Message: fmt.Sprintf("Failed to get memories for consolidation: %v", err),
		})
		return
	}

	// Group memories by associations for intelligent consolidation
	groupedMemories := s.groupMemoriesByAssociations(memories)
	
	totalProcessed := 0
	groupsConsolidated := 0
	
	for _, group := range groupedMemories {
		// Only consolidate groups with multiple memories
		if len(group) > 1 {
			err = s.deps.Journal.ConsolidateMemories(ctx, group)
			if err != nil {
				slog.Warn("Failed to consolidate memory group", "error", err, "group_size", len(group))
				continue
			}
			totalProcessed += len(group)
			groupsConsolidated++
		}
	}

	c.JSON(http.StatusOK, models.ConsolidateResponse{
		Message:            "Intelligent consolidation completed",
		GroupsFormed:       len(groupedMemories),
		GroupsConsolidated: groupsConsolidated,
		MemoriesProcessed:  totalProcessed,
		TotalMemories:      len(memories),
	})
}

// handleGetMemoryStats handles GET /api/v1/journal/stats
func (s *Server) handleGetMemoryStats(c *gin.Context) {
	ctx := c.Request.Context()
	stats, err := s.deps.Journal.GetMemoryStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "stats_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.StatsResponse{
		Stats: stats,
	})
}

// groupMemoriesByAssociations groups memories that share associations for targeted consolidation
func (s *Server) groupMemoriesByAssociations(memories []*models.MemoryEntry) [][]*models.MemoryEntry {
	// Track which memories have been grouped
	grouped := make(map[string]bool)
	var groups [][]*models.MemoryEntry
	
	for _, memory := range memories {
		if grouped[memory.ID] {
			continue // Skip already grouped memories
		}
		
		// Start a new group with this memory
		group := []*models.MemoryEntry{memory}
		grouped[memory.ID] = true
		
		// Find related memories through associations
		for _, candidate := range memories {
			if grouped[candidate.ID] {
				continue
			}
			
			// Check if memories share associations (bidirectional)
			if s.memoriesShareAssociations(memory, candidate) {
				group = append(group, candidate)
				grouped[candidate.ID] = true
			}
		}
		
		groups = append(groups, group)
	}
	
	return groups
}

// memoriesShareAssociations checks if two memories have overlapping associations
func (s *Server) memoriesShareAssociations(memory1, memory2 *models.MemoryEntry) bool {
	// Check direct association IDs
	for _, id1 := range memory1.AssociationIDs {
		for _, id2 := range memory2.AssociationIDs {
			if id1 == id2 {
				return true
			}
		}
		// Also check if memory2 is directly associated with memory1
		if id1 == memory2.ID {
			return true
		}
	}
	
	// Check if memory1 is directly associated with memory2
	for _, id := range memory2.AssociationIDs {
		if id == memory1.ID {
			return true
		}
	}
	
	return false
}


// Start starts the HTTP server
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
