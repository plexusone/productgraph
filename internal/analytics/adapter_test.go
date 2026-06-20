package analytics

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/plexusone/omnidxi"
	"github.com/plexusone/productgraph/pkg/schema"
)

// mockTracker implements omnidxi.Tracker for testing.
type mockTracker struct {
	mu       sync.Mutex
	events   []omnidxi.Event
	users    []omnidxi.User
	groups   []omnidxi.Group
	aliases  []omnidxi.Alias
	flushed  int
	closed   bool
	trackErr error
}

func (m *mockTracker) Track(ctx context.Context, event omnidxi.Event) error {
	if m.trackErr != nil {
		return m.trackErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
	return nil
}

func (m *mockTracker) Identify(ctx context.Context, user omnidxi.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users = append(m.users, user)
	return nil
}

func (m *mockTracker) Group(ctx context.Context, group omnidxi.Group) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.groups = append(m.groups, group)
	return nil
}

func (m *mockTracker) Alias(ctx context.Context, alias omnidxi.Alias) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.aliases = append(m.aliases, alias)
	return nil
}

func (m *mockTracker) Flush(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flushed++
	return nil
}

func (m *mockTracker) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *mockTracker) getEvents() []omnidxi.Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]omnidxi.Event{}, m.events...)
}

func TestAdapter_Publish(t *testing.T) {
	mock := &mockTracker{}
	adapter := NewAdapter(mock)

	now := time.Now()
	events := []schema.Event{
		{
			EventID:   "evt-1",
			ProductID: "proj-1",
			SessionID: "sess-1",
			UserID:    "user-1",
			EventType: schema.EventTypePageView,
			EventName: "Home Page View",
			Timestamp: now,
			PagePath:  "/home",
			PageTitle: "Home",
			PageURL:   "https://example.com/home",
		},
		{
			EventID:         "evt-2",
			ProductID:       "proj-1",
			SessionID:       "sess-1",
			UserID:          "user-1",
			EventType:       schema.EventTypeUIClick,
			EventName:       "Button Click",
			Timestamp:       now.Add(time.Second),
			UIComponentName: "SignupButton",
			UIAction:        "click",
		},
	}

	err := adapter.Publish(context.Background(), events)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	tracked := mock.getEvents()
	if len(tracked) != 2 {
		t.Fatalf("expected 2 events, got %d", len(tracked))
	}

	// Verify first event
	if tracked[0].ID != "evt-1" {
		t.Errorf("expected ID evt-1, got %s", tracked[0].ID)
	}
	if tracked[0].Type != omnidxi.EventTypePageView {
		t.Errorf("expected PageView type, got %s", tracked[0].Type)
	}
	if tracked[0].Name != "Home Page View" {
		t.Errorf("expected name 'Home Page View', got %s", tracked[0].Name)
	}
	if tracked[0].UserID != "user-1" {
		t.Errorf("expected UserID user-1, got %s", tracked[0].UserID)
	}
	if tracked[0].Context == nil {
		t.Error("expected Context to be set")
	} else {
		if tracked[0].Context.PagePath != "/home" {
			t.Errorf("expected PagePath /home, got %s", tracked[0].Context.PagePath)
		}
	}

	// Verify second event
	if tracked[1].Type != omnidxi.EventTypeUIClick {
		t.Errorf("expected UIClick type, got %s", tracked[1].Type)
	}
	if tracked[1].Properties["component_name"] != "SignupButton" {
		t.Errorf("expected component_name SignupButton, got %v", tracked[1].Properties["component_name"])
	}
}

func TestAdapter_Flush(t *testing.T) {
	mock := &mockTracker{}
	adapter := NewAdapter(mock)

	err := adapter.Flush(context.Background())
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	if mock.flushed != 1 {
		t.Errorf("expected 1 flush, got %d", mock.flushed)
	}
}

func TestAdapter_Close(t *testing.T) {
	mock := &mockTracker{}
	adapter := NewAdapter(mock)

	err := adapter.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if !mock.closed {
		t.Error("expected tracker to be closed")
	}
}

func TestMapEventType(t *testing.T) {
	tests := []struct {
		input    schema.EventType
		expected omnidxi.EventType
	}{
		{schema.EventTypePageView, omnidxi.EventTypePageView},
		{schema.EventTypePageLeave, omnidxi.EventTypePageLeave},
		{schema.EventTypeUIClick, omnidxi.EventTypeUIClick},
		{schema.EventTypeUIInput, omnidxi.EventTypeUIInput},
		{schema.EventTypeUIScroll, omnidxi.EventTypeUIScroll},
		{schema.EventTypeUISubmit, omnidxi.EventTypeUISubmit},
		{schema.EventTypeStateChange, omnidxi.EventTypeStateChange},
		{schema.EventTypeAPIRequest, omnidxi.EventTypeAPIRequest},
		{schema.EventTypeAPIResponse, omnidxi.EventTypeAPIResponse},
		{schema.EventTypeJourneyStep, omnidxi.EventTypeJourneyStep},
		{schema.EventTypeError, omnidxi.EventTypeError},
		{schema.EventTypePerformance, omnidxi.EventTypePerformance},
		{schema.EventTypeCustom, omnidxi.EventTypeCustom},
		{"unknown.type", omnidxi.EventTypeCustom}, // Unknown maps to custom
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			result := mapEventType(tt.input)
			if result != tt.expected {
				t.Errorf("mapEventType(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuildProperties(t *testing.T) {
	event := schema.Event{
		ProductID:       "proj-1",
		OrgID:           "org-1",
		Sequence:        42,
		UIComponentName: "Header",
		UIAction:        "click",
		JourneyID:       "journey-1",
		JourneyStepName: "Checkout",
		APIMethod:       "POST",
		APIPath:         "/api/users",
		APIStatusCode:   201,
		APIDurationMs:   150,
		ErrorType:       "ValidationError",
		ErrorMessage:    "Invalid email",
		PerformanceLCP:  2500.5,
		DurationMs:      1000,
		Metadata: map[string]any{
			"custom_field": "custom_value",
		},
	}

	props := buildProperties(event)

	// Check standard properties
	if props["product_id"] != "proj-1" {
		t.Errorf("expected product_id proj-1, got %v", props["product_id"])
	}
	if props["org_id"] != "org-1" {
		t.Errorf("expected org_id org-1, got %v", props["org_id"])
	}
	if props["sequence"] != int64(42) {
		t.Errorf("expected sequence 42, got %v", props["sequence"])
	}

	// Check UI properties
	if props["component_name"] != "Header" {
		t.Errorf("expected component_name Header, got %v", props["component_name"])
	}
	if props["action"] != "click" {
		t.Errorf("expected action click, got %v", props["action"])
	}

	// Check journey properties
	if props["journey_id"] != "journey-1" {
		t.Errorf("expected journey_id journey-1, got %v", props["journey_id"])
	}

	// Check API properties
	if props["api_method"] != "POST" {
		t.Errorf("expected api_method POST, got %v", props["api_method"])
	}
	if props["api_status_code"] != 201 {
		t.Errorf("expected api_status_code 201, got %v", props["api_status_code"])
	}

	// Check error properties
	if props["error_type"] != "ValidationError" {
		t.Errorf("expected error_type ValidationError, got %v", props["error_type"])
	}

	// Check performance properties
	if props["lcp_ms"] != 2500.5 {
		t.Errorf("expected lcp_ms 2500.5, got %v", props["lcp_ms"])
	}

	// Check custom metadata
	if props["custom_field"] != "custom_value" {
		t.Errorf("expected custom_field custom_value, got %v", props["custom_field"])
	}
}

func TestBuildContext(t *testing.T) {
	t.Run("with context fields", func(t *testing.T) {
		event := schema.Event{
			PagePath:     "/checkout",
			PageTitle:    "Checkout",
			PageURL:      "https://example.com/checkout",
			PageReferrer: "https://example.com/cart",
		}

		ctx := buildContext(event)
		if ctx == nil {
			t.Fatal("expected context to be non-nil")
		}
		if ctx.PagePath != "/checkout" {
			t.Errorf("expected PagePath /checkout, got %s", ctx.PagePath)
		}
		if ctx.PageTitle != "Checkout" {
			t.Errorf("expected PageTitle Checkout, got %s", ctx.PageTitle)
		}
		if ctx.PageURL != "https://example.com/checkout" {
			t.Errorf("expected PageURL, got %s", ctx.PageURL)
		}
		if ctx.PageReferrer != "https://example.com/cart" {
			t.Errorf("expected PageReferrer, got %s", ctx.PageReferrer)
		}
	})

	t.Run("without context fields", func(t *testing.T) {
		event := schema.Event{
			EventID:   "evt-1",
			ProductID: "proj-1",
			SessionID: "sess-1",
			EventType: schema.EventTypeUIClick,
		}

		ctx := buildContext(event)
		if ctx != nil {
			t.Error("expected context to be nil when no page fields set")
		}
	})
}
