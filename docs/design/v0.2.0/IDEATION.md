# v0.2.0 Ideation: Analytics Adapters & SDK Design

This document captures architectural decisions and design principles for the v0.2.0 milestone, focusing on analytics tool integration and SDK design.

## Core Philosophy

### Product Graph as Source of Truth

The key insight is to avoid building a thin wrapper around analytics tools. Instead:

- Define a **canonical event + state model** (the Product Graph)
- Treat external tools as **downstream consumers**, not peers
- Adapters translate rich graph data into vendor-specific formats

```
Application
    ↓
Product Graph (canonical model)
    ↓
Export / Adapters
    ↓
Mixpanel / Amplitude / Pendo
```

This approach ensures:

- No lowest-common-denominator limitations
- Consistent semantics across all outputs
- Rich data that can be downsampled for specific tools
- Future extensibility for AI agents, observability, etc.

### Graph Semantics Over Raw Events

Traditional analytics track actions:

```ts
track("clicked_button")
```

Product Graph captures **meaning**:

```ts
graph.enterNode("CheckoutStep", { cartTotal: 150 })
graph.transition("Cart", "Checkout")
graph.setState("cart", { items: [...] })
```

This semantic upgrade enables:

- State transition tracking (not just events)
- Component context awareness
- UI snapshots and journey visualization
- Full journey graph reconstruction

## Target Integrations

### Phase 1: Amplitude & Mixpanel

Both tools are well-suited for backend integration:

| Capability | Amplitude | Mixpanel |
|------------|-----------|----------|
| Backend API | Strong | Strong |
| Event streaming | Native fit | Native fit |
| Identity management | Server-side | Server-side |

**Mapping strategy:**

```
Graph transition → Event
Node → Event name / type
State → Event properties
User → user_id / distinct_id
```

### Phase 2+ (Deferred): Pendo

Pendo is more frontend-centric:

- Requires DOM awareness and selectors
- Designed for in-app guides and tooltips
- Backend integration is limited

Recommendation: Defer Pendo until core graph model is validated.

## SDK Design Principles

### API Surface

The SDK should expose graph semantics, not raw event tracking:

```ts
// Core operations
graph.enterNode(nodeName, state)
graph.transition(from, to)
graph.setState(key, value)
graph.annotate(node, metadata)

// React integration
useGraphNode("CheckoutPage")
withGraph(Component, { node: "ProductView" })
```

### Automatic Instrumentation Hooks

- **Navigation**: Route changes → transitions
- **Network**: fetch/axios interception → API events
- **State management**: Redux/Zustand → state updates
- **React lifecycle**: Component mount/unmount → node enter/exit

### Traffic Management

To avoid overwhelming the backend:

| Technique | Purpose |
|-----------|---------|
| Batching | Collect multiple events before sending |
| Debouncing | Wait 1-5 seconds between sends |
| Sampling | 10-100% based on event priority |
| `sendBeacon` | Reliable delivery on page unload |
| `requestIdleCallback` | Send during browser idle time |

**Priority levels:**

- **High**: Errors, checkout, signup (100% capture)
- **Medium**: Feature usage, navigation (50-100%)
- **Low**: Generic browsing (10-50% sample)

## Architecture

### Data Flow

```
[TypeScript SDK]
        ↓
[Ingestion API (Edge/Node)]  ← thin, fast (~10-50ms response)
        ↓
[Event Queue (Kafka/SQS)]    ← buffering, retry, replay
        ↓
[Go Processing Layer]        ← normalize, enrich, route
        ↓
[Adapters]
   ↙        ↓        ↘
Amplitude  Mixpanel  (future)
```

### Layer Responsibilities

| Layer | Responsibilities | Anti-patterns |
|-------|------------------|---------------|
| SDK | Capture, batch, debounce, send | Direct vendor calls |
| Ingestion API | Validate, enqueue, respond fast | Heavy transformations |
| Queue | Buffer, retry, enable replay | Synchronous processing |
| Processor | Normalize, enrich, route | Vendor-specific logic |
| Adapters | Translate graph → vendor format | Core model pollution |

### Schema Requirements

The canonical schema must be:

- **Consistent**: Same structure across all events
- **Strongly typed**: TypeScript types shared across layers
- **Versioned**: Include `version` field for evolution

Essential fields:

```ts
{
  event_id: string      // Deduplication
  timestamp: number     // Client timestamp
  server_time: number   // Server timestamp
  version: string       // Schema version
  type: "node_enter" | "transition" | "state_update"
  // ... event-specific fields
}
```

## Key Design Decisions

### Backend-First Integration

All analytics tool integration happens server-side:

**Advantages:**

- Not blocked by ad blockers
- Consistent data quality
- Centralized control and debugging
- Single source of truth

**Trade-offs:**

- Requires SDK → backend → vendor flow
- Slightly higher latency than direct integration

### Adapters as Lossy Translators

The Product Graph is richer than any single vendor tool. Adapters:

- Flatten graph structures to events
- Drop unused metadata
- Reshape semantics to fit vendor models

This is expected and acceptable—you can always map rich → simple.

### Monorepo with Service Boundaries

Start with one repository, but design for future separation:

```
productgraph/
├── packages/
│   ├── sdk/           # TypeScript SDK (npm publish)
│   ├── schema/        # Shared types + validation
│   └── adapters/      # Amplitude, Mixpanel adapters
├── services/
│   ├── ingest-api/    # Node/Edge ingestion
│   └── processor/     # Go event processing
└── apps/
    └── dashboard/     # Runtime canvas (future)
```

Split repos only when:

- Independent scaling is needed
- Team boundaries emerge
- Release cycles diverge

## Success Criteria

After implementing v0.2.0:

1. Define a journey once in Product Graph
2. See it visualized in:
   - ProductGraph runtime canvas
   - Amplitude funnels
   - Mixpanel reports
3. Data consistency across all outputs

If this works cleanly, the abstraction layer is correct.

## References

- [v0.1.0 Release](../../../releases/v0.1.0.md) - Event ingestion foundation
- [Scaling Guide](../../architecture/scaling.md) - Infrastructure decisions
- [Event API](../../api/events.md) - Current ingestion endpoint
