package schema

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEventValidate(t *testing.T) {
	tests := []struct {
		name    string
		event   Event
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid event",
			event: Event{
				EventID:   "evt_123",
				ProductID: "proj_abc",
				SessionID: "sess_xyz",
				EventType: EventTypePageView,
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing event_id",
			event: Event{
				ProductID: "proj_abc",
				SessionID: "sess_xyz",
				EventType: EventTypePageView,
				Timestamp: time.Now(),
			},
			wantErr: true,
			errMsg:  "event_id: required",
		},
		{
			name: "missing product_id",
			event: Event{
				EventID:   "evt_123",
				SessionID: "sess_xyz",
				EventType: EventTypePageView,
				Timestamp: time.Now(),
			},
			wantErr: true,
			errMsg:  "product_id: required",
		},
		{
			name: "missing session_id",
			event: Event{
				EventID:   "evt_123",
				ProductID: "proj_abc",
				EventType: EventTypePageView,
				Timestamp: time.Now(),
			},
			wantErr: true,
			errMsg:  "session.id: required",
		},
		{
			name: "missing event_type",
			event: Event{
				EventID:   "evt_123",
				ProductID: "proj_abc",
				SessionID: "sess_xyz",
				Timestamp: time.Now(),
			},
			wantErr: true,
			errMsg:  "event.type: required",
		},
		{
			name: "missing timestamp",
			event: Event{
				EventID:   "evt_123",
				ProductID: "proj_abc",
				SessionID: "sess_xyz",
				EventType: EventTypePageView,
			},
			wantErr: true,
			errMsg:  "event.timestamp: required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error, got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Validate() error = %q, want %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestEventJSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)

	event := Event{
		EventID:         "evt_123",
		ProductID:       "proj_abc",
		SessionID:       "sess_xyz",
		UserID:          "user_456",
		EventType:       EventTypeUIClick,
		EventName:       "button_click",
		Timestamp:       now,
		Sequence:        42,
		PagePath:        "/checkout",
		PageTitle:       "Checkout",
		UIComponentName: "CheckoutButton",
		UIComponentPath: "App/Checkout/CheckoutButton",
		UIAction:        "click",
		JourneyID:       "journey_001",
		JourneyStepID:   "step_003",
		Metadata: map[string]any{
			"button_text": "Complete Purchase",
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(&event)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify fields
	if decoded.EventID != event.EventID {
		t.Errorf("EventID = %q, want %q", decoded.EventID, event.EventID)
	}
	if decoded.SessionID != event.SessionID {
		t.Errorf("SessionID = %q, want %q", decoded.SessionID, event.SessionID)
	}
	if decoded.EventType != event.EventType {
		t.Errorf("EventType = %q, want %q", decoded.EventType, event.EventType)
	}
	if decoded.UIComponentPath != event.UIComponentPath {
		t.Errorf("UIComponentPath = %q, want %q", decoded.UIComponentPath, event.UIComponentPath)
	}
	if decoded.JourneyID != event.JourneyID {
		t.Errorf("JourneyID = %q, want %q", decoded.JourneyID, event.JourneyID)
	}

	// Verify timestamp roundtrip (within reasonable tolerance)
	if decoded.Timestamp.Sub(now).Abs() > time.Second {
		t.Errorf("Timestamp = %v, want %v", decoded.Timestamp, now)
	}
}

func TestEventBatch(t *testing.T) {
	batch := EventBatch{
		Events: []Event{
			{
				EventID:   "evt_1",
				ProductID: "proj_abc",
				SessionID: "sess_xyz",
				EventType: EventTypePageView,
				Timestamp: time.Now(),
			},
			{
				EventID:   "evt_2",
				ProductID: "proj_abc",
				SessionID: "sess_xyz",
				EventType: EventTypeUIClick,
				Timestamp: time.Now(),
			},
		},
	}

	data, err := json.Marshal(batch)
	if err != nil {
		t.Fatalf("Marshal batch failed: %v", err)
	}

	var decoded EventBatch
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal batch failed: %v", err)
	}

	if len(decoded.Events) != 2 {
		t.Errorf("Events count = %d, want 2", len(decoded.Events))
	}
}
