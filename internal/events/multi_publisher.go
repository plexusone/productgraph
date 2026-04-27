package events

import (
	"context"
	"errors"
	"sync"

	"github.com/plexusone/productgraph/pkg/schema"
)

// MultiPublisher dispatches events to multiple publishers in parallel.
type MultiPublisher struct {
	publishers []Publisher
	mu         sync.RWMutex
}

// NewMultiPublisher creates a new multi-publisher with the given publishers.
func NewMultiPublisher(publishers ...Publisher) *MultiPublisher {
	return &MultiPublisher{
		publishers: publishers,
	}
}

// Add adds a publisher to the multi-publisher.
func (m *MultiPublisher) Add(p Publisher) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishers = append(m.publishers, p)
}

// Publish dispatches events to all publishers in parallel.
// Returns an error if any publisher fails, but does not stop other publishers.
func (m *MultiPublisher) Publish(ctx context.Context, events []schema.Event) error {
	m.mu.RLock()
	publishers := m.publishers
	m.mu.RUnlock()

	if len(publishers) == 0 {
		return nil
	}

	if len(publishers) == 1 {
		return publishers[0].Publish(ctx, events)
	}

	// Parallel dispatch
	var wg sync.WaitGroup
	errs := make([]error, len(publishers))

	for i, p := range publishers {
		wg.Add(1)
		go func(idx int, pub Publisher) {
			defer wg.Done()
			errs[idx] = pub.Publish(ctx, events)
		}(i, p)
	}

	wg.Wait()

	// Collect errors
	var combined []error
	for _, err := range errs {
		if err != nil {
			combined = append(combined, err)
		}
	}

	if len(combined) == 0 {
		return nil
	}
	if len(combined) == 1 {
		return combined[0]
	}
	return errors.Join(combined...)
}

// Len returns the number of publishers.
func (m *MultiPublisher) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.publishers)
}
