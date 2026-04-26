# ProductGraph - Product Requirements Document

> Moved from root PRD.md - See PLAN.md for current implementation status.

## Executive Summary

ProductGraph is an AI-native runtime product intelligence platform that unifies design visualization, user journey analytics, and application observability into a single system. It replaces the fragmented toolchain of Figma (design), Amplitude/Pendo (analytics), and Datadog (monitoring) with a unified runtime graph that serves as both documentation and measurement system for how products actually perform.

## Vision Statement

> A living, visual map of user journeys that doubles as both documentation and a real-time measurement system for how well those journeys actually perform.

## Problem Statement

### Current Tool Fragmentation

Teams today use disconnected tools that each capture a partial view of the product:

| Tool Category | Examples | Limitation |
|---------------|----------|------------|
| **Design** | Figma, Sketch | Static pixels, not runtime-aware |
| **Analytics** | Amplitude, Mixpanel | Abstract tables/charts, detached from UI |
| **User Guidance** | Pendo, WalkMe | Overlay system, manual event definition |
| **Monitoring** | Datadog, New Relic | Backend-focused, limited UI context |
| **Session Replay** | FullStory, LogRocket | Post-hoc debugging, no journey context |

### Core Problems

1. **Design-Reality Gap**: Figma shows intended flows; users experience something different
2. **Manual Event Instrumentation**: Analytics require developers to manually add tracking
3. **No Unified Journey View**: Funnels, debugging, and design exist in separate systems
4. **AI Agents Lack Observability**: Coding assistants can edit code but can't observe real UI behavior

## Goals

### Primary Goals

1. **Unified Runtime Canvas**: Visual graph of user journeys generated from actual application state
2. **Automatic Event Capture**: Component-level tracking without manual instrumentation
3. **Spec vs Reality**: Compare intended journeys with actual user behavior
4. **AI Development Loop**: Enable AI agents to observe, debug, and improve products

### Non-Goals (v1)

- Full Figma design replacement (focus on runtime, not static design)
- Mobile app support (web-first)
- Backend distributed tracing (focus on frontend journeys)

## User Personas

### 1. Product Manager

- **Needs**: Understand user behavior, identify drop-offs, validate features
- **Current Pain**: Switching between Amplitude dashboards and Figma specs
- **ProductGraph Value**: See journey success rates overlaid on visual flows

### 2. UX Designer

- **Needs**: Validate designs with real user data, identify friction
- **Current Pain**: Design specs don't reflect actual user paths
- **ProductGraph Value**: Visual diff between intended and actual journeys

### 3. Frontend Developer

- **Needs**: Debug UI issues, understand state transitions, fix user-reported problems
- **Current Pain**: Reproducing user issues, correlating logs with UI state
- **ProductGraph Value**: Step-by-step replay with component state at each step

### 4. AI Coding Agent

- **Needs**: Observe UI changes after code edits, understand runtime behavior
- **Current Pain**: Can edit code but can't see results; limited feedback loop
- **ProductGraph Value**: Visual feedback on code changes, automated testing

### 5. Growth Engineer

- **Needs**: Optimize funnels, A/B test UI changes, improve conversion
- **Current Pain**: Manual funnel setup, delayed insights
- **ProductGraph Value**: Automatic funnel detection, real-time metrics

## Product Concepts

### Runtime Canvas

A visual graph where each node represents a step in the user journey:

```
[Landing Page]
     │
     ├──60%──► [Product List]
     │              │
     │              ├──30%──► [Product Detail]
     │              │              │
     │              │              └──50%──► [Add to Cart]
     │              │                             │
     │              │                             └──40%──► [Checkout]
     │              │
     │              └──70%──► [Exit]
     │
     └──40%──► [Exit]
```

Each node contains:
- UI snapshot
- Component tree
- State values
- Performance metrics
- Success rate

### Journey Graph

The underlying data model connecting:
- **Pages**: Route-based application screens
- **Steps**: User actions (clicks, inputs, navigations)
- **States**: Application state at each step
- **Metrics**: Conversion rates, timing, errors

## Requirements

See [TRD.md](../architecture/TRD.md) for detailed technical requirements.

## Success Metrics

| Metric | Target | Timeline |
|--------|--------|----------|
| Instrumented applications | 5+ | Q3 2025 |
| Events ingested/day | 1M+ | Q3 2025 |
| Journey insights generated | 100+ | Q4 2025 |
| User drop-off identification accuracy | >90% | Q4 2025 |
| Time to first insight | < 5 minutes | Q3 2025 |

## Competitive Analysis

| Capability | Amplitude | Pendo | Datadog | FullStory | ProductGraph |
|------------|-----------|-------|---------|-----------|--------------|
| Event analytics | ★★★★★ | ★★★☆☆ | ★★☆☆☆ | ★★★☆☆ | ★★★★★ |
| Visual journeys | ★★☆☆☆ | ★★★☆☆ | ★☆☆☆☆ | ★★★★☆ | ★★★★★ |
| Session replay | ☆☆☆☆☆ | ★★☆☆☆ | ☆☆☆☆☆ | ★★★★★ | ★★★★☆ |
| Component state | ☆☆☆☆☆ | ☆☆☆☆☆ | ☆☆☆☆☆ | ★★☆☆☆ | ★★★★★ |
| AI integration | ★★☆☆☆ | ★☆☆☆☆ | ★★☆☆☆ | ★☆☆☆☆ | ★★★★★ |
| Auto-instrumentation | ☆☆☆☆☆ | ★★★☆☆ | ☆☆☆☆☆ | ★★★★☆ | ★★★★★ |
| Spec vs reality | ☆☆☆☆☆ | ☆☆☆☆☆ | ☆☆☆☆☆ | ☆☆☆☆☆ | ★★★★★ |
