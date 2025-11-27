// Filename: cmd/api/errors.go

package main

import (
	"fmt"
	"net/http"
)

/************************************************************************************************************/
// General helper functions for error handling
/************************************************************************************************************/

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
	if err != nil {
		app.logError(r, err) // log the error
		w.WriteHeader(500)   // send a 500 Internal Server Error status code
	}
}

/************************************************************************************************************/
// Individual error response functions
/************************************************************************************************************/
// error response for total server failure with a 500 status code
func (app *app) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)                                                             // log the error
	message := "the server encountered a problem and could not process your request" // client message
	app.errorResponseJSON(w, r, http.StatusInternalServerError, message)             // send the error response
}

// send an error response if our client messes up with a 404
func (app *app) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponseJSON(w, r, http.StatusNotFound, message)
}

// send an error response if our client messes up with a 405
func (app *app) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
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

// Return a 401 status code
func (a *app) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	a.errorResponseJSON(w, r, http.StatusUnauthorized, message)
}

// Return an authentication required status code 401
func (a *app) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	message := "invalid or missing authentication token"
	a.errorResponseJSON(w, r, http.StatusUnauthorized, message)
}

// Return an authentication required status code 401
func (a *app) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	a.errorResponseJSON(w, r, http.StatusUnauthorized, message)
}

// Return a 403 status code
func (a *app) inactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account must be activated to access this resource"
	a.errorResponseJSON(w, r, http.StatusForbidden, message)
}

// Return a 403 status code
func (a *app) notPermittedResponse(w http.ResponseWriter, r *http.Request) {
	message := "you do not have the necessary permissions to access this resource"
	a.errorResponseJSON(w, r, http.StatusForbidden, message)
}

// Return a 409 status code
func (a *app) conflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "the request could not be completed due to a conflict with the current state of the resource"
	a.errorResponseJSON(w, r, http.StatusConflict, message)
}
