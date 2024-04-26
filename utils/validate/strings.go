package validate

import (
	"strings"
)

// IsBlank checks if a string is empty or contains only whitespace
func IsBlank(s string) bool {
	if len(strings.TrimSpace(s)) == 0 {
		return true
	}
	return false
}

// IsNotBlank checks if a string is not empty and does not contain only whitespace
func IsNotBlank(s string) bool {
	return !IsBlank(s)
}
