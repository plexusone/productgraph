package events

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/plexusone/productgraph/pkg/schema"
)

type mockPublisher struct {
	events    []schema.Event
	callCount atomic.Int32
	err       error
	delay     time.Duration
}

func (m *mockPublisher) Publish(ctx context.Context, events []schema.Event) error {
	m.callCount.Add(1)
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	if m.err != nil {
		return m.err
	}
	m.events = append(m.events, events...)
	return nil
}

func TestMultiPublisher_Publish(t *testing.T) {
	t.Run("single publisher", func(t *testing.T) {
		pub := &mockPublisher{}
		multi := NewMultiPublisher(pub)

		events := []schema.Event{{EventID: "1"}, {EventID: "2"}}
		err := multi.Publish(context.Background(), events)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(pub.events) != 2 {
			t.Errorf("expected 2 events, got %d", len(pub.events))
		}
	})

	t.Run("multiple publishers", func(t *testing.T) {
		pub1 := &mockPublisher{}
		pub2 := &mockPublisher{}
		pub3 := &mockPublisher{}
		multi := NewMultiPublisher(pub1, pub2, pub3)

		events := []schema.Event{{EventID: "1"}}
		err := multi.Publish(context.Background(), events)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if pub1.callCount.Load() != 1 {
			t.Errorf("pub1 expected 1 call, got %d", pub1.callCount.Load())
		}
		if pub2.callCount.Load() != 1 {
			t.Errorf("pub2 expected 1 call, got %d", pub2.callCount.Load())
		}
		if pub3.callCount.Load() != 1 {
			t.Errorf("pub3 expected 1 call, got %d", pub3.callCount.Load())
		}
	})

	t.Run("parallel execution", func(t *testing.T) {
		// Each publisher has 50ms delay
		pub1 := &mockPublisher{delay: 50 * time.Millisecond}
		pub2 := &mockPublisher{delay: 50 * time.Millisecond}
		pub3 := &mockPublisher{delay: 50 * time.Millisecond}
		multi := NewMultiPublisher(pub1, pub2, pub3)

		events := []schema.Event{{EventID: "1"}}
		start := time.Now()
		err := multi.Publish(context.Background(), events)
		elapsed := time.Since(start)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// If sequential, would take 150ms+. Parallel should be ~50ms.
		if elapsed > 100*time.Millisecond {
			t.Errorf("expected parallel execution (~50ms), took %v", elapsed)
		}
	})

	t.Run("single error", func(t *testing.T) {
		expectedErr := errors.New("publish failed")
		pub1 := &mockPublisher{}
		pub2 := &mockPublisher{err: expectedErr}
		multi := NewMultiPublisher(pub1, pub2)

		events := []schema.Event{{EventID: "1"}}
		err := multi.Publish(context.Background(), events)

		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}

		// Both publishers should still be called
		if pub1.callCount.Load() != 1 {
			t.Errorf("pub1 should still be called")
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		pub1 := &mockPublisher{err: err1}
		pub2 := &mockPublisher{err: err2}
		multi := NewMultiPublisher(pub1, pub2)

		events := []schema.Event{{EventID: "1"}}
		err := multi.Publish(context.Background(), events)

		if err == nil {
			t.Fatal("expected error")
		}

		// Should contain both errors
		if !errors.Is(err, err1) {
			t.Errorf("error should contain err1")
		}
		if !errors.Is(err, err2) {
			t.Errorf("error should contain err2")
		}
	})

	t.Run("no publishers", func(t *testing.T) {
		multi := NewMultiPublisher()

		events := []schema.Event{{EventID: "1"}}
		err := multi.Publish(context.Background(), events)

		if err != nil {
			t.Errorf("expected nil error for empty publishers, got %v", err)
		}
	})
}

func TestMultiPublisher_Add(t *testing.T) {
	multi := NewMultiPublisher()
	if multi.Len() != 0 {
		t.Errorf("expected 0 publishers, got %d", multi.Len())
	}

	multi.Add(&mockPublisher{})
	if multi.Len() != 1 {
		t.Errorf("expected 1 publisher, got %d", multi.Len())
	}

	multi.Add(&mockPublisher{})
	if multi.Len() != 2 {
		t.Errorf("expected 2 publishers, got %d", multi.Len())
	}
}
