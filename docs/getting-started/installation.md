# Installation

ProductGraph can be deployed in several configurations depending on your scale requirements.

## Development

### From Source

```bash
# Clone the repository
git clone https://github.com/plexusone/productgraph.git
cd productgraph

# Install dependencies
go mod download

# Start PostgreSQL
make docker-up

# Build and run
make run-ingestion
```

### Using Docker Compose

```bash
# Start all services
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

## Production Deployment

### Prerequisites

- PostgreSQL 16+ with the following extensions:
    - `uuid-ossp`
    - `pgcrypto`
- Go 1.22+ (for building from source)

### Database Setup

1. Create the database:

```sql
CREATE DATABASE productgraph;
```

2. Run the initialization script to set up RLS functions:

```sql
-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create roles
CREATE ROLE service_role NOLOGIN;
CREATE ROLE app_role NOLOGIN;

-- Function to set current org context
CREATE OR REPLACE FUNCTION set_current_org(org_id uuid)
RETURNS void AS $$
BEGIN
    PERFORM set_config('app.current_org_id', org_id::text, false);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

3. Run Ent migrations:

```bash
make migrate
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://pg:pg@localhost:5432/productgraph?sslmode=disable` |
| `DEBUG` | Enable debug logging | `false` |

### Build the Binary

```bash
make build
```

The binary is output to `bin/ingestion`.

### Run the Service

```bash
./bin/ingestion -port 8080
```

With debug logging:

```bash
./bin/ingestion -port 8080 -debug
```

## Health Checks

The service exposes two health endpoints:

| Endpoint | Purpose |
|----------|---------|
| `GET /health` | Basic health check |
| `GET /ready` | Readiness check (database connectivity) |

Example health check script:

```bash
#!/bin/bash
curl -sf http://localhost:8080/health || exit 1
```

## Scaling

See the [Scaling Guide](../architecture/scaling.md) for production deployment patterns:

- **Starter**: Single PostgreSQL instance
- **Growth**: Add Kafka, ClickHouse, Redis
- **Scale**: Sharded, multi-region deployment

## Next Steps

- [Quick Start](quickstart.md) - Send your first events
- [Architecture Overview](../architecture/overview.md) - Understand the system
- [Scaling Guide](../architecture/scaling.md) - Plan for growth
