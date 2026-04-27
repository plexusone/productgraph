# ProductGraph - Task Tracker

## Status Legend

- [ ] Not started
- [~] In progress
- [x] Complete
- [-] Blocked
- [!] Needs review

---

## Phase 1: Foundation

### 1.1 Project Setup

- [ ] Initialize Go module
  ```bash
  go mod init github.com/plexusone/productgraph
  ```

- [ ] Create directory structure
  ```
  productgraph/
  ├── cmd/
  │   ├── ingestion/
  │   ├── processor/
  │   └── api/
  ├── internal/
  │   ├── events/
  │   ├── sessions/
  │   ├── journeys/
  │   ├── analytics/
  │   └── storage/
  ├── pkg/
  │   └── schema/
  ├── web/                  # React frontend
  ├── sdk/                  # TypeScript SDK
  ├── deployments/
  │   ├── docker/
  │   └── helm/
  └── docs/
  ```

- [ ] Set up Docker Compose for local development
  - Kafka (KRaft mode)
  - ClickHouse
  - PostgreSQL
  - Redis
  - MinIO (S3-compatible)

- [ ] Configure GitHub Actions CI
  - Go lint + test
  - TypeScript lint + test
  - Docker build
  - Helm lint

- [ ] Create Helm charts
  - ingestion service
  - processor service
  - api service
  - frontend

### 1.2 Event Ingestion Service

- [ ] Define event schema
  ```
  pkg/schema/
  ├── event.go           # Go types
  ├── event_test.go      # Validation tests
  └── openapi.yaml       # OpenAPI spec
  ```

- [ ] Implement ingestion handler
  - POST /v1/events endpoint
  - Batch support (up to 1000 events)
  - Schema validation
  - Error responses with details

- [ ] Add Kafka producer
  - Async publishing
  - Retry logic
  - Dead letter queue

- [ ] Implement rate limiting
  - Per-project limits
  - Redis-based counter
  - 429 response with retry-after

- [ ] Add API key authentication
  - X-PG-API-Key header
  - Project lookup from key
  - Key rotation support

- [ ] Write tests
  - Unit tests for validation
  - Integration tests with Kafka
  - Load tests

### 1.3 Event Storage (ClickHouse)

- [ ] Create ClickHouse schema
  ```sql
  -- migrations/001_events.sql
  CREATE TABLE events (...)
  CREATE TABLE events_buffer (...)
  ```

- [ ] Implement Kafka consumer
  - Consumer group: pg-clickhouse-writer
  - Batch inserts (1000 events or 1 second)
  - Error handling with DLQ

- [ ] Add materialized views
  - Events by project/day
  - Events by session
  - Events by page

- [ ] Benchmark performance
  - Target: 10K events/sec/node
  - Measure query latency

### 1.4 Basic SDK

- [ ] Create SDK package structure
  ```
  sdk/
  ├── src/
  │   ├── index.ts
  │   ├── emitter.ts
  │   ├── schema.ts
  │   ├── providers/
  │   │   └── react.tsx
  │   └── sinks/
  │       └── http.ts
  ├── package.json
  └── tsconfig.json
  ```

- [ ] Implement event emitter
  - Type-safe event creation
  - Auto-populated fields (timestamp, session_id)
  - Batching (10 events or 5 seconds)

- [ ] Add HTTP sink
  - Configurable endpoint
  - API key header
  - Retry logic
  - sendBeacon on unload

- [ ] Create React provider
  - Context for emitter
  - User ID resolution
  - Automatic page views

- [ ] Publish to npm
  - Package as `@productgraph/sdk`
  - Include TypeScript types
  - Write README

---

## Phase 1.5: Analytics Integration (v0.2.0)

### 1.5.1 OmniDXI Adapter

- [x] Create analytics adapter
  - Implements `events.Publisher` interface
  - Maps ProductGraph events to omnidxi events
  - Supports all event types and properties

- [ ] Add adapter tests
  - Unit tests for event mapping
  - Mock tracker for isolation
  - Test all event types
  - Test context/properties extraction

- [ ] Wire adapter to ingestion service
  - Add to `cmd/ingestion/main.go`
  - Create tracker from config
  - Chain with existing publisher (fan-out)

### 1.5.2 Analytics Configuration

- [ ] Add config schema
  ```yaml
  analytics:
    enabled: true
    providers:
      amplitude:
        enabled: true
        api_key: ${AMPLITUDE_API_KEY}
      mixpanel:
        enabled: true
        token: ${MIXPANEL_TOKEN}
  ```

- [ ] Implement config loading
  - Viper or env-based config
  - Validate required fields when enabled
  - Graceful degradation if provider unavailable

- [ ] Add provider health checks
  - Include in `/ready` endpoint
  - Log provider status on startup

### 1.5.3 Fan-out Publisher

- [ ] Create multi-publisher
  - Wraps multiple `events.Publisher` implementations
  - Parallel dispatch to all publishers
  - Aggregate errors without blocking

- [ ] Integrate with event handler
  - Memory publisher (existing)
  - Analytics adapter (new)
  - Future: Kafka publisher

### 1.5.4 Documentation & Release

- [ ] Update CHANGELOG for v0.2.0
  - Add omnidxi integration feature
  - Document new dependencies
  - Note configuration options

- [ ] Update README
  - Add analytics section
  - Configuration examples
  - Provider setup instructions

- [ ] Update API docs
  - Document event flow to analytics
  - Architecture diagram update

---

## Phase 2: Session & Journey Processing

### 2.1 Session Builder Service

- [ ] Define session data model
  ```go
  type Session struct {
      ID          string
      ProjectID   string
      UserID      *string
      StartedAt   time.Time
      EndedAt     time.Time
      Events      []Event
      Converted   bool
      EntryPage   string
      ExitPage    string
  }
  ```

- [ ] Implement session builder
  - Kafka consumer (events.raw topic)
  - Group events by session_id
  - Detect session end (30 min timeout)
  - Emit to sessions topic

- [ ] Store sessions in ClickHouse
  - Sessions table
  - Session-event mapping

- [ ] Cache active sessions in Redis
  - Key: session:{project_id}:{session_id}
  - TTL: 35 minutes
  - Update on each event

### 2.2 Journey Processor Service

- [ ] Define journey model (PostgreSQL)
  ```sql
  CREATE TABLE journeys (
      id UUID PRIMARY KEY,
      project_id UUID NOT NULL,
      name VARCHAR(255),
      nodes JSONB,
      edges JSONB,
      ...
  );
  ```

- [ ] Implement journey matching
  - Load journey definitions
  - Match sessions to journeys
  - Calculate conversion per journey

- [ ] Auto-detect common paths
  - Analyze event sequences
  - Cluster similar paths
  - Suggest journey definitions

- [ ] Update journey metrics
  - Scheduled job (every 5 min)
  - Calculate conversion rates
  - Identify top drop-off nodes

### 2.3 Basic Query API

- [ ] Set up GraphQL server
  - Use gqlgen
  - Define schema
  - Implement resolvers

- [ ] Journey queries
  - Get journey by ID
  - List journeys by project
  - Journey metrics

- [ ] Session queries
  - Get session by ID
  - List sessions (filtered)
  - Session events

- [ ] Funnel queries
  - Define funnel steps
  - Calculate conversion rates
  - Filter by date range

---

## Phase 3: Visual Canvas

### 3.1 Frontend Foundation

- [ ] Initialize React app
  ```bash
  pnpm create vite web --template react-ts
  ```

- [ ] Set up routing
  - /login
  - /projects
  - /projects/:id/journeys
  - /projects/:id/journeys/:journeyId
  - /projects/:id/sessions
  - /projects/:id/analytics

- [ ] Integrate TanStack Query
  - GraphQL client setup
  - Query hooks for journeys/sessions

- [ ] Authentication
  - OAuth flow (GitHub, Google)
  - JWT token handling
  - Protected routes

- [ ] Apply CoreForge design tokens
  - Install @coreforge/design-tokens
  - Configure Tailwind preset

### 3.2 Canvas Viewer

- [ ] Install React Flow
  ```bash
  pnpm add reactflow
  ```

- [ ] Implement journey graph
  - Custom node types (page, action, decision)
  - Edge labels with conversion rates
  - Auto-layout algorithm

- [ ] Node detail panel
  - Show page path
  - Show component info
  - Display metrics (visits, duration, errors)
  - Show screenshot

- [ ] Canvas controls
  - Zoom in/out
  - Fit view
  - Toggle minimap
  - Export as image

### 3.3 Snapshot Capture

- [ ] Add screenshot capture to SDK
  - html2canvas integration
  - Configurable quality
  - Privacy masking

- [ ] Set up S3/R2 storage
  - Create bucket
  - Configure CORS
  - Set up CDN

- [ ] Create upload endpoint
  - Presigned URL generation
  - Size limits
  - Deduplication

- [ ] Display in canvas
  - Lazy load images
  - Placeholder while loading
  - Zoom on click

---

## Phase 4: Analytics Dashboard

### 4.1 Funnel Analytics

- [ ] Funnel builder UI
  - Step selector (page, action, component)
  - Drag-and-drop reordering
  - Save funnel definitions

- [ ] Funnel visualization
  - Horizontal bar chart
  - Percentage labels
  - Click to drill down

- [ ] Date range picker
  - Presets (7d, 30d, 90d)
  - Custom range
  - Compare periods

### 4.2 Retention Analytics

- [ ] Retention query
  - Cohort by signup date
  - Return events (page view, action)
  - Daily/weekly/monthly

- [ ] Cohort table
  - Row per cohort
  - Column per period
  - Color gradient

- [ ] Export
  - CSV download
  - JSON export

### 4.3 Real-time Metrics

- [ ] WebSocket server
  - Connection management
  - Project-scoped channels
  - Authentication

- [ ] Live dashboard widgets
  - Active users (5 min window)
  - Events per minute
  - Active journeys

- [ ] Live event stream
  - Filterable by type
  - Pause/resume
  - Click to view details

---

## Phase 5: Session Replay

### 5.1 Replay Backend

- [ ] Session events endpoint
  - GET /sessions/:id/events
  - Include state snapshots
  - Order by timestamp

- [ ] Optimize for large sessions
  - Pagination
  - Streaming response
  - Compression

### 5.2 Replay Frontend

- [ ] Replay player component
  - Event timeline
  - Play/pause button
  - Speed controls (1x, 2x, 4x)
  - Timeline scrubber

- [ ] State inspector
  - Show state at current step
  - Diff with previous state
  - Expand/collapse nested objects

- [ ] Component tree viewer
  - React DevTools-like view
  - Highlight current component

### 5.3 Session Search

- [ ] Search API
  - Full-text on actions
  - Filter by user
  - Filter by page
  - Filter by conversion
  - Filter by errors

- [ ] Search UI
  - Search input with suggestions
  - Filter chips
  - Results list
  - Quick preview

---

## Phase 6: AI Integration

### 6.1 MCP Server

- [ ] Implement MCP protocol
  - Tool registration
  - Request handling
  - Response formatting

- [ ] Define tools
  - pg.list_journeys
  - pg.get_journey
  - pg.get_funnel
  - pg.get_drop_offs
  - pg.get_session
  - pg.search_sessions
  - pg.get_suggestions

- [ ] Write tool documentation
  - Description
  - Parameters
  - Example responses

### 6.2 AI Insights

- [ ] Anomaly detection
  - Baseline metrics
  - Detect deviations
  - Score severity

- [ ] Suggestion engine
  - Identify drop-offs
  - Suggest improvements
  - Prioritize by impact

- [ ] Insight cards UI
  - Card layout
  - Severity indicator
  - Action buttons

### 6.3 AI Testing

- [ ] Journey test specs
  - Define expected paths
  - Define success criteria

- [ ] Test runner
  - Simulate user flows
  - Compare to spec
  - Report failures

---

## Phase 7: Alerting & Polish

### 7.1 Alert System

- [ ] Alert rules schema
  - Metric type
  - Threshold
  - Time window
  - Notification channels

- [ ] Alert evaluation
  - Scheduled checks
  - Threshold comparison
  - Deduplication

- [ ] Notifications
  - Slack integration
  - Webhook support
  - Email (future)

- [ ] Alert UI
  - Create/edit rules
  - View alert history
  - Acknowledge/silence

### 7.2 Documentation

- [ ] Set up documentation site
  - MkDocs or Docusaurus
  - API reference
  - Guides

- [ ] Write content
  - Getting started
  - SDK integration
  - GraphQL API
  - MCP tools
  - Self-hosting guide

- [ ] Create examples
  - React app example
  - Vue app example
  - Next.js example

### 7.3 Performance & Security

- [ ] Load testing
  - Ingestion: 10K events/sec
  - Queries: 1K req/sec
  - Canvas: 100 concurrent users

- [ ] Security audit
  - OWASP checklist
  - Dependency scan
  - Penetration testing

- [ ] GDPR compliance
  - Data export API
  - Data deletion API
  - Consent management

- [ ] PII redaction
  - Automatic detection
  - Configurable rules
  - Redaction in storage

---

## Bugs & Issues

| ID | Description | Status | Component |
|----|-------------|--------|-----------|
| | | | |

---

## Notes

### Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-04-27 | Use omnidxi for analytics adapters | Unified interface for Amplitude/Mixpanel, backend-first integration, avoids ad blockers |
| | Use ClickHouse over TimescaleDB | Better analytics query performance |
| | Use gqlgen over graphql-go | Type-safe, better tooling |
| | Use React Flow over custom canvas | Mature, well-documented |

### Open Questions

- [ ] Should we support real-time DOM diffing for replay?
- [ ] How to handle screenshot privacy (blur sensitive data)?
- [ ] Self-hosted vs cloud-only for initial release?
- [ ] Pricing model (events-based, seats, hybrid)?

### Research Notes

- ClickHouse sharding strategy: https://clickhouse.com/docs/en/guides/sre/scaling-clusters
- MCP protocol spec: https://spec.modelcontextprotocol.io/
- React Flow performance: https://reactflow.dev/docs/guides/performance/
