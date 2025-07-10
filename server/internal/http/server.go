package http

import (
	"context"
	"net/http"
	"time"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	server *http.Server
	config *config.Config
	engine *gin.Engine
}

// HealthChecker interface for checking service health
type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// Dependencies holds all the dependencies for the HTTP server
type Dependencies struct {
	QdrantHealth HealthChecker
	OllamaHealth HealthChecker
}

// NewServer creates a new HTTP server using Gin
func NewServer(cfg *config.Config, deps *Dependencies) *Server {
	// Set Gin mode based on config
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// Add middleware
	engine.Use(gin.Recovery())
	if cfg.Logging.Level == "debug" {
		engine.Use(gin.Logger())
	}

	s := &Server{
		config: cfg,
		engine: engine,
		server: &http.Server{
			Addr:         ":" + cfg.HTTP.Port,
			Handler:      engine,
			ReadTimeout:  time.Duration(cfg.HTTP.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.HTTP.WriteTimeout) * time.Second,
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
		qdrantStatus := "healthy"
		ollamaStatus := "healthy"

		if deps.QdrantHealth != nil {
			if err := deps.QdrantHealth.HealthCheck(ctx); err != nil {
				qdrantStatus = "unhealthy"
			}
		} else {
			qdrantStatus = "unknown"
		}

		if deps.OllamaHealth != nil {
			if err := deps.OllamaHealth.HealthCheck(ctx); err != nil {
				ollamaStatus = "unhealthy"
			}
		} else {
			ollamaStatus = "unknown"
		}

		// Determine overall readiness
		ready := qdrantStatus == "healthy" && ollamaStatus == "healthy"
		status := http.StatusOK
		if !ready {
			status = http.StatusServiceUnavailable
		}

		c.JSON(status, gin.H{
			"status": map[bool]string{true: "ready", false: "not_ready"}[ready],
			"ready":  ready,
			"dependencies": gin.H{
				"qdrant": qdrantStatus,
				"ollama": ollamaStatus,
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
