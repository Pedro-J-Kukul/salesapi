// File: internal/data/exports.go
package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Pedro-J-Kukul/salesapi/internal/validator"
)

// ----------------------------------------------------------------------
//
//	Definitions
//
// ----------------------------------------------------------------------

// ExportHistory represents an export record in the system.
type ExportHistory struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"user_id"`
	ExportType    string     `json:"export_type"`
	SpreadsheetID string     `json:"spreadsheet_id"`
	SheetName     string     `json:"sheet_name"`
	RowCount      int64      `json:"row_count"`
	StartDate     *time.Time `json:"start_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	Status        string     `json:"status"`
	ErrorMessage  string     `json:"error_message,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ExportHistoryModel wraps a sql.DB connection pool.
type ExportHistoryModel struct {
	DB *sql.DB
}

// ExportFilter represents filtering criteria for querying export history.
type ExportFilter struct {
	Filter     Filter `json:"filter"`
	UserID     int64  `json:"user_id"`
	ExportType string `json:"export_type"`
	Status     string `json:"status"`
	MinDate    string `json:"min_date"`
	MaxDate    string `json:"max_date"`
}

// SaleExportRecord represents a single sale record formatted for export.
type SaleExportRecord struct {
	TransactionID int64     `json:"transaction_id"`
	Date          string    `json:"date"`
	Time          string    `json:"time"`
	UserEmail     string    `json:"user_email"`
	UserName      string    `json:"user_name"`
	ProductName   string    `json:"product_name"`
	ProductPrice  float64   `json:"product_price"`
	Quantity      int64     `json:"quantity"`
	TotalAmount   float64   `json:"total_amount"`
	SoldAt        time.Time `json:"sold_at"`
}

// ----------------------------------------------------------------------
//
//	Methods
//
// ----------------------------------------------------------------------

// ValidateExportHistory checks the fields of an ExportHistory struct to ensure they meet the required criteria.
func ValidateExportHistory(v *validator.Validator, export *ExportHistory) {
	v.Check(export.UserID > 0, "user_id", "must be a positive integer")
	v.Check(export.ExportType != "", "export_type", "must be provided")
	v.Check(export.SpreadsheetID != "", "spreadsheet_id", "must be provided")
	v.Check(export.SheetName != "", "sheet_name", "must be provided")
	v.Check(export.Status != "", "status", "must be provided")
	v.Check(export.Status == "pending" || export.Status == "completed" || export.Status == "failed", "status", "must be pending, completed, or failed")
}

// Insert adds a new export history record to the database.
func (m *ExportHistoryModel) Insert(export *ExportHistory) error {
	query := `
		INSERT INTO export_history (user_id, export_type, spreadsheet_id, sheet_name, row_count, start_date, end_date, status, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := m.DB.QueryRowContext(
		ctx,
		query,
		export.UserID,
		export.ExportType,
		export.SpreadsheetID,
		export.SheetName,
		export.RowCount,
		export.StartDate,
		export.EndDate,
		export.Status,
		export.ErrorMessage,
	).Scan(&export.ID, &export.CreatedAt); err != nil {
		return err
	}
	return nil
}

// Update modifies an existing export history record in the database.
func (m *ExportHistoryModel) Update(export *ExportHistory) error {
	query := `
		UPDATE export_history
		SET status = $1, error_message = $2, row_count = $3
		WHERE id = $4
		RETURNING created_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := m.DB.QueryRowContext(
		ctx,
		query,
		export.Status,
		export.ErrorMessage,
		export.RowCount,
		export.ID,
	).Scan(&export.CreatedAt); err != nil {
		return err
	}
	return nil
}

// Get retrieves an export history record by its ID.
func (m *ExportHistoryModel) Get(id int64) (*ExportHistory, error) {
	query := `
		SELECT id, user_id, export_type, spreadsheet_id, sheet_name, row_count, start_date, end_date, status, error_message, created_at
		FROM export_history
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	export := &ExportHistory{}

	if err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&export.ID,
		&export.UserID,
		&export.ExportType,
		&export.SpreadsheetID,
		&export.SheetName,
		&export.RowCount,
		&export.StartDate,
		&export.EndDate,
		&export.Status,
		&export.ErrorMessage,
		&export.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return export, nil
}

// GetAll retrieves export history records based on filtering criteria and pagination.
func (m *ExportHistoryModel) GetAll(filter ExportFilter) ([]*ExportHistory, MetaData, error) {
	query := fmt.Sprintf(`
		SELECT id, user_id, export_type, spreadsheet_id, sheet_name, row_count, start_date, end_date, status, error_message, created_at
		FROM export_history
		WHERE (user_id = $1 OR $1 = 0)
		  AND (export_type = $2 OR $2 = '')
		  AND (status = $3 OR $3 = '')
		  AND (created_at >= COALESCE(NULLIF($4, ''), created_at))
		  AND (created_at <= COALESCE(NULLIF($5, ''), created_at))
		ORDER BY %s %s
		LIMIT $6 OFFSET $7
	`, filter.Filter.SortColumn(), filter.Filter.SortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(
		ctx,
		query,
		filter.UserID,
		filter.ExportType,
		filter.Status,
		filter.MinDate,
		filter.MaxDate,
		filter.Filter.Limit(),
		filter.Filter.Offset(),
	)
	if err != nil {
		return nil, MetaData{}, err
	}
	defer rows.Close()

	exports := []*ExportHistory{}
	totalRecords := int64(0)

	for rows.Next() {
		export := &ExportHistory{}
		if err := rows.Scan(
			&export.ID,
			&export.UserID,
			&export.ExportType,
			&export.SpreadsheetID,
			&export.SheetName,
			&export.RowCount,
			&export.StartDate,
			&export.EndDate,
			&export.Status,
			&export.ErrorMessage,
			&export.CreatedAt,
		); err != nil {
			return nil, MetaData{}, err
		}
		exports = append(exports, export)
		totalRecords++
	}

	if err := rows.Err(); err != nil {
		return nil, MetaData{}, err
	}

	metadata := CalculateMetaData(totalRecords, filter.Filter.Page, filter.Filter.PageSize)

	return exports, metadata, nil
}

// GetSalesForExport retrieves sales records with joined user and product information for export.
func (m *SaleModel) GetSalesForExport(startDate, endDate *time.Time) ([]*SaleExportRecord, error) {
	query := `
		SELECT 
			s.id,
			s.sold_at,
			u.email,
			u.first_name || ' ' || u.last_name as user_name,
			p.name,
			p.price,
			s.quantity,
			p.price * s.quantity as total_amount
		FROM sales s
		JOIN users u ON s.user_id = u.id
		JOIN products p ON s.product_id = p.id
		WHERE (s.sold_at >= $1 OR $1 IS NULL)
		  AND (s.sold_at <= $2 OR $2 IS NULL)
		ORDER BY s.sold_at DESC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := []*SaleExportRecord{}

	for rows.Next() {
		record := &SaleExportRecord{}
		if err := rows.Scan(
			&record.TransactionID,
			&record.SoldAt,
			&record.UserEmail,
			&record.UserName,
			&record.ProductName,
			&record.ProductPrice,
			&record.Quantity,
			&record.TotalAmount,
		); err != nil {
			return nil, err
		}

		// Format date and time
		record.Date = record.SoldAt.Format("2006-01-02")
		record.Time = record.SoldAt.Format("15:04:05")

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}
