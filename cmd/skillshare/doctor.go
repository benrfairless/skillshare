package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	// Check each target
	issues += checkTargets(cfg)

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
