package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/logger"
	"github.com/JaimeStill/persistent-context/internal/types"
)

// ProcessingPipeline handles async capture processing with batching and prioritization
type ProcessingPipeline struct {
	config        config.MCPConfig
	filter        *FilterEngine
	httpClient    *http.Client
	eventQueue    chan *types.CaptureEvent
	priorityQueue chan *types.CaptureEvent
	batchBuffer   []*types.CaptureEvent
	batchMutex    sync.Mutex
	batchTimer    *time.Timer
	workers       []*Worker
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	logger        *logger.Logger

	// Performance metrics
	metrics *PipelineMetrics
}

// PipelineMetrics tracks pipeline performance
type PipelineMetrics struct {
	mutex            sync.RWMutex
	TotalEvents      int64         `json:"total_events"`
	FilteredEvents   int64         `json:"filtered_events"`
	ProcessedEvents  int64         `json:"processed_events"`
	FailedEvents     int64         `json:"failed_events"`
	AverageLatency   time.Duration `json:"average_latency"`
	BatchesSent      int64         `json:"batches_sent"`
	CurrentQueueSize int           `json:"current_queue_size"`
	CurrentBatchSize int           `json:"current_batch_size"`
}

// Worker handles event processing
type Worker struct {
	id       int
	pipeline *ProcessingPipeline
	logger   *logger.Logger
}

// NewProcessingPipeline creates a new capture processing pipeline
func NewProcessingPipeline(cfg config.MCPConfig, filter *FilterEngine, log *logger.Logger) *ProcessingPipeline {
	ctx, cancel := context.WithCancel(context.Background())

	pipeline := &ProcessingPipeline{
		config:        cfg,
		filter:        filter,
		httpClient:    &http.Client{Timeout: cfg.Timeout},
		eventQueue:    make(chan *types.CaptureEvent, cfg.BufferSize),
		priorityQueue: make(chan *types.CaptureEvent, cfg.PriorityQueueSize),
		batchBuffer:   make([]*types.CaptureEvent, 0, cfg.MaxBatchSize),
		ctx:           ctx,
		cancel:        cancel,
		logger:        log,
		metrics:       &PipelineMetrics{},
	}

	// Start workers
	pipeline.startWorkers()

	// Start batch timer
	pipeline.startBatchTimer()

	return pipeline
}

// startWorkers launches the worker goroutines
func (p *ProcessingPipeline) startWorkers() {
	p.workers = make([]*Worker, p.config.WorkerCount)

	for i := 0; i < p.config.WorkerCount; i++ {
		worker := &Worker{
			id:       i,
			pipeline: p,
			logger:   p.logger,
		}
		p.workers[i] = worker

		p.wg.Add(1)
		go worker.run()
	}

	p.logger.Info("Started MCP processing pipeline",
		"workers", p.config.WorkerCount,
		"buffer_size", p.config.BufferSize,
		"batch_size", p.config.MaxBatchSize)
}

// startBatchTimer starts the batching timer
func (p *ProcessingPipeline) startBatchTimer() {
	batchWindow := time.Duration(p.config.BatchWindowMs) * time.Millisecond
	p.batchTimer = time.AfterFunc(batchWindow, p.flushBatch)
}

// ProcessEvent adds an event to the processing pipeline
func (p *ProcessingPipeline) ProcessEvent(event *types.CaptureEvent) error {
	start := time.Now()

	// Update metrics
	p.metrics.mutex.Lock()
	p.metrics.TotalEvents++
	p.metrics.mutex.Unlock()

	// Apply filtering
	shouldCapture, priority := p.filter.ShouldCapture(event)
	if !shouldCapture {
		p.metrics.mutex.Lock()
		p.metrics.FilteredEvents++
		p.metrics.mutex.Unlock()
		return nil
	}

	// Set priority from filter
	event.Priority = priority
	event.Timestamp = time.Now()

	// Route to appropriate queue based on priority
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	default:
		if priority >= types.PriorityHigh {
			// High priority events bypass batching
			select {
			case p.priorityQueue <- event:
				p.logger.Debug("Event queued for priority processing",
					"type", event.Type,
					"source", event.Source,
					"priority", priority)
			default:
				p.logger.Warn("Priority queue full, dropping event",
					"type", event.Type,
					"source", event.Source)
				p.metrics.mutex.Lock()
				p.metrics.FailedEvents++
				p.metrics.mutex.Unlock()
				return fmt.Errorf("priority queue full")
			}
		} else {
			// Regular events go to batch processing
			p.addToBatch(event)
		}
	}

	// Update latency metrics
	latency := time.Since(start)
	p.metrics.mutex.Lock()
	p.metrics.AverageLatency = (p.metrics.AverageLatency + latency) / 2
	p.metrics.mutex.Unlock()

	return nil
}

// addToBatch adds an event to the current batch
func (p *ProcessingPipeline) addToBatch(event *types.CaptureEvent) {
	p.batchMutex.Lock()
	defer p.batchMutex.Unlock()

	p.batchBuffer = append(p.batchBuffer, event)

	// Update metrics
	p.metrics.mutex.Lock()
	p.metrics.CurrentBatchSize = len(p.batchBuffer)
	p.metrics.mutex.Unlock()

	// Flush if batch is full
	if len(p.batchBuffer) >= p.config.MaxBatchSize {
		p.flushBatchLocked()
	}
}

// flushBatch flushes the current batch (timer callback)
func (p *ProcessingPipeline) flushBatch() {
	p.batchMutex.Lock()
	defer p.batchMutex.Unlock()
	p.flushBatchLocked()
}

// flushBatchLocked flushes the current batch (must hold batchMutex)
func (p *ProcessingPipeline) flushBatchLocked() {
	if len(p.batchBuffer) == 0 {
		// Restart timer for next batch
		p.restartBatchTimer()
		return
	}

	// Send batch to event queue
	for _, event := range p.batchBuffer {
		select {
		case p.eventQueue <- event:
			// Event queued successfully
		case <-p.ctx.Done():
			return
		default:
			p.logger.Warn("Event queue full, dropping event",
				"type", event.Type,
				"source", event.Source)
			p.metrics.mutex.Lock()
			p.metrics.FailedEvents++
			p.metrics.mutex.Unlock()
		}
	}

	p.logger.Debug("Batch flushed",
		"batch_size", len(p.batchBuffer))

	// Update metrics
	p.metrics.mutex.Lock()
	p.metrics.BatchesSent++
	p.metrics.CurrentBatchSize = 0
	p.metrics.mutex.Unlock()

	// Clear batch
	p.batchBuffer = p.batchBuffer[:0]

	// Restart timer for next batch
	p.restartBatchTimer()
}

// restartBatchTimer restarts the batch timer
func (p *ProcessingPipeline) restartBatchTimer() {
	if p.batchTimer != nil {
		p.batchTimer.Stop()
	}
	batchWindow := time.Duration(p.config.BatchWindowMs) * time.Millisecond
	p.batchTimer = time.AfterFunc(batchWindow, p.flushBatch)
}

// run executes the worker loop
func (w *Worker) run() {
	defer w.pipeline.wg.Done()

	w.logger.Debug("Worker started", "worker_id", w.id)

	for {
		select {
		case <-w.pipeline.ctx.Done():
			w.logger.Debug("Worker stopping", "worker_id", w.id)
			return

		case event := <-w.pipeline.priorityQueue:
			// Process high priority events immediately
			w.processEvent(event)

		case event := <-w.pipeline.eventQueue:
			// Process regular events
			w.processEvent(event)
		}
	}
}

// processEvent processes a single event
func (w *Worker) processEvent(event *types.CaptureEvent) {
	start := time.Now()

	// Send to journal API
	err := w.sendToJournal(event)
	if err != nil {
		w.logger.Error("Failed to send event to journal",
			"error", err,
			"type", event.Type,
			"source", event.Source,
			"worker_id", w.id)

		w.pipeline.metrics.mutex.Lock()
		w.pipeline.metrics.FailedEvents++
		w.pipeline.metrics.mutex.Unlock()
		return
	}

	// Update metrics
	latency := time.Since(start)
	w.pipeline.metrics.mutex.Lock()
	w.pipeline.metrics.ProcessedEvents++
	w.pipeline.metrics.AverageLatency = (w.pipeline.metrics.AverageLatency + latency) / 2
	w.pipeline.metrics.mutex.Unlock()

	w.logger.Debug("Event processed successfully",
		"type", event.Type,
		"source", event.Source,
		"latency", latency,
		"worker_id", w.id)
}

// sendToJournal sends an event to the journal API
func (w *Worker) sendToJournal(event *types.CaptureEvent) error {
	// Convert CaptureEvent to journal capture format
	payload := map[string]any{
		"source":   string(event.Type), // Use event type as source
		"content":  event.Content,
		"metadata": event.Metadata,
	}

	// Marshal payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create HTTP request
	url := w.pipeline.config.WebAPIURL + "/api/v1/journal"
	req, err := http.NewRequestWithContext(w.pipeline.ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Make HTTP request with retry logic
	var lastErr error
	for attempt := 0; attempt <= w.pipeline.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-w.pipeline.ctx.Done():
				return w.pipeline.ctx.Err()
			case <-time.After(w.pipeline.config.RetryDelay):
				// Continue with retry
			}
		}

		resp, err := w.pipeline.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("HTTP request failed (attempt %d): %w", attempt+1, err)
			continue
		}

		// Check response status
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			resp.Body.Close()
			return nil // Success
		}

		// Read error response
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		lastErr = fmt.Errorf("HTTP request failed with status %d (attempt %d): %s",
			resp.StatusCode, attempt+1, string(body))

		// Don't retry on client errors (4xx)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			break
		}
	}

	return lastErr
}

// GetMetrics returns current pipeline metrics
func (p *ProcessingPipeline) GetMetrics() *PipelineMetrics {
	p.metrics.mutex.RLock()
	defer p.metrics.mutex.RUnlock()

	// Update current queue sizes
	p.metrics.CurrentQueueSize = len(p.eventQueue) + len(p.priorityQueue)

	// Return a copy of metrics
	return &PipelineMetrics{
		TotalEvents:      p.metrics.TotalEvents,
		FilteredEvents:   p.metrics.FilteredEvents,
		ProcessedEvents:  p.metrics.ProcessedEvents,
		FailedEvents:     p.metrics.FailedEvents,
		AverageLatency:   p.metrics.AverageLatency,
		BatchesSent:      p.metrics.BatchesSent,
		CurrentQueueSize: p.metrics.CurrentQueueSize,
		CurrentBatchSize: p.metrics.CurrentBatchSize,
	}
}

// Shutdown gracefully shuts down the pipeline
func (p *ProcessingPipeline) Shutdown() error {
	p.logger.Info("Shutting down MCP processing pipeline")

	// Stop accepting new events
	p.cancel()

	// Flush remaining batch
	p.flushBatch()

	// Stop batch timer
	if p.batchTimer != nil {
		p.batchTimer.Stop()
	}

	// Close channels
	close(p.eventQueue)
	close(p.priorityQueue)

	// Wait for workers to finish
	p.wg.Wait()

	p.logger.Info("MCP processing pipeline shut down successfully")
	return nil
}
