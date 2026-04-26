# ProductGraph - Implementation Plan

## Quick Links

| Document | Description |
|----------|-------------|
| [PRD](docs/design/overview/PRD.md) | Product Requirements Document |
| [TRD](docs/design/architecture/TRD.md) | Technical Requirements Document |
| [Ideation](docs/design/ideation/figma-ai-canvas.md) | Figma vs AI-Native Canvas discussion |

## Multi-Project Integration

ProductGraph operates as a unified platform receiving telemetry from multiple client SDKs:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           PRODUCTGRAPH BACKEND                              │
│  ┌─────────────────┐  ┌──────────────────┐  ┌────────────────────────────┐  │
│  │ Design Canvas   │  │ Analytics Engine │  │ Session Replay             │  │
│  │ (Journey Viz)   │  │ (Funnels/Cohorts)│  │ (Step-by-step inspection)  │  │
│  └─────────────────┘  └──────────────────┘  └────────────────────────────┘  │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │ Event Ingestion API (OTel-compatible + ProductGraph extensions)     │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
          ▲                    ▲                    ▲
          │                    │                    │
┌─────────┴─────────┐ ┌───────┴────────┐ ┌─────────┴─────────┐
│ @coreforge/       │ │ @omniobserve/  │ │ OmniObserve-Swift │
│ telemetry         │ │ typescript     │ │ (iOS/macOS)       │
│ + ProductGraph    │ │                │ │                   │
└───────────────────┘ └────────────────┘ └───────────────────┘
```

### Project Dependencies

| Project | Path | Status |
|---------|------|--------|
| **ProductGraph** | `plexusone/productgraph` | This repo |
| **CoreForge-Web** | `grokify/coreforge-web` | ✅ ProductGraph adapter created |
| **OmniObserve** | `plexusone/omniobserve` | ✅ semconv/journey created |
| **OmniObserve-TS** | `omniobserve/typescript` | ✅ @omniobserve/core created |

---

## Current Sprint: Foundation

### Priority 1: Project Setup & Event Ingestion

**Goal**: Accept events from CoreForge-Web apps and store in ClickHouse.

#### 1.1 Initialize Go Project

```
productgraph/
├── cmd/
│   ├── ingestion/          # Event ingestion service
│   ├── processor/          # Session & journey processor
│   └── api/                # GraphQL API server
├── internal/
│   ├── events/             # Event types and validation
│   ├── sessions/           # Session reconstruction
│   ├── journeys/           # Journey matching
│   ├── analytics/          # Funnel & cohort queries
│   └── storage/            # ClickHouse, Postgres, Redis
├── pkg/
│   └── schema/             # Shared event schema (OpenAPI)
├── web/                    # React frontend
├── deployments/
│   ├── docker/             # Docker Compose for local dev
│   └── helm/               # Kubernetes Helm charts
└── docs/
    └── design/             # Design documentation
```

#### 1.2 Event Schema (OTel-compatible)

```go
// pkg/schema/event.go
type Event struct {
    // Identity
    EventID   string `json:"event_id"`
    ProjectID string `json:"project_id"`
    SessionID string `json:"session.id"`
    UserID    string `json:"user.id,omitempty"`

    // Classification (OTel-compatible)
    EventType string `json:"event.type"` // page.view, ui.click, etc.
    EventName string `json:"event.name"`
    Timestamp string `json:"event.timestamp"`
    Sequence  int64  `json:"event.sequence"`

    // Page context
    PagePath  string `json:"page.path,omitempty"`
    PageTitle string `json:"page.title,omitempty"`

    // UI context (OTel ui.* namespace)
    UIComponentName string `json:"ui.component.name,omitempty"`
    UIComponentPath string `json:"ui.component.path,omitempty"`
    UIAction        string `json:"ui.action,omitempty"`

    // State tracking
    UIStateBefore string `json:"ui.state.before,omitempty"`
    UIStateAfter  string `json:"ui.state.after,omitempty"`

    // Journey context
    JourneyID     string `json:"gen_ai.journey.id,omitempty"`
    JourneyStepID string `json:"gen_ai.journey.step.id,omitempty"`

    // Performance
    DurationMs int64 `json:"duration_ms,omitempty"`

    // Errors
    ErrorType    string `json:"error.type,omitempty"`
    ErrorMessage string `json:"error.message,omitempty"`

    // Custom
    Metadata map[string]any `json:"metadata,omitempty"`
}
```

#### 1.3 Local Development Environment

```yaml
# deployments/docker/docker-compose.yml
services:
  kafka:
    image: bitnami/kafka:3.7
    ports: ["9092:9092"]
    environment:
      KAFKA_CFG_NODE_ID: 1
      KAFKA_CFG_PROCESS_ROLES: controller,broker
      KAFKA_CFG_CONTROLLER_QUORUM_VOTERS: 1@kafka:9093

  clickhouse:
    image: clickhouse/clickhouse-server:24.3
    ports: ["8123:8123", "9000:9000"]
    volumes:
      - ./clickhouse/init:/docker-entrypoint-initdb.d

  postgres:
    image: postgres:16
    ports: ["5432:5432"]
    environment:
      POSTGRES_DB: productgraph
      POSTGRES_USER: pg
      POSTGRES_PASSWORD: pg

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]
```

---

### Priority 2: Analytics Dashboard (Mixpanel-like)

**Goal**: Show real-time journey metrics with funnel and cohort analysis.

#### 2.1 Funnel Visualization

```
┌─────────────────────────────────────────────────────────────┐
│  Checkout Funnel                    Last 30 days            │
├─────────────────────────────────────────────────────────────┤
│  ████████████████████████████████  100% (10,000)  Cart      │
│  ████████████████████████          75%  (7,500)   Shipping  │
│  █████████████████                 52%  (5,200)   Payment   │
│  ██████████████                    45%  (4,500)   Review    │
│  ████████████                      38%  (3,800)   Confirm   │
│                                                             │
│  Drop-off Analysis:                                         │
│  • Cart → Shipping: -25% (shipping cost shown)              │
│  • Payment → Review: -13% (payment errors: 8%)              │
└─────────────────────────────────────────────────────────────┘
```

#### 2.2 Cohort Table

```
┌────────────────────────────────────────────────────────────────────┐
│  User Retention - Signup Cohort                                    │
├────────┬────────┬────────┬────────┬────────┬────────┬─────────────┤
│ Cohort │ Users  │ Week 1 │ Week 2 │ Week 3 │ Week 4 │ ...         │
├────────┼────────┼────────┼────────┼────────┼────────┼─────────────┤
│ Mar 1  │ 1,200  │ 45%    │ 32%    │ 28%    │ 25%    │             │
│ Mar 8  │ 1,450  │ 48%    │ 35%    │ 30%    │ -      │             │
└────────┴────────┴────────┴────────┴────────┴────────┴─────────────┘
```

#### 2.3 Real-time Widgets

```
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│ Active Users    │  │ Events/min      │  │ Conversion Rate │
│     1,247       │  │     3,420       │  │     42.3%       │
│    ↑ 12%        │  │    → steady     │  │    ↓ 2.1%       │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

---

### Priority 3: Design Canvas (Figma-like)

**Goal**: Visual journey map with zoomable screenshots and state inspection.

#### 3.1 Journey Graph Visualization

```
Overview (Zoomed Out)           Detail (Zoomed In)
┌─────────────────────┐         ┌─────────────────────────────┐
│ [📸] → [📸] → [📸] │         │ ┌─────────────────────────┐ │
│   ↓       ↓         │   →     │ │      Screenshot         │ │
│ [📸]   [📸] → [📸] │         │ │      (Full Size)        │ │
│           ↓         │         │ └─────────────────────────┘ │
│         [📸]        │         │ Component: CheckoutForm    │
└─────────────────────┘         │ State: { items: [...] }    │
                                │ [Open in Browser]          │
                                └─────────────────────────────┘
```

#### 3.2 Jupyter-like State Inspection

```typescript
interface StepInspector {
  // Click a step to load its state dynamically
  loadStepState(stepId: string): Promise<StepState>;
}

interface StepState {
  screenshot: string;           // URL to screenshot
  componentTree: ComponentNode[];
  stateSnapshot: Record<string, any>;
  apiCalls: APICallRecord[];
  timestamp: Date;
  url: string;                  // For "Open in Browser"
}
```

---

## Implementation Phases

### Phase 1: Foundation (Current)

- [x] Multi-project setup (OTel semconv, CoreForge adapter)
- [ ] Initialize Go module
- [ ] Create Docker Compose environment
- [ ] Implement event ingestion endpoint
- [ ] ClickHouse schema and writer
- [ ] Basic event validation

### Phase 2: Session & Journey Processing

- [ ] Session reconstruction (30-min timeout)
- [ ] Journey definition model (PostgreSQL)
- [ ] Journey matching algorithm
- [ ] Redis session cache

### Phase 3: Analytics Dashboard

- [ ] GraphQL API setup (gqlgen)
- [ ] Funnel query builder
- [ ] Cohort analysis queries
- [ ] Real-time WebSocket stream
- [ ] React dashboard UI

### Phase 4: Design Canvas

- [ ] React Flow integration
- [ ] Screenshot capture in SDK
- [ ] S3/R2 snapshot storage
- [ ] Zoomable journey visualization
- [ ] State inspection panel

### Phase 5: Session Replay

- [ ] Session event retrieval API
- [ ] Replay player component
- [ ] Timeline scrubber
- [ ] State diff viewer

### Phase 6: AI Integration

- [ ] MCP server implementation
- [ ] Journey query tools
- [ ] Anomaly detection
- [ ] Suggestion engine

---

## OTel Semantic Conventions

ProductGraph uses OTel-compatible semantic conventions for interoperability:

| Namespace | Purpose | Source |
|-----------|---------|--------|
| `gen_ai.journey.*` | Journey tracking | omniobserve/semconv/journey |
| `gen_ai.agent.*` | Agent observability | omniobserve/semconv/agent |
| `session.*` | User sessions | omniobserve/semconv/journey |
| `ui.*` | UI interactions | omniobserve/semconv/journey |
| `page.*` | Page navigation | omniobserve/semconv/journey |

### OTel Metrics Export

```go
// Metrics exported via OTel for Prometheus/Grafana
var (
    journeyConversionRate = otel.Float64Histogram(
        "gen_ai.journey.conversion_rate",
    )
    sessionDuration = otel.Int64Histogram(
        "session.duration_ms",
    )
    stepDropoffRate = otel.Float64Histogram(
        "gen_ai.journey.step.dropoff_rate",
    )
)
```

---

## Success Criteria

| Criterion | Target |
|-----------|--------|
| Event ingestion latency | < 100ms p99 |
| Canvas load time | < 500ms |
| Query latency | < 200ms (simple), < 2s (complex) |
| SDK bundle size | < 10KB gzipped |

---

## Next Steps

1. **Initialize Go module** and create directory structure
2. **Set up Docker Compose** with Kafka, ClickHouse, PostgreSQL, Redis
3. **Implement event ingestion** POST /v1/events endpoint
4. **Create ClickHouse schema** and Kafka consumer
5. **Test with CoreForge-Web** ProductGraph adapter
