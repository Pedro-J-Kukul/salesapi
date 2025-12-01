// File: cmd/api/chatbot.go
package main

import (
	"fmt"
	"net/http"

	"github.com/Pedro-J-Kukul/salesapi/internal/validator"
)

// ChatbotHandler handles chatbot requests
func (app *app) chatbotHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	if user == nil {
		app.notPermittedResponse(w, r)
		return
	}

	var input struct {
		Message string `json:"message"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(input.Message != "", "message", "must be provided")
	v.Check(len(input.Message) <= 500, "message", "must not exceed 500 characters")

	if !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fmt.Printf("📤 Chatbot request: '%s' from %s (%s)\n", input.Message, user.Email, user.Role)

	chatbot := app.models.ChatbotModel
	response, err := chatbot.ProcessMessage(input.Message, user) // Pass full user object
	if err != nil {
		app.logger.Error(err.Error())
		app.serverErrorResponse(w, r, err)
		return
	}

	fmt.Printf("📥 Chatbot response: %s...  (%s)\n", response.Response[:min(50, len(response.Response))], response.Type)

	err = app.writeJSON(w, http.StatusOK, envelope{"chatbot": response}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
