package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"skillshare/internal/config"
	"skillshare/internal/sync"
	"skillshare/internal/ui"
	"skillshare/internal/utils"
)

func cmdDoctor(args []string) error {
	ui.Header("Checking environment")
	issues := 0

	// Check config exists
	if _, err := os.Stat(config.ConfigPath()); os.IsNotExist(err) {
		ui.Error("Config not found: run 'skillshare init' first")
		return nil
	}
	ui.Success("Config: %s", config.ConfigPath())

	cfg, err := config.Load()
	if err != nil {
		ui.Error("Config error: %v", err)
		return nil
	}

	// Check source exists
	issues += checkSource(cfg)

	// Check symlink support
	issues += checkSymlinkSupport()

	// Check git status
	issues += checkGitStatus(cfg.Source)

	// Check skills validity
	issues += checkSkillsValidity(cfg.Source)

	// Check each target
	issues += checkTargets(cfg)

	// Check broken symlinks
	issues += checkBrokenSymlinks(cfg)

	// Check duplicate skills
	issues += checkDuplicateSkills(cfg)

	// Check backup status
	checkBackupStatus()

	// Check for updates
	checkForUpdates()

	// Summary
	ui.Header("Summary")
	if issues == 0 {
		ui.Success("All checks passed!")
	} else {
		ui.Warning("%d issue(s) found", issues)
	}

	return nil
}

func checkSource(cfg *config.Config) int {
	info, err := os.Stat(cfg.Source)
	if err != nil {
		ui.Error("Source not found: %s", cfg.Source)
		return 1
	}

	if !info.IsDir() {
		ui.Error("Source is not a directory: %s", cfg.Source)
		return 1
	}

	entries, _ := os.ReadDir(cfg.Source)
	skillCount := 0
	for _, e := range entries {
		if e.IsDir() && !utils.IsHidden(e.Name()) {
			skillCount++
		}
	}
	ui.Success("Source: %s (%d skills)", cfg.Source, skillCount)
	return 0
}

func checkSymlinkSupport() int {
	testSymlink := filepath.Join(os.TempDir(), "skillshare_symlink_test")
	testTarget := filepath.Join(os.TempDir(), "skillshare_symlink_target")
	os.Remove(testSymlink)
	os.Remove(testTarget)
	os.MkdirAll(testTarget, 0755)
	defer os.Remove(testSymlink)
	defer os.RemoveAll(testTarget)

	if err := os.Symlink(testTarget, testSymlink); err != nil {
		ui.Error("Symlink not supported: %v", err)
		return 1
	}

	ui.Success("Symlink support: OK")
	return 0
}

func checkTargets(cfg *config.Config) int {
	ui.Header("Checking targets")
	issues := 0

	for name, target := range cfg.Targets {
		// Determine mode
		mode := target.Mode
		if mode == "" {
			mode = cfg.Mode
		}
		if mode == "" {
			mode = "merge"
		}

		targetIssues := checkTargetIssues(target, cfg.Source)

		if len(targetIssues) > 0 {
			ui.Error("%s [%s]: %s", name, mode, strings.Join(targetIssues, ", "))
			issues++
		} else {
			displayTargetStatus(name, target, cfg.Source, mode)
		}
	}

	return issues
}

func checkTargetIssues(target config.TargetConfig, source string) []string {
	var targetIssues []string

	info, err := os.Lstat(target.Path)
	if err != nil {
		if os.IsNotExist(err) {
			// Check parent is writable
			parent := filepath.Dir(target.Path)
			if _, err := os.Stat(parent); err != nil {
				targetIssues = append(targetIssues, "parent directory not found")
			}
		} else {
			targetIssues = append(targetIssues, fmt.Sprintf("access error: %v", err))
		}
		return targetIssues
	}

	// Check if it's a symlink
	if info.Mode()&os.ModeSymlink != 0 {
		link, _ := os.Readlink(target.Path)
		absLink, _ := filepath.Abs(link)
		absSource, _ := filepath.Abs(source)
		if absLink != absSource {
			targetIssues = append(targetIssues, fmt.Sprintf("symlink points to wrong location: %s", link))
		}
	}

	// Check write permission
	if info.IsDir() {
		testFile := filepath.Join(target.Path, ".skillshare_write_test")
		if f, err := os.Create(testFile); err != nil {
			targetIssues = append(targetIssues, "not writable")
		} else {
			f.Close()
			os.Remove(testFile)
		}
	}

	return targetIssues
}

func displayTargetStatus(name string, target config.TargetConfig, source, mode string) {
	var statusStr string
	needsSync := false

	if mode == "merge" {
		status, linkedCount, localCount := sync.CheckStatusMerge(target.Path, source)
		switch status {
		case sync.StatusMerged:
			statusStr = fmt.Sprintf("merged (%d shared, %d local)", linkedCount, localCount)
		case sync.StatusLinked:
			statusStr = "linked (needs sync to apply merge mode)"
			needsSync = true
		default:
			statusStr = status.String()
		}
	} else {
		status := sync.CheckStatus(target.Path, source)
		statusStr = status.String()
		if status == sync.StatusMerged {
			statusStr = "merged (needs sync to apply symlink mode)"
			needsSync = true
		}
	}

	if needsSync {
		ui.Warning("%s [%s]: %s", name, mode, statusStr)
	} else {
		ui.Success("%s [%s]: %s", name, mode, statusStr)
	}
}

// checkGitStatus checks if source is a git repo and its status
func checkGitStatus(source string) int {
	gitDir := filepath.Join(source, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		ui.Warning("Git: not initialized (recommended for backup)")
		return 0 // Not an error, just a warning
	}

	// Check for uncommitted changes
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = source
	output, err := cmd.Output()
	if err != nil {
		ui.Warning("Git: unable to check status")
		return 0
	}

	if len(output) > 0 {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		ui.Warning("Git: %d uncommitted change(s)", len(lines))
		return 0 // Warning, not error
	}

	// Check for remote
	cmd = exec.Command("git", "remote")
	cmd.Dir = source
	output, err = cmd.Output()
	if err == nil && len(strings.TrimSpace(string(output))) == 0 {
		ui.Success("Git: initialized (no remote configured)")
	} else {
		ui.Success("Git: initialized with remote")
	}

	return 0
}

// checkSkillsValidity checks if all skills have valid SKILL.md files
func checkSkillsValidity(source string) int {
	entries, err := os.ReadDir(source)
	if err != nil {
		return 0
	}

	var invalid []string
	for _, entry := range entries {
		if !entry.IsDir() || utils.IsHidden(entry.Name()) {
			continue
		}

		skillFile := filepath.Join(source, entry.Name(), "SKILL.md")
		if _, err := os.Stat(skillFile); os.IsNotExist(err) {
			invalid = append(invalid, entry.Name())
		}
	}

	if len(invalid) > 0 {
		ui.Warning("Skills without SKILL.md: %s", strings.Join(invalid, ", "))
		return 0 // Warning, not error
	}

	return 0
}

// checkBrokenSymlinks finds broken symlinks in targets
func checkBrokenSymlinks(cfg *config.Config) int {
	issues := 0

	for name, target := range cfg.Targets {
		broken := findBrokenSymlinks(target.Path)
		if len(broken) > 0 {
			ui.Error("%s: %d broken symlink(s): %s", name, len(broken), strings.Join(broken, ", "))
			issues++
		}
	}

	return issues
}

func findBrokenSymlinks(dir string) []string {
	var broken []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return broken
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		info, err := os.Lstat(path)
		if err != nil {
			continue
		}

		if info.Mode()&os.ModeSymlink != 0 {
			// It's a symlink, check if target exists
			if _, err := os.Stat(path); os.IsNotExist(err) {
				broken = append(broken, entry.Name())
			}
		}
	}

	return broken
}

// checkDuplicateSkills finds skills with same name in multiple locations
func checkDuplicateSkills(cfg *config.Config) int {
	skillLocations := make(map[string][]string)

	// Collect from source
	entries, _ := os.ReadDir(cfg.Source)
	for _, entry := range entries {
		if entry.IsDir() && !utils.IsHidden(entry.Name()) {
			skillLocations[entry.Name()] = append(skillLocations[entry.Name()], "source")
		}
	}

	// Collect from targets (local-only skills)
	for name, target := range cfg.Targets {
		entries, err := os.ReadDir(target.Path)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() || utils.IsHidden(entry.Name()) {
				continue
			}

			// Check if it's a local skill (not a symlink to source)
			path := filepath.Join(target.Path, entry.Name())
			info, err := os.Lstat(path)
			if err != nil {
				continue
			}

			if info.Mode()&os.ModeSymlink == 0 {
				// It's a real directory, not a symlink
				skillLocations[entry.Name()] = append(skillLocations[entry.Name()], name)
			}
		}
	}

	// Find duplicates
	var duplicates []string
	for skill, locations := range skillLocations {
		if len(locations) > 1 {
			duplicates = append(duplicates, fmt.Sprintf("%s (%s)", skill, strings.Join(locations, ", ")))
		}
	}

	if len(duplicates) > 0 {
		sort.Strings(duplicates)
		ui.Warning("Duplicate skills: %s", strings.Join(duplicates, "; "))
	}

	return 0 // Warning only
}

// checkBackupStatus shows last backup time
func checkBackupStatus() {
	backupDir := filepath.Join(filepath.Dir(config.ConfigPath()), "backups")
	entries, err := os.ReadDir(backupDir)
	if err != nil || len(entries) == 0 {
		ui.Info("Backups: none found")
		return
	}

	// Find most recent backup
	var latest string
	var latestTime time.Time
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latest = entry.Name()
		}
	}

	if latest != "" {
		age := time.Since(latestTime)
		var ageStr string
		switch {
		case age < time.Hour:
			ageStr = fmt.Sprintf("%d minutes ago", int(age.Minutes()))
		case age < 24*time.Hour:
			ageStr = fmt.Sprintf("%d hours ago", int(age.Hours()))
		default:
			ageStr = fmt.Sprintf("%d days ago", int(age.Hours()/24))
		}
		ui.Info("Backups: last backup %s (%s)", latest, ageStr)
	}
}

// checkForUpdates checks if a newer version is available
func checkForUpdates() {
	// Use a short timeout for version check
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/runkids/skillshare/releases/latest")
	if err != nil {
		return // Silently skip if network unavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(version, "v")

	if latestVersion != currentVersion && latestVersion > currentVersion {
		ui.Info("Update available: %s -> %s", version, release.TagName)
		ui.Info("  brew upgrade skillshare  OR  curl -fsSL https://raw.githubusercontent.com/runkids/skillshare/main/install.sh | sh")
	}
}
