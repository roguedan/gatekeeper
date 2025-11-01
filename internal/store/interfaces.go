package store

import (
	"context"
)

// APIKeyRepositoryInterface defines the contract for API key storage operations
type APIKeyRepositoryInterface interface {
	CreateAPIKey(ctx context.Context, req APIKeyCreateRequest) (string, *APIKeyResponse, error)
	ValidateAPIKey(ctx context.Context, rawKey string) (*APIKey, error)
	ListAPIKeys(ctx context.Context, userID int64) ([]APIKey, error)
	GetAPIKeyByID(ctx context.Context, id int64) (*APIKey, error)
	DeleteAPIKey(ctx context.Context, id int64) error
	UpdateLastUsed(ctx context.Context, keyHash string) error
}

// UserRepositoryInterface defines the contract for user storage operations
type UserRepositoryInterface interface {
	GetOrCreateUserByAddress(ctx context.Context, address string) (*User, error)
	GetUserByAddress(ctx context.Context, address string) (*User, error)
	GetUserByID(ctx context.Context, id int64) (*User, error)
}

// AllowlistRepositoryInterface defines the contract for allowlist operations
type AllowlistRepositoryInterface interface {
	CreateAllowlist(ctx context.Context, userID int64, name, description string) (int64, error)
	ListAllowlists(ctx context.Context, userID int64) ([]Allowlist, error)
	AddAddressesToAllowlist(ctx context.Context, allowlistID int64, addresses []string) error
	RemoveAddressFromAllowlist(ctx context.Context, allowlistID int64, address string) error
	IsAddressInAllowlist(ctx context.Context, allowlistID int64, address string) (bool, error)
	DeleteAllowlist(ctx context.Context, id int64) error
}
