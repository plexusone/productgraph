# Quick Start

Get ProductGraph running locally in under 5 minutes.

## Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Make

## Start the Development Environment

### 1. Clone the Repository

```bash
git clone https://github.com/plexusone/productgraph.git
cd productgraph
```

### 2. Start PostgreSQL

```bash
make docker-up
```

This starts a PostgreSQL 16 container with:

- Database: `productgraph`
- User: `pg`
- Password: `pg`
- Port: `5432`

### 3. Run the Ingestion Service

```bash
make run-ingestion
```

The service starts on `http://localhost:8080` with debug logging enabled.

### 4. Verify Health

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{"status":"ok"}
```

## Send Your First Event

```bash
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "events": [{
      "event_id": "evt_001",
      "project_id": "proj_demo",
      "session.id": "sess_001",
      "event.type": "page.view",
      "event.timestamp": "2024-01-15T10:30:00Z",
      "page.path": "/checkout",
      "page.title": "Checkout"
    }]
  }'
```

Expected response:

```json
{
  "accepted": 1,
  "rejected": 0,
  "errors": []
}
```

## Send a Batch of Events

```bash
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "events": [
      {
        "event_id": "evt_002",
        "project_id": "proj_demo",
        "session.id": "sess_001",
        "event.type": "page.view",
        "event.timestamp": "2024-01-15T10:30:00Z",
        "page.path": "/products",
        "page.title": "Products"
      },
      {
        "event_id": "evt_003",
        "project_id": "proj_demo",
        "session.id": "sess_001",
        "event.type": "ui.click",
        "event.timestamp": "2024-01-15T10:30:15Z",
        "page.path": "/products",
        "ui.component.name": "AddToCartButton",
        "ui.action": "click"
      },
      {
        "event_id": "evt_004",
        "project_id": "proj_demo",
        "session.id": "sess_001",
        "event.type": "page.view",
        "event.timestamp": "2024-01-15T10:30:20Z",
        "page.path": "/cart",
        "page.title": "Shopping Cart"
      }
    ]
  }'
```

## Stop the Environment

```bash
make docker-down
```

To remove all data:

```bash
make docker-clean
```

## Next Steps

- [Installation](installation.md) - Production deployment options
- [Event Ingestion API](../api/events.md) - Full API reference
- [Architecture Overview](../architecture/overview.md) - Understand the system
