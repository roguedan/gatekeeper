package store

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/lib/pq"
)

var (
	// ethereumAddressRegex validates Ethereum address format (0x + 40 hex characters)
	ethereumAddressRegex = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
)

// User represents a user in the database
type User struct {
	ID        int64     `db:"id"`
	Address   string    `db:"address"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// UserRepository handles database operations for users
type UserRepository struct {
	db *DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

// validateAddress validates and normalizes an Ethereum address
func validateAddress(address string) (string, error) {
	if address == "" {
		return "", &InvalidAddressError{
			Address: address,
			Reason:  "address cannot be empty",
		}
	}

	// Normalize address to lowercase
	normalized := strings.ToLower(strings.TrimSpace(address))

	// Check format
	if !ethereumAddressRegex.MatchString(normalized) {
		return "", &InvalidAddressError{
			Address: address,
			Reason:  "must be 0x followed by 40 hexadecimal characters",
		}
	}

	return normalized, nil
}

// GetOrCreateUserByAddress gets a user by address or creates one if it doesn't exist
func (r *UserRepository) GetOrCreateUserByAddress(ctx context.Context, address string) (*User, error) {
	// Normalize address to lowercase
	address = strings.ToLower(address)

	// Try to get existing user
	user, err := r.GetUserByAddress(ctx, address)
	if err == nil {
		return user, nil
	}

	// If user doesn't exist, create it
	if err != nil && err.Error() == "user not found" {
		return r.CreateUser(ctx, address)
	}

	return nil, err
}

// GetUserByAddress retrieves a user by their Ethereum address
func (r *UserRepository) GetUserByAddress(ctx context.Context, address string) (*User, error) {
	// Validate and normalize address
	normalizedAddress, err := validateAddress(address)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, address, created_at, updated_at
		FROM users
		WHERE address = $1
	`

	user := &User{}
	err = r.db.QueryRowContext(ctx, query, normalizedAddress).Scan(
		&user.ID,
		&user.Address,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &NotFoundError{
				Resource: "user",
				ID:       normalizedAddress,
			}
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by their ID
func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT id, address, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Address,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &NotFoundError{
				Resource: "user",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return user, nil
}

// CreateUser creates a new user with the given address
func (r *UserRepository) CreateUser(ctx context.Context, address string) (*User, error) {
	// Validate and normalize address
	normalizedAddress, err := validateAddress(address)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO users (address, created_at, updated_at)
		VALUES ($1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, address, created_at, updated_at
	`

	user := &User{}
	err = r.db.QueryRowContext(ctx, query, normalizedAddress).Scan(
		&user.ID,
		&user.Address,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		// Check for duplicate key violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, &DuplicateError{
				Resource: "user",
				Field:    "address",
				Value:    normalizedAddress,
			}
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// UpdateUser updates an existing user
func (r *UserRepository) UpdateUser(ctx context.Context, user *User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	// Validate and normalize address
	normalizedAddress, err := validateAddress(user.Address)
	if err != nil {
		return err
	}

	query := `
		UPDATE users
		SET address = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		RETURNING updated_at
	`

	err = r.db.QueryRowContext(ctx, query, normalizedAddress, user.ID).Scan(&user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return &NotFoundError{
				Resource: "user",
				ID:       user.ID,
			}
		}
		// Check for duplicate key violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return &DuplicateError{
				Resource: "user",
				Field:    "address",
				Value:    normalizedAddress,
			}
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	user.Address = normalizedAddress
	return nil
}

// DeleteUser deletes a user by ID
func (r *UserRepository) DeleteUser(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &NotFoundError{
			Resource: "user",
			ID:       id,
		}
	}

	return nil
}
