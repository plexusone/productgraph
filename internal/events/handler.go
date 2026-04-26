// Package events provides event ingestion and processing.
package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/plexusone/productgraph/pkg/schema"
)

// Handler handles event ingestion HTTP requests.
type Handler struct {
	logger    *slog.Logger
	publisher Publisher
	maxBatch  int
}

// Publisher defines the interface for publishing events.
type Publisher interface {
	Publish(ctx context.Context, events []schema.Event) error
}

// NewHandler creates a new event handler.
func NewHandler(logger *slog.Logger, publisher Publisher) *Handler {
	return &Handler{
		logger:    logger,
		publisher: publisher,
		maxBatch:  1000, // Max events per batch
	}
}

// ServeHTTP handles POST /v1/events requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var batch schema.EventBatch
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		h.logger.WarnContext(ctx, "failed to decode request body", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Check batch size
	if len(batch.Events) > h.maxBatch {
		h.logger.WarnContext(ctx, "batch too large", "size", len(batch.Events), "max", h.maxBatch)
		http.Error(w, "batch too large", http.StatusBadRequest)
		return
	}

	// Validate events and collect errors
	var validEvents []schema.Event
	var errors []schema.IngestError

	for i, event := range batch.Events {
		// Set received timestamp if not set
		if event.Timestamp.IsZero() {
			event.Timestamp = time.Now()
		}

		if err := event.Validate(); err != nil {
			errors = append(errors, schema.IngestError{
				Index: i,
				Error: err.Error(),
			})
			continue
		}
		validEvents = append(validEvents, event)
	}

	// Publish valid events
	if len(validEvents) > 0 {
		if err := h.publisher.Publish(ctx, validEvents); err != nil {
			h.logger.ErrorContext(ctx, "failed to publish events", "error", err, "count", len(validEvents))
			http.Error(w, "failed to process events", http.StatusInternalServerError)
			return
		}
	}

	h.logger.InfoContext(ctx, "processed events",
		"accepted", len(validEvents),
		"rejected", len(errors),
	)

	// Send response
	response := schema.IngestResponse{
		Accepted: len(validEvents),
		Rejected: len(errors),
		Errors:   errors,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}
