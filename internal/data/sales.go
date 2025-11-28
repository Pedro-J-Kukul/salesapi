// File internal/data/sales.go
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

// Sale represents a sales record in the system.
type Sale struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	ProductID int64     `json:"product_id"`
	Quantity  int64     `json:"quantity"`
	SoldAt    time.Time `json:"sold_at"`
}

// SaleModel wraps a sql.DB connection pool.
type SaleModel struct {
	DB *sql.DB
}

// SaleFilter represents filtering criteria for querying sales.
type SaleFilter struct {
	Filter    Filter `json:"filter"`
	UserID    int64  `json:"user_id"`
	ProductID int64  `json:"product_id"`
	MinDate   string `json:"min_date"`
	MaxDate   string `json:"max_date"`
	MinQty    int64  `json:"min_qty"`
	MaxQty    int64  `json:"max_qty"`
}

// ----------------------------------------------------------------------
//
//	Methods
//
// ----------------------------------------------------------------------
// ValidateSale checks the fields of a Sale struct to ensure they meet the required criteria.
func ValidateSale(v *validator.Validator, sale *Sale) {
	v.Check(sale.UserID > 0, "user_id", "must be a positive integer")
	v.Check(sale.ProductID > 0, "product_id", "must be a positive integer")
	v.Check(sale.Quantity > 0, "quantity", "must be a positive integer")
}

// Insert adds a new sale to the database.
func (m *SaleModel) Insert(sale *Sale) error {
	query := `
		INSERT INTO sales (user_id, product_id, quantity, sold_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, sold_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := m.DB.QueryRowContext(ctx, query, sale.UserID, sale.ProductID, sale.Quantity).Scan(&sale.ID, &sale.SoldAt); err != nil {
		return err
	}
	return nil
}

// Update modifies an existing sale in the database.
func (m *SaleModel) Update(sale *Sale) error {
	query := `
		UPDATE sales
		SET user_id = $1, product_id = $2, quantity = $3, sold_at = NOW()
		WHERE id = $4
		RETURNING sold_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := m.DB.QueryRowContext(ctx, query, sale.UserID, sale.ProductID, sale.Quantity, sale.ID).Scan(&sale.SoldAt); err != nil {
		return err
	}
	return nil
}

// Delete removes a sale from the database.
func (m *SaleModel) Delete(id int64) error {
	query := `
		DELETE FROM sales
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// Get retrieves a sale by its ID.
func (m *SaleModel) Get(id int64) (*Sale, error) {
	query := `
		SELECT id, user_id, product_id, quantity, sold_at
		FROM sales
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	sale := &Sale{}

	if err := m.DB.QueryRowContext(ctx, query, id).Scan(&sale.ID, &sale.UserID, &sale.ProductID, &sale.Quantity, &sale.SoldAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return sale, nil
}

// GetAll retrieves sales based on filtering criteria and pagination.
func (m *SaleModel) GetAll(filter SaleFilter) ([]*Sale, MetaData, error) {
	query := fmt.Sprintf(`
        SELECT COUNT(*) OVER(), id, user_id, product_id, quantity, sold_at
        FROM sales
        WHERE (user_id = $1 OR $1 = 0)
          AND (product_id = $2 OR $2 = 0)
          AND (CASE WHEN $3 = '' THEN TRUE ELSE sold_at >= $3::timestamp END)
          AND (CASE WHEN $4 = '' THEN TRUE ELSE sold_at <= $4::timestamp END)
          AND (quantity >= $5 OR $5 = 0)
          AND (quantity <= $6 OR $6 = 0)
        ORDER BY %s %s
        LIMIT $7 OFFSET $8
    `, filter.Filter.SortColumn(), filter.Filter.SortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := m.DB.QueryContext(ctx, query, filter.UserID, filter.ProductID, filter.MinDate, filter.MaxDate, filter.MinQty, filter.MaxQty, filter.Filter.Limit(), filter.Filter.Offset())
	if err != nil {
		return nil, MetaData{}, err
	}
	defer rows.Close()

	sales := []*Sale{}
	totalRecords := int64(0)

	for rows.Next() {
		sale := &Sale{}
		if err := rows.Scan(&totalRecords, &sale.ID, &sale.UserID, &sale.ProductID, &sale.Quantity, &sale.SoldAt); err != nil {
			return nil, MetaData{}, err
		}
		sales = append(sales, sale)
	}

	if err := rows.Err(); err != nil {
		return nil, MetaData{}, err
	}

	metadata := CalculateMetaData(totalRecords, filter.Filter.Page, filter.Filter.PageSize)

	return sales, metadata, nil
}
