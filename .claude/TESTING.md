# Testing Strategy

This project uses a comprehensive testing approach with Unit Tests and Integration Tests (BDD).

## Unit Tests

### Framework

- **testify/assert** - Assertions
- **mockery** - Mock generation
- Pattern: **Arrange-Act-Assert**

### Test Location

Tests are located alongside source files with `*_test.go` suffix:

```
internal/
  application/
    services/
      pagination_service.go
      pagination_service_test.go  ← Test file
```

### Mock Generation

Mocks are automatically generated using Mockery based on `.mockery.yml` configuration:

```bash
# Generate all mocks
make generate-mocks
```

### Unit Test Example

```go
package services

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/mlgaray/ecommerce_api/internal/core/models"
)

func TestPaginationService_BuildCursorPagination(t *testing.T) {
    t.Run("when items equal limit then returns cursor and hasMore true", func(t *testing.T) {
        // Arrange
        limit := 3
        products := []*models.Product{
            {ID: 1, Name: "Product 1"},
            {ID: 2, Name: "Product 2"},
            {ID: 3, Name: "Product 3"},
        }
        service := NewPaginationService[*models.Product]()

        // Act
        nextCursor, hasMore := service.BuildCursorPagination(products, limit)

        // Assert
        assert.Equal(t, 3, nextCursor)
        assert.True(t, hasMore)
    })

    t.Run("when items less than limit then returns zero cursor and hasMore false", func(t *testing.T) {
        // Arrange
        limit := 5
        products := []*models.Product{
            {ID: 1, Name: "Product 1"},
            {ID: 2, Name: "Product 2"},
        }
        service := NewPaginationService[*models.Product]()

        // Act
        nextCursor, hasMore := service.BuildCursorPagination(products, limit)

        // Assert
        assert.Equal(t, 0, nextCursor)
        assert.False(t, hasMore)
    })

    t.Run("when empty list then returns zero cursor and hasMore false", func(t *testing.T) {
        // Arrange
        limit := 10
        products := []*models.Product{}
        service := NewPaginationService[*models.Product]()

        // Act
        nextCursor, hasMore := service.BuildCursorPagination(products, limit)

        // Assert
        assert.Equal(t, 0, nextCursor)
        assert.False(t, hasMore)
    })
}
```

### Test Naming Convention

Use descriptive test names that express:
1. **When** - The condition being tested
2. **Then** - The expected outcome

```go
t.Run("when user email is empty then returns bad request error", func(t *testing.T) {
    // test code
})

t.Run("when all fields are valid then returns no error", func(t *testing.T) {
    // test code
})
```

### Mock Usage vs Test Fixtures

**When to use Mocks (mockery)**:
- Complex interfaces with multiple methods
- External dependencies (repositories, services)
- Need to verify method calls

**When to use Test Fixtures**:
- Simple data structures
- Domain models
- Test-specific implementations

```go
// ✅ GOOD: Test fixture for simple data
type mockIdentifiable struct {
    ID   int
    Name string
}

func (m *mockIdentifiable) GetID() int {
    return m.ID
}

// ✅ GOOD: Mockery for complex interface
mockRepo := mocks.NewProductRepository(t)
mockRepo.On("GetByID", mock.Anything, 1).Return(product, nil)
```

## Integration Tests

### Framework

- **Cucumber/Godog** - BDD framework
- **sqlmock** - Database mocking
- Location: `tests/integration/`

### Structure

```
tests/
  integration/
    features/
      auth.feature
      product.feature
      get_products_by_shop_id.feature
    steps/
      auth_steps.go
      product_steps.go
      get_products_by_shop_id_steps.go
      test_context.go
    main_test.go
```

### Running Integration Tests

```bash
# Run all integration tests
cd tests/integration && go test

# Run with verbose output
cd tests/integration && go test -v
```

### Feature File Example

```gherkin
# features/auth.feature
Feature: Authentication
  As a user
  I want to sign in to the system
  So that I can access protected resources

  Scenario: Successful sign in with valid credentials
    Given a user with email "john@example.com" and password "password123"
    When I sign in with email "john@example.com" and password "password123"
    Then the response status code should be 200
    And the response should contain a valid JWT token

  Scenario: Failed sign in with invalid credentials
    Given a user with email "john@example.com" and password "password123"
    When I sign in with email "john@example.com" and password "wrong_password"
    Then the response status code should be 401
    And the response should contain error "invalid_credentials"
```

### Step Definitions Example

```go
package steps

import (
    "context"
    "github.com/cucumber/godog"
)

func (tc *TestContext) aUserWithEmailAndPassword(ctx context.Context, email, password string) error {
    // Setup test user
    tc.testUser = &models.User{
        Email:    email,
        Password: password,
    }

    // Mock repository behavior
    tc.mockUserRepo.On("GetByEmail", mock.Anything, email).Return(tc.testUser, nil)

    return nil
}

func (tc *TestContext) iSignInWithEmailAndPassword(ctx context.Context, email, password string) error {
    // Create sign in request
    request := contracts.SignInRequest{
        Email:    email,
        Password: password,
    }

    // Execute through handler
    tc.response = tc.executeSignIn(request)

    return nil
}

func (tc *TestContext) theResponseStatusCodeShouldBe(ctx context.Context, expectedStatus int) error {
    if tc.response.StatusCode != expectedStatus {
        return fmt.Errorf("expected status %d, got %d", expectedStatus, tc.response.StatusCode)
    }
    return nil
}
```

### Test Context Pattern

Shared context between steps for scenario state:

```go
package steps

type TestContext struct {
    // Mocks
    mockUserRepo    *mocks.ProductRepository
    mockAuthService *mocks.AuthService

    // Test data
    testUser     *models.User
    testProduct  *models.Product

    // Response
    response     *http.Response
    responseBody map[string]interface{}

    // Use cases
    signInUseCase  ports.SignInUseCase
    signUpUseCase  ports.SignUpUseCase
}

func NewTestContext() *TestContext {
    return &TestContext{
        mockUserRepo:    mocks.NewUserRepository(t),
        mockAuthService: mocks.NewAuthService(t),
    }
}
```

### BDD Best Practices

1. **Use Given-When-Then** structure in scenarios
2. **Keep scenarios focused** - One behavior per scenario
3. **Use descriptive scenario names**
4. **Avoid technical details** in feature files
5. **Reuse step definitions** across features

## Test Commands

```bash
# Unit tests
go test ./...

# Unit tests with coverage
go test -cover ./...

# Integration tests
cd tests/integration && go test

# Generate mocks
make generate-mocks

# Run all tests (unit + integration)
make test-all  # if configured in Makefile
```

## What to Test

### Unit Tests

✅ **DO test:**
- Service business logic
- Domain model methods
- Validation rules
- Error handling
- Edge cases

❌ **DON'T test:**
- Database queries (use integration tests)
- HTTP routing (use integration tests)
- External service calls (use integration tests)

### Integration Tests

✅ **DO test:**
- Full request/response cycles
- Database interactions
- Transaction handling
- Error flows
- Authentication/authorization

❌ **DON'T test:**
- Individual function logic (use unit tests)
- Complex business calculations (use unit tests)

## Coverage Goals

- **Unit Tests**: Aim for 80%+ coverage of business logic
- **Integration Tests**: Cover all critical user flows
- Focus on **meaningful tests**, not just coverage numbers

## Test Organization

```go
// ✅ GOOD: Organize by behavior
func TestProductService(t *testing.T) {
    t.Run("GetByID", func(t *testing.T) {
        t.Run("when product exists then returns product", func(t *testing.T) {})
        t.Run("when product not found then returns error", func(t *testing.T) {})
    })

    t.Run("Create", func(t *testing.T) {
        t.Run("when valid data then creates product", func(t *testing.T) {})
        t.Run("when invalid price then returns validation error", func(t *testing.T) {})
    })
}

// ❌ BAD: Flat structure
func TestGetByID(t *testing.T) {}
func TestGetByIDNotFound(t *testing.T) {}
func TestCreate(t *testing.T) {}
func TestCreateInvalidPrice(t *testing.T) {}
```

## Continuous Integration

Tests run automatically on:
- Pull requests
- Push to main branch
- Pre-commit hooks (optional)

Ensure all tests pass before merging code.
