// File> internal/data/products.go
package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Pedro-J-Kukul/salesapi/internal/validator"
	"github.com/lib/pq"
)

// ----------------------------------------------------------------------
//
//	Definitions
//
// ----------------------------------------------------------------------

// Product represents a product in the system.
type Product struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProductModel wraps a sql.DB connection pool.
type ProductModel struct {
	DB *sql.DB
}

// ProductFilter represents filtering criteria for querying products.
type ProductFilter struct {
	Filter   Filter  `json:"filter"`
	MinPrice float64 `json:"min_price"`
	MaxPrice float64 `json:"max_price"`
	Name     string  `json:"name"`
}

// ----------------------------------------------------------------------
//
//	Methods
//
// ----------------------------------------------------------------------
// ValidateProduct checks the fields of a Product struct to ensure they meet the required criteria.
func ValidateProduct(v *validator.Validator, product *Product) {
	v.Check(product.Name != "", "name", "must be provided")
	v.Check(len(product.Name) <= 200, "name", "must not be more than 200 bytes long")
	v.Check(product.Price >= 0, "price", "must be a non-negative number")
}

// Insert adds a new product to the database.
func (m *ProductModel) Insert(product *Product) error {
	query := `
		INSERT INTO products (name, price, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := m.DB.QueryRowContext(ctx, query, product.Name, product.Price).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt); err != nil {
		if pqError, ok := err.(*pq.Error); ok {
			switch pqError.Code {
			case "23514": // check_violation
				return ErrInvalidData
			case "23502": // not_null_violation
				return ErrInvalidData
			}
		}
		return err
	}
	return nil
}

// Update modifies an existing product in the database.
func (m *ProductModel) Update(product *Product) error {
	query := `
		UPDATE products
		SET name = $1, price = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := m.DB.QueryRowContext(ctx, query, product.Name, product.Price, product.ID).Scan(&product.UpdatedAt); err != nil {
		return err
	}
	return nil
}

// Delete removes a product from the database.
func (m *ProductModel) Delete(id int64) error {
	query := `
		DELETE FROM products
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	if rowsAffected, err := result.RowsAffected(); err != nil {
		return err
	} else if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Get retrieves a product by its ID.
func (m *ProductModel) Get(id int64) (*Product, error) {
	query := `
		SELECT id, name, price, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	product := &Product{}
	if err := m.DB.QueryRowContext(ctx, query, id).Scan(&product.ID, &product.Name, &product.Price, &product.CreatedAt, &product.UpdatedAt); err != nil {
		return nil, err
	}
	return product, nil
}

// GetAll retrieves products based on filtering criteria and pagination.
func (m *ProductModel) GetAll(filter ProductFilter) ([]*Product, MetaData, error) {
	query := fmt.Sprintf(`
		SELECT id, name, price, created_at, updated_at
		FROM products
		WHERE (price >= $1 OR $1 = 0)
		  AND (price <= $2 OR $2 = 0)
		  AND (name ILIKE '%%' || $3 || '%%' OR $3 = '')
		ORDER BY %s %s
		LIMIT $4 OFFSET $5
	`, filter.Filter.SortColumn(), filter.Filter.SortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, filter.MinPrice, filter.MaxPrice, filter.Name, filter.Filter.Limit(), filter.Filter.Offset())
	if err != nil {
		return nil, MetaData{}, err
	}
	defer rows.Close()

	products := []*Product{}
	totalRecords := int64(0)

	for rows.Next() {
		product := &Product{}
		if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, MetaData{}, err
		}
		products = append(products, product)
		totalRecords++
	}

	if err := rows.Err(); err != nil {
		return nil, MetaData{}, err
	}

	metadata := CalculateMetaData(totalRecords, filter.Filter.Page, filter.Filter.PageSize)

	return products, metadata, nil
}
