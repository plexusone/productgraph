# ProductGraph Architecture Scaling Guide

This document outlines the recommended architecture at different stages of scale and revenue, including trade-offs between simplicity and performance.

## Architecture Overview

ProductGraph can be deployed in three configurations:

| Configuration | Users | Events/Month | Revenue | Complexity |
|---------------|-------|--------------|---------|------------|
| **Starter** | 1-1,000 | <50M | $0-50K ARR | Low |
| **Growth** | 1,000-10,000 | 50M-500M | $50K-500K ARR | Medium |
| **Scale** | 10,000+ | 500M+ | $500K+ ARR | High |

## Starter Architecture (v0.1.0)

**Recommended for:** Early-stage, PoC, first 1000 paying users.

```
┌─────────────────────────────────────────────────────────────────┐
│                    ProductGraph Service                          │
│                      (Single Binary)                             │
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────────┐ │
│  │   Ingestion  │  │  GraphQL API │  │   Background Workers   │ │
│  │  /v1/events  │  │              │  │  (Session aggregation) │ │
│  └──────┬───────┘  └──────┬───────┘  └───────────┬────────────┘ │
│         │                 │                      │              │
│         └─────────────────┴──────────────────────┘              │
│                           │                                      │
│                    ┌──────┴──────┐                              │
│                    │   Ent ORM   │                              │
│                    └──────┬──────┘                              │
└───────────────────────────┼─────────────────────────────────────┘
                            │
                            ▼
               ┌────────────────────────┐
               │   PostgreSQL 16+ (RLS) │
               │                        │
               │  • Events (BRIN index) │
               │  • Sessions            │
               │  • Journeys            │
               │  • Projects/Orgs       │
               └────────────────────────┘
```

### Components

| Component | Technology | Purpose |
|-----------|------------|---------|
| Database | PostgreSQL 16+ | All data storage with RLS |
| ORM | Ent | Type-safe queries, migrations |
| API | Go + Chi | HTTP ingestion + GraphQL |

### Trade-offs

| Pros | Cons |
|------|------|
| Single database to manage | Analytics queries may slow >2s at scale |
| Simple deployment (single binary + DB) | No real-time streaming |
| Low operational cost | Limited horizontal scaling |
| Easy debugging | Session aggregation in-process |
| RLS provides secure multi-tenancy | Must carefully index for performance |

### PostgreSQL Optimization Tips

```sql
-- Use BRIN indexes for time-series event data
CREATE INDEX events_timestamp_brin ON events
  USING BRIN (org_id, project_id, timestamp);

-- Partition by month for large event tables
CREATE TABLE events (
  ...
) PARTITION BY RANGE (timestamp);

-- Use connection pooling (PgBouncer)
-- Tune shared_buffers, effective_cache_size
```

### When to Upgrade

Migrate to Growth architecture when:

- [ ] Event ingestion latency p99 > 200ms
- [ ] Analytics queries consistently > 2s
- [ ] Database CPU > 70% sustained
- [ ] Storage costs exceed compute savings

---

## Growth Architecture

**Recommended for:** Scaling to 10,000 users, $50K-500K ARR.

```
┌─────────────────────────────────────────────────────────────────┐
│                       API Layer                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────────┐ │
│  │   Ingestion  │  │  GraphQL API │  │      WebSocket         │ │
│  │  /v1/events  │  │              │  │   (Real-time updates)  │ │
│  └──────┬───────┘  └──────┬───────┘  └───────────┬────────────┘ │
└─────────┼─────────────────┼──────────────────────┼──────────────┘
          │                 │                      │
          ▼                 │                      │
┌─────────────────┐         │                      │
│      Kafka      │         │                      │
│  (Event Stream) │         │                      │
└────────┬────────┘         │                      │
         │                  │                      │
         ▼                  │                      │
┌─────────────────┐         │                      │
│   Processors    │         │                      │
│ • Session build │         │                      │
│ • Journey match │         │                      │
└────────┬────────┘         │                      │
         │                  │                      │
         ▼                  ▼                      ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│   ClickHouse    │  │   PostgreSQL    │  │      Redis      │
│    (Events)     │  │   (Metadata)    │  │   (Sessions)    │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

### New Components

| Component | Technology | Purpose |
|-----------|------------|---------|
| Event Stream | Kafka | Decouple ingestion from processing |
| Analytics DB | ClickHouse | Fast columnar analytics |
| Cache | Redis | Real-time session state |

### Migration Path from Starter

1. **Add Kafka** (Week 1-2)
   - Deploy Kafka cluster
   - Update ingestion to publish to Kafka
   - Add consumer for PostgreSQL writes (temporary)

2. **Add ClickHouse** (Week 3-4)
   - Deploy ClickHouse
   - Add consumer to write events to ClickHouse
   - Migrate analytics queries to ClickHouse
   - Keep recent events in PostgreSQL for joins

3. **Add Redis** (Week 5-6)
   - Deploy Redis
   - Move session state to Redis
   - Add WebSocket support for real-time

### Trade-offs

| Pros | Cons |
|------|------|
| Horizontal scaling for ingestion | 3 databases to manage |
| Sub-second analytics queries | More complex deployment |
| Real-time capabilities | Higher operational cost |
| Kafka replay for reprocessing | Need Kafka expertise |

### Cost Estimate (AWS)

| Component | Instance | Monthly Cost |
|-----------|----------|--------------|
| Kafka (MSK) | kafka.m5.large x3 | ~$600 |
| ClickHouse | r6g.xlarge x2 | ~$400 |
| PostgreSQL (RDS) | db.r6g.large | ~$200 |
| Redis (ElastiCache) | cache.r6g.large | ~$150 |
| **Total** | | **~$1,350/mo** |

---

## Scale Architecture

**Recommended for:** 10,000+ users, $500K+ ARR.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Load Balancer                                   │
└─────────────────────────────────────┬───────────────────────────────────────┘
                                      │
┌─────────────────────────────────────┴───────────────────────────────────────┐
│                           API Gateway (Kong)                                 │
│                    Rate limiting, Auth, Routing                              │
└─────────────────────────────────────┬───────────────────────────────────────┘
                                      │
         ┌────────────────────────────┼────────────────────────────┐
         │                            │                            │
         ▼                            ▼                            ▼
┌─────────────────┐          ┌─────────────────┐          ┌─────────────────┐
│   Ingestion     │          │   GraphQL API   │          │   WebSocket     │
│   (Replicas)    │          │   (Replicas)    │          │   (Replicas)    │
└────────┬────────┘          └────────┬────────┘          └────────┬────────┘
         │                            │                            │
         ▼                            │                            │
┌─────────────────────────────────────┼────────────────────────────┼─────────┐
│                    Kafka Cluster (Multi-AZ)                      │         │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐             │         │
│  │ events  │  │sessions │  │journeys │  │ alerts  │             │         │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘             │         │
└─────────────────────────────────────┬───────────────────────────┴─────────┘
                                      │
         ┌────────────────────────────┼────────────────────────────┐
         │                            │                            │
         ▼                            ▼                            ▼
┌─────────────────┐          ┌─────────────────┐          ┌─────────────────┐
│  Session        │          │  Journey        │          │  Alert          │
│  Processor      │          │  Processor      │          │  Processor      │
│  (Consumer Grp) │          │  (Consumer Grp) │          │  (Consumer Grp) │
└────────┬────────┘          └────────┬────────┘          └────────┬────────┘
         │                            │                            │
         └────────────────────────────┼────────────────────────────┘
                                      │
         ┌────────────────────────────┼────────────────────────────┐
         │                            │                            │
         ▼                            ▼                            ▼
┌─────────────────┐          ┌─────────────────┐          ┌─────────────────┐
│   ClickHouse    │          │   PostgreSQL    │          │   Redis         │
│   Cluster       │          │   (Primary +    │          │   Cluster       │
│   (Sharded)     │          │    Replicas)    │          │                 │
└─────────────────┘          └─────────────────┘          └─────────────────┘
         │
         ▼
┌─────────────────┐
│   S3 / R2       │
│  (Snapshots,    │
│   Exports)      │
└─────────────────┘
```

### Additional Components

| Component | Technology | Purpose |
|-----------|------------|---------|
| API Gateway | Kong / Traefik | Rate limiting, auth, routing |
| Object Storage | S3 / Cloudflare R2 | Screenshots, exports |
| Monitoring | Prometheus + Grafana | Observability |
| Tracing | Jaeger / Tempo | Distributed tracing |

### Trade-offs

| Pros | Cons |
|------|------|
| Unlimited horizontal scaling | Complex operations |
| Multi-region capable | Requires dedicated SRE |
| High availability | Higher infrastructure cost |
| Feature-rich (alerts, exports) | Longer development cycles |

### Cost Estimate (AWS)

| Component | Configuration | Monthly Cost |
|-----------|---------------|--------------|
| Kafka (MSK) | kafka.m5.xlarge x6 | ~$2,000 |
| ClickHouse | r6g.2xlarge x4 (sharded) | ~$2,000 |
| PostgreSQL (RDS) | db.r6g.xlarge + replica | ~$600 |
| Redis (ElastiCache) | cache.r6g.xlarge cluster | ~$500 |
| S3 | 1TB storage + transfer | ~$100 |
| Kong | t3.medium x2 | ~$100 |
| Monitoring | Managed Prometheus | ~$200 |
| **Total** | | **~$5,500/mo** |

---

## Decision Matrix

Use this matrix to decide which architecture fits your needs:

| Factor | Starter | Growth | Scale |
|--------|---------|--------|-------|
| **Setup Time** | 1 day | 2-4 weeks | 2-3 months |
| **Team Size** | 1 dev | 2-3 devs | 5+ devs + SRE |
| **Monthly Infra** | $50-200 | $1,000-2,000 | $5,000+ |
| **Query Latency** | 200ms-2s | 50-200ms | <50ms |
| **Ingestion Rate** | 1K/sec | 10K/sec | 100K+/sec |
| **Data Retention** | 90 days | 1 year | Unlimited |
| **Real-time** | No | Basic | Full |
| **Multi-region** | No | No | Yes |

## Upgrade Triggers

### Starter → Growth

| Metric | Threshold | Action |
|--------|-----------|--------|
| Events/day | >1M | Add Kafka |
| Query p99 | >2s | Add ClickHouse |
| Active sessions | >10K concurrent | Add Redis |
| Team size | >3 engineers | Worth the complexity |

### Growth → Scale

| Metric | Threshold | Action |
|--------|-----------|--------|
| Events/day | >50M | Shard ClickHouse |
| Ingestion latency | >100ms p99 | Scale Kafka partitions |
| Revenue | >$500K ARR | Can afford dedicated SRE |
| Uptime SLA | >99.9% | Multi-region deployment |

## Migration Checklist

### Starter → Growth

- [ ] Deploy Kafka cluster
- [ ] Add Kafka producer to ingestion service
- [ ] Deploy Kafka consumers for processing
- [ ] Deploy ClickHouse
- [ ] Migrate event writes to ClickHouse
- [ ] Update analytics queries to use ClickHouse
- [ ] Deploy Redis
- [ ] Migrate session state to Redis
- [ ] Add WebSocket support
- [ ] Update monitoring dashboards
- [ ] Update runbooks

### Growth → Scale

- [ ] Deploy API Gateway
- [ ] Configure rate limiting and auth
- [ ] Shard ClickHouse by project_id
- [ ] Add PostgreSQL read replicas
- [ ] Deploy Redis cluster
- [ ] Set up S3 for snapshots
- [ ] Configure multi-AZ for all components
- [ ] Add distributed tracing
- [ ] Create runbooks for each component
- [ ] Train team on operations

## Summary

Start with the **Starter** architecture. It's sufficient for most early-stage products and can handle significant scale with proper PostgreSQL tuning. Only upgrade when you hit specific performance triggers—premature optimization adds complexity without benefits.

The cost savings of Starter over Scale are significant:

| Architecture | Monthly Cost | Annual Savings vs Scale |
|--------------|--------------|------------------------|
| Starter | ~$150 | $64,200 |
| Growth | ~$1,350 | $49,800 |
| Scale | ~$5,500 | - |

Invest those savings in product development until scale demands otherwise.
