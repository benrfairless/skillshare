package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/runkids/skillshare/internal/config"
	"github.com/runkids/skillshare/internal/sync"
	"github.com/runkids/skillshare/internal/ui"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "init":
		err = cmdInit(args)
	case "sync":
		err = cmdSync(args)
	case "status":
		err = cmdStatus(args)
	case "target":
		err = cmdTarget(args)
	case "version", "-v", "--version":
		fmt.Printf("skillshare %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		ui.Error("Unknown command: %s", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		ui.Error("%v", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`skillshare - Share skills across AI CLI tools

Usage:
  skillshare <command> [options]

Commands:
  init [--source PATH]       Initialize skillshare with a source directory
  sync [--dry-run] [--force] Sync skills to all targets
  status                     Show status of all targets
  target add <name> <path>   Add a target
  target remove <name>       Remove a target
  target list                List all targets
  version                    Show version
  help                       Show this help

Examples:
  skillshare init --source ~/.skills
  skillshare target add claude ~/.claude/skills
  skillshare sync
  skillshare status`)
}

func cmdInit(args []string) error {
	home, _ := os.UserHomeDir()
	sourcePath := filepath.Join(home, ".skills") // Default source

	// Parse args
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--source", "-s":
			if i+1 >= len(args) {
				return fmt.Errorf("--source requires a path argument")
			}
			sourcePath = args[i+1]
			i++
		}
	}

	// Expand ~ in path
	if len(sourcePath) > 0 && sourcePath[0] == '~' {
		sourcePath = filepath.Join(home, sourcePath[1:])
	}

	// Check if already initialized
	if _, err := os.Stat(config.ConfigPath()); err == nil {
		return fmt.Errorf("already initialized. Config at: %s", config.ConfigPath())
	}

	// Create source directory
	if err := os.MkdirAll(sourcePath, 0755); err != nil {
		return fmt.Errorf("failed to create source directory: %w", err)
	}

	// Detect existing targets
	ui.Header("Detecting CLI skills directories")
	defaultTargets := config.DefaultTargets()
	targets := make(map[string]config.TargetConfig)

	for name, target := range defaultTargets {
		if info, err := os.Stat(target.Path); err == nil {
			if info.IsDir() {
				targets[name] = target
				ui.Success("Found: %s (%s)", name, target.Path)
			}
		} else {
			// Check if parent exists (CLI is installed but no skills yet)
			parent := filepath.Dir(target.Path)
			if _, err := os.Stat(parent); err == nil {
				targets[name] = target
				ui.Info("Available: %s (%s)", name, target.Path)
			}
		}
	}

	if len(targets) == 0 {
		ui.Warning("No CLI skills directories detected. You can add targets manually.")
	}

	// Create config
	cfg := &config.Config{
		Source:  sourcePath,
		Mode:    "merge",
		Targets: targets,
		Ignore: []string{
			"**/.DS_Store",
			"**/.git/**",
		},
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	ui.Header("Initialized successfully")
	ui.Success("Source: %s", sourcePath)
	ui.Success("Config: %s", config.ConfigPath())
	ui.Info("Run 'skillshare sync' to sync your skills")

	return nil
}

func cmdSync(args []string) error {
	dryRun := false
	force := false

	for _, arg := range args {
		switch arg {
		case "--dry-run", "-n":
			dryRun = true
		case "--force", "-f":
			force = true
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Ensure source exists
	if _, err := os.Stat(cfg.Source); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist: %s", cfg.Source)
	}

	ui.Header("Syncing skills")
	if dryRun {
		ui.Warning("Dry run mode - no changes will be made")
	}

	hasError := false
	for name, target := range cfg.Targets {
		// Determine mode: target-specific > global > default
		mode := target.Mode
		if mode == "" {
			mode = cfg.Mode
		}
		if mode == "" {
			mode = "merge"
		}

		if mode == "merge" {
			// Merge mode: create individual skill symlinks
			result, err := sync.SyncTargetMerge(name, target, cfg.Source, dryRun)
			if err != nil {
				ui.Error("%s: %v", name, err)
				hasError = true
				continue
			}

			if len(result.Linked) > 0 || len(result.Updated) > 0 {
				ui.Success("%s: merged (%d linked, %d local, %d updated)",
					name, len(result.Linked), len(result.Skipped), len(result.Updated))
			} else if len(result.Skipped) > 0 {
				ui.Success("%s: merged (%d local skills preserved)", name, len(result.Skipped))
			} else {
				ui.Success("%s: merged (no skills)", name)
			}
			continue
		}

		// Symlink mode (default)
		status := sync.CheckStatus(target.Path, cfg.Source)

		// Handle conflicts
		if status == sync.StatusConflict && !force {
			link, _ := os.Readlink(target.Path)
			ui.Error("%s: conflict - symlink points to %s (use --force to override)", name, link)
			hasError = true
			continue
		}

		if status == sync.StatusConflict && force {
			if !dryRun {
				os.Remove(target.Path)
			}
		}

		if err := sync.SyncTarget(name, target, cfg.Source, dryRun); err != nil {
			ui.Error("%s: %v", name, err)
			hasError = true
			continue
		}

		switch status {
		case sync.StatusLinked:
			ui.Success("%s: already linked", name)
		case sync.StatusNotExist:
			ui.Success("%s: symlink created", name)
		case sync.StatusHasFiles:
			ui.Success("%s: files migrated and linked", name)
		case sync.StatusBroken:
			ui.Success("%s: broken link fixed", name)
		case sync.StatusConflict:
			ui.Success("%s: conflict resolved (forced)", name)
		}
	}

	if hasError {
		return fmt.Errorf("some targets failed to sync")
	}

	return nil
}

func cmdStatus(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ui.Header("Source")
	if info, err := os.Stat(cfg.Source); err == nil {
		entries, _ := os.ReadDir(cfg.Source)
		skillCount := 0
		for _, e := range entries {
			if e.IsDir() && e.Name()[0] != '.' {
				skillCount++
			}
		}
		ui.Success("%s (%d skills, %s)", cfg.Source, skillCount, info.ModTime().Format("2006-01-02 15:04"))
	} else {
		ui.Error("%s (not found)", cfg.Source)
	}

	ui.Header("Targets")
	for name, target := range cfg.Targets {
		// Determine mode
		mode := target.Mode
		if mode == "" {
			mode = cfg.Mode
		}
		if mode == "" {
			mode = "merge"
		}

		var statusStr, detail string

		if mode == "merge" {
			status, linkedCount, localCount := sync.CheckStatusMerge(target.Path, cfg.Source)
			if status == sync.StatusMerged {
				statusStr = "merged"
				detail = fmt.Sprintf("%s (%d shared, %d local)", target.Path, linkedCount, localCount)
			} else if status == sync.StatusLinked {
				statusStr = "linked"
				detail = fmt.Sprintf("%s (using symlink mode)", target.Path)
			} else {
				statusStr = status.String()
				detail = fmt.Sprintf("%s (%d local)", target.Path, localCount)
			}
		} else {
			status := sync.CheckStatus(target.Path, cfg.Source)
			statusStr = status.String()
			detail = target.Path

			if status == sync.StatusConflict {
				link, _ := os.Readlink(target.Path)
				detail = fmt.Sprintf("%s -> %s", target.Path, link)
			}
		}

		ui.Status(name, statusStr, detail)
	}

	return nil
}

func cmdTarget(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: skillshare target <add|remove|list> [args]")
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
		return fmt.Errorf("unknown target subcommand: %s", subcmd)
	}
}

func targetAdd(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: skillshare target add <name> <path>")
	}

	name := args[0]
	path := args[1]

	// Expand ~
	if len(path) > 0 && path[0] == '~' {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[1:])
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
	return nil
}

func targetRemove(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: skillshare target remove <name>")
	}

	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if _, exists := cfg.Targets[name]; !exists {
		return fmt.Errorf("target '%s' not found", name)
	}

	delete(cfg.Targets, name)
	if err := cfg.Save(); err != nil {
		return err
	}

	ui.Success("Removed target: %s", name)
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
			mode = "symlink"
		}
		fmt.Printf("  %-12s %s (%s)\n", name, target.Path, mode)
	}

	return nil
}
