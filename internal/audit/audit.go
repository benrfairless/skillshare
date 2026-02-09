package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"skillshare/internal/utils"
)

// localhostURL matches URLs pointing to localhost/127.0.0.1 (local API examples).
var localhostURL = regexp.MustCompile(`(?i)https?://(localhost|127\.0\.0\.1)`)

// Finding represents a single security issue detected in a skill.
type Finding struct {
	Severity string `json:"severity"` // "CRITICAL", "HIGH", "MEDIUM"
	Pattern  string `json:"pattern"`  // rule name (e.g. "prompt-injection")
	Message  string `json:"message"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Snippet  string `json:"snippet"` // max 80 chars of the matched line
}

// Result holds all findings for a single skill.
type Result struct {
	SkillName string    `json:"skillName"`
	Findings  []Finding `json:"findings"`
}

// HasCritical returns true if any finding is CRITICAL severity.
func (r *Result) HasCritical() bool {
	for _, f := range r.Findings {
		if f.Severity == SeverityCritical {
			return true
		}
	}
	return false
}

// HasHigh returns true if any finding is HIGH severity.
func (r *Result) HasHigh() bool {
	for _, f := range r.Findings {
		if f.Severity == SeverityHigh {
			return true
		}
	}
	return false
}

// MaxSeverity returns the highest severity found, or "" if no findings.
func (r *Result) MaxSeverity() string {
	max := ""
	for _, f := range r.Findings {
		switch f.Severity {
		case SeverityCritical:
			return SeverityCritical
		case SeverityHigh:
			max = SeverityHigh
		case SeverityMedium:
			if max == "" {
				max = SeverityMedium
			}
		}
	}
	return max
}

// CountBySeverity returns the count of findings at each severity level.
func (r *Result) CountBySeverity() (critical, high, medium int) {
	for _, f := range r.Findings {
		switch f.Severity {
		case SeverityCritical:
			critical++
		case SeverityHigh:
			high++
		case SeverityMedium:
			medium++
		}
	}
	return
}

// ScanSkill scans all scannable files in a skill directory.
func ScanSkill(skillPath string) (*Result, error) {
	info, err := os.Stat(skillPath)
	if err != nil {
		return nil, fmt.Errorf("cannot access skill path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", skillPath)
	}

	result := &Result{
		SkillName: filepath.Base(skillPath),
	}

	// Walk the skill directory and scan text files
	err = filepath.Walk(skillPath, func(path string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if fi.IsDir() {
			if utils.IsHidden(fi.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		// Only scan text-like files
		if !isScannable(fi.Name()) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(skillPath, path)
		findings := ScanContent(data, relPath)
		result.Findings = append(result.Findings, findings...)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error scanning skill: %w", err)
	}

	return result, nil
}

// ScanContent scans raw content for security issues and returns findings.
// filename is used for reporting (e.g. "SKILL.md").
func ScanContent(content []byte, filename string) []Finding {
	var findings []Finding
	lines := strings.Split(string(content), "\n")

	for lineNum, line := range lines {
		for _, r := range rules {
			if r.Regex.MatchString(line) {
				// Skip localhost URLs for suspicious-fetch (local API examples)
				if r.Pattern == "suspicious-fetch" && localhostURL.MatchString(line) {
					continue
				}
				findings = append(findings, Finding{
					Severity: r.Severity,
					Pattern:  r.Pattern,
					Message:  r.Message,
					File:     filename,
					Line:     lineNum + 1, // 1-indexed
					Snippet:  truncate(strings.TrimSpace(line), 80),
				})
			}
		}
	}

	return findings
}

// isScannable returns true if the file should be scanned.
func isScannable(name string) bool {
	// Skip skillshare's own metadata files
	if name == ".skillshare-meta.json" {
		return false
	}

	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".md", ".txt", ".yaml", ".yml", ".json", ".toml",
		".sh", ".bash", ".zsh", ".fish",
		".py", ".js", ".ts", ".rb", ".go", ".rs":
		return true
	}
	// Also scan files without extension (e.g. Makefile, Dockerfile)
	if ext == "" {
		return true
	}
	return false
}

// truncate shortens s to maxLen characters, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
