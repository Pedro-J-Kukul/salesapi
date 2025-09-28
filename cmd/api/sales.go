package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Pedro-J-Kukul/salesapi/internal/data"
	"github.com/Pedro-J-Kukul/salesapi/internal/validator"
)

// createSale handles POST /v1/sales
func (a *app) createSale(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Cashier  string  `json:"cashier"`
		CashPaid float32 `json:"cash_paid"`
		Items    []struct {
			MenuID   int64 `json:"menu_id"`
			Quantity int64 `json:"quantity"`
		} `json:"items"`
	}

	err := a.readJson(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	// Convert input to CreateSaleRequest
	req := &data.CreateSaleRequest{
		Cashier:  input.Cashier,
		CashPaid: input.CashPaid,
		Items:    make([]data.CreateSaleItem, len(input.Items)),
	}

	for i, item := range input.Items {
		req.Items[i] = data.CreateSaleItem{
			MenuID:   item.MenuID,
			Quantity: item.Quantity,
		}
	}

	// Validate the request
	v := validator.New()
	data.ValidateCreateSaleRequest(v, req)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert the sale
	sale, err := a.models.Sales.Insert(req)
	if err != nil {
		// Check for specific business logic errors
		switch {
		case err.Error() == "insufficient cash paid":
			v.AddErrors("cash_paid", "insufficient cash paid for total amount")
			a.failedValidationResponse(w, r, v.Errors)
		case strings.Contains(err.Error(), "menu item with ID") && strings.Contains(err.Error(), "not found"):
			v.AddErrors("items", "one or more menu items not found")
			a.failedValidationResponse(w, r, v.Errors)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// Set location header
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/sales/%d", sale.ID))

	// Return the created sale
	data := envelope{"sale": sale}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

// deleteSale handles DELETE /v1/sales/{id}
func (a *app) deleteSale(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.models.Sales.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// If we reach this point, the deletion was successful
	data := envelope{"message": "sale successfully deleted"}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}
