package store

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound is returned when a requested resource is not found
	ErrNotFound = errors.New("resource not found")

	// ErrDuplicate is returned when attempting to create a resource that already exists
	ErrDuplicate = errors.New("resource already exists")

	// ErrInvalidAddress is returned when an Ethereum address is invalid
	ErrInvalidAddress = errors.New("invalid ethereum address")

	// ErrExpired is returned when a resource has expired
	ErrExpired = errors.New("resource has expired")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")
)

// NotFoundError wraps ErrNotFound with additional context
type NotFoundError struct {
	Resource string
	ID       interface{}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %v", e.Resource, e.ID)
}

func (e *NotFoundError) Unwrap() error {
	return ErrNotFound
}

// DuplicateError wraps ErrDuplicate with additional context
type DuplicateError struct {
	Resource string
	Field    string
	Value    interface{}
}

func (e *DuplicateError) Error() string {
	return fmt.Sprintf("%s already exists with %s: %v", e.Resource, e.Field, e.Value)
}

func (e *DuplicateError) Unwrap() error {
	return ErrDuplicate
}

// InvalidAddressError wraps ErrInvalidAddress with additional context
type InvalidAddressError struct {
	Address string
	Reason  string
}

func (e *InvalidAddressError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("invalid ethereum address %s: %s", e.Address, e.Reason)
	}
	return fmt.Sprintf("invalid ethereum address: %s", e.Address)
}

func (e *InvalidAddressError) Unwrap() error {
	return ErrInvalidAddress
}

// ExpiredError wraps ErrExpired with additional context
type ExpiredError struct {
	Resource string
	ID       interface{}
}

func (e *ExpiredError) Error() string {
	return fmt.Sprintf("%s has expired: %v", e.Resource, e.ID)
}

func (e *ExpiredError) Unwrap() error {
	return ErrExpired
}
