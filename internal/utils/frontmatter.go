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
