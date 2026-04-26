# ProductGraph

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Docs][docs-mkdoc-svg]][docs-mkdoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/productgraph/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/productgraph/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/productgraph/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/productgraph/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/productgraph/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/productgraph/actions/workflows/go-sast-codeql.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/plexusone/productgraph
 [goreport-url]: https://goreportcard.com/report/github.com/plexusone/productgraph
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/productgraph
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/productgraph
 [docs-mkdoc-svg]: https://img.shields.io/badge/docs-guide-blue.svg
 [docs-mkdoc-url]: https://plexusone.github.io/productgraph
 [viz-svg]: https://img.shields.io/badge/repo-visualization-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fproductgraph
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/productgraph/blob/main/LICENSE

AI-native runtime product intelligence platform combining design visualization, analytics, and observability.

## Overview

ProductGraph provides unified insights into user journeys by combining:

- **Design Canvas** - Figma-like journey visualization with zoomable screenshots and state inspection
- **Analytics Dashboard** - Mixpanel-like funnels, cohorts, and real-time metrics
- **Session Replay** - Step-by-step inspection of user sessions with state diffs

## System Architecture

The PoC architecture uses a single PostgreSQL database with Row-Level Security (RLS) for multi-tenancy. This scales to ~1000 paying users before requiring additional data stores.

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              PRODUCTGRAPH PLATFORM                              │
│                                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌─────────────────────┐  │
│  │Design Canvas │  │  Analytics   │  │   Session    │  │     MCP Server      │  │
│  │(React Flow)  │  │  Dashboard   │  │    Replay    │  │  (AI Agent Access)  │  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────────┬──────────┘  │
│         │                 │                 │                     │             │
│         └─────────────────┴─────────────────┴─────────────────────┘             │
│                                      │                                          │
│                           ┌──────────┴──────────┐                               │
│                           │     GraphQL API     │                               │
│                           └──────────┬──────────┘                               │
│                                      │                                          │
│                      ┌───────────────┴───────────────┐                          │
│                      │   ProductGraph Service (Go)   │                          │
│                      │                               │                          │
│                      │  ┌─────────────────────────┐  │                          │
│                      │  │   Ent ORM + RLS         │  │                          │
│                      │  │   Multi-tenant queries  │  │                          │
│                      │  └───────────┬─────────────┘  │                          │
│                      │              │                │                          │
│                      │  ┌───────────┴─────────────┐  │                          │
│                      │  │  Event Ingestion API    │  │                          │
│                      │  │   POST /v1/events       │  │                          │
│                      │  └─────────────────────────┘  │                          │
│                      └───────────────┬───────────────┘                          │
└──────────────────────────────────────┼──────────────────────────────────────────┘
                                       │
                                       ▼
                      ┌────────────────────────────────┐
                      │      PostgreSQL 16+ (RLS)      │
                      │                                │
                      │  ┌──────────┐  ┌────────────┐  │
                      │  │  Events  │  │  Sessions  │  │
                      │  └──────────┘  └────────────┘  │
                      │  ┌──────────┐  ┌────────────┐  │
                      │  │ Journeys │  │  Projects  │  │
                      │  └──────────┘  └────────────┘  │
                      │  ┌──────────┐  ┌────────────┐  │
                      │  │   Orgs   │  │ Snapshots  │  │
                      │  └──────────┘  └────────────┘  │
                      └────────────────────────────────┘
                                       ▲
            ┌──────────────────────────┼──────────────────────────┐
            │                          │                          │
┌───────────────────────┐  ┌───────────────────────┐  ┌───────────────────────┐
│  @coreforge/telemetry │  │   @omniobserve/core   │  │   OmniObserve-Swift   │
│  + ProductGraph       │  │     (TypeScript)      │  │     (iOS/macOS)       │
│      Adapter          │  │                       │  │                       │
└───────────────────────┘  └───────────────────────┘  └───────────────────────┘
         │                            │                          │
         ▼                            ▼                          ▼
┌───────────────────────┐  ┌───────────────────────┐  ┌───────────────────────┐
│     React Web App     │  │    Node.js Backend    │  │     iOS/macOS App     │
└───────────────────────┘  └───────────────────────┘  └───────────────────────┘
```

## Event Flow

```
┌─────────┐    ┌───────────────────┐    ┌────────────────────────────────────┐
│   SDK   │───▶│  Ingestion API    │───▶│  PostgreSQL (Events + Sessions)   │
│ (Client)│    │  POST /v1/events  │    │  Row-Level Security per org_id    │
└─────────┘    └───────────────────┘    └────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Make

### Start Development Environment

```bash
# Start PostgreSQL
make docker-up

# Run database migrations
make migrate

# Run the ingestion service
make run-ingestion

# Check health
curl http://localhost:8080/health
```

### Send Test Events

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

## Multi-Tenancy

ProductGraph uses PostgreSQL Row-Level Security (RLS) for tenant isolation:

```sql
-- Each request sets the org context
SELECT set_current_org('org-uuid-here');

-- RLS policies automatically filter by org_id
SELECT * FROM events;  -- Only returns current org's events
```

All tenant-scoped tables include an `org_id` column with RLS policies enforcing isolation.

## API Reference

### POST /v1/events

Ingest a batch of events.

**Request:**

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
      "ui.component.path": "string",
      "gen_ai.journey.id": "string",
      "gen_ai.journey.step.id": "string",
      "metadata": {}
    }
  ]
}
```

**Response:**

```json
{
  "accepted": 1,
  "rejected": 0,
  "errors": []
}
```

### GET /health

Health check endpoint.

### GET /ready

Readiness check endpoint.

## Event Types

| Type | Description |
|------|-------------|
| `page.view` | Page navigation |
| `page.leave` | Page exit |
| `ui.click` | Click interaction |
| `ui.input` | Form input |
| `ui.scroll` | Scroll event |
| `ui.submit` | Form submission |
| `state.change` | State mutation |
| `api.request` | API call initiated |
| `api.response` | API call completed |
| `journey.step` | Journey step reached |
| `error` | Error occurred |
| `performance` | Web vitals |

## OTel Semantic Conventions

ProductGraph uses OpenTelemetry-compatible semantic conventions:

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

## Project Structure

```
productgraph/
├── cmd/
│   └── ingestion/        # Event ingestion service
├── ent/
│   └── schema/           # Ent database schemas
├── internal/
│   └── events/           # Event handling
├── pkg/
│   └── schema/           # Event schema (OTel-compatible)
├── deployments/
│   └── docker/           # Docker Compose + init scripts
└── docs/
    └── design/           # PRD, TRD, ideation docs
```

## Related Projects

| Project | Description |
|---------|-------------|
| [CoreForge-Web](https://github.com/grokify/coreforge-web) | React telemetry with ProductGraph adapter |
| [OmniObserve](https://github.com/plexusone/omniobserve) | Go observability with journey semconv |

## Development

```bash
make build          # Build binaries
make test           # Run tests
make lint           # Run linter
make generate       # Generate Ent code
make migrate        # Run database migrations
make docker-up      # Start PostgreSQL
make docker-down    # Stop PostgreSQL
make docker-logs    # View logs
make health         # Check service health
```

## Scaling Beyond PoC

When the PoC architecture hits limits, add specialized data stores:

| Trigger | Solution |
|---------|----------|
| Analytics queries > 2s | Add ClickHouse for event storage |
| Write throughput > 5K/sec | Add Kafka for event streaming |
| Real-time requirements | Add Redis for session state |
| Screenshot storage > 100GB | Add S3/MinIO for object storage |

See [TRD.md](docs/design/architecture/TRD.md) for the full scale architecture.

## Documentation

- [Product Requirements (PRD)](docs/design/overview/PRD.md)
- [Technical Requirements (TRD)](docs/design/architecture/TRD.md)
- [Architecture Scaling Guide](docs/design/architecture/SCALING.md)
- [Implementation Plan](PLAN.md)

## License

Proprietary - All rights reserved.
