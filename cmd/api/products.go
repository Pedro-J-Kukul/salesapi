// File: cmd/api/products.go
// Description: products api handlers

package main

import (
	"fmt"
	"net/http"

	"github.com/Pedro-J-Kukul/salesapi/internal/data"
	"github.com/Pedro-J-Kukul/salesapi/internal/validator"
)

// createProductHandler handles the creation of a new product.
func (app *app) createProductHandler(w http.ResponseWriter, r *http.Request) {
	// Create Payload Struct
	var ProductCreatePayload struct {
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}

	err := app.readJSON(w, r, &ProductCreatePayload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	product := &data.Product{
		Name:  ProductCreatePayload.Name,
		Price: ProductCreatePayload.Price,
	}

	// Validate Product
	v := validator.New()

	if data.ValidateProduct(v, product); !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Products.Insert(product)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/products/%d", product.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"product": product}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// listProductsHandler handles listing products with optional filtering and pagination.
func (app *app) listProductsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	v := validator.New()

	ProductSortSafelist := []string{"id", "name", "price", "-id", "-name", "-price"}

	// Read Query Parameters
	filters := app.readFilters(query, "id", 20, ProductSortSafelist, v)
	// Create ProductFilter struct
	productFilter := data.ProductFilter{
		Filter:   filters,
		MinPrice: app.getSingleFloatQueryParameter(query, "min_price", 0, v),
		MaxPrice: app.getSingleFloatQueryParameter(query, "max_price", 0, v),
		Name:     app.getSingleQueryParameter(query, "name", ""),
	}

	// Validate ProductFilter
	if !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Get Products from database
	products, metadata, err := app.models.Products.GetAll(productFilter)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"products": products, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// deleteProductsHandler handles deleting a product by ID.
func (app *app) deleteProductHandler(w http.ResponseWriter, r *http.Request) {
	// Read ID parameter from URL
	id, err := app.readIDParameter(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Delete product from database
	err = app.models.Products.Delete(id)
	if err != nil {
		switch {
		case err == data.ErrRecordNotFound:
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Return a 204 No Content response
	err = app.writeJSON(w, http.StatusNoContent, nil, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// updateProductHandler handles updating an existing product by ID.
func (app *app) updateProductHandler(w http.ResponseWriter, r *http.Request) {
	// Read ID parameter from URL
	id, err := app.readIDParameter(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch existing product from database
	product, err := app.models.Products.Get(id)
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
	var ProductUpdatePayload struct {
		Name  *string  `json:"name"`
		Price *float64 `json:"price"`
	}

	err = app.readJSON(w, r, &ProductUpdatePayload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Update product fields if provided
	if ProductUpdatePayload.Name != nil {
		product.Name = *ProductUpdatePayload.Name
	}
	if ProductUpdatePayload.Price != nil {
		product.Price = *ProductUpdatePayload.Price
	}

	// Validate updated product
	v := validator.New()
	if data.ValidateProduct(v, product); !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Update product in database
	err = app.models.Products.Update(product)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Return the updated product
	err = app.writeJSON(w, http.StatusOK, envelope{"product": product}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// getProductHandler handles retrieving a product by ID.
func (app *app) getProductHandler(w http.ResponseWriter, r *http.Request) {
	// Read ID parameter from URL
	id, err := app.readIDParameter(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch product from database
	product, err := app.models.Products.Get(id)
	if err != nil {
		switch {
		case err == data.ErrRecordNotFound:
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Return the product
	err = app.writeJSON(w, http.StatusOK, envelope{"product": product}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
