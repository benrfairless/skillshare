package utils

import (
	"runtime"
	"strings"
)

// PathsEqual compares two paths for equality.
// On Windows, paths are compared case-insensitively since NTFS is case-insensitive.
// On Unix systems, paths are compared exactly.
func PathsEqual(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// PathHasPrefix checks if path starts with prefix.
// On Windows, comparison is case-insensitive.
// On Unix systems, comparison is exact.
func PathHasPrefix(path, prefix string) bool {
	if runtime.GOOS == "windows" {
		return strings.HasPrefix(strings.ToLower(path), strings.ToLower(prefix))
	}
	return strings.HasPrefix(path, prefix)
}
