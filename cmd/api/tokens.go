// Filename: cmd/api/tokens.go

package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/Pedro-J-Kukul/salesapi/internal/data"
	"github.com/Pedro-J-Kukul/salesapi/internal/validator"
)

// createAuthenticationTokenHandler handles the creation of authentication tokens.
func (app *app) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Define the structure for the expected JSON payload.
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Read and parse the JSON payload from the request body.
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Validate the input data.
	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)
	if !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Authenticate the user using the provided email and password.
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	if !user.IsActive {
		v.AddError("email", "account must be activated to login")
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// Generate a new authentication token for the authenticated user.
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send the token back in the response.
	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token.Plaintext}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// deleteAuthenticationTokenHandler handles the deletion of authentication tokens.
func (app *app) deleteAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// get user id from context
	userID := app.contextGetUser(r).ID

	// delete all authentication tokens for the user
	err := app.models.Tokens.DeleteAllForUser(data.ScopeAuthentication, userID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// send a success response
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "authentication tokens deleted successfully"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
