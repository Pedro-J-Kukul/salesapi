# Testing Documentation for Sales API

This document provides comprehensive information about testing the Sales API, including setup, running tests, and understanding the testing infrastructure.

## Table of Contents

1. [Overview](#overview)
2. [Test Infrastructure](#test-infrastructure)
3. [Database Setup](#database-setup)
4. [Running Tests](#running-tests)
5. [Test Structure](#test-structure)
6. [Live Testing](#live-testing)
7. [Test Utilities](#test-utilities)
8. [Troubleshooting](#troubleshooting)

## Overview

The Sales API includes a comprehensive testing infrastructure that covers:

- **Integration Tests**: Test handlers with real database connections
- **Middleware Tests**: Test authentication, authorization, and other middleware
- **Live API Tests**: Test the running API using curl commands
- **Database Utilities**: Clean, seed, and reset test data

## Test Infrastructure

### Components

- **`cmd/api/handlers_test.go`**: Integration tests for all API handlers
- **`cmd/api/middleware_test.go`**: Tests for middleware functionality
- **`cmd/api/test_helpers.go`**: Helper functions for testing
- **`internal/data/testutils.go`**: Database testing utilities
- **`scripts/test_commands.sh`**: Live API testing with curl commands

### Test Database

Tests use a dedicated PostgreSQL database for isolation:

- **Database**: `sales`
- **User**: `sales`
- **Password**: `sales`
- **DSN**: `postgres://sales:sales@localhost:5432/sales?sslmode=disable`

## Database Setup

### Initial Setup

1. **Create the database and user**:
   ```bash
   make db/setup
   ```

   Or manually:
   ```bash
   psql -U postgres -f scripts/database_setup.sql
   ```

2. **Initialize the schema**:
   ```bash
   make db/schema
   ```

   Or manually:
   ```bash
   psql -U sales -d sales -f scripts/schema.sql
   ```

3. **Run migrations** (alternative to schema initialization):
   ```bash
   make migrate/up
   ```

### Database Reset

To reset the test database and clear all data:

```bash
make db/reset
```

This will:
- Truncate all tables
- Reset all identity sequences to 1
- Preserve the schema

## Running Tests

### All Tests

Run all tests with race detection:

```bash
make test
```

Or:

```bash
go test -v -race ./...
```

### Integration Tests Only

Run only the integration tests:

```bash
make test/integration
```

### Test Coverage

Generate a test coverage report:

```bash
make test/coverage
```

This will:
- Run all tests with coverage
- Generate an HTML coverage report
- Open the report in your default browser

### Live API Tests

Test the running API with curl commands:

1. **Start the API server**:
   ```bash
   make run
   ```

2. **In another terminal, run the test script**:
   ```bash
   make test/live
   ```

   Or directly:
   ```bash
   ./scripts/test_commands.sh
   ```

## Test Structure

### Handler Tests

Handler tests follow this pattern:

```go
func TestHandlerName(t *testing.T) {
    app, testUtils := newTestApp(t)
    defer testUtils.CleanDatabase()
    
    // Setup test data
    userID, token := setupAdminUser(t, app, testUtils)
    
    // Define test cases
    tests := []struct {
        name           string
        payload        map[string]interface{}
        headers        map[string]string
        expectedStatus int
    }{
        // Test cases...
    }
    
    // Run test cases
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            rr := makeRequest(t, app, "POST", "/v1/endpoint", tt.payload, tt.headers)
            checkResponseCode(t, tt.expectedStatus, rr.Code)
        })
    }
}
```

### Test Categories

1. **User Tests**:
   - User registration
   - User activation
   - User listing and retrieval
   - User updates and deletion

2. **Authentication Tests**:
   - Token creation
   - Invalid credentials
   - Inactive accounts

3. **Product Tests**:
   - Product CRUD operations
   - Pagination and filtering
   - Price range queries

4. **Sales Tests**:
   - Sale creation and management
   - Filtering by user and product
   - Quantity validation

5. **Permission Tests**:
   - Role-based access control
   - Permission enforcement
   - Unauthorized access attempts

6. **Middleware Tests**:
   - Authentication middleware
   - Authorization middleware
   - CORS handling
   - Metrics collection

## Live Testing

### Test Script Usage

The `test_commands.sh` script provides automated testing of all API endpoints:

```bash
# Use default API URL (http://localhost:4000)
./scripts/test_commands.sh

# Use custom API URL
API_URL=http://api.example.com:8080 ./scripts/test_commands.sh
```

### Manual Testing Examples

#### Register a User

```bash
curl -X POST http://localhost:4000/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@example.com",
    "password": "password123",
    "role": "staff"
  }'
```

#### Authenticate

```bash
curl -X POST http://localhost:4000/v1/tokens/authentication \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

#### Create a Product (with authentication)

```bash
curl -X POST http://localhost:4000/v1/products \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Laptop",
    "price": 999.99
  }'
```

#### List Products

```bash
curl http://localhost:4000/v1/products
```

#### Create a Sale

```bash
curl -X POST http://localhost:4000/v1/sales \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "product_id": 1,
    "quantity": 5
  }'
```

## Test Utilities

### Database Test Utilities

The `TestUtils` struct provides helper methods:

#### Clean Database

```go
testUtils.CleanDatabase()  // Truncates all tables and resets sequences
```

#### Seed Test Data

```go
// Seed permissions
testUtils.SeedTestPermissions()

// Create a test user
userID, err := testUtils.SeedTestUser(
    "test@example.com",
    "Test",
    "User",
    "staff",
    passwordHash,
    true,
)

// Create a test product
productID, err := testUtils.SeedTestProduct("Product Name", 99.99)

// Create a test sale
saleID, err := testUtils.SeedTestSale(userID, productID, 5)
```

#### Assign Permissions

```go
// Assign single permission
testUtils.AssignPermissionToUser(userID, "products:create")

// Assign multiple permissions
permissions := []string{"products:create", "products:read", "products:update"}
testUtils.AssignPermissionsToUser(userID, permissions)
```

### HTTP Test Helpers

#### Create Test App

```go
app, testUtils := newTestApp(t)
defer testUtils.CleanDatabase()
```

#### Make Requests

```go
// Without authentication
rr := makeRequest(t, app, "GET", "/v1/products", nil, nil)

// With authentication
headers := createAuthHeaders(token)
rr := makeRequest(t, app, "POST", "/v1/products", payload, headers)
```

#### Setup Users

```go
// Setup admin user with full permissions
adminID, adminToken := setupAdminUser(t, app, testUtils)

// Setup staff user with limited permissions
staffID, staffToken := setupStaffUser(t, app, testUtils)
```

## Troubleshooting

### Database Connection Issues

**Error**: `Failed to connect to test database`

**Solution**:
1. Ensure PostgreSQL is running
2. Check database credentials
3. Verify database exists: `psql -U postgres -l | grep sales`
4. Create database if needed: `make db/setup`

### Permission Errors

**Error**: `permission denied for table users`

**Solution**:
Run the database setup script to grant proper permissions:
```bash
make db/setup
```

### Test Failures After Schema Changes

**Error**: Tests fail after modifying database schema

**Solution**:
1. Reset the database: `make db/reset`
2. Re-run migrations: `make migrate/up`
3. Or reinitialize schema: `make db/schema`

### Sequence Reset Issues

**Error**: Duplicate key violations in tests

**Solution**:
The test utilities automatically reset sequences, but you can manually reset:
```bash
make db/reset
```

### Live Test Authentication Failures

**Error**: Authentication fails in live tests

**Solution**:
Ensure users are activated:
```sql
psql -U sales -d sales -c "UPDATE users SET is_active = true WHERE email IN ('admin@example.com', 'staff@example.com');"
```

### Port Already in Use

**Error**: `bind: address already in use`

**Solution**:
1. Check if another instance is running: `lsof -i :4000`
2. Kill the process or use a different port
3. Update the test script: `API_URL=http://localhost:8080 ./scripts/test_commands.sh`

## Best Practices

1. **Isolation**: Each test should be independent and not rely on other tests
2. **Cleanup**: Always clean the database before and after tests
3. **Fixtures**: Use helper functions to create consistent test data
4. **Assertions**: Check both status codes and response bodies
5. **Documentation**: Comment complex test scenarios
6. **Coverage**: Aim for high coverage of critical paths
7. **Performance**: Keep tests fast by minimizing database operations

## Continuous Integration

To run tests in CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
test:
  runs-on: ubuntu-latest
  services:
    postgres:
      image: postgres:15
      env:
        POSTGRES_PASSWORD: postgres
      options: >-
        --health-cmd pg_isready
        --health-interval 10s
        --health-timeout 5s
        --health-retries 5
  steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    - name: Setup database
      run: make db/setup
    - name: Run tests
      run: make test
```

## Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [API Documentation](../README.md)

## Support

For issues or questions:
- Check this documentation
- Review test examples in the codebase
- Open an issue in the repository
