// Package schema defines the ProductGraph event schema.
// Events are OTel-compatible with ProductGraph-specific extensions.
package schema

import (
	"encoding/json"
	"time"
)

// EventType represents the type of user interaction event.
type EventType string

const (
	EventTypePageView    EventType = "page.view"
	EventTypePageLeave   EventType = "page.leave"
	EventTypeUIClick     EventType = "ui.click"
	EventTypeUIInput     EventType = "ui.input"
	EventTypeUIScroll    EventType = "ui.scroll"
	EventTypeUIFocus     EventType = "ui.focus"
	EventTypeUIBlur      EventType = "ui.blur"
	EventTypeUISubmit    EventType = "ui.submit"
	EventTypeStateChange EventType = "state.change"
	EventTypeAPIRequest  EventType = "api.request"
	EventTypeAPIResponse EventType = "api.response"
	EventTypeJourneyStep EventType = "journey.step"
	EventTypeError       EventType = "error"
	EventTypePerformance EventType = "performance"
	EventTypeCustom      EventType = "custom"
)

// Event represents a single telemetry event from the frontend.
// Field names follow OTel semantic conventions where applicable.
type Event struct {
	// Identity
	EventID   string `json:"event_id"`
	ProductID string `json:"product_id"`
	SessionID string `json:"session.id"`
	UserID    string `json:"user.id,omitempty"`
	OrgID     string `json:"org.id,omitempty"`

	// Classification (OTel event.* namespace)
	EventType EventType `json:"event.type"`
	EventName string    `json:"event.name"`
	Timestamp time.Time `json:"event.timestamp"`
	Sequence  int64     `json:"event.sequence"`

	// Page context (page.* namespace)
	PagePath     string `json:"page.path,omitempty"`
	PageTitle    string `json:"page.title,omitempty"`
	PageURL      string `json:"page.url,omitempty"`
	PageReferrer string `json:"page.referrer,omitempty"`

	// UI context (ui.* namespace)
	UIComponentName  string  `json:"ui.component.name,omitempty"`
	UIComponentPath  string  `json:"ui.component.path,omitempty"`
	UIComponentType  string  `json:"ui.component.type,omitempty"`
	UIAction         string  `json:"ui.action,omitempty"`
	UIElement        string  `json:"ui.element,omitempty"`
	UIElementText    string  `json:"ui.element.text,omitempty"`
	UIViewport       string  `json:"ui.viewport,omitempty"`
	UIScrollPosition float64 `json:"ui.scroll.position,omitempty"`

	// State tracking (ui.state.* namespace)
	UIStateKey        string `json:"ui.state.key,omitempty"`
	UIStateBefore     string `json:"ui.state.before,omitempty"` // JSON string
	UIStateAfter      string `json:"ui.state.after,omitempty"`  // JSON string
	UIStateChangeType string `json:"ui.state.change_type,omitempty"`

	// Journey context (gen_ai.journey.* namespace)
	JourneyID        string `json:"gen_ai.journey.id,omitempty"`
	JourneyStepID    string `json:"gen_ai.journey.step.id,omitempty"`
	JourneyStepName  string `json:"gen_ai.journey.step.name,omitempty"`
	ConversionStatus string `json:"gen_ai.journey.conversion.status,omitempty"`

	// API tracking (api.* namespace)
	APIMethod     string `json:"api.method,omitempty"`
	APIPath       string `json:"api.path,omitempty"`
	APIStatusCode int    `json:"api.status_code,omitempty"`
	APIDurationMs int64  `json:"api.duration_ms,omitempty"`

	// Error tracking (error.* namespace)
	ErrorType      string `json:"error.type,omitempty"`
	ErrorMessage   string `json:"error.message,omitempty"`
	ErrorStack     string `json:"error.stack,omitempty"`
	ErrorComponent string `json:"error.component,omitempty"`

	// Performance (performance.* namespace)
	PerformanceLCP  float64 `json:"performance.lcp_ms,omitempty"`
	PerformanceFID  float64 `json:"performance.fid_ms,omitempty"`
	PerformanceCLS  float64 `json:"performance.cls,omitempty"`
	PerformanceTTFB float64 `json:"performance.ttfb_ms,omitempty"`

	// Snapshot
	SnapshotURL      string `json:"snapshot.url,omitempty"`
	SnapshotViewport string `json:"snapshot.viewport,omitempty"`

	// Duration
	DurationMs int64 `json:"duration_ms,omitempty"`

	// Custom metadata
	Metadata map[string]any `json:"metadata,omitempty"`
}

// EventBatch represents a batch of events from the SDK.
type EventBatch struct {
	Events []Event `json:"events"`
}

// Validate checks if the event has required fields.
func (e *Event) Validate() error {
	if e.EventID == "" {
		return &ValidationError{Field: "event_id", Message: "required"}
	}
	if e.ProductID == "" {
		return &ValidationError{Field: "product_id", Message: "required"}
	}
	if e.SessionID == "" {
		return &ValidationError{Field: "session.id", Message: "required"}
	}
	if e.EventType == "" {
		return &ValidationError{Field: "event.type", Message: "required"}
	}
	if e.Timestamp.IsZero() {
		return &ValidationError{Field: "event.timestamp", Message: "required"}
	}
	return nil
}

// ValidationError represents a field validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// IngestResponse is the response from the event ingestion endpoint.
type IngestResponse struct {
	Accepted int           `json:"accepted"`
	Rejected int           `json:"rejected"`
	Errors   []IngestError `json:"errors,omitempty"`
}

// IngestError describes an error for a specific event in the batch.
type IngestError struct {
	Index int    `json:"index"`
	Error string `json:"error"`
}

// MarshalJSON implements custom JSON marshaling for Event.
func (e *Event) MarshalJSON() ([]byte, error) {
	type Alias Event
	return json.Marshal(&struct {
		*Alias
		Timestamp string `json:"event.timestamp"`
	}{
		Alias:     (*Alias)(e),
		Timestamp: e.Timestamp.Format(time.RFC3339Nano),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for Event.
func (e *Event) UnmarshalJSON(data []byte) error {
	type Alias Event
	aux := &struct {
		*Alias
		Timestamp string `json:"event.timestamp"`
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Timestamp != "" {
		t, err := time.Parse(time.RFC3339Nano, aux.Timestamp)
		if err != nil {
			// Try other formats
			t, err = time.Parse(time.RFC3339, aux.Timestamp)
			if err != nil {
				return err
			}
		}
		e.Timestamp = t
	}
	return nil
}
