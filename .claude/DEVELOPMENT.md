# Development Guide

Commands, tools, and setup instructions for local development.

## Prerequisites

- **Go** 1.21+
- **PostgreSQL** 14+
- **golang-migrate** CLI
- **golangci-lint**
- **Docker** (optional, for containerized development)

## Initial Setup

### 1. Clone Repository

```bash
git clone https://github.com/mlgaray/ecommerce_api.git
cd ecommerce_api
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Create Environment File

```bash
make create-env-file
```

This creates `.env.develop` with:

```env
DB_USER=postgres
DB_PASSWORD=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ecommerce_dev
ENVIRONMENT=develop
```

Adjust values as needed for your local PostgreSQL instance.

### 4. Run Migrations

```bash
# Run all migrations
make migrate-up

# Optionally, load seed data
make migrate-up-seeds
```

### 5. Run Application

```bash
go run main.go
```

Server starts on `http://localhost:8080`

## Development Commands

### Database Operations

```bash
# Create new migration
make migrate-create MIGRATION_NAME="create_table_products"

# Run migrations (up)
make migrate-up

# Rollback migrations (down)
make migrate-down

# Rollback one step
make migrate-down-one

# Force migration version (use with caution)
migrate -path database/migrations -database "postgresql://..." force VERSION
```

### Seed Data

```bash
# Create seed migration
make migrate-create-seeds MIGRATION_NAME="seed_users"

# Run seeds
make migrate-up-seeds

# Rollback seeds
make migrate-down-seeds
```

### Code Quality

```bash
# Run all code quality checks (format + lint)
make code-quality

# Format code
make fmt

# Run linter
make lint

# Run linter with auto-fix
make lint-fix
```

**Note**: `golangci-lint` uses `--out-format=tab` for clickable errors in terminal.

### Testing

```bash
# Run all unit tests
go test ./...

# Run unit tests with coverage
go test -cover ./...

# Run unit tests with verbose output
go test -v ./...

# Run integration tests (BDD)
cd tests/integration && go test

# Run integration tests with verbose
cd tests/integration && go test -v

# Generate mocks
make generate-mocks
```

### Build

```bash
# Build binary
go build -o bin/ecommerce_api main.go

# Run binary
./bin/ecommerce_api
```

### Docker (if configured)

```bash
# Build Docker image
docker build -t ecommerce_api .

# Run with Docker Compose
docker-compose up

# Run in background
docker-compose up -d

# Stop containers
docker-compose down

# View logs
docker-compose logs -f
```

## Project Structure

```
.
├── .claude/                    # Claude Code documentation
│   ├── ARCHITECTURE.md
│   ├── ERRORS.md
│   ├── LOGGING.md
│   ├── LAYERS.md
│   ├── TESTING.md
│   ├── DATABASE.md
│   └── DEVELOPMENT.md
├── database/
│   └── migrations/             # SQL migrations
│       └── seeds/              # Seed data
├── internal/
│   ├── core/                   # Core domain layer
│   │   ├── entities/
│   │   ├── errors/
│   │   ├── models/
│   │   └── ports/
│   ├── application/            # Application layer
│   │   ├── services/
│   │   └── usecases/
│   └── infraestructure/        # Infrastructure layer
│       └── adapters/
│           ├── auth/
│           ├── http/
│           ├── logs/
│           └── repositories/
├── tests/
│   └── integration/            # Integration tests (BDD)
│       ├── features/
│       └── steps/
├── .env.develop                # Local environment variables
├── .golangci.yml               # Linter configuration
├── .mockery.yml                # Mock generation configuration
├── CLAUDE.md                   # Main documentation index
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
├── main.go                     # Application entry point
└── Makefile                    # Development commands
```

## Key Technologies

- **Web Framework**: Gorilla Mux
- **Database**: PostgreSQL with lib/pq driver
- **Migrations**: golang-migrate
- **Authentication**: JWT with golang-jwt/jwt/v5
- **Logging**: Logrus with Lumberjack file rotation
- **Dependency Injection**: Uber FX
- **Testing**: Testify, Mockery, Cucumber/Godog
- **CORS**: rs/cors

## Configuration

### Environment Variables

Create `.env.develop` file:

```env
# Database
DB_USER=postgres
DB_PASSWORD=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ecommerce_dev

# Application
ENVIRONMENT=develop
PORT=8080

# JWT (optional)
JWT_SECRET=your-secret-key
JWT_EXPIRATION=2h
```

### Linter Configuration

Edit `.golangci.yml` to customize linting rules:

```yaml
linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - goimports
```

### Mock Configuration

Edit `.mockery.yml` to configure mock generation:

```yaml
with-expecter: true
packages:
  github.com/mlgaray/ecommerce_api/internal/core/ports:
    interfaces:
      ProductRepository:
      UserRepository:
      AuthService:
```

## Common Workflows

### Adding a New Feature

1. **Create migration** (if database changes needed)
   ```bash
   make migrate-create MIGRATION_NAME="add_column_to_products"
   ```

2. **Define interface** in `internal/core/ports/`
   ```go
   type ProductRepository interface {
       Create(ctx context.Context, product *models.Product) (*models.Product, error)
   }
   ```

3. **Implement in infrastructure layer**
   ```go
   // internal/infraestructure/adapters/repositories/postgresql/product_repository.go
   func (r *ProductRepository) Create(ctx context.Context, product *models.Product) (*models.Product, error) {
       // implementation
   }
   ```

4. **Add business logic in service**
   ```go
   // internal/application/services/product_service.go
   func (s *ProductService) CreateProduct(ctx context.Context, product *models.Product) (*models.Product, error) {
       // business logic
   }
   ```

5. **Create use case**
   ```go
   // internal/application/usecases/product/create.go
   func (uc *CreateProductUseCase) Execute(ctx context.Context, product *models.Product) (*models.Product, error) {
       return uc.productService.CreateProduct(ctx, product)
   }
   ```

6. **Add HTTP handler**
   ```go
   // internal/infraestructure/adapters/http/product_handler.go
   func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
       // parse request, call use case, return response
   }
   ```

7. **Register in DI container** (`main.go`)
   ```go
   fx.Provide(
       usecases.NewCreateProductUseCase,
   ),
   ```

8. **Generate mocks** (if needed)
   ```bash
   make generate-mocks
   ```

9. **Write tests**
   - Unit tests for service logic
   - Integration tests for full flow

10. **Run code quality checks**
    ```bash
    make code-quality
    ```

### Making a Database Change

1. **Create migration**
   ```bash
   make migrate-create MIGRATION_NAME="add_index_products_name"
   ```

2. **Edit up migration** (`database/migrations/NNNN_add_index_products_name.up.sql`)
   ```sql
   CREATE INDEX idx_products_name ON products(name);
   ```

3. **Edit down migration** (`database/migrations/NNNN_add_index_products_name.down.sql`)
   ```sql
   DROP INDEX IF EXISTS idx_products_name;
   ```

4. **Apply migration**
   ```bash
   make migrate-up
   ```

5. **Test rollback** (optional, in development)
   ```bash
   make migrate-down-one
   make migrate-up
   ```

## Troubleshooting

### Migration Issues

```bash
# Check current migration version
migrate -path database/migrations -database "postgresql://user:pass@localhost:5432/dbname?sslmode=disable" version

# Force to specific version (last resort)
migrate -path database/migrations -database "postgresql://..." force VERSION
```

### Database Connection Issues

- Verify PostgreSQL is running
- Check `.env.develop` credentials
- Ensure database exists: `createdb ecommerce_dev`
- Test connection: `psql -U postgres -d ecommerce_dev`

### Linter Issues

```bash
# Update golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Clear cache
golangci-lint cache clean
```

### Mock Generation Issues

```bash
# Install/update mockery
go install github.com/vektra/mockery/v2@latest

# Verify mockery is in PATH
which mockery

# Manually generate specific mock
mockery --name=ProductRepository --dir=internal/core/ports --output=mocks
```

## Git Workflow

```bash
# Create feature branch
git checkout -b feature/add-product-search

# Make changes and commit
git add .
git commit -m "feat: add product search endpoint"

# Push and create PR
git push origin feature/add-product-search
```

## Monitoring & Observability (if configured)

- **Prometheus**: Metrics at `/metrics`
- **Grafana**: Dashboards for visualization
- **Loki**: Log aggregation
- **Promtail**: Log shipping

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Gorilla Mux](https://github.com/gorilla/mux)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [Testify](https://github.com/stretchr/testify)
- [Mockery](https://github.com/vektra/mockery)
- [Cucumber/Godog](https://github.com/cucumber/godog)
