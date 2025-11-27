// File: cmd/api/users.go
package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Pedro-J-Kukul/salesapi/internal/data"
	"github.com/Pedro-J-Kukul/salesapi/internal/validator"
)

// registerUserHandler handles user registration.
func (app *app) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	// RegisterUserPayload struct to hold the incoming JSON payload
	var RegisterUserPayload struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Role      string `json:"role"`
		Email     string `json:"email"`
		Password  string `json:"password"`
	}

	if err := app.readJSON(w, r, &RegisterUserPayload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Create a new User struct
	user := &data.User{
		FirstName: RegisterUserPayload.FirstName,
		LastName:  RegisterUserPayload.LastName,
		Role:      RegisterUserPayload.Role,
		Email:     RegisterUserPayload.Email,
	}

	if err := user.Password.Set(RegisterUserPayload.Password); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Validate the user data
	v := validator.New()
	if data.ValidateUser(v, user); !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert the user into the database
	if err := app.models.Users.Insert(user); err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Check Role and assign permissions
	var permissions data.Permissions
	switch user.Role {
	case "admin":
		permissions = data.Permissions{"sale:create", "sale:view", "sale:delete", "sale:update",
			"product:create", "product:view", "product:delete", "product:update",
			"users:create", "users:view", "users:delete", "users:update",
			"self:create", "self:view", "self:delete", "self:update"}
	case "manager":
		permissions = data.Permissions{"sale:create", "sale:view", "product:create", "product:view",
			"users:view", "self:create", "self:view", "self:update"}
	case "staff":
		permissions = data.Permissions{"product:view", "sales:view"}
	default:
		permissions = data.Permissions{}
	}

	// Assign permissions to the user
	err := app.models.Permissions.AssignPermissions(user.ID, permissions)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Clear Existing Activation Tokens
	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Generate a new activation token
	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if app.mailer != nil {
		app.background(func() {
			data := map[string]any{
				"userID":          user.ID,
				"firstName":       user.FirstName,
				"lastName":        user.LastName,
				"email":           user.Email,
				"password":        RegisterUserPayload.Password,
				"activationToken": token.Plaintext,
			}
			if err := app.mailer.Send(user.Email, "user_welcome.tmpl", data); err != nil {
				app.logger.Error("failed to send activation email", "error", err)
			}
		})
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/users/%d", user.ID))
	if err := app.writeJSON(w, http.StatusCreated, envelope{"user": user}, headers); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// activateUserHandler handles user account activation.
func (app *app) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	// ActivateUserPayload struct to hold the incoming JSON payload
	var ActivateUserPayload struct {
		TokenPlaintext string `json:"token"`
	}

	if err := app.readJSON(w, r, &ActivateUserPayload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Validate the token
	v := validator.New()
	data.ValidateTokenPlaintext(v, ActivateUserPayload.TokenPlaintext)
	if !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Retrieve the user associated with the activation token
	user, err := app.models.Users.GetForToken(data.ScopeActivation, ActivateUserPayload.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecords):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Activate the user account
	user.IsActive = true
	err = app.models.Users.Update(user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Delete all activation tokens for the user
	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send a confirmation response
	if err := app.writeJSON(w, http.StatusOK, envelope{"message": "account successfully activated"}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// showCurrentUserHandler handles retrieving the authenticated user's information.
func (app *app) showCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	if err := app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// ShowUserHandler handles retrieving a user by ID.
func (app *app) showUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	user, err := app.models.Users.GetByID(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// listUsersHandler handles listing users with optional filters.
func (app *app) listUsersHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	v := validator.New()

	UsersSortSafelist := []string{"id", "first_name", "last_name", "email", "-id", "-first_name", "-last_name", "-email"}

	// Read Query Parameters
	filters := app.readFilters(query, "id", 20, UsersSortSafelist, v)
	// Validate Filters
	data.ValidateFilters(v, filters)
	if !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Create UserFilter struct
	userFilter := data.UserFilter{
		Filter:   filters,
		Name:     app.getSingleQueryParameter(query, "name", ""),
		Email:    app.getSingleQueryParameter(query, "email", ""),
		Role:     app.getSingleQueryParameter(query, "role", ""),
		IsActive: app.getOptionalBoolQueryParameter(query, "is_active", v),
	}
	// Get Users from database
	users, metadata, err := app.models.Users.GetAll(userFilter)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"users": users, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// deleteUserHandler handles deleting a user by ID.
func (app *app) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	// Read ID parameter from URL
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Delete user from database
	err = app.models.Users.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Send a confirmation response
	if err := app.writeJSON(w, http.StatusOK, envelope{"message": "user successfully deleted"}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// updateUserHandler handles updating a user by ID.
func (app *app) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Read ID parameter from URL
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Retrieve the existing user record
	user, err := app.models.Users.GetByID(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// UpdateUserPayload struct to hold the incoming JSON payload
	var UpdateUserPayload struct {
		FirstName *string `json:"first_name"`
		LastName  *string `json:"last_name"`
		Role      *string `json:"role"`
		Email     *string `json:"email"`
		Password  *string `json:"password"`
		IsActive  *bool   `json:"is_active"`
	}

	if err := app.readJSON(w, r, &UpdateUserPayload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Update fields if provided
	if UpdateUserPayload.FirstName != nil {
		user.FirstName = *UpdateUserPayload.FirstName
	}
	if UpdateUserPayload.LastName != nil {
		user.LastName = *UpdateUserPayload.LastName
	}
	if UpdateUserPayload.Role != nil {
		user.Role = *UpdateUserPayload.Role
	}
	if UpdateUserPayload.Email != nil {
		user.Email = *UpdateUserPayload.Email
	}
	if UpdateUserPayload.Password != nil {
		if err := user.Password.Set(*UpdateUserPayload.Password); err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	if UpdateUserPayload.IsActive != nil {
		user.IsActive = *UpdateUserPayload.IsActive
	}

	// Validate the updated user data
	v := validator.New()
	if data.ValidateUser(v, user); !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Update the user record in the database
	if err := app.models.Users.Update(user); err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	// Send the updated user record in the response
	if err := app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
