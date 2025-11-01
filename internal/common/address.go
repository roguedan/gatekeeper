package common

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

var (
	// ethereumAddressRegex validates Ethereum address format (0x or 0X + 40 hex characters)
	// Case-insensitive validation
	ethereumAddressRegex = regexp.MustCompile(`^0[xX][0-9a-fA-F]{40}$`)
)

// AddressError represents an error related to address validation
type AddressError struct {
	Address string
	Reason  string
}

func (e *AddressError) Error() string {
	return fmt.Sprintf("invalid address %q: %s", e.Address, e.Reason)
}

// NormalizeAddress validates and normalizes an Ethereum address.
//
// This is the canonical function for address normalization in the entire application.
// All address validation and normalization should go through this function.
//
// Validation rules:
//   - Address cannot be empty
//   - Must have "0x" prefix
//   - Must be exactly 40 hexadecimal characters (case-insensitive)
//   - If mixed case, validates EIP-55 checksum
//
// Returns:
//   - Lowercase normalized address with "0x" prefix
//   - AddressError if validation fails
func NormalizeAddress(address string) (string, error) {
	// Trim whitespace
	address = strings.TrimSpace(address)

	// Check for empty address
	if address == "" {
		return "", &AddressError{
			Address: address,
			Reason:  "address cannot be empty",
		}
	}

	// Check format: 0x + 40 hex chars (case-insensitive)
	if !ethereumAddressRegex.MatchString(address) {
		return "", &AddressError{
			Address: address,
			Reason:  "must be 0x followed by 40 hexadecimal characters",
		}
	}

	// Normalize to lowercase 0x prefix
	if address[1] == 'X' {
		address = "0x" + address[2:]
	}

	// Check if address has mixed case (potential EIP-55 checksum)
	hexPart := address[2:] // Remove "0x" prefix
	hasUpperCase := false
	hasLowerCase := false

	for _, c := range hexPart {
		if c >= 'A' && c <= 'F' {
			hasUpperCase = true
		} else if c >= 'a' && c <= 'f' {
			hasLowerCase = true
		}
		if hasUpperCase && hasLowerCase {
			break
		}
	}

	// If address has mixed case, validate EIP-55 checksum
	if hasUpperCase && hasLowerCase {
		// Use go-ethereum to validate checksum
		ethAddr := common.HexToAddress(address)
		checksummed := ethAddr.Hex()

		if address != checksummed {
			return "", &AddressError{
				Address: address,
				Reason:  fmt.Sprintf("invalid EIP-55 checksum (expected %s)", checksummed),
			}
		}
	}

	// Return lowercase normalized address with 0x prefix
	return strings.ToLower(address), nil
}

// IsValidAddress performs a quick format check on an Ethereum address.
//
// This function only validates the format (0x + 40 hex chars) without
// validating EIP-55 checksums. Use NormalizeAddress for full validation.
//
// Returns:
//   - true if address matches the basic format
//   - false otherwise
func IsValidAddress(addr string) bool {
	if addr == "" {
		return false
	}
	return ethereumAddressRegex.MatchString(addr)
}

// ChecksumAddress returns the EIP-55 checksummed version of an address.
//
// The EIP-55 checksum encoding uses mixed case to encode a checksum
// into the address itself, allowing detection of typos.
//
// Parameters:
//   - addr: Ethereum address (can be any case, with or without 0x prefix)
//
// Returns:
//   - EIP-55 checksummed address (mixed case with 0x prefix)
//   - AddressError if the address format is invalid
func ChecksumAddress(addr string) (string, error) {
	// First validate and normalize
	normalized, err := NormalizeAddress(addr)
	if err != nil {
		return "", err
	}

	// Use go-ethereum to generate checksum
	ethAddr := common.HexToAddress(normalized)
	return ethAddr.Hex(), nil
}

// CompareAddresses compares two Ethereum addresses for equality.
//
// Comparison is case-insensitive and ignores checksum formatting.
// Both addresses must be valid Ethereum addresses.
//
// Returns:
//   - true if addresses are equal (case-insensitive)
//   - false if addresses differ or if either address is invalid
func CompareAddresses(addr1, addr2 string) bool {
	norm1, err1 := NormalizeAddress(addr1)
	norm2, err2 := NormalizeAddress(addr2)

	if err1 != nil || err2 != nil {
		return false
	}

	return norm1 == norm2
}
