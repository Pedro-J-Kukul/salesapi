// Filename: internal/data/menu.go
// Description: Defines the Menu struct and related database operations.
package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Pedro-J-Kukul/salesapi/internal/validator"
)

type Menu struct {
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	Price          float32 `json:"price"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
	LastModifiedBy string  `json:"last_modified_by"`
}

// Validate checks the fields of the Menu struct for validity.
func Validate(v *validator.Validator, m *Menu) {
	v.Check(m.Name != "", "name", "must be provided")
	v.Check(len(m.Name) <= 100, "name", "must not be more than 100 bytes long")
	v.Check(m.Price > 0, "price", "must be greater than zero")
	v.Check(m.Price <= 100, "price", "must be a maximum of 100")
	v.Check(m.LastModifiedBy != "", "last_modified_by", "must be provided")
	v.Check(len(m.LastModifiedBy) <= 100, "last_modified_by", "must not be more than 100 bytes long")
}

// MenuModel wraps a sql.DB connection pool.
type MenuModel struct {
	DB *sql.DB
}

// Insert adds a new menu item to the database.
func (m *MenuModel) Insert(menu *Menu) error {
	query := `
		INSERT INTO menu (name, price, last_modified_by)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	args := []any{
		menu.Name,
		menu.Price,
		menu.LastModifiedBy,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).
		Scan(&menu.ID, &menu.CreatedAt, &menu.UpdatedAt)
}

// Update modifies an existing menu item in the database.
func (m *MenuModel) Update(menu *Menu) error {
	query := `
		UPDATE menu
		SET name = $1, price = $2, updated_at = NOW(), last_modified_by = $3
		WHERE id = $4
		RETURNING updated_at`

	args := []any{
		menu.Name,
		menu.Price,
		menu.LastModifiedBy,
		menu.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&menu.UpdatedAt)
}

// Delete removes a menu item from the database by ID.
func (m *MenuModel) Delete(id int) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM menu
		WHERE id = $1`

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

// Get retrieves a menu item from the database by ID.
func (m *MenuModel) Get(id int) (*Menu, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, name, price, created_at, updated_at, last_modified_by
		FROM menu
		WHERE id = $1`

	var menu Menu

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&menu.ID,
		&menu.Name,
		&menu.Price,
		&menu.CreatedAt,
		&menu.UpdatedAt,
		&menu.LastModifiedBy,
	)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &menu, nil
}

// GetAll retrieves all menu items from the database with optional filtering, sorting, and pagination.
func (m *MenuModel) GetAll(name string, price float32, lastModified string, filters Filters) ([]*Menu, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, name, price, created_at, updated_at, last_modified_by
		FROM menu
		WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (price = $2 OR $2 = 0)
		AND (to_tsvector('simple', last_modified_by) @@ plainto_tsquery('simple', $3) OR $3 = '')
		ORDER BY %s %s, id ASC
		LIMIT $4 OFFSET $5`, filters.SortColumn(), filters.SortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, name, price, lastModified, filters.Limit(), filters.Offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	menus := []*Menu{}

	for rows.Next() {
		var menu Menu
		err := rows.Scan(
			&totalRecords,
			&menu.ID,
			&menu.Name,
			&menu.Price,
			&menu.CreatedAt,
			&menu.UpdatedAt,
			&menu.LastModifiedBy,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		menus = append(menus, &menu)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := CalculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return menus, metadata, nil
}
