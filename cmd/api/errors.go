package main

import (
	"fmt"
	"net/http"
)

// logs the error message along with the request method and URL
func (app *app) logError(r *http.Request, err error) {
	method := r.Method                                          // get the HTTP method
	uri := r.URL.RequestURI()                                   // get the request URI
	app.logger.Error(err.Error(), "method", method, "uri", uri) // log the error with method and URI
}

// Sends an error response in JSON format
func (app *app) errorResponseJSON(w http.ResponseWriter, r *http.Request, status int, message any) {
	errorData := envelope{"error": message}         // wrap the message in an envelope
	err := app.writeJSON(w, status, errorData, nil) // write the JSON response
	//  log the error if we encounter one while trying to write the response
	if err != nil {
		app.logError(r, err) // log the error
		w.WriteHeader(500)   // send a 500 Internal Server Error status code
	}
}

// error response for total server failure with a 500 status code
func (app *app) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)                                                             // log the error
	message := "the server encountered a problem and could not process your request" // client message
	app.errorResponseJSON(w, r, http.StatusInternalServerError, message)             // send the error response
}

// send an error response if our client messes up with a 404
func (app *app) notFoundResponse(w http.ResponseWriter, r *http.Request) {

	// we only log server errors, not client errors
	// prepare a response to send to the client
	message := "the requested resource could not be found"
	app.errorResponseJSON(w, r, http.StatusNotFound, message)
}

// send an error response if our client messes up with a 405
func (app *app) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {

	// we only log server errors, not client errors
	// prepare a formatted response to send to the client
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)

	app.errorResponseJSON(w, r, http.StatusMethodNotAllowed, message)
}

// send an error response if our client messes up with a 400 (bad request)
func (a *app) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {

	a.errorResponseJSON(w, r, http.StatusBadRequest, err.Error())
}

// error response for failed validation checks with a 422 status code
func (a *app) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, errors)
}

// For rate limit exceeded errors with a 429 status code
func (a *app) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	a.errorResponseJSON(w, r, http.StatusTooManyRequests, message)
}

// for edit conflict status 409
func (a *app) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	a.errorResponseJSON(w, r, http.StatusConflict, message)
}
