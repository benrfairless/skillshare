package utils

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// ParseSkillName reads the SKILL.md and extracts the "name" from frontmatter.
func ParseSkillName(skillPath string) (string, error) {
	skillFile := filepath.Join(skillPath, "SKILL.md")
	file, err := os.Open(skillFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inFrontmatter := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Detect frontmatter delimiters
		if line == "---" {
			if inFrontmatter {
				break // End of frontmatter
			}
			inFrontmatter = true
			continue
		}

		if inFrontmatter {
			if strings.HasPrefix(line, "name:") {
				// Extract value: "name: my-skill" -> "my-skill"
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					name := strings.TrimSpace(parts[1])
					// Remove quotes if present
					name = strings.Trim(name, `"'`)
					return name, nil
				}
			}
		}
	}

	return "", nil // Name not found
}

// isYAMLBlockIndicator returns true for YAML block scalar indicators (>, >-, >+, |, |-, |+).
func isYAMLBlockIndicator(s string) bool {
	switch s {
	case ">", ">-", ">+", "|", "|-", "|+":
		return true
	}
	return false
}

// ParseFrontmatterField reads a SKILL.md file and extracts the value of a given frontmatter field.
// It supports both inline values and YAML block scalars (>, >-, |, |-).
func ParseFrontmatterField(filePath, field string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inFrontmatter := false
	prefix := field + ":"

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "---" {
			if inFrontmatter {
				break
			}
			inFrontmatter = true
			continue
		}

		if inFrontmatter && strings.HasPrefix(line, prefix) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				// Handle YAML block scalar indicators — read indented continuation lines
				if isYAMLBlockIndicator(val) {
					var blockParts []string
					for scanner.Scan() {
						next := scanner.Text()
						trimmed := strings.TrimSpace(next)
						if trimmed == "---" {
							break
						}
						// Block continues while lines are indented
						if len(next) > 0 && (next[0] == ' ' || next[0] == '\t') {
							blockParts = append(blockParts, trimmed)
						} else {
							break
						}
					}
					return strings.Join(blockParts, " ")
				}
				val = strings.Trim(val, `"'`)
				return val
			}
		}
	}

	return ""
}
