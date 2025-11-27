// File: internal/data/users.go
package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"time"

	"github.com/Pedro-J-Kukul/salesapi/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

// ----------------------------------------------------------------------
//
//	Definitions
//
// ----------------------------------------------------------------------

// Password represents a hashed password.
type Password struct {
	hash      []byte  `db:"-"`
	plaintext *string `db:"-"`
}

// User represents a user in the system.
type User struct {
	ID        int64     `db:"id"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	Password  Password  `db:"-"`
	Role      string    `db:"role"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	IsActive  bool      `db:"is_active"`
	Version   int       `db:"version"`
}

// UserModel wraps a sql.DB connection pool.
type UserModel struct {
	DB *sql.DB
}

var AnonymousUser = &User{}

type UserFilter struct {
	Filter   Filter
	Name     string
	Email    string
	Role     string
	IsActive *bool
}

// ----------------------------------------------------------------------
//
//	Methods
//
// ----------------------------------------------------------------------
// Set hashes a plaintext password and stores it in the Password struct.
func (p *Password) Set(plaintextPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hashedPassword
	return nil
}

// Matches checks if the provided plaintext password matches the stored hashed password.
func (p *Password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser // Return true if the user is the anonymous user
}

// ValidatePassword checks the strength of a plaintext password.
func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")                                                              // Check if password is not empty
	v.Check(len(password) >= 8, "password", "must be at least 8 characters long")                                        // Check if password length is at least 8 characters
	v.Check(len(password) <= 72, "password", "must not be more than 72 characters long")                                 // Check if password length is within limit
	v.Check(v.Matches(password, validator.PasswordNumberRX), "password", "must contain at least one number")             // Check if password contains at least one number
	v.Check(v.Matches(password, validator.PasswordUpperRX), "password", "must contain at least one uppercase letter")    // Check if password contains at least one uppercase letter
	v.Check(v.Matches(password, validator.PasswordLowerRX), "password", "must contain at least one lowercase letter")    // Check if password contains at least one lowercase letter
	v.Check(v.Matches(password, validator.PasswordSpecialRX), "password", "must contain at least one special character") // Check if password contains at least one special character
}

// ValidateEmail checks if the email is in a valid format.
func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(len(email) <= 254, "email", "must not be more than 254 characters long")
	v.Check(v.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

// ValidateUser checks the fields of a User struct to ensure they meet the required criteria.
func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.FirstName != "", "first_name", "must be provided")
	v.Check(len(user.FirstName) <= 100, "first_name", "must not be more than 100 characters long")

	v.Check(user.LastName != "", "last_name", "must be provided")
	v.Check(len(user.LastName) <= 100, "last_name", "must not be more than 100 characters long")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	allowedRoles := []string{"admin", "cashier", "guest"}
	v.Check(v.Permitted(user.Role, allowedRoles...), "role", "must be one of the permitted values")
}

// ----------------------------------------------------------------------
//
//	Database interaction methods
//
// ----------------------------------------------------------------------
// Insert adds a new user to the database.
func (m *UserModel) Insert(user *User) error {
	query := `
		INSERT INTO users (first_name, last_name, email, password_hash, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, created_at, updated_at, version
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password.hash,
		user.Role,
		user.IsActive,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.Version)
	if err != nil {
		return err
	}
	return nil
}

// Update modifies an existing user in the database.
func (m *UserModel) Update(user *User) error {
	query := `
		UPDATE users
		SET first_name = $1, last_name = $2, email = $3, password_hash = $4, role = $5, is_active = $6, updated_at = NOW(), version = version + 1
		WHERE id = $7 AND version = $8
		RETURNING updated_at, version
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password.hash,
		user.Role,
		user.IsActive,
		user.ID,
		user.Version,
	).Scan(&user.UpdatedAt, &user.Version)
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a user from the database.
func (m *UserModel) Delete(id int64) error {
	query := `
		DELETE FROM users
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

// Get retrieves a user by its ID.
func (m *UserModel) GetByID(id int64) (*User, error) {
	query := `
		SELECT id, first_name, last_name, email, password_hash, role, is_active, created_at, updated_at, version
		FROM users
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user := &User{}

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password.hash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Version,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return user, nil
}

// GetByEmail retrieves a user by its email.
func (m *UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, first_name, last_name, email, password_hash, role, is_active, created_at, updated_at, version
		FROM users
		WHERE email = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user := &User{}

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password.hash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Version,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return user, nil
}

// GetAll retrieves a list of users based on the provided filter and pagination parameters.
func (m *UserModel) GetAll(filter UserFilter) ([]*User, MetaData, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, first_name, last_name, email, password_hash, role, is_active, created_at, updated_at, version
		FROM users
		WHERE (first_name ILIKE '%%' || $1 || '%%' OR last_name ILIKE '%%' || $1 || '%%')
		  AND (email ILIKE '%%' || $2 || '%%')
		  AND (role = COALESCE(NULLIF($3, ''), role))
		  AND (is_active = COALESCE($4, is_active))
		ORDER BY %s %s
		LIMIT $5 OFFSET $6
	`, filter.Filter.SortColumn(), filter.Filter.SortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{
		filter.Name,
		filter.Email,
		filter.Role,
		filter.IsActive,
		filter.Filter.Limit(),
		filter.Filter.Offset(),
	}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, MetaData{}, err
	}
	defer rows.Close()

	users := []*User{}
	totalRecords := int64(0)

	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&totalRecords,
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Email,
			&user.Password.hash,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.Version,
		)
		if err != nil {
			return nil, MetaData{}, err
		}
		users = append(users, user)
		totalRecords++
	}

	if err = rows.Err(); err != nil {
		return nil, MetaData{}, err
	}

	meta := CalculateMetaData(totalRecords, filter.Filter.Page, filter.Filter.PageSize)

	return users, meta, nil
}

// GetForTokens retrieves a user based on a token scope and plaintext token.
func (m *UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	query := `
		SELECT users.id, users.first_name, users.last_name, users.email, users.password_hash, users.role, users.is_active, users.created_at, users.updated_at, users.version
		FROM users
		INNER JOIN tokens
		ON users.id = tokens.user_id
		WHERE tokens.scope = $1
		AND tokens.hash = $2
		AND tokens.expires_at > $3
	`

	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user := &User{}

	err := m.DB.QueryRowContext(ctx, query, tokenScope, tokenHash[:], time.Now()).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password.hash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Version,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return user, nil
}
