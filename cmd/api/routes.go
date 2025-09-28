// Filename: /cmd/api/routes.go
// Description: connects the routes with an api

package main

import (
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

	// Menu routes plain simple

	router.HandlerFunc(http.MethodGet, "menus:id", app.getMenuItem)
	router.HandlerFunc(http.MethodGet, "menus", app.getAllMenuItems)
	router.HandlerFunc(http.MethodPost, "menus", app.createMenuItem)
	router.HandlerFunc(http.MethodPatch, "menus:id", app.updateMenuItem)
	router.HandlerFunc(http.MethodDelete, "menus:id", app.deleteMenuItem)

	// Sales
	router.HandlerFunc(http.MethodPost, "/v1/sales", app.createSale)
	router.HandlerFunc(http.MethodDelete, "/v1/sales/:id", app.deleteSale)
	// include panic middleware
	return app.recoverPanic(app.rateLimit(app.enableCORS(router)))
}
