// File: cmd/api/handlers_test.go
// Description: Integration tests for API handlers

package main

import (
	"fmt"
	"net/http"
	"testing"
)

/************************************************************************************************************/
// User Handler Tests
/************************************************************************************************************/

func TestRegisterUserHandler(t *testing.T) {
	app, testUtils := newTestApp(t)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name: "Valid user registration",
			payload: map[string]interface{}{
				"first_name": "John",
				"last_name":  "Doe",
				"email":      "john.doe@example.com",
				"password":   "password123",
				"role":       "staff",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Duplicate email",
			payload: map[string]interface{}{
				"first_name": "Jane",
				"last_name":  "Doe",
				"email":      "john.doe@example.com",
				"password":   "password123",
				"role":       "staff",
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "Invalid email",
			payload: map[string]interface{}{
				"first_name": "Invalid",
				"last_name":  "User",
				"email":      "invalid-email",
				"password":   "password123",
				"role":       "staff",
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "Short password",
			payload: map[string]interface{}{
				"first_name": "Short",
				"last_name":  "Pass",
				"email":      "short@example.com",
				"password":   "123",
				"role":       "staff",
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, app, "POST", "/v1/users", tt.payload, nil)
			checkResponseCode(t, tt.expectedStatus, rr.Code)
		})
	}

	// Clean up after all subtests
	testUtils.CleanDatabase()
}

func TestActivateUserHandler(t *testing.T) {
	app, testUtils := newTestApp(t)

	// First, register a user
	userPayload := map[string]interface{}{
		"first_name": "Test",
		"last_name":  "User",
		"email":      "test@example.com",
		"password":   "password123",
		"role":       "staff",
	}

	rr := makeRequest(t, app, "POST", "/v1/users", userPayload, nil)
	checkResponseCode(t, http.StatusCreated, rr.Code)

	// Note: In a real test, you would need to get the activation token from the database
	// or mock the email sending to capture the token
	// For now, we'll just test invalid token

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Invalid token",
			payload: map[string]interface{}{
				"token": "invalidtoken123",
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, app, "PUT", "/v1/users/activate", tt.payload, nil)
			checkResponseCode(t, tt.expectedStatus, rr.Code)
		})
	}

	testUtils.CleanDatabase()
}

func TestListUsersHandler(t *testing.T) {
	app, testUtils := newTestApp(t)

	// Setup admin user
	_, adminToken := setupAdminUser(t, app, testUtils)

	// Create some test users
	createTestUser(t, testUtils, "user1@example.com", "User", "One", "staff", true)
	createTestUser(t, testUtils, "user2@example.com", "User", "Two", "manager", true)

	tests := []struct {
		name           string
		url            string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name:           "List users with authentication",
			url:            "/v1/users",
			headers:        createAuthHeaders(adminToken),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "List users without authentication",
			url:            "/v1/users",
			headers:        nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "List users with pagination",
			url:            "/v1/users?page=1&page_size=10",
			headers:        createAuthHeaders(adminToken),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, app, "GET", tt.url, nil, tt.headers)
			checkResponseCode(t, tt.expectedStatus, rr.Code)
		})
	}

	testUtils.CleanDatabase()
}

func TestShowUserHandler(t *testing.T) {
	app, testUtils := newTestApp(t)

	// Setup admin user
	_, adminToken := setupAdminUser(t, app, testUtils)

	// Create a test user
	userID, _ := createTestUser(t, testUtils, "view@example.com", "View", "User", "staff", true)

	tests := []struct {
		name           string
		url            string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name:           "Get user by ID with authentication",
			url:            fmt.Sprintf("/v1/users/%d", userID),
			headers:        createAuthHeaders(adminToken),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get user without authentication",
			url:            fmt.Sprintf("/v1/users/%d", userID),
			headers:        nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Get non-existent user",
			url:            "/v1/users/99999",
			headers:        createAuthHeaders(adminToken),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, app, "GET", tt.url, nil, tt.headers)
			checkResponseCode(t, tt.expectedStatus, rr.Code)
		})
	}

	testUtils.CleanDatabase()
}

/************************************************************************************************************/
// Product Handler Tests
/************************************************************************************************************/

func TestCreateProductHandler(t *testing.T) {
	app, testUtils := newTestApp(t)

	// Setup admin user
	_, adminToken := setupAdminUser(t, app, testUtils)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		headers        map[string]string
		expectedStatus int
	}{
		{
			name: "Valid product creation",
			payload: map[string]interface{}{
				"name":  "Test Product",
				"price": 99.99,
			},
			headers:        createAuthHeaders(adminToken),
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Product without authentication",
			payload: map[string]interface{}{
				"name":  "Unauthorized Product",
				"price": 49.99,
			},
			headers:        nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Invalid price",
			payload: map[string]interface{}{
				"name":  "Invalid Product",
				"price": -10,
			},
			headers:        createAuthHeaders(adminToken),
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, app, "POST", "/v1/products", tt.payload, tt.headers)
			checkResponseCode(t, tt.expectedStatus, rr.Code)
		})
	}

	testUtils.CleanDatabase()
}

func TestListProductsHandler(t *testing.T) {
	app, testUtils := newTestApp(t)

	// Create some test products
	testUtils.SeedTestProduct("Product 1", 10.00)
	testUtils.SeedTestProduct("Product 2", 20.00)
	testUtils.SeedTestProduct("Product 3", 30.00)

	tests := []struct {
		name           string
		url            string
		expectedStatus int
	}{
		{
			name:           "List all products",
			url:            "/v1/products",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "List products with pagination",
			url:            "/v1/products?page=1&page_size=2",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "List products with price filter",
			url:            "/v1/products?min_price=15&max_price=25",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, app, "GET", tt.url, nil, nil)
			checkResponseCode(t, tt.expectedStatus, rr.Code)
		})
	}

	testUtils.CleanDatabase()
}

func TestGetProductHandler(t *testing.T) {
	app, testUtils := newTestApp(t)

	// Setup admin user
	_, adminToken := setupAdminUser(t, app, testUtils)

	// Create a test product
	productID, _ := testUtils.SeedTestProduct("Get Product", 99.99)

	tests := []struct {
		name           string
		url            string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name:           "Get product by ID",
			url:            fmt.Sprintf("/v1/products/%d", productID),
			headers:        createAuthHeaders(adminToken),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get non-existent product",
			url:            "/v1/products/99999",
			headers:        createAuthHeaders(adminToken),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, app, "GET", tt.url, nil, tt.headers)
			checkResponseCode(t, tt.expectedStatus, rr.Code)
		})
	}

	testUtils.CleanDatabase()
}

func TestUpdateProductHandler(t *testing.T) {
	app, testUtils := newTestApp(t)

	// Setup admin user
	_, adminToken := setupAdminUser(t, app, testUtils)

	// Create a test product
	productID, _ := testUtils.SeedTestProduct("Update Product", 50.00)

	tests := []struct {
		name           string
		url            string
		payload        map[string]interface{}
		headers        map[string]string
		expectedStatus int
	}{
		{
			name: "Valid product update",
			url:  fmt.Sprintf("/v1/products/%d", productID),
			payload: map[string]interface{}{
				"name":  "Updated Product",
				"price": 75.00,
			},
			headers:        createAuthHeaders(adminToken),
			expectedStatus: http.StatusOK,
		},
		{
			name: "Update without authentication",
			url:  fmt.Sprintf("/v1/products/%d", productID),
			payload: map[string]interface{}{
				"price": 100.00,
			},
			headers:        nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, app, "PUT", tt.url, tt.payload, tt.headers)
			checkResponseCode(t, tt.expectedStatus, rr.Code)
		})
	}

	testUtils.CleanDatabase()
}

func TestDeleteProductHandler(t *testing.T) {
	app, testUtils := newTestApp(t)

	// Setup admin user
	_, adminToken := setupAdminUser(t, app, testUtils)

	// Create a test product
	productID, _ := testUtils.SeedTestProduct("Delete Product", 25.00)

	tests := []struct {
		name           string
		url            string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name:           "Delete product with authentication",
			url:            fmt.Sprintf("/v1/products/%d", productID),
			headers:        createAuthHeaders(adminToken),
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Delete non-existent product",
			url:            "/v1/products/99999",
			headers:        createAuthHeaders(adminToken),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, app, "DELETE", tt.url, nil, tt.headers)
			checkResponseCode(t, tt.expectedStatus, rr.Code)
		})
	}

	testUtils.CleanDatabase()
}

/************************************************************************************************************/
// Sales Handler Tests
/************************************************************************************************************/

func TestCreateSaleHandler(t *testing.T) {
	app, testUtils := newTestApp(t)

	// Setup admin user
	adminID, adminToken := setupAdminUser(t, app, testUtils)

	// Create a test product
	productID, _ := testUtils.SeedTestProduct("Sale Product", 50.00)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		headers        map[string]string
		expectedStatus int
	}{
		{
			name: "Valid sale creation",
			payload: map[string]interface{}{
				"user_id":    adminID,
				"product_id": productID,
				"quantity":   5,
			},
			headers:        createAuthHeaders(adminToken),
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Sale without authentication",
			payload: map[string]interface{}{
				"user_id":    adminID,
				"product_id": productID,
				"quantity":   3,
			},
			headers:        nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Invalid quantity",
			payload: map[string]interface{}{
				"user_id":    adminID,
				"product_id": productID,
				"quantity":   -1,
			},
			headers:        createAuthHeaders(adminToken),
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, app, "POST", "/v1/sales", tt.payload, tt.headers)
			checkResponseCode(t, tt.expectedStatus, rr.Code)
		})
	}

	testUtils.CleanDatabase()
}

func TestListSalesHandler(t *testing.T) {
	app, testUtils := newTestApp(t)

	// Setup users and products
	userID, _ := createTestUser(t, testUtils, "sale@example.com", "Sale", "User", "staff", true)
	productID, _ := testUtils.SeedTestProduct("List Sale Product", 30.00)

	// Create test sales
	testUtils.SeedTestSale(userID, productID, 2)
	testUtils.SeedTestSale(userID, productID, 3)

	tests := []struct {
		name           string
		url            string
		expectedStatus int
	}{
		{
			name:           "List all sales",
			url:            "/v1/sales",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "List sales with pagination",
			url:            "/v1/sales?page=1&page_size=10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "List sales by user",
			url:            fmt.Sprintf("/v1/sales?user_id=%d", userID),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, app, "GET", tt.url, nil, nil)
			checkResponseCode(t, tt.expectedStatus, rr.Code)
		})
	}

	testUtils.CleanDatabase()
}

func TestGetSaleHandler(t *testing.T) {
	app, testUtils := newTestApp(t)

	// Setup admin user
	_, adminToken := setupAdminUser(t, app, testUtils)

	// Create test data
	userID, _ := createTestUser(t, testUtils, "getsale@example.com", "Get", "Sale", "staff", true)
	productID, _ := testUtils.SeedTestProduct("Get Sale Product", 40.00)
	saleID, _ := testUtils.SeedTestSale(userID, productID, 4)

	tests := []struct {
		name           string
		url            string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name:           "Get sale by ID",
			url:            fmt.Sprintf("/v1/sales/%d", saleID),
			headers:        createAuthHeaders(adminToken),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get non-existent sale",
			url:            "/v1/sales/99999",
			headers:        createAuthHeaders(adminToken),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, app, "GET", tt.url, nil, tt.headers)
			checkResponseCode(t, tt.expectedStatus, rr.Code)
		})
	}

	testUtils.CleanDatabase()
}

/************************************************************************************************************/
// Token Handler Tests
/************************************************************************************************************/

func TestCreateAuthenticationTokenHandler(t *testing.T) {
	app, testUtils := newTestApp(t)

	// Create an active user
	_, password := createTestUser(t, testUtils, "auth@example.com", "Auth", "User", "staff", true)

	// Create an inactive user
	createTestUser(t, testUtils, "inactive@example.com", "Inactive", "User", "staff", false)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Valid authentication",
			payload: map[string]interface{}{
				"email":    "auth@example.com",
				"password": password,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Invalid password",
			payload: map[string]interface{}{
				"email":    "auth@example.com",
				"password": "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Non-existent user",
			payload: map[string]interface{}{
				"email":    "nonexistent@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Inactive user",
			payload: map[string]interface{}{
				"email":    "inactive@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, app, "POST", "/v1/tokens/authentication", tt.payload, nil)
			checkResponseCode(t, tt.expectedStatus, rr.Code)
		})
	}

	testUtils.CleanDatabase()
}

/************************************************************************************************************/
// Permission Tests
/************************************************************************************************************/

func TestPermissionBasedAccess(t *testing.T) {
	app, testUtils := newTestApp(t)

	// Setup admin and staff users
	_, adminToken := setupAdminUser(t, app, testUtils)
	_, staffToken := setupStaffUser(t, app, testUtils)

	// Create a test product
	productID, _ := testUtils.SeedTestProduct("Permission Test Product", 60.00)

	tests := []struct {
		name           string
		method         string
		url            string
		payload        map[string]interface{}
		headers        map[string]string
		expectedStatus int
	}{
		{
			name:           "Admin can create product",
			method:         "POST",
			url:            "/v1/products",
			payload:        map[string]interface{}{"name": "Admin Product", "price": 100.00},
			headers:        createAuthHeaders(adminToken),
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Staff cannot create product",
			method:         "POST",
			url:            "/v1/products",
			payload:        map[string]interface{}{"name": "Staff Product", "price": 50.00},
			headers:        createAuthHeaders(staffToken),
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Admin can delete product",
			method:         "DELETE",
			url:            fmt.Sprintf("/v1/products/%d", productID),
			payload:        nil,
			headers:        createAuthHeaders(adminToken),
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, app, tt.method, tt.url, tt.payload, tt.headers)
			checkResponseCode(t, tt.expectedStatus, rr.Code)
		})
	}

	testUtils.CleanDatabase()
}
