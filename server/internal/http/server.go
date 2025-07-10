package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	server *http.Server
	config *config.HTTPConfig
	engine *gin.Engine
}

// HealthChecker interface for checking service health
type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// Dependencies holds all the dependencies for the HTTP server
type Dependencies struct {
	VectorDBHealth HealthChecker
	LLMHealth      HealthChecker
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
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			Handler:      engine,
			ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		},
	}

	// Register routes
	s.registerRoutes(deps)

	return s
}

// registerRoutes sets up HTTP routes
func (s *Server) registerRoutes(deps *Dependencies) {
	// Health and monitoring endpoints
	s.engine.GET("/health", s.handleHealth)
	s.engine.GET("/ready", s.handleReady(deps))
	s.engine.GET("/metrics", s.handleMetrics)

	// API routes group (for future expansion)
	api := s.engine.Group("/api/v1")
	{
		// Memory endpoints (placeholders for Session 2)
		api.GET("/memories", s.handleGetMemories)
		api.POST("/memories", s.handleCreateMemory)

		// Persona endpoints (placeholders for Session 3)
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
func (s *Server) handleReady(deps *Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		// Check dependencies
		vectordbStatus := "healthy"
		llmStatus := "healthy"

		if deps.VectorDBHealth != nil {
			if err := deps.VectorDBHealth.HealthCheck(ctx); err != nil {
				vectordbStatus = "unhealthy"
			}
		} else {
			vectordbStatus = "unknown"
		}

		if deps.LLMHealth != nil {
			if err := deps.LLMHealth.HealthCheck(ctx); err != nil {
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

// Placeholder handlers for future sessions
func (s *Server) handleGetMemories(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "Memory retrieval endpoint - to be implemented in Session 2",
	})
}

func (s *Server) handleCreateMemory(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "Memory creation endpoint - to be implemented in Session 2",
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
