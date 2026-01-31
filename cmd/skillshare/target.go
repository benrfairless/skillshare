package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skillshare/internal/backup"
	"skillshare/internal/config"
	"skillshare/internal/sync"
	"skillshare/internal/ui"
	"skillshare/internal/utils"
	"skillshare/internal/validate"
)

func cmdTarget(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: skillshare target <add|remove|list|name> [options]")
	}

	subcmd := args[0]
	subargs := args[1:]

	switch subcmd {
	case "add":
		return targetAdd(subargs)
	case "remove", "rm":
		return targetRemove(subargs)
	case "list", "ls":
		return targetList()
	default:
		// Assume it's a target name - show info or modify settings
		return targetInfo(subcmd, subargs)
	}
}

func targetAdd(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: skillshare target add <name> <path>")
	}

	name := args[0]
	path := args[1]

	// Validate target name
	if err := validate.TargetName(name); err != nil {
		return fmt.Errorf("invalid target name: %w", err)
	}

	// Expand ~
	if utils.HasTildePrefix(path) {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot expand path: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	// Validate target path and get warnings
	warnings, err := validate.TargetPath(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Show warnings to user
	for _, w := range warnings {
		ui.Warning("%s", w)
	}

	// If path doesn't look like a skills directory, ask for confirmation
	if !validate.IsLikelySkillsPath(path) {
		ui.Warning("Path doesn't appear to be a skills directory")
		fmt.Print("  Continue anyway? [y/N]: ")
		var input string
		fmt.Scanln(&input)
		input = strings.ToLower(strings.TrimSpace(input))
		if input != "y" && input != "yes" {
			ui.Info("Cancelled")
			return nil
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if _, exists := cfg.Targets[name]; exists {
		return fmt.Errorf("target '%s' already exists", name)
	}

	cfg.Targets[name] = config.TargetConfig{Path: path}
	if err := cfg.Save(); err != nil {
		return err
	}

	ui.Success("Added target: %s -> %s", name, path)
	ui.Info("Run 'skillshare sync' to sync skills to this target")
	return nil
}

func targetRemove(args []string) error {
	// Check for --all flag
	removeAll := false
	dryRun := false
	var name string
	for _, arg := range args {
		switch arg {
		case "--all", "-a":
			removeAll = true
		case "--dry-run", "-n":
			dryRun = true
		default:
			name = arg
		}
	}

	if !removeAll && name == "" {
		return fmt.Errorf("usage: skillshare target remove <name> or --all")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var toRemove []string
	if removeAll {
		for n := range cfg.Targets {
			toRemove = append(toRemove, n)
		}
	} else {
		if _, exists := cfg.Targets[name]; !exists {
			return fmt.Errorf("target '%s' not found", name)
		}
		toRemove = []string{name}
	}

	if dryRun {
		return targetRemoveDryRun(cfg, toRemove)
	}

	// Backup before removing
	ui.Header("Backing up before unlink")
	for _, targetName := range toRemove {
		target := cfg.Targets[targetName]
		backupPath, err := backup.Create(targetName, target.Path)
		if err != nil {
			ui.Warning("Failed to backup %s: %v", targetName, err)
		} else if backupPath != "" {
			ui.Success("%s -> %s", targetName, backupPath)
		}
	}

	ui.Header("Unlinking targets")
	for _, targetName := range toRemove {
		target := cfg.Targets[targetName]

		// Check if it's a symlink (symlink mode) or has symlinked skills (merge mode)
		info, err := os.Lstat(target.Path)
		if err != nil {
			// Target doesn't exist, just remove from config
			delete(cfg.Targets, targetName)
			ui.Success("%s: removed from config", targetName)
			continue
		}

		if info.Mode()&os.ModeSymlink != 0 {
			// Symlink mode: remove symlink and copy source contents
			if err := unlinkSymlinkMode(target.Path, cfg.Source); err != nil {
				ui.Error("%s: %v", targetName, err)
				continue
			}
			ui.Success("%s: unlinked and restored", targetName)
		} else if info.IsDir() {
			// Merge mode: remove individual skill symlinks
			if err := unlinkMergeMode(target.Path, cfg.Source); err != nil {
				ui.Error("%s: %v", targetName, err)
				continue
			}
			ui.Success("%s: skill symlinks removed", targetName)
		}

		delete(cfg.Targets, targetName)
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	return nil
}

func targetRemoveDryRun(cfg *config.Config, toRemove []string) error {
	ui.Warning("Dry run mode - no changes will be made")

	ui.Header("Backing up before unlink")
	for _, targetName := range toRemove {
		ui.Info("%s: would attempt backup", targetName)
	}

	ui.Header("Unlinking targets")
	for _, targetName := range toRemove {
		target := cfg.Targets[targetName]
		info, err := os.Lstat(target.Path)
		if err != nil {
			if os.IsNotExist(err) {
				ui.Info("%s: would remove from config (path missing)", targetName)
				continue
			}
			ui.Warning("%s: %v", targetName, err)
			continue
		}

		if info.Mode()&os.ModeSymlink != 0 {
			ui.Info("%s: would unlink symlink and restore contents", targetName)
		} else if info.IsDir() {
			ui.Info("%s: would remove skill symlinks", targetName)
		}

		ui.Info("%s: would remove from config", targetName)
	}

	return nil
}

// unlinkSymlinkMode removes symlink and copies source contents back.
func unlinkSymlinkMode(targetPath, sourcePath string) error {
	// Remove the symlink
	if err := os.Remove(targetPath); err != nil {
		return fmt.Errorf("failed to remove symlink: %w", err)
	}

	// Copy source contents to target
	if err := copyDir(sourcePath, targetPath); err != nil {
		return fmt.Errorf("failed to copy skills: %w", err)
	}

	return nil
}

// unlinkMergeMode removes individual skill symlinks and copies them back.
func unlinkMergeMode(targetPath, sourcePath string) error {
	entries, err := os.ReadDir(targetPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		skillPath := filepath.Join(targetPath, entry.Name())
		info, err := os.Lstat(skillPath)
		if err != nil {
			continue
		}

		// Check if it's a symlink pointing to source
		if info.Mode()&os.ModeSymlink != 0 {
			link, _ := os.Readlink(skillPath)
			sourceSkillPath := filepath.Join(sourcePath, entry.Name())

			// Check if symlink points to our source
			absLink, _ := filepath.Abs(link)
			absSource, _ := filepath.Abs(sourceSkillPath)

			if utils.PathsEqual(absLink, absSource) {
				// Remove symlink and copy the skill back
				os.Remove(skillPath)
				if err := copyDir(sourceSkillPath, skillPath); err != nil {
					return fmt.Errorf("failed to copy %s: %w", entry.Name(), err)
				}
			}
		}
	}

	return nil
}

func targetList() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ui.Header("Configured Targets")
	for name, target := range cfg.Targets {
		mode := target.Mode
		if mode == "" {
			mode = "merge"
		}
		fmt.Printf("  %-12s %s (%s)\n", name, target.Path, mode)
	}

	return nil
}

func targetInfo(name string, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	target, exists := cfg.Targets[name]
	if !exists {
		return fmt.Errorf("target '%s' not found. Use 'skillshare target list' to see available targets", name)
	}

	// Parse flags
	var newMode string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--mode", "-m":
			if i+1 >= len(args) {
				return fmt.Errorf("--mode requires a value (merge or symlink)")
			}
			newMode = args[i+1]
			i++
		}
	}

	// If --mode is provided, update the mode
	if newMode != "" {
		return updateTargetMode(cfg, name, target, newMode)
	}

	// Show target info
	return showTargetInfo(cfg, name, target)
}

func updateTargetMode(cfg *config.Config, name string, target config.TargetConfig, newMode string) error {
	if newMode != "merge" && newMode != "symlink" {
		return fmt.Errorf("invalid mode '%s'. Use 'merge' or 'symlink'", newMode)
	}

	oldMode := target.Mode
	if oldMode == "" {
		oldMode = cfg.Mode
		if oldMode == "" {
			oldMode = "merge"
		}
	}

	target.Mode = newMode
	cfg.Targets[name] = target
	if err := cfg.Save(); err != nil {
		return err
	}

	ui.Success("Changed %s mode: %s -> %s", name, oldMode, newMode)
	ui.Info("Run 'skillshare sync' to apply the new mode")
	return nil
}

func showTargetInfo(cfg *config.Config, name string, target config.TargetConfig) error {
	mode := target.Mode
	if mode == "" {
		mode = cfg.Mode
		if mode == "" {
			mode = "merge"
		}
		mode = mode + " (default)"
	}

	status := sync.CheckStatus(target.Path, cfg.Source)

	ui.Header(fmt.Sprintf("Target: %s", name))
	fmt.Printf("  Path:   %s\n", target.Path)
	fmt.Printf("  Mode:   %s\n", mode)
	fmt.Printf("  Status: %s\n", status)

	return nil
}
