// File: internal/data/testutils.go
// Description: Database testing utilities for integration tests

package data

import (
	"database/sql"
	"fmt"
)

// TestUtils provides utility functions for testing database operations
type TestUtils struct {
	DB *sql.DB
}

// NewTestUtils creates a new TestUtils instance
func NewTestUtils(db *sql.DB) *TestUtils {
	return &TestUtils{DB: db}
}

// TruncateAllTables removes all data from all tables in the correct order to avoid foreign key constraints
func (tu *TestUtils) TruncateAllTables() error {
	tables := []string{
		"sales",
		"users_permissions",
		"tokens",
		"products",
		"permissions",
		"users",
	}

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		_, err := tu.DB.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	return nil
}

// ResetIdentitySequences resets all identity sequences to start from 1
func (tu *TestUtils) ResetIdentitySequences() error {
	sequences := []string{
		"users_id_seq",
		"permissions_id_seq",
		"tokens_id_seq",
		"products_id_seq",
		"sales_id_seq",
	}

	for _, seq := range sequences {
		query := fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH 1", seq)
		_, err := tu.DB.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to reset sequence %s: %w", seq, err)
		}
	}

	return nil
}

// CleanDatabase truncates all tables and resets sequences for a clean test environment
func (tu *TestUtils) CleanDatabase() error {
	if err := tu.TruncateAllTables(); err != nil {
		return err
	}
	if err := tu.ResetIdentitySequences(); err != nil {
		return err
	}
	return nil
}

// SeedTestPermissions creates a standard set of test permissions
func (tu *TestUtils) SeedTestPermissions() error {
	permissions := []string{
		"sales:create", "sales:read", "sales:update", "sales:delete",
		"products:create", "products:read", "products:update", "products:delete",
		"users:create", "users:read", "users:update", "users:delete",
		"self:create", "self:read", "self:update", "self:delete",
	}

	for _, perm := range permissions {
		query := `INSERT INTO permissions (code) VALUES ($1) ON CONFLICT (code) DO NOTHING`
		_, err := tu.DB.Exec(query, perm)
		if err != nil {
			return fmt.Errorf("failed to seed permission %s: %w", perm, err)
		}
	}

	return nil
}

// SeedTestUser creates a test user and returns the user ID
func (tu *TestUtils) SeedTestUser(email, firstName, lastName, role string, passwordHash []byte, isActive bool) (int64, error) {
	query := `
		INSERT INTO users (email, first_name, last_name, role, password_hash, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var userID int64
	err := tu.DB.QueryRow(query, email, firstName, lastName, role, passwordHash, isActive).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to seed test user: %w", err)
	}

	return userID, nil
}

// SeedTestProduct creates a test product and returns the product ID
func (tu *TestUtils) SeedTestProduct(name string, price float64) (int64, error) {
	query := `
		INSERT INTO products (name, price)
		VALUES ($1, $2)
		RETURNING id
	`

	var productID int64
	err := tu.DB.QueryRow(query, name, price).Scan(&productID)
	if err != nil {
		return 0, fmt.Errorf("failed to seed test product: %w", err)
	}

	return productID, nil
}

// SeedTestSale creates a test sale and returns the sale ID
func (tu *TestUtils) SeedTestSale(userID, productID, quantity int64) (int64, error) {
	query := `
		INSERT INTO sales (user_id, product_id, quantity)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var saleID int64
	err := tu.DB.QueryRow(query, userID, productID, quantity).Scan(&saleID)
	if err != nil {
		return 0, fmt.Errorf("failed to seed test sale: %w", err)
	}

	return saleID, nil
}

// AssignPermissionToUser assigns a permission to a user
func (tu *TestUtils) AssignPermissionToUser(userID int64, permissionCode string) error {
	query := `
		INSERT INTO users_permissions (user_id, permission_id)
		SELECT $1, id FROM permissions WHERE code = $2
	`

	_, err := tu.DB.Exec(query, userID, permissionCode)
	if err != nil {
		return fmt.Errorf("failed to assign permission %s to user %d: %w", permissionCode, userID, err)
	}

	return nil
}

// AssignPermissionsToUser assigns multiple permissions to a user
func (tu *TestUtils) AssignPermissionsToUser(userID int64, permissionCodes []string) error {
	for _, code := range permissionCodes {
		if err := tu.AssignPermissionToUser(userID, code); err != nil {
			return err
		}
	}
	return nil
}
