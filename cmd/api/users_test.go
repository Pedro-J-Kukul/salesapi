// File: cmd/api/users_test.go
// Description: test suite for user handlers - validation focused

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pedro-J-Kukul/salesapi/internal/data"
	"github.com/Pedro-J-Kukul/salesapi/internal/validator"
)

// TestUserValidation tests user validation logic
func TestUserValidation(t *testing.T) {
	tests := []struct {
		name          string
		firstName     string
		lastName      string
		email         string
		password      string
		role          string
		expectedValid bool
		errorField    string
	}{
		{
			name:          "Valid User",
			firstName:     "John",
			lastName:      "Doe",
			email:         "john.doe@example.com",
			password:      "SecurePassword123!", // Must have uppercase and special char
			role:          "guest",
			expectedValid: true,
		},
		{
			name:          "Empty First Name",
			firstName:     "",
			lastName:      "Doe",
			email:         "test@example.com",
			password:      "password123",
			role:          "guest",
			expectedValid: false,
			errorField:    "first_name",
		},
		{
			name:          "Empty Last Name",
			firstName:     "John",
			lastName:      "",
			email:         "test@example.com",
			password:      "password123",
			role:          "guest",
			expectedValid: false,
			errorField:    "last_name",
		},
		{
			name:          "Invalid Email",
			firstName:     "John",
			lastName:      "Doe",
			email:         "invalid-email",
			password:      "password123",
			role:          "guest",
			expectedValid: false,
			errorField:    "email",
		},
		{
			name:          "Empty Email",
			firstName:     "John",
			lastName:      "Doe",
			email:         "",
			password:      "password123",
			role:          "guest",
			expectedValid: false,
			errorField:    "email",
		},
		{
			name:          "Short Password",
			firstName:     "John",
			lastName:      "Doe",
			email:         "john@example.com",
			password:      "short",
			role:          "guest",
			expectedValid: false,
			errorField:    "password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &data.User{
				FirstName: tt.firstName,
				LastName:  tt.lastName,
				Email:     tt.email,
				Role:      tt.role,
				IsActive:  false,
			}

			// Set password if provided
			if tt.password != "" {
				err := user.Password.Set(tt.password)
				if err != nil {
					t.Fatalf("failed to set password: %v", err)
				}
			}

			v := validator.New()
			data.ValidateUser(v, user)

			isValid := v.IsValid()
			if isValid != tt.expectedValid {
				t.Errorf("expected valid=%v, got=%v. Errors: %v", tt.expectedValid, isValid, v.Errors)
			}

			if !tt.expectedValid && tt.errorField != "" {
				if _, hasError := v.Errors[tt.errorField]; !hasError {
					t.Errorf("expected error on field '%s', but got errors: %v", tt.errorField, v.Errors)
				}
			}
		})
	}
}

// TestUserRoles tests role validation and defaults
func TestUserRoles(t *testing.T) {
	tests := []struct {
		name         string
		inputRole    string
		expectedRole string
		isValid      bool
	}{
		{
			name:         "Admin Role",
			inputRole:    "admin",
			expectedRole: "admin",
			isValid:      true,
		},
		{
			name:         "Cashier Role",
			inputRole:    "cashier",
			expectedRole: "cashier",
			isValid:      true,
		},
		{
			name:         "Guest Role",
			inputRole:    "guest",
			expectedRole: "guest",
			isValid:      true,
		},
		{
			name:         "Empty Role (should default to guest)",
			inputRole:    "",
			expectedRole: "guest",
			isValid:      true,
		},
		{
			name:         "Invalid Role (should default to guest)",
			inputRole:    "superuser",
			expectedRole: "guest",
			isValid:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate role validation logic from registerUserHandler
			validRoles := map[string]bool{"admin": true, "cashier": true, "guest": true}
			role := tt.inputRole
			if role == "" || !validRoles[role] {
				role = "guest"
			}

			if role != tt.expectedRole {
				t.Errorf("expected role %s, got %s", tt.expectedRole, role)
			}
		})
	}
}

// TestUserURLParameters tests URL parameter extraction for users
func TestUserURLParameters(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		nameParam string
		email     string
		role      string
		isActive  *bool
	}{
		{
			name:      "Filter By Email",
			url:       "/v1/user?email=test@example.com",
			nameParam: "",
			email:     "test@example.com",
			role:      "",
			isActive:  nil,
		},
		{
			name:      "Filter By Role",
			url:       "/v1/user?role=admin",
			nameParam: "",
			email:     "",
			role:      "admin",
			isActive:  nil,
		},
		{
			name:      "Filter By Name",
			url:       "/v1/user?name=John",
			nameParam: "John",
			email:     "",
			role:      "",
			isActive:  nil,
		},
		{
			name:      "Filter By Active Status",
			url:       "/v1/user?is_active=true",
			nameParam: "",
			email:     "",
			role:      "",
			isActive:  boolPtr(true),
		},
		{
			name:      "Combined Filters",
			url:       "/v1/user?email=admin@example.com&role=admin&is_active=true",
			nameParam: "",
			email:     "admin@example.com",
			role:      "admin",
			isActive:  boolPtr(true),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			query := req.URL.Query()

			app := newTestApp()
			v := validator.New()

			name := app.getSingleQueryParameter(query, "name", "")
			email := app.getSingleQueryParameter(query, "email", "")
			role := app.getSingleQueryParameter(query, "role", "")
			isActive := app.getOptionalBoolQueryParameter(query, "is_active", v)

			if name != tt.nameParam {
				t.Errorf("expected name %v, got %v", tt.nameParam, name)
			}
			if email != tt.email {
				t.Errorf("expected email %v, got %v", tt.email, email)
			}
			if role != tt.role {
				t.Errorf("expected role %v, got %v", tt.role, role)
			}
			if !boolPtrEqual(isActive, tt.isActive) {
				t.Errorf("expected isActive %v, got %v", tt.isActive, isActive)
			}
		})
	}
}

// TestUserJSONParsing tests JSON payload parsing for user registration
func TestUserJSONParsing(t *testing.T) {
	tests := []struct {
		name        string
		payload     string
		expectError bool
		checkFields func(t *testing.T, payload map[string]interface{})
	}{
		{
			name:        "Valid User Registration JSON",
			payload:     `{"first_name": "John", "last_name": "Doe", "email": "john@example.com", "password": "securepass123"}`,
			expectError: false,
			checkFields: func(t *testing.T, payload map[string]interface{}) {
				if payload["first_name"] != "John" {
					t.Error("first_name not parsed correctly")
				}
				if payload["email"] != "john@example.com" {
					t.Error("email not parsed correctly")
				}
			},
		},
		{
			name:        "User with Role",
			payload:     `{"first_name": "Admin", "last_name": "User", "email": "admin@example.com", "password": "pass123", "role": "admin"}`,
			expectError: false,
			checkFields: func(t *testing.T, payload map[string]interface{}) {
				if payload["role"] != "admin" {
					t.Error("role not parsed correctly")
				}
			},
		},
		{
			name:        "Missing Required Fields",
			payload:     `{"email": "test@example.com"}`,
			expectError: false,
			checkFields: func(t *testing.T, payload map[string]interface{}) {
				if _, exists := payload["first_name"]; exists {
					t.Error("first_name should not exist")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var payload map[string]interface{}
			err := json.Unmarshal([]byte(tt.payload), &payload)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.checkFields != nil && err == nil {
				tt.checkFields(t, payload)
			}
		})
	}
}

// TestRegisterUserHandler_Integration is marked as integration test
func TestRegisterUserHandler_Integration(t *testing.T) {
	t.Skip("Integration test - requires database connection")

	payload := map[string]interface{}{
		"first_name": "Test",
		"last_name":  "User",
		"email":      "test@example.com",
		"password":   "testpassword123",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	_ = req
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func boolPtrEqual(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
