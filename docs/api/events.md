# Event Ingestion API

The Event Ingestion API allows you to send telemetry events from your application to ProductGraph.

## Endpoint

```
POST /v1/events
```

## Authentication

Include your project API key in the request header:

```
X-PG-API-Key: your-api-key
```

## Request

### Headers

| Header | Required | Description |
|--------|----------|-------------|
| `Content-Type` | Yes | Must be `application/json` |
| `X-PG-API-Key` | Yes | Your project API key |

### Body

```json
{
  "events": [
    {
      "event_id": "string (required)",
      "project_id": "string (required)",
      "session.id": "string (required)",
      "event.type": "string (required)",
      "event.timestamp": "ISO8601 (required)",
      "page.path": "string",
      "ui.component.name": "string",
      "metadata": {}
    }
  ]
}
```

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `event_id` | string | Unique identifier for this event |
| `project_id` | string | Your project identifier |
| `session.id` | string | User session identifier |
| `event.type` | string | Event type (see below) |
| `event.timestamp` | string | ISO8601 timestamp |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `user.id` | string | User identifier |
| `event.name` | string | Human-readable event name |
| `event.sequence` | integer | Event sequence number in session |
| `page.path` | string | Page URL path |
| `page.title` | string | Page title |
| `page.url` | string | Full page URL |
| `page.referrer` | string | Referrer URL |
| `ui.component.name` | string | Component name |
| `ui.component.path` | string | Component hierarchy path |
| `ui.component.type` | string | Component type |
| `ui.action` | string | User action (click, input, etc.) |
| `ui.element` | string | DOM element identifier |
| `ui.element.text` | string | Element text content |
| `ui.state.key` | string | State key that changed |
| `ui.state.before` | string | State value before change (JSON) |
| `ui.state.after` | string | State value after change (JSON) |
| `gen_ai.journey.id` | string | Journey identifier |
| `gen_ai.journey.step.id` | string | Journey step identifier |
| `gen_ai.journey.step.name` | string | Journey step name |
| `api.method` | string | HTTP method (GET, POST, etc.) |
| `api.path` | string | API endpoint path |
| `api.status_code` | integer | HTTP status code |
| `api.duration_ms` | integer | API call duration in ms |
| `error.type` | string | Error type/class |
| `error.message` | string | Error message |
| `error.stack` | string | Error stack trace |
| `duration_ms` | integer | Event duration in ms |
| `metadata` | object | Custom key-value pairs |

## Event Types

| Type | Description |
|------|-------------|
| `page.view` | User navigated to a page |
| `page.leave` | User left a page |
| `ui.click` | User clicked an element |
| `ui.input` | User typed in an input field |
| `ui.scroll` | User scrolled the page |
| `ui.focus` | Element received focus |
| `ui.blur` | Element lost focus |
| `ui.submit` | Form was submitted |
| `state.change` | Application state changed |
| `api.request` | API call was initiated |
| `api.response` | API call completed |
| `journey.step` | User reached a journey step |
| `error` | An error occurred |
| `performance` | Performance metric (Web Vitals) |
| `custom` | Custom event type |

## Response

### Success (202 Accepted)

```json
{
  "accepted": 2,
  "rejected": 0,
  "errors": []
}
```

### Partial Success (202 Accepted)

```json
{
  "accepted": 1,
  "rejected": 1,
  "errors": [
    {
      "index": 1,
      "error": "event_id: required"
    }
  ]
}
```

### Error Responses

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid JSON or batch too large |
| `401 Unauthorized` | Missing or invalid API key |
| `405 Method Not Allowed` | Must use POST |
| `500 Internal Server Error` | Server error |

## Limits

| Limit | Value |
|-------|-------|
| Max events per batch | 1,000 |
| Max request body size | 10 MB |
| Max event size | 100 KB |

## Examples

### Page View

```bash
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -H "X-PG-API-Key: your-api-key" \
  -d '{
    "events": [{
      "event_id": "evt_001",
      "project_id": "proj_demo",
      "session.id": "sess_abc123",
      "event.type": "page.view",
      "event.timestamp": "2024-01-15T10:30:00Z",
      "page.path": "/products",
      "page.title": "Products"
    }]
  }'
```

### Button Click

```bash
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -H "X-PG-API-Key: your-api-key" \
  -d '{
    "events": [{
      "event_id": "evt_002",
      "project_id": "proj_demo",
      "session.id": "sess_abc123",
      "event.type": "ui.click",
      "event.timestamp": "2024-01-15T10:30:15Z",
      "page.path": "/products",
      "ui.component.name": "AddToCartButton",
      "ui.component.path": "ProductPage/ProductCard/AddToCartButton",
      "ui.action": "click",
      "ui.element.text": "Add to Cart"
    }]
  }'
```

### State Change

```bash
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -H "X-PG-API-Key: your-api-key" \
  -d '{
    "events": [{
      "event_id": "evt_003",
      "project_id": "proj_demo",
      "session.id": "sess_abc123",
      "event.type": "state.change",
      "event.timestamp": "2024-01-15T10:30:16Z",
      "page.path": "/products",
      "ui.state.key": "cart.items",
      "ui.state.before": "[]",
      "ui.state.after": "[{\"id\":\"prod_123\",\"qty\":1}]"
    }]
  }'
```

### Journey Step

```bash
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -H "X-PG-API-Key: your-api-key" \
  -d '{
    "events": [{
      "event_id": "evt_004",
      "project_id": "proj_demo",
      "session.id": "sess_abc123",
      "event.type": "journey.step",
      "event.timestamp": "2024-01-15T10:31:00Z",
      "page.path": "/checkout",
      "gen_ai.journey.id": "checkout_flow",
      "gen_ai.journey.step.id": "step_payment",
      "gen_ai.journey.step.name": "Enter Payment Details"
    }]
  }'
```

### Batch Events

```bash
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -H "X-PG-API-Key: your-api-key" \
  -d '{
    "events": [
      {
        "event_id": "evt_005",
        "project_id": "proj_demo",
        "session.id": "sess_abc123",
        "event.type": "page.view",
        "event.timestamp": "2024-01-15T10:30:00Z",
        "page.path": "/checkout"
      },
      {
        "event_id": "evt_006",
        "project_id": "proj_demo",
        "session.id": "sess_abc123",
        "event.type": "ui.input",
        "event.timestamp": "2024-01-15T10:30:30Z",
        "page.path": "/checkout",
        "ui.component.name": "EmailInput",
        "ui.action": "input"
      },
      {
        "event_id": "evt_007",
        "project_id": "proj_demo",
        "session.id": "sess_abc123",
        "event.type": "ui.submit",
        "event.timestamp": "2024-01-15T10:31:00Z",
        "page.path": "/checkout",
        "ui.component.name": "CheckoutForm",
        "ui.action": "submit"
      }
    ]
  }'
```

## OTel Semantic Conventions

ProductGraph uses OpenTelemetry-compatible field names:

| Namespace | Purpose |
|-----------|---------|
| `session.*` | User session tracking |
| `page.*` | Page context |
| `ui.*` | UI interactions |
| `ui.state.*` | State changes |
| `gen_ai.journey.*` | Journey tracking |
| `api.*` | API call tracking |
| `error.*` | Error details |
| `performance.*` | Web vitals |

## Analytics Forwarding

When analytics integration is enabled, events are automatically forwarded to configured providers (Amplitude, Mixpanel) in addition to being stored in ProductGraph.

```
POST /v1/events
      │
      ├──▶ Storage (PostgreSQL)
      │
      └──▶ Analytics (if enabled)
              │
              ├──▶ Amplitude
              └──▶ Mixpanel
```

This forwarding is transparent to the client. No changes are required to your event payloads.

See the [Analytics Integration Guide](../integrations/analytics.md) for configuration details.

## Health Endpoints

### Health Check

```
GET /health
```

Response:
```json
{"status":"ok"}
```

### Readiness Check

```
GET /ready
```

Response:
```json
{"status":"ready"}
```
