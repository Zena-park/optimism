package utils

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestValidateNodeID tests node ID validation
func TestValidateNodeID(t *testing.T) {
	// Test valid node IDs (32 character hex strings)
	validIDs := []string{
		"0123456789abcdef0123456789abcdef", // 32 chars lowercase
		"ABCDEFABCDEFABCDEFABCDEFABCDEFAB", // 32 chars uppercase
		"00000000000000000000000000000000", // All zeros
		"ffffffffffffffffffffffffffffffff", // All f's lowercase
		"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", // All f's uppercase
		"12345678901234567890123456789012", // Numbers only
		"aAbBcCdDeEfF12345678901234567890", // Mixed case
	}

	for _, id := range validIDs {
		err := ValidateNodeID(id)
		assert.NoError(t, err, "Node ID should be valid: %s", id)
	}

	// Test invalid node IDs
	invalidIDs := []string{
		"",                                  // Empty
		"123",                               // Too short
		"0123456789abcdef0123456789abcdef0", // Too long
		"0123456789abcdef0123456789abcdeg",  // Contains invalid char 'g'
		"0123456789abcdef0123456789abcde@",  // Contains invalid char '@'
		"0123456789abcdef0123456789abcde ",  // Contains space
		"0123456789abcdef0123456789abcde\n", // Contains newline
		"node-id-not-hex-format-string",     // Not hex format
		"0123456789abcdef0123456789abcde",   // 31 chars (too short)
		"0123456789abcdef0123456789abcdef0", // 33 chars (too long)
	}

	for _, id := range invalidIDs {
		err := ValidateNodeID(id)
		assert.Error(t, err, "Node ID should be invalid: %s", id)
	}
}

// TestValidateChallengerID tests challenger ID validation
func TestValidateChallengerID(t *testing.T) {
	// Test valid challenger IDs (40 character hex strings)
	validIDs := []string{
		"0123456789abcdef0123456789abcdef01234567", // 40 chars lowercase
		"ABCDEFABCDEFABCDEFABCDEFABCDEFABCDEFABCD", // 40 chars uppercase
		"0000000000000000000000000000000000000000", // All zeros
		"ffffffffffffffffffffffffffffffffffffffff", // All f's lowercase
		"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", // All f's uppercase
		"1234567890123456789012345678901234567890", // Numbers only
		"aAbBcCdDeEfF1234567890123456789012345678", // Mixed case
	}

	for _, id := range validIDs {
		err := ValidateChallengerID(id)
		assert.NoError(t, err, "Challenger ID should be valid: %s", id)
	}

	// Test invalid challenger IDs
	invalidIDs := []string{
		"",    // Empty
		"123", // Too short
		"0123456789abcdef0123456789abcdef012345678", // Too long
		"0123456789abcdef0123456789abcdef0123456",   // Too short by 1
		"0123456789abcdef0123456789abcdef012345678", // Too long by 1
		"0123456789abcdef0123456789abcdef0123456g",  // Contains invalid char 'g'
		"0123456789abcdef0123456789abcdef0123456@",  // Contains invalid char '@'
		"0123456789abcdef0123456789abcdef0123456 ",  // Contains space
		"0123456789abcdef0123456789abcdef0123456\n", // Contains newline
		"challenger-id-not-hex-format-string",       // Not hex format
	}

	for _, id := range invalidIDs {
		err := ValidateChallengerID(id)
		assert.Error(t, err, "Challenger ID should be invalid: %s", id)
	}
}

// TestValidateNetworkAddress tests network address validation
func TestValidateNetworkAddress(t *testing.T) {
	// Test valid network addresses
	validAddresses := []string{
		"127.0.0.1:8080",
		"192.168.1.1:9999",
		"10.0.0.1:80",
		"172.16.0.1:443",
		"0.0.0.0:8080",
		"255.255.255.255:65535",
		"localhost:8080",
		"example.com:80",
		"test-server.local:12345",
		"[::1]:8080",         // IPv6 loopback
		"[2001:db8::1]:8080", // IPv6
		"[::]:8080",          // IPv6 any
	}

	for _, addr := range validAddresses {
		err := ValidateNetworkAddress(addr)
		assert.NoError(t, err, "Network address should be valid: %s", addr)
	}

	// Test invalid network addresses
	invalidAddresses := []string{
		"",                      // Empty
		"127.0.0.1",             // Missing port
		":8080",                 // Missing host
		"127.0.0.1:",            // Missing port number
		"127.0.0.1 :8080",       // Space in address
		" 127.0.0.1:8080",       // Leading space
		"127.0.0.1:8080:extra",  // Extra colon
		"not-an-address",        // Not an address format
		"http://127.0.0.1:8080", // URL format
	}

	for _, addr := range invalidAddresses {
		err := ValidateNetworkAddress(addr)
		assert.Error(t, err, "Network address should be invalid: %s", addr)
	}
}

// TestValidateVersion tests version string validation
func TestValidateVersion(t *testing.T) {
	// Test valid versions (x.y.z format)
	validVersions := []string{
		"1.0.0",
		"2.1.3",
		"0.0.1",
		"10.20.30",
		"999.999.999",
	}

	for _, version := range validVersions {
		err := ValidateVersion(version)
		assert.NoError(t, err, "Version should be valid: %s", version)
	}

	// Test invalid versions
	invalidVersions := []string{
		"",            // Empty
		"1",           // Too short
		"1.0",         // Missing patch
		"1.0.0.0",     // Too many parts
		"a.b.c",       // Non-numeric
		"1.0.-1",      // Negative number
		"1.0.0-alpha", // With suffix
		"v1.0.0",      // With prefix
		"1.0.0 ",      // Trailing space
		" 1.0.0",      // Leading space
		"1.0.0@beta",  // Invalid character
	}

	for _, version := range invalidVersions {
		err := ValidateVersion(version)
		assert.Error(t, err, "Version should be invalid: %s", version)
	}
}

// TestValidateTimeout tests timeout validation
func TestValidateTimeout(t *testing.T) {
	minTimeout := 1 * time.Second
	maxTimeout := 1 * time.Minute

	// Test valid timeouts
	validTimeouts := []time.Duration{
		1 * time.Second,
		5 * time.Second,
		30 * time.Second,
		1 * time.Minute,
	}

	for _, timeout := range validTimeouts {
		err := ValidateTimeout(timeout, minTimeout, maxTimeout)
		assert.NoError(t, err, "Timeout should be valid: %v", timeout)
	}

	// Test invalid timeouts
	invalidTimeouts := []time.Duration{
		0,                      // Zero
		-1 * time.Second,       // Negative
		500 * time.Millisecond, // Too short
		2 * time.Minute,        // Too long
	}

	for _, timeout := range invalidTimeouts {
		err := ValidateTimeout(timeout, minTimeout, maxTimeout)
		assert.Error(t, err, "Timeout should be invalid: %v", timeout)
	}
}

// TestValidateInterval tests interval validation
func TestValidateInterval(t *testing.T) {
	minInterval := 100 * time.Millisecond
	maxInterval := 10 * time.Second

	// Test valid intervals
	validIntervals := []time.Duration{
		100 * time.Millisecond,
		1 * time.Second,
		5 * time.Second,
		10 * time.Second,
	}

	for _, interval := range validIntervals {
		err := ValidateInterval(interval, minInterval, maxInterval)
		assert.NoError(t, err, "Interval should be valid: %v", interval)
	}

	// Test invalid intervals
	invalidIntervals := []time.Duration{
		0,                     // Zero
		-1 * time.Millisecond, // Negative
		50 * time.Millisecond, // Too short
		20 * time.Second,      // Too long
	}

	for _, interval := range invalidIntervals {
		err := ValidateInterval(interval, minInterval, maxInterval)
		assert.Error(t, err, "Interval should be invalid: %v", interval)
	}
}

// TestValidatePositiveInt tests positive integer validation
func TestValidatePositiveInt(t *testing.T) {
	// Test valid positive integers
	validInts := []int{1, 5, 100, 1000, 999999}
	for _, val := range validInts {
		err := ValidatePositiveInt(val, "test_value")
		assert.NoError(t, err, "Positive int should be valid: %d", val)
	}

	// Test invalid integers
	invalidInts := []int{0, -1, -100, -999}
	for _, val := range invalidInts {
		err := ValidatePositiveInt(val, "test_value")
		assert.Error(t, err, "Non-positive int should be invalid: %d", val)
	}
}

// TestValidatePositiveUint64 tests positive uint64 validation
func TestValidatePositiveUint64(t *testing.T) {
	// Test valid positive uint64s
	validUints := []uint64{1, 5, 100, 1000, 999999, ^uint64(0)} // Max uint64
	for _, val := range validUints {
		err := ValidatePositiveUint64(val, "test_value")
		assert.NoError(t, err, "Positive uint64 should be valid: %d", val)
	}

	// Test invalid uint64 (only zero is invalid)
	err := ValidatePositiveUint64(0, "test_value")
	assert.Error(t, err, "Zero uint64 should be invalid")
}

// TestValidateRange tests range validation
func TestValidateRange(t *testing.T) {
	// Test valid ranges
	testCases := []struct {
		value, min, max int
		shouldPass      bool
	}{
		{5, 1, 10, true},   // Within range
		{1, 1, 10, true},   // At minimum
		{10, 1, 10, true},  // At maximum
		{0, 0, 0, true},    // Single value range
		{-5, -10, 0, true}, // Negative range
		{0, 1, 10, false},  // Below minimum
		{11, 1, 10, false}, // Above maximum
		{-1, 0, 10, false}, // Below minimum
	}

	for _, tc := range testCases {
		err := ValidateRange(tc.value, tc.min, tc.max, "test_value")
		if tc.shouldPass {
			assert.NoError(t, err, "Value %d should be valid in range [%d, %d]", tc.value, tc.min, tc.max)
		} else {
			assert.Error(t, err, "Value %d should be invalid in range [%d, %d]", tc.value, tc.min, tc.max)
		}
	}
}

// TestValidateFloatRange tests float range validation
func TestValidateFloatRange(t *testing.T) {
	// Test valid float ranges
	testCases := []struct {
		value, min, max float64
		shouldPass      bool
	}{
		{5.5, 1.0, 10.0, true},   // Within range
		{1.0, 1.0, 10.0, true},   // At minimum
		{10.0, 1.0, 10.0, true},  // At maximum
		{0.0, 0.0, 0.0, true},    // Single value range
		{-5.5, -10.0, 0.0, true}, // Negative range
		{0.5, 1.0, 10.0, false},  // Below minimum
		{10.5, 1.0, 10.0, false}, // Above maximum
		{-0.1, 0.0, 10.0, false}, // Below minimum
	}

	for _, tc := range testCases {
		err := ValidateFloatRange(tc.value, tc.min, tc.max, "test_value")
		if tc.shouldPass {
			assert.NoError(t, err, "Value %.2f should be valid in range [%.2f, %.2f]", tc.value, tc.min, tc.max)
		} else {
			assert.Error(t, err, "Value %.2f should be invalid in range [%.2f, %.2f]", tc.value, tc.min, tc.max)
		}
	}
}

// TestValidateNonEmpty tests non-empty string validation
func TestValidateNonEmpty(t *testing.T) {
	// Test valid non-empty strings
	validStrings := []string{
		"hello",
		"test123",
		"   hello   ", // Trimmed to "hello"
		"a",
		"special@chars#allowed",
	}

	for _, str := range validStrings {
		err := ValidateNonEmpty(str, "test_string")
		assert.NoError(t, err, "Non-empty string should be valid: %s", str)
	}

	// Test invalid empty strings
	invalidStrings := []string{
		"",
		"   ",  // Only spaces
		"\t\n", // Only whitespace
	}

	for _, str := range invalidStrings {
		err := ValidateNonEmpty(str, "test_string")
		assert.Error(t, err, "Empty string should be invalid: %q", str)
	}
}

// TestValidateStringLength tests string length validation
func TestValidateStringLength(t *testing.T) {
	// Test valid string lengths
	testCases := []struct {
		value          string
		minLen, maxLen int
		shouldPass     bool
	}{
		{"hello", 1, 10, true},        // Within range
		{"a", 1, 10, true},            // At minimum
		{"1234567890", 1, 10, true},   // At maximum
		{"", 0, 5, true},              // Empty string with min=0
		{"", 1, 10, false},            // Empty string with min=1
		{"12345678901", 1, 10, false}, // Too long
		{"hello", 10, 20, false},      // Too short
	}

	for _, tc := range testCases {
		err := ValidateStringLength(tc.value, tc.minLen, tc.maxLen, "test_string")
		if tc.shouldPass {
			assert.NoError(t, err, "String %q should be valid with length range [%d, %d]", tc.value, tc.minLen, tc.maxLen)
		} else {
			assert.Error(t, err, "String %q should be invalid with length range [%d, %d]", tc.value, tc.minLen, tc.maxLen)
		}
	}
}

// TestValidateSliceLength tests slice length validation
func TestValidateSliceLength(t *testing.T) {
	// Test string slices
	stringSlice := []string{"a", "b", "c"}
	err := ValidateSliceLength(stringSlice, 1, 5, "string_slice")
	assert.NoError(t, err, "String slice should be valid")

	err = ValidateSliceLength(stringSlice, 5, 10, "string_slice")
	assert.Error(t, err, "String slice should be too short")

	// Test int slices
	intSlice := []int{1, 2, 3, 4, 5}
	err = ValidateSliceLength(intSlice, 1, 10, "int_slice")
	assert.NoError(t, err, "Int slice should be valid")

	err = ValidateSliceLength(intSlice, 1, 3, "int_slice")
	assert.Error(t, err, "Int slice should be too long")

	// Test interface slices
	interfaceSlice := []interface{}{1, "hello", 3.14}
	err = ValidateSliceLength(interfaceSlice, 1, 5, "interface_slice")
	assert.NoError(t, err, "Interface slice should be valid")

	// Test unsupported type
	err = ValidateSliceLength("not a slice", 1, 5, "invalid_type")
	assert.Error(t, err, "Non-slice type should be invalid")
}

// TestSanitizeString tests string sanitization
func TestSanitizeString(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},                           // Normal string
		{"hello123", "hello123"},                     // Alphanumeric
		{"hello@world#test", "helloworldtest"},       // Remove special chars
		{"test-name_file.txt", "test-name_file.txt"}, // Keep allowed chars
		{"hello world", "helloworld"},                // Remove spaces
		{"", ""},                                     // Empty string
		{"   ", ""},                                  // Whitespace only
		{"test$%^&*()+=[]{}|\\:;\"'<>?,/", "test"},   // Remove many special chars
	}

	for _, tc := range testCases {
		result := SanitizeString(tc.input)
		assert.Equal(t, tc.expected, result, "Sanitization failed for input: %q", tc.input)
	}
}

// TestIsValidHexString tests hex string validation
func TestIsValidHexString(t *testing.T) {
	// Test valid hex strings
	validHex := []string{
		"0123456789abcdef",
		"ABCDEF0123456789",
		"0000000000000000",
		"ffffffffffffffff",
		"a",
		"A",
		"0",
		"deadbeef",
		"DEADBEEF",
		"123ABC",
	}

	for _, hex := range validHex {
		valid := IsValidHexString(hex)
		assert.True(t, valid, "Hex string should be valid: %s", hex)
	}

	// Test invalid hex strings
	invalidHex := []string{
		"",       // Empty
		"g",      // Invalid character
		"123g",   // Contains invalid character
		"12 34",  // Contains space
		"12\n34", // Contains newline
		"12@34",  // Contains special character
		"0x123",  // With prefix (not supported by this function)
	}

	for _, hex := range invalidHex {
		valid := IsValidHexString(hex)
		assert.False(t, valid, "Hex string should be invalid: %s", hex)
	}
}

// TestGenerateRequestID tests request ID generation
func TestGenerateRequestID(t *testing.T) {
	// Test basic generation
	id1 := GenerateRequestID()
	assert.NotEmpty(t, id1, "Generated request ID should not be empty")
	assert.Contains(t, id1, "req_", "Request ID should contain prefix")

	// Test uniqueness
	id2 := GenerateRequestID()
	assert.NotEqual(t, id1, id2, "Different calls should generate different IDs")

	// Test multiple generations for uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := GenerateRequestID()
		assert.NotEmpty(t, id, "Generated ID should not be empty")
		assert.False(t, ids[id], "Generated ID should be unique: %s", id)
		ids[id] = true
	}
}

// TestIsValidAddress tests address validation
func TestIsValidAddress(t *testing.T) {
	// Test valid addresses
	validAddresses := []string{
		"127.0.0.1:8080",
		"192.168.1.1:9999",
		"localhost:8080",
		"example.com:80",
		"[::1]:8080",
		"[2001:db8::1]:8080",
	}

	for _, addr := range validAddresses {
		valid := IsValidAddress(addr)
		assert.True(t, valid, "Address should be valid: %s", addr)
	}

	// Test invalid addresses
	invalidAddresses := []string{
		"",               // Empty
		"127.0.0.1",      // Missing port
		"not-an-address", // Invalid format
	}

	for _, addr := range invalidAddresses {
		valid := IsValidAddress(addr)
		assert.False(t, valid, "Address should be invalid: %s", addr)
	}
}

// TestValidationEdgeCases tests edge cases and boundary conditions
func TestValidationEdgeCases(t *testing.T) {
	// Test with very long strings
	longString := strings.Repeat("a", 10000)

	// Node ID validation with long string
	err := ValidateNodeID(longString)
	assert.Error(t, err, "Very long node ID should be invalid")

	// Challenger ID validation with long string
	err = ValidateChallengerID(longString)
	assert.Error(t, err, "Very long challenger ID should be invalid")

	// Network address validation with long hostname
	err = ValidateNetworkAddress(longString + ":8080")
	assert.Error(t, err, "Very long hostname should be invalid")

	// Test with unicode characters
	unicodeString := "测试节点"
	err = ValidateNodeID(unicodeString)
	assert.Error(t, err, "Unicode node ID should be invalid")

	// Test with control characters
	controlString := "test\x00node"
	err = ValidateNodeID(controlString)
	assert.Error(t, err, "Node ID with control chars should be invalid")

	// Test boundary values for ranges
	err = ValidateRange(0, 0, 0, "test")
	assert.NoError(t, err, "Zero range should be valid")

	// Test empty inputs for all validators
	err = ValidateNodeID("")
	assert.Error(t, err, "Empty node ID should be invalid")

	err = ValidateChallengerID("")
	assert.Error(t, err, "Empty challenger ID should be invalid")

	err = ValidateNetworkAddress("")
	assert.Error(t, err, "Empty network address should be invalid")

	err = ValidateVersion("")
	assert.Error(t, err, "Empty version should be invalid")

	valid := IsValidHexString("")
	assert.False(t, valid, "Empty hex string should be invalid")
}

// TestValidationConcurrency tests concurrent validation operations
func TestValidationConcurrency(t *testing.T) {
	const numGoroutines = 10
	const numOperations = 100

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				// Test various validation functions concurrently
				ValidateNodeID("0123456789abcdef0123456789abcdef")
				ValidateChallengerID("0123456789abcdef0123456789abcdef01234567")
				ValidateNetworkAddress("127.0.0.1:8080")
				ValidateVersion("1.0.0")
				IsValidHexString("deadbeef")
				GenerateRequestID()
				IsValidAddress("127.0.0.1:8080")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// If we get here without issues, concurrent access is safe
	assert.True(t, true)
}
