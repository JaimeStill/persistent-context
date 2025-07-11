package services

import (
	"context"
	"fmt"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/vectordb"
)

// VectorDBService wraps vector database implementations as a managed service
type VectorDBService struct {
	BaseService
	db     vectordb.VectorDB
	config *config.VectorDBConfig
}

// NewVectorDBService creates a new vector database service
func NewVectorDBService(cfg *config.VectorDBConfig) *VectorDBService {
	return &VectorDBService{
		BaseService: NewBaseService("vectordb"),
		config:      cfg,
	}
}

// Initialize creates the appropriate vector database implementation
func (s *VectorDBService) Initialize(ctx context.Context) error {
	if s.IsInitialized() {
		return nil
	}

	// Create VectorDB config
	vdbConfig := &vectordb.Config{
		Provider:        s.config.Provider,
		URL:             s.config.URL,
		APIKey:          "", // API key not needed for development
		CollectionNames: s.config.CollectionNames,
		VectorDimension: s.config.VectorDimension,
		OnDiskPayload:   s.config.OnDiskPayload,
		Insecure:        s.config.Insecure,
	}

	// Create the vector database implementation
	db, err := vectordb.NewVectorDB(vdbConfig)
	if err != nil {
		return fmt.Errorf("failed to create vector database: %w", err)
	}

	// Initialize the database
	if err := db.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize vector database: %w", err)
	}

	s.db = db
	s.SetInitialized(true)
	return nil
}

// Start begins vector database operations
func (s *VectorDBService) Start(ctx context.Context) error {
	if !s.IsInitialized() {
		return fmt.Errorf("service not initialized")
	}

	s.SetRunning(true)
	return nil
}

// Stop gracefully shuts down the vector database service
func (s *VectorDBService) Stop(ctx context.Context) error {
	s.SetRunning(false)
	return nil
}

// HealthCheck verifies the vector database is accessible
func (s *VectorDBService) HealthCheck(ctx context.Context) error {
	if !s.IsInitialized() {
		return fmt.Errorf("service not initialized")
	}

	return s.db.HealthCheck(ctx)
}

// DB returns the vector database interface
func (s *VectorDBService) DB() vectordb.VectorDB {
	return s.db
}