// File: cmd/api/exports.go
// Description: export api handlers

package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Pedro-J-Kukul/salesapi/internal/data"
	"github.com/Pedro-J-Kukul/salesapi/internal/sheets"
	"github.com/Pedro-J-Kukul/salesapi/internal/validator"
)

// exportSalesHandler handles the export of sales records to Google Sheets.
func (app *app) exportSalesHandler(w http.ResponseWriter, r *http.Request) {
	// Check if sheets client is configured
	if app.sheetsService == nil {
		app.serverErrorResponse(w, r, errors.New("Google Sheets is not configured"))
		return
	}

	// Read request payload
	var exportPayload struct {
		ExportType string `json:"export_type"`
		StartDate  string `json:"start_date"`
		EndDate    string `json:"end_date"`
		SheetName  string `json:"sheet_name"`
	}

	err := app.readJSON(w, r, &exportPayload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Validate input
	v := validator.New()
	v.Check(exportPayload.ExportType != "", "export_type", "must be provided")
	v.Check(exportPayload.ExportType == "daily" || exportPayload.ExportType == "monthly" || exportPayload.ExportType == "custom" || exportPayload.ExportType == "all", "export_type", "must be daily, monthly, custom, or all")

	if !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Parse dates if provided
	var startDate, endDate *time.Time
	if exportPayload.StartDate != "" {
		parsedStart, err := time.Parse("2006-01-02", exportPayload.StartDate)
		if err != nil {
			v.AddError("start_date", "must be a valid date in YYYY-MM-DD format")
			app.failedValidationResponse(w, r, v.Errors)
			return
		}
		startDate = &parsedStart
	}

	if exportPayload.EndDate != "" {
		parsedEnd, err := time.Parse("2006-01-02", exportPayload.EndDate)
		if err != nil {
			v.AddError("end_date", "must be a valid date in YYYY-MM-DD format")
			app.failedValidationResponse(w, r, v.Errors)
			return
		}
		endDate = &parsedEnd
	}

	// Get authenticated user
	user := app.contextGetUser(r)

	// Create export history record
	exportHistory := &data.ExportHistory{
		UserID:        user.ID,
		ExportType:    exportPayload.ExportType,
		SpreadsheetID: app.config.sheets.spreadsheetID,
		SheetName:     exportPayload.SheetName,
		StartDate:     startDate,
		EndDate:       endDate,
		Status:        "pending",
	}

	// Generate sheet name if not provided
	if exportHistory.SheetName == "" {
		exportHistory.SheetName = sheets.GenerateSheetName(exportPayload.ExportType, startDate, endDate)
	}

	// Insert export history
	err = app.models.ExportHistory.Insert(exportHistory)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Get sales records for export
	records, err := app.models.Sales.GetSalesForExport(startDate, endDate)
	if err != nil {
		exportHistory.Status = "failed"
		exportHistory.ErrorMessage = fmt.Sprintf("Failed to fetch sales records: %v", err)
		app.models.ExportHistory.Update(exportHistory)
		app.serverErrorResponse(w, r, err)
		return
	}

	// Export to Google Sheets
	exportedBy := fmt.Sprintf("%s %s (%s)", user.FirstName, user.LastName, user.Email)
	rowCount, err := app.sheetsService.ExportSales(exportHistory.SheetName, records, exportedBy)
	if err != nil {
		exportHistory.Status = "failed"
		exportHistory.ErrorMessage = fmt.Sprintf("Failed to export to Google Sheets: %v", err)
		app.models.ExportHistory.Update(exportHistory)
		app.serverErrorResponse(w, r, err)
		return
	}

	// Update export history with success
	exportHistory.Status = "completed"
	exportHistory.RowCount = int64(rowCount)
	err = app.models.ExportHistory.Update(exportHistory)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Return success response
	err = app.writeJSON(w, http.StatusOK, envelope{
		"export": exportHistory,
		"message": fmt.Sprintf("Successfully exported %d sales records to sheet '%s'", rowCount, exportHistory.SheetName),
	}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// listExportHistoryHandler handles listing export history with optional filtering and pagination.
func (app *app) listExportHistoryHandler(w http.ResponseWriter, r *http.Request) {
	// Read Query Parameters
	query := r.URL.Query()
	v := validator.New()

	exportSafeList := []string{"id", "user_id", "export_type", "created_at"}

	filter := app.readFilters(query, "created_at", 20, exportSafeList, v)
	filters := data.ExportFilter{
		Filter:     filter,
		UserID:     app.getSingleIntQueryParameter(query, "user_id", 0, v),
		ExportType: app.getSingleQueryParameter(query, "export_type", ""),
		Status:     app.getSingleQueryParameter(query, "status", ""),
		MinDate:    app.getSingleDateQueryParameter(query, "min_date", "", v),
		MaxDate:    app.getSingleDateQueryParameter(query, "max_date", "", v),
	}

	if !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	exports, metadata, err := app.models.ExportHistory.GetAll(filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"exports": exports, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// getSheetsInfoHandler returns information about the Google Sheets configuration.
func (app *app) getSheetsInfoHandler(w http.ResponseWriter, r *http.Request) {
	// Check if sheets client is configured
	if app.sheetsService == nil {
		app.serverErrorResponse(w, r, errors.New("Google Sheets is not configured"))
		return
	}

	info, err := app.sheetsService.GetSpreadsheetInfo()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"sheets_info": info}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
