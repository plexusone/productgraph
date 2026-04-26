package events

import (
	"context"
	"log/slog"
	"sync"

	"github.com/plexusone/productgraph/pkg/schema"
)

// MemoryPublisher is an in-memory publisher for testing.
// It stores events in memory and logs them.
type MemoryPublisher struct {
	logger *slog.Logger
	mu     sync.RWMutex
	events []schema.Event
}

// NewMemoryPublisher creates a new in-memory publisher.
func NewMemoryPublisher(logger *slog.Logger) *MemoryPublisher {
	return &MemoryPublisher{
		logger: logger,
		events: make([]schema.Event, 0),
	}
}

// Publish stores events in memory.
func (p *MemoryPublisher) Publish(ctx context.Context, events []schema.Event) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, event := range events {
		p.logger.DebugContext(ctx, "received event",
			"event_id", event.EventID,
			"type", event.EventType,
			"session_id", event.SessionID,
			"page_path", event.PagePath,
		)
	}

	p.events = append(p.events, events...)
	return nil
}

// Events returns all stored events.
func (p *MemoryPublisher) Events() []schema.Event {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]schema.Event, len(p.events))
	copy(result, p.events)
	return result
}

// Count returns the number of stored events.
func (p *MemoryPublisher) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.events)
}

// Clear removes all stored events.
func (p *MemoryPublisher) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = make([]schema.Event, 0)
}
