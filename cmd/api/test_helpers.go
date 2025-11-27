// File: cmd/api/test_helpers.go
// Description: Test helper functions for API handler integration tests

package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Pedro-J-Kukul/salesapi/internal/data"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// testApp creates a test application instance with a real database connection
func newTestApp(t *testing.T) (*app, *data.TestUtils) {
	t.Helper()

	// Use environment variable or default test database
	dsn := "postgres://sales:sales@localhost:5432/sales?sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	// Create test utilities
	testUtils := data.NewTestUtils(db)

	// Clean the database before each test
	err = testUtils.CleanDatabase()
	if err != nil {
		t.Fatalf("Failed to clean database: %v", err)
	}

	// Seed test permissions
	err = testUtils.SeedTestPermissions()
	if err != nil {
		t.Fatalf("Failed to seed test permissions: %v", err)
	}

	cfg := config{
		port: 4000,
		env:  "test",
		db: struct {
			dsn          string
			maxOpenConns int
			maxIdleConns int
			maxIdleTime  time.Duration
		}{
			dsn:          dsn,
			maxOpenConns: 25,
			maxIdleConns: 25,
			maxIdleTime:  time.Minute,
		},
	}

	testApp := &app{
		config: cfg,
		logger: setUpLogger("test"),
		models: data.NewModels(db),
	}

	// Register cleanup to close database connection
	t.Cleanup(func() {
		db.Close()
	})

	return testApp, testUtils
}

// executeRequest executes an HTTP request and returns the response recorder
func executeRequest(app *app, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	app.routes().ServeHTTP(rr, req)
	return rr
}

// makeRequest creates and executes an HTTP request
func makeRequest(t *testing.T, app *app, method, url string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set default content type
	req.Header.Set("Content-Type", "application/json")

	// Add any additional headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return executeRequest(app, req)
}

// parseJSONResponse parses a JSON response into a destination struct
func parseJSONResponse(t *testing.T, rr *httptest.ResponseRecorder, dest interface{}) {
	t.Helper()

	err := json.NewDecoder(rr.Body).Decode(dest)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v. Body: %s", err, rr.Body.String())
	}
}

// checkResponseCode checks if the response has the expected status code
func checkResponseCode(t *testing.T, expected, actual int) {
	t.Helper()

	if expected != actual {
		t.Errorf("Expected status code %d, got %d", expected, actual)
	}
}

// createTestUser creates a test user with the given role and returns user ID and password
func createTestUser(t *testing.T, testUtils *data.TestUtils, email, firstName, lastName, role string, isActive bool) (int64, string) {
	t.Helper()

	password := "Password123!"
	user := &data.User{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		IsActive:  isActive,
	}

	err := user.Password.Set(password)
	if err != nil {
		t.Fatalf("Failed to set password: %v", err)
	}

	// Get the password hash by using a workaround - we'll query it back
	// or use bcrypt directly
	passwordHash, err := getBcryptHash(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	userID, err := testUtils.SeedTestUser(user.Email, user.FirstName, user.LastName, user.Role, passwordHash, user.IsActive)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return userID, password
}

// getBcryptHash generates a bcrypt hash for a password
func getBcryptHash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

// authenticateUser creates an authentication token for a user
func authenticateUser(t *testing.T, app *app, email, password string) string {
	t.Helper()

	body := map[string]string{
		"email":    email,
		"password": password,
	}

	rr := makeRequest(t, app, "POST", "/v1/tokens/authentication", body, nil)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Authentication failed: status %d, body: %s", rr.Code, rr.Body.String())
	}

	var response struct {
		AuthenticationToken string `json:"authentication_token"`
	}

	parseJSONResponse(t, rr, &response)

	if response.AuthenticationToken == "" {
		t.Fatal("No authentication token returned")
	}

	return response.AuthenticationToken
}

// createAuthHeaders creates headers with Bearer token authentication
func createAuthHeaders(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
	}
}

// setupAdminUser creates an admin user with all permissions and returns the auth token
func setupAdminUser(t *testing.T, app *app, testUtils *data.TestUtils) (int64, string) {
	t.Helper()

	userID, password := createTestUser(t, testUtils, "admin@example.com", "Admin", "User", "admin", true)

	// Assign admin permissions
	adminPerms := []string{
		"sales:create", "sales:read", "sales:update", "sales:delete",
		"products:create", "products:read", "products:update", "products:delete",
		"users:create", "users:read", "users:update", "users:delete",
	}

	err := testUtils.AssignPermissionsToUser(userID, adminPerms)
	if err != nil {
		t.Fatalf("Failed to assign admin permissions: %v", err)
	}

	token := authenticateUser(t, app, "admin@example.com", password)

	return userID, token
}

// setupStaffUser creates a staff user with limited permissions and returns the auth token
func setupStaffUser(t *testing.T, app *app, testUtils *data.TestUtils) (int64, string) {
	t.Helper()

	userID, password := createTestUser(t, testUtils, "staff@example.com", "Staff", "User", "staff", true)

	// Assign staff permissions
	staffPerms := []string{
		"products:read",
		"sales:read",
	}

	err := testUtils.AssignPermissionsToUser(userID, staffPerms)
	if err != nil {
		t.Fatalf("Failed to assign staff permissions: %v", err)
	}

	token := authenticateUser(t, app, "staff@example.com", password)

	return userID, token
}
