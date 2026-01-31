package main

import (
	"fmt"
	"os"
	"path/filepath"

	"skillshare/internal/config"
	"skillshare/internal/ui"
	"skillshare/internal/utils"
)

func cmdDiff(args []string) error {
	var targetName string
	for i := 0; i < len(args); i++ {
		if args[i] == "--target" || args[i] == "-t" {
			if i+1 < len(args) {
				targetName = args[i+1]
				i++
			}
		} else {
			targetName = args[i]
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Get source skills
	sourceSkills := getSourceSkills(cfg.Source)

	targets := cfg.Targets
	if targetName != "" {
		if t, exists := cfg.Targets[targetName]; exists {
			targets = map[string]config.TargetConfig{targetName: t}
		} else {
			return fmt.Errorf("target '%s' not found", targetName)
		}
	}

	for name, target := range targets {
		showTargetDiff(name, target, cfg.Source, sourceSkills)
	}

	return nil
}

func getSourceSkills(source string) map[string]bool {
	sourceSkills := make(map[string]bool)
	entries, _ := os.ReadDir(source)
	for _, e := range entries {
		if e.IsDir() && !utils.IsHidden(e.Name()) {
			sourceSkills[e.Name()] = true
		}
	}
	return sourceSkills
}

func showTargetDiff(name string, target config.TargetConfig, source string, sourceSkills map[string]bool) {
	ui.Header(fmt.Sprintf("Diff: %s", name))

	// Check if target is a symlink (symlink mode)
	targetInfo, err := os.Lstat(target.Path)
	if err != nil {
		ui.Warning("Cannot access target: %v", err)
		return
	}

	if targetInfo.Mode()&os.ModeSymlink != 0 {
		showSymlinkDiff(target.Path, source)
		return
	}

	// Merge mode - check individual skills
	showMergeDiff(target.Path, source, sourceSkills)
}

func showSymlinkDiff(targetPath, source string) {
	link, _ := os.Readlink(targetPath)
	absLink, _ := filepath.Abs(link)
	absSource, _ := filepath.Abs(source)
	if utils.PathsEqual(absLink, absSource) {
		ui.Success("Fully synced (symlink mode)")
	} else {
		ui.Warning("Symlink points to different location: %s", link)
	}
}

func showMergeDiff(targetPath, source string, sourceSkills map[string]bool) {
	targetSkills := make(map[string]bool)
	targetSymlinks := make(map[string]bool)
	entries, err := os.ReadDir(targetPath)
	if err != nil {
		ui.Warning("Cannot read target: %v", err)
		return
	}

	for _, e := range entries {
		if utils.IsHidden(e.Name()) {
			continue
		}
		skillPath := filepath.Join(targetPath, e.Name())
		info, _ := os.Lstat(skillPath)
		if info != nil && info.Mode()&os.ModeSymlink != 0 {
			targetSymlinks[e.Name()] = true
		}
		targetSkills[e.Name()] = true
	}

	// Compare
	hasChanges := false

	// Skills only in source (not synced)
	for skill := range sourceSkills {
		if !targetSkills[skill] {
			ui.DiffItem("add", skill, "(in source, not in target)")
			hasChanges = true
		} else if !targetSymlinks[skill] {
			ui.DiffItem("modify", skill, "(local copy, not linked)")
			hasChanges = true
		}
	}

	// Skills only in target (local only)
	for skill := range targetSkills {
		if !sourceSkills[skill] && !targetSymlinks[skill] {
			ui.DiffItem("remove", skill, "(local only, not in source)")
			hasChanges = true
		}
	}

	if !hasChanges {
		ui.Success("Fully synced (merge mode)")
	}
}
