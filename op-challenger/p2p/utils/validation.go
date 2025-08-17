package utils

import (
	"fmt"
	"math/rand"
	"net"
	"regexp"
	"strings"
	"time"
)

// ValidateNodeID validates if a node ID is valid
func ValidateNodeID(nodeID string) error {
	if nodeID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	// 길이 체크 (32자 16진수)
	if len(nodeID) != 32 {
		return fmt.Errorf("node ID must be 32 characters long, got %d", len(nodeID))
	}

	// 16진수 형식 체크
	matched, err := regexp.MatchString("^[a-fA-F0-9]+$", nodeID)
	if err != nil {
		return fmt.Errorf("failed to validate node ID format: %w", err)
	}
	if !matched {
		return fmt.Errorf("node ID must be hexadecimal")
	}

	return nil
}

// ValidateChallengerID validates if a challenger ID is valid
func ValidateChallengerID(challengerID string) error {
	if challengerID == "" {
		return fmt.Errorf("challenger ID cannot be empty")
	}

	// 길이 체크 (40자 16진수)
	if len(challengerID) != 40 {
		return fmt.Errorf("challenger ID must be 40 characters long, got %d", len(challengerID))
	}

	// 16진수 형식 체크
	matched, err := regexp.MatchString("^[a-fA-F0-9]+$", challengerID)
	if err != nil {
		return fmt.Errorf("failed to validate challenger ID format: %w", err)
	}
	if !matched {
		return fmt.Errorf("challenger ID must be hexadecimal")
	}

	return nil
}

// ValidateNetworkAddress validates if a network address is valid
func ValidateNetworkAddress(address string) error {
	if address == "" {
		return fmt.Errorf("address cannot be empty")
	}

	// IP:Port 형식인지 확인
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("invalid address format: %w", err)
	}

	// IP 주소 유효성 확인
	if net.ParseIP(host) == nil {
		// IP가 아니면 호스트명인지 확인
		if !isValidHostname(host) {
			return fmt.Errorf("invalid IP address or hostname: %s", host)
		}
	}

	// 포트 번호 확인
	if port == "" {
		return fmt.Errorf("port number is required")
	}

	return nil
}

// isValidHostname validates if a hostname is valid
func isValidHostname(hostname string) bool {
	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}

	// 정규식으로 호스트명 형식 확인
	matched, err := regexp.MatchString(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`, hostname)
	if err != nil {
		return false
	}

	return matched
}

// ValidateVersion validates if a version string is valid
func ValidateVersion(version string) error {
	if version == "" {
		return fmt.Errorf("version cannot be empty")
	}

	// 간단한 버전 형식 체크 (x.y.z)
	matched, err := regexp.MatchString(`^\d+\.\d+\.\d+$`, version)
	if err != nil {
		return fmt.Errorf("failed to validate version format: %w", err)
	}
	if !matched {
		return fmt.Errorf("version must be in format x.y.z")
	}

	return nil
}

// ValidateTimeout validates if a timeout duration is reasonable
func ValidateTimeout(timeout time.Duration, minTimeout, maxTimeout time.Duration) error {
	if timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if timeout < minTimeout {
		return fmt.Errorf("timeout %v is too short, minimum is %v", timeout, minTimeout)
	}

	if timeout > maxTimeout {
		return fmt.Errorf("timeout %v is too long, maximum is %v", timeout, maxTimeout)
	}

	return nil
}

// ValidateInterval validates if an interval duration is reasonable
func ValidateInterval(interval time.Duration, minInterval, maxInterval time.Duration) error {
	if interval <= 0 {
		return fmt.Errorf("interval must be positive")
	}

	if interval < minInterval {
		return fmt.Errorf("interval %v is too short, minimum is %v", interval, minInterval)
	}

	if interval > maxInterval {
		return fmt.Errorf("interval %v is too long, maximum is %v", interval, maxInterval)
	}

	return nil
}

// ValidatePositiveInt validates if an integer is positive
func ValidatePositiveInt(value int, name string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive, got %d", name, value)
	}
	return nil
}

// ValidatePositiveUint64 validates if a uint64 is positive
func ValidatePositiveUint64(value uint64, name string) error {
	if value == 0 {
		return fmt.Errorf("%s must be positive, got %d", name, value)
	}
	return nil
}

// ValidateRange validates if a value is within a specified range
func ValidateRange(value, min, max int, name string) error {
	if value < min {
		return fmt.Errorf("%s %d is below minimum %d", name, value, min)
	}
	if value > max {
		return fmt.Errorf("%s %d is above maximum %d", name, value, max)
	}
	return nil
}

// ValidateFloatRange validates if a float value is within a specified range
func ValidateFloatRange(value, min, max float64, name string) error {
	if value < min {
		return fmt.Errorf("%s %.2f is below minimum %.2f", name, value, min)
	}
	if value > max {
		return fmt.Errorf("%s %.2f is above maximum %.2f", name, value, max)
	}
	return nil
}

// ValidateNonEmpty validates if a string is not empty
func ValidateNonEmpty(value, name string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s cannot be empty", name)
	}
	return nil
}

// ValidateStringLength validates if a string length is within limits
func ValidateStringLength(value string, minLen, maxLen int, name string) error {
	length := len(value)
	if length < minLen {
		return fmt.Errorf("%s length %d is below minimum %d", name, length, minLen)
	}
	if length > maxLen {
		return fmt.Errorf("%s length %d is above maximum %d", name, length, maxLen)
	}
	return nil
}

// ValidateSliceLength validates if a slice length is within limits
func ValidateSliceLength(slice interface{}, minLen, maxLen int, name string) error {
	var length int

	switch s := slice.(type) {
	case []string:
		length = len(s)
	case []int:
		length = len(s)
	case []interface{}:
		length = len(s)
	default:
		return fmt.Errorf("unsupported slice type for %s", name)
	}

	if length < minLen {
		return fmt.Errorf("%s length %d is below minimum %d", name, length, minLen)
	}
	if length > maxLen {
		return fmt.Errorf("%s length %d is above maximum %d", name, length, maxLen)
	}
	return nil
}

// SanitizeString removes potentially dangerous characters from a string
func SanitizeString(input string) string {
	// 기본적인 문자만 허용 (영문, 숫자, 하이픈, 언더스코어, 점)
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-_\.]`)
	return reg.ReplaceAllString(input, "")
}

// IsValidHexString checks if a string is a valid hexadecimal string
func IsValidHexString(s string) bool {
	matched, err := regexp.MatchString("^[a-fA-F0-9]+$", s)
	return err == nil && matched
}

// GenerateRequestID generates a unique request ID
func GenerateRequestID() string {
	return fmt.Sprintf("req_%d_%s", time.Now().UnixNano(), generateRandomString(8))
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// IsValidAddress checks if a string is a valid network address (IP:Port)
func IsValidAddress(addr string) bool {
	_, _, err := net.SplitHostPort(addr)
	return err == nil
}
