package events

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/plexusone/productgraph/pkg/schema"
)

func TestHandler_ServeHTTP(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	publisher := NewMemoryPublisher(logger)
	handler := NewHandler(logger, publisher)

	tests := []struct {
		name           string
		method         string
		body           any
		wantStatus     int
		wantAccepted   int
		wantRejected   int
	}{
		{
			name:   "valid single event",
			method: http.MethodPost,
			body: schema.EventBatch{
				Events: []schema.Event{
					{
						EventID:   "evt_001",
						ProjectID: "proj_test",
						SessionID: "sess_001",
						EventType: schema.EventTypePageView,
						Timestamp: time.Now(),
						PagePath:  "/home",
					},
				},
			},
			wantStatus:   http.StatusAccepted,
			wantAccepted: 1,
			wantRejected: 0,
		},
		{
			name:   "valid batch",
			method: http.MethodPost,
			body: schema.EventBatch{
				Events: []schema.Event{
					{
						EventID:   "evt_002",
						ProjectID: "proj_test",
						SessionID: "sess_001",
						EventType: schema.EventTypePageView,
						Timestamp: time.Now(),
					},
					{
						EventID:   "evt_003",
						ProjectID: "proj_test",
						SessionID: "sess_001",
						EventType: schema.EventTypeUIClick,
						Timestamp: time.Now(),
					},
				},
			},
			wantStatus:   http.StatusAccepted,
			wantAccepted: 2,
			wantRejected: 0,
		},
		{
			name:   "partial valid batch",
			method: http.MethodPost,
			body: schema.EventBatch{
				Events: []schema.Event{
					{
						EventID:   "evt_004",
						ProjectID: "proj_test",
						SessionID: "sess_001",
						EventType: schema.EventTypePageView,
						Timestamp: time.Now(),
					},
					{
						// Missing required fields
						EventID: "evt_005",
					},
				},
			},
			wantStatus:   http.StatusAccepted,
			wantAccepted: 1,
			wantRejected: 1,
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
			body:       nil,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "invalid json",
			method:     http.MethodPost,
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "empty batch",
			method: http.MethodPost,
			body: schema.EventBatch{
				Events: []schema.Event{},
			},
			wantStatus:   http.StatusAccepted,
			wantAccepted: 0,
			wantRejected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publisher.Clear()

			var body bytes.Buffer
			if tt.body != nil {
				if s, ok := tt.body.(string); ok {
					body.WriteString(s)
				} else {
					if err := json.NewEncoder(&body).Encode(tt.body); err != nil {
						t.Fatalf("failed to encode body: %v", err)
					}
				}
			}

			req := httptest.NewRequest(tt.method, "/v1/events", &body)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusAccepted {
				var resp schema.IngestResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if resp.Accepted != tt.wantAccepted {
					t.Errorf("accepted = %d, want %d", resp.Accepted, tt.wantAccepted)
				}
				if resp.Rejected != tt.wantRejected {
					t.Errorf("rejected = %d, want %d", resp.Rejected, tt.wantRejected)
				}
				if tt.wantRejected > 0 && len(resp.Errors) != tt.wantRejected {
					t.Errorf("errors count = %d, want %d", len(resp.Errors), tt.wantRejected)
				}
			}
		})
	}
}

func TestHandler_BatchSizeLimit(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	publisher := NewMemoryPublisher(logger)
	handler := NewHandler(logger, publisher)

	// Create batch exceeding limit
	events := make([]schema.Event, 1001)
	for i := range events {
		events[i] = schema.Event{
			EventID:   "evt_" + string(rune(i)),
			ProjectID: "proj_test",
			SessionID: "sess_001",
			EventType: schema.EventTypePageView,
			Timestamp: time.Now(),
		}
	}

	body, _ := json.Marshal(schema.EventBatch{Events: events})
	req := httptest.NewRequest(http.MethodPost, "/v1/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d for batch exceeding limit", rec.Code, http.StatusBadRequest)
	}
}

func TestMemoryPublisher(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	publisher := NewMemoryPublisher(logger)

	if publisher.Count() != 0 {
		t.Errorf("initial count = %d, want 0", publisher.Count())
	}

	events := []schema.Event{
		{EventID: "evt_1"},
		{EventID: "evt_2"},
	}

	if err := publisher.Publish(nil, events); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	if publisher.Count() != 2 {
		t.Errorf("count = %d, want 2", publisher.Count())
	}

	stored := publisher.Events()
	if len(stored) != 2 {
		t.Errorf("stored events = %d, want 2", len(stored))
	}

	publisher.Clear()
	if publisher.Count() != 0 {
		t.Errorf("count after clear = %d, want 0", publisher.Count())
	}
}
