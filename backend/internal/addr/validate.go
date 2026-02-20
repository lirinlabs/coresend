package addr

import (
	"regexp"
	"strings"
)

var validAddressRegex = regexp.MustCompile(`^[a-f0-9]{16}$`)

// IsValid reports whether addr is a valid 16-character lowercase hex inbox address.
func IsValid(addr string) bool {
	return validAddressRegex.MatchString(strings.ToLower(addr))
}
