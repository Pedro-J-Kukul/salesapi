// File: cmd/api/sales_test.go
// Description: test suite for sales handlers - validation focused

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

// TestSaleValidation tests sale validation logic
func TestSaleValidation(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		productID     int64
		quantity      int64
		expectedValid bool
		errorField    string
	}{
		{
			name:          "Valid Sale",
			userID:        1,
			productID:     1,
			quantity:      5,
			expectedValid: true,
		},
		{
			name:          "Zero Quantity",
			userID:        1,
			productID:     1,
			quantity:      0,
			expectedValid: false,
			errorField:    "quantity",
		},
		{
			name:          "Negative Quantity",
			userID:        1,
			productID:     1,
			quantity:      -5,
			expectedValid: false,
			errorField:    "quantity",
		},
		{
			name:          "Zero User ID",
			userID:        0,
			productID:     1,
			quantity:      5,
			expectedValid: false,
			errorField:    "user_id",
		},
		{
			name:          "Zero Product ID",
			userID:        1,
			productID:     0,
			quantity:      5,
			expectedValid: false,
			errorField:    "product_id",
		},
		{
			name:          "Negative User ID",
			userID:        -1,
			productID:     1,
			quantity:      5,
			expectedValid: false,
			errorField:    "user_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sale := &data.Sale{
				UserID:    tt.userID,
				ProductID: tt.productID,
				Quantity:  tt.quantity,
			}

			v := validator.New()
			data.ValidateSale(v, sale)

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

// TestSaleURLParameters tests URL parameter extraction for sales
func TestSaleURLParameters(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		userID    int64
		productID int64
		minQty    int64
		maxQty    int64
	}{
		{
			name:      "User ID Filter",
			url:       "/v1/sales?user_id=1",
			userID:    1,
			productID: 0,
			minQty:    0,
			maxQty:    0,
		},
		{
			name:      "Product ID Filter",
			url:       "/v1/sales?product_id=5",
			userID:    0,
			productID: 5,
			minQty:    0,
			maxQty:    0,
		},
		{
			name:      "Quantity Range Filter",
			url:       "/v1/sales?min_qty=5&max_qty=20",
			userID:    0,
			productID: 0,
			minQty:    5,
			maxQty:    20,
		},
		{
			name:      "Combined Filters",
			url:       "/v1/sales?user_id=2&product_id=3&min_qty=1&max_qty=100",
			userID:    2,
			productID: 3,
			minQty:    1,
			maxQty:    100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			query := req.URL.Query()

			app := newTestApp()
			v := validator.New()

			userID := app.getSingleIntQueryParameter(query, "user_id", 0, v)
			productID := app.getSingleIntQueryParameter(query, "product_id", 0, v)
			minQty := app.getSingleIntQueryParameter(query, "min_qty", 0, v)
			maxQty := app.getSingleIntQueryParameter(query, "max_qty", 0, v)

			if userID != tt.userID {
				t.Errorf("expected userID %v, got %v", tt.userID, userID)
			}
			if productID != tt.productID {
				t.Errorf("expected productID %v, got %v", tt.productID, productID)
			}
			if minQty != tt.minQty {
				t.Errorf("expected minQty %v, got %v", tt.minQty, minQty)
			}
			if maxQty != tt.maxQty {
				t.Errorf("expected maxQty %v, got %v", tt.maxQty, maxQty)
			}
		})
	}
}

// TestSaleJSONParsing tests JSON payload parsing for sales
func TestSaleJSONParsing(t *testing.T) {
	tests := []struct {
		name        string
		payload     string
		expectError bool
		checkFields func(t *testing.T, payload map[string]interface{})
	}{
		{
			name:        "Valid Sale JSON",
			payload:     `{"user_id": 1, "product_id": 2, "quantity": 10}`,
			expectError: false,
			checkFields: func(t *testing.T, payload map[string]interface{}) {
				if payload["user_id"].(float64) != 1 {
					t.Error("user_id not parsed correctly")
				}
				if payload["quantity"].(float64) != 10 {
					t.Error("quantity not parsed correctly")
				}
			},
		},
		{
			name:        "Invalid JSON",
			payload:     `{"user_id": "invalid", "product_id": 2}`,
			expectError: false, // JSON will parse, but validation should fail
			checkFields: nil,
		},
		{
			name:        "Missing Fields",
			payload:     `{"user_id": 1}`,
			expectError: false,
			checkFields: func(t *testing.T, payload map[string]interface{}) {
				if _, exists := payload["product_id"]; exists {
					t.Error("product_id should not exist")
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

// TestCreateSaleHandler_Integration is marked as integration test
func TestCreateSaleHandler_Integration(t *testing.T) {
	t.Skip("Integration test - requires database connection")

	// This would test the actual endpoint with a real database
	payload := map[string]interface{}{
		"user_id":    1,
		"product_id": 1,
		"quantity":   5,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/v1/sales", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	_ = req
}
