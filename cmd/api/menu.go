package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Pedro-J-Kukul/salesapi/internal/data"
	"github.com/Pedro-J-Kukul/salesapi/internal/validator"
)

func (a *app) getMenuItem(w http.ResponseWriter, r *http.Request) {

	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	menuItem, err := a.models.Menu.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	data := envelope{"menu_item": menuItem}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *app) getAllMenuItems(w http.ResponseWriter, r *http.Request) {

	var QueryParData struct {
		Name           string
		Price          float32
		LastModifiedBy string
		data.Filters
	}
	qs := r.URL.Query()

	QueryParData.Name = a.getSingleQueryParam(qs, "name", "")
	QueryParData.LastModifiedBy = a.getSingleQueryParam(qs, "last_modified_by", "")

	v := validator.New()
	QueryParData.Price = a.getSingleFloatParam(qs, "price", 0.0, v)
	QueryParData.Page = a.getSingleIntegegerParam(qs, "page", 1, v)
	QueryParData.PageSize = a.getSingleIntegegerParam(qs, "page_size", 20, v)
	QueryParData.Sort = a.getSingleQueryParam(qs, "sort", "id")
	QueryParData.SortSafelist = []string{"id", "name", "price", "-id", "-name", "-price"}

	data.ValidateFilters(v, QueryParData.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	menuItems, metadata, err := a.models.Menu.GetAll(QueryParData.Name, QueryParData.Price, QueryParData.LastModifiedBy, QueryParData.Filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	data := envelope{"menu_items": menuItems, "metadata": metadata}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *app) createMenuItem(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name           string  `json:"name"`
		Price          float32 `json:"price"`
		LastModifiedBy string  `json:"last_modified_by"`
	}

	err := a.readJson(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	menuItem := &data.Menu{
		Name:           input.Name,
		Price:          input.Price,
		LastModifiedBy: input.LastModifiedBy,
	}

	v := validator.New()
	if data.ValidateMenu(v, menuItem); !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = a.models.Menu.Insert(menuItem)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/menu/%d", menuItem.ID))

	data := envelope{"menu_item": menuItem}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *app) updateMenuItem(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	menuItem, err := a.models.Menu.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Name           *string  `json:"name"`
		Price          *float32 `json:"price"`
		LastModifiedBy *string  `json:"last_modified_by"`
	}

	err = a.readJson(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if input.Name != nil {
		menuItem.Name = *input.Name
	}
	if input.Price != nil {
		menuItem.Price = *input.Price
	}
	if input.LastModifiedBy != nil {
		menuItem.LastModifiedBy = *input.LastModifiedBy
	}

	v := validator.New()
	if data.ValidateMenu(v, menuItem); !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = a.models.Menu.Update(menuItem)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			a.editConflictResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	data := envelope{"menu_item": menuItem}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *app) deleteMenuItem(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.models.Menu.Delete(id)
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
	data := envelope{"message": "menu item successfully deleted"}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}
