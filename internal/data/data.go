// Filename: /internal/data/data.go
// Purpose: Contains exported methods for data manipulation
package data

import "database/sql"

// Models wraps all data models for use with db
type Models struct {
	Menu  MenuModel
	Sales SalesModel
}

// NewModels initializes the Models struct with a given database connection
func NewModels(db *sql.DB) Models {
	return Models{
		Menu:  MenuModel{DB: db},
		Sales: SalesModel{DB: db},
	}
}
