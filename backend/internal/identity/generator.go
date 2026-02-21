package identity

import (
	"regexp"
	"strings"
)

var validAddressRegex = regexp.MustCompile(`^[a-f0-9]{40}$`)

// IsValidAddress validates that an address is a valid 40-character hex string
func IsValidAddress(addr string) bool {
	return validAddressRegex.MatchString(strings.ToLower(addr))
}
