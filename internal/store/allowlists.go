package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// Allowlist represents an allowlist of Ethereum addresses
type Allowlist struct {
	ID          int64     `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// AllowlistEntry represents a single address entry in an allowlist
type AllowlistEntry struct {
	ID          int64     `db:"id"`
	AllowlistID int64     `db:"allowlist_id"`
	Address     string    `db:"address"`
	AddedAt     time.Time `db:"added_at"`
}

// AllowlistWithCount includes the allowlist with entry count
type AllowlistWithCount struct {
	Allowlist
	EntryCount int64 `db:"entry_count"`
}

// AllowlistRepository provides methods for managing allowlists
type AllowlistRepository struct {
	db *DB
}

// NewAllowlistRepository creates a new AllowlistRepository
func NewAllowlistRepository(db *DB) *AllowlistRepository {
	return &AllowlistRepository{db: db}
}

// CreateAllowlist creates a new allowlist
func (r *AllowlistRepository) CreateAllowlist(ctx context.Context, name, description string) (*Allowlist, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	allowlist := &Allowlist{
		Name:        name,
		Description: description,
	}

	query := `
		INSERT INTO allowlists (name, description, created_at, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, name, description, created_at, updated_at
	`

	err := r.db.QueryRowxContext(ctx, query, name, description).StructScan(allowlist)
	if err != nil {
		// Check for duplicate key violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, &DuplicateError{
				Resource: "allowlist",
				Field:    "name",
				Value:    name,
			}
		}
		return nil, fmt.Errorf("failed to create allowlist: %w", err)
	}

	return allowlist, nil
}

// GetAllowlist retrieves an allowlist by ID
func (r *AllowlistRepository) GetAllowlist(ctx context.Context, id int64) (*Allowlist, error) {
	var allowlist Allowlist
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM allowlists
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &allowlist, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &NotFoundError{
				Resource: "allowlist",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to get allowlist: %w", err)
	}

	return &allowlist, nil
}

// ListAllowlists returns all allowlists with entry counts, ordered by created_at DESC
func (r *AllowlistRepository) ListAllowlists(ctx context.Context) ([]AllowlistWithCount, error) {
	var allowlists []AllowlistWithCount
	query := `
		SELECT
			a.id,
			a.name,
			a.description,
			a.created_at,
			a.updated_at,
			COALESCE(COUNT(ae.id), 0) as entry_count
		FROM allowlists a
		LEFT JOIN allowlist_entries ae ON a.id = ae.allowlist_id
		GROUP BY a.id, a.name, a.description, a.created_at, a.updated_at
		ORDER BY a.created_at DESC
	`

	err := r.db.SelectContext(ctx, &allowlists, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list allowlists: %w", err)
	}

	// Return empty slice instead of nil if no allowlists found
	if allowlists == nil {
		allowlists = []AllowlistWithCount{}
	}

	return allowlists, nil
}

// UpdateAllowlist updates an allowlist's name and/or description
func (r *AllowlistRepository) UpdateAllowlist(ctx context.Context, allowlist *Allowlist) error {
	if allowlist == nil {
		return fmt.Errorf("allowlist cannot be nil")
	}
	if allowlist.Name == "" {
		return fmt.Errorf("name is required")
	}

	query := `
		UPDATE allowlists
		SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING updated_at
	`

	err := r.db.QueryRowxContext(ctx, query, allowlist.Name, allowlist.Description, allowlist.ID).Scan(&allowlist.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return &NotFoundError{
				Resource: "allowlist",
				ID:       allowlist.ID,
			}
		}
		// Check for duplicate key violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return &DuplicateError{
				Resource: "allowlist",
				Field:    "name",
				Value:    allowlist.Name,
			}
		}
		return fmt.Errorf("failed to update allowlist: %w", err)
	}

	return nil
}

// DeleteAllowlist deletes an allowlist and all its entries (cascade delete)
func (r *AllowlistRepository) DeleteAllowlist(ctx context.Context, id int64) error {
	// Begin transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete allowlist entries first
	_, err = tx.ExecContext(ctx, `DELETE FROM allowlist_entries WHERE allowlist_id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete allowlist entries: %w", err)
	}

	// Delete the allowlist
	result, err := tx.ExecContext(ctx, `DELETE FROM allowlists WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete allowlist: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &NotFoundError{
			Resource: "allowlist",
			ID:       id,
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// AddAddress adds a single address to an allowlist
func (r *AllowlistRepository) AddAddress(ctx context.Context, allowlistID int64, address string) error {
	// Validate and normalize address using canonical function
	normalizedAddress, err := validateAddress(address)
	if err != nil {
		return err
	}

	// Begin transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert address (using ON CONFLICT DO NOTHING for idempotency)
	insertQuery := `
		INSERT INTO allowlist_entries (allowlist_id, address, added_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (allowlist_id, address) DO NOTHING
	`

	_, err = tx.ExecContext(ctx, insertQuery, allowlistID, normalizedAddress)
	if err != nil {
		// Check for foreign key violation (allowlist doesn't exist)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			return &NotFoundError{
				Resource: "allowlist",
				ID:       allowlistID,
			}
		}
		return fmt.Errorf("failed to add address to allowlist: %w", err)
	}

	// Update allowlist's updated_at timestamp
	updateQuery := `UPDATE allowlists SET updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err = tx.ExecContext(ctx, updateQuery, allowlistID)
	if err != nil {
		return fmt.Errorf("failed to update allowlist timestamp: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// RemoveAddress removes an address from an allowlist
func (r *AllowlistRepository) RemoveAddress(ctx context.Context, allowlistID int64, address string) error {
	// Validate and normalize address
	normalizedAddress, err := validateAddress(address)
	if err != nil {
		return err
	}

	// Begin transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete address
	deleteQuery := `DELETE FROM allowlist_entries WHERE allowlist_id = $1 AND address = $2`
	result, err := tx.ExecContext(ctx, deleteQuery, allowlistID, normalizedAddress)
	if err != nil {
		return fmt.Errorf("failed to remove address from allowlist: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &NotFoundError{
			Resource: "allowlist_entry",
			ID:       fmt.Sprintf("allowlist_id=%d, address=%s", allowlistID, normalizedAddress),
		}
	}

	// Update allowlist's updated_at timestamp
	updateQuery := `UPDATE allowlists SET updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err = tx.ExecContext(ctx, updateQuery, allowlistID)
	if err != nil {
		return fmt.Errorf("failed to update allowlist timestamp: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// AddAddresses batch adds multiple addresses to an allowlist
func (r *AllowlistRepository) AddAddresses(ctx context.Context, allowlistID int64, addresses []string) error {
	if len(addresses) == 0 {
		return nil
	}

	// Validate and normalize all addresses first
	normalizedAddresses := make([]string, len(addresses))
	for i, addr := range addresses {
		normalized, err := validateAddress(addr)
		if err != nil {
			return fmt.Errorf("invalid address at index %d: %w", i, err)
		}
		normalizedAddresses[i] = normalized
	}

	// Begin transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert addresses in batch
	insertQuery := `
		INSERT INTO allowlist_entries (allowlist_id, address, added_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (allowlist_id, address) DO NOTHING
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, addr := range normalizedAddresses {
		_, err = stmt.ExecContext(ctx, allowlistID, addr)
		if err != nil {
			// Check for foreign key violation (allowlist doesn't exist)
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
				return &NotFoundError{
					Resource: "allowlist",
					ID:       allowlistID,
				}
			}
			return fmt.Errorf("failed to insert address %s: %w", addr, err)
		}
	}

	// Update allowlist's updated_at timestamp
	updateQuery := `UPDATE allowlists SET updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err = tx.ExecContext(ctx, updateQuery, allowlistID)
	if err != nil {
		return fmt.Errorf("failed to update allowlist timestamp: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CheckAddress checks if an address exists in an allowlist (fast query)
func (r *AllowlistRepository) CheckAddress(ctx context.Context, allowlistID int64, address string) (bool, error) {
	// Validate and normalize address
	normalizedAddress, err := validateAddress(address)
	if err != nil {
		return false, err
	}

	// Use EXISTS subquery for performance
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM allowlist_entries
			WHERE allowlist_id = $1 AND address = $2
		)
	`

	var exists bool
	err = r.db.QueryRowxContext(ctx, query, allowlistID, normalizedAddress).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check address in allowlist: %w", err)
	}

	return exists, nil
}

// GetAddresses returns all addresses in an allowlist, sorted alphabetically
func (r *AllowlistRepository) GetAddresses(ctx context.Context, allowlistID int64) ([]string, error) {
	var addresses []string
	query := `
		SELECT address
		FROM allowlist_entries
		WHERE allowlist_id = $1
		ORDER BY address ASC
	`

	err := r.db.SelectContext(ctx, &addresses, query, allowlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses from allowlist: %w", err)
	}

	// Return empty slice instead of nil if no addresses found
	if addresses == nil {
		addresses = []string{}
	}

	return addresses, nil
}
