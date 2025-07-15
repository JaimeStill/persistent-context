package app

import (
	"context"
	
	"github.com/JaimeStill/persistent-context/pkg/journal"
	"github.com/JaimeStill/persistent-context/pkg/vectordb"
)

// HealthChecker interface for checking service health
type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// Dependencies holds all the dependencies for the HTTP server
type Dependencies struct {
	VectorDBHealth HealthChecker
	LLMHealth      HealthChecker
	Journal        journal.Journal
	VectorDB       vectordb.VectorDB
}