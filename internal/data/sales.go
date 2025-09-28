// Filename: internal/data/sales.go
package data

import (
	"database/sql"
	"fmt"

	"github.com/Pedro-J-Kukul/salesapi/internal/validator"
	"github.com/lib/pq"
)

// Sale struct represents the sale record for the items sold
type Sale struct {
	ID        int        `json:"id"`
	Cashier   string     `json:"cashier"`
	Total     float32    `json:"total"`
	CashPaid  float32    `json:"cash_paid"`
	ChangeDue float32    `json:"change_due"`
	CreatedAt string     `json:"created_at"`
	Items     []SaleItem `json:"items,omitempty"` // Items included in this sale
}

// SaleItem struct represents an item in a sale
type SaleItem struct {
	ID             int     `json:"id"`
	MenuID         int     `json:"menu_id"`
	SaleID         int     `json:"sale_id"`
	Quantity       int     `json:"quantity"`
	UnitPrice      float32 `json:"unit_price"`
	LastModifiedBy string  `json:"last_modified_by"`
	CreatedAt      string  `json:"created_at"`
	MenuName       string  `json:"menu_name,omitempty"` // Name of the menu item
}

// CreateSaleRequest struct represents the request payload for creating a sale
type CreateSaleRequest struct {
	Cashier  string           `json:"cashier"`
	CashPaid float32          `json:"cash_paid"`
	Items    []CreateSaleItem `json:"items"`
}

// CreateSaleItem struct represents an item in the create sale request
type CreateSaleItem struct {
	MenuID   int `json:"menu_id"`
	Quantity int `json:"quantity"`
}

// ValidateSale checks the fields of the Sale struct for validity
func ValidateSale(v *validator.Validator, s *Sale) {
	v.Check(s.Cashier != "", "cashier_name", "must be provided")
	v.Check(len(s.Cashier) <= 255, "cashier_name", "must not be more than 255 bytes long")
	v.Check(s.Total > 0, "total_amount", "must be greater than zero")
	v.Check(s.CashPaid > 0, "cash_paid", "must be greater than zero")
	v.Check(s.CashPaid >= s.Total, "cash_paid", "must be greater than or equal to total amount")
	v.Check(len(s.Items) > 0, "items", "must contain at least one item")
}

// ValidateCreateSaleRequest checks the fields of the CreateSaleRequest struct
func ValidateCreateSaleRequest(v *validator.Validator, req *CreateSaleRequest) {
	v.Check(req.Cashier != "", "cashier_name", "must be provided")
	v.Check(len(req.Cashier) <= 255, "cashier_name", "must not be more than 255 bytes long")
	v.Check(req.CashPaid > 0, "cash_paid", "must be greater than zero")
	v.Check(len(req.Items) > 0, "items", "must contain at least one item")

	for i, item := range req.Items {
		v.Check(item.MenuID > 0, fmt.Sprintf("items[%d].menu_id", i), "must be greater than zero")
		v.Check(item.Quantity > 0, fmt.Sprintf("items[%d].quantity", i), "must be greater than zero")
	}
}

func CorrectCashPaid(total, cashPaid float32) error {
	if cashPaid < total {
		return ErrInsufficientCash
	}
	return nil
}

// SalesModel wraps a sql.DB connection pool
type SalesModel struct {
	DB *sql.DB
}

// Insert adds a new sale record to the database along with its items
func (m *SalesModel) Insert(req *CreateSaleRequest) (*Sale, error) {
	// Start a transaction
	tx, err := m.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ctx, cancel := getContext()
	defer cancel()

	// Calculate the total amount by getting the unit price from the menus table
	var totalAmount float32

	// Extract menu IDs from the request
	menuIDs := make([]int, 0, len(req.Items))
	for _, item := range req.Items {
		menuIDs = append(menuIDs, item.MenuID)
	}

	// Fetch menu items from the database - CORRECTED: Use pq.Array for PostgreSQL
	query := `SELECT id, name, price FROM menu WHERE id = ANY($1)`
	rows, err := tx.QueryContext(ctx, query, pq.Array(menuIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Create a map to hold menu items for easy lookup
	menuItems := make(map[int]*Menu)
	for rows.Next() {
		var menu Menu
		// scan into menu
		if err := rows.Scan(&menu.ID, &menu.Name, &menu.Price); err != nil {
			return nil, err
		}
		menuItems[menu.ID] = &menu
	}

	// Check if all requested menu items were found
	if len(menuItems) != len(menuIDs) {
		for _, id := range menuIDs {
			if _, exists := menuItems[id]; !exists {
				return nil, fmt.Errorf("menu item with ID %d not found", id)
			}
		}
	}

	// Calculate the total amount for the sale
	for _, item := range req.Items {
		menuItem := menuItems[item.MenuID] // Safe to access since we validated above
		totalAmount += menuItem.Price * float32(item.Quantity)
	}

	// Validate that cash paid is sufficient
	err = CorrectCashPaid(totalAmount, req.CashPaid)
	if err != nil {
		return nil, err
	}

	changeDue := req.CashPaid - totalAmount

	// Insert the sale
	sale := &Sale{
		Cashier:   req.Cashier,
		Total:     totalAmount,
		CashPaid:  req.CashPaid,
		ChangeDue: changeDue,
	}

	insertSaleQuery := `
		INSERT INTO sales (cashier, total, cash_paid, change_due)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	err = tx.QueryRowContext(ctx, insertSaleQuery, sale.Cashier, sale.Total, sale.CashPaid, sale.ChangeDue).Scan(&sale.ID, &sale.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Insert sale items
	saleItems := make([]SaleItem, 0, len(req.Items))
	for _, item := range req.Items {
		menu := menuItems[item.MenuID]

		saleItemQuery := `
			INSERT INTO menu_sales (menu_id, sale_id, quantity, price, last_modified_by)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, created_at`

		var saleItem SaleItem
		err = tx.QueryRowContext(ctx, saleItemQuery,
			item.MenuID,
			sale.ID,
			item.Quantity,
			menu.Price,
			req.Cashier,
		).Scan(&saleItem.ID, &saleItem.CreatedAt)
		if err != nil {
			return nil, err
		}

		saleItem.MenuID = item.MenuID
		saleItem.SaleID = sale.ID
		saleItem.Quantity = item.Quantity
		saleItem.UnitPrice = menu.Price
		saleItem.LastModifiedBy = req.Cashier
		saleItem.MenuName = menu.Name

		saleItems = append(saleItems, saleItem)
	}

	sale.Items = saleItems

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return sale, nil
}

// Delete removes a sale and all its associated items (cascade delete handles this)
func (s *SalesModel) Delete(id int) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM sales WHERE id = $1`

	ctx, cancel := getContext()
	defer cancel()

	result, err := s.DB.ExecContext(ctx, query, id)
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
