// File: internal/data/export.go
package data

import "database/sql"

type Models struct {
	Permissions PermissionModel
	Products    ProductModel
	Tokens      TokenModel
	Users       UserModel
	Sales       SaleModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Permissions: PermissionModel{DB: db},
		Products:    ProductModel{DB: db},
		Tokens:      TokenModel{DB: db},
		Users:       UserModel{DB: db},
		Sales:       SaleModel{DB: db},
	}
}
