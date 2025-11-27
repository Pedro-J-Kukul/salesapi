// File: cmd/api/sales.go
// Description: sales api handlers

package main

import (
	"fmt"
	"net/http"

	"github.com/Pedro-J-Kukul/salesapi/internal/data"
	"github.com/Pedro-J-Kukul/salesapi/internal/validator"
)

// createSaleHandler handles the creation of a new sale.
func (app *app) createSaleHandler(w http.ResponseWriter, r *http.Request) {
	// Create Payload Struct
	var SaleCreatePayload struct {
		UserID    int64 `json:"user_id"`
		ProductID int64 `json:"product_id"`
		Quantity  int64 `json:"quantity"`
	}

	err := app.readJSON(w, r, &SaleCreatePayload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	sale := &data.Sale{
		UserID:    SaleCreatePayload.UserID,
		ProductID: SaleCreatePayload.ProductID,
		Quantity:  SaleCreatePayload.Quantity,
	}

	// Validate Sale
	v := validator.New()

	if data.ValidateSale(v, sale); !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Sales.Insert(sale)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/sales/%d", sale.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"sale": sale}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// listSalesHandler handles listing sales with optional filtering and pagination.
func (app *app) listSalesHandler(w http.ResponseWriter, r *http.Request) {
	// Read Query Parameters
	query := r.URL.Query()
	v := validator.New()

	SaleSafeList := []string{"id", "user_id", "product_id", "quantity", "sold_at"}

	filter := app.readFilters(query, "id", 20, SaleSafeList, v)
	filters := data.SaleFilter{
		Filter:    filter,
		UserID:    app.getSingleIntQueryParameter(query, "user_id", 0, v),
		ProductID: app.getSingleIntQueryParameter(query, "product_id", 0, v),
		MinQty:    app.getSingleIntQueryParameter(query, "min_qty", 0, v),
		MaxQty:    app.getSingleIntQueryParameter(query, "max_qty", 0, v),
		MinDate:   app.getSingleDateQueryParameter(query, "min_date", "", v),
		MaxDate:   app.getSingleDateQueryParameter(query, "max_date", "", v),
	}

	if !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	sales, metadata, err := app.models.Sales.GetAll(filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"sales": sales, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// deleteSalesHandler handles the deletion of a sale.
func (app *app) deleteSalesHandler(w http.ResponseWriter, r *http.Request) {
	// get the id parameter from the url
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Sales.Delete(id)
	if err != nil {
		switch {
		case err == data.ErrRecordNotFound:
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "sale successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// updateSaleHandler handles updating an existing sale.
func (app *app) updateSaleHandler(w http.ResponseWriter, r *http.Request) {
	// get the id parameter from the url
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	sales, err := app.models.Sales.Get(id)
	if err != nil {
		switch {
		case err == data.ErrRecordNotFound:
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Create Payload Struct
	var SaleUpdatePayload struct {
		UserID    *int64 `json:"user_id"`
		ProductID *int64 `json:"product_id"`
		Quantity  *int64 `json:"quantity"`
	}

	err = app.readJSON(w, r, &SaleUpdatePayload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if SaleUpdatePayload.UserID != nil {
		sales.UserID = *SaleUpdatePayload.UserID
	}
	if SaleUpdatePayload.ProductID != nil {
		sales.ProductID = *SaleUpdatePayload.ProductID
	}
	if SaleUpdatePayload.Quantity != nil {
		sales.Quantity = *SaleUpdatePayload.Quantity
	}

	// Validate Sale
	v := validator.New()

	if data.ValidateSale(v, sales); !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Sales.Update(sales)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"sale": sales}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// getSaleHandler handles retrieving a specific sale by ID.
func (app *app) getSaleHandler(w http.ResponseWriter, r *http.Request) {
	// Read the id parameter from the url
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	sale, err := app.models.Sales.Get(id)
	if err != nil {
		switch {
		case err == data.ErrRecordNotFound:
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"sale": sale}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
