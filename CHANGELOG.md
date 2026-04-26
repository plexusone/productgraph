# Changelog

All notable changes to ProductGraph will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2024-04-26

### Added

#### Event Ingestion

- Event ingestion HTTP service (`POST /v1/events`)
- Batch event ingestion with validation (max 1000 events per batch)
- OTel-compatible event schema with 40+ fields
- Event types: page.view, page.leave, ui.click, ui.input, ui.scroll, ui.submit, state.change, api.request, api.response, journey.step, error, performance, custom
- Health (`GET /health`) and readiness (`GET /ready`) endpoints
- CORS middleware for development

#### Database

- Ent ORM schemas for Organization, Project, Event, Session, Journey
- PostgreSQL Row-Level Security (RLS) for multi-tenancy
- All tenant-scoped tables include `org_id` for isolation
- Indexes optimized for time-series queries

#### Infrastructure

- Docker Compose development environment (PostgreSQL 16)
- Makefile with build, test, generate, and migrate commands

#### Documentation

- Product Requirements Document (PRD)
- Technical Requirements Document (TRD)
- Architecture Scaling Guide with cost analysis:
  - Starter (PostgreSQL only): <1000 users, ~$150/mo
  - Growth (+Kafka, ClickHouse, Redis): 1K-10K users, ~$1,350/mo
  - Scale (sharded, multi-region): 10K+ users, ~$5,500/mo
- Implementation plan (PLAN.md)

### Architecture

This release implements the **Starter** architecture:

```
┌─────────────────────────────────┐
│   ProductGraph Service (Go)    │
│   • Event Ingestion API        │
│   • Ent ORM                    │
└───────────────┬─────────────────┘
                │
                ▼
┌─────────────────────────────────┐
│   PostgreSQL 16+ (RLS)         │
│   • Events, Sessions, Journeys │
│   • Multi-tenant via org_id    │
└─────────────────────────────────┘
```

See [SCALING.md](docs/design/architecture/SCALING.md) for upgrade paths.

### Dependencies

- Go 1.22+
- PostgreSQL 16+
- Ent v0.14.6

[0.1.0]: https://github.com/plexusone/productgraph/releases/tag/v0.1.0
