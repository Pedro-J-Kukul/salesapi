// / Filename: internal/data/permissions.go
package data

import (
	"context"
	"database/sql"
	"slices"
	"time"

	"github.com/lib/pq"
)

/************************************************************************************************************/
// Permission Declarations
/************************************************************************************************************/

// Permission struct to represent a permission in the system
type Permission struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
}

// PermissionModel struct to interact with the permissions table in the database
type PermissionModel struct {
	DB *sql.DB
}

// Permissions type to represent a list of permissions
type Permissions []string

// Includes - Check if a specific permission code exists in the Permissions slice
func (p Permissions) Includes(code string) bool {
	return slices.Contains(p, code)
}

/*************************************************************************************************************/
// Methods
/*************************************************************************************************************/

// GetAllForUser - Retrieve all permissions associated with a specific user role
func (m *PermissionModel) GetAllForUser(user_id int64) (Permissions, error) {
	query := `
		SELECT p.code
		FROM permissions p
		INNER JOIN users_permissions up ON up.permission_id = p.id
		INNER JOIN users u ON up.user_id = u.id
		WHERE up.user_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // Set a 3-second timeout
	defer cancel()                                                          // Ensure the context is canceled to free resources

	// Execute the query
	rows, err := m.DB.QueryContext(ctx, query, user_id)
	if err != nil {
		return nil, err
	}

	// Ensure the rows are closed after reading
	defer rows.Close()

	var permissions Permissions // Initialize an empty slice to hold permissions
	// Iterate through the result set and scan each permission code into the slice
	for rows.Next() {
		var code string // Temporary variable to hold the scanned code
		// Scan the code from the current row
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		permissions = append(permissions, code) // Append the scanned code to the permissions slice
	}

	return permissions, nil // Return the list of permissions and nil error
}

// AssignPermissions - Assign a list of permissions to a specific role
func (m *PermissionModel) AssignPermissions(userID int64, codes Permissions) error {
	// Remove duplicate codes using slices
	cleanCodes := slices.Compact(codes)

	query := `
		INSERT INTO users_permissions (user_id, permission_id)
		SELECT $1, p.id
		FROM permissions p
		WHERE p.code = ANY($2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the insert statement with the provided role ID and permission codes
	result, err := m.DB.ExecContext(ctx, query, userID, pq.Array(cleanCodes))
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNoRecords
	}

	return nil
}

// clearPermissions - Remove all permissions associated with a specific user
func (m *PermissionModel) ClearPermissions(userID int64) error {
	query := `
		DELETE FROM users_permissions
		WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the delete statement with the provided user ID
	result, err := m.DB.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNoRecords
	}

	return nil
}
