# Architecture Overview

This document describes the technical architecture, system design, and implementation details for ProductGraph.

## Current Architecture (Starter)

The Starter architecture uses a single PostgreSQL database with Row-Level Security (RLS) for multi-tenancy. This simplifies operations and is sufficient for ~50M events/month (~1000 paying users).

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              Clients                                     │
│   ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────────┐ │
│   │ @coreforge/  │  │ @omniobserve │  │     Future: Swift/Kotlin     │ │
│   │  telemetry   │  │     /core    │  │                              │ │
│   └──────┬───────┘  └──────┬───────┘  └──────────────┬───────────────┘ │
└──────────┼─────────────────┼────────────────────────┼──────────────────┘
           │                 │                        │
           └─────────────────┴───────────┬────────────┘
                                         │ HTTPS
                                         ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                     ProductGraph Service (Single Binary)                 │
│                                                                          │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────┐ │
│  │ Event Ingestion │  │   GraphQL API   │  │   WebSocket (Future)    │ │
│  │ POST /v1/events │  │                 │  │                         │ │
│  └────────┬────────┘  └────────┬────────┘  └────────────┬────────────┘ │
│           │                    │                        │              │
│           └────────────────────┴────────────────────────┘              │
│                                │                                        │
│                    ┌───────────┴───────────┐                           │
│                    │     Ent ORM Layer     │                           │
│                    │  (Schema & Queries)   │                           │
│                    └───────────┬───────────┘                           │
└────────────────────────────────┼────────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    PostgreSQL 16+ (Single Instance)                      │
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                   Row-Level Security (RLS)                       │   │
│  │              Tenant isolation via org_id policies                │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌────────────┐ │
│  │    Events    │  │   Sessions   │  │   Journeys   │  │  Projects  │ │
│  │  (BRIN idx)  │  │              │  │              │  │            │ │
│  └──────────────┘  └──────────────┘  └──────────────┘  └────────────┘ │
└─────────────────────────────────────────────────────────────────────────┘
```

## Multi-Tenancy with Row-Level Security

All tenant-scoped tables include an `org_id` column with RLS policies:

```sql
-- Enable RLS on table
ALTER TABLE events ENABLE ROW LEVEL SECURITY;

-- Policy: users can only see their organization's data
CREATE POLICY org_isolation ON events
    USING (org_id = current_setting('app.current_org_id')::uuid);

-- Set org context per request
SET app.current_org_id = 'org-uuid-here';
```

This approach provides:

- **Security**: Tenants cannot access each other's data
- **Simplicity**: No application-level filtering required
- **Performance**: PostgreSQL optimizes queries with RLS

## Technology Stack

### Backend

| Component | Technology | Rationale |
|-----------|------------|-----------|
| **Language** | Go 1.22+ | Performance, concurrency, strong typing |
| **ORM** | Ent | Type-safe, code-generated, PostgreSQL RLS support |
| **API Framework** | Chi | Lightweight, idiomatic Go router |
| **Database** | PostgreSQL 16+ | All data, RLS for multi-tenancy |

### Frontend (Planned)

| Component | Technology | Rationale |
|-----------|------------|-----------|
| **Framework** | React 19 | Component model, ecosystem |
| **Build** | Vite 6 | Fast builds, ESM |
| **State** | TanStack Query + Zustand | Server state + client state |
| **Canvas** | React Flow | Graph visualization |
| **Charts** | Apache ECharts | Rich analytics charts |
| **Styling** | Tailwind CSS 4 | Utility-first |

## Data Models

All schemas are defined in Go using Ent and generate PostgreSQL migrations.

### Organization

```go
func (Organization) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.String("name").NotEmpty(),
        field.String("slug").Unique().NotEmpty(),
        field.Time("created_at").Default(time.Now),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}
```

### Project

```go
func (Project) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.UUID("org_id", uuid.UUID{}),
        field.String("name").NotEmpty(),
        field.String("slug").NotEmpty(),
        field.String("api_key").Unique().NotEmpty(),
        field.JSON("settings", map[string]any{}).Optional(),
        field.Time("created_at").Default(time.Now),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}
```

### Event

```go
func (Event) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.UUID("org_id", uuid.UUID{}),           // RLS column
        field.UUID("project_id", uuid.UUID{}),
        field.String("session_id").NotEmpty(),
        field.String("user_id").Optional(),
        field.Enum("event_type").Values(
            "page_view", "page_leave", "ui_click", "ui_input",
            "ui_scroll", "ui_submit", "state_change",
            "api_request", "api_response", "journey_step",
            "error", "performance", "custom",
        ),
        field.Time("timestamp"),
        // ... additional fields
    }
}

func (Event) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("org_id", "project_id", "timestamp"),
        index.Fields("org_id", "session_id"),
        index.Fields("org_id", "journey_id"),
    }
}
```

## Performance Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| Event ingestion latency | < 100ms p99 | API response time |
| Query latency (simple) | < 200ms p99 | GraphQL response |
| Query latency (complex) | < 2s p99 | GraphQL response |
| Canvas render | < 500ms | Initial load |

## Security

### Authentication

- **API Keys**: Project-scoped keys for event ingestion
- **JWT Tokens**: User authentication for dashboard (planned)
- **OAuth 2.0**: GitHub, Google SSO (planned)

### Data Protection

- **Row-Level Security**: Tenant isolation at database level
- **Encryption**: TLS 1.3 in transit
- **PII Redaction**: Configurable field scrubbing (planned)

## Future Architecture

See the [Scaling Guide](scaling.md) for:

- Growth architecture (+Kafka, ClickHouse, Redis)
- Scale architecture (sharded, multi-region)
- Migration paths and cost analysis
