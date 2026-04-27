// Package analytics provides adapters for forwarding ProductGraph events
// to external analytics providers via omnidxi.
package analytics

import (
	"context"

	"github.com/plexusone/omnidxi"
	"github.com/plexusone/productgraph/pkg/schema"
)

// Adapter forwards ProductGraph events to analytics providers.
type Adapter struct {
	tracker omnidxi.Tracker
}

// NewAdapter creates a new analytics adapter with the given tracker.
func NewAdapter(tracker omnidxi.Tracker) *Adapter {
	return &Adapter{tracker: tracker}
}

// Publish implements events.Publisher for analytics forwarding.
func (a *Adapter) Publish(ctx context.Context, events []schema.Event) error {
	for _, e := range events {
		dxiEvent := a.convert(e)
		if err := a.tracker.Track(ctx, dxiEvent); err != nil {
			return err
		}
	}
	return nil
}

// Flush forces the tracker to send any buffered events.
func (a *Adapter) Flush(ctx context.Context) error {
	return a.tracker.Flush(ctx)
}

// Close shuts down the adapter and underlying tracker.
func (a *Adapter) Close() error {
	return a.tracker.Close()
}

// convert transforms a ProductGraph event to an omnidxi event.
func (a *Adapter) convert(e schema.Event) omnidxi.Event {
	return omnidxi.Event{
		ID:         e.EventID,
		Type:       mapEventType(e.EventType),
		Name:       e.EventName,
		UserID:     e.UserID,
		SessionID:  e.SessionID,
		Timestamp:  e.Timestamp,
		Properties: buildProperties(e),
		Context:    buildContext(e),
	}
}

// mapEventType converts ProductGraph event types to omnidxi event types.
func mapEventType(t schema.EventType) omnidxi.EventType {
	switch t {
	case schema.EventTypePageView:
		return omnidxi.EventTypePageView
	case schema.EventTypePageLeave:
		return omnidxi.EventTypePageLeave
	case schema.EventTypeUIClick:
		return omnidxi.EventTypeUIClick
	case schema.EventTypeUIInput:
		return omnidxi.EventTypeUIInput
	case schema.EventTypeUIScroll:
		return omnidxi.EventTypeUIScroll
	case schema.EventTypeUISubmit:
		return omnidxi.EventTypeUISubmit
	case schema.EventTypeStateChange:
		return omnidxi.EventTypeStateChange
	case schema.EventTypeAPIRequest:
		return omnidxi.EventTypeAPIRequest
	case schema.EventTypeAPIResponse:
		return omnidxi.EventTypeAPIResponse
	case schema.EventTypeJourneyStep:
		return omnidxi.EventTypeJourneyStep
	case schema.EventTypeError:
		return omnidxi.EventTypeError
	case schema.EventTypePerformance:
		return omnidxi.EventTypePerformance
	default:
		return omnidxi.EventTypeCustom
	}
}

// buildProperties extracts event-specific properties.
func buildProperties(e schema.Event) map[string]any {
	props := make(map[string]any)

	// Project/org context
	if e.ProjectID != "" {
		props["project_id"] = e.ProjectID
	}
	if e.OrgID != "" {
		props["org_id"] = e.OrgID
	}

	// Sequence
	if e.Sequence > 0 {
		props["sequence"] = e.Sequence
	}

	// UI context
	if e.UIComponentName != "" {
		props["component_name"] = e.UIComponentName
	}
	if e.UIComponentPath != "" {
		props["component_path"] = e.UIComponentPath
	}
	if e.UIComponentType != "" {
		props["component_type"] = e.UIComponentType
	}
	if e.UIAction != "" {
		props["action"] = e.UIAction
	}
	if e.UIElement != "" {
		props["element"] = e.UIElement
	}
	if e.UIElementText != "" {
		props["element_text"] = e.UIElementText
	}
	if e.UIViewport != "" {
		props["viewport"] = e.UIViewport
	}
	if e.UIScrollPosition > 0 {
		props["scroll_position"] = e.UIScrollPosition
	}

	// State changes
	if e.UIStateKey != "" {
		props["state_key"] = e.UIStateKey
	}
	if e.UIStateBefore != "" {
		props["state_before"] = e.UIStateBefore
	}
	if e.UIStateAfter != "" {
		props["state_after"] = e.UIStateAfter
	}
	if e.UIStateChangeType != "" {
		props["state_change_type"] = e.UIStateChangeType
	}

	// Journey context
	if e.JourneyID != "" {
		props["journey_id"] = e.JourneyID
	}
	if e.JourneyStepID != "" {
		props["journey_step_id"] = e.JourneyStepID
	}
	if e.JourneyStepName != "" {
		props["journey_step_name"] = e.JourneyStepName
	}
	if e.ConversionStatus != "" {
		props["conversion_status"] = e.ConversionStatus
	}

	// API tracking
	if e.APIMethod != "" {
		props["api_method"] = e.APIMethod
	}
	if e.APIPath != "" {
		props["api_path"] = e.APIPath
	}
	if e.APIStatusCode > 0 {
		props["api_status_code"] = e.APIStatusCode
	}
	if e.APIDurationMs > 0 {
		props["api_duration_ms"] = e.APIDurationMs
	}

	// Error tracking
	if e.ErrorType != "" {
		props["error_type"] = e.ErrorType
	}
	if e.ErrorMessage != "" {
		props["error_message"] = e.ErrorMessage
	}
	if e.ErrorStack != "" {
		props["error_stack"] = e.ErrorStack
	}
	if e.ErrorComponent != "" {
		props["error_component"] = e.ErrorComponent
	}

	// Performance
	if e.PerformanceLCP > 0 {
		props["lcp_ms"] = e.PerformanceLCP
	}
	if e.PerformanceFID > 0 {
		props["fid_ms"] = e.PerformanceFID
	}
	if e.PerformanceCLS > 0 {
		props["cls"] = e.PerformanceCLS
	}
	if e.PerformanceTTFB > 0 {
		props["ttfb_ms"] = e.PerformanceTTFB
	}

	// Snapshot
	if e.SnapshotURL != "" {
		props["snapshot_url"] = e.SnapshotURL
	}
	if e.SnapshotViewport != "" {
		props["snapshot_viewport"] = e.SnapshotViewport
	}

	// Duration
	if e.DurationMs > 0 {
		props["duration_ms"] = e.DurationMs
	}

	// Custom metadata
	for k, v := range e.Metadata {
		props[k] = v
	}

	return props
}

// buildContext extracts page/device context.
func buildContext(e schema.Event) *omnidxi.EventContext {
	ctx := &omnidxi.EventContext{
		PagePath:     e.PagePath,
		PageTitle:    e.PageTitle,
		PageURL:      e.PageURL,
		PageReferrer: e.PageReferrer,
	}

	// Only return context if any field is set
	if ctx.PagePath == "" && ctx.PageTitle == "" && ctx.PageURL == "" && ctx.PageReferrer == "" {
		return nil
	}
	return ctx
}
