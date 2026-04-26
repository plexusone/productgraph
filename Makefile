.PHONY: all build test lint clean docker-up docker-down docker-logs run generate migrate

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint

# Binary names
INGESTION_BIN=bin/ingestion

# Database
DATABASE_URL=postgres://pg:pg@localhost:5432/productgraph?sslmode=disable

all: build

# Build all binaries
build:
	$(GOBUILD) -o $(INGESTION_BIN) ./cmd/ingestion

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-cover:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	$(GOLINT) run ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Go mod tidy
tidy:
	$(GOMOD) tidy

# Generate Ent code
generate:
	$(GOCMD) generate ./ent

# Generate new Ent migration
migrate-new:
	$(GOCMD) run -mod=mod entgo.io/ent/cmd/ent migrate diff --dir "file://ent/migrate/migrations" --dev-url "$(DATABASE_URL)"

# Apply database migrations
migrate:
	$(GOCMD) run -mod=mod ariga.io/atlas/cmd/atlas migrate apply --dir "file://ent/migrate/migrations" --url "$(DATABASE_URL)"

# Docker Compose commands
docker-up:
	docker compose -f deployments/docker/docker-compose.yml up -d

docker-down:
	docker compose -f deployments/docker/docker-compose.yml down

docker-logs:
	docker compose -f deployments/docker/docker-compose.yml logs -f

docker-clean:
	docker compose -f deployments/docker/docker-compose.yml down -v

# Run ingestion service locally
run-ingestion:
	$(GOBUILD) -o $(INGESTION_BIN) ./cmd/ingestion
	./$(INGESTION_BIN) -debug

# Run ingestion service with hot reload (requires air)
dev-ingestion:
	air -c .air.toml

# Database commands
db-reset:
	docker compose -f deployments/docker/docker-compose.yml down -v
	docker compose -f deployments/docker/docker-compose.yml up -d postgres
	@echo "Waiting for PostgreSQL to initialize..."
	@sleep 5
	@echo "PostgreSQL ready"

db-shell:
	docker exec -it pg-postgres psql -U pg -d productgraph

# Check service health
health:
	@echo "Checking service health..."
	@curl -s http://localhost:8080/health | jq . || echo "Ingestion service not running"
	@echo ""
	@docker exec pg-postgres pg_isready -U pg -d productgraph 2>/dev/null || echo "PostgreSQL not running"

# Help
help:
	@echo "ProductGraph Development Commands"
	@echo ""
	@echo "Build & Test:"
	@echo "  make build         - Build all binaries"
	@echo "  make test          - Run tests"
	@echo "  make test-cover    - Run tests with coverage"
	@echo "  make lint          - Run linter"
	@echo "  make clean         - Clean build artifacts"
	@echo ""
	@echo "Code Generation:"
	@echo "  make generate      - Generate Ent code from schemas"
	@echo "  make migrate-new   - Generate new database migration"
	@echo "  make migrate       - Apply database migrations"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up     - Start PostgreSQL"
	@echo "  make docker-down   - Stop PostgreSQL"
	@echo "  make docker-logs   - Follow service logs"
	@echo "  make docker-clean  - Stop and remove volumes"
	@echo ""
	@echo "Development:"
	@echo "  make run-ingestion - Build and run ingestion service"
	@echo "  make dev-ingestion - Run with hot reload (requires air)"
	@echo "  make health        - Check service health"
	@echo ""
	@echo "Database:"
	@echo "  make db-reset      - Reset database with fresh schema"
	@echo "  make db-shell      - Open PostgreSQL shell"
