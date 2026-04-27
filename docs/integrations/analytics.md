# Analytics Integration

ProductGraph integrates with external analytics providers via [omnidxi](https://github.com/plexusone/omnidxi), enabling automatic forwarding of events to Amplitude, Mixpanel, and other DXI (Digital Experience Intelligence) platforms.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Frontend                                        │
│   ┌──────────────────────────────────────────────────────────────────────┐  │
│   │                    @coreforge/telemetry                               │  │
│   │  TelemetryProvider → ProductGraphAdapter → POST /v1/events           │  │
│   └──────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         ProductGraph Service                                 │
│                                                                              │
│  ┌────────────────┐    ┌──────────────────────────────────────────────────┐│
│  │ Event Handler  │───▶│              MultiPublisher                       ││
│  │ POST /v1/events│    │                                                   ││
│  └────────────────┘    │  ┌─────────────────┐  ┌─────────────────────────┐││
│                        │  │ Memory Publisher│  │   Analytics Adapter     │││
│                        │  │   (Storage)     │  │      (omnidxi)          │││
│                        │  └─────────────────┘  └───────────┬─────────────┘││
│                        └───────────────────────────────────┼──────────────┘│
└────────────────────────────────────────────────────────────┼────────────────┘
                                                             │
                                                             ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           omnidxi MultiTracker                               │
│                                                                              │
│            ┌───────────────────────┐    ┌───────────────────────┐           │
│            │    Amplitude Tracker  │    │    Mixpanel Tracker   │           │
│            │  (omni-amplitude)     │    │   (omni-mixpanel)     │           │
│            └───────────┬───────────┘    └───────────┬───────────┘           │
└────────────────────────┼────────────────────────────┼───────────────────────┘
                         │                            │
                         ▼                            ▼
              ┌──────────────────┐         ┌──────────────────┐
              │    Amplitude     │         │     Mixpanel     │
              │   analytics.go   │         │   mixpanel-go    │
              └──────────────────┘         └──────────────────┘
```

## Benefits

- **Backend-first**: Events are sent server-side, bypassing ad blockers
- **Single source of truth**: All events flow through ProductGraph
- **Unified schema**: OTel-compatible events translate to vendor formats
- **Multi-provider**: Send to multiple analytics providers simultaneously
- **No frontend changes**: Existing `@coreforge/telemetry` integration works as-is

## Configuration

Analytics integration is configured via environment variables:

```bash
# Enable analytics integration
export ANALYTICS_ENABLED=true

# Amplitude configuration
export AMPLITUDE_ENABLED=true
export AMPLITUDE_API_KEY=your-amplitude-api-key

# Mixpanel configuration
export MIXPANEL_ENABLED=true
export MIXPANEL_TOKEN=your-mixpanel-project-token
```

### Configuration Options

| Variable | Description | Default |
|----------|-------------|---------|
| `ANALYTICS_ENABLED` | Enable analytics integration | `false` |
| `AMPLITUDE_ENABLED` | Enable Amplitude provider | `false` |
| `AMPLITUDE_API_KEY` | Amplitude API key | - |
| `MIXPANEL_ENABLED` | Enable Mixpanel provider | `false` |
| `MIXPANEL_TOKEN` | Mixpanel project token | - |

## Event Mapping

ProductGraph events are automatically mapped to provider-specific formats:

### Event Types

| ProductGraph Event | Amplitude | Mixpanel |
|--------------------|-----------|----------|
| `page.view` | Page View | Page View |
| `ui.click` | User Action | Track |
| `ui.input` | User Action | Track |
| `ui.scroll` | User Action | Track |
| `state.change` | State Change | Track |
| `journey.step` | Journey Event | Track |
| `error` | Error Event | Track |
| `performance` | Performance | Track |

### Properties

All event properties are preserved and mapped:

```go
// ProductGraph Event
{
    "event.type": "ui.click",
    "event.name": "signup_button_clicked",
    "user.id": "user-123",
    "session.id": "sess-abc",
    "page.path": "/signup",
    "ui.component.name": "SignupButton",
    "ui.action": "click"
}

// → Amplitude Event
{
    "event_type": "signup_button_clicked",
    "user_id": "user-123",
    "session_id": 1234567890,  // hashed
    "event_properties": {
        "page_path": "/signup",
        "component_name": "SignupButton",
        "action": "click"
    }
}

// → Mixpanel Event
{
    "event": "signup_button_clicked",
    "distinct_id": "user-123",
    "properties": {
        "session_id": "sess-abc",
        "page_path": "/signup",
        "component_name": "SignupButton",
        "action": "click"
    }
}
```

## Usage

### Running with Analytics

```bash
# Start with Amplitude only
ANALYTICS_ENABLED=true \
AMPLITUDE_ENABLED=true \
AMPLITUDE_API_KEY=your-key \
go run ./cmd/ingestion

# Start with both providers
ANALYTICS_ENABLED=true \
AMPLITUDE_ENABLED=true \
AMPLITUDE_API_KEY=amp-key \
MIXPANEL_ENABLED=true \
MIXPANEL_TOKEN=mp-token \
go run ./cmd/ingestion
```

### Docker Compose

```yaml
services:
  ingestion:
    build: .
    environment:
      - PORT=8080
      - ANALYTICS_ENABLED=true
      - AMPLITUDE_ENABLED=true
      - AMPLITUDE_API_KEY=${AMPLITUDE_API_KEY}
      - MIXPANEL_ENABLED=true
      - MIXPANEL_TOKEN=${MIXPANEL_TOKEN}
    ports:
      - "8080:8080"
```

### Kubernetes

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: analytics-secrets
type: Opaque
stringData:
  amplitude-api-key: "your-amplitude-key"
  mixpanel-token: "your-mixpanel-token"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: productgraph-ingestion
spec:
  template:
    spec:
      containers:
        - name: ingestion
          env:
            - name: ANALYTICS_ENABLED
              value: "true"
            - name: AMPLITUDE_ENABLED
              value: "true"
            - name: AMPLITUDE_API_KEY
              valueFrom:
                secretKeyRef:
                  name: analytics-secrets
                  key: amplitude-api-key
            - name: MIXPANEL_ENABLED
              value: "true"
            - name: MIXPANEL_TOKEN
              valueFrom:
                secretKeyRef:
                  name: analytics-secrets
                  key: mixpanel-token
```

## Troubleshooting

### Events Not Appearing in Analytics

1. **Check configuration**: Ensure `ANALYTICS_ENABLED=true` and provider is enabled
2. **Verify API keys**: Check that keys are valid and have write permissions
3. **Enable debug logging**: Run with `DEBUG=true` to see event dispatch logs
4. **Check health endpoint**: `GET /health` should return `{"status":"ok"}`

### Debug Logging

```bash
DEBUG=true ANALYTICS_ENABLED=true ... go run ./cmd/ingestion
```

Log output:
```
{"level":"info","msg":"amplitude provider enabled"}
{"level":"info","msg":"mixpanel provider enabled"}
{"level":"info","msg":"analytics enabled","publishers":2}
```

## Dependencies

| Package | Version | Description |
|---------|---------|-------------|
| `omnidxi` | v0.1.0 | Batteries-included DXI client |
| `omni-amplitude` | v0.1.0 | Amplitude adapter (indirect) |
| `omni-mixpanel` | v0.1.0 | Mixpanel adapter (indirect) |
| `omnidxi-core` | v0.1.0 | Core interfaces (indirect) |

## Related Documentation

- [Event Ingestion API](../api/events.md)
- [Architecture Overview](../architecture/overview.md)
- [v0.2.0 Design](../design/v0.2.0/IDEATION.md)
- [omnidxi Documentation](https://github.com/plexusone/omnidxi)
