package common

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test addresses
const (
	// Valid addresses in different formats
	validAddressLower     = "0xabcdef0123456789abcdef0123456789abcdef01"
	validAddressUpper     = "0xABCDEF0123456789ABCDEF0123456789ABCDEF01"
	validAddressNoPrefix  = "abcdef0123456789abcdef0123456789abcdef01"

	// EIP-55 checksummed addresses (from Ethereum examples)
	validChecksumAddr1    = "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed"
	validChecksumAddr2    = "0xfB6916095ca1df60bB79Ce92cE3Ea74c37c5d359"
	validChecksumAddr3    = "0xdbF03B407c01E7cD3CBea99509d93f8DDDC8C6FB"
	validChecksumAddr4    = "0xD1220A0cf47c7B9Be7A2E6BA89F429762e7b9aDb"

	// Invalid checksummed version of validChecksumAddr1
	invalidChecksumAddr1  = "0x5aAeb6053f3E94C9b9A09f33669435E7Ef1BeAed" // lowercase 'f' should be 'F'
)

func TestNormalizeAddress(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		want        string
		wantErr     bool
		errContains string
	}{
		// Valid cases - lowercase
		{
			name:    "valid lowercase address",
			address: validAddressLower,
			want:    validAddressLower,
			wantErr: false,
		},
		{
			name:    "valid uppercase address",
			address: validAddressUpper,
			want:    validAddressLower, // should return lowercase
			wantErr: false,
		},
		{
			name:    "valid address with whitespace",
			address: "  " + validAddressLower + "  ",
			want:    validAddressLower,
			wantErr: false,
		},

		// Valid EIP-55 checksummed addresses
		{
			name:    "valid EIP-55 checksum address 1",
			address: validChecksumAddr1,
			want:    strings.ToLower(validChecksumAddr1),
			wantErr: false,
		},
		{
			name:    "valid EIP-55 checksum address 2",
			address: validChecksumAddr2,
			want:    strings.ToLower(validChecksumAddr2),
			wantErr: false,
		},
		{
			name:    "valid EIP-55 checksum address 3",
			address: validChecksumAddr3,
			want:    strings.ToLower(validChecksumAddr3),
			wantErr: false,
		},
		{
			name:    "valid EIP-55 checksum address 4",
			address: validChecksumAddr4,
			want:    strings.ToLower(validChecksumAddr4),
			wantErr: false,
		},

		// Invalid checksum should fail
		{
			name:        "invalid EIP-55 checksum",
			address:     invalidChecksumAddr1,
			wantErr:     true,
			errContains: "invalid EIP-55 checksum",
		},

		// Invalid format cases
		{
			name:        "empty address",
			address:     "",
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:        "missing 0x prefix",
			address:     validAddressNoPrefix,
			wantErr:     true,
			errContains: "must be 0x followed by 40 hexadecimal characters",
		},
		{
			name:        "too short",
			address:     "0xabcdef",
			wantErr:     true,
			errContains: "must be 0x followed by 40 hexadecimal characters",
		},
		{
			name:        "too long",
			address:     "0xabcdef0123456789abcdef0123456789abcdef0123",
			wantErr:     true,
			errContains: "must be 0x followed by 40 hexadecimal characters",
		},
		{
			name:        "invalid characters",
			address:     "0xghijklmnopqrstuvwxyzghijklmnopqrstuvwxyz",
			wantErr:     true,
			errContains: "must be 0x followed by 40 hexadecimal characters",
		},
		{
			name:        "contains non-hex characters",
			address:     "0xabcdef0123456789abcdef0123456789abcdefgg",
			wantErr:     true,
			errContains: "must be 0x followed by 40 hexadecimal characters",
		},
		{
			name:        "only 0x prefix",
			address:     "0x",
			wantErr:     true,
			errContains: "must be 0x followed by 40 hexadecimal characters",
		},
		{
			name:        "whitespace only",
			address:     "   ",
			wantErr:     true,
			errContains: "cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeAddress(tt.address)

			if tt.wantErr {
				require.Error(t, err, "expected error but got none")
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Empty(t, got, "result should be empty on error")

				// Verify it's an AddressError
				_, ok := err.(*AddressError)
				assert.True(t, ok, "error should be of type *AddressError")
			} else {
				require.NoError(t, err, "unexpected error")
				assert.Equal(t, tt.want, got)

				// Verify result is always lowercase
				assert.Equal(t, strings.ToLower(got), got, "result should be lowercase")

				// Verify result has 0x prefix
				assert.True(t, strings.HasPrefix(got, "0x"), "result should have 0x prefix")

				// Verify result is 42 characters (0x + 40 hex)
				assert.Equal(t, 42, len(got), "result should be 42 characters")
			}
		})
	}
}

func TestIsValidAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    bool
	}{
		// Valid cases
		{
			name:    "valid lowercase address",
			address: validAddressLower,
			want:    true,
		},
		{
			name:    "valid uppercase address",
			address: validAddressUpper,
			want:    true,
		},
		{
			name:    "valid uppercase X prefix",
			address: "0XABCDEF0123456789ABCDEF0123456789ABCDEF01",
			want:    true,
		},
		{
			name:    "valid EIP-55 address",
			address: validChecksumAddr1,
			want:    true,
		},
		{
			name:    "invalid checksum still passes format check",
			address: invalidChecksumAddr1,
			want:    true, // IsValidAddress only checks format, not checksum
		},

		// Invalid cases
		{
			name:    "empty address",
			address: "",
			want:    false,
		},
		{
			name:    "missing 0x prefix",
			address: validAddressNoPrefix,
			want:    false,
		},
		{
			name:    "too short",
			address: "0xabcdef",
			want:    false,
		},
		{
			name:    "too long",
			address: "0xabcdef0123456789abcdef0123456789abcdef0123",
			want:    false,
		},
		{
			name:    "invalid characters",
			address: "0xghijklmnopqrstuvwxyzghijklmnopqrstuvwxyz",
			want:    false,
		},
		{
			name:    "only 0x prefix",
			address: "0x",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidAddress(tt.address)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestChecksumAddress(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		want        string
		wantErr     bool
		errContains string
	}{
		// Valid cases
		{
			name:    "lowercase to checksum",
			address: strings.ToLower(validChecksumAddr1),
			want:    validChecksumAddr1,
			wantErr: false,
		},
		{
			name:    "uppercase to checksum",
			address: strings.ToUpper(validChecksumAddr1),
			want:    validChecksumAddr1,
			wantErr: false,
		},
		{
			name:    "already checksummed",
			address: validChecksumAddr1,
			want:    validChecksumAddr1,
			wantErr: false,
		},
		{
			name:    "checksum address 2",
			address: strings.ToLower(validChecksumAddr2),
			want:    validChecksumAddr2,
			wantErr: false,
		},
		{
			name:    "checksum address 3",
			address: strings.ToLower(validChecksumAddr3),
			want:    validChecksumAddr3,
			wantErr: false,
		},
		{
			name:    "checksum address 4",
			address: strings.ToLower(validChecksumAddr4),
			want:    validChecksumAddr4,
			wantErr: false,
		},

		// Invalid cases
		{
			name:        "empty address",
			address:     "",
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:        "missing 0x prefix",
			address:     validAddressNoPrefix,
			wantErr:     true,
			errContains: "must be 0x followed by 40 hexadecimal characters",
		},
		{
			name:        "too short",
			address:     "0xabcdef",
			wantErr:     true,
			errContains: "must be 0x followed by 40 hexadecimal characters",
		},
		{
			name:        "invalid characters",
			address:     "0xghijklmnopqrstuvwxyzghijklmnopqrstuvwxyz",
			wantErr:     true,
			errContains: "must be 0x followed by 40 hexadecimal characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ChecksumAddress(tt.address)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Empty(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)

				// Verify checksum format
				assert.True(t, strings.HasPrefix(got, "0x"), "should have 0x prefix")
				assert.Equal(t, 42, len(got), "should be 42 characters")

				// Verify it's a valid address
				assert.True(t, IsValidAddress(got), "checksummed address should be valid")
			}
		})
	}
}

func TestCompareAddresses(t *testing.T) {
	tests := []struct {
		name  string
		addr1 string
		addr2 string
		want  bool
	}{
		// Equal cases
		{
			name:  "same lowercase addresses",
			addr1: validAddressLower,
			addr2: validAddressLower,
			want:  true,
		},
		{
			name:  "lowercase vs uppercase",
			addr1: validAddressLower,
			addr2: validAddressUpper,
			want:  true,
		},
		{
			name:  "checksummed vs lowercase",
			addr1: validChecksumAddr1,
			addr2: strings.ToLower(validChecksumAddr1),
			want:  true,
		},
		{
			name:  "two different checksummed addresses equal when same",
			addr1: validChecksumAddr1,
			addr2: validChecksumAddr1,
			want:  true,
		},

		// Not equal cases
		{
			name:  "different addresses",
			addr1: validChecksumAddr1,
			addr2: validChecksumAddr2,
			want:  false,
		},
		{
			name:  "different addresses (lowercase)",
			addr1: validAddressLower,
			addr2: "0x1234567890123456789012345678901234567890",
			want:  false,
		},

		// Invalid address cases
		{
			name:  "first address invalid",
			addr1: "invalid",
			addr2: validAddressLower,
			want:  false,
		},
		{
			name:  "second address invalid",
			addr1: validAddressLower,
			addr2: "invalid",
			want:  false,
		},
		{
			name:  "both addresses invalid",
			addr1: "invalid1",
			addr2: "invalid2",
			want:  false,
		},
		{
			name:  "empty addresses",
			addr1: "",
			addr2: "",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareAddresses(tt.addr1, tt.addr2)
			assert.Equal(t, tt.want, got)

			// Verify symmetry: CompareAddresses(a, b) == CompareAddresses(b, a)
			gotReverse := CompareAddresses(tt.addr2, tt.addr1)
			assert.Equal(t, got, gotReverse, "comparison should be symmetric")
		})
	}
}

func TestAddressError(t *testing.T) {
	err := &AddressError{
		Address: "0xinvalid",
		Reason:  "test reason",
	}

	errorMsg := err.Error()
	assert.Contains(t, errorMsg, "0xinvalid")
	assert.Contains(t, errorMsg, "test reason")
	assert.Contains(t, errorMsg, "invalid address")
}

func TestNormalizeAddressCaseSensitivity(t *testing.T) {
	// Test that all-lowercase and all-uppercase addresses pass without checksum validation
	allLower := "0xabcdef0123456789abcdef0123456789abcdef01"
	allUpper := "0xABCDEF0123456789ABCDEF0123456789ABCDEF01"

	// Both should succeed (no checksum validation for single-case)
	normalized1, err1 := NormalizeAddress(allLower)
	require.NoError(t, err1)
	assert.Equal(t, allLower, normalized1)

	normalized2, err2 := NormalizeAddress(allUpper)
	require.NoError(t, err2)
	assert.Equal(t, strings.ToLower(allUpper), normalized2)

	// They should be equal when normalized
	assert.Equal(t, normalized1, normalized2)
}

func TestNormalizeAddressEIP55Validation(t *testing.T) {
	// Test that mixed-case addresses trigger EIP-55 validation

	// This is a valid EIP-55 address
	validMixed := "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed"
	_, err := NormalizeAddress(validMixed)
	assert.NoError(t, err, "valid EIP-55 address should pass")

	// This has mixed case but invalid checksum
	invalidMixed := "0x5aAeb6053f3E94C9b9A09f33669435E7Ef1BeAed" // 'f' should be 'F'
	_, err = NormalizeAddress(invalidMixed)
	assert.Error(t, err, "invalid EIP-55 checksum should fail")
	assert.Contains(t, err.Error(), "invalid EIP-55 checksum")
}

func BenchmarkNormalizeAddress(b *testing.B) {
	addr := validAddressLower
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NormalizeAddress(addr)
	}
}

func BenchmarkIsValidAddress(b *testing.B) {
	addr := validAddressLower
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsValidAddress(addr)
	}
}

func BenchmarkChecksumAddress(b *testing.B) {
	addr := validAddressLower
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ChecksumAddress(addr)
	}
}

func BenchmarkCompareAddresses(b *testing.B) {
	addr1 := validAddressLower
	addr2 := validAddressUpper
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CompareAddresses(addr1, addr2)
	}
}
