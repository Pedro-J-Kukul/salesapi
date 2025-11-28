// Filename: /cmd/api/routes.go
// Description: connects the routes with an api

package main

import (
	"expvar"
	"net/http"

	// Importing Route Package
	"github.com/julienschmidt/httprouter"
)

func (app *app) routes() http.Handler {

	// create a new router instance
	router := httprouter.New()

	// Handle 404 errors
	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	// handling 405 errors
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// Health Check Route
	// router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	// Metrics Route
	router.Handler(http.MethodGet, "/v1/metrics", expvar.Handler())

	// Authentication and User Routes
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)                              // User Registration
	router.HandlerFunc(http.MethodPut, "/v1/users/activate", app.activateUserHandler)                      // User Activation
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler) // Login

	// Authenticated User Routes
	router.Handler(http.MethodGet, "/v1/users/profile", app.requireAuthenticatedUser(http.HandlerFunc(app.showCurrentUserHandler))) // Get Authenticated User Info
	router.Handler(http.MethodPut, "/v1/users/profile/:id", app.requireAuthenticatedUser(http.HandlerFunc(app.updateUserHandler)))  // Update Authenticated User Info

	// User Routes
	router.Handler(http.MethodGet, "/v1/user", app.requireAuthenticatedUser(app.requirePermissions("users:view")(http.HandlerFunc(app.listUsersHandler))))           // List All Users
	router.Handler(http.MethodGet, "/v1/user/:id", app.requireAuthenticatedUser(app.requirePermissions("users:view")(http.HandlerFunc(app.showUserHandler))))        // Get User by ID
	router.Handler(http.MethodDelete, "/v1/user/:id", app.requireAuthenticatedUser(app.requirePermissions("users:delete")(http.HandlerFunc(app.deleteUserHandler)))) // Delete User by ID
	router.Handler(http.MethodPut, "/v1/user/:id", app.requireAuthenticatedUser(app.requirePermissions("users:update")(http.HandlerFunc(app.updateUserHandler))))    // Update User by ID

	// Product Routes, all but view require authentication, the rest require specific permissions
	router.Handler(http.MethodGet, "/v1/products", app.requireAuthenticatedUser(app.requirePermissions("product:view")(http.HandlerFunc(app.listProductsHandler))))           // List All Products
	router.Handler(http.MethodGet, "/v1/products/:id", app.requireAuthenticatedUser(app.requirePermissions("product:view")(http.HandlerFunc(app.getProductHandler))))         // Get Product by ID
	router.Handler(http.MethodPost, "/v1/products", app.requireAuthenticatedUser(app.requirePermissions("product:create")(http.HandlerFunc(app.createProductHandler))))       // Create New Product
	router.Handler(http.MethodPut, "/v1/products/:id", app.requireAuthenticatedUser(app.requirePermissions("product:update")(http.HandlerFunc(app.updateProductHandler))))    // Update Product by ID
	router.Handler(http.MethodDelete, "/v1/products/:id", app.requireAuthenticatedUser(app.requirePermissions("product:delete")(http.HandlerFunc(app.deleteProductHandler)))) // Delete Product by ID

	// Sales Routes, all but viewall require authentication, the rest require specific permissions
	router.Handler(http.MethodGet, "/v1/sales", app.requirePermissions("sale:view")(http.HandlerFunc(app.listSalesHandler)))                                          // List All Sales
	router.Handler(http.MethodGet, "/v1/sales/:id", app.requireAuthenticatedUser(app.requirePermissions("sale:view")(http.HandlerFunc(app.getSaleHandler))))          // Get Sale by ID
	router.Handler(http.MethodPost, "/v1/sales", app.requireAuthenticatedUser(app.requirePermissions("sale:create")(http.HandlerFunc(app.createSaleHandler))))        // Create New Sale
	router.Handler(http.MethodPut, "/v1/sales/:id", app.requireAuthenticatedUser(app.requirePermissions("sale:update")(http.HandlerFunc(app.updateSaleHandler))))     // Update Sale by ID
	router.Handler(http.MethodDelete, "/v1/sales/:id", app.requireAuthenticatedUser(app.requirePermissions("sale:delete")(http.HandlerFunc(app.deleteSalesHandler)))) // Delete Sale by ID

	return app.recoverPanic(app.enableCORS(app.metrics(app.rateLimit(app.authenticate(router)))))
}
