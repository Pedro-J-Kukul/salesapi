// File: cmd/api/products_test.go
// Description: test suite for product handlers

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

// TestCreateProductHandler tests the product creation endpoint
func TestCreateProductHandler(t *testing.T) {
	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "Valid Product Creation",
			payload: map[string]interface{}{
				"name":  "Test Product",
				"price": 99.99,
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatal(err)
				}
				product := response["product"].(map[string]interface{})
				if product["name"] != "Test Product" {
					t.Errorf("expected name 'Test Product', got %v", product["name"])
				}
			},
		},
		{
			name: "Invalid JSON - Missing Price",
			payload: map[string]interface{}{
				"name": "Test",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			checkResponse:  nil,
		},
		{
			name: "Negative Price",
			payload: map[string]interface{}{
				"name":  "Invalid Product",
				"price": -10.0,
			},
			expectedStatus: http.StatusUnprocessableEntity,
			checkResponse:  nil,
		},
		{
			name: "Empty Name",
			payload: map[string]interface{}{
				"name":  "",
				"price": 50.0,
			},
			expectedStatus: http.StatusUnprocessableEntity,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/v1/products", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Note: This test requires a real database connection
			// Mark as skipped if DB not available
			t.Skip("Requires database connection - integration test")
		})
	}
}

// TestProductValidation tests validation logic specifically
func TestProductValidation(t *testing.T) {
	tests := []struct {
		name          string
		productName   string
		productPrice  float64
		expectedValid bool
	}{
		{
			name:          "Valid Product",
			productName:   "Test Product",
			productPrice:  99.99,
			expectedValid: true,
		},
		{
			name:          "Empty Name",
			productName:   "",
			productPrice:  50.0,
			expectedValid: false,
		},
		{
			name:          "Negative Price",
			productName:   "Product",
			productPrice:  -10.0,
			expectedValid: false,
		},
		{
			name:          "Zero Price",
			productName:   "Product",
			productPrice:  0.0,
			expectedValid: true, // Zero price is allowed by current validation
		},
		{
			name:          "Very Long Name",
			productName:   string(make([]byte, 1000)),
			productPrice:  10.0,
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := &data.Product{
				Name:  tt.productName,
				Price: tt.productPrice,
			}

			v := validator.New()
			data.ValidateProduct(v, product)

			isValid := v.IsValid()
			if isValid != tt.expectedValid {
				t.Errorf("expected valid=%v, got=%v. Errors: %v", tt.expectedValid, isValid, v.Errors)
			}
		})
	}
}

// TestProductEndpoints tests product endpoints using the router
func TestProductEndpoints(t *testing.T) {
	// Skip these tests as they require database
	t.Skip("Integration tests require database connection")

	tests := []struct {
		name           string
		method         string
		url            string
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name:           "Create Product",
			method:         http.MethodPost,
			url:            "/v1/products",
			body:           map[string]interface{}{"name": "Test", "price": 10.0},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "List Products",
			method:         http.MethodGet,
			url:            "/v1/products",
			body:           nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != nil {
				bodyBytes, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(tt.method, tt.url, bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.url, nil)
			}
			w := httptest.NewRecorder()

			// Would need to set up full app with database here
			_ = w
			_ = req
		})
	}
}

// TestProductURLParameters tests URL parameter extraction
func TestProductURLParameters(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		minPrice    float64
		maxPrice    float64
		productName string
	}{
		{
			name:        "Price Range Filter",
			url:         "/v1/products?min_price=10&max_price=50",
			minPrice:    10.0,
			maxPrice:    50.0,
			productName: "",
		},
		{
			name:        "Name Filter",
			url:         "/v1/products?name=Widget",
			minPrice:    0,
			maxPrice:    0,
			productName: "Widget",
		},
		{
			name:        "Combined Filters",
			url:         "/v1/products?name=Widget&min_price=5&max_price=100",
			minPrice:    5.0,
			maxPrice:    100.0,
			productName: "Widget",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			query := req.URL.Query()

			app := newTestApp()
			v := validator.New()

			minPrice := app.getSingleFloatQueryParameter(query, "min_price", 0, v)
			maxPrice := app.getSingleFloatQueryParameter(query, "max_price", 0, v)
			name := app.getSingleQueryParameter(query, "name", "")

			if minPrice != tt.minPrice {
				t.Errorf("expected minPrice %v, got %v", tt.minPrice, minPrice)
			}
			if maxPrice != tt.maxPrice {
				t.Errorf("expected maxPrice %v, got %v", tt.maxPrice, maxPrice)
			}
			if name != tt.productName {
				t.Errorf("expected name %v, got %v", tt.productName, name)
			}
		})
	}
}

// newTestApp creates a minimal app instance for testing
func newTestApp() *app {
	logger := setUpLogger("test")
	return &app{
		logger: logger,
	}
}
