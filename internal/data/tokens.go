// File: internal/data/tokens.go
package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"time"

	"github.com/Pedro-J-Kukul/salesapi/internal/validator"
)

// ----------------------------------------------------------------------
//
//	Definitions
//
// ----------------------------------------------------------------------

// Scope constants for different token types.
const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
	ScopePasswordReset  = "password_reset"
)

// Token represents a token used for various purposes in the system.
type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	Scope     string    `json:"scope"`
}

// TokenModel wraps a sql.DB connection pool.
type TokenModel struct {
	DB *sql.DB
}

// ----------------------------------------------------------------------
//
//	Methods
//
// ----------------------------------------------------------------------

func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID:    userID,
		ExpiresAt: time.Now().Add(ttl),
		Scope:     scope,
	}

	// Generate a random plaintext token (implementation omitted for brevity).
	randomBytes := make([]byte, 16) // Example size
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, err
	}
	token.Plaintext = base64.RawURLEncoding.EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}

func ValidateTokenPlaintext(v *validator.Validator, plaintext string) {
	v.Check(plaintext != "", "token", "must be provided")
	v.Check(len(plaintext) == 22, "token", "must be 22 bytes long")
}

// ----------------------------------------------------------------------
//
//	Database Operations
//
// ----------------------------------------------------------------------
// New creates a new token, inserts it into the database, and returns it.
func (m *TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}
	err = m.Insert(token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// Insert inserts a new token into the database.
func (m *TokenModel) Insert(token *Token) error {
	query := `
		INSERT INTO tokens (hash, user_id, expires_at, scope)
		VALUES ($1, $2, $3, $4)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, token.Hash, token.UserID, token.ExpiresAt, token.Scope)
	return err
}

// DeleteAllForUser deletes all tokens for a specific user and scope.
func (m *TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `
		DELETE FROM tokens
		WHERE scope = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}
