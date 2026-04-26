# ProductGraph

AI-native runtime product intelligence platform combining design visualization, analytics, and observability.

## Overview

ProductGraph provides unified insights into user journeys by combining:

- **Design Canvas** - Figma-like journey visualization with zoomable screenshots and state inspection
- **Analytics Dashboard** - Mixpanel-like funnels, cohorts, and real-time metrics
- **Session Replay** - Step-by-step inspection of user sessions with state diffs

## System Architecture

The Starter architecture uses a single PostgreSQL database with Row-Level Security (RLS) for multi-tenancy. This scales to ~1000 paying users before requiring additional data stores.

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              PRODUCTGRAPH PLATFORM                               │
│                                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │
│  │Design Canvas │  │  Analytics   │  │   Session    │  │     MCP Server       │ │
│  │(React Flow)  │  │  Dashboard   │  │    Replay    │  │  (AI Agent Access)   │ │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────────┬───────────┘ │
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
                      └────────────────────────────────┘
                                       ▲
            ┌──────────────────────────┼──────────────────────────┐
            │                          │                          │
┌───────────────────────┐  ┌───────────────────────┐  ┌───────────────────────┐
│  @coreforge/telemetry │  │   @omniobserve/core   │  │   OmniObserve-Swift   │
│  + ProductGraph       │  │     (TypeScript)      │  │     (iOS/macOS)       │
│      Adapter          │  │                       │  │                       │
└───────────────────────┘  └───────────────────────┘  └───────────────────────┘
```

## Features

### Event Ingestion

- Batch event ingestion via `POST /v1/events`
- OTel-compatible event schema with 40+ fields
- Validation with detailed error responses

### Multi-Tenancy

- PostgreSQL Row-Level Security (RLS) for tenant isolation
- All queries automatically scoped by `org_id`

### Scalable Architecture

- **Starter**: PostgreSQL only (~$150/mo, <1000 users)
- **Growth**: + Kafka, ClickHouse, Redis (~$1,350/mo, 1K-10K users)
- **Scale**: Sharded, multi-region (~$5,500/mo, 10K+ users)

See the [Scaling Guide](architecture/scaling.md) for details.

## Related Projects

| Project | Description |
|---------|-------------|
| [CoreForge-Web](https://github.com/grokify/coreforge-web) | React telemetry with ProductGraph adapter |
| [OmniObserve](https://github.com/plexusone/omniobserve) | Go observability with journey semconv |

## Next Steps

- [Quick Start](getting-started/quickstart.md) - Get up and running
- [Architecture Overview](architecture/overview.md) - Understand the system
- [API Reference](api/events.md) - Integrate your app
