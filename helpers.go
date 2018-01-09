package kvstructure

import (
	"strings"
)

// leadingSlash is adding a slash to the beginning
func leadingSlash(s string) string {
	prefix := "/"
	if !strings.HasPrefix(s, prefix) {
		return prefix + s
	}
	return s
}

// trailingSlash is adding a slash at the end
func trailingSlash(s string) string {
	suffix := "/"
	if !strings.HasSuffix(s, suffix) {
		return s + suffix
	}
	return s
}
