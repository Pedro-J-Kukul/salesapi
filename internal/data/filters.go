// File: internal/data/filters.go

package data

import (
	"strings"

	"github.com/Pedro-J-Kukul/salesapi/internal/validator"
)

// ----------------------------------------------------------------------
//
//	Definitions
//
// ----------------------------------------------------------------------

// Filter represents common filtering criteria for querying records.
type Filter struct {
	Page         int64    `json:"page"`
	PageSize     int64    `json:"page_size"`
	SortBy       string   `json:"sort_by"`
	SortSafeList []string `json:"-"`
}

// MetaData contains pagination metadata.
type MetaData struct {
	CurrentPage  int64 `json:"current_page,omitempty"`  // Current page number
	PageSize     int64 `json:"page_size,omitempty"`     // Number of records per page
	FirstPage    int64 `json:"first_page,omitempty"`    // First page number
	LastPage     int64 `json:"last_page,omitempty"`     // Last page number
	TotalRecords int64 `json:"total_records,omitempty"` // Total number of records
}

// ----------------------------------------------------------------------
//
//	Methods
//
// ----------------------------------------------------------------------

// ValidateFilters checks the validity of the filter parameters.
func ValidateFilters(v *validator.Validator, f Filter) {
	v.Check(f.Page > 0, "page", "must be greater than zero")                        // Page must be greater than 0
	v.Check(f.Page <= 500, "page", "must be a maximum of 500")                      // Page must be at most 500
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")               // PageSize must be greater than 0
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")             // PageSize must be at most 100
	v.Check(v.Permitted(f.SortBy, f.SortSafeList...), "sort", "invalid sort value") // Sort must be in the safelist
}

// Limit calculates the SQL LIMIT value based on the page size.
func (f Filter) Limit() int64 {
	return f.PageSize
}

// Offset calculates the SQL OFFSET value based on the current page and page size.
func (f Filter) Offset() int64 {
	return (f.Page - 1) * f.PageSize
}

// SortColumn returns the column name to sort by, removing any leading '-' for descending order.
func (f Filter) SortColumn() string {
	for _, safeValue := range f.SortSafeList {
		if f.SortBy == safeValue {
			return strings.TrimPrefix(f.SortBy, "-") // Remove leading '-' if present
		}
	}
	panic("unsafe sort parameter: " + f.SortBy) // Panic if the sort parameter is not in the safelist
}

// SortDirection returns the sort direction ("ASC" or "DESC") based on the SortBy field.
func (f Filter) SortDirection() string {
	if strings.HasPrefix(f.SortBy, "-") {
		return "DESC"
	}
	return "ASC"
}

// CalculateMetaData computes pagination metadata based on total records, current page, and page size.
func CalculateMetaData(totalRecords, page, pageSize int64) MetaData {
	if totalRecords == 0 {
		return MetaData{}
	}

	lastPage := (totalRecords + pageSize - 1) / pageSize // Calculate last page number

	return MetaData{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     lastPage,
		TotalRecords: totalRecords,
	}
}
