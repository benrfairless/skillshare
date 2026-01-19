package install

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// UpdateGitIgnore adds an entry to the .gitignore file in the given directory.
// If the entry already exists, it does nothing.
// Creates the .gitignore file if it doesn't exist.
func UpdateGitIgnore(dir, entry string) error {
	gitignorePath := filepath.Join(dir, ".gitignore")

	// Ensure entry ends with / for directory
	if !strings.HasSuffix(entry, "/") {
		entry = entry + "/"
	}

	// Check if entry already exists
	exists, err := gitignoreContains(gitignorePath, entry)
	if err != nil {
		return err
	}
	if exists {
		return nil // Already ignored
	}

	// Append entry to .gitignore
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open .gitignore: %w", err)
	}
	defer f.Close()

	// Check if file needs newline before entry
	needsNewline := false
	if info, err := f.Stat(); err == nil && info.Size() > 0 {
		// Read last byte to check if it's a newline
		content, err := os.ReadFile(gitignorePath)
		if err == nil && len(content) > 0 && content[len(content)-1] != '\n' {
			needsNewline = true
		}
	}

	var writeErr error
	if needsNewline {
		_, writeErr = f.WriteString("\n" + entry + "\n")
	} else {
		_, writeErr = f.WriteString(entry + "\n")
	}

	if writeErr != nil {
		return fmt.Errorf("failed to write to .gitignore: %w", writeErr)
	}

	return nil
}

// RemoveFromGitIgnore removes an entry from the .gitignore file.
// Returns true if the entry was found and removed.
func RemoveFromGitIgnore(dir, entry string) (bool, error) {
	gitignorePath := filepath.Join(dir, ".gitignore")

	// Ensure entry ends with / for directory
	if !strings.HasSuffix(entry, "/") {
		entry = entry + "/"
	}

	// Read existing content
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // No .gitignore, nothing to remove
		}
		return false, fmt.Errorf("failed to read .gitignore: %w", err)
	}

	// Find and remove the entry
	lines := strings.Split(string(content), "\n")
	var newLines []string
	found := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == entry || trimmed == strings.TrimSuffix(entry, "/") {
			found = true
			continue // Skip this line
		}
		newLines = append(newLines, line)
	}

	if !found {
		return false, nil
	}

	// Write back
	newContent := strings.Join(newLines, "\n")
	// Ensure file ends with newline if it has content
	if len(newContent) > 0 && !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}

	if err := os.WriteFile(gitignorePath, []byte(newContent), 0644); err != nil {
		return false, fmt.Errorf("failed to write .gitignore: %w", err)
	}

	return true, nil
}

// gitignoreContains checks if an entry exists in .gitignore
func gitignoreContains(path, entry string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer f.Close()

	// Also check without trailing slash
	entryNoSlash := strings.TrimSuffix(entry, "/")

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == entry || line == entryNoSlash {
			return true, nil
		}
	}

	return false, scanner.Err()
}
