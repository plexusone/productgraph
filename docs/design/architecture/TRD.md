# ProductGraph - Technical Requirements Document

> Moved from root TRD.md - See PLAN.md for current implementation status.

**Related Documents:**

- [SCALING.md](SCALING.md) - Architecture scaling guide with cost analysis and migration paths

## Overview

This document specifies the technical architecture, system design, and implementation details for ProductGraph, an AI-native runtime product intelligence platform.

## System Architecture

### Current Architecture (PoC - up to 1000 users)

The initial architecture uses a single PostgreSQL database with Row-Level Security (RLS) for multi-tenancy. This simplifies operations and is sufficient for ~50M events/month.

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
│                                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌────────────┐ │
│  │  Snapshots   │  │  Dashboards  │  │    Alerts    │  │   Users    │ │
│  │              │  │              │  │              │  │            │ │
│  └──────────────┘  └──────────────┘  └──────────────┘  └────────────┘ │
└─────────────────────────────────────────────────────────────────────────┘
```

### Multi-Tenancy with Row-Level Security

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

### Future Architecture (1000+ users, high volume)

When scaling beyond the PoC, introduce specialized data stores:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Ingestion Layer                                  │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                      API Gateway (Kong/Traefik)                  │   │
│  │   - Rate limiting    - Authentication    - Load balancing        │   │
│  └─────────────────────────────────┬───────────────────────────────┘   │
│                                    │                                    │
│  ┌─────────────────────────────────▼───────────────────────────────┐   │
│  │                     Event Ingestion Service                      │   │
│  │   - Schema validation         - Event enrichment                 │   │
│  │   - Deduplication             - Batching                         │   │
│  └─────────────────────────────────┬───────────────────────────────┘   │
└────────────────────────────────────┼────────────────────────────────────┘
                                     │
                    ┌────────────────┼────────────────┐
                    │                │                │
                    ▼                ▼                ▼
┌───────────────────────┐ ┌──────────────────┐ ┌──────────────────────────┐
│       Kafka           │ │    ClickHouse    │ │      Redis               │
│   (Event Stream)      │ │  (Event Storage) │ │   (Real-time State)      │
└───────────┬───────────┘ └──────────────────┘ └──────────────────────────┘
            │
            ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        Processing Layer                                  │
│  ┌─────────────────────┐  ┌─────────────────────┐  ┌─────────────────┐ │
│  │  Session Builder    │  │  Journey Processor  │  │ Metrics Agg     │ │
│  └─────────────────────┘  └─────────────────────┘  └─────────────────┘ │
│  ┌─────────────────────┐  ┌─────────────────────┐  ┌─────────────────┐ │
│  │  Snapshot Service   │  │  Alert Engine       │  │ AI Analyzer     │ │
│  └─────────────────────┘  └─────────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────────────────┘
```

#### When to Scale

| Trigger | Solution |
|---------|----------|
| Analytics queries > 2s | Add ClickHouse for event storage |
| Write throughput > 5K/sec | Add Kafka for event streaming |
| Real-time requirements | Add Redis for session state |
| Screenshot storage > 100GB | Add S3/MinIO for object storage |

## Technology Stack

### Backend (Current - PoC)

| Component | Technology | Rationale |
|-----------|------------|-----------|
| **Language** | Go 1.22+ | Performance, concurrency, strong typing |
| **ORM** | Ent | Type-safe, code-generated, PostgreSQL RLS support |
| **API Framework** | Huma v2 + Chi | OpenAPI generation, middleware |
| **Database** | PostgreSQL 16+ | All data, RLS for multi-tenancy |

### Backend (Future - Scale)

| Component | Technology | When to Add |
|-----------|------------|-------------|
| **Event Streaming** | Apache Kafka | Write throughput > 5K events/sec |
| **Event Storage** | ClickHouse | Analytics queries > 2s |
| **Caching** | Redis 7+ | Real-time dashboards needed |
| **Object Storage** | S3/R2 | Screenshot storage > 100GB |

### Frontend

| Component | Technology | Rationale |
|-----------|------------|-----------|
| **Framework** | React 19 | Component model, ecosystem |
| **Build** | Vite 6 | Fast builds, ESM |
| **State** | TanStack Query + Zustand | Server state + client state |
| **Canvas** | React Flow | Graph visualization |
| **Charts** | Apache ECharts | Rich analytics charts |
| **Styling** | Tailwind CSS 4 | Utility-first, CoreForge tokens |

## Data Models (Ent Schema)

All schemas are defined in Go using Ent and generate PostgreSQL migrations.

### Organization

```go
// ent/schema/organization.go
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
// ent/schema/project.go
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
// ent/schema/event.go
func (Event) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.UUID("org_id", uuid.UUID{}),           // RLS column
        field.UUID("project_id", uuid.UUID{}),
        field.String("session_id").NotEmpty(),
        field.String("user_id").Optional(),
        field.Enum("event_type").Values(
            "page.view", "page.leave", "ui.click", "ui.input",
            "ui.scroll", "ui.submit", "state.change",
            "api.request", "api.response", "journey.step",
            "error", "performance", "custom",
        ),
        field.String("event_name").Optional(),
        field.Time("timestamp"),
        field.Int64("sequence").Default(0),
        // Page context
        field.String("page_path").Optional(),
        field.String("page_title").Optional(),
        field.String("page_url").Optional(),
        // UI context
        field.String("ui_component_name").Optional(),
        field.String("ui_component_path").Optional(),
        field.String("ui_action").Optional(),
        // State tracking
        field.String("ui_state_key").Optional(),
        field.Text("ui_state_before").Optional(),
        field.Text("ui_state_after").Optional(),
        // Journey context
        field.UUID("journey_id", uuid.UUID{}).Optional().Nillable(),
        field.String("journey_step_id").Optional(),
        field.String("journey_step_name").Optional(),
        // Metadata
        field.JSON("metadata", map[string]any{}).Optional(),
        field.Int64("duration_ms").Optional(),
        field.Time("created_at").Default(time.Now),
    }
}

func (Event) Indexes() []ent.Index {
    return []ent.Index{
        // BRIN index for time-series queries (added via migration)
        index.Fields("org_id", "project_id", "timestamp"),
        index.Fields("org_id", "session_id"),
        index.Fields("org_id", "journey_id"),
    }
}
```

### Journey

```go
// ent/schema/journey.go
func (Journey) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.UUID("org_id", uuid.UUID{}),
        field.UUID("project_id", uuid.UUID{}),
        field.String("name").NotEmpty(),
        field.Text("description").Optional(),
        field.JSON("entry_conditions", []any{}).Optional(),
        field.JSON("exit_conditions", []any{}).Optional(),
        field.Int("timeout_minutes").Default(30),
        field.Bool("is_active").Default(true),
        field.Time("created_at").Default(time.Now),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}
```

### Row-Level Security Setup

```sql
-- Applied via Ent migration hooks or manual migration

-- Enable RLS on all tenant-scoped tables
ALTER TABLE events ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE journeys ENABLE ROW LEVEL SECURITY;
ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;

-- Create policies
CREATE POLICY org_isolation_events ON events
    USING (org_id = current_setting('app.current_org_id')::uuid);

CREATE POLICY org_isolation_projects ON projects
    USING (org_id = current_setting('app.current_org_id')::uuid);

CREATE POLICY org_isolation_journeys ON journeys
    USING (org_id = current_setting('app.current_org_id')::uuid);

-- Bypass for service role (migrations, admin)
CREATE POLICY service_bypass ON events
    FOR ALL TO service_role USING (true);
```

## Performance Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| Event ingestion latency | < 100ms p99 | API response time |
| Event processing latency | < 5s p99 | Kafka to ClickHouse |
| Query latency (simple) | < 200ms p99 | GraphQL response |
| Query latency (complex) | < 2s p99 | GraphQL response |
| Canvas render | < 500ms | Initial load |
| Real-time updates | < 1s | WebSocket delivery |

## Scalability

### Horizontal Scaling

- **Ingestion**: Stateless, scale via replicas
- **Kafka**: Partition by project_id
- **ClickHouse**: Sharding by project_id + date
- **Processing**: Kafka consumer groups

### Estimated Capacity (per node)

| Component | Capacity |
|-----------|----------|
| Ingestion service | 10K events/sec |
| Kafka broker | 100K events/sec |
| ClickHouse node | 1M events/sec write |
| Session builder | 5K sessions/sec |

## Security

### Authentication

- **API Keys**: Project-scoped keys for event ingestion
- **JWT Tokens**: User authentication for dashboard
- **OAuth 2.0**: GitHub, Google, CoreControl SSO

### Data Protection

- **PII Redaction**: Automatic scrubbing of sensitive fields
- **Encryption**: TLS 1.3 in transit, AES-256 at rest
- **Data Residency**: Configurable storage regions
