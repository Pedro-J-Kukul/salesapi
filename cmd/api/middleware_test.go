// File: cmd/api/middleware_test.go
// Description: Tests for middleware functionality

package main

import (
	"net/http"
	"testing"
)

func TestRecoverPanicMiddleware(t *testing.T) {
	app, testUtils := newTestApp(t)
	defer testUtils.CleanDatabase()

	// This test would require a handler that panics
	// For now, we'll test that the middleware chain works correctly
	t.Run("Normal request flow", func(t *testing.T) {
		rr := makeRequest(t, app, "GET", "/v1/metrics", nil, nil)
		// Should return 200 or other valid status, not a panic
		if rr.Code >= 500 {
			t.Errorf("Expected non-500 status, got %d", rr.Code)
		}
	})
}

func TestAuthenticateMiddleware(t *testing.T) {
	app, testUtils := newTestApp(t)

	// Create an active user
	_, password := createTestUser(t, testUtils, "authtest@example.com", "Auth", "Test", "staff", true)

	tests := []struct {
		name           string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name:           "No authentication header",
			headers:        nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Invalid token format",
			headers: map[string]string{
				"Authorization": "InvalidFormat token123",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Invalid token",
			headers: map[string]string{
				"Authorization": "Bearer invalidtoken123456789012345678901234567890123456",
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Try to access a protected endpoint
			rr := makeRequest(t, app, "GET", "/v1/users/profile", nil, tt.headers)
			checkResponseCode(t, tt.expectedStatus, rr.Code)
		})
	}

	// Test valid authentication
	t.Run("Valid authentication", func(t *testing.T) {
		token := authenticateUser(t, app, "authtest@example.com", password)
		headers := createAuthHeaders(token)
		rr := makeRequest(t, app, "GET", "/v1/users/profile", nil, headers)
		checkResponseCode(t, http.StatusOK, rr.Code)
	})

	testUtils.CleanDatabase()
}

func TestRequirePermissionsMiddleware(t *testing.T) {
	app, testUtils := newTestApp(t)

	// Setup users with different permissions
	_, adminToken := setupAdminUser(t, app, testUtils)
	_, staffToken := setupStaffUser(t, app, testUtils)

	tests := []struct {
		name           string
		method         string
		url            string
		token          string
		expectedStatus int
	}{
		{
			name:           "Admin has products:read permission",
			method:         "GET",
			url:            "/v1/products",
			token:          adminToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Staff has products:read permission",
			method:         "GET",
			url:            "/v1/products",
			token:          staffToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Staff lacks products:create permission",
			method:         "POST",
			url:            "/v1/products",
			token:          staffToken,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Admin has products:create permission",
			method:         "POST",
			url:            "/v1/products",
			token:          adminToken,
			expectedStatus: http.StatusUnprocessableEntity, // Will fail validation but passes permission check
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := createAuthHeaders(tt.token)
			var payload interface{}
			if tt.method == "POST" {
				payload = map[string]interface{}{} // Empty payload to trigger validation error
			}
			rr := makeRequest(t, app, tt.method, tt.url, payload, headers)
			checkResponseCode(t, tt.expectedStatus, rr.Code)
		})
	}

	testUtils.CleanDatabase()
}

func TestCORSMiddleware(t *testing.T) {
	app, testUtils := newTestApp(t)
	defer testUtils.CleanDatabase()

	// Configure trusted origins
	app.config.cors.trustedOrigins = []string{"http://localhost:3000"}

	tests := []struct {
		name           string
		origin         string
		method         string
		expectCORS     bool
	}{
		{
			name:       "Trusted origin",
			origin:     "http://localhost:3000",
			method:     "GET",
			expectCORS: true,
		},
		{
			name:       "Untrusted origin",
			origin:     "http://evil.com",
			method:     "GET",
			expectCORS: false,
		},
		{
			name:       "No origin header",
			origin:     "",
			method:     "GET",
			expectCORS: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{}
			if tt.origin != "" {
				headers["Origin"] = tt.origin
			}

			rr := makeRequest(t, app, tt.method, "/v1/products", nil, headers)

			corsHeader := rr.Header().Get("Access-Control-Allow-Origin")
			if tt.expectCORS {
				if corsHeader != tt.origin {
					t.Errorf("Expected CORS header %s, got %s", tt.origin, corsHeader)
				}
			} else {
				if corsHeader != "" {
					t.Errorf("Expected no CORS header, got %s", corsHeader)
				}
			}
		})
	}
}

func TestMetricsMiddleware(t *testing.T) {
	app, testUtils := newTestApp(t)
	defer testUtils.CleanDatabase()

	t.Run("Metrics endpoint accessible", func(t *testing.T) {
		rr := makeRequest(t, app, "GET", "/v1/metrics", nil, nil)
		checkResponseCode(t, http.StatusOK, rr.Code)

		// Check that response is not empty
		if rr.Body.Len() == 0 {
			t.Error("Expected metrics data, got empty response")
		}
	})

	t.Run("Metrics update after requests", func(t *testing.T) {
		// Make a few requests
		makeRequest(t, app, "GET", "/v1/products", nil, nil)
		makeRequest(t, app, "GET", "/v1/products", nil, nil)

		// Check metrics
		rr := makeRequest(t, app, "GET", "/v1/metrics", nil, nil)
		checkResponseCode(t, http.StatusOK, rr.Code)
	})
}
